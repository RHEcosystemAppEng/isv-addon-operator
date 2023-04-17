# Running the POC

Cloning is **not** required, all operators, bundles, and catalogs, are publicly accessible on *quay.io*.<br/>
Follow the instructions in this document and you will:

- Create a Kind cluster
- Install OLM
- Apply a *CatalogSource* CR referencing our *isv-addon* and *enterprise* packages
- Create a *Secret* representing the *Addon* parameters (tile inputs), i.e. the license key
- Create a *Secret* representing the *Vault* data, i.e. the enterprise manifest
- Subscribe to the *isv-addon* package (depends on *enterprise* and *ose-prometheus*)
  - Verify all operators are up
- Apply a *ISVAddon* CR
  - Verify creation of the *StarburstEnterprise* CR and owner
  - Verify creation of the starburst license *Secret* (required for the enterprise)
  - Attempt (and fail) manually deleting the *StarburstEnterprise* CR
  - Attempt (and fail) manually patching the *StarburstEnterprise* CR
  - Attempt (and fail) manually creating a new *StarburstEnterprise* CR
- Delete the *ISVAddon* CR
  - Verify deletion of the *StarburstEnterprise* CR
- Delete the previously created Kind cluster

## Prerequisites

- [kind](https://kind.sigs.k8s.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [operator-sdk](https://sdk.operatorframework.io/docs/installation/)

## Prepare POC environment

```bash
kind create cluster --name starburst
```

```bash
operator-sdk olm install
```

## Load OSE Images

Follow instructions [here](load-required-ose-images.md) to load the required images

## Create catalog source

```bash
kubectl apply -f -<< EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: starburst-combined-catalog
  namespace: olm
spec:
  sourceType: grpc
  grpcPodConfig:
    securityContextConfig: restricted
  image: quay.io/tmihalac/starburst-combined-catalog:dev
  displayName: Starburst Combined Catalog
  publisher: Me
  updateStrategy:
    registryPoll:
      interval: 60m
EOF
```

```bash
# the package manifest records can take up to a minute to appear
$ watch 'kubectl get packagemanifests.packages.operators.coreos.com | grep "starburst\|ose"'

ose-prometheus-operator                    Starburst Combined Catalog   45s
isv-addon                            Starburst Combined Catalog   45s
enterprise                       Starburst Combined Catalog   45s
```

## Subscribe to addon package

### Create target namespace

```bash
kubectl create ns starburst-playground
```

### Create the Secret for the Addon parameters

```bash
# a secret named "addon" structured from the vault keys will be manifested by osd
kubectl create secret generic addon-managed-starburst-parameters -n starburst-playground \
--from-literal starburst-license="dummy-license-value-here"
```

### Create the Secret for the Vault data

```bash
# a secret named "addon" structured from the vault keys will be manifested by osd
kubectl create secret generic isv-addon -n starburst-playground \
--from-file enterprise.yaml=enterprise/config/samples/example.com_v1alpha1_starburstenterprise.yaml \
--from-literal remote-write-url="dummy-remote-write-url" \
--from-literal token-url="dummy-token-url" \
--from-literal client-secret="dummy-client-secret" \
--from-literal client-id="dummy-client-id" \
--from-file rules=isv-addon/test-resources/rules.yaml
```

### Apply OperatorGroup for target namespace

```bash
kubectl apply -f -<< EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: my-starburst-op-group
  namespace: starburst-playground
spec:
  targetNamespaces:
  - starburst-playground
EOF
```

### Apply a Subscription to the addon package

```bash
kubectl apply -f -<< EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: isv-addon-subscription
  namespace: starburst-playground
spec:
  channel: alpha1
  name: isv-addon
  source: starburst-combined-catalog
  sourceNamespace: olm
  installPlanApproval: Automatic
EOF
```

### Verify all operators are up

```bash
# might take a couple of minutes
$ watch kubectl get pods -n starburst-playground

NAME                                                          READY   STATUS    RESTARTS   AGE
prometheus-operator-586d485cf9-d8tgl                          1/1     Running   0          45s
isv-addon-controller-manager-7c7f978ff-svqv4            2/2     Running   0          51s
isv-addon-validate-enterprise-webhook-685f4f7db-8qwjr   2/2     Running   0          51s
enterprise-controller-manager-8b4869d99-6wncl       2/2     Running   0          54s
```

### Verify all packages are deployed successfully

```bash
$ kubectl get csv -n starburst-playground

NAME                             DISPLAY                VERSION   REPLACES   PHASE
ose-prometheus-operator.4.10.0   Prometheus Operator    4.10.0               Succeeded
isv-addon.v0.0.4           ISV Addon        0.0.4                Succeeded
enterprise.v0.0.4      Starburst Enterprise   0.0.4                Succeeded
```

## Run ISV Addon

### Apply a ISVAddon CR

```bash
kubectl apply -f -<< EOF
apiVersion: example.com.example.com/v1alpha1
kind: ISVAddon
metadata:
  labels:
    app: isv-addon
  name: isvaddon-sample
  namespace: starburst-playground
EOF
```

### Verify a StarburstEnterprise CR was created

```bash
# this might take a couple of seconds to appear
$ watch kubectl get isventerprises -n starburst-playground

NAME                               AGE
isvaddon-sample-enterprise   10s
```

### Verify the owner for the StarburstEnterprise CR

```bash
$ kubectl get isventerprises -n starburst-playground isvaddon-sample-enterprise -o jsonpath='{.metadata.ownerReferences[0]}' | jq

{
  "apiVersion": "example.com.example.com/v1alpha1",
  "blockOwnerDeletion": true,
  "controller": true,
  "kind": "ISVAddon",
  "name": "isvaddon-sample",
  "uid": "..."
}
```

### Verify license secret creation (required for enterprise)

```bash
# note fetching license from enterprise secret and not addon parameter's one
$ kubectl get secret -n starburst-playground starburst-license -o jsonpath='{.data.starburstdata\.license}' | base64 --decode && echo

dummy-license-value-here
```

### Attempt (and fail) manually deleting the StarburstEnterprise CR

```bash
$ kubectl delete isventerprises -n starburst-playground isvaddon-sample-enterprise

Error from server (Unauthorized Access): admission webhook "isv-addon-validate-enterprise-webhook.example.com.example.com" denied the request: DELETE not allowed
```

### Attempt (and fail) manually patching the StarburstEnterprise CR

```bash
$ kubectl label isventerprises -n starburst-playground isvaddon-sample-enterprise  someKey=someValue

Error from server (Unauthorized Access): admission webhook "isv-addon-validate-enterprise-webhook.example.com.example.com" denied the request: UPDATE not allowed
```

### Attempt (and fail) manually creating a new StarburstEnterprise CR

```bash
$ kubectl apply -f -<< EOF
apiVersion: example.com.example.com/v1alpha1
kind: StarburstEnterprise
metadata:
  name: new-enterprise-cr
  namespace: starburst-playground
EOF

Error from server (Unauthorized Access): error when creating "STDIN": admission webhook "isv-addon-validate-enterprise-webhook.example.com.example.com" denied the request: CREATE not allowed
```

### Delete the ISVAddon CR

```bash
kubectl delete isvaddons -n starburst-playground isvaddon-sample
```

### Verify the StarburstEnterprise CR was deleted

```bash
$ kubectl get isventerprises -n starburst-playground

No resources found in starburst-playground namespace.
```

## Teardown POC environment

```bash
kind delete cluster --name starburst
```
