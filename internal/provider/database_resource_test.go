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

func TestAccDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabaseDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseResourceConfig("test-database"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_database.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-database"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_database.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_database.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDatabaseResourceConfig("test-database-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_database.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-database-updated"),
					),
				},
			},
		},
	})
}

func testCheckDatabaseDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_database" {
			continue
		}
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		resp, err := client.Client.FromDatabase().Databases().Get(ctx, rs.Primary.Attributes["project_id"], dbaasID, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Database", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("Database %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabaseResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_database" "test" {
  name       = %[1]q
  project_id = "test-project-id"
  dbaas_id   = "test-dbaas-id"
}
`, name)
}
