// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"errors"

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

	// `-force` is set so we'll delete the existing image before creating a new one.
	if config.PackerForce {
		existingImageID := stateBag.Get("existing_image_id").(string)
		existingImageName := stateBag.Get("existing_image_name").(string)

		ui.Sayf(
			"Deleting existing Oxide image %s (%s) because -force is set",
			existingImageName,
			existingImageID,
		)

		err := oxideClient.ImageDelete(ctx, oxide.ImageDeleteParams{
			Project: oxide.NameOrId(config.Project),
			Image:   oxide.NameOrId(existingImageName),
		})
		if err != nil && !errors.Is(err, oxide.ErrObjectNotFound) {
			ui.Error("Failed deleting existing Oxide image.")
			stateBag.Put("error", err)
			return multistep.ActionHalt
		}
	}

	ui.Say("Creating Oxide image")

	image, err := oxideClient.ImageCreate(ctx, oxide.ImageCreateParams{
		Project: oxide.NameOrId(config.Project),
		Body: &oxide.ImageCreate{
			Name:        oxide.Name(config.ArtifactName),
			Description: config.ArtifactDescription,
			Os:          config.ArtifactOS,
			Source: oxide.ImageSource{
				Value: &oxide.ImageSourceSnapshot{
					Id: snapshotID,
				},
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
