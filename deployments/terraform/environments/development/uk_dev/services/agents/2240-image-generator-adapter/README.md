cd ~/projects/agent-chassis/kustomize/services/image-generator-adapter/overlays/development
# Save the debug patch
cat > patch-deployment-debug.yaml << 'EOF'
- op: replace
  path: /spec/replicas
  value: 1
- op: replace
  path: /spec/template/spec/containers/0/image
  value: docker.io/aqls/image-generator-adapter:latest
- op: replace
  path: /spec/template/spec/containers/0/envFrom/0/configMapRef/name
  value: personae-dev-config
- op: replace
  path: /spec/template/spec/containers/0/envFrom/1/secretRef/name
  value: personae-dev-secrets
- op: add
  path: /spec/template/spec/serviceAccountName
  value: ai-persona-app
- op: add
  path: /spec/template/spec/containers/0/command
  value:
    - /bin/sh
    - -c
    - |
      echo "=== Environment Variables ==="
      env | sort
      echo "=== Kafka related vars ==="
      env | grep -i kafka || echo "No kafka vars found"
      echo "=== Config file content ==="
      cat /app/configs/image-adapter.yaml || echo "Config file not found"
      echo "=== Sleeping for debugging ==="
      sleep 3600
      EOF

# Update kustomization to use debug patch temporarily
cp patch-deployment-dev.yaml patch-deployment-dev.yaml.bak
cp patch-deployment-debug.yaml patch-deployment-dev.yaml

# Apply it
kubectl apply -k .

# Wait for pod to start
sleep 5

# Check the logs
kubectl logs -l app=image-generator-adapter -n ai-persona-system

--

kubectl get configmap personae-dev-config -n ai-persona-system -o yaml
# And check if the pod is actually using the updated ConfigMap:
kubectl get pod -l app=image-generator-adapter -n ai-persona-system -o yaml | grep -A 10 envFrom