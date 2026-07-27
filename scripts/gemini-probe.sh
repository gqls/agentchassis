#!/usr/bin/env bash
# FILE: scripts/gemini-probe.sh
#
# Answers the three questions that have to be settled against the LIVE key
# before any agent is pointed at Gemini. On 2026-07-24 they were answered by
# hand, mid-incident, and the answers were lost in a commit message.
#
#   1. Which models can THIS key actually reach? Google closes pinned model
#      generations to newly-issued keys with a 404, not a deprecation warning,
#      so the answer is a property of the key and not of the model list.
#   2. At each of the platform's real max_tokens tiers, how much VISIBLE text
#      comes back, and how much of the ceiling went on thinking?
#   3. Which thinking knob does the model accept — thinkingLevel (3.x) or
#      thinkingBudget (2.5)? Sending the wrong one is a 400 on every call.
#
# Usage:
#   GEMINI_API_KEY=... scripts/gemini-probe.sh              # list models only
#   GEMINI_API_KEY=... scripts/gemini-probe.sh <model>      # full probe
#   scripts/gemini-probe.sh --from-pod <model>              # read key from the cluster
#
# --from-pod pulls GEMINI_API_KEY out of a running content-creator pod, so the
# probe exercises the same key production would use. It never prints the key.

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
API_BASE="https://generativelanguage.googleapis.com/v1beta/models"

# The platform's real budgets: content-creator's four tiers after 3ea9d718c
# (twitter 100 / short 1200 / default 3000 / long 6000) plus the 500 that
# page-content-writer-class steps commonly carry. 100 and 500 are the two that
# failed in production, so they lead.
TIERS="${TIERS:-100 500 1200 3000 6000}"

# A prompt short enough to be cheap and specific enough that a thinking model
# has something to think about — an empty-ish prompt under-reports thinking.
PROMPT="${PROMPT:-Write three short sentences describing what an AI consultancy does. Plain English, no em dashes.}"

die() { printf '%s\n' "$*" >&2; exit 1; }

command -v curl >/dev/null || die "curl not found"
command -v jq   >/dev/null || die "jq not found"

if [[ "${1:-}" == "--from-pod" ]]; then
  shift
  POD=$(kubectl get pods -n "$NAMESPACE" -l app=content-creator-agent \
          -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
  [[ -n "$POD" ]] || die "no content-creator-agent pod found in $NAMESPACE (is kubectl authenticated?)"
  GEMINI_API_KEY=$(kubectl exec -n "$NAMESPACE" "$POD" -- printenv GEMINI_API_KEY 2>/dev/null | tr -d '\r\n')
  [[ -n "$GEMINI_API_KEY" ]] || die "GEMINI_API_KEY is not set in pod $POD"
  echo "# key read from pod $POD (value not printed)"
fi

[[ -n "${GEMINI_API_KEY:-}" ]] || die "GEMINI_API_KEY not set (or pass --from-pod)"

MODEL="${1:-}"

# ── 1. What this key can reach ───────────────────────────────────────────────
echo
echo "=== models reachable by THIS key (generateContent) ==="
models_json=$(curl -sS -H "x-goog-api-key: $GEMINI_API_KEY" "$API_BASE?pageSize=200")
if ! jq -e '.models' >/dev/null 2>&1 <<<"$models_json"; then
  echo "$models_json" | head -20
  die "could not list models — the response above is what the API said"
fi
jq -r '.models[]
       | select(.supportedGenerationMethods // [] | index("generateContent"))
       | "\(.name | sub("^models/";""))\t\(.outputTokenLimit // "?")"' <<<"$models_json" \
  | sort | awk -F'\t' 'BEGIN{printf "%-40s %s\n","MODEL","OUTPUT LIMIT"} {printf "%-40s %s\n",$1,$2}'

echo
echo "# Note: this list is what the API advertises. Reachability is per-key —"
echo "# a listed pinned snapshot can still 404 \"not available to new users\"."
echo "# The tier probe below is the only proof a model works for this key."

if [[ -z "$MODEL" ]]; then
  echo
  echo "Pass a model name to run the tier + thinking probe, e.g.:"
  echo "  GEMINI_API_KEY=... $0 gemini-pro-latest"
  exit 0
fi

# ── 2. Visible text per tier ─────────────────────────────────────────────────
# maxOutputTokens is a TOTAL ceiling: thinking is spent from it before any
# visible text. This table is what tells you the reserve to configure.
echo
echo "=== $MODEL — visible output vs thinking, per max_tokens tier ==="
printf '%-8s %-14s %-10s %-10s %-8s %s\n' TIER FINISH VIS_TOK THINK_TOK CHARS PREVIEW
for tier in $TIERS; do
  resp=$(curl -sS -H "x-goog-api-key: $GEMINI_API_KEY" -H 'Content-Type: application/json' \
    -X POST "$API_BASE/$MODEL:generateContent" \
    -d "$(jq -n --arg p "$PROMPT" --argjson t "$tier" \
          '{contents:[{role:"user",parts:[{text:$p}]}],generationConfig:{maxOutputTokens:$t}}')")

  if err=$(jq -er '.error.message' <<<"$resp" 2>/dev/null); then
    printf '%-8s %s\n' "$tier" "ERROR: ${err:0:100}"
    continue
  fi

  finish=$(jq -r '.candidates[0].finishReason // "-"' <<<"$resp")
  vis=$(jq -r '.usageMetadata.candidatesTokenCount // 0' <<<"$resp")
  think=$(jq -r '.usageMetadata.thoughtsTokenCount // 0' <<<"$resp")
  # Only non-thought parts are answer — the same filter the Go client applies.
  text=$(jq -r '[.candidates[0].content.parts[]? | select(.thought != true) | .text] | join("")' <<<"$resp")
  printf '%-8s %-14s %-10s %-10s %-8s %s\n' \
    "$tier" "$finish" "$vis" "$think" "${#text}" "$(printf '%s' "${text:0:48}" | tr '\n' ' ')"
done

echo
echo "# Read it this way: CHARS=0 with THINK_TOK>0 is the 2026-07-24 failure —"
echo "# thinking ate the ceiling. The reserve to configure is the largest"
echo "# THINK_TOK you see here, with headroom. It is a ceiling, not a purchase."

# ── 3. Which thinking knob is accepted ───────────────────────────────────────
echo
echo "=== $MODEL — thinking knobs ==="
probe_knob() {
  local label="$1" cfg="$2"
  local resp
  resp=$(curl -sS -H "x-goog-api-key: $GEMINI_API_KEY" -H 'Content-Type: application/json' \
    -X POST "$API_BASE/$MODEL:generateContent" \
    -d "$(jq -n --arg p "$PROMPT" --argjson tc "$cfg" \
          '{contents:[{role:"user",parts:[{text:$p}]}],generationConfig:({maxOutputTokens:2048}+$tc)}')")
  if err=$(jq -er '.error.message' <<<"$resp" 2>/dev/null); then
    printf '  %-34s REJECTED: %s\n' "$label" "${err:0:90}"
  else
    printf '  %-34s ACCEPTED (thinking=%s tokens, visible=%s)\n' "$label" \
      "$(jq -r '.usageMetadata.thoughtsTokenCount // 0' <<<"$resp")" \
      "$(jq -r '.usageMetadata.candidatesTokenCount // 0' <<<"$resp")"
  fi
}
probe_knob 'thinkingConfig.thinkingLevel="low"'  '{thinkingConfig:{thinkingLevel:"low"}}'
probe_knob 'thinkingConfig.thinkingBudget=0'     '{thinkingConfig:{thinkingBudget:0}}'
probe_knob 'thinkingConfig.thinkingBudget=512'   '{thinkingConfig:{thinkingBudget:512}}'

echo
echo "# Set the ACCEPTED one in ai_service (thinking_level / thinking_budget_tokens)."
echo "# If both are rejected, send neither and let thinking_reserve_tokens absorb it —"
echo "# that is the client's default and it needs no knob."
