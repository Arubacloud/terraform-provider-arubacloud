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

func TestAccDatabasegrantResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDatabasegrantDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabasegrantResourceConfig(projectID, "read"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("database"),
						knownvalue.StringExact("testaccgrantdb"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("role"),
						knownvalue.StringExact("read"),
					),
				},
			},
			{
				ResourceName:      "arubacloud_databasegrant.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_databasegrant.test", "project_id", "dbaas_id", "database", "user_id"),
			},
			{
				Config: testAccDatabasegrantResourceConfig(projectID, "write"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_databasegrant.test",
						tfjsonpath.New("role"),
						knownvalue.StringExact("write"),
					),
				},
			},
		},
	})
}

func testCheckDatabasegrantDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_databasegrant" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		database := rs.Primary.Attributes["database"]
		userID := rs.Primary.Attributes["user_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + database + "/grants/" + userID)
		_, err = client.Client.FromDatabase().Grants().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Databasegrant", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("DatabaseGrant %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDatabasegrantResourceConfig(projectID, role string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "dbgrant_prereq" {
  name       = "test-acc-dbgrant-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "dbgrant_prereq" {
  name       = "test-acc-dbgrant-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.dbgrant_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "dbgrant_prereq" {
  name       = "test-acc-dbgrant-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.dbgrant_prereq.id
}

resource "arubacloud_dbaas" "dbgrant_prereq" {
  name       = "test-acc-dbgrant-dbaas"
  location   = "ITBG-Bergamo"
  zone       = "ITBG-1"
  project_id = %[1]q
  engine_id  = "mysql-8.0"
  flavor     = "DBO2A4"

  storage = {
    size_gb = 20
  }

  network = {
    vpc_uri_ref            = arubacloud_vpc.dbgrant_prereq.uri
    subnet_uri_ref         = arubacloud_subnet.dbgrant_prereq.uri
    security_group_uri_ref = arubacloud_securitygroup.dbgrant_prereq.uri
  }
}

resource "arubacloud_dbaasuser" "dbgrant_prereq" {
  project_id = %[1]q
  dbaas_id   = arubacloud_dbaas.dbgrant_prereq.id
  username   = "testaccgrantuser"
  password   = "Acc3ptAbl3P@ss#01"
}

resource "arubacloud_database" "dbgrant_prereq" {
  name       = "testaccgrantdb"
  project_id = %[1]q
  dbaas_id   = arubacloud_dbaas.dbgrant_prereq.id
}

resource "arubacloud_databasegrant" "test" {
  project_id = %[1]q
  dbaas_id   = arubacloud_dbaas.dbgrant_prereq.id
  database   = arubacloud_database.dbgrant_prereq.name
  user_id    = arubacloud_dbaasuser.dbgrant_prereq.username
  role       = %[2]q
}
`, projectID, role)
}
