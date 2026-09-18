//
// Copyright (c) 2021-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//

package webterminal

import (
	"testing"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/web-terminal-operator/pkg/config"
)

func makeExecTemplate(image string) *dw.DevWorkspaceTemplate {
	return &dw.DevWorkspaceTemplate{
		ObjectMeta: v1.ObjectMeta{
			Name:      config.ExecTemplateName,
			Namespace: "test-ns",
		},
		Spec: dw.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
				Components: []dw.Component{
					{
						Name: config.ExecTemplateName,
						ComponentUnion: dw.ComponentUnion{
							Container: &dw.ContainerComponent{
								Container: dw.Container{
									Image: image,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestHandleUnmanagedExecState(t *testing.T) {
	const defaultImage = "quay.io/wto/web-terminal-exec:latest"

	tests := []struct {
		name          string
		specImage     string
		clusterImage  string
		expectedImage string
	}{
		{
			name:          "same image returns cluster unchanged",
			specImage:     defaultImage,
			clusterImage:  defaultImage,
			expectedImage: defaultImage,
		},
		{
			name:          "same repo different tag updates image",
			specImage:     defaultImage,
			clusterImage:  "quay.io/wto/web-terminal-exec:old",
			expectedImage: defaultImage,
		},
		{
			name:          "same repo different digest updates image",
			specImage:     defaultImage,
			clusterImage:  "quay.io/wto/web-terminal-exec@sha256:abc123",
			expectedImage: defaultImage,
		},
		{
			name:          "different repo does not update",
			specImage:     defaultImage,
			clusterImage:  "quay.io/other/custom-exec:latest",
			expectedImage: "quay.io/other/custom-exec:latest",
		},
		{
			name:          "empty cluster image returns cluster unchanged",
			specImage:     defaultImage,
			clusterImage:  "",
			expectedImage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RELATED_IMAGE_web_terminal_exec", tt.specImage)

			spec := makeExecTemplate(tt.specImage)
			cluster := makeExecTemplate(tt.clusterImage)

			result := handleUnmanagedExecState(spec, cluster)

			var resultImage string
			for _, component := range result.Spec.Components {
				if component.Name == config.ExecTemplateName && component.Container != nil {
					resultImage = component.Container.Image
				}
			}
			if resultImage != tt.expectedImage {
				t.Errorf("expected image %q, got %q", tt.expectedImage, resultImage)
			}
		})
	}
}

func TestHandleUnmanagedExecStateEnvUnset(t *testing.T) {
	t.Setenv("RELATED_IMAGE_web_terminal_exec", "")

	spec := makeExecTemplate("any-image")
	cluster := makeExecTemplate("quay.io/wto/web-terminal-exec:old")

	result := handleUnmanagedExecState(spec, cluster)

	var resultImage string
	for _, component := range result.Spec.Components {
		if component.Name == config.ExecTemplateName && component.Container != nil {
			resultImage = component.Container.Image
		}
	}
	if resultImage != "quay.io/wto/web-terminal-exec:old" {
		t.Errorf("expected cluster image unchanged when env var unset, got %q", resultImage)
	}
}

func TestGetSpecExecTemplate(t *testing.T) {
	const testImage = "quay.io/wto/web-terminal-exec:test"
	t.Setenv("RELATED_IMAGE_web_terminal_exec", testImage)

	dwt, err := getSpecExecTemplate("test-namespace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dwt.Name != config.ExecTemplateName {
		t.Errorf("expected name %q, got %q", config.ExecTemplateName, dwt.Name)
	}
	if dwt.Namespace != "test-namespace" {
		t.Errorf("expected namespace %q, got %q", "test-namespace", dwt.Namespace)
	}
	if len(dwt.Spec.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(dwt.Spec.Components))
	}
	if dwt.Spec.Components[0].Container.Image != testImage {
		t.Errorf("expected image %q, got %q", testImage, dwt.Spec.Components[0].Container.Image)
	}
}

func TestGetSpecExecTemplateEnvUnset(t *testing.T) {
	t.Setenv("RELATED_IMAGE_web_terminal_exec", "")

	_, err := getSpecExecTemplate("test-namespace")
	if err == nil {
		t.Error("expected error when env var is unset")
	}
}
