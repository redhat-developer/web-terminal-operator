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
	"time"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

func makeUnstructuredAPIServer(profile *configv1.TLSSecurityProfile, adherence configv1.TLSAdherencePolicy) *unstructured.Unstructured {
	apiServer := &configv1.APIServer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "config.openshift.io/v1",
			Kind:       "APIServer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.APIServerSpec{
			TLSSecurityProfile: profile,
			TLSAdherence:       adherence,
		},
	}
	obj, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(apiServer)
	return &unstructured.Unstructured{Object: obj}
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
	if result.InitialTLSProfileSpec.MinTLSVersion != configv1.VersionTLS13 {
		t.Errorf("expected profile MinTLSVersion %q, got %q", configv1.VersionTLS13, result.InitialTLSProfileSpec.MinTLSVersion)
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
	if result.InitialTLSAdherencePolicy != "" {
		t.Errorf("expected empty adherence (profile resolution failed), got %q", result.InitialTLSAdherencePolicy)
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

// --- watchAPIServer tests ---

func newFakeDynClient(fakeWatcher *watch.FakeWatcher) *dynamicfake.FakeDynamicClient {
	scheme := k8sruntime.NewScheme()
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "config.openshift.io", Version: "v1", Kind: "APIServer"},
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "config.openshift.io", Version: "v1", Kind: "APIServerList"},
		&unstructured.UnstructuredList{},
	)
	dynClient := dynamicfake.NewSimpleDynamicClient(scheme)
	dynClient.PrependWatchReactor("apiservers", func(action k8stesting.Action) (bool, watch.Interface, error) {
		return true, fakeWatcher, nil
	})
	return dynClient
}

func waitForCancel(t *testing.T, cancelled <-chan struct{}, expectCancel bool) {
	t.Helper()
	select {
	case <-cancelled:
		if !expectCancel {
			t.Error("onCancel was called but should not have been")
		}
	case <-time.After(2 * time.Second):
		if expectCancel {
			t.Error("timed out waiting for onCancel")
		}
	}
}

func TestWatchAPIServer_RestartsOnAdherencePolicyChange(t *testing.T) {
	fakeWatcher := watch.NewFakeWithChanSize(1, false)
	dynClient := newFakeDynClient(fakeWatcher)
	cancelled := make(chan struct{})

	initial := ServerTLS{
		InitialTLSAdherencePolicy: configv1.TLSAdherencePolicyStrictAllComponents,
		MinTLSVersion:             tls.VersionTLS12,
	}

	go watchAPIServer(dynClient, initial, func() { close(cancelled) }, logr.Discard())

	fakeWatcher.Modify(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType},
		configv1.TLSAdherencePolicyNoOpinion,
	))

	waitForCancel(t, cancelled, true)
}

func TestWatchAPIServer_RestartsOnAddedEventWithChangedAdherence(t *testing.T) {
	fakeWatcher := watch.NewFakeWithChanSize(1, false)
	dynClient := newFakeDynClient(fakeWatcher)
	cancelled := make(chan struct{})

	initial := ServerTLS{
		InitialTLSAdherencePolicy: configv1.TLSAdherencePolicyNoOpinion,
		MinTLSVersion:             tls.VersionTLS12,
	}

	go watchAPIServer(dynClient, initial, func() { close(cancelled) }, logr.Discard())

	fakeWatcher.Add(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType},
		configv1.TLSAdherencePolicyStrictAllComponents,
	))

	waitForCancel(t, cancelled, true)
}

func TestWatchAPIServer_RestartsOnTLSVersionChange(t *testing.T) {
	fakeWatcher := watch.NewFakeWithChanSize(1, false)
	dynClient := newFakeDynClient(fakeWatcher)
	cancelled := make(chan struct{})

	initial := ServerTLS{
		InitialTLSAdherencePolicy: configv1.TLSAdherencePolicyStrictAllComponents,
		MinTLSVersion:             tls.VersionTLS12,
	}

	go watchAPIServer(dynClient, initial, func() { close(cancelled) }, logr.Discard())

	fakeWatcher.Modify(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType},
		configv1.TLSAdherencePolicyStrictAllComponents,
	))

	waitForCancel(t, cancelled, true)
}

func TestWatchAPIServer_IgnoresProfileChangeWithNonStrictAdherence(t *testing.T) {
	fakeWatcher := watch.NewFakeWithChanSize(2, false)
	dynClient := newFakeDynClient(fakeWatcher)
	cancelled := make(chan struct{})

	initial := ServerTLS{
		InitialTLSAdherencePolicy: configv1.TLSAdherencePolicyNoOpinion,
		MinTLSVersion:             tls.VersionTLS12,
	}

	go watchAPIServer(dynClient, initial, func() { close(cancelled) }, logr.Discard())

	// This event should be ignored: profile changed but adherence is not strict
	fakeWatcher.Modify(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType},
		configv1.TLSAdherencePolicyNoOpinion,
	))

	// This event triggers cancel, proving the first event was properly skipped
	fakeWatcher.Modify(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType},
		configv1.TLSAdherencePolicyStrictAllComponents,
	))

	waitForCancel(t, cancelled, true)
}

func TestWatchAPIServer_IgnoresDeleteEvents(t *testing.T) {
	fakeWatcher := watch.NewFakeWithChanSize(2, false)
	dynClient := newFakeDynClient(fakeWatcher)
	cancelled := make(chan struct{})

	initial := ServerTLS{
		InitialTLSAdherencePolicy: configv1.TLSAdherencePolicyStrictAllComponents,
		MinTLSVersion:             tls.VersionTLS12,
	}

	go watchAPIServer(dynClient, initial, func() { close(cancelled) }, logr.Discard())

	// Delete event should be ignored
	fakeWatcher.Delete(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType},
		configv1.TLSAdherencePolicyNoOpinion,
	))

	// This Modified event triggers cancel, proving delete was ignored
	fakeWatcher.Modify(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType},
		configv1.TLSAdherencePolicyNoOpinion,
	))

	waitForCancel(t, cancelled, true)
}

func TestWatchAPIServer_NoRestartWhenTLSVersionUnchanged(t *testing.T) {
	fakeWatcher := watch.NewFakeWithChanSize(1, false)
	dynClient := newFakeDynClient(fakeWatcher)
	cancelled := make(chan struct{})

	initial := ServerTLS{
		InitialTLSAdherencePolicy: configv1.TLSAdherencePolicyStrictAllComponents,
		InitialTLSProfileSpec:     *configv1.TLSProfiles[configv1.TLSProfileIntermediateType],
		MinTLSVersion:             tls.VersionTLS12,
	}

	go watchAPIServer(dynClient, initial, func() { close(cancelled) }, logr.Discard())

	// Same adherence, same min TLS version → no restart
	fakeWatcher.Modify(makeUnstructuredAPIServer(
		&configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType},
		configv1.TLSAdherencePolicyStrictAllComponents,
	))

	waitForCancel(t, cancelled, false)
}
