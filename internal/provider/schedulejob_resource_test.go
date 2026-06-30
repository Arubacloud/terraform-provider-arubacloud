package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSchedulejobResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSchedulejobDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSchedulejobResourceConfig(projectID, "test-schedulejob"),
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
				ImportStateIdFunc: importIDFromAttrs("arubacloud_schedulejob.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccSchedulejobResourceConfig(projectID, "test-schedulejob-updated"),
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
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Schedule/jobs/" + rs.Primary.ID)
		_, err = client.Client.FromSchedule().Jobs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Schedulejob", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("ScheduleJob %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSchedulejobResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_schedulejob" "test" {
  name       = %[2]q
  project_id = %[1]q
  location   = "ITBG-Bergamo"

  properties = {
    schedule_job_type = "OneShot"
    schedule_at       = "2030-12-31T23:59:59Z"
    steps             = []
  }
}
`, projectID, name)
}
