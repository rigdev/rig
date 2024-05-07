# Object Template Plugin

The `rigdev.object_template` patches a YAML spec to a Kubernetes object defined by a Group, Kind and Name.

The config can be templated with standard Go templating and has
```
.capsule
.current
```
as its templating context where `.current` refers to the current version of the object being modified.

If the name is empty, it defaults to the capsule name. If it is '*' it will execute the object template on all objects of the given Group and Kind. For each object, `.current` will refer to that specific object when templating.

## Example
Config (in context of the rig-operator Helm values):
```
config:
  pipeline:
    steps:
      - plugins:
        - name: rigdev.object_template
          config: |
            object: | 
              spec:
               externalName: some-name 
            group: core
            kind: Service
            name: {{ .capsule.metadata.name }}
```
The resulting Service resource of the Capsule, if the Capsule is named `my-capsule`
```
kind: Service
metadata:
  name: my-capsule
...
spec:
  externalName: some-name
  ...
```
## Config



Configuration for the object_template plugin

| Field | Description |
| --- | --- |
| `object` _string_ | The yaml to apply to the object. The yaml can be templated. |
| `group` _string_ | Group to match, for which objects to apply the patch to. |
| `kind` _string_ | Kind to match, for which objects to apply the patch to. |
| `name` _string_ | Name of the object to match. Default to Capsule-name. If '*'|



