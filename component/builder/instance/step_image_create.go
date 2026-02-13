// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ multistep.Step = (*stepImageCreate)(nil)

// stepImageCreate is a Packer plugin step to create an Oxide image.
type stepImageCreate struct{}

// Run creates an Oxide image and stores its information in stateBag.
func (s *stepImageCreate) Run(
	ctx context.Context,
	stateBag multistep.StateBag,
) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	snapshotID := stateBag.Get("snapshot_id").(string)

	ui.Say("Creating Oxide image")

	image, err := oxideClient.ImageCreate(ctx, oxide.ImageCreateParams{
		Project: oxide.NameOrId(config.Project),
		Body: &oxide.ImageCreate{
			Description: "Created by Packer.",
			Name:        oxide.Name(config.ArtifactName),
			Os:          config.ArtifactOS,
			Source: oxide.ImageSource{
				Type: oxide.ImageSourceTypeSnapshot,
				Id:   snapshotID,
			},
			Version: config.ArtifactVersion,
		},
	})
	if err != nil {
		ui.Error("Failed creating Oxide image.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Sayf("Created Oxide image: %s", image.Id)

	stateBag.Put("image_id", string(image.Id))
	stateBag.Put("image_name", string(image.Name))

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepImageCreate.Run].
func (s *stepImageCreate) Cleanup(stateBag multistep.StateBag) {}
