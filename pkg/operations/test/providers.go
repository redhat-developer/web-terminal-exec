// Copyright (c) 2019-2025 Red Hat, Inc.
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
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	authenticationv1 "k8s.io/api/authentication/v1"
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
	k8stesting "k8s.io/client-go/testing"
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
// objects. SelfSubjectReview responses use the request token as the returned UID.
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
	client := fake.NewSimpleClientset(p.InitialObjs...)
	client.PrependReactor("create", "selfsubjectreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authenticationv1.SelfSubjectReview{
			Status: authenticationv1.SelfSubjectReviewStatus{
				UserInfo: authenticationv1.UserInfo{
					UID: token,
				},
			},
		}, nil
	})
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
