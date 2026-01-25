# HSM Helm Chart

This Helm chart deploys the Hytale Server Manager (HSM) on Kubernetes.

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

| Parameter          | Description             | Default     |
| ------------------ | ----------------------- | ----------- |
| `service.type`     | Kubernetes service type | `ClusterIP` |
| `service.port`     | HTTP service port       | `8080`      |
| `service.gamePort` | Game server port        | `40000`     |

### HSM Configuration

| Parameter                | Description           | Default |
| ------------------------ | --------------------- | ------- |
| `hsm.config.authEnabled` | Enable authentication | `true`  |
| `hsm.config.port`        | Port to listen on     | `8080`  |
| `hsm.config.logLevel`    | Log level             | `info`  |

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

## Upgrading

```bash
helm upgrade hsm ./helm
```

## Uninstalling

```bash
helm uninstall hsm
```
