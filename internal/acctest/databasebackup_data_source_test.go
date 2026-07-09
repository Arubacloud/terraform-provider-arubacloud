package acctest

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDatabasebackupDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasebackupDataSourceConfig(projectID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_databasebackup.test",
						tfjsonpath.New("tags"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDatabasebackupDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = "test-ds-dbbackup-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "test" {
  name       = "test-ds-dbbackup-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "test" {
  name       = "test-ds-dbbackup-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
}

resource "arubacloud_dbaas" "test" {
  name       = "test-ds-dbbackup-dbaas"
  location   = "ITBG-Bergamo"
  zone       = "ITBG-1"
  project_id = %[1]q
  engine_id  = "mysql-8.0"
  flavor     = "DBO2A4"
  tags       = ["acceptance-test"]
  timeout    = "10m"

  storage = {
    size_gb = 20
  }

  network = {
    vpc_uri_ref            = arubacloud_vpc.test.uri
    subnet_uri_ref         = arubacloud_subnet.test.uri
    security_group_uri_ref = arubacloud_securitygroup.test.uri
  }
}

resource "arubacloud_database" "test" {
  name       = "testdb"
  project_id = %[1]q
  dbaas_id   = arubacloud_dbaas.test.id
}

resource "arubacloud_databasebackup" "test" {
  name           = "test-ds-databasebackup"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  zone           = "ITBG-1"
  dbaas_id       = arubacloud_dbaas.test.id
  database       = arubacloud_database.test.name
  billing_period = "Hour"
  tags           = ["acceptance-test"]
}

data "arubacloud_databasebackup" "test" {
  id         = arubacloud_databasebackup.test.id
  project_id = %[1]q
}
`, projectID)
}
