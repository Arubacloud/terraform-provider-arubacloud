# Step 5: Create Compute Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Keypair - SSH public key for server access
resource "arubacloud_keypair" "test" {
  name       = "test-keypair"
  location    = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCzZB11JRKjbPO/1wAtJ/9+/xQtndp61EWo1T2GhIVJO0eiBbUoufdhX989hAyE0JlyGjvDloe0c8S1sK8NAeLEx/jaKwsbHMQGxkusoBFUQDGWlREsHRHn7/78Wbra45ZJi6r9uizao7HDtoq0GCB6DfleOpKMLjOLHv9NaH0Hm119ZztHIqrmWmc25e27Evy3Nht9hX0Yb/OsEWcWBKhVv6SXGdB7SCXKYIPj7357bLpb4SdW9RxQA40bjlEFtPSqZ3HNXZ7yrUZXQWtrVkpia51nR088Jz0rMlmLgH+RPTDtj8CcI/E6QgsKXfrlxswbl3cT41qZVHi0+hNxE9vg+MSAVuYKgyWWFU7qlQCvmKmDPDjivBaFn7Aaz9qw71brpIeNXRwNiEbHy2+2+A0X8iIbc1Ca3RdVQ2rBLRXQDhNMi2syJkyty0ZTiLSNt+rhl4JgFZBz88q7b34MezNNNP7HX4oG+XpwjUe4KzDjk8EbBfxiPlLy7xkBioxRe+E="
}

# Cloud Server - Virtual machine instance
# Note: vpc_uri_ref, subnet_uri_refs, securitygroup_uri_refs, key_pair_uri_ref, and elastic_ip_uri_ref use URI references
# The boot_volume_uri_ref field should reference a bootable block storage URI (created with bootable=true and image set)
# Note: Provisioner has been moved to 06-provisioning.tf to separate it from the CloudServer lifecycle
resource "arubacloud_cloudserver" "test" {
  name                  = "test-cloudserver"
  location              = "ITBG-Bergamo"  # Change to your region
  project_id            = arubacloud_project.test.id
  zone                  = "ITBG-1"  # Change to your zone
  vpc_uri_ref           = arubacloud_vpc.test.uri                    # URI reference
  flavor_name           = "CSO4A8"  # 4 CPU, 8GB RAM (see https://api.arubacloud.com/docs/metadata/#cloudserver-flavors)
  elastic_ip_uri_ref    = arubacloud_elasticip.test.uri              # URI reference
  boot_volume_uri_ref   = arubacloud_blockstorage.boot_disk.uri  # URI reference to bootable block storage
  key_pair_uri_ref      = arubacloud_keypair.test.uri                # URI reference
  subnet_uri_refs       = [arubacloud_subnet.test.uri]               # URI reference
  securitygroup_uri_refs = [arubacloud_securitygroup.test.uri]        # URI reference
  tags                  = ["compute", "test"]
}

#output "keypair_id" {
#  value       = arubacloud_keypair.test.id
#  description = "The ID of the created keypair"
#}

#output "cloudserver_id" {
#  value       = arubacloud_cloudserver.test.id
#  description = "The ID of the created cloud server"
#}

#output "cloudserver_public_ip" {
#  value       = arubacloud_elasticip.test.address
#  description = "The public IP address of the cloud server (from associated Elastic IP)"
#}

output "nginx_test_command" {
  value       = "curl http://${arubacloud_elasticip.test.address}:80"
  description = "Command to test nginx on the cloud server"
}

output "nginx_url" {
  value       = "http://${arubacloud_elasticip.test.address}"
  description = "URL to access nginx on the cloud server"
}

