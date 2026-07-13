# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/010-infrastructure/providers.tf
provider "spot" {
  token = var.rackspace_spot_token
}