package provider

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// testAccRandSuffix returns a short random hex string for use in resource names
// so that parallel or repeated test runs do not collide on unique-name constraints.
func testAccRandSuffix() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func TestAccSchedulejobDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	osImageID := os.Getenv("ARUBACLOUD_OS_IMAGE_ID")
	if projectID == "" || osImageID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_OS_IMAGE_ID must be set for acceptance tests")
	}

	suffix := testAccRandSuffix()
	sjName := "test-ds-sj-" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSchedulejobDataSourceConfig(projectID, osImageID, suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_schedulejob.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_schedulejob.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(sjName),
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

func testAccSchedulejobDataSourceConfig(projectID, osImageID, suffix string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "sj_prereq" {
  name       = "test-ds-sj-vpc-%[3]s"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "sj_prereq" {
  name       = "test-ds-sj-subnet-%[3]s"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.sj_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "sj_prereq" {
  name       = "test-ds-sj-sg-%[3]s"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.sj_prereq.id
}

resource "arubacloud_blockstorage" "sj_boot" {
  name           = "test-ds-sj-boot-%[3]s"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 30
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
  bootable       = true
  image          = %[2]q
}

resource "arubacloud_cloudserver" "sj_prereq" {
  name       = "test-ds-sj-srv-%[3]s"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  zone       = "ITBG-1"

  network = {
    vpc_uri_ref            = arubacloud_vpc.sj_prereq.uri
    subnet_uri_refs        = [arubacloud_subnet.sj_prereq.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.sj_prereq.uri]
  }

  settings = {
    flavor_name = "CSO4A8"
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.sj_boot.uri
  }
}

resource "arubacloud_schedulejob" "test" {
  name       = "test-ds-sj-%[3]s"
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
        resource_uri = "/projects/%[1]s/providers/Aruba.Compute/cloudServers/${arubacloud_cloudserver.sj_prereq.id}"
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
`, projectID, osImageID, suffix)
}
