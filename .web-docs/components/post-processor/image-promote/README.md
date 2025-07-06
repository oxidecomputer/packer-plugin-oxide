Type: `oxide-image-promote`
Artifact BuilderId: `oxide.image-promote`

The `oxide-image-promote` post-processor promotes an Oxide image built by the 
`oxide-instance` builder to make it available at the silo level or to a specific 
project. This is useful when you want to share custom images across multiple 
projects within your Oxide silo.

## Basic Example

The following example promotes an image to the silo level, making it available
to all projects within the silo:

```hcl
source "oxide-instance" "example" {
  source_image_name = "alpine-3.20"
  image_name        = "custom-alpine"
  image_description = "Customized Alpine Linux image"
  disk_size         = 20
}

build {
  sources = ["source.oxide-instance.example"]
  
  post-processor "oxide-image-promote" {
    # Image will be promoted to silo level
  }
}
```

## Promoting to a Specific Project

To promote an image to a specific project instead of the silo level:

```hcl
build {
  sources = ["source.oxide-instance.example"]
  
  post-processor "oxide-image-promote" {
    project = "production-project"
  }
}
```

## Configuration Reference

There are many configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters.

### Required

There are no required configuration options for this post-processor. The `host` and `token` fields can be provided via environment variables `OXIDE_HOST` and `OXIDE_TOKEN` respectively.

### Optional

<!-- Code generated from the comments of the Config struct in component/post-processor/image-promote/config.go; DO NOT EDIT MANUALLY -->

- `host` (string) - The Oxide API host. Defaults to the `OXIDE_HOST` environment variable.

- `token` (string) - The Oxide authentication token. Defaults to the `OXIDE_TOKEN` environment variable.

- `project` (string) - The name or ID of the project where the promoted image should be made available.
  If not specified, the image will be promoted to be available at the silo level.

<!-- End of code generated from the comments of the Config struct in component/post-processor/image-promote/config.go; -->


## Environment Variables

The post-processor respects the following environment variables:

- `OXIDE_HOST`: The Oxide API host URL
- `OXIDE_TOKEN`: The authentication token for the Oxide API

These environment variables provide default values if the corresponding 
configuration options are not specified.

## Output

The post-processor creates a new artifact containing:

- The promoted image ID
- The promoted image name
- The project name (if promoted to a specific project)

This artifact can be used by subsequent post-processors in your build chain.
