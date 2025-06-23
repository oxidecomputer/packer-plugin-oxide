packer {
  required_plugins {
    oxide = {
      version = ">= 0.1.0"
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

  # To build the image without connecting to it during tests.
  communicator = "none"
}

build {
  sources = [
    "source.oxide-instance.example",
  ]
}
