# See https://github.com/hashicorp/integration-template for schema.
integration {
  name        = "Oxide"
  description = "The Oxide Packer plugin is a multi-component plugin for building Oxide images."
  identifier  = "packer/oxidecomputer/oxide"
  flags       = []
  docs {
    process_docs    = true
    readme_location = "./README.md"
    external_url    = "https://github.com/oxidecomputer/packer-plugin-oxide"
  }
  license {
    type = "MPL-2.0"
    url  = "https://github.com/oxidecomputer/packer-plugin-oxide/blob/main/LICENSE"
  }
  component {
    type = "builder"
    name = "Oxide Instance"
    slug = "oxide-instance"
  }
  component {
    type = "data-source"
    name = "Oxide Image"
    slug = "oxide-image"
  }
}
