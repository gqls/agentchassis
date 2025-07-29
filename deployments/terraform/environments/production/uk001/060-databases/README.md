kubectl get all -n ai-persona-system

kubectl get pvc -n ai-persona-system
kubectl -n ai-persona-system describe pvc postgres-storage-postgres-clients-0
kubectl -n ai-persona-system delete pvc postgres-storage-postgres-clients-0
redeploy before delete and also do templates

kubectl describe pod postgres-clients-0 -n ai-persona-system
kubectl describe pod postgres-templates-0 -n ai-persona-system

# whitelist rackspace ip
# Get your cluster nodes' external IPs
kubectl get nodes -o wide

--

# mysql migrations
# First, find the job name if it exists
kubectl get jobs -n ai-persona-system | grep mysql-migrations
# If a job exists, import it
terraform import kubernetes_job.mysql_migrations ai-persona-system/mysql-migrations-<hash>


# Test connectivity from within the cluster
kubectl run mysql-test --image=mysql:8.0 --rm -it --restart=Never -- mysql -h rs17.uk-noc.com -u catalogu_personae -p

#  Check the actual error:
kubectl describe pod postgres-clients-0 -n ai-persona-system | grep -A 10 Events

# no persistent volume available
kubectl get pvc -n ai-persona-system

# If PVCs exist but are not bound:
kubectl describe pvc -n ai-persona-system

# Check if the storage class exists:
# Your configuration uses storage_class_name = "ssd-large" by default:
kubectl get storageclass
# In your terraform.tfvars or command line:
postgres_storage_class = "your-actual-storage-class-name"

kubectl get nodes
kubectl describe nodes | grep -A 5 "Allocated resources"

## . Find the Available StorageClass Name
kubectl get sc -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}'

ssd

--

failed migrations:
jobs
kubectl delete job postgres-migrations-3b097ae1 mysql-migrations-3b097ae1 -n ai-persona-system

# connect to db
--
clients db
kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db

templates db
kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db

mysql db
kubectl run mysql-test --image=mysql:8.0 --rm -it --restart=Never -- mysql -h rs17.uk-noc.com -u catalogu_personae -p -D catalogu_vectordbdev