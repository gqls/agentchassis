tnr create --help

ant@ant-XPS-15-9500:~/projects/sites$ POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
echo "Using pod: $POD"

# Check what tools the pod has — it's a scratch/alpine image, may not have curl
kubectl -n ai-persona-system exec $POD -- sh -c 'which curl wget; echo "---"; env | grep -i thunder | sed "s/=.*/=<redacted>/"'
Using pod: thunder-adapter-74bb4d4646-2qz27
/usr/bin/wget
---
CONSUMER_GROUP=<redacted>
THUNDER_COMPUTE_API_KEY=<redacted>
HOSTNAME=<redacted>
REQUESTS_TOPIC=<redacted>
THUNDER_ADAPTER_SERVICE_PORT_HTTP=<redacted>
THUNDER_ADAPTER_PORT_8080_TCP_ADDR=<redacted>
THUNDER_ADAPTER_SERVICE_HOST=<redacted>
POD_NAME=<redacted>
THUNDER_ADAPTER_PORT_8080_TCP_PORT=<redacted>
THUNDER_ADAPTER_PORT_8080_TCP_PROTO=<redacted>
THUNDER_ADAPTER_SERVICE_PORT=<redacted>
THUNDER_ADAPTER_PORT=<redacted>
THUNDER_ADAPTER_PORT_8080_TCP=<redacted>


ant@ant-XPS-15-9500:~/projects/sites$ tnr status
⚠ No instances found. Use 'tnr create' to create a Thunder Compute instance.
Last updated: 13:03:52
ant@ant-XPS-15-9500:~/projects/sites$

