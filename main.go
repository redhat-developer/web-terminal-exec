//
// Copyright (c) 2019-2023 Red Hat, Inc.
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
	"net/http"
	"os"

	"github.com/redhat-developer/web-terminal-exec/pkg/activity"
	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/redhat-developer/web-terminal-exec/pkg/constants"
	"github.com/redhat-developer/web-terminal-exec/pkg/handler"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/sirupsen/logrus"
)

const (
	TLSCertFile = "/var/serving-cert/tls.crt"
	TLSKeyFile  = "/var/serving-cert/tls.key"
)

func main() {
	if err := config.ParseConfig(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	clientProvider := operations.DefaultClientProvider()

	activityManager, err := activity.NewActivityManager(config.IdleTimeout, config.StopRetryPeriod, clientProvider)
	if err != nil {
		logrus.Errorf("Unable to create activity manager: %s", err)
		os.Exit(1)
	}
	activityManager.Start()

	router := handler.Router{
		ActivityManager: activityManager,
		ClientProvider:  clientProvider,
	}

	server := http.Server{
		Addr:           config.URL,
		Handler:        router.HTTPSHandler(),
		ReadTimeout:    constants.ServerReadTimeout,
		WriteTimeout:   constants.ServerWriteTimeout,
		MaxHeaderBytes: constants.MaxHeaderBytes,
	}
	if err := server.ListenAndServeTLS(TLSCertFile, TLSKeyFile); err != nil {
		logrus.Errorf("Failed to start server with TLS enabled: %s", err)
	}
}
