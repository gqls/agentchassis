#!/bin/bash


echo "=== Website Builder Agents Status ==="
echo ""
echo "JOBS:"
kubectl -n ai-persona-system get jobs | grep -E "agent-(website-builder|domain-analyst|site-architect|content|html-developer|site-publisher|visual-designer)" | tail -10
echo ""
echo "PODS:"
kubectl -n ai-persona-system get pods | grep -E "agent-(website-builder|domain-analyst|site-architect|content|html-developer|site-publisher|visual-designer)" | grep Running
echo ""
echo "To check specific agent logs:"
echo "kubectl -n ai-persona-system logs <pod-name> --tail=50"

