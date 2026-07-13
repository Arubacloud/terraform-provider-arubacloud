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

func TestAccDatabasebackupResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	databaseName := os.Getenv("ARUBACLOUD_DATABASE_NAME")
	if projectID == "" || dbaasID == "" || databaseName == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_DBAAS_ID and ARUBACLOUD_DATABASE_NAME must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabasebackupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasebackupResourceConfig(projectID, dbaasID, databaseName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
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
			{
				ResourceName:      "arubacloud_databasebackup.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_databasebackup.test", "project_id", "id"),
			},
			{
				// Backup Update is a no-op (API has no update endpoint); verify
				// the resource stays stable on a second apply.
				Config: testAccDatabasebackupResourceConfig(projectID, dbaasID, databaseName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testCheckDatabasebackupDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_databasebackup" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/backups/" + rs.Primary.ID)
		_, err = client.Client.FromDatabase().Backups().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Databasebackup", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("DatabaseBackup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabasebackupResourceConfig(projectID, dbaasID, databaseName string) string {
	return fmt.Sprintf(`
resource "arubacloud_databasebackup" "test" {
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  zone           = "ITBG-1"
  dbaas_id       = %[2]q
  database       = %[3]q
  billing_period = "Hour"
}
`, projectID, dbaasID, databaseName)
}
