// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ multistep.Step = (*stepInstanceDelete)(nil)

type stepInstanceDelete struct{}

func (s *stepInstanceDelete) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	// config := stateBag.Get("config").(*Config)

	instanceID := stateBag.Get("instance_id").(string)

	ui.Say("Deleting Oxide instance!")

	if _, err := oxideClient.InstanceStop(ctx, oxide.InstanceStopParams{
		Instance: oxide.NameOrId(instanceID),
	}); err != nil {
		ui.Error("Failed stopping Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	for range 5 {
		instance, err := oxideClient.InstanceView(ctx, oxide.InstanceViewParams{
			Instance: oxide.NameOrId(instanceID),
		})
		if err != nil {
			ui.Error("Failed viewing Oxide instance.")
			stateBag.Put("error", err)
			return multistep.ActionHalt
		}

		if instance.RunState == oxide.InstanceStateStopped {
			break
		}

		ui.Say(fmt.Sprintf("Waiting for instance to stop...(%v)", instance.RunState))
		time.Sleep(3 * time.Second)
	}

	if err := oxideClient.InstanceDelete(ctx, oxide.InstanceDeleteParams{
		Instance: oxide.NameOrId(instanceID),
	}); err != nil {
		ui.Error("Failed deleting Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	stateBag.Remove("instance_id")

	return multistep.ActionContinue
}

func (s *stepInstanceDelete) Cleanup(stateBag multistep.StateBag) {}
