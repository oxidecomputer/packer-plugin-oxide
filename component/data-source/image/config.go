// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package image

// Config represents the Packer configuration for this data source component.
type Config struct {
	// Oxide API address (e.g., https://oxide.sys.example.com).
	Host string `mapstructure:"host"`

	// Oxide API token.
	Token string `mapstructure:"token"`

	// Name of the image to fetch.
	Name string `mapstructure:"name"`

	// Name or ID of the project containing the image to fetch. Leave blank to fetch
	// a silo image instead of a project image.
	Project string `mapstructure:"project"`
}
