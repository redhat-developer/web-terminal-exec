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

func TestBuildTLSConfig_EmptyVersion(t *testing.T) {
	logrus.SetOutput(io.Discard)
	config.TLSMinVersion = ""
	assert.Nil(t, buildTLSConfig())
}

func TestBuildTLSConfig_ValidTLS12(t *testing.T) {
	logrus.SetOutput(io.Discard)
	config.TLSMinVersion = "771"
	tlsConfig := buildTLSConfig()
	assert.NotNil(t, tlsConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
}

func TestBuildTLSConfig_ValidTLS13(t *testing.T) {
	logrus.SetOutput(io.Discard)
	config.TLSMinVersion = "772"
	tlsConfig := buildTLSConfig()
	assert.NotNil(t, tlsConfig)
	assert.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MinVersion)
}

func TestBuildTLSConfig_InvalidString(t *testing.T) {
	logrus.SetOutput(io.Discard)
	config.TLSMinVersion = "not-a-number"
	assert.Nil(t, buildTLSConfig())
}

func TestBuildTLSConfig_UnknownVersion(t *testing.T) {
	logrus.SetOutput(io.Discard)
	config.TLSMinVersion = "999"
	assert.Nil(t, buildTLSConfig())
}
