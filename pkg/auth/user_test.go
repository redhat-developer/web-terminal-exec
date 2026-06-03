// Copyright (c) 2019-2025 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package auth

import (
	"net/http"
	"testing"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations/test"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
)

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		errRegexp string
		provider  operationsClientProvider
	}{
		{
			name:      "No auth token in request",
			headers:   http.Header{},
			errRegexp: "authorization header is missing",
		},
		{
			name:      "Wrong token in request",
			headers:   http.Header{"X-Access-Token": []string{"incorrect"}},
			errRegexp: "the current user is not authorized to access this web terminal",
		},
		{
			name:      "SelfSubjectReview failure",
			headers:   http.Header{"X-Access-Token": []string{testToken}},
			errRegexp: "unable to verify user: failed to get current user information",
			provider:  selfSubjectReviewErrorClientProvider{},
		},
		{
			name:    "Correct token",
			headers: http.Header{"X-Access-Token": []string{testToken}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AuthenticatedUserID = testToken
			defer config.ResetConfigForTest()

			clientProvider := tt.provider
			if clientProvider == nil {
				clientProvider = test.FakeClientProvider{
					UserToken: testToken,
				}
			}
			req := &http.Request{Header: tt.headers}
			err := Authenticate(req, clientProvider)
			if tt.errRegexp != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.errRegexp, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type operationsClientProvider interface {
	NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error)
	NewClientWithToken(token string) (kubernetes.Interface, *rest.Config, error)
	NewOpenShiftUserClient(token string) (dynamic.Interface, *rest.Config, error)
}

type selfSubjectReviewErrorClientProvider struct{}

func (selfSubjectReviewErrorClientProvider) NewDevWorkspaceClient() (dynamic.Interface, *rest.Config, error) {
	return nil, nil, nil
}

func (selfSubjectReviewErrorClientProvider) NewClientWithToken(string) (kubernetes.Interface, *rest.Config, error) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "selfsubjectreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "authentication.k8s.io", Resource: "selfsubjectreviews"}, "self")
	})
	return client, &rest.Config{}, nil
}

func (selfSubjectReviewErrorClientProvider) NewOpenShiftUserClient(string) (dynamic.Interface, *rest.Config, error) {
	return fakedynamic.NewSimpleDynamicClient(&runtime.Scheme{}), &rest.Config{}, nil
}
