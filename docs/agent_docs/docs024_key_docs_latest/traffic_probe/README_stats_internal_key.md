internal api key for e.g. stats

ssh root@167.233.33.159 "grep -E '^(INTERNAL_API_KEY|THANKS_PATH)=' /etc/site-engine/site-engine.env"

ant@ant-XPS-15-9500:~/projects/vm-sites$ ssh root@167.233.33.159 "grep -E '^(INTERNAL_API_KEY|THANKS_PATH)=' /etc/site-engine/site-engine.env"
THANKS_PATH=/gracias.html
INTERNAL_API_KEY=8194197235bfabce2ac8f50f44efd037d45aeb62f05296a7

# get stats
curl -sS -H "X-Internal-Key: 8194197235bfabce2ac8f50f44efd037d45aeb62f05296a7" https://relojistas.com/stats

