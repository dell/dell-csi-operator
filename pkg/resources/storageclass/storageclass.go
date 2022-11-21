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
package storageclass

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New - Returns a list of StorageClass objects
func New(instance csiv1.CSIDriver, customProvisionerName string, dummyClusterRole *rbacv1.ClusterRole) []*storagev1.StorageClass {
	var storageClass []*storagev1.StorageClass
	driver := instance.GetDriver()
	provisionerName := instance.GetDefaultDriverName()
	if customProvisionerName != "" {
		provisionerName = customProvisionerName
	}
	for _, sc := range driver.StorageClass {
		annotations := make(map[string]string)
		if sc.DefaultSc {
			annotations["storageclass.kubernetes.io/is-default-class"] = "true"
		}
		reclaimPolicy := corev1.PersistentVolumeReclaimDelete
		if sc.ReclaimPolicy == corev1.PersistentVolumeReclaimRetain {
			reclaimPolicy = corev1.PersistentVolumeReclaimRetain
		} else if sc.ReclaimPolicy == corev1.PersistentVolumeReclaimRecycle {
			reclaimPolicy = corev1.PersistentVolumeReclaimRecycle
		}
		ownerReferences := resources.GetDummyOwnerReferences(dummyClusterRole)
		// Over ride the blockowner deletion just for storage classes
		blockOwnerDeletion := false
		ownerReferences[0].BlockOwnerDeletion = &blockOwnerDeletion

		// Process storageClass attributes, if any
		var volumeBinding storagev1.VolumeBindingMode
		if sc.VolumeBindingMode == "Immediate" {
			volumeBinding = storagev1.VolumeBindingImmediate
		} else if sc.VolumeBindingMode == "WaitForFirstConsumer" {
			volumeBinding = storagev1.VolumeBindingWaitForFirstConsumer
		}

		storageClass = append(storageClass, &storagev1.StorageClass{
			Provisioner: provisionerName,
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("%s-%s", instance.GetName(), sc.Name),
				Annotations:     annotations,
				OwnerReferences: ownerReferences,
			},
			Parameters:           sc.Parameters,
			ReclaimPolicy:        &reclaimPolicy,
			VolumeBindingMode:    &volumeBinding,
			AllowVolumeExpansion: sc.AllowVolumeExpansion,
			AllowedTopologies:    sc.AllowedTopologies,
		})
	}
	return storageClass
}

// SyncStorageClass - Syncs StorageClass objects
func SyncStorageClass(ctx context.Context, instance csiv1.CSIDriver, storageClass []*storagev1.StorageClass,
	client client.Client, reqLogger logr.Logger, customProvisionerName string) []error {
	var errTemp = make([]error, 0)
	// List all storage classes
	existingStorageClasses := &storagev1.StorageClassList{}
	err := client.List(ctx, existingStorageClasses)
	if err != nil {
		errTemp = append(errTemp, err)
		return errTemp
	}
	scNames := make([]string, 0)
	// Form a list of SC names in the latest spec
	for _, sc := range storageClass {
		scNames = append(scNames, sc.Name)
	}
	// Form a list of SC names which we created in the past
	provisionerName := fmt.Sprintf("%s.dellemc.com", instance.GetPluginName())
	if customProvisionerName != "" {
		provisionerName = customProvisionerName
	}
	storageClassNamesToBeDeleted := make([]string, 0)
	for _, existingStorageClass := range existingStorageClasses.Items {
		if existingStorageClass.Provisioner == provisionerName {
			for _, ownerReference := range existingStorageClass.OwnerReferences {
				if ownerReference.Kind == instance.GetDriverTypeMeta().Kind {
					if !resources.IsStringInSlice(existingStorageClass.Name, scNames) {
						storageClassNamesToBeDeleted = append(
							storageClassNamesToBeDeleted, existingStorageClass.Name)
						break
					}
				}
			}
		}
	}
	for _, sc := range storageClass {
		// Check if this storage class already exists
		found := &storagev1.StorageClass{}
		err := client.Get(ctx, types.NamespacedName{Name: sc.Name}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new StorageClass", "StorageClass.Namespace", sc.Namespace, "StorageClass.Name", sc.Name)
			err = client.Create(ctx, sc)
			if err != nil {
				reqLogger.Error(err, "Updating StorageClass", "Name:", sc.Name)
				errTemp = append(errTemp, err)
			}
		} else if err != nil {
			reqLogger.Info("Unknown error.", "Error", err.Error())
			errTemp = append(errTemp, err)
		} else {
			reqLogger.Info("Updating StorageClass", "Name:", sc.Name)
			err = client.Update(ctx, sc)
			if err != nil {
				reqLogger.Error(err, "Updating StorageClass", "Name:", sc.Name)
				errTemp = append(errTemp, err)
			}
		}
	}
	// Delete any unwanted storage classes
	for _, scName := range storageClassNamesToBeDeleted {
		found := &storagev1.StorageClass{}
		err := client.Get(ctx, types.NamespacedName{Name: scName}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Storage class already deleted", scName)
		} else if err != nil {
			reqLogger.Error(err, "Failed to delete the storage class. Continuing")
		} else {
			err = client.Delete(ctx, found)
			if err != nil {
				reqLogger.Error(err, "Failed to delete the storage class. Continuing")
			} else {
				reqLogger.Info("Successfully deleted storage class", "Name:", scName)
			}
		}
	}
	return errTemp
}
