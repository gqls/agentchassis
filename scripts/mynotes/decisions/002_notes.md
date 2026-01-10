kubectl namespace is ai-persona-system

so:
kubectl -n ai-persona-system logs -l agent-type=pageflow-builder --since=30m | grep "call_site_planner" | grep -E "await_response|needsWaiting|Added awaited" | head -10

agents are labelled agent-type
adapters are labelled app