package migrate

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/api/v1/build"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/durationpb"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func migrate(ctx context.Context,
	_ *cobra.Command,
	_ []string,
	cc client.Client,
	cr client.Reader,
	oc *base.OperatorClient,
) error {
	var rc rig.Client
	var err error
	if !skipPlatform {
		rc, err = base.NewRigClient(ctx)
		if err != nil {
			return err
		}
	}

	currentResources := &CurrentResources{
		ConfigMaps: map[string]*corev1.ConfigMap{},
		Secrets:    map[string]*corev1.Secret{},
		Services:   map[string]*corev1.Service{},
		Ingresses:  map[string]*netv1.Ingress{},
		CronJobs:   map[string]*batchv1.CronJob{},
	}

	err = getDeployment(ctx, cr, currentResources)
	if err != nil || currentResources.Deployment == nil {
		return err
	}

	warnings := map[string][]*Warning{}
	changes := []*capsule.Change{}

	fmt.Print("Migrating Deployment...")
	capsuleSpec, deploymentChanges, err := migrateDeployment(ctx, currentResources, cc, warnings)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, deploymentChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Horizontal Pod Autoscaler...")
	hpaChanges, err := migrateHPA(ctx, cc, currentResources, capsuleSpec, warnings)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, hpaChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Environment...")
	environmentChanges, err := migrateEnvironment(ctx, cc, currentResources, warnings)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, environmentChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating ConfigMaps and Secrets...")
	configChanges, err := migrateConfigFilesAndSecrets(ctx, cr, currentResources, capsuleSpec, warnings)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, configChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Services and Ingress...")
	serviceChanges, err := migrateServicesAndIngresses(ctx, cr, currentResources, capsuleSpec, warnings)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, serviceChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Cronjobs...")
	cronJobChanges, err := migrateCronJobs(ctx, cc, currentResources, capsuleSpec, warnings)
	if err != nil && err.Error() == promptAborted {
		fmt.Println("Migrating Cronjobs...")
	} else if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, cronJobChanges...)
	color.Green(" ✓")

	reports := map[string]map[string]*dyff.Report{}
	currentTree := currentResources.createOverview()

	var platformResources map[string]string
	deployRequest := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId:     capsuleSpec.Name,
			ProjectId:     currentResources.Deployment.Namespace,
			EnvironmentId: base.Flags.Environment,
			Message:       "Migrated from kubernetes deployment",
			DryRun:        true,
			Changes:       changes,
		},
	}
	if !skipPlatform {

		deployResp, err := rc.Capsule().Deploy(ctx, deployRequest)
		if err != nil {
			fmt.Println("Error deploying capsule", err)
			return err
		}

		platformResources = deployResp.Msg.GetResourceYaml()
		capsuleSpec, err = processPlatformOutput(reports, currentResources, platformResources, cc.Scheme())
		if err != nil {
			return err
		}
	}

	capsuleSpecYAML, err := obj.Encode(capsuleSpec, cc.Scheme())
	if err != nil {
		return err
	}

	resp, err := oc.Pipeline.DryRun(ctx, connect.NewRequest(&pipeline.DryRunRequest{
		Namespace:   currentResources.Deployment.Namespace,
		Capsule:     capsuleSpec.Name,
		CapsuleSpec: string(capsuleSpecYAML),
		Force:       true,
	}))
	if err != nil {
		return err
	}

	err = processOperatorOutput(reports, currentResources, resp.Msg.OutputObjects, cc.Scheme())
	if err != nil {
		return err
	}

	err = processRemainingResources(reports, currentResources, cc.Scheme())
	if err != nil {
		return err
	}

	err = promptDiffingChanges(reports, warnings, currentTree)
	if err != nil && err.Error() != promptAborted {
		return err
	}

	if apply {
		apply, err := common.PromptConfirm("Do you want to apply the capsule to the rig platform?", false)
		if err != nil {
			return err
		}

		if apply {
			_, err := rc.Capsule().Create(ctx, connect.NewRequest(&capsule.CreateRequest{
				Name:      deployRequest.Msg.CapsuleId,
				ProjectId: base.Flags.Project,
			}))
			if err != nil {
				return err
			}

			buildResp, err := rc.Build().Create(ctx, connect.NewRequest(&build.CreateRequest{
				ProjectId: base.Flags.Project,
				CapsuleId: deployRequest.Msg.CapsuleId,
				Image:     capsuleSpec.Spec.Image,
			}))
			if err != nil {
				return err
			}

			change := &capsule.Change{
				Field: &capsule.Change_BuildId{
					BuildId: buildResp.Msg.BuildId,
				},
			}

			deployRequest.Msg.ProjectId = base.Flags.Project
			deployRequest.Msg.Changes = append(deployRequest.Msg.Changes, change)
			deployRequest.Msg.DryRun = false
			_, err = rc.Capsule().Deploy(ctx, deployRequest)
			if err != nil {
				return err
			}

			fmt.Println("Capsule applied to rig platform")
		}
	}

	return nil
}

func promptDiffingChanges(reports map[string]map[string]*dyff.Report,
	warnings map[string][]*Warning,
	currentOverview *tview.TreeView) error {
	choices := []string{}
	// choices = append(choices, "Overview")

	for kind := range reports {
		choices = append(choices, kind)
	}

	for {
		_, kind, err := common.PromptSelect("Select the resource kind to view the diff for. CTRL + C to continue", choices)
		if err != nil {
			return err
		}

		switch kind {
		case "Overview":
			if err := showOverview(currentOverview, reports, warnings); err != nil {
				return err
			}
			continue
		default:
			report := reports[kind]
			if len(report) == 1 {
				name := maps.Keys(report)[0]
				if err := showDiffReport(report[name], kind, name, warnings[kind]); err != nil {
					return err
				}
			} else {
				names := []string{}
				for name := range report {
					names = append(names, name)
				}

				for {
					_, name, err := common.PromptSelect("Select the resource to view the diff for. CTRL + C to continue", names)
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
	}
}

func getDeployment(ctx context.Context, cc client.Reader, currentResources *CurrentResources) error {
	deployments := &appsv1.DeploymentList{}
	err := cc.List(ctx, deployments, client.InNamespace(base.Flags.Namespace))
	if err != nil {
		return err
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
	}
	i, err := common.PromptTableSelect("Select the deployment to migrate",
		deploymentNames, headers, common.SelectEnableFilterOpt)
	if err != nil {
		return err
	}

	deployment := &deployments.Items[i]

	if deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"] != "" {
		if keepGoing, err := common.PromptConfirm("This deployment is already owned by a capsule."+
			" Do you want to continue anyways?", false); !keepGoing || err != nil {
			return err
		}

		capsule := &v1alpha2.Capsule{}
		err := cc.Get(ctx, types.NamespacedName{
			Name:      deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"],
			Namespace: deployment.GetNamespace(),
		}, capsule)
		if err != nil {
			return err
		}

		currentResources.Capsule = capsule
	}

	currentResources.Deployment = deployment

	return nil
}

func migrateDeployment(ctx context.Context,
	currentResources *CurrentResources,
	cc client.Client,
	warnings map[string][]*Warning,
) (*v1alpha2.Capsule, []*capsule.Change, error) {
	changes := []*capsule.Change{}

	if len(currentResources.Deployment.Spec.Template.Spec.Containers) > 1 {
		warnings["Deployment"] = append(warnings["Deployment"], &Warning{
			Kind:    "Deployment",
			Name:    currentResources.Deployment.Name,
			Warning: "Multiple containers in a pod are not supported by capsule. The first will be migrated",
		})
	}

	var capsuleName string
	if name == "" {
		capsuleName = currentResources.Deployment.Name
	}

	container := currentResources.Deployment.Spec.Template.Spec.Containers[0]
	capsuleSpec := &v1alpha2.Capsule{
		ObjectMeta: v1.ObjectMeta{
			Name:      capsuleName,
			Namespace: currentResources.Deployment.Namespace,
		},
		Spec: v1alpha2.CapsuleSpec{
			Image: container.Image,
			Scale: v1alpha2.CapsuleScale{
				Vertical: &v1alpha2.VerticalScale{
					CPU:    &v1alpha2.ResourceLimits{},
					Memory: &v1alpha2.ResourceLimits{},
				},
				Horizontal: v1alpha2.HorizontalScale{
					Instances: v1alpha2.Instances{
						Min: uint32(*currentResources.Deployment.Spec.Replicas),
					},
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
	cpu, memory := capsuleSpec.Spec.Scale.Vertical.CPU, capsuleSpec.Spec.Scale.Vertical.Memory
	for key, request := range container.Resources.Requests {
		switch key {
		case corev1.ResourceCPU:
			cpu.Request = &request
			containerSettings.Resources.Requests.CpuMillis = uint32(request.MilliValue())
		case corev1.ResourceMemory:
			memory.Request = &request
			containerSettings.Resources.Requests.MemoryBytes = uint64(request.Value())
		default:
			warnings["Deployment"] = append(warnings["Deployment"], &Warning{
				Kind:    "Deployment",
				Name:    currentResources.Deployment.Name,
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
			warnings["Deployment"] = append(warnings["Deployment"], &Warning{
				Kind:    "Deployment",
				Name:    currentResources.Deployment.Name,
				Warning: fmt.Sprintf("Limit of %s:%v is not supported by capsule", key, limit.Value()),
			})
		}
	}

	if len(container.Command) > 0 {
		capsuleSpec.Spec.Command = container.Command[0]
		capsuleSpec.Spec.Args = container.Command[1:]
		containerSettings.Command = capsuleSpec.Spec.Command
	}

	capsuleSpec.Spec.Args = append(capsuleSpec.Spec.Args, container.Args...)
	containerSettings.Args = capsuleSpec.Spec.Args

	// Check if the deployment has a service account, and if so add it to the current resources
	if currentResources.Deployment.Spec.Template.Spec.ServiceAccountName != "" {
		serviceAccount := &corev1.ServiceAccount{}
		err := cc.Get(ctx, types.NamespacedName{
			Name:      currentResources.Deployment.Spec.Template.Spec.ServiceAccountName,
			Namespace: currentResources.Deployment.Namespace,
		}, serviceAccount)
		if kerrors.IsNotFound(err) {
			warnings["ServiceAccount"] = append(warnings["ServiceAccount"], &Warning{
				Kind:    "ServiceAccount",
				Name:    currentResources.Deployment.Spec.Template.Spec.ServiceAccountName,
				Warning: "ServiceAccount not found",
			})
		} else if err != nil {
			return nil, nil, err
		} else {
			currentResources.ServiceAccount = serviceAccount
		}
	}

	changes = append(changes, []*capsule.Change{
		{
			Field: &capsule.Change_BuildId{
				BuildId: currentResources.Deployment.Spec.Template.Spec.Containers[0].Image,
			},
		},
		{
			Field: &capsule.Change_Replicas{
				Replicas: uint32(*currentResources.Deployment.Spec.Replicas),
			},
		},
		{
			Field: &capsule.Change_ContainerSettings{
				ContainerSettings: containerSettings,
			},
		},
	}...)

	return capsuleSpec, changes, nil
}

func migrateHPA(ctx context.Context,
	cc client.Client,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.Capsule,
	warnings map[string][]*Warning,
) ([]*capsule.Change, error) {
	// Get HPA in namespace
	hpaList := &autoscalingv2.HorizontalPodAutoscalerList{}
	err := cc.List(ctx, hpaList, client.InNamespace(base.Flags.Namespace))
	if err != nil {
		return nil, err
	}

	var changes []*capsule.Change
	for _, hpa := range hpaList.Items {
		found := false
		if hpa.Spec.ScaleTargetRef.Name == currentResources.Deployment.Name {
			hpa := hpa
			currentResources.HPA = &hpa

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
								warnings["HorizontalPodAutoscaler"] = append(warnings["HorizontalPodAutoscaler"], &Warning{
									Kind: "HorizontalPodAutoscaler",
									Name: hpa.Name,
									Warning: fmt.Sprintf("Scaling on target %s for %s is not supported",
										metric.Resource.Target.Type,
										metric.Resource.Name),
								})
							}
						default:
							warnings["HorizontalPodAutoscaler"] = append(warnings["HorizontalPodAutoscaler"], &Warning{
								Kind:    "HorizontalPodAutoscaler",
								Name:    hpa.Name,
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
								Warning: fmt.Sprintf("Scaling on target %s for object metrics is not supported", metric.Object.Target.Type),
							}
						}
						if warning == nil {
							specHorizontalScale.CustomMetrics = append(capsuleSpec.Spec.Scale.Horizontal.CustomMetrics, objectMetric)
							horizontalScale.CustomMetrics = append(horizontalScale.CustomMetrics, customMetric)
						} else {
							warnings["HorizontalPodAutoscaler"] = append(warnings["HorizontalPodAutoscaler"], warning)
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
								Warning: fmt.Sprintf("Scaling on target %s for pod metrics is not supported", metric.Pods.Target.Type),
							}
						}

						if warning == nil {
							specHorizontalScale.CustomMetrics = append(capsuleSpec.Spec.Scale.Horizontal.CustomMetrics, podMetric)
							horizontalScale.CustomMetrics = append(horizontalScale.CustomMetrics, customMetric)
						} else {
							warnings["HorizontalPodAutoscaler"] = append(warnings["HorizontalPodAutoscaler"], warning)
						}
					}
				}
			}
			if specHorizontalScale.CPUTarget != nil || len(specHorizontalScale.CustomMetrics) > 0 {
				capsuleSpec.Spec.Scale.Horizontal = specHorizontalScale
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

	return changes, nil
}

func migrateEnvironment(ctx context.Context,
	cc client.Client,
	currentResources *CurrentResources,
	warnings map[string][]*Warning,
) ([]*capsule.Change, error) {
	changes := []*capsule.Change{}

	configMap := &corev1.ConfigMap{}
	err := cc.Get(ctx, types.NamespacedName{
		Namespace: currentResources.Deployment.Namespace,
		Name:      currentResources.Deployment.Name,
	}, configMap)
	if err == nil {
		return nil, errors.New("ConfigMap already exists with the same name as the deployment" +
			"Cannot migrate environment variables")
	} else if !strings.HasSuffix(err.Error(), "not found") {
		return nil, err
	}

	if env := currentResources.Deployment.Spec.Template.Spec.Containers[0].Env; len(env) > 0 {
		for _, envVar := range env {
			if envVar.Value != "" {
				changes = append(changes, &capsule.Change{
					Field: &capsule.Change_SetEnvironmentVariable{
						SetEnvironmentVariable: &capsule.Change_KeyValue{
							Name:  envVar.Name,
							Value: envVar.Value,
						},
					},
				})
			} else if envVar.ValueFrom != nil {
				warnings["Deployment"] = append(warnings["Deployment"], &Warning{
					Kind:    "Deployment",
					Name:    currentResources.Deployment.Name,
					Warning: fmt.Sprintf("Environment variable %s has a valueFrom field. This is not supported.", envVar.Name),
				})
			}
		}
	}

	return changes, nil
}

func migrateConfigFilesAndSecrets(ctx context.Context,
	cc client.Reader,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.Capsule,
	warnings map[string][]*Warning,
) ([]*capsule.Change, error) {
	var changes []*capsule.Change
	container := currentResources.Deployment.Spec.Template.Spec.Containers[0]
	// Migrate Environment Sources
	var envReferences []v1alpha2.EnvReference
	for _, source := range container.EnvFrom {
		var environmentSource *capsule.EnvironmentSource
		if source.ConfigMapRef != nil {
			envReferences = append(envReferences, v1alpha2.EnvReference{
				Kind: "ConfigMap",
				Name: source.ConfigMapRef.Name,
			})

			environmentSource = &capsule.EnvironmentSource{
				Kind: capsule.EnvironmentSource_KIND_CONFIG_MAP,
				Name: source.ConfigMapRef.Name,
			}

			configMap := &corev1.ConfigMap{}
			err := cc.Get(ctx, types.NamespacedName{
				Name:      source.ConfigMapRef.Name,
				Namespace: currentResources.Deployment.Namespace,
			}, configMap)
			if err != nil {
				return nil, err
			}

			name := fmt.Sprintf("env-source--%s", source.ConfigMapRef.Name)
			currentResources.ConfigMaps[name] = configMap
		} else if source.SecretRef != nil {
			envReferences = append(envReferences, v1alpha2.EnvReference{
				Kind: "Secret",
				Name: source.SecretRef.Name,
			})

			environmentSource = &capsule.EnvironmentSource{
				Kind: capsule.EnvironmentSource_KIND_SECRET,
				Name: source.SecretRef.Name,
			}

			secret := &corev1.Secret{}
			err := cc.Get(ctx, types.NamespacedName{
				Name:      source.SecretRef.Name,
				Namespace: currentResources.Deployment.Namespace,
			}, secret)
			if err != nil {
				return nil, err
			}

			name := fmt.Sprintf("env-source--%s", source.SecretRef.Name)
			currentResources.Secrets[name] = secret
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
		capsuleSpec.Spec.Env = v1alpha2.Env{
			From: envReferences,
		}
	}

	// Migrate ConfigMap and Secret files
	var files []v1alpha2.File
	for _, volume := range currentResources.Deployment.Spec.Template.Spec.Volumes {
		var file v1alpha2.File
		var configFile *capsule.Change_ConfigFile

		var path string
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == volume.Name {
				path = volumeMount.MountPath
				break
			}
		}
		// If Volume is a ConfigMap
		if volume.ConfigMap != nil {
			configMap := &corev1.ConfigMap{}
			err := cc.Get(ctx, types.NamespacedName{
				Name:      volume.ConfigMap.Name,
				Namespace: currentResources.Deployment.Namespace,
			}, configMap)
			if err != nil {
				return nil, err
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
				warnings["Deployment"] = append(warnings["Deployment"], &Warning{
					Kind: "Deployment",
					Name: currentResources.Deployment.Name,
					Warning: fmt.Sprintf("Volume %s of config map %s does not have exactly one item. Cannot migrate files",
						volume.Name, volume.ConfigMap.Name),
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

			currentResources.ConfigMaps[file.Path] = configMap

			// If Volume is a Secret
		} else if volume.Secret != nil {
			secret := &corev1.Secret{}
			err := cc.Get(ctx, types.NamespacedName{
				Name:      volume.Secret.SecretName,
				Namespace: currentResources.Deployment.Namespace,
			}, secret)
			if err != nil {
				return nil, err
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
				warnings["Deployment"] = append(warnings["Deployment"], &Warning{
					Kind: "Deployment",
					Name: currentResources.Deployment.Name,
					Warning: fmt.Sprintf("Volume %s of secret %s does not have exactly one item. Cannot migrate files",
						volume.Name, volume.Secret.SecretName),
				})
			}
			currentResources.Secrets[file.Path] = secret
		} else {
			warnings["Deployment"] = append(warnings["Deployment"], &Warning{
				Kind: "Deployment",
				Name: currentResources.Deployment.Name,
				Warning: fmt.Sprintf("Volume %s is not a ConfigMap or Secret. Cannot migrate files",
					volume.Name),
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

	capsuleSpec.Spec.Files = files
	return changes, nil
}

func migrateServicesAndIngresses(ctx context.Context,
	cc client.Reader,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.Capsule,
	warnings map[string][]*Warning,
) ([]*capsule.Change, error) {
	container := currentResources.Deployment.Spec.Template.Spec.Containers[0]
	livenessProbe := container.LivenessProbe
	readinessProbe := container.ReadinessProbe

	services := &corev1.ServiceList{}
	err := cc.List(ctx, services, client.InNamespace(currentResources.Deployment.GetNamespace()))
	if err != nil {
		return nil, err
	}

	ingresses := &netv1.IngressList{}
	err = cc.List(ctx, ingresses, client.InNamespace(currentResources.Deployment.GetNamespace()))
	if err != nil {
		return nil, err
	}

	interfaces := make([]v1alpha2.CapsuleInterface, 0, len(container.Ports))
	capsuleInterfaces := make([]*capsule.Interface, 0, len(container.Ports))

	for _, port := range container.Ports {
		for _, service := range services.Items {
			match := len(service.Spec.Selector) > 0
			for key, value := range service.Spec.Selector {
				if currentResources.Deployment.Spec.Template.Labels[key] != value {
					match = false
					break
				}
			}

			if match {
				service := service
				currentResources.Services[service.GetName()] = &service
			}
		}

		if len(currentResources.Services) > 1 {
			warnings["Service"] = append(warnings["Service"], &Warning{
				Kind: "Service",
				Name: currentResources.Deployment.Name,
				Warning: fmt.Sprintf("Deployment has more than one service: %s",
					strings.Join(maps.Keys(currentResources.Services), ", ")),
			})
		}

		i := v1alpha2.CapsuleInterface{
			Name: port.Name,
			Port: port.ContainerPort,
		}

		ci := &capsule.Interface{
			Name: port.Name,
			Port: uint32(port.ContainerPort),
		}

		found := false
		for _, ingress := range ingresses.Items {
			var paths []string
			for _, path := range ingress.Spec.Rules[0].HTTP.Paths {
				if path.Backend.Service.Port.Name != port.Name {
					continue
				}

				paths = append(paths, path.Path)
			}

			if len(paths) > 0 {
				currentResources.Ingresses[ingress.GetName()] = &ingress

				if found {
					warnings["Ingress"] = append(warnings["Ingress"], &Warning{
						Kind: "Ingress",
						Name: ingress.GetName(),
						Warning: fmt.Sprintf("Previous Ingress host: %s already configured for port %s. Ingress %s is ignored.",
							i.Public.Ingress.Host, port.Name, ingress.GetName()),
					})
					continue
				}

				i.Public = &v1alpha2.CapsulePublicInterface{
					Ingress: &v1alpha2.CapsuleInterfaceIngress{
						Host:  ingress.Spec.Rules[0].Host,
						Paths: paths,
					},
				}

				ci.Public = &capsule.PublicInterface{
					Enabled: true,
					Method: &capsule.RoutingMethod{
						Kind: &capsule.RoutingMethod_Ingress_{
							Ingress: &capsule.RoutingMethod_Ingress{
								Host:  i.Public.Ingress.Host,
								Paths: paths,
							},
						},
					},
				}

				found = true
			}
		}

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

	changes := []*capsule.Change{}
	if len(interfaces) > 0 {
		capsuleSpec.Spec.Interfaces = interfaces
		changes = []*capsule.Change{
			{
				Field: &capsule.Change_Network{
					Network: &capsule.Network{
						Interfaces: capsuleInterfaces,
					},
				},
			},
		}

		if len(currentResources.Services) > 0 && name == "" {
			capsuleSpec.Name = maps.Keys(currentResources.Services)[0]
		}
	}

	return changes, nil
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

	return nil, nil, errors.InvalidArgumentErrorf("Probe for port %s is not supported", port.Name)
}

func migrateCronJobs(ctx context.Context,
	cc client.Client,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.Capsule,
	warnings map[string][]*Warning,
) ([]*capsule.Change, error) {
	cronJobList := &batchv1.CronJobList{}
	err := cc.List(ctx, cronJobList, client.InNamespace(currentResources.Deployment.GetNamespace()))
	if err != nil {
		return nil, err
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
			strings.Split(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image, "@")[0],
			lastScheduled,
			cronJob.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		})
	}

	migratedCronJobs := make([]v1alpha2.CronJob, 0, len(cronJobs))
	changes := []*capsule.Change{}
	for {
		i, err := common.PromptTableSelect("\nSelect a job to migrate or CTRL+C to continue",
			jobTitles, headers, common.SelectEnableFilterOpt, common.SelectDontShowResultOpt)
		if err != nil {
			break
		}

		cronJob := cronJobs[i]

		migratedCronJob, addCronjob, err := migrateCronJob(currentResources.Deployment, cronJob, warnings)
		if err != nil {
			fmt.Println(err)
			continue
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_AddCronJob{
				AddCronJob: addCronjob,
			},
		})
		capsuleSpec.Spec.CronJobs = append(migratedCronJobs, *migratedCronJob)
		currentResources.CronJobs[cronJob.Name] = &cronJob
		// remove the selected job from the list
		jobTitles = append(jobTitles[:i], jobTitles[i+1:]...)
		cronJobs = append(cronJobs[:i], cronJobs[i+1:]...)
	}

	return changes, nil
}

func migrateCronJob(deployment *appsv1.Deployment,
	cronJob batchv1.CronJob,
	warnings map[string][]*Warning,
) (*v1alpha2.CronJob, *capsule.CronJob, error) {
	migrated := &v1alpha2.CronJob{
		Name:           cronJob.Name,
		Schedule:       cronJob.Spec.Schedule,
		MaxRetries:     ptr.To(uint(*cronJob.Spec.JobTemplate.Spec.BackoffLimit)),
		TimeoutSeconds: ptr.To(uint(*cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)),
	}

	capsuleCronjob := &capsule.CronJob{
		JobName:    cronJob.Name,
		Schedule:   cronJob.Spec.Schedule,
		MaxRetries: *cronJob.Spec.JobTemplate.Spec.BackoffLimit,
		Timeout:    durationpb.New(time.Duration(*cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)),
	}

	if len(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers) > 1 {
		warnings["CronJob"] = append(warnings["Cronjob"], &Warning{
			Kind:    "CronJob",
			Name:    cronJob.Name,
			Warning: "CronJob has more than one container. Only the first container will be migrated",
		})
	}

	if cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image ==
		deployment.Spec.Template.Spec.Containers[0].Image {
		cmd := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[0]
		args := append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[1:],
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args...)

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
	} else if keepGoing, err := common.PromptConfirm(`The cronjob does not fit the deployment image.
		Do you want to continue with a curl based cronjob?`, false); keepGoing && err == nil {
		fmt.Printf("Migrating cronjob %s to a curl based cronjob\n", cronJob.Name)
		fmt.Printf("This will create a new job that will run a curl command to the service\n")
		fmt.Printf("Current cmd and args are: %s %s\n",
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[0],
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args)
		urlString, err := common.PromptInput("Finish the path to the service",
			common.InputDefaultOpt(fmt.Sprintf("http://%s:[PORT]/[PATH]?[PARAMS1]&[PARAM2]", deployment.Name)))
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
