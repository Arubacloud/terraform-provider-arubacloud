# Step 2: Create shared network resources (VM + DB)

resource "arubacloud_vpc" "test" {
  name       = "wp-test-vpc"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  tags       = ["network", "wordpress", "test"]
}

resource "arubacloud_subnet" "test" {
  name       = "wp-test-subnet"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"
  tags       = ["network", "wordpress", "test"]
}

resource "arubacloud_securitygroup" "vm" {
  name       = "wp-vm-sg"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "virtual-machine", "wordpress", "test"]

  depends_on = [arubacloud_subnet.test]
}

resource "arubacloud_securitygroup" "dbaas" {
  name       = "wp-db-sg"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "dbaas", "wordpress", "test"]

  depends_on = [arubacloud_subnet.test]
}

resource "arubacloud_elasticip" "vm" {
  name           = "wp-vm-eip"
  location       = "ITBG-Bergamo"
  project_id     = arubacloud_project.test.id
  billing_period = "Hour"
  tags           = ["public", "virtual-machine", "wordpress", "test"]
}

resource "arubacloud_elasticip" "dbaas" {
  name           = "wp-db-eip"
  location       = "ITBG-Bergamo"
  project_id     = arubacloud_project.test.id
  billing_period = "Hour"
  tags           = ["public", "dbaas", "wordpress", "test"]
}

resource "arubacloud_securityrule" "vm_http" {
  name              = "wp-vm-http-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.vm.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_securityrule" "vm_ssh" {
  name              = "wp-vm-ssh-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.vm.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "22"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_securityrule" "vm_egress" {
  name              = "wp-vm-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.vm.id
  properties = {
    direction = "Egress"
    protocol  = "ANY"
    port      = "*"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_securityrule" "dbaas_mysql" {
  name              = "wp-db-mysql-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.dbaas.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "3306"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_securityrule" "dbaas_egress" {
  name              = "wp-db-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.dbaas.id
  properties = {
    direction = "Egress"
    protocol  = "ANY"
    port      = "*"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}
