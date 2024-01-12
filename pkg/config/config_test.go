// Copyright (c) 2019-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package config

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetsDefaultsFromEnv(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	t.Setenv(urlEnvVar, "test-url")
	t.Setenv(authenticatedUserIdEnvVar, "test-auth-id")
	t.Setenv(podSelectorEnvVar, "test-podselector")
	t.Setenv(devworkspaceIDEnvVar, "test-id")
	t.Setenv(devworkspaceNameEnvVar, "test-name")
	t.Setenv(devworkspaceNamespaceEnvVar, "test-namespace")
	err := updateDefaultsFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "test-url", defaultURLValue)
	assert.Equal(t, "test-auth-id", defaultAuthenticatedUserID)
	assert.Equal(t, "test-podselector", defaultPodSelector)
	assert.Equal(t, "test-id", DevWorkspaceID)
	assert.Equal(t, "test-name", DevWorkspaceName)
	assert.Equal(t, "test-namespace", DevWorkspaceNamespace)
}

func TestSetsDefaultPodSelector(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	t.Setenv(devworkspaceIDEnvVar, "test-id")
	t.Setenv(devworkspaceNameEnvVar, "test-name")
	t.Setenv(devworkspaceNamespaceEnvVar, "test-namespace")
	err := updateDefaultsFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "controller.devfile.io/devworkspace_id=test-id", defaultPodSelector)
}

func TestChecksRequiredEnvVars(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	err := updateDefaultsFromEnv()
	if assert.Error(t, err) {
		assert.Regexp(t, devworkspaceNameEnvVar, err.Error())
	}
	t.Setenv(devworkspaceNameEnvVar, "test-name")
	err = updateDefaultsFromEnv()
	if assert.Error(t, err) {
		assert.Regexp(t, devworkspaceNamespaceEnvVar, err.Error())
	}
	t.Setenv(devworkspaceNamespaceEnvVar, "test-namespace")
	err = updateDefaultsFromEnv()
	if assert.Error(t, err) {
		assert.Regexp(t, devworkspaceIDEnvVar, err.Error())
	}
	t.Setenv(devworkspaceIDEnvVar, "test-id")
	err = updateDefaultsFromEnv()
	assert.NoError(t, err)
}

func TestCheckReturnsErrorIfAuthorizedUserIDNotSpecified(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	AuthenticatedUserID = "\x00"
	err := checkConfigValid()
	assert.Error(t, err)
	assert.Regexp(t, "authenticated user ID must be specified via '--authenticated-user-id'", err.Error())
}

func TestSetsBearerTokenTrue(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	AuthenticatedUserID = "test"
	err := checkConfigValid()
	assert.NoError(t, err)
	assert.True(t, UseBearerToken, "UseBearerToken should be set to true")
}

func TestSetsUseTLSTrue(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	AuthenticatedUserID = "test"
	err := checkConfigValid()
	assert.NoError(t, err)
	assert.True(t, UseTLS, "UseTLS should be set to true")
}

func TestChecksIdleAndStopRetryTimeout(t *testing.T) {
	logrus.SetOutput(io.Discard)
	defer ResetConfigForTest()
	AuthenticatedUserID = "test"
	IdleTimeout = 1
	StopRetryPeriod = -1
	err := checkConfigValid()
	assert.Error(t, err)
	assert.Regexp(t, "invalid value for '--stop-retry-period': must be greater than zero if idling is enabled", err.Error())
}
