package ctrlconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_config")

// DriverEnv - Type representing an environment variable for the drivers
type DriverEnv struct {
	Name                      string      `json:"Name"`
	Mandatory                 bool        `json:"Mandatory,omitempty"`
	CSIEnvType                EnvDataType `json:",CSIEnvType"`
	SetForController          bool        `json:"SetForController"`
	SetForNode                bool        `json:"SetForNode"`
	DefaultValueForController string      `json:"DefaultValueForController"`
	DefaultValueForNode       string      `json:"DefaultValueForNode"`
}

// DriverConfig - Type representing the default configuration of the driver
type DriverConfig struct {
	ControllerHA           bool                  `json:"controllerHA,omitempty"`
	EnableEphemeralVolumes bool                  `json:"enableEphemeralVolumes,omitempty"`
	DriverEnvs             []DriverEnv           `json:"driverEnvs"`
	NodeVolumes            []corev1.Volume       `json:"driverNodeVolumes"`
	ControllerVolumes      []corev1.Volume       `json:"driverControllerVolumes"`
	NodeVolumeMounts       []corev1.VolumeMount  `json:"driverNodeVolumeMounts"`
	ControllerVolumeMounts []corev1.VolumeMount  `json:"driverControllerVolumeMounts"`
	DriverArgs             []string              `json:"driverArgs"`
	ControllerTolerations  []corev1.Toleration   `json:"controllerTolerations"`
	NodeTolerations        []corev1.Toleration   `json:"nodeTolerations"`
	SidecarParams          []SidecarParams       `json:"sidecarParams"`
	InitContainerParams    []InitContainerParams `json:"initContainerParams"`
	StorageClassParams     []StorageClassParam   `json:"storageClassParams"`
	StorageClassAttrs      []StorageClassAttr    `json:"storageClassAttrs"`
}

// SidecarParams  - represents configuration for a side car container
type SidecarParams struct {
	Name         csiv1.ImageType      `json:"Name"`
	Optional     bool                 `json:"optional"`
	Args         []string             `json:"args"`
	Envs         []corev1.EnvVar      `json:"envs"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts"`
}

//InitContainerParams - represents configuration for InitContainers
type InitContainerParams struct {
	Name             csiv1.ImageType      `json:"Name"`
	Optional         bool                 `json:"optional"`
	SetForController bool                 `json:"SetForController"`
	SetForNode       bool                 `json:"SetForNode"`
	Args             []string             `json:"args"`
	Envs             []corev1.EnvVar      `json:"envs"`
	VolumeMounts     []corev1.VolumeMount `json:"volumeMounts"`
}

// StorageClassParam represents a single storage class parameter
type StorageClassParam struct {
	Name      string `json:"Name"`
	Mandatory bool   `json:"Mandatory"`
}

// StorageClassAttr represents attributes of a storage class (like volume expansion)
type StorageClassAttr struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// DriverConfigMap - Top level structure for reading from json
type DriverConfigMap struct {
	DriverConfig DriverConfig `json:"driverConfig"`
}

// EnvDataType represents type of an environment variable
type EnvDataType string

// Constants representing different types of acceptable types
const (
	StringType          EnvDataType = "String"
	BooleanType         EnvDataType = "Boolean"
	ListType            EnvDataType = "List"
	IntType             EnvDataType = "Int"
	FloatType           EnvDataType = "FloatType"
	EnvVarReferenceType EnvDataType = "EnvVarReferenceType"
	SecretReferenceType EnvDataType = "EnvSecretReference"
)

// readConfig - reads a config for a specified driver type
// returns an error if file not found or if there is an error
// in un-marshalling the json content
func readConfig(configDirectory string, driverVersion string, log logr.Logger) (DriverConfig, error) {
	var driverConfigMap DriverConfigMap
	jsonFileName := filepath.Join(configDirectory, fmt.Sprintf("%s.json", driverVersion))
	jsonFile, err := os.Open(filepath.Clean(jsonFileName))
	if err != nil {
		log.Error(err, "unable to find config file for driver")
		return DriverConfig{}, err
	}
	log.V(3).Info("Reading", jsonFileName, " for default config")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &driverConfigMap)
	if err != nil {
		log.Error(err, "error in umarshaling driver config")
	}
	return driverConfigMap.DriverConfig, nil
}

// Config - Represents overall configuration including driver specific configuration
type Config struct {
	ConfigVersion  string
	ConfigFileName string
	KubeAPIVersion csiv1.K8sVersion
	imageMap       map[string]string
	DriverType     csiv1.DriverType
	DriverVersion  string
	DriverConfig   *DriverConfig
	StrictCheck    bool
	Log            logr.Logger
	IsOpenShift    bool
}

// InitDriverConfig - Initializes driver config by reading files in a config directory
func (c *Config) InitDriverConfig(configDirectory string) error {
	opConfig, err := ReadOpConfig(configDirectory, c.ConfigFileName)
	if err != nil {
		log.Error(err, "error in reading operator config")
		return err
	}
	err = opConfig.IsSupportedVersion(c.DriverType, c.ConfigVersion, c.KubeAPIVersion)
	if err != nil {
		return err
	}
	imageMap, err := opConfig.GetDefaultImageTags(c.DriverType, c.ConfigVersion, c.KubeAPIVersion)
	if err != nil {
		return err
	}
	c.imageMap = imageMap
	configVersion := strings.Replace(c.ConfigVersion, ".", "", -1)
	driverVersion := fmt.Sprintf("%s_%s_%s", string(c.DriverType), configVersion, c.KubeAPIVersion)
	c.DriverVersion = driverVersion

	driverConfig, err := readConfig(configDirectory, driverVersion, c.Log)
	if err != nil {
		c.Log.Error(err, fmt.Sprintf("Failed to read config for driver: %s", string(c.DriverType)))
		return err
	}
	c.DriverConfig = &driverConfig
	return nil
}

// IsControllerHAEnabled - Determines whether Controller HA is enabled or not
func (c *Config) IsControllerHAEnabled(imageName string) bool {
	return c.DriverConfig.ControllerHA
}

// GetDefaultImageTag - Returns the default image tag for a given image name
func (c *Config) GetDefaultImageTag(imageName string) (string, error) {
	if _, ok := c.imageMap[imageName]; !ok {
		return "", fmt.Errorf("failed to find image tag for: %s", imageName)
	}
	return strings.TrimSpace(c.imageMap[imageName]), nil
}

// GetControllerEnvs - Returns an array of corev1.EnvVar
func (c *Config) GetControllerEnvs() []corev1.EnvVar {
	var envs = make([]corev1.EnvVar, 0)
	if c.DriverConfig == nil {
		return envs
	}
	for _, env := range c.DriverConfig.DriverEnvs {
		if env.SetForController {
			if env.CSIEnvType == EnvVarReferenceType {
				fields := strings.Split(env.DefaultValueForController, "/")
				if len(fields) != 2 {
					c.Log.V(2).Info("Invalid default value found. Ignoring this")
					continue
				}
				apiVersion := fields[0]
				fieldPath := fields[1]
				envs = append(envs, corev1.EnvVar{
					Name: env.Name,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: apiVersion,
							FieldPath:  fieldPath,
						},
					},
				})
			} else if env.CSIEnvType == SecretReferenceType {
				fields := strings.Split(env.DefaultValueForController, "/")
				if len(fields) != 2 {
					c.Log.V(2).Info("Invalid default value found. Ignoring this")
					continue
				}
				secretKey := fields[0]
				secretName := fields[1]
				envs = append(envs, corev1.EnvVar{
					Name: env.Name,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: secretKey,
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName},
						},
					},
				})
			} else {
				envs = append(envs, corev1.EnvVar{
					Name:  env.Name,
					Value: env.DefaultValueForController,
				})
			}
		}
	}
	return envs
}

// GetNodeEnvs - Returns an array of corev1.EnvVar for the Node
func (c *Config) GetNodeEnvs() []corev1.EnvVar {
	var envs = make([]corev1.EnvVar, 0)
	if c.DriverConfig == nil {
		return envs
	}
	for _, env := range c.DriverConfig.DriverEnvs {
		if env.SetForNode {
			if env.CSIEnvType == EnvVarReferenceType {
				fields := strings.Split(env.DefaultValueForNode, "/")
				if len(fields) != 2 {
					c.Log.V(2).Info("Invalid default value found for node. Ignoring this")
					continue
				}
				apiVersion := fields[0]
				fieldPath := fields[1]
				envs = append(envs, corev1.EnvVar{
					Name: env.Name,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: apiVersion,
							FieldPath:  fieldPath,
						},
					},
				})
			} else if env.CSIEnvType == SecretReferenceType {
				fields := strings.Split(env.DefaultValueForNode, "/")
				if len(fields) != 2 {
					c.Log.V(2).Info("Invalid default value found for node. Ignoring this")
					continue
				}
				secretKey := fields[0]
				secretName := fields[1]
				envs = append(envs, corev1.EnvVar{
					Name: env.Name,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: secretKey,
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName},
						},
					},
				})
			} else {
				envs = append(envs, corev1.EnvVar{
					Name:  env.Name,
					Value: env.DefaultValueForNode,
				})
			}
		}
	}
	return envs
}

// GetControllerVolumes - Returns an array of corev1.Volume for the Controller
func (c *Config) GetControllerVolumes() []corev1.Volume {
	controllerVolumes := make([]corev1.Volume, 0)
	if c.DriverConfig == nil {
		return controllerVolumes
	}

	return c.DriverConfig.ControllerVolumes
}

// GetNodeVolumes - Returns an array of corev1.Volume for the Node
func (c *Config) GetNodeVolumes() []corev1.Volume {
	nodeVolumes := make([]corev1.Volume, 0)
	if c.DriverConfig == nil {
		return nodeVolumes
	}
	return c.DriverConfig.NodeVolumes
}

// GetControllerVolumeMounts - Returns an array of corev1.VolumeMount for the Controller
func (c *Config) GetControllerVolumeMounts() []corev1.VolumeMount {
	controllerVolumeMounts := make([]corev1.VolumeMount, 0)
	if c.DriverConfig == nil {
		return controllerVolumeMounts
	}
	return c.DriverConfig.ControllerVolumeMounts
}

// GetNodeVolumeMounts - Returns an array of corev1.VolumeMount for the Node
func (c *Config) GetNodeVolumeMounts() []corev1.VolumeMount {
	nodeVolumeMounts := make([]corev1.VolumeMount, 0)
	if c.DriverConfig == nil {
		return nodeVolumeMounts
	}
	return c.DriverConfig.NodeVolumeMounts
}

// GetDriverArgs - Returns arguments for the driver container
func (c *Config) GetDriverArgs() []string {
	driverArgs := make([]string, 0)
	if c.DriverConfig == nil {
		return driverArgs
	}
	return c.DriverConfig.DriverArgs
}

// GetDriverControllerTolerations - Returns tolerations for the controller
func (c *Config) GetDriverControllerTolerations() []corev1.Toleration {
	tolerations := make([]corev1.Toleration, 0)
	if c.DriverConfig == nil {
		return tolerations
	}
	return c.DriverConfig.ControllerTolerations
}

// GetDriverNodeTolerations - Returns tolerations for the node
func (c *Config) GetDriverNodeTolerations() []corev1.Toleration {
	tolerations := make([]corev1.Toleration, 0)
	if c.DriverConfig == nil {
		return tolerations
	}
	return c.DriverConfig.NodeTolerations
}

// GetMandatorySideCars - Returns a slice of side car container names which are marked as default
func (c *Config) GetMandatorySideCars() []string {
	sideCarNames := make([]string, 0)
	if c.DriverConfig == nil {
		return sideCarNames
	}
	for _, sidecar := range c.DriverConfig.SidecarParams {
		if !sidecar.Optional {
			sideCarNames = append(sideCarNames, string(sidecar.Name))
		}
	}
	return sideCarNames
}

// GetAllSideCars - Returns a slice of all side car container names
func (c *Config) GetAllSideCars() []string {
	sideCarNames := make([]string, 0)
	if c.DriverConfig == nil {
		return sideCarNames
	}
	for _, sidecar := range c.DriverConfig.SidecarParams {
		sideCarNames = append(sideCarNames, string(sidecar.Name))
	}
	return sideCarNames
}

// GetSideCarArgs - Returns a slice of arguments (strings) for a side car
func (c *Config) GetSideCarArgs(sidecarType csiv1.ImageType) []string {
	sideCarArgs := make([]string, 0)
	if c.DriverConfig == nil {
		return sideCarArgs
	}
	for _, sidecar := range c.DriverConfig.SidecarParams {
		if sidecar.Name == sidecarType {
			return sidecar.Args
		}
	}
	return sideCarArgs
}

// GetSideCarEnvs - Returns a slice of env variables for a side car
func (c *Config) GetSideCarEnvs(sidecarType csiv1.ImageType) []corev1.EnvVar {
	var sideCarEnv = make([]corev1.EnvVar, 0)
	if c.DriverConfig == nil {
		return sideCarEnv
	}
	for _, sidecar := range c.DriverConfig.SidecarParams {
		if sidecar.Name == sidecarType {
			return sidecar.Envs
		}
	}
	return sideCarEnv
}

// GetSideCarVolMounts - Returns a slice of volume mounts for a side car
func (c *Config) GetSideCarVolMounts(sidecarType csiv1.ImageType) []corev1.VolumeMount {
	var sideCarParams = make([]corev1.VolumeMount, 0)
	if c.DriverConfig == nil {
		return sideCarParams
	}
	for _, sidecar := range c.DriverConfig.SidecarParams {
		if sidecar.Name == sidecarType {
			return sidecar.VolumeMounts
		}
	}
	return sideCarParams
}

// GetAllInitContainers - Returns a slice of all init container names
func (c *Config) GetAllInitContainers() []string {
	initContainerNames := make([]string, 0)
	if c.DriverConfig == nil {
		return initContainerNames
	}
	for _, initcontainer := range c.DriverConfig.InitContainerParams {
		initContainerNames = append(initContainerNames, string(initcontainer.Name))
	}
	return initContainerNames
}

// GetNodeInitContainers - Returns a slice of all init container names for nodes
func (c *Config) GetNodeInitContainers() []string {
	initContainerNames := make([]string, 0)
	if c.DriverConfig == nil {
		return initContainerNames
	}
	for _, initcontainer := range c.DriverConfig.InitContainerParams {
		if initcontainer.SetForNode {
			initContainerNames = append(initContainerNames, string(initcontainer.Name))
		}
	}
	return initContainerNames
}

// GetControllerInitContainers - Returns a slice of all init container names for controller
func (c *Config) GetControllerInitContainers() []string {
	initContainerNames := make([]string, 0)
	if c.DriverConfig == nil {
		return initContainerNames
	}
	for _, initcontainer := range c.DriverConfig.InitContainerParams {
		if initcontainer.SetForController {
			initContainerNames = append(initContainerNames, string(initcontainer.Name))
		}
	}
	return initContainerNames
}

// GetInitContainerArgs - Returns a slice of arguments (strings) for an Init Container
func (c *Config) GetInitContainerArgs(InitContainerType csiv1.ImageType) []string {
	initContainerArgs := make([]string, 0)
	if c.DriverConfig == nil {
		return initContainerArgs
	}
	for _, initContainer := range c.DriverConfig.InitContainerParams {
		if initContainer.Name == InitContainerType {
			return initContainer.Args
		}
	}
	return initContainerArgs
}

// GetInitContainerEnvs - Returns a slice of Envs for initContainer
func (c *Config) GetInitContainerEnvs(InitContainerType csiv1.ImageType) []corev1.EnvVar {
	var initContainerEnv = make([]corev1.EnvVar, 0)
	if c.DriverConfig == nil {
		return initContainerEnv
	}
	for _, initcontainer := range c.DriverConfig.InitContainerParams {
		if initcontainer.Name == InitContainerType {
			return initcontainer.Envs
		}
	}
	return initContainerEnv
}

// GetInitContainerVolMounts - Returns a slice of volume mounts for a initContainers
func (c *Config) GetInitContainerVolMounts(InitContainerType csiv1.ImageType) []corev1.VolumeMount {
	var initContainerVol = make([]corev1.VolumeMount, 0)
	if c.DriverConfig == nil {
		return initContainerVol
	}
	for _, initcontainer := range c.DriverConfig.InitContainerParams {
		if initcontainer.Name == InitContainerType {
			return initcontainer.VolumeMounts
		}
	}
	return initContainerVol
}

// GetMandatoryStorageClassParams - Returns a (string) slice of mandatory storage class parameters
func (c *Config) GetMandatoryStorageClassParams() []string {
	mandatoryStorageClassParams := make([]string, 0)
	if c.DriverConfig == nil {
		return mandatoryStorageClassParams
	}
	for _, scParam := range c.DriverConfig.StorageClassParams {
		if scParam.Mandatory {
			mandatoryStorageClassParams = append(mandatoryStorageClassParams, scParam.Name)
		}
	}
	return mandatoryStorageClassParams
}

func checkEnvType(datatype EnvDataType, value string) bool {
	if datatype == StringType {
		return true
	}
	if datatype == BooleanType {
		lowerCaseValue := strings.ToLower(value)
		if lowerCaseValue == "true" || lowerCaseValue == "false" {
			return true
		}
	}
	if datatype == IntType {
		_, err := strconv.Atoi(value)
		if err == nil {
			return true
		}
	}
	if datatype == ListType {
		return true
	}
	if datatype == FloatType {
		_, err := strconv.ParseFloat(value, 32)
		if err == nil {
			return true
		}
	}
	return false
}

// GetMandatoryEnvNames - Returns a slice of mandatory environments for each type
// of driver container - controller/node
func (c *Config) GetMandatoryEnvNames(driverContainerType string) []string {
	mandatoryEnvNames := make([]string, 0)
	if c.DriverConfig == nil {
		return mandatoryEnvNames
	}
	for _, env := range c.DriverConfig.DriverEnvs {
		if driverContainerType == "controller" {
			if env.Mandatory && env.SetForController {
				mandatoryEnvNames = append(mandatoryEnvNames, env.Name)
			}
		} else if driverContainerType == "node" {
			if env.Mandatory && env.SetForNode {
				mandatoryEnvNames = append(mandatoryEnvNames, env.Name)
			}
		}
	}
	return mandatoryEnvNames
}

// ValidateEnvironmentVarType - Used to validate environment variables specified via the user
// We dont expect the user to ever specify EnvVarReference & SecretReferenceType as
// that is always done via the config
func (c *Config) ValidateEnvironmentVarType(envFromSpec corev1.EnvVar) (bool, error) {
	for _, env := range c.DriverConfig.DriverEnvs {
		if envFromSpec.Name == env.Name {
			isValid := checkEnvType(env.CSIEnvType, envFromSpec.Value)
			if isValid {
				if env.Mandatory && envFromSpec.Value == "" {
					return false, fmt.Errorf("missing value for a mandatory env - Name: %s, Value:%s", envFromSpec.Name, envFromSpec.Value)
				}
				return true, nil
			}
			return false, fmt.Errorf("invalid value specified in spec - Name: %s, Value: %s", envFromSpec.Name, envFromSpec.Value)
		}
	}
	if c.StrictCheck {
		return false, fmt.Errorf("failed to find the environment variable: %s in the default config", envFromSpec.Name)
	}
	c.Log.V(2).Info("Warning - Unknown environment variable - %s specified by the user. Continuing",
		envFromSpec.Name)
	return true, nil
}
