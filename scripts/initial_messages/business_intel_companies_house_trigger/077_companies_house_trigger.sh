kubectl -n kafka run -i --rm kcat-ch-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.vet-intel.requests \
    -H "correlation_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "request_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "message_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "orchestration_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "orchestration_name=ch-enrich-$(date +%Y%m%d-%H%M%S)" \
    -H "step_name=start" \
    -H "client_id=vetcomparison" \
    -H "message_type=request" \
    -H "action=orchestrate" \
    -H "from_agent_type=user" \
    -H "from_agent_id=cli" \
    -H "responses_topic=system.agent.vet-intel.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"ch-enricher"},"input_data":{"batch_size":5,"vertical_slug":"veterinary"}}
JSON