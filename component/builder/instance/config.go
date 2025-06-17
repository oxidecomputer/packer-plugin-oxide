// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package instance

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

// Config represents the Packer configuration for this builder component.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Packer communicator configuration to allow users to configure how Packer will
	// communicate with an instance.
	Comm communicator.Config `mapstructure:",squash"`

	// Oxide API address (e.g., https://oxide.sys.example.com).
	Host string `mapstructure:"host"`

	// Oxide API token.
	Token string `mapstructure:"token"`

	// Image ID to boot the instance from.
	BootDiskImageID string `mapstructure:"boot_disk_image_id"`

	// Name or ID of the project where the instance will be created.
	Project string `mapstructure:"project"`

	// Size of the boot disk, in bytes.
	BootDiskSize uint64 `mapstructure:"boot_disk_size"`

	// IP pool that the instance's external IP should be allocated from.
	IPPool string `mapstructure:"ip_pool"`

	// VPC to create the instance within.
	VPC string `mapstructure:"vpc"`

	// Subnet to create the instance within.
	Subnet string `mapstructure:"subnet"`

	// Instance name.
	Name string `mapstructure:"name"`

	// Instance hostname.
	Hostname string `mapstructure:"hostname"`

	// Number of vCPUs to provision the instance with.
	CPUs uint64 `mapstructure:"cpus"`

	// Amount of memory to provision the instance with, in bytes.
	Memory uint64 `mapstructure:"memory"`

	// Names or IDs of SSH public keys to inject into the instance.
	SSHPublicKeys []string `mapstructure:"ssh_public_keys"`

	// TODO: Name this better?
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

	// Configure defaults.
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

	// Validate required configuration.
	{
		var multiErr *packer.MultiError

		if c.Host == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("host is required"))
		}

		if c.Token == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("token is required"))
		}

		if c.Project == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("project is required"))
		}

		if c.BootDiskImageID == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("boot_disk_image_id is required"))
		}

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
