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

func TestAccVpntunnelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpntunnelResourceConfig("test-vpntunnel"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpntunnel"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpntunnel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccVpntunnelResourceConfig("test-vpntunnel-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpntunnel-updated"),
					),
				},
			},
		},
	})
}

func testAccVpntunnelResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpntunnel" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
  
  properties = {
    vpn_type = "Site-To-Site"
  }
}
`, name)
}
