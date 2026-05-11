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

func TestAccDatabasegrantResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabasegrantDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabasegrantResourceConfig("test-databasegrant"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("database"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("role"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_databasegrant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDatabasegrantResourceConfig("test-databasegrant-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasegrant-updated"),
					),
				},
			},
		},
	})
}

func testCheckDatabasegrantDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_databasegrant" {
			continue
		}
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		database := rs.Primary.Attributes["database"]
		userID := rs.Primary.Attributes["user_id"]
		resp, err := client.Client.FromDatabase().Grants().Get(ctx, rs.Primary.Attributes["project_id"], dbaasID, database, userID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Databasegrant", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("DatabaseGrant %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabasegrantResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_databasegrant" "test" {
  project_id = "test-project-id"
  dbaas_id   = "test-dbaas-id"
  database   = %[1]q
  user_id    = "test-user"
  role       = "read"
}
`, name)
}
