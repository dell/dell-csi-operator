#!/bin/bash
# This is sample script to show how to create driver custom resource manifests
# First set the environment variables which will be passed by the operatorutils binary to
# generate the yaml files
# The operatorutils code can be modified to read files or read entire directory
# This is just a sample implementation to show how to generate driver manifests easily
#source ./sample_driver_config
source ./powermax_driver_config
#go build ./operatorutils
go run ./operatorutils.go
