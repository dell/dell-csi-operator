/*


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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CSIUnitySpec defines the desired state of CSIUnity
type CSIUnitySpec struct {
	// Driver is the specification for the CSI Unity XT Driver
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Driver"
	Driver Driver `json:"driver" yaml:"driver"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=csiunities,scope=Namespaced

// CSIUnity is the Schema for the csiunities API
// +operator-sdk:csv:customresourcedefinitions:displayName="CSI Unity XT",resources={{Deployment,v1,unity-controller},{DameonSet,v1,unity-node}}
type CSIUnity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIUnitySpec `json:"spec,omitempty"`
	Status DriverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CSIUnityList contains a list of CSIUnity
type CSIUnityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIUnity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSIUnity{}, &CSIUnityList{})
}
