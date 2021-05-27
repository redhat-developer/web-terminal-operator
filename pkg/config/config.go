package config

import (
	"fmt"
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
