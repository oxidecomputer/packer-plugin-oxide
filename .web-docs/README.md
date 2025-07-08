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
      version = ">= 0.3.0"
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
<!-- Code generated from the comments of the Builder struct in component/builder/instance/builder.go; DO NOT EDIT MANUALLY -->

The `oxide-instance` builder creates custom images for use with
[Oxide](https://oxide.computer). The builder launches a temporary instance
from an existing source image, connects to the instance using its external
IP, provisions the instance, and then creates a new image from the instance's
boot disk. The resulting image can be used to launch new instances on Oxide.

The builder does not manage images. Once it creates an image, it is up to you
to use it or delete it.

<!-- End of code generated from the comments of the Builder struct in component/builder/instance/builder.go; -->


### Data Sources

[`oxide-image`](/packer/integrations/oxidecomputer/oxide/latest/components/data-source/image)
<!-- Code generated from the comments of the Datasource struct in component/data-source/image/data_source.go; DO NOT EDIT MANUALLY -->

The `oxide-image` data source fetches [Oxide](https://oxide.computer) image
information for use in a Packer build. The image can be a project image or
silo image.

<!-- End of code generated from the comments of the Datasource struct in component/data-source/image/data_source.go; -->


<!-- ### Provisioners -->

<!-- ### Post-Processors -->
