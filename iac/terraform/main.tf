terraform {
  required_version = ">= 1.0"
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.53.0"
    }
  }
}

provider "openstack" {
  # OpenStack credentials should be provided via environment variables:
  # OS_AUTH_URL, OS_TENANT_ID, OS_TENANT_NAME, OS_USERNAME, OS_PASSWORD, OS_REGION_NAME
}

# Data sources for existing OpenStack resources
data "openstack_images_image_v2" "ubuntu" {
  name        = var.image_name
  most_recent = true
}

data "openstack_compute_flavor_v2" "flavor" {
  name = var.flavor_name
}

data "openstack_networking_network_v2" "external" {
  name = var.external_network_name
}

# Create a key pair
resource "openstack_compute_keypair_v2" "k8s_keypair" {
  name       = "${var.cluster_name}-keypair"
  public_key = file(var.public_key_path)
}

# Create a security group
resource "openstack_networking_secgroup_v2" "k8s_secgroup" {
  name        = "${var.cluster_name}-secgroup"
  description = "Security group for Kubernetes cluster"
}

# SSH access
resource "openstack_networking_secgroup_rule_v2" "ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.k8s_secgroup.id
}

# Kubernetes API server
resource "openstack_networking_secgroup_rule_v2" "k8s_api" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.k8s_secgroup.id
}

# NodePort services
resource "openstack_networking_secgroup_rule_v2" "nodeport" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 30000
  port_range_max    = 32767
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.k8s_secgroup.id
}

# HTTP traffic
resource "openstack_networking_secgroup_rule_v2" "http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.k8s_secgroup.id
}

# HTTPS traffic
resource "openstack_networking_secgroup_rule_v2" "https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.k8s_secgroup.id
}

# Default egress rule (allow all outbound traffic) is created automatically by OpenStack

# Create the compute instance
resource "openstack_compute_instance_v2" "k8s_master" {
  name            = "${var.cluster_name}-master"
  image_id        = data.openstack_images_image_v2.ubuntu.id
  flavor_id       = data.openstack_compute_flavor_v2.flavor.id
  key_pair        = openstack_compute_keypair_v2.k8s_keypair.name
  security_groups = [openstack_networking_secgroup_v2.k8s_secgroup.name]

  metadata = {
    Role        = "k8s-master"
    Environment = var.environment
    Cluster     = var.cluster_name
  }

  user_data = templatefile("${path.module}/cloud-init.yaml", {
    hostname = "${var.cluster_name}-master"
  })

  network {
    name = var.network_name
  }
}

# Create and assign floating IP
resource "openstack_networking_floatingip_v2" "k8s_master_fip" {
  pool = data.openstack_networking_network_v2.external.name
}

resource "openstack_compute_floatingip_associate_v2" "k8s_master_fip" {
  floating_ip = openstack_networking_floatingip_v2.k8s_master_fip.address
  instance_id = openstack_compute_instance_v2.k8s_master.id
}

# Wait for instance to be ready
resource "null_resource" "wait_for_instance" {
  provisioner "remote-exec" {
    inline = [
      "echo 'Instance is ready'",
      "cloud-init status --wait"
    ]

    connection {
      type        = "ssh"
      host        = openstack_networking_floatingip_v2.k8s_master_fip.address
      user        = var.ssh_user
      private_key = file(var.private_key_path)
      timeout     = "5m"
    }
  }

  depends_on = [openstack_compute_floatingip_associate_v2.k8s_master_fip]
}

# Generate Ansible inventory
resource "local_file" "ansible_inventory" {
  content = templatefile("${path.module}/inventory.tpl", {
    master_ip    = openstack_networking_floatingip_v2.k8s_master_fip.address
    ssh_user     = var.ssh_user
    private_key  = var.private_key_path
    cluster_name = var.cluster_name
  })
  filename = "${path.module}/../ansible/inventory.ini"

  depends_on = [null_resource.wait_for_instance]
}