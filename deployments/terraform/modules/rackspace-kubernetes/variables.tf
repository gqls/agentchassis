# The name for the Kubernetes cluster.
variable "cluster_name" {
  description = "Name for the Kubernetes cluster (Spot Cloudspace)."
  type        = string
}

# The Rackspace region for the cluster.
variable "rackspace_region" {
  description = "Rackspace region for the cluster (e.g., uk-lon-1)."
  type        = string
}

# The Kubernetes version.
variable "kubernetes_version" {
  description = "Kubernetes version for the cluster."
  type        = string
  default     = "1.28.8"
}

# A map defining multiple spot node pools. This is the key enhancement.
variable "spot_node_pools" {
  description = "A map of spot node pool configurations. Each key is a pool name."
  type = map(object({
    min_nodes = number
    max_nodes = number
    flavor    = string
    max_price = number
  }))
  default = {}
}

# Optional on-demand node configuration.
variable "ondemand_node_count" {
  description = "Number of on-demand worker nodes."
  type        = number
  default     = 0
}

variable "ondemand_node_flavor" {
  description = "Flavor for on-demand worker nodes."
  type        = string
  default     = null
}

# Slack webhook for preemption notices.
variable "preemption_webhook_url" {
  description = "Slack webhook URL for preemption notices."
  type        = string
  sensitive   = true
  default     = null
}

variable "cni" {
  description = "CNI plugin for the cluster."
  type        = string
  default     = "calico"
}

variable "hacontrol_plane" {
  description = "Enable HA control plane."
  type        = bool
  default     = false
}


variable "spot_min_nodes" {
  description = "Minimum number of spot worker nodes."
  type        = number
  default     = 4
}

variable "spot_max_nodes" {
  description = "Maximum number of spot worker nodes for autoscaling."
  type        = number
  default     = 5
}

variable "spot_node_flavor" {
  description = "Flavor (server class) for spot worker nodes."
  type        = string
}

variable "spot_max_price" {
  description = "Maximum bid price for spot instances."
  type        = number
  default     = 0.01
}

variable "ondemand_node_labels" {
  description = "Labels for on-demand worker nodes."
  type        = map(string)
  default = {
    "role"       = "general",
    "app.type"   = "stateful",
    "managed-by" = "terraform"
  }
}

variable "spot_node_labels" {
  description = "Labels for spot worker nodes."
  type        = map(string)
  default = {
    "role"       = "spot-instance",
    "app.type"   = "stateless",
    "managed-by" = "terraform"
  }
}

variable "ondemand_node_taints" {
  description = "Taints to apply to on-demand worker nodes. E.g., [{key=\"dedicated\", value=\"database\", effect=\"NoSchedule\"}]."
  type = list(object({
    key    = string
    value  = string
    effect = string
  }))
  default = []
}
