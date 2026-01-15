# ArubaCloud Terraform Provider - Nginx Example

This example demonstrates a complete CloudServer deployment with automatic nginx installation and configuration.

## What This Example Does

This Terraform configuration:
1. Creates a project and all required network infrastructure (VPC, subnet, security groups)
2. Provisions a CloudServer with a bootable block storage volume
3. Automatically installs and configures nginx on the server
4. Exposes the web server via an Elastic IP address

## Features

### Automatic Nginx Installation
The example uses **cloud-init** (`user_data`) to automatically:
- Update package lists and upgrade system packages
- Install nginx and curl
- Enable and start nginx service
- Create a custom welcome page with server information
- Configure nginx with UTF-8 charset support for proper character encoding

The cloud-init configuration includes:
- **UTF-8 charset** in both the HTML meta tag and nginx configuration for proper emoji and special character display
- **Proper HTML structure** with semantic markup
- **Automated service management** via systemd

### Network Security
Security rules configured to allow:
- **Port 80 (HTTP)**: For nginx web server access
- **Port 22 (SSH)**: For provisioner connection and server management

### Helpful Outputs
After deployment, Terraform provides:
- `nginx_test_command`: The exact curl command to test nginx
- `nginx_url`: The direct URL to access nginx in your browser
- `cloudserver_public_ip`: The public IP address

## Prerequisites

### 1. Credentials
Create `terraform.tfvars` with your ArubaCloud credentials:
```hcl
arubacloud_api_key    = "your-api-key"
arubacloud_api_secret = "your-api-secret"
```

### 2. SSH Key (Optional)
If you want to SSH into the server after deployment, make sure you have an SSH private key that corresponds to the public key in the keypair resource (defined in `05-compute.tf`).
- Update the public key value in `05-compute.tf` to match your SSH public key

### 4. Provider Binary
Build the provider if not already built:
```bash
cd /path/to/terraform-provider-arubacloud
go build -o terraform-provider-arubacloud
```

## Quick Start

1. **Set up the example directory**:
   ```bash
   cd examples/test/nginx
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

6. **Test nginx** after deployment completes:
   ```bash
   # Get the curl command
   terraform output nginx_test_command
   
   # Execute the curl command (example)
   curl http://203.0.113.45:80
   
   # Or get the URL for your browser
   terraform output nginx_url
   ```

## Expected Output

When you curl or visit the URL, you should see a nicely formatted HTML page with:
- 🚀 Hello from ArubaCloud CloudServer!
- Server name: test-cloudserver
- Provisioned with: cloud-init (user_data)
- ✅ Nginx is running successfully!

The page uses UTF-8 encoding, so emojis and special characters display correctly.

## File Structure

- `00-variables.tf` - Input variable declarations for API credentials
- `01-provider.tf` - Provider configuration with credentials
- `02-project.tf` - Project resource
- `03-main.tf` - Terraform required providers configuration
- `04-network.tf` - VPC, Subnet, Security Groups, Elastic IP, Security Rules
- `05-compute.tf` - Keypair and CloudServer resources with cloud-init user_data
- `06-storage.tf` - Block storage volumes (boot and data disks)
- `07-output.tf` - Output definitions (elastic IP, nginx URL, test command)
- `terraform.tfvars` - Your credentials (create this file)

## Configuration Details

### Step 0: Variables
Defines input variables for the ArubaCloud API credentials (`arubacloud_api_key` and `arubacloud_api_secret`), marked as sensitive.

### Step 1: Provider Configuration
Configure the ArubaCloud provider with your API credentials and default settings, using the variables defined in `00-variables.tf`.

### Step 2: Project
Creates an ArubaCloud project that will contain all resources.

### Step 3: Main Configuration
Specifies required Terraform providers:
- `arubacloud` provider (local development version)

### Step 4: Network Resources
Creates:
- VPC (Virtual Private Cloud)
- Subnet with DHCP configuration
- Security Group
- Elastic IP (public IP address)
- Security Rules (HTTP on port 80, SSH on port 22)

### Step 5: Compute
Creates:
- SSH keypair for server access
- CloudServer (4 CPU, 8GB RAM) with cloud-init user_data for automated nginx installation

The cloud-init configuration (`user_data`) automatically:
- Updates and upgrades system packages
- Installs nginx and curl
- Configures nginx with UTF-8 charset support
- Creates a welcome page with proper HTML structure and UTF-8 encoding
- Enables and starts the nginx service

**Best Practice**: The cloud-init configuration includes UTF-8 charset settings in both the HTML meta tag (`<meta charset="UTF-8">`) and nginx configuration (`charset utf-8;`) to ensure proper display of emojis and special characters.

### Step 6 (Storage): Storage Resources
Creates:
- Boot disk (100GB, bootable, with Ubuntu 22.04 image)
- Data disk (50GB, non-bootable)

### Step 7: Outputs
Defines output values that are displayed after `terraform apply`:
- `elastic_ip_address`: The public IP address assigned to the server
- `nginx_test_command`: Ready-to-use curl command to test nginx
- `nginx_url`: Direct URL to access in your browser

## Customization

### Change the Welcome Page
Edit the HTML content in the `cloud_init_config` local variable in `05-compute.tf`:
```hcl
cloud_init_config = <<-EOF
  #cloud-config
  runcmd:
    - |
      cat > /var/www/html/index.html <<'HTML'
      <!DOCTYPE html>
      <html lang="en">
      <head>
          <meta charset="UTF-8">
          <title>Your Custom Page</title>
      </head>
      <body>
          <h1>Your Custom Message!</h1>
      </body>
      </html>
      HTML
EOF
```

**Important**: Always include `<meta charset="UTF-8">` in your HTML for proper character encoding.

### Install Additional Packages
Add packages to the `packages` list in the `cloud_init_config`:
```hcl
packages:
  - nginx
  - curl
  - git
  - nodejs
  - npm
```

### Deploy Your Application
Add commands to the `runcmd` section in `05-compute.tf`:
```hcl
runcmd:
  - git clone https://github.com/your/repo.git /var/www/html
  - systemctl enable nginx
  - systemctl start nginx
```

### Configure Nginx with UTF-8 Support
When creating nginx configuration files via `write_files`, always include `charset utf-8;`:
```hcl
write_files:
  - path: /etc/nginx/sites-available/default
    content: |
      server {
          listen 80;
          charset utf-8;  # Important for proper character encoding
          root /var/www/html;
          # ... rest of config
      }
```

### Change Region/Zone
Update the `location` and `zone` values in all `.tf` files:
- Current: `ITBG-Bergamo` and `ITBG-1`
- Replace with your target region

## Troubleshooting

### Connection Timeout
- Ensure port 22 (SSH) security rule is applied
- Check that your SSH key matches the public key in the keypair resource
- Verify the CloudServer has network connectivity
- Check if the Elastic IP is properly associated

### Package Installation Fails
- Check cloud-init logs: `sudo cat /var/log/cloud-init-output.log`
- Verify the OS image supports apt-get (Debian/Ubuntu)
- For CentOS/RHEL, update the cloud-init config to use `yum`:
  ```yaml
  packages:
    - nginx
  runcmd:
    - yum update -y
    - yum install -y nginx
  ```

### Cannot Access Nginx
- Ensure port 80 (HTTP) security rule is applied and active
- Check cloud-init status: `cloud-init status` (should show "done")
- Check nginx is running: `sudo systemctl status nginx`
- Verify the elastic IP is correctly associated with the CloudServer
- Check your local firewall isn't blocking outbound port 80
- View cloud-init logs: `sudo cat /var/log/cloud-init-output.log`

### CloudServer Takes Too Long
- The provider automatically waits for resources to become active
- Default timeout is 10 minutes (configurable in provider block)
- Check the ArubaCloud console to see the server status

## Important Notes

1. **Resource Dependencies**: Resources automatically wait for dependencies to be active
2. **Timeouts**: Default timeout is 10 minutes, configurable in the provider block
3. **Cost**: This example provisions paid resources. Remember to destroy when done testing
4. **Cloud-init**: The server is automatically configured via cloud-init. No SSH access is required for provisioning
5. **UTF-8 Encoding**: The example includes UTF-8 charset settings for proper display of emojis and special characters

## Cleanup

```bash
# Destroy all resources
terraform destroy

# Or destroy specific resources
terraform destroy -target=arubacloud_cloudserver.test
terraform destroy -target=arubacloud_project.test
```

**Note**: Destroy operations may take a few minutes as Terraform waits for resources to be properly deleted.

## Additional Resources

- [ArubaCloud API Documentation](https://api.arubacloud.com/docs/)
- [CloudServer Flavors](https://api.arubacloud.com/docs/metadata/#cloudserver-flavors)
- [Provider Documentation](../../../docs/)
