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

func TestAccElasticipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccElasticipResourceConfig("test-elasticip"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-elasticip"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("address"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_elasticip.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccElasticipResourceConfig("test-elasticip-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-elasticip-updated"),
					),
				},
			},
		},
	})
}

func testAccElasticipResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_elasticip" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
}
`, name)
}
