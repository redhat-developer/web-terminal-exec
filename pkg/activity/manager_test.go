// Copyright (c) 2019-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package activity

import (
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"sigs.k8s.io/yaml"
)

var (
	testDevworkspaceGVR = schema.GroupVersionResource{
		Group:    "workspace.devfile.io",
		Version:  "v1alpha2",
		Resource: "devworkspaces",
	}
)

func TestActivityManagerTimesOut(t *testing.T) {
	logrus.SetOutput(io.Discard)
	workspace := loadDevWorkspaceFromFile(t)
	config.DevWorkspaceName = workspace.GetName()
	config.DevWorkspaceNamespace = workspace.GetNamespace()
	config.DevWorkspaceID = "test-id"
	defer config.ResetConfigForTest()

	fakeClientProvider := test.FakeClientProvider{InitialDynamic: []runtime.Object{&workspace}}
	manager, err := NewActivityManager(1*time.Millisecond, 1*time.Millisecond, fakeClientProvider)
	assert.NoError(t, err)
	manager.Start()
	time.Sleep(20 * time.Millisecond)
	client := manager.(*activityManager).devworkspaceClient
	newWorkspace, err := client.Resource(testDevworkspaceGVR).Namespace(workspace.GetNamespace()).Get(context.TODO(), workspace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.False(t, workspaceIsStarted(t, newWorkspace), "Workspace should be stopped")
}

func TestTickResetsTimer(t *testing.T) {
	logrus.SetOutput(io.Discard)
	workspace := loadDevWorkspaceFromFile(t)
	config.DevWorkspaceName = workspace.GetName()
	config.DevWorkspaceNamespace = workspace.GetNamespace()
	config.DevWorkspaceID = "test-id"
	defer config.ResetConfigForTest()

	fakeDynamicClient := fake.NewSimpleDynamicClient(&runtime.Scheme{}, &workspace)

	manager := activityManager{
		idleTimeout:        5 * time.Millisecond,
		stopRetryPeriod:    5 * time.Millisecond,
		devworkspaceClient: fakeDynamicClient,
		activityC:          make(chan bool),
	}
	activity, done := time.NewTicker(1*time.Millisecond), make(chan bool)
	go func() {
		for {
			select {
			case <-activity.C:
				manager.Tick()
			case <-done:
				activity.Stop()
				return
			}
		}
	}()
	manager.Start()
	time.Sleep(50 * time.Millisecond)
	newWorkspace, err := fakeDynamicClient.Resource(testDevworkspaceGVR).Namespace(workspace.GetNamespace()).Get(context.TODO(), workspace.GetName(), metav1.GetOptions{})
	assert.NoError(t, err)
	assert.True(t, workspaceIsStarted(t, newWorkspace), "Workspace should be stopped")
	close(done)
}

func TestActivityManagerIsNoOpIfNoIdleTimeout(t *testing.T) {
	manager, err := NewActivityManager(-1, 0, nil)
	assert.NoError(t, err)
	assert.IsType(t, &noOpManager{}, manager, "Should use no-op manager if idle timeout is less than 0")
}

func TestReturnsErrorIfStopDurationNotSpecified(t *testing.T) {
	_, err := NewActivityManager(1, -1, nil)
	assert.Error(t, err)
	assert.Regexp(t, "stop retry period must be greater than 0", err.Error())
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
