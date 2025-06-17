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

var _ multistep.Step = (*stepSnapshotCreate)(nil)

// stepSnapshotCreate is a Packer plugin step to create a snapshot from an Oxide
// instance's boot disk.
type stepSnapshotCreate struct{}

// Run creates an Oxide snapshot and stores its information in stateBag.
func (s *stepSnapshotCreate) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	ui.Say("Creating Oxide snapshot")

	bootDiskIDRaw, ok := stateBag.GetOk("boot_disk_id")
	if !ok {
		ui.Error("State does not contain boot disk ID. Cannot proceed!")
		return multistep.ActionHalt
	}
	bootDiskID := bootDiskIDRaw.(string)

	snapshot, err := oxideClient.SnapshotCreate(ctx, oxide.SnapshotCreateParams{
		Project: oxide.NameOrId(config.Project),
		Body: &oxide.SnapshotCreate{
			Name:        oxide.Name(config.Name),
			Description: "Managed by Packer.",
			Disk:        oxide.NameOrId(bootDiskID),
		},
	})
	if err != nil {
		ui.Error("Failed creating Oxide snapshot.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Sayf("Created Oxide snapshot: %s", snapshot.Id)

	stateBag.Put("snapshot_id", snapshot.Id)

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepSnapshotCreate.Run].
func (s *stepSnapshotCreate) Cleanup(stateBag multistep.StateBag) {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	if snapshotIDRaw, ok := stateBag.GetOk("snapshot_id"); ok {
		snapshotID := snapshotIDRaw.(string)

		ui.Sayf("Deleting Oxide snapshot: %s", snapshotID)

		if err := oxideClient.SnapshotDelete(context.Background(), oxide.SnapshotDeleteParams{
			Snapshot: oxide.NameOrId(snapshotID),
		}); err != nil {
			ui.Errorf("Failed deleting Oxide snapshot during cleanup. Please delete it manually: %v", err)
		}
	}
}
