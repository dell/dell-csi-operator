package resources

import (
	"bytes"
	"fmt"
	"os"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/Jeffail/gabs"
	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/pkg/constants"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"context"
	"encoding/json"
	"time"
)

// GetKubeAPIServerVersion - Returns the Kubernetes API server version
func GetKubeAPIServerVersion() (*version.Info, error) {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	// Create the discoveryClient
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	sv, err := discoveryClient.ServerVersion()
	if err != nil {
		return nil, err
	}
	return sv, nil
}

// IsOpenshift - Returns a boolean which indicates if we are running in an OpenShift cluster
func IsOpenshift() (bool, error) {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	// Create the discoveryClient
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false, err
	}
	// Get a list of all API's on the cluster
	apiGroup, _, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	openshiftAPIGroup := "security.openshift.io"
	for i := 0; i < len(apiGroup); i++ {
		if apiGroup[i].Name == openshiftAPIGroup {
			// found the api
			return true, nil
		}
	}
	return false, nil
}

// GetOwnerReferences - Returns OwnerReferences for a specific driver
func GetOwnerReferences(driver csiv1.CSIDriver) []metav1.OwnerReference {
	meta := driver.GetDriverTypeMeta()
	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(driver, schema.GroupVersionKind{
			Group:   csiv1.GroupVersion.Group,
			Version: csiv1.GroupVersion.Version,
			Kind:    meta.Kind,
		}),
	}

	return ownerReferences
}

// GetDummyOwnerReferences - returns owner references
func GetDummyOwnerReferences(clusterRole *rbacv1.ClusterRole) []metav1.OwnerReference {
	meta := &clusterRole.TypeMeta
	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(clusterRole, schema.GroupVersionKind{
			Group:   meta.GroupVersionKind().Group,
			Version: meta.GroupVersionKind().Version,
			Kind:    meta.Kind,
		}),
	}
	return ownerReferences
}

// CreateContainerElement - Creates a generic container element for the given component of the given object
func CreateContainerElement(containerName csiv1.ImageType, image string, imagePullPolicy corev1.PullPolicy, args []string, envs []corev1.EnvVar,
	volumeMounts []corev1.VolumeMount, securityContext *corev1.SecurityContext, command []string) corev1.Container {

	if securityContext == nil {
		securityContext = &corev1.SecurityContext{}
	}
	resource := corev1.Container{
		Name:            string(containerName),
		Image:           image,
		ImagePullPolicy: imagePullPolicy,
		Env:             envs,
		Args:            args,
		//Resources: nil,
		TerminationMessagePath:   constants.TerminationMessagePath,
		TerminationMessagePolicy: constants.TerminationMessagePolicy,
		//Lifecycle:
		VolumeMounts:    volumeMounts,
		SecurityContext: securityContext,
	}

	if command != nil {
		resource.Command = command
	}

	return resource
}

// NewNamespace - Returns a pointer to namespace resource
func NewNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{},
	}
}

// CreateNamespace - Creates a namespace
func CreateNamespace(name string, client client.Client, reqLogger logr.Logger) (reconcile.Result, error) {
	found := &corev1.Namespace{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: name}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace", " Namespace", name)
		err = client.Create(context.TODO(), NewNamespace(name))
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// JSONPrettyPrint - Indent the json output
func JSONPrettyPrint(obj interface{}) string {
	s, err := json.Marshal(obj)
	if err != nil {
		return "Not a instance"
	}
	var out bytes.Buffer
	err = json.Indent(&out, s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return out.String()
}

// CreateCustomResource - Creates Snapshot custom resource
func CreateCustomResource(jsonData string, name, namespace string, resourceName string) (err error) {
	var gv = schema.GroupVersion{Group: "snapshot.storage.k8s.io", Version: "v1alpha1"}
	restClient, err := GetRestClient(&gv)
	if err != nil {
		fmt.Printf("restClient error: %v\n", err)
		return err
	}
	b, err := restClient.Post().Resource("volumesnapshotclasses").Name(name).Body([]byte(jsonData)).Do(context.TODO()).Raw()
	fmt.Println("---------------", string(b), err)

	return err
}

// GetRestClient - Returns a rest client to be used with k8s APIs
func GetRestClient(schemaObj *schema.GroupVersion) (*rest.RESTClient, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("CFG error:", err)
		return nil, err
	}
	cfg.GroupVersion = schemaObj
	cfg.APIPath = "/apis"
	cfg.ContentType = runtime.ContentTypeJSON
	cfg.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(&runtime.Scheme{})}
	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		fmt.Println("restClient error:", err)
		return nil, err
	}

	return restClient, nil
}

// WaitForCustomResourceToDelete - wait for the CR to be deleted
func WaitForCustomResourceToDelete(namespace string, resourceName string, restClient rest.RESTClient) (err error) {
	fmt.Printf("Waiting for resource '%s' to get deleted ...", resourceName)

	err = Retry(DefaultRetry, func() (bool, error) {
		items, err := getItemsArray(restClient, namespace, resourceName)
		if err != nil {
			return false, err
		}
		if len(items) == 0 {
			fmt.Printf(" resource '%s' deleted !\n", resourceName)
			return true, nil
		}
		return false, nil
	})

	return err
}

// DefaultRetry is the default backoff for e2e tests.
var DefaultRetry = wait.Backoff{
	Steps:    150,
	Duration: 4 * time.Second,
	Factor:   1.0,
	Jitter:   0.1,
}

// Retry executes the provided function repeatedly, retrying until the function
// returns done = true, errors, or exceeds the given timeout.
func Retry(backoff wait.Backoff, fn wait.ConditionFunc) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		done, err := fn()
		if err != nil {
			lastErr = err
		}
		return done, err
	})
	if err == wait.ErrWaitTimeout {
		if lastErr != nil {
			err = lastErr
		}
	}
	return err
}

func getItemsArray(restClient rest.RESTClient, namespace string, resourceType string) ([]*gabs.Container, error) {
	res, err := restClient.Get().Namespace(namespace).Resource(resourceType).DoRaw(context.TODO())
	if err != nil {
		return nil, err
	}
	parsedJSON, err := gabs.ParseJSON(res)
	if err != nil {
		return nil, err
	}
	items := parsedJSON.Path("items")
	if items == nil {
		return nil, fmt.Errorf("missing \"items\" path in response for %s:%s", namespace, resourceType)
	}
	return items.Children()
}

// IsStringInSlice - Checks if a given string is present in a slice
func IsStringInSlice(str string, slice []string) bool {
	for _, ele := range slice {
		if ele == str {
			return true
		}
	}
	return false
}
