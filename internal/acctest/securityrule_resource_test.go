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

func TestAccSecurityruleResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckSecurityruleDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecurityruleResourceConfig(projectID, "test-securityrule"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_securityrule.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-securityrule"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_securityrule.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_securityrule.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_securityrule.test",
						tfjsonpath.New("security_group_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_securityrule.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_securityrule.test", "project_id", "vpc_id", "security_group_id", "id", "location"),
			},
			// Update and Read testing
			{
				Config: testAccSecurityruleResourceConfig(projectID, "test-securityrule-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_securityrule.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-securityrule-updated"),
					),
				},
			},
		},
	})
}

func testCheckSecurityruleDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_securityrule" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		vpcID := rs.Primary.Attributes["vpc_id"]
		sgID := rs.Primary.Attributes["security_group_id"]
		ref := aruba.SecurityRuleRef(projectID, vpcID, sgID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().SecurityGroupRules().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Securityrule", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("SecurityRule %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSecurityruleResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "sr_prereq" {
  name       = "test-acc-sr-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_securitygroup" "sr_prereq" {
  name       = "test-acc-sr-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.sr_prereq.id
}

resource "arubacloud_securityrule" "test" {
  name              = %[2]q
  location          = "ITBG-Bergamo"
  project_id        = %[1]q
  vpc_id            = arubacloud_vpc.sr_prereq.id
  security_group_id = arubacloud_securitygroup.sr_prereq.id

  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}
`, projectID, name)
}
