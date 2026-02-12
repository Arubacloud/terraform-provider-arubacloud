provider {
  # (Optional) Path to the generated docs directory
  docs_directory = "docs"

  # (Optional) If true, generated files replace existing ones
  generate_readmes = true
}

# Configure how resources and data sources are documented
resource {
  # Leave header empty so the template's frontmatter (with subcategory) is the only
  # one; the Registry uses the first YAML frontmatter block for nav, so we must not
  # prepend a block without subcategory.
  header = ""

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
  # Leave header empty so the template's frontmatter (with subcategory) is the only one.
  header = ""

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
