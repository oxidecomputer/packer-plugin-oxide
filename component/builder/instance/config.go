// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package instance

import (
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	"github.com/mitchellh/mapstructure"
)

// The configuration arguments for the builder. Arguments can either be required or optional.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Packer communicator configuration to configure how Packer connects to the
	// instance for provisioning.
	Comm communicator.Config `mapstructure:",squash"`

	// Oxide API URL (e.g., `https://oxide.sys.example.com`). If not specified,
	// this defaults to the value of the `OXIDE_HOST` environment variable. When
	// specified, `token` must be specified. Conflicts with `profile`.
	Host string `mapstructure:"host" required:"false"`

	// Oxide API token. If not specified, this defaults to the value of the
	// `OXIDE_TOKEN` environment variable. When specified, `host` must be specified.
	// Conflicts with `profile`.
	Token string `mapstructure:"token" required:"false"`

	// Oxide credentials profile. If not specified, this defaults to the value of
	// the `OXIDE_PROFILE` environment variable. Conflicts with `host` and `token`.
	Profile string `mapstructure:"profile" required:"false"`

	// Skip TLS certificate verification when connecting to the Oxide API.
	// Defaults to `false`.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" required:"false"`

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

	// Name of the temporary instance and its boot disk, primary network interface,
	// and snapshot. Defaults to `packer-UUID`, where `UUID` is a unique identifier
	// generated for the current build. This must be unique to prevent Oxide
	// resource name conflicts.
	Name string `mapstructure:"name"`

	// Hostname of the temporary instance. Defaults to the value of `name`.
	Hostname string `mapstructure:"hostname"`

	// Number of vCPUs to provision the instance with. Defaults to `1`.
	CPUs uint64 `mapstructure:"cpus"`

	// Amount of memory to provision the instance with, in bytes. Defaults to
	// `2147483648`, or 2 GiB.
	Memory uint64 `mapstructure:"memory"`

	// An array of names or IDs of SSH public keys to inject into the instance.
	SSHPublicKeys []string `mapstructure:"ssh_public_keys"`

	// Name of the resulting image artifact. Required unless `skip_create_image`
	// is `true`.
	ArtifactName string `mapstructure:"artifact_name" required:"true"`

	// Description of the resulting image artifact. Defaults to a description that
	// identifies the Packer build that created it.
	ArtifactDescription string `mapstructure:"artifact_description"`

	// Operating system of the resulting image artifact. Defaults to the OS of the
	// source image as retrieved from Oxide.
	ArtifactOS string `mapstructure:"artifact_os"`

	// Version of the resulting image artifact. Defaults to the version of the
	// source image as retrieved from Oxide.
	ArtifactVersion string `mapstructure:"artifact_version"`

	// Skip creating the final image. When set to `true`, the build will boot the
	// temporary instance and run all provisioners but will not create a snapshot
	// or image. The temporary instance is still cleaned up normally. This is useful
	// for testing provisioner logic without incurring the cost of image creation.
	// Defaults to `false`.
	SkipCreateImage bool `mapstructure:"skip_create_image" required:"false"`

	// User data for instance initialization systems such as cloud-init. The
	// value is a UTF-8 string and will be Base64-encoded by the plugin before
	// transmission. The maximum size is 32 KiB, measured before encoding. Use
	// Packer's built-in functions such as `file` to read content from disk.
	// Packer does not wait for user data to finish executing before shutting
	// down the instance. If your user data must complete before the image is
	// created, run `cloud-init status --wait` or an equivalent in a
	// provisioner.
	UserData string `mapstructure:"user_data" required:"false"`

	// Unique suffix shared by generated resource names for this build.
	generatedSuffix string
}

// Prepare decodes the configuration and validates it.
func (c *Config) Prepare(args ...any) ([]string, error) {
	var metadata mapstructure.Metadata

	if err := config.Decode(c, &config.DecodeOpts{
		Metadata:    &metadata,
		Interpolate: true,
		PluginType:  BuilderID,
	}, args...); err != nil {
		return nil, fmt.Errorf("failed decoding configuration: %w", err)
	}

	// Set defaults.
	{
		if c.Name == "" {
			c.Name = c.uniqueName("packer")
		}

		if c.Hostname == "" {
			c.Hostname = c.Name
		}

		if c.ArtifactDescription == "" {
			c.ArtifactDescription = c.buildDescription()
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

	// Enforce required configuration.
	{
		var multiErr *packer.MultiError

		if errs := c.Comm.Prepare(nil); len(errs) > 0 {
			multiErr = packer.MultiErrorAppend(multiErr, errs...)
		}

		if c.Comm.SSHTemporaryKeyPairName == "" {
			c.Comm.SSHTemporaryKeyPairName = c.uniqueName("packer")
		}

		c.Comm.SSHTemporaryKeyPairType = "ed25519"

		if c.Project == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("project is required"))
		}

		if c.BootDiskImageID == "" {
			multiErr = packer.MultiErrorAppend(
				multiErr,
				errors.New("boot_disk_image_id is required"),
			)
		}

		if c.ArtifactName == "" && !c.SkipCreateImage {
			multiErr = packer.MultiErrorAppend(
				multiErr,
				errors.New("artifact_name is required unless skip_create_image is true"),
			)
		}

		if len(c.UserData) > 32*1024 {
			multiErr = packer.MultiErrorAppend(
				multiErr,
				errors.New("user_data must be 32 KiB or less of unencoded data"),
			)
		}

		if multiErr != nil && len(multiErr.Errors) > 0 {
			return nil, multiErr
		}
	}

	packer.LogSecretFilter.Set(c.Token)

	return nil, nil
}

// buildDescription returns a description identifying the Packer build that
// created a resource.
func (c *Config) buildDescription() string {
	if c.PackerBuildName == "" {
		return "Created by Packer."
	}

	return fmt.Sprintf("Created by Packer build %q.", c.PackerBuildName)
}

// uniqueName returns a name containing the unique suffix for this build.
func (c *Config) uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, c.uniqueSuffix())
}

// uniqueSuffix returns the identifier used to generate unique resource names
// for this build.
func (c *Config) uniqueSuffix() string {
	if c.generatedSuffix == "" {
		c.generatedSuffix = uuid.TimeOrderedUUID()
	}

	return c.generatedSuffix
}
