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
	"testing"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	ostls "github.com/openshift/controller-runtime-common/pkg/tls"
	libgocrypto "github.com/openshift/library-go/pkg/crypto"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func defaultMinTLSVersion() uint16 {
	profile := *configv1.TLSProfiles[libgocrypto.DefaultTLSProfileType]
	tlsConfigFn, _ := ostls.NewTLSConfigFromProfile(profile)
	cfg := &tls.Config{}
	tlsConfigFn(cfg)
	return cfg.MinVersion
}

func newScheme() *k8sruntime.Scheme {
	scheme := k8sruntime.NewScheme()
	_ = configv1.Install(scheme)
	return scheme
}

func newFakeClient(objs ...crclient.Object) crclient.Client {
	return fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(objs...).Build()
}

func makeAPIServer(profileType configv1.TLSProfileType, adherence configv1.TLSAdherencePolicy) *configv1.APIServer {
	apiServer := &configv1.APIServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.APIServerSpec{
			TLSAdherence: adherence,
		},
	}
	if profileType == configv1.TLSProfileCustomType {
		apiServer.Spec.TLSSecurityProfile = &configv1.TLSSecurityProfile{
			Type: configv1.TLSProfileCustomType,
		}
	} else if profileType != "" {
		apiServer.Spec.TLSSecurityProfile = &configv1.TLSSecurityProfile{
			Type: profileType,
		}
	}
	return apiServer
}

// --- BuildServerTLSOptions tests ---

func TestBuildServerTLSOptions_APIServerNotFound(t *testing.T) {
	client := newFakeClient()
	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion %d (TLS 1.2), got %d", tls.VersionTLS12, result.MinTLSVersion)
	}
	if result.InitialTLSAdherencePolicy != "" {
		t.Errorf("expected empty adherence policy, got %q", result.InitialTLSAdherencePolicy)
	}
}

func TestBuildServerTLSOptions_StrictAdherenceAppliesClusterProfile(t *testing.T) {
	apiServer := makeAPIServer(configv1.TLSProfileModernType, configv1.TLSAdherencePolicyStrictAllComponents)
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinTLSVersion != tls.VersionTLS13 {
		t.Errorf("expected MinTLSVersion %d (TLS 1.3 from Modern profile), got %d", tls.VersionTLS13, result.MinTLSVersion)
	}
	if result.InitialTLSAdherencePolicy != configv1.TLSAdherencePolicyStrictAllComponents {
		t.Errorf("expected adherence %q, got %q", configv1.TLSAdherencePolicyStrictAllComponents, result.InitialTLSAdherencePolicy)
	}
}

func TestBuildServerTLSOptions_NonStrictAdherenceUsesDefault(t *testing.T) {
	apiServer := makeAPIServer(configv1.TLSProfileOldType, configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly)
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion %d (TLS 1.2, library-go default), got %d", tls.VersionTLS12, result.MinTLSVersion)
	}
	if result.InitialTLSAdherencePolicy != configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly {
		t.Errorf("expected adherence %q, got %q", configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly, result.InitialTLSAdherencePolicy)
	}
}

func TestBuildServerTLSOptions_InvalidCustomProfileFallsBack(t *testing.T) {
	apiServer := makeAPIServer(configv1.TLSProfileCustomType, configv1.TLSAdherencePolicyStrictAllComponents)
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion %d (TLS 1.2, fallback default), got %d", tls.VersionTLS12, result.MinTLSVersion)
	}
	if result.InitialTLSAdherencePolicy != configv1.TLSAdherencePolicyStrictAllComponents {
		t.Errorf("expected adherence %q preserved despite profile resolution failure, got %q", configv1.TLSAdherencePolicyStrictAllComponents, result.InitialTLSAdherencePolicy)
	}
}

func TestBuildServerTLSOptions_NoOpinionAdherenceUsesDefault(t *testing.T) {
	apiServer := makeAPIServer(configv1.TLSProfileIntermediateType, configv1.TLSAdherencePolicyNoOpinion)
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion %d (TLS 1.2), got %d", tls.VersionTLS12, result.MinTLSVersion)
	}
	if result.InitialTLSAdherencePolicy != configv1.TLSAdherencePolicyNoOpinion {
		t.Errorf("expected adherence %q, got %q", configv1.TLSAdherencePolicyNoOpinion, result.InitialTLSAdherencePolicy)
	}
}

// --- BuildServerTLSOptions additional tests ---

func TestBuildServerTLSOptions_NilProfileWithStrictAdherence(t *testing.T) {
	apiServer := &configv1.APIServer{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: configv1.APIServerSpec{
			TLSSecurityProfile: nil,
			TLSAdherence:       configv1.TLSAdherencePolicyStrictAllComponents,
		},
	}
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InitialTLSAdherencePolicy != configv1.TLSAdherencePolicyStrictAllComponents {
		t.Errorf("expected adherence %q, got %q", configv1.TLSAdherencePolicyStrictAllComponents, result.InitialTLSAdherencePolicy)
	}
	if result.MinTLSVersion == 0 {
		t.Error("expected a non-zero MinTLSVersion")
	}
}

func TestBuildServerTLSOptions_EmptyAdherenceUsesDefault(t *testing.T) {
	apiServer := &configv1.APIServer{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: configv1.APIServerSpec{
			TLSSecurityProfile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			},
			TLSAdherence: "",
		},
	}
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := defaultMinTLSVersion()
	if result.MinTLSVersion != expected {
		t.Errorf("expected MinTLSVersion %d (library-go default for non-strict adherence), got %d",
			expected, result.MinTLSVersion)
	}
}

func TestBuildServerTLSOptions_OldProfileWithStrictAdherence(t *testing.T) {
	apiServer := makeAPIServer(configv1.TLSProfileOldType, configv1.TLSAdherencePolicyStrictAllComponents)
	client := newFakeClient(apiServer)

	result, err := BuildServerTLSOptions(context.Background(), client, logr.Discard())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinTLSVersion != tls.VersionTLS10 {
		t.Errorf("expected MinTLSVersion %d (TLS 1.0 from Old profile), got %d", tls.VersionTLS10, result.MinTLSVersion)
	}
}

