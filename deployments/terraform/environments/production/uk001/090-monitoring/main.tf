terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-monitoring"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"  # Downgrade to 2.x
    }
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
}


provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
  }
}

# Create monitoring namespace
resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = "monitoring"
  }
}

# Create the Grafana secret here
resource "kubernetes_secret" "grafana_admin_secret" {
  metadata {
    name      = "grafana-admin-secret"
    namespace = kubernetes_namespace.monitoring.metadata[0].name
  }
  data = {
    # The Helm chart expects these specific keys
    "admin-user"     = "admin"
    "admin-password" = var.grafana_admin_password
  }
  depends_on = [kubernetes_namespace.monitoring]
}

# Deploy kube-prometheus-stack (includes Prometheus, Grafana, AlertManager)
resource "helm_release" "kube_prometheus_stack" {
  name       = "kube-prometheus-stack"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  # Repository URL is specified directly
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "51.3.0"  # Use a specific version for stability

  values = [<<EOF
    prometheus:
      prometheusSpec:
        retention: 30d
        storageSpec:
          volumeClaimTemplate:
            spec:
              storageClassName: ssd-large  # Use your Rackspace storage class
              accessModes: ["ReadWriteOnce"]
              resources:
                requests:
                  storage: 100Gi
        resources:
          requests:
            memory: 2Gi
            cpu: 1
          limits:
            memory: 4Gi
            cpu: 2

    grafana:
      persistence:
        enabled: true
        storageClassName: ssd-large
        size: 50Gi
      service:
        type: ClusterIP  # Use ClusterIP for production, expose via ingress
      resources:
        requests:
          memory: 256Mi
          cpu: 250m
        limits:
          memory: 512Mi
          cpu: 500m

    alertmanager:
      enabled: true
      alertmanagerSpec:
        storage:
          volumeClaimTemplate:
            spec:
              storageClassName: ssd-large
              accessModes: ["ReadWriteOnce"]
              resources:
                requests:
                  storage: 10Gi

    # Enable essential monitoring for production
    defaultRules:
      create: true
      rules:
        alertmanager: true
        etcd: false  # Enable if you're monitoring etcd
        kubeApiserver: true
        kubeApiserverAvailability: true
        kubeApiserverSlos: true
        kubelet: true
        kubeProxy: false
        kubePrometheusGeneral: true
        kubePrometheusNodeRecording: true
        kubernetesApps: true
        kubernetesResources: true
        kubernetesStorage: true
        kubernetesSystem: true
        kubeScheduler: false
        kubeStateMetrics: true
        network: true
        node: true
        nodeExporterAlerting: true
        nodeExporterRecording: true
        prometheus: true
        prometheusOperator: true

    # Enable for production monitoring
    kubeStateMetrics:
      enabled: true

    nodeExporter:
      enabled: true

    # Disable if not needed
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

  # This block tells Helm to use the secret you just created
  set {
    name  = "grafana.admin.existingSecret"
    value = kubernetes_secret.grafana_admin_secret.metadata[0].name
  }
  set {
    name  = "grafana.admin.passwordKey"
    value = "admin-password"
  }

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
  - personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092

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
