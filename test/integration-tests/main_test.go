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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	config "github.com/dell/dell-csi-operator/pkg/ctrlconfig"
	"github.com/dell/dell-csi-operator/test/integration-tests/framework"
	util "github.com/dell/dell-csi-operator/test/integration-tests/utils"
	"github.com/sirupsen/logrus"
)

var (
	nsCreated      = false
	testNamespaces []string
)

var log = GetLogger()
var testProp map[string]string

func TestMain(m *testing.M) {
	// Global setup
	if err := framework.Setup(); err != nil {
		log.Errorf("Failed to setup framework: %v", err)
		os.Exit(1)
	}

	// for this tutorial, we will hard code it to config.txt
	var err error
	testProp, err = readTestProperties("test.properties")
	if err != nil {
		panic("The system cannot find the file specified")
	}

	// Create the namespace if it does not exist as part of global setup (and delete it if we created it in teardown)
	if !util.NamespaceExists(framework.Global.Namespace, framework.Global.KubeClient) {
		if err := util.CreateNamespace(framework.Global.Namespace, framework.Global.KubeClient); err != nil {
			log.Errorf("Failed to create namespace %s for test: %v", framework.Global.Namespace, err)
			os.Exit(1)
		} else {
			nsCreated = true
		}
	}

	code := m.Run()

	if nsCreated && !framework.Global.SkipTeardown {
		if err := util.DeleteNamespace(framework.Global.Namespace, framework.Global.KubeClient); err != nil {
			log.Errorf("Failed to clean up integration test namespace: %v", err)
		}
	}

	for _, namespace := range testNamespaces {
		if err := util.DeleteNamespace(namespace, framework.Global.KubeClient); err != nil {
			log.Errorf("Failed to clean up integration test namespace: %v", err)
		}
	}

	// Global tear-down
	if err := framework.Teardown(); err != nil {
		log.Errorf("Failed to teardown framework: %v", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func readTestProperties(filename string) (map[string]string, error) {
	// init with some bogus data
	configPropertiesMap := map[string]string{}
	if len(filename) == 0 {
		return nil, errors.New("Error reading properties file " + filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')

		// check if the line has = sign
		// and process the line. Ignore the rest.
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				// assign the config map
				configPropertiesMap[key] = value
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return configPropertiesMap, nil
}

var singletonLog *logrus.Logger
var once sync.Once

func GetLogger() *logrus.Logger {
	once.Do(func() {
		singletonLog = logrus.New()
		fmt.Println("csi-unity logger initiated. This should be called only once.")
		singletonLog.Level = logrus.DebugLevel
		singletonLog.SetReportCaller(true)
		singletonLog.Formatter = &logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename1 := strings.Split(f.File, "dell/dell-csi-operator")
				if len(filename1) > 1 {
					return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("dell/dell-csi-operator%s:%d", filename1[1], f.Line)
				}
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", f.File, f.Line)
			},
		}
	})

	return singletonLog
}

func getDriverConfig(configDirectory, driverVersion string) (*config.DriverConfig, error) {
	jsonFileName := filepath.Join(configDirectory, driverVersion)
	jsonFile, err := os.Open(jsonFileName)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	jsonBytes, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		return nil, err
	}
	driverConfigMap := new(config.DriverConfigMap)
	err = json.Unmarshal(jsonBytes, driverConfigMap)
	if err != nil {
		return nil, err
	}
	return &driverConfigMap.DriverConfig, nil
}
