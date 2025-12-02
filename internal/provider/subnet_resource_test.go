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

func TestAccSubnetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSubnetResourceConfig("test-subnet"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-subnet"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSubnetResourceConfig("test-subnet-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-subnet-updated"),
					),
				},
			},
		},
	})
}

func testAccSubnetResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_subnet" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check subnet_resource.go for required attributes
}
`, name)
}
