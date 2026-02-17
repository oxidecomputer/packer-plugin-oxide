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

// stepInstanceCreate is a Packer plugin step to create an Oxide instance.
type stepInstanceCreate struct{}

// Run creates an Oxide instance and stores its information in stateBag.
func (o *stepInstanceCreate) Run(
	ctx context.Context,
	stateBag multistep.StateBag,
) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	ui.Say("Creating Oxide instance")

	instance, err := oxideClient.InstanceCreate(ctx, oxide.InstanceCreateParams{
		Project: oxide.NameOrId(config.Project),
		Body: &oxide.InstanceCreate{
			AntiAffinityGroups: []oxide.NameOrId{},
			BootDisk: oxide.InstanceDiskAttachment{
				Value: &oxide.InstanceDiskAttachmentCreate{
					Name:        oxide.Name(config.Name),
					Description: "Created by Packer.",
					Size:        oxide.ByteCount(config.BootDiskSize),
					DiskBackend: oxide.DiskBackend{
						Value: &oxide.DiskBackendDistributed{
							DiskSource: oxide.DiskSource{
								Value: &oxide.DiskSourceImage{
									ImageId: config.BootDiskImageID,
								},
							},
						},
					},
				},
			},
			Description: "Created by Packer.",
			ExternalIps: []oxide.ExternalIpCreate{
				{
					Value: &oxide.ExternalIpCreateEphemeral{
						PoolSelector: func() oxide.PoolSelector {
							if config.IPPool == "" {
								return oxide.PoolSelector{
									Value: &oxide.PoolSelectorAuto{
										IpVersion: oxide.IpVersionV4,
									},
								}
							}

							return oxide.PoolSelector{
								Value: &oxide.PoolSelectorExplicit{
									Pool: oxide.NameOrId(config.IPPool),
								},
							}
						}(),
					},
				},
			},
			Hostname: oxide.Hostname(config.Hostname),
			Memory:   oxide.ByteCount(config.Memory),
			Name:     oxide.Name(config.Name),
			Ncpus:    oxide.InstanceCpuCount(config.CPUs),
			NetworkInterfaces: oxide.InstanceNetworkInterfaceAttachment{
				Value: &oxide.InstanceNetworkInterfaceAttachmentCreate{
					Params: []oxide.InstanceNetworkInterfaceCreate{
						{
							Name:        oxide.Name(config.Name),
							Description: "Created by Packer.",
							SubnetName:  oxide.Name(config.Subnet),
							VpcName:     oxide.Name(config.VPC),
							IpConfig: oxide.PrivateIpStackCreate{
								Value: oxide.PrivateIpStackCreateV4{
									Value: oxide.PrivateIpv4StackCreate{
										Ip: oxide.Ipv4Assignment{
											Value: &oxide.Ipv4AssignmentAuto{},
										},
									},
								},
							},
						},
					},
				},
			},
			SshPublicKeys: func(sshPublicKeys []string) []oxide.NameOrId {
				res := make([]oxide.NameOrId, 0, len(sshPublicKeys))

				if sshKeyIDRaw, ok := stateBag.GetOk("ssh_public_key_id"); ok {
					sshKeyID := sshKeyIDRaw.(string)
					res = append(res, oxide.NameOrId(sshKeyID))
				}

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

	ui.Sayf("Waiting for Oxide instance to start: Currently %s.", instance.RunState)

	startCtx, startCtxCancel := context.WithTimeout(ctx, 30*time.Second)
	defer startCtxCancel()

	for {
		select {
		case <-startCtx.Done():
			ui.Error("Timed out waiting for Oxide instance to start.")
			return multistep.ActionHalt
		default:
		}

		instance, err := oxideClient.InstanceView(startCtx, oxide.InstanceViewParams{
			Instance: oxide.NameOrId(instance.Id),
		})
		if err != nil {
			ui.Error("Failed refreshing Oxide instance state.")
			stateBag.Put("error", err)
			return multistep.ActionHalt
		}

		if instance.RunState == oxide.InstanceStateRunning {
			ui.Say(fmt.Sprintf("Oxide instance is %s.", instance.RunState))
			break
		}

		ui.Say(fmt.Sprintf("Waiting for Oxide instance to start: Currently %s.", instance.RunState))
		time.Sleep(3 * time.Second)
	}

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepInstanceCreate.Run].
func (o *stepInstanceCreate) Cleanup(stateBag multistep.StateBag) {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	if instanceIDRaw, ok := stateBag.GetOk("instance_id"); ok {
		instanceID := instanceIDRaw.(string)

		ui.Sayf("Checking if Oxide instance is stopped: %s", instanceID)

		instanceStopCtx, instanceStopCtxCancel := context.WithTimeout(
			context.TODO(),
			30*time.Second,
		)
		defer instanceStopCtxCancel()

		instance, err := oxideClient.InstanceView(instanceStopCtx, oxide.InstanceViewParams{
			Instance: oxide.NameOrId(instanceID),
		})
		if err != nil {
			ui.Errorf("Failed viewing Oxide instance state: %v", err)
			return
		}

		if instance.RunState == oxide.InstanceStateStopped {
			goto deleteInstance
		}

		ui.Sayf("Stopping Oxide instance: %s", instanceID)

		if _, err := oxideClient.InstanceStop(instanceStopCtx, oxide.InstanceStopParams{
			Instance: oxide.NameOrId(instanceID),
		}); err != nil {
			ui.Errorf(
				"Failed stopping Oxide instance during cleanup. Please delete it manually: %v",
				err,
			)
			return
		}

		for {
			select {
			case <-instanceStopCtx.Done():
				ui.Error("Timed out waiting for Oxide instance to stop.")
				return
			default:
			}

			instance, err := oxideClient.InstanceView(instanceStopCtx, oxide.InstanceViewParams{
				Instance: oxide.NameOrId(instanceID),
			})
			if err != nil {
				ui.Errorf("Failed refreshing Oxide instance state: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			if instance.RunState == oxide.InstanceStateStopped {
				break
			}

			ui.Say(
				fmt.Sprintf(
					"Waiting for instance to stop. Instance is currently %s.",
					instance.RunState,
				),
			)
			time.Sleep(3 * time.Second)
		}

	deleteInstance:
		ui.Sayf("Deleting Oxide instance: %s", instanceID)

		instanceDeleteCtx, instanceDeleteCtxCancel := context.WithTimeout(
			context.TODO(),
			30*time.Second,
		)
		defer instanceDeleteCtxCancel()

		if err := oxideClient.InstanceDelete(instanceDeleteCtx, oxide.InstanceDeleteParams{
			Instance: oxide.NameOrId(instanceID),
		}); err != nil {
			ui.Errorf(
				"Failed deleting Oxide instance during cleanup. Please delete it manually: %v",
				err,
			)
			return
		}
	}

	if bootDiskIDRaw, ok := stateBag.GetOk("boot_disk_id"); ok {
		bootDiskID := bootDiskIDRaw.(string)

		ui.Sayf("Deleting Oxide disk: %s", bootDiskID)

		diskDeleteCtx, diskDeleteCtxCancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer diskDeleteCtxCancel()

		if err := oxideClient.DiskDelete(diskDeleteCtx, oxide.DiskDeleteParams{
			Disk: oxide.NameOrId(bootDiskID),
		}); err != nil {
			ui.Errorf(
				"Failed deleting Oxide disk during cleanup. Please delete it manually: %v",
				err,
			)
			return
		}
	}
}
