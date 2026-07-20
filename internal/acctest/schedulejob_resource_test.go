package acctest

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/terraform-provider-arubacloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSchedulejobResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	osImageID := os.Getenv("ARUBACLOUD_OS_IMAGE_ID")
	if projectID == "" || osImageID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_OS_IMAGE_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckSchedulejobDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSchedulejobResourceConfig(projectID, osImageID, "test-schedulejob"),
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
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_schedulejob.test", "project_id", "id"),
			},
			// Update and Read testing (name-only change; steps are immutable)
			{
				Config: testAccSchedulejobResourceConfig(projectID, osImageID, "test-schedulejob-updated"),
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
	client, err := AccClient()
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
		job, err := client.Client.FromSchedule().Jobs().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Schedulejob", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		// Mirror the provider's deletionChecker: the API returns HTTP 200 with
		// state=Deleted or state=Deleting instead of 404 for recently-deleted jobs.
		if job != nil {
			st := string(job.State())
			if st == "Deleted" || st == "Deleting" {
				continue
			}
		}
		return fmt.Errorf("ScheduleJob %s still exists", rs.Primary.ID)
	}
	return nil
}

// testAccSchedulejobResourceConfig builds a config with a prerequisite CloudServer so
// the schedulejob step has a valid resource_uri. The API rejects requests with zero steps.
func testAccSchedulejobResourceConfig(projectID, osImageID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "sj_rs_prereq" {
  name       = "test-rs-sj-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "sj_rs_prereq" {
  name       = "test-rs-sj-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.sj_rs_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "sj_rs_prereq" {
  name       = "test-rs-sj-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.sj_rs_prereq.id
}

resource "arubacloud_blockstorage" "sj_rs_boot" {
  name           = "test-rs-sj-boot"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 30
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
  bootable       = true
  image          = %[2]q
}

resource "arubacloud_cloudserver" "sj_rs_prereq" {
  name       = "test-rs-sj-server"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  zone       = "ITBG-1"

  network = {
    vpc_uri_ref            = arubacloud_vpc.sj_rs_prereq.uri
    subnet_uri_refs        = [arubacloud_subnet.sj_rs_prereq.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.sj_rs_prereq.uri]
  }

  settings = {
    flavor_name = "CSO4A8"
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.sj_rs_boot.uri
  }
}

resource "arubacloud_schedulejob" "test" {
  name       = %[3]q
  project_id = %[1]q
  location   = "ITBG-Bergamo"

  properties = {
    schedule_job_type = "OneShot"
    schedule_at       = "2099-12-31T23:59:59Z"
    enabled           = true
    steps = [
      {
        name         = "Power Off Server"
        resource_uri = "/projects/%[1]s/providers/Aruba.Compute/cloudServers/${arubacloud_cloudserver.sj_rs_prereq.id}"
        action_uri   = "/poweroff"
        http_verb    = "POST"
        body         = null
      }
    ]
  }
}
`, projectID, osImageID, name)
}
