package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	csiv1 "github.com/dell/dell-csi-operator/api/v1"
	controllerutils "github.com/dell/dell-csi-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

type mockTyper struct {
	gvk *schema.GroupVersionKind
	err error
}

func (t *mockTyper) ObjectKinds(obj runtime.Object) ([]schema.GroupVersionKind, bool, error) {
	if t.gvk == nil {
		return nil, false, t.err
	}
	return []schema.GroupVersionKind{*t.gvk}, false, t.err
}

func (t *mockTyper) Recognizes(_ schema.GroupVersionKind) bool {
	return false
}

// SideCar - structure of SideCar
type SideCar struct {
	Name  string
	Args  string
	Image string
}

var driverMap = make(map[string]string)
var commonEnv = make(map[string]string)
var controllerEnv = make(map[string]string)
var nodeEnv = make(map[string]string)
var sideCarMap = make(map[string]SideCar)
var storageClassMap = make(map[string]csiv1.StorageClass)
var snapshotClassMap = make(map[string]csiv1.SnapshotClass)

func readEnv() {
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "OPERATOR_ENV") {
			pair := strings.SplitN(e, "=", 2)
			tempkey := strings.TrimPrefix(pair[0], "OPERATOR_ENV_")
			if strings.HasPrefix(tempkey, "DRIVER") {
				key := strings.TrimPrefix(tempkey, "DRIVER_")
				driverMap[key] = pair[1]
			} else if strings.HasPrefix(tempkey, "COMMON") {
				key := strings.TrimPrefix(tempkey, "COMMON_")
				commonEnv[key] = pair[1]
			} else if strings.HasPrefix(tempkey, "NODE_") {
				key := strings.TrimPrefix(tempkey, "NODE_")
				nodeEnv[key] = pair[1]
			} else if strings.HasPrefix(tempkey, "CONTROLLER") {
				key := strings.TrimPrefix(tempkey, "CONTROLLER_")
				controllerEnv[key] = pair[1]
			} else if strings.HasPrefix(tempkey, "SIDECAR") {
				key := strings.TrimPrefix(tempkey, "SIDECAR")
				sideCarNumber := key[0:1]
				sideCarParamKey := key[2:]
				sideCar := SideCar{}
				if _, ok := sideCarMap[sideCarNumber]; ok {
					sideCar = sideCarMap[sideCarNumber]
				}
				if sideCarParamKey == "NAME" {
					sideCar.Name = pair[1]
				} else if sideCarParamKey == "ARGS" {
					sideCar.Args = pair[1]
				} else if sideCarParamKey == "IMAGE" {
					sideCar.Image = pair[1]
				}
				sideCarMap[sideCarNumber] = sideCar
			} else if strings.HasPrefix(tempkey, "STORAGECLASS") {
				key := strings.TrimPrefix(tempkey, "STORAGECLASS")
				storageClassNumber := key[0:1]
				storageClassKey := key[2:]
				storageClass := csiv1.StorageClass{}
				if _, ok := storageClassMap[storageClassNumber]; ok {
					storageClass = storageClassMap[storageClassNumber]
				}
				if storageClassKey == "NAME" {
					storageClass.Name = pair[1]
				} else if storageClassKey == "RECLAIM_POLICY" {
					if pair[1] == "Retain" {
						storageClass.ReclaimPolicy = corev1.PersistentVolumeReclaimRetain
					} else if pair[1] == "Delete" {
						storageClass.ReclaimPolicy = corev1.PersistentVolumeReclaimDelete
					} else if pair[1] == "Recycle" {
						storageClass.ReclaimPolicy = corev1.PersistentVolumeReclaimRecycle
					} else {
						panic(fmt.Errorf("invalid reclaim policy specified"))
					}
				} else if storageClassKey == "DEFAULTSC" {
					if strings.ToLower(pair[1]) == "true" {
						storageClass.DefaultSc = true
					} else if strings.ToLower(pair[1]) == "false" {
						storageClass.DefaultSc = false
					} else {
						panic(fmt.Errorf("invalid value specified for default sc for storage class"))
					}
				} else if storageClassKey == "PARAMETERS" {
					parameterMap := make(map[string]string)
					parameterList := strings.TrimSuffix(pair[1], "]")
					parameterList = strings.TrimPrefix(parameterList, "[")
					args := strings.Split(parameterList, ",")
					for i := range args {
						tempArg := strings.TrimSuffix(args[i], "\"")
						tempArg = strings.TrimPrefix(tempArg, "\"")
						pv := strings.Split(tempArg, "=")
						parameterMap[pv[0]] = pv[1]
					}
					storageClass.Parameters = parameterMap
				}
				storageClassMap[storageClassNumber] = storageClass
			} else if strings.HasPrefix(tempkey, "VOLUMESNAPSHOTCLASS") {
				key := strings.TrimPrefix(tempkey, "VOLUMESNAPSHOTCLASS")
				snapshotClassNumber := key[0:1]
				snapshotClassKey := key[2:]
				snapshotClass := csiv1.SnapshotClass{}
				if _, ok := snapshotClassMap[snapshotClassNumber]; ok {
					snapshotClass = snapshotClassMap[snapshotClassNumber]
				}
				if snapshotClassKey == "NAME" {
					snapshotClass.Name = pair[1]
				} else if snapshotClassKey == "PARAMETERS" {
					parameterMap := make(map[string]string)
					parameterList := strings.TrimSuffix(pair[1], "]")
					parameterList = strings.TrimPrefix(parameterList, "[")
					args := strings.Split(parameterList, ",")
					for i := range args {
						tempArg := strings.TrimSuffix(args[i], "\"")
						tempArg = strings.TrimPrefix(tempArg, "\"")
						pv := strings.Split(tempArg, "=")
						parameterMap[pv[0]] = pv[1]
					}
					snapshotClass.Parameters = parameterMap
				}
				snapshotClassMap[snapshotClassNumber] = snapshotClass
			}
		}
	}
}

func getDriver() csiv1.Driver {
	readEnv()
	commonEnvs := make([]corev1.EnvVar, 0)
	for k, v := range commonEnv {
		env := corev1.EnvVar{
			Name:  k,
			Value: v,
		}
		commonEnvs = append(commonEnvs, env)
	}
	controllerEnvs := make([]corev1.EnvVar, 0)
	for k, v := range controllerEnv {
		env := corev1.EnvVar{
			Name:  k,
			Value: v,
		}
		controllerEnvs = append(controllerEnvs, env)
	}
	nodeEnvs := make([]corev1.EnvVar, 0)
	for k, v := range nodeEnv {
		env := corev1.EnvVar{
			Name:  k,
			Value: v,
		}
		nodeEnvs = append(nodeEnvs, env)
	}
	sideCars := make([]csiv1.ContainerTemplate, 0)
	for _, v := range sideCarMap {
		name, _ := controllerutils.GetSideCarTypeFromName(v.Name)
		sideCar := csiv1.ContainerTemplate{
			Name: name,
		}
		if v.Image != "" {
			sideCar.Image = v.Image
		}
		if v.Args != "" {
			argList := strings.TrimSuffix(v.Args, "]")
			argList = strings.TrimPrefix(argList, "[")
			args := strings.Split(argList, ",")
			for i := range args {
				tempArg := strings.TrimSuffix(args[i], "\"")
				tempArg = strings.TrimPrefix(tempArg, "\"")
				args[i] = tempArg
			}
			sideCar.Args = args
		}
		sideCars = append(sideCars, sideCar)
	}
	storageClasses := make([]csiv1.StorageClass, 0)
	for _, v := range storageClassMap {
		storageClasses = append(storageClasses, v)
	}
	snapshotClasses := make([]csiv1.SnapshotClass, 0)
	for _, v := range snapshotClassMap {
		snapshotClasses = append(snapshotClasses, v)
	}
	driver := csiv1.Driver{
		ConfigVersion: driverMap["CONFIG_VERSION"],
		Replicas:      2,
		Common: csiv1.ContainerTemplate{
			Image: driverMap["IMAGE"],
			Envs:  commonEnvs,
		},
		/*		Controller: csiv1.ContainerTemplate{
					Envs: controllerEnvs,
				},
				Node: csiv1.ContainerTemplate{
					Envs: nodeEnvs,
				}, */
		SideCars:      sideCars,
		StorageClass:  storageClasses,
		SnapshotClass: snapshotClasses,
	}
	return driver
}

func getDriverObj(driver csiv1.Driver) (runtime.Object, error) {
	switch driverMap["TYPE"] {
	case "csi-powermax":
		csiPowerMax := csiv1.CSIPowerMax{}
		csiPowerMax.APIVersion = "storage.dell.com/v1"
		csiPowerMax.Namespace = driverMap["NAMESPACE"]
		csiPowerMax.Name = driverMap["NAME"]
		csiPowerMax.Kind = "CSIPowerMax"
		csiPowerMax.Spec.Driver = driver
		return &csiPowerMax, nil
	case "csi-unity":
		csiUnity := csiv1.CSIUnity{}
		csiUnity.APIVersion = "storage.dell.com/v1"
		csiUnity.Namespace = driverMap["NAMESPACE"]
		csiUnity.Name = driverMap["NAME"]
		csiUnity.Kind = "CSIUnity"
		csiUnity.Spec.Driver = driver
		return &csiUnity, nil
	case "csi-isilon":
		csiIsilon := csiv1.CSIIsilon{}
		csiIsilon.APIVersion = "storage.dell.com/v1"
		csiIsilon.Namespace = driverMap["NAMESPACE"]
		csiIsilon.Name = driverMap["NAME"]
		csiIsilon.Kind = "CSIIsilon"
		csiIsilon.Spec.Driver = driver
		return &csiIsilon, nil
	case "csi-vxflexos":
		csiVxFlexOS := csiv1.CSIVXFlexOS{}
		csiVxFlexOS.APIVersion = "storage.dell.com/v1"
		csiVxFlexOS.Namespace = driverMap["NAMESPACE"]
		csiVxFlexOS.Name = driverMap["NAME"]
		csiVxFlexOS.Kind = "CSIVXFlexOS"
		csiVxFlexOS.Spec.Driver = driver
		return &csiVxFlexOS, nil
	}
	return &csiv1.CSIPowerMax{}, fmt.Errorf("unknown driver-type specified")
}

func main() {
	driver := getDriver()
	driverObj, err := getDriverObj(driver)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	typer := &mockTyper{
		gvk: &schema.GroupVersionKind{
			Kind:    "CSIPowerMax",
			Group:   "storage.dell.com",
			Version: "v1",
		},
	}
	serializer := k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory,
		nil,
		typer,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	outfilename := driverMap["MANIFEST"]
	outFile, err := os.Create(filepath.Clean(outfilename))
	if err != nil {
		fmt.Println("Failed to create yaml file")
	}
	err = serializer.Encode(driverObj, outFile)
	if err != nil {
		fmt.Println("Failed to encode")
		panic(err)
	}
	fmt.Printf("Driver manifest: %s created successfully\n", outfilename)
}
