# Step 5: Create Compute Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Keypair - SSH public key for server access
resource "arubacloud_keypair" "test" {
  name       = "test-keypair"
  location    = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEA2No7At0tgHrcZTL0kGWyLLUqPKfOhD9hGdNV9PbJxhjOGNFxcwdQ9wCXsJ3RQaRHBuGIgVodDurrlqzxFK86yCHMgXT2YLHF0j9P4m9GDiCfOK6msbFb89p5xZExjwD2zK+w68r7iOKZeRB2yrznW5TD3KDemSPIQQIVcyLF+yxft49HWBTI3PVQ4rBVOBJ2PdC9SAOf7CYnptW24CRrC0h85szIdwMA+Kmasfl3YGzk4MxheHrTO8C40aXXpieJ9S2VQA4VJAMRyAboptIK0cKjBYrbt5YkEL0AlyBGPIu6MPYr5K/MHyDunDi9yc7VYRYRR0f46MBOSqMUiGPnMw=="
}

# Cloud Server - Virtual machine instance
# Note: vpc_uri_ref, subnet_uri_refs, securitygroup_uri_refs, key_pair_uri_ref, and elastic_ip_uri_ref use URI references
# The boot_volume_uri_ref field should reference a bootable block storage URI (created with bootable=true and image set)
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

output "keypair_id" {
  value       = arubacloud_keypair.test.id
  description = "The ID of the created keypair"
}

output "cloudserver_id" {
  value       = arubacloud_cloudserver.test.id
  description = "The ID of the created cloud server"
}

output "cloudserver_public_ip" {
  value       = arubacloud_elasticip.test.address
  description = "The public IP address of the cloud server (from associated Elastic IP)"
}

