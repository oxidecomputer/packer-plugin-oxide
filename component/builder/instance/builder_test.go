// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance_test

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/oxidecomputer/oxide.go/oxide"
)

//go:embed testdata/*.pkr.hcl.tmpl
var packerTemplates embed.FS

// TestAccBuilder_Config tests that the builder fails when required
// configuration arguments are not provided.
func TestAccBuilder_Config(t *testing.T) {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", acctest.TestEnvVar)
		return
	}

	tmpl := template.Must(template.ParseFS(packerTemplates, "testdata/*.pkr.hcl.tmpl"))

	executeTemplate := func(t *testing.T, name string, data any) string {
		var s strings.Builder
		if err := tmpl.ExecuteTemplate(&s, name, data); err != nil {
			t.Fatalf("failed executing template %s: %v", name, err)
		}
		return s.String()
	}

	tt := []*acctest.PluginTestCase{
		{
			Name: "MissingAllRequiredFields",
			Type: "oxide-instance",
			Template: executeTemplate(
				t,
				"config.pkr.hcl.tmpl",
				struct {
					Project         string
					BootDiskImageID string
				}{},
			),
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				assertFileContains(t, logfile, "project is required")
				assertFileContains(t, logfile, "boot_disk_image_id is required")

				return nil
			},
		},
		{
			Name: "MissingProject",
			Type: "oxide-instance",
			Template: executeTemplate(
				t,
				"config.pkr.hcl.tmpl",
				struct {
					Project         string
					BootDiskImageID string
				}{
					BootDiskImageID: "test-boot-disk-image-id",
				},
			),
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				assertFileContains(t, logfile, "project is required")

				return nil
			},
		},
		{
			Name: "MissingBootDiskImageID",
			Type: "oxide-instance",
			Template: executeTemplate(
				t,
				"config.pkr.hcl.tmpl",
				struct {
					Project         string
					BootDiskImageID string
				}{
					Project: "test-project",
				},
			),
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				assertFileContains(t, logfile, "boot_disk_image_id is required")

				return nil
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			acctest.TestPlugin(t, tc)
		})
	}
}

// TestAccBuilder_Instance tests that the builder can successfully boot an
// instance and create an image from it. It makes use of the Setup and Teardown
// fields within [acctest.PluginTestCase] to create and destroy the required
// resources in Oxide before running the tests.
//
// There's no requirement to use the Setup and Teardown fields within
// [acctest.PluginTestCase]. In fact, many Packer plugins don't use those fields
// and instead use custom setup logic within the test and [testing#T.Cleanup]
// for teardown logic.
func TestAccBuilder_Instance(t *testing.T) {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", acctest.TestEnvVar)
		return
	}

	requiredEnvVars := []string{"OXIDE_PROJECT", "OXIDE_BOOT_DISK_IMAGE_NAME"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s environment variable is required", envVar)
		}
	}

	tmpl := template.Must(template.ParseFS(packerTemplates, "testdata/*.pkr.hcl.tmpl"))

	var oxideClientOpts []oxide.ClientOption
	insecureSkipVerify := os.Getenv("OXIDE_INSECURE_SKIP_VERIFY") == "true"
	if insecureSkipVerify {
		oxideClientOpts = append(oxideClientOpts, oxide.WithInsecureSkipVerify())
	}

	oxideClient, err := oxide.NewClient(oxideClientOpts...)
	if err != nil {
		t.Fatalf("failed creating oxide client: %v", err)
	}

	oxideProject := os.Getenv("OXIDE_PROJECT")
	oxideBootDiskImageName := os.Getenv("OXIDE_BOOT_DISK_IMAGE_NAME")

	bootDiskImage, err := oxideClient.ImageView(t.Context(), oxide.ImageViewParams{
		Image: oxide.NameOrId(oxideBootDiskImageName),
	})
	if err != nil {
		t.Fatalf("failed to resolve boot disk image %q: %v", oxideBootDiskImageName, err)
	}
	oxideBootDiskImageID := bootDiskImage.Id

	tt := []struct {
		testName        string
		project         string
		bootDiskImageID string
		cpus            uint64
		memory          uint64
		bootDiskSize    uint64
	}{
		{
			testName:        "DefaultConfiguration",
			project:         oxideProject,
			bootDiskImageID: oxideBootDiskImageID,
		},
		{
			testName:        "CustomConfiguration",
			project:         oxideProject,
			bootDiskImageID: oxideBootDiskImageID,
			cpus:            2,
			memory:          4 * 1024 * 1024 * 1024,  // 4 GiB
			bootDiskSize:    30 * 1024 * 1024 * 1024, // 30 GiB
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			// This test ID must be unique between tests to ensure there's no naming
			// conflict when creating resources in Oxide. Having a unique, but known, test ID
			// allows the test to create and destroy resources within the Setup and Teardown
			// fields without creating some other mechanism to share state.
			testID := fmt.Sprintf("packer-%d", time.Now().UnixNano())
			artifactName := fmt.Sprintf("%s-%s", testID, "artifact")

			var packerTemplate strings.Builder
			if err := tmpl.ExecuteTemplate(&packerTemplate, "instance.pkr.hcl.tmpl", struct {
				Project            string
				BootDiskImageID    string
				CPUs               uint64
				Memory             uint64
				BootDiskSize       uint64
				ArtifactName       string
				InsecureSkipVerify bool
			}{
				Project:            tc.project,
				BootDiskImageID:    tc.bootDiskImageID,
				CPUs:               tc.cpus,
				Memory:             tc.memory,
				BootDiskSize:       tc.bootDiskSize,
				ArtifactName:       artifactName,
				InsecureSkipVerify: insecureSkipVerify,
			},
			); err != nil {
				t.Fatalf("failed rendering packer template: %v", err)
			}

			acctest.TestPlugin(t, &acctest.PluginTestCase{
				Name:     tc.testName,
				Type:     "oxide-instance",
				Template: packerTemplate.String(),
				Setup: func() error {
					return nil
				},
				Teardown: func() error {
					var joinedError error

					t.Logf("teardown: deleting oxide image %s", artifactName)
					if err := oxideClient.ImageDelete(t.Context(), oxide.ImageDeleteParams{
						Image:   oxide.NameOrId(artifactName),
						Project: oxide.NameOrId(oxideProject),
					}); err != nil {
						joinedError = errors.Join(joinedError, fmt.Errorf("failed deleting oxide image %s: %v", artifactName, err))
					}

					if joinedError != nil {
						return fmt.Errorf("failed tearing down resources: %v", joinedError)
					}

					return nil
				},
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState != nil {
						if buildCommand.ProcessState.ExitCode() != 0 {
							return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
						}
					}

					assertFileContains(t, logfile, "Build 'oxide-instance.test' finished after")

					return nil
				},
			})
		})
	}
}

func assertFileContains(t *testing.T, filename string, expected string) {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if matched, _ := regexp.MatchString(expected+".*", string(b)); !matched {
		t.Fatalf("logs doesn't contain expected value %q:\n%s", expected, string(b))
	}
}
