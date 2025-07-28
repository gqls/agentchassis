terraform {
  required_providers {
    b2 = {
      source  = "Backblaze/b2"
      version = "~> 0.8.0"  # or whatever version you prefer
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

module "storage_buckets" {
  source = "../../../../modules/s3-buckets"

  bucket_names = [
    var.image_bucket_name,
    var.site_assets_bucket_name
  ]

  b2_application_key_id = var.b2_application_key_id
  b2_application_key    = var.b2_application_key

  tags = {
    environment = "production"
    region      = var.region
    managed_by  = "terraform"
  }
}