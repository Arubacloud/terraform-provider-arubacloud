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
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	vpntunnelID := os.Getenv("ARUBACLOUD_VPNTUNNEL_ID")
	if vpntunnelID == "" {
		t.Skip("ARUBACLOUD_VPNTUNNEL_ID must be set for acceptance tests")
	}
	vpnrouteID := os.Getenv("ARUBACLOUD_VPNROUTE_ID")
	if vpnrouteID == "" {
		t.Skip("ARUBACLOUD_VPNROUTE_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnrouteDataSourceConfig(projectID, vpntunnelID, vpnrouteID),
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
						knownvalue.NotNull(),
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

func testAccVpnrouteDataSourceConfig(projectID, vpntunnelID, vpnrouteID string) string {
	return fmt.Sprintf(`
data "arubacloud_vpnroute" "test" {
  id            = %[3]q
  project_id    = %[1]q
  vpn_tunnel_id = %[2]q
}
`, projectID, vpntunnelID, vpnrouteID)
}
