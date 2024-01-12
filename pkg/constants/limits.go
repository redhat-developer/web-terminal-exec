// Copyright (c) 2019-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package constants

import "time"

const (
	MaxBodyBytes       = 1 << 20  // 1 MiB
	MaxHeaderBytes     = 16 << 10 // 16 KiB
	ServerReadTimeout  = 3 * time.Second
	ServerWriteTimeout = 3 * time.Second
)
