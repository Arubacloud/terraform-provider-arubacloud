package provider

import (
	"context"
	"fmt"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpntunnelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpntunnelDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpntunnelResourceConfig("test-vpntunnel"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpntunnel"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpntunnel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccVpntunnelResourceConfig("test-vpntunnel-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpntunnel.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpntunnel-updated"),
					),
				},
			},
		},
	})
}

func testCheckVpntunnelDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_vpntunnel" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.VPNTunnelRef(projectID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().VPNTunnels().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Vpntunnel", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("VPNTunnel %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpntunnelResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpntunnel" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"

  properties = {
    vpn_type = "Site-To-Site"
  }
}
`, name)
}
