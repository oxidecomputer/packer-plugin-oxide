// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc struct-markdown

package image

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/zclconf/go-cty/cty"
)

var _ packer.Datasource = (*DataSource)(nil)

// DataSource is the concrete type that implements the Packer data source
// component interface.
type DataSource struct {
	config Config
}

// ConfigSpec returns the HCL specification that Packer uses to validate and
// configure this plugin component.
func (d *DataSource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

// Configure decodes the configuration for this plugin component, checks whether
// the configuration is valid, and stores any necessary state for future methods
// to use during execution.
func (d *DataSource) Configure(args ...any) error {
	if err := config.Decode(&d.config, nil, args...); err != nil {
		return fmt.Errorf("failed decoding configuration: %w", err)
	}

	// Set defaults.
	{
		if d.config.Host == "" {
			d.config.Host = os.Getenv("OXIDE_HOST")
		}

		if d.config.Token == "" {
			d.config.Token = os.Getenv("OXIDE_TOKEN")
		}
	}

	// Enforce required configuration.
	{
		var multiErr *packer.MultiError

		if d.config.Host == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("host is required"))
		}

		if d.config.Token == "" {
			multiErr = packer.MultiErrorAppend(multiErr, errors.New("token is required"))
		}

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
func (d *DataSource) Execute() (cty.Value, error) {
	oxideClient, err := oxide.NewClient(&oxide.Config{
		Host:  d.config.Host,
		Token: d.config.Token,
	})
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

	output := Output{
		ImageID: image.Id,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

// OutputSpec returns the HCL specification that Packer uses to populate output
// values for this plugin component.
func (d *DataSource) OutputSpec() hcldec.ObjectSpec {
	return (&Output{}).FlatMapstructure().HCL2Spec()
}
