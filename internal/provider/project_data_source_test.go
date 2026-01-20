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

func TestAccProjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_project.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_project.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("example-project"),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_project.test",
						tfjsonpath.New("description"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_project.test",
						tfjsonpath.New("tags"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

const testAccProjectDataSourceConfig = `
data "arubacloud_project" "test" {
  id = "test-project-id"
}
`
