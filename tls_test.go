//
// Copyright (c) 2019-2026 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package main

import (
	"crypto/tls"
	"io"
	"testing"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBuildTLSConfig(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		wantNil    bool
		wantMinVer uint16
	}{
		{
			name:    "empty version returns nil",
			version: "",
			wantNil: true,
		},
		{
			name:       "valid TLS 1.2",
			version:    "771",
			wantMinVer: tls.VersionTLS12,
		},
		{
			name:       "valid TLS 1.3",
			version:    "772",
			wantMinVer: tls.VersionTLS13,
		},
		{
			name:    "invalid non-numeric string returns nil",
			version: "not-a-number",
			wantNil: true,
		},
		{
			name:       "valid TLS 1.0",
			version:    "769",
			wantMinVer: tls.VersionTLS10,
		},
		{
			name:       "valid TLS 1.1",
			version:    "770",
			wantMinVer: tls.VersionTLS11,
		},
		{
			name:    "SSLv3 is rejected",
			version: "768",
			wantNil: true,
		},
		{
			name:    "unknown version returns nil",
			version: "999",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origOutput := logrus.StandardLogger().Out
			origVersion := config.TLSMinVersion
			t.Cleanup(func() {
				logrus.SetOutput(origOutput)
				config.TLSMinVersion = origVersion
			})

			logrus.SetOutput(io.Discard)
			config.TLSMinVersion = tt.version

			result := buildTLSConfig()
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantMinVer, result.MinVersion)
			}
		})
	}
}
