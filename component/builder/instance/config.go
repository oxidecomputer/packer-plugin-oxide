// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package instance

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

// The configuration arguments for this builder component. Arguments can
// either be required or optional.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Packer communicator configuration to configure how Packer connects to the
    // instance for provisioning.
	Comm communicator.Config `mapstructure:",squash"`

	// Oxide API URL (e.g., `https://oxide.sys.example.com`). If not specified, this
	// defaults to the value of the `OXIDE_HOST` environment variable.
	Host string `mapstructure:"host" required:"true"`

	// Oxide API token. If not specified, this defaults to the value of the
	// `OXIDE_TOKEN` environment variable.
	Token string `mapstructure:"token" required:"true"`

	// Image ID to use for the instance's boot disk. This can be obtained from the
	// `oxide-image` data source.
	BootDiskImageID string `mapstructure:"boot_disk_image_id" required:"true"`

	// Name or ID of the project where the temporary instance and resulting image
	// will be created.
	Project string `mapstructure:"project" required:"true"`

	// Size of the boot disk in bytes. Defaults to `21474836480`, or 20 GiB.
	BootDiskSize uint64 `mapstructure:"boot_disk_size"`

	// IP pool to allocate the instance's external IP from. If not specified, the
	// silo's default IP pool will be used.
	IPPool string `mapstructure:"ip_pool"`

	// VPC to create the instance within. Defaults to `default`.
	VPC string `mapstructure:"vpc"`

	// Subnet to create the instance within. Defaults to `default`.
	Subnet string `mapstructure:"subnet"`

	// Name of the temporary instance. Defaults to `packer-{{timestamp}}`.
	Name string `mapstructure:"name"`

	// Hostname of the temporary instance. Defaults to `packer-{{timestamp}}`.
	Hostname string `mapstructure:"hostname"`

	// Number of vCPUs to provision the instance with. Defaults to `1`.
	CPUs uint64 `mapstructure:"cpus"`

	// Amount of memory to provision the instance with, in bytes. Defaults to
	// `2147483648`, or 2 GiB.
	Memory uint64 `mapstructure:"memory"`

	// An array of names or IDs of SSH public keys to inject into the instance.
	SSHPublicKeys []string `mapstructure:"ssh_public_keys"`

	// Name of the resulting image artifact. Defaults to `packer-{{timestamp}}`.
	ArtifactName string `mapstructure:"artifact_name"`

	ctx interpolate.Context
}

// Prepare decodes the configuration and validates it.
func (c *Config) Prepare(args ...any) ([]string, error) {
	var metadata mapstructure.Metadata

	if err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &metadata,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		PluginType:         "packer.builder.oxide",
	}, args...); err != nil {
		return nil, fmt.Errorf("failed decoding configuration: %w", err)
	}

	// Defaults.
	{
		if c.Host == "" {
			c.Host = os.Getenv("OXIDE_HOST")
		}

		if c.Token == "" {
			c.Token = os.Getenv("OXIDE_TOKEN")
		}

		if c.Name == "" {
			name, err := interpolate.Render("packer-{{timestamp}}", nil)
			if err != nil {
				return nil, fmt.Errorf("failed rendering default name, this bug should be reported: %w", err)
			}

			c.Name = name
		}

		if c.Hostname == "" {
			hostname, err := interpolate.Render("packer-{{timestamp}}", nil)
			if err != nil {
				return nil, fmt.Errorf("failed rendering default hostname, this bug should be reported: %w", err)
			}

			c.Hostname = hostname
		}

		if c.ArtifactName == "" {
			artifactName, err := interpolate.Render("packer-{{timestamp}}", nil)
			if err != nil {
				return nil, fmt.Errorf("failed rendering default artifact name, this bug should be reported: %w", err)
			}

			c.ArtifactName = artifactName
		}

		if c.CPUs == 0 {
			c.CPUs = 1
		}

		if c.Memory == 0 {
			c.Memory = 2 * 1024 * 1024 * 1024 // 2 GiB.
		}

		if c.BootDiskSize == 0 {
			c.BootDiskSize = 20 * 1024 * 1024 * 1024 // 20 GiB.
		}

		if c.VPC == "" {
			c.VPC = "default"
		}

		if c.Subnet == "" {
			c.Subnet = "default"
		}
	}

	{
		var multiErr *packer.MultiError

		if errs := c.Comm.Prepare(&c.ctx); len(errs) > 0 {
			multiErr = packer.MultiErrorAppend(multiErr, errs...)
		}

		if multiErr != nil && len(multiErr.Errors) > 0 {
			return nil, multiErr
		}
	}

	packer.LogSecretFilter.Set(c.Token)

	return nil, nil
}
