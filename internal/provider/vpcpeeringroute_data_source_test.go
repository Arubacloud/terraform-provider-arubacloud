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

func TestAccVpcpeeringrouteDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcpeeringrouteDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("example-vpcpeeringroute"),
					),
				},
			},
		},
	})
}

const testAccVpcpeeringrouteDataSourceConfig = `
data "arubacloud_vpcpeeringroute" "test" {
  name = "example-vpcpeeringroute"
  # TODO: Add required fields based on the schema
  # Check vpcpeeringroute_data_source.go for required attributes
}
`
