// Copyright (c) 2019-2022 Red Hat, Inc.
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
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// Name of the current DevWorkspace
	DevWorkspaceName string

	// Namespace for the current DevWorkspace
	DevWorkspaceNamespace string

	// ID for the current DevWorkspace
	DevWorkspaceID string

	// URL to listen on (default :4444)
	URL string

	// AuthenticatedUserID OpenShift UID of user authorized to access DevWorksapce. Required; may be specified as empty
	// string for users that do not have a UID (e.g. kubeadmin)
	AuthenticatedUserID string

	// IdleTimeout is a inactivity period after which workspace should be stopped
	// Default -1, which mean - does not stop
	IdleTimeout time.Duration

	// StopRetryPeriod is a period after which workspace should be tried to stop if the previous try failed
	// Defaults 10 second
	StopRetryPeriod time.Duration

	// PodSelector set of labels to be used as selector for getting workspace pod.
	// Default value is controller.devfile.io/devworkspace_id=${DEVWORKSPACE_ID}
	PodSelector string

	// UseTLS (deprecated) kept for compatibility but if specified must have 'true' value
	UseTLS bool

	// UseBearerToken (deprecated) kept for compatibility but if specified must have 'true' value
	UseBearerToken bool
)

const (
	urlEnvVar                   = "API_URL"
	authenticatedUserIdEnvVar   = "AUTHENTICATED_USER_ID"
	podSelectorEnvVar           = "POD_SELECTOR"
	idleTimeoutEnvVar           = "IDLE_TIMEOUT"
	stopRetryEnvVar             = "STOP_RETRY_PERIOD"
	devworkspaceIDEnvVar        = "DEVWORKSPACE_ID"
	devworkspaceNameEnvVar      = "DEVWORKSPACE_NAME"
	devworkspaceNamespaceEnvVar = "DEVWORKSPACE_NAMESPACE"
)

var (
	defaultURLValue            = ":4444"
	defaultAuthenticatedUserID = ""
	defaultPodSelector         = ""
	defaultIdleTimeout         = 5 * time.Minute
	defaultStopRetryPeriod     = 10 * time.Second
	defaultUseBearerToken      = true
	defaultUseTLS              = true
)

func ParseConfig() error {
	if err := updateDefaultsFromEnv(); err != nil {
		return fmt.Errorf("invalid configuration: %s", err)
	}
	setLogLevel()

	flag.StringVar(&URL, "url", defaultURLValue, "Host:Port address for the Web Terminal Exec server. Default is :4444")
	flag.StringVar(&AuthenticatedUserID, "authenticated-user-id", defaultAuthenticatedUserID, "OpenShift user's ID that should has access to API. Must be set.")
	flag.DurationVar(&IdleTimeout, "idle-timeout", defaultIdleTimeout, "IdleTimeout is a inactivity period after which workspace should be stopped. Use '-1' to disable idle timeout. Examples: -1, 30s, 15m, 1h")
	flag.DurationVar(&StopRetryPeriod, "stop-retry-period", defaultStopRetryPeriod, "StopRetryPeriod is a period after which workspace should be tried to stop if the previous try failed. Examples: 30s")
	flag.BoolVar(&UseBearerToken, "use-bearer-token", defaultUseBearerToken, "Use user's bearer token when communicating with OpenShift API. Option is kept for backwards-compatibility; must be set to 'true'.")
	flag.BoolVar(&UseTLS, "use-tls", defaultUseTLS, "Serve content via TLS. Option is kept for backwards-compatibility; must be set to 'true'")
	flag.StringVar(&PodSelector, "pod-selector", defaultPodSelector, "Selector that is used to find workspace pod. Default value is controller.devfile.io/devworkspace_id=${DEVWORKSPACE_ID}")
	flag.Parse()

	if err := checkConfigValid(); err != nil {
		logrus.Errorf("Invalid configuration: %s", err)
		return err
	}

	printConfig()

	return nil
}

func updateDefaultsFromEnv() error {
	urlEnvValue, isFound := os.LookupEnv(urlEnvVar)
	if isFound && len(urlEnvValue) > 0 {
		logrus.Infof("Read value %s from environment variable %s", urlEnvValue, urlEnvVar)
		defaultURLValue = urlEnvValue
	}
	authenticatedUserID, isFound := os.LookupEnv(authenticatedUserIdEnvVar)
	if isFound {
		logrus.Infof("Read value %s from environment variable %s", authenticatedUserID, authenticatedUserIdEnvVar)
		defaultAuthenticatedUserID = authenticatedUserID
	}
	podSelector, isFound := os.LookupEnv(podSelectorEnvVar)
	if isFound && len(podSelector) > 0 {
		logrus.Infof("Read value %s from environment variable %s", podSelector, podSelectorEnvVar)
		defaultPodSelector = podSelector
	} else {
		workspaceID, isFound := os.LookupEnv(devworkspaceIDEnvVar)
		if isFound {
			defaultPodSelector = fmt.Sprintf("controller.devfile.io/devworkspace_id=%s", workspaceID)
		}
	}
	DevWorkspaceName = os.Getenv(devworkspaceNameEnvVar)
	DevWorkspaceNamespace = os.Getenv(devworkspaceNamespaceEnvVar)
	DevWorkspaceID = os.Getenv(devworkspaceIDEnvVar)
	if DevWorkspaceName == "" {
		return fmt.Errorf("environment variable %s must be set", devworkspaceNameEnvVar)
	}
	if DevWorkspaceNamespace == "" {
		return fmt.Errorf("environment variable %s must be set", devworkspaceNamespaceEnvVar)
	}
	if DevWorkspaceID == "" {
		return fmt.Errorf("environment variable %s must be set", devworkspaceIDEnvVar)
	}
	return nil
}

func checkConfigValid() error {
	if AuthenticatedUserID == "" {
		return fmt.Errorf("authenticated user ID must be specified via '--authenticated-user-id")
	}
	if !UseBearerToken {
		logrus.Warn("Flag '--use-bearer-token' is kept for backwards compatibility and must be set to true. Ignoring configured value")
		UseBearerToken = true
	}
	if !UseTLS {
		logrus.Warn("Flag '--use-tls' is kept for backwards compatibility and must be set to true. Ignoring configured value")
		UseTLS = true
	}
	if IdleTimeout >= 0 && StopRetryPeriod < 0 {
		return fmt.Errorf("invalid value for '--stop-retry-period': must be greater than zero if idling is enabled")
	}
	return nil
}

func setLogLevel() {
	logLevel, isFound := os.LookupEnv("LOG_LEVEL")
	if isFound && len(logLevel) > 0 {
		parsedLevel, err := logrus.ParseLevel(logLevel)
		if err == nil {
			logrus.SetLevel(parsedLevel)
			logrus.Infof("Using log level '%s'", logLevel)
		} else {
			logrus.Errorf("Failed to parse log level '%s'. Possible values: panic, fatal, error, warn, info, debug. Default 'info' is applied", logLevel)
			logrus.SetLevel(logrus.InfoLevel)
		}
	} else {
		logrus.Infof("Using default log level 'info'")
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func printConfig() {
	logrus.Info("Web Terminal Exec configuration:")

	logrus.Infof("==> Debug level %s", logrus.GetLevel().String())
	logrus.Infof("==> Application url %s", URL)
	logrus.Infof("==> Use bearer token: %t", UseBearerToken)
	logrus.Infof("==> Authenticated user ID: %s", AuthenticatedUserID)
	logrus.Infof("==> Pod selector: %s", PodSelector)
	logrus.Infof("==> Idle timeout: %s", IdleTimeout)
	logrus.Infof("==> Stop retry period: %s", StopRetryPeriod)
}
