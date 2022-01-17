package rbac

import (
	"context"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SyncRole - Creates/Updates a Role
func SyncRole(ctx context.Context, role *rbacv1.Role, client client.Client, reqLogger logr.Logger) error {
	found := &rbacv1.Role{}
	err := client.Get(ctx, types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Role", "Name", role.Name)
		err = client.Create(ctx, role)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating Role", "Name:", role.Name)
		err = client.Update(ctx, role)
		if err != nil {
			return err
		}
	}

	return nil
}
