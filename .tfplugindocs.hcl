provider {
  # (Optional) Path to the generated docs directory
  docs_directory = "docs"

  # (Optional) If true, generated files replace existing ones
  generate_readmes = true
}

# Configure how resources and data sources are documented
resource {
  # Common header added to all resource docs
  header = <<EOF
---
page_title: "%s Resource - Terraform Provider"
subcategory: "Resources"
description: |-
  %s Terraform resource.
---

EOF

  # Template for resource examples written at bottom of generated docs
  # You can remove this block entirely if you don’t want examples included
  example = <<EOF
# Example: %s
resource "%s" "%s" {
  # TODO: add example configuration
}
EOF
}

data_source {
  # Common header added to all data-source docs
  header = <<EOF
---
page_title: "%s Data Source - Terraform Provider"
subcategory: "Data Sources"
description: |-
  %s Terraform data source.
---

EOF

  example = <<EOF
# Example: %s
data "%s" "%s" {
  # TODO: add example configuration
}
EOF
}

# Optional: files to ignore (useful if manually maintaining some docs)
ignore_paths = [
  "README.md",
  "templates/**",
]
