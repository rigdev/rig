package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	capsule_service "github.com/rigdev/rig/internal/service/capsule"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	capsuleName string
	buildImage  string

	listOffset int
	listLimit  int
)

func init() {
	capsule := &cobra.Command{
		Use: "capsule",
	}

	create := &cobra.Command{
		Use:  "create <capsule-name>",
		Args: cobra.ExactArgs(1),
		RunE: register(CapsuleCreate),
	}
	capsule.AddCommand(create)

	delete := &cobra.Command{
		Use:  "delete <capsule-id>",
		Args: cobra.ExactArgs(1),
		RunE: register(CapsuleDelete),
	}
	capsule.AddCommand(delete)

	update := &cobra.Command{
		Use:  "update <capsule-id>",
		Args: cobra.ExactArgs(1),
		RunE: register(CapsuleUpdate),
	}
	update.PersistentFlags().StringVarP(&capsuleName, "name", "n", "", "name of the capsule")
	capsule.AddCommand(update)

	list := &cobra.Command{
		Use:  "list",
		Args: cobra.ExactArgs(0),
		RunE: register(CapsuleList),
	}
	list.PersistentFlags().IntVarP(&listOffset, "offset", "o", 0, "offset for the list")
	list.PersistentFlags().IntVarP(&listLimit, "limit", "l", 10, "limit for the list")
	capsule.AddCommand(list)

	createBuild := &cobra.Command{
		Use:  "create-build <capsule-id>",
		Args: cobra.ExactArgs(1),
		RunE: register(CapsuleCreateBuild),
	}
	createBuild.PersistentFlags().StringVarP(&buildImage, "image", "i", "", "image to build")
	capsule.AddCommand(createBuild)

	listBuilds := &cobra.Command{
		Use:  "list-builds <capsule-id>",
		Args: cobra.ExactArgs(1),
		RunE: register(CapsuleListBuilds),
	}
	capsule.AddCommand(listBuilds)

	deleteBuild := &cobra.Command{
		Use:  "delete-build <capsule-id> <build-id>",
		Args: cobra.ExactArgs(2),
		RunE: register(CapsuleDeleteBuild),
	}
	capsule.AddCommand(deleteBuild)

	deployBuild := &cobra.Command{
		Use:  "deploy <capsule-id> <build-id>",
		Args: cobra.ExactArgs(2),
		RunE: register(CapsuleDeployBuild),
	}
	capsule.AddCommand(deployBuild)

	rootCmd.AddCommand(capsule)

	pushImage := &cobra.Command{
		Use:  "push-image <image>",
		Args: cobra.ExactArgs(1),
		RunE: register(PushImage),
	}
	rootCmd.AddCommand(pushImage)
}

func CapsuleCreate(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	var is []*capsule.Update
	id, err := cs.CreateCapsule(ctx, args[0], is)
	if err != nil {
		return err
	}

	logger.Info("created capsule", zap.Stringer("capsule_id", id), zap.String("name", args[0]))

	return nil
}

func CapsuleDelete(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	capsuleID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	if err := cs.DeleteCapsule(ctx, capsuleID); err != nil {
		return err
	}

	logger.Info("capsule deleted", zap.Stringer("capsule_id", capsuleID))

	return nil
}

func CapsuleUpdate(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	capsuleID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	updates := []*capsule.Update{}
	if err := cs.UpdateCapsule(ctx, capsuleID, updates); err != nil {
		return err
	}

	logger.Info("capsule updated", zap.Stringer("capsule_id", capsuleID))

	return nil
}

func CapsuleList(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	it, total, err := cs.ListCapsules(ctx, &model.Pagination{
		Offset: uint32(listOffset),
		Limit:  uint32(listLimit),
	})
	if err != nil {
		return err
	}

	logger.Info("capsules listed", zap.Int("total", int(total)))
	for {
		c, err := it.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		logger.Info("Capsule Listed", zap.String("name", c.GetName()), zap.String("id", c.GetCapsuleId()),
			zap.Uint64("current rollout", c.GetCurrentRollout()))
	}
	return nil
}

func CapsuleCreateBuild(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	capsuleID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	buildID, err := cs.CreateBuild(ctx, capsuleID, buildImage, "", nil, nil, true)
	if err != nil {
		return err
	}

	logger.Info("build created", zap.String("build_id", buildID))

	return nil
}

func CapsuleListBuilds(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	capsuleID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	it, total, err := cs.ListBuilds(ctx, capsuleID, &model.Pagination{
		Offset: uint32(listOffset),
		Limit:  uint32(listLimit),
	})
	if err != nil {
		return err
	}

	logger.Info("Builds listed", zap.Int("total", int(total)), zap.Stringer("capsule_id", capsuleID))
	for {
		c, err := it.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		logger.Info("Build Listed", zap.String("id", c.GetBuildId()), zap.String("image", c.GetBuildId()))
	}
	return nil
}

func CapsuleDeleteBuild(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	capsuleID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	buildID := args[1]

	if err := cs.DeleteBuild(ctx, capsuleID, buildID); err != nil {
		return err
	}

	logger.Info("build deleted", zap.String("build_id", buildID))

	return nil
}

func CapsuleDeployBuild(ctx context.Context, cmd *cobra.Command, args []string, cs *capsule_service.Service, logger *zap.Logger) error {
	capsuleID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	buildID := args[1]
	cgs := []*capsule.Change{{
		Field: &capsule.Change_BuildId{BuildId: buildID},
	}}
	if err := cs.Deploy(ctx, capsuleID, cgs); err != nil {
		return err
	}

	logger.Info("build deployed", zap.String("build_id", buildID))

	return nil
}

func PushImage(ctx context.Context, cmd *cobra.Command, args []string) error {
	image := args[0]

	rigImage := fmt.Sprint("localhost:5001/rig/", image)

	if err := run(ctx, "docker", "tag", image, rigImage); err != nil {
		return err
	}

	if err := run(ctx, "docker", "push", rigImage); err != nil {
		return err
	}

	cmd.Println("Image available in Rig as", rigImage)
	return nil
}

func run(ctx context.Context, name string, args ...string) error {
	p := exec.CommandContext(ctx, name, args...)
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	return p.Run()
}
