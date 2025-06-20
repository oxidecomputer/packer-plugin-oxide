Type: `oxide-image`

<!-- Code generated from the comments of the Datasource struct in component/data-source/image/data_source.go; DO NOT EDIT MANUALLY -->

The Oxide `image` data source fetches the image ID for an Oxide image using
its name. The image can be a project image or silo image.

<!-- End of code generated from the comments of the Datasource struct in component/data-source/image/data_source.go; -->


## Configuration

<!-- Code generated from the comments of the Config struct in component/data-source/image/config.go; DO NOT EDIT MANUALLY -->

The configuration arguments for this data source component. Arguments can
either be required or optional.

<!-- End of code generated from the comments of the Config struct in component/data-source/image/config.go; -->


### Required

<!-- Code generated from the comments of the Config struct in component/data-source/image/config.go; DO NOT EDIT MANUALLY -->

- `host` (string) - Oxide API URL (e.g., `https://oxide.sys.example.com`). If not specified, this
  defaults to the value of the `OXIDE_HOST` environment variable.

- `token` (string) - Oxide API token. If not specified, this defaults to the value of the
  `OXIDE_TOKEN` environment variable.

- `name` (string) - Name of the image to fetch.

<!-- End of code generated from the comments of the Config struct in component/data-source/image/config.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in component/data-source/image/config.go; DO NOT EDIT MANUALLY -->

- `project` (string) - Name or ID of the project containing the image to fetch. Leave blank to fetch
  a silo image instead of a project image.

<!-- End of code generated from the comments of the Config struct in component/data-source/image/config.go; -->


## Outputs

<!-- Code generated from the comments of the DatasourceOutput struct in component/data-source/image/output.go; DO NOT EDIT MANUALLY -->

- `image_id` (string) - ID of the image that was fetched.

<!-- End of code generated from the comments of the DatasourceOutput struct in component/data-source/image/output.go; -->


## Examples

Fetch a project image.

```hcl
data "oxide-image" "example" {
  name    = "ubuntu"
  project = "oxide"
}
```

Fetch a silo image.

```hcl
data "oxide-image" "example" {
  name = "ubuntu"
}
```
