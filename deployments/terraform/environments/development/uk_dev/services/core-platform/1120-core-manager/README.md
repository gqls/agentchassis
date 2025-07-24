docker build -f ./build/docker/backend/core-manager.dockerfile -t aqls/core-manager:latest .
docker push aqls/core-manager:latest 

# see auth service README.md

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/services/core-platform/1120-core-manager$ 
terraform taint module.core_manager_deployment_dev.null_resource.apply_kustomization

make create-dev-secrets
make create-dev-configs

--

kubectl get pods -n ai-persona-system -l app=auth-service
# Replace <auth-service-pod-name> with the name from the previous command
kubectl exec auth-service-5dbcf9f85f-tzhs9 -n ai-persona-system -- printenv | grep AUTH_DB_PASSWORD
 
# Now, we'll try to connect to your external MySQL database from inside the cluster using the credentials from your config files and the password we just retrieved.
 Run a temporary network-testing pod:
kubectl run -it --rm --image=ubuntu network-test -n ai-persona-system -- bash
# Run these commands inside the network-test pod
apt-get update && apt-get install -y mysql-client
# Attempt to log in using the hostname, username, and database name from your auth-service.yaml file.
mysql -h rs17.uk-noc.com -u catalogu_agent-chassis -p catalogu_vectordbdev
