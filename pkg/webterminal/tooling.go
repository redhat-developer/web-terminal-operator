//
// Copyright (c) 2021 Red Hat, Inc.
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

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-developer/web-terminal-operator/pkg/config"
)

func syncToolingTemplate(ctx context.Context, client crclient.Client) error {
	clusterDWT, err := getClusterToolingTemplate(ctx, client)
	if err != nil {
		return err
	}
	specDWT, err := getSpecToolingTemplate()
	if err != nil {
		return nil
	}

	if clusterDWT == nil {
		log.Info("DevWorkspaceTemplate for Web Terminal Tooling does not exist; creating.")
		return client.Create(ctx, specDWT)
	}

	// TODO: Figure out way to sync DWTs for updates to WTO.
	log.Info("Web Terminal Tooling template already exists; skipping.")
	return nil
}

func getSpecToolingTemplate() (*dw.DevWorkspaceTemplate, error) {
	image, err := config.GetDefaultToolingImage()
	if err != nil {
		return nil, err
	}
	boolFalse := false

	dwt := &dw.DevWorkspaceTemplate{
		ObjectMeta: v1.ObjectMeta{
			Name:      config.ToolingTemplateName,
			Namespace: config.DefaultTemplatesNamespace,
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

func getClusterToolingTemplate(ctx context.Context, client crclient.Client) (*dw.DevWorkspaceTemplate, error) {
	toolingTemplate := &dw.DevWorkspaceTemplate{}
	toolingRef := types.NamespacedName{
		Name:      config.ToolingTemplateName,
		Namespace: config.DefaultTemplatesNamespace,
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
