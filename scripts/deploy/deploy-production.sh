#!/bin/bash
set -e

# Load environment
source .env.production

echo "🚀 Deploying to Production UK001..."

# Deploy infrastructure
echo "📦 Step 1: Infrastructure..."
make deploy-010-infrastructure ENVIRONMENT=production REGION=uk001

# Get kubeconfig
echo "🔑 Step 2: Getting kubeconfig..."
cd ~/projects/agent-chassis/deployments/terraform/environments/production/uk001/010-infrastructure
terraform output -raw kubeconfig_raw > ~/.kube/config_production_uk001
cd -
export KUBECONFIG=~/.kube/config_production_uk001

# Deploy remaining infrastructure
echo "🏗️ Step 3: Platform components..."
make deploy-020-ingress ENVIRONMENT=production REGION=uk001
make deploy-030-strimzi ENVIRONMENT=production REGION=uk001
make deploy-040-kafka ENVIRONMENT=production REGION=uk001
make deploy-045-kafka-users ENVIRONMENT=production REGION=uk001
make deploy-047-platform-configs ENVIRONMENT=production REGION=uk001
make deploy-050-storage ENVIRONMENT=production REGION=uk001
make deploy-060-databases ENVIRONMENT=production REGION=uk001
make deploy-070-schemas ENVIRONMENT=production REGION=uk001
make deploy-080-topics ENVIRONMENT=production REGION=uk001
make deploy-090-monitoring ENVIRONMENT=production REGION=uk001

# Build and push images
echo "🐳 Step 4: Building and pushing images..."
make build-all push-all IMAGE_TAG=v1.0.0

# Deploy applications
echo "🚀 Step 5: Deploying applications..."
make deploy-core ENVIRONMENT=production REGION=uk001
make deploy-agents ENVIRONMENT=production REGION=uk001

echo "✅ Production deployment complete!"
echo "📊 Check status with: kubectl get pods -n ai-persona-system"