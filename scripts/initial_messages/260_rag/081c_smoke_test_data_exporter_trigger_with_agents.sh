CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

echo "CORRELATION_ID=$CORRELATION_ID"
echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm kcat-export-smoke-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=training-export-smoke-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=demo_client" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"training-data-exporter"},"input_data":{"agent_type":"page-content-writer","step_name":"process_sections_loop_iter_0_generate_content","model_filter":"claude-sonnet-4-6","output_path":"/tmp/training_exports/smoke_test.jsonl","max_rows":5}}
JSON

echo ""
echo "Saved:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"





SELECT status, current_step, LEFT(error, 300) as error_preview,
       jsonb_pretty(collected_data->'export_result') as summary
FROM orchestration_states
WHERE orchestration_id = '$ORCHESTRATION_ID'::uuid;