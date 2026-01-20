# Datasources and Import Testing Example

This example demonstrates how to:
1. Use datasources to read existing resources
2. Import existing resources into Terraform state
3. Verify datasources match created resources

## Prerequisites

- Terraform >= 1.0
- ArubaCloud account with API credentials
- ArubaCloud Terraform Provider installed locally

## Resources Created

- **Project**: Test project for organizing resources
- **KMS**: Key Management Service for encryption
- **Schedule Job**: Scheduled job for automation

## Files

- `00-variables.tf`: Input variables for API credentials
- `01-provider.tf`: Provider configuration
- `02-project.tf`: Project resource
- `03-resources.tf`: Test resources (KMS, Schedule Job)
- `04-datasources.tf`: Datasources to read the created resources
- `05-outputs.tf`: Output values comparing resources and datasources
- `terraform.tfvars`: Variable values (credentials)

## Usage

### 1. Configure Credentials

Edit `terraform.tfvars` with your ArubaCloud API credentials:

```hcl
arubacloud_api_key    = "your-api-key"
arubacloud_api_secret = "your-api-secret"
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Plan and Apply

```bash
terraform plan
terraform apply
```

This will:
- Create a project, KMS, and schedule job
- Read them back using datasources
- Output comparison results

### 4. Verify Datasources

Check the outputs to verify datasources match resources:

```bash
terraform output
```

You should see `*_match = true` for all comparisons.

## Testing Datasources

### Read Existing Resource by ID

After creating resources, you can test datasources independently:

```hcl
data "arubacloud_project" "existing" {
  id = "project-id-from-output"
}

output "existing_project_name" {
  value = data.arubacloud_project.existing.name
}
```

## Testing Import

### Import Existing Resources

You can import existing resources into Terraform state:

#### 1. Import Project

```bash
terraform import arubacloud_project.imported <project-id>
```

#### 2. Import KMS

```bash
terraform import arubacloud_kms.imported <kms-id>
```

#### 3. Import Schedule Job

```bash
terraform import arubacloud_schedulejob.imported <schedulejob-id>
```

### Import Example Workflow

1. Add resource definition without creating it:

```hcl
resource "arubacloud_kms" "imported" {
  name       = "existing-kms"
  project_id = "existing-project-id"
  location   = "ITBG-Bergamo"
  tags       = []
  properties = {
    algorithm = "AES256"
    key_size  = 256
  }
}
```

2. Import the existing resource:

```bash
terraform import arubacloud_kms.imported <kms-id>
```

3. Verify the import:

```bash
terraform plan
```

Should show no changes if the configuration matches the existing resource.

4. Update the configuration to match actual state:

```bash
terraform show
```

Copy the actual values from the state to your configuration.

## Import ID Format

Different resources use different import ID formats:

| Resource | Import ID Format | Example |
|----------|-----------------|---------|
| Project | `<project-id>` | `68398923fb2cb026400d4d31` |
| KMS | `<kms-id>` | `690083754e7d691466d86331` |
| Schedule Job | `<schedulejob-id>` | `690083754e7d691466d86332` |

## Testing Import with Datasource Validation

Best practice: Use datasources to get resource attributes before importing:

```bash
# Step 1: Use datasource to read existing resource
data "arubacloud_kms" "to_import" {
  id = "existing-kms-id"
}

output "kms_details" {
  value = data.arubacloud_kms.to_import
}

# Step 2: Apply to see the actual configuration
terraform apply

# Step 3: Copy the output values to your resource definition
resource "arubacloud_kms" "imported" {
  name       = data.arubacloud_kms.to_import.name
  project_id = data.arubacloud_kms.to_import.project_id
  location   = data.arubacloud_kms.to_import.location
  tags       = data.arubacloud_kms.to_import.tags
  properties = data.arubacloud_kms.to_import.properties
}

# Step 4: Import the resource
# terraform import arubacloud_kms.imported <kms-id>
```

## Common Issues

### Datasource Returns Null

If a datasource returns null values:
- Verify the resource ID is correct
- Ensure the resource exists in your account
- Check API credentials have proper permissions

### Import Shows Differences

If `terraform plan` shows changes after import:
- The configuration doesn't match the actual resource
- Update your configuration to match the actual state
- Use `terraform show` to see the imported state

### Import Fails

Common reasons for import failures:
- Invalid resource ID format
- Resource doesn't exist
- Insufficient permissions
- Wrong provider configuration

## Cleanup

To destroy all created resources:

```bash
terraform destroy
```

## Notes

- Datasources are read-only and don't create resources
- Import adds existing resources to Terraform state
- Always verify configuration matches imported state
- Use datasources before import to get accurate configuration
- Some resource attributes may not be importable (computed values)

## Next Steps

After testing datasources and import:
1. Try importing resources from your existing infrastructure
2. Use datasources to reference external resources
3. Combine datasources with resources for complex configurations
4. Practice disaster recovery by importing resources to new Terraform state
