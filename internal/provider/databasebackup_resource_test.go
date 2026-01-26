package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDatabasebackupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
