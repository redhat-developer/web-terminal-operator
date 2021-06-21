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
	"github.com/devfile/api/v2/pkg/attributes"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-developer/web-terminal-operator/pkg/config"
)

func syncExecTemplate(ctx context.Context, client crclient.Client) error {
	clusterDWT, err := getClusterExecTemplate(ctx, client)
	if err != nil {
		return err
	}
	specDWT, err := getSpecExecTemplate()
	if err != nil {
		return nil
	}

	if clusterDWT == nil {
		log.Info("DevWorkspaceTemplate for Web Terminal Exec does not exist; creating.")
		return client.Create(ctx, specDWT)
	}
	// TODO: Figure out way to sync DWTs for updates to WTO.
	log.Info("Web Terminal Exec template already exists; skipping.")
	return nil
}

func getSpecExecTemplate() (*dw.DevWorkspaceTemplate, error) {
	image, err := config.GetDefaultExecImage()
	if err != nil {
		return nil, err
	}
	boolFalse := false
	mainAttr := attributes.Attributes{}
	mainAttr.PutString("type", "main")

	dwt := &dw.DevWorkspaceTemplate{
		ObjectMeta: v1.ObjectMeta{
			Name:      config.ExecTemplateName,
			Namespace: config.DefaultTemplatesNamespace,
			Annotations: map[string]string{
				config.PermittedNamespacesAnnotation: "*",
			},
		},
		Spec: dw.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
				Components: []dw.Component{
					{
						Name: config.ExecTemplateName,
						ComponentUnion: dw.ComponentUnion{
							Container: &dw.ContainerComponent{
								Container: dw.Container{
									Image:         image,
									Env:           nil,
									MemoryLimit:   config.ExecMemoryLimit,
									MemoryRequest: config.ExecMemoryRequest,
									CpuLimit:      config.ExecCPULimit,
									CpuRequest:    config.ExecCPURequest,
									Command: []string{
										"/go/bin/che-machine-exec",
										"--authenticated-user-id", "$(DEVWORKSPACE_CREATOR)",
										"--idle-timeout", "$(DEVWORKSPACE_IDLE_TIMEOUT)",
										"--pod-selector", "controller.devfile.io/devworkspace_id=$(DEVWORKSPACE_ID)",
										"--use-tls",
										"--use-bearer-token",
									},
									Args:         nil,
									MountSources: &boolFalse,
								},
								Endpoints: []dw.Endpoint{
									{
										Name:       "exec",
										TargetPort: 4444,
										Attributes: mainAttr,
									},
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

func getClusterExecTemplate(ctx context.Context, client crclient.Client) (*dw.DevWorkspaceTemplate, error) {
	execTemplate := &dw.DevWorkspaceTemplate{}
	execRef := types.NamespacedName{
		Name:      config.ExecTemplateName,
		Namespace: config.DefaultTemplatesNamespace,
	}
	err := client.Get(ctx, execRef, execTemplate)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return execTemplate, nil
}
