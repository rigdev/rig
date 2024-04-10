# Annotations Plugin

The annotations plugin can insert annotations and labels into a given object.
If any of the given annotations or labels are already present in the object, they will be overwritten.

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
        - name: rigdev.annotations
          config: |
            annotations:
              key1: value1
            labels:
              key2: value2
            group: apps
            kind: Deployment
```

If the name of the capsule in the request context is `my-capsule` with corresponding `Deployment`
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-capsule
  annotations:
    key1: some-other-value
  labels:
    label: value
  ....
```
The resulting config of the `Deployment` is
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-capsule
  annotations:
    key1: value1
  labels:
    label: value
    key2: value2
  ....
```
## Config



Configuration for the annotations plugin

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations are the annotations to insert into the object |
| `labels` _object (keys:string, values:string)_ | Labels are the labels to insert into the object |
| `group` _string_ | Group to match, for which objects to apply the patch to. |
| `kind` _string_ | Kind to match, for which objects to apply the patch to. |
| `name` _string_ | Name of the object to match. Defaults to Capsule-name. |



