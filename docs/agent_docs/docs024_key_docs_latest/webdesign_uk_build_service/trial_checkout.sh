#!/usr/bin/env bash
# trial_checkout.sh — mint a ruled voucher and create one order + Stripe
# checkout link through the REAL admin API (PAY-009 has no screen yet, and
# both routes are admin-JWT-gated by design — so the OWNER runs this himself,
# in his own terminal; the password and JWT live in this process only and are
# never written anywhere).
#
# Usage:
#   ./trial_checkout.sh <client_id> <BR-reference> [pence 1000|3000|5500] [recipient] [customer-email]
#
# Worked example — owner customer-zero trial run 1 (2026-08-27, chat brief
# BR-9AUZ59, client row created that day):
#   ./trial_checkout.sh a7395f69-e735-4390-98d7-9f17085338f4 BR-9AUZ59 3000 "Boxing Online" aaa@designconsultancy.co.uk
#
# What it does: port-forwards auth-service, logs YOU in, mints a single-use
# 14-day voucher at the given pence (ruled variants only — the API refuses
# anything else), creates the order carrying your BR- reference and the
# voucher, and prints the Stripe checkout URL. Pay there; the webhook marks
# the order paid; collect_external_orders releases the brief on a PAID order
# carrying the reference.
set -euo pipefail

CLIENT_ID=${1:?usage: trial_checkout.sh <client_id> <BR-reference> [pence] [recipient] [email]}
REF=${2:?a BR- reference from the chat}
PENCE=${3:-3000}
RECIPIENT=${4:-owner trial}
EMAIL=${5:-}
NS=ai-persona-system
PORT=18081

kubectl -n "$NS" port-forward svc/auth-service "$PORT:8081" >/dev/null 2>&1 &
PF=$!
trap 'kill "$PF" 2>/dev/null || true' EXIT
for _ in $(seq 1 20); do
  curl -s -o /dev/null "http://127.0.0.1:$PORT/" 2>/dev/null && break
  sleep 0.5
done

read -rp  "admin email: " ADMIN_EMAIL
read -rsp "admin password: " ADMIN_PASS; echo
JWT=$(curl -sf -X POST "http://127.0.0.1:$PORT/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])')
unset ADMIN_PASS
[ -n "$JWT" ] || { echo "login failed"; exit 1; }

VOUCHER=$(curl -sf -X POST "http://127.0.0.1:$PORT/api/v1/admin/billing/vouchers" \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d "{\"drops_price_to_pence\":$PENCE,\"recipient_name\":\"$RECIPIENT\",\"ttl_days\":14}")
CODE=$(echo "$VOUCHER" | python3 -c 'import json,sys; print(json.load(sys.stdin)["code"])')
echo "voucher: $CODE  (single-use, expires in 14 days, drops the price to ${PENCE}p)"

ORDER=$(curl -sf -X POST "http://127.0.0.1:$PORT/api/v1/admin/billing/orders" \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d "{\"client_id\":\"$CLIENT_ID\",\"voucher_code\":\"$CODE\",\"email\":\"$EMAIL\",\"external_reference\":\"$REF\"}")
echo "$ORDER" | python3 -c '
import json, sys
o = json.load(sys.stdin)
print("order:", json.dumps(o.get("order", o)))
print()
print("CHECKOUT:", o.get("checkout_url", "(no checkout_url in response — read the order JSON above)"))'
