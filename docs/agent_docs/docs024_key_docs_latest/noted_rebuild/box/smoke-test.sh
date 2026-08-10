#!/usr/bin/env bash
set -uo pipefail
B="http://127.0.0.1:8082"; H="Host: app.noted.co.uk"; J="Content-Type: application/json"
PSQL="sudo -u postgres psql -d noted -q -tAc"
$PSQL "DELETE FROM accounts WHERE email_canonical LIKE 'smoke%@example.com'" >/dev/null 2>&1

hdrs() { curl -s -D- -o /dev/null -H "$H" -H "$J" -X POST -d "$2" "$B$1"; }
tok()  { sed -n 's/.*noted_session=\([^;]*\).*/\1/p' | tr -d '\r' | head -1; }

CRED='{"email":"smoke@example.com","password":"a-long-enough-password"}'
RESP=$(hdrs /api/register "$CRED")
TOK=$(printf '%s' "$RESP" | tok)
echo "1. register            : $([ -n "$TOK" ] && echo "session issued" || echo FAILED)"
echo "   cookie flags        : $(printf '%s' "$RESP" | grep -io 'HttpOnly\|Secure\|SameSite=[A-Za-z]*' | tr '\n' ' ')"

CK="Cookie: noted_session=$TOK"
echo "2. save a note         : $(curl -s -H "$H" -H "$CK" -H "$J" -X POST \
     -d '{"title":"Shopping","content":"milk and bread"}' $B/api/notes -w ' [%{http_code}]' | head -c 70)"

IMPORT='{"format":"noted.co.uk/full-backup","version":1,"notes":[{"id":"x1","title":"From my phone","content":"imported"}],"audio":{"x1":["data:audio/webm;base64,YWJjZA=="]},"images":{}}'
echo "3. import a backup     : $(curl -s -H "$H" -H "$CK" -H "$J" -X POST -d "$IMPORT" $B/api/import -w ' [%{http_code}]')"

T2=$(hdrs /api/login "$CRED" | tok)
echo "4. sign in AGAIN (as a different browser would) — are the notes there?"
curl -s -H "$H" -H "Cookie: noted_session=$T2" $B/api/notes \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); [print("     -", n["title"], "| recordings:", len(n.get("audio") or [])) for n in d["notes"]]'

echo "5. another account sees NOTHING of theirs:"
T3=$(hdrs /api/register '{"email":"smoke2@example.com","password":"a-long-enough-password"}' | tok)
curl -s -H "$H" -H "Cookie: noted_session=$T3" $B/api/notes \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); print("     notes visible to the other account:", len(d["notes"]))'

$PSQL "DELETE FROM accounts WHERE email_canonical LIKE 'smoke%@example.com'" >/dev/null 2>&1
echo "6. cleaned up"
