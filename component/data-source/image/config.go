// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package image

// The configuration arguments for this data source component. Arguments can
// either be required or optional.
type Config struct {
	// Oxide API URL (e.g., `https://oxide.sys.example.com`). If not specified, this
	// defaults to the value of the `OXIDE_HOST` environment variable.
	Host string `mapstructure:"host" required:"true"`

	// Oxide API token. If not specified, this defaults to the value of the
	// `OXIDE_TOKEN` environment variable.
	Token string `mapstructure:"token" required:"true"`

	// Name of the image to fetch.
	Name string `mapstructure:"name" required:"true"`

	// Name or ID of the project containing the image to fetch. Leave blank to fetch
	// a silo image instead of a project image.
	Project string `mapstructure:"project"`
}
