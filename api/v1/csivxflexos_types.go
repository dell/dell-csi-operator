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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CSIVXFlexOSSpec defines the desired state of CSIVXFlexOS
type CSIVXFlexOSSpec struct {
	// Driver is the specification for the CSI PowerFlex Driver
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Driver"
	Driver Driver `json:"driver" yaml:"driver"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=csivxflexoses,scope=Namespaced

// CSIVXFlexOS is the Schema for the csivxflexos API
// +operator-sdk:csv:customresourcedefinitions:displayName="CSI PowerFlex",resources={{Deployment,v1,vxflexos-controller},{DameonSet,v1,vxflexos-node}}
type CSIVXFlexOS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIVXFlexOSSpec `json:"spec,omitempty"`
	Status DriverStatus    `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CSIVXFlexOSList contains a list of CSIVXFlexOS
type CSIVXFlexOSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIVXFlexOS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSIVXFlexOS{}, &CSIVXFlexOSList{})
}
