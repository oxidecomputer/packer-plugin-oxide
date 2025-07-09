# See https://github.com/hashicorp/integration-template for schema.
integration {
  name        = "Oxide"
  description = "The Oxide multi-component plugin can be used with HashiCorp (IBM) Packer to create custom Oxide images."
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
    slug = "instance"
  }
  component {
    type = "data-source"
    name = "Oxide Image"
    slug = "image"
  }
}
