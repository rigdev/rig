# Placement Plugin

The `rigdev.placement` plugin adds placement configuration to the Deployment resource of your capsule.
It can modify the `nodeSelector` and `tolerations` fields of the deployment.
It has a `requireTag` bool config value. If set to `true`, the plugin will only run on capsules `rigdev.placement/tag` annotation matches the `tag` of the placement plugin. This also means the `tag` must be set on the plugin if `requireTag` is true.

## Example

Config:

```yaml title="Helm values - Operator"
config:
  pipeline:
    steps:
      - plugins:
          - plugin: rigdev.placement
            config: |
              nodeSelector:
                key1: value1
              tolerations:
                - key: some-key
                  value: some-value
```

The `Deployment` resource of the Capsule

```
kind: Deployment
...
spec:
  template:
    spec:
      nodeSelector:
        key2: value2
      tolerations:
        - key: some-other-key
          value: some-other-value
   ...
```

The resulting config of the `Deployment` is

```
kind: Deployment
...
spec:
  template:
    spec:
      nodeSelector:
        key1: value1
        key2: value2
      tolerations:
        - key: some-other-key
          value: some-other-value
        - key: some-key
          value: some-value
   ...
```

## Config



Configuration for the placement plugin

| Field | Description |
| --- | --- |
| `nodeSelector` _object (keys:string, values:string)_ | Nodeselectors which will be inserted into the deployment's podSpec |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#toleration-v1-core) array_ | Tolerations which will be appended to the deployment's podSpec |
| `requireTag` _boolean_ | True if a capsule needs a Tag annotation to be run |



