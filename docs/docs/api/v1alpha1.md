# API Reference

## Packages
- [rig.dev/v1alpha1](#rigdevv1alpha1)


## rig.dev/v1alpha1

Package v1alpha1 contains API Schema definitions for the  v1alpha1 API group

### Resource Types
- [Capsule](#capsule)



#### CPUTarget



CPUTarget defines an autoscaler target for the CPU metric If empty, no autoscaling will be done

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `averageUtilizationPercentage` _integer_ |  |


#### Capsule



Capsule is the Schema for the capsules API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.dev/v1alpha1`
| `kind` _string_ | `Capsule`
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
| `public` _[CapsulePublicInterface](#capsulepublicinterface)_ |  |


#### CapsuleInterfaceIngress

_Underlying type:_ _[struct{Host string "json:\"host\""}](#struct{host-string-"json:\"host\""})_

CapsuleInterfaceIngress defines that the interface should be exposed as http ingress

_Appears in:_
- [CapsulePublicInterface](#capsulepublicinterface)



#### CapsuleInterfaceLoadBalancer

_Underlying type:_ _[struct{Port int32 "json:\"port\""; NodePort int32 "json:\"nodePort,omitempty\""}](#struct{port-int32-"json:\"port\"";-nodeport-int32-"json:\"nodeport,omitempty\""})_

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


#### CapsuleSpec



CapsuleSpec defines the desired state of Capsule

_Appears in:_
- [Capsule](#capsule)

| Field | Description |
| --- | --- |
| `replicas` _integer_ |  |
| `image` _string_ |  |
| `command` _string_ |  |
| `args` _string array_ |  |
| `interfaces` _[CapsuleInterface](#capsuleinterface) array_ |  |
| `env` _[Env](#env)_ |  |
| `files` _[File](#file) array_ |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)_ |  |
| `imagePullSecret` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#localobjectreference-v1-core)_ |  |
| `horizontalScale` _[HorizontalScale](#horizontalscale)_ |  |
| `serviceAccountName` _string_ |  |
| `nodeSelector` _object (keys:string, values:string)_ |  |


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
| `automatic` _boolean_ | Automatic sets wether the capsule should automatically use existing secrets and configmaps which share the same name as the capsule as environment variables. |
| `from` _[EnvSource](#envsource) array_ | From holds a list of references to secrets and configmaps which should be mounted as environment variables. |


#### EnvSource



EnvSource holds a reference to either a ConfigMap or a Secret

_Appears in:_
- [Env](#env)

| Field | Description |
| --- | --- |
| `configMapName` _string_ | ConfigMapName is the name of a ConfigMap in the same namespace as the Capsule |
| `secretName` _string_ | SecretName is the name of a Secret in the same namespace as the Capsule |


#### File



File defines a mounted file and where to retrieve the contents from

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `path` _string_ |  |
| `configMap` _[FileContentRef](#filecontentref)_ |  |
| `secret` _[FileContentRef](#filecontentref)_ |  |


#### FileContentRef



FileContentRef defines the name of a config resource and the key from which to retrieve the contents

_Appears in:_
- [File](#file)

| Field | Description |
| --- | --- |
| `name` _string_ |  |
| `key` _string_ |  |


#### HorizontalScale



HorizontalScale defines the policy for the number of replicas of the capsule It can both be configured with autoscaling and with a static number of replicas

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `minReplicas` _integer_ |  |
| `maxReplicas` _integer_ |  |
| `cpuTarget` _[CPUTarget](#cputarget)_ |  |


#### OwnedResource





_Appears in:_
- [CapsuleStatus](#capsulestatus)

| Field | Description |
| --- | --- |
| `ref` _[TypedLocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#typedlocalobjectreference-v1-core)_ |  |
| `state` _string_ |  |
| `message` _string_ |  |




