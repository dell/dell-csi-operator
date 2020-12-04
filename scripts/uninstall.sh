#!/bin/bash
echo "** Deleting the Operator Deployment **"
echo
echo kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/operator.yaml
echo
echo "** Deleting ConfigMap **"
echo
echo kubectl delete configmap dell-csi-operator-config
kubectl delete configmap dell-csi-operator-config
echo
echo "Removing temporary archive"
echo rm -f config.tar.gz
rm -f config.tar.gz
