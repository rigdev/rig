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



### Capsule







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `Capsule`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name,Project,Environment is unique Project,Name referes to an existing Capsule type with the given name and project Will throw an error (in the platform) if the Capsule does not exist |
| `project` _string_ | Project references an existing Project type with the given name Will throw an error (in the platform) if the Project does not exist |
| `environment` _string_ | Environment references an existing Environment type with the given name Will throw an error (in the platform) if the Environment does not exist The environment also needs to be present in the parent Capsule |
| `spec` _[CapsuleSpec](#capsulespec)_ |  |


### CapsuleSet







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `CapsuleSet`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name,Project is unique |
| `project` _string_ | Project references an existing Project type with the given name Will throw an error (in the platform) if the Project does not exist |
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
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `annotations` _object (keys:string, values:string)_ |  |
| `image` _string_ | Image specifies what image the Capsule should run. |
| `command` _string_ | Command is run as a command in the shell. If left unspecified, the container will run using what is specified as ENTRYPOINT in the Dockerfile. |
| `args` _string array_ | Args is a list of arguments either passed to the Command or if Command is left empty the arguments will be passed to the ENTRYPOINT of the docker image. |
| `interfaces` _CapsuleInterface array_ | Interfaces specifies the list of interfaces the the container should have. Specifying interfaces will create the corresponding kubernetes Services and Ingresses depending on how the interface is configured. nolint:lll |
| `files` _[File](#file) array_ | Files is a list of files to mount in the container. These can either be based on ConfigMaps or Secrets. |
| `env` _[EnvironmentVariables](#environmentvariables)_ | Env defines the environment variables set in the Capsule |
| `scale` _[Scale](#scale)_ | Scale specifies the scaling of the Capsule. |
| `cronJobs` _CronJob array_ |  |
| `autoAddRigServiceAccounts` _boolean_ | TODO Move to plugin |


### Environment







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `Environment`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
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
| `sources` _[EnvironmentSource](#environmentsource) array_ | Sources is a list of source files which will be injected as environment variables. They can be references to either ConfigMaps or Secrets. |


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


### HorizontalScale



HorizontalScale defines the policy for the number of replicas of the capsule It can both be configured with autoscaling and with a static number of replicas

_Appears in:_
- [Scale](#scale)

| Field | Description |
| --- | --- |
| `min` _integer_ | Min specifies the minimum amount of instances to run. |
| `max` _integer_ | Max specifies the maximum amount of instances to run. Omit to disable autoscaling. |
| `instances` _[Instances](#instances)_ | Instances specifies minimum and maximum amount of Capsule instances. Deprecated; use `min` and `max` instead. |
| `cpuTarget` _[CPUTarget](#cputarget)_ | CPUTarget specifies that this Capsule should be scaled using CPU utilization. |
| `customMetrics` _CustomMetric array_ | CustomMetrics specifies custom metrics emitted by the custom.metrics.k8s.io API which the autoscaler should scale on |


### HostCapsule







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.platform/v1`
| `kind` _string_ | `HostCapsule`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `name` _string_ | Name,Project,Environment is unique Project,Name referes to an existing Capsule type with the given name and project Will throw an error (in the platform) if the Capsule does not exist |
| `project` _string_ | Project references an existing Project type with the given name Will throw an error (in the platform) if the Project does not exist |
| `environment` _string_ | Environment references an existing Environment type with the given name Will throw an error (in the platform) if the Environment does not exist The environment also needs to be present in the parent Capsule |
| `network` _[HostNetwork](#hostnetwork)_ | Network mapping between the host network and the Kubernetes cluster network. When activated, traffic between the two networks will be tunneled according to the rules specified here. |


### HostNetwork





_Appears in:_
- [HostCapsule](#hostcapsule)

| Field | Description |
| --- | --- |
| `hostInterfaces` _[ProxyInterface](#proxyinterface) array_ | HostInterfaces are interfaces activated on the local machine (the host) and forwarded to the Kubernetes cluster capsules. |
| `capsuleInterfaces` _[ProxyInterface](#proxyinterface) array_ | CapsuleInterfaces are interfaces activated on the Capsule within the Kubernetes cluster and forwarded to the local machine (the host). The traffic is directed to a single target, e.g. `localhost:8080`. |
| `tunnelPort` _integer_ | TunnelPort for which the proxy-capsule should listen on. This is automatically set by the tooling. |


### InterfaceOptions





_Appears in:_
- [ProxyInterface](#proxyinterface)

| Field | Description |
| --- | --- |
| `tcp` _boolean_ | TCP enables layer-4 proxying in favor of layer-7 HTTP proxying. |
| `allowOrigin` _string_ | AllowOrigin sets the `Access-Control-Allow-Origin` Header on responses to the provided value, allowing local by-pass of CORS rules. Ignored if TCP is enabled. |
| `changeOrigin` _boolean_ | ChangeOrigin changes the Host header to match the given target. If not set, the Host header will be that of the original request. This does not impact the Origin header - use `Headers` to set that. Ignored if TCP is enabled. |
| `headers` _object (keys:string, values:string)_ | Headers to set on the proxy-requests. Ignored if TCP is enabled. |


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
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
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


### Scale





_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `horizontal` _[HorizontalScale](#horizontalscale)_ | Horizontal specifies the horizontal scaling of the Capsule. |
| `vertical` _[VerticalScale](#verticalscale)_ | Vertical specifies the vertical scaling of the Capsule. |




<hr class="solid" />


:::info generated from source code
This page is generated based on go source code. If you have suggestions for
improvements for this page, please open an issue at
[github.com/rigdev/rig](https://github.com/rigdev/rig/issues/new), or a pull
request with changes to [the go source
files](https://github.com/rigdev/rig/tree/main/pkg/api).
:::