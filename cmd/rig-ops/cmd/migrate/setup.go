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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Setup(parent *cobra.Command) {
	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate you kubernetes deployments to Rig Capsules",
		RunE:  base.Register(migrate),
	}

	parent.AddCommand(migrate)
}

func migrate(ctx context.Context,
	_ *cobra.Command,
	_ []string,
	cc client.Client,
	cr client.Reader,
	rc rig.Client,
	oc *base.OperatorClient) error {
	deployment, err := getDeployment(ctx, cr)
	if err != nil || deployment == nil {
		return err
	}

	changes := []*capsule.Change{}

	fmt.Print("Migrating Deployment...")
	capsuleSpec, capsuleID, deploymentChanges, err := migrateDeployment(ctx, deployment, rc)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, deploymentChanges...)
	color.Green(" ✓")

	defer func() {
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

	fmt.Print("Migrating ConfigMaps and Secrets...")
	configChanges, err := migrateEnvironmentAndConfigFiles(ctx, cc, deployment, &capsuleSpec.Spec)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, configChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Services and Ingress...")
	serviceChanges, err := migrateServicesAndIngresses(ctx, cc, deployment, &capsuleSpec.Spec)
	if err != nil {
		color.Red(" ✗")
		return err
	}

	changes = append(changes, serviceChanges...)
	color.Green(" ✓")

	fmt.Print("Migrating Cronjobs...")
	cronJobChanges, err := migrateCronJobs(ctx, cc, deployment, &capsuleSpec.Spec)
	if err != nil {
		fmt.Println(err)
	}

	changes = append(changes, cronJobChanges...)
	fmt.Print("Migrating Cronjobs...")
	color.Green(" ✓")

	capsuleSpecYaml, err := obj.Encode(capsuleSpec, cc.Scheme())
	if err != nil {
		return err
	}

	resp, err := oc.Pipeline.DryRun(ctx, connect.NewRequest(&pipeline.DryRunRequest{
		Namespace:   deployment.Namespace,
		Capsule:     capsuleID,
		CapsuleSpec: string(capsuleSpecYaml),
	}))
	if err != nil {
		return err
	}

	fmt.Println(resp.Msg.OutputObjects)

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

	resources := deployResp.Msg.GetResourceYaml()
	fmt.Println("KUBERNETES RESOURCES:\n", resources)
	fmt.Println("CAPSULE SPEC:\n", string(capsuleSpecYaml))

	return nil
}

func migrateDeployment(ctx context.Context,
	deployment *appsv1.Deployment,
	rc rig.Client) (*v1alpha2.Capsule, string, []*capsule.Change, error) {
	capsuleID := deployment.Name
	{
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
			capsuleID = fmt.Sprintf("%s-migrated", deployment.Name)
		}
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
		Image:          deployment.Spec.Template.Spec.Containers[0].Image,
		SkipImageCheck: false,
		ProjectId:      base.Flags.Project,
	}))
	if err != nil {
		return nil, "", nil, err
	}

	capsuleSpec := &v1alpha2.Capsule{
		ObjectMeta: v1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		},
		Spec: v1alpha2.CapsuleSpec{
			Image: deployment.Spec.Template.Spec.Containers[0].Image,
			Scale: v1alpha2.CapsuleScale{
				Vertical: &v1alpha2.VerticalScale{
					CPU: &v1alpha2.ResourceLimits{
						Request: deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu(),
						Limit:   deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu(),
					},
					Memory: &v1alpha2.ResourceLimits{
						Request: deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory(),
						Limit:   deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory(),
					},
				},
				Horizontal: v1alpha2.HorizontalScale{
					Instances: v1alpha2.Instances{
						Min: uint32(*deployment.Spec.Replicas),
					},
				},
			},
		},
	}

	containerSettings := &capsule.ContainerSettings{
		Resources: &capsule.Resources{
			Requests: &capsule.ResourceList{
				CpuMillis:   uint32(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()),
				MemoryBytes: uint64(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Value()),
			},
			Limits: &capsule.ResourceList{
				CpuMillis:   uint32(deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()),
				MemoryBytes: uint64(deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value()),
			},
		},
	}

	if len(deployment.Spec.Template.Spec.Containers[0].Command) > 0 {
		capsuleSpec.Spec.Command = deployment.Spec.Template.Spec.Containers[0].Command[0]
		capsuleSpec.Spec.Args = deployment.Spec.Template.Spec.Containers[0].Args
		containerSettings.Command = capsuleSpec.Spec.Command
		containerSettings.Args = capsuleSpec.Spec.Args
	}

	capsuleChanges := []*capsule.Change{
		{
			Field: &capsule.Change_BuildId{
				BuildId: resp.Msg.GetBuildId(),
			},
		},
		{
			Field: &capsule.Change_Replicas{
				Replicas: uint32(*deployment.Spec.Replicas),
			},
		},
		{
			Field: &capsule.Change_ContainerSettings{
				ContainerSettings: containerSettings,
			},
		},
	}

	return capsuleSpec, capsuleID, capsuleChanges, nil
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
		// if deployment.GetObjectMeta().GetLabels()["rig.dev/owned-by-capsule"] == "" {
		deploymentNames = append(deploymentNames, []string{
			deployment.GetName(),
			deployment.GetNamespace(),
			fmt.Sprintf("    %d/%d    ", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas),
			fmt.Sprintf("     %d     ", deployment.Status.UpdatedReplicas),
			fmt.Sprintf("     %d     ", deployment.Status.AvailableReplicas),
			deployment.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		})
		// }
	}

	fmt.Printf("There are %d deployments of which %d are not owned by a capsule\n",
		len(deployments.Items), len(deploymentNames))
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

func migrateEnvironmentAndConfigFiles(ctx context.Context,
	cc client.Client,
	deployment *appsv1.Deployment,
	capsuleSpec *v1alpha2.CapsuleSpec) ([]*capsule.Change, error) {
	changes := []*capsule.Change{}
	// Migrate Environment Sources
	envReferences := []v1alpha2.EnvReference{}
	for _, source := range deployment.Spec.Template.Spec.Containers[0].EnvFrom {
		var envReference *v1alpha2.EnvReference
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
		} else if source.SecretRef != nil {
			envReferences = append(envReferences, v1alpha2.EnvReference{
				Kind: "Secret",
				Name: source.SecretRef.Name,
			})

			environmentSource = &capsule.EnvironmentSource{
				Kind: capsule.EnvironmentSource_KIND_CONFIG_MAP,
				Name: source.ConfigMapRef.Name,
			}
		}

		if envReference != nil {
			envReferences = append(envReferences, *envReference)
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_SetEnvironmentSource{
					SetEnvironmentSource: environmentSource,
				},
			})
		}
	}
	capsuleSpec.Env = &v1alpha2.Env{
		From: envReferences,
	}

	// Migrate ConfigMap and Secret files
	files := []v1alpha2.File{}
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		var file v1alpha2.File
		var configFile *capsule.Change_ConfigFile
		// If Volume is a ConfigMap
		if volume.ConfigMap != nil {
			configMap := &corev1.ConfigMap{}
			err := cc.Get(ctx, types.NamespacedName{Name: volume.ConfigMap.Name, Namespace: deployment.Namespace}, configMap)
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

			for _, volumeMount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
				if volumeMount.Name == volume.Name {
					file.Path = volumeMount.MountPath
					break
				}
			}

			configFile = &capsule.Change_ConfigFile{
				Path:     file.Path,
				IsSecret: false,
			}

			if len(configMap.BinaryData) > 0 {
				configFile.Content = configMap.BinaryData[volume.ConfigMap.Items[0].Key]
			} else if len(configMap.Data) > 0 {
				configFile.Content = []byte(configMap.Data[volume.ConfigMap.Items[0].Key])
			}
			// If Volume is a Secret
		} else if volume.Secret != nil {
			secret := &corev1.Secret{}
			err := cc.Get(ctx, types.NamespacedName{Name: volume.Secret.SecretName, Namespace: deployment.Namespace}, secret)
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

			for _, volumeMount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
				if volumeMount.Name == volume.Name {
					file.Path = volumeMount.MountPath
					break
				}
			}

			configFile = &capsule.Change_ConfigFile{
				Path:     file.Path,
				IsSecret: true,
			}

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
	deployment *appsv1.Deployment,
	capsuleSpec *v1alpha2.CapsuleSpec) ([]*capsule.Change, error) {
	livenessProbe := deployment.Spec.Template.Spec.Containers[0].LivenessProbe
	readinessProbe := deployment.Spec.Template.Spec.Containers[0].ReadinessProbe

	ingresses := &netv1.IngressList{}
	err := cc.List(ctx, ingresses, client.InNamespace(deployment.GetNamespace()))
	if err != nil {
		return nil, err
	}

	interfaces := make([]v1alpha2.CapsuleInterface, 0, len(deployment.Spec.Template.Spec.Containers[0].Ports))
	capsuleInterfaces := make([]*capsule.Interface, 0, len(deployment.Spec.Template.Spec.Containers[0].Ports))

	for _, port := range deployment.Spec.Template.Spec.Containers[0].Ports {
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
	port corev1.ContainerPort) (*v1alpha2.InterfaceProbe, *capsule.InterfaceProbe, error) {
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
	deployment *appsv1.Deployment,
	capsuleSpec *v1alpha2.CapsuleSpec) ([]*capsule.Change, error) {
	cronJobList := &batchv1.CronJobList{}
	err := cc.List(ctx, cronJobList, client.InNamespace(deployment.GetNamespace()))
	if err != nil {
		return nil, err
	}

	cronJobs := cronJobList.Items
	headers := []string{"NAME", "SCHEDULE", "IMAGE", "LAST SCHEDULE", "AGE"}

	jobTitles := [][]string{}
	for _, cronJob := range cronJobList.Items {
		jobTitles = append(jobTitles, []string{
			cronJob.GetName(),
			cronJob.Spec.Schedule,
			strings.Split(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image, "@")[0],
			cronJob.Status.LastScheduleTime.Format("2006-01-02 15:04:05"),
			cronJob.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		})
	}

	migratedCronJobs := make([]v1alpha2.CronJob, 0, len(cronJobs))
	changes := []*capsule.Change{}
	// fmt.Printf("There are %d cron jobs in the same namespace with the same image\n", len(jobTitles))
	for {
		i, err := common.PromptTableSelect("\nSelect a job to migrate or CTRL+C to continue",
			jobTitles, headers, common.SelectEnableFilterOpt, common.SelectDontShowResultOpt)
		if err != nil {
			break
		}

		cronJob := cronJobs[i]

		migratedCronJob, addCronjob, err := migrateCronJob(deployment, cronJob)
		if err != nil {
			fmt.Println(err)
			continue
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_AddCronJob{
				AddCronJob: addCronjob,
			},
		})
		migratedCronJobs = append(migratedCronJobs, *migratedCronJob)
		// remove the selected job from the list
		jobTitles = append(jobTitles[:i], jobTitles[i+1:]...)
		cronJobs = append(cronJobs[:i], cronJobs[i+1:]...)
	}

	capsuleSpec.CronJobs = migratedCronJobs

	return changes, nil
}

func migrateCronJob(deployment *appsv1.Deployment,
	cronJob batchv1.CronJob) (*v1alpha2.CronJob, *capsule.CronJob, error) {
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
