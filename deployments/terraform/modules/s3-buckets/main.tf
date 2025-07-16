# main.tf for the s3-buckets module

terraform {
  required_providers {
    b2 = {
      source  = "Backblaze/b2"
      version = "~> 0.6" # Use a recent version of the Backblaze provider
    }
  }
}

# Create a Backblaze B2 bucket for each name in the list
resource "b2_bucket" "storage_buckets" {
  for_each = toset(var.bucket_names)

  bucket_name = each.key
  bucket_type = "allPrivate" # Default to private, can be overridden

  lifecycle {
    prevent_destroy = false # Set to true in production for safety
  }

  cors_rules {
    cors_rule_name = "allowAll"
    allowed_origins = ["*"]
    allowed_operations = [
      "s3_delete",
      "s3_get",
      "s3_head",
      "s3_post",
      "s3_put",
    ]
    allowed_headers = ["*"]
    expose_headers = ["x-bz-content-sha1"]
    max_age_seconds = 3600
  }
}