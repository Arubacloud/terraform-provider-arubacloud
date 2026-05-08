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

func TestAccSchedulejobDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSchedulejobDataSourceConfig(projectID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_schedulejob.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_schedulejob.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_schedulejob.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
				},
			},
		},
	})
}

func testAccSchedulejobDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
resource "arubacloud_schedulejob" "test" {
  name       = "test-ds-schedulejob"
  project_id = %[1]q
  location   = "ITBG-Bergamo"

  properties = {
    schedule_job_type = "OneShot"
    schedule_at       = "2099-12-31T23:59:59Z"
    steps             = []
  }
}

data "arubacloud_schedulejob" "test" {
  id         = arubacloud_schedulejob.test.id
  project_id = %[1]q
}
`, projectID)
}
