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

	"github.com/redhat-developer/web-terminal-operator/pkg/config"

	ctrl "sigs.k8s.io/controller-runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log = ctrl.Log.WithName("sync-templates")
)

func SetupDefaultWebTerminalTemplates(ctx context.Context, client crclient.Client) error {
	namespace, err := config.GetNamespace()
	if err != nil {
		return err
	}

	log.Info("Syncing DevWorkspaceTemplate for Web Terminal Tooling")
	if err := syncToolingTemplate(ctx, client, namespace); err != nil {
		return err
	}
	log.Info("Syncing DevWorkspaceTemplate for Web Terminal Exec")
	if err := syncExecTemplate(ctx, client, namespace); err != nil {
		return err
	}

	return nil
}
