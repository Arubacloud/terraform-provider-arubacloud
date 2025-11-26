resource "arubacloud_securityrule" "example" {
  name              = "example-security-rule"
  vpc_id            = "example-vpc-id"
  security_group_id = "example-security-group-id"
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
