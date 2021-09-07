package solver

import (
	"github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	controllerv1alpha1 "github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	"github.com/devfile/devworkspace-operator/controllers/controller/devworkspacerouting/solvers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const webTerminalRoutingClass = "web-terminal"

// WebTerminalGetter negotiates the solver with the calling code
type WebTerminalGetter struct {
	scheme *runtime.Scheme
}

// Getter creates a new WebTerminalGetter
func Getter(scheme *runtime.Scheme) *WebTerminalGetter {
	return &WebTerminalGetter{
		scheme: scheme,
	}
}

func (w WebTerminalGetter) SetupControllerManager(mgr *builder.Builder) error {
	return nil
}

func (w WebTerminalGetter) HasSolver(routingClass v1alpha1.DevWorkspaceRoutingClass) bool {
	return isSupported(routingClass)
}

func (w WebTerminalGetter) GetSolver(client crclient.Client, routingClass v1alpha1.DevWorkspaceRoutingClass) (solver solvers.RoutingSolver, err error) {
	if !isSupported(routingClass) {
		return nil, solvers.RoutingNotSupported
	}
	return &solvers.ClusterSolver{TLS: true}, nil
}

func isSupported(routingClass controllerv1alpha1.DevWorkspaceRoutingClass) bool {
	return routingClass == webTerminalRoutingClass
}
