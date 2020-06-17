#!/bin/bash
echo "Creating Service Account for Operator"
echo "kubectl create -f deploy/service_account.yaml"
kubectl create -f deploy/service_account.yaml
echo "Creating ClusterRole for the Operator"
echo "kubectl create -f deploy/role.yaml"
kubectl create -f deploy/role.yaml
echo "Creating ClusterRoleBinding for the Operator"
echo "kubectl create -f deploy/role_binding.yaml"
kubectl create -f deploy/role_binding.yaml
echo "Create a temporary archive from config files"
echo "tar -czvf config.tar.gz config/"
tar -czf config.tar.gz config/
echo "Create ConfigMap using the archive"
echo "kubectl create configmap config-dell-csi-operator --from-file config.tar.gz"
kubectl create configmap config-dell-csi-operator --from-file config.tar.gz
echo "Delete the temporary archive"
echo "rm -f config.tar.gz"
rm -f config.tar.gz
echo "********"
echo "Creating CRDs for all driver types"
kubectl apply -f deploy/crds/csiisilons.storage.dell.com.crd.yaml
kubectl apply -f deploy/crds/csipowermaxes.storage.dell.com.crd.yaml
kubectl apply -f deploy/crds/csiunities.storage.dell.com.crd.yaml
kubectl apply -f deploy/crds/csivxflexoses.storage.dell.com.crd.yaml
echo "********"
echo "Deploying the Operator"
echo "kubectl create -f deploy/operator.yaml"
kubectl create -f deploy/operator.yaml
