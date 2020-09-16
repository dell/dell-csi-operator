#!/bin/bash

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
k8s16=16

skip_beta_crd_flag=""
if [ $# -ne 0 ]; then
  skip_beta_crd_flag=$1
fi

if [ "$skip_beta_crd_flag" == "--skip-betacrd-validation" ]; then
  skip_beta_crd_validation=true
fi

VOLUMESNAPSHOTCLASSCRD="volumesnapshotclasses.snapshot.storage.k8s.io"
VOLUMESNAPSHOTCONTENTCRD="volumesnapshotcontents.snapshot.storage.k8s.io"
VOLUMESNAPSHOTSCRD="volumesnapshots.snapshot.storage.k8s.io"

# print header information
function header() {
	echo
	echo "******"
	echo "Verifying configuration"
	echo "Kubernetes Version: ${kMajorVersion}.${kMinorVersion}"
	echo "Openshift: ${isOpenShift}"
	echo
}

# verify minimum k8s version
function verify_min_k8s_version() {
	log step "Verifying minimum Kubernetes version."
	if [[ "${1}" -gt "${kMajorVersion}" ]]; then
		RESULT_K8S_MIN_VER="Failed"
    AddError "Kubernetes version, ${kMajorVersion}.${kMinorVersion}, is too old. Minimum required version is: ${1}.${2}"
	fi
	if [[ "${2}" -gt "${kMinorVersion}" ]]; then
		RESULT_K8S_MIN_VER="Failed"
    AddError "Kubernetes version, ${kMajorVersion}.${kMinorVersion}, is too old. Minimum required version is: ${1}.${2}"
  fi
  # Following check is temporary as we still support OpenShift 4.3
  if [[ "${kMinorVersion}" -eq "${k8s16}" ]]; then
    if ! "$isOpenShift43"; then
		  RESULT_K8S_MIN_VER="Failed"
      AddError "Upstream Kubernetes version, ${kMajorVersion}.${kMinorVersion}, is too old. Minimum required version is: 17"
    fi
  fi
	if [ $RESULT_K8S_MIN_VER == "Failed" ]; then
	  log step_failure
	else
	  log step_success
	fi
}

# verify maximum k8s version
function verify_max_k8s_version() {
	log step "Verifying maximum Kubernetes version."
	RESULT_K8S_MAX_VER="Passed"
	if [[ "${1}" -lt "${kMajorVersion}" ]]; then
		RESULT_K8S_MAX_VER="Failed"
    AddWarning "Kubernetes version, ${kMajorVersion}.${kMinorVersion}, is newer than the latest tested version (${1}.${2})"
	fi
	if [[ "${2}" -lt "${kMinorVersion}" ]]; then
		RESULT_K8S_MAX_VER="Failed"
		AddWarning "Kubernetes version, ${kMajorVersion}.${kMinorVersion}, is newer than the latest tested version (${1}.${2})"
	fi
	if [ $RESULT_K8S_MAX_VER == "Failed" ]; then
	  log step_failure
	else
	  log step_success
	fi
}

# verify that the alpha snap CRDs are not installed
verify_alpha_snap_crds() {
	# check for the alpha snapshot CRDs. These shouldn't be present for installation to proceed with
  CRDS=("VolumeSnapshotClasses" "VolumeSnapshotContents" "VolumeSnapshots")
  for C in "${CRDS[@]}"; do
    log step "Checking for alpha $C CRD"
    # Verify that alpha snapshot related CRDs/CRs are not there on the system.
    kubectl explain ${C} 2> /dev/null | grep "^VERSION.*v1alpha1$" --quiet
    if [ $? -eq 0 ]; then
      AddError "The alpha CRD for ${C} is installed. Please uninstall it"
      RESULT_ALPHA_SNAP_CRDS="Failed"
      log step_failure
      log step "Checking for Custom Resources of alpha $C CRD"
      if [[ $(kubectl get ${C} -A --no-headers 2>/dev/null | wc -l) -ne 0 ]]; then
        AddError "Found Custom Resource for alpha CRD ${C}. Please delete it before continuing with installation"
        RESULT_ALPHA_SNAP_CRDS="Failed"
        log step_failure
      else
        log step_success
      fi
    else
      log step_success
    fi
  done
}

# verify that the alpha snap CRDs are not installed
verify_beta_snap_crds() {
	# check for the beta snapshot CRDs. These shouldn't be present for installation to proceed with
	if [ "$skip_beta_crd_validation" == true ]; then
	  echo "Skipping check to see if beta snapshot CRDs are installed"
	else
    CRDS=("VolumeSnapshotClasses" "VolumeSnapshotContents" "VolumeSnapshots")
    for C in "${CRDS[@]}"; do
      log step "Checking for beta $C CRD"
      # Verify that beta snapshot related CRDs/CRs are not there on the system.
      kubectl explain ${C} 2> /dev/null | grep "^VERSION.*v1beta1$" --quiet
      if [ $? -ne 0 ]; then
        AddError "The beta CRD for ${C} is not installed. Please install it"
        RESULT_BETA_SNAP_CRDS="Failed"
        log step_failure
      else
        log step_success
      fi
    done
  fi
}

function verify_beta_snapshot_controller() {
  log step "Checking if snapshot controller is deployed"
  # check for the snapshot-controller. These are strongly suggested but not required
	kubectl get pods -A | grep snapshot-controller --quiet
	if [ $? -ne 0 ]; then
		AddWarning "The Snapshot Controller does not seem to be deployed"
		RESULT_BETA_SNAP_CONTROLLER="Failed"
		log step_failure
	else
	  log step_success
	fi
}

# error, installation will not continue
function AddError() {
  for N in "$@"; do
    ERRORS+=("${N}")
  done
}

# warning, installation can continue
function AddWarning() {
  for N in "$@"; do
    WARNINGS+=("${N}")
  done
}

# Print a nice summary at the end
function summary() {
	echo
	log separator
	echo "Verification Complete"
	# print all the WARNINGS
	if [ "${#WARNINGS[@]}" -ne 0 ]; then
		echo
		echo "Warnings:"
		for E in "${WARNINGS[@]}"; do
  			echo "- ${E}"
		done
		RC=$EXIT_WARNING
	fi

	# print all the ERRORS
	if [ "${#ERRORS[@]}" -ne 0 ]; then
		echo
		echo "Errors:"
		for E in "${ERRORS[@]}"; do
  			echo "- ${E}"
		done
		RC=$EXIT_ERROR
	fi

	echo
	log separator
	echo "Summary"
	log step "Kubernetes Min version:"
	log ${RESULT_K8S_MIN_VER}
	log step "Kubernetes Max version:"
	log ${RESULT_K8S_MAX_VER}

	if [[ "$kMinorVersion" -gt 16 ]]; then
	  log step "Beta Snapshot CRDs:"
	  if [ "$skip_beta_crd_validation" == true ]; then
      log step_warning
      log warning "Skipped because of user request"
    else
      log ${RESULT_BETA_SNAP_CRDS}
    fi
    log step "Beta Snapshot Controller:"
    log ${RESULT_BETA_SNAP_CONTROLLER}
  fi

	echo
}

#
# main
#
# default values
RESULT_K8S_MIN_VER="Passed"
RESULT_K8S_MAX_VER="Passed"
RESULT_ALPHA_SNAP_CRDS="Passed"
RESULT_BETA_SNAP_CRDS="Passed"
RESULT_BETA_SNAP_CONTROLLER="Passed"

# exit codes
EXIT_SUCCESS=0
EXIT_WARNING=1
EXIT_ERROR=99

# arrays of messages
WARNINGS=()
ERRORS=()

# return code
RC=0

NAMESPACE="default"
# Determine the kubernetes version
source $SCRIPTDIR/common.bash

header
log separator
verify_min_k8s_version "1" "16"
verify_max_k8s_version "1" "19"
log separator

if [[ "$kMinorVersion" -gt 16 ]]; then
  verify_alpha_snap_crds
  verify_beta_snap_crds
  verify_beta_snapshot_controller
fi

summary

if [ ${RESULT_ALPHA_SNAP_CRDS} == "Failed" ]; then
  echo "Please uninstall alpha CRDs for Volume Snapshots before continuing."
fi

if [ ${RESULT_BETA_SNAP_CRDS} == "Failed" ]; then
  echo "Please install CSI VolumeSnapshot Beta CRDs before continuing."
  echo "Run the install script with the option --snapshot-crd to install the beta snapshot CRD"
fi
echo
exit $RC
