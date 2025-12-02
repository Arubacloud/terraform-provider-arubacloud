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

func TestAccContainerregistryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerregistryResourceConfig("test-containerregistry"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-containerregistry"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_containerregistry.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccContainerregistryResourceConfig("test-containerregistry-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-containerregistry-updated"),
					),
				},
			},
		},
	})
}

func testAccContainerregistryResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_containerregistry" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check containerregistry_resource.go for required attributes
}
`, name)
}
