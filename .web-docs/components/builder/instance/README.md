Type: `oxide-instance`

<!-- Code generated from the comments of the Builder struct in component/builder/instance/builder.go; DO NOT EDIT MANUALLY -->

This builder creates custom images on Oxide. The builder launches a temporary
instance, connects to it using its external IP, provisions it, and then
creates an image from the instance's boot disk. The resulting image can be
used to launch new instances.

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


## Examples

Basic build using environment variables for Oxide credentials.

```hcl
source "oxide-instance" "example" {
  project            = "oxide"
  boot_disk_image_id = "feb2c8ee-5a1d-4d66-beeb-289b860561bf"

  ssh_public_keys = [
    "529885a0-2919-463a-a588-ac48f100a165",
  ]

  ssh_username   = "ubuntu"
}

build {
  sources = [
    "source.oxide-instance.example",
  ]

  provisioner "shell" {
    inline = [
      "echo 'Hello from Packer!'",
    ]
  }
}
```
