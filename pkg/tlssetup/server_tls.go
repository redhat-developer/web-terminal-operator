//
// Copyright (c) 2021-2026 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package tlssetup

import (
	"context"
	"crypto/tls"
	"math"
	"time"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	ostls "github.com/openshift/controller-runtime-common/pkg/tls"
	libgocrypto "github.com/openshift/library-go/pkg/crypto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var apiServerGVR = schema.GroupVersionResource{
	Group:    "config.openshift.io",
	Version:  "v1",
	Resource: "apiservers",
}

type ServerTLS struct {
	MinTLSVersion             uint16
	InitialTLSProfileSpec     configv1.TLSProfileSpec
	InitialTLSAdherencePolicy configv1.TLSAdherencePolicy
}

func BuildServerTLSOptions(ctx context.Context, client crclient.Client, log logr.Logger) (ServerTLS, error) {
	var profile configv1.TLSProfileSpec
	var adherence configv1.TLSAdherencePolicy

	apiServer := &configv1.APIServer{}
	if err := client.Get(ctx, crclient.ObjectKey{Name: ostls.APIServerName}, apiServer); err != nil {
		log.Error(err, "Failed to read APIServer/cluster, falling back to library-go default TLS profile")
	} else if p, err := ostls.GetTLSProfileSpec(apiServer.Spec.TLSSecurityProfile); err != nil {
		log.Error(err, "Failed to resolve TLS profile spec, falling back to library-go default TLS profile")
	} else {
		profile = p
		adherence = apiServer.Spec.TLSAdherence
	}

	serverTLS := ServerTLS{
		InitialTLSProfileSpec:     profile,
		InitialTLSAdherencePolicy: adherence,
	}

	if libgocrypto.ShouldHonorClusterTLSProfile(adherence) {
		tlsConfigFn, unsupported := ostls.NewTLSConfigFromProfile(profile)
		if len(unsupported) > 0 {
			log.Info("TLS profile contains ciphers unsupported by Go", "unsupported", unsupported)
		}

		if len(profile.Ciphers) > 0 && len(unsupported) == len(profile.Ciphers) {
			log.Error(nil, "No ciphers from the cluster TLS profile are supported by Go; server will use library-go defaults, which may not satisfy tlsAdherence",
				"profileCiphers", profile.Ciphers)
		}

		cfg := &tls.Config{}
		tlsConfigFn(cfg)
		serverTLS.MinTLSVersion = cfg.MinVersion

		log.Info("Applying cluster TLS profile",
			"minTLSVersion", profile.MinTLSVersion)
		log.V(1).Info("TLS cipher list from cluster profile", "ciphers", profile.Ciphers)
	} else {
		defaultProfile := *configv1.TLSProfiles[libgocrypto.DefaultTLSProfileType]
		defaultTLSConfigFn, unsupported := ostls.NewTLSConfigFromProfile(defaultProfile)
		if len(unsupported) > 0 {
			log.Info("Default TLS profile contains ciphers unsupported by Go", "unsupported", unsupported)
		}

		cfg := &tls.Config{}
		defaultTLSConfigFn(cfg)
		serverTLS.MinTLSVersion = cfg.MinVersion

		log.Info("Using library-go default TLS profile",
			"minTLSVersion", defaultProfile.MinTLSVersion,
			"adherencePolicy", adherence)
	}

	return serverTLS, nil
}

func RegisterSecurityProfileWatcher(cfg *rest.Config, initial ServerTLS, onCancel context.CancelFunc, log logr.Logger) error {
	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}
	go watchAPIServer(dynClient, initial, onCancel, log)
	return nil
}

func watchAPIServer(dynClient dynamic.Interface, initial ServerTLS, onCancel context.CancelFunc, log logr.Logger) {
	const (
		maxRetries     = 10
		initialBackoff = 30 * time.Second
		maxBackoff     = 5 * time.Minute
	)
	retries := 0
	for {
		watcher, err := dynClient.Resource(apiServerGVR).Watch(context.Background(), metav1.ListOptions{
			FieldSelector: "metadata.name=" + ostls.APIServerName,
		})
		if err != nil {
			retries++
			if retries > maxRetries {
				log.Error(err, "Failed to watch APIServer, max retries exceeded, giving up")
				return
			}
			backoff := time.Duration(math.Min(float64(initialBackoff)*math.Pow(2, float64(retries-1)), float64(maxBackoff)))
			log.Error(err, "Failed to watch APIServer, retrying", "attempt", retries, "maxRetries", maxRetries, "backoff", backoff)
			time.Sleep(backoff)
			continue
		}
		log.Info("APIServer watcher established successfully")
		retries = 0

		for event := range watcher.ResultChan() {
			if event.Type != watch.Modified {
				continue
			}
			unstrObj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			apiServer := &configv1.APIServer{}
			if err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstrObj.Object, apiServer); err != nil {
				log.Error(err, "Failed to convert watch event to APIServer")
				continue
			}

			if apiServer.Spec.TLSAdherence != initial.InitialTLSAdherencePolicy {
				log.Info("TLS adherence policy changed, restarting operator",
					"adherencePolicy", apiServer.Spec.TLSAdherence)
				onCancel()
				return
			}

			if !libgocrypto.ShouldHonorClusterTLSProfile(initial.InitialTLSAdherencePolicy) {
				log.V(1).Info("Cluster TLS profile changed but adherence policy is not strict, not restarting")
				continue
			}

			newProfile, err := ostls.GetTLSProfileSpec(apiServer.Spec.TLSSecurityProfile)
			if err != nil {
				log.Error(err, "Failed to resolve TLS profile from watch event")
				continue
			}
			tlsConfigFn, _ := ostls.NewTLSConfigFromProfile(newProfile)
			cfg := &tls.Config{}
			tlsConfigFn(cfg)
			if cfg.MinVersion != initial.MinTLSVersion {
				log.Info("TLS security profile changed, restarting operator",
					"oldMinTLSVersion", initial.MinTLSVersion,
					"newMinTLSVersion", cfg.MinVersion)
				onCancel()
				return
			}
		}
		log.Info("APIServer watch closed, reconnecting")
	}
}
