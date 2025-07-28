kubectl get all -n ai-persona-system

kubectl get pvc -n ai-persona-system
kubectl -n ai-persona-system describe pvc postgres-storage-postgres-clients-0
kubectl -n ai-persona-system delete pvc postgres-storage-postgres-clients-0
redeploy before delete and also do templates

kubectl describe pod postgres-clients-0 -n ai-persona-system
kubectl describe pod postgres-templates-0 -n ai-persona-system

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