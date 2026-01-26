package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSchedulejobResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSchedulejobResourceConfig("test-schedulejob"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_schedulejob.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-schedulejob"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_schedulejob.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_schedulejob.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_schedulejob.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSchedulejobResourceConfig("test-schedulejob-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_schedulejob.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-schedulejob-updated"),
					),
				},
			},
		},
	})
}

func testAccSchedulejobResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_schedulejob" "test" {
  name       = %[1]q
  project_id = "test-project-id"
  location   = "it-1"
  
  properties = {
    schedule_job_type = "OneShot"
    schedule_at       = "2025-12-31T23:59:59Z"
    steps             = []
  }
}
`, name)
}
