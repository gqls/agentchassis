# terraform.tfvars for Strimzi 0.47.0
strimzi_operator_dev_namespace = "strimzi"
watched_namespaces_dev = ["kafka", "personae", "strimzi"]
strimzi_yaml_bundle_path_dev = "../../../../modules/strimzi-operator/strimzi-0.47.0/"
strimzi_operator_deployment_yaml_filename_dev = "060-Deployment-strimzi-cluster-operator.yaml"