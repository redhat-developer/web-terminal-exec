// Copyright (c) 2019-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package test

import (
	"fmt"

	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	fakehttp "k8s.io/client-go/rest/fake"
)

var userGVK = schema.GroupVersionKind{
	Group:   "user.openshift.io",
	Version: "v1",
	Kind:    "User",
}

// NoOpClientProvider returns nil for all functions in the interface, for use in tests
// where clients are not expected to be used
type NoOpClientProvider struct{}

var _ operations.ClientProvider = (*NoOpClientProvider)(nil)

func (NoOpClientProvider) NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error) {
	return nil, nil, nil
}

func (NoOpClientProvider) NewClientWithToken(token string) (kubernetes.Interface, *rest.Config, error) {
	return nil, nil, nil
}

func (NoOpClientProvider) NewOpenShiftUserClient(token string) (dynamic.Interface, *rest.Config, error) {
	return nil, nil, nil
}

// FakeClientProvider returns fake clientsets and dynamic clients that are initialized with
// objects. UserToken is used to verify that expected token is passed to NewClientWithToken()
type FakeClientProvider struct {
	InitialObjs    []runtime.Object
	InitialDynamic []runtime.Object
	UserToken      string
}

var _ operations.ClientProvider = (*FakeClientProvider)(nil)

func (p FakeClientProvider) NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error) {
	client := fakedynamic.NewSimpleDynamicClient(&runtime.Scheme{}, p.InitialDynamic...)
	return client, &rest.Config{}, nil
}

func (p FakeClientProvider) NewClientWithToken(token string) (kubernetes.Interface, *rest.Config, error) {
	if token != p.UserToken {
		return nil, nil, fmt.Errorf("(TEST) Invalid token")
	}
	client := fake.NewSimpleClientset(p.InitialObjs...)
	return &WrapFakeClientCoreV1{client}, &rest.Config{}, nil
}

func (p FakeClientProvider) NewOpenShiftUserClient(token string) (dynamic.Interface, *rest.Config, error) {
	// Fake OpenShift User API -- Use '~' as name since API endpoint apis/user.openshift.io/v1/users/~ serves
	// current user. Use token as authorized user ID for convenience
	fakeUser := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "~",
				"uid":  token,
			},
		},
	}
	fakeUser.SetGroupVersionKind(userGVK)
	client := fakedynamic.NewSimpleDynamicClient(&runtime.Scheme{}, fakeUser)
	return client, &rest.Config{}, nil
}

// Functions below are to wrap the RESTClient in fake.Clientset (which is by default nil)
// This is required to allow allow resolving requests for pods/exec in tests.
type WrapFakeClientCoreV1 struct {
	*fake.Clientset
}

func (f *WrapFakeClientCoreV1) CoreV1() typedcorev1.CoreV1Interface {
	return &wrapFakeClientHTTPClient{f.Clientset.CoreV1()}
}

type wrapFakeClientHTTPClient struct {
	typedcorev1.CoreV1Interface
}

func (f *wrapFakeClientHTTPClient) RESTClient() restclient.Interface {
	// return &fakeKubeRESTClient{f.CoreV1Interface.RESTClient()}
	return &fakehttp.RESTClient{}
}

type fakeKubeRESTClient struct {
	restclient.Interface
}

func (*fakeKubeRESTClient) Post() *restclient.Request {
	return &restclient.Request{}
}
