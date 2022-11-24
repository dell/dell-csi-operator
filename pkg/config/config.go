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
