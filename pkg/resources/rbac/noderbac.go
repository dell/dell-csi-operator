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

package rbac

import (
	"fmt"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbachelper "k8s.io/kubernetes/pkg/apis/rbac/v1"
)

// NewNodeClusterRole - Returns a clusterRole for the Node plugin
func NewNodeClusterRole(instance csiv1.CSIDriver, customControllerName bool, dummyClusterRole *rbacv1.ClusterRole) *rbacv1.ClusterRole {
	driverName := instance.GetName()
	driverNameSpace := instance.GetNamespace()
	clusterRoleName := fmt.Sprintf("%s-node", driverName)
	if customControllerName {
		clusterRoleName = fmt.Sprintf("%s-%s-node", driverNameSpace, driverName)
	}
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterRoleName,
			OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
		},
		Rules: []rbacv1.PolicyRule{
			rbachelper.NewRule("list", "watch", "create", "update", "patch").Groups("").Resources("events").RuleOrDie(),
			rbachelper.NewRule("get", "list", "watch", "create", "update", "patch").Groups("").Resources("nodes").RuleOrDie(),
			rbachelper.NewRule("get", "list", "watch", "create", "delete", "update").Groups("").Resources("persistentvolumes").RuleOrDie(),
			rbachelper.NewRule("get", "list", "watch", "update").Groups("").Resources("persistentvolumeclaims").RuleOrDie(),
			rbachelper.NewRule("get", "list", "watch").Groups("storage.k8s.io").Resources("storageclasses").RuleOrDie(),
			rbachelper.NewRule("get", "list", "watch", "update").Groups("storage.k8s.io").Resources("volumeattachments").RuleOrDie(),
			rbachelper.NewRule("use").Groups("security.openshift.io").Resources("securitycontextconstraints").Names("privileged").RuleOrDie(),
		},
	}
	return clusterRole
}

// NewLimitedClusterRole - Returns a clusterRole for the Node plugin
func NewLimitedClusterRole(instance csiv1.CSIDriver, customControllerName bool, dummyClusterRole *rbacv1.ClusterRole) *rbacv1.ClusterRole {
	driverName := instance.GetName()
	driverNameSpace := instance.GetNamespace()
	clusterRoleName := fmt.Sprintf("%s-node", driverName)
	if customControllerName {
		clusterRoleName = fmt.Sprintf("%s-%s-node", driverNameSpace, driverName)
	}
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterRoleName,
			OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
		},
		Rules: []rbacv1.PolicyRule{
			rbachelper.NewRule("use").Groups("security.openshift.io").Resources("securitycontextconstraints").Names("privileged").RuleOrDie(),
		},
	}
	return clusterRole
}
