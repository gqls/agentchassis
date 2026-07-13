cd ~/projects/agent-chassis/deployments/terraform/environments/production/uk001/010-infrastructure

# Import the cloudspace into Terraform state
terraform import module.kubernetes_cluster.spot_cloudspace.cluster "uk001-prod-agent-chassis-cluster"
terraform import 'module.kubernetes_cluster.spot_spotnodepool.spot_pools["spot_worker_pool"]' "NODEPOOL_ID"

rackspace_spot_token = "SpCQoPRlnKZcNJ8dM1tKUDO3uvmJ1JqZ3197AO6R4cJro"

make continue-deployment ENVIRONMENT=production REGION=uk001
make deploy-infrastructure-from STEP=040-kafka ENVIRONMENT=production REGION=uk001
make deploy-infrastructure-from-ingress ENVIRONMENT=production REGION=uk001

download kubeconfig from rackspace into Downloads
cd ~/Downloads
mv uk001 tab ~/.kube/config_production_uk001
export KUBECONFIG=~/.kube/config_production_uk001

/home/ant/.kube/config_production_uk001