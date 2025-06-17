The [Oxide](https://oxide.computer) Packer plugin is a multi-component plugin
for building Oxide images.

## Installation

To install this plugin, add the following to your Packer configuration and run
`packer init`.

```hcl
packer {
  required_plugins {
    oxide = {
      source  = "github.com/oxidecomputer/oxide"
      version = ">= 1.0.0"
    }
  }
}
```

Alternatively, use `packer plugins install` to install this plugin.

```sh
packer plugins install github.com/oxidecomputer/oxide
```

## Components

### Builders

- [`instance`](/packer/integrations/oxidecomputer/oxide/latest/components/builder/instance) -
Builds an Oxide image by creating an instance from a specified image,
provisioning it, and snapshotting its boot disk into an image. The resulting
image can be used as the source image for new Oxide instances.

### Data Sources

- [`image`](/packer/integrations/oxidecomputer/oxide/latest/components/data-source/image) -
Fetches the image ID for an Oxide image using its name. The image can be a
project image or silo image.

<!-- ### Provisioners -->

<!-- ### Post-Processors -->
