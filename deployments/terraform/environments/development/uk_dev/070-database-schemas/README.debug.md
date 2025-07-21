mysql is in Clook catalogues.
mysql -ucatalogu_agent-chassis -p -hrs17.uk-noc.com catalogu_vectordbdev
external_mysql_password = "agent-chassis123!"
DROP TABLE auth_tokens;
DROP TABLE projects;
DROP TABLE users;
make ENVIRONMENT=development REGION=uk_dev deploy-010-infrastructure

---

kubectl -n personae-dev-db get jobs

# Check the PostgreSQL migration logs
kubectl logs -n personae-dev-db job/postgres-migrations-80e55012 --all-containers=true

# Check the MySQL migration logs
kubectl logs -n personae-dev-db job/mysql-migrations-80e55012

# Check the PostgreSQL secrets
kubectl get secret -n personae-dev-db postgres-clients-dev-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
echo
kubectl get secret -n personae-dev-db postgres-templates-dev-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
echo


# Go back to the databases directory
cd ../060-databases

# Destroy the databases
terraform destroy -auto-approve -var-file=terraform.tfvars.secret

# Recreate them
terraform apply -auto-approve -var-file=terraform.tfvars.secret

cd ../070-database-schemas
terraform apply -auto-approve

--
kubectl logs -n personae-dev-db job/mysql-migrations-daf34be0
Found 3 pods, using pod/mysql-migrations-daf34be0-sgzm4
Applying migrations to MySQL database...
Host: rs17.uk-noc.com
User: catalogu_agent-chassis
Database: catalogu_vectordbdev


# Then try the schemas again
cd ../070-database-schemas
terraform destroy -auto-approve
terraform apply -auto-approve