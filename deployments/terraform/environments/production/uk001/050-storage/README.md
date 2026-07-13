image_bucket_id = "92f302a29468550e9e8d0917"
site_assets_bucket_id = "423302a29468550e9e8d0917"

cd ~/projects/agent-chassis/deployments/terraform/environments/production/uk001/050-storage
terraform import -var-file=terraform.tfvars.secret module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-images\"] 92f302a29468550e9e8d0917
terraform import -var-file=terraform.tfvars.secret module.storage_buckets_dev.b2_bucket.storage_buckets[\"personae-dev-uk-site-assets\"] 423302a29468550e9e8d0917

bash> b2 account authorize
Backblaze application key ID: 0052322485eed970000000001
Backblaze application key:
{
"accountAuthToken": "4_0052322485eed970000000001_01be1fbf_90278a_acct_q3wMTSZB9E6SUH6BFFF5n1s4CsA=",
"accountFilePath": "/home/ant/.config/b2/account_info",
"accountId": "2322485eed97",
"allowed": {
"buckets": null,
"capabilities": [
"bypassGovernance",
"deleteBuckets",
"deleteFiles",
"deleteKeys",
"listBuckets",
"listFiles",
"listKeys",
"readBucketEncryption",
"readBucketLogging",
"readBucketNotifications",
"readBucketReplications",
"readBucketRetentions",
"readBuckets",
"readFileLegalHolds",
"readFileRetentions",
"readFiles",
"shareFiles",
"writeBucketEncryption",
"writeBucketLogging",
"writeBucketNotifications",
"writeBucketReplications",
"writeBucketRetentions",
"writeBuckets",
"writeFileLegalHolds",
"writeFileRetentions",
"writeFiles",
"writeKeys"
],
"namePrefix": "personae"
},
"apiUrl": "https://api005.backblazeb2.com",
"applicationKey": "K005AAjVc74NBn+mitLkgLIo5U/yDbg",
"applicationKeyId": "0052322485eed970000000001",
"downloadUrl": "https://f005.backblazeb2.com",
"isMasterKey": false,
"s3endpoint": "https://s3.us-east-005.backblazeb2.com"
}


--
# 5. Let's trace through exactly what credentials the adapter is using:
# 6. Add some debug logging to your s3.go temporarily:
// In NewS3Client, after getting the credentials:
accessKey := os.Getenv(cfg.AccessKeyEnvVar)
secretKey := os.Getenv(cfg.SecretKeyEnvVar)

// Add debug logging
fmt.Printf("DEBUG: AccessKeyEnvVar=%s, SecretKeyEnvVar=%s\n", cfg.AccessKeyEnvVar, cfg.SecretKeyEnvVar)
fmt.Printf("DEBUG: AccessKey=%s...\n", accessKey[:10])
fmt.Printf("DEBUG: Endpoint=%s, Bucket=%s\n", cfg.Endpoint, cfg.Bucket)