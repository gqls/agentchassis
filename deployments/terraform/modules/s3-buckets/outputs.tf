output "bucket_ids" {
  description = "A map of bucket names to their Backblaze B2 IDs."
  value = {
    for bucket in b2_bucket.storage_buckets : bucket.bucket_name => bucket.bucket_id
  }
}

output "bucket_names" {
  description = "A list of the names of the created buckets."
  value = [for bucket in b2_bucket.storage_buckets : bucket.bucket_name]
}

variable "b2_application_key_id" {
  description = "The application key ID for Backblaze B2."
  type        = string
  sensitive   = true
}

variable "b2_application_key" {
  description = "The application key for Backblaze B2."
  type        = string
  sensitive   = true
}