package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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

func TestTargetKindNormalizePlanModifier(t *testing.T) {
	ctx := context.Background()
	m := targetKindNormalizePlanModifier{}

	cases := []struct {
		name       string
		planValue  types.String
		stateValue types.String
		wantValue  string
		wantNull   bool
	}{
		{
			name:      "Ip is normalized to IP (no prior state)",
			planValue: types.StringValue("Ip"),
			wantValue: "IP",
		},
		{
			name:      "ip is normalized to IP",
			planValue: types.StringValue("ip"),
			wantValue: "IP",
		},
		{
			name:      "IP stays IP",
			planValue: types.StringValue("IP"),
			wantValue: "IP",
		},
		{
			name:      "securitygroup normalized to SecurityGroup",
			planValue: types.StringValue("securitygroup"),
			wantValue: "SecurityGroup",
		},
		{
			name:       "Ip in plan with Ip in state — preserves state value (no spurious replace)",
			planValue:  types.StringValue("Ip"),
			stateValue: types.StringValue("Ip"),
			wantValue:  "Ip",
		},
		{
			name:       "Ip in plan with IP in state (post-import) — plan becomes IP, no drift",
			planValue:  types.StringValue("Ip"),
			stateValue: types.StringValue("IP"),
			wantValue:  "IP",
		},
		{
			name:       "null plan falls back to state",
			planValue:  types.StringNull(),
			stateValue: types.StringValue("IP"),
			wantValue:  "IP",
		},
		{
			name:       "null plan and null state — stays null",
			planValue:  types.StringNull(),
			stateValue: types.StringNull(),
			wantNull:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateVal := tc.stateValue
			if stateVal == (types.String{}) {
				stateVal = types.StringNull()
			}
			req := planmodifier.StringRequest{
				PlanValue:  tc.planValue,
				StateValue: stateVal,
			}
			resp := &planmodifier.StringResponse{PlanValue: tc.planValue}
			m.PlanModifyString(ctx, req, resp)
			if tc.wantNull {
				if !resp.PlanValue.IsNull() {
					t.Errorf("expected null, got %q", resp.PlanValue.ValueString())
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
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
				ImportStateIdFunc: importIDFromAttrs("arubacloud_securityrule.test", "project_id", "vpc_id", "security_group_id", "id", "location"),
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
	client, err := testAccClient()
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
		if provErr := CheckResponseErr("get", "Securityrule", err); provErr != nil {
			if IsNotFound(provErr) {
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
