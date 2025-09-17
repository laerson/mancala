[k8s_master]
${cluster_name}-master ansible_host=${master_ip} ansible_user=${ssh_user} ansible_ssh_private_key_file=${private_key}

[k8s_nodes:children]
k8s_master

[all:vars]
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'
cluster_name=${cluster_name}