package config

import csiv1 "github.com/dell/dell-csi-operator/api/v1"

// DriverType - Represents the type of the driver
type DriverType string

// Constants for driver types
const (
	PowerMax   DriverType = "powermax"
	Unity      DriverType = "unity"
	VxFlexOs   DriverType = "vxflexos"
	Isilon     DriverType = "isilon"
	PowerStore DriverType = "powerstore"
	Unknown    DriverType = "unknown"
)

// Config - Holds configuration information for an operator
type Config struct {
	ConfigDirectory      string
	ConfigFile           string
	KubeAPIServerVersion csiv1.K8sVersion
	EnabledDrivers       []csiv1.DriverType
	RetryCount           int32
	IsOpenShift          bool
}

// GetDriverType - gets the driver type from a string
func GetDriverType(driverName string) csiv1.DriverType {
	switch driverName {
	case "powermax":
		return csiv1.PowerMax
	case "unity":
		return csiv1.Unity
	case "isilon":
		return csiv1.Isilon
	case "vxflexos":
		return csiv1.VXFlexOS
	case "powerstore":
		return csiv1.PowerStore
	}
	return csiv1.Unknown
}
