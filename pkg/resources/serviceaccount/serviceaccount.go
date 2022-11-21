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
package serviceaccount

import (
	"context"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New - Returns a ServiceAccount object
func New(instance csiv1.CSIDriver, saName string) *corev1.ServiceAccount {
	//var driver *csiv1.Driver = instance.GetDriver()
	driverNamespace := instance.GetNamespace()
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            saName,
			Namespace:       driverNamespace,
			OwnerReferences: resources.GetOwnerReferences(instance),
		},
	}
}

// SyncServiceAccount - Syncs a ServiceAccount
func SyncServiceAccount(ctx context.Context, sa *corev1.ServiceAccount, client client.Client, reqLogger logr.Logger) error {
	found := &corev1.ServiceAccount{}
	err := client.Get(ctx, types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "Namespace", sa.Namespace, "Name", sa.Name)
		err = client.Create(ctx, sa)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		// Updating the service account keeps regenerating the secrets.
		// We dont have to update the service account if it exists.
		reqLogger.Info("ServiceAccount already exists", "Name:", sa.Name)
	}
	return nil
}
