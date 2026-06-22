## 0.1.7 (June 22, 2026)

BUG FIXES:

* **All stateful resources**: Resources that reach a terminal `Failed` state no longer block `terraform destroy` ([#116](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/116)). Previously, `Read()` returned a hard error for `Failed`-state resources, which prevented every subsequent Terraform operation — including `destroy` — from running and left CI pipelines stuck with no automated recovery path. `Read()` now emits a warning instead, populates state normally (the API returns valid metadata for failed resources), and allows Terraform to build a full destroy plan so that `Delete()` is invoked in the correct reverse-dependency order.

## 0.1.6 (May 27, 2026)

BUG FIXES:

* `arubacloud_elasticip`: Fix `billing_period` value sent to the API after backend breaking change — the ElasticIP API now uses `Hour`/`Month`/`Year` (same canonical form as all other resources) instead of the legacy `hourly`/`monthly`/`yearly` lowercase values. The create request was incorrectly sending the old lowercase form, causing API rejections.

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
