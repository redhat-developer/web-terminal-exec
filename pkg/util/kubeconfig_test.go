// Copyright (c) 2019-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKubeConfigText(t *testing.T) {
	token, namespace, username := "test-token", "test-namespace", "test-username"
	testHost, testPort := "0.0.0.0", "9999"
	t.Setenv("KUBERNETES_SERVICE_HOST", testHost)
	t.Setenv("KUBERNETES_SERVICE_PORT", testPort)
	result, err := CreateKubeConfigText(token, namespace, username)
	assert.NoError(t, err)
	assert.Contains(t, result, "token: "+token)
	assert.Contains(t, result, "namespace: "+namespace)
	assert.Contains(t, result, "user: "+username)
	hostAndPort := fmt.Sprintf("https://%s:%s", testHost, testPort)
	assert.Contains(t, result, "cluster: "+hostAndPort)
}
