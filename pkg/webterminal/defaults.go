package webterminal

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log = ctrl.Log.WithName("sync-templates")
)

func SetupDefaultWebTerminalTemplates(ctx context.Context, client crclient.Client) error {
	log.Info("Syncing DevWorkspaceTemplate for Web Terminal Tooling")
	if err := syncToolingTemplate(ctx, client); err != nil {
		return err
	}
	log.Info("Syncing DevWorkspaceTemplate for Web Terminal Exec")
	if err := syncExecTemplate(ctx, client); err != nil {
		return err
	}

	return nil
}
