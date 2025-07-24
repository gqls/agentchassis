mysql is in Clook catalogues.
mysql -ucatalogu_agent-chassis -p -hrs17.uk-noc.com catalogu_vectordbdev
external_mysql_password = "agent-chassis123!"
# be careful B E  C A R E F U L
DROP TABLE auth_tokens;
DROP TABLE auth_tokens;
DROP TABLE projects;
DROP TABLE subscriptions;
DROP TABLE subscription_tiers;
DROP TABLE user_profiles;
DROP TABLE user_permissions;
DROP TABLE permissions;
DROP TABLE users;
DROP TABLE users;

1. namespace exists
2. terraform force-unlock 78a0e5b1-f27e-4517-dbd5-7807102d60ac
   cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/060-databases
   terraform import -var-file="terraform.tfvars.secret" kubernetes_namespace.db_namespace ai-persona-system
# dev-clients-password
# agent-chassis123!
# dev-templates-password

make ENVIRONMENT=development REGION=uk_dev deploy-010-infrastructure

---

First, let's verify the secret exists and has the correct values:
# Check if the secret exists
kubectl get secret postgres-passwords -n ai-persona-system

# Decode and check the actual password value
kubectl get secret postgres-passwords -n ai-persona-system -o jsonpath='{.data.clients-password}' | base64 -d; echo

Let's also check what password the Postgres database is actually expecting:
# Check the postgres pod environment
kubectl exec -n ai-persona-system postgres-clients-dev-0 -- env | grep POSTGRES

Check how the postgres module is setting up the database:
# Look at the postgres instance module
cat ~/projects/agent-chassis/deployments/terraform/modules/postgres-instance/main.tf | grep -A10 -B10 "POSTGRES_PASSWORD"

---

kubectl -n ai-persona-system get jobs

# Check the PostgreSQL migration logs
kubectl logs -n ai-persona-system job/postgres-migrations-80e55012 --all-containers=true

# Check the MySQL migration logs
kubectl logs -n ai-persona-system job/mysql-migrations-80e55012

# Check the PostgreSQL secrets
kubectl get secret -n ai-persona-system postgres-clients-dev-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
echo
kubectl get secret -n ai-persona-system postgres-templates-dev-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
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
kubectl logs -n ai-persona-system job/mysql-migrations-daf34be0
Found 3 pods, using pod/mysql-migrations-daf34be0-sgzm4
Applying migrations to MySQL database...
Host: rs17.uk-noc.com
User: catalogu_agent-chassis
Database: catalogu_vectordbdev


# Then try the schemas again
cd ../070-database-schemas
terraform destroy -auto-approve
terraform apply -auto-approve