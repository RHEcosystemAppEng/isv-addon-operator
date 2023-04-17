## Create a placeholder for the Starburst Enterprise Operator

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [podman](https://podman.io/) or [docker](https://www.docker.com/)

## Deploy the Operator

```bash
(cd enterprise && \
make docker-build docker-push IMG="quay.io/tomerfi/enterprise-operator:v0.0.4")
```

> Don't forget to make `enterprise-operator` PUBLIC.

## Generate the Bundle

```bash
(cd enterprise && \
make bundle IMG="quay.io/tomerfi/enterprise-operator:v0.0.4" CHANNELS="alpha" DEFAULT_CHANNEL="alpha" VERSION="0.0.4")
```

## Deploy the Bundle

```bash
(cd enterprise && \
make bundle-build bundle-push BUNDLE_IMG="quay.io/tomerfi/enterprise-operator-bundle:v0.0.4")
```

> Don't forget to make `enterprise-operator-bundle` PUBLIC.
