- [Dell CSI Operator](#dell-csi-operator)
  - [Support](#support)
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
* CSI Driver for Dell Unity XT
* CSI Driver for Dell PowerFlex (formerly VxFlex OS)
* CSI Driver for Dell PowerStore

Additionally, the Dell CSI Operator can also deploy Storage Classes and Volume Snapshot Classes as part of the driver deployment.
The Dell CSI Operator is itself installed as a Kubernetes deployment.

**NOTE**: You can refer to additional information about the Dell CSI Operator on the new documentation website [here](https://dell.github.io/csm-docs/docs/csidriver/installation/operator/)

## Support
The Dell CSI Operator image is available on Dockerhub and is officially supported by Dell.
For any CSI operator and driver issues, questions or feedback, join the [Dell EMC Container community](https://www.dell.com/community/Containers/bd-p/Containers).

## Supported Platforms
Dell CSI Operator has been tested and qualified with 

    * Upstream Kubernetes cluster v1.22, v1.23, v1.24
    * OpenShift Clusters 4.9, 4.10 with RHEL 7.x & RHCOS worker nodes

## Installation
To install Dell CSI Operator please refer the steps given here at [https://dell.github.io/csm-docs/docs/csidriver/installation/operator/](https://dell.github.io/csm-docs/docs/csidriver/installation/operator/)

## Upgrading Dell CSI Operator
To upgrade the driver to the latest version (across supported Kubernetes/OpenShift versions), please refer [https://dell.github.io/csm-docs/docs/csidriver/upgradation/drivers/operator/](https://dell.github.io/csm-docs/docs/csidriver/upgradation/drivers/operator/)

## Install CSI Drivers
To install CSI drivers using operator please refer here at [https://dell.github.io/csm-docs/docs/csidriver/installation/operator/#installing-csi-driver-via-operator](https://dell.github.io/csm-docs/docs/csidriver/installation/operator/#installing-csi-driver-via-operator)

## Uninstall CSI Drivers
To uninstall CSI drivers installed using operator please refer here at [https://dell.github.io/csm-docs/docs/csidriver/uninstall/#uninstall-a-csi-driver-installed-via-dell-csi-operator](https://dell.github.io/csm-docs/docs/csidriver/uninstall/#uninstall-a-csi-driver-installed-via-dell-csi-operator)
