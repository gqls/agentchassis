1. Check the MySQL migration job logs:

kubectl get jobs -n personae-dev-db
kubectl -n personae-dev-db logs jobs/mysql-migrations-38b2f9c3

# Get the pod name for the failed job
kubectl get pods -n personae-dev-db -l job-name=mysql-migrations-53dbf62a

# Check the logs (replace the pod name from above)
kubectl logs -n personae-dev-db -l job-name=mysql-migrations-53dbf62a

# Or if multiple attempts:
kubectl logs -n personae-dev-db -l job-name=mysql-migrations-53dbf62a --all-containers=true --prefix=true



2. Check job details:

# Describe the job to see what happened
kubectl describe job mysql-migrations-53dbf62a -n personae-dev-db

# Check events
kubectl get events -n personae-dev-db --sort-by='.lastTimestamp' | grep mysql



3. Common issues and fixes:
Check if MySQL is running:

# Check MySQL pod status
kubectl get pods -n personae-dev-db -l app=mysql

# Check if MySQL service is available
kubectl get svc -n personae-dev-db | grep mysql



Check MySQL connection:

# Test MySQL connection
kubectl run -it --rm mysql-test --image=mysql:8.0 --restart=Never -n personae-dev-db -- \
mysql -h <mysql-service-name> -u root -p<password> -e "SHOW DATABASES;"



Check the migration script:
The issue might be in the migration script itself. Look at what the migration job is trying to do:

# Check the ConfigMap with migration scripts
kubectl get configmap -n personae-dev-db | grep migration
kubectl describe configmap <migration-configmap-name> -n personae-dev-db



4. Quick fixes:
   Option A: Skip schemas for now and continue

Option B: Retry just the schemas

# First, delete the failed job
kubectl delete job mysql-migrations-53dbf62a -n personae-dev-db

# Then retry
cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/070-database-schemas
terraform destroy -auto-approve  # Clean up
terraform apply -auto-approve    # Retry


Option C: Check if databases exist

The postgres migration succeeded but MySQL failed. This suggests MySQL might not be properly set up:
# Check what's in the databases namespace
kubectl get all -n personae-dev-db

--
index exists
Add a flag to skip migrations if the schema already exists:

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/070-database-schemas

# Destroy just the MySQL migration job
terraform destroy -target=kubernetes_job.mysql_migrations -auto-approve

# Or mark it as complete manually
terraform state rm kubernetes_job.mysql_migrations