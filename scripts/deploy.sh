kubectl create -f deploy/service_account.yaml

kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml

#TODO identify k8s version and use properties file accordingly
kubectl create configmap config-csi-operator --from-file config/
kubectl create -f deploy/operator.yaml

kubectl apply -f deploy/crds/csiisilons.storage.dell.com.crd.yaml
kubectl apply -f deploy/crds/csipowermaxes.storage.dell.com.crd.yaml
kubectl apply -f deploy/crds/csiunities.storage.dell.com.crd.yaml
kubectl apply -f deploy/crds/csivxflexoses.storage.dell.com.crd.yaml

