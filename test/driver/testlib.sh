#!/bin/sh
#
#  Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
# 
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#       http://www.apache.org/licenses/LICENSE-2.0
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# Function to check the secrets present in the system

check_unity_secret()
{
cred_secret=$(kubectl get secrets -n "$3" |grep -i "$1")
cert_secret=$(kubectl get secrets -n "$3" |grep -i "$2")
if  [ -n "$cred_secret" ] && [ -n "$cert_secret" ]
 then
	retval=0
 else
	echo "problem with secrets;secrets may not be present or corrupted"
	retval=1
fi
return "$retval"
}

check_powermax_secret()
{
echo "secodn: $2 and one: $1"
cert_secret=$(kubectl get secrets -n "$2" | grep -i "$1")
echo "printing secretvalue:"$cert_secret""
if [ -n "$cert_secret" ]
then
  retval=0
else
 echo "problem with secrets;secrets may not be present or corrupted"
 retval=1
fi
return "$retval"
}


# ------------------------------------------------------------------------------------
# Function to check whether required namespace is present
checknamespace()
{
 echo "Given namespace is $1"
 namespaceExists=$(kubectl get namespace | grep -i "$1")
 if [ -n "$namespaceExists" ]
 then
	nsretval=0
 else
	nsretval=1
 fi
 return "$nsretval"
}


# Function for validating driver pods
# ----------------------------------------

validate_driver_pods()
{
  mycontroller=$(kubectl get pods -n "$1" | grep controller-)
  controller_running_state=`echo $mycontroller | cut -d ' ' -f3`
  controller_ready_state=`echo $mycontroller | cut -d ' ' -f2`
  
  mynode=$(kubectl get pods -n "$1" | grep node)
  node_running_state=`echo $mynode | cut -d ' ' -f3`
  node_ready_state=`echo $mynode | cut -d ' ' -f2`
  
  if [[ $controller_running_state == "Running" ]] && [[ $controller_ready_state == "5/5" ]]; then
   echo "Driver controller statefulset is working fine"
   driver_pod_health=0
  else
   echo "Driver controller may not be running or not in a good status"
   driver_pod_health=1
  fi
 
  if [[ $node_running_state == "Running" ]] && [[ $node_ready_state == "2/2" ]]; then
   echo "Node pod is working fine"
   driver_pod_health=0
  else
   echo "Node pod may not be running or not in a good status"
   driver_pod_health=1
  fi
  return "$driver_pod_health"
}

#---------------------------------------------------------------------------

#Function for furnishing the operator.yaml file.

furnish_operator_yaml()
{
rm -f ./../../deploy/operator.yaml
echo "Given parameter in furnish_operator_yaml is $1"
cat <<EOF >> ./../../deploy/operator.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: csi-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: csi-operator
  template:
    metadata:
      labels:
        name: csi-operator
    spec:
      serviceAccountName: csi-operator
      containers:
        - name: csi-operator
          #Replace this with the built image name
          image: "$1"
          imagePullPolicy: IfNotPresent
          args:
            - "--zap-level=debug"
            - "--zap-encoder=console"
          env:
            - name: WATCH_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPERATOR_NAME
              value: "csi-operator"
            - name: OPERATOR_DRIVERS
              value: "unity,powermax,isilon,vxflexos"
          volumeMounts:
            - name: config-volume
              mountPath: /etc/config/csi-operator
      volumes:
        - name: config-volume
          configMap:
            # Provide the name of the ConfigMap containing the files you want
            # to add to the container
            name: config-csi-operator
EOF
sleep 2

	FF=`ls ../../deploy/operator.yaml  | awk -F "/" '{print $4}'`
	if [ "$FF" == "operator.yaml" ];then 
		echo "Operator.yaml got created"
	else
		echo "Operator.yaml file is not present"
		exit 1
	fi
}

#------------------------------------------------------------------
