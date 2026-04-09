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

func TestAccRestoreDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	backupID := os.Getenv("ARUBACLOUD_BACKUP_ID")
	restoreID := os.Getenv("ARUBACLOUD_RESTORE_ID")
	if projectID == "" || backupID == "" || restoreID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_BACKUP_ID and ARUBACLOUD_RESTORE_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreDataSourceConfig(projectID, backupID, restoreID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_restore.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_restore.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_restore.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_restore.test",
						tfjsonpath.New("backup_id"),
						knownvalue.StringExact(backupID),
					),
				},
			},
		},
	})
}

func testAccRestoreDataSourceConfig(projectID, backupID, restoreID string) string {
	return fmt.Sprintf(`
data "arubacloud_restore" "test" {
  id         = %[1]q
  project_id = %[2]q
  backup_id  = %[3]q
}
`, restoreID, projectID, backupID)
}
