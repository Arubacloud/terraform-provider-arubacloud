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

func TestAccVpnrouteResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	vpnTunnelID := os.Getenv("ARUBACLOUD_VPNTUNNEL_ID")
	if projectID == "" || vpnTunnelID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_VPNTUNNEL_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpnrouteDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpnrouteResourceConfig(projectID, vpnTunnelID, "test-vpnroute"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpnroute.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpnroute"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpnroute.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpnroute.test",
						tfjsonpath.New("vpn_tunnel_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpnroute.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_vpnroute.test", "project_id", "vpn_tunnel_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccVpnrouteResourceConfig(projectID, vpnTunnelID, "test-vpnroute-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpnroute.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpnroute-updated"),
					),
				},
			},
		},
	})
}

func testCheckVpnrouteDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_vpnroute" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		tunnelID := rs.Primary.Attributes["vpn_tunnel_id"]
		ref := aruba.VPNRouteRef(projectID, tunnelID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().VPNRoutes().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Vpnroute", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("VPNRoute %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpnrouteResourceConfig(projectID, vpnTunnelID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpnroute" "test" {
  name          = %[3]q
  location      = "ITBG-Bergamo"
  project_id    = %[1]q
  vpn_tunnel_id = %[2]q

  properties = {
    cloud_subnet   = "10.0.0.0/24"
    on_prem_subnet = "192.168.0.0/24"
  }
}
`, projectID, vpnTunnelID, name)
}
