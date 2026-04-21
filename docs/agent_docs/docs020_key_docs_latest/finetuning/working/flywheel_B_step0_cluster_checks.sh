#!/bin/bash
# ============================================================================
# Flywheel B — Step 0: Cluster-side infrastructure checks
# ============================================================================
# Run these commands and share the output. These are read-only.
#
# The goal: confirm the Ollama adapter is serving nomic-embed-text and that
# the chassis image has rag_lookup + rag_index registered.
# ============================================================================

echo "=== A. Ollama adapter pod status ==="
kubectl -n ai-persona-system get pods -l app=ollama-adapter

echo ""
echo "=== B. Models loaded on the Ollama adapter ==="
# Hit /api/tags — lists installed models
kubectl -n ai-persona-system exec deploy/ollama-adapter -- \
    curl -s http://localhost:11434/api/tags

echo ""
echo "=== C. Can we actually embed something? ==="
# If nomic-embed-text is installed, this returns a 768-float array
kubectl -n ai-persona-system exec deploy/ollama-adapter -- \
    curl -s http://localhost:11434/api/embeddings \
    -d '{"model":"nomic-embed-text","prompt":"test embedding"}' \
    | head -c 200

echo ""
echo ""
echo "=== D. Chassis pod + image tag (what's actually running) ==="
kubectl -n ai-persona-system get pods -l app=agent-chassis -o wide
kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}'
echo ""

echo ""
echo "=== E. Chassis pod logs — search for rag action registration ==="
# If registry.go has the rag entries, the chassis should log action names on boot
# (if it logs them at all — may need a different check)
kubectl -n ai-persona-system logs deploy/agent-chassis --tail=500 2>&1 \
    | grep -iE "rag_lookup|rag_index|registered|action" \
    | head -40

echo ""
echo "=== F. Recent chassis activity (sanity check pods are processing) ==="
kubectl -n ai-persona-system logs deploy/agent-chassis --tail=30 2>&1 \
    | head -40
