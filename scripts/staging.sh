#  Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
 
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#       http://www.apache.org/licenses/LICENSE-2.0
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

#!/bin/bash

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOTDIR="$(dirname "$SCRIPTDIR")"

RANDOM=$$
generated_dir_name="Release/temp$RANDOM"

if [ -z ${STAGING_DIR+x} ]; then
        STAGING_DIRECTORY=$generated_dir_name
else
        STAGING_DIRECTORY=${STAGING_DIR}
        if [ "$STAGING_DIRECTORY" == "" ]; then
                STAGING_DIRECTORY=$generated_dir_name
        fi
fi

echo "Staging directory is set to: $STAGING_DIRECTORY"
echo "Creating directory"
echo mkdir -p $STAGING_DIRECTORY
mkdir -p $STAGING_DIRECTORY

echo "** Copying scripts folder **"
echo mkdir -p $STAGING_DIRECTORY/scripts
mkdir -p $STAGING_DIRECTORY/scripts
echo cp -f "${ROOTDIR}/scripts/install.sh" "${STAGING_DIRECTORY}/scripts/"
echo cp -f "${ROOTDIR}/scripts/uninstall.sh" "${STAGING_DIRECTORY}/scripts/"
echo cp -f "${ROOTDIR}/scripts/common.bash" "${STAGING_DIRECTORY}/scripts/"
echo cp -f "${ROOTDIR}/scripts/verify.sh" "${STAGING_DIRECTORY}/scripts/"
echo cp -f "${ROOTDIR}/scripts/delete_crds.sh" "${STAGING_DIRECTORY}/scripts/"
echo cp -f "${ROOTDIR}/scripts/csi-offline-bundle.sh" "${STAGING_DIRECTORY}/scripts/"
echo cp -f "${ROOTDIR}/scripts/csi-offline-bundle.md" "${STAGING_DIRECTORY}/scripts/"

cp -f "${ROOTDIR}/scripts/install.sh" "${STAGING_DIRECTORY}/scripts/"
cp -f "${ROOTDIR}/scripts/uninstall.sh" "${STAGING_DIRECTORY}/scripts/"
cp -f "${ROOTDIR}/scripts/common.bash" "${STAGING_DIRECTORY}/scripts/"
cp -f "${ROOTDIR}/scripts/verify.sh" "${STAGING_DIRECTORY}/scripts/"
cp -f "${ROOTDIR}/scripts/delete_crds.sh" "${STAGING_DIRECTORY}/scripts/"
cp -f "${ROOTDIR}/scripts/csi-offline-bundle.sh" "${STAGING_DIRECTORY}/scripts/"
cp -f "${ROOTDIR}/scripts/csi-offline-bundle.md" "${STAGING_DIRECTORY}/scripts/"

echo
echo "** Copying deploy folder **"
echo cp -r "${ROOTDIR}/deploy" "${STAGING_DIRECTORY}/"
cp -r "${ROOTDIR}/deploy" "${STAGING_DIRECTORY}/"

echo
echo "** Copying driverconfig folder **"
echo cp -r "${ROOTDIR}/driverconfig" "${STAGING_DIRECTORY}/"
cp -r "${ROOTDIR}/driverconfig" "${STAGING_DIRECTORY}/"

echo
echo "** Copying samples folder **"
echo cp -r "${ROOTDIR}/samples" "${STAGING_DIRECTORY}/"
cp -r "${ROOTDIR}/samples" "${STAGING_DIRECTORY}/"

echo
echo "** Copying LICENSE file **"
echo cp "${ROOTDIR}/licenses/LICENSE" "${STAGING_DIRECTORY}/"
cp "${ROOTDIR}/licenses/LICENSE" "${STAGING_DIRECTORY}/"

echo
echo "** Copying Readme from documentation folder **"
echo cp "${ROOTDIR}/documentation/Readme.md" "${STAGING_DIRECTORY}/README.md"
cp "${ROOTDIR}/documentation/Readme.md" "${STAGING_DIRECTORY}/README.md"
