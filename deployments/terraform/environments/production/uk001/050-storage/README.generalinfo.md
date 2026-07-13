# Find S3/storage related code
find . -type f -name "*.go" | xargs grep -l "s3\|S3\|bucket\|storage" | head -10

# Or specifically
ls -la platform/storage/
ls -la platform/s3/
ls -la cmd/agent-chassis/storage/

# Find the site publisher implementation
find . -type f -name "*.go" | xargs grep -l "site-publisher\|SitePublisher"

# Or check actions
ls -la platform/actions/
cat platform/actions/publish_actions.go  # if it exists

# Find HTML generation code
find . -type f -name "*.go" | xargs grep -l "html-developer\|GenerateHTML"

# Check the storage secrets
kubectl -n ai-persona-system get secret personae-storage-secrets -o jsonpath='{.data}' | jq -r 'keys[]'

# Check the storage config
kubectl -n ai-persona-system get configmap storage-config -o yaml

# Check if agents have these environment variables
kubectl -n ai-persona-system describe pod $(kubectl -n ai-persona-system get pods | grep agent-site-publisher | awk '{print $1}') | grep -A20 "Environment:"


ant@ant-XPS-15-9500:~/projects/agentchassis$ # Check the storage secrets
kubectl -n ai-persona-system get secret personae-storage-secrets -o jsonpath='{.data}' | jq -r 'keys[]'

# Check the storage config
kubectl -n ai-persona-system get configmap storage-config -o yaml

kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT type,
default_config->>'storage' as storage_config,
env_vars
FROM agent_definitions
WHERE type IN ('site-publisher', 'html-developer', 'visual-designer');"

# how website builder is structure
kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db -c "
SELECT id, name,
orchestration_workflow->'steps'->>'publish_site' as publish_step
FROM agent_groups
WHERE group_type = 'website-builder';"

# Check if agents have these environment variables
kubectl -n ai-persona-system describe pod $(kubectl -n ai-persona-system get pods | grep agent-site-publisher | awk '{print $1}') | grep -A20 "Environment:"
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
B2_APPLICATION_KEY
B2_APPLICATION_KEY_ID
S3-ENDPOINT
S3-REGION
apiVersion: v1
data:
S3-ENDPOINT: https://s3.us-east-005.backblazeb2.com
S3-REGION: us-west-004
S3_USE_PATH_STYLE: "false"
assets_bucket: personae-prod-uk001-site-assets
image_bucket: personae-prod-uk001-images
immutable: false
kind: ConfigMap
metadata:
creationTimestamp: "2025-08-01T18:57:05Z"
name: storage-config
namespace: ai-persona-system
resourceVersion: "2642552"
uid: d0239129-02e0-4d16-97ec-423cbfddfe40
Environment:
AGENT_TYPE:             site-publisher
AGENT_ID:               8bfdcf2e-af78-4977-953f-dcc9d6a63867
CLIENT_ID:              demo_client
KAFKA_TOPIC:            system.agent.site-publisher.process
KAFKA_CONSUMER_GROUP:   site-publisher-group-8bfdcf2e
HEALTH_PORT:            8080
METRICS_PORT:           9090
CLIENTS_DB_PASSWORD:    <set to the key 'CLIENTS_DB_PASSWORD' in secret 'personae-platform-secrets'>    Optional: false
TEMPLATES_DB_PASSWORD:  <set to the key 'TEMPLATES_DB_PASSWORD' in secret 'personae-platform-secrets'>  Optional: false
AUTH_DB_PASSWORD:       <set to the key 'AUTH_DB_PASSWORD' in secret 'personae-platform-secrets'>       Optional: false
ANTHROPIC_API_KEY:      <set to the key 'ANTHROPIC_API_KEY' in secret 'personae-default-secrets'>       Optional: false
AGENT_BOOTSTRAP_KEY:    <set to the key 'agent-bootstrap-key' in secret 'personae-platform-secrets'>    Optional: false
CORE_MANAGER_URL:       http://core-manager.ai-persona-system.svc.cluster.local:8088
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-5fw2x (ro)
Conditions:
Type                        Status
PodReadyToStartContainers   True
Initialized                 True
Ready                       True 
