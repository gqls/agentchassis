cd deployments/terraform/environments/development/uk_dev/050-storage
terraform force-unlock f6b704ff-2385-5028-3d24-88935cabb016
cd -

# List secrets related to terraform state
kubectl get secrets -A | grep tfstate

# The lock will be in a secret with a name like "tfstate-storage-dev-lock-<some-hash>"
# Delete the lock secret
kubectl delete secret -n default tfstate-storage-dev-lock-default

--
# Bucket names already exist.
Option 1: Import existing buckets (if you own them)
https://secure.backblaze.com/b2_buckets.htm
aaa@designconsultancy.co.uk AaB...
personae-dev-uk-images  Bucket ID: d213e2e2a46815de9e8d0917
personae-dev-uk-site-assets   Bucket ID: 5223e2e2a46815de9e8d0917

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/050-storage
terraform import module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-images\"] <bucket-id>
terraform import module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-site-assets\"] <bucket-id>

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/050-storage
terraform import module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-images\"] d213e2e2a46815de9e8d0917
terraform import module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-site-assets\"] 5223e2e2a46815de9e8d0917

Option 2: Use different bucket names
Update your bucket configuration to use unique names. Add a random suffix or your username:

# In your 050-storage configuration, update the bucket names
locals {
bucket_suffix = "ant-${random_string.bucket_suffix.result}"
}

resource "random_string" "bucket_suffix" {
length  = 6
special = false
upper   = false
}

# Then update your bucket names to include the suffix
# e.g., "personae-dev-uk-images-${local.bucket_suffix}"


--

# To get the bucket ids
# Install B2 CLI if you haven't
pip install b2

# Authorize (you'll need your B2 account ID and application key)
b2 authorize-account <accountId> <applicationKey>

# List buckets to get their IDs
b2 list-buckets

or

Option 2: Check Terraform state from previous runs

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/050-storage
terraform state list
terraform state show module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-images\"]