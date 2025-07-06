// Copyright (c) Oxide Computer Company
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package imagepromote

import (
	"errors"
	"os"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The Oxide API host. Defaults to the `OXIDE_HOST` environment variable.
	Host string `mapstructure:"host" required:"false"`
	// The Oxide authentication token. Defaults to the `OXIDE_TOKEN` environment variable.
	Token string `mapstructure:"token" required:"false"`
	// The name or ID of the project where the promoted image should be made available.
	// If not specified, the image will be promoted to be available at the silo level.
	Project string `mapstructure:"project" required:"false"`

	ctx interpolate.Context
}

func (c *Config) ConfigSpec() hcldec.ObjectSpec {
	return c.FlatMapstructure().HCL2Spec()
}

func (c *Config) Prepare(raws ...interface{}) error {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         "oxide-image-promote",
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return err
	}

	var errs *packer.MultiError

	// Set defaults from environment if not provided
	if c.Host == "" {
		c.Host = os.Getenv("OXIDE_HOST")
	}
	if c.Token == "" {
		c.Token = os.Getenv("OXIDE_TOKEN")
	}

	// Validate required fields
	if c.Host == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("host is required"))
	}
	if c.Token == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("token is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}
