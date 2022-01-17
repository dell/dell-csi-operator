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

// NewControllerClusterRoleBindings - Returns a new ClusterRoleBinding for controller
func NewControllerClusterRoleBindings(instance csiv1.CSIDriver, customClusterRoleBinding bool, dummyClusterRole *rbacv1.ClusterRole) *rbacv1.ClusterRoleBinding {
	//var driver *csiv1.Driver = instance.GetDriver()
	driverType := instance.GetDriverType()
	driverName := instance.GetName()
	driverNamespace := instance.GetNamespace()
	clusterRoleBindingName := fmt.Sprintf("%s-controller", driverName)
	if customClusterRoleBinding {
		clusterRoleBindingName = fmt.Sprintf("%s-%s-controller", driverNamespace, driverName)
	}
	clusterRoleName := fmt.Sprintf("%s-controller", driverName)
	if customClusterRoleBinding {
		clusterRoleName = fmt.Sprintf("%s-%s-controller", driverNamespace, driverName)
	}
	clusterRoleBindings := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterRoleBindingName,
			OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      fmt.Sprintf("%s-controller", driverType),
			Namespace: driverNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
	}
	return clusterRoleBindings
}

// NewNodeClusterRoleBindings - Returns a new ClusterRoleBinding for the node plugin
func NewNodeClusterRoleBindings(instance csiv1.CSIDriver, customClusterRoleBinding bool, dummyClusterRole *rbacv1.ClusterRole) *rbacv1.ClusterRoleBinding {
	driverType := instance.GetDriverType()
	driverName := instance.GetName()
	driverNamespace := instance.GetNamespace()
	clusterRoleBindingName := fmt.Sprintf("%s-node", driverName)
	if customClusterRoleBinding {
		clusterRoleBindingName = fmt.Sprintf("%s-%s-node", driverNamespace, driverName)
	}
	clusterRoleName := fmt.Sprintf("%s-node", driverName)
	if customClusterRoleBinding {
		clusterRoleName = fmt.Sprintf("%s-%s-node", driverNamespace, driverName)
	}
	clusterRoleBindings := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterRoleBindingName,
			OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      fmt.Sprintf("%s-node", driverType),
			Namespace: driverNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
	}
	return clusterRoleBindings
}

// SyncClusterRoleBindings - Syncs the ClusterRoleBindings
func SyncClusterRoleBindings(ctx context.Context, rb *rbacv1.ClusterRoleBinding, client client.Client, reqLogger logr.Logger) error {
	found := &rbacv1.ClusterRoleBinding{}
	err := client.Get(ctx, types.NamespacedName{Name: rb.Name, Namespace: rb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ClusterRoleBinding", "Namespace", rb.Namespace, "Name", rb.Name)
		err = client.Create(ctx, rb)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating ClusterRoleBinding", "Name:", rb.Name)
		err = client.Update(ctx, rb)
		if err != nil {
			return err
		}
	}
	return nil
}

// SyncRoleBindings - Syncs the RoleBindings
func SyncRoleBindings(ctx context.Context, rb *rbacv1.RoleBinding, client client.Client, reqLogger logr.Logger) error {
	found := &rbacv1.RoleBinding{}
	err := client.Get(ctx, types.NamespacedName{Name: rb.Name, Namespace: rb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new RoleBinding", "Namespace", rb.Namespace, "Name", rb.Name)
		err = client.Create(ctx, rb)
		return err
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating RoleBinding", "Name:", rb.Name)
		err = client.Update(ctx, rb)
		if err != nil {
			return err
		}
	}
	return nil
}
