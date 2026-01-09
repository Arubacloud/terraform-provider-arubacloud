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
The CloudServer configuration includes a `remote-exec` provisioner that:
- Waits for cloud-init to complete
- Updates package lists
- Installs nginx
- Enables and starts nginx service
- Creates a custom welcome page with server information

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

### 2. SSH Key
Make sure you have an SSH private key at `~/.ssh/id_rsa` that corresponds to the public key in the keypair resource (defined in `05-compute.tf`).
- Update the path in the `connection` block if your key is elsewhere
- The public key in the example is: `ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEA2No7At0t...`

### 3. OS User
The provisioner uses `root` user by default.
- Update the `user` field in the `connection` block if your image uses a different default user (e.g., `ubuntu`, `debian`, `centos`)

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

When you curl or visit the URL, you should see:
```html
<h1>Hello from ArubaCloud CloudServer!</h1>
<p>Server: test-cloudserver</p>
<p>Public IP: <your-elastic-ip></p>
```

## File Structure

- `01-provider.tf` - Provider configuration with credentials
- `02-project.tf` - Project resource
- `03-network.tf` - VPC, Subnet, Security Groups, Elastic IP, Security Rules
- `04-storage.tf` - Block storage volumes (boot and data disks)
- `05-compute.tf` - Keypair and CloudServer with nginx provisioner
- `terraform.tfvars` - Your credentials (create this file)

## Configuration Details

### Step 1: Provider Configuration
Configure the ArubaCloud provider with your API credentials and default settings.

### Step 2: Project
Creates an ArubaCloud project that will contain all resources.

### Step 3: Network Resources
Creates:
- VPC (Virtual Private Cloud)
- Subnet with DHCP configuration
- Security Group
- Elastic IP (public IP address)
- Security Rules (HTTP on port 80, SSH on port 22)

### Step 4: Storage
Creates:
- Boot disk (100GB, bootable, with Ubuntu 22.04 image)
- Data disk (50GB, non-bootable)

### Step 5: Compute
Creates:
- SSH keypair for server access
- CloudServer (4 CPU, 8GB RAM)
- Provisioner to install nginx automatically

## Customization

### Change the Welcome Page
Edit the `echo` commands in `05-compute.tf`:
```hcl
"echo '<h1>Your Custom Message!</h1>' > /var/www/html/index.html",
```

### Install Additional Packages
Add more `apt-get install` commands:
```hcl
"apt-get install -y nginx git nodejs npm",
```

### Deploy Your Application
Add commands to clone and deploy:
```hcl
provisioner "remote-exec" {
  inline = [
    "sleep 30",
    "apt-get update",
    "apt-get install -y nginx git",
    "git clone https://github.com/your/repo.git /var/www/html",
    "systemctl enable nginx",
    "systemctl start nginx"
  ]
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
- Increase the sleep time if cloud-init needs more time (change from 30 to 60 seconds)
- Check the OS image supports apt-get (Debian/Ubuntu)
- For CentOS/RHEL, replace `apt-get` with `yum`:
  ```hcl
  "yum update -y",
  "yum install -y nginx",
  ```

### Cannot Access Nginx
- Ensure port 80 (HTTP) security rule is applied and active
- Check nginx is running: `ssh root@<elastic-ip> systemctl status nginx`
- Verify the elastic IP is correctly associated with the CloudServer
- Check your local firewall isn't blocking outbound port 80

### CloudServer Takes Too Long
- The provider automatically waits for resources to become active
- Default timeout is 10 minutes (configurable in provider block)
- Check the ArubaCloud console to see the server status

## Important Notes

1. **Resource Dependencies**: Resources automatically wait for dependencies to be active
2. **Timeouts**: Default timeout is 10 minutes, configurable in the provider block
3. **Cost**: This example provisions paid resources. Remember to destroy when done testing
4. **SSH Access**: The provisioner requires SSH access to work. The server must be accessible from your machine

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
