// Copyright (c) 2019-2023 Red Hat, Inc.
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
	"regexp"
	"strings"

	"github.com/redhat-developer/web-terminal-exec/pkg/errors"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	getShellCommand     = "echo $SHELL"
	getUserIdCommand    = "id -u"
	catEtcPasswdCommand = "cat /etc/passwd"
)

func DetectShell(client kubernetes.Interface, restconfig *rest.Config, podName, containerName string) (string, error) {
	// Try to get shell from $SHELL env var
	stdout, stderr, err := operations.ExecCommandInPod(client, restconfig, podName, containerName, getShellCommand)
	if err == nil {
		shellEnv := strings.TrimSuffix(stdout.String(), "\n")
		logrus.Debugf("Detected shell '%s' from $SHELL environment variable", shellEnv)
		if shellEnv != "" {
			return shellEnv, nil
		}
	} else {
		logrus.Infof("Failed to read $SHELL environment variable in container %s in pod %s: %s", containerName, podName, err)
		logrus.Debugf("Command stdout: %s", stdout.String())
		logrus.Debugf("Command stderr: %s", stderr.String())
	}

	// Try to read shell from /etc/passwd directly
	stdout, stderr, err = operations.ExecCommandInPod(client, restconfig, podName, containerName, getUserIdCommand)
	if err != nil {
		logrus.Errorf("Failed to get user ID in container %s in pod %s: %s", containerName, podName, err)
		logrus.Debugf("Command stdout: %s", stdout.String())
		logrus.Debugf("Command stderr: %s", stderr.String())
		return "", errors.NewInternalErrorf("failed to get user ID in container %s in pod %s", containerName, podName)
	}
	userID := strings.TrimSuffix(stdout.String(), "\n")
	logrus.Debugf("Detected user ID: '%s'", userID)

	stdout, stderr, err = operations.ExecCommandInPod(client, restconfig, podName, containerName, catEtcPasswdCommand)
	if err != nil {
		logrus.Errorf("Failed to read /etc/passwd in container %s in pod %s: %s", containerName, podName, err)
		logrus.Debugf("Command stdout: %s", stdout.String())
		logrus.Debugf("Command stderr: %s", stderr.String())
		return "", errors.NewInternalErrorf("failed to read /etc/passwd in container %s in pod %s", containerName, podName)
	}

	etcPasswd := stdout.String()
	logrus.Debugf("Content of /etc/passwd: '%s'", etcPasswd)

	shell, err := parseShellFromEtcPasswd(etcPasswd, userID)
	if err != nil {
		logrus.Errorf("Error parsing /etc/passwd: %s", err)
		return "", errors.NewInternalErrorf("failed to parse shell from /etc/passwd in container %s", containerName)
	}
	logrus.Debugf("Detected shell %s from /etc/passwd", shell)

	return shell, nil
}

// Use content of the "/etc/passwd" file to parse shell by user ID.
// For each user /etc/passwd file stores information in the separated line. Information split with help ":".
// Row information:
// - User name
// - Encrypted password
// - User ID number (UID)
// - User's group ID number (GID)
// - Full name of the user (GECOS)
// - User home directory
// - Login shell
// Read more: https://www.ibm.com/support/knowledgecenter/en/ssw_aix_72/com.ibm.aix.security/passwords_etc_passwd_file.htm
func parseShellFromEtcPasswd(etcPasswd, userID string) (string, error) {
	re, err := regexp.Compile(fmt.Sprintf(`.*:.*:%s:.*:.*:.*:(?P<ShellPath>.*)`, userID))
	if err != nil {
		return "", err
	}
	result := re.FindStringSubmatch(etcPasswd)
	if len(result) != 2 || result[1] == "" {
		return "", fmt.Errorf("failed to parse shell from /etc/passwd")
	}
	return result[1], nil
}
