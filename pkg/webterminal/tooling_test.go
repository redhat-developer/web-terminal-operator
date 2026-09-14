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
	"os"
	"testing"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/web-terminal-operator/pkg/config"
)

func makeToolingTemplate(image string) *dw.DevWorkspaceTemplate {
	return &dw.DevWorkspaceTemplate{
		ObjectMeta: v1.ObjectMeta{
			Name:      config.ToolingTemplateName,
			Namespace: "test-ns",
		},
		Spec: dw.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
				Components: []dw.Component{
					{
						Name: config.ToolingTemplateName,
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

func TestHandleUnmanagedToolingState(t *testing.T) {
	const defaultImage = "quay.io/wto/web-terminal-tooling:latest"

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
			clusterImage:  "quay.io/wto/web-terminal-tooling:old",
			expectedImage: defaultImage,
		},
		{
			name:          "same repo different digest updates image",
			specImage:     defaultImage,
			clusterImage:  "quay.io/wto/web-terminal-tooling@sha256:abc123",
			expectedImage: defaultImage,
		},
		{
			name:          "different repo does not update",
			specImage:     defaultImage,
			clusterImage:  "quay.io/other/custom-tooling:latest",
			expectedImage: "quay.io/other/custom-tooling:latest",
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
			t.Setenv("RELATED_IMAGE_web_terminal_tooling", tt.specImage)

			spec := makeToolingTemplate(tt.specImage)
			cluster := makeToolingTemplate(tt.clusterImage)

			result := handleUnmanagedToolingState(spec, cluster)

			var resultImage string
			for _, component := range result.Spec.Components {
				if component.Name == config.ToolingTemplateName && component.Container != nil {
					resultImage = component.Container.Image
				}
			}
			if resultImage != tt.expectedImage {
				t.Errorf("expected image %q, got %q", tt.expectedImage, resultImage)
			}
		})
	}
}

func TestHandleUnmanagedToolingStateEnvUnset(t *testing.T) {
	os.Unsetenv("RELATED_IMAGE_web_terminal_tooling")

	spec := makeToolingTemplate("any-image")
	cluster := makeToolingTemplate("quay.io/wto/web-terminal-tooling:old")

	result := handleUnmanagedToolingState(spec, cluster)

	var resultImage string
	for _, component := range result.Spec.Components {
		if component.Name == config.ToolingTemplateName && component.Container != nil {
			resultImage = component.Container.Image
		}
	}
	if resultImage != "quay.io/wto/web-terminal-tooling:old" {
		t.Errorf("expected cluster image unchanged when env var unset, got %q", resultImage)
	}
}

func TestGetSpecToolingTemplate(t *testing.T) {
	const testImage = "quay.io/wto/web-terminal-tooling:test"
	t.Setenv("RELATED_IMAGE_web_terminal_tooling", testImage)

	dwt, err := getSpecToolingTemplate("test-namespace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dwt.Name != config.ToolingTemplateName {
		t.Errorf("expected name %q, got %q", config.ToolingTemplateName, dwt.Name)
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

func TestGetSpecToolingTemplateEnvUnset(t *testing.T) {
	os.Unsetenv("RELATED_IMAGE_web_terminal_tooling")

	_, err := getSpecToolingTemplate("test-namespace")
	if err == nil {
		t.Error("expected error when env var is unset")
	}
}
