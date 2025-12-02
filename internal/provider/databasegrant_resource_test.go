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

func TestAccDatabasegrantResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabasegrantResourceConfig("test-databasegrant"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasegrant"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_databasegrant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDatabasegrantResourceConfig("test-databasegrant-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasegrant-updated"),
					),
				},
			},
		},
	})
}

func testAccDatabasegrantResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_databasegrant" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check databasegrant_resource.go for required attributes
}
`, name)
}
