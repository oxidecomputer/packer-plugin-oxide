Type: `oxide-instance`

<!-- Code generated from the comments of the Builder struct in component/builder/instance/builder.go; DO NOT EDIT MANUALLY -->

The `oxide-instance` builder creates custom images for use with
[Oxide](https://oxide.computer). The builder launches a temporary instance
from an existing source image, connects to the instance using its external
IP, provisions the instance, and then creates a new image from the instance's
boot disk. The resulting image can be used to launch new instances on Oxide.

The builder does not manage images. Once it creates an image, it is up to you
to use it or delete it.

<!-- End of code generated from the comments of the Builder struct in component/builder/instance/builder.go; -->


## Configuration

<!-- Code generated from the comments of the Config struct in component/builder/instance/config.go; DO NOT EDIT MANUALLY -->

The configuration arguments for this builder component. Arguments can
either be required or optional.

<!-- End of code generated from the comments of the Config struct in component/builder/instance/config.go; -->


### Required

<!-- Code generated from the comments of the Config struct in component/builder/instance/config.go; DO NOT EDIT MANUALLY -->

- `host` (string) - Oxide API URL (e.g., `https://oxide.sys.example.com`). If not specified, this
  defaults to the value of the `OXIDE_HOST` environment variable.

- `token` (string) - Oxide API token. If not specified, this defaults to the value of the
  `OXIDE_TOKEN` environment variable.

- `boot_disk_image_id` (string) - Image ID to use for the instance's boot disk. This can be obtained from the
  `oxide-image` data source.

- `project` (string) - Name or ID of the project where the temporary instance and resulting image
  will be created.

<!-- End of code generated from the comments of the Config struct in component/builder/instance/config.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in component/builder/instance/config.go; DO NOT EDIT MANUALLY -->

- `boot_disk_size` (uint64) - Size of the boot disk in bytes. Defaults to `21474836480`, or 20 GiB.

- `ip_pool` (string) - IP pool to allocate the instance's external IP from. If not specified, the
  silo's default IP pool will be used.

- `vpc` (string) - VPC to create the instance within. Defaults to `default`.

- `subnet` (string) - Subnet to create the instance within. Defaults to `default`.

- `name` (string) - Name of the temporary instance. Defaults to `packer-{{timestamp}}`.

- `hostname` (string) - Hostname of the temporary instance. Defaults to `packer-{{timestamp}}`.

- `cpus` (uint64) - Number of vCPUs to provision the instance with. Defaults to `1`.

- `memory` (uint64) - Amount of memory to provision the instance with, in bytes. Defaults to
  `2147483648`, or 2 GiB.

- `ssh_public_keys` ([]string) - An array of names or IDs of SSH public keys to inject into the instance.

- `artifact_name` (string) - Name of the resulting image artifact. Defaults to `packer-{{timestamp}}`.

<!-- End of code generated from the comments of the Config struct in component/builder/instance/config.go; -->


## Communicator

A [`communicator`](/docs/communicators) can be configured for the builder.

### SSH

The builder automatically generates a temporary SSH key pair that's used to
connect to the temporary instance unless one of the following SSH communicator
attributes are set.

- [`ssh_password`](/docs/communicators/ssh#ssh_password)
- [`ssh_private_key_file`](/docs/communicators/ssh#ssh_private_key_file)
- [`ssh_agent_auth`](/docs/communicators/ssh#ssh_agent_auth)

The temporary SSH public key is uploaded to Oxide, injected into the emphemeral
instance, and deleted during cleanup. The temporary SSH private key is used by
Packer to connect to the instance.

The name of the temporary SSH public key uploaded to Oxide can be set using the
[`temporary_key_pair_name`](/docs/communicators/ssh#temporary_key_pair_name)
attribute. Generally there's no reason to set this but it's available should it
be necessary.

## Provisioner

A [`provisioner`](/docs/provisioners) can be configured for the builder.

### File System Consistency

In [oxidecomputer/packer-plugin-oxide#33](https://github.com/oxidecomputer/packer-plugin-oxide/issues/33) 
it was reported that images built by the builder were missing files uploaded
during provisioning. To prevent missing files add a provisioner that forces the
instance to flush any buffered data to its storage devices. This provisioner
must be defined after all other provisioners that modify the file system.

A provisioner that reboots the instance will work. Packer will attempt to
connect to the instance once it's back up.

```hcl
# Linux. 
provisioner "shell" {
  # Required so Packer doesn't error when the SSH connection terminates.
  expect_disconnect = true
  inline = [
    "sudo reboot",
  ]
}

# Windows.
provisioner "windows-restart" {}
```

A provisioner that explictly forces a file synchronize will work.

```hcl
# Linux. 
provisioner "shell" {
  inline = [
    "sudo sync",
  ]
}

# Windows.
provisioner "powershell" {
  inline = [
    "Get-Volume | Write-VolumeCache"
  ]
}
```

## Examples

This example uses environment variables for Oxide credentials and uses
the automatically generated SSH key pair to connect to the instance for
provisioning. The provisioners shown demonstrate common patterns when working
with the builder.

```hcl
source "oxide-instance" "example" {
  project            = "packer-acc-test"
  boot_disk_image_id = "feb2c8ee-5a1d-4d66-beeb-289b860561bf"

  # SSH communicator configuration.
  ssh_username = "ubuntu"
}

build {
  sources = [
    "source.oxide-instance.example",
  ]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y ca-certificates",
    ]
  }

  # Packer recommends using a 2-step workflow for uploading files.

  # 1) Upload the file to a location the provisioning user has access to.
  provisioner "file" {
    content     = "Hello from Packer!"
    destination = "/tmp/hello.txt"
  }

  # 2) Use the shell provisioner to move the files and set permissions.
  provisioner "shell" {
    expect_disconnect = true
    inline = [
      "sudo cp /tmp/hello.txt /opt/hello.txt",
      "sudo chmod 0644 /opt/hello.txt",
    ]
  }

  # Reboot the instance to flush any buffered data to its storage devices.
  #
  # https://github.com/oxidecomputer/packer-plugin-oxide/issues/33
  provisioner "shell" {
    # Required so Packer doesn't error when the SSH connection terminates.
    expect_disconnect = true
    inline = [
      "sudo reboot",
    ]
  }
}
```
