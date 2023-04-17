# ISV Addon operator template

[Sequence Diagram Quick Access](#sequence-diagram)

## Prologue

This repository holds the ISV Addon operator template from which future isv addon operator should start from.<br/>
The new design leverages *OLM* dependencies mechanism for operator dependencies.<br/>
A *Dynamic Admission Webhook* for provisioning the *Enterprise* CR modifications.<br/>
And a *Finalizer* for cleaning up the *Enterprise* CR.

## Repository content

Next to this document, you'll find two operators with their respective bundle manifests (deployed to *quay.io*).

The [isv-addon](/isv-addon) controller is the shell for the new design, this part will be implemented in our current operator, if this design gets accepted.

The [olm-catalog](/olm-catalog) folder holds the sources for the OLM catalog.

The [docs](/docs) folder holds documentation for running this POC, as well as building it from scratch.

> This POC is demonstrating the usage of webhooks, dependencies, and finalizers, it doesn't hold the code for our current operator.

## Design walkthrough

The *isv-addon* describes the *enterprise* package as its dependency and runs 2 controllers.<br/>
The *addon* controller creates the *StarburstEnterprise* CR.<br/>
The *webhook* controller provisions all *StarburstEnterprise* CRs.<br/>
The *addon* controller leverages a *finalizer* for cleaning up existing *StarburstEnterprise* CRs.

In this design, the *addon* controller will handle the actions currently handled by our operator.<br/>
Reconciling the required resources, i.e. secrets, prometheus servers, and service monitors.<br/>
And eventually deploying the required *StarburstEnterprise* CR.

The *webhook*, makes our current running *cronjob* redundant by intercepting all API requests for creating/patching/deleting *StarburstEnterprise* CRs.<br/>
Prohibiting any modifications to our CR.

## Entity Relationship

```mermaid
erDiagram
  STARBURST-ADDON ||--o{ enterprise : depends-on
```

## Sequence Diagram

```mermaid
sequenceDiagram
  participant P as Platform
  participant A as Addon Controller
  participant E as Enterprise Controller
  participant V as Validation Webhook
  actor U as User

  P->>+A: Create Addon CR
  activate A
  A->>V: Enterprise CR Create Allowed?
  V-->>A: Create Allowed!
  A->>E: Create Enterprise CR
  activate E

  U->>V: Enterprise CR Delete Allowed?
  V--xU: Delete Denied
  U->>V: Enterprise CR Update Allowed?
  V--xU: Update Denied!
  U->>V: New Enterprise CR Create Allowed?
  V--xU: Create Denied!

  P->>A: Delete Addon CR
  A->>V: Enterprise CR Delete Allowed?
  V-->>A: Delete Allowed!
  A->>E: Delete Enterprise CR
  deactivate E
  deactivate A
```
