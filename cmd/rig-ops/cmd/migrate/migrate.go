package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	rerrors "github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	envmapping "github.com/rigdev/rig/plugins/env_mapping"
	"github.com/rigdev/rig/plugins/ingress_routes"
	"github.com/rivo/tview"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/durationpb"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Migration struct {
	currentResources  *Resources
	migratedResources *Resources
	capsule           *v1alpha2.Capsule
	changes           []*capsule.Change
	warnings          map[string][]*Warning
	containerIndex    int
	operatorConfig    *v1alpha1.OperatorConfig
	plugins           []string
}

func (c *Cmd) migrate(ctx context.Context, _ *cobra.Command, _ []string) error {
	// TODO Move rig.Client into FX as well
	var rc rig.Client
	var err error
	if !skipPlatform || apply {
		rc, err = base.NewRigClient(ctx, afero.NewOsFs(), c.Prompter)
		if err != nil {
			return err
		}
	}

	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}

	migration := &Migration{
		currentResources:  NewResources(),
		migratedResources: NewResources(),
		warnings:          map[string][]*Warning{},
		changes:           []*capsule.Change{},
		operatorConfig:    cfg,
	}

	for key, value := range annotations {
		migration.changes = append(migration.changes, &capsule.Change{
			Field: &capsule.Change_SetAnnotation{
				SetAnnotation: &capsule.Change_KeyValue{
					Name:  key,
					Value: value,
				},
			},
		})
	}

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

	fmt.Print("Migrating Deployment...")
	if err := c.migrateDeployment(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Migrating Services and Ingress...")
	if err := c.migrateServicesAndIngresses(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}

	color.Green(" ✓")

	fmt.Print("Migrating Horizontal Pod Autoscaler...")
	if err := c.migrateHPA(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Migrating Environment...")
	if err := c.migrateEnvironment(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Migrating ConfigMaps and Secrets...")
	if err := c.migrateConfigFilesAndSecrets(ctx, migration); err != nil {
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	fmt.Print("Migrating Cronjobs...")
	if err := c.migrateCronJobs(ctx, migration); err != nil && err.Error() != promptAborted {
		fmt.Print("Migrating Cronjobs...")
		color.Red(" ✗")
		return err
	}
	color.Green(" ✓")

	currentTree := migration.currentResources.CreateOverview("Current Resources")
	deployRequest := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId:     migration.capsule.Name,
			ProjectId:     base.Flags.Project,
			EnvironmentId: base.Flags.Environment,
			Message:       "Migrated from kubernetes deployment",
			DryRun:        true,
			Changes:       migration.changes,
			ForceOverride: true,
		},
	}
	platformObjects := []*pipeline.Object{}
	if !skipPlatform {
		deployResp, err := rc.Capsule().Deploy(ctx, deployRequest)
		if err != nil {
			return err
		}

		platformResources := deployResp.Msg.GetResourceYaml()
		migration.capsule, platformObjects, err = c.processPlatformOutput(migration.migratedResources, platformResources)
		if err != nil {
			return fmt.Errorf("error performing dry-run on platform: %v", err)
		}
	}

	// Add "fake" CapsuleSpec status field, to inject current objects into the Pipeline.
	capsuleSpec := migration.capsule.DeepCopy()
	capsuleSpec.Status = &v1alpha2.CapsuleStatus{}
	for _, res := range migration.currentResources.All() {
		capsuleSpec.Status.OwnedResources = append(capsuleSpec.Status.OwnedResources, v1alpha2.OwnedResource{
			Ref: &corev1.TypedLocalObjectReference{
				APIGroup: ptr.To(res.GetObjectKind().GroupVersionKind().Group),
				Kind:     res.GetObjectKind().GroupVersionKind().Kind,
				Name:     res.GetName(),
			},
		})
	}

	capsuleSpecYAML, err := obj.Encode(capsuleSpec, c.Scheme)
	if err != nil {
		return err
	}

	request := &pipeline.DryRunRequest{
		Namespace:         base.Flags.Project,
		Capsule:           migration.capsule.Name,
		CapsuleSpec:       string(capsuleSpecYAML),
		AdditionalObjects: platformObjects,
		Force:             true,
	}

	if base.Flags.OperatorConfig != "" {
		cfgBytes, err := yaml.Marshal(migration.operatorConfig)
		if err != nil {
			return err
		}

		request.OperatorConfig = string(cfgBytes)
	}

	resp, err := c.OperatorClient.Pipeline.DryRun(ctx, connect.NewRequest(request))
	if err != nil {
		return fmt.Errorf("error performing dry-run on operator: %v", err)
	}

	if err := ProcessOperatorOutput(migration.migratedResources, resp.Msg.OutputObjects, c.Scheme); err != nil {
		return err
	}

	migratedTree := migration.migratedResources.CreateOverview("New Resources")

	reports, err := migration.migratedResources.Compare(migration.currentResources, c.Scheme)
	if err != nil {
		return err
	}

	if err := PromptDiffingChanges(reports,
		migration.warnings,
		currentTree,
		migratedTree, c.Prompter); err != nil && err.Error() != promptAborted {
		return err
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

	change := &capsule.Change{
		Field: &capsule.Change_AddImage_{
			AddImage: &capsule.Change_AddImage{
				Image: migration.capsule.Spec.Image,
			},
		},
	}

	deployRequest.Msg.ProjectId = base.Flags.Project
	deployRequest.Msg.Changes = append(deployRequest.Msg.Changes, change)
	deployRequest.Msg.DryRun = false
	if _, err = rc.Capsule().Deploy(ctx, deployRequest); err != nil {
		return err
	}

	fmt.Println("Capsule applied to rig platform")

	return nil
}

func (c *Cmd) setCapsulename(migration *Migration) error {
	migration.capsule.Name = migration.currentResources.Deployment.Name
	switch nameOrigin {
	case CapsuleName(""):
		if migration.currentResources.Service != nil {
			migration.capsule.Name = migration.currentResources.Service.Name
		}
	case CapsuleNameDeployment:
	case CapsuleNameService:
		if migration.currentResources.Service == nil {
			return rerrors.FailedPreconditionErrorf("No services found to inherit name from")
		}
		migration.capsule.Name = migration.currentResources.Service.Name
	case CapsuleNameInput:
		inputName, err := c.Prompter.Input("Enter the name for the capsule", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}

		migration.capsule.Name = inputName
	}

	return nil
}

func PromptDiffingChanges(
	reports *ReportSet,
	warnings map[string][]*Warning,
	currentOverview *tview.TreeView,
	migratedOverview *tview.TreeView,
	prompter common.Prompter,
) error {
	choices := []string{"Overview"}

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
		}

		report, _ := reports.GetKind(kind)
		if len(report) == 0 {
			if err := showDiffReport(nil, kind, "", warnings[kind]); err != nil {
				return err
			}
			continue
		}
		if len(report) == 1 {
			name := maps.Keys(report)[0]
			if err := showDiffReport(report[name], kind, name, warnings[kind]); err != nil {
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

			if err := showDiffReport(report[name], kind, name, warnings[kind]); err != nil {
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

		capsule := &v1alpha2.Capsule{}
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

	migration.capsule = &v1alpha2.Capsule{
		ObjectMeta: v1.ObjectMeta{
			Namespace:   migration.currentResources.Deployment.Namespace,
			Annotations: annotations,
		},
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
	changes := []*capsule.Change{}
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

		fmt.Print("Migrating Deployment...")
	}

	container := migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex]

	migration.capsule.Spec = v1alpha2.CapsuleSpec{
		Image: container.Image,
		Scale: v1alpha2.CapsuleScale{
			Vertical: &v1alpha2.VerticalScale{
				CPU:    &v1alpha2.ResourceLimits{},
				Memory: &v1alpha2.ResourceLimits{},
			},
			Horizontal: v1alpha2.HorizontalScale{
				Instances: v1alpha2.Instances{
					Min: uint32(*migration.currentResources.Deployment.Spec.Replicas),
				},
			},
		},
	}

	containerSettings := &capsule.ContainerSettings{
		Resources: &capsule.Resources{
			Requests: &capsule.ResourceList{},
			Limits:   &capsule.ResourceList{},
		},
	}
	cpu, memory := migration.capsule.Spec.Scale.Vertical.CPU, migration.capsule.Spec.Scale.Vertical.Memory
	for key, request := range container.Resources.Requests {
		switch key {
		case corev1.ResourceCPU:
			cpu.Request = &request
			containerSettings.Resources.Requests.CpuMillis = uint32(request.MilliValue())
		case corev1.ResourceMemory:
			memory.Request = &request
			containerSettings.Resources.Requests.MemoryBytes = uint64(request.Value())
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
			cpu.Limit = &limit
			containerSettings.Resources.Limits.CpuMillis = uint32(limit.MilliValue())
		case corev1.ResourceMemory:
			memory.Limit = &limit
			containerSettings.Resources.Limits.MemoryBytes = uint64(limit.Value())
		default:
			migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
				Kind:    "Deployment",
				Name:    migration.currentResources.Deployment.Name,
				Field:   fmt.Sprintf("spec.template.spec.containers.%s.resources.limit", container.Name),
				Warning: fmt.Sprintf("Limit of %s:%v is not supported by capsule", key, limit.Value()),
			})
		}
	}

	if len(container.Command) > 0 {
		migration.capsule.Spec.Command = container.Command[0]
		migration.capsule.Spec.Args = container.Command[1:]
		containerSettings.Command = migration.capsule.Spec.Command
	}

	migration.capsule.Spec.Args = append(migration.capsule.Spec.Args, container.Args...)
	containerSettings.Args = migration.capsule.Spec.Args

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

	changes = append(changes, []*capsule.Change{
		{
			Field: &capsule.Change_ImageId{
				ImageId: migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex].Image,
			},
		},
		{
			Field: &capsule.Change_Replicas{
				Replicas: uint32(*migration.currentResources.Deployment.Spec.Replicas),
			},
		},
		{
			Field: &capsule.Change_ContainerSettings{
				ContainerSettings: containerSettings,
			},
		},
	}...)

	migration.changes = append(migration.changes, changes...)

	return nil
}

func (c *Cmd) migrateHPA(ctx context.Context, migration *Migration) error {
	// Get HPA in namespace
	hpaList := &autoscalingv2.HorizontalPodAutoscalerList{}
	err := c.K8sReader.List(ctx, hpaList, client.InNamespace(base.Flags.Namespace))
	if err != nil {
		return onCCListError(err, "HorizontalPodAutoscaler", base.Flags.Namespace)
	}

	var changes []*capsule.Change
	for _, hpa := range hpaList.Items {
		found := false
		if hpa.Spec.ScaleTargetRef.Name == migration.currentResources.Deployment.Name {
			hpa := hpa
			if err := migration.currentResources.AddObject("HorizontalPodAutoscaler", hpa.Name, &hpa); err != nil {
				return err
			}

			horizontalScale := &capsule.HorizontalScale{
				MaxReplicas: uint32(hpa.Spec.MaxReplicas),
				MinReplicas: uint32(*hpa.Spec.MinReplicas),
			}

			specHorizontalScale := v1alpha2.HorizontalScale{
				Instances: v1alpha2.Instances{
					Max: ptr.To(uint32(hpa.Spec.MaxReplicas)),
					Min: uint32(*hpa.Spec.MinReplicas),
				},
			}
			if metrics := hpa.Spec.Metrics; len(metrics) > 0 {
				for _, metric := range metrics {
					if metric.Resource != nil {
						switch metric.Resource.Name {
						case corev1.ResourceCPU:
							switch metric.Resource.Target.Type {
							case autoscalingv2.UtilizationMetricType:
								specHorizontalScale.CPUTarget = &v1alpha2.CPUTarget{
									Utilization: ptr.To(uint32(*metric.Resource.Target.AverageUtilization)),
								}
								horizontalScale.CpuTarget = &capsule.CPUTarget{
									AverageUtilizationPercentage: uint32(*metric.Resource.Target.AverageUtilization),
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

						customMetric := &capsule.CustomMetric{
							Metric: &capsule.CustomMetric_Object{
								Object: &capsule.ObjectMetric{
									MetricName: metric.Object.Metric.Name,
									ObjectReference: &capsule.ObjectReference{
										Kind:       metric.Object.DescribedObject.Kind,
										Name:       metric.Object.DescribedObject.Name,
										ApiVersion: metric.Object.DescribedObject.APIVersion,
									},
								},
							},
						}

						objectMetric := v1alpha2.CustomMetric{
							ObjectMetric: &v1alpha2.ObjectMetric{
								MetricName:      metric.Object.Metric.Name,
								DescribedObject: metric.Object.DescribedObject,
							},
						}
						switch metric.Object.Target.Type {
						case autoscalingv2.AverageValueMetricType:
							objectMetric.ObjectMetric.AverageValue = metric.Object.Target.AverageValue.String()
							customMetric.GetObject().AverageValue = metric.Object.Target.AverageValue.String()
						case autoscalingv2.ValueMetricType:
							objectMetric.ObjectMetric.Value = metric.Object.Target.Value.String()
							customMetric.GetObject().Value = metric.Object.Target.Value.String()
						default:
							warning = &Warning{
								Kind:    "HorizontalPodAutoscaler",
								Name:    hpa.Name,
								Field:   "spec.metrics.object.target.type",
								Warning: fmt.Sprintf("Scaling on target %s for object metrics is not supported", metric.Object.Target.Type),
							}
						}
						if warning == nil {
							specHorizontalScale.CustomMetrics = append(migration.capsule.Spec.Scale.Horizontal.CustomMetrics, objectMetric)
							horizontalScale.CustomMetrics = append(horizontalScale.CustomMetrics, customMetric)
						} else {
							migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], warning)
						}
					}

					if metric.Pods != nil {
						var warning *Warning
						podMetric := v1alpha2.CustomMetric{
							InstanceMetric: &v1alpha2.InstanceMetric{
								MetricName: metric.Pods.Metric.Name,
							},
						}

						customMetric := &capsule.CustomMetric{
							Metric: &capsule.CustomMetric_Instance{
								Instance: &capsule.InstanceMetric{
									MetricName: metric.Pods.Metric.Name,
								},
							},
						}

						switch metric.Pods.Target.Type {
						case autoscalingv2.AverageValueMetricType:
							podMetric.InstanceMetric.AverageValue = metric.Pods.Target.AverageValue.String()
							customMetric.GetInstance().AverageValue = metric.Pods.Target.AverageValue.String()
						default:
							warning = &Warning{
								Kind:    "HorizontalPodAutoscaler",
								Name:    hpa.Name,
								Field:   "spec.metrics.pods.target.type",
								Warning: fmt.Sprintf("Scaling on target %s for pod metrics is not supported", metric.Pods.Target.Type),
							}
						}

						if warning == nil {
							specHorizontalScale.CustomMetrics = append(migration.capsule.Spec.Scale.Horizontal.CustomMetrics, podMetric)
							horizontalScale.CustomMetrics = append(horizontalScale.CustomMetrics, customMetric)
						} else {
							migration.warnings["HorizontalPodAutoscaler"] = append(migration.warnings["HorizontalPodAutoscaler"], warning)
						}
					}
				}
			}
			if specHorizontalScale.CPUTarget != nil || len(specHorizontalScale.CustomMetrics) > 0 {
				migration.capsule.Spec.Scale.Horizontal = specHorizontalScale
				changes = append(changes, &capsule.Change{
					Field: &capsule.Change_HorizontalScale{
						HorizontalScale: horizontalScale,
					},
				})
			}
			found = true
		}
		if found {
			break
		}
	}

	migration.changes = append(migration.changes, changes...)

	return nil
}

func (c *Cmd) migrateEnvironment(ctx context.Context, migration *Migration) error {
	changes := []*capsule.Change{}

	configMapMappings := map[string]map[string]string{}
	secretMappings := map[string]map[string]string{}

	env := migration.currentResources.Deployment.Spec.Template.Spec.Containers[migration.containerIndex].Env

	if len(env) == 0 {
		return nil
	}

	for _, envVar := range env {
		switch {
		case envVar.Value != "":
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_SetEnvironmentVariable{
					SetEnvironmentVariable: &capsule.Change_KeyValue{
						Name:  envVar.Name,
						Value: envVar.Value,
					},
				},
			})

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
					return onCCGetError(err, "ConfigMap",
						cfgMap.Name,
						migration.currentResources.Deployment.Namespace,
					)
				}

				if err := migration.currentResources.AddObject("ConfigMap",
					cfgMap.Name, configMap); err != nil && !rerrors.IsAlreadyExists(err) {
					return err
				}

				migration.changes = append(migration.changes, &capsule.Change{
					Field: &capsule.Change_SetEnvironmentVariable{
						SetEnvironmentVariable: &capsule.Change_KeyValue{
							Name:  envVar.Name,
							Value: configMap.Data[cfgMap.Key],
						},
					},
				})

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

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_SetAnnotation{
				SetAnnotation: &capsule.Change_KeyValue{
					Name:  envmapping.AnnotationEnvMapping,
					Value: string(annotationValueJSON),
				},
			},
		})
	}

	migration.changes = append(migration.changes, changes...)

	return nil
}

func (c *Cmd) migrateConfigFilesAndSecrets(ctx context.Context, migration *Migration) error {
	var changes []*capsule.Change
	container := migration.currentResources.Deployment.Spec.
		Template.Spec.Containers[migration.containerIndex]
	// Migrate Environment Sources
	var envReferences []v1alpha2.EnvReference
	for _, source := range container.EnvFrom {
		var environmentSource *capsule.EnvironmentSource
		switch {
		case source.ConfigMapRef != nil:
			configMap := &corev1.ConfigMap{}
			err := c.K8sReader.Get(ctx, client.ObjectKey{
				Name:      source.ConfigMapRef.Name,
				Namespace: migration.currentResources.Deployment.Namespace,
			}, configMap)
			if err != nil {
				return onCCGetError(err, "ConfigMap",
					source.ConfigMapRef.Name,
					migration.currentResources.Deployment.Namespace)
			}

			if err := migration.currentResources.AddObject("ConfigMap", source.ConfigMapRef.Name, configMap); err != nil {
				return err
			}

			if keepEnvConfigMaps {
				envReferences = append(envReferences, v1alpha2.EnvReference{
					Kind: "ConfigMap",
					Name: source.ConfigMapRef.Name,
				})

				environmentSource = &capsule.EnvironmentSource{
					Kind: capsule.EnvironmentSource_KIND_CONFIG_MAP,
					Name: source.ConfigMapRef.Name,
				}

				if err := migration.migratedResources.AddObject("ConfigMap", source.ConfigMapRef.Name, configMap); err != nil {
					return err
				}
			} else {
				for key, value := range configMap.Data {
					changes = append(changes, &capsule.Change{
						Field: &capsule.Change_SetEnvironmentVariable{
							SetEnvironmentVariable: &capsule.Change_KeyValue{
								Name:  key,
								Value: value,
							},
						},
					})
				}
			}
		case source.SecretRef != nil:
			envReferences = append(envReferences, v1alpha2.EnvReference{
				Kind: "Secret",
				Name: source.SecretRef.Name,
			})

			environmentSource = &capsule.EnvironmentSource{
				Kind: capsule.EnvironmentSource_KIND_SECRET,
				Name: source.SecretRef.Name,
			}
		}

		if environmentSource != nil {
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_SetEnvironmentSource{
					SetEnvironmentSource: environmentSource,
				},
			})
		}
	}

	if len(envReferences) > 0 {
		migration.capsule.Spec.Env = v1alpha2.Env{
			From: envReferences,
		}
	}

	// Migrate ConfigMap and Secret files
	var files []v1alpha2.File
	for _, volume := range migration.currentResources.Deployment.Spec.Template.Spec.Volumes {
		var file v1alpha2.File
		var configFile *capsule.Change_ConfigFile

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
		// If Volume is a ConfigMap
		if volume.ConfigMap != nil {
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

			configFile = &capsule.Change_ConfigFile{
				Path:     path,
				IsSecret: false,
			}

			if len(volume.ConfigMap.Items) == 1 {
				if len(configMap.BinaryData) > 0 {
					configFile.Content = configMap.BinaryData[volume.ConfigMap.Items[0].Key]
				} else if len(configMap.Data) > 0 {
					configFile.Content = []byte(configMap.Data[volume.ConfigMap.Items[0].Key])
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

			file = v1alpha2.File{
				Ref: &v1alpha2.FileContentReference{
					Kind: "ConfigMap",
					Name: volume.ConfigMap.Name,
					Key:  "content",
				},
				Path: path,
			}
			// If Volume is a Secret
		} else if volume.Secret != nil {
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

			file = v1alpha2.File{
				Ref: &v1alpha2.FileContentReference{
					Kind: "Secret",
					Name: volume.Secret.SecretName,
					Key:  "content",
				},
				Path: path,
			}

			configFile = &capsule.Change_ConfigFile{
				Path:     path,
				IsSecret: true,
			}

			if len(volume.Secret.Items) == 1 {
				if len(secret.Data) > 0 {
					configFile.Content = secret.Data[volume.Secret.Items[0].Key]
				} else if len(secret.StringData) > 0 {
					configFile.Content = []byte(secret.StringData[volume.Secret.Items[0].Key])
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
		} else {
			migration.warnings["Deployment"] = append(migration.warnings["Deployment"], &Warning{
				Kind: "Deployment",
				Name: migration.currentResources.Deployment.Name,
				Field: fmt.Sprintf("spec.template.spec.volumes.%s",
					volume.Name),
				Warning: "Volume is not a ConfigMap or Secret. Cannot migrate files",
			})
		}

		if file.Path != "" && file.Ref != nil {
			files = append(files, file)
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_SetConfigFile{
					SetConfigFile: configFile,
				},
			})
		}
	}

	migration.capsule.Spec.Files = files
	migration.changes = append(migration.changes, changes...)

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
			if err := migration.currentResources.AddObject("Service", service.GetName(), &service); err != nil &&
				rerrors.IsAlreadyExists(err) {
				if service.Name != migration.currentResources.Service.Name {
					migration.warnings["Service"] = append(migration.warnings["Service"], &Warning{
						Kind:    "Service",
						Name:    migration.currentResources.Deployment.Name,
						Warning: "More than one service is configured for the deployment",
					})
				}
				continue
			} else if err != nil {
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

	interfaces := make([]v1alpha2.CapsuleInterface, 0, len(container.Ports))
	capsuleInterfaces := make([]*capsule.Interface, 0, len(container.Ports))
	changes := []*capsule.Change{}

	ingresses := &netv1.IngressList{}
	err := c.K8sReader.List(ctx, ingresses, client.InNamespace(migration.currentResources.Deployment.GetNamespace()))
	if err != nil {
		return onCCListError(err, "Ingress", migration.currentResources.Deployment.GetNamespace())
	}

	for _, port := range container.Ports {
		i := v1alpha2.CapsuleInterface{
			Name: port.Name,
			Port: port.ContainerPort,
		}

		ci := &capsule.Interface{
			Name: port.Name,
			Port: uint32(port.ContainerPort),
		}

		routes := []v1alpha2.HostRoute{}
		capsuleRoutes := []*capsule.HostRoute{}
		for _, ingress := range ingresses.Items {
			ingress := ingress
			routePaths := []v1alpha2.HTTPPathRoute{}
			capsuleRoutePaths := []*capsule.HTTPPathRoute{}

			annotations := maps.Clone(ingress.Annotations)
			delete(annotations, "kubernetes.io/ingress.class")
			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")

			for _, path := range ingress.Spec.Rules[0].HTTP.Paths {
				if path.Backend.Service.Name != migration.currentResources.Service.Name {
					continue
				}

				if path.Backend.Service.Port.Name != port.Name && path.Backend.Service.Port.Number != port.ContainerPort {
					continue
				}

				if err := migration.currentResources.AddObject("Ingress", ingress.GetName(), &ingress); err != nil &&
					!rerrors.IsAlreadyExists(err) {
					return err
				}

				var pathType v1alpha2.PathMatchType
				var capsulePathType capsule.PathMatchType
				switch *path.PathType {
				case netv1.PathTypeImplementationSpecific:
					// TODO(ajohnsen): This is actually per path.
					annotations[ingress_routes.AnnotationImplementationSpecificPathType] = "true"
				case netv1.PathTypePrefix:
					pathType = v1alpha2.PathPrefix
					capsulePathType = capsule.PathMatchType_PATH_MATCH_TYPE_PATH_PREFIX
				case netv1.PathTypeExact:
					pathType = v1alpha2.Exact
					capsulePathType = capsule.PathMatchType_PATH_MATCH_TYPE_EXACT
				}

				routePaths = append(routePaths, v1alpha2.HTTPPathRoute{
					Path:  path.Path,
					Match: pathType,
				})

				capsuleRoutePaths = append(capsuleRoutePaths, &capsule.HTTPPathRoute{
					Path:  path.Path,
					Match: capsulePathType,
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

				routes = append(routes, v1alpha2.HostRoute{
					ID:   ingress.GetName(),
					Host: ingress.Spec.Rules[0].Host,
					RouteOptions: v1alpha2.RouteOptions{
						Annotations: annotations,
					},
					Paths: routePaths,
				})
			}

			if len(capsuleRoutePaths) > 0 {
				capsuleRoutes = append(capsuleRoutes, &capsule.HostRoute{
					Id:   ingress.GetName(),
					Host: ingress.Spec.Rules[0].Host,
					Options: &capsule.RouteOptions{
						Annotations: annotations,
					},
					Paths: capsuleRoutePaths,
				})
			}
		}

		i.Routes = routes
		ci.Routes = capsuleRoutes

		if livenessProbe != nil {
			i.Liveness, ci.Liveness, err = migrateProbe(livenessProbe, port)
			if err == nil {
				livenessProbe = nil
			}
		}

		if readinessProbe != nil {
			i.Readiness, ci.Readiness, err = migrateProbe(readinessProbe, port)
			if err == nil {
				readinessProbe = nil
			}
		}

		capsuleInterfaces = append(capsuleInterfaces, ci)
		interfaces = append(interfaces, i)
	}

	if len(interfaces) > 0 {
		migration.capsule.Spec.Interfaces = interfaces
		changes = []*capsule.Change{
			{
				Field: &capsule.Change_Network{
					Network: &capsule.Network{
						Interfaces: capsuleInterfaces,
					},
				},
			},
		}
	}

	migration.changes = append(migration.changes, changes...)

	return nil
}

func migrateProbe(probe *corev1.Probe,
	port corev1.ContainerPort,
) (*v1alpha2.InterfaceProbe, *capsule.InterfaceProbe, error) {
	TCPAndCorrectPort := probe.TCPSocket != nil &&
		(probe.TCPSocket.Port.StrVal == port.Name || probe.TCPSocket.Port.IntVal == port.ContainerPort)
	if TCPAndCorrectPort {
		if probe.TCPSocket.Port.StrVal == port.Name || probe.TCPSocket.Port.IntVal == port.ContainerPort {
			return &v1alpha2.InterfaceProbe{
					TCP: true,
				},
				&capsule.InterfaceProbe{
					Kind: &capsule.InterfaceProbe_Tcp{
						Tcp: &capsule.InterfaceProbe_TCP{},
					},
				}, nil
		}
	}

	HTTPAndCorrectPort := probe.HTTPGet != nil &&
		(probe.HTTPGet.Port.StrVal == port.Name || probe.HTTPGet.Port.IntVal == port.ContainerPort)
	if HTTPAndCorrectPort {
		return &v1alpha2.InterfaceProbe{
				Path: probe.HTTPGet.Path,
			},
			&capsule.InterfaceProbe{
				Kind: &capsule.InterfaceProbe_Http{
					Http: &capsule.InterfaceProbe_HTTP{
						Path: probe.HTTPGet.Path,
					},
				},
			}, nil
	}

	GRPCAndCorrectPort := probe.GRPC != nil && probe.GRPC.Port == port.ContainerPort
	if GRPCAndCorrectPort {
		var service string
		if probe.GRPC.Service != nil {
			service = *probe.GRPC.Service
		}

		return &v1alpha2.InterfaceProbe{
				GRPC: &v1alpha2.InterfaceGRPCProbe{
					Service: service,
				},
			}, &capsule.InterfaceProbe{
				Kind: &capsule.InterfaceProbe_Grpc{
					Grpc: &capsule.InterfaceProbe_GRPC{
						Service: service,
					},
				},
			}, nil
	}

	return nil, nil, rerrors.InvalidArgumentErrorf("Probe for port %s is not supported", port.Name)
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

	migratedCronJobs := make([]v1alpha2.CronJob, 0, len(cronJobs))
	changes := []*capsule.Change{}
	for {
		i, err := c.Prompter.TableSelect("\nSelect a job to migrate or CTRL+C to continue",
			jobTitles, headers, common.SelectEnableFilterOpt, common.SelectDontShowResultOpt)
		if err != nil {
			break
		}

		cronJob := cronJobs[i]

		migratedCronJob, addCronjob, err := c.migrateCronJob(migration, cronJob)
		if err != nil {
			fmt.Println(err)
			continue
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_AddCronJob{
				AddCronJob: addCronjob,
			},
		})
		migration.capsule.Spec.CronJobs = append(migratedCronJobs, *migratedCronJob)
		if err := migration.currentResources.AddObject("CronJob", cronJob.Name, &cronJob); err != nil {
			return err
		}

		// remove the selected job from the list
		jobTitles = append(jobTitles[:i], jobTitles[i+1:]...)
		cronJobs = append(cronJobs[:i], cronJobs[i+1:]...)
	}

	migration.changes = append(migration.changes, changes...)

	return nil
}

func (c *Cmd) migrateCronJob(
	migration *Migration,
	cronJob batchv1.CronJob,
) (*v1alpha2.CronJob, *capsule.CronJob, error) {
	migrated := &v1alpha2.CronJob{
		Name:     cronJob.Name,
		Schedule: cronJob.Spec.Schedule,
	}

	capsuleCronjob := &capsule.CronJob{
		JobName:  cronJob.Name,
		Schedule: cronJob.Spec.Schedule,
	}

	if cronJob.Spec.JobTemplate.Spec.BackoffLimit != nil {
		migrated.MaxRetries = ptr.To(uint(*cronJob.Spec.JobTemplate.Spec.BackoffLimit))
		capsuleCronjob.MaxRetries = *cronJob.Spec.JobTemplate.Spec.BackoffLimit
	}

	if cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds != nil {
		migrated.TimeoutSeconds = ptr.To(uint(*cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds))
		capsuleCronjob.Timeout = durationpb.New(time.Duration(*cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds))
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

		migrated.Command = &v1alpha2.JobCommand{
			Command: cmd,
			Args:    args,
		}

		capsuleCronjob.JobType = &capsule.CronJob_Command{
			Command: &capsule.JobCommand{
				Command: cmd,
				Args:    args,
			},
		}
	} else if keepGoing, err := c.Prompter.Confirm(`The cronjob does not fit the deployment image.
		Do you want to continue with a curl based cronjob?`, false); keepGoing && err == nil {
		fmt.Printf("Migrating cronjob %s to a curl based cronjob\n", cronJob.Name)
		fmt.Printf("This will create a new job that will run a curl command to the service\n")
		fmt.Printf("Current cmd and args are: %s %s\n",
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Command[0],
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[migration.containerIndex].Args)
		urlString, err := c.Prompter.Input("Finish the path to the service",
			common.InputDefaultOpt(fmt.Sprintf("http://%s:[PORT]/[PATH]?[PARAMS1]&[PARAM2]",
				migration.currentResources.Deployment.Name)))
		if err != nil {
			return nil, nil, err
		}

		// parse url and get port and path
		url, err := url.Parse(urlString)
		if err != nil {
			return nil, nil, err
		}

		portInt, err := strconv.ParseUint(url.Port(), 10, 16)
		if err != nil {
			return nil, nil, err
		}
		port := uint16(portInt)

		queryParams := make(map[string]string)
		for key, values := range url.Query() {
			queryParams[key] = values[0]
		}

		migrated.URL = &v1alpha2.URL{
			Port:            port,
			Path:            url.Path,
			QueryParameters: queryParams,
		}

		capsuleCronjob.JobType = &capsule.CronJob_Url{
			Url: &capsule.JobURL{
				Port:            uint64(port),
				Path:            url.Path,
				QueryParameters: queryParams,
			},
		}
	}

	return migrated, capsuleCronjob, nil
}

func onCCGetError(err error, kind, name, namespace string) error {
	return errors.Wrapf(err, "Error getting %s %s in namespace %s", kind, name, namespace)
}

func onCCListError(err error, kind, namespace string) error {
	return errors.Wrapf(err, "Error listing %s in namespace %s", kind, namespace)
}
