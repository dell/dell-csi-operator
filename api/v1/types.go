package v1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// +k8s:deepcopy-gen=false

// CSIDriver is the interface which extends the runtime object interfaces for each of the driver instances
type CSIDriver interface {
	GetNamespace() string
	SetNamespace(namespace string)
	GetName() string
	SetName(name string)
	GetGenerateName() string
	SetGenerateName(name string)
	GetUID() types.UID
	SetUID(uid types.UID)
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetGeneration() int64
	SetGeneration(generation int64)
	GetSelfLink() string
	SetSelfLink(selfLink string)
	GetCreationTimestamp() metav1.Time
	SetCreationTimestamp(timestamp metav1.Time)
	GetDeletionTimestamp() *metav1.Time
	SetDeletionTimestamp(timestamp *metav1.Time)
	GetDeletionGracePeriodSeconds() *int64
	SetDeletionGracePeriodSeconds(*int64)
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	GetFinalizers() []string
	SetFinalizers(finalizers []string)
	GetOwnerReferences() []metav1.OwnerReference
	SetOwnerReferences([]metav1.OwnerReference)
	GetClusterName() string
	SetClusterName(clusterName string)
	GetManagedFields() []metav1.ManagedFieldsEntry
	SetManagedFields(managedFields []metav1.ManagedFieldsEntry)
	GetDriver() *Driver
	GetDriverTypeMeta() *metav1.TypeMeta
	GetDriverType() DriverType
	GetDriverStatus() *DriverStatus
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
	GetResourcePlural() string
	GetPluginName() string
	GetCertVolumeName() string
	GetControllerName() string
	GetDaemonSetName() string
	GetUserEnvName() string
	GetPasswordEnvName() string
	GetDefaultDriverName() string
	GetDriverEnvName() string
}

// GetDriverTypeMeta - Returns TypeMeta for the driver
func (cr *CSIPowerMax) GetDriverTypeMeta() *metav1.TypeMeta {
	return &cr.TypeMeta
}

// GetDriver - Returns a pointer to the driver instance
func (cr *CSIPowerMax) GetDriver() *Driver {
	return &cr.Spec.Driver
}

// GetDriverStatus - Returns the driver status
func (cr *CSIPowerMax) GetDriverStatus() *DriverStatus {
	return &cr.Status
}

// GetDriverType - Returns the driver type
func (cr *CSIPowerMax) GetDriverType() DriverType {
	return PowerMax
}

// GetResourcePlural - Returns the plural form of the driver resource
func (cr *CSIPowerMax) GetResourcePlural() string {
	return "powermaxes"
}

// GetPluginName - Returns the plugin name
func (cr *CSIPowerMax) GetPluginName() string {
	return "csi-powermax"
}

// GetCertVolumeName - Returns the volume name for the TLS certs
func (cr *CSIPowerMax) GetCertVolumeName() string {
	return "certs"
}

// GetControllerName - Returns the name of the controller for the driver
func (cr *CSIPowerMax) GetControllerName() string {
	return fmt.Sprintf("%s-controller", cr.GetDriverType())
}

// GetDaemonSetName - Returns the name of the daemonset for the driver
func (cr *CSIPowerMax) GetDaemonSetName() string {
	return fmt.Sprintf("%s-node", cr.GetDriverType())
}

// GetUserEnvName - Returns the environment variable which is set to username in credentials secret
func (cr *CSIPowerMax) GetUserEnvName() string {
	return "X_CSI_POWERMAX_USER"
}

// GetPasswordEnvName - Returns the environment variable which is set to password in credentials secret
func (cr *CSIPowerMax) GetPasswordEnvName() string {
	return "X_CSI_POWERMAX_PASSWORD"
}

// GetDefaultDriverName - Returns the default driver name
func (cr *CSIPowerMax) GetDefaultDriverName() string {
	return fmt.Sprintf("%s.dellemc.com", cr.GetPluginName())
}

// GetDriverEnvName - Returns the name for the env which is used to set the driver name
func (cr *CSIPowerMax) GetDriverEnvName() string {
	return "X_CSI_POWERMAX_DRIVER_NAME"
}

// GetDriverTypeMeta - Returns TypeMeta for the driver
func (cr *CSIIsilon) GetDriverTypeMeta() *metav1.TypeMeta {
	return &cr.TypeMeta
}

// GetDriver - Returns a pointer to the driver instance
func (cr *CSIIsilon) GetDriver() *Driver {
	return &cr.Spec.Driver
}

// GetDriverStatus - Returns the driver status
func (cr *CSIIsilon) GetDriverStatus() *DriverStatus {
	return &cr.Status
}

// GetDriverType - Returns the driver type
func (cr *CSIIsilon) GetDriverType() DriverType {
	return Isilon
}

// GetResourcePlural - Returns the plural form of the driver resource
func (cr *CSIIsilon) GetResourcePlural() string {
	return "isilons"
}

// GetPluginName - Returns the plugin name
func (cr *CSIIsilon) GetPluginName() string {
	return "csi-isilon"
}

// GetCertVolumeName - Returns the volume name for the TLS certs
func (cr *CSIIsilon) GetCertVolumeName() string {
	return "certs"
}

// GetControllerName - Returns the name of the controller for the driver
func (cr *CSIIsilon) GetControllerName() string {
	return fmt.Sprintf("%s-controller", cr.GetDriverType())
}

// GetDaemonSetName - Returns the name of the daemonset for the driver
func (cr *CSIIsilon) GetDaemonSetName() string {
	return fmt.Sprintf("%s-node", cr.GetDriverType())
}

// GetUserEnvName - Returns the environment variable which is set to username in credentials secret
func (cr *CSIIsilon) GetUserEnvName() string {
	return "X_CSI_ISI_USER"
}

// GetPasswordEnvName - Returns the environment variable which is set to password in credentials secret
func (cr *CSIIsilon) GetPasswordEnvName() string {
	return "X_CSI_ISI_PASSWORD"
}

// GetDefaultDriverName - Returns the default driver name
func (cr *CSIIsilon) GetDefaultDriverName() string {
	return fmt.Sprintf("%s.dellemc.com", cr.GetPluginName())
}

// GetDriverEnvName - Returns the name for the env which is used to set the driver name
func (cr *CSIIsilon) GetDriverEnvName() string {
	return ""
}

// GetDriverTypeMeta - Returns TypeMeta for the driver
func (cr *CSIUnity) GetDriverTypeMeta() *metav1.TypeMeta {
	return &cr.TypeMeta
}

// GetDriver - Returns a pointer to the driver instance
func (cr *CSIUnity) GetDriver() *Driver {
	return &cr.Spec.Driver
}

// GetDriverStatus - Returns the driver status
func (cr *CSIUnity) GetDriverStatus() *DriverStatus {
	return &cr.Status
}

// GetDriverType - Returns the driver type
func (cr *CSIUnity) GetDriverType() DriverType {
	return Unity
}

// GetResourcePlural - Returns the plural form of the driver resource
func (cr *CSIUnity) GetResourcePlural() string {
	return "csiunities"
}

// GetPluginName - Returns the plugin name
func (cr *CSIUnity) GetPluginName() string {
	return "csi-unity"
}

// GetCertVolumeName - Returns the volume name for the TLS certs
func (cr *CSIUnity) GetCertVolumeName() string {
	return "certs"
}

// GetControllerName - Returns the name of the controller for the driver
func (cr *CSIUnity) GetControllerName() string {
	return fmt.Sprintf("%s-controller", cr.GetDriverType())
}

// GetDaemonSetName - Returns the name of the daemonset for the driver
func (cr *CSIUnity) GetDaemonSetName() string {
	return fmt.Sprintf("%s-node", cr.GetDriverType())
}

// GetUserEnvName - Returns the environment variable which is set to username in credentials secret
func (cr *CSIUnity) GetUserEnvName() string {
	return "X_CSI_UNITY_USER"
}

// GetPasswordEnvName - Returns the environment variable which is set to password in credentials secret
func (cr *CSIUnity) GetPasswordEnvName() string {
	return "X_CSI_UNITY_PASSWORD"
}

// GetDefaultDriverName - Returns the default driver name
func (cr *CSIUnity) GetDefaultDriverName() string {
	return fmt.Sprintf("%s.dellemc.com", cr.GetPluginName())
}

// GetDriverEnvName - Returns the name for the env which is used to set the driver name
func (cr *CSIUnity) GetDriverEnvName() string {
	return ""
}

// GetDriverTypeMeta - Returns TypeMeta for the driver
func (cr *CSIVXFlexOS) GetDriverTypeMeta() *metav1.TypeMeta {
	return &cr.TypeMeta
}

// GetDriver - Returns a pointer to the driver instance
func (cr *CSIVXFlexOS) GetDriver() *Driver {
	return &cr.Spec.Driver
}

// GetDriverStatus - Returns the driver status
func (cr *CSIVXFlexOS) GetDriverStatus() *DriverStatus {
	return &cr.Status
}

// GetDriverType - Returns the driver type
func (cr *CSIVXFlexOS) GetDriverType() DriverType {
	return "vxflexos"
}

// GetResourcePlural - Returns the plural form of the driver resource
func (cr *CSIVXFlexOS) GetResourcePlural() string {
	return "vxflexoses"
}

// GetPluginName - Returns the plugin name
func (cr *CSIVXFlexOS) GetPluginName() string {
	return "csi-vxflexos"
}

// GetCertVolumeName - Returns the volume name for the TLS certs
func (cr *CSIVXFlexOS) GetCertVolumeName() string {
	return "certs"
}

// GetControllerName - Returns the name of the controller for the driver
func (cr *CSIVXFlexOS) GetControllerName() string {
	return fmt.Sprintf("%s-controller", cr.GetDriverType())
}

// GetDaemonSetName - Returns the name of the daemonset for the driver
func (cr *CSIVXFlexOS) GetDaemonSetName() string {
	return fmt.Sprintf("%s-node", cr.GetDriverType())
}

// GetUserEnvName - Returns the environment variable which is set to username in credentials secret
func (cr *CSIVXFlexOS) GetUserEnvName() string {
	return "X_CSI_VXFLEXOS_USER"
}

// GetPasswordEnvName - Returns the environment variable which is set to password in credentials secret
func (cr *CSIVXFlexOS) GetPasswordEnvName() string {
	return "X_CSI_VXFLEXOS_PASSWORD"
}

// GetDefaultDriverName - Returns the default driver name
func (cr *CSIVXFlexOS) GetDefaultDriverName() string {
	return fmt.Sprintf("%s.dellemc.com", cr.GetPluginName())
}

// GetDriverEnvName - Returns the name for the env which is used to set the driver name
func (cr *CSIVXFlexOS) GetDriverEnvName() string {
	return ""
}

// GetDriverTypeMeta - Returns TypeMeta for the driver
func (cr *CSIPowerStore) GetDriverTypeMeta() *metav1.TypeMeta {
	return &cr.TypeMeta
}

// GetDriver - Returns a pointer to the driver instance
func (cr *CSIPowerStore) GetDriver() *Driver {
	return &cr.Spec.Driver
}

// GetDriverStatus - Returns the driver status
func (cr *CSIPowerStore) GetDriverStatus() *DriverStatus {
	return &cr.Status
}

// GetDriverType - Returns the driver type
func (cr *CSIPowerStore) GetDriverType() DriverType {
	return PowerStore
}

// GetResourcePlural - Returns the plural form of the driver resource
func (cr *CSIPowerStore) GetResourcePlural() string {
	return "powerstores"
}

// GetPluginName - Returns the plugin name
func (cr *CSIPowerStore) GetPluginName() string {
	return "csi-powerstore"
}

// GetCertVolumeName - Returns the volume name for the TLS certs
func (cr *CSIPowerStore) GetCertVolumeName() string {
	return "certs"
}

// GetControllerName - Returns the name of the controller for the driver
func (cr *CSIPowerStore) GetControllerName() string {
	return fmt.Sprintf("%s-controller", cr.GetDriverType())
}

// GetDaemonSetName - Returns the name of the daemonset for the driver
func (cr *CSIPowerStore) GetDaemonSetName() string {
	return fmt.Sprintf("%s-node", cr.GetDriverType())
}

// GetUserEnvName - Returns the environment variable which is set to username in credentials secret
func (cr *CSIPowerStore) GetUserEnvName() string {
	return "X_CSI_POWERSTORE_USER"
}

// GetPasswordEnvName - Returns the environment variable which is set to password in credentials secret
func (cr *CSIPowerStore) GetPasswordEnvName() string {
	return "X_CSI_POWERMAX_PASSWORD"
}

// GetDefaultDriverName - Returns the default driver name
func (cr *CSIPowerStore) GetDefaultDriverName() string {
	return fmt.Sprintf("%s.dellemc.com", cr.GetPluginName())
}

// GetDriverEnvName - Returns the name for the env which is used to set the driver name
func (cr *CSIPowerStore) GetDriverEnvName() string {
	return ""
}

// DriverType - type representing the type of the driver. e.g. - powermax, unity
type DriverType string

// K8sVersion - type representing the kubernetes version.
type K8sVersion string

// CSIOperatorConditionType defines the type of the last status update
type CSIOperatorConditionType string

// Constants for driver types and condition types
const (
	Unity          DriverType               = "unity"
	PowerMax       DriverType               = "powermax"
	VXFlexOS       DriverType               = "vxflexos"
	Isilon         DriverType               = "isilon"
	PowerStore     DriverType               = "powerstore"
	Common         DriverType               = "common"
	Unknown        DriverType               = "unknown"
	Succeeded      CSIOperatorConditionType = "Succeeded"
	InvalidConfig  CSIOperatorConditionType = "InvalidConfig"
	Running        CSIOperatorConditionType = "Running"
	Error          CSIOperatorConditionType = "Error"
	Updating       CSIOperatorConditionType = "Updating"
	Failed         CSIOperatorConditionType = "Failed"
	k8sv117        K8sVersion               = "v117"
	k8sv118        K8sVersion               = "v118"
	k8sv119        K8sVersion               = "v119"
	BaseK8sVersion K8sVersion               = "v117"
)

// SideCarType - type representing type of the sidecar container
type SideCarType string

// Constants for each of the sidecar container type
const (
	Provisioner   = "provisioner"
	Attacher      = "attacher"
	Snapshotter   = "snapshotter"
	Registrar     = "registrar"
	Resizer       = "resizer"
	Sdcmonitor    = "sdc-monitor"
	Healthmonitor = "external-health-monitor"
)

// InitContainerType - type representing type of initcontainer
type InitContainerType string

// Constants for powerflex container types
const (
	Sdc = "sdc"
)

// PodSchedulingConstraints - constraints applied on pod spec wrt to scheduling
type PodSchedulingConstraints struct {
	Tolerations  []corev1.Toleration
	NodeSelector map[string]string
}

// Driver of CSIDriver
// +k8s:openapi-gen=true
type Driver struct {

	// ConfigVersion is the configuration version of the driver
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Config Version"
	ConfigVersion string `json:"configVersion" yaml:"configVersion"`

	// Replicas is the count of controllers for Controller plugin
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Controller count"
	Replicas int32 `json:"replicas" yaml:"replicas"`

	// DNSPolicy is the dnsPolicy of the daemonset for Node plugin
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="DNSPolicy"
	DNSPolicy string `json:"dnsPolicy,omitempty" yaml:"dnsPolicy"`

	// FsGroupPolicy specifies fs group permission changes while mounting volume
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="FSGroupPolicy"
	FSGroupPolicy string `json:"fsGroupPolicy,omitempty" yaml:"fsGroupPolicy"`

	// Common is the common specification for both controller and node plugins
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Common specification"
	Common ContainerTemplate `json:"common" yaml:"common"`

	// Controller is the specification for Controller plugin only
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Controller Specification"
	Controller ContainerTemplate `json:"controller,omitempty" yaml:"controller"`

	// Node is the specification for Node plugin only
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Node specification"
	Node ContainerTemplate `json:"node,omitempty" yaml:"node"`

	// SideCars is the specification for CSI sidecar containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="CSI SideCars specification"
	SideCars []ContainerTemplate `json:"sideCars,omitempty" yaml:"sideCars"`

	// InitContainers is the specification for Driver InitContainers
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="InitContainers"
	InitContainers []ContainerTemplate `json:"initContainers,omitempty" yaml:"initContainers"`

	// StorageClass is the specification for Storage Classes
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage Classes"
	StorageClass []StorageClass `json:"storageClass,omitempty" yaml:"storageClass"`

	// SnapshotClass is the specification for Snapshot Classes
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Snapshot Classes"
	SnapshotClass []SnapshotClass `json:"snapshotClass,omitempty" yaml:"snapshotClass"`

	// ForceUpdate is the boolean flag used to force an update of the driver instance
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Force update"
	ForceUpdate bool `json:"forceUpdate,omitempty" yaml:"forceUpdate"`

	// AuthSecret is the name of the credentials secret for the driver
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Auth Secret"
	AuthSecret string `json:"authSecret,omitempty" yaml:"authSecret"`

	// TLSCertSecret is the name of the TLS Cert secret
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLSCert Secret"
	TLSCertSecret string `json:"tlsCertSecret,omitempty" yaml:"tlsCertSecret"`
}

// ContainerTemplate - Structure representing a container
// +k8s:openapi-gen=true
type ContainerTemplate struct {

	// Name is the name of Container
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Name"
	Name ImageType `json:"name,omitempty" yaml:"name"`

	// Image is the image tag for the Container
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Image"
	Image string `json:"image,omitempty" yaml:"image"`

	// ImagePullPolicy is the image pull policy for the image
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Image Pull Policy",xDescriptors="urn:alm:descriptor:com.tectonic.ui:imagePullPolicy"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Args is the set of arguments for the container
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Arguments"
	Args []string `json:"args,omitempty" yaml:"args"`

	// Envs is the set of environment variables for the container
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Environment vars"
	Envs []corev1.EnvVar `json:"envs,omitempty" yaml:"envs"`

	// Tolerations is the list of tolerations for the driver pods
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tolerations"
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" yaml:"tolerations"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="NodeSelector"
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector"`
}

// ImageType - represents type of image
type ImageType string

// Constants for image types
const (
	ImageTypeDriver        ImageType = "driver"
	ImageTypeProvisioner   ImageType = Provisioner
	ImageTypeAttacher      ImageType = Attacher
	ImageTypeRegistrar     ImageType = Registrar
	ImageTypeSnapshotter   ImageType = Snapshotter
	ImageTypeResizer       ImageType = Resizer
	ImageTypeSdcmonitor    ImageType = Sdcmonitor
	ImageTypeSdc           ImageType = Sdc
	ImageTypeHealthmonitor ImageType = Healthmonitor
)

// DriverState - type representing the state of the driver (in status)
type DriverState string

// PodStatus - Represents a list of PodStatus
type PodStatus struct {
	Available []string `json:"available,omitempty"`
	Ready     []string `json:"ready,omitempty"`
	Starting  []string `json:"starting,omitempty"`
	Stopped   []string `json:"stopped,omitempty"`
}

// DriverStatus defines the observed state of CSIDriver
// +k8s:openapi-gen=true
type DriverStatus struct {
	// ControllerStatus is the status of Controller pods
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="ControllerStatus",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	ControllerStatus PodStatus `json:"controllerStatus,omitempty"`

	// NodeStatus is the status of Controller pods
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="NodeStatus",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	NodeStatus PodStatus `json:"nodeStatus,omitempty"`

	// DriverHash is a hash of the driver specification
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="DriverHash",xDescriptors="urn:alm:descriptor:text"
	DriverHash uint64 `json:"driverHash,omitempty" yaml:"driverHash"`

	// State is the state of the driver installation
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="State",xDescriptors="urn:alm:descriptor:text"
	State DriverState `json:"state,omitempty" yaml:"state"`

	// LastUpdate is the last updated state of the driver
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="LastUpdate"
	LastUpdate LastUpdate `json:"lastUpdate,omitempty" yaml:"lastUpdate"`
}

// LastUpdate - Stores the last update condition for the driver status
// +k8s:openapi-gen=true
type LastUpdate struct {

	// Condition is the last known condition of the Custom Resource
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Condition",xDescriptors="urn:alm:descriptor:text"
	Condition CSIOperatorConditionType `json:"condition,omitempty" yaml:"type"`

	// Time is the time stamp for the last condition update
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Time",xDescriptors="urn:alm:descriptor:text"
	Time metav1.Time `json:"time,omitempty" yaml:"time"`

	// ErrorMessage is the last error message associated with the condition
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="ErrorMessage",xDescriptors="urn:alm:descriptor:text"
	ErrorMessage string `json:"errorMessage,omitempty" yaml:"errorMessage"`
}

// StorageClass represents a kubernetes storage class
// +k8s:openapi-gen=true
type StorageClass struct {
	// Name is the name of the StorageClass
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage Class Name"
	Name string `json:"name" yaml:"name"`

	// DefaultSc is a boolean flag to indicate if the storage class is going to be marked as default
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Default"
	DefaultSc bool `json:"default,omitempty" yaml:"default"`

	// ReclaimPolicy is the reclaim policy for the storage class
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="PersistentVolumeReclaimPolicy"
	ReclaimPolicy corev1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy,omitempty" yaml:"reclaimPolicy"`

	// Parameters is a map of driver specific storage class
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage Class Parameters"
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters"`

	// AllowVolumeExpansion is a boolean flag which indicates if volumes can be expanded
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Allow Volume Expansion"
	AllowVolumeExpansion *bool `json:"allowVolumeExpansion,omitempty" yaml:"allowVolumeExpansion"`

	// VolumeBindingMode field controls when volume binding and dynamic provisioning should occur.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Volume Binding Mode"
	VolumeBindingMode string `json:"volumeBindingMode,omitempty" yaml:"volumeBindingMode"`

	// Restrict the node topologies where volumes can be dynamically provisioned.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Allowed Topologies"
	AllowedTopologies []corev1.TopologySelectorTerm `json:"allowedTopologies,omitempty" yaml:"allowedTopologies"`
}

// SnapshotClass represents a VolumeSnapshotClass
// +k8s:openapi-gen=true
type SnapshotClass struct {
	// Name is the name of the Snapshot Class
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Snapshot Class Name"
	Name string `json:"name" yaml:"name"`

	// Parameters is a map of driver specific parameters for snapshot class
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Snapshot Class Parameters"
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters"`
}
