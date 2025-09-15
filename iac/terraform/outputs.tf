output "master_public_ip" {
  description = "Public IP address of the Kubernetes master node"
  value       = openstack_networking_floatingip_v2.k8s_master_fip.address
}

output "master_private_ip" {
  description = "Private IP address of the Kubernetes master node"
  value       = openstack_compute_instance_v2.k8s_master.network[0].fixed_ip_v4
}

output "instance_id" {
  description = "ID of the created instance"
  value       = openstack_compute_instance_v2.k8s_master.id
}

output "security_group_id" {
  description = "ID of the security group"
  value       = openstack_networking_secgroup_v2.k8s_secgroup.id
}

output "ansible_inventory_path" {
  description = "Path to the generated Ansible inventory file"
  value       = local_file.ansible_inventory.filename
}

output "ssh_connection_command" {
  description = "SSH command to connect to the master node"
  value       = "ssh -i ${var.private_key_path} ${var.ssh_user}@${openstack_networking_floatingip_v2.k8s_master_fip.address}"
}

output "kubectl_config_command" {
  description = "Command to copy kubectl config from master"
  value       = "scp -i ${var.private_key_path} ${var.ssh_user}@${openstack_networking_floatingip_v2.k8s_master_fip.address}:~/.kube/config ~/.kube/config"
}