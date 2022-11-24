// Copyright (c) 2019-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package api

type InitParams struct {
	ContainerName    string `json:"container"` // optional, Will be first suitable container in pod if not set
	KubeConfigParams `json:"kubeconfig"`
}

type KubeConfigParams struct {
	Namespace   string `json:"namespace"`   //optional, Is not set into kubeconfig file if is not set or empty
	Username    string `json:"username"`    //optional, Developer in kubeconfig if empty
	BearerToken string `json:"bearertoken"` //evaluated from header
}

type ExecInitResponse struct {
	PodName       string   `json:"pod"`
	ContainerName string   `json:"container"`
	Cmd           []string `json:"cmd"`
}
