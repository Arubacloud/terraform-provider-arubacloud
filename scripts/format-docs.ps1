# Script to format Terraform documentation to separate Arguments from Attributes
# This script post-processes the generated documentation to make it clearer
# which fields are inputs (Arguments) and which are outputs (Attributes)

param(
    [string]$DocsDir = "docs"
)

Write-Host "Formatting Terraform provider documentation..."

# Function to process a markdown file
function Format-TerraformDoc {
    param([string]$FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    
    # Pattern for resources: Convert Schema sections to Arguments/Attributes
    # Replace "## Schema" followed by "### Required" with Arguments section
    $content = $content -replace '(?m)^## Schema\s*\n\s*\n### Required', @"
## Schema

### Arguments

The following arguments are supported:

#### Required
"@
    
    # Change "### Optional" to "#### Optional" when it follows Required
    $content = $content -replace '(?m)^### Optional(?=\s*\n\s*\n-)', '#### Optional'
    
    # Replace "### Read-Only" with Attributes Reference section
    $content = $content -replace '(?m)^### Read-Only', @"
### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only
"@
    
    # Save the modified content back
    Set-Content -Path $FilePath -Value $content -NoNewline
}

# Process all resource documentation files
$resourceFiles = Get-ChildItem -Path "$DocsDir\resources" -Filter "*.md" -ErrorAction SilentlyContinue
foreach ($file in $resourceFiles) {
    Write-Host "  Processing: $($file.Name)"
    Format-TerraformDoc -FilePath $file.FullName
}

# Process all data source documentation files
$dataSourceFiles = Get-ChildItem -Path "$DocsDir\data-sources" -Filter "*.md" -ErrorAction SilentlyContinue
foreach ($file in $dataSourceFiles) {
    Write-Host "  Processing: $($file.Name)"
    Format-TerraformDoc -FilePath $file.FullName
}

Write-Host "Documentation formatting complete!"
