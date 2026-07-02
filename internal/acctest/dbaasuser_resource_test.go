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

// testAccDBaaSPassword returns the password to use for DBaaS user acceptance tests.
// The password is read from ARUBACLOUD_DBAAS_PASSWORD; the test is skipped if the
// var is unset so callers can provide a password that satisfies the current API
// policy without requiring a code change.
func testAccDBaaSPassword(t *testing.T) string {
	t.Helper()
	pw := os.Getenv("ARUBACLOUD_DBAAS_PASSWORD")
	if pw == "" {
		t.Skip("ARUBACLOUD_DBAAS_PASSWORD must be set for acceptance tests that create DBaaS users")
	}
	return pw
}

func TestAccDbaasuserResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	dbaasID := os.Getenv("ARUBACLOUD_DBAAS_ID")
	if projectID == "" || dbaasID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_DBAAS_ID must be set for acceptance tests")
	}
	password := testAccDBaaSPassword(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckDbaasuserDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDbaasuserResourceConfig(projectID, dbaasID, "testaccdbuser", password),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("testaccdbuser"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaasuser.test",
						tfjsonpath.New("dbaas_id"),
						knownvalue.StringExact(dbaasID),
					),
				},
			},
			{
				ResourceName:            "arubacloud_dbaasuser.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc:       ImportIDFromAttrs("arubacloud_dbaasuser.test", "project_id", "dbaas_id", "id"),
			},
		},
	})
}

func testCheckDbaasuserDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_dbaasuser" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		dbaasID := rs.Primary.Attributes["dbaas_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + rs.Primary.ID)
		_, err = client.Client.FromDatabase().Users().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Dbaasuser", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("DBaaSUser %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDbaasuserResourceConfig(projectID, dbaasID, username, password string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaasuser" "test" {
  project_id = %[1]q
  dbaas_id   = %[2]q
  username   = %[3]q
  password   = %[4]q
}
`, projectID, dbaasID, username, password)
}
