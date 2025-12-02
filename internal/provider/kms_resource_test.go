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

func TestAccKmsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKmsResourceConfig("test-kms"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kms"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kms.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_kms.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKmsResourceConfig("test-kms-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kms-updated"),
					),
				},
			},
		},
	})
}

func testAccKmsResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_kms" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check kms_resource.go for required attributes
}
`, name)
}
