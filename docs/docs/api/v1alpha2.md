# API Reference

## Packages
- [rig.dev/v1alpha2](#rigdevv1alpha2)


## rig.dev/v1alpha2

Package v1alpha2 contains API Schema definitions for the v1alpha2 API group

### Resource Types
- [Capsule](#capsule)



#### CPUTarget



CPUTarget defines an autoscaler target for the CPU metric If empty, no autoscaling will be done

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `utilization` _integer_ |  |


#### Capsule



Capsule is the Schema for the capsules API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.dev/v1alpha2`
| `kind` _string_ | `Capsule`
| `kind` _string_ | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[CapsuleSpec](#capsulespec)_ |  |


#### CapsuleInterface



CapsuleInterface defines an interface for a capsule

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `name` _string_ |  |
| `port` _integer_ |  |
| `liveness` _[InterfaceProbe](#interfaceprobe)_ |  |
| `readiness` _[InterfaceProbe](#interfaceprobe)_ |  |
| `public` _[CapsulePublicInterface](#capsulepublicinterface)_ |  |


#### CapsuleInterfaceIngress

_Underlying type:_ _[struct{Host string "json:\"host\""}](#struct{host-string-"json:\"host\""})_

CapsuleInterfaceIngress defines that the interface should be exposed as http ingress

_Appears in:_
- [CapsulePublicInterface](#capsulepublicinterface)



#### CapsuleInterfaceLoadBalancer

_Underlying type:_ _[struct{Port int32 "json:\"port\""}](#struct{port-int32-"json:\"port\""})_

CapsuleInterfaceLoadBalancer defines that the interface should be exposed as a L4 loadbalancer

_Appears in:_
- [CapsulePublicInterface](#capsulepublicinterface)



#### CapsulePublicInterface



CapsulePublicInterface defines how to publicly expose the interface

_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `ingress` _[CapsuleInterfaceIngress](#capsuleinterfaceingress)_ |  |
| `loadBalancer` _[CapsuleInterfaceLoadBalancer](#capsuleinterfaceloadbalancer)_ |  |


#### CapsuleScale





_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `horizontal` _[HorizontalScale](#horizontalscale)_ |  |
| `vertical` _[VerticalScale](#verticalscale)_ |  |


#### CapsuleSpec



CapsuleSpec defines the desired state of Capsule

_Appears in:_
- [Capsule](#capsule)

| Field | Description |
| --- | --- |
| `image` _string_ |  |
| `command` _string_ |  |
| `args` _string array_ |  |
| `interfaces` _[CapsuleInterface](#capsuleinterface) array_ |  |
| `files` _[File](#file) array_ |  |
| `scale` _[CapsuleScale](#capsulescale)_ |  |
| `nodeSelector` _object (keys:string, values:string)_ |  |
| `env` _[Env](#env)_ |  |


#### DeploymentStatus





_Appears in:_
- [CapsuleStatus](#capsulestatus)

| Field | Description |
| --- | --- |
| `state` _string_ |  |
| `message` _string_ |  |


#### Env



Env defines what secrets and configmaps should be used for environment variables in the capsule.

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `disable_automatic` _boolean_ | DisableAutomatic sets wether the capsule should disable automatically use of existing secrets and configmaps which share the same name as the capsule as environment variables. |
| `from` _[EnvReference](#envreference) array_ | From holds a list of references to secrets and configmaps which should be mounted as environment variables. |


#### EnvReference



EnvSource holds a reference to either a ConfigMap or a Secret

_Appears in:_
- [Env](#env)

| Field | Description |
| --- | --- |
| `kind` _string_ | Kind is the resource kind of the env reference, must be ConfigMap or Secret. |
| `name` _string_ | Name is the name of a ConfigMap or Secret in the same namespace as the Capsule. |


#### File



File defines a mounted file and where to retrieve the contents from

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `ref` _[FileContentReference](#filecontentreference)_ |  |
| `path` _string_ |  |


#### FileContentReference



FileContentRef defines the name of a config resource and the key from which to retrieve the contents

_Appears in:_
- [File](#file)

| Field | Description |
| --- | --- |
| `kind` _string_ |  |
| `name` _string_ |  |
| `key` _string_ |  |


#### HorizontalScale



HorizontalScale defines the policy for the number of replicas of the capsule It can both be configured with autoscaling and with a static number of replicas

_Appears in:_
- [CapsuleScale](#capsulescale)

| Field | Description |
| --- | --- |
| `instances` _[Instances](#instances)_ |  |
| `cpuTarget` _[CPUTarget](#cputarget)_ |  |


#### Instances





_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `min` _integer_ |  |
| `max` _integer_ |  |


#### InterfaceGRPCProbe

_Underlying type:_ _[struct{Service string "json:\"service\""}](#struct{service-string-"json:\"service\""})_



_Appears in:_
- [InterfaceProbe](#interfaceprobe)



#### InterfaceProbe





_Appears in:_
- [CapsuleInterface](#capsuleinterface)

| Field | Description |
| --- | --- |
| `path` _string_ |  |
| `tcp` _boolean_ |  |
| `grpc` _[InterfaceGRPCProbe](#interfacegrpcprobe)_ |  |


#### OwnedResource





_Appears in:_
- [CapsuleStatus](#capsulestatus)

| Field | Description |
| --- | --- |
| `ref` _[TypedLocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#typedlocalobjectreference-v1-core)_ |  |
| `state` _string_ |  |
| `message` _string_ |  |


#### ResourceLimits





_Appears in:_
- [VerticalScale](#verticalscale)

| Field | Description |
| --- | --- |
| `request` _[Quantity](#quantity)_ |  |
| `limit` _[Quantity](#quantity)_ |  |


#### ResourceRequest





_Appears in:_
- [VerticalScale](#verticalscale)

| Field | Description |
| --- | --- |
| `request` _[Quantity](#quantity)_ |  |


#### VerticalScale





_Appears in:_
- [CapsuleScale](#capsulescale)

| Field | Description |
| --- | --- |
| `cpu` _[ResourceLimits](#resourcelimits)_ |  |
| `memory` _[ResourceLimits](#resourcelimits)_ |  |
| `gpu` _[ResourceRequest](#resourcerequest)_ |  |


