Docker secret already exists:
terraform import <resource_type>.<resource_name> <namespace>/<secret_name>
#terraform import kubernetes_secret.docker_hub_creds ai-persona-system/docker-hub-creds
terraform import -var-file=terraform.tfvars.secret kubernetes_secret.docker_hub_creds ai-persona-system/docker-hub-creds

terraform force-unlock d9604.....