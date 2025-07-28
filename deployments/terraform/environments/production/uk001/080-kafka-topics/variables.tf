# infrastructure/environments/production/uk001/080-kafka-topics/variables.tf

variable "namespace" {
  description = "Kubernetes namespace where Kafka is deployed"
  type        = string
  default     = "kafka"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "/home/ant/.kube/config_production_uk001"
}

variable "context_name" {
  description = "Kubernetes context name"
  type        = string
  default     = "default"
}

variable "platform_topics" {
  description = "A list of Kafka topic names to be created for the application."
  type        = list(string)
  default = [
    # System & Orchestration Topics
    "system.commands.workflow.resume",
    "system.events.workflow.paused",
    "system.events.workflow.completed",

    # Core Service Topics
    "requests.auth.user.create",
    "events.auth.user.created",

    # Agent Communication Topics
    "requests.agent.task.execute",
    "events.agent.task.completed",
    "events.agent.task.failed",
    "events.agent.task.progress",

    # Specialized Agent Topics
    "requests.agent.reasoning",
    "requests.agent.web-search",
    "requests.agent.image-generation",
  ]
}

variable "topic_config" {
  description = "Default configuration for topics"
  type = object({
    partitions        = number
    replication_factor = number
    retention_ms      = string
    segment_ms        = string
  })
  default = {
    partitions        = 3
    replication_factor = 3
    retention_ms      = "604800000"  # 7 days
    segment_ms        = "86400000"   # 1 day
  }
}