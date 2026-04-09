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

func TestAccVpnrouteDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	tunnelID := os.Getenv("ARUBACLOUD_VPNTUNNEL_ID")
	routeID := os.Getenv("ARUBACLOUD_VPNROUTE_ID")
	if projectID == "" || tunnelID == "" || routeID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_VPNTUNNEL_ID and ARUBACLOUD_VPNROUTE_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnrouteDataSourceConfig(projectID, tunnelID, routeID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpnroute.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpnroute.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpnroute.test",
						tfjsonpath.New("vpn_tunnel_id"),
						knownvalue.StringExact(tunnelID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpnroute.test",
						tfjsonpath.New("destination"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpnroute.test",
						tfjsonpath.New("gateway"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccVpnrouteDataSourceConfig(projectID, tunnelID, routeID string) string {
	return fmt.Sprintf(`
data "arubacloud_vpnroute" "test" {
  id             = %[1]q
  project_id     = %[2]q
  vpn_tunnel_id  = %[3]q
}
`, routeID, projectID, tunnelID)
}
