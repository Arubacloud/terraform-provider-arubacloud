package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpcpeeringrouteResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpcpeeringrouteDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpcpeeringrouteResourceConfig("test-vpcpeeringroute"),
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
			},
			// Update and Read testing
			{
				Config: testAccVpcpeeringrouteResourceConfig("test-vpcpeeringroute-updated"),
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
		vpcID := rs.Primary.Attributes["vpc_id"]
		peeringID := rs.Primary.Attributes["vpc_peering_id"]
		resp, err := client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, rs.Primary.Attributes["project_id"], vpcID, peeringID, rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Vpcpeeringroute", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("VPCPeeringRoute %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpcpeeringrouteResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpcpeeringroute" "test" {
  name                   = %[1]q
  project_id             = "test-project-id"
  vpc_id                 = "test-vpc-id"
  vpc_peering_id         = "test-peering-id"
  local_network_address  = "10.0.0.0/24"
  remote_network_address = "10.1.0.0/24"
  billing_period         = "Hour"
}
`, name)
}
