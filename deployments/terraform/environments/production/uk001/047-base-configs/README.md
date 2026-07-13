Docker secret already exists:
terraform import <resource_type>.<resource_name> <namespace>/<secret_name>
#terraform import kubernetes_secret.docker_hub_creds ai-persona-system/docker-hub-creds
terraform import -var-file=terraform.tfvars.secret kubernetes_secret.docker_hub_creds ai-persona-system/docker-hub-creds

terraform force-unlock d9604.....

# check keys in cluster
# Check each key in the secret
for key in jwt-secret-key auth-db-password templates-db-password clients-db-password agent-bootstrap-key; do
value=$(kubectl get secret personae-platform-secrets -n ai-persona-system -o json | jq -r ".data[\"$key\"]" | base64 -d)
echo "$key: $(echo -n "$value" | wc -c) characters"
if [ -z "$value" ]; then
echo "  WARNING: $key is EMPTY!"
fi
done