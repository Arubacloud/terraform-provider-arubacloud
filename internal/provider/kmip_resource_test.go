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

func TestAccKmipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKmipResourceConfig("test-kmip"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kmip.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kmip"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kmip.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_kmip.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKmipResourceConfig("test-kmip-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kmip.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kmip-updated"),
					),
				},
			},
		},
	})
}

func testAccKmipResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_kmip" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check kmip_resource.go for required attributes
}
`, name)
}
