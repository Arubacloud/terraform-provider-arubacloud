# Step 3: Create boot disk for the VM
resource "arubacloud_blockstorage" "boot_disk" {
  name           = "wp-boot-disk"
  project_id     = arubacloud_project.test.id
  location       = "ITBG-Bergamo"
  size_gb        = 100
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Performance"
  bootable       = true
  image          = "LU22-001"
  tags           = ["storage", "boot", "wordpress", "test"]
}
