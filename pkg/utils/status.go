package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dell/dell-csi-operator/pkg/ctrlconfig"
	"hash/fnv"
	"math"
	"reflect"
	"time"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func getInt32(pointer *int32) int32 {
	if pointer == nil {
		return 0
	}
	return *pointer
}

func getControllerStatus(ctx context.Context, instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config, r ReconcileCSI) (int32, csiv1.PodStatus, error) {
	var available, ready, starting, stopped []string
	var controllerReplicas, readyCount int32

	// TODO: This is a hack and should be removed after we remove all statefulset related code
	// We make an assumption (fair) that if DriverConfig is nil then ControllerHA is enabled
	if driverConfig.DriverConfig == nil || driverConfig.DriverConfig.ControllerHA {
		controller := &appsv1.Deployment{}
		err := r.GetClient().Get(ctx, types.NamespacedName{Name: instance.GetControllerName(),
			Namespace: instance.GetNamespace()}, controller)
		if err != nil {
			return 0, csiv1.PodStatus{}, err
		}
		controllerReplicas = getInt32(controller.Spec.Replicas)
		readyCount = controller.Status.ReadyReplicas
	} else {
		controller := &appsv1.StatefulSet{}
		err := r.GetClient().Get(ctx, types.NamespacedName{Name: instance.GetControllerName(),
			Namespace: instance.GetNamespace()}, controller)
		if err != nil {
			return 0, csiv1.PodStatus{}, err
		}
		controllerReplicas = getInt32(controller.Spec.Replicas)
		readyCount = controller.Status.ReadyReplicas
	}

	if controllerReplicas == 0 || readyCount == 0 {
		stopped = append(stopped, instance.GetControllerName())
	} else {
		podList := &v1.PodList{}
		opts := []client.ListOption{
			client.InNamespace(instance.GetNamespace()),
			client.MatchingLabels{"app": instance.GetControllerName()},
		}
		err := r.GetClient().List(ctx, podList, opts...)
		if err != nil {
			return controllerReplicas, csiv1.PodStatus{}, err
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				running := true
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.State.Running == nil {
						running = false
						break
					}
				}
				if running {
					available = append(available, pod.Name)
				} else {
					ready = append(ready, pod.Name)
				}
			} else if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodUnknown || pod.Status.Phase == corev1.PodRunning {
				starting = append(starting, pod.Name)
			} else if pod.Status.Phase == corev1.PodFailed {
				stopped = append(stopped, pod.Name)
			}
		}
	}
	return controllerReplicas, csiv1.PodStatus{
		Available: available,
		Stopped:   stopped,
		Starting:  starting,
		Ready:     ready,
	}, nil
}

func getDeploymentStatus(ctx context.Context, deploymentName, namespace string, crcClient client.Client) (int32, csiv1.PodStatus, error) {
	var available, ready, starting, stopped []string
	deployment := &appsv1.Deployment{}
	err := crcClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespace}, deployment)
	if err != nil {
		return 0, csiv1.PodStatus{}, err
	}
	replicas := getInt32(deployment.Spec.Replicas)
	readyCount := deployment.Status.ReadyReplicas
	if replicas == 0 || readyCount == 0 {
		stopped = append(stopped, deploymentName)
	} else {
		podList := &v1.PodList{}
		opts := []client.ListOption{
			client.InNamespace(namespace),
			client.MatchingLabels{"name": deploymentName},
		}
		err = crcClient.List(ctx, podList, opts...)
		if err != nil {
			return replicas, csiv1.PodStatus{}, err
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				running := true
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.State.Running == nil {
						running = false
						break
					}
				}
				if running {
					available = append(available, pod.Name)
				} else {
					ready = append(ready, pod.Name)
				}
			} else if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodUnknown || pod.Status.Phase == corev1.PodRunning {
				starting = append(starting, pod.Name)
			} else if pod.Status.Phase == corev1.PodFailed {
				stopped = append(stopped, pod.Name)
			}
		}
	}
	return replicas, csiv1.PodStatus{
		Available: available,
		Stopped:   stopped,
		Starting:  starting,
		Ready:     ready,
	}, nil
}

func getDaemonSetStatus(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI) (int32, csiv1.PodStatus, error) {
	var available, ready, starting, stopped []string
	node := &appsv1.DaemonSet{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: instance.GetDaemonSetName(),
		Namespace: instance.GetNamespace()}, node)
	if err != nil {
		return 0, csiv1.PodStatus{}, err
	}
	if node.Status.DesiredNumberScheduled == 0 || node.Status.NumberReady == 0 {
		stopped = append(stopped, instance.GetDaemonSetName())
	} else {
		podList := &v1.PodList{}
		opts := []client.ListOption{
			client.InNamespace(instance.GetNamespace()),
			client.MatchingLabels{"app": instance.GetDaemonSetName()},
		}
		err = r.GetClient().List(ctx, podList, opts...)
		if err != nil {
			return node.Status.DesiredNumberScheduled, csiv1.PodStatus{}, err
		}
		for _, pod := range podList.Items {
			if podutil.IsPodAvailable(&pod, node.Spec.MinReadySeconds, metav1.Now()) {
				available = append(available, pod.Name)
			} else if podutil.IsPodReady(&pod) {
				ready = append(ready, pod.Name)
			} else {
				starting = append(starting, pod.Name)
			}
		}
	}
	return node.Status.DesiredNumberScheduled, csiv1.PodStatus{
		Available: available,
		Stopped:   stopped,
		Starting:  starting,
		Ready:     ready,
	}, nil
}

func calculateState(ctx context.Context, instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config, r ReconcileCSI, newStatus *csiv1.DriverStatus) (bool, error) {
	running := false
	controllerReplicas, controllerStatus, statefulSetErr := getControllerStatus(ctx, instance, driverConfig, r)
	newStatus.ControllerStatus = controllerStatus
	expected, nodeStatus, daemonSetErr := getDaemonSetStatus(ctx, instance, r)
	newStatus.NodeStatus = nodeStatus
	if ((controllerReplicas != 0) && (controllerReplicas == int32(len(controllerStatus.Available)))) && ((expected != 0) && (expected == int32(len(nodeStatus.Available)))) {
		// Even if there is an error message, it is okay to overwrite that as all the pods are in running state
		running = true
	}
	var err error
	if statefulSetErr != nil {
		if daemonSetErr != nil {
			err = fmt.Errorf("statefulseterror: %s, daemonseterror: %s", statefulSetErr.Error(), daemonSetErr.Error())
		} else {
			err = statefulSetErr
		}
	} else {
		if daemonSetErr != nil {
			err = daemonSetErr
		} else {
			err = nil
		}
	}
	return running, err
}

// CalculateProxyState - Calculates the state of the Reverse Proxy CR
func CalculateProxyState(ctx context.Context, deploymentName, namespace string, client client.Client,
	newStatus *csiv1.CSIPowerMaxRevProxyStatus) (bool, error) {
	running := false
	replicas, deploymentStatus, err := getDeploymentStatus(ctx, deploymentName, namespace, client)
	newStatus.ProxyStatus = deploymentStatus
	if (replicas != 0) && (replicas == int32(len(deploymentStatus.Available))) {
		// Even if there is an error message, it is okay to overwrite that as all the pods are in running state
		running = true
	}
	return running, err
}

// HashDriver returns the hash of the driver specification
// This is used to detect if the driver spec has changed and any updates are required
func HashDriver(instance csiv1.CSIDriver) uint64 {
	hash := fnv.New32a()
	driverJSON, _ := json.Marshal(instance.GetDriver())
	hashutil.DeepHashObject(hash, driverJSON)
	return uint64(hash.Sum32())
}

// HashProxy returns the hash of the proxy specification
// This is used to detect if the proxy spec has changed and any updates are required
func HashProxy(instance *csiv1.CSIPowerMaxRevProxy) uint64 {
	hash := fnv.New32a()
	proxySpecJSON, _ := json.Marshal(instance.Spec)
	hashutil.DeepHashObject(hash, proxySpecJSON)
	return uint64(hash.Sum32())
}

func setStatus(instance csiv1.CSIDriver, newStatus *csiv1.DriverStatus) {
	instance.GetDriverStatus().State = newStatus.State
	instance.GetDriverStatus().LastUpdate.ErrorMessage = newStatus.LastUpdate.ErrorMessage
	instance.GetDriverStatus().LastUpdate.Condition = newStatus.LastUpdate.Condition
	instance.GetDriverStatus().LastUpdate.Time = newStatus.LastUpdate.Time
	instance.GetDriverStatus().ControllerStatus = newStatus.ControllerStatus
	instance.GetDriverStatus().NodeStatus = newStatus.NodeStatus
	instance.GetDriverStatus().DriverHash = newStatus.DriverHash
}

func setLastStatusUpdate(status *csiv1.DriverStatus, conditionType csiv1.CSIOperatorConditionType, errorMsg string) csiv1.LastUpdate {
	// If the condition has not changed, then don't update the time
	if status.LastUpdate.Condition == conditionType && status.LastUpdate.ErrorMessage == errorMsg {
		return csiv1.LastUpdate{
			Condition:    conditionType,
			ErrorMessage: errorMsg,
			Time:         status.LastUpdate.Time,
		}
	}
	return csiv1.LastUpdate{
		Condition:    conditionType,
		ErrorMessage: errorMsg,
		Time:         metav1.Now(),
	}
}

func updateStatus(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI, reqLogger logr.Logger, newStatus, oldStatus *csiv1.DriverStatus) error {
	//running := calculateState(ctx, instance, r, newStatus)
	if !reflect.DeepEqual(oldStatus, newStatus) {
		statusString := fmt.Sprintf("Status: (State - %s, Error Message - %s, Driver Hash - %d)",
			newStatus.State, newStatus.LastUpdate.ErrorMessage, newStatus.DriverHash)
		reqLogger.Info(statusString)
		reqLogger.Info("State", "Controller",
			newStatus.ControllerStatus, "Node", newStatus.NodeStatus)
		setStatus(instance, newStatus)
		reqLogger.Info("Attempting to update CR status")
		err := r.GetClient().Status().Update(ctx, instance)
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

func handleValidationError(ctx context.Context, instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config, r ReconcileCSI, reqLogger logr.Logger,
	validationError error) (reconcile.Result, error) {
	reqLogger.Error(validationError, "Validation error")
	status := instance.GetDriverStatus()
	oldStatus := status.DeepCopy()
	newStatus := status.DeepCopy()
	// Update the status
	reqLogger.Info("Marking the driver status as InvalidConfig")
	_, _ = calculateState(ctx, instance, driverConfig, r, newStatus)
	newStatus.LastUpdate = setLastStatusUpdate(oldStatus, csiv1.InvalidConfig, validationError.Error())
	newStatus.State = constants.InvalidConfig
	_ = updateStatus(ctx, instance, r, reqLogger, newStatus, oldStatus)
	reqLogger.Error(validationError, fmt.Sprintf("*************Create/Update %s failed ********",
		instance.GetDriverType()))
	return logBannerAndReturn(reconcile.Result{Requeue: false}, nil, reqLogger)
}

// GetOperatorConditionTypeFromState - Returns operator condition type
func GetOperatorConditionTypeFromState(state csiv1.DriverState) csiv1.CSIOperatorConditionType {
	switch state {
	case constants.Succeeded:
		return csiv1.Succeeded
	case constants.Running:
		return csiv1.Running
	case constants.InvalidConfig:
		return csiv1.InvalidConfig
	case constants.Updating:
		return csiv1.Updating
	case constants.Failed:
		return csiv1.Failed
	}
	return csiv1.Error
}

func handleSuccess(ctx context.Context, instance csiv1.CSIDriver, driverConfig *ctrlconfig.Config, r ReconcileCSI, reqLogger logr.Logger, newStatus, oldStatus *csiv1.DriverStatus) (reconcile.Result, error) {
	errorMsg := ""
	running, err := calculateState(ctx, instance, driverConfig, r, newStatus)
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
		GetOperatorConditionTypeFromState(newStatus.State), errorMsg)
	retryInterval := constants.DefaultRetryInterval
	requeue := true
	if newStatus.State == constants.Running {
		// If previously we were in running state
		if oldStatus.State == constants.Running {
			requeue = false
			reqLogger.Info("Driver state didn't change from Running")
		}
	} else if newStatus.State == constants.Succeeded {
		if oldStatus.State == constants.Running {
			// We went back to succeeded from running
			reqLogger.Info("Driver migrated from Running state to Succeeded state")
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
	updateStatusError := updateStatus(ctx, instance, r, reqLogger, newStatus, oldStatus)
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
