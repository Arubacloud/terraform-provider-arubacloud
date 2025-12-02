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

func TestAccBlockStorageResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBlockStorageResourceConfig("test-blockstorage", 100, "Standard"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-blockstorage"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("properties").AtMapKey("size_gb"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("properties").AtMapKey("type"),
						knownvalue.StringExact("Standard"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("blockstorage-id"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_blockstorage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccBlockStorageResourceConfig("test-blockstorage-updated", 200, "Performance"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-blockstorage-updated"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("properties").AtMapKey("size_gb"),
						knownvalue.Int64Exact(200),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("properties").AtMapKey("type"),
						knownvalue.StringExact("Performance"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccBlockStorageResource_Bootable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageResourceConfigBootable("test-bootable", 50),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-bootable"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("properties").AtMapKey("bootable"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("properties").AtMapKey("image"),
						knownvalue.StringExact("ubuntu-22.04"),
					),
				},
			},
		},
	})
}

func testAccBlockStorageResourceConfig(name string, sizeGB int, storageType string) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "test" {
  name       = %[1]q
  project_id = "project-123"
  
  properties = {
    size_gb        = %[2]d
    billing_period = "Hour"
    zone           = "it-1"
    type           = %[3]q
  }
}
`, name, sizeGB, storageType)
}

func testAccBlockStorageResourceConfigBootable(name string, sizeGB int) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "test" {
  name       = %[1]q
  project_id = "project-123"
  
  properties = {
    size_gb        = %[2]d
    billing_period = "Hour"
    zone           = "it-1"
    type           = "Standard"
    bootable       = true
    image          = "ubuntu-22.04"
  }
}
`, name, sizeGB)
}
