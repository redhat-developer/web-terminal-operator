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
	"flag"
	controllerv1alpha1 "github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	"github.com/devfile/devworkspace-operator/controllers/controller/devworkspacerouting"
	"github.com/devfile/devworkspace-operator/pkg/infrastructure"
	"github.com/redhat-developer/web-terminal-operator/pkg/solver"
	"os"
	"os/signal"
	"syscall"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dwv1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha1"
	dwv2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/web-terminal-operator/pkg/webterminal"
)

var (
	scheme   = k8sruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	setupLog.Info("Setting up the thing")
	if err := infrastructure.Initialize(); err != nil {
		setupLog.Error(nil, "unable to detect the Kubernetes infrastructure type", "error", err)
		os.Exit(1)
	}
	setupLog.Info("Finished setting up")
	ctrl.Log.Info("asdasdads")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dwv1.AddToScheme(scheme))
	utilruntime.Must(dwv2.AddToScheme(scheme))
	utilruntime.Must(controllerv1alpha1.AddToScheme(scheme))

	if infrastructure.IsOpenShift() {
		setupLog.Info("asdzzz")
		ctrl.Log.Info("asdasdadszzzzz")
		utilruntime.Must(routev1.AddToScheme(scheme))
	}
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.Log.Info("asdasdadszzzzzzzz")
	setupLog.Info("Setting up the controller")

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	setupLog.Info("asdasdasdasd")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "7d328f43.devfile.io",
	})

	ctrl.Log.Info("asdasdadszzzz")
	setupLog.Info("asdasdasdasdasdasaazzzz")

	if err != nil {
		setupLog.Error(err, "unable to start the operator manager")
		os.Exit(1)
	}

	routingReconciler := &devworkspacerouting.DevWorkspaceRoutingReconciler{
		Client:       mgr.GetClient(),
		Log:          ctrl.Log.WithName("controllers").WithName("DevWorkspaceRouting"),
		Scheme:       mgr.GetScheme(),
		SolverGetter: solver.Getter(scheme),
	}

	if err = routingReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WebTerminalDevWorkspaceRoutingSolver")
		os.Exit(1)
	}

	err = webterminal.SetupDefaultWebTerminalTemplates(context.Background(), mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "Failed to set up default DevWorkspaceTemplates for Web Terminal")
		os.Exit(1)
	}
	setupLog.Info("Web Terminal DevWorkspaceTemplates successfully set up.")
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	sigHandler := ctrl.SetupSignalHandler()

	setupLog.Info("starting manager")
	if err := mgr.Start(sigHandler); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
