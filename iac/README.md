# Mancala Infrastructure as Code

This directory contains Terraform and Ansible configurations for deploying a single-node Kubernetes cluster on OpenStack and deploying the Mancala application.

## Architecture

- **Terraform**: Creates OpenStack VM with floating IP and security groups
- **Ansible**: Configures Kubernetes cluster using kubeadm and deploys Mancala services

## Prerequisites

1. **Required Tools**:
   - Terraform >= 1.0
   - Ansible >= 2.9
   - kubectl

2. **OpenStack Environment**:
   - Source your OpenStack RC file with the following variables:
     ```bash
     export OS_AUTH_URL="..."
     export OS_TENANT_ID="..."
     export OS_TENANT_NAME="..."
     export OS_USERNAME="..."
     export OS_PASSWORD="..."
     export OS_REGION_NAME="..."
     ```

3. **SSH Key Pair**:
   - Generate or use existing SSH key pair
   - Default location: `~/.ssh/id_rsa` (configurable)

## Configuration

1. Copy the example variables file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your specific values:
   ```hcl
   cluster_name          = "mancala-k8s"
   image_name           = "Ubuntu 22.04"
   flavor_name          = "m1.medium"
   network_name         = "private"
   external_network_name = "public"
   public_key_path      = "~/.ssh/id_rsa.pub"
   private_key_path     = "~/.ssh/id_rsa"
   ssh_user            = "ubuntu"
   ```

## Usage

### Full Deployment

Deploy complete infrastructure and application:
```bash
./deploy.sh deploy
```

### Individual Components

1. **Plan Only** (see what will be created):
   ```bash
   ./deploy.sh plan
   ```

2. **Kubernetes Setup Only** (assumes VM exists):
   ```bash
   ./deploy.sh k8s-only
   ```

3. **Application Deployment Only** (assumes cluster exists):
   ```bash
   ./deploy.sh app-only
   ```

4. **Destroy Infrastructure**:
   ```bash
   ./deploy.sh destroy
   ```

## Manual Steps

### Terraform Only

```bash
cd terraform
terraform init
terraform plan -var-file="../terraform.tfvars"
terraform apply -var-file="../terraform.tfvars"
```

### Ansible Only

```bash
cd ansible
ansible-galaxy collection install -r requirements.yml
ansible-playbook k8s-setup.yml
ansible-playbook deploy-mancala.yml
```

## Outputs

After successful deployment:

- **Master Node IP**: Public IP of the Kubernetes master
- **SSH Access**: `ssh -i ~/.ssh/id_rsa ubuntu@<master-ip>`
- **Games Service**: Available at `<master-ip>:30052`
- **Kubectl Config**: Copy from master node to access cluster locally

## Architecture Details

### Infrastructure (Terraform)

- **VM**: Ubuntu 22.04 on specified flavor
- **Networking**: Attached to private network with floating IP
- **Security Groups**: Configured for SSH, Kubernetes API, and NodePort services
- **Cloud-Init**: Basic system preparation

### Kubernetes Setup (Ansible)

- **Container Runtime**: containerd with systemd cgroup driver
- **Kubernetes**: Latest 1.28.x with kubeadm
- **CNI**: Flannel for pod networking
- **Configuration**: Single-node cluster (control-plane taint removed)
- **Tools**: Helm installed for package management

### Application Deployment

- **Namespace**: `mancala`
- **Services**: Redis, Engine, Games
- **Networking**: NodePort service for external access on port 30052
- **Storage**: Redis with persistent volume

## Troubleshooting

1. **Terraform Issues**:
   - Verify OpenStack credentials
   - Check resource quotas and availability
   - Validate network and image names

2. **Ansible Issues**:
   - Test SSH connectivity: `ansible all -m ping`
   - Check cloud-init completion: `cloud-init status`
   - Verify Python installation on target

3. **Kubernetes Issues**:
   - Check node status: `kubectl get nodes`
   - View pod logs: `kubectl logs -n mancala <pod-name>`
   - Check service endpoints: `kubectl get svc -n mancala`

## Customization

- Modify `terraform/variables.tf` for different instance sizes or configurations
- Update `ansible/k8s-setup.yml` for different Kubernetes versions or CNI
- Adjust security group rules in `terraform/main.tf` for specific access requirements