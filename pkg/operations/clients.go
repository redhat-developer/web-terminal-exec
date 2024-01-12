// Copyright (c) 2019-2024 Red Hat, Inc.
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
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	devworkspaceGroupVersion = schema.GroupVersion{
		Group:   "workspace.devfile.io",
		Version: "v1alpha2",
	}

	devworkspaceGVR = schema.GroupVersionResource{
		Group:    "workspace.devfile.io",
		Version:  "v1alpha2",
		Resource: "devworkspaces",
	}

	userGroupVersion = schema.GroupVersion{
		Group:   "user.openshift.io",
		Version: "v1",
	}

	userGVR = schema.GroupVersionResource{
		Group:    "user.openshift.io",
		Version:  "v1",
		Resource: "users",
	}
)

// NewSPDYExecutor is the function used to create a new SPDY Executor, with default
// value remotecommand.NewSPDYExecutor. Exposed only to enable testing code that uses
// the operations package.
var NewSPDYExecutor = remotecommand.NewSPDYExecutor

type ClientProvider interface {
	NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error)
	NewClientWithToken(token string) (kubernetes.Interface, *rest.Config, error)
	NewOpenShiftUserClient(token string) (dynamic.Interface, *rest.Config, error)
}

type defaultClientProvider struct {
	newSPDYExecutor func(*rest.Config, string, *url.URL) (remotecommand.Executor, error)
}

func DefaultClientProvider() ClientProvider {
	return defaultClientProvider{
		newSPDYExecutor: remotecommand.NewSPDYExecutor,
	}
}

func (defaultClientProvider) NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	config.APIPath = "/apis"
	config.GroupVersion = &devworkspaceGroupVersion

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return client, config, nil
}

func (defaultClientProvider) NewClientWithToken(token string) (kubernetes.Interface, *rest.Config, error) {
	if len(token) == 0 {
		return nil, nil, fmt.Errorf("failed to create client -- token must not be empty")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	config.BearerToken = token
	config.BearerTokenFile = ""

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return client, config, nil
}

func (defaultClientProvider) NewOpenShiftUserClient(token string) (dynamic.Interface, *rest.Config, error) {
	if len(token) == 0 {
		return nil, nil, fmt.Errorf("failed to create client -- token must not be empty")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	config.APIPath = "/apis"
	config.GroupVersion = &devworkspaceGroupVersion
	config.BearerToken = token
	config.BearerTokenFile = ""

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return client, config, nil
}
