package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	rigAutoscalingv2 "github.com/rigdev/rig-go-api/k8s.io/api/autoscaling/v2"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	v1alpha2pkg "github.com/rigdev/rig/pkg/api/v1alpha2"
	rerrors "github.com/rigdev/rig/pkg/errors"
	envmapping "github.com/rigdev/rig/plugins/builtin/env_mapping"
	"github.com/rigdev/rig/plugins/capsulesteps/deployment"
	"github.com/rigdev/rig/plugins/capsulesteps/ingress_routes"
	"github.com/rivo/tview"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Migration struct {
	currentResources  *Resources
	migratedResources *Resources
	capsuleSpec       *platformv1.CapsuleSpec
	capsuleName       string
	warnings          map[string][]*Warning
	containerIndex    int
	operatorConfig    *v1alpha1.OperatorConfig
	plugins           []string
}

func (c *Cmd) migrate(ctx context.Context, _ *cobra.Command, _ []string) error {
	// TODO Move rig.Client into FX as well
	rc, err := base.NewRigClient(ctx, afero.NewOsFs(), c.Prompter)
	if err != nil {
		return err
	}

	if helmDir != "" && base.Flags.KubeDir != "" {
		return fmt.Errorf("--helm-dir and --kube-dir cannot both be supplied")
	}

	if helmDir != "" {
		reader, err := createHelmReader(c.Scheme, helmDir, valuesFiles)
		if err != nil {
			return err
		}
		c.K8sReader = reader
	}

	if base.Flags.KubeDir != "" {
		reader, err := createRegularDirReader(c.Scheme, base.Flags.KubeDir)
		if err != nil {
			return err
		}
		c.K8sReader = reader
	}

	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}
	migration := &Migration{
		currentResources:  NewResources(),
		migratedResources: NewResources(),
		capsuleSpec: &platformv1.CapsuleSpec{
			Annotations: map[string]string{},
			Env: &platformv1.EnvironmentVariables{
				Raw: map[string]string{},
			},
		},
		warnings:       map[string][]*Warning{},
		operatorConfig: cfg,
	}
	maps.Copy(migration.capsuleSpec.Annotations, annotations)

	if err := c.getDeployment(ctx, migration); err != nil || migration.currentResources.Deployment == nil {
		return err
	}

	if err := c.getService(ctx, migration); err != nil {
		return err
	}

	if err := c.setCapsulename(migration); err != nil {
		return err
	}

	if err = c.getPlugins(ctx, migration); err != nil {
		return nil
	}
	fmt.Println("Enabled Plugins:", strings.Join(migration.plugins, ", "))

	fmt.Print("Scanning for Deployment...")
	if err := c.migrateDeployment(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Scanning for Services and Ingress...")
	if err := c.migrateServicesAndIngresses(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Scanning for Horizontal Pod Autoscaler...")
	if err := c.migrateHPA(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Scanning for Environment...")
	if err := c.migrateEnvironment(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Scanning for ConfigMaps and Secrets...")
	if err := c.migrateConfigFilesAndSecrets(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Scanning for Cronjobs...")
	if err := c.migrateCronJobs(ctx, migration); err != nil && err.Error() != common.AbortedErrMsg {
		fmt.Print("Scanning for Cronjobs...")
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	platformCapsule := &platformv1.Capsule{
		Kind:        "Capsule",
		ApiVersion:  "platform.rig.dev/v1",
		Name:        migration.capsuleName,
		Project:     base.Flags.Project,
		Environment: base.Flags.Environment,
		Spec:        migration.capsuleSpec,
	}

	if skipDryRun {
		// Export the capsule to a file if export is set
		if export != "" {
			if err := exportCapsule(platformCapsule, export); err != nil {
				return err
			}
			return nil
		}
		fmt.Println("")
		return common.FormatPrint(platformCapsule, common.OutputTypeYAML)
	}

	currentTree := migration.currentResources.CreateOverview("Current Resources")
	deployRequest := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId:     migration.capsuleName,
			ProjectId:     base.Flags.Project,
			EnvironmentId: base.Flags.Environment,
			Message:       "Migrated from kubernetes deployment",
			Changes: []*capsule.Change{
				{
					Field: &capsule.Change_Spec{
						Spec: migration.capsuleSpec,
					},
				},
			},
			DryRun:        true,
			ForceOverride: true,
		},
	}
	if base.Flags.OperatorConfig != "" {
		bytes, err := os.ReadFile(base.Flags.OperatorConfig)
		if err != nil {
			return err
		}
		deployRequest.Msg.OperatorConfig = string(bytes)
	}

	deployResp, err := rc.Capsule().Deploy(ctx, deployRequest)
	if err != nil {
		return err
	}
	if err = c.processPlatformOutput(migration.migratedResources, deployResp.Msg.GetOutcome()); err != nil {
		return fmt.Errorf("error performing dry-run on platform: %v", err)
	}

	if deployResp.Msg.GetOutcome().GetKubernetesError() != "" {
		return errors.New("Kubernetes Error: " + deployResp.Msg.GetOutcome().GetKubernetesError())
	}

	migratedTree := migration.migratedResources.CreateOverview("New Resources")

	reports, err := migration.migratedResources.Compare(migration.currentResources, c.Scheme)
	if err != nil {
		return err
	}

	if err := PromptDiffingChanges(reports,
		migration.warnings,
		currentTree,
		migratedTree,
		platformCapsule,
		c.Prompter); err != nil && err.Error() != promptAborted {
		return err
	}

	// Export the capsule to a file if export is set
	if export != "" {
		if err := exportCapsule(platformCapsule, export); err != nil {
			return err
		}
	}

	if !apply {
		return nil
	}

	apply, err := c.Prompter.Confirm("Do you want to apply the capsule to the rig platform?", false)
	if err != nil {
		return err
	}

	if !apply {
		return nil
	}

	if _, err := rc.Capsule().Create(ctx, connect.NewRequest(&capsule.CreateRequest{
		Name:      deployRequest.Msg.CapsuleId,
		ProjectId: base.Flags.Project,
	})); rerrors.IsAlreadyExists(err) {
		// Okay, just do the deployment.
	} else if err != nil {
		return err
	}

	deployRequest.Msg.DryRun = false
	if _, err = rc.Capsule().Deploy(ctx, deployRequest); err != nil {
		return err
	}

	fmt.Println("Capsule applied to rig platform")

	return nil
}

func (c *Cmd) setCapsulename(migration *Migration) error {
	migration.capsuleName = migration.currentResources.Deployment.Name
	switch nameOrigin {
	case CapsuleName(""):
		if len(migration.currentResources.Services) == 1 {
			for name := range migration.currentResources.Services {
				migration.capsuleName = name
			}
		}
	case CapsuleNameDeployment:
	case CapsuleNameService:
		if len(migration.currentResources.Services) == 0 {
			return rerrors.FailedPreconditionErrorf("No services found to inherit name from")
		} else if len(migration.currentResources.Services) == 1 {
			for name := range migration.currentResources.Services {
				migration.capsuleName = name
			}
		} else {
			_, name, err := c.Prompter.Select(
				"Which service to inherit the name for", maps.Keys(migration.currentResources.Services),
			)
			if err != nil {
				return err
			}
			migration.capsuleName = name
		}
	case CapsuleNameInput:
		inputName, err := c.Prompter.Input("Enter the name for the capsule", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}

		migration.capsuleName = inputName
	}

	return nil
}

func PromptDiffingChanges(
	reports *ReportSet,
	warnings map[string][]*Warning,
	currentOverview *tview.TreeView,
	migratedOverview *tview.TreeView,
	platformCapsule *platformv1.Capsule,
	prompter common.Prompter,
) error {
	choices := []string{"Overview"}

	if platformCapsule != nil {
		choices = append(choices, "Platform Capsule")
	}

	for kind := range reports.reports {
		choices = append(choices, kind)
	}

	for kind, warnings := range warnings {
		choice := fmt.Sprintf("%s (%v Warnings)", kind, len(warnings))
		i := slices.Index(choices, kind)
		if i == -1 {
			choices = append(choices, choice)
		} else {
			choices[i] = choice
		}
	}

	for {
		_, kind, err := prompter.Select("Select the resource kind to view the diff for. CTRL + C to continue",
			choices,
			common.SelectDontShowResultOpt,
			common.SelectPageSizeOpt(8))
		if err != nil {
			return err
		}

		// if kind contains warnings, strip it from the kind
		if i := strings.Index(kind, " ("); i != -1 {
			kind = kind[:i]
		}

		switch kind {
		case "Overview":
			if err := showOverview(currentOverview, migratedOverview); err != nil {
				return err
			}
			continue
		case "Platform Capsule":
			if err := showCapsule(platformCapsule); err != nil {
				return err
			}
			continue
		}

		report, _ := reports.GetKind(kind)
		if len(report) == 0 {
			if err := ShowDiffReport(nil, kind, "", warnings[kind]); err != nil {
				return err
			}
			continue
		}
		if len(report) == 1 {
			name := maps.Keys(report)[0]
			if err := ShowDiffReport(report[name], kind, name, warnings[kind]); err != nil {
				return err
			}
			continue
		}
		names := []string{}
		for name := range report {
			names = append(names, name)
		}

		for {
			_, name, err := prompter.Select("Select the resource to view the diff for. CTRL + C to continue", names,
				common.SelectDontShowResultOpt,
				common.SelectPageSizeOpt(8))
			if err != nil && err.Error() == promptAborted {
				break
			} else if err != nil {
				return err
			}

			if err := ShowDiffReport(report[name], kind, name, warnings[kind]); err != nil {
				return err
			}
		}
	}
}

func (c *Cmd) promptDeploymentSelect(ctx context.Context) (*appsv1.Deployment, error) {
	deployments := &appsv1.DeploymentList{}
	err := c.K8sReader.List(ctx, deployments, client.InNamespace(base.Flags.Namespace))
	if err != nil {
		return nil, onCCListError(err, "Deployment", base.Flags.Namespace)
	}

	if len(deployments.Items) == 1 {
		return &deployments.Items[0], nil
	}

	headers := []string{"NAME", "NAMESPACE", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"}
	deploymentNames := make([][]string, 0, len(deployments.Items))
	for _, deployment := range deployments.Items {
		deploymentNames = append(deploymentNames, []string{
			deployment.GetName(),
			deployment.GetNamespace(),
			fmt.Sprintf("    %d/%d    ", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas),
			fmt.Sprintf("     %d     ", deployment.Status.UpdatedReplicas),
			fmt.Sprintf("     %d     ", deployment.Status.AvailableReplicas),
			deployment.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		})

		if deployment.GetName() == deploymentName {
			return &deployment, nil
		}
	}
	i, err := c.Prompter.TableSelect("Select the deployment to migrate",
		deploymentNames, headers, common.SelectEnableFilterOpt, common.SelectPageSizeOpt(10))
	if err != nil {
		return nil, err
	}

	return &deployments.Items[i], nil
}

func (c *Cmd) getDeployment(ctx context.Context, migration *Migration) error {
	deployment, err := c.promptDeploymentSelect(ctx)
	if err != nil {
		return err
	}

	if deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"] != "" {
		if keepGoing, err := c.Prompter.Confirm("This deployment is already owned by a capsule."+
			" Do you want to continue anyways?", false); !keepGoing || err != nil {
			return err
		}

		capsule := &v1alpha2pkg.Capsule{}
		err := c.K8sReader.Get(ctx, client.ObjectKey{
			Name:      deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"],
			Namespace: deployment.GetNamespace(),
		}, capsule)
		if err != nil {
			return onCCGetError(err, "Capsule",
				deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"],
				deployment.GetNamespace())
		}

		if err := migration.currentResources.AddObject("Capsule", capsule.GetName(), capsule); err != nil {
			return err
		}
	}

	if err := migration.currentResources.AddObject("Deployment", deployment.GetName(), deployment); err != nil {
		return err
	}

	return nil
}

func (c *Cmd) promptContainerSelectIndex(containers []corev1.Container) (int, error) {
	containerNames := make([]string, 0, len(containers))
	for _, container := range containers {
		containerNames = append(containerNames, container.Name)
	}

	i, _, err := c.Prompter.Select("\nThe deployment has more than 1 container. Select the primary container to migrate",
		containerNames,
		common.SelectDontShowResultOpt,
	)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c *Cmd) migrateDeployment(
	ctx context.Context,
	migration *Migration,
) error {
	var err error
	migration.containerIndex = 0
	if containers := migration.currentResources.Deployment.Spec.Template.Spec.Containers; len(containers) > 1 {
		migration.containerIndex, err = c.promptContainerSelectIndex(containers)
		if err != nil {
			return err
		}

		allContainerNames := make([]string, 0, len(containers))
		for _, container := range containers {
			allContainerNames = append(allContainerNames, container.Name)
		}

		migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
			Kind:  "Deployment",
			Name:  migration.currentResources.Deployment.Name,
			Field: "spec.template.spec.containers",
			Warning: fmt.Sprintf("Multiple containers: %s in a deployment are not supported by capsule. "+
				"Only container: %s will be migrated",
				strings.Join(allContainerNames, ", "), containers[migration.containerIndex].Name),
			Suggestion: "Use the rigdev.init_container plugin or the rigdev.sidecar plugin to migrate the other containers",
		})

		fmt.Print("Scanning for Deployment...")
	}

	container := migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex]

	migration.capsuleSpec.Image = container.Image
	migration.capsuleSpec.Scale = &platformv1.Scale{
		Horizontal: &platformv1.HorizontalScale{
			Min: 1,
		},
	}

	if migration.currentResources.Deployment.Spec.Replicas != nil {
		migration.capsuleSpec.Scale.Horizontal.Min = uint32(*migration.currentResources.Deployment.Spec.Replicas)
	}

	if len(container.Resources.Requests) > 0 || len(container.Resources.Limits) > 0 {
		migration.capsuleSpec.Scale.Vertical = &platformv1.VerticalScale{
			Cpu:    &platformv1.ResourceLimits{},
			Memory: &platformv1.ResourceLimits{},
		}

		cpu, memory := migration.capsuleSpec.Scale.Vertical.Cpu, migration.capsuleSpec.Scale.Vertical.Memory
		for key, request := range container.Resources.Requests {
			switch key {
			case corev1.ResourceCPU:
				cpu.Request = request.String()
			case corev1.ResourceMemory:
				memory.Request = request.String()
			default:
				migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
					Kind:    "Deployment",
					Name:    migration.currentResources.Deployment.Name,
					Field:   fmt.Sprintf("spec.template.spec.containers.%s.resources.requests", container.Name),
					Warning: fmt.Sprintf("Request of %s:%v is not supported by capsule", key, request.Value()),
				})
			}
		}

		for key, limit := range container.Resources.Limits {
			switch key {
			case corev1.ResourceCPU:
				cpu.Limit = limit.String()
			case corev1.ResourceMemory:
				memory.Limit = limit.String()
			default:
				migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
					Kind:    "Deployment",
					Name:    migration.currentResources.Deployment.Name,
					Field:   fmt.Sprintf("spec.template.spec.containers.%s.resources.limit", container.Name),
					Warning: fmt.Sprintf("Limit of %s:%v is not supported by capsule", key, limit.Value()),
				})
			}
		}
	}

	if len(container.Command) > 0 {
		migration.capsuleSpec.Command = container.Command[0]
		migration.capsuleSpec.Args = container.Command[1:]
	}

	migration.capsuleSpec.Args = append(migration.capsuleSpec.Args, container.Args...)

	// Check if the deployment has a service account, and if so add it to the current resources
	if migration.currentResources.Deployment.Spec.Template.Spec.ServiceAccountName != "" {
		serviceAccount := &corev1.ServiceAccount{}
		err := c.K8sReader.Get(ctx, client.ObjectKey{
			Name:      migration.currentResources.Deployment.Spec.Template.Spec.ServiceAccountName,
			Namespace: migration.currentResources.Deployment.Namespace,
		}, serviceAccount)
		if kerrors.IsNotFound(err) {
			migration.warnings["ServiceAccount"] = append(migration.warnings["ServiceAccount"], &Warning{
				Kind:    "ServiceAccount",
				Name:    migration.currentResources.Deployment.Spec.Template.Spec.ServiceAccountName,
				Field:   "spec.template.spec.serviceAccountName",
				Warning: "ServiceAccount not found",
			})
		} else if err != nil {
			return onCCGetError(err, "ServiceAccount",
				migration.currentResources.Deployment.Spec.Template.Spec.ServiceAccountName,
				migration.currentResources.Deployment.Namespace)
		} else {
			migration.currentResources.ServiceAccount = serviceAccount
		}
	}

	return nil
}

func (c *Cmd) migrateHPA(ctx context.Context, migration *Migration) error {
	// Get HPA in namespace
	hpaList := &autoscalingv2.HorizontalPodAutoscalerList{}
	err := c.K8sReader.List(ctx, hpaList, client.InNamespace(base.Flags.Namespace))
	if err != nil {
		return onCCListError(err, "HorizontalPodAutoscaler", base.Flags.Namespace)
	}

	for _, hpa := range hpaList.Items {
		found := false
		if hpa.Spec.ScaleTargetRef.Name == migration.currentResources.Deployment.Name {
			hpa := hpa
			if err := migration.currentResources.AddObject("HorizontalPodAutoscaler", hpa.Name, &hpa); err != nil {
				return err
			}

			if hpa.TypeMeta.APIVersion != "autoscaling/v2" {
				migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], &Warning{
					Kind:    "HorizontalPodAutoscaler",
					Name:    hpa.Name,
					Field:   "apiVersion",
					Warning: "Only autoscaling/v2 API version is supported",
				})
			}

			specHorizontalScale := &platformv1.HorizontalScale{
				Max: uint32(hpa.Spec.MaxReplicas),
				Min: uint32(*hpa.Spec.MinReplicas),
			}
			if metrics := hpa.Spec.Metrics; len(metrics) > 0 {
				for _, metric := range metrics {
					if metric.Resource != nil {
						switch metric.Resource.Name {
						case corev1.ResourceCPU:
							switch metric.Resource.Target.Type {
							case autoscalingv2.UtilizationMetricType:
								specHorizontalScale.CpuTarget = &platformv1.CPUTarget{
									Utilization: uint32(*metric.Resource.Target.AverageUtilization),
								}
							default:
								migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], &Warning{
									Kind:  "HorizontalPodAutoscaler",
									Name:  hpa.Name,
									Field: fmt.Sprintf("spec.metrics.resource.%s.target.type", metric.Resource.Name),
									Warning: fmt.Sprintf("Scaling on target type %s is not supported",
										metric.Resource.Target.Type),
								})
							}
						default:
							migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], &Warning{
								Kind:    "HorizontalPodAutoscaler",
								Name:    hpa.Name,
								Field:   fmt.Sprintf("spec.metrics.resource.%s", metric.Resource.Name),
								Warning: fmt.Sprintf("Scaling on resource %s is not supported", metric.Resource.Name),
							})
						}
					}
					if metric.Object != nil {
						var warning *Warning
						objectMetric := &platformv1.CustomMetric{
							ObjectMetric: &platformv1.ObjectMetric{
								MetricName: metric.Object.Metric.Name,
								ObjectReference: &rigAutoscalingv2.CrossVersionObjectReference{
									ApiVersion: metric.Object.DescribedObject.APIVersion,
									Kind:       metric.Object.DescribedObject.Kind,
									Name:       metric.Object.DescribedObject.Name,
								},
							},
						}
						switch metric.Object.Target.Type {
						case autoscalingv2.AverageValueMetricType:
							objectMetric.ObjectMetric.AverageValue = metric.Object.Target.AverageValue.String()
						case autoscalingv2.ValueMetricType:
							objectMetric.ObjectMetric.Value = metric.Object.Target.Value.String()
						default:
							warning = &Warning{
								Kind:    "HorizontalPodAutoscaler",
								Name:    hpa.Name,
								Field:   "spec.metrics.object.target.type",
								Warning: fmt.Sprintf("Scaling on target %s for object metrics is not supported", metric.Object.Target.Type),
							}
						}
						if warning == nil {
							specHorizontalScale.CustomMetrics = append(migration.capsuleSpec.Scale.Horizontal.CustomMetrics, objectMetric)
						} else {
							migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], warning)
						}
					}

					if metric.Pods != nil {
						var warning *Warning
						podMetric := &platformv1.CustomMetric{
							InstanceMetric: &platformv1.InstanceMetric{
								MetricName: metric.Pods.Metric.Name,
							},
						}

						switch metric.Pods.Target.Type {
						case autoscalingv2.AverageValueMetricType:
							podMetric.InstanceMetric.AverageValue = metric.Pods.Target.AverageValue.String()
						default:
							warning = &Warning{
								Kind:    "HorizontalPodAutoscaler",
								Name:    hpa.Name,
								Field:   "spec.metrics.pods.target.type",
								Warning: fmt.Sprintf("Scaling on target %s for pod metrics is not supported", metric.Pods.Target.Type),
							}
						}

						if warning == nil {
							specHorizontalScale.CustomMetrics = append(migration.capsuleSpec.Scale.Horizontal.CustomMetrics, podMetric)
						} else {
							migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], warning)
						}
					}
				}
			}
			if specHorizontalScale.CpuTarget != nil || len(specHorizontalScale.CustomMetrics) > 0 {
				migration.capsuleSpec.Scale.Horizontal = specHorizontalScale
			}
			found = true
		}
		if found {
			break
		}
	}

	return nil
}

func (c *Cmd) migrateEnvironment(ctx context.Context, migration *Migration) error {
	configMapMappings := map[string]map[string]string{}
	secretMappings := map[string]map[string]string{}

	env := migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex].Env

	if len(env) == 0 {
		return nil
	}

	migration.capsuleSpec.Env = &platformv1.EnvironmentVariables{
		Raw: map[string]string{},
	}

	for _, envVar := range env {
		switch {
		case envVar.Value != "":
			migration.capsuleSpec.Env.Raw[envVar.Name] = envVar.Value

		case envVar.ValueFrom != nil:
			from := envVar.ValueFrom
			switch {
			case from.ConfigMapKeyRef != nil:
				cfgMap := from.ConfigMapKeyRef

				configMap := &corev1.ConfigMap{}
				err := c.K8sReader.Get(ctx, client.ObjectKey{
					Name:      cfgMap.Name,
					Namespace: migration.currentResources.Deployment.Namespace,
				}, configMap)
				if err != nil {
					if kerrors.IsNotFound(err) {
						migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
							Kind: "Deployment",
							Name: migration.currentResources.Deployment.Name,
							Field: fmt.Sprintf("spec.template.spec.containers.%s.env.%s.valueFrom.configMapRef",
								migration.currentResources.Deployment.Spec.
									Template.Spec.Containers[migration.containerIndex].Name, envVar.Name),
							Warning: err.Error(),
						})
					} else {
						return onCCGetError(err, "ConfigMap",
							cfgMap.Name,
							migration.currentResources.Deployment.Namespace,
						)
					}
				}

				if err := migration.currentResources.AddObject("ConfigMap",
					cfgMap.Name, configMap); err != nil && !rerrors.IsAlreadyExists(err) {
					return err
				}

				migration.capsuleSpec.Env.Raw[envVar.Name] = configMap.Data[cfgMap.Key]

				// TODO(anders): Add under flag?
				// if !slices.Contains(migration.plugins, "rigdev.env_mapping") {
				// 	migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
				// 		Kind: "Deployment",
				// 		Name: migration.currentResources.Deployment.Name,
				// 		Field: fmt.Sprintf("spec.template.spec.containers.%s.env.%s.valueFrom.configMapKeyRef",
				// 			migration.currentResources.Deployment.Spec.
				// 				Template.Spec.Containers[migration.containerIndex].Name, envVar.Name),
				// 		Warning:    "valueFrom configMap field is not natively supported.",
				// 		Suggestion: "Enable the rigdev.env_mapping plugin to migrate envVars from configMaps",
				// 	})
				// } else {
				// 	if _, ok := configMapMappings[cfgMap.Name]; !ok {
				// 		configMapMappings[cfgMap.Name] = map[string]string{}
				// 	}

				// 	configMapMappings[cfgMap.Name][envVar.Name] = cfgMap.Key
				// }

			case from.SecretKeyRef != nil:
				secretRef := from.SecretKeyRef

				if !slices.Contains(migration.plugins, "rigdev.env_mapping") {
					migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
						Kind: "Deployment",
						Name: migration.currentResources.Deployment.Name,
						Field: fmt.Sprintf("spec.template.spec.containers.%s.env.%s.valueFrom.secretKeyRef",
							migration.currentResources.Deployment.Spec.
								Template.Spec.Containers[migration.containerIndex].Name, envVar.Name),
						Warning:    "valueFrom secret field is not natively supported.",
						Suggestion: "Enable the rigdev.env_mapping plugin to migrate envVars from secrets",
					})
				} else {
					if _, ok := secretMappings[secretRef.Name]; !ok {
						secretMappings[secretRef.Name] = map[string]string{}
					}
					secretMappings[secretRef.Name][envVar.Name] = secretRef.Key
				}
			default:
				migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
					Kind: "Deployment",
					Name: migration.currentResources.Deployment.Name,
					Field: fmt.Sprintf("spec.template.spec.containers.%s.env.%s.valueFrom",
						migration.currentResources.Deployment.Spec.
							Template.Spec.Containers[migration.containerIndex].Name, envVar.Name),
					Warning: "ValueFrom field is not supported",
				})
			}
		}
	}

	if len(configMapMappings) > 0 || len(secretMappings) > 0 {
		annotationValue := envmapping.AnnotationValue{}
		for configmap, mappings := range configMapMappings {
			annotationValue.Sources = append(annotationValue.Sources, envmapping.AnnotationSource{
				ConfigMap: configmap,
				Mappings:  mappings,
			})
		}

		for secret, mappings := range secretMappings {
			annotationValue.Sources = append(annotationValue.Sources, envmapping.AnnotationSource{
				Secret:   secret,
				Mappings: mappings,
			})
		}

		annotationValueJSON, err := json.Marshal(annotationValue)
		if err != nil {
			return err
		}

		migration.capsuleSpec.Annotations[envmapping.AnnotationEnvMapping] = string(annotationValueJSON)
	}

	return nil
}

func (c *Cmd) migrateConfigFilesAndSecrets(ctx context.Context, migration *Migration) error {
	container := migration.currentResources.Deployment.Spec.
		Template.Spec.Containers[migration.containerIndex]
	// Migrate Environment Sources
	for _, source := range container.EnvFrom {
		switch {
		case source.ConfigMapRef != nil:
			configMap := &corev1.ConfigMap{}
			err := c.K8sReader.Get(ctx, client.ObjectKey{
				Name:      source.ConfigMapRef.Name,
				Namespace: migration.currentResources.Deployment.Namespace,
			}, configMap)
			if err != nil {
				if kerrors.IsNotFound(err) {
					migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
						Kind: "Deployment",
						Name: migration.currentResources.Deployment.Name,
						Field: fmt.Sprintf("spec.template.spec.containers.%s.envFrom",
							migration.currentResources.Deployment.Spec.
								Template.Spec.Containers[migration.containerIndex].Name),
						Warning: err.Error(),
					})
				} else {
					return onCCGetError(err, "ConfigMap",
						source.ConfigMapRef.Name,
						migration.currentResources.Deployment.Namespace)
				}
			}

			if err := migration.currentResources.AddObject("ConfigMap", source.ConfigMapRef.Name, configMap); err != nil {
				return err
			}

			if keepEnvConfigMaps {
				migration.capsuleSpec.Env.Sources = append(migration.capsuleSpec.Env.Sources, &platformv1.EnvironmentSource{
					Name: source.ConfigMapRef.Name,
					Kind: "ConfigMap",
				})

				if err := migration.migratedResources.AddObject("ConfigMap", source.ConfigMapRef.Name, configMap); err != nil {
					return err
				}
			} else {
				for key, value := range configMap.Data {
					migration.capsuleSpec.Env.Raw[key] = value
				}
			}
		case source.SecretRef != nil:
			migration.capsuleSpec.Env.Sources = append(migration.capsuleSpec.Env.GetSources(), &platformv1.EnvironmentSource{
				Kind: "Secret",
				Name: source.SecretRef.Name,
			})
		}
	}

	// Migrate ConfigMap and Secret files
	var emptyPaths []string
	for _, volume := range migration.currentResources.Deployment.Spec.Template.Spec.Volumes {
		var file *platformv1.File
		var path string
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == volume.Name {
				path = volumeMount.MountPath
				break
			}
		}

		if path == "" {
			migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
				Kind: "Deployment",
				Name: migration.currentResources.Deployment.Name,
				Field: fmt.Sprintf("spec.template.spec.volumes.%s",
					volume.Name),
				Warning: "Volume is not mounted in the container. It is removed from the capsule",
				Suggestion: "If the volume is mounted in another container (init or sidecar)," +
					"use the rigdev.object_template plugin to add the volume",
			})
			continue
		}

		switch {
		// If Volume is a ConfigMap
		case volume.ConfigMap != nil:
			file = &platformv1.File{
				Path:     path,
				AsSecret: false,
			}
			configMap := &corev1.ConfigMap{}
			err := c.K8sReader.Get(ctx, client.ObjectKey{
				Name:      volume.ConfigMap.Name,
				Namespace: migration.currentResources.Deployment.Namespace,
			}, configMap)
			if err != nil {
				return onCCGetError(err, "ConfigMap", volume.ConfigMap.Name,
					migration.currentResources.Deployment.Namespace)
			}

			if err := migration.currentResources.AddObject("ConfigMap", path, configMap); err != nil {
				return err
			}

			if len(volume.ConfigMap.Items) == 1 {
				if len(configMap.BinaryData) > 0 {
					file.Bytes = configMap.BinaryData[volume.ConfigMap.Items[0].Key]
				} else if len(configMap.Data) > 0 {
					file.String_ = configMap.Data[volume.ConfigMap.Items[0].Key]
				}
			} else {
				migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
					Kind: "Deployment",
					Name: migration.currentResources.Deployment.Name,
					Field: fmt.Sprintf("spec.template.spec.volumes.%s.configMap",
						volume.Name),
					Warning: "Volume does not have exactly one item. Cannot migrate files",
				})
				continue
			}
		// If Volume is a Secret
		case volume.Secret != nil:
			file = &platformv1.File{
				Path:     path,
				AsSecret: true,
			}
			secret := &corev1.Secret{}
			err := c.K8sReader.Get(ctx, client.ObjectKey{
				Name:      volume.Secret.SecretName,
				Namespace: migration.currentResources.Deployment.Namespace,
			}, secret)
			if err != nil {
				return onCCGetError(err, "Secret", volume.Secret.SecretName, migration.currentResources.Deployment.Namespace)
			}

			if err := migration.currentResources.AddObject("Secret", path, secret); err != nil {
				return err
			}

			if len(volume.Secret.Items) == 1 {
				if len(secret.Data) > 0 {
					file.Bytes = secret.Data[volume.Secret.Items[0].Key]
				} else if len(secret.StringData) > 0 {
					file.String_ = secret.StringData[volume.Secret.Items[0].Key]
				}
			} else {
				migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
					Kind: "Deployment",
					Name: migration.currentResources.Deployment.Name,
					Field: fmt.Sprintf("spec.template.spec.volumes.%s.secret",
						volume.Name),
					Warning: "Volume does not have exactly one item. Cannot migrate files",
				})
			}
		case volume.EmptyDir != nil:
			for _, v := range container.VolumeMounts {
				if v.Name == volume.Name {
					emptyPaths = append(emptyPaths, v.MountPath)
					break
				}
			}
		default:
			migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
				Kind: "Deployment",
				Name: migration.currentResources.Deployment.Name,
				Field: fmt.Sprintf("spec.template.spec.volumes.%s",
					volume.Name),
				Warning: "Volume is not a ConfigMap or Secret. Cannot migrate files",
			})
		}

		if file != nil {
			migration.capsuleSpec.Files = append(migration.capsuleSpec.Files, file)
		}
	}

	if len(emptyPaths) > 0 {
		migration.capsuleSpec.Annotations[deployment.AnnotationEmptyDirs] = strings.Join(emptyPaths, ",")
	}

	return nil
}

func (c *Cmd) getService(ctx context.Context, migration *Migration) error {
	services := &corev1.ServiceList{}
	err := c.K8sReader.List(ctx, services, client.InNamespace(migration.currentResources.Deployment.GetNamespace()))
	if err != nil {
		return onCCListError(err, "Service", migration.currentResources.Deployment.GetNamespace())
	}

	// Find service
	for _, service := range services.Items {
		match := len(service.Spec.Selector) > 0
		for key, value := range service.Spec.Selector {
			if migration.currentResources.Deployment.Spec.Template.Labels[key] != value {
				match = false
				break
			}
		}

		if match {
			service := service
			if len(migration.currentResources.Services) > 0 {
				hasDifferentService := false
				for name := range migration.currentResources.Services {
					if name != service.Name {
						hasDifferentService = true
						break
					}
				}
				if hasDifferentService {
					migration.warnings["Service"] = append(migration.warnings["Service"], &Warning{
						Kind:    "Service",
						Name:    migration.currentResources.Deployment.Name,
						Warning: "More than one service is configured for the deployment",
					})
				}
				continue
			}
			if err := migration.currentResources.AddObject("Service", service.GetName(), &service); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cmd) migrateServicesAndIngresses(ctx context.Context,
	migration *Migration,
) error {
	container := migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex]
	livenessProbe := container.LivenessProbe
	readinessProbe := container.ReadinessProbe
	startupProbe := container.StartupProbe
	ingressEnabled := migration.operatorConfig.Pipeline.RoutesStep.Plugin != ""

	if container.StartupProbe != nil {
		migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
			Kind: "Deployment",
			Name: migration.currentResources.Deployment.Name,
			Field: fmt.Sprintf("spec.template.spec.containers.%s.startupProbe",
				container.Name),
			Warning: "StartupProbe is not supported",
		})
	}

	interfaces := make([]*platformv1.CapsuleInterface, 0, len(container.Ports))

	ingresses := &netv1.IngressList{}
	err := c.K8sReader.List(ctx, ingresses, client.InNamespace(migration.currentResources.Deployment.GetNamespace()))
	if err != nil {
		return onCCListError(err, "Ingress", migration.currentResources.Deployment.GetNamespace())
	}

	var serviceName string
	// There should only be one service
	for name := range migration.currentResources.Services {
		serviceName = name
		break
	}
	for _, port := range container.Ports {
		i := &platformv1.CapsuleInterface{
			Name: port.Name,
			Port: port.ContainerPort,
		}

		routes := []*platformv1.HostRoute{}
		for _, ingress := range ingresses.Items {
			ingress := ingress
			routePaths := []*platformv1.HTTPPathRoute{}

			annotations := maps.Clone(ingress.Annotations)
			if annotations == nil {
				annotations = map[string]string{}
			}

			delete(annotations, "kubernetes.io/ingress.class")
			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")

			for _, path := range ingress.Spec.Rules[0].HTTP.Paths {
				if path.Backend.Service.Name != serviceName {
					continue
				}

				// nolint:lll
				// TODO What if the port is not the same but the service names match above. Then the ingress would still send traffic to the service
				// if path.Backend.Service.Port.Name != port.Name && path.Backend.Service.Port.Number != port.ContainerPort {
				// 	continue
				// }

				if err := migration.currentResources.AddObject("Ingress", ingress.GetName(), &ingress); err != nil &&
					!rerrors.IsAlreadyExists(err) {
					return err
				}

				var pathType string
				switch *path.PathType {
				case netv1.PathTypeImplementationSpecific:
					// TODO(ajohnsen): This is actually per path.
					annotations[ingress_routes.AnnotationImplementationSpecificPathType] = "true"
				case netv1.PathTypePrefix:
					pathType = "PathPrefix"
				case netv1.PathTypeExact:
					pathType = "Exact"
				}

				routePaths = append(routePaths, &platformv1.HTTPPathRoute{
					Path:  path.Path,
					Match: pathType,
				})
			}

			if len(routePaths) > 0 {
				if !ingressEnabled {
					migration.warnings["Ingress"] = append(migration.warnings["Ingress"],
						&Warning{
							Kind:    "Ingress",
							Name:    ingress.GetName(),
							Field:   "",
							Warning: "Routes are not configured in the operator config, and thus the ingresses cannot be migrated.",
							Suggestion: "Enable routes in the operator pipeline. If you want to create ingresses" +
								" use the rigdev.ingress_routes plugin",
						},
					)

					continue
				}

				routes = append(routes, &platformv1.HostRoute{
					Id:          ingress.GetName(),
					Host:        ingress.Spec.Rules[0].Host,
					Annotations: annotations,
					Paths:       routePaths,
				})
			}
		}

		i.Routes = routes

		if livenessProbe != nil {
			i.Liveness, err = migrateLivenessProbe(livenessProbe, port)
			if err == nil {
				livenessProbe = nil
				if startupProbe != nil && i.Liveness != nil {
					i.Liveness.StartupDelay = uint32(startupProbe.FailureThreshold * startupProbe.PeriodSeconds)
					startupProbe = nil
				}
			}
		}

		if readinessProbe != nil {
			i.Readiness, err = migrateReadinessProbe(readinessProbe, port)
			if err == nil {
				readinessProbe = nil
			}
		}

		interfaces = append(interfaces, i)
	}

	if len(interfaces) > 0 {
		migration.capsuleSpec.Interfaces = interfaces
	}

	return nil
}

func migrateLivenessProbe(probe *corev1.Probe,
	port corev1.ContainerPort,
) (*platformv1.InterfaceLivenessProbe, error) {
	TCPAndCorrectPort := probe.TCPSocket != nil &&
		(probe.TCPSocket.Port.StrVal == port.Name || probe.TCPSocket.Port.IntVal == port.ContainerPort)
	if TCPAndCorrectPort {
		if probe.TCPSocket.Port.StrVal == port.Name || probe.TCPSocket.Port.IntVal == port.ContainerPort {
			return &platformv1.InterfaceLivenessProbe{
				Tcp: true,
			}, nil
		}
	}

	HTTPAndCorrectPort := probe.HTTPGet != nil &&
		(probe.HTTPGet.Port.StrVal == port.Name || probe.HTTPGet.Port.IntVal == port.ContainerPort)
	if HTTPAndCorrectPort {
		return &platformv1.InterfaceLivenessProbe{
			Path: probe.HTTPGet.Path,
		}, nil
	}

	GRPCAndCorrectPort := probe.GRPC != nil && probe.GRPC.Port == port.ContainerPort
	if GRPCAndCorrectPort {
		var service string
		if probe.GRPC.Service != nil {
			service = *probe.GRPC.Service
		}

		return &platformv1.InterfaceLivenessProbe{
			Grpc: &platformv1.InterfaceGRPCProbe{
				Service: service,
			},
		}, nil
	}

	return nil, rerrors.InvalidArgumentErrorf("Probe for port %s is not supported", port.Name)
}

func migrateReadinessProbe(probe *corev1.Probe,
	port corev1.ContainerPort,
) (*platformv1.InterfaceReadinessProbe, error) {
	TCPAndCorrectPort := probe.TCPSocket != nil &&
		(probe.TCPSocket.Port.StrVal == port.Name || probe.TCPSocket.Port.IntVal == port.ContainerPort)
	if TCPAndCorrectPort {
		if probe.TCPSocket.Port.StrVal == port.Name || probe.TCPSocket.Port.IntVal == port.ContainerPort {
			return &platformv1.InterfaceReadinessProbe{
				Tcp: true,
			}, nil
		}
	}

	HTTPAndCorrectPort := probe.HTTPGet != nil &&
		(probe.HTTPGet.Port.StrVal == port.Name || probe.HTTPGet.Port.IntVal == port.ContainerPort)
	if HTTPAndCorrectPort {
		return &platformv1.InterfaceReadinessProbe{
			Path: probe.HTTPGet.Path,
		}, nil
	}

	GRPCAndCorrectPort := probe.GRPC != nil && probe.GRPC.Port == port.ContainerPort
	if GRPCAndCorrectPort {
		var service string
		if probe.GRPC.Service != nil {
			service = *probe.GRPC.Service
		}

		return &platformv1.InterfaceReadinessProbe{
			Grpc: &platformv1.InterfaceGRPCProbe{
				Service: service,
			},
		}, nil
	}

	return nil, rerrors.InvalidArgumentErrorf("Probe for port %s is not supported", port.Name)
}

func (c *Cmd) migrateCronJobs(ctx context.Context, migration *Migration) error {
	cronJobList := &batchv1.CronJobList{}
	err := c.K8sReader.List(ctx, cronJobList,
		client.InNamespace(migration.currentResources.Deployment.GetNamespace()),
	)
	if err != nil {
		return onCCListError(err, "CronJob", migration.currentResources.Deployment.GetNamespace())
	}

	cronJobs := cronJobList.Items
	headers := []string{"NAME", "SCHEDULE", "IMAGE", "LAST SCHEDULE", "AGE"}

	jobTitles := [][]string{}
	for _, cronJob := range cronJobList.Items {
		lastScheduled := "Never"
		if cronJob.Status.LastScheduleTime != nil {
			lastScheduled = cronJob.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
		}

		jobTitles = append(jobTitles, []string{
			cronJob.GetName(),
			cronJob.Spec.Schedule,
			strings.Split(cronJob.Spec.JobTemplate.Spec.
				Template.Spec.Containers[migration.containerIndex].Image, "@")[0],
			lastScheduled,
			cronJob.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		})
	}

	migratedCronJobs := make([]*platformv1.CronJob, 0, len(cronJobs))
	for {
		i, err := c.Prompter.TableSelect("\nSelect a job to migrate or CTRL+C to continue",
			jobTitles, headers, common.SelectEnableFilterOpt, common.SelectDontShowResultOpt)
		if err != nil {
			break
		}

		cronJob := cronJobs[i]

		migratedCronJob, err := c.migrateCronJob(migration, cronJob)
		if err != nil {
			fmt.Println(err)
			continue
		}

		migration.capsuleSpec.CronJobs = append(migratedCronJobs, migratedCronJob)
		if err := migration.currentResources.AddObject("CronJob", cronJob.Name, &cronJob); err != nil {
			return err
		}

		// remove the selected job from the list
		jobTitles = append(jobTitles[:i], jobTitles[i+1:]...)
		cronJobs = append(cronJobs[:i], cronJobs[i+1:]...)
	}

	return nil
}

func (c *Cmd) migrateCronJob(
	migration *Migration,
	cronJob batchv1.CronJob,
) (*platformv1.CronJob, error) {
	migrated := &platformv1.CronJob{
		Name:     cronJob.Name,
		Schedule: cronJob.Spec.Schedule,
	}

	if cronJob.Spec.JobTemplate.Spec.BackoffLimit != nil {
		migrated.MaxRetries = uint64(*cronJob.Spec.JobTemplate.Spec.BackoffLimit)
	}

	if cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds != nil {
		migrated.TimeoutSeconds = uint64(*cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)
	}

	if len(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers) > 1 {
		migration.warnings["CronJob"] = append(migration.warnings["Cronjob"], &Warning{
			Kind:    "CronJob",
			Name:    cronJob.Name,
			Field:   "spec.template.spec.containers",
			Warning: "CronJob has more than one container. Only the first container will be migrated",
		})
	}

	if cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Image ==
		migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex].Image {
		cmd := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Command[0]
		args := append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Command[1:],
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Args...)

		migrated.Command = &platformv1.JobCommand{
			Command: cmd,
			Args:    args,
		}
	} else if keepGoing, err := c.Prompter.Confirm(`The cronjob does not fit the deployment image.
		Do you want to continue with a curl based cronjob?`, false); keepGoing && err == nil {
		fmt.Printf("Scanning for cronjob %s to a curl based cronjob\n", cronJob.Name)
		fmt.Printf("This will create a new job that will run a curl command to the service\n")
		fmt.Printf("Current cmd and args are: %s %s\n",
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Command[0],
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Args)
		urlString, err := c.Prompter.Input("Finish the path to the service",
			common.InputDefaultOpt(fmt.Sprintf("http://%s:[PORT]/[PATH]?[PARAMS1]&[PARAM2]",
				migration.currentResources.Deployment.Name)))
		if err != nil {
			return nil, err
		}

		// parse url and get port and path
		url, err := url.Parse(urlString)
		if err != nil {
			return nil, err
		}

		portInt, err := strconv.ParseUint(url.Port(), 10, 32)
		if err != nil {
			return nil, err
		}
		port := uint32(portInt)

		queryParams := make(map[string]string)
		for key, values := range url.Query() {
			queryParams[key] = values[0]
		}

		migrated.Url = &platformv1.URL{
			Port:            port,
			Path:            url.Path,
			QueryParameters: queryParams,
		}
	}

	return migrated, nil
}

func onCCGetError(err error, kind, name, namespace string) error {
	return errors.Wrapf(err, "Error getting %s %s in namespace %s", kind, name, namespace)
}

func onCCListError(err error, kind, namespace string) error {
	return errors.Wrapf(err, "Error listing %s in namespace %s", kind, namespace)
}
