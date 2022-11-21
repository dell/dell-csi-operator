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
package daemonset

import (
	"context"
	customError "errors"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/dell/dell-csi-operator/pkg/ctrlconfig"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New - Returns a daemonset object
func New(instance csiv1.CSIDriver,
	nodeEnv []corev1.EnvVar, driverVolumeMounts []corev1.VolumeMount, volumes []corev1.Volume, driverArgs []string, initContainerMap map[csiv1.ImageType]ctrlconfig.InitContainerParams,
	sidecarMap map[csiv1.ImageType]ctrlconfig.SidecarParams, rbacRequired bool, podConstraints csiv1.PodSchedulingConstraints, reqLogger logr.Logger) (*appsv1.DaemonSet, error) {
	var driver = instance.GetDriver()

	driverNamespace := instance.GetNamespace()
	daemonSetName := instance.GetDaemonSetName()

	var dnsPolicy string

	if driver.DNSPolicy == "" {
		dnsPolicy = string(corev1.DNSClusterFirstWithHostNet)
	} else {
		dnsPolicy = driver.DNSPolicy
	}

	validPolicy := isValidDNSPolicy(dnsPolicy)

	if !validPolicy {
		return nil, customError.New("invalid DNS Policy provided")
	}

	labels := make(map[string]string)
	labels["app"] = daemonSetName
	containers := make([]corev1.Container, 0)
	initContainers := make([]corev1.Container, 0)
	sa := daemonSetName
	if !rbacRequired {
		// Use default service account if RBAC is not required
		sa = "default"
	}
	privileged := true
	driverSecurityContext := &corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Add: []corev1.Capability{"SYS_ADMIN"},
		},
		Privileged:               &privileged,
		SELinuxOptions:           nil,
		WindowsOptions:           nil,
		RunAsUser:                nil,
		RunAsGroup:               nil,
		RunAsNonRoot:             nil,
		ReadOnlyRootFilesystem:   nil,
		AllowPrivilegeEscalation: nil,
		ProcMount:                nil,
	}
	// Add the node driver container
	containers = append(containers, resources.CreateContainerElement(
		csiv1.ImageTypeDriver, driver.Common.Image, driver.Common.ImagePullPolicy,
		driverArgs, nodeEnv, driverVolumeMounts, driverSecurityContext, nil))
	// Add the registrar
	for _, sideCarContainer := range driver.SideCars {
		if sideCarContainer.Name == csiv1.ImageTypeRegistrar || sideCarContainer.Name == csiv1.ImageTypeSdcmonitor {
			args := sidecarMap[sideCarContainer.Name].Args
			envs := sidecarMap[sideCarContainer.Name].Envs
			volMounts := sidecarMap[sideCarContainer.Name].VolumeMounts
			containers = append(containers, resources.CreateContainerElement(sideCarContainer.Name,
				sideCarContainer.Image, sideCarContainer.ImagePullPolicy,
				args, envs, volMounts, nil, nil))
		}
	}
	//Adding sdc initcontainer
	for _, initContainer := range driver.InitContainers {
		initcontainerName := initContainer.Name
		imageName := initContainer.Image
		initContainers = append(initContainers, resources.CreateContainerElement(
			initcontainerName, imageName, initContainer.ImagePullPolicy,
			initContainerMap[initcontainerName].Args, initContainerMap[initcontainerName].Envs,
			initContainerMap[initcontainerName].VolumeMounts, driverSecurityContext, nil))
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            daemonSetName,
			Namespace:       driverNamespace,
			OwnerReferences: resources.GetOwnerReferences(instance),
		},

		Spec: appsv1.DaemonSetSpec{
			RevisionHistoryLimit: &constants.RevisionHistoryLimit,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &constants.MaxUnavailableUpdateStrategy,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers:                initContainers,
					Containers:                    containers,
					DNSPolicy:                     corev1.DNSPolicy(dnsPolicy),
					HostNetwork:                   true,
					ServiceAccountName:            sa,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 corev1.DefaultSchedulerName,
					TerminationGracePeriodSeconds: &constants.TerminationGracePeriodSeconds,
					Volumes:                       volumes,
					Tolerations:                   podConstraints.Tolerations,
					NodeSelector:                  podConstraints.NodeSelector,
				},
			},
		},
	}, nil
}

// SyncDaemonset - Syncs a daemonset object
func SyncDaemonset(ctx context.Context, daemonset *appsv1.DaemonSet, client client.Client, reqLogger logr.Logger) error {
	//fmt.Println("Creating DaemonSet:", daemonset.Name, daemonset.Namespace)
	found := &appsv1.DaemonSet{}
	err := client.Get(ctx, types.NamespacedName{Name: daemonset.Name, Namespace: daemonset.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DaemonSet", "Namespace", daemonset.Namespace, "Name", daemonset.Name)
		err = client.Create(ctx, daemonset)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating DaemonSet", "Name:", daemonset.Name)
		err = client.Update(ctx, daemonset)
		if err != nil {
			return err
		}
	}
	return nil
}

func isValidDNSPolicy(str string) bool {
	allowedDNSPolicies := []string{string(corev1.DNSClusterFirst), string(corev1.DNSClusterFirstWithHostNet),
		string(corev1.DNSNone), string(corev1.DNSDefault)}

	for _, v := range allowedDNSPolicies {
		if v == str {
			return true
		}
	}
	return false
}
