---
custom_edit_url: null
---


# rig.dev/v1alpha2

Package v1alpha2 contains API Schema definitions for the v1alpha2 API group

## Resource Types
- [Capsule](#capsule)



### CPUTarget



CPUTarget defines an autoscaler target for the CPU metric If empty, no autoscaling will be done

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `utilization` _integer_ | Utilization specifies the average CPU target. If the average exceeds this number new instances will be added. |


### Capsule



Capsule is the Schema for the capsules API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.dev/v1alpha2`
| `kind` _string_ | `Capsule`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[CapsuleSpec](#capsulespec)_ | Spec holds the specification of the Capsule. |


### CapsuleInterface



CapsuleInterface defines an interface for a capsule

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `name` _string_ | Name specifies a descriptive name of the interface. |
| `port` _integer_ | Port specifies what port the interface should have. |
| `liveness` _[InterfaceProbe](#interfaceprobe)_ | Liveness specifies that this interface should be used for liveness probing. Only one of the Capsule interfaces can be used as liveness probe. |
| `readiness` _[InterfaceProbe](#interfaceprobe)_ | Readiness specifies that this interface should be used for readiness probing. Only one of the Capsule interfaces can be used as readiness probe. |
| `public` _[CapsulePublicInterface](#capsulepublicinterface)_ | Public specifies if and how the interface should be published. |
| `routes` _[HostRoute](#hostroute) array_ | Host routes that are mapped to this interface. |


### CapsuleInterfaceIngress

_Underlying type:_ _[struct{Host string "json:\"host\""; Paths []string "json:\"paths,omitempty\""}](#struct{host-string-"json:\"host\"";-paths-[]string-"json:\"paths,omitempty\""})_

CapsuleInterfaceIngress defines that the interface should be exposed as http ingress

_Appears in:_
- [CapsulePublicInterface](#capsulepublicinterface)



### CapsuleInterfaceLoadBalancer

_Underlying type:_ _[struct{Port int32 "json:\"port\""}](#struct{port-int32-"json:\"port\""})_

CapsuleInterfaceLoadBalancer defines that the interface should be exposed as a L4 loadbalancer

_Appears in:_
- [CapsulePublicInterface](#capsulepublicinterface)



### CapsulePublicInterface



CapsulePublicInterface defines how to publicly expose the interface

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `ingress` _[CapsuleInterfaceIngress](#capsuleinterfaceingress)_ | Ingress specifies that this interface should be exposed through an Ingress resource. The Ingress field is mutually exclusive with the LoadBalancer field. |
| `loadBalancer` _[CapsuleInterfaceLoadBalancer](#capsuleinterfaceloadbalancer)_ | LoadBalancer specifies that this interface should be exposed through a LoadBalancer Service. The LoadBalancer field is mutually exclusive with the Ingress field. |


### CapsuleScale



CapsuleScale specifies the horizontal and vertical scaling of the Capsule.

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `horizontal` _[HorizontalScale](#horizontalscale)_ | Horizontal specifies the horizontal scaling of the Capsule. |
| `vertical` _[VerticalScale](#verticalscale)_ | Vertical specifies the vertical scaling of the Capsule. |


### CapsuleSpec



CapsuleSpec defines the desired state of Capsule

_Appears in:_
- [Capsule](#capsule)

| Field | Description |
| --- | --- |
| `image` _string_ | Image specifies what image the Capsule should run. |
| `command` _string_ | Command is run as a command in the shell. If left unspecified, the container will run using what is specified as ENTRYPOINT in the Dockerfile. |
| `args` _string array_ | Args is a list of arguments either passed to the Command or if Command is left empty the arguments will be passed to the ENTRYPOINT of the docker image. |
| `interfaces` _[CapsuleInterface](#capsuleinterface) array_ | Interfaces specifies the list of interfaces the the container should have. Specifying interfaces will create the corresponding kubernetes Services and Ingresses depending on how the interface is configured. |
| `files` _[File](#file) array_ | Files is a list of files to mount in the container. These can either be based on ConfigMaps or Secrets. |
| `scale` _[CapsuleScale](#capsulescale)_ | Scale specifies the scaling of the Capsule. |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector for what nodes the Capsule should live on. |
| `env` _[Env](#env)_ | Env specifies configuration for how the container should obtain environment variables. |
| `cronJobs` _[CronJob](#cronjob) array_ |  |


### CronJob





_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `name` _string_ |  |
| `schedule` _string_ |  |
| `url` _[URL](#url)_ |  |
| `command` _[JobCommand](#jobcommand)_ |  |
| `maxRetries` _integer_ | Defaults to 6 |
| `timeoutSeconds` _integer_ |  |


### CustomMetric



CustomMetric defines a custom metrics emitted by the custom.metrics.k8s.io API which the autoscaler should scale on Exactly one of InstanceMetric and ObjectMetric must be provided

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `instanceMetric` _[InstanceMetric](#instancemetric)_ | InstanceMetric defines a custom instance-based metric (pod-metric in Kubernetes lingo) |
| `objectMetric` _[ObjectMetric](#objectmetric)_ | ObjectMetric defines a custom object-based metric |


### Env



Env defines what secrets and configmaps should be used for environment variables in the capsule.

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `disable_automatic` _boolean_ | DisableAutomatic sets wether the capsule should disable automatically use of existing secrets and configmaps which share the same name as the capsule as environment variables. |
| `from` _[EnvReference](#envreference) array_ | From holds a list of references to secrets and configmaps which should be mounted as environment variables. |


### EnvReference



EnvSource holds a reference to either a ConfigMap or a Secret

_Appears in:_
- [Env](#env)

| Field | Description |
| --- | --- |
| `kind` _string_ | Kind is the resource kind of the env reference, must be ConfigMap or Secret. |
| `name` _string_ | Name is the name of a ConfigMap or Secret in the same namespace as the Capsule. |


### File



File defines a mounted file and where to retrieve the contents from

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `ref` _[FileContentReference](#filecontentreference)_ | Ref specifies a reference to a ConfigMap or Secret key which holds the contents of the file. |
| `path` _string_ | Path specifies the full path where the File should be mounted including the file name. |


### FileContentReference



FileContentRef defines the name of a config resource and the key from which to retrieve the contents

_Appears in:_
- [File](#file)

| Field | Description |
| --- | --- |
| `kind` _string_ | Kind of reference. Can be either ConfigMap or Secret. |
| `name` _string_ | Name of reference. |
| `key` _string_ | Key in reference which holds file contents. |


### HTTPPathRoute

_Underlying type:_ _[struct{Path string "json:\"path\""; Match PathMatchType "json:\"match,omitempty\""}](#struct{path-string-"json:\"path\"";-match-pathmatchtype-"json:\"match,omitempty\""})_

A HTTP path routing.

_Appears in:_
- [HostRoute](#hostroute)



### HorizontalScale



HorizontalScale defines the policy for the number of replicas of the capsule It can both be configured with autoscaling and with a static number of replicas

_Appears in:_
- [CapsuleScale](#capsulescale)

| Field | Description |
| --- | --- |
| `instances` _[Instances](#instances)_ | Instances specifies minimum and maximum amount of Capsule instances. |
| `cpuTarget` _[CPUTarget](#cputarget)_ | CPUTarget specifies that this Capsule should be scaled using CPU utilization. |
| `customMetrics` _[CustomMetric](#custommetric) array_ | CustomMetrics specifies custom metrics emitted by the custom.metrics.k8s.io API which the autoscaler should scale on |


### HostRoute



HostRoute is the configuration of a route to the network interface it's configured on.

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `name` _string_ | Name of the route. This field is required and cannot be empty, and must be unique for the interface. If this field is changed, it may result in downtime, as it is used to generate resources. |
| `host` _string_ | Host of the route. This field is required and cannot be empty. |
| `paths` _[HTTPPathRoute](#httppathroute) array_ | HTTP paths of the host that maps to the interface. If empty, all paths are automatically matched. |
| `annotations` _object (keys:string, values:string)_ | Annotations of the route option. This can be plugin-specific configuration that allows custom plugins to add non-standard behavior. |


### InstanceMetric

_Underlying type:_ _[struct{MetricName string "json:\"metricName\""; MatchLabels map[string]string "json:\"matchLabels,omitempty\""; AverageValue string "json:\"averageValue\""}](#struct{metricname-string-"json:\"metricname\"";-matchlabels-map[string]string-"json:\"matchlabels,omitempty\"";-averagevalue-string-"json:\"averagevalue\""})_

InstanceMetric defines a custom instance-based metric (pod-metric in Kubernetes lingo)

_Appears in:_
- [CustomMetric](#custommetric)



### Instances



Instances specifies the minimum and maximum amount of capsule instances.

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `min` _integer_ | Min specifies the minimum amount of instances to run. |
| `max` _integer_ | Max specifies the maximum amount of instances to run. Omit to disable autoscaling. |


### InterfaceGRPCProbe

_Underlying type:_ _[struct{Service string "json:\"service\""}](#struct{service-string-"json:\"service\""})_

InterfaceGRPCProbe specifies a GRPC probe.

_Appears in:_
- [InterfaceProbe](#interfaceprobe)



### InterfaceProbe



InterfaceProbe specifies an interface probe

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `path` _string_ | Path is the HTTP path of the probe. Path is mutually exclusive with the TCP and GCRP fields. |
| `tcp` _boolean_ | TCP specifies that this is a simple TCP listen probe. |
| `grpc` _[InterfaceGRPCProbe](#interfacegrpcprobe)_ | GRPC specifies that this is a GRCP probe. |


### JobCommand





_Appears in:_
- [CronJob](#cronjob)

| Field | Description |
| --- | --- |
| `command` _string_ |  |
| `args` _string array_ |  |


### ObjectMetric

_Underlying type:_ _[struct{MetricName string "json:\"metricName\""; MatchLabels map[string]string "json:\"matchLabels,omitempty\""; AverageValue string "json:\"averageValue,omitempty\""; Value string "json:\"value,omitempty\""; DescribedObject k8s.io/api/autoscaling/v2.CrossVersionObjectReference "json:\"objectReference\""}](#struct{metricname-string-"json:\"metricname\"";-matchlabels-map[string]string-"json:\"matchlabels,omitempty\"";-averagevalue-string-"json:\"averagevalue,omitempty\"";-value-string-"json:\"value,omitempty\"";-describedobject-k8sioapiautoscalingv2crossversionobjectreference-"json:\"objectreference\""})_

ObjectMetric defines a custom object metric for the autoscaler

_Appears in:_
- [CustomMetric](#custommetric)





### ResourceLimits



ResourceLimits specifies the request and limit of a resource.

_Appears in:_
- [VerticalScale](#verticalscale)

| Field | Description |
| --- | --- |
| `request` _[Quantity](#quantity)_ | Request specifies the resource request. |
| `limit` _[Quantity](#quantity)_ | Limit specifies the resource limit. |


### ResourceRequest



ResourceRequest specifies the request of a resource.

_Appears in:_
- [VerticalScale](#verticalscale)

| Field | Description |
| --- | --- |
| `request` _[Quantity](#quantity)_ | Request specifies the request of a resource. |


### RouteOptions



Route options.

_Appears in:_
- [HostRoute](#hostroute)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations of the route option. This can be plugin-specific configuration that allows custom plugins to add non-standard behavior. |


### URL





_Appears in:_
- [CronJob](#cronjob)

| Field | Description |
| --- | --- |
| `port` _integer_ |  |
| `path` _string_ |  |
| `queryParameters` _object (keys:string, values:string)_ |  |


### UsedResource





_Appears in:_
- [CapsuleStatus](#capsulestatus)

| Field | Description |
| --- | --- |
| `ref` _[TypedLocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#typedlocalobjectreference-v1-core)_ |  |
| `state` _string_ |  |
| `message` _string_ |  |


### VerticalScale



VerticalScale specifies the vertical scaling of the Capsule.

_Appears in:_
- [CapsuleScale](#capsulescale)

| Field | Description |
| --- | --- |
| `cpu` _[ResourceLimits](#resourcelimits)_ | CPU specifies the CPU resource request and limit |
| `memory` _[ResourceLimits](#resourcelimits)_ | Memory specifies the Memory resource request and limit |
| `gpu` _[ResourceRequest](#resourcerequest)_ | GPU specifies the GPU resource request and limit |




<hr class="solid" />


:::info generated from source code
This page is generated based on go source code. If you have suggestions for
improvements for this page, please open an issue at
[github.com/rigdev/rig](https://github.com/rigdev/rig/issues/new), or a pull
request with changes to [the go source
files](https://github.com/rigdev/rig/tree/main/pkg/api).
:::