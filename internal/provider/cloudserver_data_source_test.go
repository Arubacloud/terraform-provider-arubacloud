package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCloudserverDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudserverDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					// Test flattened fields (not nested in properties)
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("zone"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

const testAccCloudserverDataSourceConfig = `
data "arubacloud_cloudserver" "test" {
  id = "test-cloudserver-id"
}
`
