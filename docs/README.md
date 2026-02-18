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
      version = ">= 0.7.0"
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

[`oxide-instance`](/packer/integrations/oxidecomputer/oxide/latest/components/builder/instance)
@include 'component/builder/instance/Builder.mdx'

### Data Sources

[`oxide-image`](/packer/integrations/oxidecomputer/oxide/latest/components/data-source/image)
@include 'component/data-source/image/Datasource.mdx'

<!-- ### Provisioners -->

<!-- ### Post-Processors -->
