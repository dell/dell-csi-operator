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

package controllers

import (
	"context"
	"os"
	"sync/atomic"

	"github.com/dell/dell-csi-operator/pkg/utils"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	storagev1 "github.com/dell/dell-csi-operator/api/v1"
	operatorconfig "github.com/dell/dell-csi-operator/pkg/config"
)

// CSIIsilonReconciler reconciles a CSIIsilon object
type CSIIsilonReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Config      operatorconfig.Config
	updateCount int32
}

// +kubebuilder:rbac:groups=storage.dell.com,resources=csiisilons;csiisilons/finalizers;csiisilons/status,verbs=*
// +kubebuilder:rbac:groups="",resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts,verbs=*
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;patch;update
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims/status,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;delete;patch;update
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;replicasets;rolebindings,verbs=get;list;watch;update;create;delete;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles/finalizers,verbs=get;list;watch;update;create;delete;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;update;create;delete;patch
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups="apps",resources=deployments/finalizers,resourceNames=dell-csi-operator-controller-manager,verbs=update
// +kubebuilder:rbac:groups="storage.k8s.io",resources=csidrivers,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups="storage.k8s.io",resources=storageclasses,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=volumeattachments,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="storage.k8s.io",resources=csinodes,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotclasses,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotcontents,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotcontents/status,verbs=update;patch
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshots;volumesnapshots/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=create;list;watch;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=volumeattachments/status,verbs=patch
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,resourceNames=privileged,verbs=use

// Reconcile function reconciles a CSIIsilon object
func (r *CSIIsilonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("csiisilon", req.NamespacedName)

	// your logic here
	csiIsilon := new(storagev1.CSIIsilon)
	return utils.Reconcile(ctx, csiIsilon, req, r, log)
}

// GetConfig - returns the config
func (r *CSIIsilonReconciler) GetConfig() operatorconfig.Config {
	return r.Config
}

// GetClient - returns the split client
func (r *CSIIsilonReconciler) GetClient() client.Client {
	return r.Client
}

// GetScheme - Returns k8s runtime scheme
func (r *CSIIsilonReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// GetUpdateCount - Returns the current update count
func (r *CSIIsilonReconciler) GetUpdateCount() int32 {
	return r.updateCount
}

// SetClient - Sets the split client (only for testing)
func (r *CSIIsilonReconciler) SetClient(c client.Client) {
	r.Client = c
}

// SetScheme - Sets k8s runtime scheme (only for testing)
func (r *CSIIsilonReconciler) SetScheme(s *runtime.Scheme) {
	r.Scheme = s
}

// SetConfig - returns the config (only for testing)
func (r *CSIIsilonReconciler) SetConfig(c operatorconfig.Config) {
	r.Config = c
}

// IncrUpdateCount - Increments the update count
func (r *CSIIsilonReconciler) IncrUpdateCount() {
	atomic.AddInt32(&r.updateCount, 1)
}

// InitializeDriverSpec - Initializes any uninitialized elements of the instance spec.
// Also initialize the defaults if user didn't set the values
func (r *CSIIsilonReconciler) InitializeDriverSpec(instance storagev1.CSIDriver, reqLogger logr.Logger) (bool, error) {
	//@TODO User need to perform parameter/env/args validations here
	return false, nil
}

// ValidateDriverSpec - Validates the driver spec
// returns error if the spec is not valid
func (r *CSIIsilonReconciler) ValidateDriverSpec(ctx context.Context, instance storagev1.CSIDriver, reqLogger logr.Logger) error {
	//Return nil, if the driver do not want to validate any params
	return nil
}

// SetupWithManager - sets up the controller
func (r *CSIIsilonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("CSIIsilon", mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.Log.Error(err, "Unable to setup CSIIsilon controller")
		os.Exit(1)
	}

	err = c.Watch(
		&source.Kind{Type: &storagev1.CSIIsilon{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		r.Log.Error(err, "Unable to watch CSIIsilon Driver")
		os.Exit(1)
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIIsilon{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Deployment")
		os.Exit(1)
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIIsilon{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Daemonset")
		os.Exit(1)
	}
	return nil
}
