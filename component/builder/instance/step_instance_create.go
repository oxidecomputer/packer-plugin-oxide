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

var _ multistep.Step = (*stepInstanceCreate)(nil)

// stepInstanceCreate is a Packer plugin step to create an Oxide instance.
type stepInstanceCreate struct{}

// Run creates an Oxide instance and stores its information in stateBag.
func (o *stepInstanceCreate) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	ui.Say("Creating Oxide instance")

	instance, err := oxideClient.InstanceCreate(ctx, oxide.InstanceCreateParams{
		Project: oxide.NameOrId(config.Project),
		Body: &oxide.InstanceCreate{
			AntiAffinityGroups: []oxide.NameOrId{},
			BootDisk: &oxide.InstanceDiskAttachment{
				Type:        oxide.InstanceDiskAttachmentTypeCreate,
				Name:        oxide.Name(config.Name),
				Description: "Created by Packer.",
				Size:        oxide.ByteCount(config.BootDiskSize),
				DiskSource: oxide.DiskSource{
					Type:    oxide.DiskSourceTypeImage,
					ImageId: config.BootDiskImageID,
				},
			},
			Description: "Created by Packer.",
			ExternalIps: []oxide.ExternalIpCreate{
				{
					Type: oxide.ExternalIpCreateTypeEphemeral,
					Pool: oxide.NameOrId(config.IPPool),
				},
			},
			Hostname: oxide.Hostname(config.Hostname),
			Memory:   oxide.ByteCount(config.Memory),
			Name:     oxide.Name(config.Name),
			Ncpus:    oxide.InstanceCpuCount(config.CPUs),
			NetworkInterfaces: oxide.InstanceNetworkInterfaceAttachment{
				Type: oxide.InstanceNetworkInterfaceAttachmentTypeCreate,
				Params: []oxide.InstanceNetworkInterfaceCreate{
					{
						Name:        oxide.Name(config.Name),
						Description: "Created by Packer.",
						SubnetName:  oxide.Name(config.Subnet),
						VpcName:     oxide.Name(config.VPC),
					},
				},
			},
			SshPublicKeys: func(sshPublicKeys []string) []oxide.NameOrId {
				res := make([]oxide.NameOrId, 0, len(sshPublicKeys))

				for _, sshPublicKey := range sshPublicKeys {
					res = append(res, oxide.NameOrId(sshPublicKey))
				}

				return res
			}(config.SSHPublicKeys),
		},
	})
	if err != nil {
		ui.Error("Failed creating Oxide instance.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Sayf("Created Oxide instance: %s", instance.Id)

	stateBag.Put("instance_id", instance.Id)
	stateBag.Put("boot_disk_id", instance.BootDiskId)

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepInstanceCreate.Run].
func (o *stepInstanceCreate) Cleanup(stateBag multistep.StateBag) {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	if instanceIDRaw, ok := stateBag.GetOk("instance_id"); ok {
		instanceID := instanceIDRaw.(string)

		ui.Sayf("Deleting Oxide instance: %s", instanceID)

		if err := oxideClient.InstanceDelete(context.TODO(), oxide.InstanceDeleteParams{
			Instance: oxide.NameOrId(instanceID),
		}); err != nil {
			ui.Errorf("Failed deleting Oxide instance during cleanup. Please delete it manually: %v", err)
		}
	}

	if bootDiskIDRaw, ok := stateBag.GetOk("boot_disk_id"); ok {
		bootDiskID := bootDiskIDRaw.(string)

		ui.Sayf("Deleting Oxide disk: %s", bootDiskID)

		if err := oxideClient.DiskDelete(context.TODO(), oxide.DiskDeleteParams{
			Disk: oxide.NameOrId(bootDiskID),
		}); err != nil {
			ui.Errorf("Failed deleting Oxide disk during cleanup. Please delete it manually: %v", err)
		}
	}
}
