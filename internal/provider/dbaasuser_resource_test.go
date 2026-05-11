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

func TestAccDbaasuserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDbaasuserDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbaasuserResourceConfig("test-dbaasuser"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("username"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_dbaasuser.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDbaasuserResourceConfig("test-dbaasuser-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-dbaasuser-updated"),
					),
				},
			},
		},
	})
}

func testCheckDbaasuserDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_dbaasuser" {
			continue
		}
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		resp, err := client.Client.FromDatabase().Users().Get(ctx, rs.Primary.Attributes["project_id"], dbaasID, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Dbaasuser", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("DBaaSUser %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDbaasuserResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaasuser" "test" {
  project_id = "test-project-id"
  dbaas_id   = "test-dbaas-id"
  username   = %[1]q
  password   = "test-password-123"
}
`, name)
}
