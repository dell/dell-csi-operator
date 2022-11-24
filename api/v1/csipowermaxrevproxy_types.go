/*
 Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProxyLimits is used for storing the various types of limits
// applied for a particular proxy instance
type ProxyLimits struct {
	MaxActiveRead       int `json:"maxActiveRead,omitempty" yaml:"maxActiveRead,omitempty"`
	MaxActiveWrite      int `json:"maxActiveWrite,omitempty" yaml:"maxActiveWrite,omitempty"`
	MaxOutStandingRead  int `json:"maxOutStandingRead,omitempty" yaml:"maxOutStandingRead,omitempty"`
	MaxOutStandingWrite int `json:"maxOutStandingWrite,omitempty" yaml:"maxOutStandingWrite,omitempty"`
}

// ManagementServerConfig - represents a management server configuration for the management server
type ManagementServerConfig struct {
	URL                       string      `json:"url" yaml:"url"`
	ArrayCredentialSecret     string      `json:"arrayCredentialSecret,omitempty" yaml:"arrayCredentialSecret,omitempty"`
	SkipCertificateValidation bool        `json:"skipCertificateValidation,omitempty" yaml:"skipCertificateValidation,omitempty"`
	CertSecret                string      `json:"certSecret,omitempty" yaml:"certSecret,omitempty"`
	Limits                    ProxyLimits `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// StorageArrayConfig represents a storage array managed by reverse proxy
type StorageArrayConfig struct {
	StorageArrayID         string   `json:"storageArrayId" yaml:"storageArrayId"`
	PrimaryURL             string   `json:"primaryURL" yaml:"primaryURL"`
	BackupURL              string   `json:"backupURL,omitempty" yaml:"backupURL,omitempty"`
	ProxyCredentialSecrets []string `json:"proxyCredentialSecrets" yaml:"proxyCredentialSecrets"`
}

// LinkConfig is one of the configuration modes for reverse proxy
type LinkConfig struct {
	Primary ManagementServerConfig `json:"primary" yaml:"primary"`
	Backup  ManagementServerConfig `json:"backup,omitempty" yaml:"backup,omitempty"`
}

// StandAloneConfig is one of the configuration modes for reverse proxy
type StandAloneConfig struct {
	StorageArrayConfig     []StorageArrayConfig     `json:"storageArrays" yaml:"storageArrays"`
	ManagementServerConfig []ManagementServerConfig `json:"managementServers" yaml:"managementServers"`
}

// RevProxyConfig represents the reverse proxy configuration
type RevProxyConfig struct {
	Mode             string            `json:"mode,omitempty" yaml:"mode,omitempty"`
	Port             int32             `json:"port,omitempty" yaml:"port,omitempty"`
	LinkConfig       *LinkConfig       `json:"linkConfig,omitempty" yaml:"linkConfig,omitempty"`
	StandAloneConfig *StandAloneConfig `json:"standAloneConfig,omitempty" yaml:"standAloneConfig,omitempty"`
}

// CSIPowerMaxRevProxySpec defines the desired state of CSIPowerMaxRevProxy
type CSIPowerMaxRevProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image           string            `json:"image" yaml:"image"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	TLSSecret       string            `json:"tlsSecret" yaml:"tlsSecret"`
	RevProxy        RevProxyConfig    `json:"config" yaml:"config"`
}

// CSIPowerMaxRevProxyStatus defines the observed state of CSIPowerMaxRevProxy
type CSIPowerMaxRevProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// ProxyStatus is the status of proxy pod
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="ProxyStatus"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	ProxyStatus PodStatus `json:"proxyStatus,omitempty"`

	// DriverHash is a hash of the driver specification
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="ProxyHash"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:text"
	ProxyHash uint64 `json:"proxyHash,omitempty" yaml:"proxyHash"`

	// State is the state of the driver installation
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="State"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:text"
	State DriverState `json:"state,omitempty" yaml:"state"`

	// LastUpdate is the last updated state of the driver
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="LastUpdate"
	LastUpdate LastUpdate `json:"lastUpdate,omitempty" yaml:"lastUpdate"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=csipowermaxrevproxies,scope=Namespaced

// CSIPowerMaxRevProxy is the Schema for the csipowermaxrevproxies API
// +operator-sdk:csv:customresourcedefinitions:displayName="CSI PowerMax ReverseProxy",resources={{Deployment,v1,powermax-reverseproxy}}
type CSIPowerMaxRevProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIPowerMaxRevProxySpec   `json:"spec,omitempty"`
	Status CSIPowerMaxRevProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CSIPowerMaxRevProxyList contains a list of CSIPowerMaxRevProxy
type CSIPowerMaxRevProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIPowerMaxRevProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSIPowerMaxRevProxy{}, &CSIPowerMaxRevProxyList{})
}
