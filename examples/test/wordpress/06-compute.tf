# Step 5: Create VM and bootstrap WordPress with cloud-init

resource "arubacloud_keypair" "wordpress" {
  name       = "wp-keypair"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCzZB11JRKjbPO/1wAtJ/9+/xQtndp61EWo1T2GhIVJO0eiBbUoufdhX989hAyE0JlyGjvDloe0c8S1sK8NAeLEx/jaKwsbHMQGxkusoBFUQDGWlREsHRHn7/78Wbra45ZJi6r9uizao7HDtoq0GCB6DfleOpKMLjOLHv9NaH0Hm119ZztHIqrmWmc25e27Evy3Nht9hX0Yb/OsEWcWBKhVv6SXGdB7SCXKYIPj7357bLpb4SdW9RxQA40bjlEFtPSqZ3HNXZ7yrUZXQWtrVkpia51nR088Jz0rMlmLgH+RPTDtj8CcI/E6QgsKXfrlxswbl3cT41qZVHi0+hNxE9vg+MSAVuYKgyWWFU7qlQCvmKmDPDjivBaFn7Aaz9qw71brpIeNXRwNiEbHy2+2+A0X8iIbc1Ca3RdVQ2rBLRXQDhNMi2syJkyty0ZTiLSNt+rhl4JgFZBz88q7b34MezNNNP7HX4oG+XpwjUe4KzDjk8EbBfxiPlLy7xkBioxRe+E="
}

locals {
  vm_name       = "wp-vm"
  wordpress_url = arubacloud_elasticip.vm.address != null ? format("http://%s", arubacloud_elasticip.vm.address) : "http://<vm-elastic-ip>"

  cloud_init_config = <<-EOF
    #cloud-config
    package_update: true
    package_upgrade: true
    packages:
      - apache2
      - libapache2-mod-php
      - php
      - php-mysql
      - php-xml
      - php-mbstring
      - mysql-client
      - rsync
      - wget
      - tar
      - unzip
    runcmd:
      - systemctl enable apache2
      - systemctl start apache2
      - cd /tmp
      - wget https://wordpress.org/latest.tar.gz
      - tar -xzf latest.tar.gz
      - rsync -a /tmp/wordpress/ /var/www/html/
      - rm -f /var/www/html/index.html
      - chown -R www-data:www-data /var/www/html
      - chmod -R 755 /var/www/html
      # Wait until MySQL accepts TCP connections (DBaaS may still be starting when cloud-init runs)
      - |
        DBHOST="${arubacloud_elasticip.dbaas.address}"
        for i in $(seq 1 90); do
          (echo >/dev/tcp/$DBHOST/3306) >/dev/null 2>&1 && exit 0
          sleep 10
        done
        exit 1
      - cp /var/www/html/wp-config-sample.php /var/www/html/wp-config.php
      - sed -i "s/database_name_here/wordpress/g" /var/www/html/wp-config.php
      - sed -i "s/username_here/wordpress/g" /var/www/html/wp-config.php
      - sed -i "s|password_here|${replace(var.database_password, "|", "\\|")}|g" /var/www/html/wp-config.php
      - sed -i "s/localhost/${arubacloud_elasticip.dbaas.address}/g" /var/www/html/wp-config.php
      - |
        cat > /var/www/html/wp-cli-install.sh <<'SCRIPT'
        #!/usr/bin/env bash
        set -euo pipefail
        cd /var/www/html
        curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar
        chmod +x wp-cli.phar
        mv wp-cli.phar /usr/local/bin/wp
        sudo -u www-data wp core install \
          --url="${local.wordpress_url}" \
          --title="ArubaCloud WordPress" \
          --admin_user="admin" \
          --admin_password="${var.wordpress_admin_password}" \
          --admin_email="admin@example.com" \
          --skip-email || true
        SCRIPT
      - chmod +x /var/www/html/wp-cli-install.sh
      - /var/www/html/wp-cli-install.sh
      - systemctl restart apache2
    final_message: "WordPress bootstrap completed."
  EOF
}

resource "arubacloud_cloudserver" "wordpress" {
  name       = local.vm_name
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  zone       = "ITBG-1"
  tags       = ["compute", "virtual-machine", "wordpress", "test"]

  network = {
    vpc_uri_ref            = arubacloud_vpc.test.uri
    elastic_ip_uri_ref     = arubacloud_elasticip.vm.uri
    subnet_uri_refs        = [arubacloud_subnet.test.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.vm.uri]
  }

  settings = {
    flavor_name      = "CSO4A8"
    key_pair_uri_ref = arubacloud_keypair.wordpress.uri
    user_data        = local.cloud_init_config
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.boot_disk.uri
  }

  depends_on = [
    arubacloud_securityrule.vm_http,
    arubacloud_securityrule.vm_ssh,
    arubacloud_securityrule.vm_egress,
    arubacloud_securityrule.dbaas_mysql,
    arubacloud_securityrule.dbaas_egress,
    arubacloud_dbaas.wordpress,
    arubacloud_database.wordpress,
    arubacloud_dbaasuser.wordpress,
    arubacloud_databasegrant.wordpress
  ]
}
