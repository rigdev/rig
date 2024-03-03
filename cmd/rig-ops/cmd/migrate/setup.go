package migrate

import (
	"context"
	"fmt"
	"net/url"
	"os"
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
	"github.com/spf13/cobra"
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

var promptAborted = "prompt aborted"

var platformDryRun bool

func Setup(parent *cobra.Command) {
	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate you kubernetes deployments to Rig Capsules",
		RunE:  base.Register(migrate),
	}

	migrate.Flags().BoolVar(&platformDryRun,
		"platform-dryrun",
		false,
		`Additionally perform a platform dryrun to compare to k8s resources from rig-platform.
		If true:
			- The rig platform must be running
			- A valid rig-cli config and context must be set either according to defaults or through flags
			- Valid access- and refresh tokens must be provided in the context. Otherwise a login is required.`,
	)

	parent.AddCommand(migrate)
}

func migrate(ctx context.Context,
	_ *cobra.Command,
	_ []string,
	cc client.Client,
	cr client.Reader,
	oc *base.OperatorClient,
) error {
	var rc rig.Client
	var err error
	if platformDryRun {
		rc, err = base.NewRigClient(ctx)
		if err != nil {
			return err
		}
	}

	deployment, err := getDeployment(ctx, cr)
	if err != nil || deployment == nil {
		return err
	}

	currentResources := &CurrentResources{
		Deployment: deployment,
		ConfigMaps: map[string]*corev1.ConfigMap{},
		Secrets:    map[string]*corev1.Secret{},
		Services:   map[string]*corev1.Service{},
		Ingresses:  map[string]*netv1.Ingress{},
		CronJobs:   map[string]*batchv1.CronJob{},
	}

	changes := []*capsule.Change{}

	fmt.Print("Migrating Deployment...")
	capsuleSpec, capsuleID, deploymentChanges, err := migrateDeployment(ctx, currentResources, cc, rc)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, deploymentChanges...)
	color.Green(" ✓")

	defer func() {
		if rc == nil {
			return
		}

		_, err := rc.Capsule().Delete(ctx, &connect.Request[capsule.DeleteRequest]{
			Msg: &capsule.DeleteRequest{
				CapsuleId: capsuleID,
				ProjectId: base.Flags.Project,
			},
		})
		if err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Print("Migrating Horizontal Pod Autoscaler...")
	hpaChanges, err := migrateHPA(ctx, cc, currentResources, &capsuleSpec.Spec)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, hpaChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating ConfigMaps and Secrets...")
	configChanges, err := migrateEnvironmentAndConfigFiles(ctx, cc, currentResources, &capsuleSpec.Spec)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, configChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Services and Ingress...")
	serviceChanges, err := migrateServicesAndIngresses(ctx, cc, currentResources, &capsuleSpec.Spec)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, serviceChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Cronjobs...")
	cronJobChanges, err := migrateCronJobs(ctx, cc, currentResources, &capsuleSpec.Spec)
	if err != nil && err.Error() == promptAborted {
		fmt.Println("Migrating Cronjobs...")
	} else if err != nil {
		color.Red(" ✗")
		return err
	}
	changes = append(changes, cronJobChanges...)
	color.Green(" ✓")

	capsuleSpecYAML, err := obj.Encode(capsuleSpec, cc.Scheme())
	if err != nil {
		return err
	}

	resp, err := oc.Pipeline.DryRun(ctx, connect.NewRequest(&pipeline.DryRunRequest{
		Namespace:   deployment.Namespace,
		Capsule:     capsuleID,
		CapsuleSpec: string(capsuleSpecYAML),
		Force:       true,
	}))
	if err != nil {
		return err
	}

	reports := map[string]map[string]*dyff.HumanReport{}
	err = processOperatorOutput(reports, currentResources, resp.Msg.OutputObjects)
	if err != nil {
		return err
	}

	var platformResources map[string]string
	if platformDryRun {
		deployResp, err := rc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
			Msg: &capsule.DeployRequest{
				CapsuleId:     capsuleID,
				ProjectId:     base.Flags.Project,
				EnvironmentId: base.Flags.Environment,
				Message:       "Migrated from kubernetes deployment",
				DryRun:        true,
				Changes:       changes,
			},
		})
		if err != nil {
			fmt.Println("Error deploying capsule", err)
			return err
		}

		platformResources = deployResp.Msg.GetResourceYaml()
		err = processPlatformOutput(reports, currentResources, platformResources)
		if err != nil {
			return err
		}
	}

	err = processRemainingResources(reports, currentResources)
	if err != nil {
		return err
	}

	err = promptDiffingChanges(reports)
	if err != nil && err.Error() != promptAborted {
		return err
	}

	return nil
}

func promptDiffingChanges(reports map[string]map[string]*dyff.HumanReport) error {
	choices := []string{}
	for kind := range reports {
		choices = append(choices, kind)
	}

	for {
		_, kind, err := common.PromptSelect("Select the resource kind to view the diff for. CTRL + C to continue", choices)
		if err != nil {
			return err
		}

		report := reports[kind]
		if len(report) == 1 {
			for _, r := range report {
				if err := r.WriteReport(os.Stdout); err != nil {
					return err
				}
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

				if err := report[name].WriteReport(os.Stdout); err != nil {
					return err
				}
			}
		}
	}
}

func migrateDeployment(ctx context.Context,
	currentResources *CurrentResources,
	cc client.Client,
	rc rig.Client,
) (*v1alpha2.Capsule, string, []*capsule.Change, error) {
	capsuleID := currentResources.Deployment.Name
	container := currentResources.Deployment.Spec.Template.Spec.Containers[0]
	changes := []*capsule.Change{}

	capsuleSpec := &v1alpha2.Capsule{
		ObjectMeta: v1.ObjectMeta{
			Name:      currentResources.Deployment.Name,
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
						// TODO Handle autoscaler
						Min: uint32(*currentResources.Deployment.Spec.Replicas),
					},
				},
			},
		},
	}

	cpu, memory := capsuleSpec.Spec.Scale.Vertical.CPU, capsuleSpec.Spec.Scale.Vertical.Memory
	if q, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
		cpu.Request = &q
	}
	if q, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
		memory.Request = &q
	}
	if q, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
		cpu.Limit = &q
	}
	if q, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
		memory.Limit = &q
	}

	containerSettings := &capsule.ContainerSettings{
		Resources: &capsule.Resources{
			Requests: &capsule.ResourceList{
				CpuMillis:   uint32(container.Resources.Requests.Cpu().MilliValue()),
				MemoryBytes: uint64(container.Resources.Requests.Memory().Value()),
			},
			Limits: &capsule.ResourceList{
				CpuMillis:   uint32(container.Resources.Limits.Cpu().MilliValue()),
				MemoryBytes: uint64(container.Resources.Limits.Memory().Value()),
			},
		},
	}

	if len(container.Command) > 0 {
		capsuleSpec.Spec.Command = container.Command[0]
		capsuleSpec.Spec.Args = container.Args
		containerSettings.Command = capsuleSpec.Spec.Command
		containerSettings.Args = capsuleSpec.Spec.Args
	}

	// Check if the deployment has a service account, and if so add it to the current resources
	if currentResources.Deployment.Spec.Template.Spec.ServiceAccountName != "" {
		serviceAccount := &corev1.ServiceAccount{}
		err := cc.Get(ctx, types.NamespacedName{
			Name:      currentResources.Deployment.Spec.Template.Spec.ServiceAccountName,
			Namespace: currentResources.Deployment.Namespace,
		}, serviceAccount)
		if kerrors.IsNotFound(err) {
			// TODO: warning we can't find it?
		} else if err != nil {
			return nil, "", nil, err
		} else {
			currentResources.ServiceAccount = serviceAccount
		}
	}

	if rc != nil {
		res, err := rc.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
			Msg: &capsule.GetRequest{
				CapsuleId: capsuleID,
				ProjectId: base.Flags.Project,
			},
		})
		if err != nil && !errors.IsNotFound(err) {
			return nil, "", nil, err
		}

		if res != nil && res.Msg.GetCapsule() != nil {
			capsuleID = fmt.Sprintf("%s-migrated", currentResources.Deployment.Name)

			capsule := &v1alpha2.Capsule{}
			err := cc.Get(ctx, types.NamespacedName{
				Name:      currentResources.Deployment.GetName(),
				Namespace: currentResources.Deployment.GetNamespace(),
			}, capsule)
			if err != nil {
				return nil, "", nil, err
			}

			currentResources.Capsule = capsule
		}

		if _, err := rc.Capsule().Create(ctx, &connect.Request[capsule.CreateRequest]{
			Msg: &capsule.CreateRequest{
				Name:      capsuleID,
				ProjectId: base.Flags.Project,
			},
		}); err != nil {
			return nil, "", nil, err
		}

		resp, err := rc.Build().Create(ctx, connect.NewRequest(&build.CreateRequest{
			CapsuleId:      capsuleID,
			Image:          currentResources.Deployment.Spec.Template.Spec.Containers[0].Image,
			SkipImageCheck: false,
			ProjectId:      base.Flags.Project,
		}))
		if err != nil {
			return nil, "", nil, err
		}

		changes = append(changes, []*capsule.Change{
			{
				Field: &capsule.Change_BuildId{
					BuildId: resp.Msg.GetBuildId(),
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
	}

	return capsuleSpec, capsuleID, changes, nil
}

func getDeployment(ctx context.Context, cc client.Reader) (*appsv1.Deployment, error) {
	deployments := &appsv1.DeploymentList{}
	err := cc.List(ctx, deployments, client.InNamespace(base.Flags.Namespace))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	deployment := &deployments.Items[i]

	if deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"] != "" {
		if keepGoing, err := common.PromptConfirm("This deployment is already owned by a capsule."+
			" Do you want to continue anyways?", false); !keepGoing || err != nil {
			return nil, err
		}
	}

	return deployment, nil
}

func migrateHPA(ctx context.Context,
	cc client.Client,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.CapsuleSpec,
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
			capsuleSpec.Scale.Horizontal.Instances.Max = ptr.To(uint32(hpa.Spec.MaxReplicas))
			capsuleSpec.Scale.Horizontal.Instances.Min = uint32(*hpa.Spec.MinReplicas)
			if metrics := hpa.Spec.Metrics; len(metrics) > 0 {
				for _, metric := range metrics {
					if metric.Resource != nil {
						if metric.Resource.Name == corev1.ResourceCPU && metric.Resource.Target.AverageUtilization != nil {
							capsuleSpec.Scale.Horizontal.CPUTarget = &v1alpha2.CPUTarget{
								Utilization: ptr.To(uint32(*metric.Resource.Target.AverageUtilization)),
							}
							horizontalScale.CpuTarget = &capsule.CPUTarget{
								AverageUtilizationPercentage: uint32(*metric.Resource.Target.AverageUtilization),
							}
						}
					}
					if metric.Object != nil {
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
						if metric.Object.Target.AverageValue != nil {
							objectMetric.ObjectMetric.AverageValue = metric.Object.Target.AverageValue.String()
							customMetric.GetObject().AverageValue = metric.Object.Target.AverageValue.String()
						} else if metric.Object.Target.Value != nil {
							objectMetric.ObjectMetric.Value = metric.Object.Target.Value.String()
							customMetric.GetObject().Value = metric.Object.Target.Value.String()
						}

						capsuleSpec.Scale.Horizontal.CustomMetrics = append(capsuleSpec.Scale.Horizontal.CustomMetrics, objectMetric)
						horizontalScale.CustomMetrics = append(horizontalScale.CustomMetrics, customMetric)
					}

					if metric.Pods != nil {
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

						if metric.Pods.Target.AverageValue != nil {
							podMetric.InstanceMetric.AverageValue = metric.Pods.Target.AverageValue.String()
							customMetric.GetInstance().AverageValue = metric.Pods.Target.AverageValue.String()
						}

						capsuleSpec.Scale.Horizontal.CustomMetrics = append(capsuleSpec.Scale.Horizontal.CustomMetrics, podMetric)
						horizontalScale.CustomMetrics = append(horizontalScale.CustomMetrics, customMetric)
					}
				}
			}
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_HorizontalScale{
					HorizontalScale: horizontalScale,
				},
			})
			found = true
		}
		if found {
			break
		}
	}

	return changes, nil
}

func migrateEnvironmentAndConfigFiles(ctx context.Context,
	cc client.Client,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.CapsuleSpec,
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

			currentResources.ConfigMaps["env-source"] = configMap
		} else if source.SecretRef != nil {
			envReferences = append(envReferences, v1alpha2.EnvReference{
				Kind: "Secret",
				Name: source.SecretRef.Name,
			})

			environmentSource = &capsule.EnvironmentSource{
				Kind: capsule.EnvironmentSource_KIND_CONFIG_MAP,
				Name: source.ConfigMapRef.Name,
			}

			secret := &corev1.Secret{}
			err := cc.Get(ctx, types.NamespacedName{
				Name:      source.SecretRef.Name,
				Namespace: currentResources.Deployment.Namespace,
			}, secret)
			if err != nil {
				return nil, err
			}

			currentResources.Secrets["env-source"] = secret
		}

		if environmentSource != nil {
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_SetEnvironmentSource{
					SetEnvironmentSource: environmentSource,
				},
			})
		}
	}
	capsuleSpec.Env = v1alpha2.Env{
		From: envReferences,
	}

	// Migrate ConfigMap and Secret files
	var files []v1alpha2.File
	for _, volume := range currentResources.Deployment.Spec.Template.Spec.Volumes {
		var file v1alpha2.File
		var configFile *capsule.Change_ConfigFile
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

			file = v1alpha2.File{
				Ref: &v1alpha2.FileContentReference{
					Kind: "ConfigMap",
					Name: volume.ConfigMap.Name,
					Key:  "content",
				},
			}

			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == volume.Name {
					file.Path = volumeMount.MountPath
					break
				}
			}

			configFile = &capsule.Change_ConfigFile{
				Path:     file.Path,
				IsSecret: false,
			}

			currentResources.ConfigMaps[file.Path] = configMap

			if len(configMap.BinaryData) > 0 {
				configFile.Content = configMap.BinaryData[volume.ConfigMap.Items[0].Key]
			} else if len(configMap.Data) > 0 {
				configFile.Content = []byte(configMap.Data[volume.ConfigMap.Items[0].Key])
			}
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
			}

			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == volume.Name {
					file.Path = volumeMount.MountPath
					break
				}
			}

			configFile = &capsule.Change_ConfigFile{
				Path:     file.Path,
				IsSecret: true,
			}

			currentResources.Secrets[file.Path] = secret

			if len(secret.Data) > 0 {
				configFile.Content = secret.Data[volume.Secret.Items[0].Key]
			} else if len(secret.StringData) > 0 {
				configFile.Content = []byte(secret.StringData[volume.Secret.Items[0].Key])
			}
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

	capsuleSpec.Files = files
	return changes, nil
}

func migrateServicesAndIngresses(ctx context.Context,
	cc client.Client,
	currentResources *CurrentResources,
	capsuleSpec *v1alpha2.CapsuleSpec,
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
			for _, servicePort := range service.Spec.Ports {
				found := false
				if servicePort.Name == port.Name {
					service := service
					currentResources.Services[service.GetName()] = &service
					found = true
				}
				if found {
					break
				}
			}
		}

		i := v1alpha2.CapsuleInterface{
			Name: port.Name,
			Port: port.ContainerPort,
		}

		ci := &capsule.Interface{
			Name: port.Name,
			Port: uint32(port.ContainerPort),
		}

		for _, ingress := range ingresses.Items {
			if i.Public != nil {
				break
			}

			var paths []string
			for _, path := range ingress.Spec.Rules[0].HTTP.Paths {
				if path.Backend.Service.Port.Name != port.Name {
					continue
				}

				paths = append(paths, path.Path)
			}

			if len(paths) > 0 {
				currentResources.Ingresses[ingress.GetName()] = &ingress

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

	capsuleSpec.Interfaces = interfaces
	changes := []*capsule.Change{
		{
			Field: &capsule.Change_Network{
				Network: &capsule.Network{
					Interfaces: capsuleInterfaces,
				},
			},
		},
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
	capsuleSpec *v1alpha2.CapsuleSpec,
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

		migratedCronJob, addCronjob, err := migrateCronJob(currentResources.Deployment, cronJob)
		if err != nil {
			fmt.Println(err)
			continue
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_AddCronJob{
				AddCronJob: addCronjob,
			},
		})
		capsuleSpec.CronJobs = append(migratedCronJobs, *migratedCronJob)
		currentResources.CronJobs[cronJob.Name] = &cronJob
		// remove the selected job from the list
		jobTitles = append(jobTitles[:i], jobTitles[i+1:]...)
		cronJobs = append(cronJobs[:i], cronJobs[i+1:]...)
	}

	return changes, nil
}

func migrateCronJob(deployment *appsv1.Deployment,
	cronJob batchv1.CronJob,
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

	if cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image ==
		deployment.Spec.Template.Spec.Containers[0].Image {
		migrated.Command = &v1alpha2.JobCommand{
			Command: cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[0],
			Args:    cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args,
		}

		capsuleCronjob.JobType = &capsule.CronJob_Command{
			Command: &capsule.JobCommand{
				Command: migrated.Command.Command,
				Args:    migrated.Command.Args,
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
