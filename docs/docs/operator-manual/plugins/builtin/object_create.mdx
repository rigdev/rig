# Object Template Plugin

The `rigdev.object_create` creates a new arbitrary Kubernetes object from a YAML spec. The YAML spec must contain group, version and kind.

The config can be templated with standard Go templating and has
```
.capsule
```

If the name is empty, it defaults to the capsule name.

## Example
Config:
```yaml title="Helm values - Operator"
config:
  pipeline:
    steps:
      - plugins:
        - name: rigdev.object_create
          config: |
            object: |
              apiVersion: vpcresources.k8s.aws/v1beta1
              kind: SecurityGroupPolicy
              spec:
                podSelector:
                  matchLabels:
                    rig.dev/owned-by-capsule: {{ .capsule.metadata.name }}
                  securityGroups:
                    groupIds: {{ .capsule.metadata.annotations.groupIDs }}
```
The resulting Service resource of the Capsule, if the Capsule is named `my-capsule` and has `groupIDs: [id1, id2]` in its annotations:
```
apiVersion: vpcresources.k8s.aws/v1beta1
kind: SecurityGroupPolicy
metadata:
  name: my-capsule
spec:
  podSelector:
    matchLabels:
      rig.dev/owned-by-capsule: my-capsule
    securityGroups:
      groupIds: [id1, id2]
```

## Config



Configuration for the object_create plugin

| Field | Description |
| --- | --- |
| `object` _string_ | The yaml to apply as an object. The yaml can be templated. |



