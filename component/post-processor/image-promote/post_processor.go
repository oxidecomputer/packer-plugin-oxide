// Copyright (c) Oxide Computer Company
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown

package imagepromote

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/oxidecomputer/oxide.go/oxide"
)

const BuilderID = "oxide.image-promote"

type PostProcessor struct {
	config Config
}

var _ packer.PostProcessor = (*PostProcessor)(nil)

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.config.ConfigSpec()
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	return p.config.Prepare(raws...)
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	// Only process artifacts from the oxide-instance builder
	if artifact.BuilderId() != "oxide.instance" {
		err := fmt.Errorf("Unknown artifact type %s, this post-processor only works with oxide-instance artifacts", artifact.BuilderId())
		return nil, false, false, err
	}

	// Extract the image ID from the artifact
	imageID := artifact.State("ImageID")
	if imageID == nil {
		return nil, false, false, fmt.Errorf("No ImageID found in artifact")
	}

	imageIDStr, ok := imageID.(string)
	if !ok {
		return nil, false, false, fmt.Errorf("ImageID is not a string")
	}

	// Create Oxide client
	client, err := oxide.NewClient(&oxide.Config{
		Host:  p.config.Host,
		Token: p.config.Token,
	})
	if err != nil {
		return nil, false, false, fmt.Errorf("Failed to create Oxide client: %w", err)
	}

	ui.Say(fmt.Sprintf("Promoting image %s", imageIDStr))

	// Prepare promotion parameters
	params := oxide.ImagePromoteParams{
		Image: oxide.NameOrId(imageIDStr),
	}

	// If project is specified, set it
	if p.config.Project != "" {
		params.Project = oxide.NameOrId(p.config.Project)
		ui.Say(fmt.Sprintf("Promoting to project: %s", p.config.Project))
	} else {
		ui.Say("Promoting to silo level (no project specified)")
	}

	// Promote the image
	promotedImage, err := client.ImagePromote(ctx, params)
	if err != nil {
		return nil, false, false, fmt.Errorf("Failed to promote image: %w", err)
	}

	ui.Say(fmt.Sprintf("Successfully promoted image: %s", string(promotedImage.Name)))

	// Create a new artifact with the promoted image information
	newArtifact := &Artifact{
		ImageID:   promotedImage.Id,
		ImageName: string(promotedImage.Name),
		Project:   p.config.Project,
	}

	// Return the new artifact, keep the original, and indicate success
	return newArtifact, true, false, nil
}
