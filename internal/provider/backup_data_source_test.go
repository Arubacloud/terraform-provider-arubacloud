// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccBackupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_backup.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("backup-id"),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_backup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("example-backup"),
					),
				},
			},
		},
	})
}

const testAccBackupDataSourceConfig = `
data "arubacloud_backup" "test" {
  name = "example-backup"
}
`
