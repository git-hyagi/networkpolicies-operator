# About this repository
I've created this project to study and learn golang and the operators framework.

## The networkpolicies-operator
The project contains a Kubernetes operator that creates 2 network policies on the namespaces configured in the custom resource (`forcenetpols.spec.projects`):
* allow-from-same-namespace 
* deny-by-default

These two network policies together provide the same behavior as using the ovs-multitentant sdn plugin and they were created based on [this document](https://docs.openshift.com/container-platform/3.11/admin_guide/managing_networking.html#admin-guide-networking-networkpolicy).

It is developed using the [operator-framework](https://operatorframework.io/).

# Pre-requisites
- A cluster-admin access to a running OpenShift 4.x cluster
- Access to the container registry used by the ocp cluster
- git to clone this repo
- The operator-sdk just to build this project (WIP push a working image to a public registry)
- docker/podman/buildah to pull the built image to the registry (WIP push a working image to a public registry)


# Installing the networkpolicies-operator
* Clone the networkpolicies-operator repo
~~~
git clone git@github.com:git-hyagi/networkpolicies-operator.git
~~~

* Build it
~~~
cd networkpolicies-operator
operator-sdk build <registry address>/network-policies-operator/forcenetpol:v1
~~~

* Push the built image to the registry
~~~
docker push <registry address>/network-policies-operator/forcenetpol:v1
~~~

* Create the cluster objects
~~~
oc new-project network-policies-operator
oc create -f  deploy/crds/lab.local_forcenetpols_crd.yaml 
oc create -f  deploy/service_account.yaml
oc create -f  deploy/role.yaml
oc create -f  deploy/rolebinding.yaml
oc create -f  deploy/operator.yaml
~~~

* Create a **forcenetpol** `custom resource` with the projects that should have the `network policies`
~~~
cat<<EOF> lab.local_v1_forcenetpol_cr.yaml
apiVersion: lab.local/v1
kind: ForceNetPol
metadata:
  name: forcenetpol
spec:
  projects:
  - <my project A>
  - <my project B>
EOF

oc create -f  lab.local_v1_forcenetpol_cr.yaml 
~~~

# Configuring the labels-operator
For now, the only configuration available is to add or remove projects from the **forcenetpol** `custom resource` which will, in turn, sync or not the `network policies` (*allow-from-same-namespace* and *deny-by-default*):
~~~
oc edit forcenetpol forcenetpol
~~~
