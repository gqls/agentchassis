# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/backend.tf
# Configure Terraform state backend (e.g., local, S3, Terraform Cloud)
terraform {
  backend "local" {
    path = "terraform.tfstate.uk001-prod-cluster" # Specific state file for this instance
  }
}