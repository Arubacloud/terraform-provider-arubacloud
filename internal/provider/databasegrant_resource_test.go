package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDatabasegrantResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	if projectID == "" || dbaasID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_DBAAS_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabasegrantDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasegrantResourceConfig(projectID, dbaasID, "read"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("database"),
						knownvalue.StringExact("testaccgrantdb"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("role"),
						knownvalue.StringExact("read"),
					),
				},
			},
			{
				ResourceName:      "arubacloud_databasegrant.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_databasegrant.test", "project_id", "dbaas_id", "database", "user_id"),
			},
			{
				Config: testAccDatabasegrantResourceConfig(projectID, dbaasID, "write"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("role"),
						knownvalue.StringExact("write"),
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
		projectID := rs.Primary.Attributes["project_id"]
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		database := rs.Primary.Attributes["database"]
		userID := rs.Primary.Attributes["user_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + database + "/grants/" + userID)
		_, err = client.Client.FromDatabase().Grants().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Databasegrant", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("DatabaseGrant %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabasegrantResourceConfig(projectID, dbaasID, role string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaasuser" "dbgrant_prereq" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  username   = "testaccgrantuser"
  password   = "TestAcc0untP4ss!"
}

resource "arubacloud_database" "dbgrant_prereq" {
  name       = "testaccgrantdb"
  project_id = %[1]q
  dbaas_id   = %[2]q
}

resource "arubacloud_databasegrant" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  database   = arubacloud_database.dbgrant_prereq.name
  user_id    = arubacloud_dbaasuser.dbgrant_prereq.username
  role       = %[3]q
}
`, projectID, dbaasID, role)
}
