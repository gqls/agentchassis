output "image_bucket_id" {
  description = "The ID of the image storage bucket."
  value       = module.storage_buckets.bucket_ids[var.image_bucket_name]
}

output "site_assets_bucket_id" {
  description = "The ID of the site assets storage bucket."
  value       = module.storage_buckets.bucket_ids[var.site_assets_bucket_name]
}