# values.yaml.tpl
# See all possible values here:
# https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-prometheus-stack/values.yaml

# Grafana configuration
grafana:
  # Use the password from our variables.tf
  adminPassword: "${grafana_admin_password}"

  # To access Grafana, you'll typically set up an Ingress.
  # For now, we can expose it via a LoadBalancer for direct access.
  # In a real production setup, you would use an Ingress controller.
  service:
    type: LoadBalancer

# Prometheus configuration
prometheus:
  prometheusSpec:
    # Set retention for production workloads
    retention: 30d
    storageSpec:
      volumeClaimTemplate:
        spec:
          # Use your production storage class
          storageClassName: premium-storage
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 50Gi