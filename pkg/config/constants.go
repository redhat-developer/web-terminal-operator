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

package config

const (
	PermittedNamespacesAnnotation = "controller.devfile.io/allow-import-from"

	ToolingTemplateName       = "web-terminal-tooling"
	ExecTemplateName          = "web-terminal-exec"
	DefaultTemplatesNamespace = "openshift-operators"

	ToolingMemoryRequest = "128Mi"
	ToolingMemoryLimit   = "256Mi"
	ToolingCPURequest    = "100m"
	ToolingCPULimit      = "400m"

	ExecMemoryRequest = "128Mi"
	ExecMemoryLimit   = "128Mi"
	ExecCPURequest    = "100m"
	ExecCPULimit      = "400m"
)

var (
	DefaultTemplatesLabels = map[string]string{
		"console.openshift.io/terminal": "true",
	}
)
