//
// Copyright (c) 2021-2024 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//

package config

import (
	"os"
	"testing"
)

func TestGetDefaultToolingImage(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    string
		wantErr bool
	}{
		{
			name:    "returns image when env var is set",
			envVal:  "quay.io/wto/web-terminal-tooling:latest",
			want:    "quay.io/wto/web-terminal-tooling:latest",
			wantErr: false,
		},
		{
			name:    "returns error when env var is unset",
			envVal:  "",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv(toolingImageEnvVar, tt.envVal)
			} else {
				os.Unsetenv(toolingImageEnvVar)
			}
			got, err := GetDefaultToolingImage()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultToolingImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDefaultToolingImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultExecImage(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    string
		wantErr bool
	}{
		{
			name:    "returns image when env var is set",
			envVal:  "quay.io/wto/web-terminal-exec:latest",
			want:    "quay.io/wto/web-terminal-exec:latest",
			wantErr: false,
		},
		{
			name:    "returns error when env var is unset",
			envVal:  "",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv(execImageEnvVar, tt.envVal)
			} else {
				os.Unsetenv(execImageEnvVar)
			}
			got, err := GetDefaultExecImage()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultExecImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDefaultExecImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNamespace(t *testing.T) {
	_, err := GetNamespace()
	if err == nil {
		t.Log("GetNamespace() succeeded — running in a Kubernetes pod")
		return
	}
	t.Log("GetNamespace() returned expected error in non-cluster environment:", err)
}
