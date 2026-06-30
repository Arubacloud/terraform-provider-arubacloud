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

func TestAccVpcpeeringResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpcpeeringDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpcpeeringResourceConfig(projectID, "test-vpcpeering"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeering.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpcpeering"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeering.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeering.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpcpeering.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_vpcpeering.test", "project_id", "vpc_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccVpcpeeringResourceConfig(projectID, "test-vpcpeering-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpcpeering.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpcpeering-updated"),
					),
				},
			},
		},
	})
}

func testCheckVpcpeeringDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_vpcpeering" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		vpcID := rs.Primary.Attributes["vpc_id"]
		ref := aruba.VPCPeeringRef(projectID, vpcID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().VPCPeerings().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Vpcpeering", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("VPCPeering %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpcpeeringResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "peering_local" {
  name       = "test-acc-peering-local-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_vpc" "peering_peer" {
  name       = "test-acc-peering-peer-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_vpcpeering" "test" {
  name       = %[2]q
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.peering_local.id
  peer_vpc   = arubacloud_vpc.peering_peer.id
}
`, projectID, name)
}
