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

func TestAccCloudserverResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCloudserverResourceConfig("test-cloudserver"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-cloudserver"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_cloudserver.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCloudserverResourceConfig("test-cloudserver-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-cloudserver-updated"),
					),
				},
			},
		},
	})
}

func testAccCloudserverResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_cloudserver" "test" {
  name = %[1]q
  # TODO: Add required fields based on the schema
  # Check cloudserver_resource.go for required attributes
}
`, name)
}
