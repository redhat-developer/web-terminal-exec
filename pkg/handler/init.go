// Copyright (c) 2019-2023 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/redhat-developer/web-terminal-exec/pkg/api"
	"github.com/redhat-developer/web-terminal-exec/pkg/auth"
	"github.com/redhat-developer/web-terminal-exec/pkg/constants"
	"github.com/redhat-developer/web-terminal-exec/pkg/errors"
	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/redhat-developer/web-terminal-exec/pkg/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const createKubeConfigCommandFmt = `
set -ex
echo "test"
if [ -z "$KUBECONFIG" ]; then
	KUBECONFIG_DIR="$HOME/.kube"
	KUBECONFIG_FILE="config"
else
	KUBECONFIG_DIR="$(dirname "$KUBECONFIG")"
	KUBECONFIG_FILE="$(basename "$KUBECONFIG")"
fi
mkdir -p $KUBECONFIG_DIR
cat <<EOF > "$KUBECONFIG_DIR/$KUBECONFIG_FILE"
%s
EOF
`

func (s *Router) handleExecInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params, err := readInitParams(w, r)
	if err != nil {
		handleError(w, err)
		return
	}

	userClient, userConfig, err := s.ClientProvider.NewClientWithToken(params.KubeConfigParams.BearerToken)
	if err != nil {
		logrus.Errorf("Failed to create client: %s", err)
		http.Error(w, "Failed to create API client", http.StatusInternalServerError)
		return
	}

	workspacePod, err := operations.GetCurrentWorkspacePod(userClient)
	if err != nil {
		logrus.Errorf("Failed to get current workspace pod: %s", err)
		http.Error(w, "Failed to find workspace pod", http.StatusInternalServerError)
		return
	}
	logrus.Debugf("Found workspace pod %s", workspacePod.Name)

	containerName, err := getContainerNameForExec(params, workspacePod)
	if err != nil {
		handleError(w, err)
		return
	}
	logrus.Debugf("Found container name %s", containerName)

	kubeconfig, err := util.CreateKubeConfigText(params.BearerToken, params.Namespace, params.Username)
	if err != nil {
		handleError(w, err)
		return
	}
	createKubeConfigCommand := fmt.Sprintf(createKubeConfigCommandFmt, kubeconfig)
	if stdout, stderr, err := operations.ExecCommandInPod(userClient, userConfig, workspacePod.Name, containerName, createKubeConfigCommand); err != nil {
		logrus.Errorf("Failed to create kubeconfig in container %s workspace pod %s: %s", containerName, workspacePod.Name, err)
		logrus.Debugf("Command stdout: %s", stdout.String())
		logrus.Debugf("Command stderr: %s", stderr.String())
		http.Error(w, "Failed to create kubeconfig in pod", http.StatusInternalServerError)
		return
	}
	logrus.Debugf("Created kubeconfig in container %s", containerName)

	shell, err := util.DetectShell(userClient, userConfig, workspacePod.Name, containerName)
	if err != nil {
		handleError(w, err)
	}
	logrus.Debugf("Detected shell %s in container %s", shell, containerName)

	response := api.ExecInitResponse{
		PodName:       workspacePod.Name,
		ContainerName: containerName,
		Cmd:           []string{shell},
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("Failed to marshal json response: %s", err)
		http.Error(w, "Failed to marshal json response", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(responseJson); err != nil {
		logrus.Errorf("Failed to write response to /exec/init request")
	}
}

func getContainerNameForExec(params *api.InitParams, pod *corev1.Pod) (string, error) {
	if params.ContainerName != "" {
		for _, container := range pod.Spec.Containers {
			if container.Name == params.ContainerName {
				return container.Name, nil
			}
		}
		return "", errors.NewHTTPErrorf(http.StatusBadRequest, "container '%s' not found in pod '%s'", params.ContainerName, pod.Name)
	}

	// Attempt to find a container:
	// 1. If there is only one non-exec container in the pod, return its name
	// 2. Otherwise, if web-terminal-tooling container is present, return its name
	// 3. Otherwise, return first container in list
	filteredContainers := filterContainerList(pod.Spec.Containers)
	switch len(filteredContainers) {
	case 0:
		return "", errors.NewHTTPErrorf(http.StatusBadRequest, "no suitable container found in pod '%s'", pod.Name)
	case 1:
		return filteredContainers[0].Name, nil
	default:
		for _, container := range filteredContainers {
			if container.Name == constants.WebTerminalToolingContainerName {
				return container.Name, nil
			}
		}
		return filteredContainers[0].Name, nil
	}
}

func readInitParams(w http.ResponseWriter, r *http.Request) (*api.InitParams, error) {
	params := &api.InitParams{}
	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxBodyBytes)
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			return nil, errors.NewHTTPError(http.StatusRequestEntityTooLarge, "Request body too large")
		}
		return nil, errors.NewInternalErrorf("failed to read exec/init parameters from request: %s", err)
	}
	if len(reqBody) > 0 {
		if err := json.Unmarshal(reqBody, params); err != nil {
			return nil, errors.NewInternalErrorf("failed to unmarshal exec/init parameters from request: %s", err)
		}
	}
	token, err := auth.ExtractToken(r)
	if err != nil {
		return nil, errors.NewHTTPErrorf(http.StatusUnauthorized, "failed to get token from request: %s", err)
	}
	params.BearerToken = token
	// Set defaults
	if params.Username == "" {
		params.Username = "Developer"
	}
	return params, nil
}

func filterContainerList(containers []corev1.Container) []corev1.Container {
	var filtered []corev1.Container
	for _, container := range containers {
		if container.Name == constants.WebTerminalExecContainerName {
			continue
		}
		filtered = append(filtered, container)
	}
	return filtered
}
