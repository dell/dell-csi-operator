package secrets

import (
	"context"
	"fmt"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/resources"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// New - Returns a secret object
func New(instance csiv1.CSIDriver, secretName string) *corev1.Secret {
	//var driver *csiv1.Driver = instance.GetDriver()
	//driverType := strings.ToLower(instance.GetDriverTypeMeta().Kind)
	driverNamespace := instance.GetNamespace()

	data := make(map[string][]byte)
	//data ["username"] = []byte(driver.UnisphereUserName)
	//data["password"] = []byte(driver.UnispherePassword)

	return &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:            secretName,
			Namespace:       driverNamespace,
			OwnerReferences: resources.GetOwnerReferences(instance),
		},
		Data: data,
	}
}

// NewTLS - Returns a new TLS secret object
func NewTLS(instance csiv1.CSIDriver, secretName string) *corev1.Secret {
	//var driver *csiv1.Driver = instance.GetDriver()
	data := make(map[string][]byte)
	//@TODO get TLS from driver spec
	//data ["TLS"] = []byte(driver.TLSCert)
	//data["password"] = []byte(driver.UnispherePassword)
	driverNamespace := instance.GetNamespace()
	return &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:            secretName,
			Namespace:       driverNamespace,
			OwnerReferences: resources.GetOwnerReferences(instance),
		},
		Data: data,
	}
}

// SyncSecret - Syncs a secret
func SyncSecret(ctx context.Context, secret *corev1.Secret, client crclient.Client, reqLogger logr.Logger) error {
	found := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Namespace", secret.Namespace, "Name", secret.Name)
		err = client.Create(ctx, secret)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		reqLogger.Info("Unknown error.", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Updating Secret", "Name:", secret.Name)
		err = client.Update(ctx, secret)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSecret - Returns a secret object
func GetSecret(ctx context.Context, name, namespace string, client crclient.Client, reqLogger logr.Logger) (*corev1.Secret, error) {
	found := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return nil, fmt.Errorf("no secrets found or error: %v", err)
	}
	return found, nil
}

// GetSecrets - Returns all secrets in a namespace
func GetSecrets(ctx context.Context, namespace string, client crclient.Client, reqLogger logr.Logger) ([]corev1.Secret, error) {
	slist := &corev1.SecretList{}
	err := client.List(ctx, slist, crclient.InNamespace(namespace))
	if err != nil && len(slist.Items) == 0 {
		return nil, fmt.Errorf("No secrets found or error: %v", err)
	}
	return slist.Items, nil
}
