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

func TestAccDatabasebackupDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	if projectID == "" || dbaasID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_DBAAS_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasebackupDataSourceConfig(projectID, dbaasID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("tags"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDatabasebackupDataSourceConfig(projectID, dbaasID string) string {
	return fmt.Sprintf(`
resource "arubacloud_database" "test" {
  name       = "testaccbackupds"
  project_id = %[1]q
  dbaas_id   = %[2]q
  timeout    = "15m"
}

resource "arubacloud_databasebackup" "test" {
  name           = "test-ds-databasebackup"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  zone           = "ITBG-1"
  dbaas_id       = %[2]q
  database       = arubacloud_database.test.name
  billing_period = "Hour"
  tags           = ["acceptance-test"]
}

data "arubacloud_databasebackup" "test" {
  id         = arubacloud_databasebackup.test.id
  project_id = %[1]q
}
`, projectID, dbaasID)
}
