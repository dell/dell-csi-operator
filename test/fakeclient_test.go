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
package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	snaps "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"

	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type errorInjector interface {
	shouldFail(method string, obj runtime.Object) error
}

type storageKey struct {
	Namespace string
	Name      string
	Kind      string
}

type fakeClient struct {
	objects       map[storageKey]runtime.Object
	errorInjector errorInjector
}

// blank assignment to verify that ReconcileCSIPowerMax implements client.Client
var _ client.Client = &fakeClient{}

func getKey(obj runtime.Object) (storageKey, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return storageKey{}, err
	}
	gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
	if err != nil {
		return storageKey{}, err
	}
	return storageKey{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
		Kind:      gvk.Kind,
	}, nil
}

func newFakeClient(initialObjects []runtime.Object, errorInjector errorInjector) (*fakeClient, error) {
	client := &fakeClient{
		objects:       map[storageKey]runtime.Object{},
		errorInjector: errorInjector,
	}

	for _, obj := range initialObjects {
		key, err := getKey(obj)
		if err != nil {
			return nil, err
		}
		client.objects[key] = obj
	}
	return client, nil
}

func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	if f.errorInjector != nil {
		if err := f.errorInjector.shouldFail("Get", obj); err != nil {
			return err
		}
	}

	gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}
	k := storageKey{
		Name:      key.Name,
		Namespace: key.Namespace,
		Kind:      gvk.Kind,
	}
	o, found := f.objects[k]
	if !found {
		gvr := schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		}
		return errors.NewNotFound(gvr, key.Name)
	}

	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

func (f *fakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if f.errorInjector != nil {
		if err := f.errorInjector.shouldFail("List", list); err != nil {
			return err
		}
	}
	switch list.(type) {
	case *storagev1.StorageClassList:
		return f.listStorageClasses(list.(*storagev1.StorageClassList))
	case *snaps.VolumeSnapshotClassList:
		return f.listVolumeSnapshots(list.(*snaps.VolumeSnapshotClassList))
	default:
		return fmt.Errorf("Unknown type: %s", reflect.TypeOf(list))
	}
}

func (f *fakeClient) listStorageClasses(list *storagev1.StorageClassList) error {
	for k, v := range f.objects {
		if k.Kind == "StorageClass" {
			list.Items = append(list.Items, *v.(*storagev1.StorageClass))
		}
	}
	return nil
}

func (f *fakeClient) listVolumeSnapshots(list *snaps.VolumeSnapshotClassList) error {
	for k, v := range f.objects {
		if k.Kind == "VolumeSnapshot" {
			list.Items = append(list.Items, *v.(*snaps.VolumeSnapshotClass))
		}
	}
	return nil
}

func (f *fakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if f.errorInjector != nil {
		if err := f.errorInjector.shouldFail("Create", obj); err != nil {
			return err
		}
	}
	k, err := getKey(obj)
	if err != nil {
		return err
	}
	_, found := f.objects[k]
	if found {
		gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
		if err != nil {
			return err
		}
		gvr := schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		}
		return errors.NewAlreadyExists(gvr, k.Name)
	}
	f.objects[k] = obj
	return nil
}

func (f *fakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if len(opts) > 0 {
		return fmt.Errorf("delete options are not supported")
	}
	if f.errorInjector != nil {
		if err := f.errorInjector.shouldFail("Delete", obj); err != nil {
			return err
		}
	}

	k, err := getKey(obj)
	if err != nil {
		return err
	}
	_, found := f.objects[k]
	if !found {
		gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
		if err != nil {
			return err
		}
		gvr := schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		}
		return errors.NewNotFound(gvr, k.Name)
	}
	delete(f.objects, k)
	return nil
}

func (f *fakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if f.errorInjector != nil {
		if err := f.errorInjector.shouldFail("Update", obj); err != nil {
			return err
		}
	}
	k, err := getKey(obj)
	if err != nil {
		return err
	}
	_, found := f.objects[k]
	if !found {
		gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
		if err != nil {
			return err
		}
		gvr := schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		}
		return errors.NewNotFound(gvr, k.Name)
	}
	f.objects[k] = obj
	return nil
}

func (f *fakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	panic("implement me")
}

func (f *fakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	panic("implement me")
}

func (f *fakeClient) Status() client.StatusWriter {
	return f
}

func (f *fakeClient) Scheme() *runtime.Scheme {
	panic("implement me")
}

func (f *fakeClient) RESTMapper() meta.RESTMapper {
	panic("implement me")
}
