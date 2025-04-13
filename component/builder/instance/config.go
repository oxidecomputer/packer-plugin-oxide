// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package instance

import (
	"errors"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Comm communicator.Config `mapstructure:",squash"`

	Host  string `mapstructure:"host"`
	Token string `mapstructure:"token"`

	Project    string `mapstructure:"project"`
	ImageID    string `mapstructure:"image_id"`

	ctx interpolate.Context
}

func (c *Config) Prepare(args ...any) ([]string, error) {
	var metadata mapstructure.Metadata

	if err := config.Decode(c, &config.DecodeOpts{
		Metadata:    &metadata,
		Interpolate: true,
		PluginType:  "packer.builder.oxide",
	}, args...); err != nil {
		return nil, err
	}

	if c.Host == "" {
		c.Host = os.Getenv("OXIDE_HOST")
	}

	if c.Token == "" {
		c.Token = os.Getenv("OXIDE_TOKEN")
	}

	if errs := c.Comm.Prepare(&c.ctx); len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return nil, nil
}
