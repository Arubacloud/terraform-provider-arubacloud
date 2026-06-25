package provider

import (
	"context"
	"fmt"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
		projectID := rs.Primary.Attributes["project_id"]
		backupID := rs.Primary.Attributes["backup_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID + "/restores/" + rs.Primary.ID)
		_, err = client.Client.FromStorage().Restores().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Restore", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
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
