variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "mancala-k8s"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "image_name" {
  description = "Name of the OS image to use"
  type        = string
  default     = "Ubuntu 22.04"
}

variable "flavor_name" {
  description = "Name of the flavor to use for the instance"
  type        = string
  default     = "m1.medium"
}

variable "network_name" {
  description = "Name of the network to attach the instance to"
  type        = string
  default     = "private"
}

variable "external_network_name" {
  description = "Name of the external network for floating IPs"
  type        = string
  default     = "public"
}

variable "public_key_path" {
  description = "Path to the public key file"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "private_key_path" {
  description = "Path to the private key file"
  type        = string
  default     = "~/.ssh/id_rsa"
}

variable "ssh_user" {
  description = "SSH user for the instance"
  type        = string
  default     = "ubuntu"
}