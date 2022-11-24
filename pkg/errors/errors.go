// Copyright (c) 2019-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package errors

import "fmt"

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func NewHTTPErrorf(statusCode int, messageFmt string, args ...interface{}) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf(messageFmt, args...),
	}
}

type InternalError struct {
	errMsg string
}

func (e InternalError) Error() string {
	return e.errMsg
}

func NewInternalError(message string) *InternalError {
	return &InternalError{message}
}

func NewInternalErrorf(messageFmt string, args ...interface{}) *InternalError {
	return &InternalError{fmt.Sprintf(messageFmt, args...)}
}
