KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_1781177552891442378"}'


(/decline to reject one instead.)
--------------
cd docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files

KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_1781177552891442378"}'

-------------
# get order detail from box (or from clook leopardess mailbox)
ssh root@116.203.204.115 "python3 -c \"import json;o=json.load(open('/var/lib/idea/orders.json'))['orders'];r=sorted([x for x in o.values() if x['status']=='requested'],key=lambda x:x['created_at']);print(r[-1]['id'], r[-1]['email'])\""

# or if on the box - without backslashes
python3 -c "import json;o=json.load(open('/var/lib/idea/orders.json'))['orders'];r=sorted([x for x in o.values() if x['status']=='requested'],key=lambda x:x['created_at']);print(r[-1]['id'], r[-1]['email'])"

ord_1781205749466777546 aaa@designconsultancy.co.uk


4. Confirm it on the box (this is what sends the pay-link, via SES):
5. ssh root@116.203.204.115
   KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
   curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' \
   -d '{"order_id":"ord_1781205749466777546"}'

------------
# what went wrong

journalctl -u idea --since "20 min ago" --no-pager | grep -iE "email|smtp|deliver|fail"

--------------

# Quick sequence:

On https://idea.uk, place a request — use a recipient you can check easily (aaa@ is fine, or a plain Gmail).
Grab the new order id from the box:

ssh root@116.203.204.115 "python3 -c \"import json;o=json.load(open('/var/lib/idea/orders.json'))['orders'];r=sorted([x for x in o.values() if x['status']=='requested'],key=lambda x:x['created_at']);print(r[-1]['id'], r[-1]['email'])\""

Confirm that id:

KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"<new id>"}'

Read the log straight after:

journalctl -u idea --since "3 min ago" --no-pager | grep -iE "email|smtp|fail"