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

	if clusterDWT.Annotations != nil && clusterDWT.Annotations[config.UnmanagedStateAnnotation] == "true" {
		log.Info("Found unmanaged template for Web Terminal Tooling; skipping.")
		return nil
	}

	specDWT.ResourceVersion = clusterDWT.ResourceVersion
	err = client.Update(ctx, specDWT)
	if err != nil {
		return fmt.Errorf("error updating Web Terminal Tooling template: %w", err)
	}
	log.Info("Web Terminal Tooling template updated.")
	return nil
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
