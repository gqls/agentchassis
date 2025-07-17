variable "region" {
  description = "The region where resources will be deployed."
  type        = string
  default     = "uk-dev"
}

variable "image_bucket_name" {
  description = "Name for the bucket to store generated images for development."
  type        = string
  default     = "personae-dev-uk-images"
}

variable "site_assets_bucket_name" {
  description = "Name for the bucket to store generated static site assets for development."
  type        = string
  default     = "personae-dev-uk-site-assets"
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