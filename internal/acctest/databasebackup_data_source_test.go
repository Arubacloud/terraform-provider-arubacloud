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
	databaseName := os.Getenv("ARUBACLOUD_DATABASE_NAME")
	if projectID == "" || dbaasID == "" || databaseName == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_DBAAS_ID and ARUBACLOUD_DATABASE_NAME must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasebackupDataSourceConfig(projectID, dbaasID, databaseName),
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

func testAccDatabasebackupDataSourceConfig(projectID, dbaasID, databaseName string) string {
	return fmt.Sprintf(`
resource "arubacloud_databasebackup" "test" {
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  zone           = "ITBG-1"
  dbaas_id       = %[2]q
  database       = %[3]q
  billing_period = "Hour"
  tags           = ["acceptance-test"]
}

data "arubacloud_databasebackup" "test" {
  id         = arubacloud_databasebackup.test.id
  project_id = %[1]q
}
`, projectID, dbaasID, databaseName)
}
