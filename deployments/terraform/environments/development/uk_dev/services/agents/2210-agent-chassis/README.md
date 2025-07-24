cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/services/agents/2210-agent-chassis/


terraform destroy \
-var="service_name=agent-chassis" \
-var="namespace=ai-persona-system" \
-var="image_repository=aqls/agent-chassis"

kubectl rollout restart deployment/agent-chassis

make redeploy-agents ENVIRONMENT=development REGION=uk_dev

--

kubectl get pods -n ai-persona-system -l app=agent-chassis
# Replace <pod-name> with the name you got from the first command
kubectl exec agent-chassis-7f7597dbff-52fzz -n ai-persona-system -- printenv | grep PASSWORD

make build-image-generator-adapter
docker push aqls/image-generator-adapter:latest

--
{"level":"fatal","ts":"2025-07-24T16:04:56.134Z","caller":"image-generator-adapter/main.go:39","msg":"Failed to initialize image generator adapter","error":"failed to create kafka consumer: kafka brokers list cannot be empty","stacktrace":"main.main\n\t/app/cmd/image-generator-adapter/main.go:39\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:283"}

# Look at the application code to see what it's looking for
kubectl exec -it deployment/image-generator-adapter -n ai-persona-system -- env | grep -i kafka
error: unable to upgrade connection: container not found ("image-generator-adapter")
crashes too frequently

# Check if Kafka is deployed
kubectl get pods -n kafka

# If Kafka is in a different namespace, check there
kubectl get services --all-namespaces | grep kafka