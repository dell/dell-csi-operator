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
	"fmt"
	"math"
	"net/url"
	"os"
	"reflect"
	"time"

	storagev1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/dell/dell-csi-operator/pkg/resources/configmap"
	"github.com/dell/dell-csi-operator/pkg/resources/deployment"
	"github.com/dell/dell-csi-operator/pkg/resources/rbac"
	"github.com/dell/dell-csi-operator/pkg/resources/service"
	"github.com/dell/dell-csi-operator/pkg/resources/serviceaccount"
	"github.com/dell/dell-csi-operator/pkg/utils"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Constants for the reverseproxy
const (
	ReverseProxyName         = "powermax-reverseproxy"
	ConfigMapName            = "powermax-reverseproxy-config"
	DefaultMode              = "Linked"
	DefaultPort              = int32(2222)
	ConfigFileName           = "config.yaml"
	ConfigMapVolumeName      = "configmap-volume"
	ConfigMapVolumeMountPath = "/etc/config/configmap"
	TLSSecretVolumeName      = "tls-secret"
	TLSSecretMountPath       = "/app/tls"
	CertVolumeName           = "cert-dir"
	CertVolumeMountPath      = "/app/certs"
)

var log = klogr.New().WithName("CSIPowerMaxReverseProxy")

// CSIPowerMaxRevProxyReconciler reconciles a CSIPowerMaxRevProxy object
type CSIPowerMaxRevProxyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=storage.dell.com,resources=csipowermaxrevproxies;csipowermaxrevproxies/finalizers;csipowermaxrevproxies/status,verbs=*

// Reconcile function reconciles a CSIPowerMax object
func (r *CSIPowerMaxRevProxyReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CSIPowerMaxRevProxy")
	retryInterval := constants.DefaultRetryInterval
	reqLogger.Info("################Starting Reconcile##############")
	// Fetch the CSIPowerMaxRevProxy instance
	instance := &storagev1.CSIPowerMaxRevProxy{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	status := instance.Status
	// newStatus is the status object which is modified and finally used to update the Status
	// in case the instance or the status is updated
	newStatus := status.DeepCopy()
	// oldStatus is the previous status of the CR instance
	// This is used to compare if there is a need to update the status
	oldStatus := status.DeepCopy()
	oldState := oldStatus.State
	reqLogger.Info(fmt.Sprintf("Proxy was previously in (%s) state", oldState))
	// Check if the proxy spec has changed
	expectedHash, actualHash, changed := utils.ProxyChanged(instance)
	if changed {
		message := fmt.Sprintf("Proxy spec has changed (%d vs %d)", actualHash, expectedHash)
		newStatus.ProxyHash = expectedHash
		reqLogger.Info(message)
	} else {
		reqLogger.Info("No changes detected in the proxy spec")
	}
	checkStateOnly := false
	switch oldState {
	case constants.Running:
		fallthrough
	case constants.Succeeded:
		if changed {
			// If the proxy hash has changed, we need to update the proxy again
			newStatus.State = constants.Updating
			reqLogger.Info("Changed state to Updating as proxy spec changed")
		} else {
			// Just check the state of the proxy and update status accordingly
			reqLogger.Info("Recalculating proxy state(only) as there is no change in proxy spec")
			checkStateOnly = true
		}
	case constants.InvalidConfig:
		fallthrough
	case constants.Failed:
		if changed {
			// Do a reconcile as we detected a change
			newStatus.State = constants.Updating
		} else {
			reqLogger.Info(fmt.Sprintf("CR is in (%s) state. Reconcile request won't be requeued",
				newStatus.State))
			return logBannerAndReturn(reconcile.Result{}, nil, reqLogger)
		}
	case constants.NoState:
		newStatus.State = constants.Updating
	case constants.Updating:
		reqLogger.Info("Proxy already in Updating state")
	}
	// Check if proxy is in running state (only if the status was previously set to Succeeded or Running)
	if checkStateOnly {
		return handleSuccess(context.TODO(), instance, r.Client, reqLogger, newStatus, oldStatus)
	}
	if changed {
		// Also update the status as we calculate the hash every time
		newStatus.LastUpdate = setLastStatusUpdate(oldStatus, storagev1.Updating, "")
		updateStatusError := updateStatus(context.TODO(), instance, r.Client, reqLogger, newStatus, oldStatus)
		if updateStatusError != nil {
			newStatus.LastUpdate.ErrorMessage = updateStatusError.Error()
			reqLogger.Info(fmt.Sprintf("\n################End Reconcile %s %s##############\n",
				"CSIPowerMaxReverseProxy", request))
			return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, updateStatusError, reqLogger)
		}
	}
	// Always validate the spec
	err = ValidateProxySpec(context.TODO(), r.Client, instance)
	if err != nil {
		return handleValidationError(context.TODO(), instance, r.Client, reqLogger, err)
	}
	// Set the proxy status to updating
	newStatus.State = constants.Updating
	syncErr := SyncProxy(instance, r.Client, reqLogger)
	if syncErr == nil {
		// Mark the proxy state as succeeded
		newStatus.State = constants.Succeeded
		errorMsg := ""
		running, err := utils.CalculateProxyState(context.TODO(), ReverseProxyName, instance.Namespace, r.Client, newStatus)
		if err != nil {
			errorMsg = err.Error()
		}
		if running {
			newStatus.State = constants.Running
		}
		newStatus.LastUpdate = setLastStatusUpdate(oldStatus,
			utils.GetOperatorConditionTypeFromState(newStatus.State), errorMsg)
		updateStatusError := updateStatus(context.TODO(), instance, r.Client, reqLogger, newStatus, oldStatus)
		if updateStatusError != nil {
			return reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, updateStatusError
		}
		if newStatus.State != constants.Running {
			return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, nil, reqLogger)
		}
		return logBannerAndReturn(reconcile.Result{}, nil, reqLogger)
	}
	// Failed to sync proxy deployment
	// Look at the last condition
	_, _ = utils.CalculateProxyState(context.TODO(), ReverseProxyName, instance.Namespace, r.Client, newStatus)
	newStatus.LastUpdate = setLastStatusUpdate(oldStatus, storagev1.Error, syncErr.Error())
	// Check the last condition
	if oldStatus.LastUpdate.Condition == storagev1.Error {
		reqLogger.Info(" Proxy previously encountered an error")
		timeSinceLastConditionChange := metav1.Now().Sub(oldStatus.LastUpdate.Time.Time).Round(time.Second)
		reqLogger.Info(fmt.Sprintf("Time since last condition change :%v", timeSinceLastConditionChange))
		if timeSinceLastConditionChange >= constants.MaxRetryDuration {
			// Mark the proxy as failed and update the condition
			newStatus.State = constants.Failed
			newStatus.LastUpdate = setLastStatusUpdate(oldStatus,
				utils.GetOperatorConditionTypeFromState(newStatus.State), syncErr.Error())
			// This will trigger a reconcile again
			_ = updateStatus(context.TODO(), instance, r.Client, reqLogger, newStatus, oldStatus)
			return logBannerAndReturn(reconcile.Result{Requeue: false}, nil, reqLogger)
		}
		retryInterval = time.Duration(math.Min(float64(timeSinceLastConditionChange.Nanoseconds()*2),
			float64(constants.MaxRetryInterval.Nanoseconds())))
	} else {
		_ = updateStatus(context.TODO(), instance, r.Client, reqLogger, newStatus, oldStatus)
	}
	reqLogger.Info(fmt.Sprintf("Retry Interval: %v", retryInterval))

	// Don't return an error here. Controller runtime will immediately requeue the request
	// Also the requeueAfter setting only is effective after an amount of time
	return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, nil, reqLogger)
}

// SetupWithManager - sets up controller
func (r *CSIPowerMaxRevProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("CSIPowerMaxRevProxy", mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.Log.Error(err, "Unable to setup CSIPowerMaxRevProxy controller")
		os.Exit(1)
	}

	err = c.Watch(
		&source.Kind{Type: &storagev1.CSIPowerMaxRevProxy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		r.Log.Error(err, "Unable to watch CSIPowerMaxRevProxy Driver")
		os.Exit(1)
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIPowerMaxRevProxy{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Deployment")
		os.Exit(1)
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIPowerMaxRevProxy{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Daemonset")
		os.Exit(1)
	}
	return nil
}

func setStatus(instance *storagev1.CSIPowerMaxRevProxy, newStatus *storagev1.CSIPowerMaxRevProxyStatus) {
	instance.Status.State = newStatus.State
	instance.Status.LastUpdate.ErrorMessage = newStatus.LastUpdate.ErrorMessage
	instance.Status.LastUpdate.Condition = newStatus.LastUpdate.Condition
	instance.Status.LastUpdate.Time = newStatus.LastUpdate.Time
	instance.Status.ProxyStatus = newStatus.ProxyStatus
	instance.Status.ProxyHash = newStatus.ProxyHash
}

// ValidateProxySpec - Validates the proxy specification
func ValidateProxySpec(ctx context.Context, client client.Client, instance *storagev1.CSIPowerMaxRevProxy) error {
	proxySpec := instance.Spec
	err := checkIfSecretExists(ctx, client, proxySpec.TLSSecret, instance.Namespace)
	if err != nil {
		return err
	}
	// Validate the mode
	switch proxySpec.RevProxy.Mode {
	case "":
		fallthrough
	case "Linked":
		return validateLinkedProxySpec(ctx, client, instance)
	case "StandAlone":
		return validateStandAloneProxySpec(ctx, client, instance)
	default:
		return fmt.Errorf("unknown mode specified")
	}
}

func validateLinkedProxySpec(ctx context.Context, client client.Client, instance *storagev1.CSIPowerMaxRevProxy) error {
	linkConfig := instance.Spec.RevProxy.LinkConfig
	if linkConfig == nil {
		return fmt.Errorf("link config can't be nil")
	}
	// Primary
	_, err := url.Parse(linkConfig.Primary.URL)
	if err != nil {
		return fmt.Errorf("linkConfig primary URL is not of the proper format. Error: %s", err.Error())
	}
	if linkConfig.Primary.SkipCertificateValidation == false {
		if linkConfig.Primary.CertSecret == "" {
			return fmt.Errorf("link config Primary: SkipCertificateValidation is set to false and cert secret has not been specified")
		}
		err = checkIfSecretExists(ctx, client, linkConfig.Primary.CertSecret, instance.Namespace)
		if err != nil {
			return err
		}
	}
	// Backup
	if linkConfig.Backup.URL != "" {
		_, err = url.Parse(linkConfig.Backup.URL)
		if err != nil {
			return fmt.Errorf("linkConfig backup URL is not of the proper format. Error: %s", err.Error())
		}
		if linkConfig.Backup.SkipCertificateValidation == false {
			if linkConfig.Backup.CertSecret == "" {
				return fmt.Errorf("link config Backup: SkipCertificateValidation is set to false and cert secret has not been specified")
			}
			err = checkIfSecretExists(ctx, client, linkConfig.Backup.CertSecret, instance.Namespace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateStandAloneProxySpec(ctx context.Context, client client.Client, instance *storagev1.CSIPowerMaxRevProxy) error {
	standAloneConfig := instance.Spec.RevProxy.StandAloneConfig
	if standAloneConfig == nil {
		return fmt.Errorf("stand-alone config can't be nil")
	}

	if len(standAloneConfig.ManagementServerConfig) == 0 {
		return fmt.Errorf("no management server(s) specified")
	}

	if len(standAloneConfig.StorageArrayConfig) == 0 {
		return fmt.Errorf("no storage array config(s) specified")
	}

	for _, managementServer := range standAloneConfig.ManagementServerConfig {
		// Check URL
		_, err := url.Parse(managementServer.URL)
		if err != nil {
			return fmt.Errorf("one of the management server's URL is not in proper format. Error: %s", err.Error())
		}

		// Check cert secret
		if managementServer.SkipCertificateValidation == false {
			if managementServer.CertSecret == "" {
				return fmt.Errorf("one of the management server's SkipCertificateValidation is set to false and cert secret has not been specified")
			}
			err = checkIfSecretExists(ctx, client, managementServer.CertSecret, instance.Namespace)
			if err != nil {
				return err
			}
		}

		// Check array credential secret
		if managementServer.ArrayCredentialSecret != "" {
			err = checkIfSecretExists(ctx, client, managementServer.ArrayCredentialSecret, instance.Namespace)
			if err != nil {
				return err
			}
		}
	}

	for _, arrayConfig := range standAloneConfig.StorageArrayConfig {
		// Check array id
		if arrayConfig.StorageArrayID == "" {
			return fmt.Errorf("array-id empty for one of the array configs")
		}

		// Check primary and backup URLs
		if arrayConfig.PrimaryURL == "" {
			return fmt.Errorf("invalid primary URL for one of the array configs")
		}
		_, err := url.Parse(arrayConfig.PrimaryURL)
		if err != nil {
			return fmt.Errorf("invalid primary URL for one of the array configs. Error: %s", err.Error())
		}

		if arrayConfig.BackupURL != "" {
			_, err := url.Parse(arrayConfig.BackupURL)
			if err != nil {
				return fmt.Errorf("invalid backup URL for one of the array configs. Error: %s", err.Error())
			}
		}

		// Check proxy credentials
		if len(arrayConfig.ProxyCredentialSecrets) == 0 {
			return fmt.Errorf("no proxy credential(s) speficied for authentication")
		}
		for _, credSecret := range arrayConfig.ProxyCredentialSecrets {
			err = checkIfSecretExists(ctx, client, credSecret, instance.Namespace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func boolPtr(i bool) *bool { return &i }

func handleValidationError(ctx context.Context, instance *storagev1.CSIPowerMaxRevProxy, client client.Client, reqLogger logr.Logger,
	validationError error) (reconcile.Result, error) {
	reqLogger.Error(validationError, "Validation error")
	status := instance.Status
	oldStatus := status.DeepCopy()
	newStatus := status.DeepCopy()
	// Update the status
	reqLogger.Info("Marking the proxy status as InvalidConfig")
	_, _ = utils.CalculateProxyState(ctx, ReverseProxyName, instance.Namespace, client, newStatus)
	newStatus.LastUpdate = setLastStatusUpdate(oldStatus, storagev1.InvalidConfig, validationError.Error())
	newStatus.State = constants.InvalidConfig
	_ = updateStatus(ctx, instance, client, reqLogger, newStatus, oldStatus)
	reqLogger.Error(validationError, "*************Create/Update failed ********")
	return logBannerAndReturn(reconcile.Result{Requeue: false}, nil, reqLogger)
}

func setLastStatusUpdate(status *storagev1.CSIPowerMaxRevProxyStatus, conditionType storagev1.CSIOperatorConditionType, errorMsg string) storagev1.LastUpdate {
	// If the condition has not changed, then don't update the time
	if status.LastUpdate.Condition == conditionType && status.LastUpdate.ErrorMessage == errorMsg {
		return storagev1.LastUpdate{
			Condition:    conditionType,
			ErrorMessage: errorMsg,
			Time:         status.LastUpdate.Time,
		}
	}
	return storagev1.LastUpdate{
		Condition:    conditionType,
		ErrorMessage: errorMsg,
		Time:         metav1.Now(),
	}
}

func updateStatus(ctx context.Context, instance *storagev1.CSIPowerMaxRevProxy, client client.Client, reqLogger logr.Logger,
	newStatus, oldStatus *storagev1.CSIPowerMaxRevProxyStatus) error {
	//running := calculateState(ctx, instance, r, newStatus)
	if !reflect.DeepEqual(oldStatus, newStatus) {
		statusString := fmt.Sprintf("Status: (State - %s, Error Message - %s, Proxy Hash - %d)",
			newStatus.State, newStatus.LastUpdate.ErrorMessage, newStatus.ProxyHash)
		reqLogger.Info(statusString)
		reqLogger.Info("State", "Proxy Status", newStatus.ProxyStatus)
		setStatus(instance, newStatus)
		reqLogger.Info("Attempting to update CR status")
		err := client.Status().Update(ctx, instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update CR status")
			return err
		}
		reqLogger.Info("Successfully updated CR status")
	} else {
		reqLogger.Info("No change to status. No updates will be applied to CR status")
	}
	return nil
}

func handleSuccess(ctx context.Context, instance *storagev1.CSIPowerMaxRevProxy, client client.Client, reqLogger logr.Logger, newStatus, oldStatus *storagev1.CSIPowerMaxRevProxyStatus) (reconcile.Result, error) {
	errorMsg := ""
	running, err := utils.CalculateProxyState(ctx, ReverseProxyName, instance.Namespace, client, newStatus)
	if err != nil {
		errorMsg = err.Error()
	}
	if running {
		newStatus.State = constants.Running
	} else if err != nil {
		newStatus.State = constants.Updating
	} else {
		newStatus.State = constants.Succeeded
	}
	newStatus.LastUpdate = setLastStatusUpdate(oldStatus,
		utils.GetOperatorConditionTypeFromState(newStatus.State), errorMsg)
	retryInterval := constants.DefaultRetryInterval
	requeue := true
	if newStatus.State == constants.Running {
		// If previously we were in running state
		if oldStatus.State == constants.Running {
			requeue = false
			reqLogger.Info("Proxy state didn't change from Running")
		}
	} else if newStatus.State == constants.Succeeded {
		if oldStatus.State == constants.Running {
			// We went back to succeeded from running
			reqLogger.Info("Proxy migrated from Running state to Succeeded state")
		} else if oldStatus.State == constants.Succeeded {
			timeSinceLastConditionChange := metav1.Now().Sub(oldStatus.LastUpdate.Time.Time).Round(time.Millisecond)
			reqLogger.Info(fmt.Sprintf("Time since last condition change: %v", timeSinceLastConditionChange))
			if timeSinceLastConditionChange >= constants.MaxRetryDuration {
				// Don't requeue again
				requeue = false
				reqLogger.Info("Time elapsed since last condition change is more than max limit. Not going to reconcile")
			} else {
				// set to the default retry interval at minimum
				retryInterval = time.Duration(math.Max(float64(timeSinceLastConditionChange.Nanoseconds()*2),
					float64(constants.DefaultRetryInterval)))
				// Maximum set to MaxRetryInterval
				retryInterval = time.Duration(math.Min(float64(retryInterval), float64(constants.MaxRetryInterval)))
			}
		}
	} else {
		requeue = true
	}
	updateStatusError := updateStatus(ctx, instance, client, reqLogger, newStatus, oldStatus)
	if updateStatusError != nil {
		reqLogger.Error(updateStatusError, "failed to update the status")
		// Don't return error as controller runtime will immediately requeue the request
		return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, nil, reqLogger)
	}
	if requeue {
		reqLogger.Info(fmt.Sprintf("Requeue interval: %v", retryInterval))
		return logBannerAndReturn(reconcile.Result{Requeue: true, RequeueAfter: retryInterval}, nil, reqLogger)
	}
	return logBannerAndReturn(reconcile.Result{}, nil, reqLogger)
}

func checkIfSecretExists(ctx context.Context, client client.Client, secretName, namespace string) error {
	found := &v1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return fmt.Errorf("failed to find secret: [%s]", secretName)
	} else if err != nil {
		log.Error(err, "Failed to query for secret. Warning - the proxy pod may not start")
	}
	return nil
}

func logBannerAndReturn(result reconcile.Result, err error, reqLogger logr.Logger) (reconcile.Result, error) {
	reqLogger.Info("################End Reconcile##############")
	return result, err
}

// Creates a new Service object for the custom resource
func newServiceForCR(cr *storagev1.CSIPowerMaxRevProxy) *v1.Service {
	labels := map[string]string{
		"name": ReverseProxyName,
	}
	port := cr.Spec.RevProxy.Port
	if port == 0 {
		port = DefaultPort
	}
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ReverseProxyName,
			Namespace:       cr.Namespace,
			Labels:          labels,
			OwnerReferences: getOwnerReferences(cr),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:     port,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: DefaultPort,
					},
				},
			},
			Type:     v1.ServiceTypeClusterIP,
			Selector: labels,
		},
	}
}

// newDeploymentForCR - Creates a new deployment object for the Custom Resource
func newDeploymentForCR(cr *storagev1.CSIPowerMaxRevProxy) *appsv1.Deployment {
	labels := map[string]string{
		"name": ReverseProxyName,
	}

	var imagePullPolicy v1.PullPolicy
	if cr.Spec.ImagePullPolicy == "" {
		imagePullPolicy = v1.PullIfNotPresent
	} else {
		imagePullPolicy = cr.Spec.ImagePullPolicy
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ReverseProxyName,
			Namespace:       cr.Namespace,
			OwnerReferences: getOwnerReferences(cr),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: ReverseProxyName,
					Containers: []v1.Container{
						{
							Name:            "csireverseproxy",
							Image:           cr.Spec.Image,
							Env:             proxyEnvs(cr.Namespace),
							VolumeMounts:    volumeMounts(),
							ImagePullPolicy: imagePullPolicy,
						},
					},
					Volumes: volumes(cr),
				},
			},
		},
	}
}

// newRoleBindingForCR - Creates a Role Binding object for the Custom Resource
func newRoleBindingForCR(cr *storagev1.CSIPowerMaxRevProxy) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ReverseProxyName,
			Namespace:       cr.Namespace,
			OwnerReferences: getOwnerReferences(cr),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      ReverseProxyName,
			Namespace: cr.Namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     ReverseProxyName,
		},
	}
}

// newRoleForCR - Creates a new role object for the Custom Resource
func newRoleForCR(cr *storagev1.CSIPowerMaxRevProxy) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ReverseProxyName,
			Namespace:       cr.Namespace,
			OwnerReferences: getOwnerReferences(cr),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"list", "watch", "get"},
			},
		},
	}
}

// newServiceAccount - Creates a new service account for the Custom Resource
func newServiceAccount(cr *storagev1.CSIPowerMaxRevProxy) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ReverseProxyName,
			Namespace:       cr.Namespace,
			OwnerReferences: getOwnerReferences(cr),
		},
	}
}

// volumeMounts - Returns a slice of volume mounts for the reverseproxy container
func volumeMounts() []v1.VolumeMount {
	volumeMounts := make([]v1.VolumeMount, 0)
	configMapMount := v1.VolumeMount{
		Name:      ConfigMapVolumeName,
		MountPath: ConfigMapVolumeMountPath,
	}
	secretMount := v1.VolumeMount{
		Name:      TLSSecretVolumeName,
		MountPath: TLSSecretMountPath,
	}
	emptyDirMount := v1.VolumeMount{
		Name:      CertVolumeName,
		MountPath: CertVolumeMountPath,
	}
	volumeMounts = append(volumeMounts, configMapMount)
	volumeMounts = append(volumeMounts, secretMount)
	volumeMounts = append(volumeMounts, emptyDirMount)
	return volumeMounts
}

// volumes - Returns a slice of volumes for the reverseproxy pod
func volumes(cr *storagev1.CSIPowerMaxRevProxy) []v1.Volume {
	volumes := make([]v1.Volume, 0)
	configMapVol := v1.Volume{
		Name: ConfigMapVolumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: ConfigMapName,
				},
				Optional: boolPtr(false),
			},
		},
	}
	secretVol := v1.Volume{
		Name: TLSSecretVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: cr.Spec.TLSSecret,
				Optional:   boolPtr(false),
			},
		},
	}
	emptyDirVol := v1.Volume{
		Name: CertVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: "",
			},
		},
	}
	volumes = append(volumes, configMapVol)
	volumes = append(volumes, secretVol)
	volumes = append(volumes, emptyDirVol)
	return volumes
}

// Marshals the proxy configuration from the Custom Resource
// and returns a configmap object
func newConfigMapForCR(cr *storagev1.CSIPowerMaxRevProxy) (*v1.ConfigMap, error) {
	config := cr.Spec.RevProxy
	if config.Mode == "" {
		config.Mode = DefaultMode
	}
	if config.Port == 0 {
		config.Port = DefaultPort
	}
	out, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}
	configMapData := make(map[string]string)
	configMapData[ConfigFileName] = string(out)
	labels := map[string]string{
		"name": ReverseProxyName,
	}
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ConfigMapName,
			Namespace:       cr.Namespace,
			Labels:          labels,
			OwnerReferences: getOwnerReferences(cr),
		},
		Data: configMapData,
	}, nil
}

func proxyEnvs(namespace string) []v1.EnvVar {
	envVars := make([]v1.EnvVar, 0)
	envVars = append(envVars, v1.EnvVar{Name: "X_CSI_REVPROXY_CONFIG_DIR", Value: ConfigMapVolumeMountPath})
	envVars = append(envVars, v1.EnvVar{Name: "X_CSI_REVPROXY_CONFIG_FILE_NAME", Value: ConfigFileName})
	envVars = append(envVars, v1.EnvVar{Name: "X_CSI_REVRPOXY_IN_CLUSTER", Value: "true"})
	envVars = append(envVars, v1.EnvVar{Name: "X_CSI_REVPROXY_TLS_CERT_DIR", Value: TLSSecretMountPath})
	envVars = append(envVars, v1.EnvVar{Name: "X_CSI_REVPROXY_WATCH_NAMESPACE", Value: namespace})
	return envVars
}

// getOwnerReference - returns owner references for k8s objects
func getOwnerReferences(cr *storagev1.CSIPowerMaxRevProxy) []metav1.OwnerReference {
	meta := &cr.TypeMeta
	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(cr, schema.GroupVersionKind{
			Group:   storagev1.GroupVersion.Group,
			Version: storagev1.GroupVersion.Version,
			Kind:    meta.Kind,
		}),
	}
	return ownerReferences
}

// SyncProxy - syncs the proxy instance
func SyncProxy(cr *storagev1.CSIPowerMaxRevProxy, client client.Client, reqLogger logr.Logger) error {
	// Create the configmap
	configMap, err := newConfigMapForCR(cr)
	if err != nil {
		return err
	}
	err = configmap.SyncConfigMap(context.TODO(), configMap, client, reqLogger)
	if err != nil {
		return err
	}
	// Create service object
	proxyService := newServiceForCR(cr)
	err = service.SyncService(context.TODO(), proxyService, client, reqLogger)
	if err != nil {
		return err
	}
	sa := newServiceAccount(cr)
	err = serviceaccount.SyncServiceAccount(context.TODO(), sa, client, reqLogger)
	if err != nil {
		return err
	}
	role := newRoleForCR(cr)
	err = rbac.SyncRole(context.TODO(), role, client, reqLogger)
	if err != nil {
		return err
	}
	roleBinding := newRoleBindingForCR(cr)
	err = rbac.SyncRoleBindings(context.TODO(), roleBinding, client, reqLogger)
	if err != nil {
		return err
	}
	proxyDeployment := newDeploymentForCR(cr)
	err = deployment.SyncDeployment(context.TODO(), proxyDeployment, client, reqLogger)
	if err != nil {
		return err
	}
	return nil
}

// SetClient sets the client for CSIPowerMaxRevProxyReconciler
func (r *CSIPowerMaxRevProxyReconciler) SetClient(client client.Client) *CSIPowerMaxRevProxyReconciler {
	r.Client = client
	return r
}

// SetScheme sets the scheme for ReconcileCSIPowerMaxRevProxy
func (r *CSIPowerMaxRevProxyReconciler) SetScheme(scheme *runtime.Scheme) *CSIPowerMaxRevProxyReconciler {
	r.Scheme = scheme
	return r
}
