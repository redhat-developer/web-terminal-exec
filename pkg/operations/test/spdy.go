// Copyright (c) 2019-2023 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package test

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// FakeSPDYExecutorProvider provides a function that allows replacing remotecommand.NewSPDYExecutor
type FakeSPDYExecutorProvider struct {
	FakeSPDYExecutor
}

func (f *FakeSPDYExecutorProvider) NewFakeSPDYExecutor(_ *rest.Config, _ string, _ *url.URL) (remotecommand.Executor, error) {
	return &f.FakeSPDYExecutor, nil
}

// FakeSPDYExecutor is a fake SPDYExecutor that allows configuring output for a given input
// Inputs are saved in order to allow verifying commands that were executed.
type FakeSPDYExecutor struct {
	InputBuffers    []string
	ResponseOutputs map[string]string
	ResponseStdErr  map[string]string
	ErrInputs       []string
}

var _ remotecommand.Executor = (*FakeSPDYExecutor)(nil)

func (f *FakeSPDYExecutor) Stream(options remotecommand.StreamOptions) error {
	stdinBytes, err := io.ReadAll(options.Stdin)
	if err != nil {
		return fmt.Errorf("(TEST) failed to read stdin from command: %w", err)
	}
	stdin := string(stdinBytes)
	f.InputBuffers = append(f.InputBuffers, string(stdin))

	if output, ok := f.ResponseOutputs[stdin]; ok {
		_, err := options.Stdout.Write([]byte(output))
		if err != nil {
			return fmt.Errorf("(TEST) failed to write to stdout: %w", err)
		}
	}

	if outerr, ok := f.ResponseStdErr[stdin]; ok {
		_, err := options.Stderr.Write([]byte(outerr))
		if err != nil {
			return fmt.Errorf("(TEST) failed to write to stdout: %w", err)
		}
	}

	for _, badInput := range f.ErrInputs {
		if strings.Contains(stdin, badInput) {
			return fmt.Errorf("bad input in test")
		}
	}

	return nil
}
