package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccRestoreResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRestoreResourceConfig("test-restore"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_restore.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-restore"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_restore.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_restore.test",
						tfjsonpath.New("backup_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_restore.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccRestoreResourceConfig("test-restore-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_restore.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-restore-updated"),
					),
				},
			},
		},
	})
}

func testAccRestoreResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_restore" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
  backup_id  = "test-backup-id"
  volume_id  = "test-volume-id"
}
`, name)
}
