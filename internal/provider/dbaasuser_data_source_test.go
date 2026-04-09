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
	username := os.Getenv("ARUBACLOUD_DBAAS_USERNAME")
	if projectID == "" || dbaasID == "" || username == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_DBAAS_ID and ARUBACLOUD_DBAAS_USERNAME must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbaasuserDataSourceConfig(projectID, dbaasID, username),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_dbaasuser.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_dbaasuser.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
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

func testAccDbaasuserDataSourceConfig(projectID, dbaasID, username string) string {
	return fmt.Sprintf(`
data "arubacloud_dbaasuser" "test" {
  username   = %[1]q
  project_id = %[2]q
  dbaas_id   = %[3]q
}
`, username, projectID, dbaasID)
}
