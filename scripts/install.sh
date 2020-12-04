#!/bin/bash

VERIFYSCRIPT="verify.sh"
SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
PROG="${0}"
ROOTDIR="$(dirname "$SCRIPTDIR")"
SNAPCLASSDIR="$ROOTDIR/deploy/beta-snapshot-crds"
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
  echo "  --snapshot-crd                           Install snapshot CRDs. Default will not install beta snapshot crds."
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
  if [ "$INSTALL_CRD" == "yes" ]; then
    bash "${SCRIPTDIR}/${VERIFYSCRIPT}" --skip-betacrd-validation
  else
    bash "${SCRIPTDIR}/${VERIFYSCRIPT}"
  fi
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
  kubectl delete deployment dell-csi-operator 2>&1 >/dev/null
  kubectl delete clusterrolebinding dell-csi-operator 2>&1 >/dev/null
  kubectl delete clusterrole dell-csi-operator 2>&1 >/dev/null
  kubectl delete serviceaccount dell-csi-operator 2>&1 >/dev/null
  kubectl delete configmap config-dell-csi-operator 2>&1 >/dev/null
}

function check_for_operator() {
  log step "Checking for existing installations"
  kubectl get pods -n default | grep "dell-csi-operator" --quiet
  if [ $? -eq 0 ]; then
    operator_in_default_namespace=true
  else
    kubectl get pods -A | grep "dell-csi-operator" --quiet
    if [ $? -eq 0 ]; then
      operator_in_different_namespace=true
    fi
  fi
  if [ "$MODE" == "upgrade" ] && [ "$operator_in_default_namespace" = true ]; then
    log step_warning
    log warning "Found existing installation of Operator in default namespace"
    echo "Attempting to upgrade the Operator as --upgrade option was specified"
    kubectl get deployment dell-csi-operator | grep "dell-csi-operator" --quiet
    if [[ $? -eq 0 ]]; then
      delete_old_deployment
    fi
  elif [ "$operator_in_default_namespace" = true ]; then
    log step_failure
    log warning "Found existing installation of dell-csi-operator in default namespace"
    log error "Remove the existing installation or use the --upgrade option to upgrade the Operator"
  elif [ "$operator_in_different_namespace" = true ]; then
    log step_failure
    log warning "Found existing installation of dell-csi-operator"
    log error "Remove the existing installation and then proceed with installation"
  else
    log step_success
  fi
}

function install_beta_snapshot_crd() {
  declare -A SNAPCLASSES=(
    ["volumesnapshotclasses"]="snapshot.storage.k8s.io_volumesnapshotclasses.yaml"
    ["volumesnapshotcontents"]="snapshot.storage.k8s.io_volumesnapshotcontents.yaml"
    ["volumesnapshots"]="snapshot.storage.k8s.io_volumesnapshots.yaml"
  )
  for C in "${!SNAPCLASSES[@]}"; do
    if [ ! -f "${SNAPCLASSDIR}/${SNAPCLASSES[$C]}" ]; then
        echo "Unable to to find Snapshot Classes CRD file at ${SNAPCLASSDIR}"
        exit 1
      fi
    kubectl get crd | grep "${C}" --quiet
    if [[ $? -ne 0 ]]; then
      # make sure CRD exists
      log step "Installing CRD: $C"
      # create the custom resource
      kubectl create -f "${SNAPCLASSDIR}/${SNAPCLASSES[$C]}" > /dev/null 2>&1
      if [ $? -ne 0 ]; then
        log step_failure
        log error "Failed to install $CRD"
      else
        log step_success
      fi
    else
      log step "Updating CRD: $C"
      kubectl apply -f "${SNAPCLASSDIR}/${SNAPCLASSES[$C]}" > /dev/null 2>&1
      if [ $? -ne 0 ]; then
        log error "Failed to update $CRD"
      fi
      log step_success
    fi
  done
  sleep 3s
}

function install_or_update_driver_crd() {
  log step "Install/Update CRDs"
  kubectl apply -f ${DEPLOYDIR}/crds/storage.dell.com.crds.all.yaml 2>&1 >/dev/null
  if [ $? -ne 0 ]; then
    log error "Failed to create cluster role binding for operator"
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
  kubectl create configmap dell-csi-operator-config --from-file "$ROOTDIR/config.tar.gz" -o yaml --dry-run | kubectl apply -f - > /dev/null 2>&1
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
    log error "Failed to create cluster role binding for operator"
  fi
  log step_success
}

function install_operator() {
  log separator
  echo "Installing Operator"
  log separator
  install_or_update_driver_crd
  log separator
  create_or_update_configmap
  create_operator_deployment
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
NAMESPACE="default"

while getopts ":h-:" optchar; do
  case "${optchar}" in
  -)
    case "${OPTARG}" in
      # SNAPSHOT_CRD
    snapshot-crd)
      INSTALL_CRD="yes"
      ;;
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
if [ "$INSTALL_CRD" == "yes" ]; then
  install_beta_snapshot_crd
fi
install_operator
check_progress

summary