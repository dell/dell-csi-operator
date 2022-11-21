/*
 Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
 
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/
#!/bin/bash
set -e
if [ $# -eq 0 ]; then
    echo "You must specify an Operator version"
    exit 1
fi
operator_version=$1
printf "\n*****\n"
echo "Creating temporary directory for storing the sample yaml files"
echo
echo mkdir -p config/temp_sample_manifest_dir
mkdir -p config/temp_sample_manifest_dir
printf "\n*****\n"
echo "Moving the OpenShift sample yaml files to the temp directory"
echo
echo mv -f config/samples/*ops*.yaml config/temp_sample_manifest_dir
mv -f config/samples/*ops*.yaml config/temp_sample_manifest_dir
printf "\n*****\n"
echo "Current set of files in the config/samples folder"
echo
find config/samples -type f -printf "%f\n"
printf "\n*****\n"
echo "Generating CSV file"
echo
echo ./operator-sdk generate csv --update-crds --csv-version "$operator_version" --default-channel --csv-channel stable --operator-name dell-csi-operator
./operator-sdk generate csv --update-crds --csv-version "$operator_version" --default-channel --csv-channel stable --operator-name dell-csi-operator
printf "\n*****\n"
echo "Moving the sample files back to original directory"
echo
echo mv -f config/temp_sample_manifest_dir/*.yaml config/samples
mv -f config/temp_sample_manifest_dir/*.yaml config/samples
printf "\n*****\n"
echo "Deleting the temporary directory"
echo
echo rm -rf deploy/temp_sample_manifest_dir
rm -rf deploy/temp_sample_manifest_dir
printf "\n*****\n"
echo "**** git status after the update ****"
git status -s

