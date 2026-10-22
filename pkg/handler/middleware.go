// Copyright (c) 2019-2025 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package handler

import (
	"net/http"
	"time"

	"github.com/redhat-developer/web-terminal-exec/pkg/auth"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/sirupsen/logrus"
)

type middleware interface {
	addMiddleware(http.Handler) http.Handler
}

type logRequestMiddleware struct {
	path string
}

func (m *logRequestMiddleware) addMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		handler.ServeHTTP(w, r)
		duration := time.Since(startTime)
		logrus.WithFields(logrus.Fields{
			"endpoint": m.path,
			"duration": duration.String(),
			"method":   r.Method,
		}).Info()
	})
}

type authMiddleware struct {
	clientProvider operations.ClientProvider
}

func (m *authMiddleware) addMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth.Authenticate(r, m.clientProvider); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
