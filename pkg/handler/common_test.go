// Copyright (c) 2019-2025 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/redhat-developer/web-terminal-exec/pkg/activity"
	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	optest "github.com/redhat-developer/web-terminal-exec/pkg/operations/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	testUserToken = "test-user-token"
)

func setConfigForTest() {
	config.DevWorkspaceName = "test-workspace"
	config.DevWorkspaceNamespace = "test-namespace"
	config.AuthenticatedUserID = "test-user-token"
	config.DevWorkspaceID = "test-workspace-id"
	config.PodSelector = "controller.devfile.io/devworkspace_id=test-workspace-id"
}

func TestRouter(t *testing.T) {
	logrus.SetOutput(io.Discard)
	// Workarounds to emulate running in a pod within a cluster.
	t.Setenv("KUBERNETES_SERVICE_HOST", "0.0.0.0")
	t.Setenv("KUBERNETES_SERVICE_PORT", "8443")

	noOpActivityManager, err := activity.NewActivityManager(-1, -1, optest.NoOpClientProvider{})
	assert.NoError(t, err)

	tests := []struct {
		name        string
		initialObjs []runtime.Object
		spdy        optest.FakeSPDYExecutorProvider
		req         *http.Request
		headers     http.Header
		respCode    int
		respBody    string
	}{
		{
			name:     "test /healthz returns 200",
			req:      httptest.NewRequest("GET", "/healthz", nil),
			respCode: http.StatusOK,
		},
		{
			name:     "test /activity/tick does nothing when not configured",
			req:      httptest.NewRequest("POST", "/activity/tick", nil),
			respCode: http.StatusNoContent,
			headers:  http.Header{"X-Access-Token": []string{testUserToken}},
		},
		{
			name:        "test basic /exec/init behavior",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusOK,
			respBody:    `{"pod": "test-terminal-pod", "container": "web-terminal-tooling", "cmd": ["test_shellcommand"]}`,
			headers:     http.Header{"X-Forwarded-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL": "test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test resolves container when multiple",
			initialObjs: loadPodFromFile(t, "multi-container-pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusOK,
			respBody:    `{"pod": "test-terminal-pod", "container": "web-terminal-tooling", "cmd": ["test_shellcommand"]}`,
			headers:     http.Header{"X-Forwarded-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL": "test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test resolves user-defined tooling container",
			initialObjs: loadPodFromFile(t, "generic-pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusOK,
			respBody:    `{"pod": "test-terminal-pod", "container": "test-user-defined", "cmd": ["test_shellcommand"]}`,
			headers:     http.Header{"X-Forwarded-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL": "test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test fails when only exec container",
			initialObjs: loadPodFromFile(t, "only-exec-pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusBadRequest,
			headers:     http.Header{"X-Forwarded-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL": "test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test uses default username",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"namespace": "test-namespace"}}`))),
			respCode:    http.StatusOK,
			respBody:    `{"pod": "test-terminal-pod", "container": "web-terminal-tooling", "cmd": ["test_shellcommand"]}`,
			headers:     http.Header{"X-Forwarded-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL": "test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test shell detection when no SHELL env var",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusOK,
			respBody:    `{"pod": "test-terminal-pod", "container": "web-terminal-tooling", "cmd": ["test_shellcommand"]}`,
			headers:     http.Header{"X-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL":     "\n",
						"id -u":           "1234\n",
						"cat /etc/passwd": "user:x:1234:0:user user:/home/user:test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test specifies container name",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"container": "web-terminal-exec", "kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusOK,
			respBody:    `{"pod": "test-terminal-pod", "container": "web-terminal-exec", "cmd": ["test_shellcommand"]}`,
			headers:     http.Header{"X-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL":     "\n",
						"id -u":           "1234\n",
						"cat /etc/passwd": "user:x:1234:0:user user:/home/user:test_shellcommand\n",
					},
				},
			},
		},
		{
			name:        "test specifies invalid container name",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"container": "not-exist", "kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusBadRequest,
			headers:     http.Header{"X-Access-Token": []string{testUserToken}},
		},
		{
			name:        "test cannot resolve shell",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusInternalServerError,
			headers:     http.Header{"X-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL":     "\n",
						"id -u":           "1234\n",
						"cat /etc/passwd": "\n",
					},
				},
			},
		},
		{
			name:        "test cannot create kubeconfig",
			initialObjs: loadPodFromFile(t, "pod.yaml"),
			req:         httptest.NewRequest("POST", "/exec/init", bytes.NewBuffer([]byte(`{"kubeconfig": {"username": "test", "namespace": "test-namespace"}}`))),
			respCode:    http.StatusInternalServerError,
			headers:     http.Header{"X-Access-Token": []string{testUserToken}},
			spdy: optest.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: optest.FakeSPDYExecutor{
					ResponseOutputs: map[string]string{
						"echo $SHELL": "test_shellcommand\n",
					},
					ErrInputs: []string{"$KUBECONFIG"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (%s %s)", tt.name, tt.req.Method, tt.req.URL.Path), func(t *testing.T) {
			router := Router{
				ActivityManager: noOpActivityManager,
				ClientProvider: optest.FakeClientProvider{
					InitialObjs: tt.initialObjs,
					UserToken:   testUserToken,
				},
			}
			handler := router.HTTPSHandler()

			setConfigForTest()
			defer config.ResetConfigForTest()
			oldSPDYExecutor := operations.NewSPDYExecutor
			operations.NewSPDYExecutor = tt.spdy.NewFakeSPDYExecutor
			defer func() { operations.NewSPDYExecutor = oldSPDYExecutor }()

			if tt.headers != nil {
				tt.req.Header = tt.headers
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, tt.req)
			actualRespCode := recorder.Code
			actualBodyBytes, err := io.ReadAll(recorder.Body)
			actualBody := string(actualBodyBytes)
			if !assert.NoError(t, err, "failed to read recorder body") {
				return
			}
			if !assert.Equal(t, tt.respCode, actualRespCode, "Status code should match") {
				t.Logf("Response body: %s", actualBody)
				return
			}
			if tt.respBody != "" {
				assert.Equal(t, strings.ReplaceAll(tt.respBody, " ", ""), actualBody)
			}
		})
	}
}

func TestAuthorizedAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)

	noOpActivityManager, err := activity.NewActivityManager(-1, -1, optest.NoOpClientProvider{})
	assert.NoError(t, err)
	router := Router{
		ActivityManager: noOpActivityManager,
		ClientProvider: optest.FakeClientProvider{
			UserToken: testUserToken,
		},
	}
	handler := router.HTTPSHandler()

	tests := []struct {
		name     string
		headers  http.Header
		respBody string
	}{
		{
			name:     "no access headers",
			headers:  http.Header{},
			respBody: "authorization header is missing",
		},
		{
			name:     "wrong user token",
			headers:  http.Header{"X-Access-Token": []string{"bad token"}},
			respBody: "the current user is not authorized to access this web terminal",
		},
		{
			name:     "wrong forwarded user token",
			headers:  http.Header{"X-Forwarded-Access-Token": []string{"bad token"}},
			respBody: "the current user is not authorized to access this web terminal",
		},
		{
			name:     "ignores Authorization: header",
			headers:  http.Header{"Authorization": []string{"Bearer bad token"}},
			respBody: "authorization header is missing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activityRecorder := httptest.NewRecorder()
			activityReq := httptest.NewRequest("GET", "/activity/tick", nil)
			activityReq.Header = tt.headers
			handler.ServeHTTP(activityRecorder, activityReq)
			assert.Equal(t, http.StatusUnauthorized, activityRecorder.Code, "Should return unauthorized on /activity/tick")
			respBytes, err := io.ReadAll(activityRecorder.Body)
			assert.Nil(t, err)
			assert.Regexp(t, tt.respBody, string(respBytes))

			execRecorder := httptest.NewRecorder()
			execReq := httptest.NewRequest("POST", "/exec/init", nil)
			execReq.Header = tt.headers
			handler.ServeHTTP(execRecorder, execReq)
			assert.Equal(t, http.StatusUnauthorized, execRecorder.Code, "Should return unauthorized on /exec/init")
			respBytes, err = io.ReadAll(execRecorder.Body)
			assert.Nil(t, err)
			assert.Regexp(t, tt.respBody, string(respBytes))
		})
	}
}

func TestRouterEndpoints(t *testing.T) {
	logrus.SetOutput(io.Discard)

	noOpActivityManager, err := activity.NewActivityManager(-1, -1, optest.NoOpClientProvider{})
	assert.NoError(t, err)
	router := Router{
		ActivityManager: noOpActivityManager,
		ClientProvider: optest.FakeClientProvider{
			UserToken: testUserToken,
		},
	}
	handler := router.HTTPSHandler()

	httpMethods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}
	methodIsSupported := func(method string, supported []string) bool {
		for _, supportedMethod := range supported {
			if method == supportedMethod {
				return true
			}
		}
		return false
	}
	tests := []struct {
		endpoint         string
		supportedMethods []string
		respCode         int
	}{
		{
			endpoint:         "/healthz",
			supportedMethods: []string{"GET"},
			respCode:         http.StatusMethodNotAllowed,
		},
		{
			endpoint:         "/activity/tick",
			supportedMethods: []string{"POST"},
			respCode:         http.StatusMethodNotAllowed,
		},
		{
			endpoint:         "/exec/init",
			supportedMethods: []string{"POST"},
			respCode:         http.StatusMethodNotAllowed,
		},
		{
			endpoint:         "/activity/",
			supportedMethods: []string{},
			respCode:         http.StatusNotFound,
		},
		{
			endpoint:         "/exec/",
			supportedMethods: []string{},
			respCode:         http.StatusNotFound,
		},
		{
			endpoint:         "/",
			supportedMethods: []string{},
			respCode:         http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Test allowed methods %s (%s)", tt.endpoint, strings.Join(tt.supportedMethods, ",")), func(t *testing.T) {
			for _, method := range httpMethods {
				setConfigForTest()
				defer config.ResetConfigForTest()
				if methodIsSupported(method, tt.supportedMethods) {
					continue
				}
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(method, tt.endpoint, nil)
				req.Header.Add("X-Access-Token", testUserToken)
				handler.ServeHTTP(recorder, req)
				assert.Equal(t, tt.respCode, recorder.Code, "Wrong code returned")
			}
		})
	}
}

func loadPodFromFile(t *testing.T, filepath string) []runtime.Object {
	podbytes, err := os.ReadFile(path.Join("testdata", filepath))
	if err != nil {
		t.Fatal(err)
	}
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal(podbytes, pod); err != nil {
		t.Fatal(err)
	}

	return []runtime.Object{pod}
}
