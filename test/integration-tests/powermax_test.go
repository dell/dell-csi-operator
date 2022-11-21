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
package integrationtest

import (
	ctx "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	v1 "github.com/dell/dell-csi-operator/api/v1"
	config "github.com/dell/dell-csi-operator/pkg/ctrlconfig"
	"github.com/dell/dell-csi-operator/pkg/resources/rbac"
	"github.com/dell/dell-csi-operator/test/integration-tests/framework"
	util "github.com/dell/dell-csi-operator/test/integration-tests/utils"
	snapv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	V1                 = "v1"
	V1WOPassword       = "v1Password"
	V2                 = "v2"
	V2CustomDrivername = "v2DriverName"
	V2WOCHAPSecret     = "v2Chapsecret"
)

type ContextMap map[string]*DriverContext

var contextMap = ContextMap{
	V1:                 nil,
	V1WOPassword:       nil,
	V2:                 nil,
	V2CustomDrivername: nil,
	V2WOCHAPSecret:     nil,
}

func (context ContextMap) Context(version string) (*DriverContext, error) {
	if _, ok := context[version]; !ok {
		return nil, fmt.Errorf("Invalid configuration version")
	}
	if context[version] == nil {
		err := context.SetContext(version)
		if err != nil {
			return nil, err
		}
	}
	return context[version], nil
}

func (context ContextMap) SetContext(version string) error {
	context[version] = new(DriverContext)
	if strings.Contains(version, V1) {
		context[version].ConfigVersion = V1
	} else if strings.Contains(version, V2) {
		context[version].ConfigVersion = V2
	} else {
		return fmt.Errorf("Failed to determine context version")
	}
	var err error
	if version == V1 {
		err = testBasicDriver("powermax", context[version], new(v1.CSIPowerMax), new(v1.CSIPowerMax), "", "", "", testProp["X_CSI_POWERMAX_USER"], testProp["X_CSI_POWERMAX_PASSWORD"], "", "", false)
	} else if version == V2 {
		err = testBasicDriver("powermax", context[version], new(v1.CSIPowerMax), new(v1.CSIPowerMax), "", "", "", testProp["X_CSI_POWERMAX_USER"], testProp["X_CSI_POWERMAX_PASSWORD"], testProp["X_CSI_POWERMAX_CHAP_SECRET"], testProp["X_CSI_POWERMAX_DRIVER_NAME"], true)
	} else if version == V2CustomDrivername {
		err = testBasicDriver("powermax", context[version], new(v1.CSIPowerMax), new(v1.CSIPowerMax), "", "", "", testProp["X_CSI_POWERMAX_USER"], testProp["X_CSI_POWERMAX_PASSWORD"], testProp["X_CSI_POWERMAX_CHAP_SECRET"], "", true)
	} else if version == V2WOCHAPSecret {
		err = testBasicDriver("powermax", context[version], new(v1.CSIPowerMax), new(v1.CSIPowerMax), "", "", "", testProp["X_CSI_POWERMAX_USER"], testProp["X_CSI_POWERMAX_PASSWORD"], "", testProp["X_CSI_POWERMAX_DRIVER_NAME"], true)
	} else if version == V1WOPassword {
		err = testBasicDriver("powermax", context[version], new(v1.CSIPowerMax), new(v1.CSIPowerMax), "", "", "", testProp["X_CSI_POWERMAX_USER"], "", "", "", false)
	}
	return err
}

func TestBasicPowermax(t *testing.T) {
	_, err := contextMap.Context(V1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("Driver installed successfully")
}

func TestSidecarArgsV1(t *testing.T) {
	testSidecarArgs(t, V1)
}

func TestSidecarArgsV2(t *testing.T) {
	testSidecarArgs(t, V2)
}

func testSidecarArgs(t *testing.T, configVersion string) {
	f := framework.Global
	context, err := contextMap.Context(configVersion)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Checking controller args
	controller, err := f.KubeClient2.AppsV1().StatefulSets(context.FinalSpec.GetNamespace()).Get(ctx.Background(), "powermax-controller", metaV1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to fetch controller daemon-set. (%s)", err.Error())
		t.Error(err.Error())
		return
	}
	ok := checkSidecarArgs(controller.Spec.Template.Spec.Containers, context.DriverConfig.SidecarParams)
	if !ok {
		t.Errorf("Sidecar args, not set properly for the controller.")
		return
	}
	// Checking node args
	node, err := f.KubeClient2.ExtensionsV1beta1().DaemonSets(context.FinalSpec.GetNamespace()).Get(ctx.Background(), "powermax-node", metaV1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to fetch node daemon-set. (%s)", err.Error())
		t.Errorf(err.Error())
		return
	}
	ok = checkSidecarArgs(node.Spec.Template.Spec.Containers, context.DriverConfig.SidecarParams)
	if !ok {
		t.Errorf("Sidecar args, not set properly for the node.")
		return
	}
	fmt.Println("Sidecar args set properly.")
}

func checkSidecarArgs(containers []coreV1.Container, sidecarParams []config.SidecarParams) bool {
	ok := true
	for _, container := range containers {
		for _, param := range sidecarParams {
			if string(param.Name) == container.Name {
				if !reflect.DeepEqual(param.Args, container.Args) {
					ok = false
					log.Debugf("Container args: %+v\n", container.Args)
					log.Debugf("Config args: %+v\n", param.Args)
					log.Errorf("%s args are not properly set", container.Name)
				}
				break
			}
		}
	}
	return ok
}

func TestStorageClass(t *testing.T) {
	context, err := contextMap.Context(V1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	storageClass, err := getStorageClass(context)
	if err != nil {
		log.Errorf("Failed to fetch storage class")
		t.Errorf(err.Error())
	}
	if reflect.DeepEqual(storageClass.Parameters, context.FinalSpec.GetDriver().StorageClass[0].Parameters) {
		fmt.Println("StorageClass properly configured")
	} else {
		log.Errorf("StorageClass params are not set properly. (%s)", err.Error())
		t.Error(err.Error())
	}
}

func getStorageClass(context *DriverContext) (*storageV1.StorageClass, error) {
	storageClassName := context.FinalSpec.GetNamespace() + "-" + context.FinalSpec.GetDriver().StorageClass[0].Name
	return framework.Global.KubeClient2.StorageV1().StorageClasses().Get(ctx.Background(), storageClassName, metaV1.GetOptions{})
}

func TestSnapshotClass(t *testing.T) {
	context, err := contextMap.Context(V1)
	if err != nil {
		log.Errorf("Failed to fetch driver context, (%s)", err.Error())
		t.Error(err.Error())
		return
	}
	snapshotClass, err := getSnapshotClass(context)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Printf("%+v", snapshotClass)
}

func getSnapshotClass(context *DriverContext) (*snapv1.VolumeSnapshotClass, error) {
	snapclassName := context.FinalSpec.GetNamespace() + "-" + context.FinalSpec.GetDriver().SnapshotClass[0].Name
	return framework.Global.SnapshotClientBeta.SnapshotV1().VolumeSnapshotClasses().Get(ctx.Background(), snapclassName, metaV1.GetOptions{})
}

func TestCHAPAuthenticationV2(t *testing.T) {
	f := framework.Global
	context, err := contextMap.Context(V2)
	if err != nil {
		log.Errorf("Failed to fetch driver context, (%s)", err.Error())
		t.Error(err.Error())
		return
	}
	daemonSet, err := f.KubeClient2.ExtensionsV1beta1().DaemonSets(context.InitialSpec.GetNamespace()).Get(ctx.Background(), "powermax-node", metaV1.GetOptions{})
	if err != nil {
		log.Error("Failed to fetch the driver pods.")
		t.Error(err)
		return
	}
	var isCHAPEnabled, isCHAPUsed, isCHAPSecretSet bool
	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		if container.Name == "driver" {
			if chapEnv := getEnvObject(container.Env, envVars["powermaxCHAPEnable"]); chapEnv != nil {
				isCHAPUsed = true
				if chapEnv.Value == "true" {
					isCHAPEnabled = true
					chapSecretEnv := getEnvObject(container.Env, envVars["powremaxCHAPSecret"])
					if chapSecretEnv != nil && chapSecretEnv.ValueFrom.SecretKeyRef.Key == "chapsecret" {
						isCHAPSecretSet = true
					}
				}
			}
		}
		break
	}
	if isCHAPUsed {
		message := fmt.Sprintf("%s set to %t", envVars["powermaxCHAPEnable"], isCHAPEnabled)
		if isCHAPEnabled && isCHAPSecretSet {
			message = fmt.Sprintf("%s and the secret is being read from `chapsecret`", message)
		} else {
			t.Error(fmt.Errorf("CHAP authentication enabled but secret not provided"))
			return
		}
		fmt.Println(message)
	} else {
		t.Error("CHAP env variable not set.")
	}
}

func getEnvObject(envs []coreV1.EnvVar, variableName string) *coreV1.EnvVar {
	var envObject *coreV1.EnvVar
	for _, env := range envs {
		if env.Name == variableName {
			envObject = &env
			break
		}
	}
	return envObject
}

func TestRBACPriveledgesV1(t *testing.T) {
	f := framework.Global
	context, err := contextMap.Context(V1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	clusterRole, err := f.KubeClient2.RbacV1().ClusterRoles().Get(ctx.Background(), context.FinalSpec.GetNamespace()+"-node", metaV1.GetOptions{})
	if err != nil {
		log.Error("Failed to fetch cluster role")
		t.Error(err.Error())
		return
	}
	expectedRole := rbac.NewNodeClusterRole(context.FinalSpec, false, rbac.NewDummyClusterRole("powermax-powermax-dummy"))
	if reflect.DeepEqual(clusterRole.Rules, expectedRole.Rules) {
		fmt.Println("Cluster roles set properly")
	} else {
		t.Error("Roles not set properly")
	}
}

func TestRBACPriveledgesV2(t *testing.T) {
	f := framework.Global
	context, err := contextMap.Context(V2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	clusterRole, err := f.KubeClient2.RbacV1().ClusterRoles().Get(ctx.Background(), context.FinalSpec.GetNamespace()+"-node", metaV1.GetOptions{})
	if err != nil {
		log.Error("Failed to fetch cluster role")
		t.Error(err.Error())
		return
	}
	expectedRole := rbac.NewLimitedClusterRole(context.FinalSpec, false, rbac.NewDummyClusterRole("powermax-powermax-dummy"))
	if reflect.DeepEqual(clusterRole.Rules, expectedRole.Rules) {
		fmt.Println("Cluster roles set properly")
	} else {
		t.Error("Roles not set properly")
	}
}

func TestDefaultDriverNameV2(t *testing.T) {
	f := framework.Global
	context, err := contextMap.Context(V2CustomDrivername)
	if err != nil {
		t.Error(err.Error())
		return
	}
	driverName := getDriverName(f, context)
	if driverName == "" {
		t.Error("Failed to fetch driver name")
		return
	}
	if driverName != "csi-powermax.dellemc.com" {
		t.Error("Default driver name not set properly")
		return
	}
	fmt.Println("Default driver name set properly")
}

func TestDriverNameV2(t *testing.T) {
	f := framework.Global
	context, err := contextMap.Context(V2)
	if err != nil {
		log.Errorf("Failed to fetch driver context, (%s)", err.Error())
		t.Error(err.Error())
		return
	}
	// Get driver name environment variable
	driverName := getDriverName(f, context)
	if driverName == "" {
		t.Error("Failed to fetch driver name")
		return
	}
	if !strings.Contains(driverName, "dellemc.com") && !strings.Contains(driverName, context.FinalSpec.GetNamespace()) {
		log.Error("Driver name is missing namespace prefix and `dellemc.com` suffix.")
		t.Error(fmt.Errorf("driver name is missing prefix and `dellemc.com` suffix"))
		return
	}
	// Check if the driver name matches storageclass provisioner name.
	storageClass, err := getStorageClass(context)
	if err != nil {
		log.Errorf("Failed to fetch storage class. (%s)", err.Error())
		t.Error(err.Error())
	}
	if storageClass.Provisioner == driverName {
		fmt.Println("Driver name set properly in storage class")
	} else {
		t.Errorf("Provisioner name not set properly in storage class. Expected (%s), got (%s)", driverName, storageClass.Provisioner)
	}

	snapshotClass, err := getSnapshotClass(context)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if snapshotClass.Driver == driverName {
		fmt.Println("Driver name set properly in volumesnapshot class")
	} else {
		t.Errorf("Snapshotter name not set properly in volumesnapshot class. Expected (%s), got (%s)", driverName, snapshotClass.Driver)
	}
}

func getDriverName(f *framework.Framework, context *DriverContext) string {
	statefulSet, err := f.KubeClient2.AppsV1().StatefulSets(context.FinalSpec.GetNamespace()).Get(ctx.Background(), "powermax-controller", metaV1.GetOptions{})
	if err != nil {
		log.Errorf("Error, fetching node Stateful set. (%s)", err.Error())
		return ""
	}
	var driverName string
	for _, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == "driver" {
			driverEnvObject := getEnvObject(container.Env, envVars["powermaxDriverName"])
			driverName = driverEnvObject.Value
			break
		}
	}
	return driverName
}

func TestStorageClassModification(t *testing.T) {
	f := framework.Global
	context, err := contextMap.Context(V1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	storageClass, err := getStorageClass(context)
	if err != nil {
		log.Errorf("Failed to fetch storage class. (%s)", err.Error())
		t.Error(err.Error())
		return
	}
	newStorageClass := storageClass.DeepCopy()
	newStorageClass.Provisioner = "powermax"
	_, err = f.KubeClient2.StorageV1().StorageClasses().Update(ctx.Background(), newStorageClass, metaV1.UpdateOptions{})
	if err != nil {
		fmt.Println("Storage class can't be modified.")
	} else {
		t.Error("Storage class modified successfully.")
	}
}

func TestMissingPassword(t *testing.T) {
	_, err := contextMap.Context(V1WOPassword)
	if err == nil {
		t.Error("Driver installation should have failed because password is missing")
		return
	}
	fmt.Println("Cannot install driver without password")
}

func TestMissingCHAPSecret(t *testing.T) {
	_, err := contextMap.Context(V2WOCHAPSecret)
	if err == nil {
		t.Error("Driver installation should have failed because CHAP secret is missing")
		return
	}
	fmt.Println("Cannot install driver without CHAP secret")
}

func TestUserSpecificSecret(t *testing.T) {
	f := framework.Global
	driverType := "powermax"
	customSecretName := "custom-" + driverType + "-creds"
	namespace := driverType + "-" + f.RunID + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	//create namespace
	err := util.CreateNamespace(namespace, f.KubeClient)
	if err != nil {
		log.Errorf("Error creating Namespace. (%s)", err.Error())
		return
	}
	testNamespaces = append(testNamespaces, namespace)
	//assert.False(t, err != nil, "Create namespace failed %v", err)

	//create custom secret
	log.Info("Creating custom secret")
	customSecret := util.NewSecret(customSecretName, namespace, testProp["X_CSI_POWERMAX_USER"], testProp["X_CSI_POWERMAX_PASSWORD"], "")
	_, err = f.KubeClient2.CoreV1().Secrets(namespace).Create(ctx.Background(), customSecret, metaV1.CreateOptions{})
	//assert.False(t, err != nil, "Secret creation failed %v", err)
	if err != nil {
		t.Errorf("Error creating Secret. (%s)", err.Error())
		return
	}
	//read custom resource json file
	initialDriverSpec := new(v1.CSIPowerMax)
	log.Info("Reading CR")
	dat, err := ioutil.ReadFile("cr/" + driverType + ".json")
	if err != nil {
		util.DeleteNamespace(namespace, f.KubeClient)
		//assert.Error(t, err, "Reading CR file failed %v", err)
		t.Errorf("Error Deleting namespace. (%s)", err.Error())
		return

	}
	err = json.Unmarshal(dat, initialDriverSpec)
	//assert.False(t, err != nil, "Parse Driver Json failed. %v", err)
	if err != nil {
		t.Errorf("Error Parse Driver Json failed. (%s)", err.Error())
		return
	}
	initialDriverSpec.SetNamespace(namespace)
	initialDriverSpec.SetName(namespace)
	initialDriverSpec.GetDriver().AuthSecret = customSecret.Name
	initialDriverSpec.GetDriver().TLSCertSecret = customSecret.Name

	dat, err = json.Marshal(initialDriverSpec)
	//assert.False(t, err != nil, "Parse Driver Json failed. %v", err)
	if err != nil {
		t.Errorf("Error Parse Driver Json failed. (%s)", err.Error())
		return
	}
	plural := initialDriverSpec.GetResourcePlural()
	// Add prefix for plural if not
	if strings.Index(plural, "csi") == -1 {
		plural = "csi" + plural
	}
	//Create custom resource
	log.Info("Creating CR")
	_, err = f.RestClient.Post().Namespace(namespace).Resource(plural).Body(dat).DoRaw(ctx.Background())
	//assert.False(t, err != nil, "Create Driver failed %v", err)
	if err != nil {
		t.Errorf("Error Creating Driver failed. (%s)", err.Error())
		return
	}

	log.Info("Waiting resources to create")
	err = util.WaitForDaemonSetAvailable(namespace, driverType+"-node", util.DefaultRetry, f.KubeClient, log)
	if err != nil {
		util.DeleteNamespace(namespace, f.KubeClient)
		t.Errorf("Error in waiting for Daemon Set. (%s)", err.Error())
		return
	}

	log.Info("Waiting pods to create")
	err = util.WaitForPods(namespace, f.KubeClient, driverType, log)
	if err != nil {
		if !strings.Contains(err.Error(), "timeout") {
			util.DeleteNamespace(namespace, f.KubeClient)
			t.Errorf("Error in waiting for Pod. (%s)", err.Error())
			return
		}
	}
	afterDriverSpec := new(v1.CSIPowerMax)
	log.Info("Reading CR again after creating")
	b, err := f.RestClient.Get().Namespace(namespace).Resource(plural).Name(initialDriverSpec.GetName()).DoRaw(ctx.Background())
	err = json.Unmarshal(b, afterDriverSpec)
	//assert.False(t, err != nil, "Parse Driver Json failed. %v", err)
	if err != nil {
		t.Errorf("Error Parse Driver Json failed. (%s)", err.Error())
		return
	}

	statefulSet, err := f.KubeClient2.AppsV1().StatefulSets(namespace).Get(ctx.Background(), "powermax-controller", metaV1.GetOptions{})
	if err != nil {
		t.Errorf("Error, fetching node stateful set. (%s)", err.Error())
		return
	}
	// Performing Check for Custom Auth Secret
	log.Info("Performing Check for Custom Auth Secret")
	for _, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == "driver" {
			driverEnvObject := getEnvObject(container.Env, "X_CSI_POWERMAX_USER")
			secretKeyRef := driverEnvObject.ValueFrom.SecretKeyRef
			driverSecretName := secretKeyRef.LocalObjectReference.Name
			if driverSecretName != customSecret.Name {
				t.Errorf("Using wrong auth secret: %s", afterDriverSpec.GetDriver().AuthSecret)
				return
			}
			break
		}
	}
	log.Info("Custom Auth Secret being used successfully")

	// Performing Check for Custom TLS cert Secret
	log.Info("Performing Check for Custom TLS cert Secret")
	for _, volume := range statefulSet.Spec.Template.Spec.Volumes {
		if volume.Name == "certs" {
			if volume.Secret.SecretName != customSecret.Name {
				t.Errorf("Using wrong TLS cert secret: %s", afterDriverSpec.GetDriver().AuthSecret)
				return
			}
			break
		}
	}
	log.Info("Custom Auth Secret being used successfully")
	return
}
