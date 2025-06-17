// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ packer.Builder = (*Builder)(nil)

// Builder is an implementation of [packer.Builder] for use with Oxide.
type Builder struct {
	config Config
	runner multistep.Runner
}

// ConfigSpec returns the HCL configuration specification for the builder.
func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	return b.config.FlatMapstructure().HCL2Spec()
}

// Prepare configures the builder and validates its configuration.
func (b *Builder) Prepare(args ...any) ([]string, []string, error) {
	warnings, err := b.config.Prepare(args...)
	if err != nil {
		return nil, warnings, err
	}

	return nil, warnings, nil
}

// Run executes the builder steps to create an Oxide image.
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	oxideClient, err := oxide.NewClient(&oxide.Config{
		Host:  b.config.Host,
		Token: b.config.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating oxide client: %w", err)
	}

	steps := []multistep.Step{
		&stepInstanceCreate{},
		&stepInstanceExternalIPList{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost(b.config.Comm.Host(), "external_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&stepInstanceStop{},
		&stepSnapshotCreate{},
		&stepImageCreate{},
	}

	stateBag := &multistep.BasicStateBag{}
	stateBag.Put("hook", hook)
	stateBag.Put("ui", ui)
	stateBag.Put("client", oxideClient)
	stateBag.Put("config", &b.config)

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, stateBag)

	if err, ok := stateBag.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		ImageID:   stateBag.Get("image_id").(string),
		ImageName: stateBag.Get("image_name").(string),
	}

	return artifact, nil
}
