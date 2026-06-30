package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpcpeeringrouteResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpcpeeringrouteDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpcpeeringrouteResourceConfig(projectID, "test-vpcpeeringroute"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpcpeeringroute"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("vpc_peering_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpcpeeringroute.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_vpcpeeringroute.test", "project_id", "vpc_id", "vpc_peering_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccVpcpeeringrouteResourceConfig(projectID, "test-vpcpeeringroute-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpcpeeringroute-updated"),
					),
				},
			},
		},
	})
}

func testCheckVpcpeeringrouteDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_vpcpeeringroute" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		vpcID := rs.Primary.Attributes["vpc_id"]
		peeringID := rs.Primary.Attributes["vpc_peering_id"]
		ref := aruba.VPCPeeringRouteRef(projectID, vpcID, peeringID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Vpcpeeringroute", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("VPCPeeringRoute %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpcpeeringrouteResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "prtroute_local" {
  name       = "test-acc-prtroute-local"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_vpc" "prtroute_peer" {
  name       = "test-acc-prtroute-peer"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_vpcpeering" "prtroute_prereq" {
  name       = "test-acc-prtroute-peering"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.prtroute_local.id
  peer_vpc   = arubacloud_vpc.prtroute_peer.id
}

resource "arubacloud_vpcpeeringroute" "test" {
  name                   = %[2]q
  project_id             = %[1]q
  vpc_id                 = arubacloud_vpc.prtroute_local.id
  vpc_peering_id         = arubacloud_vpcpeering.prtroute_prereq.id
  local_network_address  = "10.0.0.0/24"
  remote_network_address = "10.1.0.0/24"
  billing_period         = "Hour"
}
`, projectID, name)
}
