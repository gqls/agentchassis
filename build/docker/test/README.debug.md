kubectl port-forward deployment/agent-chassis-debug 40000:40000
# Then connect VS Code or dlv to localhost:40000

apiVersion: apps/v1
kind: Deployment
metadata:
name: agent-chassis-debug
spec:
template:
spec:
containers:
- name: agent
image: your-registry/agent-chassis:debug
ports:
- containerPort: 40000
name: delve