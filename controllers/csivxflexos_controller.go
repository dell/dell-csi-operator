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

package controllers

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"

	"sigs.k8s.io/yaml"

	"github.com/dell/dell-csi-operator/pkg/utils"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	storagev1 "github.com/dell/dell-csi-operator/api/v1"
	operatorconfig "github.com/dell/dell-csi-operator/pkg/config"
	"github.com/dell/dell-csi-operator/pkg/resources/secrets"
)

// CSIVXFlexOSReconciler reconciles a CSIVXFlexOS object
type CSIVXFlexOSReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Config      operatorconfig.Config
	updateCount int32
}

// +kubebuilder:rbac:groups=storage.dell.com,resources=csivxflexoses;csivxflexoses/finalizers;csivxflexoses/status,verbs=*
// +kubebuilder:rbac:groups="",resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts,verbs=*
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;patch;update
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims/status,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;delete;patch;update
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;replicasets;rolebindings,verbs=get;list;watch;update;create;delete;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles/finalizers,verbs=get;list;watch;update;create;delete;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;update;create;delete;patch
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups="apps",resources=deployments/finalizers,resourceNames=dell-csi-operator-controller-manager,verbs=update
// +kubebuilder:rbac:groups="storage.k8s.io",resources=csidrivers,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups="storage.k8s.io",resources=storageclasses,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=volumeattachments,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="storage.k8s.io",resources=csinodes,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotclasses,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotcontents,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshotcontents/status,verbs=update;patch
// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshots;volumesnapshots/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=create;list;watch;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=volumeattachments/status,verbs=patch
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,resourceNames=privileged,verbs=use

// Reconcile function reconciles a CSIVFlex Object
func (r *CSIVXFlexOSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("csivxflexos", req.NamespacedName)

	// your logic here
	csiVXFlexOS := &storagev1.CSIVXFlexOS{}
	return utils.Reconcile(ctx, csiVXFlexOS, req, r, log)
}

// GetConfig - returns the config
func (r *CSIVXFlexOSReconciler) GetConfig() operatorconfig.Config {
	return r.Config
}

// GetClient - returns the split client
func (r *CSIVXFlexOSReconciler) GetClient() client.Client {
	return r.Client
}

// GetScheme - Returns k8s runtime scheme
func (r *CSIVXFlexOSReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// GetUpdateCount - Returns the current update count
func (r *CSIVXFlexOSReconciler) GetUpdateCount() int32 {
	return r.updateCount
}

// SetClient - Sets the split client (only for testing)
func (r *CSIVXFlexOSReconciler) SetClient(c client.Client) {
	r.Client = c
}

// SetScheme - Sets k8s runtime scheme (only for testing)
func (r *CSIVXFlexOSReconciler) SetScheme(s *runtime.Scheme) {
	r.Scheme = s
}

// SetConfig - Sets the config (only for testing)
func (r *CSIVXFlexOSReconciler) SetConfig(c operatorconfig.Config) {
	r.Config = c
}

//IncrUpdateCount - Increments the update count
func (r *CSIVXFlexOSReconciler) IncrUpdateCount() {
	atomic.AddInt32(&r.updateCount, 1)
}

// InitializeDriverSpec - Initializes any uninitialized elements of the instance spec.
// Also initialize the defaults if user didn't set the values
func (r *CSIVXFlexOSReconciler) InitializeDriverSpec(instance storagev1.CSIDriver, reqLogger logr.Logger) (bool, error) {
	// Getting the MDM value from Sdc-Initcontainer and Passing it to Sdc-monitor
	isDriverupdate := false
	driver := instance.GetDriver()
	ctx := context.Background()
	if driver.ConfigVersion == "v2.1.0" {
		var newmdm corev1.EnvVar
		mdmVar, err := r.GetMDMFromSecret(ctx, instance, reqLogger)
		if err != nil {
			return false, err
		}
		for _, initcontainer := range driver.InitContainers {
			if initcontainer.Name == "sdc" {
				k := 0
				initenv := initcontainer.Envs
				for c, env := range initenv {
					if env.Name == "MDM" {
						env.Value = mdmVar
						newmdm = env
						k = c
						break
					}
				}
				initenv[k] = newmdm
				isDriverupdate = true
				break
			}
		}

		for _, sidecar := range driver.SideCars {
			if sidecar.Name == "sdc-monitor" {
				sidenv := sidecar.Envs
				var updatenv corev1.EnvVar
				j := 0
				for c, env := range sidenv {
					if env.Name == "MDM" {
						env.Value = mdmVar
						updatenv = env
						j = c
						break
					}
				}
				sidenv[j] = updatenv
				isDriverupdate = true
			}
		}
		return isDriverupdate, nil
	}
	mdmIP := ""
	mdmFin := ""
	ismdmip := false
	for _, initcontainer := range driver.InitContainers {
		if initcontainer.Name == "sdc" {
			initenv := initcontainer.Envs
			for i, env := range initenv {
				if env.Name == "MDM" {
					mdmIP = env.Value
					mdmFin, ismdmip = ValidateIPAddress(mdmIP)
					if !ismdmip {
						return false, fmt.Errorf("Invalid MDM value. Ip address should be nummeric and comma separated without space")
					}
					env.Value = mdmFin
					initenv[i] = env
					break
				}
			}
		}
	}
	for i, sidecar := range driver.SideCars {
		if sidecar.Name == "sdc-monitor" {
			sidenv := sidecar.Envs
			updatenv, err := Getmdmipforsdc(sidenv, mdmFin, reqLogger)
			if err != nil {
				return false, err
			}
			if len(updatenv.Value) > 0 {
				sidenv = append(sidenv, updatenv)
			}
			sidecar.Envs = sidenv
			driver.SideCars[i] = sidecar
			isDriverupdate = true
		}
	}
	return isDriverupdate, nil
}

// Getmdmipforsdc - Appends MDM value to sdc-monitor
// Function to append MDM varibale to sdc-monitor environment varibale if not provided
func Getmdmipforsdc(sidenv []corev1.EnvVar, mdmFin string, reqLogger logr.Logger) (corev1.EnvVar, error) {
	var updatenv corev1.EnvVar
	var mdmFound bool
	for _, env := range sidenv {
		if env.Name == "MDM" {
			mdmFound = true
			existingIP, ismdmip := ValidateIPAddress(env.Value)
			if !ismdmip {
				return updatenv, fmt.Errorf("Invalid MDM value, ip address should be nummeric, comma separated without space")
			}
			if mdmFin != existingIP {
				env.Value = mdmFin
				updatenv = env
				break
			}
		}
	}
	if !mdmFound {
		updatenv = corev1.EnvVar{
			Name:  "MDM",
			Value: mdmFin,
		}
	}
	return updatenv, nil
}

// ValidateDriverSpec - Validates the driver spec
// returns error if the spec is not valid
func (r *CSIVXFlexOSReconciler) ValidateDriverSpec(ctx context.Context, instance storagev1.CSIDriver, reqLogger logr.Logger) error {
	// Validates the HOST_PID value from manifest file
	isfound := false
	driver := instance.GetDriver()
	for _, sideCar := range driver.SideCars {
		if sideCar.Name == "sdc-monitor" {
			for _, env := range sideCar.Envs {
				if env.Name == "HOST_PID" {
					if env.Value == "1" || env.Value == "0" {
						isfound = true
						break
					} else {
						return fmt.Errorf("Invalid HOST_PID value, it should be 0 or 1")
					}
				}
			}
		}
		if isfound {
			break
		}
	}
	return nil
}

// ValidateIPAddress - Validates the Ip Address
// returns error if the Ip Address is not valid
func ValidateIPAddress(ipAdd string) (string, bool) {
	trimIP := strings.Split(ipAdd, ",")
	if len(trimIP) < 1 {
		return "", false
	}
	newIP := ""
	for i := range trimIP {
		trimIP[i] = strings.TrimSpace(trimIP[i])
		istrueip := IsIpv4Regex(trimIP[i])
		if istrueip {
			newIP = strings.Join(trimIP[:], ",")
		} else {
			return newIP, false
		}
	}
	return newIP, true
}

var (
	ipRegex, _ = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
)

// IsIpv4Regex - Matches Ipaddress with regex
// returns error if the Ip Address doesn't match regex
func IsIpv4Regex(ipAddress string) bool {
	return ipRegex.MatchString(ipAddress)
}

// GetMDMFromSecret - Returns MDM value
func (r *CSIVXFlexOSReconciler) GetMDMFromSecret(ctx context.Context, instance storagev1.CSIDriver, log logr.Logger) (string, error) {
	client := r.GetClient()
	secretName := fmt.Sprintf("%s-config", instance.GetDriverType())
	credSecret, err := secrets.GetSecret(ctx, secretName, instance.GetNamespace(), client, log)
	if err != nil {
		return "", fmt.Errorf("reading secret [%s] error [%s]", secretName, err)
	}

	type StorageArrayConfig struct {
		Username                  string `json:"username"`
		Password                  string `json:"password"`
		SystemID                  string `json:"systemId"`
		Endpoint                  string `json:"endpoint"`
		SkipCertificateValidation bool   `json:"skipCertificateValidation,omitempty"`
		AllSystemNames            string `json:"allSystemNames"`
		IsDefault                 bool   `json:"isDefault,omitempty"`
		MDM                       string `json:"mdm"`
	}

	//To parse the secret json file
	data := credSecret.Data
	configBytes := data["config"]
	mdmVal := ""
	mdmFin := ""
	ismdmip := false

	if string(configBytes) != "" {
		yamlConfig := make([]StorageArrayConfig, 0)
		configs, _ := yaml.JSONToYAML(configBytes)
		err := yaml.Unmarshal(configs, &yamlConfig)
		if err != nil {
			return "", fmt.Errorf("unable to parse multi-array configuration[%v]", err)
		}

		if len(yamlConfig) == 0 {
			return "", fmt.Errorf("Arrays details are not provided in vxflexos-config secret")
		}

		var noOfDefaultArrays int
		tempMapToFindDuplicates := make(map[string]interface{}, 0)
		for i, config := range yamlConfig {
			if config.SystemID == "" {
				return "", fmt.Errorf("invalid value for ArrayID at index [%d]", i)
			}
			if config.Username == "" {
				return "", fmt.Errorf("invalid value for Username at index [%d]", i)
			}
			if config.Password == "" {
				return "", fmt.Errorf("invalid value for Password at index [%d]", i)
			}
			if config.Endpoint == "" {
				return "", fmt.Errorf("invalid value for RestGateway at index [%d]", i)
			}
			if config.MDM != "" {
				mdmFin, ismdmip = ValidateIPAddress(config.MDM)
				if !ismdmip {
					return "", fmt.Errorf("Invalid MDM value. Ip address should be numeric and comma separated without space")
				}
				if i == 0 {
					mdmVal += mdmFin
				} else {
					mdmVal += "&" + mdmFin
				}
			}
			if config.AllSystemNames != "" {
				names := strings.Split(config.AllSystemNames, ",")
				log.Info("For systemID %s configured System Names found %#v ", config.SystemID, names)
			}

			if _, ok := tempMapToFindDuplicates[config.SystemID]; ok {
				return "", fmt.Errorf("Duplicate ArrayID [%s] found in storageArrayList parameter", config.SystemID)
			}
			tempMapToFindDuplicates[config.SystemID] = nil

			if config.IsDefault {
				noOfDefaultArrays++
			}

			if noOfDefaultArrays > 1 {
				return "", fmt.Errorf("'isDefaultArray' parameter located in multiple places ArrayID: %s. 'isDefaultArray' parameter should present only once in the storageArrayList", config.SystemID)
			}
		}
	} else {
		return "", fmt.Errorf("Arrays details are not provided in vxflexos-config secret")
	}
	fmt.Printf("mdmValFin: %s", mdmVal)
	return mdmVal, nil
}

// SetupWithManager - sets up the controller
func (r *CSIVXFlexOSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("CSIVXFlexOS", mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.Log.Error(err, "Unable to setup CSIVXFlexOS controller")
		os.Exit(1)
	}

	err = c.Watch(
		&source.Kind{Type: &storagev1.CSIVXFlexOS{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		r.Log.Error(err, "Unable to watch CSIVXFlexOS Driver")
		os.Exit(1)
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIVXFlexOS{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Deployment")
		os.Exit(1)
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storagev1.CSIVXFlexOS{},
	})
	if err != nil {
		r.Log.Error(err, "Unable to watch Daemonset")
		os.Exit(1)
	}
	return nil
}
