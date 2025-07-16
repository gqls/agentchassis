variable "rackspace_spot_token" {
  description = "Rackspace Spot API token for this environment."
  type        = string
  sensitive   = true
  # No default, provide via terraform.tfvars.secret or TF_VAR_... env var
}

variable "instance_cluster_name" { // Renamed to avoid conflict with module input name
  description = "Specific name for this instance of the Kubernetes cluster (e.g., sydney-prod-k8s)."
  type        = string
}

variable "instance_rackspace_region" { // Renamed
  description = "Specific Rackspace region for this instance (e.g., aus-syd-1)."
  type        = string
}

variable "instance_ondemand_node_flavor" { // Renamed
  description = "Flavor for on-demand nodes for this instance."
  type        = string
}

variable "instance_spot_node_flavor" { // Renamed
  description = "Flavor for spot nodes for this instance."
  type        = string
}

variable "instance_slack_webhook_url" { // Renamed
  description = "Slack webhook URL for preemption notices for this instance (optional)."
  type        = string
  sensitive   = true
  default     = null
}

variable "instance_ondemand_node_taints" {
  description = "Taints for on-demand nodes for this uk001 instance."
  type = list(object({
    key    = string
    value  = string
    effect = string
  }))
  default = []
}

variable "instance_ondemand_node_count" {
  description = "No. of on-demand instances for this uk001 deployment"
  type = number
  default = 2
}

variable "instance_spot_min_nodes" {
  description = "Min number of spot nodes for uk001 deployment"
  type = number
  default = 3
}

variable "instance_spot_max_nodes" {
  description = "Min number of spot nodes for uk001 deployment"
  type = number
  default = 4
}

variable "instance_spot_max_price" {
  description = "Max price for spot nodes in uk001 deployment"
  type = number
  default = 0.035
}

// Add variables here if you want to override module defaults for this instance
// e.g., var.instance_k8s_version, var.instance_ondemand_node_count