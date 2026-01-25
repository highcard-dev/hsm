# Helm Charts Repository

This branch contains the Helm chart repository index for HSM (Hytale Server Manager).

## Usage

Add this repository to Helm:

```bash
helm repo add hsm https://highcard-dev.github.io/hsm/
helm repo update
```

Install the chart:

```bash
helm install hsm hsm/hsm
```

## Charts

- **hsm**: Hytale Server Manager Helm chart

