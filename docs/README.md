The [Oxide](https://oxide.computer) multi-component plugin can be used
with HashiCorp (IBM) Packer to create custom Oxide images.

## Installation

To install this plugin, add the following to your Packer configuration and run
`packer init`.

```hcl
packer {
  required_plugins {
    oxide = {
      source  = "github.com/oxidecomputer/oxide"
      version = ">= 0.1.0"
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
This builder creates custom images on Oxide. The builder launches a temporary
instance, connects to it using its external IP, provisions it, and then
creates an image from the instance's boot disk. The resulting image can be
used to launch new instances.

### Data Sources

- [`image`](/packer/integrations/oxidecomputer/oxide/latest/components/data-source/image) -
This data source fetches the image ID for an Oxide image using its name. The
image can be a project image or silo image.

<!-- ### Provisioners -->

<!-- ### Post-Processors -->
