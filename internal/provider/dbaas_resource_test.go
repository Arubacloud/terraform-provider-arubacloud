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

func TestAccDbaasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbaasResourceConfig("test-dbaas"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-dbaas"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_dbaas.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDbaasResourceConfig("test-dbaas-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-dbaas-updated"),
					),
				},
			},
		},
	})
}

func testAccDbaasResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaas" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check dbaas_resource.go for required attributes
}
`, name)
}
