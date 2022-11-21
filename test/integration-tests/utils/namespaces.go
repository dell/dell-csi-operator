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
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates a namespace with the specified name using the provided client.
func CreateNamespace(namespace string, kubeClient kubernetes.Interface) error {
	var err error
	fmt.Printf("Creating namespace '%s'...\n", namespace)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	ns, err = kubeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Failed to create namespace '%s': %v", namespace, err)
		return err
	}

	fmt.Printf("Successfully created namespace '%s'\n", ns.Name)
	return nil
}

// DeleteNamespace deletes a namespace with the specified name using the provided client.
func DeleteNamespace(namespace string, kubeClient kubernetes.Interface) error {
	var err error
	fmt.Printf("Deleting namespace '%s'...\n", namespace)

	err = kubeClient.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Failed to delete namespace '%s': %v", namespace, err)
		return err
	}

	fmt.Printf("Successfully deleted namespace '%s'\n", namespace)
	return nil
}

// NamespaceExists determines whether a namespace with the specified name exists using the provided client.
func NamespaceExists(namespace string, kubeClient kubernetes.Interface) bool {
	var err error

	nss, err := kubeClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list namespaces %v", err)
		return false
	}
	for _, ns := range nss.Items {
		if ns.Name == namespace {
			return true
		}
	}

	return false
}
