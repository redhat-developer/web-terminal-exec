// Copyright (c) 2019-2025 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package operations

import (
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/stretchr/testify/assert"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/yaml"
)

func TestGetCurrentUserUID(t *testing.T) {
	const expectedUID = "test-user-uid"
	tests := []struct {
		name        string
		provider    ClientProvider
		errRegexp   string
		expectedUID string
	}{
		{
			name: "Should return UID from SelfSubjectReview",
			provider: testUserIDClientProvider{
				userUID: expectedUID,
			},
			expectedUID: expectedUID,
		},
		{
			name: "Should return error when client creation fails",
			provider: testUserIDClientProvider{
				returnClientError: true,
			},
			errRegexp: "failed to create client to check user info",
		},
		{
			name: "Should fall back to OpenShift User API when SelfSubjectReview is unavailable",
			provider: testUserIDClientProvider{
				returnReviewError: apierrors.NewNotFound(schema.GroupResource{Group: "authentication.k8s.io", Resource: "selfsubjectreviews"}, "self"),
				userAPIUID:        expectedUID,
			},
			expectedUID: expectedUID,
		},
		{
			name: "Should return error when both user lookups fail",
			provider: testUserIDClientProvider{
				returnReviewError:  apierrors.NewNotFound(schema.GroupResource{Group: "authentication.k8s.io", Resource: "selfsubjectreviews"}, "self"),
				returnUserAPIError: apierrors.NewNotFound(schema.GroupResource{Group: "user.openshift.io", Resource: "users"}, "~"),
			},
			errRegexp: "failed to get current user information",
		},
		{
			name: "Should allow empty UID from SelfSubjectReview for kube:admin",
			provider: testUserIDClientProvider{
				userUID: "",
			},
			expectedUID: "",
		},
		{
			name: "Should allow empty UID from OpenShift User API for kube:admin",
			provider: testUserIDClientProvider{
				returnReviewError: apierrors.NewNotFound(schema.GroupResource{Group: "authentication.k8s.io", Resource: "selfsubjectreviews"}, "self"),
				userAPIUID:        "",
				emptyUserAPIUID:   true,
			},
			expectedUID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, err := GetCurrentUserUID("test-token", tt.provider)
			if tt.errRegexp != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.errRegexp, err.Error())
				if tt.name == "Should return error when both user lookups fail" {
					assert.Contains(t, err.Error(), "SelfSubjectReview error:")
					assert.Contains(t, err.Error(), "OpenShift User API error:")
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedUID, uid)
		})
	}
}

type testUserIDClientProvider struct {
	userUID            string
	userAPIUID         string
	returnClientError  bool
	returnReviewError  error
	returnUserAPIError error
	emptyUserAPIUID    bool
}

func (p testUserIDClientProvider) NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error) {
	return nil, nil, nil
}

func (p testUserIDClientProvider) NewClientWithToken(string) (kubernetes.Interface, *rest.Config, error) {
	if p.returnClientError {
		return nil, nil, fmt.Errorf("(TEST) expected error")
	}
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "selfsubjectreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		if p.returnReviewError != nil {
			return true, nil, p.returnReviewError
		}
		return true, &authenticationv1.SelfSubjectReview{
			Status: authenticationv1.SelfSubjectReviewStatus{
				UserInfo: authenticationv1.UserInfo{
					UID: p.userUID,
				},
			},
		}, nil
	})
	return client, &rest.Config{}, nil
}

func (p testUserIDClientProvider) NewOpenShiftUserClient(string) (dynamic.Interface, *rest.Config, error) {
	if p.returnUserAPIError != nil {
		return fakedynamic.NewSimpleDynamicClient(&runtime.Scheme{}), &rest.Config{}, nil
	}
	if p.userAPIUID == "" && !p.emptyUserAPIUID {
		return nil, nil, fmt.Errorf("(TEST) OpenShift User API not configured")
	}
	fakeUser := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "~",
				"uid":  p.userAPIUID,
			},
		},
	}
	fakeUser.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "user.openshift.io",
		Version: "v1",
		Kind:    "User",
	})
	client := fakedynamic.NewSimpleDynamicClient(&runtime.Scheme{}, fakeUser)
	return client, &rest.Config{}, nil
}

func setConfigForTest() {
	config.DevWorkspaceName = "test-workspace"
	config.DevWorkspaceNamespace = "test-namespace"
	config.DevWorkspaceID = "test-id"
}

func loadDevWorkspaceFromFile(t *testing.T) unstructured.Unstructured {
	bytes, err := os.ReadFile("testdata/devworkspace.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var workspace unstructured.Unstructured
	if err := yaml.Unmarshal(bytes, &workspace); err != nil {
		t.Fatal(err)
	}
	return workspace
}

func TestStopDevWorkspace(t *testing.T) {
	setConfigForTest()
	defer config.ResetConfigForTest()
	workspace := loadDevWorkspaceFromFile(t)
	fakeDynamic := fakedynamic.NewSimpleDynamicClient(&runtime.Scheme{}, &workspace)
	err := StopDevWorkspace(fakeDynamic)
	assert.NoError(t, err, "Should not return error when stopping workspace")
	result, err := fakeDynamic.Resource(devworkspaceGVR).Namespace(workspace.GetNamespace()).Get(context.TODO(), workspace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err, "Unexpected error getting devworkspace")
	assert.False(t, workspaceIsStarted(t, result), "Workspace should be stopped")
}

func TestGetCurrentWorkspacePod(t *testing.T) {
	const expectedPodName = "test-terminal-pod"
	t.Setenv("HOSTNAME", expectedPodName)
	tests := []struct {
		name         string
		podFilenames []string
		errRegexp    string
	}{
		{
			name:         "Simple test with one pod",
			podFilenames: []string{"pod.yaml"},
		},
		{
			name:         "No pods in namespace",
			podFilenames: []string{},
			errRegexp:    "no workspace pods found",
		},
		{
			name:         "Multiple pods in namespace",
			podFilenames: []string{"pod.yaml", "alternate-pod.yaml"},
		},
		{
			name:         "Multiple pods in namespace but no terminal",
			podFilenames: []string{"alternate-pod.yaml", "alternate-pod-2.yaml"},
			errRegexp:    "failed to get current workspace pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setConfigForTest()
			defer config.ResetConfigForTest()
			var pods []runtime.Object
			for _, filename := range tt.podFilenames {
				pods = append(pods, loadPodFromFile(t, filename))
			}
			client := fake.NewSimpleClientset(pods...)
			actualPod, err := GetCurrentWorkspacePod(client)
			if tt.errRegexp != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.errRegexp, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedPodName, actualPod.Name)
			}
		})
	}
}

func workspaceIsStarted(t *testing.T, workspace *unstructured.Unstructured) bool {
	return readUnstructuredPath(t, workspace, reflect.TypeOf(false), "spec", "started").(bool)
}

func readUnstructuredPath(t *testing.T, obj *unstructured.Unstructured, resultType reflect.Type, fields ...string) interface{} {
	var innerField map[string]interface{}
	for i := 0; i < len(fields)-1; i++ {
		temp, ok := obj.Object[fields[i]].(map[string]interface{})
		if !ok {
			t.Fatalf("Failed to read field '%s' in object", strings.Join(fields, ", "))
		}
		innerField = temp
	}
	result := innerField[fields[len(fields)-1]]
	if reflect.TypeOf(result) != resultType {
		t.Fatalf("Failed to read into parameter, types don't match")
	}
	return result
}

func loadPodFromFile(t *testing.T, filepath string) runtime.Object {
	podbytes, err := os.ReadFile(path.Join("testdata", filepath))
	if err != nil {
		t.Fatal(err)
	}
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal(podbytes, pod); err != nil {
		t.Fatal(err)
	}

	return pod
}
