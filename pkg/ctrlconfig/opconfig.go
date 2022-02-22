package ctrlconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"gopkg.in/yaml.v2"
)

// OpConfig - Represents the image & RBAC config used by Operator
type OpConfig struct {
	Drivers              []DriverConfigParams `yaml:"drivers"`
	CSISideCars          []ContainerApp       `yaml:"csiSideCars"`
	Extensions           []ContainerApp       `yaml:"extensions,omitempty"`
	SupportedK8sVersions []csiv1.K8sVersion   `yaml:"supportedK8sVersions"`
}

// DriverConfigParams - Represents the driver and its config versions
type DriverConfigParams struct {
	Name           csiv1.DriverType      `yaml:"name"`
	ConfigVersions []ConfigVersionParams `yaml:"configVersions"`
}

// ConfigVersionParams - Represents a specific config version of the driver
type ConfigVersionParams struct {
	ConfigVersion     string                   `yaml:"configVersion"`
	UseDefaults       bool                     `yaml:"useDefaults,omitempty"`
	SupportedVersions []SupportedVersionParams `yaml:"supportedVersions"`
	Attacher          string                   `yaml:"attacher,omitempty"`
	Provisoner        string                   `yaml:"provisioner,omitempty"`
	Resizer           string                   `yaml:"resizer,omitempty"`
	Snapshotter       string                   `yaml:"snapshotter,omitempty"`
	Registrar         string                   `yaml:"registrar,omitempty"`
}

// SupportedVersionParams - Represents the supported versions and corresponding sidecars
type SupportedVersionParams struct {
	Version     csiv1.K8sVersion `yaml:"version"`
	Attacher    string           `yaml:"attacher,omitempty"`
	Provisioner string           `yaml:"provisioner,omitempty"`
	Resizer     string           `yaml:"resizer,omitempty"`
	Snapshotter string           `yaml:"snapshotter,omitempty"`
	Registrar   string           `yaml:"registrar,omitempty"`
}

// ImageTag - Image tag with associated K8s version
type ImageTag struct {
	K8sVersion csiv1.K8sVersion `yaml:"version"`
	Tag        string           `yaml:"tag,omitempty"`
}

// ContainerApp - Represents any other container
type ContainerApp struct {
	Name   string     `yaml:"name"`
	Images []ImageTag `yaml:"images"`
}

// ReadOpConfig - Reads Operator config
func ReadOpConfig(configDirectory, configFileName string) (*OpConfig, error) {
	configFile := filepath.Join(configDirectory, configFileName)
	log.Info("Reading file for default image tags", "Filename", configFile)
	jsonFile, err := os.Open(filepath.Clean(configFile))
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	var opConfig OpConfig
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = yaml.Unmarshal(byteValue, &opConfig)
	if err != nil {
		return nil, err
	}
	return &opConfig, nil
}

// GetSupportedK8sVersions - Returns supported k8s versions
func (opConfig *OpConfig) GetSupportedK8sVersions() []csiv1.K8sVersion {
	return opConfig.SupportedK8sVersions
}

// IsSupportedVersion - Returns if a specific driver config version is supported
func (opConfig *OpConfig) IsSupportedVersion(driverType csiv1.DriverType, driverConfigVersion string,
	version csiv1.K8sVersion) error {
	found := false
	for _, supportedVersion := range opConfig.SupportedK8sVersions {
		if supportedVersion == version {
			found = true
			break
		}
	}
	if found {
		driverFound := false
		for _, driver := range opConfig.Drivers {
			if driver.Name == driverType {
				driverFound = true
				configVersionFound := false
				for _, configVersion := range driver.ConfigVersions {
					if configVersion.ConfigVersion == driverConfigVersion {
						configVersionFound = true
						configVersionSupported := false
						for _, supportedVersion := range configVersion.SupportedVersions {
							if supportedVersion.Version == version {
								configVersionSupported = true
								break
							}
						}
						if configVersionSupported {
							return nil
						}
						return fmt.Errorf("driver config version not supported on this K8s version")
					}
				}
				if !configVersionFound {
					return fmt.Errorf("unknown driver config version")
				}
				break
			}
		}
		if !driverFound {
			return fmt.Errorf("unknown driver type")
		}
	}
	return fmt.Errorf("k8s version not supported by operator")
}

// GetDefaultImageTags - Returns an image map
func (opConfig *OpConfig) GetDefaultImageTags(driverType csiv1.DriverType, configVersion string,
	k8sVersion csiv1.K8sVersion) (map[string]string, error) {
	err := opConfig.IsSupportedVersion(driverType, configVersion, k8sVersion)
	if err != nil {
		return nil, err
	}
	imageMap := make(map[string]string)
	for _, driver := range opConfig.Drivers {
		if driver.Name == driverType {
			for _, configVersionParams := range driver.ConfigVersions {
				if configVersionParams.ConfigVersion == configVersion {
					if configVersionParams.UseDefaults {
						for _, sideCar := range opConfig.CSISideCars {
							for _, image := range sideCar.Images {
								if image.K8sVersion == k8sVersion {
									imageMap[string(sideCar.Name)] = image.Tag
								}
							}
						}
					} else {
						imageMap[csiv1.Provisioner] = configVersionParams.Provisoner
						imageMap[csiv1.Attacher] = configVersionParams.Attacher
						imageMap[csiv1.Resizer] = configVersionParams.Resizer
						imageMap[csiv1.Snapshotter] = configVersionParams.Snapshotter
						imageMap[csiv1.Registrar] = configVersionParams.Registrar
						for _, supportedVersion := range configVersionParams.SupportedVersions {
							if supportedVersion.Version == k8sVersion {
								if supportedVersion.Provisioner != "" {
									imageMap[csiv1.Provisioner] = supportedVersion.Provisioner
								}
								if supportedVersion.Attacher != "" {
									imageMap[csiv1.Attacher] = supportedVersion.Attacher
								}
								if supportedVersion.Resizer != "" {
									imageMap[csiv1.Resizer] = supportedVersion.Resizer
								}
								if supportedVersion.Snapshotter != "" {
									imageMap[csiv1.Snapshotter] = supportedVersion.Snapshotter
								}
								if supportedVersion.Registrar != "" {
									imageMap[csiv1.Registrar] = supportedVersion.Registrar
								}
							}
						}
					}
					break
				}
			}
			break
		}
	}
	// Also populate any container application image tags
	for _, app := range opConfig.Extensions {
		for _, image := range app.Images {
			if k8sVersion == image.K8sVersion {
				imageMap[app.Name] = image.Tag
			}
		}
	}
	// Validate if all image tags were populated
	for k, v := range imageMap {
		if v == "" {
			return nil, fmt.Errorf("default image tag not found for: %s", k)
		}
	}
	return imageMap, nil
}
