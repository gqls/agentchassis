#!/usr/bin/env bash
# Dynadot RESTful v2 client. Reads API_KEY + API_SECRET from
# ~/.config/dynadot/credentials, never prints either. Signs per the API docs:
# X-Signature = base64(HMAC-SHA256(secret, "API_KEY\n<path+query>\n<X-Request-ID or ''>\n<body or ''>")).
#
# Usage: dynadot-restful.sh <METHOD> </restful/v2/path?query> [json-body]
#   dynadot-restful.sh GET /restful/v2/domains/example.com/appraisal
#
# Prints the response body, then "HTTP <code>" on the last line. Exit 0 on
# HTTP 2xx, 1 otherwise (the body is still printed for diagnosis).
set -euo pipefail

CRED="${DYNADOT_CREDENTIALS:-$HOME/.config/dynadot/credentials}"
if [[ ! -r "$CRED" ]]; then
  echo "dynadot-restful.sh: no credentials at $CRED" >&2
  exit 2
fi
API_KEY=$(sed -n 's/^API_KEY=//p' "$CRED" | tr -d '[:space:]')
API_SECRET=$(sed -n 's/^API_SECRET=//p' "$CRED" | tr -d '[:space:]')
if [[ -z "$API_KEY" || -z "$API_SECRET" ]]; then
  echo "dynadot-restful.sh: need both API_KEY= and API_SECRET= lines in $CRED" >&2
  exit 2
fi

if [[ $# -lt 2 ]]; then
  echo "usage: dynadot-restful.sh <METHOD> </restful/v2/path?query> [json-body]" >&2
  exit 2
fi
method="$1"; path="$2"; body="${3:-}"

# Empty X-Request-ID line between path and body, per the documented sign string.
sign_string="$API_KEY
$path

$body"
sig=$(SEC="$API_SECRET" STR="$sign_string" python3 -c '
import os, hmac, hashlib, base64
print(base64.b64encode(hmac.new(os.environ["SEC"].encode(), os.environ["STR"].encode(), hashlib.sha256).digest()).decode())')

curl_args=(-sS --max-time 60 -X "$method"
  -H "Authorization: Bearer $API_KEY"
  -H "Content-Type: application/json"
  -H "Accept: application/json"
  -H "X-Signature: $sig"
  -w '\nHTTP %{http_code}\n')
if [[ -n "$body" ]]; then curl_args+=(--data "$body"); fi

out=$(curl "${curl_args[@]}" "https://api.dynadot.com$path")
printf '%s\n' "$out"
[[ "$out" == *"HTTP 2"* ]] || exit 1
