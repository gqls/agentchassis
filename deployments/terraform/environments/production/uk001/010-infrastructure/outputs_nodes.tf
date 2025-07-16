# In terraform/environments/production/uk001/010-infrastructure/outputs_nodes.tf

# This provider block allows Terraform to connect to your newly created/existing
# Kubernetes cluster using the kubeconfig file that your 'infra-kubeconfig'
# Makefile target should be generating.
# Ensure KUBECONFIG_FILE in your Makefile points to the correct path for this environment.
provider "kubernetes" {
  alias = "k8s_cluster_access" # Alias to avoid conflict with other k8s provider configs if any
  config_path = abspath(pathexpand("~/.kube/config_production_uk001")) # Adjust if your Makefile saves it elsewhere
  # This assumes the kubeconfig exists from a previous apply.
}

data "kubernetes_nodes" "worker_nodes" {
  provider = kubernetes.k8s_cluster_access # Use the aliased provider

  # Optional: Filter for nodes that are expected to run Ingress controllers.
  # Your DaemonSet for ingress-nginx might have a nodeSelector or tolerations.
  # If it runs on all general worker nodes, you might select based on a common worker label.
  # Your nodes show "node-role.kubernetes.io/worker=".
  # If your Ingress DaemonSet is scheduled on all nodes with this role (and no other specific selector),
  # then this is appropriate. If your DaemonSet has a more specific nodeSelector (e.g., role=ingress-node),
  # you should match that here.
  metadata {
    labels = {
      # Adjust this label selector if your Ingress DaemonSet targets specific nodes.
      # If it runs on all schedulable worker nodes without a specific selector,
      # you might rely on all nodes returned or a general worker role.
      "node-role.kubernetes.io/worker" = ""
    }
  }
  # Ensure this data source depends on the cluster being fully provisioned.
  # The provider configuration using the kubeconfig implies this dependency.
  depends_on = [module.kubernetes_cluster]
}

output "ingress_controller_node_external_ips" {
  description = "List of external IP addresses (labelled as INTERNAL in Rackspace) for Kubernetes nodes expected to run Ingress controllers."
  value = [
    for node in data.kubernetes_nodes.worker_nodes.nodes :
    one([ # Use one() to ensure exactly one ExternalIP is found per node, or fail if none/multiple
      for addr in node.status[0].addresses : addr.address if addr.type == "InternalIP"
    ]) if length([for addr in node.status[0].addresses : addr.address if addr.type == "InternalIP"]) > 0
    # This 'if' at the end of the list comprehension filters out nodes that might not have an ExternalIP
  ]
  depends_on = [data.kubernetes_nodes.worker_nodes]
}