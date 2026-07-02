package acctest

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/terraform-provider-arubacloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVpntunnelResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpntunnelDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpntunnelResourceConfig(projectID, "test-vpntunnel"),
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
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_vpntunnel.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccVpntunnelResourceConfig(projectID, "test-vpntunnel-updated"),
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
	client, err := AccClient()
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
		if provErr := provider.CheckResponseErr("get", "Vpntunnel", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("VPNTunnel %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpntunnelResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpntunnel" "test" {
  name       = %[2]q
  location   = "ITBG-Bergamo"
  project_id = %[1]q

  properties = {
    vpn_type = "Site-To-Site"
  }
}
`, projectID, name)
}
