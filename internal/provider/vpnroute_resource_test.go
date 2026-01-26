package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpnrouteResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpnrouteResourceConfig("test-vpnroute"),
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
			},
			// Update and Read testing
			{
				Config: testAccVpnrouteResourceConfig("test-vpnroute-updated"),
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

func testAccVpnrouteResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpnroute" "test" {
  name           = %[1]q
  location       = "it-1"
  project_id     = "test-project-id"
  vpn_tunnel_id  = "test-tunnel-id"
  
  properties = {
    cloud_subnet   = "10.0.0.0/24"
    on_prem_subnet = "192.168.0.0/24"
  }
}
`, name)
}
