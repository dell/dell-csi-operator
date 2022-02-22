/*


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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"sigs.k8s.io/yaml"

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
	"github.com/dell/dell-csi-operator/pkg/resources/secrets"
)

// CSIUnityReconciler reconciles a CSIUnity object
type CSIUnityReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Config      operatorconfig.Config
	updateCount int32
}

// +kubebuilder:rbac:groups=storage.dell.com,resources=csiunities;csiunities/finalizers;csiunities/status,verbs=*
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
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotclasses;volumesnapshotcontents,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotcontents/status,verbs=update
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshots;volumesnapshots/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=create;list;watch;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=volumeattachments/status,verbs=patch
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,resourceNames=privileged,verbs=use

// Reconcile function reconciles a CSIUnity Object
func (r *CSIUnityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("csiunity", req.NamespacedName)

	csiUnity := &storagev1.CSIUnity{}
	return utils.Reconcile(ctx, csiUnity, req, r, log)
}

// SetupWithManager - sets up controller
func (r *CSIUnityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("CSIUnity", mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.Log.Error(err, "Unable to setup CSIUnity controller")
		os.Exit(1)
	}

	err = c.Watch(
		&source.Kind{Type: &storagev1.CSIUnity{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		r.Log.Error(err, "Unable to watch CSIUnity Driver")
		os.Exit(1)
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIUnity{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Deployment")
		os.Exit(1)
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIUnity{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Daemonset")
		os.Exit(1)
	}
	return nil
}

// GetConfig - returns the config
func (r *CSIUnityReconciler) GetConfig() operatorconfig.Config {
	return r.Config
}

// GetClient - returns the split client
func (r *CSIUnityReconciler) GetClient() client.Client {
	return r.Client
}

// GetScheme - Returns k8s runtime scheme
func (r *CSIUnityReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// GetUpdateCount - Returns the current update count
func (r *CSIUnityReconciler) GetUpdateCount() int32 {
	return r.updateCount
}

// SetClient - Sets the split client (only for testing)
func (r *CSIUnityReconciler) SetClient(c client.Client) {
	r.Client = c
}

// SetScheme - Sets k8s runtime scheme (only for testing)
func (r *CSIUnityReconciler) SetScheme(s *runtime.Scheme) {
	r.Scheme = s
}

// SetConfig - Sets the config
func (r *CSIUnityReconciler) SetConfig(c operatorconfig.Config) {
	r.Config = c
}

//IncrUpdateCount - Increments the update count
func (r *CSIUnityReconciler) IncrUpdateCount() {
	atomic.AddInt32(&r.updateCount, 1)
}

// InitializeDriverSpec - Initializes any uninitialized elements of the instance spec.
// Also initialize the defaults if user didn't set the values
func (r *CSIUnityReconciler) InitializeDriverSpec(instance storagev1.CSIDriver, reqLogger logr.Logger) (bool, error) {
	//@TODO User need to perform parameter/env/args validations here
	return false, nil
}

// ValidateDriverSpec does driver specific validation of the spec
func (r *CSIUnityReconciler) ValidateDriverSpec(ctx context.Context, instance storagev1.CSIDriver, reqLogger logr.Logger) error {
	driver := instance.GetDriver()

	err := r.validateMultiArrayUnityCredsSecret(ctx, instance, reqLogger)
	if err != nil {
		return err
	}

	scs := driver.StorageClass
	for _, sc := range scs {
		scParams := sc.Parameters
		if instance.GetDriverType() == storagev1.Unity && instance.GetDriver().ConfigVersion == "v2" {
			pool, ok := scParams["storagePool"]
			if !ok {
				return fmt.Errorf("storagePool paramter is mandatory in StorageClass [%s]", sc.Name)
			}

			if pool == "" {
				return fmt.Errorf("storagePool paramter should not be empty in StorageClass [%s]", sc.Name)
			}

			arrayID, ok := scParams["arrayId"]
			if !ok {
				return fmt.Errorf("arrayId paramter is mandatory in StorageClass [%s]", sc.Name)
			}

			if arrayID == "" {
				return fmt.Errorf("arrayId paramter should not be empty in StorageClass [%s]", sc.Name)
			}
		}
		val, ok := scParams["tieringPolicy"]
		if ok {
			i, err := strconv.Atoi(val)
			if err != nil {
				return errors.New("tieringPolicy should be numeric and values should be 0,1,2 for instance " + instance.GetName())
			}

			if i < 0 || i > 2 {
				return errors.New("tieringPolicy should be numeric and values should be 0,1,2 for instance " + instance.GetName())
			}
		}
	}
	return nil
}

func (r *CSIUnityReconciler) validateMultiArrayUnityCredsSecret(ctx context.Context, instance storagev1.CSIDriver, log logr.Logger) error {
	client := r.GetClient()
	secretName := fmt.Sprintf("%s-creds", instance.GetDriverType())
	credSecret, err := secrets.GetSecret(ctx, secretName, instance.GetNamespace(), client, log)
	if err != nil {
		return fmt.Errorf("reading secret [%s] error [%s]", secretName, err)
	}

	type StorageArrayConfig struct {
		ArrayID                   string `yaml:"arrayId"`
		Username                  string `yaml:"username"`
		Password                  string `yaml:"password"`
		RestGateway               string `yaml:"restGateway"`
		Insecure                  bool   `yaml:"insecure,omitempty"`
		IsDefaultArray            bool   `yaml:"isDefaultArray,omitempty"`
		Endpoint                  string `yaml:"endpoint,omitempty"`
		IsDefault                 bool   `yaml:"isDefault,omitempty"`
		SkipCertificateValidation bool   `yaml:"skipCertificateValidation,omitempty"`
	}

	//To parse the secret yaml file
	type StorageArrayList struct {
		StorageArrayList []StorageArrayConfig `yaml:"storageArrayList"`
	}
	data := credSecret.Data
	configBytes := data["config"]

	if string(configBytes) != "" {
		secretConfig := new(StorageArrayList)

		if instance.GetDriver().ConfigVersion == "v4" || instance.GetDriver().ConfigVersion == "v5" {
			err := json.Unmarshal(configBytes, &secretConfig)
			if err != nil {
				return fmt.Errorf("Unable to parse the credentials [%v]", err)
			}
		} else {
			err := yaml.Unmarshal(configBytes, &secretConfig)
			if err != nil {
				return fmt.Errorf("Unable to parse the credentials [%v]", err)
			}
		}

		if len(secretConfig.StorageArrayList) == 0 {
			return fmt.Errorf("Arrays details are not provided in unity-creds secret")
		}

		var noOfDefaultArrays int
		tempMapToFindDuplicates := make(map[string]interface{}, 0)
		for i, config := range secretConfig.StorageArrayList {
			if config.ArrayID == "" {
				return fmt.Errorf("invalid value for ArrayID at index [%d]", i)
			}
			if config.Username == "" {
				return fmt.Errorf("invalid value for Username at index [%d]", i)
			}
			if config.Password == "" {
				return fmt.Errorf("invalid value for Password at index [%d]", i)
			}
			if config.RestGateway == "" && config.Endpoint == "" {
				return fmt.Errorf("invalid value for RestGateway at index [%d]", i)
			}

			if _, ok := tempMapToFindDuplicates[config.ArrayID]; ok {
				return fmt.Errorf("Duplicate ArrayID [%s] found in storageArrayList parameter", config.ArrayID)
			}
			tempMapToFindDuplicates[config.ArrayID] = nil

			if config.IsDefaultArray || config.IsDefault {
				noOfDefaultArrays++
			}

			if noOfDefaultArrays > 1 {
				return fmt.Errorf("'isDefaultArray' parameter located in multiple places ArrayID: %s. 'isDefaultArray' parameter should present only once in the storageArrayList", config.ArrayID)
			}
		}
	} else {
		return fmt.Errorf("Arrays details are not provided in unity-creds secret")
	}
	return nil
}
