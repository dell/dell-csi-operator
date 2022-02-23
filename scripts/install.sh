#!/bin/bash

VERIFYSCRIPT="verify.sh"
SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
PROG="${0}"
ROOTDIR="$(dirname "$SCRIPTDIR")"
DEPLOYDIR="$ROOTDIR/deploy"
CONFIGDIR="$ROOTDIR/config"
POWERMAX_CRD="csipowermaxes.storage.dell.com"
POWERMAX_REVPROXY_CRD="csipowermaxrevproxies.storage.dell.com"
ISILON_CRD="csiisilons.storage.dell.com"
UNITY_CRD="csiunities.storage.dell.com"
VXFLEXOS_CRD="csivxflexoses.storage.dell.com"
POWERSTORE_CRD="csipowerstores.storage.dell.com"

#
# usage will print command execution help and then exit
function usage() {
  echo
  echo "Help for $PROG"
  echo
  echo "Usage: $PROG options..."
  echo "Options:"
  echo "  Optional"
  echo "  --upgrade                                Perform an upgrade of the specified driver, default is false"
  echo "  -h                                       Help"
  echo

  exit 0
}

# warning, with an option for users to continue
function warning() {
  echo "*****************************************"
  echo "WARNING:"
  for N in "$@"; do
    echo $N
  done
  echo
  if [ "${ASSUMEYES}" == "true" ]; then
    echo "Continuing as '-Y' argument was supplied"
    return
  fi
  read -n 1 -p "Press 'y' to continue or any other key to exit: " CONT
  echo
  if [ "${CONT}" != "Y" -a "${CONT}" != "y" ]; then
    echo "quitting at user request"
    exit 2
  fi
}

# error, installation will not continue
function errors() {
  echo
  echo "*****************************************"
  printf "${RED}ERROR:"
  for N in "$@"; do
    printf "${RED}$N\n"
  done
  printf "${RED}Installation cannot continue${NC}\n"
  exit 1
}

# print header information
function header() {
  echo "******"
  echo "Installing Dell CSI Operator"
  echo "Kubernetes Version: ${kMajorVersion}.${kMinorVersion}"
  echo
}

# verify K8s configuration
function verify_kubernetes() {
  if [ ! -f "${SCRIPTDIR}/${VERIFYSCRIPT}" ]; then
    log error "Unable to locate ${VERIFYSCRIPT} script in ${SCRIPTDIR}"
  fi
  bash "${SCRIPTDIR}/${VERIFYSCRIPT}"
  case $? in
  0) ;;
  
  1)
    warning "Kubernetes validation failed but installation can continue. " \
      "This may affect driver installation."
    ;;
  *)
    log error "Kubernetes validation failed."
    ;;
  esac
}

function check_for_kubectl() {
  log step "Checking for kubectl installation"
  out=$(command -v kubectl)
  if [ $? -eq 0 ]; then
    log step_success
  else
    log error "Couldn't find kubectl binary in path"
  fi
}

function delete_old_deployment() {
    # deployments
    kubectl delete -n $1 deployment dell-csi-operator-controller-manager 2>&1 >/dev/null
    # clusterRoleBindings
    kubectl delete clusterrolebinding dell-csi-operator-manager-rolebinding 2>&1 >/dev/null
    kubectl delete clusterrolebinding dell-csi-operator-proxy-rolebinding 2>&1 >/dev/null
    # clusterRoles
    kubectl delete clusterrole dell-csi-operator-manager-role 2>&1 >/dev/null
    kubectl delete clusterrole dell-csi-operator-metrics-reader 2>&1 >/dev/null
    kubectl delete clusterrole dell-csi-operator-proxy-role 2>&1 >/dev/null
    # service and serviceAccount
    kubectl delete -n $1 service dell-csi-operator-controller-manager-metrics-service 2>&1 >/dev/null
    kubectl delete -n $1 serviceaccount dell-csi-operator-manager-service-account 2>&1 >/dev/null
    # configMap
    kubectl delete -n $1 configmap dell-csi-operator-config 2>&1 >/dev/null
}

function check_or_create_namespace() {
  # Check if namespace exists
  kubectl get namespace $1 > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Namespace '$1' doesn't exist"
    echo "Creating namespace '$1'"
    kubectl create namespace $1 2>&1 >/dev/null
    if [ $? -ne 0 ]; then
      echo "Failed to create namespace: '$1'"
      echo "Exiting with failure"
      exit 1
    fi
  else
    echo "Namespace '$1' already exists"
  fi
}

function check_for_operator() {
  # check for old operator (ver < 1.2) and fail if the operator are older than 1.2  
  kubectl get deployment -A --no-headers | grep dell-csi-operator | awk {'print $2'} | grep -w ^dell-csi-operator$ --quiet
  if [ $? -eq 0 ]; then
    log error "Found dell-csi-operator version < 1.2. Please remove the old operator manually and retry installation."
    exit 1
  fi
  # get namespace from YAML file for deployment
  NS_STRING=$(cat ${DEPLOYDIR}/operator.yaml | grep "namespace:" | head -1)
  if [ -z "${NS_STRING}" ]; then
    echo "Couldn't find any target namespace in ${DEPLOYDIR}/operator.yaml"
    exit 1
  fi
  # find the namespace from the filtered string
  NAMESPACE=$(echo $NS_STRING | cut -d ' ' -f2)
  EXISTING_OP_NS=''
  
  # check for existing installations in supported namespaces - default, dell-csi-operator
  log step "Checking for existing installations"
  # check operator in default namespace first
  kubectl get pods -n default | grep "dell-csi-operator" --quiet
  if [ $? -eq 0 ]; then
    operator_in_default_namespace=true
    EXISTING_OP_NS='default'
  else
    # check for operator in dell-csi-operator namespace
    kubectl get pods -n dell-csi-operator | grep "dell-csi-operator" --quiet
    if [ $? -eq 0 ]; then
      EXISTING_OP_NS='dell-csi-operator'
      operator_in_default_namespace=true
    else
      kubectl get pods -A | grep "dell-csi-operator" --quiet
      if [ $? -eq 0 ]; then
        operator_in_different_namespace=true
      fi
    fi
  fi
  if [ "$MODE" == "upgrade" ] && [ "$operator_in_default_namespace" = true ]; then
    log step_warning
    log warning "Found existing installation of Operator in '$EXISTING_OP_NS' namespace"
    echo "Attempting to upgrade the Operator as --upgrade option was specified"
    kubectl get deployment -n $EXISTING_OP_NS dell-csi-operator-controller-manager | grep "dell-csi-operator" --quiet
    if [[ $? -eq 0 ]]; then
      delete_old_deployment $EXISTING_OP_NS
    fi
  elif [ "$operator_in_default_namespace" = true ]; then
    log step_failure
    log warning "Found existing installation of dell-csi-operator in '$EXISTING_OP_NS' namespace"
    log error "Remove the existing installation manually, or use the --upgrade option to upgrade the Operator"
    exit 1
  elif [ "$operator_in_different_namespace" = true ]; then
    log step_failure
    log warning "Found existing installation of dell-csi-operator in a different namespace"
    log error "Remove the existing installation manually, and then proceed with installation"
    exit 1
  else
    log step_success
  fi
}

function install_or_update_driver_crd() {
  log step "Install/Update CRDs"
  kubectl apply -f ${DEPLOYDIR}/crds/storage.dell.com.crds.all.yaml 2>&1 >/dev/null
  if [ $? -ne 0 ]; then
    log error "Failed to install/update CRD"
  fi
  log step_success
}

function create_or_update_configmap() {
  log step "Create temporary archive"
  (cd "$ROOTDIR" && tar -cf - driverconfig/* | gzip > config.tar.gz)
  if [ $? -ne 0 ]; then
    log error "Failed to create temporary archive for operator"
  fi
  log step_success
  log step "Create/Update ConfigMap"
  kubectl create configmap dell-csi-operator-config --from-file "$ROOTDIR/config.tar.gz" -o yaml --dry-run=client | kubectl apply -n $1 -f - > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    log error "Failed to create/update ConfigMap for operator"
  fi
  log step_success
  log step "Removing temporary archive"
  yes | rm "$ROOTDIR/config.tar.gz" 2>&1 >/dev/null
  if [ $? -ne 0 ]; then
    log step_failure
    log warning "Failed to remove temporary archive"
  else
    log step_success
  fi
}

function create_operator_deployment() {
  log step "Install Operator"
  kubectl apply -f ${DEPLOYDIR}/operator.yaml 2>&1 >/dev/null
  if [ $? -ne 0 ]; then
    log error "Failed to deploy operator"
  fi
  log step_success
}

function install_operator() {
  log separator
  echo "Installing Operator"
  log separator
  install_or_update_driver_crd
  log separator
  create_or_update_configmap $NAMESPACE
  create_operator_deployment $NAMESPACE
  log separator
}

function check_progress() {
  # find out the deployment name
  # wait for the deployment to finish, use the default timeout
  waitOnRunning "${NAMESPACE}" "deployment dell-csi-operator-controller-manager"
  if [ $? -eq 1 ]; then
    warning "Timed out waiting for installation of the operator to complete." \
      "This does not indicate a fatal error, pods may take a while to start." \
      "Progress can be checked by running \"kubectl get pods\""
  fi
}

# Print a nice summary at the end
function summary() {
  echo
  echo "******"
  echo "Installation complete"
  echo
}

#
# main
#
ASSUMEYES="false"
OPERATOR="dell-csi-operator"

while getopts ":h-:" optchar; do
  case "${optchar}" in
  -)
    case "${OPTARG}" in
    upgrade)
      MODE="upgrade"
      ;;
    *)
      echo "Unknown option --${OPTARG}"
      echo "For help, run $PROG -h"
      exit 1
      ;;
    esac
    ;;
  h)
    usage
    ;;
  *)
    echo "Unknown option -${OPTARG}"
    echo "For help, run $PROG -h"
    exit 1
    ;;
  esac
done

source "$SCRIPTDIR"/common.bash

header
check_for_kubectl
check_for_operator
verify_kubernetes
check_or_create_namespace $NAMESPACE
install_operator $NAMESPACE
check_progress

summary
