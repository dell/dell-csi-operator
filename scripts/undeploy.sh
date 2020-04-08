kubectl delete -f deploy/service_account.yaml

kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml

kubectl delete -f deploy/operator.yaml

kubectl delete configmap config-csi-operator
