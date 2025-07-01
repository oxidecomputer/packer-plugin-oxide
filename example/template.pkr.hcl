packer {
  required_plugins {
    oxide = {
      version = ">= 0.2.0"
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

  # Do not wait for SSH to connect. Useful in automated tests.
  communicator = "none"

  # Wait for SSH to connect. Useful in manual tests.
  # communicator = "ssh"
  # ssh_username = "ubuntu"
}

build {
  sources = [
    "source.oxide-instance.example",
  ]
}
