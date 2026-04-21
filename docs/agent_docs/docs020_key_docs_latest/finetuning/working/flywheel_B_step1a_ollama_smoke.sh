#!/bin/bash
# ============================================================================
# Flywheel B — Step 1a: Prove Ollama embedding works
# ============================================================================
# The ollama-adapter container has no curl. We spin up a throwaway curlimages/curl
# pod in the same namespace to hit the adapter's embedding endpoint.
#
# Expected: a JSON response with an "embedding" array of 768 floats.
#
# On failure: either the service isn't reachable, or nomic-embed-text isn't loaded.
# ============================================================================

NS=ai-persona-system
ADAPTER_SVC=ollama-adapter.ai-persona-system.svc.cluster.local:11434

echo "=== 1. List models installed on the adapter ==="
kubectl -n $NS run ollama-check --rm -i --restart=Never --quiet \
    --image=curlimages/curl:latest -- \
    -s http://$ADAPTER_SVC/api/tags

echo ""
echo ""
echo "=== 2. Request an embedding for a short test string ==="
kubectl -n $NS run ollama-embed-check --rm -i --restart=Never --quiet \
    --image=curlimages/curl:latest -- \
    -s http://$ADAPTER_SVC/api/embeddings \
    -H 'Content-Type: application/json' \
    -d '{"model":"nomic-embed-text","prompt":"French Bulldogs are brachycephalic."}' \
    | head -c 400

echo ""
echo ""
echo "=== 3. Confirm embedding dimension is 768 ==="
# This extracts the embedding array length
kubectl -n $NS run ollama-dim-check --rm -i --restart=Never --quiet \
    --image=curlimages/curl:latest -- \
    sh -c "curl -s http://$ADAPTER_SVC/api/embeddings \
      -H 'Content-Type: application/json' \
      -d '{\"model\":\"nomic-embed-text\",\"prompt\":\"test\"}' \
      | grep -o '[0-9.-]*,' | wc -l"
# Not perfect but gives us approximate count. Real check: should output ~768.

echo ""
