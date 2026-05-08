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

func TestAccDatabasegrantDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	if projectID == "" || dbaasID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_DBAAS_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasegrantDataSourceConfig(projectID, dbaasID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.StringExact(dbaasID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("database"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("user_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("role"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDatabasegrantDataSourceConfig(projectID, dbaasID string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaasuser" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  username   = "test-ds-grantuser"
  password   = "TestPassword123!"
}

resource "arubacloud_database" "test" {
  name       = "testdsgrantdb"
  project_id = %[1]q
  dbaas_id   = %[2]q
}

resource "arubacloud_databasegrant" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  database   = arubacloud_database.test.name
  user_id    = arubacloud_dbaasuser.test.username
  role       = "read"
}

data "arubacloud_databasegrant" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  database   = arubacloud_database.test.name
  user_id    = arubacloud_dbaasuser.test.username
}
`, projectID, dbaasID)
}
