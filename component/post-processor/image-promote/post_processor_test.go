// Copyright (c) Oxide Computer Company
// SPDX-License-Identifier: MPL-2.0

package imagepromote

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestPostProcessor_Configure(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]interface{}
		envVars   map[string]string
		expectErr bool
	}{
		{
			name: "valid config with explicit values",
			config: map[string]interface{}{
				"host":    "https://api.oxide.example",
				"token":   "test-token",
				"project": "test-project",
			},
			expectErr: false,
		},
		{
			name: "valid config with env vars",
			config: map[string]interface{}{
				"project": "test-project",
			},
			envVars: map[string]string{
				"OXIDE_HOST":  "https://api.oxide.example",
				"OXIDE_TOKEN": "test-token",
			},
			expectErr: false,
		},
		{
			name:      "missing required host",
			config:    map[string]interface{}{},
			expectErr: true,
		},
		{
			name: "missing required token",
			config: map[string]interface{}{
				"host": "https://api.oxide.example",
			},
			expectErr: true,
		},
		{
			name: "project is optional",
			config: map[string]interface{}{
				"host":  "https://api.oxide.example",
				"token": "test-token",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			p := &PostProcessor{}
			err := p.Configure(tt.config)

			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestArtifact(t *testing.T) {
	tests := []struct {
		name      string
		artifact  *Artifact
		wantID    string
		wantStr   string
		wantState map[string]interface{}
	}{
		{
			name: "artifact with project",
			artifact: &Artifact{
				ImageID:   "12345678-1234-1234-1234-123456789012",
				ImageName: "custom-image",
				Project:   "production",
			},
			wantID:  "12345678-1234-1234-1234-123456789012",
			wantStr: "Promoted Oxide image: custom-image (ID: 12345678-1234-1234-1234-123456789012) in project: production",
			wantState: map[string]interface{}{
				"ImageID":   "12345678-1234-1234-1234-123456789012",
				"ImageName": "custom-image",
				"Project":   "production",
			},
		},
		{
			name: "artifact without project (silo level)",
			artifact: &Artifact{
				ImageID:   "87654321-4321-4321-4321-210987654321",
				ImageName: "base-image",
				Project:   "",
			},
			wantID:  "87654321-4321-4321-4321-210987654321",
			wantStr: "Promoted Oxide image: base-image (ID: 87654321-4321-4321-4321-210987654321) at silo level",
			wantState: map[string]interface{}{
				"ImageID":   "87654321-4321-4321-4321-210987654321",
				"ImageName": "base-image",
				"Project":   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.artifact.BuilderId(); got != BuilderID {
				t.Errorf("BuilderId() = %v, want %v", got, BuilderID)
			}

			if got := tt.artifact.Id(); got != tt.wantID {
				t.Errorf("Id() = %v, want %v", got, tt.wantID)
			}

			if got := tt.artifact.String(); got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}

			for k, want := range tt.wantState {
				if got := tt.artifact.State(k); got != want {
					t.Errorf("State(%q) = %v, want %v", k, got, want)
				}
			}

			if got := tt.artifact.State("unknown"); got != nil {
				t.Errorf("State(\"unknown\") = %v, want nil", got)
			}

			if got := tt.artifact.Files(); got != nil {
				t.Errorf("Files() = %v, want nil", got)
			}

			if err := tt.artifact.Destroy(); err != nil {
				t.Errorf("Destroy() error = %v, want nil", err)
			}
		})
	}
}

func TestPostProcessor_InvalidArtifact(t *testing.T) {
	p := &PostProcessor{
		config: Config{
			Host:  "https://api.oxide.example",
			Token: "test-token",
		},
	}

	// Create a mock artifact with wrong builder ID
	mockArtifact := &packer.MockArtifact{
		BuilderIdValue: "wrong.builder",
	}

	_, _, _, err := p.PostProcess(nil, nil, mockArtifact)
	if err == nil {
		t.Error("expected error for wrong artifact type")
	}
}
