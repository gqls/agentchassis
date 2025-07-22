terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-monitoring-dev"
    config_path   = "~/.kube/config"
  }

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"  # Downgrade to 2.x
    }
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = var.kube_context_name
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = var.kube_context_name
  }
}

# Create monitoring namespace
resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = "monitoring"
  }
}

# Deploy kube-prometheus-stack (includes Prometheus, Grafana, AlertManager)
resource "helm_release" "kube_prometheus_stack" {
  name       = "kube-prometheus-stack"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  # Repository URL is specified directly
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "51.3.0"  # Use a specific version for stability

  # Basic configuration for development
  values = [<<EOF
prometheus:
  prometheusSpec:
    storageSpec:
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 10Gi

grafana:
  adminPassword: ${var.grafana_admin_password}
  persistence:
    enabled: true
    size: 5Gi
  service:
    type: ClusterIP  # Change to LoadBalancer or NodePort if needed

# Reduce resources for development
defaultRules:
  create: true
  rules:
    alertmanager: false
    etcd: false
    kubeApiserver: false
    kubeApiserverAvailability: false
    kubeApiserverSlos: false
    kubelet: false
    kubeProxy: false
    kubePrometheusGeneral: false
    kubePrometheusNodeRecording: false
    kubernetesApps: false
    kubernetesResources: false
    kubernetesStorage: false
    kubernetesSystem: false
    kubeScheduler: false
    kubeStateMetrics: false
    network: false
    node: false
    nodeExporterAlerting: false
    nodeExporterRecording: false
    prometheus: false
    prometheusOperator: false

alertmanager:
  enabled: false  # Disable for dev

kubeEtcd:
  enabled: false

kubeScheduler:
  enabled: false

kubeProxy:
  enabled: false

kubeControllerManager:
  enabled: false
EOF
  ]

  depends_on = [kubernetes_namespace.monitoring]
}

# Optional: Deploy Kafka exporter for Kafka metrics
resource "helm_release" "kafka_exporter" {
  name       = "kafka-exporter"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus-kafka-exporter"
  version    = "2.1.0"

  values = [<<EOF
kafkaServer:
  - personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092

service:
  port: 9308

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi
EOF
  ]

  depends_on = [kubernetes_namespace.monitoring]
}
