# scale down
for dep in agent-chassis auth-service content-creator-agent core-manager image-generator-adapter reasoning-agent web-search-adapter; do
  kubectl scale deployment/$dep --replicas=0 -n ai-persona-system
done


# scale up
for dep in agent-chassis auth-service content-creator-agent core-manager image-generator-adapter reasoning-agent web-search-adapter; do
  kubectl scale deployment/$dep --replicas=3 -n ai-persona-system
done