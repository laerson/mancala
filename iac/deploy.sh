#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TERRAFORM_DIR="terraform"
ANSIBLE_DIR="ansible"
TFVARS_FILE="terraform.tfvars"

# Function to print colored output
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if required tools are installed
check_prerequisites() {
    print_message $BLUE "Checking prerequisites..."

    local missing_tools=()

    if ! command -v terraform &> /dev/null; then
        missing_tools+=("terraform")
    fi

    if ! command -v ansible-playbook &> /dev/null; then
        missing_tools+=("ansible")
    fi

    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_message $RED "Missing required tools: ${missing_tools[*]}"
        print_message $YELLOW "Please install the missing tools and try again."
        exit 1
    fi

    print_message $GREEN "All prerequisites are installed."
}

# Function to check OpenStack environment variables
check_openstack_env() {
    print_message $BLUE "Checking OpenStack environment variables..."

    local required_vars=(
        "OS_AUTH_URL"
        "OS_TENANT_ID"
        "OS_TENANT_NAME"
        "OS_USERNAME"
        "OS_PASSWORD"
        "OS_REGION_NAME"
    )

    local missing_vars=()

    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done

    if [ ${#missing_vars[@]} -ne 0 ]; then
        print_message $RED "Missing OpenStack environment variables: ${missing_vars[*]}"
        print_message $YELLOW "Please source your OpenStack RC file and try again."
        exit 1
    fi

    print_message $GREEN "OpenStack environment variables are set."
}

# Function to setup Terraform variables
setup_terraform_vars() {
    if [ ! -f "$TFVARS_FILE" ]; then
        print_message $YELLOW "terraform.tfvars not found. Copying from example..."
        cp terraform.tfvars.example "$TFVARS_FILE"
        print_message $YELLOW "Please edit $TFVARS_FILE with your specific values before proceeding."
        print_message $YELLOW "Press Enter to continue when ready..."
        read
    fi
}

# Function to deploy infrastructure with Terraform
deploy_infrastructure() {
    print_message $BLUE "Deploying infrastructure with Terraform..."

    cd "$TERRAFORM_DIR"

    # Initialize Terraform
    print_message $BLUE "Initializing Terraform..."
    terraform init

    # Plan deployment
    print_message $BLUE "Creating Terraform plan..."
    terraform plan -var-file="../$TFVARS_FILE"

    # Apply deployment
    print_message $BLUE "Applying Terraform configuration..."
    terraform apply -var-file="../$TFVARS_FILE" -auto-approve

    cd ..

    print_message $GREEN "Infrastructure deployed successfully!"
}

# Function to setup Kubernetes cluster with Ansible
setup_kubernetes() {
    print_message $BLUE "Setting up Kubernetes cluster with Ansible..."

    cd "$ANSIBLE_DIR"

    # Install Ansible collections
    print_message $BLUE "Installing Ansible collections..."
    ansible-galaxy collection install -r requirements.yml

    # Wait for instance to be fully ready
    print_message $BLUE "Waiting for instance to be ready..."
    sleep 30

    # Test connectivity
    print_message $BLUE "Testing Ansible connectivity..."
    ansible all -m ping

    # Setup Kubernetes cluster
    print_message $BLUE "Installing and configuring Kubernetes..."
    ansible-playbook k8s-setup.yml

    cd ..

    print_message $GREEN "Kubernetes cluster setup completed!"
}

# Function to deploy Mancala application
deploy_application() {
    print_message $BLUE "Deploying Mancala application..."

    cd "$ANSIBLE_DIR"

    # Deploy the application
    ansible-playbook deploy-mancala.yml

    cd ..

    print_message $GREEN "Mancala application deployed successfully!"
}

# Function to get cluster information
get_cluster_info() {
    print_message $BLUE "Getting cluster information..."

    cd "$TERRAFORM_DIR"

    MASTER_IP=$(terraform output -raw master_public_ip)
    SSH_COMMAND=$(terraform output -raw ssh_connection_command)
    KUBECTL_COMMAND=$(terraform output -raw kubectl_config_command)

    cd ..

    print_message $GREEN "Deployment completed successfully!"
    echo
    print_message $BLUE "Cluster Information:"
    echo "Master Node IP: $MASTER_IP"
    echo "SSH Command: $SSH_COMMAND"
    echo "Kubectl Config: $KUBECTL_COMMAND"
    echo
    print_message $BLUE "Application Access:"
    echo "Games Service: $MASTER_IP:30052"
    echo
    print_message $YELLOW "To access your cluster locally:"
    echo "1. Copy kubectl config: $KUBECTL_COMMAND"
    echo "2. Use kubectl port-forward for local access to services"
}

# Function to destroy infrastructure
destroy_infrastructure() {
    print_message $YELLOW "This will destroy all infrastructure. Are you sure? (y/N)"
    read -r response

    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_message $BLUE "Destroying infrastructure..."

        cd "$TERRAFORM_DIR"
        terraform destroy -var-file="../$TFVARS_FILE" -auto-approve
        cd ..

        print_message $GREEN "Infrastructure destroyed successfully!"
    else
        print_message $BLUE "Destruction cancelled."
    fi
}

# Main function
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            check_openstack_env
            setup_terraform_vars
            deploy_infrastructure
            setup_kubernetes
            deploy_application
            get_cluster_info
            ;;
        "destroy")
            destroy_infrastructure
            ;;
        "plan")
            check_prerequisites
            check_openstack_env
            setup_terraform_vars
            cd "$TERRAFORM_DIR"
            terraform plan -var-file="../$TFVARS_FILE"
            cd ..
            ;;
        "k8s-only")
            check_prerequisites
            setup_kubernetes
            deploy_application
            ;;
        "app-only")
            deploy_application
            ;;
        *)
            echo "Usage: $0 {deploy|destroy|plan|k8s-only|app-only}"
            echo "  deploy    - Deploy complete infrastructure and application"
            echo "  destroy   - Destroy all infrastructure"
            echo "  plan      - Show Terraform execution plan"
            echo "  k8s-only  - Setup Kubernetes and deploy app (assumes VM exists)"
            echo "  app-only  - Deploy application only (assumes cluster exists)"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"