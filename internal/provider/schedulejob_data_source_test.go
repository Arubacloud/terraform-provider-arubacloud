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
	cloudserverID := os.Getenv("ARUBACLOUD_CLOUDSERVER_ID")
	if cloudserverID == "" {
		t.Skip("ARUBACLOUD_CLOUDSERVER_ID must be set for schedulejob acceptance tests (step resource_uri requires an existing cloud server)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSchedulejobDataSourceConfig(projectID, cloudserverID),
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

func testAccSchedulejobDataSourceConfig(projectID, cloudserverID string) string {
	return fmt.Sprintf(`
resource "arubacloud_schedulejob" "test" {
  name       = "test-ds-schedulejob"
  project_id = %[1]q
  location   = "ITBG-Bergamo"
  tags       = []

  properties = {
    schedule_job_type = "OneShot"
    schedule_at       = "2099-12-31T23:59:59+00:00"
    enabled           = true
    steps = [
      {
        name         = "Power Off Server"
        resource_uri = "/projects/%[1]s/providers/Aruba.Compute/cloudServers/%[2]s"
        action_uri   = "/poweroff"
        http_verb    = "POST"
        body         = null
      }
    ]
  }
}

data "arubacloud_schedulejob" "test" {
  id         = arubacloud_schedulejob.test.id
  project_id = %[1]q
}
`, projectID, cloudserverID)
}
