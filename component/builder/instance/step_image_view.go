// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ multistep.Step = (*stepImageView)(nil)

// stepImageView is a Packer plugin step to fetch an Oxide image.
type stepImageView struct{}

// Run fetches an Oxide image and populate configuration arguments.
func (s *stepImageView) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	ui.Say("Fetching Oxide image metadata")

	image, err := oxideClient.ImageView(ctx, oxide.ImageViewParams{
		Image: oxide.NameOrId(config.BootDiskImageID),
	})
	if err != nil {
		ui.Error("Failed fetching Oxide image metadata.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Sayf("Fetched Oxide image: %s", image.Id)

	timestamp, err := interpolate.Render("{{timestamp}}", &config.ctx)
	if err != nil {
		ui.Error("Failed rendering timestamp interpolation.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	if config.ArtifactName == "" {
		config.ArtifactName = fmt.Sprintf("%s-%s", image.Name, timestamp)
	}

	if config.ArtifactOS == "" {
		config.ArtifactOS = image.Os
	}

	if config.ArtifactVersion == "" {
		config.ArtifactVersion = fmt.Sprintf("%s-%s", image.Version, timestamp)
	}

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepImageView.Run].
func (s *stepImageView) Cleanup(stateBag multistep.StateBag) {}
