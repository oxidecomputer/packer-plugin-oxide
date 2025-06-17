packer {
  required_plugins {
    oxide = {
      version = ">= 0.0.1"
      source  = "github.com/oxidecomputer/oxide"
    }
  }
}

data "oxide-image" "ubuntu" {
  name = "noble"
}

source "oxide-instance" "example" {
  project            = "matthewsanabria"
  boot_disk_image_id = data.oxide-image.ubuntu.image_id
  ip_pool            = "eng-vpn"

  ssh_public_keys = ["529885a0-2919-463a-a588-ac48f100a165"]

  ssh_username   = "ubuntu"
  ssh_agent_auth = true
}

build {
  sources = [
    "source.oxide-instance.example",
  ]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y ripgrep",
    ]
  }
}
