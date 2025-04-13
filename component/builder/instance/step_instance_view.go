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

var _ multistep.Step = (*stepInstanceCreate)(nil)

type stepInstanceView struct{}

func (s *stepInstanceView) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	instanceID := stateBag.Get("instance_id").(string)

	ui.Say("Viewing Oxide instance!")

	results, err :=oxideClient.InstanceExternalIpList(ctx, oxide.InstanceExternalIpListParams{
		Instance: oxide.NameOrId(instanceID),
	})
	if err != nil {
		ui.Error("Failed viewing Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	if len(results.Items) == 0 {
		ui.Error("Failed retrieving external IPs for Oxide instance.")
		stateBag.Put("error", errors.New("empty external IPs"))
		return multistep.ActionHalt
	}

	stateBag.Put("external_ip", results.Items[0].Ip)

	return multistep.ActionContinue
}

func (s *stepInstanceView) Cleanup(stateBag multistep.StateBag) {}
