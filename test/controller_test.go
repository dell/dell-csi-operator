package controller_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	v1 "github.com/dell/dell-csi-operator/api/v1"
	"github.com/dell/dell-csi-operator/controllers"
	operatorconfig "github.com/dell/dell-csi-operator/pkg/config"
	"github.com/dell/dell-csi-operator/pkg/resources/statefulset"
	"github.com/dell/dell-csi-operator/pkg/utils"
	snaps "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

type ControllerTestSuite struct {
	suite.Suite
	configDir  string
	configFile string
	drivers    []Driver
}

type Driver struct {
	driverType v1.DriverType
	reconciler utils.ReconcileCSI
	findCR     func(inObjects []runtime.Object) (string, string)
	copyTime   func(expObj runtime.Object, gotObj runtime.Object) error
	k8sVersion v1.K8sVersion
}

func (suite *ControllerTestSuite) SetupSuite() {
	// Add VolumeSnapshots
	if err := snaps.AddToScheme(scheme.Scheme); err != nil {
		suite.Fail("cannot add to scheme: ", err)
	}

	// Add everything else
	if err := v1.AddToScheme(scheme.Scheme); err != nil {
		suite.Fail("cannot add to scheme: ", err)
	}

	// Add CSIDriver
	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		suite.Fail("cannot add to scheme: ", err)
	}

	// Set statefulset sleep timeout to 0
	statefulset.SleepTime = 0 * time.Second
	suite.configFile = "config.yaml"
	suite.configDir = "../driverconfig"
	suite.drivers = []Driver{
		{
			driverType: v1.PowerMax,
			reconciler: &controllers.CSIPowerMaxReconciler{
				Log: ctrl.Log.WithName("controllers").WithName("CSIPowerMax"),
			},
			k8sVersion: "v125",
			findCR: func(inObjects []runtime.Object) (string, string) {
				for _, o := range inObjects {
					if cr, ok := o.(*v1.CSIPowerMax); ok {
						return cr.Name, cr.Namespace
					}
				}
				return "", ""
			},
			copyTime: func(expObj runtime.Object, gotObj runtime.Object) error {
				expPowerMax, ok := expObj.(*v1.CSIPowerMax)
				if !ok {
					return fmt.Errorf("can't convert object to CSIPowerMax")
				}
				gotPowerMax, ok := gotObj.(*v1.CSIPowerMax)
				if !ok {
					return fmt.Errorf("can't convert object to CSIPowerMax")
				}
				expPowerMax.Status.LastUpdate.Time.Time = gotPowerMax.Status.LastUpdate.Time.Time.Truncate(time.Second)
				expPowerMax.Status.DriverHash = gotPowerMax.Status.DriverHash
				return nil
			},
		},
		{
			driverType: v1.PowerStore,
			reconciler: &controllers.CSIPowerStoreReconciler{
				Log: ctrl.Log.WithName("controllers").WithName("CSIPowerStore"),
			},
			k8sVersion: "v121",
			findCR: func(inObjects []runtime.Object) (string, string) {
				for _, o := range inObjects {
					if cr, ok := o.(*v1.CSIPowerStore); ok {
						return cr.Name, cr.Namespace
					}
				}
				return "", ""
			},
			copyTime: func(expObj runtime.Object, gotObj runtime.Object) error {
				expPowerStore, ok := expObj.(*v1.CSIPowerStore)
				if !ok {
					return fmt.Errorf("can't convert object to CSIPowerStore")
				}
				gotPowerStore, ok := gotObj.(*v1.CSIPowerStore)
				if !ok {
					return fmt.Errorf("can't convert object to CSIPowerStore")
				}
				expPowerStore.Status.LastUpdate.Time.Time = gotPowerStore.Status.LastUpdate.Time.Time.Truncate(time.Second)
				expPowerStore.Status.DriverHash = gotPowerStore.Status.DriverHash
				return nil
			},
		},
		{
			driverType: v1.VXFlexOS,
			reconciler: &controllers.CSIVXFlexOSReconciler{
				Log: ctrl.Log.WithName("controllers").WithName("CSIVXFlexOS"),
			},
			k8sVersion: "v125",
			findCR: func(inObjects []runtime.Object) (string, string) {
				for _, o := range inObjects {
					if cr, ok := o.(*v1.CSIVXFlexOS); ok {
						return cr.Name, cr.Namespace
					}
				}
				return "", ""
			},
			copyTime: func(expObj runtime.Object, gotObj runtime.Object) error {
				expCSIVXFlexOS, ok := expObj.(*v1.CSIVXFlexOS)
				if !ok {
					return fmt.Errorf("can't convert object to CSIVXFlexOS")
				}
				gotCSIVXFlexOS, ok := gotObj.(*v1.CSIVXFlexOS)
				if !ok {
					return fmt.Errorf("can't convert object to CSIVXFlexOS")
				}
				expCSIVXFlexOS.Status.LastUpdate.Time.Time = gotCSIVXFlexOS.Status.LastUpdate.Time.Time.Truncate(time.Second)
				expCSIVXFlexOS.Status.DriverHash = gotCSIVXFlexOS.Status.DriverHash
				return nil
			},
		},
		//{
		//	driverType: operatorconfig.Unity,
		//	reconciler: &csiunity.ReconcileCSIUnity{},
		//	findCR: func(inObjects []runtime.Object) (string, string) {
		//		for _, o := range inObjects {
		//			if cr, ok := o.(*v1.CSIUnity); ok {
		//				return cr.Name, cr.Namespace
		//			}
		//		}
		//		return "", ""
		//	},
		//	copyTime: func(expObj runtime.Object, gotObj runtime.Object) error {
		//		// expUnity, ok := expObj.(*v1.CSIUnity)
		//		// if !ok {
		//		// 	return fmt.Errorf("can't convert object to CSIUnity")
		//		// }
		//		// gotUnity, ok := gotObj.(*v1.CSIUnity)
		//		// if !ok {
		//		// 	return fmt.Errorf("can't convert object to CSIUnity")
		//		// }
		//		// expUnity.Status.Status.CreationTime.Time = gotUnity.Status.Status.CreationTime.Time.Truncate(time.Second)
		//		// expUnity.Status.Status.StartTime.Time = gotUnity.Status.Status.StartTime.Time.Truncate(time.Second)
		//		return nil
		//	},
		//},
		{
			driverType: v1.Isilon,
			reconciler: &controllers.CSIIsilonReconciler{
				Log: ctrl.Log.WithName("controllers").WithName("CSIIsilon"),
			},
			k8sVersion: "v125",
			findCR: func(inObjects []runtime.Object) (string, string) {
				for _, o := range inObjects {
					if cr, ok := o.(*v1.CSIIsilon); ok {
						return cr.Name, cr.Namespace
					}
				}
				return "", ""
			},
			copyTime: func(expObj runtime.Object, gotObj runtime.Object) error {
				expIsilon, ok := expObj.(*v1.CSIIsilon)
				if !ok {
					return fmt.Errorf("can't convert object to CSIIsilon")
				}
				gotIsilon, ok := gotObj.(*v1.CSIIsilon)
				if !ok {
					return fmt.Errorf("can't convert object to CSIIsilon")
				}
				expIsilon.Status.LastUpdate.Time.Time = gotIsilon.Status.LastUpdate.Time.Time.Truncate(time.Second)
				expIsilon.Status.DriverHash = gotIsilon.Status.DriverHash
				return nil
			},
		},
	}
}

func (suite *ControllerTestSuite) TestAllControllers() {
	for _, driver := range suite.drivers {
		suite.Run(string(driver.driverType), func() {
			folder := fmt.Sprintf("testdata/csi%s", string(driver.driverType))
			files, err := ioutil.ReadDir(folder)
			suite.NoError(err)

			for _, file := range files {
				if strings.HasPrefix(file.Name(), ".") {
					continue
				}
				if file.IsDir() {
					suite.Run(file.Name(), func() {
						suite.testDirectory(&driver, filepath.Join(folder, file.Name()))
					})
				}
			}
		})
	}
}

func (suite *ControllerTestSuite) TestAllControllersWithErrors() {
	for _, driver := range suite.drivers {
		suite.Run(string(driver.driverType), func() {
			folder := fmt.Sprintf("testdata/csi%s", string(driver.driverType))
			files, err := ioutil.ReadDir(folder)
			suite.NoError(err)

			for _, file := range files {
				if strings.HasPrefix(file.Name(), ".") {
					continue
				}
				if file.IsDir() {
					suite.Run(file.Name(), func() {
						suite.testDirectoryWithErrors(&driver, filepath.Join(folder, file.Name()))
					})
				}
			}
		})
	}
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (suite *ControllerTestSuite) testDirectory(driver *Driver, path string) {
	suite.T().Logf("processing directory %s", path)
	inObjects, outObjects := suite.parseDirectory(path)

	name, namespace := driver.findCR(inObjects)

	if name == "" && namespace == "" {
		suite.Failf("could not find CSIDrivers in input objects in %s", path)
		return
	}

	c, err := newFakeClient(inObjects, nil)
	if err != nil {
		suite.T().Fatal(err)
	}

	cfg := operatorconfig.Config{
		ConfigDirectory:      suite.configDir,
		ConfigFile:           suite.configFile,
		KubeAPIServerVersion: driver.k8sVersion,
		EnabledDrivers: []v1.DriverType{
			driver.driverType,
		},
		RetryCount: 1,
	}

	driver.reconciler.SetClient(c)
	driver.reconciler.SetScheme(scheme.Scheme)
	driver.reconciler.SetConfig(cfg)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}

	// Update the spec
	_, err = driver.reconciler.Reconcile(context.Background(), req)
	if err != nil {
		suite.T().Fatal(err)
	}

	// Sync the driver
	_, err = driver.reconciler.Reconcile(context.Background(), req)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.checkObjects(driver, c, outObjects)
}

func (suite *ControllerTestSuite) testDirectoryWithErrors(driver *Driver, path string) {
	suite.T().Logf("processing directory %s", path)
	inObjects, outObjects := suite.parseDirectory(path)

	name, namespace := driver.findCR(inObjects)

	if name == "" && namespace == "" {
		suite.Failf("could not find CSIDrivers in input objects in %s", path)
		return
	}

	c, err := newFakeClient(inObjects, newStableErrorInjector(suite.T()))
	if err != nil {
		suite.T().Fatal(err)
	}

	cfg := operatorconfig.Config{
		ConfigDirectory: suite.configDir,
		ConfigFile:      suite.configFile,
		EnabledDrivers: []v1.DriverType{
			driver.driverType,
		},
		KubeAPIServerVersion: driver.k8sVersion,
		RetryCount:           100,
	}

	driver.reconciler.SetClient(c)
	driver.reconciler.SetScheme(scheme.Scheme)
	driver.reconciler.SetConfig(cfg)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}

	// Update spec Reconcile call
	for attempts := 0; attempts < 100; attempts++ {
		_, err = driver.reconciler.Reconcile(context.Background(), req)
		if err != nil {
			suite.T().Logf("update spec %d failed with: %s", attempts, err)
		}
	}
	if err != nil {
		suite.T().Errorf("unexpected reconcile error: %s", err)
	}

	suite.checkObjects(driver, c, outObjects)
}

// parseFile parses *one* object out of a YAML file and returns it
func parseFile(path string) (runtime.Object, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	obj, _, err := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode(content, nil, nil)
	return obj, err
}

// parseDirectory parses *one* object out of each yaml file in a directory returns array of them
func (suite *ControllerTestSuite) parseDirectory(path string) (inObjects, outObjects []runtime.Object) {
	inObjects = []runtime.Object{}
	outObjects = []runtime.Object{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		suite.T().Fatal(err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if file.IsDir() {
			suite.T().Errorf("subdirectory %s is not allowed in %s", file.Name(), path)
		}

		switch {
		case strings.HasPrefix(file.Name(), "in-"):
			obj, err := parseFile(filepath.Join(path, file.Name()))
			if err != nil {
				suite.T().Error(err)
				continue
			}
			inObjects = append(inObjects, obj)

		case strings.HasPrefix(file.Name(), "out-"):
			obj, err := parseFile(filepath.Join(path, file.Name()))
			if err != nil {
				suite.T().Error(err)
				continue
			}
			outObjects = append(outObjects, obj)

		case strings.HasSuffix(file.Name(), ".txt"):
		case strings.HasSuffix(file.Name(), ".md"):
			// Ignore text files
		default:
			suite.T().Errorf("file %s/%s has unknown prefix", path, file.Name())
		}
	}
	return
}

func (suite *ControllerTestSuite) checkObjects(driver *Driver, client *fakeClient, expectedObjects []runtime.Object) {
	// Get list of output objects in the same way as `client` has them for easy comparison,
	// i.e. using dummy fakeClient.
	expectedClient, err := newFakeClient(expectedObjects, nil)
	suite.NoError(err)

	driverType := fmt.Sprint("csi", driver.driverType)
	// Compare the objects
	expectedObjs := expectedClient.objects
	gotObjs := client.objects
	for k, gotObj := range gotObjs {
		expectedObj, found := expectedObjs[k]
		if !found {
			suite.T().Errorf("unexpected object %+v created:\n%s", k, objectYAML(gotObj))
			continue
		}
		suite.T().Logf("found expected object %+v", k)

		// We can't really set precise time in the out- manifest, so we just copy it
		if strings.ToLower(k.Kind) == driverType {
			err := driver.copyTime(expectedObj, gotObj)
			if err != nil {
				suite.T().Error(err)
			}
		} else if strings.ToLower(k.Kind) == "csidriver" {
			// ignore owner reference links
			expDriver, ok := expectedObj.(*storagev1.CSIDriver)
			if !ok {
				suite.T().Error("can't convert object to CSIDriver")
			}
			gotDriver, ok := gotObj.(*storagev1.CSIDriver)
			if !ok {
				suite.T().Error("can't convert object to CSIDriver")
			}
			if expDriver != nil && gotDriver != nil {
				expDriver.SetOwnerReferences(gotDriver.GetOwnerReferences())
			}
		}

		// gotObj does not have TypeMeta. Fill it.
		buf := new(bytes.Buffer)
		err := serializer.
			NewCodecFactory(scheme.Scheme).
			LegacyCodec(
				corev1.SchemeGroupVersion,
				appsv1.SchemeGroupVersion,
				storagev1.SchemeGroupVersion,
				rbacv1.SchemeGroupVersion,
				v1.GroupVersion,
				snaps.SchemeGroupVersion,
				v1beta1.SchemeGroupVersion,
			).Encode(gotObj, buf)

		if err != nil {
			suite.T().Error(err)
			continue
		}
		gotObj, _, err = serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode(buf.Bytes(), nil, nil)
		if err != nil {
			suite.T().Error(err)
			continue
		}

		if !equality.Semantic.DeepEqual(expectedObj, gotObj) {
			suite.T().Errorf("unexpected object %+v content:\n%s", k, diff.ObjectDiff(expectedObj, gotObj))
		}
		// Delete processed objects to keep track of the unprocessed ones.
		delete(expectedObjs, k)
	}
	// Unprocessed objects.
	for k := range expectedObjs {
		suite.T().Errorf("expected object %+v but none was created", k)
	}
}

// objectYAML prints YAML of an object
func objectYAML(obj runtime.Object) string {
	objString := ""
	j, err := json.Marshal(obj)
	if err != nil {
		objString = err.Error()
	} else {
		y, err := yaml.JSONToYAML(j)
		if err != nil {
			objString = err.Error()
		} else {
			objString = string(y)
		}
	}
	return objString
}

// stableErrorInjector fails every call exactly once.
// It uses call stack to check what calls it has already failed.
type stableErrorInjector struct {
	t     *testing.T
	calls sets.String
}

func newStableErrorInjector(t *testing.T) *stableErrorInjector {
	return &stableErrorInjector{
		t:     t,
		calls: sets.NewString(),
	}
}
func (s *stableErrorInjector) shouldFail(method string, object runtime.Object) error {
	_, file, line, _ := goruntime.Caller(2)
	callID := fmt.Sprintf("%s:%d", file, line)
	if s.calls.Has(callID) {
		return nil
	}
	s.calls.Insert(callID)
	return fmt.Errorf("call %s failed", callID)
}

// ReverseProxyControllerTestSuite - suite for testing reverse-proxy controller
type ReverseProxyControllerTestSuite struct {
	suite.Suite
	revProxyReconciler *controllers.CSIPowerMaxRevProxyReconciler
	k8sVersion         string
	name               string
}

func (revSuite *ReverseProxyControllerTestSuite) SetupSuite() {
	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		revSuite.Fail("cannot add to scheme", err)
	}

	if err := v1.AddToScheme(scheme.Scheme); err != nil {
		revSuite.Fail("cannot add to scheme", err)
	}
	revSuite.revProxyReconciler = &controllers.CSIPowerMaxRevProxyReconciler{
		Log: ctrl.Log.WithName("controllers").WithName("CSIPowerMaxRevProxy"),
	}
	revSuite.k8sVersion = "v122"
	revSuite.name = "powermaxrevproxy"
}

func (revSuite *ReverseProxyControllerTestSuite) TestController() {
	revSuite.Run("powermax_reverseproxy", func() {
		folder := fmt.Sprintf("testdata/csi%s", revSuite.name)
		files, err := ioutil.ReadDir(folder)
		revSuite.NoError(err)

		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			if file.IsDir() {
				revSuite.testDirectory(filepath.Join(folder, file.Name()))
			}
		}
	})
}

func (revSuite *ReverseProxyControllerTestSuite) TestControllerWithError() {
	revSuite.Run("powermax_reverseproxy", func() {
		folder := fmt.Sprintf("testdata/csi%s", revSuite.name)
		files, err := ioutil.ReadDir(folder)
		revSuite.NoError(err)

		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			if file.IsDir() {
				revSuite.testDirectoryWithErrors(filepath.Join(folder, file.Name()))
			}
		}
	})
}

func (revSuite *ReverseProxyControllerTestSuite) testDirectory(path string) {
	revSuite.T().Logf("processing directory %s", path)
	inObjects, outObjects := revSuite.parseDirectory(path)

	name, namespace := revSuite.findCR(inObjects)
	if name == "" || namespace == "" {
		revSuite.Failf("couldn't find CSIPowerMaxRevProxy in input objects in %s", path)
		return
	}

	fakeClient, err := newFakeClient(inObjects, nil)
	if err != nil {
		revSuite.T().Fatal(err)
	}

	revSuite.revProxyReconciler.SetClient(fakeClient).SetScheme(scheme.Scheme)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	// Update revproxy spec
	_, err = revSuite.revProxyReconciler.Reconcile(context.Background(), req)
	if err != nil {
		revSuite.T().Fatal(err)
	}

	// Sync revproxy
	_, err = revSuite.revProxyReconciler.Reconcile(context.Background(), req)
	if err != nil {
		revSuite.T().Fatal(err)
	}

	revSuite.checkObjects(fakeClient, outObjects)
}

func (revSuite *ReverseProxyControllerTestSuite) testDirectoryWithErrors(path string) {
	revSuite.T().Logf("processing directory %s", path)
	inObjects, outObjects := revSuite.parseDirectory(path)

	name, namespace := revSuite.findCR(inObjects)
	if name == "" || namespace == "" {
		revSuite.Failf("couldn't find CSIPowerMaxRevProxy in input objects in %s", path)
		return
	}

	fakeClient, err := newFakeClient(inObjects, newStableErrorInjector(revSuite.T()))
	if err != nil {
		revSuite.T().Fatal(err)
	}

	revSuite.revProxyReconciler.SetClient(fakeClient)
	revSuite.revProxyReconciler.SetScheme(scheme.Scheme)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	// Update revproxy spec
	for attempt := 0; attempt < 100; attempt++ {
		_, err = revSuite.revProxyReconciler.Reconcile(context.Background(), req)
		if err != nil {
			revSuite.T().Logf("update spec %d failed with %s", attempt, err.Error())
		}
	}
	if err != nil {
		revSuite.T().Errorf("unexpected reconcile error: %s", err.Error())
	}

	revSuite.checkObjects(fakeClient, outObjects)
}

// parseDirectory parses *one* object out of each yaml file in a directory returns array of them
func (revSuite *ReverseProxyControllerTestSuite) parseDirectory(path string) (inObjects, outObjects []runtime.Object) {
	inObjects = []runtime.Object{}
	outObjects = []runtime.Object{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		revSuite.T().Fatal(err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if file.IsDir() {
			revSuite.T().Errorf("subdirectory %s is not allowed in %s", file.Name(), path)
		}

		switch {
		case strings.HasPrefix(file.Name(), "in-"):
			obj, err := parseFile(filepath.Join(path, file.Name()))
			if err != nil {
				revSuite.T().Error(err)
				continue
			}
			inObjects = append(inObjects, obj)

		case strings.HasPrefix(file.Name(), "out-"):
			obj, err := parseFile(filepath.Join(path, file.Name()))
			if err != nil {
				revSuite.T().Error(err)
				continue
			}
			outObjects = append(outObjects, obj)

		case strings.HasSuffix(file.Name(), ".txt"):
		case strings.HasSuffix(file.Name(), ".md"):
			// Ignore text files
		default:
			revSuite.T().Errorf("file %s/%s has unknown prefix", path, file.Name())
		}
	}
	return
}

func (revSuite *ReverseProxyControllerTestSuite) checkObjects(client *fakeClient, expectedObjects []runtime.Object) {
	// Get list of output objects in the same way as `client` has them for easy comparison,
	// i.e. using dummy fakeClient.
	expectedClient, err := newFakeClient(expectedObjects, nil)
	revSuite.NoError(err)
	name := fmt.Sprint("csi", revSuite.name)
	// Compare the objects
	expectedObjs := expectedClient.objects
	gotObjs := client.objects
	for k, gotObj := range gotObjs {
		expectedObj, found := expectedObjs[k]
		if !found {
			revSuite.T().Errorf("unexpected object %+v created:\n%s", k, objectYAML(gotObj))
			continue
		}
		revSuite.T().Logf("found expected object %+v", k)

		// We can't really set precise time in the out- manifest, so we just copy it
		if strings.ToLower(k.Kind) == name {
			err := revSuite.copyTime(expectedObj, gotObj)
			if err != nil {
				revSuite.T().Error(err)
			}
		}

		// gotObj does not have TypeMeta. Fill it.
		buf := new(bytes.Buffer)
		err = serializer.
			NewCodecFactory(scheme.Scheme).
			LegacyCodec(
				corev1.SchemeGroupVersion,
				appsv1.SchemeGroupVersion,
				storagev1.SchemeGroupVersion,
				rbacv1.SchemeGroupVersion,
				v1.GroupVersion,
				snaps.SchemeGroupVersion,
				v1beta1.SchemeGroupVersion,
			).Encode(gotObj, buf)

		if err != nil {
			revSuite.T().Error(err)
			continue
		}
		gotObj, _, err = serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode(buf.Bytes(), nil, nil)
		if err != nil {
			revSuite.T().Error(err)
			continue
		}

		if !equality.Semantic.DeepEqual(expectedObj, gotObj) {
			revSuite.T().Errorf("unexpected object %+v content:\n%s", k, diff.ObjectDiff(expectedObj, gotObj))
		}
		// Delete processed objects to keep track of the unprocessed ones.
		delete(expectedObjs, k)
	}
	// Unprocessed objects.
	for k := range expectedObjs {
		revSuite.T().Errorf("expected object %+v but none was created", k)
	}
}

func (revSuite *ReverseProxyControllerTestSuite) findCR(inObjects []runtime.Object) (string, string) {
	for _, o := range inObjects {
		if cr, ok := o.(*v1.CSIPowerMaxRevProxy); ok {
			return cr.Name, cr.Namespace
		}
	}
	return "", ""
}

func (revProxy *ReverseProxyControllerTestSuite) copyTime(expObj runtime.Object, gotObj runtime.Object) error {
	expPowerMax, ok := expObj.(*v1.CSIPowerMaxRevProxy)
	if !ok {
		return fmt.Errorf("can't convert object to CSIPowerMaxRevProxy")
	}
	gotPowerMax, ok := gotObj.(*v1.CSIPowerMaxRevProxy)
	if !ok {
		return fmt.Errorf("can't convert object to CSIPowerMaxRevProxy")
	}
	expPowerMax.Status.LastUpdate.Time.Time = gotPowerMax.Status.LastUpdate.Time.Time.Truncate(time.Second)
	return nil
}

func TestReverseProxyControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ReverseProxyControllerTestSuite))
}
