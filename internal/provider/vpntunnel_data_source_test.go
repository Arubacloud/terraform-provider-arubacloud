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

func TestAccVpntunnelDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	vpntunnelID := os.Getenv("ARUBACLOUD_VPNTUNNEL_ID")
	if vpntunnelID == "" {
		t.Skip("ARUBACLOUD_VPNTUNNEL_ID must be set for acceptance tests (VPN tunnel requires public IP, subnet CIDR and full vpnClientSettings — use a pre-provisioned fixture)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpntunnelDataSourceConfig(projectID, vpntunnelID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpntunnel.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpntunnel.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpntunnel.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
				},
			},
		},
	})
}

func testAccVpntunnelDataSourceConfig(projectID, vpntunnelID string) string {
	return fmt.Sprintf(`
data "arubacloud_vpntunnel" "test" {
  id         = %[2]q
  project_id = %[1]q
}
`, projectID, vpntunnelID)
}
