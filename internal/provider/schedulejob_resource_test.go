package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSchedulejobResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSchedulejobDestroyed,
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

func testCheckSchedulejobDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_schedulejob" {
			continue
		}
		resp, err := client.Client.FromSchedule().Jobs().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Schedulejob", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("ScheduleJob %s still exists", rs.Primary.ID)
	}
	return nil
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
