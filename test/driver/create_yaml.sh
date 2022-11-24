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

# This is sample script to show how to create driver custom resource manifests
# First set the environment variables which will be passed by the operatorutils binary to
# generate the yaml files
# The operatorutils code can be modified to read files or read entire directory
# This is just a sample implementation to show how to generate driver manifests easily
#source ./sample_driver_config
source ./powermax_driver_config
#go build ./operatorutils
go run ./operatorutils.go
