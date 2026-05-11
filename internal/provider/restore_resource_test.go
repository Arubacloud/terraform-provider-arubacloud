package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccRestoreResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckRestoreDestroyed,
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

func testCheckRestoreDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_restore" {
			continue
		}
		backupID := rs.Primary.Attributes["backup_id"]
		resp, err := client.Client.FromStorage().Restores().Get(ctx, rs.Primary.Attributes["project_id"], backupID, rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Restore", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("Restore %s still exists", rs.Primary.ID)
	}
	return nil
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
