// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ multistep.Step = (*stepArtifactValidate)(nil)

// stepArtifactValidate is a Packer plugin step to validate whether there's no
// existing image that will conflict with the specified artifact name.
type stepArtifactValidate struct{}

// Run validates the artifact name has no conflicts, honoring Packer's `-force`
// flag to replace an existing image.
func (s *stepArtifactValidate) Run(
	ctx context.Context,
	stateBag multistep.StateBag,
) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	ui.Sayf("Validating artifact name: %s", config.ArtifactName)

	image, err := oxideClient.ImageView(ctx, oxide.ImageViewParams{
		Project: oxide.NameOrId(config.Project),
		Image:   oxide.NameOrId(config.ArtifactName),
	})
	if err != nil {
		if !errors.Is(err, oxide.ErrObjectNotFound) {
			ui.Error("Failed validating artifact name.")
			stateBag.Put("error", err)
			return multistep.ActionHalt
		}

		// The artifact name has no conflicts.
		return multistep.ActionContinue
	}

	// The artifact name conflicts with an existing image and `-force` was not
	// set. Bail so the user can decide what to do.
	if !config.PackerForce {
		ui.Errorf("Artifact name conflicts with existing image: %s (%s)", image.Name, image.Id)
		stateBag.Put(
			"error",
			fmt.Errorf(
				"Artifact name conflicts with existing image: %s (%s): use -force to overwrite",
				config.ArtifactName,
				image.Id,
			),
		)
	}

	// `-force` was set so we record the existing image information and tell the
	// user we'll overwrite the existing image later in the build.
	stateBag.Put("existing_image_id", string(image.Id))
	stateBag.Put("existing_image_name", string(image.Name))

	ui.Sayf(
		"Existing image will be overwritten (-force): %s (%s)",
		image.Name,
		image.Id,
	)

	return multistep.ActionContinue
}

// Cleanup does nothing as [stepArtifactValidate.Run] creates no resources.
func (s *stepArtifactValidate) Cleanup(stateBag multistep.StateBag) {}
