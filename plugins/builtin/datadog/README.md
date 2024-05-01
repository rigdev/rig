# Datadog Plugin

The `rigdev.datadog` plugin adds Datadog specific tags to the Deployment and Pods of the capsule as requested [here](https://docs.datadoghq.com/tracing/trace_collection/library_injection_local/?tab=kubernetes). It can enable/disable the execution of the [Datadog Admission Controller](https://docs.datadoghq.com/containers/cluster_agent/admission_controller/?tab=operator) on the pods and sets the necessary library and unified service tags.

The config can be templated with standard Go templating and has
```
.capsule
```
as its templating context.

## Example
Config (in context of the rig-operator Helm values):
```
config:
  pipeline:
    steps:
      - plugins:
        - name: rigdev.datadog
          config: |
            libraryTag:
              java: v1.31.0
            unifiedServiceTags:
              env: my-env
              service: my-service
              versin: my-version
```

The resulting `Deployment` resource of the Capsule
```
kind: Deployment
metadata:
  ...
  labels:
		tags.datadoghq.com/env:     my-env,
		tags.datadoghq.com/service: my-service,
		tags.datadoghq.com/version: my-version,
spec:
  template:
    metadata:
      labels:
				admission.datadoghq.com/enabled: true,
				tags.datadoghq.com/env:          my-env,
				tags.datadoghq.com/service:      my-name,
				tags.datadoghq.com/version:     my-version,
      annotations:
				admission.datadoghq.com/java-lib.version: v1.31.0,
   ...
```
## Config



Configuration for the datadog plugin

| Field | Description |
| --- | --- |
| `dontAddEnabledAnnotation` _boolean_ | DontAddEnabledAnnotation toggles if the pods should have an annotation allowing the Datadog Admission controller to modify them. |
| `libraryTag` _[LibraryTag](#librarytag)_ | LibraryTag defines configuration for which datadog libraries to inject into the pods. |
| `unifiedServiceTags` _[UnifiedServiceTags](#unifiedservicetags)_ | UnifiedServiceTags configures the values for the Unified Service datadog tags. |



### LibraryTag

LibraryTag defines configuration for which datadog libraries to let the admission controller inject into the pods The admission controller will inject libraries from a container with the specified tag if the field is set.

| Field | Description |
| --- | --- |
| `java` _string_ | Tag of the Java library container |
| `javascript` _string_ | Tag of the JavaScript library container |
| `python` _string_ | Tag of the Python library container |
| `net` _string_ | Tag of the .NET library container |
| `ruby` _string_ | Tag of the Ruby library container |





### UnifiedServiceTags

UnifiedServiceTags configures the values of the Unified Service datadog tags on both Deployment and Pods

| Field | Description |
| --- | --- |
| `env` _string_ | The env tag |
| `service` _string_ | The service tag |
| `version` _string_ | The version tag |

