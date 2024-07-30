# Sidecar Plugin

The `rigdev.sidecar` plugin adds a sidecar to the Capsule's deployment. Specifically, it appends the configured container to the Deployment's initcontainers with a `restartPolicy` of `Always`.

The config can be templated with standard Go templating and has
```
.capsule
```
as its templating context.

## Example
Config:
```yaml title="Helm values - Operator"
config:
  pipeline:
    steps:
      - plugins:
        - name: rigdev.sidecar
          config: |
            container:
              name: my-sidecar
              image: my-container-image:v1.1
```
The resulting Deployment resource of the Capsule
```
kind: Deployment
...
spec:
  template:
    spec:
      initContainers:
        - name: my-sidecar
          image: my-container-image:v1.1
          restartPolicy: Always
      ...
```
## Config



Configuration for the sidecar plugin

| Field | Description |
| --- | --- |
| `container` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#container-v1-core)_ | Container is the configuration of the sidecar injected into the deployment |



