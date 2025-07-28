image_bucket_id = "92f302a29468550e9e8d0917"
site_assets_bucket_id = "423302a29468550e9e8d0917"

cd ~/projects/agent-chassis/deployments/terraform/environments/production/uk001/050-storage
terraform import -var-file=terraform.tfvars.secret module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-images\"] 92f302a29468550e9e8d0917
terraform import -var-file=terraform.tfvars.secret module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-site-assets\"] 423302a29468550e9e8d0917
