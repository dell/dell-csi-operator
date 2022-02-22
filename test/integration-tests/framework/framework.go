package framework

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"

	betaSnapClient "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Constants
const (
	Before = "before"
	After  = "after"
)

// Global framework.
var Global *Framework

// Framework handles communication with the kube cluster in e2e tests.
type Framework struct {
	KubeClient         kubernetes.Interface
	KubeClient2        kubernetes.Clientset
	SnapshotClientBeta betaSnapClient.Clientset
	RestClient         rest.RESTClient
	Client             client.Client
	Namespace          string
	OperatorNamespace  string
	SkipTeardown       bool
	RunID              string
	Phase              string
}

// Setup sets up a test framework and initialises framework.Global.
func Setup() error {
	namespace := flag.String("namespace", "default", "Integration test namespace")
	operatorNamespace := flag.String("operatorNamespace", "", "Local test run mimicks prod environments")
	skipTeardown := flag.Bool("skipteardown", false, "Skips tearing down instances created by the tests")
	runid := flag.String("runid", "test-"+generateRandomID(3), "Optional string that will be used to uniquely identify this test run.")
	kubeconfig := os.Getenv("KUBECONFIG")
	if home := homeDir(); home != "" {
		kubeconfig = *(flag.String("kubeconf", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file"))
	} else {
		kubeconfig = *(flag.String("kubeconf", "", "absolute path to the kubeconfig file"))
	}
	flag.Parse()

	if *operatorNamespace == "" {
		operatorNamespace = namespace
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	restClient, err := GetRestClientForDellStorage(kubeconfig)
	if err != nil {
		return err
	}

	betaSnapClient, err := betaSnapClient.NewForConfig(cfg)
	if err != nil {
		return err
	}

	Global = &Framework{
		KubeClient:         kubeClient,
		KubeClient2:        *kubeClient,
		RestClient:         *restClient,
		SnapshotClientBeta: *betaSnapClient,
		Namespace:          *namespace,
		OperatorNamespace:  *operatorNamespace,
		SkipTeardown:       *skipTeardown,
		RunID:              *runid,
	}

	return nil
}

// GetRestClientForDellStorage - returns the rest client for storage
func GetRestClientForDellStorage(kubeconfig string) (*rest.RESTClient, error) {
	var gv = schema.GroupVersion{Group: "storage.dell.com", Version: "v1"}
	return GetRestClient(&gv, kubeconfig)
}

// GetRestClient - returns rest client
func GetRestClient(schemaObj *schema.GroupVersion, kubeconfig string) (*rest.RESTClient, error) {
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

// Teardown shuts down the test framework and cleans up.
func Teardown() error {
	// TODO: wait for all cr deleted.
	Global = nil
	return nil
}

func generateRandomID(n int) string {
	rand.Seed(time.Now().Unix())
	var letter = []rune("abcdefghijklmnopqrstuvwxyz")

	id := make([]rune, n)
	for i := range id {
		id[i] = letter[rand.Intn(len(letter))]
	}
	return string(id)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
