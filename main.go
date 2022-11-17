//
// Copyright (c) 2019-2022 Red Hat, Inc.
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
	"os"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := config.ParseConfig(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

}
