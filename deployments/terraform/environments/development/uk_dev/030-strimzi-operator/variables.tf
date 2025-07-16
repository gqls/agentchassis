variable "kube_context_name" {
  description = "The Kubernetes context name for Kind."
  type        = string
  default = "kind-personae-dev"
}

variable "kubeconfig_path" {
  description = "Optional path to kubeconfig YAML file."
  type        = string
  default     = null # If null, a default single-node cluster is created
}

variable "strimzi_operator_dev_namespace" {
  description = "Namespace for the Strimzi operator in dev."
  type        = string
  default     = "strimzi" // Operator's own namespace
}

variable "watched_namespaces_dev" {
  description = "List of namespaces for the Strimzi operator to watch in dev."
  type        = list(string)
  default     = ["kafka", "personae"] // Strimzi will watch 'kafka' for Kafka CRs and 'personae' if KafkaUsers are there
}

variable "strimzi_yaml_bundle_path_dev" {
  description = "Path to the Strimzi YAML files directory for dev."
  type        = string
  # Path relative to this file's directory, pointing to the module's shared Strimzi YAMLs
  default     = "../../../../modules/strimzi_operator/strimzi-yaml-0.45.0/"
}

variable "strimzi_operator_deployment_yaml_filename_dev" {
description = "Filename of the main operator deployment YAML."
type        = string
default     = "060-Deployment-strimzi-cluster-operator.yaml"
}