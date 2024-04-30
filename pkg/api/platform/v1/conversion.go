package v1

import (
	"fmt"
	"maps"
	"reflect"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	v2 "github.com/rigdev/rig-go-api/k8s.io/api/autoscaling/v2"
	"github.com/rigdev/rig-go-api/k8s.io/apimachinery/pkg/api/resource"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig-go-api/v1alpha2"
	types_v1alpha2 "github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

func RolloutConfigToCapsuleSpecExtension(rc *capsule.RolloutConfig) (*platformv1.CapsuleSpecExtension, error) {
	replicas := rc.GetReplicas()
	spec := &platformv1.CapsuleSpecExtension{
		Kind:       "CapsuleSpecExtension",
		ApiVersion: "v1", // TODO
		Image:      rc.GetImageId(),
		Command:    rc.GetContainerSettings().GetCommand(),
		Args:       rc.GetContainerSettings().GetArgs(),
		Scale: &v1alpha2.CapsuleScale{
			Vertical: makeVerticalScale(
				rc.GetContainerSettings().GetResources().GetRequests(),
				rc.GetContainerSettings().GetResources().GetLimits(),
				rc.GetContainerSettings().GetResources().GetGpuLimits(),
			),
			Horizontal: &v1alpha2.HorizontalScale{
				Instances: &v1alpha2.Instances{
					Min: rc.GetReplicas(),
				},
			},
		},
		Annotations:               maps.Clone(rc.GetAnnotations()),
		AutoAddRigServiceAccounts: rc.GetAutoAddRigServiceAccounts(),
	}

	horizontal := rc.GetHorizontalScale()
	if horizontal.GetCpuTarget().GetAverageUtilizationPercentage() > 0 {
		spec.Scale.Horizontal.CpuTarget = &v1alpha2.CPUTarget{
			Utilization: horizontal.GetCpuTarget().GetAverageUtilizationPercentage(),
		}
	}

	for _, m := range horizontal.GetCustomMetrics() {
		var metric v1alpha2.CustomMetric
		if obj := m.GetObject(); obj != nil {
			metric.ObjectMetric = &v1alpha2.ObjectMetric{
				MetricName:   obj.MetricName,
				MatchLabels:  obj.MatchLabels,
				AverageValue: obj.AverageValue,
				Value:        obj.Value,
				ObjectReference: &v2.CrossVersionObjectReference{
					Kind:       obj.ObjectReference.Kind,
					Name:       obj.ObjectReference.Name,
					ApiVersion: obj.ObjectReference.ApiVersion,
				},
			}
		} else if instance := m.GetInstance(); instance != nil {
			metric.InstanceMetric = &v1alpha2.InstanceMetric{
				MetricName:   instance.MetricName,
				MatchLabels:  instance.MatchLabels,
				AverageValue: instance.AverageValue,
			}
		}
		spec.Scale.Horizontal.CustomMetrics = append(spec.Scale.Horizontal.CustomMetrics, &metric)
	}

	if len(spec.Scale.Horizontal.CustomMetrics) > 0 || spec.Scale.Horizontal.CpuTarget != nil {
		if horizontal.GetMinReplicas() > replicas {
			spec.Scale.Horizontal.Instances.Min = horizontal.GetMinReplicas()
		}
		spec.Scale.Horizontal.Instances.Max = horizontal.GetMaxReplicas()
	}

	spec.EnvironmentVariables = &platformv1.EnvironmentVariables{
		Direct: maps.Clone(rc.GetContainerSettings().GetEnvironmentVariables()),
	}

	for _, es := range rc.GetContainerSettings().GetEnvironmentSources() {
		ref := &platformv1.EnvironmentSource{
			Name: es.GetName(),
		}
		switch es.GetKind() {
		case capsule.EnvironmentSource_KIND_CONFIG_MAP:
			ref.Kind = string(EnvironmentSourceKindConfigMap)
		case capsule.EnvironmentSource_KIND_SECRET:
			ref.Kind = string(EnvironmentSourceKindSecret)
		default:
			return nil, errors.InvalidArgumentErrorf("invalid environment source kind '%s'", es.GetKind())
		}
		spec.EnvironmentVariables.Sources = append(spec.EnvironmentVariables.Sources, ref)
	}

	for _, cf := range rc.GetConfigFiles() {
		spec.ConfigFiles = append(spec.ConfigFiles, &platformv1.ConfigFile{
			Path:     cf.GetPath(),
			Content:  cf.GetContent(),
			IsSecret: cf.GetIsSecret(),
		})
	}

	for _, i := range rc.GetNetwork().GetInterfaces() {
		capIf := &v1alpha2.CapsuleInterface{
			Name: i.GetName(),
			Port: int32(i.GetPort()),
		}

		// Deprecate Public interface by porting the ingress to a route.
		if i.GetPublic().GetEnabled() {
			switch v := i.GetPublic().GetMethod().GetKind().(type) {
			case *capsule.RoutingMethod_Ingress_:
				route := &v1alpha2.HostRoute{
					Host: v.Ingress.GetHost(),
				}
				for _, p := range v.Ingress.GetPaths() {
					route.Paths = append(route.Paths, &v1alpha2.HTTPPathRoute{
						Path:  p,
						Match: string(types_v1alpha2.PathPrefix),
					})
				}

				capIf.Routes = append(capIf.Routes, route)
			}
		}

		for _, r := range i.GetRoutes() {
			route := &v1alpha2.HostRoute{
				Id:          r.GetId(),
				Host:        r.GetHost(),
				Annotations: r.GetOptions().GetAnnotations(),
			}

			for _, p := range r.GetPaths() {
				path := &v1alpha2.HTTPPathRoute{
					Path: p.GetPath(),
				}

				switch p.Match {
				case capsule.PathMatchType_PATH_MATCH_TYPE_EXACT:
					path.Match = string(types_v1alpha2.Exact)
				case capsule.PathMatchType_PATH_MATCH_TYPE_PATH_PREFIX,
					capsule.PathMatchType_PATH_MATCH_TYPE_UNSPECIFIED:
					path.Match = string(types_v1alpha2.PathPrefix)
				default:
					return nil, errors.InvalidArgumentErrorf("invalid path match type '%v'", p.Match)
				}

				route.Paths = append(route.Paths, path)
			}

			capIf.Routes = append(capIf.Routes, route)
		}

		var err error
		if capIf.Liveness, err = getInterfaceProbe(i.GetLiveness()); err != nil {
			return nil, err
		}

		if capIf.Readiness, err = getInterfaceProbe(i.GetReadiness()); err != nil {
			return nil, err
		}

		spec.Interfaces = append(spec.Interfaces, capIf)
	}

	for _, j := range rc.GetCronJobs() {
		var timeoutSeconds uint64
		if t := j.GetTimeout(); t != nil {
			timeoutSeconds = uint64(t.AsDuration().Seconds())
		}
		job := &v1alpha2.CronJob{
			Name:           j.GetJobName(),
			Schedule:       j.GetSchedule(),
			MaxRetries:     uint64(j.GetMaxRetries()),
			TimeoutSeconds: timeoutSeconds,
		}
		switch v := j.GetJobType().(type) {
		case *capsule.CronJob_Command:
			job.Command = &v1alpha2.JobCommand{
				Command: v.Command.GetCommand(),
				Args:    v.Command.GetArgs(),
			}
		case *capsule.CronJob_Url:
			job.Url = &v1alpha2.URL{
				Port:            uint32(v.Url.GetPort()),
				Path:            v.Url.GetPath(),
				QueryParameters: v.Url.GetQueryParameters(),
			}
		}
		spec.CronJobs = append(spec.CronJobs, job)
	}

	return spec, nil
}

func getInterfaceProbe(p *capsule.InterfaceProbe) (*v1alpha2.InterfaceProbe, error) {
	switch v := p.GetKind().(type) {
	case nil:
		return nil, nil
	case *capsule.InterfaceProbe_Http:
		return &v1alpha2.InterfaceProbe{
			Path: v.Http.GetPath(),
		}, nil
	case *capsule.InterfaceProbe_Tcp:
		return &v1alpha2.InterfaceProbe{
			Tcp: true,
		}, nil
	case *capsule.InterfaceProbe_Grpc:
		return &v1alpha2.InterfaceProbe{
			Grpc: &v1alpha2.InterfaceGRPCProbe{
				Service: v.Grpc.GetService(),
			},
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("unknown interface probe '%v'", reflect.TypeOf(v))
	}
}

func makeVerticalScale(
	requests *capsule.ResourceList,
	limits *capsule.ResourceList,
	gpuLimits *capsule.GpuLimits,
) *v1alpha2.VerticalScale {
	vs := &v1alpha2.VerticalScale{
		Cpu:    &v1alpha2.ResourceLimits{},
		Memory: &v1alpha2.ResourceLimits{},
	}

	if cpu := limits.GetCpuMillis(); cpu > 0 {
		vs.Cpu.Limit = &resource.Quantity{String_: fmt.Sprintf("%v", float64(cpu)/1000.)}
	}
	if memory := limits.GetMemoryBytes(); memory > 0 {
		vs.Memory.Limit = &resource.Quantity{String_: fmt.Sprintf("%v", memory)}
	}

	if cpu := requests.GetCpuMillis(); cpu > 0 {
		vs.Cpu.Request = &resource.Quantity{String_: fmt.Sprintf("%v", float64(cpu)/1000.)}
	}
	if memory := requests.GetMemoryBytes(); memory > 0 {
		vs.Memory.Request = &resource.Quantity{String_: fmt.Sprintf("%v", memory)}
	}

	if gpuLimits != nil {
		if gpu := gpuLimits.GetCount(); gpu > 0 {
			vs.Gpu = &v1alpha2.ResourceRequest{
				Request: &resource.Quantity{String_: fmt.Sprintf("%v", gpu)},
			}
		}
	}

	if vs.Cpu.Limit == nil && vs.Cpu.Request == nil {
		vs.Cpu = nil
	}
	if vs.Memory.Limit == nil && vs.Memory.Request == nil {
		vs.Memory = nil
	}

	return vs
}

func CapsuleSpecExtensionToRolloutConfig(spec *platformv1.CapsuleSpecExtension) (*capsule.RolloutConfig, error) {
	config := &capsule.RolloutConfig{
		ImageId: spec.GetImage(),
		Network: makeNetworks(spec.GetInterfaces()),
		ContainerSettings: &capsule.ContainerSettings{
			EnvironmentVariables: maps.Clone(spec.GetEnvironmentVariables().GetDirect()),
			Command:              spec.GetCommand(),
			Args:                 spec.GetArgs(),
			EnvironmentSources:   makeEnvironmentSources(spec.GetEnvironmentVariables().GetSources()),
		},
		AutoAddRigServiceAccounts: spec.GetAutoAddRigServiceAccounts(),
		ConfigFiles:               makeConfigFiles(spec.GetConfigFiles()),
		HorizontalScale:           makeHorizontalScale(spec.GetScale().GetHorizontal()),
		CronJobs:                  makeCronJobs(spec.GetCronJobs()),
		Annotations:               maps.Clone(spec.GetAnnotations()),
	}

	resources, err := makeResources(spec.GetScale().GetVertical())
	if err != nil {
		return nil, err
	}
	config.ContainerSettings.Resources = resources

	return config, nil
}

func makeEnvironmentSources(spec []*platformv1.EnvironmentSource) []*capsule.EnvironmentSource {
	var res []*capsule.EnvironmentSource
	for _, e := range spec {
		ee := &capsule.EnvironmentSource{
			Name: e.Name,
		}
		switch e.GetKind() {
		case string(EnvironmentSourceKindConfigMap):
			ee.Kind = capsule.EnvironmentSource_KIND_CONFIG_MAP
		case string(EnvironmentSourceKindSecret):
			ee.Kind = capsule.EnvironmentSource_KIND_SECRET
		default:
			ee.Kind = capsule.EnvironmentSource_KIND_UNSPECIFIED
		}
		res = append(res, ee)
	}
	return res
}

func makeResources(vertical *v1alpha2.VerticalScale) (*capsule.Resources, error) {
	res := &capsule.Resources{
		Requests:  &capsule.ResourceList{},
		Limits:    &capsule.ResourceList{},
		GpuLimits: &capsule.GpuLimits{},
	}

	cpuReq, cpuLimit, err := parseLimits(vertical.GetCpu())
	if err != nil {
		return nil, err
	}
	res.Limits.CpuMillis = uint32(cpuLimit.MilliValue())
	res.Requests.CpuMillis = uint32(cpuReq.MilliValue())

	memReq, memLim, err := parseLimits(vertical.GetMemory())
	if err != nil {
		return nil, err
	}
	res.Requests.MemoryBytes = uint64(memReq.Value())
	res.Limits.MemoryBytes = uint64(memLim.Value())

	gpu, err := k8sresource.ParseQuantity(vertical.GetGpu().GetRequest().String_)
	if err != nil {
		return nil, err
	}
	// TODO Type
	res.GpuLimits.Count = uint32(gpu.Value())

	return res, nil
}

func parseLimits(r *v1alpha2.ResourceLimits) (k8sresource.Quantity, k8sresource.Quantity, error) {
	var empty k8sresource.Quantity
	req, err := k8sresource.ParseQuantity(r.GetRequest().String_)
	if err != nil {
		return empty, empty, err
	}
	limit, err := k8sresource.ParseQuantity(r.GetLimit().String_)
	if err != nil {
		return empty, empty, err
	}
	return req, limit, err
}

func makeHorizontalScale(spec *v1alpha2.HorizontalScale) *capsule.HorizontalScale {
	res := &capsule.HorizontalScale{
		MaxReplicas: spec.GetInstances().GetMax(),
		MinReplicas: spec.GetInstances().GetMin(),
	}
	if cpu := spec.GetCpuTarget(); cpu != nil {
		res.CpuTarget = &capsule.CPUTarget{
			AverageUtilizationPercentage: cpu.GetUtilization(),
		}
	}

	for _, metric := range spec.GetCustomMetrics() {
		m := &capsule.CustomMetric{}
		if instance := metric.GetInstanceMetric(); instance != nil {
			m.Metric = &capsule.CustomMetric_Instance{
				Instance: &capsule.InstanceMetric{
					MetricName:   instance.GetMetricName(),
					MatchLabels:  maps.Clone(instance.GetMatchLabels()),
					AverageValue: instance.GetAverageValue(),
				},
			}
		} else if object := metric.GetObjectMetric(); object != nil {
			m.Metric = &capsule.CustomMetric_Object{
				Object: &capsule.ObjectMetric{
					MetricName:   object.GetMetricName(),
					MatchLabels:  maps.Clone(object.GetMatchLabels()),
					AverageValue: object.GetAverageValue(),
					Value:        object.GetValue(),
					ObjectReference: &capsule.ObjectReference{
						Kind:       object.GetObjectReference().GetKind(),
						Name:       object.GetObjectReference().GetName(),
						ApiVersion: object.GetObjectReference().GetApiVersion(),
					},
				},
			}
		} else {
			continue
		}
		res.CustomMetrics = append(res.CustomMetrics, m)
	}

	return res
}

func makeConfigFiles(configFiles []*platformv1.ConfigFile) []*capsule.ConfigFile {
	var res []*capsule.ConfigFile

	for _, c := range configFiles {
		res = append(res, &capsule.ConfigFile{
			Path:     c.GetPath(),
			Content:  c.GetContent(),
			IsSecret: c.GetIsSecret(),
		})
	}

	return res
}

func makeCronJobs(cronJobs []*v1alpha2.CronJob) []*capsule.CronJob {
	var res []*capsule.CronJob

	for _, j := range cronJobs {
		job := &capsule.CronJob{
			JobName:    j.GetName(),
			Schedule:   j.GetSchedule(),
			MaxRetries: int32(j.GetMaxRetries()),
			// Timeout:    j.GetTimeoutSeconds(),
			JobType: nil,
		}
		if j.GetTimeoutSeconds() != 0 {
			job.Timeout = durationpb.New(time.Second * time.Duration(j.GetTimeoutSeconds()))
		}
		if cmd := j.GetCommand(); cmd != nil {
			job.JobType = &capsule.CronJob_Command{
				Command: &capsule.JobCommand{
					Command: cmd.GetCommand(),
					Args:    cmd.GetArgs(),
				},
			}
		} else if url := j.GetUrl(); url != nil {
			job.JobType = &capsule.CronJob_Url{
				Url: &capsule.JobURL{
					Port:            uint64(url.GetPort()),
					Path:            url.GetPath(),
					QueryParameters: maps.Clone(url.GetQueryParameters()),
				},
			}
		} else {
			continue
		}
		res = append(res, job)
	}

	return res
}

func makeNetworks(spec []*v1alpha2.CapsuleInterface) *capsule.Network {
	res := &capsule.Network{}
	for _, i := range spec {
		ii := &capsule.Interface{
			Port:      uint32(i.GetPort()),
			Name:      i.GetName(),
			Liveness:  makeInterfaceProbe(i.GetLiveness()),
			Readiness: makeInterfaceProbe(i.GetReadiness()),
			Routes:    []*capsule.HostRoute{},
		}
		for _, r := range i.GetRoutes() {
			ii.Routes = append(ii.Routes, makeHostRoute(r))
		}
		if len(ii.Routes) == 0 {
			ii.Routes = nil
		}
		res.Interfaces = append(res.Interfaces, ii)
	}
	return res
}

func makeInterfaceProbe(probe *v1alpha2.InterfaceProbe) *capsule.InterfaceProbe {
	if probe == nil {
		return nil
	}

	r := &capsule.InterfaceProbe{}
	if path := probe.GetPath(); path != "" {
		r.Kind = &capsule.InterfaceProbe_Http{
			Http: &capsule.InterfaceProbe_HTTP{
				Path: path,
			},
		}
	} else if grpc := probe.GetGrpc(); grpc != nil {
		r.Kind = &capsule.InterfaceProbe_Grpc{
			Grpc: &capsule.InterfaceProbe_GRPC{
				Service: grpc.GetService(),
			},
		}
	} else if probe.GetTcp() {
		r.Kind = &capsule.InterfaceProbe_Tcp{
			Tcp: &capsule.InterfaceProbe_TCP{},
		}
	} else {
		return nil
	}

	return r
}

func makeHostRoute(route *v1alpha2.HostRoute) *capsule.HostRoute {
	res := &capsule.HostRoute{
		Host: route.GetHost(),
		Options: &capsule.RouteOptions{
			Annotations: maps.Clone(route.GetAnnotations()),
		},
		Id: route.GetId(),
	}
	for _, p := range route.GetPaths() {
		pp := &capsule.HTTPPathRoute{
			Path: p.GetPath(),
		}
		switch p.GetMatch() {
		case string(types_v1alpha2.Exact):
			pp.Match = capsule.PathMatchType_PATH_MATCH_TYPE_EXACT
		case string(types_v1alpha2.PathPrefix):
			pp.Match = capsule.PathMatchType_PATH_MATCH_TYPE_PATH_PREFIX
		}
		res.Paths = append(res.Paths, pp)
	}

	return res
}
