// Copyright (c) 2019-2022 Red Hat, Inc.
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
	"net"
	"os"

	"github.com/redhat-developer/web-terminal-exec/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type KubeConfig struct {
	APIVersion     string     `json:"apiVersion"`
	Clusters       []Clusters `json:"clusters"`
	Users          []Users    `json:"users"`
	Contexts       []Contexts `json:"contexts"`
	CurrentContext string     `json:"current-context"`
	Kind           string     `json:"kind"`
}

type Clusters struct {
	Cluster ClusterInfo `json:"cluster"`
	Name    string      `json:"name"`
}

type ClusterInfo struct {
	CertificateAuthority string `json:"certificate-authority"`
	Server               string `json:"server"`
}

type Users struct {
	Name string `json:"name"`
	User User   `json:"user"`
}

type User struct {
	Token string `json:"token"`
}

type Contexts struct {
	Context Context `json:"context"`
	Name    string  `json:"name"`
}

type Context struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	User      string `json:"user"`
}

func CreateKubeConfigText(token, namespace, username string) (string, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return "", errors.NewInternalError("Could not find $KUBERNETES_SERVICE_HOST or $KUBERNETES_SERVICE_PORT")
	}

	server := "https://" + net.JoinHostPort(host, port)
	kubeconfig := generateKubeConfig(token, server, namespace, username)

	bytes, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		logrus.Errorf("Failed to marshal kubeconfig: %s", err)
		return "", err
	}
	return string(bytes), nil
}

func generateKubeConfig(token, server, namespace, username string) *KubeConfig {
	currentContext := fmt.Sprintf("%s-context", username)
	return &KubeConfig{
		APIVersion: "v1",
		Clusters: []Clusters{
			{
				Cluster: ClusterInfo{
					CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
					Server:               server,
				},
				Name: server,
			},
		},
		Users: []Users{
			{
				Name: username,
				User: User{
					Token: token,
				},
			},
		},
		Contexts: []Contexts{
			{
				Context: Context{
					Cluster:   server,
					Namespace: namespace,
					User:      username,
				},
				Name: currentContext,
			},
		},
		CurrentContext: currentContext,
		Kind:           "Config",
	}
}
