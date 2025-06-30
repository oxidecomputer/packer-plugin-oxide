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

// stepSSHKeyCreate is a Packer plugin step to create an Oxide SSH public key.
type stepSSHKeyCreate struct{}

// Run creates an Oxide SSH public key and stores its information in stateBag.
func (s *stepSSHKeyCreate) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)
	config := stateBag.Get("config").(*Config)

	if config.Comm.SSHPublicKey == nil {
		ui.Say("No SSH public key found. Skipping SSH public key create...")
		return multistep.ActionContinue
	}

	ui.Say("Creating Oxide SSH public key")

	sshKey, err := oxideClient.CurrentUserSshKeyCreate(ctx, oxide.CurrentUserSshKeyCreateParams{
		Body: &oxide.SshKeyCreate{
			Description: "Created by Packer.",
			Name:        oxide.Name(config.Comm.SSHTemporaryKeyPairName),
			PublicKey:   string(config.Comm.SSHPublicKey),
		},
	})
	if err != nil {
		ui.Error("Failed creating Oxide SSH public key.")
		stateBag.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Sayf("Created Oxide SSH public key: %s", sshKey.Id)

	stateBag.Put("ssh_public_key_name", sshKey.Name)
	stateBag.Put("ssh_public_key_id", sshKey.Id)

	return multistep.ActionContinue
}

// Cleanup deletes the resources created by [stepSSHKeyCreate.Run].
func (s *stepSSHKeyCreate) Cleanup(stateBag multistep.StateBag) {
	oxideClient := stateBag.Get("client").(*oxide.Client)
	ui := stateBag.Get("ui").(packer.Ui)

	if sshPublicKeyIDRaw, ok := stateBag.GetOk("ssh_public_key_id"); ok {
		sshPublicKeyID := sshPublicKeyIDRaw.(string)

		ui.Sayf("Deleting Oxide SSH public key: %s", sshPublicKeyID)

		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancel()

		if err := oxideClient.CurrentUserSshKeyDelete(ctx, oxide.CurrentUserSshKeyDeleteParams{
			SshKey: oxide.NameOrId(sshPublicKeyID),
		}); err != nil {
			ui.Errorf("Failed deleting Oxide SSH public key %s during cleanup. Please delete it manually: %v", sshPublicKeyID, err)
			return
		}
	}
}
