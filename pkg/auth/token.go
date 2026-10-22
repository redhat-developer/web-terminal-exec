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
	"fmt"
	"net/http"
	"strings"
)

const (
	accessTokenHeader          = "X-Access-Token"
	forwardedAccessTokenHeader = "X-Forwarded-Access-Token"
)

func ExtractToken(r *http.Request) (string, error) {
	token := r.Header.Get(accessTokenHeader)
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
		return token, nil
	}

	token = r.Header.Get(forwardedAccessTokenHeader)
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("authorization header is missing")
}
