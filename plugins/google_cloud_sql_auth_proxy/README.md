# Google Cloud SQL Auth Proxy Plugin

The `rigdev.google_cloud_sql_auth_proxy` plugins injects a Google Cloud SQL auth proxy container into your deployment as a [sidecar](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/). See [here](https://cloud.google.com/sql/docs/mysql/sql-proxy) for a description of the auth proxy.
It will append a container named `google-cloud-sql-proxy` running the `gcr.io/cloud-sql-connectors/cloud-sql-proxy` image to your deployment and set its arguments, environment variables and config files according the the configuration of this plugin.

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
        - name: rigdev.google_cloud_sql_auth_proxy
          config: |
            tag: 2.9  
            args
              - arg1
              - arg2
            envFromSource:
              - kind: ConfigMap
                name: my-configmap
            envVars:
              name: MY_ENV_VAR
              value: some-value
            resources:
              cpu: 0.1
              memory: 128M
            instanceConnectionNames:
              - instance-name
```
Resulting Deployment
```
kind: Deployment
spec:
  initContainers:
    ...
    - name: google-cloud-sql-proxy
      image: gcr.io/cloud-sql-connectors/cloud-sql-proxy
      args: ['instance-name', 'arg1', 'arg2']
      envFrom:
        - configMapRef:
            name: my-configmap
      env:
        - name: MY_ENV_VAR
          value: some-value
      resources:
        requests:
          cpu: 0.1
          memory: 128M
      securityContext:
        runAsNonRoot: true
      restartPolicy: Always
```
## Config



Configuration for the google_cloud_sql_auth_proxy plugin

| Field | Description |
| --- | --- |
| `image` _string_ | The image running on the new container. Defaults to gcr.io/cloud-sql-connectors/cloud-sql-proxy |
| `tag` _string_ | The tag of the image |
| `args` _string array_ | Arguments to pass to the cloud sql proxy. These will be appended after the instance connection names. |
| `envFromSource` _EnvReference array_ | A list of either ConfigMaps or Secrets which will be mounted in as environment variables to the container. It's a reuse of the Capsule CRD |
| `envVars` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envvar-v1-core) array_ | A list of environment variables to set in the container |
| `files` _File array_ | Files is a list of files to mount in the container. These can either be based on ConfigMaps or Secrets. It's a reuse of the Capsule CRD |
| `resources` _[Resources](#resources)_ | Resources defines how large the container request should be. Defaults to the Kubernetes defaults. |
| `instanceConnectionNames` _string array_ | The instance_connection_names passed to the cloud_sql_proxy. |





### Resources

Resources configures the size of the request of the cloud_sql_proxy container

| Field | Description |
| --- | --- |
| `cpu` _string_ | The number of CPU cores to request. |
| `memory` _string_ | The bytes of memory to request. |

