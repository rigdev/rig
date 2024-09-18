# Ingress Routes Plugin

The `rigdev.ingress_routes` plugin handles the routes by creating an Ingress resource for each route. The Ingress resource is created with the annotations specified in the `RouteOptions`, with the class specified in the `ingressClassName` field in the config, and has tls specified if the `disableTLS` field is not set to true. Furthermore, the plugin will create a Certificate resource for the route hosts if the `createCertificateResources` field is set to true and a clusterIssuer is specified in the `clusterIssuer` field.

## Example
Config:
```yaml title="Helm values - Operator"
config:
  pipeline:
    routesStep:
      plugin: "rigdev.ingress_routes"
      config: |
        clusterIssuer: letsencrypt-prod
        createCertificateResources: true
        ingressClassName: nginx
        disableTLS: false
        annotations:
          my-annotation: "my-value"
```

## ALB Ingress Controller
If using The ALB Ingress Controller with `target-type=instance`, an additional service with the suffix `alb` is automatically created. This is done as the traffic is routed to NodePorts and then proxied to the pods. `instance` is The default type, but can be changed to `ip` if the annotation `alb.ingress.kubernetes.io/target-type: ip` is set.

## Nginx Ingress Controller
When using the Nginx Ingress Controller, it is possible to use rewrite-targets. This is done by setting the annotation `nginx.ingress.kubernetes.io/rewrite-target: /` to the desired URI.

If the path-type is set to `RegularExpression`, the annotation `nginx.ingress.kubernetes.io/use-regex: "true"` is automatically set.

## Config



Configuration for the ingress_routes plugin

| Field | Description |
| --- | --- |
| `clusterIssuer` _string_ | ClusterIssuer to use for issueing ingress certificates |
| `createCertificateResources` _boolean_ | CreateCertificateResources specifies wether to create Certificate<br />resources. If this is not enabled we will use ingress annotations. This<br />is handy in environments where the ingress-shim isn't enabled. |
| `ingressClassName` _string_ | ClassName specifies the default ingress class to use for all ingress<br />resources created. |
| `disableTLS` _boolean_ | DisableTLS for ingress resources generated. This is useful if a 3rd-party component<br />is handling the HTTPS TLS termination and certificates. |
| `annotations` _object (keys:string, values:string)_ | Annotations to be added to all ingress resources created. |



