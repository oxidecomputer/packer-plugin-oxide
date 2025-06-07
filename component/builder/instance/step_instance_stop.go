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

var _ multistep.Step = (*stepInstanceStop)(nil)

// stepInstanceStop is a Packer plugin step to stop an Oxide instance.
type stepInstanceStop struct{}

// Run stops an Oxide instance and waits for it to be stopped.
func (s *stepInstanceStop) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	ui.Say("Stopping Oxide instance")

	instanceIDRaw, ok := stateBag.GetOk("instance_id")
	if !ok {
		ui.Error("State does not contain instance ID. Cannot proceed!")
		return multistep.ActionHalt
	}
	instanceID := instanceIDRaw.(string)

	if _, err := oxideClient.InstanceStop(ctx, oxide.InstanceStopParams{
		Instance: oxide.NameOrId(instanceID),
	}); err != nil {
		ui.Error("Failed stopping Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			ui.Error("Timed out waiting for Oxide instance to stop.")
			return multistep.ActionHalt
		default:
		}

		instance, err := oxideClient.InstanceView(timeoutCtx, oxide.InstanceViewParams{
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

		ui.Say(fmt.Sprintf("Waiting for instance to stop. Instance is currently %s.", instance.RunState))
		time.Sleep(3 * time.Second)
	}

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepInstanceStop.Run].
func (s *stepInstanceStop) Cleanup(stateBag multistep.StateBag) {}
