package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpcpeeringrouteDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcpeeringrouteDataSourceConfig(projectID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("vpc_peering_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccVpcpeeringrouteDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "local" {
  name       = "test-ds-route-vpc-local"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_vpc" "peer" {
  name       = "test-ds-route-vpc-peer"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_vpcpeering" "test" {
  name       = "test-ds-route-vpcpeering"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.local.id
  peer_vpc   = arubacloud_vpc.peer.id
}

resource "arubacloud_vpcpeeringroute" "test" {
  name                   = "test-ds-vpcpeeringroute"
  location               = "ITBG-Bergamo"
  project_id             = %[1]q
  vpc_id                 = arubacloud_vpc.local.id
  vpc_peering_id         = arubacloud_vpcpeering.test.id
  local_network_address  = "10.0.0.0/24"
  remote_network_address = "10.1.0.0/24"
  billing_period         = "Hour"
}

data "arubacloud_vpcpeeringroute" "test" {
  id             = arubacloud_vpcpeeringroute.test.id
  project_id     = %[1]q
  vpc_id         = arubacloud_vpc.local.id
  vpc_peering_id = arubacloud_vpcpeering.test.id
}
`, projectID)
}
