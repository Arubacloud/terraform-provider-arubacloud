package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSecuritygroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecuritygroupDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					// Test flattened fields
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

const testAccSecuritygroupDataSourceConfig = `
data "arubacloud_securitygroup" "test" {
  id = "test-securitygroup-id"
}
`
