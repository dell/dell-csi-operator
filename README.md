# Dell Storage CSI Operator
Dell Storage CSI Operator is a Kubernetes native application which helps in installing and managing CSI Drivers provided by Dell EMC for its various storage platforms. The CSI Operator uses Kubernetes CRDs (Custom Resource Definitions) to define a manifest that describes the deployment specifications for each driver to be deployed. Multiple CSI drivers provided by Dell EMC and multiple instances of each driver can be deployed by the operator by defining a manifest for each deployment.

The CSI Operator is built using the [operator framework](https://github.com/operator-framework) and runs custom Kubernetes controllers to manage the driver installations. These controllers listen for any create/update/delete request for the respective CRDs and try to reconcile the request.

Currently the CSI Operator can be used to deploy the following CSI drivers provided by Dell EMC

* CSI Driver for Dell EMC PowerMax
* CSI Driver for Dell EMC Isilon
* CSI Driver for Dell EMC Unity
* CSI Driver for Dell EMC VxFlexOS

Additionally, the CSI Operator can also deploy storage classes and snapshot classes as part of the driver deployment.
The CSI Operator is itself installed as a Kubernetes deployment.

## Overview
The major steps in installation process are:

1. Installing the Operator from OperatorHub or manually using the installation scripts provided in this repository
2. Make sure that the pre-requisites for the drivers being installed are met. For e.g. - creation of namespace, secrets, installation of packages
3. Configuring the driver manifest and then install the driver using the manifest


## Installation
The CSI Operator is available on OperatorHub on upstream Kubernetes as well as Openshift clusters and can be installed using OLM (Operator Lifecycle Manager). It can also be deployed manually using the installation and configuration files available in this repository

### Pre-requisites
The CSI Operator has been tested and verified on 

    * Upstream Kubernetes cluster v1.14
    * Openshift Clusters 4.2 with RHEL 7.x worker nodes

#### OperatorHub
The CSI Operator requires a configmap to be created in the same namespace where the operator is deployed. This is an optional step but highly recommended.

Please run the following commands for creating the configmap
```
$ git clone github.com/dell/dell-csi-operator
$ cd dell-csi-operator
# Replace operator-namespace in the below command with the actual namespace where the operator is being deployed
$ kubectl create configmap config-csi-operator --from-file config/ -n <operator-namespace>
```
#### Manual Installation
There are no pre-requisites to be fulfilled before the manual installation of the operator

### Install the Operator
#### OperatorHub
The CSI Operator can be deployed via OperatorHub for both upstream Kubernetes installations as well as Openshift 4.2 clusters using OLM.

OLM is bundled with Openshift clusters and the CSI Operator is available as a community operator.

For upstream Kubernetes installations, OLM is not installed as a default component. Please follow the [instructions](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md) on how to add OLM to your upstream Kubernetes cluster. Once OLM has been installed and configured, the CSI Operator can be installed via OperatorHub.io

#### Manual Installation
For manual installations, an install script has been provided which will create the appropriate ServiceAccount, ClusterRole, ClusterRolebinding before deploying the operator. The operator will be deployed in the default namespace.

Note - If OLM is not configured in your upstream Kubernetes cluster, we recommend to use the manual installation method

```
# Clone this repository
$ git clone github.com/dell/dell-csi-operator
$ cd dell-csi-operator
# Make sure you are at the root of the cloned dell-csi-operator repository
$ bash scripts/deploy.sh
```
Post the installation, the operator should be deployed successfully in the default namespace. 
```
$ kubectl get deployment
```

##### Advanced configuration
By default, the CSI Operator will deploy one Kubernetes controller to manage each type of the CSI driver it can manage. If you would like to run only specific controllers, then you can modify the environment variable **OPERATOR_DRIVERS** in deploy/operator.yaml

```
            - name: OPERATOR_DRIVERS
              value: "unity,powermax,isilon,vxflexos"
```
For e.g. - If you only want to install the unity controller as you know that you are only going to install CSI Driver for Dell EMC Unity using the operator, then remove all other storage array types.  
Here is an example
```
            - name: OPERATOR_DRIVERS
              value: "unity"
```
Note - The above configuration is only supported in the manual installation method and you can't configure specific controllers in an OperatorHub installation. 

Once this configuration has been changed then the only way to add any more controllers for any specific storage array type is by redeploying the operator.

### Custom Resource Definitions
As part of the CSI Operator installation, a CRD representing each driver installation is also installed.  
List of CRDs which are installed in API Group `storage.dell.com`
* csipowermax
* csiunity
* csivxflexos
* csiisilon

### Driver manifest examples

#### OperatorHub
An example manifest for each driver is available in OperatorHub GUI after the operator has been installed.  
They can be accessed while trying to create a Custom Resource (CR)

#### Manual Installation
There are a few sample manifest files specified for each of the drivers in the samples folder. All the yaml files ending with the suffix _ops should be used for installation in an Openshift Cluster.

* sample/powermax.yaml - Sample manifest for CSI PowerMax v1.2 driver for upstream kubernetes cluster
* sample/powermax_ops.yaml - Sample manifest for CSI PowerMax v1.2 driver for Openshift 4.2 cluster
* sample/unity.yaml - Sample manifest for CSI Unity v1.1 driver for upstream kubernetes cluster
* sample/unity_ops.yaml - Sample manifest for CSI Unity v1.1 driver for Openshift 4.2 cluster
* sample/isilon.yaml - Sample manifest for CSI Isilon v1.1 driver for upstream kubernetes cluster
* sample/isilon_ops.yaml - Sample manifest for CSI Isilon v1.1 driver for Openshift 4.2 cluster
* sample/vxflexos.yaml - Sample manifest for CSI VxFlexOS v1.1.4 driver for upstream kubernetes cluster
* sample/vxflexos_ops.yaml - Sample manifest for CSI VxFlexOS v1.1.4 driver for Openshift 4.2 cluster


## Install CSI Drivers

For installing any CSI Driver, follow these steps in general. (Steps specific to each driver is documented in each of the driver sections)

### Fulfill any pre-requisites for the driver installation
These typically include (but not limited to) -
* Create a namespace for the driver installation
* Create a secret containing credentials for the storage array’s management interface
* Install any packages on nodes (if required)

Please follow the driver specific instructions to fulfill these requirements

### Custom Resource Specification
Each CSI Driver installation is represented by a Custom Resource.  

The specification for the Custom Resource is the same for all the drivers.   
Below is a list of all the mandatory and optional fields in the Custom Resource specification

#### Mandatory fields
**configVersion** - Configuration version  - Must be set to v1  
**replicas**  - Number of replicas for controller plugin - Must be set to 1 for all drivers  
**common**  
This field is mandatory and is used to specify common properties for both controller and the node plugin
* image - driver image
* imagePullPolicy - Image Pull Policy of the driver image
* envs - List of environment variables and their values
#### Optional fields
**controller** - List of environment variables and values which are applicable only for controller  
**node** - List of environment variables and values which are applicable only for node  
**sideCars** - optional - Specification for CSI sidecar containers.  
**storageClass**  
List of storage class specification

   1. name - name of the storage class
   2. default - Used to specify if the storage class will be marked as default (only set one storage class as default in a cluster)
   3. reclaimPolicy - Sets the Reclaim Policy for the PVCs. Defaults to Delete if not specified
   4. parameters - driver specific parameters. Refer individual driver section for more details

**snapshotclass**  
List of snapshot class specifications  

   1. name - name of the snapshot class
   2. parameters - driver specific parameters. Refer individual driver section for more details

**forceUpdate**  
Boolean value which must be set to `true` in order to update the Custom Resource if the status (state) of the Custom Resource is `Failed`


Here is a sample specification with annotated comments to explain each field
```
apiVersion: storage.dell.com/v1
kind: CSIPowerMax <- Type of the driver
metadata:
  name: test-powermax <- Name of the driver
  namespace: test-powermax <- Namespace where driver is installed
spec:
  driver:
    # Used to specify configuration version
    configVersion: v1 <- Set to v1. Shouldn't be changed
    replicas: 1 <- Always set to 1 for all drivers
    forceUpdate: false <- Set to true in case the installation failed
    common: <- All common specification
      image: "dellemc/csi-powermax:v1.2.0.000R" <- driver image
      imagePullPolicy: IfNotPresent
      envs:
        - name: X_CSI_POWERMAX_ENDPOINT
          value: "https://0.0.0.0:8443/"
        - name: X_CSI_K8S_CLUSTER_PREFIX
          value: "XYZ"
    sideCars:
      - name: snapshotter <- Installs snapshotter sidecar
    storageClass:
      - name: bronze
        default: true
        reclaimPolicy: Delete
        parameters:
          SYMID: "000000000001"
          SRP: DEFAULT_SRP
          ServiceLevel: Bronze
```
Note - The name of the storage class or the snapshot class (which are created in the Kubernetes/Openshift cluster) is created using the name of the driver and the name provided for these classes in the manifest. This is done in order to ensure name of these classes are unique if there are multiple drivers installed in the same cluster. For e.g. - With the above sample manifest, the name of the storage class which is created in the cluster will be `test-powermax-bronze`.  
You can get the name of the storageclass and snapshotclass created by the operator by running the commands - `kubectl get storageclass` and `kubectl get volumesnapshotclass`

### SideCars
Although the sidecars field in the driver specification is optional, it is **strongly** recommended to not modify any details related to sidecars provided (if present) in the sample manifests. Any modifications to this should be only done after consulting with Dell EMC support.

#### Snapshotter sidecar
All the CSI Drivers which can be installed by the CSI Operator don't support creating VolumeSnapshots in an Openshift cluster because the Snapshot feature is still in Technical Preview in Openshift 4.2 clusters.Because of this, the snapshotter sidecar is missing from the sample manifests (for Openshift) for the drivers. Make sure that the snapshotter sidecar name (and any other agruments) is not present in the driver manifest before installation in an Openshift cluster.

Similarly, in an upstream Kubernetes cluster, the snapshotter sidecar must be present in the driver manifest in order to use the snapshot functionality of the driver.

### Create Custom Resource manifest using example manifests
#### OperatorHub
Use the OperatorHub GUI to create a new manifest using the example manifest provided for the driver you wish to install

#### Manual installation
Copy the example manifest provided in the dell-csi-operator repository and use this to install the driver
For e.g. – Copy the PowerMax example manifest file
```
$ cp dell-csi-operator/sample/powermax.yaml .
```
##### Modify the driver specification
* Provide the namespace (in metadata section) where you want to install the driver
* Provide a name (in metadata section) for the driver. This will be the name of the Custom Resource
* Edit the values for mandatory configuration parameters specific to your installation
* Edit/Add any values for optional configuration parameters to customize your installation

### Create Custom Resource
#### OperatorHub
Use the OperatorHub GUI to create the Custom Resource once you have created the CR manifest

#### Manual installation
Create the custom resource using the following command
```
$ kubectl create -f powermax.yaml
```

### Verification
Once the driver Custom Resource has been created, you can verify the installation

*  Check if Driver CR is created successfully

    For e.g. – If you installed the PowerMax driver
    ```
    $ kubectl get csipowermax -n <driver-namespace>
    ```
* Check the status of the Custom Resource to verify if the driver was installed successfully

If the driver-namespace was set to test-powermax and the name of the driver is powermax, then run the command `kubectl get csipowermax/powermax -n test-powermax -o yaml` to get the details of the Custom Resource.  
Here is a sample output of the above command
```
apiVersion: storage.dell.com/v1
kind: CSIPowerMax
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"storage.dell.com/v1","kind":"CSIPowerMax","metadata":{"annotations":{},"name":"powermax","namespace":"test-powermax"},"spec":{"driver":{"common":{"envs":[{"name":"X_CSI_POWERMAX_ENDPOINT","value":"https://0.0.0.0:8443/"},{"name":"X_CSI_K8S_CLUSTER_PREFIX","value":"XYZ"}],"image":"dellemc/csi-powermax:v1.2.0.000R","imagePullPolicy":"IfNotPresent"},"configVersion":"v1","forceUpdate":false,"replicas":1,"sideCars":[{"name":"snapshotter"}],"snapshotClass":[{"name":"powermax-snapclass"}],"storageClass":[{"default":true,"name":"bronze","parameters":{"SRP":"SRP_1","SYMID":"000000000001","ServiceLevel":"Bronze"},"reclaimPolicy":"Delete"}]}}}
  creationTimestamp: "2020-04-06T09:56:02Z"
  generation: 2
  name: powermax
  namespace: test-powermax
  resourceVersion: "10225198"
  selfLink: /apis/storage.dell.com/v1/namespaces/test-powermax/csipowermaxes/powermax
  uid: d31c2fba-77ec-11ea-abc1-005056a3aee3
spec:
  driver:
    common:
      envs:
      - name: X_CSI_POWERMAX_ENDPOINT
        value: https://0.0.0.0:8443/
      - name: X_CSI_K8S_CLUSTER_PREFIX
        value: XYZ
      image: dellemc/csi-powermax:v1.2.0.000R
      imagePullPolicy: IfNotPresent
    configVersion: v1
    controller: {}
    node: {}
    replicas: 1
    sideCars:
    - image: quay.io/k8scsi/csi-snapshotter:v1.2.2
      imagePullPolicy: IfNotPresent
      name: snapshotter
    - image: quay.io/k8scsi/csi-provisioner:v1.2.1
      imagePullPolicy: IfNotPresent
      name: provisioner
    - image: quay.io/k8scsi/csi-attacher:v1.1.1
      imagePullPolicy: IfNotPresent
      name: attacher
    - image: quay.io/k8scsi/csi-node-driver-registrar:v1.1.0
      imagePullPolicy: IfNotPresent
      name: registrar
    snapshotClass:
    - name: powermax-snapclass
    storageClass:
    - default: true
      name: bronze
      parameters:
        SRP: SRP_1
        SYMID: "000000000001"
        ServiceLevel: Bronze
      reclaimPolicy: Delete
status:
  status:
    creationTime: "2020-04-06T09:56:02Z"
    envName: powermax
    startTime: "2020-04-06T09:56:02Z"
    state: Succeeded

```

* Driver statefulset & daemonset should be created in this namespace
    ```
    $ kubectl get statefulset -n <driver-namespace>
    $ kubectl get daemonset -n <driver-namespace>
    ```
## Update CSI Drivers

The CSI Drivers installed by the CSI Operator can be updated like any Kubernetes resource. This can be achieved in various ways which include –

* Modifying the original CR manifest file (used to deploy the driver) and running a `kubectl apply` command

    For e.g. - Modify the unity.yaml used to install the Unity driver and run

    ```
    $ kubectl apply -f unity.yaml
    ```
* Modifying the installation directly via `kubectl edit`
    For e.g. - If the name of the installed unity driver is unity, then run
    ```
    # Replace driver-namespace with the namespace where the Unity driver is installed
    $ kubectl edit csiunity/unity -n <driver-namespace>
    ```
    and modify the installation
* Modify the API object in-place via `kubectl patch`

### Supported modifications
* Changing environment variable values for driver
* Adding (supported) environment variables
* Updating the image of the driver

### Unsupported modifications
Kubernetes doesn’t allow to update a storage class once it has been created. Any attempt to update a storage class will result in a failure.

Note – Any attempt to rename a storage class or snapshot class will result in the deletion of older class and creation of a new class.

## Uninstall CSI Drivers
For uninstalling any CSI drivers deployed the CSI Operator, just delete the respective Custom Resources.  
This can be done using OperatorHub GUI by deleting the CR or via kubectl.
    
For e.g. – To uninstall a vxflexos driver installed via the operator, delete the Custom Resource(CR)

```
# Replace driver-name and driver-namespace with their respective values
$ kubectl delete vxflexos/<driver-name> -n <driver-namespace>
```

## Limitations
* The CSI Operator can't manage any existing driver installed using Helm charts. If you already have installed one of the DellEMC CSI driver in your cluster and  want to use the operator based deployment, uninstall the driver and then redeploy the driver following the installation procedure described above
* The CSI Operator can't update storage classes as it is prohibited by Kubernetes. Any attempt to do so will cause an error and the driver Custom Resource will be left in a `Failed` state. Refer the Troubleshooting section to fix the driver CR.
* The CSI Operator is not fully compliant with the OperatorHub React UI elements and some of the Custom Resource fields may show up as invalid or unsupported in the OperatorHub GUI. To get around this problem, use kubectl/oc commands to get details about the Custom Resource(CR). This issue will be fixed in the upcoming releases of the CSI Operator


## Troubleshooting
* Before installing the drivers, CSI Operator tries to validate the Custom Resource being created. If some mandatory environment variables is missing or there is a type mismatch, then the Operator will report an error during the reconciliation attempts.  
Because of this, the status of the Custom Resource will change to "Failed" and the error captured in the "ErrorMessage" field in the status.  
For e.g. - If the PowerMax driver was installed in the namespace test-powermax and has the name powermax, then run the command `kubectl get csipowermax/powermax -n test-powermax -o yaml` to get the Custom Resource details.  
If there was an error while installing the driver, then you would see a status like this -
  ```
  status:
    status:
      creationTime: "2020-04-06T09:51:17Z"
      envName: powermax
      errorMessage: mandatory Env - X_CSI_K8S_CLUSTER_PREFIX not specified in user spec
      startTime: "2020-04-06T09:51:17Z"
      state: Failed
  ```
    Any further updates to the Custom Resource are rejected as long as the status of the Custom Resource is `Failed`. This is done to avoid endless retries and avoid overloading the kube-api server.  

    The state of the Custom Resource can also change to `Failed` because of any other prohibited updates or any failure while installing the driver. In order to recover from this failure, there are two options -

    1. Delete the Custom Resource and create it again after correcting any mistakes
    2.  Add the `forceUpdate` boolean field to the Custom Resource and set it to `true`. This will force an update of the Custom Resource regardless of the failed state. Once the Custom Resource is updated, the CSI Operator sets this field to `false` automatically.

* After an update to the driver, the controller pod may not have the latest desired specification  
The above happens when the controller pod was in a failed state before applying the update. Even though the CSI Operator updates the pod template specification for the StatefulSet, the StatefulSet controller does not apply the update to the pod. This happens because of the unique nature of StatefulSets where the controller tries to retain the last known working state. 

    To get around this problem, the CSI Operator forces an update of the pod specification by deleting the older pod. In case the CSI Operator fails to do so, delete the controller pod to force an update of the controller pod specification

# Driver details
## CSI Isilon
### Pre-requisites
#### Create secret to store Isilon credentials
Create a secret named `isilon-creds`, in the namespace where the CSI Isilon driver will be installed, using the following manifest
```
apiVersion: v1
kind: Secret
metadata:
  name: isilon-creds
  # Replace driver-namespace with the namespace where driver is being deployed
  namespace: <driver-namespace>
type: Opaque
data:
  # set username to the base64 encoded username
  username: <base64 username>
  # set password to the base64 encoded password
  password: <base64 password>
```
The base64 username and password can be obtained by running the following commands
```
# If myusername is the username
echo -n "myusername" | base64
# If mypassword is the password
echo -n "mypassword" | base64
```
#### Optional - Create secret for client side TLS verification
Create a secret named `isilon-certs` in the namespace where the CSI Isilon driver will be installed. This is an optional step and is only required if you are setting the env variable `X_CSI_ISI_INSECURE` to `false`. Please refer detailed documentation on how to create this secret in the Product Guide [here](https://github.com/dell/csi-isilon)

### Set the following *Mandatory* Environment variables
| Variable name      | Section | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_ISI_ENDPOINT | common    | HTTPS endpoint of the Isilon OneFS API server | 1.1.1.1 |
| X_CSI_ISI_PORT     | common    | HTTPS port number of the Isilon OneFS API server (string) | 8080 |

### Modify/Set the following *optional* environment variables
| Variable name      | Section    | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_VERBOSE      | common | Indicates what content of the OneFS REST API message should be logged in debug level logs (string) | 1 |
| X_CSI_ISI_PATH     | common | The default base path for the volumes to be created, this will be used if a storage class does not have the IsiPath parameter specified | /ifs/data/csi |
| X_CSI_ISILON_NO_PROBE_ON_START | common | Indicates whether a probe should be attempted upon start (string) | false |
| X_CSI_ISI_AUTO_PROBE | common | Indicates whether the controller service should be automatically probed (string) | true |
| X_CSI_ISI_INSECURE | common | Indicates the certificate should not or should be verified (string) | true |
| X_CSI_DEBUG | common | Indicates whether the driver is in debug mode (string) | false |
| X_CSI_ISI_QUOTA_ENABLED | controller | Indicates whether the provisioner should attempt to set (later unset) quota on a newly provisioned volume | true |
| X_CSI_ISI_ACCESSZONE | controller | The default name of the access zone a volume can be created in, this will be used if a storage class does not have the AccessZone parameter specified | System |
| X_CSI_ISILON_NFS_V3 | node | Indicates whether to add "-o ver=3" option to the mount command when mounting an NFS export (string) | false |

### StorageClass parameters
| Name | Mandatory | Description | Example |
| ---- | :-------: | ----------- | ------- |
| AccessZone | no | The name of the access zone a volume can be created in through this storage class | System |
| IsiPath | no | The base path for the volumes to be created through this storage class | /ifs/data/csi |
| AzServiceIP | no | Access zone service IP. Need to specify if different from X_CSI_ISI_ENDPOINT, also it can be the same as X_CSI_ISI_ENDPOINT | 1.1.1.1 |
| RootClientEnabled | no | Indicates when a node mounts the PVC, in NodeStageVolume, whether to add the k8s node to the "Root clients" field (when true) or "Clients" field (when false) of the NFS export (string) | false |

### SnapshotClass parameters

| Name | Mandatory | Description | Example |
| ---- | :-------: | ----------- | ------- |
| IsiPath | no | The base path for the volumes to create snapshots, and the value should match with the IsiPath of respective storage class | /ifs/data/csi |

## CSI Unity
### Pre-requisites
#### Create secret to store Unity credentials

Create a secret named `unity-creds`, in the namespace where the CSI Unity driver will be installed, using the following manifest

```
apiVersion: v1
kind: Secret
metadata:
  name: unity-creds
  # Replace driver-namespace with the namespace where driver is being deployed
  namespace: <driver-namespace>
type: Opaque
data:
  # set username to the base64 encoded username
  username: <base64 username>
  # set password to the base64 encoded password
  password: <base64 password>
```
The base64 username and password can be obtained by running the following commands
```
# If myusername is the username
echo -n "myusername" | base64
# If mypassword is the password
echo -n "mypassword" | base64
```
#### Optional - Create secret for client side TLS verification
Create a secret named `unity-certs` in the namespace where the CSI Unity driver will be installed. This is an optional step and is only required if you are setting the env variable `X_CSI_UNITY_INSECURE` to `false`. Please refer detailed documentation on how to create this secret in the Product Guide [here](https://github.com/dell/csi-unity)


### Set the following *Mandatory* Environment variables

| Variable name      | Section | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_UNITY_ENDPOINT | common | Must provide a UNITY HTTPS unisphere url | https://127.0.0.1:443 |

### Modify/Set the following *optional* environment variables

| Variable name      | Section    | Description | Default |
| ------------------ | ---------- | ----------- | ------- |
| X_CSI_DEBUG | common | To enable debug mode | false |
| X_CSI_UNITY_INSECURE | common | Specifies that the Unity's hostname and certificate chain | true |
| GOUNITY_DEBUG | common | To enable debug mode for gounity library | false |

### StorageClass Parameters
Following parameters are not present in values.yaml in the Helm based installer

| Parameter | Description | Required | Default |
| --------- | ----------- | -------- |-------- |
| FsType | To set File system type. Possible values are ext3,ext4,xfs | false | ext4 |
| volumeThinProvisioned | To set volume thinProvisioned | false | true |
| isVolumeDataReductionEnabled | To set volume data reduction | false | false |
| volumeTieringPolicy | To set volume tiering policy | false | 0 |
| hostIOLimitName | To set unity host IO limit | false | "" |

### SnapshotClass parameters
Following parameters are not present in values.yaml in the Helm based installer

| Parameter | Description | Required | Default |
| --------- | ----------- | -------- |-------- |
| snapshotRetentionDuration | TO set snapshot retention duration. Format:"1:23:52:50" (number of days:hours:minutes:sec)| false | "" |

## CSI VxFlexOS
### Pre-requisites

#### Create secret to store VxFlexOS credentials
Create a secret named `vxflexos-creds`,  in the namespace where the CSI VxFlexOS driver will be installed, using the following manifest
```
apiVersion: v1
kind: Secret
metadata:
  name: vxflexos-creds
  # Replace driver-namespace with the namespace where driver is being deployed
  namespace: <driver-namespace>
type: Opaque
data:
  # set username to the base64 encoded username
  username: <base64 username>
  # set password to the base64 encoded password
  password: <base64 password>
```
The base64 username and password can be obtained by running the following commands
```
# If myusername is the username
echo -n "myusername" | base64
# If mypassword is the password
echo -n "mypassword" | base64
```

#### Install VxFlex OS Storage Data Client
Install the VxFlex OS Storage Data Client (SDC) on all Kubernetes nodes.  
For detailed VxFlex OS installation procedure, and current version of the driver see the [Dell EMC VxFlex OS Deployment Guide.](https://github.com/dell/csi-vxflexos/blob/master/CSI%20Driver%20for%20VxFlex%20OS%20Product%20Guide.pdf)


Procedure:
1. Download the VxFlex OS SDC from Dell EMC Online support. The filename is EMC-ScaleIO-sdc-*.rpm, where * is the SDC name corresponding to the VxFlex OS installation version.
2. Export the shell variable MDM_IP in a comma-separated list. This list contains the IP addresses of the MDMs.
    export MDM_IP=xx.xxx.xx.xx,xx.xxx.xx.xx, where xxx represents the actual IP address in your environment variable.
3. Install the SDC using the following commands:
l For Red Hat Enterprise Linux and Cent OS, run rpm -iv ./EMC-ScaleIO-sdc-*.x86_64.rpm, where * is the SDC name corresponding to the VxFlex OS installation version.
l For Ubuntu, run EMC-ScaleIO-sdc-3.0-0.769.Ubuntu.18.04.x86_64.deb.

### Set the following *Mandatory* Environment variables

| Variable name      | Section | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_VXFLEXOS_SYSTEMNAME | common | defines the name of the VxFlex OS system from which volumes will be provisioned. This must either be set to the VxFlex OS system name or system ID | systemname |
| X_CSI_VXFLEXOS_ENDPOINT | common | defines the VxFlex OS REST API endpoint, with full URL, typically leveraging HTTPS. You must set this for your VxFlex OS installations REST gateway | https://127.0.0.1 |
|

### Modify/Set the following **optional environment variables**


| Variable name      | Section    | Description | Example |
| ------------------ | ---------- | ----------- | ------- |
| X_CSI_DEBUG | common | To enable debug mode | false |
| X_CSI_VXFLEXOS_ENABLELISTVOLUMESNAPSHOT | common |  Enable list volume operation to include snapshots (since creating a volume from a snap actually results in a new snap) | false |
| X_CSI_VXFLEXOS_ENABLESNAPSHOTCGDELETE | common | Enable this to automatically delete all snapshots in a consistency group when a snap in the group is deleted | false |

### StorageClass parameters

| Name | Mandatory | Description | Example |
| ---- | :-------: | ----------- | ------- |
| storagePool | yes | defines the VxFlex OS storage pool from which this driver will provision volumes. You must set this for the primary storage pool to be used | sp |
| FsType | No  | To set File system type. Possible values are ext3,ext4,xfs | xfs |

## CSI PowerMax
### Pre-requisites
#### Create secret to store Unisphere for PowerMax credentials
Create a secret named `powermax-creds`,  in the namespace where the CSI PowerMax driver will be installed, using the following manifest
```
apiVersion: v1
kind: Secret
metadata:
  name: powermax-creds
  # Replace driver-namespace with the namespace where driver is being deployed
  namespace: <driver-namespace>
type: Opaque
data:
  # set username to the base64 encoded username
  username: <base64 username>
  # set password to the base64 encoded password
  password: <base64 password>
```
The base64 username and password can be obtained by running the following commands
```
# If myusername is the username
echo -n "myusername" | base64
# If mypassword is the password
echo -n "mypassword" | base64
```
#### Optional - Create secret for client side TLS verification
Create a secret named `powermax-certs` in the namespace where the CSI PowerMax driver will be installed. This is an optional step and is only required if you are setting the env variable `X_CSI_POWERMAX_SKIP_CERTIFICATE_VALIDATION` to `false`. Please refer detailed documentation on how to create this secret in the Product Guide [here](https://github.com/dell/csi-powermax)

#### Node requirements
For node specific requirements, please refer detailed instructions in the Product Guide [here](https://github.com/dell/csi-powermax "CSI PowerMax")

### Set the following *Mandatory* Environment variables

| Variable name      | Section | Description | Example |
| ------------------ | ------ | -----------  | ------- |
| X_CSI_POWERMAX_ENDPOINT | common | IP address of the Unisphere for PowerMax | https://0.0.0.0:8443 |
| X_CSI_K8S_CLUSTER_PREFIX | common | defines a prefix that is appended onto all resources created in the Array; unique per K8s/CSI deployment; max length - 3 characters | XYZ |

### Modify/Set the following *Optional* environment variables

| Variable name      | Section    | Description | Example |
| ------------------ | ---------- | ----------- | ------- |
| X_CSI_POWERMAX_DEBUG | common | determines if HTTP Request/Response is logged | "false" |
| X_CSI_POWERMAX_SKIP_CERTIFICATE_VALIDATION | common | skip client side TLS verification of Unisphere certificates | "true" |
| X_CSI_POWERMAX_PORTGROUPS | common | List of comma separated port groups (ISCSI only) | "PortGroup1,PortGroup2" |
| X_CSI_POWERMAX_ARRAYS | common | list of comma separated array id(s) which will be managed by the driver. If left blank, driver will manage all arrays managed by Unisphere for PowerMax | "000000000001,000000000002" |
| X_CSI_TRANSPORT_PROTOCOL | common | preferred transport protocol - FC/FIBRE or ISCSI/iSCSI. If left blank, driver will autoselect | "FC" |
| X_CSI_ENABLE_BLOCK | common | enable Block Volume capability which is in experimental phase | "true" |

### StorageClass parameters

| Name          | Mandatory | Description | Example |
| ------------- | :-------: | ----------- | ------- |
| SYMID         | yes       | Symmetrix ID | 000000000001 |
| SRP           | yes       | Storage Resource Pool name | DEFAULT_SRP|
| ServiceLevel  | no        | Service Level | Bronze |

### SnapshotClass parameters
There are no required parameters for snapshot class for PowerMax
