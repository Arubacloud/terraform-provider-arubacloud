#!/bin/bash
# Script to format Terraform documentation to separate Arguments from Attributes
# This script post-processes the generated documentation to make it clearer
# which fields are inputs (Arguments) and which are outputs (Attributes)

DOCS_DIR="${1:-docs}"

echo "Formatting Terraform provider documentation..."

# Function to process a markdown file
format_terraform_doc() {
    local file="$1"
    echo "  Processing: $(basename "$file")"
    
    # Use sed with extended regex to format the documentation
    # Create a temporary file
    local tmp_file="${file}.tmp"
    
    # Process the file
    awk '
    BEGIN { in_schema = 0; found_required = 0; found_optional = 0; found_readonly = 0 }
    
    # Detect Schema section
    /^## Schema$/ {
        print
        in_schema = 1
        next
    }
    
    # When we find Required after Schema
    in_schema && !found_required && /^### Required$/ {
        print ""
        print "### Arguments"
        print ""
        print "The following arguments are supported:"
        print ""
        print "#### Required"
        found_required = 1
        next
    }
    
    # Change Optional to subheading
    in_schema && found_required && /^### Optional$/ {
        print "#### Optional"
        found_optional = 1
        next
    }
    
    # Replace Read-Only with Attributes Reference
    in_schema && /^### Read-Only$/ {
        print "### Attributes Reference"
        print ""
        print "In addition to all arguments above, the following attributes are exported:"
        print ""
        print "#### Read-Only"
        found_readonly = 1
        next
    }
    
    # Print all other lines as-is
    { print }
    
    # Reset when we exit schema section (next ## heading)
    /^## [^S]/ { in_schema = 0 }
    ' "$file" > "$tmp_file"
    
    # Replace original with formatted version
    mv "$tmp_file" "$file"
}

# Process all resource documentation files
if [ -d "$DOCS_DIR/resources" ]; then
    for file in "$DOCS_DIR/resources"/*.md; do
        if [ -f "$file" ]; then
            format_terraform_doc "$file"
        fi
    done
fi

# Process all data source documentation files
if [ -d "$DOCS_DIR/data-sources" ]; then
    for file in "$DOCS_DIR/data-sources"/*.md; do
        if [ -f "$file" ]; then
            format_terraform_doc "$file"
        fi
    done
fi

echo "Documentation formatting complete!"
