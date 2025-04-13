packer {
  required_plugins {
    oxide = {
      version = ">=v0.0.1"
      source  = "github.com/oxidecomputer/oxide"
    }
  }
}

source "oxide-instance" "example" {
  project     = "matthewsanabria"
  image_id    = "feb2c8ee-5a1d-4d66-beeb-289b860561bf"

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
