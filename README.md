# Dell CSI Operator
Dell CSI Operator is a Kubernetes native application which helps in installing and managing CSI Drivers provided by Dell EMC for its various storage platforms. 
Dell CSI Operator uses Kubernetes CRDs (Custom Resource Definitions) to define a manifest that describes the deployment specifications for each driver to be deployed. Multiple CSI drivers provided by Dell EMC and multiple instances of each driver can be deployed by the operator by defining a manifest for each deployment.

Dell CSI Operator is built using the [operator framework](https://github.com/operator-framework) and runs custom Kubernetes controllers to manage the driver installations. These controllers listen for any create/update/delete request for the respective CRDs and try to reconcile the request.

Currently, the Dell CSI Operator can be used to deploy the following CSI drivers provided by Dell EMC

* CSI Driver for Dell EMC PowerMax
* CSI Driver for Dell EMC PowerScale
* CSI Driver for Dell EMC Unity
* CSI Driver for Dell EMC PowerFlex (formerly VxFlex OS)
* CSI Driver for Dell EMC PowerStore

Additionally, the Dell CSI Operator can also deploy Storage Classes and Volume Snapshot Classes as part of the driver deployment.
The Dell CSI Operator is itself installed as a Kubernetes deployment.

## Support
The Dell CSI Operator image is available on Dockerhub and is officially supported by Dell EMC.
For any CSI operator and driver issues, questions or feedback, join the [Dell EMC Container community](https://www.dell.com/community/Containers/bd-p/Containers).

## Overview
The major steps in installation process are:

1. Installing the Operator from OperatorHub or manually using the installation scripts provided in this repository.
2. Ensure pre-requisites for the drivers are met. For e.g. - creation of namespace, secrets, installation of packages.
3. Configuring the driver manifest and then install the driver using the manifest.

## Before Installation
`Dell CSI Operator` was previously available, with the name `CSI Operator`, for both manual and OLM installation.  
`CSI Operator` has been discontinued and has been renamed to `Dell CSI Operator`.  This is just a name change and as a result,
the Kubernetes resources created as part of the Operator deployment will use the name `dell-csi-operator` instead of `csi-operator`.

Before proceeding with the installation of the new `Dell CSI Operator`, any existing `CSI Operator` installation has to be completely 
removed from the cluster.

Note - This **doesn't** impact any of the CSI Drivers which have been installed in the cluster

If the old `CSI Operator` was installed manually, then run the following command from the root of the repository which was used 
originally for installation

    bash scripts/undeploy.sh

If you don't have the original repository available, then run the following commands

    git clone https://github.com/dell/dell-csi-operator.git
    cd dell-csi-operator
    git checkout csi-operator-v1.0.0
    bash scripts/undeploy.sh

Note - Once you have removed the old `CSI Operator`, then for installing the new `Dell CSI Operator`, you will need to pull/checkout the latest code

If you had installed old CSI Operator using OLM, then please follow un-installation instructions provided by OperatorHub. This will mostly involve:

    * Deleting the CSI Operator Subscription  
    * Deleting the CSI Operator CSV  

## Installation
`Dell CSI Operator` is available on OperatorHub on upstream Kubernetes as well as OpenShift clusters and can be installed using OLM (Operator Lifecycle Manager). It can also be deployed manually using the installation and configuration files available in this repository


### Pre-requisites
Dell CSI Operator has been tested and qualified with 

    * Upstream Kubernetes cluster v1.17, v1.18, v1.19
    * OpenShift Clusters 4.3, 4.4 with RHEL 7.x & RHCOS worker nodes

#### Installation using OperatorHub
Dell CSI Operator requires a ConfigMap to be created in the same namespace where the operator is deployed. This is an optional step but highly recommended.

Please run the following commands for creating the ConfigMap
```
$ git clone github.com/dell/dell-csi-operator
$ cd dell-csi-operator
$ tar -czf config.tar.gz config/
# Replace operator-namespace in the below command with the actual namespace where the operator is being deployed
$ kubectl create configmap config-dell-csi-operator --from-file config.tar.gz -n <operator-namespace>
```

#### Manual Installation
For upstream k8s clusters, make sure to install

    Beta VolumeSnapshot CRDs (can be installed using the Operator installation script)
    External Volume Snapshot Controller

### Install the Operator

### A note about Operator upgrade
Dell CSI Operator v1.1 can only manage driver installations which do not support alpha VolumeSnapshots. It can only create/manage beta VolumeSnapshotClass objects.  
This restriction is because of the breaking changes in the migration of VolumeSnapshot APIs from `v1alpha1` to `v1beta`  
Because of these restrictions, you shouldn't upgrade the Operator from an older release.  A suggested upgrade path could be:

* If you created any VolumeSnapshotClass along with your driver installation, then
    Delete the driver installation
* If you have any alpha VolumeSnapshot CRDs or CRs present in your cluster, delete them
* If required, upgrade your cluster to a supported version
* If `dell-csi-operator` was installed using OLM, then you can now upgrade it
* Only for upstream clusters - Run `install.sh` script with the option `--upgrade`. 
* Optionally, if you wish to install beta VolumeSnapshot CRDs, then use option `--snapshot-crd` while running the `install.sh` script

#### OperatorHub
Dell CSI Operator can be deployed via OperatorHub for both upstream Kubernetes installations and OpenShift 4.3, 4.4 clusters using OLM.

OLM is bundled with OpenShift clusters and the Dell CSI Operator is available as a community operator.

For upstream Kubernetes installations, OLM is not available as a default component. Please follow the [instructions](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md) on how to add OLM to your upstream Kubernetes cluster.  
Once OLM has been installed and configured, the Dell CSI Operator can be installed via OperatorHub.io

#### Manual Installation
For manual installations, an install script - `install.sh` has been provided which will create the appropriate ServiceAccount, ClusterRole, ClusterRolebinding before deploying the operator. The operator will be deployed in the default namespace.

Note - The install script `install.sh` checks if you have any alpha VolumeSnapshot CRDs or corresponding CRs present in your cluster and fails if it finds any.  
Make sure to completely remove all alpha VolumeSnapshot CRs and CRDs before proceeding with the installation.

Note - If OLM has not configured in your upstream Kubernetes cluster, we recommend using the manual installation method

```
# Clone this repository
$ git clone github.com/dell/dell-csi-operator
$ cd dell-csi-operator
# Make sure you are at the root of the cloned dell-csi-operator repository
$ bash scripts/install.sh
```
Post the installation, the operator should be deployed successfully in the default namespace. 
```
$ kubectl get deployment
```

Note - If you wish to install the VolumeSnapshot beta CRDs (release 2.1) in the cluster, then run the command - `bash scripts/install.sh --snapshot-crd`

##### Advanced configuration
By default, the Dell CSI Operator will deploy one Kubernetes controller to manage each type of the CSI driver it can manage. If you would like to run only specific controllers, then you can modify the environment variable **OPERATOR_DRIVERS** in deploy/operator.yaml

```
            - name: OPERATOR_DRIVERS
              value: "unity,powermax,isilon,vxflexos,powerstore"
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
As part of the Dell CSI Operator installation, a CRD representing each driver installation is also installed.  
List of CRDs which are installed in API Group `storage.dell.com`
* csipowermax
* csiunity
* csivxflexos
* csiisilon
* csipowerstore
* csipowermaxrevproxy

### Driver manifest examples

#### OperatorHub
An example manifest for each driver is available in OperatorHub GUI after the operator has been installed.  
They can be accessed while trying to create a Custom Resource (CR)

#### Manual Installation
A lot of sample manifest files have been provided in the samples folder to help with the installation of various CSI Drivers  
They follow the naming convention

    {driver name}_{driver version}_k8s_{k8 version}.yaml

Or

    {driver name}_{driver version}_ops_{OpenShift version}.yaml

Use the correct sample manifest based on the driver, driver version and Kubernetes/OpenShift version


For e.g.  
*sample/powermax_v140_k8s_117.yaml* <- To install CSI PowerMax driver v1.4.0 on a Kubernetes 1.17 cluster  
*sample/powermax_v140_ops_43.yaml* <- To install CSI PowerMax driver v1.4.0 on an OpenShift 4.3 cluster


## Install CSI Drivers

### Full list of CSI Drivers and versions supported by the Dell CSI Operator
| CSI Driver         | Version | ConfigVersion | Kubernetes Version       | OpenShift Version |
| ------------------ | ------  | --------------| ------------------------ | ----------------- |
| CSI PowerMax       | 1.4     | v3            | 1.17, 1.18, 1.19         | 4.3, 4.4          |
| CSI PowerFlex      | 1.2     | v2            | 1.17, 1.18, 1.19         | 4.3, 4.4          |
| CSI PowerScale     | 1.3     | v3            | 1.17, 1.18, 1.19         | 4.3, 4.4          |
| CSI Unity          | 1.3     | v2            | 1.17, 1.18, 1.19         | 4.3, 4.4          |
| CSI PowerStore     | 1.1     | v1            | 1.17, 1.18, 1.19         | 4.3, 4.4          |

For installing any CSI Driver, follow these steps in general   
(Steps specific to each driver have been documented in each of the driver sections)

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
**configVersion** - Configuration version  - Refer full list of supported driver for finding out the appropriate config version
**replicas**  - Number of replicas for controller plugin - Must be set to 1 for all drivers  
**common**  
This field is mandatory and is used to specify common properties for both controller and the node plugin
* image - driver container image
* imagePullPolicy - Image Pull Policy of the driver image
* envs - List of environment variables and their values
#### Optional fields
**controller** - List of environment variables and values which are applicable only for controller  
**node** - List of environment variables and values which are applicable only for node  
**sideCars** - Specification for CSI sidecar containers.  
**authSecret** - Name of the secret holding credentials for use by the driver. If not specified, the default secret *-creds must exist in the same namespace as driver  
**tlsCertSecret** - Name of the TLS cert secret for use by the driver. If not specified, a secret *-certs must exist in the namespace as driver
**storageclass**  
List of Storage Class specification

   1. name - name of the Storage Class
   2. default - Used to specify if the storage class will be marked as default (only set one storage class as default in a cluster)
   3. reclaimPolicy - Sets the PersistentVolumeReclaim Policy for the PVCs. Defaults to Delete if not specified
   4. parameters - driver specific parameters. Refer individual driver section for more details
   5. allowVolumeExpansion - Set to true for allowing volume expansion for PVC
   6. allowedTopologies - Sets the topology keys and values which allows the pods/and volumes to be scheduled on nodes that have access to the storage. Use this only with the CSI PowerFlex driver as rest of the drivers do not support the VOLUME_ACCESSIBILITY_CONSTRAINTS capability

Note: The volumeBindingMode for all storage classes created for the CSI PowerFlex driver will be set to `WaitForFirstConsumer`

**snapshotclass**  
List of Snapshot Class specifications  

   1. name - name of the snapshot class
   2. parameters - driver specific parameters. Refer individual driver section for more details

**forceUpdate**  
Boolean value which can be set to `true` in order to force update the status of the CSI Driver 

**tolerations**
List of tolerations which should be applied to the driver StatefulSet and DaemonSet  
It should be set separately in the controller and node sections if you want separate set of tolerations for them

**nodeSelector**
Used to specify node selectors for the driver StatefulSet and DaemonSet  

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
    configVersion: v3 <- Refer the table containing the full list of supported drivers to find the appropriate config version 
    replicas: 1 <- Always set to 1 for all drivers
    forceUpdate: false <- Set to true in case you want to force an update of driver status
    common: <- All common specification
      image: "dellemc/csi-powermax:v1.4.0.000R" <- driver image for a particular release
      imagePullPolicy: IfNotPresent
      envs:
        - name: X_CSI_POWERMAX_ENDPOINT
          value: "https://0.0.0.0:8443/"
        - name: X_CSI_K8S_CLUSTER_PREFIX
          value: "XYZ"
    storageClass:
      - name: bronze
        default: true
        reclaimPolicy: Delete
        parameters:
          SYMID: "000000000001"
          SRP: DEFAULT_SRP
          ServiceLevel: Bronze
```

Note - The `image` field should point to the correct image tag for version of the driver you are installing.  
For e.g. - If you wish to install v1.4 of the CSI PowerMax driver, use the image tag `dellemc/csi-powermax:v1.4.0.000R`

Note - The name of the Storage Class or the Volume Snapshot Class (which are created in the Kubernetes/OpenShift cluster) is created using the name of the driver and the name provided for these classes in the manifest. This is done in order to ensure that these names are unique if there are multiple drivers installed in the same cluster.
For e.g. - With the above sample manifest, the name of the storage class which is created in the cluster will be `test-powermax-bronze`.  
You can get the name of the StorageClass and SnapshotClass created by the operator by running the commands - `kubectl get storageclass` and `kubectl get volumesnapshotclass`

### SideCars
Although the sidecars field in the driver specification is optional, it is **strongly** recommended to not modify any details related to sidecars provided (if present) in the sample manifests. Any modifications to this should be only done after consulting with Dell EMC support.

#### Snapshotter sidecar
Snapshotter sidecar will not be deployed as part of any driver installation on OpenShift 4.3 clusters. Volume Snapshots are a Technology Preview feature in OpenShift 4.3 and are not officially supported.  
Any attempt to create VolumeSnapshotClass (alpha) as part of the driver installation on OpenShift 4.3 cluster would fail.

### Create Custom Resource manifest using example manifests
#### OperatorHub
Use the OperatorHub GUI to create a new manifest using the example manifest provided for the driver you wish to install

#### Manual installation
Copy the example manifest provided in the dell-csi-operator repository and use this to install the driver
For e.g. – Copy the PowerMax example manifest file
```
$ cp dell-csi-operator/sample/powermax_v140_k8s_v117.yaml .
```
##### Modify the driver specification
* Choose the correct configVersion. Refer the table containing the full list of supported drivers and versions.
* Provide the namespace (in metadata section) where you want to install the driver.
* Provide a name (in metadata section) for the driver. This will be the name of the Custom Resource.
* Edit the values for mandatory configuration parameters specific to your installation.
* Edit/Add any values for optional configuration parameters to customize your installation.

### Create Custom Resource
#### OperatorHub
Use the OperatorHub GUI to create the Custom Resource once you have created the CR manifest

   Or
   
Use one of the sample files in the samples folder and follow the instructions below to install the driver using `kubectl`


#### Manual installation
Create the custom resource using the following command
```
$ kubectl create -f powermax_v140_k8s_v117.yaml
```

### Verification
Once the driver Custom Resource has been created, you can verify the installation

*  Check if Driver CR got created successfully

    For e.g. – If you installed the PowerMax driver
    ```
    $ kubectl get csipowermax -n <driver-namespace>
    ```
* Check the status of the Custom Resource to verify if the driver installation was successful

If the driver-namespace was set to test-powermax, and the name of the driver is powermax, then run the command `kubectl get csipowermax/powermax -n test-powermax -o yaml` to get the details of the Custom Resource.  
Here is a sample output of the above command
```
apiVersion: storage.dell.com/v1
kind: CSIPowerMax
metadata:
  creationTimestamp: "2020-06-04T09:56:02Z"
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
      image: dellemc/csi-powermax:v1.4.0.000R
      imagePullPolicy: IfNotPresent
    configVersion: v2
    controller: {}
    node: {}
    replicas: 1
    sideCars:
    - image: quay.io/k8scsi/csi-snapshotter:v2.1.1
      imagePullPolicy: IfNotPresent
      name: snapshotter
    - image: quay.io/k8scsi/csi-provisioner:v1.6.0
      imagePullPolicy: IfNotPresent
      name: provisioner
    - image: quay.io/k8scsi/csi-attacher:v2.2.0
      imagePullPolicy: IfNotPresent
      name: attacher
    - image: quay.io/k8scsi/csi-node-driver-registrar:v1.2.0
      imagePullPolicy: IfNotPresent
      name: registrar
    - image: quay.io/k8scsi/csi-resizer:v0.5.0
      imagePullPolicy: IfNotPresent
      name: resizer
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
  controllerStatus:
    available:
    - powermax-controller-0
  driverHash: 3166303568
  lastUpdate:
    condition: Running
    time: "2020-09-05T01:53:01Z"
  nodeStatus:
    available:
    - powermax-node-kphlf
    - powermax-node-x4f74
    - powermax-node-zz2h5
    - powermax-node-xlt46
  state: Running
```

* Driver statefulset & daemonset are created in the same namespace automatically by the Operator
    ```
    $ kubectl get statefulset -n <driver-namespace>
    $ kubectl get daemonset -n <driver-namespace>
    ```
## Update CSI Drivers

The CSI Drivers installed by the Dell CSI Operator can be updated like any Kubernetes resource. This can be achieved in various ways which include –

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
For uninstalling any CSI drivers deployed the Dell CSI Operator, just delete the respective Custom Resources.  
This can be done using OperatorHub GUI by deleting the CR or via kubectl.
    
For e.g. – To uninstall a PowerFlex driver installed via the operator, delete the Custom Resource(CR)

```
# Replace driver-name and driver-namespace with their respective values
$ kubectl delete vxflexos/<driver-name> -n <driver-namespace>
```

## Limitations
* The Dell CSI Operator can't manage any existing driver installed using Helm charts. If you already have installed one of the DellEMC CSI driver in your cluster and  want to use the operator based deployment, uninstall the driver and then redeploy the driver following the installation procedure described above
* The Dell CSI Operator can't update storage classes as it is prohibited by Kubernetes. Any attempt to do so will cause an error and the driver Custom Resource will be left in a `Failed` state. Refer the Troubleshooting section to fix the driver CR.
* The Dell CSI Operator is not fully compliant with the OperatorHub React UI elements and some of the Custom Resource fields may show up as invalid or unsupported in the OperatorHub GUI. To get around this problem, use kubectl/oc commands to get details about the Custom Resource(CR). This issue will be fixed in the upcoming releases of the Dell CSI Operator


## Troubleshooting
* Before installing the drivers, Dell CSI Operator tries to validate the Custom Resource being created. If some mandatory environment variables are missing or there is a type mismatch, then the Operator will report an error during the reconciliation attempts.  
Because of this, the status of the Custom Resource will change to "Failed" and the error captured in the "ErrorMessage" field in the status.  
For e.g. - If the PowerMax driver was installed in the namespace test-powermax and has the name powermax, then run the command `kubectl get csipowermax/powermax -n test-powermax -o yaml` to get the Custom Resource details.  
If there was an error while installing the driver, then you would see a status like this -
  ```
  status:
    status:
      errorMessage: mandatory Env - X_CSI_K8S_CLUSTER_PREFIX not specified in user spec
      state: Failed
  ```  

    The state of the Custom Resource can also change to `Failed` because of any other prohibited updates or any failure while installing the driver. In order to recover from this failure, 
    fix the error in the manifest and update/patch the Custom Resource

* After an update to the driver, the controller pod may not have the latest desired specification  
The above happens when the controller pod was in a failed state before applying the update. Even though the Dell CSI Operator updates the pod template specification for the StatefulSet, the StatefulSet controller does not apply the update to the pod. This happens because of the unique nature of StatefulSets where the controller tries to retain the last known working state. 

    To get around this problem, the Dell CSI Operator forces an update of the pod specification by deleting the older pod. In case the Dell CSI Operator fails to do so, delete the controller pod to force an update of the controller pod specification

* The Status of the CSI Driver Custom Resource shows the state of the driver pods after installation. This state will not be updated automatically if there are any changes to the driver pods outside any Operator operations
At times because of inconsistencies in fetching data from the Kubernetes cache, state of some driver pods may not be updated correctly in the status. To force an update of the state, you can update
the Custom Resource forcefully by setting forceUpdate to true. If all the driver pods are in `Available` State, then the state of the Custom Resource will be updated as `Running`


# Driver details
## CSI PowerScale
### Pre-requisites
#### Create secret to store PowerScale credentials
Create a secret named `isilon-creds`, in the namespace where the CSI PowerScale driver will be installed, using the following manifest
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
Create a secret named `isilon-certs` in the namespace where the CSI PowerScale driver will be installed. This is an optional step and is only required if you are setting the env variable `X_CSI_ISI_INSECURE` to `false`. Please refer detailed documentation on how to create this secret in the Product Guide [here](https://github.com/dell/csi-isilon)

### Set the following *Mandatory* Environment variables
| Variable name      | Section | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_ISI_ENDPOINT | common    | HTTPS endpoint of the PowerScale OneFS API server | 1.1.1.1 |
| X_CSI_ISI_PORT     | common    | HTTPS port number of the PowerScale OneFS API server (string) | 8080 |

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

### StorageClass attributes
| Name | Mandatory | Description | Default |
| ---- | :-------: | ----------- | ------- |
| allowVolumeExpansion | no | Allow volume expansion for pvc backed by this storage class | true |

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
Create a namespace called unity (it can be any user-defined name; But commands in this section assumes that the namespace is unity)
Prepare the secret.json for driver configuration.
The following table lists driver configuration parameters for multiple storage arrays.

| Parameter | Description | Required | Default |
| --------- | ----------- | -------- |-------- |   
| username | Username for accessing unity system  | true | - |
| password | Password for accessing unity system  | true | - |
| restGateway | REST API gateway HTTPS endpoint Unity system| true | - |
| arrayId | ArrayID for unity system | true | - |
| insecure | "unityInsecure" determines if the driver is going to validate unisphere certs while connecting to the Unisphere REST API interface If it is set to false, then a secret unity-certs has to be created with a X.509 certificate of CA which signed the Unisphere certificate | true | true |
| isDefaultArray | An array having isDefaultArray=true is for backward compatibility. This parameter should occur once in the list. | false | false |

Ex: secret.json

```json5

   {
     "storageArrayList": [
       {
         "username": "user",
         "password": "password",
         "restGateway": "https://10.1.1.1",
         "arrayId": "APM00******1",
         "insecure": true,
         "isDefaultArray": true
       },
       {
         "username": "user",
         "password": "password",
         "restGateway": "https://10.1.1.2",
         "arrayId": "APM00******2",
         "insecure": true
       }
     ]
   }
  
```

`kubectl create secret generic unity-creds -n unity --from-file=config=secret.json`

Use the following command to replace or update the secret

`kubectl create secret generic unity-creds -n unity --from-file=config=secret.json -o yaml --dry-run | kubectl replace -f -`

**Note**: The user needs to validate the JSON syntax and array related key/values while replacing the unity-creds secret.
The driver will continue to use previous values in case of an error found in the JSON file.

#### Optional - Create secret for client side TLS verification

Create an empty secret. Ex: empty-secret.yaml

```
  apiVersion: v1
  kind: Secret
  metadata:
    name: unity-certs-0
    namespace: unity
  type: Opaque
  data:
    cert-0: ""
```

Please refer detailed documentation on how to create this secret in the Product Guide [here](https://github.com/dell/csi-unity#certificate-validation-for-unisphere-rest-api-calls)

### Modify/Set the following *optional* environment variables

| Variable name      | Section    | Description | Default |
| ------------------ | ---------- | ----------- | ------- |
| X_CSI_DEBUG | common | To enable debug mode | false |
| X_CSI_UNITY_SYNC_NODEINFO_INTERVAL | node |  Time interval to add node info to array. Default 15 minutes. Minimum value should be 1.  If your specifies 0, then time is set to default value. | true |
| X_CSI_UNITY_DEBUG | common | To enable debug logging mode  | false |


### StorageClass Parameters
        
| Parameter | Description | Required | Default |
| --------- | ----------- | -------- |-------- |
| storagePool | Unity Storage Pool CLI ID to use with in the Kubernetes storage class | true | - |
| thinProvisioned | To set volume thinProvisioned | false | "true" |    
| isDataReductionEnabled | To set volume data reduction | false | "false" |
| volumeTieringPolicy | To set volume tiering policy | false | 0 |
| FsType | Block volume related parameter. To set File system type. Possible values are ext3,ext4,xfs. Supported for FC/iSCSI protocol only. | false | ext4 |
| hostIOLimitName | Block volume related parameter.  To set unity host IO limit. Supported for FC/iSCSI protocol only. | false | "" |    
| nasServer | NFS related parameter. NAS Server CLI ID for filesystem creation. | true | "" |
| hostIoSize | NFS related parameter. To set filesystem host IO Size. | false | "8192" |
| reclaimPolicy | What should happen when a volume is removed | false | Delete |

### SnapshotClass parameters
Following parameters are not present in values.yaml in the Helm based installer

| Parameter | Description | Required | Default |
| --------- | ----------- | -------- |-------- |
| snapshotRetentionDuration | TO set snapshot retention duration. Format:"1:23:52:50" (number of days:hours:minutes:sec)| false | "" |

## CSI PowerFlex
### Pre-requisites

#### Create secret to store PowerFlex credentials
Create a secret named `vxflexos-creds`,  in the namespace where the CSI PowerFlex driver will be installed, using the following manifest
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

#### Install PowerFlex Storage Data Client
Install the PowerFlex Storage Data Client (SDC) on all Kubernetes nodes.  
For detailed PowerFlex installation procedure, and current version of the driver see the [Dell EMC PowerFlex Deployment Guide.](https://github.com/dell/csi-vxflexos/blob/master/CSI%20Driver%20for%20VxFlex%20OS%20Product%20Guide.pdf)


Procedure:
1. Download the PowerFlex SDC from Dell EMC Online support. The filename is EMC-ScaleIO-sdc-*.rpm, where * is the SDC name corresponding to the PowerFlex installation version.
2. Export the shell variable MDM_IP in a comma-separated list. This list contains the IP addresses of the MDMs.
    export MDM_IP=xx.xxx.xx.xx,xx.xxx.xx.xx, where xxx represents the actual IP address in your environment variable.
3. Install the SDC using the following commands:
l For Red Hat Enterprise Linux and Cent OS, run rpm -iv ./EMC-ScaleIO-sdc-*.x86_64.rpm, where * is the SDC name corresponding to the PowerFlex installation version.
l For Ubuntu, run EMC-ScaleIO-sdc-3.0-0.769.Ubuntu.18.04.x86_64.deb.

### Set the following *Mandatory* Environment variables

| Variable name      | Section | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_VXFLEXOS_SYSTEMNAME | common | defines the name of the PowerFlex system from which volumes will be provisioned. This must either be set to the PowerFlex system name or system ID | systemname |
| X_CSI_VXFLEXOS_ENDPOINT | common | defines the PowerFlex REST API endpoint, with full URL, typically leveraging HTTPS. You must set this for your PowerFlex installations REST gateway | https://127.0.0.1 |


### Modify/Set the following **optional environment variables**


| Variable name      | Section    | Description | Example |
| ------------------ | ---------- | ----------- | ------- |
| X_CSI_DEBUG | common | To enable debug mode | false |
| X_CSI_VXFLEXOS_ENABLELISTVOLUMESNAPSHOT | common |  Enable list volume operation to include snapshots (since creating a volume from a snap actually results in a new snap) | false |
| X_CSI_VXFLEXOS_ENABLESNAPSHOTCGDELETE | common | Enable this to automatically delete all snapshots in a consistency group when a snap in the group is deleted | false |

### StorageClass parameters

| Name | Mandatory | Description | Example |
| ---- | :-------: | ----------- | ------- |
| storagePool | yes | defines the PowerFlex storage pool from which this driver will provision volumes. You must set this for the primary storage pool to be used | sp |
| FsType | No  | To set File system type. Possible values are ext3,ext4,xfs | xfs |
| allowedTopologies | no | Once the allowed topology is modified in storage class, pods/and volumes will always be scheduled on nodes that have access to the storage | allowedTopology: key: xxx/xx value: xxx |

Note:  key : csi-vxflexos.dellemc.com/X_CSI_VXFLEXOS_SYSTEMNAME , change the X_CSI_VXFLEXOS_SYSTEMNAME parameter in allowed topology to it's value(System name) 

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
  # Uncomment the following key if you wish to use ISCSI CHAP authentication (v1.3.0 onwards)
  # chapsecret: <base64 CHAP secret>
```
The base64 username and password can be obtained by running the following commands
```
# If myusername is the username
echo -n "myusername" | base64
# If mypassword is the password
echo -n "mypassword" | base64
# If mychapsecret is the ISCSI CHAP secret
echo -n "mychapsecret" | base64
```
#### Optional - Create secret for client side TLS verification
Create a secret named `powermax-certs` in the namespace where the CSI PowerMax driver will be installed. This is an optional step and is only required if you are setting the env variable `X_CSI_POWERMAX_SKIP_CERTIFICATE_VALIDATION` to `false`. Please refer detailed documentation on how to create this secret in the Product Guide [here](https://github.com/dell/csi-powermax)

#### Node requirements
For node specific requirements, please refer detailed instructions in the Product Guide [here](https://github.com/dell/csi-powermax "CSI PowerMax")  
Choose the Product Guide for the version you are installing by selecting the corresponding release [here](https://github.com/dell/csi-powermax/releases "CSI PowerMax Releases")

### Set the following *Mandatory* Environment variables

| Variable name      | Section | Description | Example |
| ------------------ | ------ | -----------  | ------- |
| X_CSI_POWERMAX_ENDPOINT | common | IP address of the Unisphere for PowerMax | https://0.0.0.0:8443 |
| X_CSI_K8S_CLUSTER_PREFIX | common | defines a prefix that is appended onto all resources created in the Array; unique per K8s/CSI deployment; max length - 3 characters | XYZ |

### Modify/Set the following *Optional* environment variables

| Variable name      | Section    | Description | Example | Comments |
| ------------------ | ---------- | ----------- | ------- | -------- |
| X_CSI_POWERMAX_DEBUG | common | determines if HTTP Request/Response is logged | "false" | |
| X_CSI_POWERMAX_SKIP_CERTIFICATE_VALIDATION | common | skip client side TLS verification of Unisphere certificates | "true" | |
| X_CSI_POWERMAX_PORTGROUPS | common | List of comma separated port groups (ISCSI only) | "PortGroup1,PortGroup2" | |
| X_CSI_POWERMAX_ARRAYS | common | list of comma separated array id(s) which will be managed by the driver | "000000000001,000000000002" | |
| X_CSI_TRANSPORT_PROTOCOL | common | preferred transport protocol - FC/FIBRE or ISCSI/iSCSI. If left blank, driver will autoselect | "FC" | |
| X_CSI_ENABLE_BLOCK | common | enable Block Volume capability which is in experimental phase | "true" | |
| X_CSI_POWERMAX_ISCSI_ENABLE_CHAP | node | enable ISCSI CHAP authentication | "true" | Only supported from v1.3.0 onwards |
| X_CSI_POWERMAX_DRIVER_NAME | common | Custom CSI driver name | "csi-powermax" | Only supported from v1.3.0 onwards |
| X_CSI_IG_NODENAME_TEMPLATE | common | Template used for creating Hosts on PowerMax | "a-b-c-%foo%-xyz" | The text between % symbols(foo) is replaced by actual host name |
| X_CSI_IG_MODIFY_HOSTNAME | node | determines if node plugin can rename any existing Host on PowerMax array | "true" | Use it along with node name template to rename existing Hosts |
| X_CSI_POWERMAX_PROXY_SERVICE_NAME | common | Name of CSI PowerMax ReverseProxy service | "powermax-reverseproxy" | Leave blank if not using reverse proxy |
| X_CSI_GRPC_MAX_THREADS | common | Number of concurrent grpc requests allowed per client | "4" | Set to a higher number (<50) when using reverse proxy |

Note -  Please refer the Product guide for CSI PowerMax v1.4.0 for detailed instructions before setting X_CSI_POWERMAX_ISCSI_ENABLE_CHAP & X_CSI_POWERMAX_DRIVER_NAME

### StorageClass parameters

| Name          | Mandatory | Description | Example |
| ------------- | :-------: | ----------- | ------- |
| SYMID         | yes       | Symmetrix ID | 000000000001 |
| SRP           | yes       | Storage Resource Pool name | DEFAULT_SRP|
| ServiceLevel  | no        | Service Level | Bronze |
| FsType        | no        | File System type (xfs/ext4) | xfs |

### SnapshotClass parameters
No parameters have to be specified for Volume Snapshot Class for PowerMax

## CSI PowerStore
### Pre-requisites
#### Create secret to store PowerStore API credentials

Create a secret named `powerstore-creds`, in the namespace where the CSI PowerStore driver will be installed, using the following manifest

```
apiVersion: v1
kind: Secret
metadata:
  name: powerstore-creds
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

### Set the following *Mandatory* Environment variables

| Variable name      | Section | Description | Example |
| ------------------ | ------ | ----------- | ------- |
| X_CSI_POWERSTORE_ENDPOINT | common | Must provide a PowerStore HTTPS API url | https://127.0.0.1/api/rest |
| X_CSI_TRANSPORT_PROTOCOL | node | Choose what transport protocol to use (`ISCSI`, `FC`, `auto` or `None`) | auto |
| X_CSI_POWERSTORE_NODE_NAME_PREFIX | node | Prefix to add to each node registered by the CSI driver | `"csi-node"` |

### Modify/Set the following *optional* environment variables

| Variable name      | Section    | Description | Default |
| ------------------ | ---------- | ----------- | ------- |
| X_CSI_DEBUG | common | To enable debug mode | true |
| X_CSI_POWERSTORE_INSECURE | common | To enable insecure mode without client side verification of certificates | true |
| X_CSI_FC_PORTS_FILTER_FILE_PATH | node | To set path to the file which provide list of WWPN which should be used by the driver for FC connection on this node | `"/etc/fc-ports-filter"`

### StorageClass Parameters

| Parameter | Description | Required | Default |
| --------- | ----------- | -------- |-------- |
| FsType | To set File system type. Possible values are ext3,ext4,xfs,nfs | false | ext4 |
| nasName | To set what NAS Server to use for NFS | false | "" |

### SnapshotClass parameters
No parameters have to be specified for Volume Snapshot Class for PowerStore

## CSI PowerMax ReverseProxy
Starting v1.1 of 'dell-csi-operator', you can install the new CSI PowerMax ReverseProxy service using the `dell-csi-operator`
CSI PowerMax ReverseProxy is a new optional component which can be installed along with the CSI PowerMax driver  
Please refer CSI PowerMax product guide [here](https://github.com/dell/csi-powermax "CSI PowerMax")  for more details.

When you install CSI PowerMax ReverseProxy, `dell-csi-operator` is going to create a `Deployment` and `ClusterIP` service as part of the installation

Note - If you wish to use the ReverseProxy with CSI PowerMax driver, the ReverseProxy service should be created before you install the CSIPowerMax driver 

### Pre-requisites
Create a TLS secret which holds a SSL certificate & a private key which is required by the reverse proxy server. 
Use a tool like `openssl` to generate this secret using the example below:

```
    openssl genrsa -out tls.key 2048
    openssl req -new -x509 -sha256 -key tls.key -out tls.crt -days 3650
    kubectl create secret -n powermax tls revproxy-certs --cert=tls.crt --key=tls.key
```

### Set the following parameters in the CSI PowerMaxReverseProxy Spec
**tlsSecret** : Provide the name of the TLS secret. If using the above example, it should be set to `revproxy-certs`  
**config** : This section contains the details of the Reverse Proxy configuration  
**mode** : This value is set to `Linked` by default. Don't change this value  
**linkConfig** : This section contains the configuration of the `Linked` mode  
**primary** : This section holds details for the primary Unisphere which the Reverse Proxy will connect to
**backup** : This optional section holds details for a backup Unisphere which the Reverse Proxy can connect to if Primary Unisphere is unreachable  
**url** : URL of the Unisphere server
**skipCertificateValidation**: This setting determines if the client side Unisphere certificate validation is required
**certSecret**: Secret name which holds the CA certificates which was used to sign Unisphere SSL certificates. Mandatory if skipCertificateValidation is set to `false`

Here is a sample manifest with each field annotated. A copy of this manifest is provided in the `samples` folder
```
apiVersion: storage.dell.com/v1
kind: CSIPowerMaxRevProxy
metadata:
  name: powermax-reverseproxy # <- Name of the CSIPowerMaxRevProxy object
  namespace: test-powermax # <- Set the namespace to where you will install the CSI PowerMax driver
spec:
  # Image for CSI PowerMax ReverseProxy
  image: dellemc/csipowermax-reverseproxy:v1.0.0.000R # <- CSI PowerMax Reverse Proxy image 
  imagePullPolicy: Always
  # TLS secret which contains SSL certificate and private key for the Reverse Proxy server
  tlsSecret: csirevproxy-tls-secret
  config:
    # Mode for the proxy - only supported mode for now is "Linked"
    mode: Linked
    linkConfig:
      primary:
        url: https://0.0.0.0:8443 #Unisphere URL
        skipCertificateValidation: true # This setting determines if client side Unisphere certificate validation is to be skipped
        certSecret: "" # Provide this value if skipCertificateValidation is set to false
      backup: # This is an optional field and lets you configure a backup unisphere which can be used by proxy server
        url: https://0.0.0.0:8443 #Unisphere URL
        skipCertificateValidation: true
```

### Installation
Copy the sample file - `powermax_reverseproxy.yaml` from the `samples` folder or use the sample available in the `OperatorHub` GUI  
Edit and input all required parameters and then use the `OperatorHub` GUI or run the following command to install the CSI PowerMax Reverse Proxy service

    kubectl create -f powermax_reverseproxy.yaml

You can query for the deployment and service created as part of the installation using the following commands:
  
    kubectl get deployment -n <namespace>
    kubectl get svc -n <namespace>
