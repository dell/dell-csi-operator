#!/bin/bash
# This script does the following:
# 1. Create a CatalogSource containing index for various Operator versions
# 2. Create an OperatorGroup
# 3. Create a subscription for the Operator

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOTDIR="$(dirname "$SCRIPTDIR")"
DEPLOYDIR="$ROOTDIR/config/olm"

# Constants
CATALOGSOURCE="dellemc-registry"
OPERATORGROUP="dellemc-operators"
SUBSCRIPTION="dellemc-subscription"
COMMUNITY_MANIFEST="operator_community.yaml"
CERTIFIED_MANIFEST="operator_certified.yaml"

MANIFEST_FILE="$DEPLOYDIR/$COMMUNITY_MANIFEST"

if [ "$1" == "--certified" ]; then
  MANIFEST_FILE="$DEPLOYDIR/$CERTIFIED_MANIFEST"
fi

unableToFindKubectlErrorMsg="Install kubectl before running this script"
uninstallComponentErrorMsg="Remove all existing installations before running this script"
installOLMErrorMsg="Install all OLM components correctly before running this script"


catsrccrd="catalogsources.operators.coreos.com"
csvcrd="clusterserviceversions.operators.coreos.com"
ipcrd="installplans.operators.coreos.com"
opgroupcrd="operatorgroups.operators.coreos.com"
subcrd="subscriptions.operators.coreos.com"


function check_for_kubectl() {
  out=$(command -v kubectl)
  if [ $? -ne 0 ]; then
    exit_with_error "Couldn't find kubectl binary in path" $unableToFindKubectlErrorMsg
  fi
}

function check_for_olm_components() {
  kubectl get crd | grep $catsrccrd --quiet
  if [ $? -ne 0 ]; then
    exit_with_error "Couldn't find $catsrccrd" "$installOLMErrorMsg"
  fi
  kubectl get crd | grep $csvcrd --quiet
  if [ $? -ne 0 ]; then
    exit_with_error "Couldn't find csvcrd" "$installOLMErrorMsg"
  fi
  kubectl get crd | grep $ipcrd --quiet
  if [ $? -ne 0 ]; then
    exit_with_error "Couldn't find $ipcrd" "$installOLMErrorMsg"
  fi
  kubectl get crd | grep $opgroupcrd --quiet
  if [ $? -ne 0 ]; then
    exit_with_error "Couldn't find $opgroupcrd" "$installOLMErrorMsg"
  fi
  kubectl get crd | grep $subcrd --quiet
  if [ $? -ne 0 ]; then
    exit_with_error "Couldn't find $subcrd" "$installOLMErrorMsg"
  fi
}

function exit_with_error() {
  echo "$1"
  echo "$2"
  echo
  echo "Exiting with error"
  exit 1
}

function check_existing_installation() {
  kubectl get catalogsource "$CATALOGSOURCE" -n $NS > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    exit_with_error "A CatalogSource with name $CATALOGSOURCE already exists in namespace $NS" $uninstallComponentErrorMsg
  fi
  kubectl get operatorgroup "$OPERATORGROUP" -n $NS > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    exit_with_error "An OperatorGroup with name $OPERATORGROUP already exists in namespace $NS" $uninstallComponentErrorMsg
  fi
  kubectl get Subscription "$SUBSCRIPTION" -n "$NS" > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    exit_with_error "A Subscription with name $SUBSCRIPTION already exists in namespace $NS" $uninstallComponentErrorMsg
  fi
}

function set_namespace() {
  NS="test-olm"
  echo "Operator will be installed in namespace: $NS"
}

function check_or_create_namespace() {
  # Check if namespace exists
  kubectl get namespace $NS > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Namespace $NS doesn't exist"
    echo "Creating namespace $NS"
    echo "kubectl create namespace $NS"
    kubectl create namespace $NS 2>&1 >/dev/null
    if [ $? -ne 0 ]; then
      echo "Failed to create namespace: $NS"
      echo "Exiting with failure"
      exit 1
    fi
  else
    echo "Namespace $NS already exists"
  fi
}

function install_operator() {
  echo "*****"
  echo
  echo "Installing Operator"
  echo "kubectl apply -f $MANIFEST_FILE"
  kubectl apply -f $MANIFEST_FILE
}

check_for_kubectl
check_for_olm_components
set_namespace
check_or_create_namespace
check_existing_installation
install_operator

echo "The installation will take some time to complete"
echo "If the installation is successful, a CSV with the status 'Succeeded' should be created in the namespace $NS"
