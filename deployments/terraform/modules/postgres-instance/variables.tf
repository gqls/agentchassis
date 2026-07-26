variable "instance_name" {
  description = "The unique name for the PostgreSQL StatefulSet and related resources."
  type        = string
}

variable "namespace" {
  description = "The Kubernetes namespace to deploy the resources into."
  type        = string
}

variable "database_name" {
  description = "The name of the database to create."
  type        = string
}

variable "database_user" {
  description = "The username for the database."
  type        = string
}

variable "database_pass" {
  description = "The password for the database user."
  type        = string
  sensitive   = true
}

variable "storage_class_name" {
  description = "The name of the StorageClass to use for the PersistentVolumeClaim."
  type        = string
}

variable "storage_size" {
  description = "The size of the persistent volume (e.g., '10Gi')."
  type        = string
}

# --- Container resources (bugs_open/082) ---
# These have defaults on purpose: an instance that forgets to set them must
# still get a CPU floor. Leaving them unset is what produced a BestEffort
# database that a co-tenant could starve until kubelet killed it.

variable "cpu_request" {
  description = "Guaranteed CPU floor. Must never be empty: a container with no CPU request is BestEffort (cpu.shares = 2) and loses every contended scheduling round to any co-tenant."
  type        = string
  default     = "500m"
}

variable "cpu_limit" {
  description = "CPU burst ceiling. Well above the ~30m these instances idle at, so that CFS throttling is never the reason a probe times out."
  type        = string
  default     = "2000m"
}

variable "memory_request" {
  description = "Guaranteed memory. Also what lifts the pod out of BestEffort, which is the first QoS class evicted under node memory pressure."
  type        = string
  default     = "512Mi"
}

variable "memory_limit" {
  description = "Memory ceiling. Deliberately generous against observed use (~210Mi RSS): a memory limit that is reached kills the database outright, so it is insurance, not a budget."
  type        = string
  default     = "2Gi"
}