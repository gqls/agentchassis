# deployments/terraform/environments/production/uk001/100-bootstrap-agents/terraform.tfvars

namespace          = "ai-persona-system"
registry           = "docker.io/aqls"
image_tag          = "v1.0.47"
enable_autoscaling = false
environment        = "production"
region             = "uk001"