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

var _ multistep.Step = (*stepInstanceExternalIPList)(nil)

// stepInstanceExternalIPList is a Packer plugin step to list the external IP
// addresses for an Oxide instance.
type stepInstanceExternalIPList struct{}

// Run lists the external IP addresses for an Oxide instance and stores its
// information in stateBag.
func (s *stepInstanceExternalIPList) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	ui.Say("Listing external IPs for Oxide instance")

	instanceIDRaw, ok := stateBag.GetOk("instance_id")
	if !ok {
		ui.Error("State does not contain instance ID. Cannot proceed!")
		return multistep.ActionHalt
	}
	instanceID := instanceIDRaw.(string)

	results, err := oxideClient.InstanceExternalIpList(ctx, oxide.InstanceExternalIpListParams{
		Instance: oxide.NameOrId(instanceID),
	})
	if err != nil {
		ui.Error("Failed listing external IPs for Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	if len(results.Items) == 0 {
		ui.Error("Instance does not have any external IPs. Packer will be unable to connect to this instance.")
		return multistep.ActionHalt
	}

	stateBag.Put("external_ip", results.Items[0].Ip)

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepInstanceExternalIPList.Run].
func (s *stepInstanceExternalIPList) Cleanup(stateBag multistep.StateBag) {}
