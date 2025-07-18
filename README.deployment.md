# 1. Create the Kind cluster using Terraform
make ENVIRONMENT=development REGION=uk_dev deploy-010-infrastructure

# 2. Verify the cluster is running
kubectl config use-context kind-personae-dev
kubectl get nodes

# 3. Continue with the rest of the infrastructure
make ENVIRONMENT=development REGION=uk_dev deploy-infrastructure

# 4. Build and deploy services
make build-all
make kind-load-images  # If using local images
make ENVIRONMENT=development REGION=uk_dev deploy-all