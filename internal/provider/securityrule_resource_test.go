package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestNormalizeProtocol(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"any", "Any"},
		{"ANY", "Any"},
		{"Any", "Any"},
		{"tcp", "TCP"},
		{"TCP", "TCP"},
		{"udp", "UDP"},
		{"UDP", "UDP"},
		{"icmp", "ICMP"},
		{"ICMP", "ICMP"},
		{"", ""},
		{"other", "Other"},
	}
	for _, tc := range cases {
		if got := normalizeProtocol(tc.in); got != tc.want {
			t.Errorf("normalizeProtocol(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeTargetKind(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"IP", "IP"},
		{"ip", "IP"},
		{"Ip", "IP"},
		{"SecurityGroup", "SecurityGroup"},
		{"securitygroup", "SecurityGroup"},
		{"SECURITYGROUP", "SecurityGroup"},
		{"", ""},
		{"unknown", "unknown"},
	}
	for _, tc := range cases {
		if got := normalizeTargetKind(tc.in); got != tc.want {
			t.Errorf("normalizeTargetKind(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestProtocolNormalizePlanModifier(t *testing.T) {
	ctx := context.Background()
	m := protocolNormalizePlanModifier{}

	cases := []struct {
		name       string
		planValue  types.String
		stateValue types.String
		wantValue  string
		wantNull   bool
	}{
		{
			name:       "known plan value is uppercased",
			planValue:  types.StringValue("tcp"),
			stateValue: types.StringNull(),
			wantValue:  "TCP",
		},
		{
			name:       "known plan Any is uppercased to ANY",
			planValue:  types.StringValue("Any"),
			stateValue: types.StringNull(),
			wantValue:  "ANY",
		},
		{
			name:       "null plan falls back to state value",
			planValue:  types.StringNull(),
			stateValue: types.StringValue("UDP"),
			wantValue:  "UDP",
		},
		{
			name:       "null plan and null state — plan stays null",
			planValue:  types.StringNull(),
			stateValue: types.StringNull(),
			wantNull:   true,
		},
		{
			name:       "unknown plan falls back to state value",
			planValue:  types.StringUnknown(),
			stateValue: types.StringValue("ICMP"),
			wantValue:  "ICMP",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				PlanValue:  tc.planValue,
				StateValue: tc.stateValue,
			}
			resp := &planmodifier.StringResponse{PlanValue: tc.planValue}
			m.PlanModifyString(ctx, req, resp)
			if tc.wantNull {
				if !resp.PlanValue.IsNull() {
					t.Errorf("expected null plan value, got %q", resp.PlanValue.ValueString())
				}
				return
			}
			if resp.PlanValue.ValueString() != tc.wantValue {
				t.Errorf("PlanModifyString: got %q, want %q", resp.PlanValue.ValueString(), tc.wantValue)
			}
		})
	}
}

func TestAccSecurityruleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSecurityruleDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecurityruleResourceConfig("test-securityrule"),
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
			},
			// Update and Read testing
			{
				Config: testAccSecurityruleResourceConfig("test-securityrule-updated"),
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
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_securityrule" {
			continue
		}
		vpcID := rs.Primary.Attributes["vpc_id"]
		sgID := rs.Primary.Attributes["security_group_id"]
		resp, err := client.Client.FromNetwork().SecurityGroupRules().Get(ctx, rs.Primary.Attributes["project_id"], vpcID, sgID, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Securityrule", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("SecurityRule %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSecurityruleResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_securityrule" "test" {
  name               = %[1]q
  location           = "it-1"
  project_id         = "test-project-id"
  vpc_id             = "test-vpc-id"
  security_group_id  = "test-sg-id"
  
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "IP"
      value = "0.0.0.0/0"
    }
  }
}
`, name)
}
