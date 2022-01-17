/*


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

package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	osruntime "runtime"
	"strconv"
	"strings"
	"time"

	operatorconfig "github.com/dell/dell-csi-operator/pkg/config"

	"github.com/dell/dell-csi-operator/core"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/dell/dell-csi-operator/pkg/resources"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kubernetes-csi/external-snapshotter/client/v3/apis/volumesnapshot/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	storagev1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// BaseKubernetesVersion - Kubernetes Version which the operator defaults to
const BaseKubernetesVersion = "v118"
const DefaultConfigFile = "config.yaml"

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(storagev1.AddToScheme(scheme))

	utilruntime.Must(v1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info("Operator Version", "Version", core.SemVer, "Commit ID", core.CommitSha32, "Commit SHA", string(core.CommitTime.Format(time.RFC1123)))
	log.Info(fmt.Sprintf("Go Version: %s", osruntime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", osruntime.GOOS, osruntime.GOARCH))
	//log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func kubeAPIServerVersion(configDirectory, configFile string) (storagev1.K8sVersion, error) {
	kubeAPIServerVersion, err := resources.GetKubeAPIServerVersion()
	if err != nil {
		// Default to the base k8s version
		log.Info(fmt.Sprintf("Failed to get KubeAPI Server versions. Defaulting to %s", BaseKubernetesVersion))
		return storagev1.BaseK8sVersion, nil

	}
	majorVersion := kubeAPIServerVersion.Major
	minorVersion := strings.TrimSuffix(kubeAPIServerVersion.Minor, "+")
	kubeVersion := fmt.Sprintf("v%s%s", majorVersion, minorVersion)
	log.Info(fmt.Sprintf("Kubernetes Version: %s", kubeVersion))
	supportedVersionFilePath := filepath.Join(configDirectory, configFile)
	file, err := os.Open(supportedVersionFilePath)
	if err != nil {
		log.Info(fmt.Sprintf("Failed to find the config file %s", configFile))
		return "", fmt.Errorf("missing config file")
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	configData := make(map[string]interface{})
	err = yaml.Unmarshal(byteValue, &configData)
	if err != nil {
		return "", err
	}
	if versions, ok := configData["supportedK8sVersions"]; ok {
		if reflect.TypeOf(versions).Kind() == reflect.Slice {
			supportedVersions := reflect.ValueOf(versions)
			found := false
			for i := 0; i < supportedVersions.Len(); i++ {
				if supportedVersions.Index(i).Elem().Interface().(string) == kubeVersion {
					found = true
					break
				}
			}
			if !found {
				minorVersionNumber, _ := strconv.Atoi(minorVersion)
				tempKubeVersion := fmt.Sprintf("v%s%d", majorVersion, minorVersionNumber-1)
				for i := 0; i < supportedVersions.Len(); i++ {
					if supportedVersions.Index(i).Elem().Interface().(string) == tempKubeVersion {
						log.Info(fmt.Sprintf("%s is not supported by dell-csi-operator. Falling back to config files for %s",
							kubeVersion, tempKubeVersion))
						kubeVersion = tempKubeVersion
						found = true
						break
					}
				}
				if !found {
					return "", fmt.Errorf("unsupported k8s version. can't run operator")
				}
			}
		}
	} else {
		return "", fmt.Errorf("list of supported K8s versions missing from config file")
	}
	return storagev1.K8sVersion(kubeVersion), nil
}

func getOperatorConfig() operatorconfig.Config {
	cfg := operatorconfig.Config{}
	// Set the default retry count from constant
	cfg.RetryCount = constants.RetryCount
	// Get the environment variable config file
	configFile := os.Getenv("X_CSI_OPERATOR_CONFIG_FILE")
	if configFile == "" {
		configFile = DefaultConfigFile
	}
	cfg.ConfigFile = configFile
	// Get the environment variable config dir
	configDir := os.Getenv("X_CSI_OPERATOR_CONFIG_DIR")
	if configDir == "" {
		// Set the config dir to the folder pkg/config
		configDir = "driverconfig/"
	} else {
		// Check if the directory is empty
		fileName := configDir + "/" + configFile
		_, err := os.Open(fileName)
		if err != nil {
			// This means that the configmap is not mounted
			// fall back to the local copy
			log.Error(err, "Error reading file from the configmap mount")
			log.Info("Falling back to local copy of config files")
			configDir = "/etc/config/local/dell-csi-operator"
		}
	}
	cfg.ConfigDirectory = configDir
	kubeVersion, err := kubeAPIServerVersion(cfg.ConfigDirectory, cfg.ConfigFile)
	if err != nil {
		panic(err.Error())
	}
	cfg.KubeAPIServerVersion = kubeVersion
	isOpenShift, err := resources.IsOpenshift()
	if err != nil {
		log.Error(err, "Failed to determine if it is an Openshift cluster. Assuming it is")
		isOpenShift = true
	}
	cfg.IsOpenShift = isOpenShift
	if cfg.IsOpenShift {
		log.Info("Detected OpenShift API groups")
	}
	enabledDrivers := make([]storagev1.DriverType, 0)
	enabledDriverEnvValue := os.Getenv("OPERATOR_DRIVERS")
	if enabledDriverEnvValue == "" {
		// if the environment variable is unset or set to empty, enabled all drivers
		enabledDrivers = append(enabledDrivers, storagev1.PowerMax,
			storagev1.Unity, storagev1.Isilon, storagev1.VXFlexOS, storagev1.PowerStore)
	} else {
		tempEnabledDrivers := strings.Split(enabledDriverEnvValue, ",")
		for _, driver := range tempEnabledDrivers {
			driverType := operatorconfig.GetDriverType(driver)
			if driverType == storagev1.Unknown {
				fmt.Println("Unknown driver type specified. Ignoring...")
				continue
			}
			enabledDrivers = append(enabledDrivers, driverType)
		}
	}
	cfg.EnabledDrivers = enabledDrivers
	return cfg
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":9999", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	printVersion()
	operatorConfig := getOperatorConfig()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "7e980ba4.dell.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.CSIPowerMaxReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CSIPowerMax"),
		Scheme: mgr.GetScheme(),
		Config: operatorConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIPowerMax")
		os.Exit(1)
	}
	if err = (&controllers.CSIPowerMaxRevProxyReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CSIPowerMaxRevProxy"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIPowerMaxRevProxy")
		os.Exit(1)
	}
	if err = (&controllers.CSIIsilonReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CSIIsilon"),
		Scheme: mgr.GetScheme(),
		Config: operatorConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIIsilon")
		os.Exit(1)
	}
	if err = (&controllers.CSIUnityReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CSIUnity"),
		Scheme: mgr.GetScheme(),
		Config: operatorConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIUnity")
		os.Exit(1)
	}
	if err = (&controllers.CSIVXFlexOSReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CSIVXFlexOS"),
		Scheme: mgr.GetScheme(),
		Config: operatorConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIVXFlexOS")
		os.Exit(1)
	}
	if err = (&controllers.CSIPowerStoreReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CSIPowerStore"),
		Scheme: mgr.GetScheme(),
		Config: operatorConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIPowerStore")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
