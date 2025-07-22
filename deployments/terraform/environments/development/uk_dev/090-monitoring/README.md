there is a password for grafana hardcoded

To apply:
cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/090-monitoring
terraform init -upgrade
terraform apply -auto-approve

# Port forward Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Access at http://localhost:3000
# Username: admin
# Password: admin or YOUR_SECURE_DEV_GRAFANA_PASSWORD