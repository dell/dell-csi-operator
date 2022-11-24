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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	v1 "github.com/dell/dell-csi-operator/api/v1"
	config "github.com/dell/dell-csi-operator/pkg/ctrlconfig"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/dell/dell-csi-operator/test/integration-tests/framework"
	util "github.com/dell/dell-csi-operator/test/integration-tests/utils"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var envVars = map[string]string{
	"powermaxCHAPEnable": "X_CSI_POWERMAX_ISCSI_ENABLE_CHAP",
	"powremaxCHAPSecret": "X_CSI_POWERMAX_ISCSI_CHAP_PASSWORD",
	"powermaxDriverName": "X_CSI_POWERMAX_DRIVER_NAME",
}

type DriverContext struct {
	InitialSpec   v1.CSIDriver
	FinalSpec     v1.CSIDriver
	DriverConfig  *config.DriverConfig
	ConfigVersion string
}

func TestBasic1Unity(t *testing.T) {
	initialUnitySpec := new(v1.CSIUnity)
	afterUnitySpec := new(v1.CSIUnity)
	driverContext := &DriverContext{
		ConfigVersion: "v1",
	}
	err := testBasicDriver("unity", driverContext, initialUnitySpec, afterUnitySpec, testProp["UNITY_UNISPHERE_URL"], "", "", testProp["UNITY_UNISPHERE_USER"], testProp["UNITY_UNISPHERE_PASSWORD"], "", "", false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestBasic1Isilon(t *testing.T) {
	initialIsilonSpec := new(v1.CSIIsilon)
	afterIsilonSpec := new(v1.CSIIsilon)
	driverContext := &DriverContext{
		ConfigVersion: "v3",
	}
	err := testBasicDriver("isilon", driverContext, initialIsilonSpec, afterIsilonSpec, "", "", "", testProp["X_CSI_ISI_USER"], testProp["X_CSI_ISI_PASSWORD"], "", "", false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestBasic1Vxflexos(t *testing.T) {
	initialVxflexosSpec := new(v1.CSIVXFlexOS)
	afterVxflexosSpec := new(v1.CSIVXFlexOS)
	driverContext := &DriverContext{
		ConfigVersion: "v3",
	}
	err := testBasicDriver("vxflexos", driverContext, initialVxflexosSpec, afterVxflexosSpec, "", "", "", testProp["X_CSI_VXFLEXOS_USER"], testProp["X_CSI_VXFLEXOS_PASSWORD"], "", "", false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestBasic1PowerStore(t *testing.T) {
	initialPowerstoreSpec := new(v1.CSIPowerStore)
	afterPowerstoreSpec := new(v1.CSIPowerStore)
	driverContext := &DriverContext{
		ConfigVersion: "v2",
	}
	err := testBasicDriver("powerstore", driverContext, initialPowerstoreSpec, afterPowerstoreSpec, "", "", "", testProp["X_CSI_POWERSTORE_USER"], testProp["X_CSI_POWERSTORE_PASSWORD"], "", "", false)
	if err != nil {
		t.Error(err.Error())
	}
}

func testBasicDriver(driverType string, driverContext *DriverContext, initialDriverSpec v1.CSIDriver, afterDriverSpec v1.CSIDriver, unisphereURL, isiEndpoint, isiPort, username, password, chapsecret, customDriverName string, isCHAPEnabled bool) error {
	f := framework.Global

	secretName := driverType + "-creds"
	namespace := driverType + "-" + f.RunID + "-" + strconv.FormatInt(time.Now().Unix(), 10)

	//create namespace
	err := util.CreateNamespace(namespace, f.KubeClient)
	if err != nil {
		return err
	}
	testNamespaces = append(testNamespaces, namespace)

	//create secret
	log.Info("Creating secret")
	secret := util.NewSecret(secretName, namespace, username, password, chapsecret)
	_, err = f.KubeClient2.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	//read custom resource json file
	log.Info("Reading CR")
	dat, err := ioutil.ReadFile("cr/" + driverType + ".json")
	if err != nil {
		util.DeleteNamespace(namespace, f.KubeClient)
		return err
	}
	err = json.Unmarshal(dat, initialDriverSpec)
	if err != nil {
		return err
	}

	initialDriverSpec.SetNamespace(namespace)
	initialDriverSpec.SetName(namespace)
	initialDriverSpec.GetDriver().ConfigVersion = driverContext.ConfigVersion

	if driverContext.ConfigVersion == "v2" {
		if isCHAPEnabled {
			initialDriverSpec.GetDriver().Common.Envs = append(initialDriverSpec.GetDriver().Common.Envs, coreV1.EnvVar{
				Name:  envVars[driverType+"CHAPEnable"],
				Value: "true",
			})
		}
		if customDriverName != "" {
			initialDriverSpec.GetDriver().Common.Envs = append(initialDriverSpec.GetDriver().Common.Envs, coreV1.EnvVar{
				Name:  envVars[driverType+"DriverName"],
				Value: customDriverName,
			})
		}
	}

	dat, err = json.Marshal(initialDriverSpec)
	if err != nil {
		return err
	}
	plural := initialDriverSpec.GetResourcePlural()
	// Add prefix for plural if not
	if strings.Index(plural, "csi") == -1 {
		plural = "csi" + plural
	}
	//Create custom resource
	log.Info("Creating CR")
	_, err = f.RestClient.Post().Namespace(namespace).Resource(plural).Body(dat).DoRaw(context.Background())
	if err != nil {
		return err
	}

	log.Info("Waiting resources to create")
	err = util.WaitForDaemonSetAvailable(namespace, driverType+"-node", util.DefaultRetry, f.KubeClient, log)
	if err != nil {
		util.DeleteNamespace(namespace, f.KubeClient)
		return err
	}

	log.Info("Waiting pods to create")
	err = util.WaitForPods(namespace, f.KubeClient, driverType, log)
	if err != nil {
		if !strings.Contains(err.Error(), "timeout") {
			util.DeleteNamespace(namespace, f.KubeClient)
			return err
		}
	}

	log.Info("Reading CR again after creating")
	b, err := f.RestClient.Get().Namespace(namespace).Resource(plural).Name(initialDriverSpec.GetName()).DoRaw(context.Background())
	err = json.Unmarshal(b, afterDriverSpec)
	if err != nil {
		return err
	}

	versionInfo, err := resources.GetKubeAPIServerVersion()
	if err != nil {
		return err
	}
	version := strings.Replace(fmt.Sprintf("v%s%s", versionInfo.Major, versionInfo.Minor), "+", "", 1)
	configFile := fmt.Sprintf("%s_%s_%s.json", driverType, initialDriverSpec.GetDriver().ConfigVersion, version)
	driverConfig, err := getDriverConfig("../../driverconfig", configFile)
	if err != nil {
		return err
	}

	driverContext.InitialSpec = initialDriverSpec
	driverContext.FinalSpec = afterDriverSpec
	driverContext.DriverConfig = driverConfig
	return nil
}

func TestBasic2Unity(t *testing.T) {
	fmt.Println("----Test2------")
}
