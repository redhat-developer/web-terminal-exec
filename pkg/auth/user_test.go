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
)

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		errRegexp string
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
			name:    "Correct token",
			headers: http.Header{"X-Access-Token": []string{testToken}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AuthenticatedUserID = testToken
			defer config.ResetConfigForTest()

			clientProvider := test.FakeClientProvider{
				UserToken: testToken,
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
