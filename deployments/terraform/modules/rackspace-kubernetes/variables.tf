# ~/projects/terraform/rackspace_generic/terraform/modules/kubernetes_cluster_rackspace/variables.tf

variable "cluster_name" {
  description = "Name for the Kubernetes cluster (Spot Cloudspace)."
  type        = string
}

variable "rackspace_region" {
  description = "Rackspace region for the cluster (e.g., aus-syd-1)."
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version for the cluster."
  type        = string
  default     = "1.31.1"
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

variable "preemption_webhook_url" {
  description = "Preemption webhook URL (e.g., Slack webhook)."
  type        = string
  sensitive   = true
  default     = null
}

# --- REMOVED OLD ON-DEMAND VARIABLES ---
# variable "ondemand_node_count" { ... }
# variable "ondemand_node_flavor" { ... }
# variable "ondemand_node_labels" { ... }
# variable "ondemand_node_taints" { ... }

# --- REMOVED OLD SPOT VARIABLES ---
# variable "spot_min_nodes" { ... }
# variable "spot_max_nodes" { ... }
# variable "spot_node_flavor" { ... }
# variable "spot_max_price" { ... }
# variable "spot_node_labels" { ... }


# +++ ADDED NEW FLEXIBLE ON-DEMAND POOL VARIABLE +++
variable "ondemand_node_pools" {
  description = "A map of on-demand node pools to create."
  type = map(object({
    node_count = number
    flavor     = string
    labels     = map(string)
    taints = list(object({
      key    = string
      value  = string
      effect = string
    }))
  }))
  default = {
    "default_pool" = {
      node_count = 0
      flavor     = "gp.small" # Example, replace with your actual flavor
      labels = {
        "role"       = "general",
        "app.type"   = "stateful",
        "managed-by" = "terraform"
      }
      taints = []
    }
  }
}

# +++ ADDED NEW FLEXIBLE SPOT POOL VARIABLE +++
variable "spot_node_pools" {
  description = "A map of spot node pools to create."
  type = map(object({
    min_nodes = number
    max_nodes = number
    flavor    = string
    max_price = number
    labels    = map(string)
  }))
  default = {
    "spot_worker_pool" = {
      min_nodes = 3
      max_nodes = 5
      flavor    = "c.large" # Example, replace with your actual flavor
      max_price = 0.01
      labels = {
        "role"       = "spot-instance",
        "app.type"   = "stateless",
        "managed-by" = "terraform"
      }
    }
  }
}