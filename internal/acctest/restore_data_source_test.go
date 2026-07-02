package acctest

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
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreDataSourceConfig(projectID),
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
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccRestoreDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "source" {
  name           = "test-ds-restore-source"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 10
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_blockstorage" "target" {
  name           = "test-ds-restore-target"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 10
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_backup" "test" {
  name           = "test-ds-restore-backup"
  location       = "ITBG-Bergamo"
  project_id     = %[1]q
  type           = "Full"
  volume_id      = arubacloud_blockstorage.source.id
  billing_period = "Hour"
  retention_days = 7
  tags           = ["acceptance-test"]
}

resource "arubacloud_restore" "test" {
  name       = "test-ds-restore"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  backup_id  = arubacloud_backup.test.id
  volume_id  = arubacloud_blockstorage.target.id
}

data "arubacloud_restore" "test" {
  id         = arubacloud_restore.test.id
  project_id = %[1]q
  backup_id  = arubacloud_backup.test.id
}
`, projectID)
}
