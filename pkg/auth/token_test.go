// Copyright (c) 2019-2024 Red Hat, Inc.
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

	"github.com/stretchr/testify/assert"
)

const (
	testToken = "test-token"
)

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		expected  string
		errRegexp string
	}{
		{
			name:     "Extracts from X-Access-Token header",
			headers:  http.Header{"X-Access-Token": []string{testToken}},
			expected: testToken,
		},
		{
			name:     "Extracts from X-Forwarded-Access-Token header",
			headers:  http.Header{"X-Access-Token": []string{testToken}},
			expected: testToken,
		},
		{
			name:      "Error if no access token header",
			headers:   http.Header{"Access": []string{testToken}},
			errRegexp: "authorization header is missing",
		},
		{
			name:     "Strips Bearer from X-Access-Token header",
			headers:  http.Header{"X-Access-Token": []string{"Bearer " + testToken}},
			expected: testToken,
		},
		{
			name:     "Does not strip Bearer from X-Access-Token header",
			headers:  http.Header{"X-Forwarded-Access-Token": []string{"Bearer " + testToken}},
			expected: "Bearer " + testToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{Header: tt.headers}
			result, err := ExtractToken(req)
			if tt.errRegexp != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.errRegexp, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
