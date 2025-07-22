
variable "kube_context_name" {
  description = "The Kubernetes context name for Kind."
  type        = string
  default     = "kind-personae-dev"
}

variable "namespace" {
  description = "The namespace for the Kafka topics job."
  type        = string
  default     = "personae"
}