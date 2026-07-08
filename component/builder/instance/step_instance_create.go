// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"encoding/base64"
	"errors"
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
			UserData: func(userData string) string {
				if userData == "" {
					return ""
				}
				return base64.StdEncoding.EncodeToString([]byte(userData))
			}(config.UserData),
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

	startCtx, startCtxCancel := context.WithTimeout(ctx, 60*time.Second)
	defer startCtxCancel()

	for {
		refreshedInstance, err := oxideClient.InstanceView(startCtx, oxide.InstanceViewParams{
			Instance: oxide.NameOrId(instance.Id),
		})
		if err != nil {
			ui.Error("Failed refreshing Oxide instance state.")
			stateBag.Put("error", err)
			return multistep.ActionHalt
		}

		if refreshedInstance.RunState == oxide.InstanceStateRunning {
			ui.Say(fmt.Sprintf("Oxide instance is %s.", refreshedInstance.RunState))
			break
		}

		ui.Say(
			fmt.Sprintf(
				"Waiting for Oxide instance to start: Currently %s.",
				refreshedInstance.RunState,
			),
		)

		select {
		case <-startCtx.Done():
			ui.Error("Timed out waiting for Oxide instance to start.")
			stateBag.Put("error", startCtx.Err())
			return multistep.ActionHalt
		case <-time.After(5 * time.Second):
		}
	}

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepInstanceCreate.Run].
func (o *stepInstanceCreate) Cleanup(stateBag multistep.StateBag) {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	ctx := context.Background()

	if instanceIDRaw, ok := stateBag.GetOk("instance_id"); ok {
		instanceID := instanceIDRaw.(string)

		instanceCtx, instanceCtxCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer instanceCtxCancel()

		ui.Sayf("Cleaning up Oxide instance: %s", instanceID)

		if err := o.cleanupInstance(instanceCtx, oxideClient, instanceID); err != nil {
			ui.Errorf(
				"Failed cleaning up Oxide instance: %s\n\tPlease delete it manually: %v",
				instanceID,
				err,
			)
		}
	}

	if bootDiskIDRaw, ok := stateBag.GetOk("boot_disk_id"); ok {
		bootDiskID := bootDiskIDRaw.(string)

		bootDiskCtx, bootDiskCtxCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer bootDiskCtxCancel()

		ui.Sayf("Cleaning up Oxide disk: %s", bootDiskID)

		if err := oxideClient.DiskDelete(bootDiskCtx, oxide.DiskDeleteParams{
			Disk: oxide.NameOrId(bootDiskID),
		}); err != nil && !errors.Is(err, oxide.ErrObjectNotFound) {
			ui.Errorf(
				"Failed cleaning up Oxide disk: %s\n\tPlease delete it manually: %v",
				bootDiskID,
				err,
			)
		}
	}
}

// cleanupInstance deletes the Oxide instance specified by instanceID.
func (o *stepInstanceCreate) cleanupInstance(
	ctx context.Context,
	oxideClient *oxide.Client,
	instanceID string,
) error {
	instance, err := oxideClient.InstanceView(ctx, oxide.InstanceViewParams{
		Instance: oxide.NameOrId(instanceID),
	})
	if err != nil {
		if errors.Is(err, oxide.ErrObjectNotFound) {
			return nil
		}

		return fmt.Errorf("failed fetching instance details: %w", err)
	}

	if !instanceDeletable(instance.RunState) {
		if instance.RunState != oxide.InstanceStateStopping {
			if _, err := oxideClient.InstanceStop(ctx, oxide.InstanceStopParams{
				Instance: oxide.NameOrId(instanceID),
			}); err != nil {
				return fmt.Errorf("failed issuing instance stop request: %w", err)
			}
		}

		for {
			refreshedInstance, err := oxideClient.InstanceView(ctx, oxide.InstanceViewParams{
				Instance: oxide.NameOrId(instanceID),
			})
			if err != nil {
				if errors.Is(err, oxide.ErrObjectNotFound) {
					return nil
				}
				// Transient error. Wait and retry below.
			} else if instanceDeletable(refreshedInstance.RunState) {
				break
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("timed out waiting for instance to stop: %w", ctx.Err())
			case <-time.After(5 * time.Second):
			}
		}
	}

	if err := oxideClient.InstanceDelete(ctx, oxide.InstanceDeleteParams{
		Instance: oxide.NameOrId(instanceID),
	}); err != nil && !errors.Is(err, oxide.ErrObjectNotFound) {
		return err
	}

	return nil
}

// instanceDeletable reports whether an instance in the given run state can be
// deleted without first being stopped.
func instanceDeletable(state oxide.InstanceState) bool {
	switch state {
	case oxide.InstanceStateStopped,
		oxide.InstanceStateFailed,
		oxide.InstanceStateDestroyed:
		return true
	default:
		return false
	}
}
