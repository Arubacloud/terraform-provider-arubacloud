# ArubaCloud Terraform Provider - Container Example

This example demonstrates a complete container infrastructure deployment with a Container Registry and a managed Kubernetes (KaaS) cluster.

## What This Example Does

This Terraform configuration:
1. Creates a project and all required network infrastructure (VPC, subnet, security groups)
2. Provisions a Container Registry for storing Docker images
3. Deploys a managed Kubernetes cluster (KaaS) with node pools
4. Configures network security rules for container registry and Kubernetes
5. Sets up storage for the container registry

## Features

### Container Registry
The example provisions a private container registry with:
- **Dedicated Elastic IP**: Public IP address for registry access
- **HTTPS access**: Port 443 configured for secure image push/pull
- **Dedicated storage**: Block storage volume for registry data
- **Admin user**: Configurable admin user for registry management

### Kubernetes as a Service (KaaS)
The example deploys a managed Kubernetes cluster with:
- **Kubernetes version 1.33.2**: Latest supported version (see [Kubernetes Version documentation](https://api.arubacloud.com/docs/metadata#kubernetes-version))
- **K2A4 node flavor**: 2 CPU, 4GB RAM, 40GB storage per node (see [KaaS flavors documentation](https://api.arubacloud.com/docs/metadata#kaas-flavors))
- **Node pools**: Configurable node pools with autoscaling
- **High Availability**: HA mode enabled for production readiness
- **Network configuration**: Separate CIDR blocks for nodes and pods

### Network Security
Security rules configured to allow:
- **Port 443 (HTTPS)**: Inbound access for container registry from anywhere (0.0.0.0/0)
- **Egress traffic**: All outbound traffic allowed for registry and Kubernetes operations

## Prerequisites

### 1. Credentials
Create `terraform.tfvars` with your ArubaCloud credentials:
```hcl
arubacloud_api_key    = "your-api-key"
arubacloud_api_secret = "your-api-secret"
```

### 2. Provider Binary
Build the provider if not already built:
```bash
cd /path/to/terraform-provider-arubacloud
go build -o terraform-provider-arubacloud
```

## Quick Start

1. **Set up the example directory**:
   ```bash
   cd examples/test/container
   ```

2. **Create terraform.tfvars** with your credentials (see Prerequisites above)

3. **Initialize Terraform**:
   ```bash
   terraform init
   ```

4. **Review the plan**:
   ```bash
   terraform plan
   ```

5. **Apply the configuration**:
   ```bash
   terraform apply
   ```

6. **Get connection information** after deployment completes:
   ```bash
   # Get the Container Registry Elastic IP
   terraform output container_registry_elastic_ip
   
   # Get the Container Registry ID
   terraform output container_registry_id
   
   # Get the KaaS cluster ID
   terraform output kaas_id
   
   # Get the KaaS cluster URI
   terraform output kaas_uri
   ```

## Expected Output

After deployment, you'll have:
- A running Container Registry accessible via HTTPS
- A managed Kubernetes cluster with configured node pools
- Network security configured for container operations
- Storage volumes for the container registry

## File Structure

- `00-variables.tf` - Input variable declarations for API credentials
- `01-provider.tf` - Provider configuration with credentials
- `02-project.tf` - Project resource
- `03-network.tf` - VPC, Subnet, Security Groups, Elastic IP, Security Rules
- `04-storage.tf` - Block storage volume for container registry
- `05-container.tf` - Container Registry and KaaS resources
- `06-output.tf` - Output definitions (registry ID, KaaS ID, Elastic IP, etc.)

## Configuration Details

### Step 0: Variables
Defines input variables for the ArubaCloud API credentials (`arubacloud_api_key` and `arubacloud_api_secret`), marked as sensitive.

### Step 1: Provider Configuration
Configure the ArubaCloud provider with your API credentials and default settings, using the variables defined in `00-variables.tf`.

### Step 2: Project
Creates an ArubaCloud project that will contain all resources.

### Step 3: Network Resources
Creates:
- VPC (Virtual Private Cloud)
- Subnet with Basic type
- Security Groups:
  - Container Registry security group
  - KaaS security group
- Elastic IP for container registry
- Security Rules:
  - HTTPS ingress on port 443 from 0.0.0.0/0 for container registry
  - Egress traffic for all outbound connections

### Step 4: Storage Resources
Creates:
- Block storage volume for container registry data

### Step 5: Container Resources
Creates:
- **Container Registry**: Private Docker registry with HTTPS access
- **KaaS Cluster**: Managed Kubernetes cluster with:
  - Kubernetes version 1.33.2
  - Node pool with K2A4 flavor (2 CPU, 4GB RAM, 40GB storage)
  - Autoscaling enabled (1-5 nodes)
  - High Availability enabled
  - Separate CIDR blocks for nodes (10.0.0.0/24) and pods (10.0.3.0/24)

## Using the Container Registry

After deployment, you can push and pull Docker images:

```bash
# Get the registry endpoint
REGISTRY_IP=$(terraform output -raw container_registry_elastic_ip)

# Login to the registry (use the admin_user and password from ArubaCloud console)
docker login $REGISTRY_IP:443

# Tag and push an image
docker tag myimage:latest $REGISTRY_IP:443/myimage:latest
docker push $REGISTRY_IP:443/myimage:latest

# Pull an image
docker pull $REGISTRY_IP:443/myimage:latest
```

**Note**: You'll need the admin credentials from the ArubaCloud console to authenticate with the registry.

## Using the Kubernetes Cluster

After deployment, you can connect to your Kubernetes cluster:

1. **Get cluster credentials** from the ArubaCloud console
2. **Configure kubectl** using the provided kubeconfig
3. **Deploy applications** to your cluster

```bash
# Example: Deploy a simple nginx pod
kubectl create deployment nginx --image=nginx

# Check cluster nodes
kubectl get nodes

# Check pods
kubectl get pods
```

## Customization

### Change Kubernetes Version
Update the `kubernetes_version` in `05-container.tf`:
```hcl
kubernetes_version = "1.33.2"  # See https://api.arubacloud.com/docs/metadata#kubernetes-version
```

### Change Node Pool Flavor
Update the `instance` field in the node pool configuration:
```hcl
node_pools = [
  {
    name        = "pool-1"
    nodes       = 2
    instance    = "K4A8"  # 4 CPU, 8GB RAM, 80GB storage
    # ...
  }
]
```

Available KaaS flavors (see [KaaS flavors documentation](https://api.arubacloud.com/docs/metadata#kaas-flavors)):
- `K2A4`: 2 CPU, 4GB RAM, 40GB storage
- `K4A8`: 4 CPU, 8GB RAM, 80GB storage
- And more

### Configure Node Pool Autoscaling
Modify the autoscaling settings in the node pool:
```hcl
node_pools = [
  {
    name        = "pool-1"
    nodes       = 3
    instance    = "K2A4"
    autoscaling = true
    min_count   = 2
    max_count   = 10
  }
]
```

### Add Multiple Node Pools
Add additional node pools for different workloads:
```hcl
node_pools = [
  {
    name        = "pool-1"
    nodes       = 2
    instance    = "K2A4"
    autoscaling = true
    min_count   = 1
    max_count   = 5
  },
  {
    name        = "pool-2"
    nodes       = 1
    instance    = "K4A8"
    autoscaling = false
    min_count   = 1
    max_count   = 1
  }
]
```

### Change Container Registry Admin User
Update the `admin_user` in the container registry resource:
```hcl
resource "arubacloud_containerregistry" "test" {
  # ...
  admin_user = "myadmin"
}
```

### Change Container Registry Network or Storage
Update the `network` or `storage` blocks:
```hcl
resource "arubacloud_containerregistry" "test" {
  # ...
  network = {
    public_ip_uri_ref      = arubacloud_elasticip.new.uri
    vpc_uri_ref            = arubacloud_vpc.test.uri
    subnet_uri_ref         = arubacloud_subnet.test.uri
    security_group_uri_ref = arubacloud_securitygroup.new.uri
  }
  
  storage = {
    block_storage_uri_ref = arubacloud_blockstorage.new.uri
  }
}
```

### Configure Network CIDR Blocks
Update the node and pod CIDR blocks:
```hcl
resource "arubacloud_kaas" "test" {
  # ...
  node_cidr = {
    address = "10.0.1.0/24"
    name    = "kaas-node-cidr"
  }
  pod_cidr = "10.0.4.0/24"
}
```

### Restrict Container Registry Access
Update the security rule in `03-network.tf` to restrict HTTPS access:
```hcl
target = {
  kind  = "Ip"
  value = "203.0.113.0/24"  # Your office IP range
}
```

## Important Notes

### Security Group References
- **Container Registry**: Uses URI reference for security group (`security_group_uri_ref`)
- **KaaS**: Uses security group name (`security_group_name`) - must match an existing security group

### Resource Dependencies
- Resources automatically wait for dependencies to be active
- Default timeout is 10 minutes, configurable in the provider block
- KaaS cluster creation may take 15-20 minutes

### Cost
This example provisions paid resources:
- Container Registry (hourly billing)
- KaaS cluster (hourly billing per node)
- Elastic IP (hourly billing)
- Block storage (based on size)

Remember to destroy resources when done testing:
```bash
terraform destroy
```

### High Availability
The KaaS cluster is configured with `controlplane_ha = true` for production readiness. This ensures:
- Multiple control plane nodes
- Improved reliability and fault tolerance
- Higher resource usage

## Troubleshooting

### Container Registry Creation Fails
- Verify the Elastic IP is correctly associated
- Check security rules allow port 443
- Ensure block storage is properly attached
- Verify the project has sufficient quota

### KaaS Cluster Creation Fails
- Check the Kubernetes version is valid: https://api.arubacloud.com/docs/metadata#kubernetes-version
- Verify the KaaS flavor is available: https://api.arubacloud.com/docs/metadata#kaas-flavors
- Ensure the node CIDR doesn't conflict with existing networks
- Check if the project has sufficient quota for the requested node count

### Cannot Push/Pull Images
- Verify the Elastic IP is accessible
- Check security rules allow port 443 from your IP
- Ensure you're using the correct admin credentials
- Verify the registry is in "Active" state

### Kubernetes Cluster Not Accessible
- Get cluster credentials from the ArubaCloud console
- Verify the cluster is in "Active" state
- Check network connectivity to the cluster endpoint
- Ensure kubectl is properly configured

### Node Pool Scaling Issues
- Verify autoscaling is enabled
- Check min_count and max_count are within valid ranges
- Ensure the project has sufficient quota for scaling
- Monitor cluster resource usage

### Resource Timeout
- KaaS cluster creation can take 15-20 minutes
- Increase the timeout in the provider block if needed:
  ```hcl
  provider "arubacloud" {
    # ...
    timeout = "30m"
  }
  ```

## Cleanup

```bash
# Destroy all resources
terraform destroy

# Or destroy specific resources
terraform destroy -target=arubacloud_kaas.test
terraform destroy -target=arubacloud_containerregistry.test
terraform destroy -target=arubacloud_project.test
```

**Note**: Destroy operations may take several minutes as Terraform waits for resources to be properly deleted. KaaS cluster deletion can take 10-15 minutes.

## Additional Resources

- [ArubaCloud API Documentation](https://api.arubacloud.com/docs/)
- [KaaS Flavors](https://api.arubacloud.com/docs/metadata#kaas-flavors)
- [Kubernetes Versions](https://api.arubacloud.com/docs/metadata#kubernetes-version)
- [Provider Documentation](../../../docs/)
