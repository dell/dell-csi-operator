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
package utils

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dell/dell-csi-operator/pkg/resources/deployment"
	"github.com/dell/dell-csi-operator/pkg/resources/statefulset"

	k8serror "k8s.io/apimachinery/pkg/api/errors"

	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dell/dell-csi-operator/pkg/config"
	"github.com/dell/dell-csi-operator/pkg/ctrlconfig"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/dell/dell-csi-operator/pkg/resources/csidriver"
	"github.com/dell/dell-csi-operator/pkg/resources/daemonset"
	"github.com/dell/dell-csi-operator/pkg/resources/rbac"
	"github.com/dell/dell-csi-operator/pkg/resources/secrets"
	"github.com/dell/dell-csi-operator/pkg/resources/serviceaccount"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcileCSI is the interface which extends each of the respective Reconcile interfaces
// for drivers
type ReconcileCSI interface {
	reconcile.Reconciler
	GetClient() crclient.Client
	GetScheme() *runtime.Scheme
	GetConfig() config.Config
	SetClient(crclient.Client)
	SetScheme(*runtime.Scheme)
	SetConfig(config.Config)
	GetUpdateCount() int32
	IncrUpdateCount()
	InitializeDriverSpec(instance csiv1.CSIDriver, reqLogger logr.Logger) (bool, error)
	ValidateDriverSpec(ctx context.Context, instance csiv1.CSIDriver, reqLogger logr.Logger) error
}

// MetadataPrefix - prefix for all labels & annotations
const MetadataPrefix = "storage.dell.com"

var configVersionKey = fmt.Sprintf("%s/%s", MetadataPrefix, "CSIDriverConfigVersion")

func checkAndApplyConfigVersionAnnotations(instance csiv1.CSIDriver, log logr.Logger, update bool) (bool, error) {
	driver := instance.GetDriver()
	if driver.ConfigVersion == "" {
		// fail immediately
		return false, fmt.Errorf("mandatory argument: ConfigVersion missing")
	}
	// If driver has not been initialized yet, we first annotate the driver with the config version annotation
	if instance.GetDriverStatus().DriverHash == 0 || update {
		annotations := instance.GetAnnotations()
		isUpdated := false
		if annotations == nil {
			annotations = make(map[string]string)
			isUpdated = true
		}
		if configVersion, ok := annotations[configVersionKey]; !ok {
			annotations[configVersionKey] = driver.ConfigVersion
			isUpdated = true
			instance.SetAnnotations(annotations)
			log.Info(fmt.Sprintf("Installing CSI Driver %s with config Version %s. Updating Annotations with Config Version",
				instance.GetName(), driver.ConfigVersion))
		} else {
			if configVersion != driver.ConfigVersion {
				annotations[configVersionKey] = driver.ConfigVersion
				isUpdated = true
				instance.SetAnnotations(annotations)
				log.Info(fmt.Sprintf("Config Version changed from %s to %s. Updating Annotations",
					configVersion, driver.ConfigVersion))
			}
		}
		return isUpdated, nil
	}
	return false, nil
}

func deleteDummyClusterRoleAndRemoveFinalizer(ctx context.Context, instance csiv1.CSIDriver,
	r ReconcileCSI, log logr.Logger) (reconcile.Result, error) {
	dummyClusterRoleName := getDummyClusterRoleName(instance)
	found := &rbacv1.ClusterRole{}
	err := r.GetClient().Get(
		ctx,
		types.NamespacedName{
			Name:      dummyClusterRoleName,
			Namespace: "",
		}, found)
	if err != nil && k8serror.IsNotFound(err) {
		// Not found
		// We have a finalizer set but no dummy clusterrole
		// Try to sync the driver again to update any obsolete ownerreferences
		configVersion := instance.GetDriver().ConfigVersion
		configDirectory := r.GetConfig().ConfigDirectory
		driverConfig := &ctrlconfig.Config{
			ConfigVersion:  configVersion,
			KubeAPIVersion: r.GetConfig().KubeAPIServerVersion,
			DriverType:     instance.GetDriverType(),
			Log:            log,
			IsOpenShift:    r.GetConfig().IsOpenShift,
			ConfigFileName: r.GetConfig().ConfigFile,
		}
		err = driverConfig.InitDriverConfig(configDirectory)
		if err != nil || instance.GetDriverStatus().State == constants.InvalidConfig {
			// Unable to initialize config or the state is already invalid config
			// Remove finalizer, log a warning and return
			log.Info("Warning: Objects with invalid OwnerReference may be left behind in the cluster. Delete them manually")
			instance.SetFinalizers(nil)
			// Update the object
			err = r.GetClient().Update(ctx, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			log.Info("Successfully removed the finalizer")
			return reconcile.Result{}, nil
		}
		// Update the object
		syncErr := SyncDriver(ctx, instance, r, driverConfig, log)
		if syncErr == nil {
			err = r.GetClient().Delete(ctx, found)
			if err != nil {
				log.Error(err, "Failed to delete the dummy clusterrole")
				return reconcile.Result{}, err
			}
			log.Info("Successfully deleted the dummy clusterRole", "Name", dummyClusterRoleName)
		} else {
			log.Error(err, "Failed to sync driver")
			return reconcile.Result{}, err
		}
	} else if err != nil {
		// Lets requeue
		// Requeue the request
		return reconcile.Result{}, err
	}
	// We found it
	err = r.GetClient().Delete(ctx, found)
	if err != nil {
		log.Error(err, "failed to delete the dummy clusterrole")
		return reconcile.Result{}, err
	}
	log.Info("Successfully deleted the dummy clusterRole", "Name", dummyClusterRoleName)
	// Remove the finalizers
	instance.SetFinalizers(nil)
	// Update the object
	err = r.GetClient().Update(ctx, instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// Reconcile - Common Reconcile method for all drivers
func Reconcile(ctx context.Context, instance csiv1.CSIDriver, request reconcile.Request, r ReconcileCSI, log logr.Logger) (reconcile.Result, error) {
	driverType := instance.GetDriverType()
	reqLogger := log.WithValues("Namespace", request.Namespace)
	reqLogger = reqLogger.WithValues("Name", request.Name)
	reqLogger = reqLogger.WithValues("Attempt", r.GetUpdateCount())
	reqLogger.Info(fmt.Sprintf("Reconciling %s ", driverType), "request", request.String())

	retryInterval := constants.DefaultRetryInterval
	// atomic.AddInt32(&r.updateCount, 1)
	reqLogger.Info("################Starting Reconcile##############")
	r.IncrUpdateCount()
	var err error
	// Fetch the CSIDriver instance
	err = r.GetClient().Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if k8serror.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	isCustomResourceMarkedForDeletion := instance.GetDeletionTimestamp() != nil
	if isCustomResourceMarkedForDeletion {
		return deleteDummyClusterRoleAndRemoveFinalizer(ctx, instance, r, reqLogger)
	}
	// Add finalizer
	instance.SetFinalizers([]string{"finalizer.dell.emc.com"})
	// Update CR
	err = r.GetClient().Update(ctx, instance)
	if err != nil {
		reqLogger.Error(err, "Failed to update CR with finalizer")
		return reconcile.Result{}, err
	}

	configVersion := instance.GetDriver().ConfigVersion
	configDirectory := r.GetConfig().ConfigDirectory
	driverConfig := &ctrlconfig.Config{
		ConfigVersion:  configVersion,
		KubeAPIVersion: r.GetConfig().KubeAPIServerVersion,
		DriverType:     instance.GetDriverType(),
		Log:            log,
		IsOpenShift:    r.GetConfig().IsOpenShift,
		ConfigFileName: r.GetConfig().ConfigFile,
	}

	err = driverConfig.InitDriverConfig(configDirectory)
	if err != nil {
		log.Error(err, "Failed to initialize driver config")
		return handleValidationError(ctx, instance, driverConfig, r, reqLogger, err)
	}

	// Before doing anything else, check for config version and apply annotation if not set
	isUpdated, err := checkAndApplyConfigVersionAnnotations(instance, log, false)
	if err != nil {
		return handleValidationError(ctx, instance, driverConfig, r, reqLogger, err)
	} else if isUpdated {
		_ = r.GetClient().Update(ctx, instance)
		return reconcile.Result{Requeue: true}, nil
	}

	status := instance.GetDriverStatus()
	// newStatus is the status object which is modified and finally used to update the Status
	// in case the instance or the status is updated
	newStatus := status.DeepCopy()
	// oldStatus is the previous status of the CR instance
	// This is used to compare if there is a need to update the status
	oldStatus := status.DeepCopy()
	oldState := oldStatus.State
	reqLogger.Info(fmt.Sprintf("Driver was previously in (%s) state", string(oldState)))

	// Check if the driver has changed
	expectedHash, actualHash, changed := driverChanged(instance)
	if changed {
		message := fmt.Sprintf("Driver spec has changed (%d vs %d)", actualHash, expectedHash)
		newStatus.DriverHash = expectedHash
		reqLogger.Info(message)
	} else {
		reqLogger.Info("No changes detected in the driver spec")
	}
	// Check if force update was requested
	forceUpdate := instance.GetDriver().ForceUpdate
	checkStateOnly := false
	switch oldState {
	case constants.Running:
		fallthrough
	case constants.Succeeded:
		if changed {
			// If the driver hash has changed, we need to update the driver again
			newStatus.State = constants.Updating
			reqLogger.Info("Changed state to Updating as driver spec changed")
		} else {
			// Just check the state of the driver and update status accordingly
			reqLogger.Info("Recalculating driver state(only) as there is no change in driver spec")
			checkStateOnly = true
		}
	case constants.InvalidConfig:
		fallthrough
	case constants.Failed:
		// Check if force update was requested
		if forceUpdate {
			reqLogger.Info("Force update requested")
			newStatus.State = constants.Updating
		} else {
			if changed {
				// Do a reconcile as we detected a change
				newStatus.State = constants.Updating
			} else {
				reqLogger.Info(fmt.Sprintf("CR is in (%s) state. Reconcile request won't be requeued",
					newStatus.State))
				return logBannerAndReturn(reconcile.Result{}, nil, reqLogger)
			}
		}
	case constants.NoState:
		newStatus.State = constants.Updating
	case constants.Updating:
		reqLogger.Info("Driver already in Updating state")
	}

	// Always initialize the spec
	isUpdated, err = InitializeSpec(instance, r, driverConfig, reqLogger)
	if err != nil {
		log.Error(err, "Failed to initialize common spec")
		return handleValidationError(ctx, instance, driverConfig, r, reqLogger, err)
	}

	// Check if driver is in running state (only if the status was previously set to Succeeded or Running)
	if checkStateOnly {
		return handleSuccess(ctx, instance, driverConfig, r, reqLogger, newStatus, oldStatus)
	}
	// Remove the force update field if set
	// The assumption is that we will not have a spec with Running/Succeeded state
	// and the forceUpdate field set
	if forceUpdate {
		instance.GetDriver().ForceUpdate = false
		isUpdated = true
	}
	if changed {
		isUpdated = true
	}
	// Update the instance
	if isUpdated {
		updateInstanceError := updateInstance(ctx, instance, r, reqLogger, isUpdated)
		if updateInstanceError != nil {
			newStatus.LastUpdate.ErrorMessage = updateInstanceError.Error()
			return logBannerAndReturn(reconcile.Result{
				Requeue: true, RequeueAfter: retryInterval}, updateInstanceError, reqLogger)
		}
		// Also update the status as we calculate the hash every time
		newStatus.LastUpdate = setLastStatusUpdate(oldStatus, csiv1.Updating, "")
		updateStatusError := updateStatus(ctx, instance, r, reqLogger, newStatus, oldStatus)
		if updateStatusError != nil {
			newStatus.LastUpdate.ErrorMessage = updateStatusError.Error()
			reqLogger.Info(fmt.Sprintf("\n################End Reconcile %s %s##############\n", driverType, request))
			return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, updateStatusError, reqLogger)
		}
	}
	// Validate Spec
	err = ValidateSpec(ctx, instance, r, driverConfig, reqLogger)
	if err != nil {
		return handleValidationError(ctx, instance, driverConfig, r, reqLogger, err)
	}
	// Validate any driver specific things
	err = r.ValidateDriverSpec(ctx, instance, reqLogger)
	if err != nil {
		return handleValidationError(ctx, instance, driverConfig, r, reqLogger, err)
	}
	// Set the driver status to updating
	newStatus.State = constants.Updating
	// Update the driver
	syncErr := SyncDriver(ctx, instance, r, driverConfig, reqLogger)
	if syncErr == nil {
		// Mark the driver state as succeeded
		newStatus.State = constants.Succeeded
		errorMsg := ""
		running, err := calculateState(ctx, instance, driverConfig, r, newStatus)
		if err != nil {
			errorMsg = err.Error()
		}
		if running {
			newStatus.State = constants.Running
		}
		newStatus.LastUpdate = setLastStatusUpdate(oldStatus,
			GetOperatorConditionTypeFromState(newStatus.State), errorMsg)
		updateStatusError := updateStatus(ctx, instance, r, reqLogger, newStatus, oldStatus)
		if updateStatusError != nil {
			return reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, updateStatusError
		}
		if newStatus.State != constants.Running {
			return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, nil, reqLogger)
		}
		return logBannerAndReturn(reconcile.Result{}, nil, reqLogger)
	}
	// Failed to sync driver deployment
	// Look at the last condition
	_, _ = calculateState(ctx, instance, driverConfig, r, newStatus)
	newStatus.LastUpdate = setLastStatusUpdate(oldStatus, csiv1.Error, syncErr.Error())
	// Check the last condition
	if oldStatus.LastUpdate.Condition == csiv1.Error {
		reqLogger.Info(" Driver previously encountered an error")
		timeSinceLastConditionChange := metav1.Now().Sub(oldStatus.LastUpdate.Time.Time).Round(time.Second)
		reqLogger.Info(fmt.Sprintf("Time since last condition change :%v", timeSinceLastConditionChange))
		if timeSinceLastConditionChange >= constants.MaxRetryDuration {
			// Mark the driver as failed and update the condition
			newStatus.State = constants.Failed
			newStatus.LastUpdate = setLastStatusUpdate(oldStatus,
				GetOperatorConditionTypeFromState(newStatus.State), syncErr.Error())
			// This will trigger a reconcile again
			_ = updateStatus(ctx, instance, r, reqLogger, newStatus, oldStatus)
			return logBannerAndReturn(reconcile.Result{Requeue: false}, nil, reqLogger)
		}
		retryInterval = time.Duration(math.Min(float64(timeSinceLastConditionChange.Nanoseconds()*2),
			float64(constants.MaxRetryInterval.Nanoseconds())))
	} else {
		_ = updateStatus(ctx, instance, r, reqLogger, newStatus, oldStatus)
	}
	reqLogger.Info(fmt.Sprintf("Retry Interval: %v", retryInterval))

	// Don't return an error here. Controller runtime will immediately requeue the request
	// Also the requeueAfter setting only is effective after an amount of time

	return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, nil, reqLogger)
}

// InitializeSpec - Initializes common and driver specific elements in spec
func InitializeSpec(instance csiv1.CSIDriver, r ReconcileCSI, driverConfig *ctrlconfig.Config, reqLogger logr.Logger) (bool, error) {
	isUpdated := false
	reqLogger.Info("Initializing the spec")
	isCommonSpecUpdated, err := InitializeCommonSpec(instance, driverConfig, reqLogger)
	isUpdated = isUpdated || isCommonSpecUpdated
	if err != nil {
		reqLogger.Error(err, "Failed to initialize common spec")
		return isUpdated, err
	}
	isDriverUpdated, err := r.InitializeDriverSpec(instance, reqLogger)
	isUpdated = isUpdated || isDriverUpdated
	if err != nil {
		reqLogger.Error(err, "Initializing error")
		return isUpdated, err
	}
	return isUpdated, nil
}

func updateInstance(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI, reqLogger logr.Logger, isUpdated bool) error {
	if isUpdated {
		reqLogger.Info("Attempting to update CR instance")
		err := r.GetClient().Update(ctx, instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update CR instance")
		} else {
			reqLogger.Info("Successfully updated CR instance")
		}
		return err
	}
	reqLogger.Info("No updates to instance at this point")
	return nil
}

// GetNodeEnv - Returns a list of environment variables for Node by merging
// the common environment variables, driver specific environment variables
// and environment variable specified via the Custom Resource spec
func GetNodeEnv(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) []corev1.EnvVar {
	envs := driverConfig.GetNodeEnvs()
	envs = mergeEnvironmentVars(envs, driver.GetDriver().Common.Envs)
	envs = mergeEnvironmentVars(envs, driver.GetDriver().Node.Envs)
	authSecretName := driver.GetDriver().AuthSecret
	if authSecretName != "" {
		for i, env := range envs {
			if env.Name == driver.GetUserEnvName() {
				envs[i].ValueFrom.SecretKeyRef.Name = authSecretName
			} else if env.Name == driver.GetPasswordEnvName() {
				envs[i].ValueFrom.SecretKeyRef.Name = authSecretName
			}
		}
	}
	envs = mergeEnvironmentVars(envs, GetCustomEnvVars(driver, driverConfig.ConfigVersion, envs))
	// Code only for PowerMax
	if driver.GetDriverType() == csiv1.PowerMax && driverConfig.ConfigVersion != "v1" {
		iscsiCHAPEnvName := "X_CSI_POWERMAX_ISCSI_ENABLE_CHAP"
		iscsiCHAPEnv, err := getEnvVar(iscsiCHAPEnvName, envs)
		if err == nil {
			if iscsiCHAPEnv.Value == "" || strings.Compare(strings.ToUpper(iscsiCHAPEnv.Value), "FALSE") == 0 {
				// Remove the ISCSI CHAP secret env
				envs = removeEnvVar(envs, "X_CSI_POWERMAX_ISCSI_CHAP_PASSWORD")
			}
		}
	}
	return envs
}

// GetControllerEnv - Returns a list of environment variables for Controller by merging
// the common environment variables, driver specific environment variables
// and environment variable specified via the Custom Resource spec
func GetControllerEnv(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) []corev1.EnvVar {
	envs := driverConfig.GetControllerEnvs()
	// First merge with the common environment variables
	envs = mergeEnvironmentVars(envs, driver.GetDriver().Common.Envs)
	// Merge with the Controller specific environment variables
	envs = mergeEnvironmentVars(envs, driver.GetDriver().Controller.Envs)
	authSecretName := driver.GetDriver().AuthSecret
	if authSecretName != "" {
		for i, env := range envs {
			if env.Name == driver.GetUserEnvName() {
				envs[i].ValueFrom.SecretKeyRef.Name = authSecretName
			} else if env.Name == driver.GetPasswordEnvName() {
				envs[i].ValueFrom.SecretKeyRef.Name = authSecretName
			}
		}
	}
	envs = mergeEnvironmentVars(envs, GetCustomEnvVars(driver, driverConfig.ConfigVersion, envs))
	return envs
}

// GetControllerArgs - Returns a list of arguments for Controller by merging
// the common arguments, driver specific arguments
// and arguments specified via the Custom Resource spec
func GetControllerArgs(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) []string {
	args := driverConfig.GetDriverArgs()
	// First merge with the common args
	args = mergeArgs(args, driver.GetDriver().Common.Args)
	// Merge with the controller specific args
	args = mergeArgs(args, driver.GetDriver().Controller.Args)
	return args
}

// GetNodeArgs - Returns a list of arguments for Node by merging
// the common arguments, driver specific arguments
// and arguments specified via the Custom Resource spec
func GetNodeArgs(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) []string {
	args := driverConfig.GetDriverArgs()
	args = mergeArgs(args, driver.GetDriver().Common.Args)
	args = mergeArgs(args, driver.GetDriver().Node.Args)
	return args
}

// GetControllerTolerations - Returns a list of tolerations for Controller
// by merging the common tolerations, the controller specific tolerations
// and the driver specific tolerations
func GetControllerTolerations(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) []corev1.Toleration {
	tolerations := driverConfig.GetDriverControllerTolerations()
	tolerations = mergeTolerations(tolerations, driver.GetDriver().Common.Tolerations)
	tolerations = mergeTolerations(tolerations, driver.GetDriver().Controller.Tolerations)
	return tolerations
}

// GetNodeTolerations - Returns a list of tolerations for Node
// by merging the common tolerations, the node specific tolerations
// and the driver specific tolerations
func GetNodeTolerations(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) []corev1.Toleration {
	tolerations := driverConfig.GetDriverNodeTolerations()
	tolerations = mergeTolerations(tolerations, driver.GetDriver().Common.Tolerations)
	tolerations = mergeTolerations(tolerations, driver.GetDriver().Node.Tolerations)
	return tolerations
}

// GetControllerNodeSelector - Returns NodeSelector for the statefulset pod
func GetControllerNodeSelector(driver csiv1.CSIDriver) map[string]string {
	if driver.GetDriver().Controller.NodeSelector != nil {
		return driver.GetDriver().Controller.NodeSelector
	}
	return driver.GetDriver().Common.NodeSelector
}

// GetNodeNodeSelector - Returns NodeSelector for the daemonset pod
func GetNodeNodeSelector(driver csiv1.CSIDriver) map[string]string {
	if driver.GetDriver().Node.NodeSelector != nil {
		return driver.GetDriver().Node.NodeSelector
	}
	return driver.GetDriver().Common.NodeSelector
}

// GetSideCarParams - Returns a map of parameters for each side car
// Parameters include arguments, environments and volume mounts
func GetSideCarParams(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config, reqLogger logr.Logger) map[csiv1.ImageType]ctrlconfig.SidecarParams {
	sideCarMap := make(map[csiv1.ImageType]ctrlconfig.SidecarParams)
	for _, sidecar := range driver.GetDriver().SideCars {
		// process args
		args := driverConfig.GetSideCarArgs(sidecar.Name)
		args = mergeArgs(args, sidecar.Args)

		// process environment variables
		envs := driverConfig.GetSideCarEnvs(sidecar.Name)
		envs = mergeEnvironmentVars(envs, sidecar.Envs)

		// note: volume mounts are not specified via CR spec
		vols := driverConfig.GetSideCarVolMounts(sidecar.Name)
		sideCarMap[sidecar.Name] = ctrlconfig.SidecarParams{Name: sidecar.Name, Args: args, Envs: envs, VolumeMounts: vols}
	}
	return sideCarMap
}

func updateAnnotationMap(annotations map[string]string, key string, value string) bool {
	if val, ok := annotations[key]; ok {
		if val != value {
			annotations[key] = value
			return true
		}
	} else {
		annotations[key] = value
		return true
	}
	return false
}

func updateAnnotations(annotations map[string]string, isDefault string, imageType, imageName string) bool {
	isUpdated := false
	key := fmt.Sprintf("%s/%s.Image.IsDefault", MetadataPrefix, imageType)
	annotationsUpdated := updateAnnotationMap(annotations, key, isDefault)
	if annotationsUpdated {
		isUpdated = true
	}
	key = fmt.Sprintf("%s/%s.Image", MetadataPrefix, imageType)
	annotationsUpdated = updateAnnotationMap(annotations, key, imageName)
	if annotationsUpdated {
		isUpdated = true
	}
	return isUpdated
}

// ApplyDefaultsForSideCars - Applies any missing defaults for side cars
func ApplyDefaultsForSideCars(instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config, annotations map[string]string,
	isUpgrade bool, reqLogger logr.Logger) ([]csiv1.ContainerTemplate, bool, error) {
	isUpdated := false
	driver := instance.GetDriver()

	sideCars := driver.SideCars
	sidecarNamesSpecifiedInSpec := make([]string, 0)
	for i, sideCar := range sideCars {
		sidecarNamesSpecifiedInSpec = append(sidecarNamesSpecifiedInSpec, string(sideCar.Name))
		if isStringInSlice(string(sideCar.Name), driverConfig.GetAllSideCars()) {
			defaultImage, err := driverConfig.GetDefaultImageTag(string(sideCar.Name))
			if err != nil {
				return []csiv1.ContainerTemplate{}, isUpdated, err
			}
			updated := false
			if sideCar.Image == "" {
				sideCars[i].Image = defaultImage
				isUpdated = true
				_ = updateAnnotations(annotations, "true", string(sideCar.Name), defaultImage)
			} else {
				if sideCar.Image != defaultImage {
					if isUpgrade {
						sideCars[i].Image = defaultImage
						isUpdated = true
						_ = updateAnnotations(annotations, "true", string(sideCar.Name), defaultImage)
					} else {
						// user specified image
						updated = updateAnnotations(annotations, "false", string(sideCar.Name), sideCar.Image)
					}
				} else {
					// Update the annotations just in case
					updated = updateAnnotations(annotations, "true", string(sideCar.Name), defaultImage)
				}
			}
			if updated {
				isUpdated = true
			}
			if sideCar.ImagePullPolicy == "" {
				sideCars[i].ImagePullPolicy = corev1.PullIfNotPresent
				isUpdated = true
			}
		}

	}
	// Add any missing mandatory sidecars
	for _, sideCarName := range driverConfig.GetMandatorySideCars() {
		if !isStringInSlice(sideCarName, sidecarNamesSpecifiedInSpec) {
			imageType, err := GetSideCarTypeFromName(sideCarName)
			if err != nil {
				reqLogger.Error(err, "Invalid image type found in default side cars")
				continue
			}
			defaultImage, err := driverConfig.GetDefaultImageTag(sideCarName)
			if err != nil {
				reqLogger.Error(err, "Failed to get default image tag. Continuing")
				continue
			}
			sideCar := csiv1.ContainerTemplate{
				Name:            imageType,
				Image:           defaultImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
			}
			sideCars = append(sideCars, sideCar)
			_ = updateAnnotations(annotations, "true", string(sideCar.Name), defaultImage)
			isUpdated = true
		}
	}
	if isUpdated {
		reqLogger.Info(fmt.Sprintf("Setting SideCars from default config: %v", sideCars))
	} else {
		reqLogger.Info("All default side car information is already present")
	}
	instance.SetAnnotations(annotations)
	return sideCars, isUpdated, nil
}

// ApplyDefaultsForInitContainers - Applies any missing defaults for InitContainers
func ApplyDefaultsForInitContainers(instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config,
	reqLogger logr.Logger) ([]csiv1.ContainerTemplate, bool, error) {
	isUpdated := false
	driver := instance.GetDriver()
	initContainers := driver.InitContainers
	initContainerNamesSpecifiedInSpec := make([]string, 0)
	for i, initcontainer := range initContainers {
		initContainerNamesSpecifiedInSpec = append(initContainerNamesSpecifiedInSpec, string(initcontainer.Name))
		// Try to fill in default values for all initcontainers
		if initcontainer.Image == "" {
			fmt.Printf("initcontainer.Name")
			image, err := driverConfig.GetDefaultImageTag(string(initcontainer.Name))
			if err != nil {
				return []csiv1.ContainerTemplate{}, isUpdated, err
			}
			initContainers[i].Image = image
			isUpdated = true
		}
		if initcontainer.ImagePullPolicy == "" {
			initContainers[i].ImagePullPolicy = corev1.PullIfNotPresent
			isUpdated = true
		}
	}
	if isUpdated {
		reqLogger.Info(fmt.Sprintf("Setting InitContainer from default config: %v", initContainers))
	} else {
		reqLogger.Info("All default InitContainer information is already present")
	}
	return initContainers, isUpdated, nil
}

// ApplyDefaultsForStorageClass - Applies defaults for storage classes
func ApplyDefaultsForStorageClass(instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config,
	isUpgrade bool, reqLogger logr.Logger) ([]csiv1.StorageClass, bool) {
	isUpdated := false
	driver := instance.GetDriver()
	scAttr := driverConfig.DriverConfig.StorageClassAttrs
	storageClasses := driver.StorageClass
	for i, sc := range storageClasses {
		for _, attr := range scAttr {
			switch attr.Name {
			case "allowVolumeExpansion":
				value := attr.Value.(bool)
				// Over ride the default value with user specified spec value
				if sc.AllowVolumeExpansion == nil {
					storageClasses[i].AllowVolumeExpansion = &value
					sc.AllowVolumeExpansion = &value
					isUpdated = true
				}
			case "volumeBindingMode":
				value := attr.Value.(string)
				if sc.VolumeBindingMode == "" && !isUpgrade {
					storageClasses[i].VolumeBindingMode = value
					sc.VolumeBindingMode = value
					isUpdated = true
				}
			}
		}
		if sc.VolumeBindingMode == "" {
			storageClasses[i].VolumeBindingMode = "Immediate"
			sc.VolumeBindingMode = "Immediate"
			isUpdated = true
		}
	}
	if isUpdated {
		reqLogger.Info(fmt.Sprintf("Setting StorageClass from default config: %v", driver.StorageClass))
	} else {
		reqLogger.Info("All default storageClass information is already present")
	}
	return storageClasses, isUpdated
}

// InitializeCommonSpec - Initializes status and applies defaults for sidecars
func InitializeCommonSpec(instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config, reqLogger logr.Logger) (bool, error) {
	status := instance.GetDriverStatus()
	isUpdated := false
	annotations := instance.GetAnnotations()
	annotationsUpdated := false
	if annotations == nil {
		annotations = make(map[string]string)
		annotationsUpdated = true
		instance.SetAnnotations(annotations)
	}
	driver := instance.GetDriver()
	isUpgrade := false
	// Check if it is an upgrade
	// Check for the annotations with the config version and if it matches with the current one
	if len(annotations) != 0 {
		if configVersionFromAnnotation, ok := annotations[configVersionKey]; ok {
			if configVersionFromAnnotation != "" && configVersionFromAnnotation != driver.ConfigVersion {
				// This means that it is an upgrade
				isUpgrade = true
			}
		}
	} else {
		// lack of annotation could mean that the driver was installed via the older operator
		// and we have not applied the annotations yet
		//TODO: The below code is a hack to check if we are upgrading from a driver installed via older operator
		if driverConfig.IsControllerHAEnabled("") {
			// Controller HA is enabled for all new drivers && all drivers installed via new operator will have the annotations set always
			// If annotations don't exist, then it means that this is an upgrade
			isUpgrade = true
		}
	}

	sideCars, sideCarsUpdated, err := ApplyDefaultsForSideCars(instance, driverConfig, annotations, isUpgrade, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Failure during applying defaults for sidecars")
		return false, err
	}
	initContainers, initContainersUpdated, err := ApplyDefaultsForInitContainers(instance, driverConfig, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Failure during applying defaults for initContainers")
		return false, err
	}

	driver.SideCars = sideCars
	driver.InitContainers = initContainers
	configVersionApplied, err := checkAndApplyConfigVersionAnnotations(instance, reqLogger, true)
	if err != nil {
		return false, err
	}
	isUpdated = annotationsUpdated || sideCarsUpdated || configVersionApplied || initContainersUpdated
	status.LastUpdate.ErrorMessage = ""
	return isUpdated, nil
}

// GetControllerInitContainersParams - Returns a map of parameters for each initcontainers
// Parameters include arguments, environments and volume mounts
func GetControllerInitContainersParams(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) map[csiv1.ImageType]ctrlconfig.InitContainerParams {
	initContainerMap := make(map[csiv1.ImageType]ctrlconfig.InitContainerParams)
	controllerInitContainers := driverConfig.GetControllerInitContainers()
	for _, initContainer := range driver.GetDriver().InitContainers {
		if isStringInSlice(string(initContainer.Name), controllerInitContainers) {
			// process args
			args := driverConfig.GetInitContainerArgs(initContainer.Name)
			args = mergeArgs(args, initContainer.Args)
			// process environment variables
			envs := driverConfig.GetInitContainerEnvs(initContainer.Name)
			envs = mergeEnvironmentVars(envs, initContainer.Envs)
			// process volume mounts
			vols := driverConfig.GetInitContainerVolMounts(initContainer.Name)
			initContainerMap[initContainer.Name] = ctrlconfig.InitContainerParams{
				Name:         initContainer.Name,
				Args:         args,
				Envs:         envs,
				VolumeMounts: vols,
			}
		}
	}
	return initContainerMap
}

// GetNodeInitContainersParams - Returns a map of parameters for the node init containers
func GetNodeInitContainersParams(driver csiv1.CSIDriver, driverConfig *ctrlconfig.Config) map[csiv1.ImageType]ctrlconfig.InitContainerParams {
	initContainerMap := make(map[csiv1.ImageType]ctrlconfig.InitContainerParams)
	nodeInitContainers := driverConfig.GetNodeInitContainers()
	for _, initContainer := range driver.GetDriver().InitContainers {
		// process args
		if isStringInSlice(string(initContainer.Name), nodeInitContainers) {
			args := driverConfig.GetInitContainerArgs(initContainer.Name)
			args = mergeArgs(args, initContainer.Args)
			// process environment variables
			envs := driverConfig.GetInitContainerEnvs(initContainer.Name)
			envs = mergeEnvironmentVars(envs, initContainer.Envs)
			// process volume mounts
			vols := driverConfig.GetInitContainerVolMounts(initContainer.Name)
			initContainerMap[initContainer.Name] = ctrlconfig.InitContainerParams{
				Name:         initContainer.Name,
				Args:         args,
				Envs:         envs,
				VolumeMounts: vols,
			}
		}
	}
	return initContainerMap
}

func getMultipleCertSecretVolume(ctx context.Context, instance csiv1.CSIDriver, client crclient.Client, reqLogger logr.Logger) corev1.Volume {
	secrets, err := secrets.GetSecrets(ctx, instance.GetNamespace(), client, reqLogger)
	//var CertSecretCount int
	var volume corev1.Volume
	if err != nil {
		return corev1.Volume{}
	}
	secretPrefix := fmt.Sprintf("%s-certs-", instance.GetDriverType())
	volSources := make([]corev1.VolumeProjection, 0)
	for i := 0; i >= 0; i++ {
		found := false
		for _, s := range secrets {
			if s.Name == fmt.Sprintf("%s%d", secretPrefix, i) {
				found = true
				certName := fmt.Sprintf("cert-%d", i)
				if _, ok := s.Data[certName]; ok {
					volSources = append(volSources, corev1.VolumeProjection{
						Secret: &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{Name: s.Name},
							Items: []corev1.KeyToPath{
								{Key: certName, Path: certName},
							},
						},
					})
				} else {
					reqLogger.Error(fmt.Errorf("cert Secret [%s] dosen't have key [%s] in the data filed", s.Name, certName), "")
				}
				break
			}
		}

		if !found {
			break
		}
	}

	projectedVolSource := corev1.ProjectedVolumeSource{
		Sources: volSources,
	}
	volume = corev1.Volume{
		Name: "certs",
		VolumeSource: corev1.VolumeSource{
			Projected: &projectedVolSource,
		},
	}

	return volume
}

func getDummyClusterRoleName(instance csiv1.CSIDriver) string {
	return fmt.Sprintf("%s-%s-dummy", instance.GetName(), instance.GetNamespace())
}

// SyncDriver - Sync the current installation - this can lead to a create or update
func SyncDriver(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI, driverConfig *ctrlconfig.Config,
	reqLogger logr.Logger) error {
	var err error
	client := r.GetClient()
	// First get the envs
	controllerEnvs := GetControllerEnv(instance, driverConfig)
	daemonSetEnvs := GetNodeEnv(instance, driverConfig)

	controllerPodConstraints := csiv1.PodSchedulingConstraints{}
	controllerPodConstraints.Tolerations = GetControllerTolerations(instance, driverConfig)
	controllerPodConstraints.NodeSelector = GetControllerNodeSelector(instance)

	nodePodConstraints := csiv1.PodSchedulingConstraints{}
	nodePodConstraints.Tolerations = GetNodeTolerations(instance, driverConfig)
	nodePodConstraints.NodeSelector = GetNodeNodeSelector(instance)

	// Get the name only using one set of env values
	customDriverName := GetCustomDriverName(instance, driverConfig.ConfigVersion, controllerEnvs, reqLogger)
	customRBACNames := false
	if customDriverName != "" {
		customRBACNames = true
	}
	// Create the dummy config map
	dummyClusterRole := rbac.NewDummyClusterRole(getDummyClusterRoleName(instance))
	dummyClusterRoleInstance, err := rbac.SyncClusterRole(context.TODO(), dummyClusterRole, client, reqLogger)
	if err != nil {
		return err
	}
	controllerClusterRole := rbac.NewControllerClusterRole(instance, customRBACNames,
		driverConfig.DriverConfig.ControllerHA, dummyClusterRoleInstance)
	_, err = rbac.SyncClusterRole(ctx, controllerClusterRole, client, reqLogger)
	if err != nil {
		return err
	}

	// Create controller ServiceAccount
	controllerSa := serviceaccount.New(instance, fmt.Sprintf("%s-controller", instance.GetDriverType()))
	err = serviceaccount.SyncServiceAccount(ctx, controllerSa, client, reqLogger)
	if err != nil {
		return err
	}

	controllerClusterRoleBinding := rbac.NewControllerClusterRoleBindings(instance, customRBACNames, dummyClusterRoleInstance)
	err = rbac.SyncClusterRoleBindings(ctx, controllerClusterRoleBinding, client, reqLogger)
	if err != nil {
		return err
	}
	isOpenshift := r.GetConfig().IsOpenShift
	isLimitedNodeRBAC := IsLimitedNodeRBAC(instance.GetDriverType(), instance.GetDriver().ConfigVersion)
	createServiceAccount := false
	if !isLimitedNodeRBAC {
		createServiceAccount = true
	} else if isOpenshift {
		createServiceAccount = true
	}
	if createServiceAccount {
		// Create Node ServiceAccount
		nodeSa := serviceaccount.New(instance, instance.GetDaemonSetName())
		err = serviceaccount.SyncServiceAccount(ctx, nodeSa, client, reqLogger)
		if err != nil {
			return err
		}
		if !isLimitedNodeRBAC {
			nodeClusterRole := rbac.NewNodeClusterRole(instance, customRBACNames, dummyClusterRoleInstance)
			_, err = rbac.SyncClusterRole(ctx, nodeClusterRole, client, reqLogger)
			if err != nil {
				return err
			}
		} else {
			limitedClusterRole := rbac.NewLimitedClusterRole(instance, customRBACNames, dummyClusterRoleInstance)
			_, err = rbac.SyncClusterRole(ctx, limitedClusterRole, client, reqLogger)
			if err != nil {
				return err
			}
		}
		nodeClusterRoleBinding := rbac.NewNodeClusterRoleBindings(instance, customRBACNames, dummyClusterRoleInstance)
		err = rbac.SyncClusterRoleBindings(ctx, nodeClusterRoleBinding, client, reqLogger)
		if err != nil {
			return err
		}
	}

	// Create CSI Driver entry
	csiDriver := csidriver.New(instance, driverConfig.DriverConfig.EnableEphemeralVolumes, dummyClusterRoleInstance)
	err = csidriver.SyncCSIDriver(ctx, csiDriver, client, reqLogger)
	if err != nil {
		return err
	}

	// Create StatefulSet
	secretVolumes := make([]corev1.Volume, 0)
	controllerVolumes := driverConfig.GetControllerVolumes()
	controllerVolumeMounts := driverConfig.GetControllerVolumeMounts()
	tlsCertSecretName := instance.GetDriver().TLSCertSecret
	if tlsCertSecretName != "" {
		reqLogger.Info(fmt.Sprintf("User specified TLS Cert Secret: %s", tlsCertSecretName))
		var mode int32
		mode = 420
		booleanTrue := true
		tlsSecretVolumeSource := corev1.SecretVolumeSource{
			SecretName:  tlsCertSecretName,
			Optional:    &booleanTrue,
			DefaultMode: &mode,
		}
		tlsVolume := corev1.Volume{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &tlsSecretVolumeSource,
			},
		}
		secretVolumes = append(secretVolumes, tlsVolume)
	}

	multipleCertSecretVolume := make([]corev1.Volume, 0)
	if instance.GetDriverType() == csiv1.Unity || instance.GetDriverType() == csiv1.Isilon {
		certVolume := getMultipleCertSecretVolume(ctx, instance, client, reqLogger)
		if certVolume.Name != "" {
			multipleCertSecretVolume = append(multipleCertSecretVolume, certVolume)
		}
	}
	if len(multipleCertSecretVolume) != 0 {
		controllerVolumes = mergeVolumes(controllerVolumes, multipleCertSecretVolume)
	}

	if len(secretVolumes) != 0 {
		controllerVolumes = mergeVolumes(controllerVolumes, secretVolumes)
	}
	args := GetControllerArgs(instance, driverConfig)
	sidecarMap := GetSideCarParams(instance, driverConfig, reqLogger)
	if driverConfig.DriverConfig.ControllerHA {
		deploy := deployment.New(instance, controllerEnvs, controllerVolumeMounts,
			controllerVolumes, args, sidecarMap, controllerPodConstraints)

		err = deployment.SyncControllerDeployment(ctx, deploy, client, reqLogger)
		if err != nil {
			return err
		}

		//Searching for statefulset, if found delete the statefulset
		statefulSet, err := statefulset.GetStatefulset(ctx, instance, client, reqLogger)
		if err == nil && statefulSet != nil {
			reqLogger.Info("Statefulset found", statefulSet.Namespace, statefulSet.Name)
			err = statefulset.DeleteStatefulset(ctx, statefulSet, client, reqLogger)
			if err != nil {
				reqLogger.Info("Delete statefulset failed", statefulSet.Namespace, statefulSet.Name)
			}
			reqLogger.Info("Statefulset deleted", statefulSet.Namespace, statefulSet.Name)
		}
	} else {
		ss := statefulset.New(instance, controllerEnvs, controllerVolumeMounts,
			controllerVolumes, args, sidecarMap, controllerPodConstraints)

		err = statefulset.SyncStatefulset(ctx, ss, client, reqLogger)
		if err != nil {
			return err
		}
	}

	// Create daemonset
	daemonSetVolumes := driverConfig.GetNodeVolumes()
	if len(multipleCertSecretVolume) != 0 {
		daemonSetVolumes = mergeVolumes(daemonSetVolumes, multipleCertSecretVolume)
	}
	daemonSetDriverVolumeMounts := driverConfig.GetNodeVolumeMounts()

	args = GetNodeArgs(instance, driverConfig)
	reqLogger.Info("calling GetInitContainerParams")
	nodeInitContainers := GetNodeInitContainersParams(instance, driverConfig)
	ds, err := daemonset.New(instance, daemonSetEnvs, daemonSetDriverVolumeMounts, daemonSetVolumes,
		args, nodeInitContainers, sidecarMap, createServiceAccount, nodePodConstraints, reqLogger)
	if err != nil {
		return err
	}
	err = daemonset.SyncDaemonset(ctx, ds, client, reqLogger)
	if err != nil {
		return err
	}
	return nil
}
