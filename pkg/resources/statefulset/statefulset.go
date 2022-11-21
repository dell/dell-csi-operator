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
package statefulset

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/dell/dell-csi-operator/pkg/ctrlconfig"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SleepTime - minimum time to sleep before checking the state of controller pod
var SleepTime = 10 * time.Second

// New - Creates a statefulset element for the given driver and component
func New(instance csiv1.CSIDriver, driverEnv []corev1.EnvVar, driverVolumeMounts []corev1.VolumeMount, podVolumes []corev1.Volume,
	args []string, sidecarMap map[csiv1.ImageType]ctrlconfig.SidecarParams, podConstraints csiv1.PodSchedulingConstraints) *appsv1.StatefulSet {
	var driver = instance.GetDriver()
	driverNamespace := instance.GetNamespace()
	replicas := driver.Replicas
	controllerName := instance.GetControllerName()
	labels := make(map[string]string)
	labels["app"] = controllerName
	containers := make([]corev1.Container, 0)

	containers = append(containers, resources.CreateContainerElement(
		csiv1.ImageTypeDriver, driver.Common.Image, driver.Common.ImagePullPolicy,
		args, driverEnv, driverVolumeMounts, nil, nil))
	for _, sideCarContainer := range driver.SideCars {
		// Add all sidecars except registrar for controller
		if sideCarContainer.Name != csiv1.ImageTypeRegistrar {
			containerName := sideCarContainer.Name
			imageName := sideCarContainer.Image
			containers = append(containers, resources.CreateContainerElement(
				containerName, imageName, sideCarContainer.ImagePullPolicy,
				sidecarMap[containerName].Args, sidecarMap[containerName].Envs,
				sidecarMap[containerName].VolumeMounts, nil, nil))
		}
	}
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			//Labels:          ,
			Name:            controllerName,
			Namespace:       driverNamespace,
			OwnerReferences: resources.GetOwnerReferences(instance),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &constants.RevisionHistoryLimit,
			PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
			ServiceName:          controllerName,

			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					DNSPolicy:                     corev1.DNSClusterFirst,
					ServiceAccountName:            controllerName,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 corev1.DefaultSchedulerName,
					TerminationGracePeriodSeconds: &constants.TerminationGracePeriodSeconds,
					Volumes:                       podVolumes,
					Tolerations:                   podConstraints.Tolerations,
					NodeSelector:                  podConstraints.NodeSelector,
				},
			},
		},
	}
}

// GetStatefulset -- Gets  a Statefulset
func GetStatefulset(ctx context.Context, instance csiv1.CSIDriver, client client.Client, reqLogger logr.Logger) (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	driverNamespace := instance.GetNamespace()
	controllerName := instance.GetControllerName()
	err := client.Get(ctx, types.NamespacedName{Name: controllerName, Namespace: driverNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return nil, err
	}

	return found, nil
}

// DeleteStatefulset -- Deletes a StatefulSet
func DeleteStatefulset(ctx context.Context, statefulset *appsv1.StatefulSet, client client.Client, reqLogger logr.Logger) error {
	err := client.Delete(ctx, statefulset)
	if err != nil && errors.IsNotFound(err) {
		return err
	}
	return nil
}

// SyncStatefulset - Syncs a StatefulSet
func SyncStatefulset(ctx context.Context, statefulset *appsv1.StatefulSet, client client.Client, reqLogger logr.Logger) error {
	found := &appsv1.StatefulSet{}
	err := client.Get(ctx, types.NamespacedName{Name: statefulset.Name, Namespace: statefulset.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Statefulset", "Namespace", statefulset.Namespace, "Name", statefulset.Name)
		err = client.Create(ctx, statefulset)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating StatefulSet", "Name:", statefulset.Name)
		err = client.Update(ctx, statefulset)
		if err != nil {
			return err
		}
		if statefulset.Status.ReadyReplicas != statefulset.Status.Replicas {
			// Check if the pod spec is same as pod spec from stateful spec
			reqLogger.Info("Waiting 10 seconds before checking the status of controller pods")
			time.Sleep(SleepTime)
		}
		err := client.Get(ctx, types.NamespacedName{Name: statefulset.Name, Namespace: statefulset.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Error(err, "Failed to find the statefulset after upgrade. Internal error!")
			return err
		}
		podTemplateSpec := found.Spec.Template.Spec
		for i := found.Status.Replicas - 1; i >= 0; i-- {
			controllerPod := &corev1.Pod{}
			controllerPodName := fmt.Sprintf("%s-%d", statefulset.Name, i)
			err = client.Get(ctx, types.NamespacedName{Name: controllerPodName, Namespace: statefulset.Namespace}, controllerPod)
			if err == nil {
				podSpec := controllerPod.Spec
				if !comparePodSpec(podTemplateSpec, podSpec, reqLogger) {
					reqLogger.Info("Deleting the controller pod", controllerPodName)
					err = client.Delete(ctx, controllerPod)
					if err != nil {
						reqLogger.Error(err, "Failed to delete the pod. Continuing")
					}
				}
			} else {
				reqLogger.Error(err, "Failed to get the controller pod. Continuing")
			}
		}
	}
	return nil
}

func comparePodSpec(spec1, spec2 corev1.PodSpec, reqLogger logr.Logger) bool {
	for _, container1 := range spec1.Containers {
		for _, container2 := range spec2.Containers {
			if container1.Name == container2.Name {
				if !reflect.DeepEqual(container1.Env, container2.Env) {
					reqLogger.Info("Environments don't match for", container1.Name)
					return false
				}
				reqLogger.Info(fmt.Sprintf("Environment variables match for %s", container1.Name))
				if container1.Image != container2.Image {
					reqLogger.Info(fmt.Sprintf("Image (%s, %s) don't match for container %s",
						string(container1.Image), string(container2.Image), container1.Name))
					return false
				}
				if container1.ImagePullPolicy != container2.ImagePullPolicy {
					reqLogger.Info(fmt.Sprintf("ImagePullPolicy (%s, %s) don't match for container %s",
						string(container1.ImagePullPolicy), string(container2.ImagePullPolicy), container1.Name))
					return false
				}
			}
			break
		}
	}
	return true
}
