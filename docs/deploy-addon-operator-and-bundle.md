## Create the Addon Operator for the POC

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [podman](https://podman.io/) or [docker](https://www.docker.com/)

## Deploy the Operator

```bash
(cd isv-addon && \
make docker-build docker-push IMG="quay.io/tmihalac/isv-addon-operator:v0.0.5")
```

> Don't forget to make `isv-addon-operator` PUBLIC.

## Generate the Bundle

```bash
(cd isv-addon && \
make bundle IMG="quay.io/tmihalac/isv-addon-operator:v0.0.5" CHANNELS="alpha" DEFAULT_CHANNEL="alpha" VERSION="0.0.5")
```

## Deploy the Bundle

```bash
(cd isv-addon && \
make bundle-build bundle-push BUNDLE_IMG="quay.io/tmihalac/isv-addon-operator-bundle:v0.0.5")
```

> Don't forget to make `isv-addon-operator-bundle` PUBLIC.
