package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpcResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpcResourceConfig("test-vpc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpc"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpc.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccVpcResourceConfig("test-vpc-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpc-updated"),
					),
				},
			},
		},
	})
}

func testAccVpcResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
}
`, name)
}
