#!/bin/bash

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

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

# verify that the snap CRDs are installed
function verify_snap_crds() {
	# check for the snapshot CRDs.
    CRDS=("VolumeSnapshotClasses" "VolumeSnapshotContents" "VolumeSnapshots")
    for C in "${CRDS[@]}"; do
      log step "Checking $C CRD"
      # Verify that snapshot related CRDs/CRs exist on the system.
      kubectl explain ${C} > /dev/null 2>&1
      if [ $? -ne 0 ]; then
        AddError "The CRD for ${C} is not Found. These need to be installed by the Kubernetes administrator"
        RESULT_SNAP_CRDS="Failed"
        log step_failure
      else
        log step_success
      fi
    done
}

function verify_snapshot_controller() {
  log step "Checking if snapshot controller is deployed"
  # check for the snapshot-controller. These are strongly suggested but not required
	kubectl get pods -A | grep snapshot-controller --quiet
	if [ $? -ne 0 ]; then
		AddWarning "The Snapshot Controller was not found on the system. These need to be installed by the Kubernetes administrator."
		RESULT_SNAP_CONTROLLER="Failed"
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

    log step "Snapshot CRDs:"
    log ${RESULT_SNAP_CRDS}
    log step "Snapshot Controller:"
    log ${RESULT_SNAP_CONTROLLER}

	echo
}

#
# main
#
# default values
RESULT_K8S_MIN_VER="Passed"
RESULT_K8S_MAX_VER="Passed"
RESULT_SNAP_CRDS="Passed"
RESULT_SNAP_CONTROLLER="Passed"

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
verify_min_k8s_version "1" "19"
verify_max_k8s_version "1" "24"
verify_snap_crds
verify_snapshot_controller
log separator

summary

if [ ${RESULT_SNAP_CRDS} == "Failed" ]; then
  echo "Some of the CRDs are not found on the system. These need to be installed by the Kubernetes administrator."
fi
echo
exit $RC
