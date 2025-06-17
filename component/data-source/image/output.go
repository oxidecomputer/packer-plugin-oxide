// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Output
//go:generate packer-sdc struct-markdown

package image

// Output represents the information returned by this plugin component for use
// in other Packer plugin components.
type Output struct {
	// ID of the image that was fetched.
	ImageID string `mapstructure:"image_id"`
}
