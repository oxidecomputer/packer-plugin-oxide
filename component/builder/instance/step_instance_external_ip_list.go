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
func (s *stepInstanceExternalIPList) Run(
	ctx context.Context,
	stateBag multistep.StateBag,
) multistep.StepAction {
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

	// Filter out invalid external IPs (e.g., SNAT) and extract the IP.
	var externalIP string
	for _, eip := range results.Items {
		switch eip.Kind() {
		case oxide.ExternalIpKindEphemeral:
			if v, ok := eip.AsEphemeral(); ok {
				externalIP = v.Ip
			}
		case oxide.ExternalIpKindFloating:
			if v, ok := eip.AsFloating(); ok {
				externalIP = v.Ip
			}
		}
		if externalIP != "" {
			break
		}
	}

	if externalIP == "" {
		ui.Error(
			"Instance does not have any valid external IPs. Packer will be unable to connect to this instance.",
		)
		return multistep.ActionHalt
	}

	stateBag.Put("external_ip", externalIP)

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepInstanceExternalIPList.Run].
func (s *stepInstanceExternalIPList) Cleanup(stateBag multistep.StateBag) {}
