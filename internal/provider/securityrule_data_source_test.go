package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSecurityruleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityruleDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("security_group_id"),
						knownvalue.NotNull(),
					),
					// Test flattened fields
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("direction"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("protocol"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

const testAccSecurityruleDataSourceConfig = `
data "arubacloud_securityrule" "test" {
  id = "test-securityrule-id"
}
`
