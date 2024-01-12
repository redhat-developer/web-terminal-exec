// Copyright (c) 2019-2024 Red Hat, Inc.
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
	"testing"

	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestDetectShell(t *testing.T) {
	const testPodName, testContainerName = "test-pod", "test-container"
	mapContains := func(cmdMap map[string]string, contains string) bool {
		for cmd := range cmdMap {
			if contains == cmd {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name               string
		commandsAndOutputs map[string]string
		erroredCommands    []string
		expectedShell      string
		errRegexp          string
	}{
		{
			name: "Resolves shell from env var",
			commandsAndOutputs: map[string]string{
				"echo $SHELL": "myshell",
			},
			expectedShell: "myshell",
		},
		{
			name: "Strips trailing newline from SHELL env var",
			commandsAndOutputs: map[string]string{
				"echo $SHELL": "myshell\n",
			},
			expectedShell: "myshell",
		},
		{
			name: "Resolves shell from /etc/passwd when SHELL env var is not available",
			commandsAndOutputs: map[string]string{
				"echo $SHELL":     "",
				"id -u":           "1234",
				"cat /etc/passwd": "user:x:1234:0:user user:/home/user:/bin/myshell",
			},
			expectedShell: "/bin/myshell",
		},
		{
			name: "Resolves trailing newline from command output",
			commandsAndOutputs: map[string]string{
				"echo $SHELL":     "\n",
				"id -u":           "1234\n",
				"cat /etc/passwd": "user:x:1234:0:user user:/home/user:/bin/myshell\n",
			},
			expectedShell: "/bin/myshell",
		},
		{
			name: "Error reading SHELL env var",
			commandsAndOutputs: map[string]string{
				"echo $SHELL":     "\n",
				"id -u":           "1234\n",
				"cat /etc/passwd": "user:x:1234:0:user user:/home/user:/bin/myshell\n",
			},
			erroredCommands: []string{"echo $SHELL"},
			expectedShell:   "/bin/myshell",
		},
		{
			name: "Error reading user ID",
			commandsAndOutputs: map[string]string{
				"echo $SHELL":     "\n",
				"id -u":           "1234\n",
				"cat /etc/passwd": "user:x:1234:0:user user:/home/user:/bin/myshell\n",
			},
			erroredCommands: []string{"id -u"},
			errRegexp:       "failed to get user ID",
		},
		{
			name: "Error reading /etc/passwd",
			commandsAndOutputs: map[string]string{
				"echo $SHELL":     "\n",
				"id -u":           "1234\n",
				"cat /etc/passwd": "user:x:1234:0:user user:/home/user:/bin/myshell\n",
			},
			erroredCommands: []string{"cat /etc/passwd"},
			errRegexp:       "failed to read /etc/passwd",
		},
		{
			name: "Unparseable /etc/passwd",
			commandsAndOutputs: map[string]string{
				"echo $SHELL":     "\n",
				"id -u":           "1234\n",
				"cat /etc/passwd": "user:x:/bin/myshell\n",
			},
			errRegexp: "failed to parse shell from /etc/passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			fakeClient := &test.WrapFakeClientCoreV1{}
			fakeClient.Clientset = client

			fakeSPDY := test.FakeSPDYExecutorProvider{
				FakeSPDYExecutor: test.FakeSPDYExecutor{
					ResponseOutputs: tt.commandsAndOutputs,
					ErrInputs:       tt.erroredCommands,
				},
			}
			oldSPDYExecutor := operations.NewSPDYExecutor
			operations.NewSPDYExecutor = fakeSPDY.NewFakeSPDYExecutor
			defer func() { operations.NewSPDYExecutor = oldSPDYExecutor }()

			result, err := DetectShell(fakeClient, &rest.Config{}, testPodName, testContainerName)
			if tt.errRegexp != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.errRegexp, err.Error())
			} else {
				assert.Equal(t, tt.expectedShell, result)
				for _, ranCommand := range fakeSPDY.InputBuffers {
					assert.True(t, mapContains(tt.commandsAndOutputs, ranCommand), "Unexpected command ran")
				}
			}
		})
	}
}

func TestParseShellFromEtcPass(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		etcPass       string
		expectedShell string
		errRegexp     string
	}{
		{
			name:          "one-line /etc/passwd",
			userID:        "1234",
			etcPass:       `user:x:1234:0:user user:/home/user:/bin/bash`,
			expectedShell: "/bin/bash",
		},
		{
			name:          "parses whitespace in shell",
			userID:        "1234",
			etcPass:       `user:x:1234:0:user user:/home/user:/bin/my special bash/bash`,
			expectedShell: "/bin/my special bash/bash",
		},
		{
			name:      "no userID in /etc/passwd",
			userID:    "9999",
			etcPass:   `user:x:1234:0:user user:/home/user:/bin/bash`,
			errRegexp: "failed to parse shell from /etc/passwd",
		},
		{
			name:      "invalid /etc/passwd",
			userID:    "9999",
			etcPass:   `user:x:/home/user:/bin/bash`,
			errRegexp: "failed to parse shell from /etc/passwd",
		},
		{
			name:   "multi-line /etc/passwd",
			userID: "1234",
			etcPass: `
				root:x:0:0:root:/root:/bin/bash
				bin:x:1:1:bin:/bin:/sbin/nologin
				user:x:1234:0:user user:/home/user:/bin/myshell
				daemon:x:2:2:daemon:/sbin:/sbin/nologin
				adm:x:3:4:adm:/var/adm:/sbin/nologin`,
			expectedShell: "/bin/myshell",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseShellFromEtcPasswd(tt.etcPass, tt.userID)
			if tt.errRegexp != "" {
				assert.Error(t, err)
				assert.Regexp(t, tt.errRegexp, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedShell, result)
			}
		})
	}
}
