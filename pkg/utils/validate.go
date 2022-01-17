package utils

import (
	"context"
	"fmt"
	"strings"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/ctrlconfig"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// ValidateSpec - Validates the user specified spec
func ValidateSpec(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI, driverConfig *ctrlconfig.Config, log logr.Logger) error {
	driver := instance.GetDriver()
	common := driver.Common
	controller := driver.Controller
	node := driver.Node
	if common.Image == "" {
		return fmt.Errorf("driver image not specified in spec")
	}
	// Check is the credentials secret exists for controller
	err := checkIfCredentialsSecretExists(ctx, instance, r, driverConfig, "controller", log)
	if err != nil {
		return err
	}
	combinedControllerEnvs := mergeEnvironmentVars(common.Envs, controller.Envs)
	err = validateUserEnv(driverConfig, combinedControllerEnvs, "controller")
	if err != nil {
		return err
	}
	// Check for controller secret
	if isCertificateValidationRequested(combinedControllerEnvs, string(instance.GetDriverType())) {
		err = checkCertSecret(ctx, instance, r, driverConfig, "controller", log)
		if err != nil {
			return err
		}
	}
	// Check is the credentials secret exists for node
	err = checkIfCredentialsSecretExists(ctx, instance, r, driverConfig, "node", log)
	if err != nil {
		return err
	}
	combinedNodeEnvs := mergeEnvironmentVars(common.Envs, node.Envs)
	err = validateUserEnv(driverConfig, combinedNodeEnvs, "node")
	if err != nil {
		return err
	}
	// Check for node secret
	if isCertificateValidationRequested(combinedNodeEnvs, string(instance.GetDriverType())) {
		err = checkCertSecret(ctx, instance, r, driverConfig, "node", log)
		if err != nil {
			return err
		}
	}
	if len(instance.GetDriver().StorageClass) > 0 {
		log.Info("Warning: Creation of storage class via operator is deprecated")
	}
	if len(instance.GetDriver().SnapshotClass) > 0 {
		log.Info("Warning: Creation of snapshot class via operator is deprecated")
	}
	return nil
}

func checkIfCredentialsSecretExists(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI,
	driverConfig *ctrlconfig.Config, driverContainerType string, log logr.Logger) error {
	driver := instance.GetDriver()
	// assumption is that username and password come from the same secret
	// so we just check username
	if instance.GetDriverType() == csiv1.Unity && (driverConfig.ConfigVersion != "v1") {
		log.Info("Unity_creds secret is different from unity v1.2.* onwards")
		return nil
	}
	if instance.GetDriverType() == csiv1.PowerStore && driverConfig.ConfigVersion != "v1" && driverConfig.ConfigVersion != "v2" {
		log.Info("Since version v1.3.0 PowerStore driver expects config to be placed into secret and mounted to the container")
		return nil
	}
	if instance.GetDriverType() == csiv1.Isilon && driverConfig.ConfigVersion != "v1" && driverConfig.ConfigVersion != "v2" &&
		driverConfig.ConfigVersion != "v3" && driverConfig.ConfigVersion != "v4" {
		log.Info("Since version v1.5.0 PowerScale driver expects config to be placed into secret and mounted to the container")
		return nil
	}
	if instance.GetDriverType() == csiv1.VXFlexOS && driverConfig.ConfigVersion != "v2" && driverConfig.ConfigVersion != "v3" {
		log.Info("From version v1.4.0 PowerFlex driver expects config file to be created into secret and mounted to the container")
		return nil
	}
	credentialsEnvName := instance.GetUserEnvName()
	credentialsSecretName := ""
	envs := make([]corev1.EnvVar, 0)
	if driverContainerType == "controller" {
		envs = driverConfig.GetControllerEnvs()
	} else {
		envs = driverConfig.GetNodeEnvs()
	}
	credentialEnvValue, err := getEnvVar(credentialsEnvName, envs)
	if err != nil {
		log.Error(err, "Internal error: credential details missing from driver configuration")
		return err
	}
	if credentialEnvValue.ValueFrom != nil {
		credentialsSecretName = credentialEnvValue.ValueFrom.SecretKeyRef.LocalObjectReference.Name
	}
	log.Info(fmt.Sprintf("Default secret name: %s", credentialsSecretName))
	// The user provided secret takes priority over the default one
	if driver.AuthSecret != "" {
		credentialsSecretName = driver.AuthSecret
		log.Info(fmt.Sprintf("User specified secret name: %s", credentialsSecretName))
	}
	found := &corev1.Secret{}
	err = r.GetClient().Get(ctx, types.NamespacedName{Name: credentialsSecretName,
		Namespace: instance.GetNamespace()}, found)
	if err != nil && errors.IsNotFound(err) {
		return fmt.Errorf("failed to find secret: [%s] for connecting to the API endpoint", credentialsSecretName)
	} else if err != nil {
		log.Error(err, "Failed to query for secret. Warning - the driver pod may not start")
	} else {
		secretData := found.Data
		if _, found := secretData["username"]; !found {
			return fmt.Errorf("username key not found in secret")
		}
		if _, found := secretData["password"]; !found {
			return fmt.Errorf("password key not found in secret")
		}
		// Code only for PowerMax
		if driverContainerType == "node" {
			if instance.GetDriverType() == csiv1.PowerMax && driverConfig.ConfigVersion != "v1" {
				// We need to check the user specificed envs and not just in the default ones
				mergedEnvs := mergeEnvironmentVars(envs, instance.GetDriver().Common.Envs)
				mergedEnvs = mergeEnvironmentVars(envs, instance.GetDriver().Node.Envs)
				iscsiCHAPEnvName := "X_CSI_POWERMAX_ISCSI_ENABLE_CHAP"
				iscsiCHAPEnv, err := getEnvVar(iscsiCHAPEnvName, mergedEnvs)
				if err == nil {
					if strings.Compare(strings.ToUpper(iscsiCHAPEnv.Value), "TRUE") == 0 {
						if _, found := secretData["chapsecret"]; !found {
							return fmt.Errorf("chapsecret key not found in secret")
						}
					}
				}
			}
		}
	}
	return nil
}

// validateUserSpec - Validate the user specified environment variables
func validateUserEnv(driverConfig *ctrlconfig.Config, userSpecifiedEnvs []corev1.EnvVar, driverContainerType string) error {
	err := ensureMandatoryEnvsSpecified(driverConfig, userSpecifiedEnvs, driverContainerType)
	if err != nil {
		return err
	}
	for _, env := range userSpecifiedEnvs {
		isValid, err := driverConfig.ValidateEnvironmentVarType(env)
		if err != nil {
			return err
		}
		if !isValid {
			return fmt.Errorf("invalid value specified for the environment variable for %s",
				driverContainerType)
		}
	}
	return nil
}

// checkCertSecret - Checks if the secret for cert is present in the specified namespace
func checkCertSecret(ctx context.Context, instance csiv1.CSIDriver, r ReconcileCSI, driverConfig *ctrlconfig.Config,
	driverContainerType string, log logr.Logger) error {
	certVolume := instance.GetCertVolumeName()
	if certVolume == "" {
		return nil
	}
	volumes := make([]corev1.Volume, 0)
	if driverContainerType == "controller" {
		volumes = driverConfig.GetControllerVolumes()
	} else if driverContainerType == "node" {
		volumes = driverConfig.GetNodeVolumes()
	} else {
		return fmt.Errorf("invalid container type specified: %s", driverContainerType)
	}
	certName := ""
	for _, volume := range volumes {
		if volume.Name == certVolume {
			certName = volume.Secret.SecretName
			break
		}
	}
	if certName == "" {
		log.Info(fmt.Sprintf(
			"Volume - %s not found in config. Secret name for validation of certs not found. Continuing",
			certVolume))
		return nil
	}
	found := &corev1.Secret{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: certName,
		Namespace: instance.GetNamespace()}, found)
	if err != nil && errors.IsNotFound(err) {
		return fmt.Errorf("failed to find secret %s and certificate validation is requested", certName)
	} else if err != nil {
		log.Error(err, "Failed to query for secret. Warning - the controller pod may not start")
	}
	return nil
}

// isCertificateValidationRequested - Checks if validation of certificate is requested
func isCertificateValidationRequested(envs []corev1.EnvVar, driverName string) bool {
	envName := fmt.Sprintf("X_CSI_%s", strings.ToUpper(driverName))
	envNameSuffix := ""
	if driverName == "powermax" {
		envNameSuffix = "SKIP_CERTIFICATE_VALIDATION"
	} else {
		envNameSuffix = "INSECURE"
	}
	envName = fmt.Sprintf("%s_%s", envName, envNameSuffix)
	envValue, err := getEnvValue(envName, envs)
	if err != nil {
		return false
	}
	if strings.ToLower(envValue) == "false" {
		return true
	}
	return false
}

func ensureMandatoryEnvsSpecified(driverConfig *ctrlconfig.Config, userSpecifiedEnvs []corev1.EnvVar, driverContainerType string) error {
	mandatoryEnvNames := driverConfig.GetMandatoryEnvNames(driverContainerType)
	for _, mandatoryEnvName := range mandatoryEnvNames {
		present, value := isEnvPresent(mandatoryEnvName, userSpecifiedEnvs)
		if !present {
			return fmt.Errorf("mandatory Env - %s not specified in user spec", mandatoryEnvName)
		}
		if value == "" {
			return fmt.Errorf("value for mandatory Env - %s not specified in user spec", mandatoryEnvName)
		}
	}
	return nil
}
