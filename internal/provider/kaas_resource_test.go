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

func TestAccKaasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKaasResourceConfig("test-kaas"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kaas"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_kaas.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKaasResourceConfig("test-kaas-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kaas-updated"),
					),
				},
			},
		},
	})
}

func testAccKaasResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_kaas" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check kaas_resource.go for required attributes
}
`, name)
}
