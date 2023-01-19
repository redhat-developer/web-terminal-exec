// Copyright (c) 2019-2023 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package operations

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/web-terminal-exec/pkg/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func StopDevWorkspace(devworkspaceClient dynamic.Interface) error {
	stopWorkspacePatch := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"controller.devfile.io/stopped-by": "inactivity",
				},
			},
			"spec": map[string]interface{}{
				"started": false,
			},
		},
	}
	patchJSON, err := stopWorkspacePatch.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = devworkspaceClient.Resource(devworkspaceGVR).Namespace(config.DevWorkspaceNamespace).Patch(context.TODO(), config.DevWorkspaceName, types.MergePatchType, patchJSON, v1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch DevWorkspace: %s", err)
	}

	return nil
}

func ExecCommandInPod(client kubernetes.Interface, restconfig *rest.Config, podName, containerName, command string) (stdout, stderr *bytes.Buffer, err error) {
	req := client.CoreV1().RESTClient().
		Post().
		Namespace(config.DevWorkspaceNamespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"/bin/sh"},
			Stdout:    true,
			Stderr:    true,
			Stdin:     true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up executor for command: %s", err)
	}

	input := strings.NewReader(command)
	var outBuf, errBuf bytes.Buffer
	if err := executor.Stream(remotecommand.StreamOptions{
		Stdin:  input,
		Stdout: &outBuf,
		Stderr: &errBuf,
	}); err != nil {
		return &outBuf, &errBuf, fmt.Errorf("error executing command in container: %s", err)
	}
	return &outBuf, &errBuf, nil
}

func GetCurrentWorkspacePod(client kubernetes.Interface) (*corev1.Pod, error) {
	filterOptions := metav1.ListOptions{LabelSelector: config.PodSelector, FieldSelector: "status.phase=Running"}
	podList, err := client.CoreV1().Pods(config.DevWorkspaceNamespace).List(context.TODO(), filterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace '%s': %s", config.DevWorkspaceNamespace, err)
	}
	switch len(podList.Items) {
	case 0:
		return nil, fmt.Errorf("no workspace pods found in namespace '%s'", config.DevWorkspaceNamespace)
	case 1:
		return &podList.Items[0], nil
	default:
		// Multiple pods found -- try to get pod that exec is running in; may occur if dedicated pods are used
		// Workaround as current pod name is not available -- hostname is substitute
		podName := os.Getenv("HOSTNAME")
		if podName == "" {
			return &podList.Items[0], nil
		}
		for idx, pod := range podList.Items {
			if pod.Name == podName {
				return &podList.Items[idx], nil
			}
		}
		return nil, fmt.Errorf("failed to get current workspace pod")
	}
}

func GetCurrentUserUID(token string, clientProvider ClientProvider) (string, error) {
	userClient, _, err := clientProvider.NewOpenShiftUserClient(token)
	if err != nil {
		return "", fmt.Errorf("failed to create client to check user info")
	}
	userInfo, err := userClient.Resource(userGVR).Namespace("").Get(context.TODO(), "~", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get current user information: %s", err)
	}

	return string(userInfo.GetUID()), nil
}
