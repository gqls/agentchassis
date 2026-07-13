// kubeconfig_path = "/home/ant/.kube/config_production_uk001" // Will be set by Makefile
// strimzi_operator_target_namespace = "strimzi" // Using default
// watched_namespaces_for_sydney = ["kafka", "personae", "strimzi"] // Using default
// strimzi_yaml_bundle_path_for_sydney = "../../../../modules/strimzi-operator/strimzi-yaml-0.45.0/" // Using default

kubeconfig_path = "/home/ant/.kube/config_production_uk001"
kube_context_name = "uk001-prod-agent-chassis-cluster"
strimzi_version = "0.47.0"
strimzi_operator_target_namespace = "strimzi"
watched_namespaces_for_uk001 = ["kafka", "personae", "strimzi"]