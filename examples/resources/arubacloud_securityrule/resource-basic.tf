# Example: Ingress rule allowing HTTP traffic
resource "arubacloud_securityrule" "http" {
  name              = "example-http-rule"
  location          = "ITBG-Bergamo"
  project_id        = "example-project-id"
  vpc_id            = "example-vpc-id"
  security_group_id = "example-security-group-id"
  tags              = ["security", "example"]
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

# Example: Egress rule allowing all outbound traffic
# Note: For ANY/ICMP protocols, the port field is automatically omitted from the API request
resource "arubacloud_securityrule" "egress_all" {
  name              = "example-egress-all"
  location          = "ITBG-Bergamo"
  project_id        = "example-project-id"
  vpc_id            = "example-vpc-id"
  security_group_id = "example-security-group-id"
  properties = {
    direction = "Egress"
    protocol  = "ANY"  # Case-insensitive: ANY, any, Any all work
    port      = "*"    # Automatically omitted for ANY/ICMP protocols
    target = {
      kind  = "Ip"     # Case-insensitive: Ip, IP, ip all normalized to IP
      value = "0.0.0.0/0"
    }
  }
}
