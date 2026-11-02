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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dwv1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha1"
	dwv2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	configv1 "github.com/openshift/api/config/v1"

	"github.com/redhat-developer/web-terminal-operator/pkg/tlssetup"
	"github.com/redhat-developer/web-terminal-operator/pkg/webterminal"
)

var (
	scheme   = k8sruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dwv1.AddToScheme(scheme))
	utilruntime.Must(dwv2.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
}

func main() {
	ctrl.SetLogger(zap.New())

	cfg, err := ctrlconfig.GetConfig()
	if err != nil {
		setupLog.Error(err, "Failed to read cluster configuration")
		os.Exit(1)
	}
	client, err := crclient.New(cfg, crclient.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Failed to create Kubernetes client")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverTLS, err := tlssetup.BuildServerTLSOptions(ctx, client, setupLog)
	if err != nil {
		setupLog.Error(err, "Failed to read TLS profile from APIServer")
		os.Exit(1)
	}

	err = webterminal.SetupDefaultWebTerminalTemplates(ctx, client, serverTLS)
	if err != nil {
		setupLog.Error(err, "Failed to set up default DevWorkspaceTemplates for Web Terminal")
		os.Exit(1)
	}
	setupLog.Info("Web Terminal DevWorkspaceTemplates successfully set up.")

	if err := tlssetup.RegisterSecurityProfileWatcher(cfg, serverTLS, cancel, setupLog); err != nil {
		setupLog.Error(err, "Unable to set up TLS security profile watcher")
	}

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		setupLog.Info("TLS security profile changed, operator will restart")
	case <-exitSignal:
	}
}
