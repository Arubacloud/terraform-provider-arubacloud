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

func TestAccSubnetDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	vpcID := os.Getenv("ARUBACLOUD_VPC_ID")
	subnetID := os.Getenv("ARUBACLOUD_SUBNET_ID")
	if projectID == "" || vpcID == "" || subnetID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_VPC_ID and ARUBACLOUD_SUBNET_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetDataSourceConfig(projectID, vpcID, subnetID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_subnet.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_subnet.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_subnet.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.StringExact(vpcID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_subnet.test",
						tfjsonpath.New("type"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_subnet.test",
						tfjsonpath.New("tags"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccSubnetDataSourceConfig(projectID, vpcID, subnetID string) string {
	return fmt.Sprintf(`
data "arubacloud_subnet" "test" {
  id         = %[1]q
  project_id = %[2]q
  vpc_id     = %[3]q
}
`, subnetID, projectID, vpcID)
}
