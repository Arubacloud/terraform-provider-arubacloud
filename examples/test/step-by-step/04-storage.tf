# Step 4: Create Storage Resources (Optional)
# Note: The boot volume can be specified directly in the cloud server using the image ID
# If you need additional storage volumes, create them here

# Optional: Additional Block Storage (if needed for data volumes)
# resource "arubacloud_blockstorage" "test" {
#   name          = "test-data-disk"
#   project_id    = arubacloud_project.test.id
#   location      = "ITBG-Bergamo"  # Change to your region
#   size_gb       = 50
#   billing_period = "Hour"
#   zone          = "ITBG-1"  # Change to your zone
#   type          = "Standard"
#   tags          = ["storage", "data", "test"]
# }

