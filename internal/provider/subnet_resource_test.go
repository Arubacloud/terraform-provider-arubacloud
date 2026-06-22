package provider

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSubnetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSubnetDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSubnetResourceConfig("test-subnet"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-subnet"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("type"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSubnetResourceConfig("test-subnet-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-subnet-updated"),
					),
				},
			},
		},
	})
}

func testCheckSubnetDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_subnet" {
			continue
		}
		vpcID := rs.Primary.Attributes["vpc_id"]
		resp, err := client.Client.FromNetwork().Subnets().Get(ctx, rs.Primary.Attributes["project_id"], vpcID, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Subnet", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("Subnet %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSubnetResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_subnet" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
  vpc_id     = "test-vpc-id"
  type       = "Basic"

  network = {
    address = "10.0.0.0/24"
  }
}
`, name)
}

// ── cidrContains ──────────────────────────────────────────────────────────────

func TestCidrContains(t *testing.T) {
	parseCIDR := func(s string) *net.IPNet {
		t.Helper()
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("ParseCIDR(%q): %v", s, err)
		}
		return n
	}

	cases := []struct {
		parent string
		child  string
		want   bool
	}{
		// Subnet is contained in the parent.
		{"10.0.0.0/8", "10.1.2.0/24", true},
		// Equal networks are considered contained.
		{"10.0.0.0/24", "10.0.0.0/24", true},
		// Child prefix longer than parent — still a subset.
		{"192.168.0.0/16", "192.168.1.0/24", true},
		// Child has a different network address — not contained.
		{"10.0.0.0/24", "10.0.1.0/24", false},
		// Parent is a subnet of child (narrower mask) — not contained.
		{"10.0.0.0/24", "10.0.0.0/16", false},
		// IPv4 vs IPv6 — different bit widths, never contained.
		{"10.0.0.0/8", "::1/128", false},
	}

	for _, tc := range cases {
		parent, child := parseCIDR(tc.parent), parseCIDR(tc.child)
		if got := cidrContains(parent, child); got != tc.want {
			t.Errorf("cidrContains(%q, %q) = %v, want %v", tc.parent, tc.child, got, tc.want)
		}
	}
}
