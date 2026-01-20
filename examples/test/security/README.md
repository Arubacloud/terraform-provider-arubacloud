# ArubaCloud Terraform Provider - Security Example

This example demonstrates encryption and key management services deployment with KMS.

## What This Example Does

This Terraform configuration:
1. Creates a project as the foundation
2. Sets up Key Management Service (KMS) for encryption

## Features

### Encryption & Key Management

#### KMS (Key Management Service)
- Managed key service for encryption operations
- Configured with hourly billing
- Provides centralized key management

## Prerequisites

### 1. Credentials
Create `terraform.tfvars` with your ArubaCloud credentials:
```hcl
arubacloud_api_key    = "your-api-key"
arubacloud_api_secret = "your-api-secret"
```

### 2. Provider Binary
Build the provider if not already built:
```bash
cd /path/to/terraform-provider-arubacloud
go build -o terraform-provider-arubacloud
```

## Quick Start

### 1. Initialize Terraform
```bash
terraform init
```

### 2. Review the Execution Plan
```bash
terraform plan
```

### 3. Apply the Configuration
```bash
terraform apply
```

When prompted, type `yes` to confirm.

### 4. View Outputs
After successful deployment:
```bash
terraform output
```

## Configuration Details

### Location and Region
The example uses `ITBG-Bergamo` as the default location. Change this to match your target region:
```hcl
location = "ITBG-Bergamo"  # Change to your region
```

Common locations:
- `ITBG-Bergamo` - Italy, Bergamo
- `CZTX-Prague` - Czech Republic, Prague
- `PLWZ-Warsaw` - Poland, Warsaw

### Billing Period
KMS supports hourly billing for flexible cost management:
```hcl
billing_period = "Hour"
```

## Resource Dependencies

The example follows this simple dependency chain:
```
Project
  └── KMS
```

## Outputs

The configuration provides outputs for:
- KMS instance ID and URI

These outputs can be used as inputs for other Terraform modules or for reference.

## Testing

### Verify KMS Setup
```bash
terraform output kms_id
terraform output kms_uri
```

## Cleanup

To destroy all resources:
```bash
terraform destroy
```

When prompted, type `yes` to confirm deletion.

## Common Issues

### Issue: KMS creation fails
**Solution**: Ensure you have proper permissions and billing setup for encryption services in your ArubaCloud account.

## Additional Resources

- [ArubaCloud API Documentation](https://api.arubacloud.com/docs/)
- [KMS Documentation](https://api.arubacloud.com/docs/resources/#kms)
- [Terraform Provider Documentation](../../docs/)

## Notes

- KMS resources require proper billing setup in your ArubaCloud account
- Keys and encryption services are managed centrally through KMS
- KMS provides secure key management for encrypting your ArubaCloud resources
