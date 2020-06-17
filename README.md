# About this repository
I've created this project to study and learn golang and the operators framework.

## The networkpolicy-operator
The project contains a Kubernetes operator that creates 2 network policies on the namespaces configured in the custom resource (`forcenetpols.spec.projects`):
* allow-from-same-namespace 
* deny-by-default

These two network policies together provide the same behavior as using the ovs-multitentant sdn plugin and they were created based on [this document](https://docs.openshift.com/container-platform/3.11/admin_guide/managing_networking.html#admin-guide-networking-networkpolicy).

It is developed using the [operator-framework](https://operatorframework.io/).

# Pre-requisites
[WIP]

# Installing the labels-operator
[WIP]

# Configuring the labels-operator
[WIP]
