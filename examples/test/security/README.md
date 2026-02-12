# ArubaCloud Terraform Provider - Security Example

This example demonstrates encryption and key management services deployment with KMS and encryption keys.

## What This Example Does

This Terraform configuration:
1. Creates a project as the foundation
2. Sets up Key Management Service (KMS) for encryption
3. Creates encryption keys within the KMS

**Note:** KMIP (Key Management Interoperability Protocol) endpoint creation is currently commented out due to API validation requirements.

## Features

### Encryption & Key Management

#### KMS (Key Management Service)
- Managed key service for encryption operations
- Configured with hourly billing
- Provides centralized key management
- Location-based deployment (ITBG-Bergamo default)
- Tag support for resource organization (minimum 4 characters per tag)

#### Encryption Keys
- AES and RSA encryption algorithms supported
- Keys are managed within KMS
- Automatic key lifecycle management

#### KMIP (Currently Unavailable)
- KMIP endpoint creation returns "invalid status" validation error
- May require specific KMS configuration or feature enablement
- Resource definition is commented out in the example

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

The example follows this dependency chain:
```
Project
  └── KMS
       └── Key (encryption key)
       └── KMIP (commented out - validation issues)
```

## Outputs

The configuration provides outputs for:
- KMS instance ID and URI
- Key ID and URI (when created)

These outputs can be used as inputs for other Terraform modules or for reference.

## Testing

### Verify KMS Setup
```bash
terraform output kms_id
terraform output kms_uri
```

### Verify Key Creation
```bash
terraform output key_id
terraform output key_uri
```

## Cleanup

To destroy all resources:
```bash
terraform destroy
```

When prompted, type `yes` to confirm deletion.

## Common Issues

### Issue: KMS creation fails with "length must be at least 4 char"
**Solution**: Ensure all tags have at least 4 characters. Example: use "encryption" instead of "kms".

### Issue: KMIP creation fails with "invalid status"
**Solution**: KMIP endpoint creation currently has API validation issues. The feature may require specific KMS configuration or account-level enablement. The resource is commented out in this example.

### Issue: Key creation returns "no ID returned from API"
**Solution**: The provider will attempt to use the key name as ID if KeyID is not returned. Check the debug logs for more information.

### Issue: KMS not accessible immediately after creation
**Solution**: The provider waits for KMS to be in "Active" state before proceeding. If issues persist, check KMS state in the ArubaCloud console.

## Additional Resources

- [ArubaCloud API Documentation](https://api.arubacloud.com/docs/)
- [KMS Documentation](https://api.arubacloud.com/docs/resources/#kms)
- [Terraform Provider Documentation](../../docs/)

## Notes

- **SDK Version**: This example requires SDK v0.1.20 or later with KMS response parsing fixes
- KMS resources require proper billing setup in your ArubaCloud account
- Keys and encryption services are managed centrally through KMS
- KMS provides secure key management for encrypting your ArubaCloud resources
- The provider automatically waits for KMS to be active before creating child resources (Keys, KMIP)
- **Algorithm names**: Use "Aes" or "Rsa" (case-sensitive) for key algorithms
- **Immutable fields**: KMS `id` and `uri` are set once at creation and never change until resource deletion
