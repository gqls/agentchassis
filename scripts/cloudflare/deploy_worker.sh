#!/bin/bash
# Deploy scripts/cloudflare/worker.js to the portfolio-sites-router worker.
#
# Safer than the README's minimal recipe in one respect: the metadata is BUILT FROM
# THE LIVE SETTINGS (~/.cloudflare/portfolio-sites-router.settings.json) rather than
# hand-typed, so the two B2 bindings are carried forward verbatim — the README's
# warning is that an update without them strips the worker's credentials and takes
# every site down — and so are compatibility_date, compatibility_flags and the
# observability block, which the README's hand-typed metadata would silently reset.
#
# Usage: ./deploy_worker.sh <path-to-worker.js>
set -euo pipefail
SRC="$1"
ACC=13044f178ae0b156961065f55c8fada8
SETTINGS=~/.cloudflare/portfolio-sites-router.settings.json

[ -f "$SRC" ] || { echo "no such file: $SRC" >&2; exit 1; }

set -a; . ~/.cloudflare/404-token.env; set +a
: "${CLOUDFLARE_API_404_TOKEN:?token not loaded}"

META=$(python3 - "$SETTINGS" <<'PY'
import json, sys
r = json.load(open(sys.argv[1]))["result"]
bindings = r.get("bindings", [])
names = sorted(b.get("name") for b in bindings)
# refuse to deploy without BOTH B2 bindings — this is the outage guard
assert names == ["B2_APP_KEY", "B2_KEY_ID"], f"expected both B2 bindings, got {names}"
for b in bindings:
    assert b.get("text"), f"binding {b.get('name')} has no value"
meta = {
    "main_module": "worker.js",
    "compatibility_date": r.get("compatibility_date"),
    "bindings": bindings,
}
if r.get("compatibility_flags"):
    meta["compatibility_flags"] = r["compatibility_flags"]
if r.get("observability"):
    meta["observability"] = r["observability"]
print(json.dumps(meta))
PY
)

echo "metadata built: bindings=$(python3 -c "import json,sys;print(len(json.loads(sys.argv[1])['bindings']))" "$META"), compat=$(python3 -c "import json,sys;print(json.loads(sys.argv[1])['compatibility_date'])" "$META")"

RESP=$(curl -s -X PUT \
  "https://api.cloudflare.com/client/v4/accounts/$ACC/workers/scripts/portfolio-sites-router" \
  -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" \
  -F "metadata=${META};type=application/json" \
  -F "worker.js=@${SRC};type=application/javascript+module")

python3 - "$RESP" <<'PY'
import json, sys
try:
    d = json.loads(sys.argv[1])
except Exception:
    print("UNPARSEABLE RESPONSE:", sys.argv[1][:500]); raise SystemExit(1)
print("success:", d.get("success"))
if d.get("errors"):   print("errors:", json.dumps(d["errors"])[:600])
if d.get("messages"): print("messages:", json.dumps(d["messages"])[:300])
res = d.get("result") or {}
print("deployed_on:", res.get("modified_on") or res.get("created_on"))
print("bindings now:", sorted(b.get("name") for b in (res.get("bindings") or [])))
raise SystemExit(0 if d.get("success") else 1)
PY
