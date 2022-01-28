package utils

import (
	"fmt"
	"reflect"
	"strings"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// appendIfMissingString - Appends a string to a slice if not already present
func appendIfMissingString(slice []string, str string) []string {
	for _, ele := range slice {
		if ele == str {
			return slice
		}
	}
	return append(slice, str)
}

// appendIfMissingEnvVar - Appends a environment variable to a slice if not already present
func appendIfMissingEnvVar(slice []corev1.EnvVar, env corev1.EnvVar) []corev1.EnvVar {
	for i := 0; i < len(slice); i++ {
		if slice[i].Name == env.Name {
			slice[i] = env
			return slice
		}
	}
	return append(slice, env)
}

// mergeEnvironmentVars - Merges a sourcelist with a new list
// merge strategy - If env is present in both lists, env from new list takes priority
// If env is not present in the source list, it is added
func mergeEnvironmentVars(sourceEnvList []corev1.EnvVar, newEnvList []corev1.EnvVar) []corev1.EnvVar {
	if sourceEnvList == nil {
		sourceEnvList = make([]corev1.EnvVar, 0)
	}
	mergedEnvList := sourceEnvList
	for _, newenv := range newEnvList {
		mergedEnvList = appendIfMissingEnvVar(mergedEnvList, newenv)
	}
	return mergedEnvList
}

func appendIfMissingToleration(slice []corev1.Toleration, toleration corev1.Toleration) []corev1.Toleration {
	for i := 0; i < len(slice); i++ {
		if reflect.DeepEqual(slice[i], toleration) {
			return slice
		}
	}
	return append(slice, toleration)
}

// mergeTolerations - Merges a sourcelist with a new list
// merge strategy - If toleration is present in both lists, toleration from new list takes priority
// If toleration is not present in the source list, it is added
func mergeTolerations(sourceTolerationList []corev1.Toleration, newTolerationList []corev1.Toleration) []corev1.Toleration {
	if sourceTolerationList == nil {
		sourceTolerationList = make([]corev1.Toleration, 0)
	}
	mergedTolerationList := sourceTolerationList
	for _, newToleration := range newTolerationList {
		mergedTolerationList = appendIfMissingToleration(mergedTolerationList, newToleration)
	}
	return mergedTolerationList
}

// removeEnvVar - removes a env from the env list
func removeEnvVar(envList []corev1.EnvVar, envVarName string) []corev1.EnvVar {
	i := 0 // output index
	isUpated := false
	for _, env := range envList {
		if env.Name != envVarName {
			envList[i] = env
			i++
		} else {
			isUpated = true
		}
	}
	if isUpated {
		envList = envList[:i]
	}
	return envList
}

// appendIfMissingVolume - Appends a volume to a slice if not already present
func appendIfMissingVolume(slice []corev1.Volume, vol corev1.Volume) []corev1.Volume {
	for i := 0; i < len(slice); i++ {
		if slice[i].Name == vol.Name {
			slice[i] = vol
			return slice
		}
	}
	return append(slice, vol)
}

// mergeVolumes - Merges a sourcelist with a new list
// merge strategy - If volume is present in both lists, volume from new list takes priority
// If volume is not present in the source list, it is added
func mergeVolumes(sourceVolumeList []corev1.Volume, newVolumeList []corev1.Volume) []corev1.Volume {
	mergedVolumeList := sourceVolumeList
	for _, newvolume := range newVolumeList {
		mergedVolumeList = appendIfMissingVolume(mergedVolumeList, newvolume)
	}
	return mergedVolumeList
}

// appendIfMissingVolumeMount - Appends a volume to a slice if not already present
func appendIfMissingVolumeMount(slice []corev1.VolumeMount, volumeMount corev1.VolumeMount) []corev1.VolumeMount {
	for i := 0; i < len(slice); i++ {
		if slice[i].Name == volumeMount.Name {
			slice[i] = volumeMount
			return slice
		}
	}
	return append(slice, volumeMount)
}

// mergeVolumeMounts - Merges a sourcelist with a new list
// merge strategy - If VolumeMount is present in both lists, VolumeMount from new list takes priority
// If VolumeMount is not present in the source list, it is added
func mergeVolumeMounts(sourceVolumeMountList []corev1.VolumeMount, newVolumeMountList []corev1.VolumeMount) []corev1.VolumeMount {
	mergedVolumeMountList := sourceVolumeMountList
	for _, newVolumeMount := range newVolumeMountList {
		mergedVolumeMountList = appendIfMissingVolumeMount(mergedVolumeMountList, newVolumeMount)
	}
	return mergedVolumeMountList
}

func driverChanged(instance csiv1.CSIDriver) (uint64, uint64, bool) {
	expectedHash := HashDriver(instance)
	return expectedHash, instance.GetDriverStatus().DriverHash, instance.GetDriverStatus().DriverHash != expectedHash
}

// ProxyChanged - Checks if proxy spec has changed
func ProxyChanged(instance *csiv1.CSIPowerMaxRevProxy) (uint64, uint64, bool) {
	expectedHash := HashProxy(instance)
	return expectedHash, instance.Status.ProxyHash, instance.Status.ProxyHash != expectedHash
}

func logBannerAndReturn(result reconcile.Result, err error, reqLogger logr.Logger) (reconcile.Result, error) {
	reqLogger.Info("################End Reconcile##############")
	return result, err
}

// mergeArgs - Merges a source list with a new list
// merge strategy - If argument is present in both lists, arg from new list takes priority
// If arg is not present in the source list, it is added
func mergeArgs(sourceArgs []string, newArgList []string) []string {
	mergedArgs := sourceArgs
	for _, newArg := range newArgList {
		mergedArgs = appendIfMissingArgs(mergedArgs, newArg)
	}
	return mergedArgs
}

// appendIfMissingArgs - Appends an argument to a slice if not already present
func appendIfMissingArgs(slice []string, arg string) []string {
	argName := strings.Split(arg, "=")[0]
	for i := 0; i < len(slice); i++ {
		tempArgName := strings.Split(slice[i], "=")[0]
		if argName == tempArgName {
			slice[i] = arg
			return slice
		}
	}
	return append(slice, arg)
}

func isStringInSlice(str string, slice []string) bool {
	for _, ele := range slice {
		if ele == str {
			return true
		}
	}
	return false
}

func getEnvVar(name string, envs []corev1.EnvVar) (corev1.EnvVar, error) {
	for _, env := range envs {
		if env.Name == name {
			return env, nil
		}
	}
	return corev1.EnvVar{}, fmt.Errorf("env name: %s not present", name)
}

func getEnvValue(name string, envs []corev1.EnvVar) (string, error) {
	for _, env := range envs {
		if env.Name == name {
			return env.Value, nil
		}
	}
	return "", fmt.Errorf("env name: %s not present", name)
}

func isEnvPresent(name string, envs []corev1.EnvVar) (bool, string) {
	for _, env := range envs {
		if env.Name == name {
			return true, env.Value
		}
	}
	return false, ""
}

func GetSideCarTypeFromName(sideCarName string) (csiv1.ImageType, error) {
	switch sideCarName {
	case csiv1.Provisioner:
		return csiv1.ImageTypeProvisioner, nil
	case csiv1.Attacher:
		return csiv1.ImageTypeAttacher, nil
	case csiv1.Snapshotter:
		return csiv1.ImageTypeSnapshotter, nil
	case csiv1.Resizer:
		return csiv1.ImageTypeResizer, nil
	case csiv1.Registrar:
		return csiv1.ImageTypeRegistrar, nil
	case csiv1.Sdcmonitor:
		return csiv1.ImageTypeSdcmonitor, nil
	case csiv1.Healthmonitor:
		return csiv1.ImageTypeHealthmonitor, nil
	}
	return "", fmt.Errorf("invalid image type specified")
}

func getinitContainerTypeFromName(initContainerName string) (csiv1.ImageType, error) {
	switch initContainerName {
	case csiv1.Sdc:
		return csiv1.ImageTypeSdc, nil
	}
	return "", fmt.Errorf("invalid image type specified")
}

// IsLimitedNodeRBAC - Returns a boolean which indicates if limited node RBAC is required
func IsLimitedNodeRBAC(driverType csiv1.DriverType, configVersion string) bool {
	limitedNodeRBAC := true
	switch driverType {
	case csiv1.PowerMax:
		if configVersion == "v1" {
			limitedNodeRBAC = false
		}
	case csiv1.Isilon:
		limitedNodeRBAC = false
	case csiv1.Unity:
		limitedNodeRBAC = false
	case csiv1.VXFlexOS:
		limitedNodeRBAC = false
	}
	return limitedNodeRBAC
}

// IsCustomDriverNameSupported - Returns a boolean which indicates if custom driver names are supported
// by a specific config version for a driver
func IsCustomDriverNameSupported(instance csiv1.CSIDriver, configVersion string) bool {
	customNameEnabled := false
	switch instance.GetDriverType() {
	case csiv1.PowerMax:
		if configVersion != "v1" {
			customNameEnabled = true
		}
	}
	return customNameEnabled
}

// GetCustomDriverName - Returns a custom driver name
// If no custom driver name is requested, then an empty string is returned
func GetCustomDriverName(instance csiv1.CSIDriver, configVersion string, envs []corev1.EnvVar, reqLogger logr.Logger) string {
	if IsCustomDriverNameSupported(instance, configVersion) {
		env, err := getEnvVar("X_CSI_POWERMAX_DRIVER_NAME", envs)
		if err != nil {
			// We didn't get the environment
			// This is unexpected. Log a warning and return an empty string
			reqLogger.Error(err, "Unexpected error. Custom driver name supported by driver but env not detected")
			return ""
		}
		if env.Value != instance.GetDefaultDriverName() {
			return env.Value
		}
	}
	return ""
}

// GetCustomEnvVars - Returns a list of custom environment variables
func GetCustomEnvVars(instance csiv1.CSIDriver, configVersion string,
	envs []corev1.EnvVar) []corev1.EnvVar {
	customEnvVars := make([]corev1.EnvVar, 0)
	if IsCustomDriverNameSupported(instance, configVersion) {
		driverName := ""
		env, err := getEnvVar(instance.GetDriverEnvName(), envs)
		if err == nil {
			if env.Value != instance.GetDefaultDriverName() {
				driverName = fmt.Sprintf("%s.%s.dellemc.com", instance.GetNamespace(), env.Value)
			} else {
				driverName = instance.GetDefaultDriverName()
			}
			driverEnv := corev1.EnvVar{
				Name:  instance.GetDriverEnvName(),
				Value: driverName,
			}
			customEnvVars = append(customEnvVars, driverEnv)
		}
	}
	return customEnvVars
}
