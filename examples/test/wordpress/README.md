# ArubaCloud Terraform Provider - WordPress Example

This example deploys a complete WordPress stack in ArubaCloud:
- one CloudServer VM for the web application
- one managed MySQL DBaaS instance
- shared network infrastructure (VPC, subnet, security groups, Elastic IPs)

## What This Example Creates

1. Project, VPC, subnet, security groups, and security rules
2. A bootable block storage disk for the VM
3. A MySQL DBaaS instance (`mysql-8.0`) plus WordPress database and DB user
4. A CloudServer VM that installs Apache/PHP/WordPress via cloud-init
5. `wp-config.php` configured to use the provisioned DBaaS endpoint

## Prerequisites

Create `terraform.tfvars` in this folder:

```hcl
arubacloud_api_key          = "your-api-key"
arubacloud_api_secret       = "your-api-secret"
database_password           = "YourStrongDBPassword123!"
wordpress_admin_password    = "YourStrongWpAdminPassword123!"
```

## Quick Start

```bash
cd examples/test/wordpress
terraform init
terraform plan
terraform apply
```

After apply:

```bash
terraform output wordpress_url
terraform output wordpress_admin_user
terraform output wordpress_admin_password
```

Open `wordpress_url` in your browser and log in with the output credentials.

## Files

- `00-variables.tf` - API and app/database credentials
- `01-provider.tf` - Provider and Terraform requirements
- `02-project.tf` - ArubaCloud project
- `03-network.tf` - VPC/subnet/security groups/rules/Elastic IPs
- `04-storage.tf` - boot block storage volume
- `05-database.tf` - MySQL DBaaS + database + database user
- `06-compute.tf` - VM and cloud-init WordPress bootstrap
- `07-output.tf` - URLs and key credentials/connection outputs

## Notes

- This example opens SSH (22), HTTP (80), and MySQL (3306) to `0.0.0.0/0` for testing convenience.
- For production, restrict ingress CIDRs and manage secrets securely (do not keep defaults).
- Provisioning can take several minutes because DBaaS and VM creation are asynchronous.
- Resource **display names** in ArubaCloud are kept short (e.g. `wp-vm-sg`, `wp-vm-eip`) to stay within typical API length limits. If `apply` fails with a name conflict, rename those strings or destroy leftovers from a previous run.

## Troubleshooting

### `Error establishing a database connection`

Typical causes:

1. **Missing database grant** — The DB user must be granted on the logical database (see `arubacloud_databasegrant` in `05-database.tf`). Without it, MySQL rejects the login even if the password is correct.
2. **Cloud-init ran before DBaaS was ready** — The example waits for TCP port **3306** on the DBaaS Elastic IP before writing `wp-config.php` and running WP-CLI.
3. **Special characters in `database_password`** — If your password contains characters that break `sed` (e.g. `\`, `&`), prefer alphanumeric/safe symbols or switch to generating `wp-config.php` with `wp config create` manually.

After changing Terraform, run `terraform apply` so the grant exists; to refresh the VM bootstrap you must **recreate the CloudServer** (new `user_data` / cloud-init) or fix `wp-config.php` by hand on the server.

### Apache “It works!” / default page instead of WordPress

Ubuntu’s Apache package leaves **`/var/www/html/index.html`** in place. With the default `DirectoryIndex`, that file is served **before** WordPress’s `index.php`, so you only see the Apache welcome page.

The example cloud-init removes that file after copying WordPress. If you already applied an older version, either:

- SSH in and run: `sudo rm -f /var/www/html/index.html` then refresh the browser, or  
- change `user_data` and run `terraform apply` so the instance is replaced (or re-run cloud-init if you prefer).

### `One or more validation errors` / `Validation Errors: - <nil>: <nil>`

The provider may not echo every validation field from the API. Use **Terraform/provider debug logs** to see the full error payload (see below).

Also check:

1. **Partial apply / duplicate names** — a failed run may have created a project or VPC; either finish `terraform destroy` or change the `name` fields in the `.tf` files.
2. **Elastic IP quota** — if the account cannot allocate more public IPs, creation of `arubacloud_elasticip` fails; release unused Elastic IPs in the console.
3. **Debug logging** — run `export TF_LOG=INFO` (or `DEBUG`) and retry `terraform apply` to see the logged full API error JSON from the provider.

## Cleanup

```bash
terraform destroy
```
