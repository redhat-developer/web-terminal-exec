// Copyright (c) 2019-2023 Red Hat, Inc.
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

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/sirupsen/logrus"
)

func Authenticate(r *http.Request, clientProvider operations.ClientProvider) error {
	token, err := ExtractToken(r)
	if err != nil {
		return err
	}
	uid, err := operations.GetCurrentUserUID(token, clientProvider)
	if err != nil {
		return fmt.Errorf("unable to verify user: %s", err)
	}
	if uid != config.AuthenticatedUserID {
		logrus.Debugf("User failed to authenticate: authorized user = '%s', requested user = '%s'", config.AuthenticatedUserID, uid)
		return fmt.Errorf("the current user is not authorized to access this web terminal")
	}
	logrus.Debugf("User '%s' authenticated", uid)
	return nil
}
