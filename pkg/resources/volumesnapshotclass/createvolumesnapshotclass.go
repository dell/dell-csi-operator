package volumesnapshotclass

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/types"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New - Returns a list of VolumeSnapshotClass objects
func New(instance csiv1.CSIDriver, customSnapshotterName string, dummyClusterRole *rbacv1.ClusterRole) []*v1.VolumeSnapshotClass {
	var vsClass []*v1.VolumeSnapshotClass
	driver := instance.GetDriver()
	snapshotterName := fmt.Sprintf("%s.dellemc.com", instance.GetPluginName())
	if customSnapshotterName != "" {
		snapshotterName = customSnapshotterName
	}
	for _, vc := range driver.SnapshotClass {
		sc := &v1.VolumeSnapshotClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("%s-%s", instance.GetName(), vc.Name),
				OwnerReferences: resources.GetDummyOwnerReferences(dummyClusterRole),
			},
			Driver:         snapshotterName,
			Parameters:     vc.Parameters,
			DeletionPolicy: "Delete",
		}
		sc.APIVersion = "snapshot.storage.k8s.io/v1beta1"
		sc.Kind = "VolumeSnapshotClass"
		vsClass = append(vsClass, sc)
	}
	return vsClass
}

// SyncSnapshotClass - Syncs snapshot class objects
func SyncSnapshotClass(ctx context.Context, instance csiv1.CSIDriver, snapClass []*v1.VolumeSnapshotClass,
	client client.Client, reqLogger logr.Logger, customDriverName string) []error {
	var errTemp = make([]error, 0)
	driverName := fmt.Sprintf("%s.dellemc.com", instance.GetPluginName())
	if customDriverName != "" {
		driverName = customDriverName
	}
	// List all snapshot classes
	existingSnapshotClasses := &v1.VolumeSnapshotClassList{}
	err := client.List(ctx, existingSnapshotClasses)
	if err != nil {
		errTemp = append(errTemp, err)
		return errTemp
	}
	scNames := make([]string, 0)
	// Form a list of SnapshotClass names in the latest spec
	for _, sc := range snapClass {
		scNames = append(scNames, sc.Name)
	}
	// Form a list of SnapshotClass names which we created in the past
	snapshotClassNamesToBeDeleted := make([]string, 0)
	for _, existingSnapshotClass := range existingSnapshotClasses.Items {
		if existingSnapshotClass.Driver == driverName {
			for _, ownerReference := range existingSnapshotClass.OwnerReferences {
				if ownerReference.Kind == instance.GetDriverTypeMeta().Kind {
					if !resources.IsStringInSlice(existingSnapshotClass.Name, scNames) {
						snapshotClassNamesToBeDeleted = append(
							snapshotClassNamesToBeDeleted, existingSnapshotClass.Name)
						break
					}
				}
			}
		}
	}

	for _, sc := range snapClass {
		// Check if this snapshot class already exists
		found := &v1.VolumeSnapshotClass{}
		err := client.Get(ctx, types.NamespacedName{Name: sc.Name}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new SnapshotClass", "SnapshotClass.Namespace", sc.Namespace, "SnapshotClass.Name", sc.Name)
			err = client.Create(ctx, sc)
			if err != nil {
				reqLogger.Error(err, "Updating SnapshotClass", "Name:", sc.Name)
				errTemp = append(errTemp, err)
			}
		} else if err != nil {
			reqLogger.Info("Unknown error.", "Error", err.Error())
			errTemp = append(errTemp, err)
		} else {
			reqLogger.Info("Updating SnapshotClass", "Name:", sc.Name)
			sc.ResourceVersion = found.ResourceVersion
			err = client.Update(ctx, sc)
			if err != nil {
				reqLogger.Error(err, "Updating SnapshotClass", "Name:", sc.Name)
				errTemp = append(errTemp, err)
			}
		}
	}
	// Delete any unwanted snapshot classes
	for _, scName := range snapshotClassNamesToBeDeleted {
		found := &v1.VolumeSnapshotClass{}
		err := client.Get(ctx, types.NamespacedName{Name: scName}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Snapshot class already deleted", scName)
		} else if err != nil {
			reqLogger.Error(err, "Failed to delete the snapshot class. Continuing")
		} else {
			err = client.Delete(ctx, found)
			if err != nil {
				reqLogger.Error(err, "Failed to delete the snapshot class. Continuing")
			} else {
				reqLogger.Info("Successfully deleted snapshot class", "SnapshotClass", scName)
			}
		}
	}
	return errTemp
}
