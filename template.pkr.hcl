packer {
  required_plugins {
    oxide = {
      version = ">= 0.0.1"
      source  = "github.com/oxidecomputer/oxide"
    }
  }
}

data "oxide-image" "ubuntu" {
  name    = "noble"
}

source "oxide-instance" "example" {
  project     = "matthewsanabria"
  image_id    = data.oxide-image.ubuntu.image_id

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
