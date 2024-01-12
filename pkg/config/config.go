//
// Copyright (c) 2021-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package config

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	toolingImageEnvVar = "RELATED_IMAGE_web_terminal_tooling"
	execImageEnvVar    = "RELATED_IMAGE_web_terminal_exec"
)

func GetDefaultToolingImage() (string, error) {
	val := os.Getenv(toolingImageEnvVar)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s is unset", toolingImageEnvVar)
	}
	return val, nil
}

func GetDefaultExecImage() (string, error) {
	val := os.Getenv(execImageEnvVar)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s is unset", execImageEnvVar)
	}
	return val, nil
}

func GetNamespace() (string, error) {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return string(namespace), err
}
