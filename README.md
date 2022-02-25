- [Dell CSI Operator](#dell-csi-operator)
  - [Supported Platforms](#supported-platforms)
  - [Installation](#installation)
  - [Upgrading Dell CSI Operator](#upgrading-dell-csi-operator)
  - [Install CSI Drivers](#install-csi-drivers)
  - [Uninstall CSI Drivers](#uninstall-csi-drivers)

# Dell CSI Operator
Dell CSI Operator is a Kubernetes native application which helps in installing and managing CSI Drivers provided by Dell for its various storage platforms. 
Dell CSI Operator uses Kubernetes CRDs (Custom Resource Definitions) to define a manifest that describes the deployment specifications for each driver to be deployed. Multiple CSI drivers provided by Dell and multiple instances of each driver can be deployed by the operator by defining a manifest for each deployment.

Dell CSI Operator is built using the [operator framework](https://github.com/operator-framework) and runs custom Kubernetes controllers to manage the driver installations. These controllers listen for any create/update/delete request for the respective CRDs and try to reconcile the request.

Currently, the Dell CSI Operator can be used to deploy the following CSI drivers provided by Dell 

* CSI Driver for Dell PowerMax
* CSI Driver for Dell PowerScale
* CSI Driver for Dell Unity
* CSI Driver for Dell PowerFlex (formerly VxFlex OS)
* CSI Driver for Dell PowerStore

Additionally, the Dell CSI Operator can also deploy Storage Classes and Volume Snapshot Classes as part of the driver deployment.
The Dell CSI Operator is itself installed as a Kubernetes deployment.

**NOTE**: You can refer to additional information about the Dell CSI Operator on the new documentation website [here](https://dell.github.io/csm-docs/docs/csidriver/installation/operator/)

## Table of Contents

* [Code of Conduct](https://github.com/dell/csm/blob/main/docs/CODE_OF_CONDUCT.md)
* [Maintainer Guide](https://github.com/dell/csm/blob/main/docs/MAINTAINER_GUIDE.md)
* [Committer Guide](https://github.com/dell/csm/blob/main/docs/COMMITTER_GUIDE.md)
* [Contributing Guide](https://github.com/dell/csm/blob/main/docs/CONTRIBUTING.md)
* [Branching Strategy](https://github.com/dell/csm/blob/main/docs/BRANCHING.md)
* [List of Adopters](https://github.com/dell/csm/blob/main/docs/ADOPTERS.md)
* [Maintainers](https://github.com/dell/csm/blob/main/docs/MAINTAINERS.md)
* [Support](https://dell.github.io/csm-docs/docs/support/)
* [Security](https://github.com/dell/csm/blob/main/docs/SECURITY.md)

## Supported Platforms
Dell CSI Operator has been tested and qualified with 

    * Upstream Kubernetes cluster v1.21, v1.22, v1.23
    * OpenShift Clusters 4.8, 4.8 EUS, 4.9 with RHEL 7.x & RHCOS worker nodes

## Installation
To install Dell CSI Operator please refer the steps given here at [https://dell.github.io/csm-docs/docs/csidriver/installation/operator/](https://dell.github.io/csm-docs/docs/csidriver/installation/operator/)

## Upgrading Dell CSI Operator
To upgrade the driver to the latest version (across supported Kubernetes/OpenShift versions), please refer [https://dell.github.io/csm-docs/docs/csidriver/upgradation/drivers/operator/](https://dell.github.io/csm-docs/docs/csidriver/upgradation/drivers/operator/)

## Install CSI Drivers
To install CSI drivers using operator please refer here at [https://dell.github.io/csm-docs/docs/csidriver/installation/operator/#installing-csi-driver-via-operator](https://dell.github.io/csm-docs/docs/csidriver/installation/operator/#installing-csi-driver-via-operator)

## Uninstall CSI Drivers
To uninstall CSI drivers installed using operator please refer here at [https://dell.github.io/csm-docs/docs/csidriver/uninstall/#uninstall-a-csi-driver-installed-via-dell-csi-operator](https://dell.github.io/csm-docs/docs/csidriver/uninstall/#uninstall-a-csi-driver-installed-via-dell-csi-operator)
