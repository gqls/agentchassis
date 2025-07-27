# Create the required namespaces that Strimzi will watch
resource "kubernetes_namespace" "watched_ns" {
  for_each = toset(var.watched_namespaces_for_uk001)
  metadata {
    name = each.key
  }
}

# Read the entire multi-document YAML file as a single string
data "local_file" "strimzi_install_yaml" {
  filename = "../../../../modules/strimzi-operator/strimzi-0.47.0-static/strimzi-install.yaml"
}

# Split the file content into a list of individual YAML documents
locals {
  strimzi_docs = split("---", data.local_file.strimzi_install_yaml.content)
}

# Create a manifest for each non-empty document in the list
resource "kubernetes_manifest" "strimzi_resources" {
  for_each = { for i, doc in local.strimzi_docs : i => doc if trimspace(doc) != "" }

  manifest = yamldecode(each.value)

  depends_on = [kubernetes_namespace.watched_ns]
}