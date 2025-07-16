variable "bucket_names" {
  description = "A list of bucket names to create in Backblaze B2."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "A map of tags to assign to the resources."
  type        = map(string)
  default     = {}
}