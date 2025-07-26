variable "operator_namespace" {
  description = "Namespace to deploy the Strimzi Kafka operator into."
  type        = string
}

variable "watched_namespaces_list" {
  description = "List of namespaces for the Strimzi operator to watch."
  type        = list(string)
}

variable "strimzi_yaml_source_path" {
  description = "Path to the directory containing the Strimzi YAML files to apply (e.g., ./strimzi-0.47.0/install/cluster-operator/)."
  type        = string
}

variable "operator_deployment_yaml_filename" {
  description = "Filename of the main operator deployment YAML within the strimzi_yaml_source_path (used for trigger)."
  type        = string
  default     = "060-Deployment-strimzi-cluster-operator.yaml"
}

variable "cluster_kubeconfig_path" {
  description = "Path to the kubeconfig file for the target Kubernetes cluster."
  type        = string
  sensitive   = true
}