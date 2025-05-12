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

var _ multistep.Step = (*stepInstanceCreate)(nil)

type stepInstanceCreate struct{}

func (o *stepInstanceCreate) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	ui.Say("Creating Oxide instance!")

	now := time.Now().Unix()

	instance, err := oxideClient.InstanceCreate(ctx, oxide.InstanceCreateParams{
		Project: oxide.NameOrId(config.Project),
		Body: &oxide.InstanceCreate{
			AntiAffinityGroups: []oxide.NameOrId{},
			BootDisk: &oxide.InstanceDiskAttachment{
				Description: "Created by Packer.",
				DiskSource: oxide.DiskSource{
					Type:    oxide.DiskSourceTypeImage,
					ImageId: config.ImageID,
				},
				Name: oxide.Name(fmt.Sprintf("packer-%d", now)),
				Size: oxide.ByteCount(21474836480),
				Type: oxide.InstanceDiskAttachmentTypeCreate,
			},
			Description: "Created by Packer.",
			ExternalIps: []oxide.ExternalIpCreate{
				{
					Type: oxide.ExternalIpCreateTypeEphemeral,
					Pool: oxide.NameOrId("eng-vpn"),
				},
			},
			Hostname: oxide.Hostname(fmt.Sprintf("packer-%d", now)),
			Memory:   oxide.ByteCount(8589934592),
			Name:     oxide.Name(fmt.Sprintf("packer-%d", now)),
			Ncpus:    oxide.InstanceCpuCount(2),
			NetworkInterfaces: oxide.InstanceNetworkInterfaceAttachment{
				Params: []oxide.InstanceNetworkInterfaceCreate{
					{
						Description: "Created by Packer.",
						Name:        oxide.Name(fmt.Sprintf("packer-%d", now)),
						SubnetName:  oxide.Name("default"),
						VpcName:     oxide.Name("default"),
					},
				},
				Type: oxide.InstanceNetworkInterfaceAttachmentTypeCreate,
			},
			SshPublicKeys: []oxide.NameOrId{
				oxide.NameOrId("529885a0-2919-463a-a588-ac48f100a165"),
			},
		},
	})
	if err != nil {
		ui.Error("Failed creating Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	stateBag.Put("instance_id", instance.Id)

	return multistep.ActionContinue
}

func (o *stepInstanceCreate) Cleanup(stateBag multistep.StateBag) {}
