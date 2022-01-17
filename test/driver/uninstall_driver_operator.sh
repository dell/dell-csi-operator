#!/bin/sh
# Shell script to deploy operators and to create CSI Drivers

source ./sample_driver_config

#Uninstall teh driver 

kubectl  delete -f  $OPERATOR_ENV_DRIVER_MANIFEST
 
# Uninstall the operator 
cd ../../ && sh scripts/uninstall.sh > ./test/driver/deploy_output.log



