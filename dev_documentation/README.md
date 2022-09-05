
# Dell CSI Operator for Dell CSI Drivers
**Repository for development of common operator for Dell CSI Drivers**

## Description
`dell-csi-operator` is a Kubernetes operator which can be used to install and manage various CSI Drivers provided by Dell for different storage arrays. 

It is built, deployed and tested using the toolset provided by Operator [framework](https://github.com/operator-framework) which include:
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)
* [operator-registry](https://github.com/operator-framework/operator-registry)

Note - **operator-sdk-vv1.18.2-x86_64-linux-gnu** has been used to build the `dell-csi-operator` and is available at the root of the repository

## Custom Resource Definitions
`dell-csi-operator` manages a set of [Custom Resource Definitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/) (CRDs)  
These CRDs represent a specific CSI Driver installation and are part of the API group `storage.dell.com`  
The current set of CRDs managed by the `dell-csi-operator` are:
* CSIUnityXT
* CSIIsilon
* CSIVXFlexOS
* CSIPowerStore
* CSIPowerMax
* CSIPowerMaxRevProxy

## Controllers
`dell-csi-operator` utilizes Kubernetes [controller runtime](https://github.com/kubernetes-sigs/controller-runtime) libraries for building controllers which
run as part of the Operator deployment.  
These controllers watch for any requests to create/modify/delete instances of the Custom Resource Definitions (CRDs) and handle the [reconciliation](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/reconcile)
of these requests.

Each instance of a CRD is called a Custom Resource (CR) and can be managed by a client like `kubectl` in the same way a native
Kubernetes resource is managed.  
When you create a Custom Resource, then the corresponding Controller will create the Kubernetes objects required for the driver installation.  

This includes:
* Service Accounts and RBAC configuration
* StatefulSet
* DaemonSet
* Deployment and Service (only for CSIPowerMaxRevProxy)

Note - There is one controller per Custom Resource type and each controller runs a single worker 


## Build and Deploy

### Pre-requisites
* golang >= 1.18.x
* operator-sdk v1.15.0 [Download link](https://github.com/operator-framework/operator-sdk/releases/download/v1.15.0/operator-sdk_linux_amd64)
* opm v1.14.0 [Download link](https://github.com/operator-framework/operator-registry/releases/download/v1.14.0/linux-amd64-opm)

While installing `operator-sdk` & `opm`, make sure they are available in your PATH & the binaries have the right names and permissions.

There are multiple Makefile targets available for building the Operator
### Building Operator binary
If you wish to build the Operator binary, run the command - `make manager`

### Build Operator image
There are a few available Makefile targets which let you build a docker image for the Operator. 
The docker image is built using the `operator-sdk` binary (which is available in the repository). 
The base image used to build the docker image is UBI (Universal Base Image) provided by Red Hat.


Run the command `make docker-build` to build a docker image for the Operator which will be tagged with git semantic versioning  
The official builds of Operator which are hosted on artifactory are built using the `make docker-build` command

By default, this target will tag the newly built images with the artifactory repo  
Run the command `IMG=my-image-repo/dell-csi-operator:<your-tag> make docker-build` to tag the docker image with your own repository

### Push docker image to private repository
Run the command `IMG=my-image-repo/dell-csi-operator:<your-tag> make docker-push` to build and push the docker image to `my-image-repo` private repository

### Push docker image to artifactory
Run  the command `REGISTRY=<your-registry> make docker-push` to build and push the operator image to an image registry

### Build Operator image along with bundle and index images
Run the command `REGISTRY=<your-registry> make index` to build the operator image, operator bundle image & an index image which refers to the bundle image  
The above command will also push the images to the specified registry.

### Run the Operator locally without deploying any image

#### Pre-requisites  
Make sure that a [**KubeConfig**](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file pointing to your Kubernetes/OpenShift cluster is present in the default location

Run either of the following commands to run the Operator in your cluster without creating a deployment

* `make install run`

The above command will update the CRDs, install the CRDs and then run the Operator. 

#### Deploy the operator
There are primarily two ways of deploying the Operator -
1. Use Operator manifests and installation scripts
2. Use [OLM](https://github.com/operator-framework/operator-lifecycle-manager) to install the Operator  
     During development -  
     a) Use internal Operator Catalog hosted on artifactory  
     b) Use OperatorHub to deploy publicly available Operator
3. Offline installation of operator

### Deploy Operator without OLM
* To deploy the Operator with the latest build, run the command `make deploy`. 

* To deploy the Operator with the docker image or namespace of your choice,
  Run the command `make deploy-dev IMG=my-image-repo/dell-csi-operator:<your-tag> NAMESPACE=<your-namespace>` 

* Run the command `bash scripts/install.sh` to deploy the Operator in the default namespace

The installation script does the following:
* Create a Service Account and setup the RBAC using ClusterRole and ClusterRoleBindings
* Create a ConfigMap which is used by the Operator
* Install the various CRDs managed by the Operator
* Create the Operator deployment (single replica)

### Deploy Operator using OLM

#### Pre-requisites
OLM is not available on upstream Kubernetes cluster by default and has to be installed before installing the Operator.

Run the command `operator-sdk olm install` to install OLM in your cluster

Note - OLM is available as a default component in OpenShift clusters and you don't need to install it separately

Note - Please refer Operator SDK documentation for more help on using operator-sdk

#### Deploy Operator using internal Catalog
Run the command `bash scripts/install_olm.sh` to install the Operator using OLM in an upstream Kubernetes cluster
Run the command `bash scripts/install_olm.sh --openshift` to install the Operator using OLM in an OpenShift cluster

The above scripts will create instances of the following CRDs:

CatalogSource
OperatorGroup
Subscription
InstallPlan
ClusterServiceVersion
You can query the status of these objects in the test-olm namespace to get more information or troubleshoot any issues found during installation.

### Offline Installation of Operator
Operator can be installed offline from the internal repo by following the below steps
* Create a directory for offline bundle deployment, for example  `mkdir -p offlinebundle/dell-csi-operator`
* Set the path for STAGING_DIR by running the command `export STAGING_DIR=(path of above folder like offlinebundle/dell-csi-operator)`
* Clone the internal repo to another folder, for example `mkdir codebase` followed by `git clone https://eos2git.cec.lab.emc.com/DevCon/dell-csi-operator.git` (which should be executed from codebase folder).
* Execute `bash scripts/staging.sh` from `codebase/dell-csi-operator`
* Follow offline procedure steps noted in https://dell.github.io/storage-plugin-docs/docs/installation/offline/ from `offlinebundle/dell-csi-operator` folder

Note - `csi-offline-bundle.sh` is intended to work with the external repo [https://github.com/dell/dell-csi-operator]. Above mentioned steps are prerequisites for the script to work with the internal repo.

### Verify Installation
Post a successful installation of the Operator, there should be an Operator deployment created in the cluster.
Depending on the method of installation, this would be created in the namespace:

Installation without OLM - default
Installation with OLM - test-olm
Query for CRDs

#### Query for CRDs
You can also query for the CRDs installed in the cluster by running the command
`kubectl get crd`. You should see the following CRDs in the list of CRDs installed in the cluster:
* csiisilons.storage.dell.com
* csipowermaxes.storage.dell.com
* csipowermaxrevproxies.storage.dell.com
* csipowerstores.storage.dell.com
* csiunities.storage.dell.com
* csivxflexoses.storage.dell.com

## Install a CSI driver
Once the CRDs and Operator has been installed in the cluster, you can install any CSI driver by creating a Custom Resource (CR)  
For e.g. - If you want to install the CSI PowerMax driver, you should create a CR of the Kind `CSIPowerMax`

Here are the steps involved in installing a driver:
* Make sure all dependencies for the driver have been met. This can include installation of specific packages, configuration of services, nodes
* Create the namespace where you wish to install the driver
* Create any mandatory and optional secrets required for the driver installation
* Create the Custom Resource using the sample manifests provided for the driver  

### Create custom resource
A lot of sample manifest files have been provided in the `samples` folder to help with the installation of various CSI Drivers
They follow the naming convention

    {driver name}_{driver version}_k8s_{k8 version}.yaml

Or

    {driver name}_{driver version}_ops_{OpenShift version}.yaml

Use the correct sample manifest based on the driver, driver version and Kubernetes/OpenShift version

For e.g.  
powermax_v140_k8s_117.yaml <- To install CSI PowerMax driver v1.4.0 on a Kubernetes 1.17 cluster
powermax_v140_ops_43.yaml <- To install CSI PowerMax driver v1.4.0 on an OpenShift 4.3 cluster

Note - For this example, we will assume that the Custom Resource will be created in the namespace `powermax` with the name `powermax`

Create a new file `powermax.yaml` by copying the relevant sample file and edit the contents (specific to your installation).

Run the command `kubectl create -f powermax.yaml` to create the Custom Resource (CR) 

Check the status of the Custom Resource by running the command `kubectl get csipowermax -n powermax`

If the status of the CR shows `Running` then the driver installation completed successfully with all driver pods running  
If the status of the CR shows `Succeeded` then the driver installation succeeded but all driver pods are not up and running  
If the status of the CR shows `InvalidConfig` then there is an incorrect value specified in the Custom Resource manifest  

Note - The driver status can take some time to migrate from `Succeeded` to `Running` because of the time taken for the driver pods to completely start up.  
In case some pods are not up and running, the Operator will query for their status for at least an hour before giving up on updating the status of the Custom Resource.

You can also check the status of the driver pods by running the command `kubectl get pods -n powermax`

Note - Run the command `kubectl get csipowermax --all-namespaces` to query for all Custom Resources of the type CSIPowerMax in your cluster

### Update Custom Resource
If you want to update the driver installation or fix any issues in the Custom Resource (for e.g. - InValidConfig), then you can update the Custom Resource  
This can be done in multiple ways
* Run the command `kubectl edit csipowermax powermax -n powermax` and edit any desired field(s)
* Update the manifest file and run the command `kubectl apply -f powermax.yaml`
* Directly patch the Custom Resource (refer Kubernetes documentation)

Once the update has been applied to the Custom Resource, the Operator will try to `reconcile` the desired state with the observed state in the cluster and apply required changes (if any) to the various resources part of the driver installation

### Delete Custom Resource
Run the command `kubectl delete -f powermax.yaml` to delete the Custom Resource. This will delete the Custom Resource and delete all the driver pods.

# Dell CSI Operator Certified Bundle 
The package name used is `dell-csi-operator-certified` and the CSV file has some additional fields compared to the CSVs for the community operators

The CRD files have been copied from deploy/crds folder & the annotations file has been generated via opm

## Steps used for generating bundle for certified operators
* Create a directory called cert_bundle to hold the manifests
* Create a folder called 1.1.0 and added the CRD & the CSV files
* Copied the package file to the cert_bundle directory
* Run the command `opm alpha bundle generate -d ./1.1.0/ -u ./1.1.0/` from the cert_bundle directory
* The above command generates the manifests & metadata directory & the bundle.Dockerfile (in the cert_bundle directory)
* Add the required labels to the bundle.Dockerfile
    LABEL com.redhat.openshift.versions="v4.5,v4.6"
    LABEL com.redhat.delivery.backport=true
    LABEL com.redhat.delivery.operator.bundle=true
* Use the newly created bundle.Dockerfile to build the bundle image for the certified operator
    podman build -f bundle.Dockerfile -t <your_repo>/csiopbundle_ops_certified:v1.1.0 .
    

## Installing the certified operator
Currently there are no scripts provided to create the certified bundle image & add it to any catalog source.  
If you wish to install the certified operator, use the one available on OpenShift OperatorHub  
The installation can be done via the OpenShift GUI or by running the following command from the root of the `dell-csi-operator` repository
```
oc apply -f deploy/olm/operator_ops_certified.yaml
```

