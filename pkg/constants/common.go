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

package constants

import (
	"time"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Constants for driver states etc
const (
	RetryCount           = 3
	Running              = csiv1.DriverState("Running")
	Succeeded            = csiv1.DriverState("Succeeded")
	Creating             = csiv1.DriverState("Creating")
	Failed               = csiv1.DriverState("Failed")
	InvalidConfig        = csiv1.DriverState("InvalidConfig")
	NoState              = csiv1.DriverState("")
	Updating             = csiv1.DriverState("Updating")
	DefaultRetryInterval = 5 * time.Second
	MaxRetryInterval     = 10 * time.Minute
	MaxRetryDuration     = 30 * time.Minute
)

// DriverReplicas - Replica count for controller
var DriverReplicas = int32(1)

// RevisionHistoryLimit - Max revision history limit for driver daemonset
var RevisionHistoryLimit = int32(10)

// MaxUnavailableUpdateStrategy - Maximum unavailable update strategy
var MaxUnavailableUpdateStrategy = intstr.IntOrString{IntVal: 1, StrVal: "1"}

// TerminationMessagePath for the container
const TerminationMessagePath = "/dev/termination-log"

// TerminationMessagePolicy determines the policy for termination message
const TerminationMessagePolicy = "File"

// DriverMountPath - Mount path for the driver container
const DriverMountPath = "/var/run/csi"

// DriverMountName - Socket directory volume mount name
const DriverMountName = "socket-dir"

// TerminationGracePeriodSeconds - grace period in seconds
var TerminationGracePeriodSeconds = int64(30)
