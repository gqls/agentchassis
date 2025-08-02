terraform {
  required_providers {
    b2 = {
      source  = "Backblaze/b2"
      version = "~> 0.8.0"  # or whatever version you prefer
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
  backend "kubernetes" {
    secret_suffix    = "tfstate-storage"
    config_path      = "/home/ant/.kube/config_production_uk001"
    # In a real CI/CD pipeline, you might use in_cluster_config = true
  }
}

provider "b2" {
  application_key_id = var.b2_application_key_id
  application_key    = var.b2_application_key
}

provider "kubernetes" {
  config_path   = "/home/ant/.kube/config_production_uk001"
}

data "kubernetes_namespace" "ai_persona_system" {
  metadata {
    name = "ai-persona-system"
  }
}

module "storage_buckets" {
  source = "../../../../modules/s3-buckets"

  bucket_names = [
    var.image_bucket_name,
    var.site_assets_bucket_name
  ]

  B2_APPLICATION_KEY_ID = var.b2_application_key_id
  B2_APPLICATION_KEY    = var.b2_application_key

  tags = {
    environment = "production"
    region      = var.region
    managed_by  = "terraform"
  }
}

# Create storage secrets for applications
resource "kubernetes_secret" "storage_credentials" {
  metadata {
    name      = "personae-storage-secrets"
    namespace = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  }

  data = {
    # B2 credentials for S3-compatible API
    B2_APPLICATION_KEY_ID = var.b2_application_key_id
    B2_APPLICATION_KEY    = var.b2_application_key

    # S3-compatible environment variables
    AWS_ACCESS_KEY_ID     = var.b2_application_key_id
    AWS_SECRET_ACCESS_KEY = var.b2_application_key

    # B2 endpoint configuration
    S3-ENDPOINT = "https://s3.us-east-005.backblazeb2.com"
    S3-REGION   = "us-west-004"
  }
}

# Create a ConfigMap with non-sensitive storage configuration
resource "kubernetes_config_map" "storage_config" {
  metadata {
    name      = "storage-config"
    namespace = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  }

  data = {
    # Bucket names from the module outputs
    image_bucket     = var.image_bucket_name
    assets_bucket    = var.site_assets_bucket_name

    # S3 configuration
    S3-ENDPOINT      = "https://s3.us-east-005.backblazeb2.com"
    S3-REGION       = "us-west-004"
    S3_USE_PATH_STYLE = "false"
  }
}