# Step 5: Create Compute Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Keypair - SSH public key for server access
resource "arubacloud_keypair" "test" {
  name       = "test-keypair"
  location    = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCzZB11JRKjbPO/1wAtJ/9+/xQtndp61EWo1T2GhIVJO0eiBbUoufdhX989hAyE0JlyGjvDloe0c8S1sK8NAeLEx/jaKwsbHMQGxkusoBFUQDGWlREsHRHn7/78Wbra45ZJi6r9uizao7HDtoq0GCB6DfleOpKMLjOLHv9NaH0Hm119ZztHIqrmWmc25e27Evy3Nht9hX0Yb/OsEWcWBKhVv6SXGdB7SCXKYIPj7357bLpb4SdW9RxQA40bjlEFtPSqZ3HNXZ7yrUZXQWtrVkpia51nR088Jz0rMlmLgH+RPTDtj8CcI/E6QgsKXfrlxswbl3cT41qZVHi0+hNxE9vg+MSAVuYKgyWWFU7qlQCvmKmDPDjivBaFn7Aaz9qw71brpIeNXRwNiEbHy2+2+A0X8iIbc1Ca3RdVQ2rBLRXQDhNMi2syJkyty0ZTiLSNt+rhl4JgFZBz88q7b34MezNNNP7HX4oG+XpwjUe4KzDjk8EbBfxiPlLy7xkBioxRe+E="
}

# Local variables for server configuration
locals {
  server_name = "test-cloudserver"
  
  # Cloud-init configuration for automated nginx setup
  # Provide raw cloud-init YAML content (not base64-encoded)
  cloud_init_config = <<-EOF
    #cloud-config
    package_update: true
    package_upgrade: true
    packages:
      - nginx
      - curl
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
      - |
        cat > /var/www/html/index.html <<'HTML'
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>ArubaCloud CloudServer</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    max-width: 800px;
                    margin: 50px auto;
                    padding: 20px;
                    background-color: #f5f5f5;
                }
                .container {
                    background-color: white;
                    padding: 30px;
                    border-radius: 8px;
                    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                }
                h1 {
                    color: #0066cc;
                }
                .info {
                    margin: 10px 0;
                    padding: 10px;
                    background-color: #f0f8ff;
                    border-left: 4px solid #0066cc;
                }
            </style>
        </head>
        <body>
            <div class="container">
                <h1>🚀 Hello from ArubaCloud CloudServer!</h1>
                <div class="info">
                    <strong>Server:</strong> ${local.server_name}
                </div>
                <div class="info">
                    <strong>Provisioned with:</strong> cloud-init (user_data)
                </div>
                <p>✅ Nginx is running successfully!</p>
                <p><em>This page was automatically configured using cloud-init during instance creation.</em></p>
            </div>
        </body>
        </html>
        HTML
    write_files:
      - path: /etc/nginx/sites-available/default
        content: |
          server {
              listen 80 default_server;
              listen [::]:80 default_server;
              root /var/www/html;
              index index.html index.htm;
              server_name _;
              charset utf-8;
              location / {
                  try_files $uri $uri/ =404;
              }
          }
    final_message: "Cloud-init setup completed successfully. Nginx is ready!"
  EOF
}

# Cloud Server - Virtual machine instance
# Note: Uses nested network, settings, and storage blocks for better organization
# The boot_volume_uri_ref field should reference a bootable block storage URI (created with bootable=true and image set)
# Using cloud-init user_data for nginx installation (raw cloud-init YAML content)
resource "arubacloud_cloudserver" "test" {
  name       = local.server_name
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  zone       = "ITBG-1"  # Change to your zone
  tags       = ["compute", "test"]

  network = {
    vpc_uri_ref              = arubacloud_vpc.test.uri
    elastic_ip_uri_ref       = arubacloud_elasticip.test.uri
    subnet_uri_refs          = [arubacloud_subnet.test.uri]
    securitygroup_uri_refs   = [arubacloud_securitygroup.test.uri]
  }

  settings = {
    flavor_name      = "CSO4A8"  # 4 CPU, 8GB RAM (see https://api.arubacloud.com/docs/metadata/#cloudserver-flavors)
    key_pair_uri_ref = arubacloud_keypair.test.uri
    user_data        = local.cloud_init_config  # Cloud-init configuration (raw YAML content)
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.boot_disk.uri
    
    # Note: Additional data volumes (like arubacloud_blockstorage.data_disk) are created
    # in 06-storage.tf but cannot be attached during CloudServer creation.
    # Data volumes must be attached to the CloudServer through the ArubaCloud console
    # or API after the server is created, as the CloudServer API doesn't support
    # attaching data volumes at creation time.
  }
  
  # Ensure all security rules are created before the cloudserver
  depends_on = [
    arubacloud_securityrule.test,
    arubacloud_securityrule.ssh,
    arubacloud_securityrule.default_egress
  ]
}



