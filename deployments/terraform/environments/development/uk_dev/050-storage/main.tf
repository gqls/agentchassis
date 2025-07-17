terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-storage-dev"
    config_path   = "~/.kube/config"
  }
}

provider "b2" {
  application_key_id = var.b2_application_key_id
  application_key    = var.b2_application_key
}

module "storage_buckets_dev" {
  source = "../../../../modules/s3-buckets"

  bucket_names = [
    var.image_bucket_name,
    var.site_assets_bucket_name
  ]

  tags = {
    environment = "development"
    region      = var.region
    managed_by  = "terraform"
  }
}