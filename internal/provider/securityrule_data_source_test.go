package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSecurityruleDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	vpcID := os.Getenv("ARUBACLOUD_VPC_ID")
	sgID := os.Getenv("ARUBACLOUD_SECURITYGROUP_ID")
	ruleID := os.Getenv("ARUBACLOUD_SECURITYRULE_ID")
	if projectID == "" || vpcID == "" || sgID == "" || ruleID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_VPC_ID, ARUBACLOUD_SECURITYGROUP_ID and ARUBACLOUD_SECURITYRULE_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityruleDataSourceConfig(projectID, vpcID, sgID, ruleID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.StringExact(vpcID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("security_group_id"),
						knownvalue.StringExact(sgID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("direction"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securityrule.test",
						tfjsonpath.New("protocol"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccSecurityruleDataSourceConfig(projectID, vpcID, sgID, ruleID string) string {
	return fmt.Sprintf(`
data "arubacloud_securityrule" "test" {
  id                = %[1]q
  project_id        = %[2]q
  vpc_id            = %[3]q
  security_group_id = %[4]q
}
`, ruleID, projectID, vpcID, sgID)
}
