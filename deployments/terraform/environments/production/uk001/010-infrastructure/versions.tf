# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/010-infrastructure/versions.tf
terraform {
  required_providers {
    # Configure the Rackspace Spot provider
    spot = {
      source  = "rackerlabs/spot"
      version = "~> 0.1.4" # Match the version constraint in your module
    }
    # Add kubernetes and helm here if you plan to deploy services from this root module later
    # For now, just the spot provider is needed for the cluster module.
  }
  required_version = ">= 1.0"
}