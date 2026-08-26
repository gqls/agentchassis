#!/usr/bin/env bash
# find_customer_domain.sh — the domain-finding workflow (owner brief 2026-08-26):
# choose the BEST GENERIC domain to register for a client. Rules, owner-verbatim:
# never their brand names ("every domain must be generic"); as short as possible,
# as few words as possible; hyphens only if the good clean .co.uk/.uk options are
# taken.
#
# Shape (severable, zero platform surface): candidate stems come from the
# framework's own LLM step fired as an INLINE workflow over generic requests (the
# proven 077/envelope pattern — no new agent type, nothing dispatches this unless
# an operator runs it); availability is the registry's own answer via the proven
# EPP check client (VMB-016) from a cluster pod. Part of the opt-in domain layer:
# BRIEF_2026-08-26, rulings section.
#
# Usage:
#   ./find_customer_domain.sh --brief-file <txt> --ban "Brand One,BrandTwo" [--stems N]
#   ./find_customer_domain.sh --site <domain>            # brief + brand bans from the DB
#
# Output: a ranked table of AVAILABLE domains — tier 1 clean .uk shortest-first,
# then clean .co.uk, then (only if clean options ran out) hyphenated fallbacks.
# Read-only against Nominet; registration stays a separate owner-gated step
# (nominet-epp-domain-register.py --apply).
set -euo pipefail
cd "$(git -C "$(dirname "$0")" rev-parse --show-toplevel 2>/dev/null || echo /home/ant/projects/agentchassis)"

BRIEF_FILE=""; SITE=""; BAN=""; STEMS=12
while [ $# -gt 0 ]; do case "$1" in
  --brief-file) BRIEF_FILE="$2"; shift 2;;
  --site) SITE="$2"; shift 2;;
  --ban) BAN="$2"; shift 2;;
  --stems) STEMS="$2"; shift 2;;
  *) echo "unknown arg $1" >&2; exit 2;;
esac; done
[ -n "$BRIEF_FILE" ] || [ -n "$SITE" ] || { echo "need --brief-file or --site" >&2; exit 2; }

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA)
W=$(mktemp -d)
trap 'rm -rf "$W"' EXIT

# ── 1. the brief text + brand ban tokens ──
if [ -n "$SITE" ]; then
  "${PSQL[@]}" -c "SELECT COALESCE(ss.data->>'about_us','') || E'\n' || COALESCE((SELECT x.data->>'reasoning' FROM site_specs x JOIN sites s2 ON s2.id=x.site_id WHERE s2.domain='$SITE' AND x.aspect='classification' AND x.is_current),'') FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='$SITE' AND ss.aspect='briefing' AND ss.is_current;" </dev/null > "$W/brief.txt"
  DBBAN=$("${PSQL[@]}" -c "SELECT COALESCE(company_name,'') || ',' || COALESCE(name,'') FROM sites WHERE domain='$SITE';" </dev/null)
  BAN="$BAN,$DBBAN,$SITE"
else
  cp "$BRIEF_FILE" "$W/brief.txt"
fi
[ -s "$W/brief.txt" ] || { echo "empty brief" >&2; exit 1; }

# ── 2. fire the inline generation workflow ──
CORR=$(cat /proc/sys/kernel/random/uuid)
BAN="$BAN" BRIEF="$(cat "$W/brief.txt")" CORR="$CORR" W="$W" python3 - << 'PYEOF'
import json, os, uuid, datetime
ban = sorted({t.strip() for t in os.environ["BAN"].split(",") if t.strip()})
prompt = (
"You are choosing GENERIC domain-name stems for a small UK business website.\n"
"THE BUSINESS, from its brief:\n---\n" + os.environ["BRIEF"][:4000] + "\n---\n"
"RULES, all binding:\n"
"1. GENERIC ONLY. The stem describes what the business DOES (its trade, service, product,"
" optionally its locality) in words any competitor could also use. NEVER use the business's"
" own name, any part of it, or any brand token. FORBIDDEN tokens (and close variants): "
+ (", ".join(ban) if ban else "(none supplied)") + ".\n"
"2. AS SHORT AS POSSIBLE, AS FEW WORDS AS POSSIBLE. One tight word or compound is best"
" (plumbleeds, bakewell); two short words acceptable (leedsplumber); never three or more.\n"
"3. Lowercase a-z only. No digits, no punctuation. UK English spelling.\n"
"4. NO HYPHENS in tier 1. Tier 2 is the hyphenated fallback only: the same ideas as"
" two-word hyphenated forms (leeds-plumber).\n"
"5. Stems only, no TLD: .uk and .co.uk are appended by the checker.\n"
"6. Aim for stems a person can say aloud and type first time.\n"
"OUTPUT: STRICT JSON only, no fences, no commentary: an array of at least 20 tier-1 and"
" 8 tier-2 objects, ordered best-first: [{\"stem\":\"leedsplumber\",\"tier\":1}, ...]"
)
wf = {"start_step":"generate","processing_mode":"orchestrator","timeout_seconds":300,
 "steps":{"generate":{"action":"execute_llm_prompt",
   "config":{"prompt":prompt,"ai_service":{"provider":"anthropic","model":"claude-sonnet-4-6","api_key_env_var":"ANTHROPIC_API_KEY","max_tokens":4000},"input_fields":[]},
   "output_field":"candidates_raw","next_step":"complete"},
  "complete":{"action":"complete_workflow","config":{"output_fields":["candidates_raw"]}}}}
msg = {"headers":{"correlation_id":os.environ["CORR"],"orchestration_id":str(uuid.uuid4()),
 "request_id":str(uuid.uuid4()),"message_id":str(uuid.uuid4()),"message_type":"request",
 "client_id":"system","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},
 "timestamp":datetime.datetime.now(datetime.timezone.utc).isoformat()},
 "config":{"workflow":wf},"input_data":{"task":"domain-candidate-generation"}}
open(os.environ["W"]+"/gen.json","w").write(json.dumps(msg,separators=(",",":")))
print("prompt bytes:", len(prompt))
PYEOF
. scripts/kafka-publish-lib.sh
kafka_publish_checked --topic system.agent.generic.requests --correlation "$CORR" \
  --header "orchestration_id=$(cat /proc/sys/kernel/random/uuid)" --header "request_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "message_id=$(cat /proc/sys/kernel/random/uuid)" --header "message_type=request" \
  --header "client_id=system" --header "action=process" --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$(date -u +%Y-%m-%dT%H:%M:%S%z)" --payload "$(cat "$W/gen.json")"

echo "generation fired ($CORR); waiting (first row can take minutes)..." >&2
RAW=""
for i in $(seq 1 20); do
  ROW=$("${PSQL[@]}" -c "SELECT status FROM orchestration_states WHERE correlation_id='$CORR';" </dev/null || true)
  if [ "$ROW" = "COMPLETED" ]; then
    "${PSQL[@]}" -c "SELECT COALESCE(final_result->>'candidates_raw', collected_data->>'candidates_raw','') FROM orchestration_states WHERE correlation_id='$CORR';" </dev/null > "$W/raw.txt"
    RAW=ok; break
  elif [ "$ROW" = "FAILED" ]; then
    "${PSQL[@]}" -c "SELECT left(COALESCE(error, collected_data->>'__step_error',''),300) FROM orchestration_states WHERE correlation_id='$CORR';" </dev/null >&2
    echo "generation FAILED" >&2; exit 1
  fi
  sleep 30
done
[ "$RAW" = ok ] || { echo "no COMPLETED row after 10 min (receipt was seen — do not re-fire blindly; check the topic/orchestration)" >&2; exit 1; }

# ── 3. rank stems, build the check list ──
STEMS="$STEMS" W="$W" python3 - << 'PYEOF'
import json, os, re
raw = open(os.environ["W"]+"/raw.txt").read().strip()
m = re.search(r"\[.*\]", raw, re.S)
cands = json.loads(m.group(0) if m else raw)
seen, t1, t2 = set(), [], []
for c in cands:
    stem = str(c.get("stem","")).strip().lower()
    tier = int(c.get("tier",1))
    if not stem or stem in seen: continue
    if tier == 1 and not re.fullmatch(r"[a-z]{3,20}", stem): continue
    if tier == 2 and not re.fullmatch(r"[a-z]{2,14}-[a-z]{2,14}", stem): continue
    seen.add(stem); (t1 if tier==1 else t2).append(stem)
t1.sort(key=len); t2.sort(key=len)
n = int(os.environ["STEMS"])
picked = t1[:n] + t2[:max(4, n//3)]
with open(os.environ["W"]+"/checklist.txt","w") as f:
    for s in picked:
        f.write(s+".uk\n"); f.write(s+".co.uk\n")
json.dump({"t1":t1,"t2":t2}, open(os.environ["W"]+"/stems.json","w"))
print(f"stems: tier1={len(t1)} tier2={len(t2)}; checking {len(picked)*2} names")
PYEOF

# ── 4. availability at the registry (read-only), from a cluster pod ──
POD="epp-find-$RANDOM"
kubectl -n ai-persona-system run "$POD" --image=python:3.12-slim --restart=Never --command -- sleep 600 >/dev/null
kubectl -n ai-persona-system wait --for=condition=Ready "pod/$POD" --timeout=90s >/dev/null
kubectl -n ai-persona-system cp docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/box/nominet-epp-domain-check.py "$POD":/tmp/check.py
# shellcheck disable=SC2046
kubectl -n ai-persona-system exec "$POD" -- env NOMINET_EPP_PW="$(cat ~/.config/nominet/epp-password)" \
  python3 /tmp/check.py --tag DESIGNCONSULT $(tr '\n' ' ' < "$W/checklist.txt") > "$W/avail.txt" 2>"$W/avail.err" || true
kubectl -n ai-persona-system delete pod "$POD" --wait=false >/dev/null

# ── 5. the ranked answer ──
W="$W" python3 - << 'PYEOF'
import os, re
avail = set()
for line in open(os.environ["W"]+"/avail.txt"):
    m = re.match(r"AVAILABLE\s+(\S+)", line)
    if m: avail.add(m.group(1))
def rank(names):  # shortest stem first, .uk before .co.uk at equal stems
    return sorted(names, key=lambda d: (len(d.split(".")[0]), 0 if d.endswith(".uk") and not d.endswith(".co.uk") else 1, d))
t1uk  = rank([d for d in avail if "-" not in d and not d.endswith(".co.uk")])
t1co  = rank([d for d in avail if "-" not in d and d.endswith(".co.uk")])
t2    = rank([d for d in avail if "-" in d])
print("\n=== AVAILABLE, ranked (register with nominet-epp-domain-register.py --apply) ===")
for label, group in (("CLEAN .uk", t1uk), ("CLEAN .co.uk", t1co), ("HYPHENATED FALLBACK (use only if the above are unsuitable)", t2)):
    if group:
        print(f"-- {label}:")
        for d in group: print(f"     {d}")
if not (t1uk or t1co):
    print("!! no clean option available — widen --stems or revisit the brief's vocabulary")
print("(taken names and full transcript: see the run's avail.txt if kept)")
PYEOF
