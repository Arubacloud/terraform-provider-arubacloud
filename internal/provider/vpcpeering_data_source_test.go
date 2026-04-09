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

func TestAccVpcpeeringDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	vpcID := os.Getenv("ARUBACLOUD_VPC_ID")
	peeringID := os.Getenv("ARUBACLOUD_VPCPEERING_ID")
	if projectID == "" || vpcID == "" || peeringID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_VPC_ID and ARUBACLOUD_VPCPEERING_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcpeeringDataSourceConfig(projectID, vpcID, peeringID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeering.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeering.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeering.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeering.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.StringExact(vpcID),
					),
				},
			},
		},
	})
}

func testAccVpcpeeringDataSourceConfig(projectID, vpcID, peeringID string) string {
	return fmt.Sprintf(`
data "arubacloud_vpcpeering" "test" {
  id         = %[1]q
  project_id = %[2]q
  vpc_id     = %[3]q
}
`, peeringID, projectID, vpcID)
}
