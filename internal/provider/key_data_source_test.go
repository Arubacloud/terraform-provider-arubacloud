package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeyDataSource(t *testing.T) {
	t.Skip("Skipping test - Key resource requires SDK fix to return project_id and kms_id in API response")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccKeyDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.arubacloud_key.test", "id"),
					resource.TestCheckResourceAttrSet("data.arubacloud_key.test", "name"),
					resource.TestCheckResourceAttrSet("data.arubacloud_key.test", "algorithm"),
					resource.TestCheckResourceAttrSet("data.arubacloud_key.test", "size"),
				),
			},
		},
	})
}

func testAccKeyDataSourceConfig() string {
	return `
resource "arubacloud_project" "test" {
  name        = "test-project"
  description = "Test project for Key data source"
}

resource "arubacloud_kms" "test" {
  name           = "test-kms"
  project_id     = arubacloud_project.test.id
  location       = "it-mil1"
  billing_period = "monthly"
}

resource "arubacloud_key" "test" {
  name        = "test-key"
  project_id  = arubacloud_project.test.id
  kms_id      = arubacloud_kms.test.id
  algorithm   = "AES"
  size        = 256
  description = "Test key for data source"
}

data "arubacloud_key" "test" {
  id         = arubacloud_key.test.id
  project_id = arubacloud_project.test.id
  kms_id     = arubacloud_kms.test.id
}
`
}
