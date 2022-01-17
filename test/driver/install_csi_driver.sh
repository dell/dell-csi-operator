#!/bin/sh
# Shell script to deploy operators and to create CSI Drivers
# Author: gokul.srivathsan@dell.com
# Furnish proper values in config.properties file before executing this script
# Way to execute: sh install_csi_driver.sh
source ./testlib.sh
source ./config.properties
source ./sample_driver_config

echo "-----------------------------------------------"

echo "namespace is $OPERATOR_ENV_DRIVER_NAMESPACE"
echo "manifest is $OPERATOR_ENV_DRIVER_MANIFEST"

parent_folder='./../../'
# Validating the values of config properties
# ---------------------------------------------

shopt -s nocasematch

if [ -z "$driver_type" ]; then
 echo "Driver type parameter has empty value; Terminating the execution"
 exit 1
 elif [[ $driver_type == "unity" ]] || [[ $driver_type == "powermax" ]] || [[ $driver_type == "vxflex" ]] || [[ $driver_type == "isilon" ]]; then
 echo "Driver type is valid-Good to go..."
 else
 echo "Driver type is not valid. Terminating the execution"
 exit 1
fi


# powermax specific parameter validation
# --------------------------------------

if [[ "$driver_type" == "powermax" ]]
then

 if [ -z "$powermax_secret" ]; then
   echo "powermax secret is not set. Terminating the execution"
   exit 1
 fi

 if [ -z "$X_CSI_K8S_CLUSTER_PREFIX" ]; then
    echo "The cluster prefix variable is not set. Terminating the execution"
    exit 1
 fi

# if [ -z "$symmetrixID" ]; then
#  echo "symmetrixID is not set. Terminating the execution"
#  exit 1
# fi


fi 



# unity specific parameter validation
# ------------------------------------

if [[ "$driver_type" == "unity" ]]; then
 if [ -z "$unity_creds_secret" ] || [ -z "unity_certs_secret" ]; then
   echo "unity secret is not set. Terminating the execution"
   exit 1
 fi
 
fi


# General parameter validation
# ----------------------------

if [ -z "$storage_array" ];then
 echo "Storage array value is empty in config.properties; Terminating the execution"
 exit 1
fi

if [ -z "$OPERATOR_ENV_DRIVER_NAMESPACE" ]; then
 echo "Namespace value is empty in config.properties; Terminating the execution"
 exit 1
fi


if [ -z "$OPERATOR_ENV_DRIVER_MANIFEST" ]; then
 echo "Driver yaml file name is empty in config.properties; Terminating the execution"
 exit 1
fi

# Deleting existing temporary files
# ---------------------------------

[ -e docker-build-output.log ] && rm -f docker-build-output.log
[ -e deploy_output.log ] && rm -f deploy_output.log

sleep 2
# Furnishing operator build name in operator.yaml
operator_build="$1"
echo "operator image is $operator_build"
sed -i "s/dell-csi-operator:.*/${operator_build}/g" ../../deploy/operator.yaml

# Running install script
echo "About to execute install.sh..."
echo "current directory is $PWD"
cd ../../ && sh scripts/install.sh > ./test/driver/deploy_output.log
cd test/driver
sleep 50

# Validation of csi-operator
# -----------------------------------
echo "About to validate csi-operator..."
csi_operator_validation=$(kubectl get pods | grep -i "dell-csi-operator" | grep -i "Running")
if [ -n "$csi_operator_validation" ]
then
	csi_operator_name=${csi_operator_validation:0:34}
	echo
	echo "*****************************************************"
	echo "csi operator is $csi_operator_name"
	echo "*****************************************************"
	echo
	csi_operator_status=0
else
	echo "Problem with csi-operator creation"
	csi_operator_status=1
	echo "Exiting the execution"
	exit 1
fi

# Checking the status of CSI-Operator
# --------------------------------------

if [ "$csi_operator_status" == "0" ]
then
	temp_status=$(kubectl describe pod $csi_operator_name)
	csi_operator_description_status=$(echo $temp_status | grep -i "State: Running")
	if [ -n "$csi_operator_description_status" ]
	then
		echo
		echo "*********************************"
		echo "csi operator is up and running"
		echo "*********************************"
		echo
        else
		echo "Problem in csi operator creation"
		echo "Exiting the execution"
		exit 1
	fi
else
	echo "Exiting without checking running status of csi-operator object"
	exit 1
fi

# checking the presence of namespace by calling checknamespace function
# ----------------------------------------------------------------------
echo "About to validate namespace..."
checknamespace "$OPERATOR_ENV_DRIVER_NAMESPACE"
nsretval=$?
if [ $nsretval -eq 0 ]
then
   echo
   echo "********************************************************"
   echo "Required namespace is present"
   echo "********************************************************"
   echo
else
   echo "Required namespace not present;Terminating the execution"
   exit 1
fi 

# checking the presence of secrets by calling checksecret function
# ----------------------------------------------------------------

if [[ "$driver_type" == "unity" ]]; then
 echo "About to check unity secrets..."
 check_unity_secret "$unity_creds_secret" "$unity_certs_secret" "$OPERATOR_ENV_DRIVER_NAMESPACE"
 retval=$?
 if [ $retval -eq 0 ]
 then
	echo
	echo "*******************************************************"
	echo "Both secrets are fine. Can proceed to driver creation"
	echo "*******************************************************"
	echo
 else
	echo "secrets are not present; Terminating the execution"
	exit 1
 fi
else
 echo "About to check powermax secrets..."
 check_powermax_secret "$powermax_secret" "$namespace"
 retval=$?
 if [ $retval -eq 0 ]
 then
	echo
	echo "*********************************************************"
        echo "Secret is present. Can proceed to driver creation"
	echo "********************************************************"
	echo
 else
	echo "Secret is not present or corrupted; Terminating the execution"
	exit 1
 fi
fi
 
#add code to validate  secrets for isilon
# add code to validate secrets for vxflex

# calling furnish driver yaml function from testlib.sh
# ----------------------------------------------------
echo "Furnishing driver yaml"

if [[ "$driver_type" == "unity" ]];then
	echo "driver_build_name is $driver_build_name"
	go run ./operatorutils.go
	retval=$?
elif [[ "$driver_type" == "powermax" ]];then
	echo "driver_build_name is $driver_build_name"
	echo "Powermax driver deployment is yet to be done...................."
	go run ./operatorutils.go
	return=$?
 elif [[ "$driver_type" == "vxflex" ]];then
	echo "driver_build_name is $driver_build_name"
	echo "vxflex driver deployment is yet to be done...."
	go run ./operatorutils.go
 	retval=$?
else
	echo "driver_build_name is $driver_build_name"
	echo "Isilon driver deployment is yet to be done ...."
	go run ./operatorutils.go
 	retval=$?
fi

if [ $retval -eq 0 ];then

	echo "About to create CSI Driver.Please wait..."
	kubectl create -f $OPERATOR_ENV_DRIVER_MANIFEST
	if [ $? -eq 0 ]; then
		  sleep 50
		echo "kubectl get pods -n  $OPERATOR_ENV_DRIVER_NAMESPACE"
	  	kubectl get pods -n  "$OPERATOR_ENV_DRIVER_NAMESPACE"
		validate_driver_pods "$OPERATOR_ENV_DRIVER_NAMESPACE"
        	pod_validation=$?
	  if [ "$pod_validation" == "0" ];then
	    echo "CSI Driver creation in $OPERATOR_ENV_DRIVER_NAMESPACE is success"
            exit 0
          else
	    echo "CSI Driver creation in $OPERATOR_ENV_DRIVER_NAMESPACE has problems; Terminating the execution"
            exit 1
	  fi
	  
	  pod_validation=`kubectl get pods -n "$OPERATOR_ENV_DRIVER_NAMESPACE" | grep Running`
	  if [ -z "$pod_validation" ]; then
	  	echo "Pod creation has problem; Please execute kubectl get pods -n <namespace>"
	  else
 	  	echo
		echo "*************************************************************************************************"
		echo "CSI Driver Pod creation in $OPERATOR_ENV_DRIVER_NAMESPACE is success"
		echo
		kubectl get pods -n "$OPERATOR_ENV_DRIVER_NAMESPACE" -o wide
		echo "**************************************************************************************************"
		echo ""
          fi
        else
	  echo "Problem in creating csi driver pods"
	  echo "Exiting the execution"
	  shopt -u nocasematch
	  exit 1
	fi
shopt -u nocasematch
exit 0
else
 echo "Exiting the execution without installing csi-drivers"
 exit 1
fi
