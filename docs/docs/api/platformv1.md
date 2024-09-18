---
custom_edit_url: null
---


# rig.platform/v1



## Resource Types
- [Capsule](#capsule)
- [CapsuleSet](#capsuleset)
- [CapsuleSpec](#capsulespec)
- [Environment](#environment)
- [HostCapsule](#hostcapsule)
- [Project](#project)



### CPUTarget



CPUTarget defines an autoscaler target for the CPU metric
If empty, no autoscaling will be done

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `utilization` _integer_ | Utilization specifies the average CPU target. If the average<br /><br />exceeds this number new instances will be added. |


### Capsule







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `Capsule`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br /><br />Servers may infer this from the endpoint the client submits requests to.<br /><br />Cannot be updated.<br /><br />In CamelCase.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br /><br />Servers should convert recognized schemas to the latest internal value, and<br /><br />may reject unrecognized values.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name,Project,Environment is unique<br /><br />Project,Name referes to an existing Capsule type with the given name and project<br /><br />Will throw an error (in the platform) if the Capsule does not exist |
| `project` _string_ | Project references an existing Project type with the given name<br /><br />Will throw an error (in the platform) if the Project does not exist |
| `environment` _string_ | Environment references an existing Environment type with the given name<br /><br />Will throw an error (in the platform) if the Environment does not exist<br /><br />The environment also needs to be present in the parent Capsule |
| `spec` _[CapsuleSpec](#capsulespec)_ |  |


### CapsuleInterface



CapsuleInterface defines an interface for a capsule

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `name` _string_ | Name specifies a descriptive name of the interface. |
| `port` _integer_ | Port specifies what port the interface should have. |
| `liveness` _[InterfaceLivenessProbe](#interfacelivenessprobe)_ | Liveness specifies that this interface should be used for<br /><br />liveness probing. Only one of the Capsule interfaces can be<br /><br />used as liveness probe. |
| `readiness` _[InterfaceReadinessProbe](#interfacereadinessprobe)_ | Readiness specifies that this interface should be used for<br /><br />readiness probing. Only one of the Capsule interfaces can be<br /><br />used as readiness probe. |
| `routes` _[HostRoute](#hostroute) array_ | Host routes that are mapped to this interface. |


### CapsuleSet







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `CapsuleSet`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br /><br />Servers may infer this from the endpoint the client submits requests to.<br /><br />Cannot be updated.<br /><br />In CamelCase.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br /><br />Servers should convert recognized schemas to the latest internal value, and<br /><br />may reject unrecognized values.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name,Project is unique |
| `project` _string_ | Project references an existing Project type with the given name<br /><br />Will throw an error (in the platform) if the Project does not exist |
| `spec` _[CapsuleSpec](#capsulespec)_ | Capsule-level defaults |
| `environments` _object (keys:string, values:[CapsuleSpec](#capsulespec))_ |  |
| `environmentRefs` _string array_ |  |


### CapsuleSpec





_Appears in:_
- [Capsule](#capsule)
- [CapsuleSet](#capsuleset)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `CapsuleSpec`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br /><br />Servers may infer this from the endpoint the client submits requests to.<br /><br />Cannot be updated.<br /><br />In CamelCase.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br /><br />Servers should convert recognized schemas to the latest internal value, and<br /><br />may reject unrecognized values.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `annotations` _object (keys:string, values:string)_ |  |
| `image` _string_ | Image specifies what image the Capsule should run. |
| `command` _string_ | Command is run as a command in the shell. If left unspecified, the<br /><br />container will run using what is specified as ENTRYPOINT in the<br /><br />Dockerfile. |
| `args` _string array_ | Args is a list of arguments either passed to the Command or if Command<br /><br />is left empty the arguments will be passed to the ENTRYPOINT of the<br /><br />docker image. |
| `interfaces` _[CapsuleInterface](#capsuleinterface) array_ | Interfaces specifies the list of interfaces the the container should<br /><br />have. Specifying interfaces will create the corresponding kubernetes<br /><br />Services and Ingresses depending on how the interface is configured.<br /><br />nolint:lll |
| `files` _[File](#file) array_ | Files is a list of files to mount in the container. These can either be<br /><br />based on ConfigMaps or Secrets. |
| `env` _[EnvironmentVariables](#environmentvariables)_ | Env defines the environment variables set in the Capsule |
| `scale` _[Scale](#scale)_ | Scale specifies the scaling of the Capsule. |
| `cronJobs` _[CronJob](#cronjob) array_ |  |
| `autoAddRigServiceAccounts` _boolean_ |  |
| `extensions` _object (keys:string, values:RawMessage)_ | Extensions are extra, typed fields defined by the platform for custom behaviour implemented through plugins |


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



CustomMetric defines a custom metrics emitted by the custom.metrics.k8s.io API
which the autoscaler should scale on
Exactly one of InstanceMetric and ObjectMetric must be provided

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `instanceMetric` _[InstanceMetric](#instancemetric)_ | InstanceMetric defines a custom instance-based metric (pod-metric in Kubernetes lingo) |
| `objectMetric` _[ObjectMetric](#objectmetric)_ | ObjectMetric defines a custom object-based metric |


### Environment







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `Environment`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br /><br />Servers may infer this from the endpoint the client submits requests to.<br /><br />Cannot be updated.<br /><br />In CamelCase.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br /><br />Servers should convert recognized schemas to the latest internal value, and<br /><br />may reject unrecognized values.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name is unique |
| `namespaceTemplate` _string_ |  |
| `operatorVersion` _string_ |  |
| `cluster` _string_ |  |
| `spec` _[ProjEnvCapsuleBase](#projenvcapsulebase)_ | Environment level defaults |
| `ephemeral` _boolean_ |  |
| `activeProjects` _string array_ |  |
| `global` _boolean_ |  |


### EnvironmentSource





_Appears in:_
- [EnvironmentVariables](#environmentvariables)

| Field | Description |
| --- | --- |
| `name` _string_ | Name is the name of the kubernetes object containing the environment source. |
| `kind` _[EnvironmentSourceKind](#environmentsourcekind)_ | Kind is the kind of source, either ConfigMap or Secret. |


### EnvironmentSourceKind

_Underlying type:_ _string_



_Appears in:_
- [EnvironmentSource](#environmentsource)



### EnvironmentVariables



EnvironmentVariables defines the environment variables injected into a Capsule.

_Appears in:_
- [CapsuleSpec](#capsulespec)
- [ProjEnvCapsuleBase](#projenvcapsulebase)

| Field | Description |
| --- | --- |
| `raw` _object (keys:string, values:string)_ | Raw is a list of environment variables as key-value pairs. |
| `sources` _[EnvironmentSource](#environmentsource) array_ | Sources is a list of source files which will be injected as environment variables.<br /><br />They can be references to either ConfigMaps or Secrets. |


### File





_Appears in:_
- [CapsuleSpec](#capsulespec)
- [ProjEnvCapsuleBase](#projenvcapsulebase)

| Field | Description |
| --- | --- |
| `path` _string_ |  |
| `asSecret` _boolean_ |  |
| `bytes` _integer_ |  |
| `string` _string_ |  |
| `ref` _[FileReference](#filereference)_ |  |


### FileReference



FileReference defines the name of a k8s config resource and the key from which
to retrieve the contents

_Appears in:_
- [File](#file)

| Field | Description |
| --- | --- |
| `kind` _string_ | Kind of reference. Can be either ConfigMap or Secret. |
| `name` _string_ | Name of reference. |
| `key` _string_ | Key in reference which holds file contents. |




### HorizontalScale



HorizontalScale defines the policy for the number of replicas of
the capsule It can both be configured with autoscaling and with a
static number of replicas

_Appears in:_
- [Scale](#scale)

| Field | Description |
| --- | --- |
| `min` _integer_ | Min specifies the minimum amount of instances to run. |
| `max` _integer_ | Max specifies the maximum amount of instances to run. Omit to<br /><br />disable autoscaling. |
| `instances` _[Instances](#instances)_ | Instances specifies minimum and maximum amount of Capsule<br /><br />instances.<br /><br />Deprecated; use `min` and `max` instead. |
| `cpuTarget` _[CPUTarget](#cputarget)_ | CPUTarget specifies that this Capsule should be scaled using CPU<br /><br />utilization. |
| `customMetrics` _[CustomMetric](#custommetric) array_ | CustomMetrics specifies custom metrics emitted by the custom.metrics.k8s.io API<br /><br />which the autoscaler should scale on |


### HostCapsule







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `HostCapsule`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br /><br />Servers may infer this from the endpoint the client submits requests to.<br /><br />Cannot be updated.<br /><br />In CamelCase.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br /><br />Servers should convert recognized schemas to the latest internal value, and<br /><br />may reject unrecognized values.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name,Project,Environment is unique<br /><br />Project,Name referes to an existing Capsule type with the given name and project<br /><br />Will throw an error (in the platform) if the Capsule does not exist |
| `project` _string_ | Project references an existing Project type with the given name<br /><br />Will throw an error (in the platform) if the Project does not exist |
| `environment` _string_ | Environment references an existing Environment type with the given name<br /><br />Will throw an error (in the platform) if the Environment does not exist<br /><br />The environment also needs to be present in the parent Capsule |
| `network` _[HostNetwork](#hostnetwork)_ | Network mapping between the host network and the Kubernetes cluster network. When activated,<br /><br />traffic between the two networks will be tunneled according to the rules specified here. |


### HostNetwork





_Appears in:_
- [HostCapsule](#hostcapsule)

| Field | Description |
| --- | --- |
| `hostInterfaces` _[ProxyInterface](#proxyinterface) array_ | HostInterfaces are interfaces activated on the local machine (the host) and forwarded<br /><br />to the Kubernetes cluster capsules. |
| `capsuleInterfaces` _[ProxyInterface](#proxyinterface) array_ | CapsuleInterfaces are interfaces activated on the Capsule within the Kubernetes cluster<br /><br />and forwarded to the local machine (the host). The traffic is directed to a single target,<br /><br />e.g. `localhost:8080`. |
| `tunnelPort` _integer_ | TunnelPort for which the proxy-capsule should listen on. This is automatically set by the tooling. |


### HostRoute



HostRoute is the configuration of a route to the network interface
it's configured on.

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `id` _string_ | ID of the route. This field is required and cannot be empty, and must be unique for the interface.<br /><br />If this field is changed, it may result in downtime, as it is used to generate resources. |
| `host` _string_ | Host of the route. This field is required and cannot be empty. |
| `paths` _[HTTPPathRoute](#httppathroute) array_ | HTTP paths of the host that maps to the interface. If empty, all paths are<br /><br />automatically matched. |




### Instances



Instances specifies the minimum and maximum amount of capsule
instances.

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `min` _integer_ | Min specifies the minimum amount of instances to run. |
| `max` _integer_ | Max specifies the maximum amount of instances to run. Omit to<br /><br />disable autoscaling. |




### InterfaceLivenessProbe



InterfaceLivenessProbe specifies an interface probe for liveness checks.

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `path` _string_ | Path is the HTTP path of the probe. Path is mutually<br /><br />exclusive with the TCP and GCRP fields. |
| `tcp` _boolean_ | TCP specifies that this is a simple TCP listen probe. |
| `grpc` _[InterfaceGRPCProbe](#interfacegrpcprobe)_ | GRPC specifies that this is a GRCP probe. |
| `startupDelay` _integer_ | For slow-starting containers, the startup delay allows liveness<br /><br />checks to fail for a set duration before restarting the instance. |


### InterfaceOptions





_Appears in:_
- [ProxyInterface](#proxyinterface)

| Field | Description |
| --- | --- |
| `tcp` _boolean_ | TCP enables layer-4 proxying in favor of layer-7 HTTP proxying. |
| `allowOrigin` _string_ | AllowOrigin sets the `Access-Control-Allow-Origin` Header on responses to<br /><br />the provided value, allowing local by-pass of CORS rules.<br /><br />Ignored if TCP is enabled. |
| `changeOrigin` _boolean_ | ChangeOrigin changes the Host header to match the given target. If not set,<br /><br />the Host header will be that of the original request.<br /><br />This does not impact the Origin header - use `Headers` to set that.<br /><br />Ignored if TCP is enabled. |
| `headers` _object (keys:string, values:string)_ | Headers to set on the proxy-requests.<br /><br />Ignored if TCP is enabled. |


### InterfaceReadinessProbe



InterfaceReadinessProbe specifies an interface probe for readiness checks.

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `path` _string_ | Path is the HTTP path of the probe. Path is mutually<br /><br />exclusive with the TCP and GCRP fields. |
| `tcp` _boolean_ | TCP specifies that this is a simple TCP listen probe. |
| `grpc` _[InterfaceGRPCProbe](#interfacegrpcprobe)_ | GRPC specifies that this is a GRCP probe. |


### JobCommand





_Appears in:_
- [CronJob](#cronjob)

| Field | Description |
| --- | --- |
| `command` _string_ |  |
| `args` _string array_ |  |




### PathMatchType

_Underlying type:_ _string_

PathMatchType specifies the semantics of how HTTP paths should be compared.

_Appears in:_
- [HTTPPathRoute](#httppathroute)



### ProjEnvCapsuleBase





_Appears in:_
- [Environment](#environment)
- [Project](#project)

| Field | Description |
| --- | --- |
| `files` _[File](#file) array_ |  |
| `env` _[EnvironmentVariables](#environmentvariables)_ |  |


### Project







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `Project`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br /><br />Servers may infer this from the endpoint the client submits requests to.<br /><br />Cannot be updated.<br /><br />In CamelCase.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br /><br />Servers should convert recognized schemas to the latest internal value, and<br /><br />may reject unrecognized values.<br /><br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name is unique |
| `spec` _[ProjEnvCapsuleBase](#projenvcapsulebase)_ | Project level defaults |


### ProxyInterface





_Appears in:_
- [HostNetwork](#hostnetwork)

| Field | Description |
| --- | --- |
| `port` _integer_ | Port to accept traffic from. |
| `target` _string_ | Target is the address:port to forward traffic to. |
| `options` _[InterfaceOptions](#interfaceoptions)_ | Options to further configure the proxying aspects of the interface. |


### ResourceLimits

_Underlying type:_ _[struct{Request *k8s.io/apimachinery/pkg/api/resource.Quantity "json:\"request,omitempty\" protobuf:\"1\""; Limit *k8s.io/apimachinery/pkg/api/resource.Quantity "json:\"limit,omitempty\" protobuf:\"2\""}](#struct{request-*k8sioapimachinerypkgapiresourcequantity-"json:\"request,omitempty\"-protobuf:\"1\"";-limit-*k8sioapimachinerypkgapiresourcequantity-"json:\"limit,omitempty\"-protobuf:\"2\""})_

ResourceLimits specifies the request and limit of a resource.

_Appears in:_
- [VerticalScale](#verticalscale)



### ResourceRequest

_Underlying type:_ _[struct{Request k8s.io/apimachinery/pkg/api/resource.Quantity "json:\"request,omitempty\" protobuf:\"1\""}](#struct{request-k8sioapimachinerypkgapiresourcequantity-"json:\"request,omitempty\"-protobuf:\"1\""})_

ResourceRequest specifies the request of a resource.

_Appears in:_
- [VerticalScale](#verticalscale)



### RouteOptions

_Underlying type:_ _[struct{Annotations map[string]string "json:\"annotations,omitempty\" protobuf:\"4\""}](#struct{annotations-map[string]string-"json:\"annotations,omitempty\"-protobuf:\"4\""})_

Route options.

_Appears in:_
- [HostRoute](#hostroute)



### Scale





_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `horizontal` _[HorizontalScale](#horizontalscale)_ | Horizontal specifies the horizontal scaling of the Capsule. |
| `vertical` _[VerticalScale](#verticalscale)_ | Vertical specifies the vertical scaling of the Capsule. |


### URL





_Appears in:_
- [CronJob](#cronjob)

| Field | Description |
| --- | --- |
| `port` _integer_ |  |
| `path` _string_ |  |
| `queryParameters` _object (keys:string, values:string)_ |  |


### VerticalScale



VerticalScale specifies the vertical scaling of the Capsule.

_Appears in:_
- [Scale](#scale)

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