package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccBlockStorageDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_blockstorage.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("example-blockstorage"),
					),
					// Test flattened fields (not nested in properties)
					statecheck.ExpectKnownValue(
						"data.arubacloud_blockstorage.test",
						tfjsonpath.New("size_gb"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_blockstorage.test",
						tfjsonpath.New("billing_period"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_blockstorage.test",
						tfjsonpath.New("zone"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_blockstorage.test",
						tfjsonpath.New("type"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

const testAccBlockStorageDataSourceConfig = `
data "arubacloud_blockstorage" "test" {
  id = "test-blockstorage-id"
}
`
