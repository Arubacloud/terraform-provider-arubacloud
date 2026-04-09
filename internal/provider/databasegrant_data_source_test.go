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
	database := os.Getenv("ARUBACLOUD_DATABASE_ID")
	userID := os.Getenv("ARUBACLOUD_DBAAS_USER_ID")
	if projectID == "" || dbaasID == "" || database == "" || userID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_DBAAS_ID, ARUBACLOUD_DATABASE_ID and ARUBACLOUD_DBAAS_USER_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasegrantDataSourceConfig(projectID, dbaasID, database, userID),
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
						knownvalue.StringExact(database),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasegrant.test",
						tfjsonpath.New("user_id"),
						knownvalue.StringExact(userID),
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

func testAccDatabasegrantDataSourceConfig(projectID, dbaasID, database, userID string) string {
	return fmt.Sprintf(`
data "arubacloud_databasegrant" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  database   = %[3]q
  user_id    = %[4]q
}
`, projectID, dbaasID, database, userID)
}
