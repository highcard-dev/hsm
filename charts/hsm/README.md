# HSM Helm Chart

This Helm chart deploys the Hytale Server Manager (HSM) on Kubernetes as a Service.

## Installation

```bash
# Install with default values
helm install hsm ./helm

# Install with custom values
helm install hsm ./helm -f custom-values.yaml

# Install in a specific namespace
helm install hsm ./helm --namespace hytale --create-namespace
```

## Configuration

The following table lists the configurable parameters of the HSM chart and their default values.

### General Parameters

| Parameter          | Description            | Default                      |
| ------------------ | ---------------------- | ---------------------------- |
| `replicaCount`     | Number of HSM replicas | `1`                          |
| `image.repository` | HSM image repository   | `hsm`                        |
| `image.tag`        | HSM image tag          | `""` (uses Chart.appVersion) |
| `image.pullPolicy` | Image pull policy      | `IfNotPresent`               |

### Service Parameters

| Parameter      | Description             | Default     |
| -------------- | ----------------------- | ----------- |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | HTTP service port       | `8080`      |

### HSM Configuration

| Parameter                 | Description                                     | Default |
| ------------------------- | ----------------------------------------------- | ------- |
| `hsm.config.authEnabled`  | Enable authentication                           | `true`  |
| `hsm.config.port`         | Port to listen on                               | `8080`  |
| `hsm.config.logLevel`     | Log level                                       | `info`  |
| `hsm.useServiceAccount`   | Use Kubernetes service account authentication   | `false` |
| `hsm.jwks_endpoint`       | JWKS endpoint URL for JWT validation (optional) | `""`    |
| `hsm.jwks_ca_cert`        | CA certificate file path for JWKS endpoint      | `""`    |
| `hsm.jwks_ca_cert_secret` | Secret name containing CA certificate           | `""`    |

### Persistence

| Parameter                      | Description        | Default         |
| ------------------------------ | ------------------ | --------------- |
| `hsm.persistence.enabled`      | Enable persistence | `true`          |
| `hsm.persistence.storageClass` | Storage class      | `""`            |
| `hsm.persistence.accessMode`   | Access mode        | `ReadWriteOnce` |
| `hsm.persistence.size`         | Storage size       | `10Gi`          |

### Ingress

| Parameter           | Description        | Default                                                           |
| ------------------- | ------------------ | ----------------------------------------------------------------- |
| `ingress.enabled`   | Enable ingress     | `false`                                                           |
| `ingress.className` | Ingress class name | `""`                                                              |
| `ingress.hosts`     | Ingress hosts      | `[{host: "hsm.local", paths: [{path: "/", pathType: "Prefix"}]}]` |

## Examples

### Enable Ingress

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: hsm.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: hsm-tls
      hosts:
        - hsm.example.com
```

### Custom Resource Limits

```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

### Use Existing PVC

```yaml
hsm:
  persistence:
    enabled: true
    existingClaim: my-existing-pvc
```

### Enable Kubernetes Service Account Authentication

For native Kubernetes service account authentication (simplest setup):

```yaml
hsm:
  useServiceAccount: true
```

This automatically configures:

- JWKS endpoint: `https://kubernetes.default.svc/openid/v1/jwks`
- CA cert: `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`

### Enable JWT Authentication with Custom CA Certificate

For Kubernetes environments with custom CA certificates:

```yaml
hsm:
  jwks_endpoint: "https://your-auth-server/.well-known/jwks.json"
  jwks_ca_cert: "/etc/ssl/certs/ca.crt"
  jwks_ca_cert_secret: "your-ca-cert-secret"
```

First, create the secret containing your CA certificate:

```bash
kubectl create secret generic your-ca-cert-secret --from-file=ca.crt=/path/to/your/ca.crt
```

## Upgrading

```bash
helm upgrade hsm ./helm
```

## Uninstalling

```bash
helm uninstall hsm
```
