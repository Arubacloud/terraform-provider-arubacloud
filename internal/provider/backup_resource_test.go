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

func TestAccBackupResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckBackupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupResourceConfig(projectID, "test-backup", "ITBG-Bergamo", "Full", 30),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-backup"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact("ITBG-Bergamo"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("Full"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("retention_days"),
						knownvalue.Int64Exact(30),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				ResourceName:      "arubacloud_backup.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_backup.test", "project_id", "id"),
			},
			{
				Config: testAccBackupResourceConfig(projectID, "test-backup-updated", "ITBG-Bergamo", "Incremental", 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-backup-updated"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("Incremental"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("retention_days"),
						knownvalue.Int64Exact(60),
					),
				},
			},
		},
	})
}

func TestAccBackupResource_WithTags(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckBackupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupResourceConfigWithTags(projectID, "test-backup-tags", "ITBG-Bergamo"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-backup-tags"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}

func testCheckBackupDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_backup" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + rs.Primary.ID)
		_, err = client.Client.FromStorage().Backups().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Backup", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("Backup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccBackupResourceConfig(projectID, name, location, backupType string, retentionDays int) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "backup_vol" {
  name           = "test-acc-backup-vol"
  project_id     = %[1]q
  location       = %[3]q
  size_gb        = 10
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_backup" "test" {
  name           = %[2]q
  location       = %[3]q
  project_id     = %[1]q
  type           = %[4]q
  volume_id      = arubacloud_blockstorage.backup_vol.id
  billing_period = "Month"
  retention_days = %[5]d
}
`, projectID, name, location, backupType, retentionDays)
}

func testAccBackupResourceConfigWithTags(projectID, name, location string) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "backup_vol" {
  name           = "test-acc-backup-tags-vol"
  project_id     = %[1]q
  location       = %[3]q
  size_gb        = 10
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_backup" "test" {
  name           = %[2]q
  location       = %[3]q
  project_id     = %[1]q
  type           = "Full"
  volume_id      = arubacloud_blockstorage.backup_vol.id
  billing_period = "Month"
  retention_days = 30
  tags           = ["env:test", "managed:terraform"]
}
`, projectID, name, location)
}
