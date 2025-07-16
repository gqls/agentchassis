variable "region" {
  description = "The region where resources will be deployed."
  type        = string
}

variable "s3_use_path_style" {
  description = "Whether to use path-style addressing for S3. Set to true for MinIO."
  type        = bool
  default     = false
}

variable "image_bucket_name" {
  description = "Name for the bucket to store generated images."
  type        = string
  default     = "personae-prod-uk001-images"
}

variable "site_assets_bucket_name" {
  description = "Name for the bucket to store generated static site assets."
  type        = string
  default     = "personae-prod-uk001-site-assets"
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