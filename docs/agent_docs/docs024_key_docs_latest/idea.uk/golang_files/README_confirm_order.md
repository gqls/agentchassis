KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_1781177552891442378"}'


(/decline to reject one instead.)
--------------
cd docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files

KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_1781177552891442378"}'