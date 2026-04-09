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

func TestAccSecuritygroupDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	vpcID := os.Getenv("ARUBACLOUD_VPC_ID")
	sgID := os.Getenv("ARUBACLOUD_SECURITYGROUP_ID")
	if projectID == "" || vpcID == "" || sgID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_VPC_ID and ARUBACLOUD_SECURITYGROUP_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecuritygroupDataSourceConfig(projectID, vpcID, sgID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.StringExact(vpcID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.test",
						tfjsonpath.New("tags"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccSecuritygroupDataSourceConfig(projectID, vpcID, sgID string) string {
	return fmt.Sprintf(`
data "arubacloud_securitygroup" "test" {
  id         = %[1]q
  project_id = %[2]q
  vpc_id     = %[3]q
}
`, sgID, projectID, vpcID)
}
