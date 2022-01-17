# Operator Framework

Operator Framework is a framework used for developing and managing entire lifecycle of Operators
This [framework](https://github.com/operator-framework) contains various projects like
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)
* [operator-registry](https://github.com/operator-framework/operator-registry)

* `operator-sdk` is used for the development of Operators
* `Operator Lifecycle Manager(OLM)` extends Kubernetes to provide a declarative way to install, manage, and upgrade Operators and their dependencies in a cluster
* `Operator Registry` runs in a Kubernetes or OpenShift cluster to provide operator catalog data to Operator Lifecycle Manager.


## Supported Operator Registries
`dell-csi-operator` is supported officially on the following public registries:
* OpenShift Certified Operators
* Upstream Community Operators
* OpenShift Community Operators (discontinued after v1.1.0)

## Packaging an Operator
This section describes the various components involved in packaging an operator

### Packages
OLM identifies operators by package names. The package names used for `dell-csi-operator` in the various public Operator Registries are:

* OpenShift Certified Operators - `dell-csi-operator-certified`
* Upstream Community Operators  - `dell-csi-operator`
* OpenShift Community Operators - `dell-csi-operator`

### Bundles
A directory of files with one `ClusterServiceVersion` is referred to as a "bundle". A bundle typically includes a ClusterServiceVersion and the CRDs that define the owned APIs of the CSV in its manifest directory, though additional objects may be included. It also includes an annotations file in its metadata folder which defines some higher level aggregate data that helps to describe the format and package information about how the bundle should be added into an index of bundles.

>*NOTE:* Each bundle represents one unique version of the operator 

```
 # Example dell-csi-operator-certified bundle
 etcd
 ├── manifests
 │   ├── storage.dell.com_csipowermaxes.yaml
 │   └── dell-csi-operator-certified.clusterserviceversion.yaml
 └── metadata
     └── annotations.yaml
```

#### Bundle & Index images
`Index` image encapsulates a catalog registry database which can contain multiple packages. These are built using a tool called [opm](https://github.com/operator-framework/operator-registry)
These images are additive & an older image can be used as a base for building a new image.

The index images are used while creating the `CatalogSource` object in the Cluster which are used to deploy Operator Registries.

Using `bundle` Dockerfiles, bundle images are built which are then packaged into the `Index` image using `opm`

Each bundle image contains a single bundle for one Operator package. 
Thus a single index image can contain multiple operator packages and multiple bundles representing various versions of the operators.

#### Updating Certified and Community bundles
* Use `make update-bundle-keep-base` command to update the certified bundles, it will update the CSV and manifest files for certified bundle which exists in `bundle` directory.
* Once certified bundles are updated, execute the script `scripts\build_olm_community_images.sh` to update community bundle CSV and manifest files which exists in `community_bundle` directory. Then script builds community index and bundle images which can be used for testing of operator with OLM.

**NOTE:** For `community_bundles/manifests`, some of the files which are already present in `bundle/manifests` directory are skipped for check-in. Only two files: `dell-csi-operator.clusterserviceversion.yaml` and `dell-csi-operator.package.yaml` need to be checked in.

### Package Manifests
This format precedes the bundle format and is still in use by `community-operators`

The `*package.yaml` file contains the package definition and includes the relevant `CSV` & `Channel` details
```
# Example from upstream-community-operators
├── 1.0.0
│   ├── csiisilons.storage.dell.com.crd.yaml
│   ├── csipowermaxes.storage.dell.com.crd.yaml
│   ├── csiunities.storage.dell.com.crd.yaml
│   ├── csivxflexoses.storage.dell.com.crd.yaml
│   ├── dell-csi-operator.v1.0.0.clusterserviceversion.yaml
│   └── volumesnapshotclasses.snapshot.storage.k8s.io.crd.yaml
├── 1.1.0
│   ├── csiisilons.storage.dell.com.crd.yaml
│   ├── csipowermaxes.storage.dell.com.crd.yaml
│   ├── csipowermaxrevproxies.storage.dell.com.crd.yaml
│   ├── csipowerstores.storage.dell.com.crd.yaml
│   ├── csiunities.storage.dell.com.crd.yaml
│   ├── csivxflexoses.storage.dell.com.crd.yaml
│   └── dell-csi-operator.v1.1.0.clusterserviceversion.yaml
├── ci.yaml
└── dell-csi-operator.package.yaml
```

*NOTE:* The `dell-csi-operator.package.yaml` is not used in bundle generation and is only useful with the catalogs which use package manifests

## Testing `dell-csi-operator` using OLM
For the purpose of testing installation & upgrade of `dell-csi-operator`, Operator registries are deployed using index images built as part of the build process.
These Operator Registries are used for the purpose of testing the operator and there are no plans to publish them (or the associated bundle/index images)

### Certified Operator Package
The Bundle & Index images can be built using the Makefile target `make index` which builds both the bundle & the index images
Dockerfile used for the bundle builds - `bundle.Dockerfile`
The bundle manifests are located in the `bundle` folder which is updated by the Makefile targets - 
* `update-bundle` - Used for just updating the bundle manifests
* `index` - Used only while building the images. This updates the manifests before building the images

### Upstream Community Operator Package
>*NOTE:* Even though the upstream-community-operators registry uses `package manifests`, we build bundles to test the CSV/Package files internally.
This is done in order to provide a common way of testing the installation of `dell-csi-operator` on both upstream kubernetes as well as Red Hat OpenShift clusters.

The Bundle & Index images can be built using the script `build_olm_images_k8.sh`.  
Dockerfile used for the bundle builds - `upstream.bundle.Dockerfile`

The bundle manifests are located in the `upstream_bundle` folder. 
These manifests can't be updated automatically because of new restrictions in Operator SDK 1.0. The CSV file needs to be updated *manually* using the contents of CSV file in `bundles` directory

The `build_olm_images_k8.sh` script copies the CRD (and other manifests) files from the `bundles` directory to `upstream_bundles` before making the bundle build.
This is done in order to avoid duplication of CRD files

## Submission

### Red Hat OpenShift Certified
A bundle image which points to the public image of the `dell-csi-operator` has to be submitted to Red Hat certification process

### Upstream Community Operators
`<TBD>`
