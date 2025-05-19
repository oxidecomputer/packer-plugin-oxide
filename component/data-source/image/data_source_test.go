// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package image_test

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

// TestAccDataSource_Config tests that the oxide-image data source fails when
// required configuration attributes are not provided.
func TestAccDataSource_Config(t *testing.T) {
	requiredEnvVars := []string{"OXIDE_HOST", "OXIDE_TOKEN"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s environment variable is required", envVar)
		}
	}

	tmpl := template.Must(template.ParseFS(packerTemplates, "testdata/*.pkr.hcl.tmpl"))

	// executeTemplate is a helper function that renders a Go template and returns
	// the rendered template to the caller, failing the test when the template
	// cannot be rendered. This helps assign a Packer template to the Template
	// field in test cases. Tests should probably use [fmt.Sprintf] to build
	// Packer templates instead of Go templates but it's an interesting experiment
	// to try and use Go templates.
	executeTemplate := func(t *testing.T, name string, data any) string {
		var s strings.Builder
		if err := tmpl.ExecuteTemplate(&s, name, data); err != nil {
			t.Fatalf("failed executing template %s: %v", name, err)
		}
		return s.String()
	}

	// Hold on to these so we can reset their values in between tests.
	oxideHost := os.Getenv("OXIDE_HOST")
	oxideToken := os.Getenv("OXIDE_TOKEN")

	tt := []*acctest.PluginTestCase{
		{
			Name: "MissingAllRequiredFields",
			Type: "oxide-image",
			Template: executeTemplate(
				t,
				"config.pkr.hcl.tmpl",
				struct{ Name string }{Name: ""},
			),
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				assertFileContains(t, logfile, "host is required")
				assertFileContains(t, logfile, "token is required")
				assertFileContains(t, logfile, "name is required")

				return nil
			},
			Setup: func() error {
				os.Unsetenv("OXIDE_HOST")
				os.Unsetenv("OXIDE_TOKEN")
				return nil
			},
			Teardown: func() error {
				os.Setenv("OXIDE_HOST", oxideHost)
				os.Setenv("OXIDE_TOKEN", oxideToken)
				return nil
			},
		},
		{
			Name: "MissingAPICredentials",
			Type: "oxide-image",
			Template: executeTemplate(
				t,
				"config.pkr.hcl.tmpl",
				struct{ Name string }{Name: "test-image"},
			),
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				assertFileContains(t, logfile, "host is required")
				assertFileContains(t, logfile, "token is required")

				return nil
			},
			Setup: func() error {
				os.Unsetenv("OXIDE_HOST")
				os.Unsetenv("OXIDE_TOKEN")
				return nil
			},
			Teardown: func() error {
				os.Setenv("OXIDE_HOST", oxideHost)
				os.Setenv("OXIDE_TOKEN", oxideToken)
				return nil
			},
		},
		{
			Name: "MissingName",
			Type: "oxide-image",
			Template: executeTemplate(
				t,
				"config.pkr.hcl.tmpl",
				struct{ Name string }{Name: ""},
			),
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				assertFileContains(t, logfile, "name is required")

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

// TestAccDataSource_Image tests that the oxide-image data source can
// successfully read a project image and a silo image from the Oxide API. It
// makes use of the Setup and Teardown fields within [acctest.PluginTestCase] to
// create the required image in Oxide before running Packer to retreive it.
//
// There's no requirement to use the Setup and Teardown fields within
// [acctest.PluginTestCase]. In fact, many Packer plugins don't use those fields
// and instead use custom setup logic within the test and [testing#T.Cleanup]
// for teardown logic. However, the implementation of this test was a good
// change to experiment with the APIs provided by [acctest.PluginTestCase] to
// see how they work.
func TestAccDataSource_Image(t *testing.T) {
	requiredEnvVars := []string{"OXIDE_HOST", "OXIDE_TOKEN", "OXIDE_PROJECT"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s environment variable is required", envVar)
		}
	}

	tmpl := template.Must(template.ParseFS(packerTemplates, "testdata/*.pkr.hcl.tmpl"))

	oxideClient, err := oxide.NewClient(nil)
	if err != nil {
		t.Fatalf("failed creating oxide client: %v", err)
	}

	oxideProject := os.Getenv("OXIDE_PROJECT")

	tt := []struct {
		testName  string
		project   string
		siloImage bool
	}{
		{
			testName: "ProjectImage",
			project:  oxideProject,
		},
		{
			testName:  "SiloImage",
			project:   oxideProject,
			siloImage: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			// This test ID must be unique between tests to ensure there's no naming
			// conflict when creating resources in Oxide. Having a unique, but known,
			// test ID allows the test to create and destroy disks, snapshots, and
			// images within the Setup and Teardown fields without creating some other
			// mechanism to share state.
			testID := fmt.Sprintf("packer-%d", time.Now().UnixNano())

			diskName := fmt.Sprintf("%s-%s", testID, "disk")
			snapshotName := fmt.Sprintf("%s-%s", testID, "snapshot")
			imageName := fmt.Sprintf("%s-%s", testID, "image")

			var packerTemplate strings.Builder
			if err := tmpl.ExecuteTemplate(&packerTemplate, "image.pkr.hcl.tmpl", struct {
				Name      string
				Project   string
				SiloImage bool
			}{
				Name:      imageName,
				Project:   tc.project,
				SiloImage: tc.siloImage,
			}); err != nil {
				t.Fatalf("failed rendering packer template: %v", err)
			}

			acctest.TestPlugin(t, &acctest.PluginTestCase{
				Name:     tc.testName,
				Type:     "oxide-image",
				Template: packerTemplate.String(),
				Setup: func() error {
					t.Logf("setup: creating oxide disk %s", diskName)
					disk, err := oxideClient.DiskCreate(t.Context(), oxide.DiskCreateParams{
						Project: oxide.NameOrId(tc.project),
						Body: &oxide.DiskCreate{
							Name:        oxide.Name(diskName),
							Description: fmt.Sprintf("Packer acceptance test %s.", testID),
							DiskSource: oxide.DiskSource{
								BlockSize: 4096,
								Type:      oxide.DiskSourceTypeBlank,
							},
							Size: 1024 * 1024 * 1024, // 1 GiB.
						},
					})
					if err != nil {
						return fmt.Errorf("failed creating disk %s: %v", diskName, err)
					}

					t.Logf("setup: creating oxide snapshot %s", snapshotName)
					snapshot, err := oxideClient.SnapshotCreate(t.Context(), oxide.SnapshotCreateParams{
						Project: oxide.NameOrId(tc.project),
						Body: &oxide.SnapshotCreate{
							Name:        oxide.Name(snapshotName),
							Description: fmt.Sprintf("Packer acceptance test %s.", testID),
							Disk:        oxide.NameOrId(disk.Id),
						},
					})
					if err != nil {
						return fmt.Errorf("failed creating snapshot %s: %v", snapshotName, err)
					}

					t.Logf("setup: creating oxide image %s", imageName)
					_, err = oxideClient.ImageCreate(t.Context(), oxide.ImageCreateParams{
						Project: oxide.NameOrId(tc.project),
						Body: &oxide.ImageCreate{
							Name:        oxide.Name(imageName),
							Description: fmt.Sprintf("Packer acceptance test %s.", testID),
							Os:          "Blank",
							Version:     "0.0.0",
							Source: oxide.ImageSource{
								Id:   snapshot.Id,
								Type: oxide.ImageSourceTypeSnapshot,
							},
						},
					})
					if err != nil {
						return fmt.Errorf("failed creating image %s: %v", imageName, err)
					}

					if tc.siloImage {
						t.Logf("setup: promoting oxide image %s", imageName)
						if _, err := oxideClient.ImagePromote(t.Context(), oxide.ImagePromoteParams{
							Image:   oxide.NameOrId(imageName),
							Project: oxide.NameOrId(oxideProject),
						}); err != nil {
							return fmt.Errorf("failed promoting oxide image %s: %v", imageName, err)
						}
					}

					return nil
				},
				Teardown: func() error {
					var joinedError error

					if tc.siloImage {
						t.Logf("teardown: demoting oxide image %s", imageName)
						if _, err := oxideClient.ImageDemote(t.Context(), oxide.ImageDemoteParams{
							Image:   oxide.NameOrId(imageName),
							Project: oxide.NameOrId(oxideProject),
						}); err != nil {
							joinedError = errors.Join(joinedError, fmt.Errorf("failed demoting oxide image %s: %v", imageName, err))
						}
					}

					t.Logf("teardown: deleting oxide image %s", imageName)
					if err := oxideClient.ImageDelete(t.Context(), oxide.ImageDeleteParams{
						Image:   oxide.NameOrId(imageName),
						Project: oxide.NameOrId(oxideProject),
					}); err != nil {
						joinedError = errors.Join(joinedError, fmt.Errorf("failed deleting oxide image %s: %v", imageName, err))
					}

					t.Logf("teardown: deleting oxide snapshot %s", snapshotName)
					if err := oxideClient.SnapshotDelete(t.Context(), oxide.SnapshotDeleteParams{
						Snapshot: oxide.NameOrId(snapshotName),
						Project:  oxide.NameOrId(oxideProject),
					}); err != nil {
						joinedError = errors.Join(joinedError, fmt.Errorf("failed deleting oxide snapshot %s: %v", snapshotName, err))
					}

					t.Logf("teardown: deleting oxide disk %s", diskName)
					if err := oxideClient.DiskDelete(t.Context(), oxide.DiskDeleteParams{
						Disk:    oxide.NameOrId(diskName),
						Project: oxide.NameOrId(oxideProject),
					}); err != nil {
						joinedError = errors.Join(joinedError, fmt.Errorf("failed deleting oxide disk %s: %v", imageName, err))
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
