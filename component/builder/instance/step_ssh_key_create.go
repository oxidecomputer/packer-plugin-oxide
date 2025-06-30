// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"context"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ multistep.Step = (*stepSSHKeyCreate)(nil)

// stepSSHKeyCreate is a Packer plugin step to create an SSH key.
type stepSSHKeyCreate struct{}

// Run creates an SSH key and stores its information in stateBag.
func (s *stepSSHKeyCreate) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	if config.Comm.SSHPublicKey == nil {
		ui.Say("No public SSH key found; skipping SSH public key create...")
		return multistep.ActionContinue
	}

	ui.Say("Creating Oxide SSH key")

	sshKey, err := oxideClient.CurrentUserSshKeyCreate(ctx, oxide.CurrentUserSshKeyCreateParams{
		Body: &oxide.SshKeyCreate{
			Description: "Created by Packer.",
			Name:        oxide.Name(config.Name),
			PublicKey:   string(config.Comm.SSHPublicKey),
		},
	})
	if err != nil {
		ui.Error("Failed creating Oxide SSH key.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Sayf("Created Oxide SSH key: %s", sshKey.Id)

	stateBag.Put("ssh_key_id", sshKey.Id)

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepSSHKeyCreate.Run].
func (s *stepSSHKeyCreate) Cleanup(stateBag multistep.StateBag) {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	if sshKeyIDRaw, ok := stateBag.GetOk("ssh_key_id"); ok {
		sshKeyID := sshKeyIDRaw.(string)

		ui.Sayf("Deleting Oxide SSH key: %s", sshKeyID)

		instanceDeleteCtx, instanceDeleteCtxCancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer instanceDeleteCtxCancel()

		if err := oxideClient.CurrentUserSshKeyDelete(instanceDeleteCtx, oxide.CurrentUserSshKeyDeleteParams{
			SshKey: oxide.NameOrId(sshKeyID),
		}); err != nil {
			ui.Errorf("Failed deleting Oxide SSH key during cleanup. Please delete it manually: %v", err)
			return
		}
	}
}
