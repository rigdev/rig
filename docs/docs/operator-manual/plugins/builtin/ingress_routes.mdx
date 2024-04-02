# Ingress Routes Plugin
The way the operator handles routes in the reconcilliation pipeline is interchangable. This means, that the operator can be configured to handle routes in different ways by using different plugins. This is specified in the `routes_step` in the pipeline in the operator config.

The `rigdev.ingress_routes` plugin handles the routes by creating an Ingress resource for each route. The Ingress resource is created with the annotations specified in the `RouteOptions`, with the class specified in the `ingressClassName` field in the config, and has tls specified if the `disableTLS` field is not set to true. Furthermore, the plugin will create a Certificate resource for the route hosts if the `createCertificateResources` field is set to true and a clusterIssuer is specified in the `clusterIssuer` field.

## Example
Config:
```
clusterIssuer: letsencrypt-prod
createCertificateResources: true
ingressClassName: nginx
disableTLS: false
```

## Config



Configuration for the ingress_routes plugin

| Field | Description |
| --- | --- |
| `clusterIssuer` _string_ | ClusterIssuer to use for issueing ingress certificates |
| `createCertificateResources` _boolean_ | CreateCertificateResources specifies wether to create Certificate resources. If this is not enabled we will use ingress annotations. This is handy in environments where the ingress-shim isn't enabled. |
| `ingressClassName` _string_ | ClassName specifies the default ingress class to use for all ingress resources created. |
| `disableTLS` _boolean_ | DisableTLS for ingress resources generated. This is useful if a 3rd-party component is handling the HTTPS TLS termination and certificates. |



