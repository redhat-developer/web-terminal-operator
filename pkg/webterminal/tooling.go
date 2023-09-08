//
// Copyright (c) 2021-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package webterminal

import (
	"context"
	"fmt"
	"strings"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-developer/web-terminal-operator/pkg/config"
)

func syncToolingTemplate(ctx context.Context, client crclient.Client, namespace string) error {
	clusterDWT, err := getClusterToolingTemplate(ctx, client, namespace)
	if err != nil {
		return err
	}
	specDWT, err := getSpecToolingTemplate(namespace)
	if err != nil {
		return nil
	}

	if clusterDWT == nil {
		log.Info("DevWorkspaceTemplate for Web Terminal Tooling does not exist; creating.")
		return client.Create(ctx, specDWT)
	}

	var updatedDWT *dw.DevWorkspaceTemplate
	if clusterDWT.Annotations != nil && clusterDWT.Annotations[config.UnmanagedStateAnnotation] == "true" {
		log.Info("Found unmanaged template for Web Terminal Tooling.")
		updatedDWT = handleUnmanagedToolingState(specDWT, clusterDWT)
	} else {
		specDWT.ResourceVersion = clusterDWT.ResourceVersion
		updatedDWT = specDWT
	}

	err = client.Update(ctx, updatedDWT)
	if err != nil {
		return fmt.Errorf("error updating Web Terminal Tooling template: %w", err)
	}
	log.Info("Web Terminal Tooling template updated.")
	return nil
}

func handleUnmanagedToolingState(spec, cluster *dw.DevWorkspaceTemplate) *dw.DevWorkspaceTemplate {
	result := cluster.DeepCopy()

	// Even though the template is marked as unmanaged, still try to update the image used if the current
	// image is a default image to ensure CVE fixes are propagated
	specImage, err := config.GetDefaultToolingImage()
	if err != nil {
		// Should never happen, since getSpecToolingTemplate depends on the same function
		log.Info("warning: Could not get default Web Terminal Tooling container image")
		return cluster
	}
	var clusterImage string
	for _, component := range cluster.Spec.Components {
		if component.Name == config.ToolingTemplateName && component.Container != nil {
			clusterImage = component.Container.Image
		}
	}
	if specImage == clusterImage {
		// No update needed
		return cluster
	}

	if clusterImage == "" {
		// The cluster template is potentially _very_ different from default; do the safe thing and make no changes
		return cluster
	}
	// Strip digest or tag from images
	specRepo := strings.Split(specImage, `@`)[0]
	specRepo = strings.Split(specRepo, `:`)[0]
	clusterRepo := strings.Split(clusterImage, `@`)[0]
	clusterRepo = strings.Split(clusterRepo, `:`)[0]
	if specRepo == clusterRepo {
		log.Info(fmt.Sprintf("Found default image %s in Web Terminal Tooling template. Updating image to %s", clusterImage, specImage))
		for idx, component := range result.Spec.Components {
			if component.Name == config.ToolingTemplateName {
				result.Spec.Components[idx].Container.Image = specImage
			}
		}
	}

	return result
}

func getSpecToolingTemplate(namespace string) (*dw.DevWorkspaceTemplate, error) {
	image, err := config.GetDefaultToolingImage()
	if err != nil {
		return nil, err
	}
	boolFalse := false

	dwt := &dw.DevWorkspaceTemplate{
		ObjectMeta: v1.ObjectMeta{
			Name:      config.ToolingTemplateName,
			Namespace: namespace,
			Annotations: map[string]string{
				config.PermittedNamespacesAnnotation: "*",
			},
			Labels: config.DefaultTemplatesLabels,
		},
		Spec: dw.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
				Components: []dw.Component{
					{
						Name: config.ToolingTemplateName,
						ComponentUnion: dw.ComponentUnion{
							Container: &dw.ContainerComponent{
								Container: dw.Container{
									Image:         image,
									Env:           nil,
									MemoryLimit:   config.ToolingMemoryLimit,
									MemoryRequest: config.ToolingMemoryRequest,
									CpuLimit:      config.ToolingCPULimit,
									CpuRequest:    config.ToolingCPURequest,
									Args:          []string{"tail", "-f", "/dev/null"},
									MountSources:  &boolFalse,
								},
							},
						},
					},
				},
			},
		},
	}
	return dwt, nil
}

func getClusterToolingTemplate(ctx context.Context, client crclient.Client, namespace string) (*dw.DevWorkspaceTemplate, error) {
	toolingTemplate := &dw.DevWorkspaceTemplate{}
	toolingRef := types.NamespacedName{
		Name:      config.ToolingTemplateName,
		Namespace: namespace,
	}
	err := client.Get(ctx, toolingRef, toolingTemplate)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return toolingTemplate, nil
}
