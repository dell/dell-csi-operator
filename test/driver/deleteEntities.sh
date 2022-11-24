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

#This script is to delete the objects created by install script
#Parameters are taken from config.properties

source ./unity_driver_config
echo "Driver file is $OPERATOR_ENV_DRIVER_MANIFEST"

# Deletion of Controller and Deletion pod
# -----------------------------------------
if [ -z "$OPERATOR_ENV_DRIVER_MANIFEST" ]; then
 echo "Driver yaml file name is not present in config.properties file; Exiting execution"
 exit 1
fi

if [ -f $OPERATOR_ENV_DRIVER_MANIFEST ];then
 echo "$OPERATOR_ENV_DRIVER_MANIFEST present;proceeding further"
else
 echo "$OPERATOR_ENV_DRIVER_MANIFEST is not present;Exiting the execution"
 exit 1
fi


#cd ../../test/driver/

target_yaml_file=$OPERATOR_ENV_DRIVER_MANIFEST
echo "About to delete the driver pods..."
kubectl delete -f  $target_yaml_file
delete_pod_validation=$?
sleep 5
if [ "$delete_pod_validation" -ne 0 ]; then
   echo "Pod deletion - command failed; Terminating the script"
   exit 1
else
   echo "Deleting the pod - command successful; Proceeding further..."
fi
sleep 60

pod_validation=`kubectl get pods -n $OPERATOR_ENV_DRIVER_NAMESPACE`

if [ -z "$pod_validation" ] 
then
   echo "**************************************************"
   echo "Controller and node pod got deleted successfully"
   echo "**************************************************"
   echo "About to execute Operator uninstallation"
else
   echo "Pods still exists;Problem in Deleting Controller/node pod"
   exit 1
fi

# Deletion of CSI-Operator
# -------------------------------

echo "Deleting csi-operator via uninstall script"
cd ../../ && sh scripts/uninstall.sh
if [ "$?" -ne 0 ]; then
 echo "Problem in executing uninstall.sh"
 echo "Check the presence of csi-operator by the command kubectl get pods"
 echo "Exiting the script without checking csi-drivers"
 exit 1
fi
sleep 60
cd test/driver
operator_pod_validation=`kubectl get pods`
sleep 10
if [ -z "$operator_pod_validation" ]
then
    echo "******************************************"
    echo "csi-operator got uninstalled successfully"
    echo "******************************************"
else
   echo "Problem in undeploying dell-csi-operator; operator may still be present.."
   echo "Check the existence of dell-csi-operator by the command kubectl get pods"
   echo "Exiting the script without checking csi-drivers"
   exit 1
fi

