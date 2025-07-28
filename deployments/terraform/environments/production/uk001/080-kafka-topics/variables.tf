variable "context_name" {
  description = "The Kubernetes context name for Kind."
  type        = string
  default     = "personae-uk001-prod-agent-chassis-cluster"
}

variable "namespace" {
  description = "The namespace for the Kafka topics job in the prod environment."
  type        = string
  default     = "kafka"
}

variable "kubeconfig_path" {
  description = "Kubeconfig path"
  type        = string
  default     = "/home/ant/.kube/config_production_uk001"
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