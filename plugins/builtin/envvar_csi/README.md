# Environment Variable from CSI plugin

The `rigdev.envvar_csi` plugin loads environment variables from a [CSI provider](https://secrets-store-csi-driver.sigs.k8s.io/concepts) into a Pod using a [synced Kubernetes secret](https://secrets-store-csi-driver.sigs.k8s.io/topics/sync-as-kubernetes-secret).

The plugin currently supports the `aws` driver which needs to be installed in an AWS cluster. See [here](https://docs.aws.amazon.com/systems-manager/latest/userguide/integrating_csi_driver.html#integrating_csi_driver_install) for a guide on how to install the AWS Secrets and Configuration Provider. 

## AWS Provider
The plugin reads the environment variables set in the `.spec.env.raw` field of the Platform Capsule and decides if they should be injected as a CSI environment variable. The syntax for the envionment variables is 
- `ENV_VAR: __ssmParameter__=<MY-PARAMETER>`: Will try to load an object of type `ssmparameter` with name `<MY-PARAMETER>` and store it in the environment varable `ENV_VAR`
- `ENV_VAR: __secretName__=<MY-SECRET>`: Will try to load an object of type `secretsmanager` with name `<MY-SECRET>` and store it in the environment variable `ENV_VAR`

## Example

Config:

```yaml title="Helm values - Operator"
config:
  pipeline:
    steps:
      - plugins:
          - plugin: rigdev.envvar_csi
            config: |
              provider: aws


# You have to give the Rig Operator permission to read/write SecretProviderClass objects
rbac:
  rules: 
  - apiGroups:
    - secrets-store.csi.x-k8s.io
    resources:
    - secretproviderclasses
```

```yaml title="Platform Capsule"
apiVersion: platform.rig.dev/v1
kind: Capsule
project: myproject
environment: myenv
name: mycapsule
spec:
  image: myimage
  env:
    raw:
      NORMAL_VAR: some_value
      SSM_PARAMETER: __ssmParameter__=SomeParameter
      SECRET_PARAMETER: __secretName__=SomeSecret
```

The resulting `Deployment` and `SecretProviderClass` resource of the Capsule

```yaml title=Deployment
kind: Deployment
metadata:
  name: mycapsule
  namespace: myproject
spec:
  template:
    spec:
      containers:
        name: mycapsule
        image: myimage
        envFrom:
        - configMapRef:
            name: cap
        - secretRef:
            name: csi-envvars-cap
      volumes:
      - csi:
          driver: secrets-store.csi.k8s.io
          readOnly: true
          volumeAttributes:
            secretProviderClass: mycapsule
        name: csi
   ...
```

```yaml title=SecretProviderClass
kind: SecretProviderClass
metadata:
  name: mycapsule
  namespace: myproject
spec:
  parameters:
    objects: |
      - objectName: SomeParameter
        objectType: ssmparameter
      - objectName: SomeSecret
        objectType: secretsmanager
  provider: aws
  secretObjects:
  - secretName: csi-envvars-mycapsule
    type: Opaque
    data:
    - key: SSM_PARAMETER
      objectName: MyParameter
    - key: SECRET_PARAMETER
      objectName: SomeSecret
```

The SecretProviderClass will then construct a Kubernetes secret named `csi-envvars-mycapsule` and inject it into the Capsule's pods.

```yaml title="Secret owned by SecretProviderClass"
kind: Secret
metadata:
  name: csi-envvars-mycapsule
  namespace: myproject
type: opaque
data:
  SSM_PARAMETER: ...
  SECRET_PARAMETER: ...
```
## Config
