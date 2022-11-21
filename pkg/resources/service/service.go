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
package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SyncService - Creates/Updates a service
func SyncService(ctx context.Context, service *corev1.Service, client client.Client, reqLogger logr.Logger) error {
	found := &corev1.Service{}
	err := client.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Name", service.Name)
		err = client.Create(ctx, service)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating Service", "Name:", service.Name)
		err = client.Update(ctx, service)
		if err != nil {
			return err
		}
	}

	return nil
}
