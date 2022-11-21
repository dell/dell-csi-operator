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

package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// NewSecret creates a basic auth secret
func NewSecret(secretName, namespace, username, password, chapsecret string) *corev1.Secret {
	secret := &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"username": []byte(username),
		},
	}
	if password != "" {
		secret.Data["password"] = []byte(password)
	}
	if chapsecret != "" {
		secret.Data["chapsecret"] = []byte(chapsecret)
	}
	return secret
}

// JSONPrettyPrint - Indent the json output
func JSONPrettyPrint(obj interface{}) string {
	s, err := json.Marshal(obj)
	if err != nil {
		return "Not a instance"
	}
	var out bytes.Buffer
	err = json.Indent(&out, s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return out.String()
}

// DefaultRetry is the default backoff for e2e tests.
var DefaultRetry = wait.Backoff{
	Steps:    50,
	Duration: 8 * time.Second,
	Factor:   1.0,
	Jitter:   0.1,
}

// SmallRetry holds parameters applied to a Backoff function
var SmallRetry = wait.Backoff{
	Steps:    5,
	Duration: 8 * time.Second,
	Factor:   1.0,
	Jitter:   0.1,
}

// Retry executes the provided function repeatedly, retrying until the function
// returns done = true, errors, or exceeds the given timeout.
func Retry(backoff wait.Backoff, fn wait.ConditionFunc) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		done, err := fn()
		if err != nil {
			lastErr = err
		}
		return done, err
	})
	if err == wait.ErrWaitTimeout {
		if lastErr != nil {
			err = lastErr
		}
	}
	return err
}

// WaitForDaemonSetAvailable - Waits for the given daemonset to available
func WaitForDaemonSetAvailable(namespace string, daemonSetName string, backoff wait.Backoff, kubeClient kubernetes.Interface, log *logrus.Logger) error {
	var err error
	log.Infof("Waiting for daemonset '%s' to available...\n", daemonSetName)
	err = Retry(backoff, func() (bool, error) {
		daemonSets, err := kubeClient.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		if len(daemonSets.Items) == 1 {
			log.Infof("Daemonset '%s' created successfully.\n", daemonSetName)
			return true, nil
		}
		return false, nil
	})
	return err
}

// WaitForPods Waits for the given deployment to contain the given PVC
func WaitForPods(namespace string, kubeClient kubernetes.Interface, driverType string, log *logrus.Logger) error {
	var err error
	log.Infof("Waiting for pods in namespace '%s' to start...\n", namespace)
	err = Retry(SmallRetry, func() (bool, error) {
		podItems, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})

		if err != nil {
			return false, err
		}
		var finalErr string
		pods := podItems.Items
		var controllerContainers, nodeContainers int
		controllerContainers = 0
		nodeContainers = 0
		for _, pod := range pods {
			fmt.Println("-----------", pod.Name, pod.Status.ContainerStatuses[1].Ready)
			if strings.Contains(pod.Name, driverType+"-controller") {
				controllerContainers++
				if pod.Status.Phase != corev1.PodRunning {
					finalErr = fmt.Sprintf("%s Pod [%s] is not running status :[%s].\n", finalErr, pod.Name, pod.Status.Phase)
				}
			} else if strings.Contains(pod.Name, driverType+"-node") {
				nodeContainers++
				if pod.Status.Phase != corev1.PodRunning {
					finalErr = fmt.Sprintf("%s Pod [%s] is not running status :[%s].\n", finalErr, pod.Name, pod.Status.Phase)
				}
			}
		}
		if controllerContainers == 0 || nodeContainers == 0 {
			finalErr = fmt.Sprintf("Controller or node pods are not creating properly")
			return false, errors.New(finalErr)
		}
		if finalErr != "" {
			return false, errors.New(finalErr)
		}
		return true, nil
	})
	return err
}
