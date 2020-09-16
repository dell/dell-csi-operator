#!/bin/bash
echo "Deleting the Operator Deployment"
echo "kubectl delete -f deploy/operator.yaml"
kubectl delete -f deploy/operator.yaml
echo "Deleting the ConfigMap"
echo "kubectl delete configmap config-dell-csi-operator"
kubectl delete configmap config-dell-csi-operator
echo "Deleting the ClusterRoleBinding"
echo "kubectl delete -f deploy/role_binding.yaml"
kubectl delete -f deploy/role_binding.yaml
echo "Deleting the ClusterRole"
echo "kubectl delete -f deploy/role.yaml"
kubectl delete -f deploy/role.yaml
echo "Deleting the ServiceAccount"
echo "kubectl delete -f deploy/service_account.yaml"
kubectl delete -f deploy/service_account.yaml
