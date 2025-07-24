docker build -f ./build/docker/backend/core-manager.dockerfile -t aqls/core-manager:latest .
docker push aqls/core-manager:latest 

# see auth service README.md

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/services/core-platform/1120-core-manager$ 
terraform taint module.core_manager_deployment_dev.null_resource.apply_kustomization

make create-dev-secrets
