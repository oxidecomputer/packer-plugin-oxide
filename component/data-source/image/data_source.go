// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc struct-markdown

package image

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/zclconf/go-cty/cty"
)

var _ packer.Datasource = (*Datasource)(nil)

// The `oxide-image` data source fetches [Oxide](https://oxide.computer) image
// information for use in a Packer build. The image can be a project image or
// silo image.
type Datasource struct {
	config Config
}

// ConfigSpec returns the HCL specification that Packer uses to validate and
// configure this plugin component.
func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

// Configure decodes the configuration for this plugin component, checks whether
// the configuration is valid, and stores any necessary state for future methods
// to use during execution.
func (d *Datasource) Configure(args ...any) error {
	if err := config.Decode(&d.config, nil, args...); err != nil {
		return fmt.Errorf("failed decoding configuration: %w", err)
	}

	// Enforce required configuration.
	{
		var multiErr *packer.MultiError

		if d.config.Name == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("name is required"))
		}

		if multiErr != nil && len(multiErr.Errors) > 0 {
			return multiErr
		}
	}

	return nil
}

// Execute fetches image information from the Oxide API and returns that
// information in the format specified by [OutputSpec].
func (d *Datasource) Execute() (cty.Value, error) {
	opts := make([]oxide.ClientOption, 0)
	if d.config.Host != "" {
		opts = append(opts, oxide.WithHost(d.config.Host))
	}
	if d.config.Token != "" {
		opts = append(opts, oxide.WithToken(d.config.Token))
	}
	if d.config.Profile != "" {
		opts = append(opts, oxide.WithProfile(d.config.Profile))
	}
	if d.config.InsecureSkipVerify {
		opts = append(opts, oxide.WithInsecureSkipVerify())
	}
	oxideClient, err := oxide.NewClient(opts...)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed creating oxide client: %w", err)
	}

	image, err := oxideClient.ImageView(context.TODO(), oxide.ImageViewParams{
		Image:   oxide.NameOrId(d.config.Name),
		Project: oxide.NameOrId(d.config.Project), // This relies on the Go SDK omitting empty strings from serialization to fetch silo images.
	})
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed fetching image %q within project %q: %w", d.config.Name, d.config.Project, err)
	}

	output := DatasourceOutput{
		ImageID: image.Id,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

// OutputSpec returns the HCL specification that Packer uses to populate output
// values for this plugin component.
func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}
