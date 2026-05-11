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

func TestAccDatabasebackupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabasebackupDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabasebackupResourceConfig("test-databasebackup"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasebackup"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_databasebackup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDatabasebackupResourceConfig("test-databasebackup-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasebackup-updated"),
					),
				},
			},
		},
	})
}

func testCheckDatabasebackupDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_databasebackup" {
			continue
		}
		resp, err := client.Client.FromDatabase().Backups().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Databasebackup", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("DatabaseBackup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabasebackupResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_databasebackup" "test" {
  name           = %[1]q
  project_id     = "test-project-id"
  location       = "it-1"
  zone           = "it-1"
  dbaas_id       = "test-dbaas-id"
  database       = "test-db"
  billing_period = "Hour"
}
`, name)
}
