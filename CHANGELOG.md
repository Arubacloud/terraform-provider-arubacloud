## 0.0.1 (January 26, 2026)

FEATURES:

**Resources:**
* `arubacloud_project` - Manage ArubaCloud projects
* `arubacloud_cloudserver` - Manage Cloud Servers (VMs)
* `arubacloud_keypair` - Manage SSH keypairs
* `arubacloud_elasticip` - Manage Elastic IPs
* `arubacloud_blockstorage` - Manage Block Storage volumes
* `arubacloud_snapshot` - Manage Block Storage snapshots
* `arubacloud_vpc` - Manage Virtual Private Clouds
* `arubacloud_subnet` - Manage VPC subnets
* `arubacloud_securitygroup` - Manage Security Groups
* `arubacloud_securityrule` - Manage Security Group rules
* `arubacloud_vpcpeering` - Manage VPC peering connections
* `arubacloud_vpcpeeringroute` - Manage VPC peering routes
* `arubacloud_vpntunnel` - Manage VPN tunnels
* `arubacloud_vpnroute` - Manage VPN routes
* `arubacloud_kaas` - Manage Kubernetes as a Service clusters
* `arubacloud_containerregistry` - Manage Container Registry
* `arubacloud_dbaas` - Manage DBaaS instances
* `arubacloud_database` - Manage databases within DBaaS
* `arubacloud_dbaasuser` - Manage DBaaS users
* `arubacloud_databasegrant` - Manage database grants
* `arubacloud_databasebackup` - Manage database backups
* `arubacloud_backup` - Manage volume backups
* `arubacloud_restore` - Manage volume restores
* `arubacloud_schedulejob` - Manage scheduled jobs
* `arubacloud_kms` - Manage Key Management Service

**Data Sources:**
* All resources have corresponding data sources for importing existing infrastructure

NOTES:
* Initial release of the ArubaCloud Terraform Provider
* Supports all major ArubaCloud services: Compute, Storage, Networking, Kubernetes, DBaaS, and Security
* Provider uses the official ArubaCloud Go SDK
* Resources include automatic waiting for active state after creation
* Comprehensive examples and documentation included
* Known limitation: Key and KMIP resources temporarily disabled pending SDK updates
