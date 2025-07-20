cd deployments/terraform/environments/development/uk_dev/050-storage
terraform force-unlock f6b704ff-2385-5028-3d24-88935cabb016
cd -

# List secrets related to terraform state
kubectl get secrets -A | grep tfstate

# The lock will be in a secret with a name like "tfstate-storage-dev-lock-<some-hash>"
# Delete the lock secret
kubectl delete secret -n default tfstate-storage-dev-lock-default

