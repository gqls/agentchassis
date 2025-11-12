echo ""
echo "Simple Test Without LLM"
echo "---------------------------------"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="eborg-content-approval-$(date +%Y%m%d-%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_eborg_content_request"
CLIENT_ID="demo_client"

kubectl -n kafka run -i --rm kcat-producer-eborg \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H step_name=$STEP_NAME \
-H client_id=$CLIENT_ID \
-H message_type=request \
-H action=orchestrate \
-H from_agent_type=user \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.generic.responses <<JSON
{"action":"orchestrate","config":{"workflow":{"start_step":"transform","steps":{"transform":{"action":"transform_data","config":{"transformation":"uppercase"},"next_step":"wait"},"wait":{"action":"await_approval","next_step":"done"},"done":{"action":"complete_workflow"}}}},"input_data":{"message":"test"}}
JSON


kubectl -n kafka run -i --rm kcat-producer-eborg-approval \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=0c142ce5-dc89-402f-b239-fe2a4c39b185 \
-H request_id=$(cat /proc/sys/kernel/random/uuid) \
-H message_id=$(cat /proc/sys/kernel/random/uuid) \
-H orchestration_id=3101ce68-2c8d-4f27-857e-743f2d124662 \
-H request_id=858909e3-467c-4c2e-9488-77e7072f98b0 \
-H approval_token=858909e3-467c-4c2e-9488-77e7072f98b0 \
-H step_name="wait" \
-H message_type=response \
-H action=approved \
-H from_agent_type=human \
-H from_agent_id="hitl_system" \
-H responses_topic=system.generic.responses \
-H client_id=demo_client \
<<JSON
{"message":"Approved by human"}
JSON
