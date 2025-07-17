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

variable "default_partitions" {
  description = "Default number of partitions for development topics."
  type        = number
  default     = 1
}

variable "default_replication_factor" {
  description = "Default replication factor for development topics. Should be 1 for a single-broker dev cluster."
  type        = number
  default     = 1
}