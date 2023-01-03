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

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOTDIR="$(dirname "$SCRIPTDIR")"

# This should be updated with every release
OPERATOR_VERSION="1.11.0"
DEFAULT_REPO=dellemc/dell-csi-operator
SOURCE_INDEX_IMG=dellemc/dell-csi-operator/dellemcregistry_community:v1.10.0

# Using docker for building the images as there were some issues with podman
command -v docker
if [ $? -eq 0 ]; then
  echo "Using docker for building image"
  BUILDER="docker"
else
  echo "Couldn't find docker for building the image"
  exit 1
fi

if [ -z ${REGISTRY+x} ]; then
        echo "REGISTRY is unset"
        REPO=$DEFAULT_REPO
else
        REPO=${REGISTRY}
        if [ "$REPO" == "" ]; then
                echo "REGISTRY is set to empty. Defaulting to $DEFAULT_REPO"
                REPO=$DEFAULT_REPO
        fi
fi

set -e

VERSION=v"$OPERATOR_VERSION"."${BUILD_NUMBER}"
BUNDLE_IMG_NAME="csiopbundle"
INDEX_IMG_NAME="dellemcregistry_community"

BUNDLE_IMG=${REPO}/${BUNDLE_IMG_NAME}:${VERSION}
INDEX_IMG=${REPO}/${INDEX_IMG_NAME}:${VERSION}
INDEX_REL_IMG=${REPO}/${INDEX_IMG_NAME}:v"$OPERATOR_VERSION"

echo "*****"
echo
echo "Bundle images will be pushed to: $REPO"
echo "Bundle Image Tag: $BUNDLE_IMG"
echo "Index Image Tag: $INDEX_IMG"
echo "Source Index Image Tag: $SOURCE_INDEX_IMG"
echo

echo "Updating the community bundle CSV file"
rm -f "$ROOTDIR"/community_bundle/manifests/dell-csi-operator.clusterserviceversion.yaml
cp "$ROOTDIR"/bundle/manifests/dell-csi-operator-certified.clusterserviceversion.yaml "$ROOTDIR"/community_bundle/manifests/dell-csi-operator.clusterserviceversion.yaml
sed -i s/dell-csi-operator-certified/dell-csi-operator/g "$ROOTDIR"/community_bundle/manifests/dell-csi-operator.clusterserviceversion.yaml
sed -i /certified:/d "$ROOTDIR"/community_bundle/manifests/dell-csi-operator.clusterserviceversion.yaml

echo "Copying files into the community bundle manifests folder"
mkdir -p temp/community_bundle_manifests
cp -r "$ROOTDIR"/community_bundle/manifests/ temp/community_bundle_manifests/
cp "$ROOTDIR"/bundle/manifests/*.yaml "$ROOTDIR"/community_bundle/manifests/
rm -f "$ROOTDIR"/community_bundle/manifests/dell-csi-operator-certified.clusterserviceversion.yaml

echo "**** Building bundle image for version: $OPERATOR_VERSION"
echo $BUILDER build -f community.bundle.Dockerfile -t "$BUNDLE_IMG" .
$BUILDER build -f community.bundle.Dockerfile -t "$BUNDLE_IMG" .
echo

echo "In order to build the index image, $BUNDLE_IMG must exist on the remote repository"
echo "If $BUNDLE_IMG doesn't exist on the remote repo, then the next step to update the index image will fail"
echo
if [ "${1}" == "autopush" ]; then
RESP=y
else
read -n 1 -p "Do you wish to push $BUNDLE_IMG (pres Y/y to continue): " RESP
fi

if [ "${RESP}" == "y" ] || [ "${RESP}" == "Y" ]; then
  echo $BUILDER push "$BUNDLE_IMG"
  $BUILDER push "$BUNDLE_IMG"
fi
echo

echo "Cleaning up the copied files"
echo rm -f "$ROOTDIR"/community_bundle/manifests/*
rm -f "$ROOTDIR"/community_bundle/manifests/*
echo cp temp/community_bundle_manifests/manifests/* "$ROOTDIR"/community_bundle/manifests/
cp temp/community_bundle_manifests/manifests/* "$ROOTDIR"/community_bundle/manifests/
echo rm -rf temp/community_bundle_manifests
rm -rf temp/community_bundle_manifests
echo

echo "**** Updating index with the bundle image: $BUNDLE_IMG"
echo "opm index add --bundles $BUNDLE_IMG --from-index $SOURCE_INDEX_IMG --tag $INDEX_IMG --container-tool $BUILDER"
opm index add --bundles "$BUNDLE_IMG" --from-index "$SOURCE_INDEX_IMG" --tag "$INDEX_IMG" --container-tool $BUILDER
$BUILDER tag "$INDEX_IMG" "$INDEX_REL_IMG"
echo
echo

if [ "${1}" == "autopush" ]; then
RESP=y
else
read -n 1 -p "Do you wish to push $INDEX_IMG (pres Y/y to continue): " RESP
fi

if [ "${RESP}" == "y" ] || [ "${RESP}" == "Y" ]; then
  echo
  echo $BUILDER push "$INDEX_IMG"
  $BUILDER push "$INDEX_IMG"
  echo $BUILDER push "$INDEX_REL_IMG"
  $BUILDER push "$INDEX_REL_IMG"
fi
echo
echo "**** End of script ****"

