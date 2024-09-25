package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	v2 "github.com/rigdev/rig-go-api/k8s.io/api/autoscaling/v2"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func RolloutConfigToCapsuleSpec(rc *capsule.RolloutConfig) (*platformv1.CapsuleSpec, error) {
	spec := &platformv1.CapsuleSpec{
		Image: rc.GetImageId(),
		Scale: &platformv1.Scale{
			Horizontal: HorizontalScaleConversion(rc.GetHorizontalScale(), rc.GetReplicas()),
		},
		Annotations:               maps.Clone(rc.GetAnnotations()),
		AutoAddRigServiceAccounts: rc.GetAutoAddRigServiceAccounts(),
	}

	if err := FeedContainerSettings(spec, rc.GetContainerSettings()); err != nil {
		return nil, err
	}

	for _, cf := range rc.GetConfigFiles() {
		f := &platformv1.File{
			Path:     cf.GetPath(),
			AsSecret: cf.GetIsSecret(),
		}
		if ValidString(cf.GetContent()) {
			f.String_ = string(cf.GetContent())
		} else {
			f.Bytes = cf.GetContent()
		}
		spec.Files = append(spec.Files, f)
	}

	for _, i := range rc.GetNetwork().GetInterfaces() {
		capI, err := InterfaceConversion(i)
		if err != nil {
			return nil, err
		}
		spec.Interfaces = append(spec.Interfaces, capI)
	}

	for _, j := range rc.GetCronJobs() {
		spec.CronJobs = append(spec.CronJobs, CronJobConversion(j))
	}

	return spec, nil
}

func CronJobConversion(j *capsule.CronJob) *platformv1.CronJob {
	var timeoutSeconds uint64
	if t := j.GetTimeout(); t != nil {
		timeoutSeconds = uint64(t.AsDuration().Seconds())
	}
	job := &platformv1.CronJob{
		Name:           j.GetJobName(),
		Schedule:       j.GetSchedule(),
		MaxRetries:     uint64(j.GetMaxRetries()),
		TimeoutSeconds: timeoutSeconds,
	}
	switch v := j.GetJobType().(type) {
	case *capsule.CronJob_Command:
		job.Command = &platformv1.JobCommand{
			Command: v.Command.GetCommand(),
			Args:    v.Command.GetArgs(),
		}
	case *capsule.CronJob_Url:
		job.Url = &platformv1.URL{
			Port:            uint32(v.Url.GetPort()),
			Path:            v.Url.GetPath(),
			QueryParameters: v.Url.GetQueryParameters(),
		}
	}
	return job
}

func HorizontalScaleConversion(horizontal *capsule.HorizontalScale, replicas uint32) *platformv1.HorizontalScale {
	res := &platformv1.HorizontalScale{
		Min: max(replicas, horizontal.GetMinReplicas()),
	}

	if horizontal.GetCpuTarget().GetAverageUtilizationPercentage() > 0 {
		res.CpuTarget = &platformv1.CPUTarget{
			Utilization: horizontal.GetCpuTarget().GetAverageUtilizationPercentage(),
		}
	}

	for _, m := range horizontal.GetCustomMetrics() {
		var metric platformv1.CustomMetric
		if obj := m.GetObject(); obj != nil {
			metric.ObjectMetric = &platformv1.ObjectMetric{
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
			metric.InstanceMetric = &platformv1.InstanceMetric{
				MetricName:   instance.MetricName,
				MatchLabels:  instance.MatchLabels,
				AverageValue: instance.AverageValue,
			}
		}
		res.CustomMetrics = append(res.CustomMetrics, &metric)
	}

	if len(res.CustomMetrics) > 0 || res.CpuTarget != nil {
		res.Max = horizontal.GetMaxReplicas()
	}

	return res
}

func FeedContainerSettings(spec *platformv1.CapsuleSpec, containerSettings *capsule.ContainerSettings) error {
	if spec.Scale == nil {
		spec.Scale = &platformv1.Scale{}
	}
	spec.Scale.Vertical = makeVerticalScale(
		containerSettings.GetResources().GetRequests(),
		containerSettings.GetResources().GetLimits(),
		containerSettings.GetResources().GetGpuLimits(),
	)
	spec.Env = &platformv1.EnvironmentVariables{
		Raw: maps.Clone(containerSettings.GetEnvironmentVariables()),
	}
	if spec.Env.Raw == nil {
		spec.Env.Raw = map[string]string{}
	}

	for _, es := range containerSettings.GetEnvironmentSources() {
		ref, err := EnvironmentSourceConversion(es)
		if err != nil {
			return err
		}
		spec.Env.Sources = append(spec.Env.Sources, ref)
	}
	spec.Command = containerSettings.GetCommand()
	spec.Args = containerSettings.GetArgs()
	return nil
}

func EnvironmentSourceConversion(source *capsule.EnvironmentSource) (*platformv1.EnvironmentSource, error) {
	if err := common.ValidateSystemName(source.Name); err != nil {
		return nil, errors.InvalidArgumentErrorf("invalid environment source name; %v", err)
	}

	ref := &platformv1.EnvironmentSource{
		Name: source.GetName(),
	}
	switch source.GetKind() {
	case capsule.EnvironmentSource_KIND_CONFIG_MAP:
		ref.Kind = string(EnvironmentSourceKindConfigMap)
	case capsule.EnvironmentSource_KIND_SECRET:
		ref.Kind = string(EnvironmentSourceKindSecret)
	default:
		return nil, errors.InvalidArgumentErrorf("invalid environment source kind '%s'", source.GetKind())
	}

	return ref, nil
}

func EnvironmentSourceSpecConversion(source *platformv1.EnvironmentSource) *capsule.EnvironmentSource {
	ref := &capsule.EnvironmentSource{
		Name: source.GetName(),
	}
	switch source.GetKind() {
	case string(EnvironmentSourceKindConfigMap):
		ref.Kind = capsule.EnvironmentSource_KIND_CONFIG_MAP
	case string(EnvironmentSourceKindSecret):
		ref.Kind = capsule.EnvironmentSource_KIND_SECRET
	}
	return ref
}

func InterfaceConversion(i *capsule.Interface) (*platformv1.CapsuleInterface, error) {
	capIf := &platformv1.CapsuleInterface{
		Name: i.GetName(),
		Port: int32(i.GetPort()),
	}

	// Deprecate Public interface by porting the ingress to a route.
	if i.GetPublic().GetEnabled() {
		switch v := i.GetPublic().GetMethod().GetKind().(type) {
		case *capsule.RoutingMethod_Ingress_:
			route := &platformv1.HostRoute{
				Host: v.Ingress.GetHost(),
			}
			for _, p := range v.Ingress.GetPaths() {
				route.Paths = append(route.Paths, &platformv1.HTTPPathRoute{
					Path:  p,
					Match: string(PathPrefix),
				})
			}

			capIf.Routes = append(capIf.Routes, route)
		}
	}

	for _, r := range i.GetRoutes() {
		route := &platformv1.HostRoute{
			Id:          r.GetId(),
			Host:        r.GetHost(),
			Annotations: r.GetOptions().GetAnnotations(),
		}

		for _, p := range r.GetPaths() {
			path := &platformv1.HTTPPathRoute{
				Path: p.GetPath(),
			}

			switch p.Match {
			case capsule.PathMatchType_PATH_MATCH_TYPE_EXACT:
				path.Match = string(Exact)
			case capsule.PathMatchType_PATH_MATCH_TYPE_REGULAR_EXPRESSION:
				path.Match = string(RegularExpression)
			case capsule.PathMatchType_PATH_MATCH_TYPE_PATH_PREFIX,
				capsule.PathMatchType_PATH_MATCH_TYPE_UNSPECIFIED:
				path.Match = string(PathPrefix)
			default:
				return nil, errors.InvalidArgumentErrorf("invalid path match type '%v'", p.Match)
			}

			route.Paths = append(route.Paths, path)
		}

		capIf.Routes = append(capIf.Routes, route)
	}

	var err error
	if capIf.Liveness, err = getInterfaceLivenessProbe(i.GetLiveness()); err != nil {
		return nil, err
	}

	if capIf.Readiness, err = getInterfaceReadinessProbe(i.GetReadiness()); err != nil {
		return nil, err
	}

	return capIf, nil
}

func getInterfaceLivenessProbe(p *capsule.InterfaceProbe) (*platformv1.InterfaceLivenessProbe, error) {
	switch v := p.GetKind().(type) {
	case nil:
		return nil, nil
	case *capsule.InterfaceProbe_Http:
		return &platformv1.InterfaceLivenessProbe{
			Path: v.Http.GetPath(),
		}, nil
	case *capsule.InterfaceProbe_Tcp:
		return &platformv1.InterfaceLivenessProbe{
			Tcp: true,
		}, nil
	case *capsule.InterfaceProbe_Grpc:
		return &platformv1.InterfaceLivenessProbe{
			Grpc: &platformv1.InterfaceGRPCProbe{
				Service: v.Grpc.GetService(),
			},
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("unknown interface probe '%v'", reflect.TypeOf(v))
	}
}

func getInterfaceReadinessProbe(p *capsule.InterfaceProbe) (*platformv1.InterfaceReadinessProbe, error) {
	switch v := p.GetKind().(type) {
	case nil:
		return nil, nil
	case *capsule.InterfaceProbe_Http:
		return &platformv1.InterfaceReadinessProbe{
			Path: v.Http.GetPath(),
		}, nil
	case *capsule.InterfaceProbe_Tcp:
		return &platformv1.InterfaceReadinessProbe{
			Tcp: true,
		}, nil
	case *capsule.InterfaceProbe_Grpc:
		return &platformv1.InterfaceReadinessProbe{
			Grpc: &platformv1.InterfaceGRPCProbe{
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
) *platformv1.VerticalScale {
	vs := &platformv1.VerticalScale{
		Cpu:    &platformv1.ResourceLimits{},
		Memory: &platformv1.ResourceLimits{},
	}

	if cpu := limits.GetCpuMillis(); cpu > 0 {
		vs.Cpu.Limit = fmt.Sprintf("%v", float64(cpu)/1000.)
	}
	if memory := limits.GetMemoryBytes(); memory > 0 {
		vs.Memory.Limit = fmt.Sprintf("%v", memory)
	}

	if cpu := requests.GetCpuMillis(); cpu > 0 {
		vs.Cpu.Request = fmt.Sprintf("%v", float64(cpu)/1000.)
	}
	if memory := requests.GetMemoryBytes(); memory > 0 {
		vs.Memory.Request = fmt.Sprintf("%v", memory)
	}

	if gpuLimits != nil {
		if gpu := gpuLimits.GetCount(); gpu > 0 {
			vs.Gpu = &platformv1.ResourceRequest{
				Request: fmt.Sprintf("%v", gpu),
			}
		}
	}

	if vs.Cpu.Limit == "" && vs.Cpu.Request == "" {
		vs.Cpu = nil
	}
	if vs.Memory.Limit == "" && vs.Memory.Request == "" {
		vs.Memory = nil
	}

	return vs
}

func CapsuleSpecToRolloutConfig(spec *platformv1.CapsuleSpec) (*capsule.RolloutConfig, error) {
	config := &capsule.RolloutConfig{
		ImageId:                   spec.GetImage(),
		Network:                   makeNetworks(spec.GetInterfaces()),
		AutoAddRigServiceAccounts: spec.GetAutoAddRigServiceAccounts(),
		ConfigFiles:               makeConfigFiles(spec.GetFiles()),
		HorizontalScale:           HorizontalScaleSpecConversion(spec.GetScale().GetHorizontal()),
		CronJobs:                  makeCronJobs(spec.GetCronJobs()),
		Annotations:               maps.Clone(spec.GetAnnotations()),
	}
	config.Replicas = config.GetHorizontalScale().GetMinReplicas()
	var err error
	if config.ContainerSettings, err = ContainerSettingsSpecConversion(spec); err != nil {
		return nil, err
	}

	return config, nil
}

func ContainerSettingsSpecConversion(spec *platformv1.CapsuleSpec) (*capsule.ContainerSettings, error) {
	resources, err := makeResources(spec.GetScale().GetVertical())
	if err != nil {
		return nil, err
	}
	return &capsule.ContainerSettings{
		EnvironmentVariables: maps.Clone(spec.GetEnv().GetRaw()),
		Command:              spec.GetCommand(),
		Args:                 spec.GetArgs(),
		Resources:            resources,
		EnvironmentSources:   makeEnvironmentSources(spec.GetEnv().GetSources()),
	}, nil
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

func makeResources(vertical *platformv1.VerticalScale) (*capsule.Resources, error) {
	res := &capsule.Resources{
		Requests:  &capsule.ResourceList{},
		Limits:    &capsule.ResourceList{},
		GpuLimits: &capsule.GpuLimits{},
	}

	cpuReq, cpuLimit, err := parseLimits(vertical.GetCpu())
	if err != nil {
		return nil, err
	}
	res.Requests.CpuMillis = uint32(cpuReq * 1000)
	res.Limits.CpuMillis = uint32(cpuLimit * 1000)

	memReq, memLim, err := parseLimits(vertical.GetMemory())
	if err != nil {
		return nil, err
	}
	res.Requests.MemoryBytes = uint64(memReq)
	res.Limits.MemoryBytes = uint64(memLim)

	if gpu := vertical.GetGpu().GetRequest(); gpu != "" {
		gpu, err := k8sresource.ParseQuantity(gpu)
		if err != nil {
			return nil, err
		}
		// TODO Type
		res.GpuLimits.Count = uint32(gpu.Value())
	}

	return res, nil
}

func parseLimits(r *platformv1.ResourceLimits) (float64, float64, error) {
	var req, limit float64
	if s := r.GetRequest(); s != "" {
		qq, err := k8sresource.ParseQuantity(s)
		if err != nil {
			return 0, 0, err
		}
		req = qq.AsApproximateFloat64()
	}

	if s := r.GetLimit(); s != "" {
		qq, err := k8sresource.ParseQuantity(s)
		if err != nil {
			return 0, 0, err
		}
		limit = qq.AsApproximateFloat64()
	}

	return req, limit, nil
}

func HorizontalScaleSpecConversion(spec *platformv1.HorizontalScale) *capsule.HorizontalScale {
	res := &capsule.HorizontalScale{
		MaxReplicas: max(spec.GetMax(), spec.GetInstances().GetMax()),
		MinReplicas: max(spec.GetMin(), spec.GetInstances().GetMin()),
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
					ObjectReference: &model.ObjectReference{
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

func makeConfigFiles(configFiles []*platformv1.File) []*capsule.ConfigFile {
	var res []*capsule.ConfigFile

	for _, c := range configFiles {
		res = append(res, ConfigFileSpecConversion(c))
	}

	return res
}

func ConfigFileSpecConversion(c *platformv1.File) *capsule.ConfigFile {
	f := &capsule.ConfigFile{
		Path:     c.GetPath(),
		IsSecret: c.GetAsSecret(),
	}
	switch {
	case c.String_ != "":
		f.Content = []byte(c.GetString_())
	case len(c.Bytes) != 0:
		f.Content = c.GetBytes()
	}

	return f
}

func makeCronJobs(cronJobs []*platformv1.CronJob) []*capsule.CronJob {
	var res []*capsule.CronJob

	for _, j := range cronJobs {
		res = append(res, CronJobSpecConversion(j))
	}

	return res
}

func CronJobSpecConversion(j *platformv1.CronJob) *capsule.CronJob {
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
	}
	return job
}

func makeNetworks(spec []*platformv1.CapsuleInterface) *capsule.Network {
	res := &capsule.Network{}
	for _, i := range spec {
		res.Interfaces = append(res.Interfaces, InterfaceSpecConversion(i))
	}
	return res
}

func InterfaceSpecConversion(i *platformv1.CapsuleInterface) *capsule.Interface {
	ii := &capsule.Interface{
		Port:      uint32(i.GetPort()),
		Name:      i.GetName(),
		Liveness:  makeInterfaceLivenessProbe(i.GetLiveness()),
		Readiness: makeInterfaceReadinessProbe(i.GetReadiness()),
		Routes:    []*capsule.HostRoute{},
	}
	for _, r := range i.GetRoutes() {
		ii.Routes = append(ii.Routes, makeHostRoute(r))
	}
	if len(ii.Routes) == 0 {
		ii.Routes = nil
	}
	return ii
}

func makeInterfaceLivenessProbe(probe *platformv1.InterfaceLivenessProbe) *capsule.InterfaceProbe {
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

func makeInterfaceReadinessProbe(probe *platformv1.InterfaceReadinessProbe) *capsule.InterfaceProbe {
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

func makeHostRoute(route *platformv1.HostRoute) *capsule.HostRoute {
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
		case string(Exact):
			pp.Match = capsule.PathMatchType_PATH_MATCH_TYPE_EXACT
		case string(PathPrefix):
			pp.Match = capsule.PathMatchType_PATH_MATCH_TYPE_PATH_PREFIX
		case string(RegularExpression):
			pp.Match = capsule.PathMatchType_PATH_MATCH_TYPE_REGULAR_EXPRESSION
		}
		res.Paths = append(res.Paths, pp)
	}

	return res
}

func ChangesFromSpecPair(curSpec, newSpec *platformv1.CapsuleSpec) ([]*capsule.Change, error) {
	var res []*capsule.Change

	if curSpec.GetImage() != newSpec.GetImage() {
		res = append(res, &capsule.Change{
			Field: &capsule.Change_ImageId{
				ImageId: newSpec.GetImage(),
			},
		})
	}

	d := curSpec.GetEnv().GetRaw()
	for k, v := range newSpec.GetEnv().GetRaw() {
		if vv, ok := d[k]; !ok || v != vv {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_SetEnvironmentVariable{
					SetEnvironmentVariable: &capsule.Change_KeyValue{
						Name:  k,
						Value: v,
					},
				},
			})
		}
	}

	d = newSpec.GetEnv().GetRaw()
	for k, v := range curSpec.GetEnv().GetRaw() {
		if _, ok := d[k]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_RemoveEnvironmentVariable{
					RemoveEnvironmentVariable: v,
				},
			})
		}
	}

	s := map[source]struct{}{}
	for _, ss := range newSpec.GetEnv().GetSources() {
		s[news(ss)] = struct{}{}
	}
	for _, ss := range curSpec.GetEnv().GetSources() {
		if _, ok := s[news(ss)]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_RemoveEnvironmentSource{
					RemoveEnvironmentSource: EnvironmentSourceSpecConversion(ss),
				},
			})
		}
	}

	clear(s)
	for _, ss := range curSpec.GetEnv().GetSources() {
		s[news(ss)] = struct{}{}
	}
	for _, ss := range newSpec.GetEnv().GetSources() {
		if _, ok := s[news(ss)]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_SetEnvironmentSource{
					SetEnvironmentSource: EnvironmentSourceSpecConversion(ss),
				},
			})
		}
	}

	a := curSpec.GetAnnotations()
	for k, v := range newSpec.GetAnnotations() {
		if vv, ok := a[k]; !ok || v != vv {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_SetAnnotation{
					SetAnnotation: &capsule.Change_KeyValue{
						Name:  k,
						Value: vv,
					},
				},
			})
		}
	}

	a = newSpec.GetAnnotations()
	for k := range curSpec.GetAnnotations() {
		if _, ok := a[k]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_RemoveAnnotation{
					RemoveAnnotation: k,
				},
			})
		}
	}

	ints := map[string]*platformv1.CapsuleInterface{}
	for _, i := range curSpec.GetInterfaces() {
		ints[i.GetName()] = i
	}
	for _, i := range newSpec.GetInterfaces() {
		if ii, ok := ints[i.GetName()]; !ok || !proto.Equal(i, ii) {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_SetInterface{
					SetInterface: InterfaceSpecConversion(i),
				},
			},
			)
		}
	}

	clear(ints)
	for _, i := range newSpec.GetInterfaces() {
		ints[i.GetName()] = i
	}
	for _, i := range curSpec.GetInterfaces() {
		if _, ok := ints[i.GetName()]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_RemoveInterface{
					RemoveInterface: i.GetName(),
				},
			})
		}
	}

	if newSpec.GetCommand() != curSpec.GetCommand() || !slices.Equal(newSpec.GetArgs(), curSpec.GetArgs()) {
		res = append(res, &capsule.Change{
			Field: &capsule.Change_CommandArguments_{
				CommandArguments: &capsule.Change_CommandArguments{
					Command: curSpec.GetCommand(),
					Args:    curSpec.GetArgs(),
				},
			},
		})
	}

	c := map[string]*platformv1.CronJob{}
	for _, cc := range curSpec.GetCronJobs() {
		c[cc.GetName()] = cc
	}
	for _, c1 := range newSpec.GetCronJobs() {
		if c2, ok := c[c1.GetName()]; !ok || !proto.Equal(c1, c2) {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_AddCronJob{
					AddCronJob: CronJobSpecConversion(c1),
				},
			})
		}
	}

	clear(c)
	for _, cc := range newSpec.GetCronJobs() {
		c[cc.GetName()] = cc
	}
	for _, cc := range curSpec.GetCronJobs() {
		if _, ok := c[cc.GetName()]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_RemoveCronJob_{
					RemoveCronJob: &capsule.Change_RemoveCronJob{
						JobName: cc.GetName(),
					},
				},
			})
		}
	}

	f := map[string]*platformv1.File{}
	for _, ff := range curSpec.GetFiles() {
		f[ff.GetPath()] = ff
	}
	for _, f1 := range newSpec.GetFiles() {
		if f2, ok := f[f1.GetPath()]; !ok || !proto.Equal(f1, f2) {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_SetConfigFile{
					SetConfigFile: &capsule.Change_ConfigFile{
						Path:     f1.GetPath(),
						Content:  f1.GetBytes(),
						IsSecret: f1.GetAsSecret(),
					},
				},
			})
		}
	}
	clear(f)
	for _, ff := range newSpec.GetFiles() {
		f[ff.GetPath()] = ff
	}
	for _, f1 := range curSpec.GetFiles() {
		if _, ok := f[f1.GetPath()]; !ok {
			res = append(res, &capsule.Change{
				Field: &capsule.Change_RemoveConfigFile{
					RemoveConfigFile: f1.GetPath(),
				},
			})
		}
	}

	if !proto.Equal(curSpec.GetScale().GetHorizontal(), newSpec.GetScale().GetHorizontal()) {
		res = append(res, &capsule.Change{
			Field: &capsule.Change_HorizontalScale{
				HorizontalScale: HorizontalScaleSpecConversion(newSpec.GetScale().GetHorizontal()),
			},
		})
	}

	if curSpec.GetAutoAddRigServiceAccounts() != newSpec.GetAutoAddRigServiceAccounts() {
		res = append(res, &capsule.Change{
			Field: &capsule.Change_AutoAddRigServiceAccounts{
				AutoAddRigServiceAccounts: newSpec.GetAutoAddRigServiceAccounts(),
			},
		})
	}

	if !proto.Equal(curSpec.GetScale().GetVertical(), newSpec.GetScale().GetVertical()) {
		containerSettings, err := ContainerSettingsSpecConversion(newSpec)
		if err != nil {
			return nil, err
		}
		res = append(res, &capsule.Change{
			Field: &capsule.Change_ContainerSettings{
				ContainerSettings: containerSettings,
			},
		})
	}

	return res, nil
}

type source struct {
	kind string
	name string
}

func news(s *platformv1.EnvironmentSource) source {
	return source{
		kind: s.GetKind(),
		name: s.GetName(),
	}
}

const (
	// Characters in 0..31, not in "\x07\x08\x09\x10\x12\x13\x27"
	_invalidChars = "\x00\x01\x02\x03\x04\x05\x06\x11\x14\x15\x16\x17\x18\x19\x20\x21\x22\x23\x24\x25\x26\x28\x29\x30\x31"
)

func ValidString(bs []byte) bool {
	if !utf8.Valid(bs) {
		return false
	}

	if bytes.ContainsAny(bs, _invalidChars) {
		return false
	}

	return true
}

// nolint:lll
// MergeProjectEnv merges a ProjEnvCapsuleBase into a CapsuleSpec and returns a new object with the merged result
// It uses StrategicMergePatch (https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/)
func MergeProjectEnv(patch *platformv1.ProjEnvCapsuleBase, into *platformv1.CapsuleSpec) (*platformv1.CapsuleSpec, error) {
	return mergeCapsuleSpec(patch, into)
}

// nolint:lll
// MergeCapsuleSpec merges a CapsuleSpec into another CapsuleSpec and returns a new object with the merged result
// It uses StrategicMergePatch (https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/)
func MergeCapsuleSpecs(patch, into *platformv1.CapsuleSpec) (*platformv1.CapsuleSpec, error) {
	return mergeCapsuleSpec(patch, into)
}

func mergeCapsuleSpec(patch any, into *platformv1.CapsuleSpec) (*platformv1.CapsuleSpec, error) {
	// It would be possible to do much faster merging by manualling overwriting protobuf fields.
	// This is tedius to maintain so until it becomes an issue, we use json marshalling to leverage StrategicMergePatch
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	intoBytes, err := json.Marshal(into)
	if err != nil {
		return nil, err
	}

	outBytes, err := strategicpatch.StrategicMergePatch(intoBytes, patchBytes, &CapsuleSpec{})
	if err != nil {
		return nil, err
	}

	out := &platformv1.CapsuleSpec{}
	if err := json.Unmarshal(outBytes, out); err != nil {
		return nil, err
	}

	return out, nil
}

func CapsuleProtoToK8s(spec *platformv1.Capsule, scheme *runtime.Scheme) (*Capsule, error) {
	bs, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return obj.DecodeIntoT(bs, &Capsule{}, scheme)
}
