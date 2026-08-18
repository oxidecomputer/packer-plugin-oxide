// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"regexp"
	"strings"
	"testing"
)

func TestUniqueSuffix(t *testing.T) {
	t.Setenv("PACKER_RUN_UUID", "packer-run-uuid")

	config := &Config{}
	suffix := config.uniqueSuffix()

	if suffix != config.uniqueSuffix() {
		t.Fatal("uniqueSuffix returned different values for the same config")
	}

	uuidPattern := regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	)
	if !uuidPattern.MatchString(suffix) {
		t.Fatalf("uniqueSuffix returned invalid UUID %q", suffix)
	}
}

func TestUniqueName(t *testing.T) {
	config := &Config{}
	prefix := "packer"
	name := config.uniqueName(prefix)

	if got, want := name, prefix+"-"+config.uniqueSuffix(); got != want {
		t.Fatalf("uniqueName returned %q; want %q", got, want)
	}
}

func TestBuildDescription(t *testing.T) {
	tests := []struct {
		name      string
		buildName string
		want      string
	}{
		{
			name:      "build name",
			buildName: "service-image",
			want:      `Created by Packer build "service-image".`,
		},
		{
			name: "missing build name",
			want: "Created by Packer.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &Config{}
			config.PackerBuildName = test.buildName

			if got := config.buildDescription(); got != test.want {
				t.Fatalf("buildDescription returned %q; want %q", got, test.want)
			}
		})
	}
}

func TestPrepareGeneratedNames(t *testing.T) {
	config := &Config{}
	_, err := config.Prepare(map[string]any{
		"artifact_name":      "artifact",
		"boot_disk_image_id": "image-id",
		"communicator":       "none",
		"packer_build_name":  "service-image",
		"project":            "project",
	})
	if err != nil {
		t.Fatalf("Prepare returned an error: %v", err)
	}

	if config.Hostname != config.Name {
		t.Fatalf("Hostname is %q; want %q", config.Hostname, config.Name)
	}
	if got, want := config.ArtifactDescription,
		`Created by Packer build "service-image".`; got != want {
		t.Fatalf("ArtifactDescription is %q; want %q", got, want)
	}
	if config.Comm.SSHTemporaryKeyPairName != config.Name {
		t.Fatalf(
			"SSHTemporaryKeyPairName is %q; want %q",
			config.Comm.SSHTemporaryKeyPairName,
			config.Name,
		)
	}
}

func TestPrepareRequiresArtifactName(t *testing.T) {
	config := &Config{}
	_, err := config.Prepare(map[string]any{
		"boot_disk_image_id": "image-id",
		"communicator":       "none",
		"project":            "project",
	})
	if err == nil || !strings.Contains(err.Error(), "artifact_name is required") {
		t.Fatalf("Prepare returned error %v", err)
	}
}

func TestPrepareAllowsMissingArtifactNameWhenSkippingImage(t *testing.T) {
	config := &Config{}
	_, err := config.Prepare(map[string]any{
		"boot_disk_image_id": "image-id",
		"communicator":       "none",
		"project":            "project",
		"skip_create_image":  true,
	})
	if err != nil {
		t.Fatalf("Prepare returned an error: %v", err)
	}
}
