// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeyResource(t *testing.T) {
	t.Skip("Skipping test - Key resource requires SDK fix to return project_id and kms_id in API response")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKeyResourceConfig("test-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arubacloud_key.test", "name", "test-key"),
					resource.TestCheckResourceAttrSet("arubacloud_key.test", "id"),
					resource.TestCheckResourceAttrSet("arubacloud_key.test", "algorithm"),
					resource.TestCheckResourceAttrSet("arubacloud_key.test", "size"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKeyResourceConfig("test-key-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arubacloud_key.test", "name", "test-key-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccKeyResourceConfig(name string) string {
	return `
resource "arubacloud_project" "test" {
  name        = "test-project"
  description = "Test project for Key resource"
}

resource "arubacloud_kms" "test" {
  name           = "test-kms"
  project_id     = arubacloud_project.test.id
  location       = "it-mil1"
  billing_period = "monthly"
}

resource "arubacloud_key" "test" {
  name        = "` + name + `"
  project_id  = arubacloud_project.test.id
  kms_id      = arubacloud_kms.test.id
  algorithm   = "AES"
  size        = 256
  description = "Test key for terraform provider"
}
`
}
