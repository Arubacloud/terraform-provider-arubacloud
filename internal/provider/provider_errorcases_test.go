package provider

// Error-case unit tests — exercise schema-level validation that is enforced
// at plan time before any API call is made. These use resource.UnitTest so
// they run without TF_ACC=1 set.
//
// Pattern: each test omits or misspells a Required attribute and asserts that
// the plan-time error message matches the expected regexp.
//
// To add coverage for a new resource, copy one of the functions below and
// adjust the resource type, missing field, and error regexp.

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// --- Network ---

func TestUnitVpcResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_vpc" "test" {
  name     = "test-vpc"
  location = "ITBG-Bergamo"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitSubnetResource_MissingVpcID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_subnet" "test" {
  name       = "test-subnet"
  location   = "ITBG-Bergamo"
  project_id = "test-project"
  type       = "Basic"
}
`,
				ExpectError: regexp.MustCompile(`vpc_id`),
			},
		},
	})
}

func TestUnitElasticipResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_elasticip" "test" {
  name     = "test-eip"
  location = "ITBG-Bergamo"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitVpntunnelResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_vpntunnel" "test" {
  name     = "test-tunnel"
  location = "ITBG-Bergamo"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitSecuritygroupResource_MissingVpcID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_securitygroup" "test" {
  name       = "test-sg"
  location   = "ITBG-Bergamo"
  project_id = "test-project"
}
`,
				ExpectError: regexp.MustCompile(`vpc_id`),
			},
		},
	})
}

// --- Compute ---

func TestUnitKeypairResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_keypair" "test" {
  name     = "test-keypair"
  location = "ITBG-Bergamo"
  value    = "ssh-rsa AAAA..."
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

// --- Storage ---

func TestUnitBlockstorageResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_blockstorage" "test" {
  name           = "test-vol"
  location       = "ITBG-Bergamo"
  size_gb        = 50
  billing_period = "Hour"
  zone           = "ITBG-Bergamo"
  type           = "Standard"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitSnapshotResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_snapshot" "test" {
  name           = "test-snap"
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  volume_uri     = "/projects/p/providers/Aruba.Storage/volumes/v"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitBackupResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_backup" "test" {
  name           = "test-backup"
  location       = "ITBG-Bergamo"
  type           = "full"
  volume_id      = "vol-123"
  billing_period = "Hour"
  retention_days = 7
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitRestoreResource_MissingBackupID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_restore" "test" {
  name       = "test-restore"
  location   = "ITBG-Bergamo"
  project_id = "test-project"
}
`,
				ExpectError: regexp.MustCompile(`backup_id`),
			},
		},
	})
}

// --- Security ---

func TestUnitKmsResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_kms" "test" {
  name           = "test-kms"
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

// --- Management ---

func TestUnitSchedulejobResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_schedulejob" "test" {
  name     = "test-job"
  location = "ITBG-Bergamo"
  properties = {
    schedule_job_type = "OneShot"
    schedule_at       = "2099-12-31T23:59:59Z"
    steps             = []
  }
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

// --- Database ---

func TestUnitDbaasResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_dbaas" "test" {
  name     = "test-db"
  location = "ITBG-Bergamo"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitDatabaseResource_MissingDbaasID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_database" "test" {
  project_id = "test-project"
  name       = "testdb"
}
`,
				ExpectError: regexp.MustCompile(`dbaas_id`),
			},
		},
	})
}

func TestUnitDbaasuserResource_MissingDbaasID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_dbaasuser" "test" {
  project_id = "test-project"
  username   = "testuser"
  password   = "TestPass123!"
}
`,
				ExpectError: regexp.MustCompile(`dbaas_id`),
			},
		},
	})
}

// --- Container ---

func TestUnitKaasResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_kaas" "test" {
  name     = "test-kaas"
  location = "ITBG-Bergamo"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}

func TestUnitContainerregistryResource_MissingProjectID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "arubacloud_containerregistry" "test" {
  name     = "test-registry"
  location = "ITBG-Bergamo"
}
`,
				ExpectError: regexp.MustCompile(`project_id`),
			},
		},
	})
}
