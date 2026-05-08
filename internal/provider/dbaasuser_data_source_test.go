package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDbaasuserDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	if projectID == "" || dbaasID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_DBAAS_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbaasuserDataSourceConfig(projectID, dbaasID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_dbaasuser.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_dbaasuser.test",
						tfjsonpath.New("username"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_dbaasuser.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_dbaasuser.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.StringExact(dbaasID),
					),
				},
			},
		},
	})
}

func testAccDbaasuserDataSourceConfig(projectID, dbaasID string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaasuser" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  username   = "test-ds-user"
  password   = "TestPassword123!"
}

data "arubacloud_dbaasuser" "test" {
  username   = arubacloud_dbaasuser.test.username
  project_id = %[1]q
  dbaas_id   = %[2]q
}
`, projectID, dbaasID)
}
