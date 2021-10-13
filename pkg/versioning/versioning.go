package versioning

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("versioning")

	kubeMajorVersion string
	kubeMinorVersion string
	openShiftVersion string

	// Correspondence between Kubernetes server versions and OpenShift versions
	// Pulled from https://access.redhat.com/solutions/4870701
	k8sToOpenShiftVersions = map[string]string{
		"1.13": "4.1",
		"1.14": "4.2",
		"1.16": "4.3",
		"1.17": "4.4",
		"1.18": "4.5",
		"1.19": "4.6",
		"1.20": "4.7",
		"1.21": "4.8",
	}
)

func InitVersioning() error {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	disclient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}
	ver, err := disclient.ServerVersion()
	if err != nil {
		return err
	}

	kubeMajorVersion = ver.Major
	kubeMinorVersion = ver.Minor
	openShiftVersion = k8sToOpenShiftVersions[fmt.Sprintf("%s.%s", ver.Major, ver.Minor)]
	if openShiftVersion == "" {
		openShiftVersion = "version unknown"
	}

	log.Info(fmt.Sprintf("Detected Kubernetes version %s.%s, corresponding to OpenShift %s", ver.Major, ver.Minor, openShiftVersion))

	return nil
}
