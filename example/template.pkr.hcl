packer {
  required_plugins {
    oxide = {
      version = ">= 0.6.1"
      source  = "github.com/oxidecomputer/oxide"
    }
  }
}

data "oxide-image" "ubuntu" {
  name = "noble"
}

source "oxide-instance" "example" {
  project            = "packer-acc-test"
  boot_disk_image_id = data.oxide-image.ubuntu.image_id

  # Do not connect to the temporary instance.
  communicator = "none"

  # Connect to the temporary instance using SSH.
  # communicator = "ssh"
  # ssh_username = "ubuntu"
}

build {
  sources = [
    "source.oxide-instance.example",
  ]
}
