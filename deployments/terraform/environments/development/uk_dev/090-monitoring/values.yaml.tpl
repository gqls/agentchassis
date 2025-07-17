# values.yaml.tpl for development
# This configuration is lightweight and suitable for local clusters.

# Grafana configuration
grafana:
  adminPassword: "${grafana_admin_password}"
  # For dev, we use ClusterIP and access via `kubectl port-forward`
  service:
    type: ClusterIP

# Prometheus configuration
prometheus:
  prometheusSpec:
    # Disable persistent storage for development to keep it lightweight
    storageSpec: {}
    retention: 1d # Lower retention for dev