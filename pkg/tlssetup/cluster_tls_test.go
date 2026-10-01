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

package tlssetup

import (
	"crypto/tls"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	libgocrypto "github.com/openshift/library-go/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultProfile() configv1.TLSProfileSpec {
	return *configv1.TLSProfiles[libgocrypto.DefaultTLSProfileType]
}

func oldProfile() configv1.TLSProfileSpec {
	return *configv1.TLSProfiles[configv1.TLSProfileOldType]
}

func modernProfile() configv1.TLSProfileSpec {
	return *configv1.TLSProfiles[configv1.TLSProfileModernType]
}

func tlsVersionFromOpenShift(v configv1.TLSProtocolVersion) uint16 {
	switch v {
	case configv1.VersionTLS10:
		return tls.VersionTLS10
	case configv1.VersionTLS11:
		return tls.VersionTLS11
	case configv1.VersionTLS12:
		return tls.VersionTLS12
	case configv1.VersionTLS13:
		return tls.VersionTLS13
	default:
		return 0
	}
}

func TestBuildTLSConfigFromProfile_OldProfile(t *testing.T) {
	profile := oldProfile()
	result := buildTLSConfigFromProfile(profile)

	require.NotNil(t, result)
	assert.Equal(t, tlsVersionFromOpenShift(profile.MinTLSVersion), result.MinVersion)
}

func TestBuildTLSConfigFromProfile_ModernProfile(t *testing.T) {
	profile := modernProfile()
	result := buildTLSConfigFromProfile(profile)

	require.NotNil(t, result)
	assert.Equal(t, uint16(tls.VersionTLS13), result.MinVersion,
		"modern profile should set TLS 1.3 minimum")
}

func TestBuildTLSConfigFromProfile_DefaultProfile(t *testing.T) {
	profile := defaultProfile()
	result := buildTLSConfigFromProfile(profile)

	require.NotNil(t, result)
	assert.Equal(t, tlsVersionFromOpenShift(profile.MinTLSVersion), result.MinVersion)
}

func TestBuildTLSConfigFromProfile_IntermediateProfile(t *testing.T) {
	intermediate := *configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
	result := buildTLSConfigFromProfile(intermediate)

	require.NotNil(t, result)
	assert.Equal(t, tlsVersionFromOpenShift(intermediate.MinTLSVersion), result.MinVersion)
}

func TestBuildTLSConfigFromProfile_AlwaysReturnsNonNil(t *testing.T) {
	profiles := []struct {
		name    string
		profile configv1.TLSProfileSpec
	}{
		{"old profile", oldProfile()},
		{"modern profile", modernProfile()},
		{"default profile", defaultProfile()},
	}

	for _, tt := range profiles {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTLSConfigFromProfile(tt.profile)
			assert.NotNil(t, result)
		})
	}
}
