# Step 6: Provision CloudServer with Nginx
# Using null_resource to separate provisioning from CloudServer lifecycle
# This prevents CloudServer recreation if provisioning fails

resource "null_resource" "nginx_install" {
  # Re-run provisioner if CloudServer is replaced
  triggers = {
    cloudserver_id = arubacloud_cloudserver.test.id
  }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = arubacloud_elasticip.test.address
    private_key = file("~/.ssh/id_rsa")  # Adjust path to your SSH private key
    timeout     = "5m"
  }

  provisioner "remote-exec" {
    inline = [
      "sleep 30",  # Wait for cloud-init to complete
      "sudo apt-get update",
      "sudo apt-get install -y nginx",
      "sudo systemctl enable nginx",
      "sudo systemctl start nginx",
      "echo '<h1>Hello from ArubaCloud CloudServer!</h1>' | sudo tee /var/www/html/index.html",
      "echo '<p>Server: ${arubacloud_cloudserver.test.name}</p>' | sudo tee -a /var/www/html/index.html",
      "echo '<p>Public IP: ${arubacloud_elasticip.test.address}</p>' | sudo tee -a /var/www/html/index.html"
    ]
  }

  depends_on = [arubacloud_cloudserver.test]
}
