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

func TestAccDatabasebackupResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabasebackupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasebackupResourceConfig(projectID, "test-databasebackup"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasebackup"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				ResourceName:      "arubacloud_databasebackup.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_databasebackup.test", "project_id", "id"),
			},
			{
				Config: testAccDatabasebackupResourceConfig(projectID, "test-databasebackup-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasebackup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-databasebackup-updated"),
					),
				},
			},
		},
	})
}

func testCheckDatabasebackupDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_databasebackup" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/backups/" + rs.Primary.ID)
		_, err = client.Client.FromDatabase().Backups().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Databasebackup", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("DatabaseBackup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabasebackupResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "dbbackup_prereq" {
  name       = "test-acc-dbbackup-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "dbbackup_prereq" {
  name       = "test-acc-dbbackup-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.dbbackup_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "dbbackup_prereq" {
  name       = "test-acc-dbbackup-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.dbbackup_prereq.id
}

resource "arubacloud_dbaas" "dbbackup_prereq" {
  name       = "test-acc-dbbackup-dbaas"
  location   = "ITBG-Bergamo"
  zone       = "ITBG-1"
  project_id = %[1]q
  engine_id  = "mysql-8.0"
  flavor     = "DBO2A4"

  storage = {
    size_gb = 20
  }

  network = {
    vpc_uri_ref            = arubacloud_vpc.dbbackup_prereq.uri
    subnet_uri_ref         = arubacloud_subnet.dbbackup_prereq.uri
    security_group_uri_ref = arubacloud_securitygroup.dbbackup_prereq.uri
  }
}

resource "arubacloud_database" "dbbackup_prereq" {
  name       = "testaccdbbackupdb"
  project_id = %[1]q
  dbaas_id   = arubacloud_dbaas.dbbackup_prereq.id
}

resource "arubacloud_databasebackup" "test" {
  name           = %[2]q
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  zone           = "ITBG-1"
  dbaas_id       = arubacloud_dbaas.dbbackup_prereq.id
  database       = arubacloud_database.dbbackup_prereq.name
  billing_period = "Hour"
}
`, projectID, name)
}
