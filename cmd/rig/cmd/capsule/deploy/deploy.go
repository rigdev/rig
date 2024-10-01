package deploy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"
	container_name "github.com/google/go-containerregistry/pkg/name"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	api_image "github.com/rigdev/rig-go-api/api/v1/image"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	v1 "github.com/rigdev/rig/pkg/api/platform/v1"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

var fileRefRegExp = regexp.MustCompile("path=([^,]*),obj=(.*)/(.*)/(.*)")

func parseEnvironmentSource(value string) (*platformv1.EnvironmentSource, error) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) != 2 {
		return nil, errors.InvalidArgumentErrorf("invalid --env-source format: %s", value)
	}

	switch strings.ToLower(parts[0]) {
	case "configmap":
		return &platformv1.EnvironmentSource{
			Name: parts[1],
			Kind: "ConfigMap",
		}, nil
	case "secret":
		return &platformv1.EnvironmentSource{
			Name: parts[1],
			Kind: "Secret",
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("invalid --env-source kind, must be ConfigMap or Secret: %s", value)
	}
}

func (c *Cmd) deploy(ctx context.Context, cmd *cobra.Command, args []string) error {
	if currentFingerprint != "" && currentRolloutID != 0 {
		return errors.InvalidArgumentErrorf("cannot give both --fingerprint and --current-rollout")
	}

	capsule, err := c.getNewSpec(ctx, cmd, args)
	if err != nil {
		return err
	}

	if _, err := c.Rig.Capsule().Get(ctx, &connect.Request[capsule_api.GetRequest]{
		Msg: &capsule_api.GetRequest{
			CapsuleId: capsule.GetName(),
			ProjectId: capsule.GetProject(),
		},
	}); errors.IsNotFound(err) {
		fmt.Printf("Capsule `%s` doesn't exist, creating Capsule\n", capsule.GetName())
		if _, err := c.Rig.Capsule().Create(ctx, &connect.Request[capsule_api.CreateRequest]{
			Msg: &capsule_api.CreateRequest{
				Name:      capsule.GetName(),
				ProjectId: capsule.GetProject(),
			},
		}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	respGit, err := c.Rig.Capsule().GetEffectiveGitSettings(
		ctx, connect.NewRequest(&capsule_api.GetEffectiveGitSettingsRequest{
			ProjectId:     capsule.GetProject(),
			EnvironmentId: capsule.GetEnvironment(),
			CapsuleId:     capsule.GetName(),
		}),
	)
	if errors.IsUnimplemented(err) {
		respGit = &connect.Response[capsule_api.GetEffectiveGitSettingsResponse]{}
	} else if err != nil {
		return err
	}

	changes := []*capsule_api.Change{{
		Field: &capsule_api.Change_Spec{
			Spec: capsule.GetSpec(),
		},
	}}

	if respGit.Msg.GetEnvironmentEnabled() && prBranchName != "" {
		resp, err := c.Rig.Capsule().ProposeRollout(ctx, connect.NewRequest(&capsule_api.ProposeRolloutRequest{
			CapsuleId:     capsule.GetName(),
			Changes:       changes,
			ProjectId:     capsule.GetProject(),
			EnvironmentId: capsule.GetProject(),
			BranchName:    prBranchName,
		}))
		if err != nil {
			return err
		}
		url := resp.Msg.GetProposal().GetMetadata().GetReviewUrl()
		fmt.Println("New pull request created at", url)
		return nil
	} else if !respGit.Msg.GetEnvironmentEnabled() && prBranchName != "" {
		return errors.InvalidArgumentErrorf("--pr-branch was set, but the capsule is not git backed")
	}

	var rollbackID uint64
	if !noRollback {
		// TODO: Get this from the Deploy command instead.
		rollbackID = currentRolloutID
		if rollbackID == 0 {
			res, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule_api.ListRolloutsRequest]{
				Msg: &capsule_api.ListRolloutsRequest{
					Pagination: &model.Pagination{
						Limit:      1,
						Descending: true,
					},
					ProjectId:     capsule.GetProject(),
					EnvironmentId: capsule.GetEnvironment(),
					CapsuleId:     capsule.GetName(),
				},
			})
			if err != nil {
				return err
			}

			if len(res.Msg.GetRollouts()) > 0 {
				rollbackID = res.Msg.GetRollouts()[0].GetRolloutId()
			}
		}
	}

	baseInput := capsule_cmd.BaseInput{
		Ctx:           ctx,
		Rig:           c.Rig,
		ProjectID:     capsule.GetProject(),
		EnvironmentID: capsule.GetEnvironment(),
		CapsuleID:     capsule.GetName(),
	}

	if dry {
		input := capsule_cmd.DeployDryInput{
			BaseInput:          baseInput,
			Changes:            changes,
			Scheme:             c.Scheme,
			CurrentRolloutID:   currentRolloutID,
			CurrentFingerprint: parseFingerprint(currentFingerprint),
			IsInteractive:      c.Scope.IsInteractive(),
		}

		return capsule_cmd.DeployDry(input)
	}

	input := capsule_cmd.DeployAndWaitInput{
		DeployInput: capsule_cmd.DeployInput{
			BaseInput:          baseInput,
			Changes:            changes,
			ForceDeploy:        true,
			ForceOverride:      forceOverride,
			CurrentRolloutID:   currentRolloutID,
			CurrentFingerprint: parseFingerprint(currentFingerprint),
		},
		Timeout:    timeout,
		RollbackID: rollbackID,
		NoWait:     noWait,
	}
	return capsule_cmd.DeployAndWait(input)
}

func (c *Cmd) getNewSpec(ctx context.Context, cmd *cobra.Command, args []string) (*platformv1.Capsule, error) {
	if file != "" {
		if environmentVariables != nil ||
			removeEnvironmentVariables != nil ||
			environmentSources != nil ||
			removeEnvironmentSources != nil ||
			annotations != nil ||
			removeAnnotations != nil ||
			cmd.Flag("replicas").Changed ||
			configFiles != nil ||
			configFileRefs != nil ||
			removeConfigFiles != nil ||
			networkInterfaces != nil ||
			removeNetworkInterfaces != nil {
			return nil, errors.InvalidArgumentErrorf("cannot supply both --file and another configuration flag")
		}
		bytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		spec, err := v1.CapsuleYAMLToProto(bytes)
		if err != nil {
			return nil, err
		}
		return spec, nil
	}

	// Prompt for project and env if necessary
	projectID := c.Scope.GetCurrentContext().GetProject()
	environmentID := c.Scope.GetCurrentContext().GetEnvironment()
	capsuleID, err := c.getCapsuleID(ctx, args)
	if err != nil {
		return nil, err
	}

	spec := v1.NewCapsuleProto(projectID, environmentID, capsuleID, nil)
	resp, err := c.Rig.Capsule().Get(ctx, connect.NewRequest(&capsule_api.GetRequest{
		CapsuleId: capsuleID,
		ProjectId: projectID,
	}))
	if errors.IsNotFound(err) {
	} else if err != nil {
		return nil, err
	} else {
		for _, env := range resp.Msg.GetEnvironmentRevisions() {
			if env.GetSpec().GetEnvironment() == environmentID {
				spec = env.GetSpec()
				v1.InitialiseProto(spec)
			}
		}
	}

	// Annotations.
	for _, key := range removeAnnotations {
		delete(spec.Spec.Annotations, key)
	}
	for key, value := range annotations {
		if spec.Spec.Annotations == nil {
			spec.Spec.Annotations = map[string]string{}
		}
		spec.Spec.Annotations[key] = value
	}

	// Environment variables.
	for _, key := range removeEnvironmentVariables {
		delete(spec.Spec.GetEnv().GetRaw(), key)
	}
	for key, value := range environmentVariables {
		spec.Spec.Env.Raw[key] = value
	}

	// Environment sources.
	for _, value := range removeEnvironmentSources {
		source, err := parseEnvironmentSource(value)
		if err != nil {
			return nil, err
		}
		spec.Spec.Env.Sources = removeFromList(
			spec.Spec.Env.Sources, source,
			func(s1, s2 *platformv1.EnvironmentSource) bool {
				return s1.Name == s2.Name && s1.Kind == s2.Kind
			},
		)
	}
	for _, value := range environmentSources {
		source, err := parseEnvironmentSource(value)
		if err != nil {
			return nil, err
		}
		spec.Spec.Env.Sources = insertInList(
			spec.Spec.Env.Sources, source,
			func(s *platformv1.EnvironmentSource, s2 *platformv1.EnvironmentSource) bool {
				return s.GetName() == s2.GetName() && s.GetKind() == s2.GetKind()
			},
		)
	}

	// Image.
	if imageID != "" {
		spec.Spec.Image = imageID
	}

	// Network interfaces.
	for _, file := range networkInterfaces {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, errors.InvalidArgumentErrorf("errors reading network interface: %v", err)
		}

		raw, err := yaml.YAMLToJSON(data)
		if err != nil {
			return nil, err
		}

		ci := &platformv1.CapsuleInterface{}
		if err := protojson.Unmarshal(raw, ci); err != nil {
			// Backwards compatibility if anyone used the old format to deploy interfaces
			oldCI := &capsule.Interface{}
			if err2 := protojson.Unmarshal(raw, oldCI); err2 != nil {
				return nil, fmt.Errorf("error parsing network interface from %s: %w", file, err)
			}
			ci, err = v1.InterfaceConversion(oldCI)
			if err != nil {
				return nil, fmt.Errorf("error parsing network interface from %s: %w", file, err)
			}
		}
		found := false
		for idx, i := range spec.Spec.GetInterfaces() {
			if i.GetName() == ci.GetName() {
				spec.Spec.Interfaces[idx] = ci
				found = true
				break
			}
		}
		if !found {
			spec.Spec.Interfaces = append(spec.Spec.Interfaces, ci)
		}
	}

	for _, name := range removeNetworkInterfaces {
		for idx, i := range spec.Spec.GetInterfaces() {
			if i.GetName() == name {
				spec.Spec.Interfaces = append(spec.Spec.Interfaces[:idx], spec.Spec.Interfaces[idx+1:]...)
			}
		}
	}

	// Config Files
	for _, target := range removeConfigFiles {
		for idx, f := range spec.Spec.GetFiles() {
			if f.GetPath() == target {
				spec.Spec.Files = append(spec.Spec.Files[:idx], spec.Spec.Files[idx+1:]...)
			}
		}
	}

	for _, configFile := range configFiles {
		var target string
		var source string
		var secret bool
		for _, opt := range strings.Split(configFile, ",") {
			opt = strings.TrimSpace(opt)
			if v, ok := strings.CutPrefix(opt, "path="); ok {
				target = v
			} else if v, ok := strings.CutPrefix(opt, "src="); ok {
				source = v
			} else if opt == "secret" {
				secret = true
			} else {
				return nil, errors.InvalidArgumentErrorf("invalid config-file argument: %v", configFile)
			}
		}
		if err := validateConfigFilePath(target, "config-file"); err != nil {
			return nil, err
		}
		bs, err := os.ReadFile(source)
		if err != nil {
			return nil, err
		}

		if !utf8.Valid(bs) {
			return nil, errors.InvalidArgumentErrorf("source file is not valid UTF-8: %v", source)
		}
		file := &platformv1.File{
			Path:     target,
			AsSecret: secret,
			Bytes:    bs,
			String_:  string(bs),
		}
		spec.Spec.Files = insertInList(spec.Spec.Files, file, func(f1, f2 *platformv1.File) bool {
			return f1.GetPath() == f2.GetPath()
		})
	}
	for _, configFileRef := range configFileRefs {
		matches := fileRefRegExp.FindStringSubmatch(configFileRef)
		if len(matches) != 5 {
			return nil, errors.InvalidArgumentErrorf("config-file-ref does not match the format")
		}

		target, kind, name, key := matches[1], matches[2], matches[3], matches[4]
		if kind != "Secret" && kind != "ConfigMap" {
			return nil, errors.InvalidArgumentErrorf(
				"config-file-ref kind must be either Secret or Configmap, was '%s'", kind,
			)
		}
		if err := common.ValidateKubernetesName(name); err != nil {
			return nil, errors.InvalidArgumentErrorf("config-file-ref name '%s' was invalid: %w", name, err)
		}
		if err := validateConfigFilePath(target, "config-file-ref"); err != nil {
			return nil, err
		}

		file := &platformv1.File{
			Path: target,
			Ref: &platformv1.FileReference{
				Kind: kind,
				Name: name,
				Key:  key,
			},
		}
		spec.Spec.Files = insertInList(spec.Spec.Files, file, func(f1, f2 *platformv1.File) bool {
			return f1.GetPath() == f2.GetPath()
		})
	}

	// Replicas.
	if cmd.Flag("replicas").Changed {
		if replicas < 0 {
			return nil, errors.InvalidArgumentErrorf("number of replicas cannot be negative: %v", replicas)
		}
		spec.Spec.Scale.Horizontal.Instances.Min = uint32(replicas)
	}

	// Command and arguments.
	if idx := cmd.ArgsLenAtDash(); idx >= 0 {
		extraArgs := args[idx:]
		if len(extraArgs) == 0 {
			spec.Spec.Command = ""
			spec.Spec.Args = nil
		} else {
			spec.Spec.Command = extraArgs[0]
			spec.Spec.Args = extraArgs[1:]
		}
	}

	if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
		return nil, err
	}

	return spec, nil
}

func insertInList[T any](existing []T, obj T, equal func(T, T) bool) []T {
	for idx, o := range existing {
		if equal(o, obj) {
			existing[idx] = obj
			return existing
		}
	}
	return append(existing, obj)
}

func removeFromList[T any, K any](existing []T, key K, equal func(T, K) bool) []T {
	for idx, o := range existing {
		if equal(o, key) {
			return append(existing[:idx], existing[idx+1:]...)
		}
	}
	return existing
}

func validateConfigFilePath(p string, s string) error {
	if !path.IsAbs(p) {
		return errors.InvalidArgumentErrorf("%s path is not absolute: %v", s, p)
	}
	if path.Clean(p) != p {
		return errors.InvalidArgumentErrorf(
			"%s path is not clean: %v should be %s", s, p, path.Clean(p),
		)
	}
	if strings.HasSuffix(p, "/") {
		return errors.InvalidArgumentErrorf("%s path should not end with a '/': %v", s, p)
	}
	return nil
}

func (c *Cmd) GetImageID(ctx context.Context, capsuleID string) (string, error) {
	if imageID != "" {
		// TODO Figure out pagination
		resp, err := c.Rig.Image().List(ctx, connect.NewRequest(&api_image.ListRequest{
			CapsuleId: capsuleID,
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
		}))
		if err != nil {
			return "", err
		}
		images := resp.Msg.GetImages()
		return expandImageID(images, imageID)
	}

	return c.promptForDockerOrImage(ctx, capsuleID)
}

func expandImageID(images []*capsule_api.Image, imageID string) (string, error) {
	if strings.HasPrefix(imageID, "sha256:") {
		return expandByDigestPrefix(imageID, images)
	}
	if isHexString(imageID) {
		return expandByDigestPrefix("sha256:"+imageID, images)
	}
	if strings.Contains(imageID, "@") {
		return expandByDigestName(imageID, images)
	}
	if ref, err := container_name.NewTag(imageID); err == nil {
		return expandByLatestTag(ref, images)
	}

	return "", errors.New("unable to parse image")
}

func expandByDigestName(imageID string, images []*capsule_api.Image) (string, error) {
	idx := strings.Index(imageID, "@")
	name := imageID[:idx]
	digest := imageID[idx+1:]
	tag, err := container_name.NewTag(name)
	if err != nil {
		return "", err
	}
	var validImages []*capsule_api.Image
	for _, b := range images {
		repoMatch := b.GetRepository() == fmt.Sprintf("%s/%s", tag.RegistryStr(), tag.RepositoryStr())
		tagMatch := b.GetTag() == tag.TagStr()
		digMatch := strings.HasPrefix(b.GetDigest(), digest)
		if repoMatch && tagMatch && digMatch {
			validImages = append(validImages, b)
		}
	}

	if len(validImages) == 0 {
		return "", errors.New("no images matched the image name and digest prefix")
	}
	if len(validImages) > 1 {
		return "", errors.New("the image name and digest prefix was not unique")
	}

	return validImages[0].GetImageId(), nil
}

func expandByLatestTag(ref container_name.Reference, images []*capsule_api.Image) (string, error) {
	var latest *capsule_api.Image
	for _, i := range images {
		if i.GetRepository() != fmt.Sprintf("%s/%s", ref.Context().RegistryStr(), ref.Context().RepositoryStr()) ||
			i.GetTag() != ref.Identifier() {
			continue
		}
		if latest == nil || latest.CreatedAt.AsTime().Before(i.CreatedAt.AsTime()) {
			latest = i
		}
	}

	if latest == nil {
		return "", errors.New("no images matched the given image name")
	}

	return latest.GetImageId(), nil
}

func expandByDigestPrefix(digestPrefix string, images []*capsule_api.Image) (string, error) {
	var validImages []*capsule_api.Image
	for _, b := range images {
		if strings.HasPrefix(b.GetDigest(), digestPrefix) {
			validImages = append(validImages, b)
		}
	}
	if len(validImages) > 1 {
		return "", errors.New("digest prefix was not unique")
	}
	if len(validImages) == 0 {
		return "", errors.New("no images had a matching digest prefix")
	}
	return validImages[0].GetImageId(), nil
}

func isHexString(s string) bool {
	s = strings.ToLower(s)
	for _, c := range s {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f')) {
			return false
		}
	}
	return true
}

func (c *Cmd) promptForDockerOrImage(ctx context.Context, capsuleID string) (string, error) {
	i, _, err := c.Prompter.Select(
		"Deploy from docker image rig-registered image?",
		[]string{"Docker", "Rig registered"},
	)
	if err != nil {
		return "", err
	}
	switch i {
	case 0:
		imgRef, err := c.promptForImage(ctx)
		if err != nil {
			return "", err
		}
		return c.createImageInner(ctx, capsuleID, imgRef)
	case 1:
		return c.promptForExistingImage(ctx, capsuleID)
	default:
		return "", errors.New("something went wrong")
	}
}

func (c *Cmd) promptForExistingImage(ctx context.Context, capsuleID string) (string, error) {
	resp, err := c.Rig.Image().List(ctx, connect.NewRequest(&api_image.ListRequest{
		CapsuleId:  capsuleID,
		Pagination: &model.Pagination{},
		ProjectId:  c.Scope.GetCurrentContext().GetProject(),
	}))
	if err != nil {
		return "", err
	}
	images := resp.Msg.GetImages()
	slices.SortFunc(images, func(b1, b2 *capsule_api.Image) int {
		t1 := b1.CreatedAt.AsTime()
		t2 := b2.CreatedAt.AsTime()
		if t1.Equal(t2) {
			return 0
		}
		if t1.Before(t2) {
			return 1
		}
		return -1
	})

	if len(images) == 0 {
		return "", errors.New("capsule has no images")
	}

	var rows [][]string
	for _, b := range images {
		rows = append(rows, []string{
			fmt.Sprint(b.GetRepository(), ":", b.GetTag()),
			capsule_cmd.TruncatedFixed(b.GetDigest(), 19),
			common.FormatDuration(time.Since(b.GetCreatedAt().AsTime())),
		})
	}

	idx, err := c.Prompter.TableSelect(
		"Select a Rig image",
		rows,
		[]string{"Image name", "Digest", "Age"},
		common.SelectFuzzyFilterOpt,
	)
	if err != nil {
		return "", err
	}

	return images[idx].GetImageId(), nil
}

func (c *Cmd) pushLocalImageToDevRegistry(ctx context.Context, image string) (string, string, error) {
	resp, err := c.Rig.Cluster().GetConfigs(ctx, connect.NewRequest(&cluster.GetConfigsRequest{}))
	if err != nil {
		return "", "", err
	}

	clusters := resp.Msg.Clusters
	if len(clusters) != 1 {
		return "", "", errors.New("cannot push local images to dev registry if there are more than one cluster")
	}
	config := clusters[0]

	switch config.GetDevRegistry().(type) {
	case *cluster.GetConfigResponse_Docker:
		return image, "", nil
	}
	devRegistry := config.GetRegistry()
	if devRegistry == nil {
		return "", "", fmt.Errorf("no dev-registry configured") // TODO Help the user with fixing this
	}

	newImageName, err := makeDevRegistryImageName(image, devRegistry.Host)
	if err != nil {
		return "", "", err
	}

	fmt.Printf("Pushing the image to the dev docker registry under the new name %q\n", newImageName)

	if err := c.DockerClient.ImageTag(ctx, image, newImageName); err != nil {
		return "", "", err
	}

	digest, err := c.pushToDevRegistry(ctx, newImageName, devRegistry.Host)
	if err != nil {
		return "", "", err
	}

	return newImageName, digest, nil
}

func makeDevRegistryImageName(image string, devRegistryHost string) (string, error) {
	r, err := container_name.NewRegistry(devRegistryHost, container_name.Insecure)
	if err != nil {
		return "", err
	}
	ref, err := container_name.ParseReference(image)
	if err != nil {
		return "", err
	}
	repo := r.Repo(ref.Context().RepositoryStr())
	tag := repo.Tag(ref.Identifier())
	return tag.String(), nil
}

func (c *Cmd) pushToDevRegistry(ctx context.Context, image string, host string) (string, error) {
	ac := registry.AuthConfig{
		ServerAddress: host,
	}
	secret, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	rc, err := c.DockerClient.ImagePush(ctx, image, types.ImagePushOptions{
		RegistryAuth: base64.StdEncoding.EncodeToString(secret),
	})
	if err != nil {
		return "", err
	}

	defer rc.Close()

	decoder := json.NewDecoder(rc)
	progressWriter := progress.NewWriter()
	progressWriter.SetAutoStop(true)
	trackers := map[string]*progress.Tracker{}

	go progressWriter.Render()
	var digest string
	for decoder.More() {
		var p dockerProgress
		if err := decoder.Decode(&p); err != nil {
			return "", err
		}
		if p.ID == "" || p.ProgressDetail.Total == 0 {
			continue
		}
		tracker, ok := trackers[p.ID]
		if !ok {
			tracker = &progress.Tracker{
				Message: p.ID,
				Total:   int64(p.ProgressDetail.Total),
				Units:   progress.UnitsBytes,
			}
			trackers[p.ID] = tracker
			progressWriter.AppendTracker(tracker)
		}
		if p.ProgressDetail.Current != 0 {
			tracker.SetValue(int64(p.ProgressDetail.Current))
		}
		if p.Aux.Digest != "" {
			digest = p.Aux.Digest
		}
	}

	return digest, nil
}

type dockerProgress struct {
	Status         string
	ID             string
	ProgressDetail struct {
		Current uint64
		Total   uint64
	}
	Aux struct {
		Tag    string
		Digest string
	}
}

func (c *Cmd) promptForImage(ctx context.Context) (imageRef, error) {
	var empty imageRef

	ok, err := c.Prompter.Confirm("Use a local image?", true)
	if err != nil {
		return empty, err
	}

	if ok {
		img, err := c.getDaemonImage(ctx)
		if err != nil {
			return empty, err
		}
		return imageRef{
			Image:        img.tag,
			IsKnownLocal: ptr.New(true),
		}, nil
	}

	imageName, err := c.Prompter.Input("Enter image:", common.ValidateImageOpt)
	if err != nil {
		return empty, nil
	}
	return imageRef{
		Image:        imageName,
		IsKnownLocal: ptr.New(false),
	}, nil
}

func (c *Cmd) getDaemonImage(ctx context.Context) (*imageInfo, error) {
	images, prompts, err := c.getImagePrompts(ctx)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, errors.New("no local docker images found")
	}
	idx, err := c.Prompter.TableSelect(
		"Select image:", prompts, []string{"Image name", "Age"}, common.SelectEnableFilterOpt,
	)
	if err != nil {
		return nil, err
	}
	return &images[idx], nil
}

func (c *Cmd) getImagePrompts(ctx context.Context) ([]imageInfo, [][]string, error) {
	res, err := c.DockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("dangling", "false")),
	})
	if err != nil {
		return nil, nil, err
	}

	var images []imageInfo
	var prompts [][]string

	for _, image := range res {
		for _, tag := range image.RepoTags {
			t := time.Unix(image.Created, 0)
			ii, _, err := c.DockerClient.ImageInspectWithRaw(ctx, tag)
			if err != nil {
				return nil, nil, err
			}
			if !ii.Metadata.LastTagTime.IsZero() {
				t = ii.Metadata.LastTagTime
			}
			images = append(images, imageInfo{
				tag:     tag,
				created: t,
			})
		}
	}

	slices.SortFunc(images, func(i, j imageInfo) int {
		if i.created.Equal(j.created) {
			return 0
		}
		if i.created.Before(j.created) {
			return 1
		}
		return -1
	})

	for idx, image := range images {
		if idx >= 50 {
			break
		}
		t := time.Since(image.created)
		prompts = append(prompts, []string{image.tag, common.FormatDuration(t)})
	}
	return images, prompts, nil
}

type imageInfo struct {
	tag     string
	created time.Time
}

type imageRef struct {
	Image string
	// &true: we know it's local
	// &false: we know it's remote
	// nil: we don't know
	IsKnownLocal *bool
}

func (c *Cmd) createImageInner(ctx context.Context, capsuleID string, imageRef imageRef) (string, error) {
	if strings.Contains(imageRef.Image, "@") {
		return "", errors.UnimplementedErrorf("referencing images by digest is not yet supported")
	}

	var err error
	var isLocalImage bool
	if imageRef.IsKnownLocal == nil {
		isLocalImage, _, err = utils.ImageExistsNatively(ctx, c.DockerClient, imageRef.Image)
		if err != nil {
			return "", err
		}
	} else {
		isLocalImage = *imageRef.IsKnownLocal
	}

	var digest string
	if isLocalImage {
		imageRef.Image, digest, err = c.pushLocalImageToDevRegistry(ctx, imageRef.Image)
		if err != nil {
			return "", err
		}
	}

	res, err := c.Rig.Image().Add(ctx, connect.NewRequest(&api_image.AddRequest{
		CapsuleId:      capsuleID,
		Image:          imageRef.Image,
		Digest:         digest,
		SkipImageCheck: skipImageCheck,
		ProjectId:      c.Scope.GetCurrentContext().GetProject(),
	}))
	if err != nil {
		return "", err
	}

	if res.Msg.GetAddedNewImage() {
		fmt.Println("Added new image:", res.Msg.GetImageId())
	} else {
		fmt.Println("Image already exists, using existing image")
	}

	return res.Msg.GetImageId(), nil
}

func (c *Cmd) getCapsuleID(ctx context.Context, args []string) (string, error) {
	var capsuleName string
	if len(args) > 0 {
		capsuleName = args[0]
	}

	if len(capsuleName) == 0 {
		if !c.Scope.IsInteractive() {
			return "", errors.InvalidArgumentErrorf("missing capsule name argument")
		}

		name, err := capsule_cmd.SelectCapsule(ctx, c.Rig, c.Prompter, c.Scope)
		if err != nil {
			return "", err
		}

		capsuleName = name
	}

	return capsuleName, nil
}

func parseFingerprint(s string) *model.Fingerprint {
	if s == "" {
		return nil
	}
	return &model.Fingerprint{
		Data: s,
	}
}
