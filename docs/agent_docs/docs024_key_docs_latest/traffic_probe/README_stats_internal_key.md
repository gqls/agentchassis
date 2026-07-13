internal api key for e.g. stats

ssh root@167.233.33.159 "grep -E '^(INTERNAL_API_KEY|THANKS_PATH)=' /etc/site-engine/site-engine.env"

ant@ant-XPS-15-9500:~/projects/vm-sites$ ssh root@167.233.33.159 "grep -E '^(INTERNAL_API_KEY|THANKS_PATH)=' /etc/site-engine/site-engine.env"
THANKS_PATH=/gracias.html
INTERNAL_API_KEY=8194197235bfabce2ac8f50f44efd037d45aeb62f05296a7

# get stats
curl -sS -H "X-Internal-Key: 8194197235bfabce2ac8f50f44efd037d45aeb62f05296a7" https://relojistas.com/stats


---


ssh root@167.233.33.159 "wc -l /var/log/nginx/access.log; grep -c ' /api/hit' /var/log/nginx/access.log"
ssh root@167.233.33.159 "awk '{print \$9}' /var/log/nginx/access.log | sort | uniq -c | sort -rn"          # status codes — 301 flood = redirect loop
ssh root@167.233.33.159 "awk -F'\"' '{print \$6}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head"  # user-agent mix
ssh root@167.233.33.159 "awk '{print \$7}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head"    # which paths


ssh root@167.233.33.159  'tail -3 /var/lib/site-engine/events-*.jsonl; cat /var/lib/site-engine/counters.json'


----


