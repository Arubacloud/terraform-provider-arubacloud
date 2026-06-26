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
				ResourceName:            "arubacloud_dbaasuser.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc:       importIDFromAttrs("arubacloud_dbaasuser.test", "project_id", "dbaas_id", "id"),
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
		projectID := rs.Primary.Attributes["project_id"]
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + rs.Primary.ID)
		_, err = client.Client.FromDatabase().Users().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Dbaasuser", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
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
