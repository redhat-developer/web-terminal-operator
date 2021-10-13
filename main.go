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

	"github.com/redhat-developer/web-terminal-operator/pkg/versioning"
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
}

func main() {
	ctrl.SetLogger(zap.New())

	err := versioning.InitVersioning()
	if err != nil {
		setupLog.Error(err, "Failed to detect cluster version")
		os.Exit(1)
	}

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
	err = webterminal.SetupDefaultWebTerminalTemplates(context.Background(), client)
	if err != nil {
		setupLog.Error(err, "Failed to set up default DevWorkspaceTemplates for Web Terminal")
		os.Exit(1)
	}
	setupLog.Info("Web Terminal DevWorkspaceTemplates successfully set up.")
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
