
# rig.dev/v1alpha1

Package v1alpha1 contains API Schema definitions for the  v1alpha1 API group

## Resource Types
- [Capsule](#capsule)



### CPUTarget



CPUTarget defines an autoscaler target for the CPU metric If empty, no autoscaling will be done

_Appears in:_
- [HorizontalScale](#horizontalscale)

| Field | Description |
| --- | --- |
| `averageUtilizationPercentage` _integer_ | AverageUtilizationPercentage sets the utilization which when exceeded will trigger autoscaling. |


### Capsule



Capsule is the Schema for the capsules API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `rig.dev/v1alpha1`
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
| `public` _[CapsulePublicInterface](#capsulepublicinterface)_ | Public specifies if and how the interface should be published. |


### CapsuleInterfaceIngress

_Underlying type:_ _[struct{Host string "json:\"host\""}](#struct{host-string-"json:\"host\""})_

CapsuleInterfaceIngress defines that the interface should be exposed as http ingress

_Appears in:_
- [CapsulePublicInterface](#capsulepublicinterface)



### CapsuleInterfaceLoadBalancer

_Underlying type:_ _[struct{Port int32 "json:\"port\""; NodePort int32 "json:\"nodePort,omitempty\""}](#struct{port-int32-"json:\"port\"";-nodeport-int32-"json:\"nodeport,omitempty\""})_

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


### CapsuleSpec



CapsuleSpec defines the desired state of Capsule

_Appears in:_
- [Capsule](#capsule)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | Replicas specifies how many replicas the Capsule should have. |
| `image` _string_ | Image specifies what image the Capsule should run. |
| `command` _string_ | Command is run as a command in the shell. If left unspecified, the container will run using what is specified as ENTRYPOINT in the Dockerfile. |
| `args` _string array_ | Args is a list of arguments either passed to the Command or if Command is left empty the arguments will be passed to the ENTRYPOINT of the docker image. |
| `interfaces` _[CapsuleInterface](#capsuleinterface) array_ | Interfaces specifies the list of interfaces the the container should have. Specifying interfaces will create the corresponding kubernetes Services and Ingresses depending on how the interface is configured. |
| `env` _[Env](#env)_ | Env specifies configuration for how the container should obtain environment variables. |
| `files` _[File](#file) array_ | Files is a list of files to mount in the container. These can either be based on ConfigMaps or Secrets. |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)_ | Resources describes what resources the Capsule should have access to. |
| `imagePullSecret` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#localobjectreference-v1-core)_ | ImagePullSecret is a reference to a secret holding docker credentials for the registry of the image. |
| `horizontalScale` _[HorizontalScale](#horizontalscale)_ | HorizontalScale describes how the Capsule should scale out |
| `serviceAccountName` _string_ | ServiceAccountName specifies the name of an existing ServiceAccount which the Capsule should run as. |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector for what nodes the Capsule should live on. |


### Env



Env defines what secrets and configmaps should be used for environment variables in the capsule.

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `automatic` _boolean_ | Automatic sets wether the capsule should automatically use existing secrets and configmaps which share the same name as the capsule as environment variables. |
| `from` _[EnvSource](#envsource) array_ | From holds a list of references to secrets and configmaps which should be mounted as environment variables. |


### EnvSource



EnvSource holds a reference to either a ConfigMap or a Secret

_Appears in:_
- [Env](#env)

| Field | Description |
| --- | --- |
| `configMapName` _string_ | ConfigMapName is the name of a ConfigMap in the same namespace as the Capsule |
| `secretName` _string_ | SecretName is the name of a Secret in the same namespace as the Capsule |


### File



File defines a mounted file and where to retrieve the contents from

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `path` _string_ | Path specifies the full path where the File should be mounted including the file name. |
| `configMap` _[FileContentRef](#filecontentref)_ | ConfigMap specifies that this file is based on a key in a ConfigMap. The ConfigMap field is mutually exclusive with Secret. |
| `secret` _[FileContentRef](#filecontentref)_ | Secret specifies that this file is based on a key in a Secret. The Secret field is mutually exclusive with ConfigMap. |


### FileContentRef



FileContentRef defines the name of a config resource and the key from which to retrieve the contents

_Appears in:_
- [File](#file)

| Field | Description |
| --- | --- |
| `name` _string_ | Name specifies the name of the Secret or ConfigMap. |
| `key` _string_ | Key specifies the key holding the file contents. |


### HorizontalScale



HorizontalScale defines the policy for the number of replicas of the capsule It can both be configured with autoscaling and with a static number of replicas

_Appears in:_
- [CapsuleSpec](#capsulespec)

| Field | Description |
| --- | --- |
| `minReplicas` _integer_ | MinReplicas is the minimum amount of replicas that the Capsule should have. |
| `maxReplicas` _integer_ | MaxReplicas is the maximum amount of replicas that the Capsule should have. |
| `cpuTarget` _[CPUTarget](#cputarget)_ | CPUTarget specifies that this Capsule should be scaled using CPU utilization. |




