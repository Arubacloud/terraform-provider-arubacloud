# Step 3: Create Storage Resources

# Block Storage - Storage disk for container registry
resource "arubacloud_blockstorage" "container_registry" {
  name          = "container-registry-storage"
  project_id    = arubacloud_project.test.id
  location      = "ITBG-Bergamo"  # Change to your region
  size_gb       = 100
  billing_period = "Hour"
  zone          = "ITBG-1"  # Change to your zone
  type          = "Standard"
  tags          = ["storage", "container", "registry", "test"]
}
