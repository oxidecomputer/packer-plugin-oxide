// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput
//go:generate packer-sdc struct-markdown

package image

// The outputs returned by this data source component.
type DatasourceOutput struct {
	// ID of the image that was fetched.
	ImageID string `mapstructure:"image_id"`
}
