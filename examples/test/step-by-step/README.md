# Step-by-Step Testing Examples

This directory contains incremental examples for testing the Terraform provider.

## Quick Start

1. **Copy files to a test directory**:
   ```bash
   mkdir -p ~/terraform-test
   cd ~/terraform-test
   cp examples/test/step-by-step/*.tf .
   ```

2. **Set up credentials**:
   Create `terraform.tfvars`:
   ```hcl
   arubacloud_api_key    = "your-api-key"
   arubacloud_api_secret = "your-api-secret"
   ```

3. **Build the provider** (if not already built):
   ```bash
   cd /path/to/terraform-provider-arubacloud
   go build -o terraform-provider-arubacloud
   ```

4. **Initialize Terraform**:
   ```bash
   # Option 1: Use local provider binary
   terraform init -plugin-dir=../terraform-provider-arubacloud
   
   # Option 2: Install to default location
   mkdir -p ~/.terraform.d/plugins/hashicorp.com/arubacloud/arubacloud/1.0.0/darwin_arm64
   cp ../terraform-provider-arubacloud/terraform-provider-arubacloud \
      ~/.terraform.d/plugins/hashicorp.com/arubacloud/arubacloud/1.0.0/darwin_arm64/
   terraform init
   ```

## Testing Steps

### Step 1: Provider Configuration
- File: `01-provider.tf`
- Test: `terraform init`

### Step 2: Create Project
- File: `02-project.tf`
- Test: `terraform plan` then `terraform apply`
- Verify: `terraform state show arubacloud_project.test`

### Step 3: Create Network Resources
- File: `03-network.tf`
- Test: `terraform apply`
- Note: Resources will wait for dependencies to be active

### Continue with more steps...
See `TESTING_GUIDE.md` in the root directory for complete step-by-step instructions.

## Important Notes

1. **Adjust locations**: Change `ITBG-Bergamo` to your actual ArubaCloud region
2. **Adjust zones**: Change `ITBG-1` to your actual zone
3. **Wait mechanism**: Resources automatically wait for dependencies to be active
4. **Timeout**: Default is 10 minutes, configurable in provider block

## Cleanup

```bash
# Destroy all resources
terraform destroy

# Or destroy specific resource
terraform destroy -target=arubacloud_project.test
```

