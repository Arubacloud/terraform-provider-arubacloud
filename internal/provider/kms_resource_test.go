package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccKmsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckKmsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsResourceConfig("test-kms"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("arubacloud_kms.test", tfjsonpath.New("name"), knownvalue.StringExact("test-kms")),
					statecheck.ExpectKnownValue("arubacloud_kms.test", tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("arubacloud_kms.test", tfjsonpath.New("location"), knownvalue.NotNull()),
				},
			},
			{ResourceName: "arubacloud_kms.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccKmsResourceConfig("test-kms-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("arubacloud_kms.test", tfjsonpath.New("name"), knownvalue.StringExact("test-kms-updated")),
				},
			},
		},
	})
}

func testCheckKmsDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_kms" {
			continue
		}
		resp, err := client.Client.FromSecurity().KMS().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Kms", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("KMS key %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccKmsResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_kms" "test" {
  name           = %[1]q
  project_id     = "test-project-id"
  location       = "it-1"
  billing_period = "Hour"
}
`, name)
}
