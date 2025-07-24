1. namespace exists
   cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/060-databases
   terraform import -var-file="terraform.tfvars.secret" kubernetes_namespace.db_namespace ai-persona-system

# dev-clients-password
# agent-chassis123!
# dev-templates-password

make create-dev-secrets
# RBAC
kubectl apply -f deployments/kustomize/base/rbac-security.yaml -n ai-persona-system

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/services/core-platform/1120-core-manager 
terraform taint module.core_manager_deployment_dev.null_resource.apply_kustomization

kubectl -n ai-persona-system get replicaset
kubectl describe replicaset core-manager-7c589689d -n ai-persona-system

kubectl create secret generic docker-hub-creds \
-n ai-persona-system \
--from-file=.dockerconfigjson=$HOME/.docker/config.json \
--type=kubernetes.io/dockerconfigjson


~/.docker/config.json should be
{
"auths": {
"docker.io": {
"username": "aqls",
"password": "AaD02432123!",
"email": "aaa@designconsultancy.co.uk",
"auth": "YXFsczpBYUQwMjQzMjEyMyE="
}
}
}

4. Check the MySQL migration job logs:

kubectl get jobs -n ai-persona-system
kubectl -n ai-persona-system logs jobs/mysql-migrations-38b2f9c3

# Get the pod name for the failed job
kubectl get pods -n ai-persona-system -l job-name=mysql-migrations-53dbf62a

# Check the logs (replace the pod name from above)
kubectl logs -n ai-persona-system -l job-name=mysql-migrations-53dbf62a

# Or if multiple attempts:
kubectl logs -n ai-persona-system -l job-name=mysql-migrations-53dbf62a --all-containers=true --prefix=true



2. Check job details:

# Describe the job to see what happened
kubectl describe job mysql-migrations-53dbf62a -n ai-persona-system

# Check events
kubectl get events -n ai-persona-system --sort-by='.lastTimestamp' | grep mysql



3. Common issues and fixes:
Check if MySQL is running:

# Check MySQL pod status
kubectl get pods -n ai-persona-system -l app=mysql

# Check if MySQL service is available
kubectl get svc -n ai-persona-system | grep mysql



Check MySQL connection:

# Test MySQL connection
kubectl run -it --rm mysql-test --image=mysql:8.0 --restart=Never -n ai-persona-system -- \
mysql -h <mysql-service-name> -u root -p<password> -e "SHOW DATABASES;"



Check the migration script:
The issue might be in the migration script itself. Look at what the migration job is trying to do:

# Check the ConfigMap with migration scripts
kubectl get configmap -n ai-persona-system | grep migration
kubectl describe configmap <migration-configmap-name> -n ai-persona-system



4. Quick fixes:
   Option A: Skip schemas for now and continue

Option B: Retry just the schemas

# First, delete the failed job
kubectl delete job mysql-migrations-53dbf62a -n ai-persona-system

# Then retry
cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/070-database-schemas
terraform destroy -auto-approve  # Clean up
terraform apply -auto-approve    # Retry


Option C: Check if databases exist

The postgres migration succeeded but MySQL failed. This suggests MySQL might not be properly set up:
# Check what's in the databases namespace
kubectl get all -n ai-persona-system

--
index exists
Add a flag to skip migrations if the schema already exists:

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/070-database-schemas

# Destroy just the MySQL migration job
terraform destroy -target=kubernetes_job.mysql_migrations -auto-approve

# Or mark it as complete manually
terraform state rm kubernetes_job.mysql_migrations


--
make create-dev-secrets
kubectl apply -f deployments/kustomize/base/rbac-security.yaml -n ai-persona-system