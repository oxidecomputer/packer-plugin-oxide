// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package image

// The configuration arguments for the data source. Arguments can either be
// required or optional.
type Config struct {
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

	// Name of the image to fetch.
	Name string `mapstructure:"name" required:"true"`

	// Name or ID of the project containing the image to fetch. Leave blank to fetch
	// a silo image instead of a project image.
	Project string `mapstructure:"project"`
}
