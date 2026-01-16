# ArubaCloud Terraform Provider - Database Example

This example demonstrates a complete Database as a Service (DBaaS) deployment with MySQL, including database creation, user management, and network configuration.

## What This Example Does

This Terraform configuration:
1. Creates a project and all required network infrastructure (VPC, subnet, security groups)
2. Provisions a DBaaS instance with MySQL 8.0
3. Creates a database within the DBaaS instance
4. Creates a database user with authentication
5. Configures network security rules for MySQL access
6. Associates an Elastic IP for public access

## Features

### Managed Database Service
The example provisions a fully managed MySQL database instance with:
- **MySQL 8.0** engine (configurable via `engine_id`)
- **DBO2A4 flavor**: 2 CPU, 4GB RAM (see [DBaaS flavors documentation](https://api.arubacloud.com/docs/metadata/#dbaas-flavors))
- **Automatic scaling**: Enabled with 100GB available space and 10GB step size
- **Elastic IP**: Public IP address for external access

### Database Management
- **Database creation**: Creates a database named `testdb` within the DBaaS instance
- **User management**: Creates a database user `dbuser` with password authentication
- **Security**: Database access controlled via security groups

### Network Security
Security rules configured to allow:
- **Port 3306 (MySQL)**: Inbound access from anywhere (0.0.0.0/0) - **Note**: In production, restrict this to specific IPs
- **Egress traffic**: All outbound traffic allowed for database operations

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

1. **Set up the example directory**:
   ```bash
   cd examples/test/database
   ```

2. **Create terraform.tfvars** with your credentials (see Prerequisites above)

3. **Initialize Terraform**:
   ```bash
   terraform init
   ```

4. **Review the plan**:
   ```bash
   terraform plan
   ```

5. **Apply the configuration**:
   ```bash
   terraform apply
   ```

6. **Get connection information** after deployment completes:
   ```bash
   # Get the Elastic IP address
   terraform output dbaas_elastic_ip
   
   # Get the database name
   terraform output database_name
   
   # Get the DBaaS ID
   terraform output dbaas_id
   ```

## Expected Output

After deployment, you'll have:
- A running MySQL 8.0 database instance
- A database named `testdb`
- A database user `dbuser` with the configured password
- An Elastic IP address for public access
- Network security configured for MySQL access

## File Structure

- `00-variables.tf` - Input variable declarations for API credentials
- `01-provider.tf` - Provider configuration with credentials
- `02-project.tf` - Project resource
- `03-network.tf` - VPC, Subnet, Security Groups, Elastic IP, Security Rules
- `04-database.tf` - DBaaS, Database, and DBaaS User resources
- `05-output.tf` - Output definitions (DBaaS ID, database name, Elastic IP, etc.)

## Configuration Details

### Step 0: Variables
Defines input variables for the ArubaCloud API credentials (`arubacloud_api_key` and `arubacloud_api_secret`), marked as sensitive.

### Step 1: Provider Configuration
Configure the ArubaCloud provider with your API credentials and default settings, using the variables defined in `00-variables.tf`.

### Step 2: Project
Creates an ArubaCloud project that will contain all resources.

### Step 3: Network Resources
Creates:
- VPC (Virtual Private Cloud)
- Subnet with Basic type
- Security Group for DBaaS
- Elastic IP (public IP address)
- Security Rules:
  - MySQL ingress on port 3306 from 0.0.0.0/0
  - Egress traffic for all outbound connections

### Step 4: Database Resources
Creates:
- **DBaaS instance**: Managed MySQL 8.0 database with storage and autoscaling configuration
- **Database**: A database named `testdb` within the DBaaS instance
- **DBaaS User**: A user `dbuser` with password authentication
- **Database Grant**: Currently commented out due to provider limitations (see notes below)

## Connecting to the Database

After deployment, you can connect to your MySQL database using:

```bash
mysql -h <elastic-ip> -u dbuser -p
```

You'll be prompted for the password configured in `04-database.tf` (default: `supersecretpassword123!`).

**Note**: In production, use a secure password stored in a secret management system or Terraform variables.

## Customization

### Change Database Engine
Update the `engine_id` in `04-database.tf`:
```hcl
engine_id = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
```

### Change Database Flavor
Update the `flavor` in `04-database.tf`:
```hcl
flavor = "DBO2A4"  # See https://api.arubacloud.com/docs/metadata/#dbaas-flavors
```

Available flavors include:
- `DBO2A4`: 2 CPU, 4GB RAM
- `DBO4A8`: 4 CPU, 8GB RAM
- And more (check the documentation)

### Change Storage Size
Update the `storage.size_gb` in `04-database.tf`:
```hcl
storage = {
  size_gb = 200  # Storage size in GB
  # ... autoscaling configuration
}
```

### Configure Storage and Autoscaling
Modify the `storage` block in `04-database.tf`:
```hcl
storage = {
  size_gb = 200  # Storage size in GB
  autoscaling = {
    enabled         = true
    available_space = 200  # GB
    step_size       = 20   # GB
  }
}
```

**Note**: The `autoscaling` block is optional within `storage`. If omitted, autoscaling will be disabled.

### Restrict Database Access
Update the security rule in `03-network.tf` to restrict MySQL access to specific IPs:
```hcl
target = {
  kind  = "Ip"
  value = "203.0.113.0/24"  # Your office IP range
}
```

### Change Database Name

**Important:** Databases **cannot be updated**. To change a database's name, you must delete the existing database and create a new one.

**To delete a database:**
```bash
terraform destroy -target=arubacloud_database.test
```

**To create a new database with a different name:**
```hcl
resource "arubacloud_database" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  name       = "myappdb"
}
```

### Change User Credentials

**Important:** DBaaS users **cannot be updated**. To change a user's password or username, you must delete the existing user and create a new one.

**To delete a user:**
```bash
terraform destroy -target=arubacloud_dbaasuser.test
```

**To create a new user with different credentials:**
```hcl
resource "arubacloud_dbaasuser" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  username   = "myuser"
  password   = base64encode(var.database_password)  # Password must be 8-20 chars with number, uppercase, lowercase, and special char.
}
```

**Password Requirements:**
- 8-20 characters
- At least one number
- At least one uppercase letter
- At least one lowercase letter
- At least one special character
- No spaces allowed

The password must be base64 encoded using the `base64encode()` function in Terraform before passing to the provider.

## Important Notes

### Database Grant Resource
The `arubacloud_databasegrant` resource is currently commented out in `04-database.tf` due to provider limitations with GrantRole type conversion. Once this is resolved, you can uncomment it to associate the user with the database and grant permissions:

```hcl
resource "arubacloud_databasegrant" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  database   = arubacloud_database.test.id
  user_id    = arubacloud_dbaasuser.test.id
  role       = "admin"  # Role: read, write, or admin
}
```

### Security Considerations
1. **Password Security**: Never commit passwords to version control. Use Terraform variables or secret management systems.
2. **Network Access**: The example allows MySQL access from anywhere (0.0.0.0/0). In production, restrict this to specific IP ranges.
3. **SSL/TLS**: Consider enabling SSL/TLS for database connections in production environments.

### Resource Dependencies
- Resources automatically wait for dependencies to be active
- Default timeout is 10 minutes, configurable in the provider block
- DBaaS creation may take several minutes

### Cost
This example provisions paid resources:
- DBaaS instance (hourly billing)
- Elastic IP (hourly billing)
- Storage (based on autoscaling configuration)

Remember to destroy resources when done testing:
```bash
terraform destroy
```

## Troubleshooting

### Database Creation Fails
- Check the DBaaS engine ID is valid: https://api.arubacloud.com/docs/metadata/#dbaas-engines
- Verify the flavor is available: https://api.arubacloud.com/docs/metadata/#dbaas-flavors
- Ensure the project has sufficient quota

### Cannot Connect to Database
- Verify the Elastic IP is correctly associated with the DBaaS instance
- Check security rules allow port 3306 from your IP address
- Verify the database user credentials are correct
- Check if the DBaaS instance is in "Active" state

### Database Grant Issues
- The `arubacloud_databasegrant` resource is currently disabled due to provider limitations
- You may need to manually grant permissions via the ArubaCloud console or API

### Resource Timeout
- DBaaS creation can take 10-15 minutes
- Increase the timeout in the provider block if needed:
  ```hcl
  provider "arubacloud" {
    # ...
    timeout = "20m"
  }
  ```

## Cleanup

```bash
# Destroy all resources
terraform destroy

# Or destroy specific resources
terraform destroy -target=arubacloud_dbaas.test
terraform destroy -target=arubacloud_project.test
```

**Note**: Destroy operations may take a few minutes as Terraform waits for resources to be properly deleted.

## Additional Resources

- [ArubaCloud API Documentation](https://api.arubacloud.com/docs/)
- [DBaaS Engines](https://api.arubacloud.com/docs/metadata/#dbaas-engines)
- [DBaaS Flavors](https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
- [Provider Documentation](../../../docs/)
