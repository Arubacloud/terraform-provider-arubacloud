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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnrouteDataSourceConfig(projectID),
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

func testAccVpnrouteDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpntunnel" "test" {
  name       = "test-ds-vpntunnel"
  location   = "ITBG-Bergamo"
  project_id = %[1]q

  properties = {
    vpn_type = "Site-To-Site"
  }
}

resource "arubacloud_vpnroute" "test" {
  name          = "test-ds-vpnroute"
  location      = "ITBG-Bergamo"
  project_id    = %[1]q
  vpn_tunnel_id = arubacloud_vpntunnel.test.id

  properties = {
    cloud_subnet   = "10.0.0.0/24"
    on_prem_subnet = "192.168.0.0/24"
  }
}

data "arubacloud_vpnroute" "test" {
  id            = arubacloud_vpnroute.test.id
  project_id    = %[1]q
  vpn_tunnel_id = arubacloud_vpntunnel.test.id
}
`, projectID)
}
