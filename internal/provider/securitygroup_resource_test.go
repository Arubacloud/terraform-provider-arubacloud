package provider

import (
	"context"
	"fmt"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSecuritygroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSecuritygroupDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecuritygroupResourceConfig("test-securitygroup"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_securitygroup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-securitygroup"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_securitygroup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_securitygroup.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_securitygroup.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_securitygroup.test", "project_id", "vpc_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccSecuritygroupResourceConfig("test-securitygroup-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_securitygroup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-securitygroup-updated"),
					),
				},
			},
		},
	})
}

func testCheckSecuritygroupDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_securitygroup" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		vpcID := rs.Primary.Attributes["vpc_id"]
		ref := aruba.SecurityGroupRef(projectID, vpcID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().SecurityGroups().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Securitygroup", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("SecurityGroup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSecuritygroupResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_securitygroup" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
  vpc_id     = "test-vpc-id"
}
`, name)
}
