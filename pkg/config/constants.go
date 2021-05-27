package config

const (
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
