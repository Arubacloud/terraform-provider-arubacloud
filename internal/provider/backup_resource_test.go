// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccBackupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBackupResourceConfig("test-backup", "de-1", "full", 30),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-backup"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact("de-1"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("full"),
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
			// ImportState testing
			{
				ResourceName:      "arubacloud_backup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccBackupResourceConfig("test-backup-updated", "de-1", "incremental", 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-backup-updated"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("incremental"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_backup.test",
						tfjsonpath.New("retention_days"),
						knownvalue.Int64Exact(60),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccBackupResource_WithTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupResourceConfigWithTags("test-backup-tags", "de-1"),
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

func testAccBackupResourceConfig(name, location, backupType string, retentionDays int) string {
	return fmt.Sprintf(`
resource "arubacloud_backup" "test" {
  name           = %[1]q
  location       = %[2]q
  project_id     = "project-123"
  type           = %[3]q
  volume_id      = "volume-123"
  billing_period = "monthly"
  retention_days = %[4]d
}
`, name, location, backupType, retentionDays)
}

func testAccBackupResourceConfigWithTags(name, location string) string {
	return fmt.Sprintf(`
resource "arubacloud_backup" "test" {
  name           = %[1]q
  location       = %[2]q
  project_id     = "project-123"
  type           = "full"
  volume_id      = "volume-123"
  billing_period = "monthly"
  retention_days = 30
  tags           = ["env:test", "managed:terraform"]
}
`, name, location)
}
