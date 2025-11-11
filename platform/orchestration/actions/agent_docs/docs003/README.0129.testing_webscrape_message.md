Test Smart Analyzer with Basic Scraping:

kubectl -n kafka run -i --rm kcat-producer-analyzer \
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
-H responses_topic=system.responses.generic <<EOF
{"action":"orchestrate","config":{"group_type":"website-analyzer"},"input_data":{"target_url":"boxing-tickets.com","extract_structured":false,"crawl_pages":false}}
EOF


{
"action": "orchestrate",
"config": {
"group_type": "website-analyzer"
},
"input_data": {
"target_url": "boxing-tickets.com",
"extract_structured": false,
"crawl_pages": false
}
}


-----

Test with Structured Extraction and Crawling:


kubectl -n kafka run -i --rm kcat-producer-full \
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
-H responses_topic=system.responses.generic <<EOF
{"action":"orchestrate","config":{"group_type":"website-analyzer"},"input_data":{"target_url":"boxing-tickets.com","extract_structured":true,"crawl_pages":true,"crawl_limit":3,"crawl_depth":1}}
EOF


{
"action": "orchestrate",
"config": {
"group_type": "website-analyzer"
},
"input_data": {
"target_url": "boxing-tickets.com",
"extract_structured": true,
"crawl_pages": true,
"crawl_limit": 3,
"crawl_depth": 1
}
}
