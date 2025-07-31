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

--
# get outbound IP address
# Run a test pod that checks your external IP
kubectl run test-ip --image=curlimages/curl --rm -it --restart=Never -- curl -s ifconfig.me

# Or using a different service
kubectl run test-ip --image=curlimages/curl --rm -it --restart=Never -- curl -s ipinfo.io/ip

# Or more detailed info
kubectl run test-ip --image=curlimages/curl --rm -it --restart=Never -- curl -s ipinfo.io

# check from existing pod
# From any running pod
kubectl exec -it -n ai-persona-system auth-service-[pod-name] -- wget -qO- ifconfig.me

# Use a Bastion/Jump Host
# Instead of direct connections, use a bastion host with a fixed IP:
# Bastion host
# In your auth-service config
auth_db_host: "bastion.yourdomain.com"  # Instead of direct MySQL host
auth_db_port: "33306"  # Port forwarded through bastion

# On bastion host
ssh -L 33306:rs17.uk-noc.com:3306 user@bastion

---

# Debug mysql connection
# Test if you can reach the MySQL host from a pod
kubectl run mysql-test --image=mysql:8.0 --rm -it --restart=Never -- \
mysql -h rs17.uk-noc.com -P 3306 -u catalogu_personae -p$AUTH_DB_PASSWORD -e "SELECT 1"

# Or just test TCP connectivity
kubectl run net-test --image=nicolaka/netshoot --rm -it --restart=Never -- \
nc -zv rs17.uk-noc.com 3306

# Test DNS resolution
kubectl run dns-test --image=busybox --rm -it --restart=Never -- \
nslookup rs17.uk-noc.com

# . Check if the auth-service has the correct configuration:
# Check what environment variables the pod is seeing
kubectl exec -n ai-persona-system deployment/auth-service -- env | grep -E "(AUTH_DB|auth_db)"

# Check the configmap
kubectl get configmap personae-prod-config -n ai-persona-system -o yaml | grep auth_db

# Check if the password is set correctly
kubectl get secret personae-platform-secrets -n ai-persona-system -o yaml | grep auth-db-password

# Test from inside the auth-service pod:
# Get a shell in the auth-service pod
kubectl exec -it -n ai-persona-system deployment/auth-service -- /bin/sh

# Inside the pod, test connectivity
apk add --no-cache mysql-client
mysql -h rs17.uk-noc.com -P 3306 -u catalogu_personae -p

# Or test with telnet
apk add --no-cache busybox-extras
telnet rs17.uk-noc.com 3306

# 6. Check for network policies blocking outbound:
# Check if there are any network policies
kubectl get networkpolicies -n ai-persona-system

# Check if there's an egress policy blocking MySQL
kubectl describe networkpolicy -n ai-persona-system

===
# . First, check if the pod is running and has the env vars:
# Check pod status
kubectl get pod mysql-connection-test -n ai-persona-system

# Check environment variables are set
kubectl exec mysql-connection-test -n ai-persona-system -- env | grep MYSQL

# Get into the pod first
kubectl exec -it mysql-connection-test -n ai-persona-system -- /bin/bash

# 2. Once inside the pod, test the connection:
# Inside the pod
mysql -h rs17.uk-noc.com -P 3306 -u catalogu_personae -p
# Enter password when prompted
PpC47410423123!
# Or use the environment variables
mysql -h $MYSQL_HOST -P 3306 -u $MYSQL_USER -p$MYSQL_PASSWORD -e "SELECT 1"
