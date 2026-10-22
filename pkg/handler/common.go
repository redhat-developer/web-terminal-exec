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

	"github.com/redhat-developer/web-terminal-exec/pkg/activity"
	"github.com/redhat-developer/web-terminal-exec/pkg/constants"
	"github.com/redhat-developer/web-terminal-exec/pkg/errors"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
)

type Router struct {
	ActivityManager activity.ActivityManager
	ClientProvider  operations.ClientProvider
}

func (s *Router) HTTPSHandler() http.Handler {
	mux := http.NewServeMux()
	handle := func(path string, handler http.Handler, middlewares ...middleware) {
		loggingMiddleware := logRequestMiddleware{path}
		composedHandler := handler
		// Note this will work in reverse order (i.e. the middleware specified last will be used first)
		for _, m := range middlewares {
			composedHandler = m.addMiddleware(composedHandler)
		}
		composedHandler = loggingMiddleware.addMiddleware(composedHandler)
		mux.Handle(path, composedHandler)
	}
	handleFunc := func(path string, handler http.HandlerFunc, middlewares ...middleware) {
		handle(path, handler, middlewares...)
	}

	// Serve /activity/tick endpoint
	handleFunc(constants.ActivityTickEndpoint, s.handleActivityTick, &authMiddleware{s.ClientProvider})

	// Serve /exec/init endpoint
	handleFunc(constants.ExecInitEndpoint, s.handleExecInit, &authMiddleware{s.ClientProvider})

	// Serve /healthz endpoint
	handleFunc(constants.HealthzEndpoint, s.handleHealthCheck)
	return http.Handler(mux)
}

func handleError(w http.ResponseWriter, err error) {
	switch t := err.(type) {
	case *errors.HTTPError:
		http.Error(w, t.Message, t.StatusCode)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
