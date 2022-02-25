package csidriver

import (
	"context"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New - returns a instance of CSIDriver object
func New(instance csiv1.CSIDriver, ephemeralEnabled bool, dummyClusterRole *rbacv1.ClusterRole) *storagev1.CSIDriver {
	fsgrouppolicy := instance.GetDriver().FSGroupPolicy
	if fsgrouppolicy == "" {
		fsgrouppolicy = "ReadWriteOnceWithFSType"
	}
	fsgroup := storagev1.FSGroupPolicy(fsgrouppolicy)

	b := true
	modes := []storagev1.VolumeLifecycleMode{storagev1.VolumeLifecyclePersistent}

	if ephemeralEnabled {
		modes = append(modes, storagev1.VolumeLifecycleEphemeral)
	}

	spec := storagev1.CSIDriverSpec{
		AttachRequired:       &b,
		PodInfoOnMount:       &b,
		VolumeLifecycleModes: modes,
	}

	if instance.GetDriverType() == "powerstore" || instance.GetDriverType() == "isilon" {
		spec.FSGroupPolicy = &fsgroup
	}

	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetDefaultDriverName(),
			OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
		},
		Spec: spec,
	}
}

// SyncCSIDriver - Syncs a CSI Driver object
func SyncCSIDriver(ctx context.Context, csi *storagev1.CSIDriver, client client.Client, reqLogger logr.Logger) error {
	found := &storagev1.CSIDriver{}
	err := client.Get(ctx, types.NamespacedName{Name: csi.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new CSIDriver", "Name:", csi.Name)
		err = client.Create(ctx, csi)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		isUpdateRequired := false
		ownerRefs := found.GetOwnerReferences()
		for _, ownerRef := range ownerRefs {
			if ownerRef.APIVersion != "rbac.authorization.k8s.io/v1" {
				// Lets overwrite everything
				isUpdateRequired = true
				break
			}
		}
		if isUpdateRequired {
			found.OwnerReferences = csi.OwnerReferences
			err = client.Update(ctx, found)
			if err != nil {
				reqLogger.Error(err, "Failed to update CSIDriver object")
			} else {
				reqLogger.Info("Successfully updated CSIDriver object", "Name:", csi.Name)
			}
		}
	}
	return nil
}
