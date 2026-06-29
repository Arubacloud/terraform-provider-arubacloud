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

func TestAccDatabaseResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	if projectID == "" || dbaasID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_DBAAS_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabaseDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseResourceConfig(projectID, dbaasID, "test-database"),
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
				ImportStateIdFunc: importIDFromAttrs("arubacloud_database.test", "project_id", "dbaas_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseResourceConfig(projectID, dbaasID, "test-database-updated"),
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
		projectID := rs.Primary.Attributes["project_id"]
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + rs.Primary.ID)
		_, err = client.Client.FromDatabase().Databases().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Database", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("Database %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabaseResourceConfig(projectID, dbaasID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_database" "test" {
  name       = %[3]q
  project_id = %[1]q
  dbaas_id   = %[2]q
}
`, projectID, dbaasID, name)
}
