#!/bin/bash
# Script to generate test file templates for resources that don't have tests yet
# Usage: ./generate-test-stubs.sh

PROVIDER_DIR="internal/provider"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Scanning for resources without test files..."

# Find all resource files
for resource_file in "$PROVIDER_DIR"/*_resource.go; do
    # Skip if file doesn't exist
    [ -e "$resource_file" ] || continue
    
    # Extract base name (e.g., backup from backup_resource.go)
    base_name=$(basename "$resource_file" _resource.go)
    test_file="$PROVIDER_DIR/${base_name}_resource_test.go"
    
    # Check if test file exists
    if [ ! -f "$test_file" ]; then
        echo -e "${YELLOW}Missing test for: ${base_name}_resource${NC}"
        
        # Extract resource type from the file
        resource_type=$(grep -oP 'resp.TypeName = req.ProviderTypeName \+ "_\K\w+' "$resource_file" | head -1)
        
        if [ -n "$resource_type" ]; then
            echo "  Creating stub test file: $test_file"
            
            # Generate test file from template
            cat > "$test_file" << EOF
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAcc${base_name^}Resource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAcc${base_name^}ResourceConfig("test-${base_name}"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_${resource_type}.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-${base_name}"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_${resource_type}.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_${resource_type}.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAcc${base_name^}ResourceConfig("test-${base_name}-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_${resource_type}.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-${base_name}-updated"),
					),
				},
			},
		},
	})
}

func testAcc${base_name^}ResourceConfig(name string) string {
	return fmt.Sprintf(\`
resource "arubacloud_${resource_type}" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check ${base_name}_resource.go for required attributes
}
\`, name)
}
EOF
            echo -e "${GREEN}  ✓ Created test stub${NC}"
        else
            echo "  ⚠ Could not determine resource type"
        fi
    fi
done

# Do the same for data sources
echo ""
echo "Scanning for data sources without test files..."

for datasource_file in "$PROVIDER_DIR"/*_data_source.go; do
    # Skip if file doesn't exist
    [ -e "$datasource_file" ] || continue
    
    # Extract base name
    base_name=$(basename "$datasource_file" _data_source.go)
    test_file="$PROVIDER_DIR/${base_name}_data_source_test.go"
    
    # Check if test file exists
    if [ ! -f "$test_file" ]; then
        echo -e "${YELLOW}Missing test for: ${base_name}_data_source${NC}"
        
        # Extract resource type from the file
        resource_type=$(grep -oP 'resp.TypeName = req.ProviderTypeName \+ "_\K\w+' "$datasource_file" | head -1)
        
        if [ -n "$resource_type" ]; then
            echo "  Creating stub test file: $test_file"
            
            cat > "$test_file" << EOF
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAcc${base_name^}DataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAcc${base_name^}DataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_${resource_type}.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_${resource_type}.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("example-${base_name}"),
					),
				},
			},
		},
	})
}

const testAcc${base_name^}DataSourceConfig = \`
data "arubacloud_${resource_type}" "test" {
  name = "example-${base_name}"
  # TODO: Add required fields based on the schema
  # Check ${base_name}_data_source.go for required attributes
}
\`
EOF
            echo -e "${GREEN}  ✓ Created test stub${NC}"
        else
            echo "  ⚠ Could not determine resource type"
        fi
    fi
done

echo ""
echo -e "${GREEN}Done! Remember to:${NC}"
echo "  1. Fill in the required fields in the generated config functions"
echo "  2. Add additional test scenarios for optional fields"
echo "  3. Run 'go test ./internal/provider/ -v' to verify tests compile"
echo "  4. Run 'TF_ACC=1 go test ./internal/provider/ -v' to run acceptance tests"
