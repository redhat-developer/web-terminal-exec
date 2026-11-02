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
	"strconv"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/sirupsen/logrus"
)

var supportedTLSVersions = map[uint16]bool{
	tls.VersionTLS10: true,
	tls.VersionTLS11: true,
	tls.VersionTLS12: true,
	tls.VersionTLS13: true,
}

func buildTLSConfig() *tls.Config {
	if config.TLSMinVersion == "" {
		return nil
	}
	minVersion, err := strconv.ParseUint(config.TLSMinVersion, 10, 16)
	if err != nil {
		logrus.Warnf("Invalid TLS min version %q, using Go default", config.TLSMinVersion)
		return nil
	}
	v := uint16(minVersion)
	if !supportedTLSVersions[v] {
		logrus.Warnf("Unsupported TLS version %q, ignoring", config.TLSMinVersion)
		return nil
	}
	logrus.Infof("Setting TLS min version to %s", tls.VersionName(v))
	return &tls.Config{MinVersion: v}
}
