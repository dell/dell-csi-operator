package rbac

import (
	"context"
	"fmt"
	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewDummyClusterRole - returns the cluster role
func NewDummyClusterRole(name string) *rbacv1.ClusterRole {
	labels := map[string]string{
		"name": name,
	}
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
	return clusterRole
}

// NewControllerClusterRole - Returns a ClusterRole for the controller plugin
func NewControllerClusterRole(instance csiv1.CSIDriver, customClusterRoleName bool, haRequired bool, dummyClusterRole *rbacv1.ClusterRole) *rbacv1.ClusterRole {
	driverName := instance.GetName()
	driverNamespace := instance.GetNamespace()
	driverType := instance.GetDriverType()
	clusterRoleName := fmt.Sprintf("%s-controller", driverName)
	if customClusterRoleName {
		clusterRoleName = fmt.Sprintf("%s-%s-controller", driverNamespace, driverName)
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterRoleName,
			OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "create", "delete", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "create", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments/status"},
				Verbs:     []string{"patch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"csinodes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotcontents"},
				Verbs:     []string{"create", "get", "list", "watch", "update", "delete", "patch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshots/status"},
				Verbs:     []string{"watch", "update", "get", "list"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotcontents/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshots"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"create", "list", "watch", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims/status"},
				Verbs:     []string{"update", "patch"},
			},
		},
	}
	if haRequired {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
			Verbs:     []string{"create", "get", "list", "watch", "delete", "update"},
		})
	}
	if driverType == "powerstore" {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"storage.k8s.io"},
			Resources: []string{"csistoragecapacities"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		})
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"apps"},
			Resources: []string{"replicasets"},
			Verbs:     []string{"get"},
		})
	}
	return clusterRole
}

// SyncClusterRole - Syncs a ClusterRole
func SyncClusterRole(ctx context.Context, clusterRole *rbacv1.ClusterRole, client client.Client, reqLogger logr.Logger) (*rbacv1.ClusterRole, error) {
	found := &rbacv1.ClusterRole{}
	err := client.Get(ctx, types.NamespacedName{Name: clusterRole.Name, Namespace: clusterRole.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ClusterRole", "Name", clusterRole.Name)
		err = client.Create(ctx, clusterRole)
		if err != nil {
			return nil, err
		}
		// we need to return found object
		err := client.Get(ctx, types.NamespacedName{Name: clusterRole.Name, Namespace: clusterRole.Namespace}, found)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return nil, err
	} else {
		reqLogger.Info("Updating ClusterRole", "Name:", clusterRole.Name)
		err = client.Update(ctx, clusterRole)
		if err != nil {
			return nil, err
		}
	}

	return found, nil
}
