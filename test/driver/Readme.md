# Readme
Files present
1. install_csi_driver.sh -> main script which calls other scripts
2. testlib.sh -> contains certain library functions
3. operatorutils.go -> contains golang function for creating driver related yaml file.
4. sample_driver_config -> Sample File which contains key value pairs for creating driver yaml files


## How to Execute?
1. Take a copy of sample_driver_config file and rename (according to your csi-driver)
2. Furnish the newly created driver config file with all the required values (including driver build number and driver yaml file name)
3. Manually furnish the pre-requisites (like namespaes, secrets etc.)
4. Execute the following command with operator build number (in the format of dell-csi-operator:v1.2.0) to create csi-operator and then csi drivers (controller and node)
    1. sh install_csi_driver.sh "<operator build name>"
5. Wait for the script to finish and check the existence of csi-operator and controller/node by normal kubectl commands.
6. This script will pull the specified operator build
5.As of unity/powermax driver has been tested. Other drivers are yet to be tested.
