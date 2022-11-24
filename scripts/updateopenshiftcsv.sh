#!/bin/bash
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

set -e
if [ $# -eq 0 ]; then
    echo "You must specify an Operator version"
    exit 1
fi
operator_version=$1
CSVFileName="dell-csi-operator.clusterserviceversion.yaml"
OpenShiftCSVDir=config/OpenShiftCSV
printf "\n*****\n"
echo "Creating temporary directory for storing the sample yaml files"
echo
echo mkdir -p config/temp_sample_manifest_dir
mkdir -p config/temp_sample_manifest_dir
printf "\n*****\n"
echo "Moving the k8s sample yaml files to the temp directory"
echo
echo mv -f config/samples/*k8s*.yaml config/temp_sample_manifest_dir
mv -f config/samples/*k8s*.yaml config/temp_sample_manifest_dir
printf "\n*****\n"
echo "Current set of files in the deploy/crds folder"
echo
find config/samples -type f -printf "%f\n"
printf "\n*****\n"
printf "\n*****\n"
echo
echo "**** Copying the OpenShift CSV file into to the manifests folder temporarily ****"
echo mkdir -p config/temp/
mkdir -p config/temp
echo cp config/olm-catalog/dell-csi-operator/manifests/"$CSVFileName" config/temp/
cp config/olm-catalog/dell-csi-operator/manifests/"$CSVFileName" config/temp/
echo cp -f "$OpenShiftCSVDir"/"$CSVFileName" config/olm-catalog/dell-csi-operator/manifests
cp -f "$OpenShiftCSVDir"/"$CSVFileName" deploy/olm-catalog/dell-csi-operator/manifests
printf "\n*****\n"
echo "Generating CSV file"
echo
echo ./operator-sdk generate csv --update-crds --csv-version "$operator_version" --default-channel --csv-channel stable --operator-name dell-csi-operator
./operator-sdk generate csv --update-crds --csv-version "$operator_version" --default-channel --csv-channel stable --operator-name dell-csi-operator
printf "\n*****\n"
echo "Moving the new CSV file back to OpenShift folder"
echo
echo cp -f config/olm-catalog/dell-csi-operator/manifests/dell-csi-operator.clusterserviceversion.yaml config/OpenShiftCSV
cp -f config/olm-catalog/dell-csi-operator/manifests/dell-csi-operator.clusterserviceversion.yaml config/OpenShiftCSV
echo
printf "\n*****\n"
echo "**** Restore the original CSV File"
echo mv -f config/temp/"$CSVFileName" config/olm-catalog/dell-csi-operator/manifests/
mv -f config/temp/"$CSVFileName" config/olm-catalog/dell-csi-operator/manifests/
echo rm -rf config/temp
rm -rf config/temp
printf "\n*****\n"
echo "Moving the sample files back to original directory"
echo
echo mv -f config/temp_sample_manifest_dir/*.yaml config/samples
mv -f config/temp_sample_manifest_dir/*.yaml config/samples
printf "\n*****\n"
echo "Deleting the temporary directory"
echo
echo rm -rf config/temp_sample_manifest_dir
rm -rf config/temp_sample_manifest_dir
printf "\n*****\n"
echo "**** git status after the update ****"
git status -s
