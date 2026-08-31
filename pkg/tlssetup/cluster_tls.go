//
// Copyright (c) 2019-2026 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package tlssetup

import (
	"context"
	"crypto/tls"

	configv1 "github.com/openshift/api/config/v1"
	ostls "github.com/openshift/controller-runtime-common/pkg/tls"
	libgocrypto "github.com/openshift/library-go/pkg/crypto"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BuildClusterTLSConfig fetches TLS settings from the OpenShift APIServer and builds tls.Config.
// Returns nil on non-OpenShift clusters.
// Falls back to library-go default TLS profile when adherence policy doesn't require strict adherence.
func BuildClusterTLSConfig(ctx context.Context) *tls.Config {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		logrus.WithError(err).Info("Not running in cluster; using Go default TLS configuration")
		return nil
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = configv1.AddToScheme(scheme)

	defaultProfile := *configv1.TLSProfiles[libgocrypto.DefaultTLSProfileType]

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		logrus.WithError(err).Error("Failed to create Kubernetes client; using library-go default TLS configuration")
		return buildTLSConfigFromProfile(defaultProfile)
	}

	adherence, err := ostls.FetchAPIServerTLSAdherencePolicy(ctx, k8sClient)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch TLS adherence policy; using library-go default TLS configuration")
		return buildTLSConfigFromProfile(defaultProfile)
	}

	if !libgocrypto.ShouldHonorClusterTLSProfile(adherence) {
		logrus.WithFields(logrus.Fields{
			"minTLSVersion":   defaultProfile.MinTLSVersion,
			"adherencePolicy": adherence,
		}).Info("Using library-go default TLS profile")
		return buildTLSConfigFromProfile(defaultProfile)
	}

	profile := fetchTLSProfile(ctx, k8sClient, defaultProfile)

	logrus.WithFields(logrus.Fields{
		"minTLSVersion":   profile.MinTLSVersion,
		"adherencePolicy": adherence,
	}).Info("Applying cluster TLS profile to HTTPS server")

	return buildTLSConfigFromProfile(profile)
}

func fetchTLSProfile(ctx context.Context, k8sClient client.Client, fallback configv1.TLSProfileSpec) configv1.TLSProfileSpec {
	profile, err := ostls.FetchAPIServerTLSProfile(ctx, k8sClient)
	if err != nil {
		logrus.WithError(err).Info("Failed to fetch TLS profile from APIServer; using library-go default TLS configuration")
		return fallback
	}
	return profile
}

func buildTLSConfigFromProfile(profile configv1.TLSProfileSpec) *tls.Config {
	tlsConfig := &tls.Config{}
	tlsConfigFn, unsupported := ostls.NewTLSConfigFromProfile(profile)
	if len(unsupported) > 0 {
		logrus.WithField("unsupported", unsupported).Info("TLS profile contains ciphers unsupported by Go")
	}

	if len(profile.Ciphers) > 0 && len(unsupported) == len(profile.Ciphers) {
		logrus.WithField("profileCiphers", profile.Ciphers).Error("No ciphers from the cluster TLS profile are supported by Go; server will use library-go defaults, which may not satisfy tlsAdherence")
	}

	tlsConfigFn(tlsConfig)
	return tlsConfig
}
