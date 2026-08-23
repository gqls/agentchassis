#!/usr/bin/env bash
# kafka-publish-lib.sh — publishing to Kafka with a RECEIPT, single-sourced (bugs_open/327).
#
# SOURCED, NEVER EXECUTED (the one exception is `bash scripts/kafka-publish-lib.sh
# --self-test`, guarded at the foot of this file). The file is mode 664 for the same
# reason scripts/council-scope.sh is: a fragment that is not meant to be run should not
# look runnable.
#
#   . "$REPO_ROOT/scripts/kafka-publish-lib.sh"
#
# ---------------------------------------------------------------------------
# WHY THIS FILE EXISTS
# ---------------------------------------------------------------------------
# `kubectl run -i` attaches stdin ASYNCHRONOUSLY. If the container reaches `kcat -P`
# before stdin is wired, kcat sees EOF, publishes ZERO messages and exits 0 — and
# `--rm` then deletes the evidence. Measured 2026-07-26: four publishes in five lost.
# LANDMINES.md, "`kubectl run -i --rm … kcat -P < file` drops roughly 4 publishes in 5
# AT EXIT 0".
#
# The remedy has been documented for a month. Measured 2026-08-23, across the repo:
#
#     218  shell scripts publish via `kcat -P`
#     200  still use the racing `kubectl run -i` + stdin form
#      25  print a PUBLISH_OK receipt
#       2  ACTUALLY ASSERT ON IT
#
# That last line is why this is a library and not another documented recipe. Twenty-three
# scripts followed the guidance, emitted the receipt, and still exit 0 when it is absent —
# because the receipt was written as advice to copy rather than as something you can call.
# A receipt nobody asserts on is a log line, not a control.
#
# ---------------------------------------------------------------------------
# THE THING THIS BUYS YOU: (a) AND (b) STOP LOOKING ALIKE
# ---------------------------------------------------------------------------
# Two failures produce an identical missing row, and the correct response to each is the
# OPPOSITE of the other:
#
#   (a) NEVER PUBLISHED      -> retry immediately; nothing landed, a retry collides with
#                               nothing.
#   (b) PUBLISHED, NOT LANDED -> do NOT retry; a duplicate costs a whole round. This is
#                               also the documented signature of ordinary queue latency,
#                               which CLAUDE.md tells you not to retry on.
#
# Until now the operator could not tell which they were in. Now the exit code says:
#
#     0   published, receipt seen                    (kafka_publish_checked)
#    10   NOT PUBLISHED — retry now                  (kafka_publish_checked)
#    11   RECEIPT INDETERMINATE — verify, don't retry (kafka_publish_checked)
#     0   landed                                     (kafka_verify_landing)
#    12   CONSUMED AND REFUSED — do not resend as-is (kafka_verify_landing)
#    13   PUBLISHED, NOT LANDED — wait and re-poll   (kafka_verify_landing)
#     2   usage / precondition error                 (both)
#
# Codes start at 10 deliberately. Consumers of this library already own 1 and 2 — the
# council trigger's contract is "1 = hard error, 2 = REFUSED out of scope"
# (097_TRIGGER_council_review_v1.sh:111-114, mirrored in council-scope.sh) — so a publish
# outcome must never collide with a caller's own vocabulary.
#
# ---------------------------------------------------------------------------
# THREE TRAPS, ONE CONTROL
# ---------------------------------------------------------------------------
# Verified 2026-08-23 against an unreachable broker (so nothing was published):
#
#   trap                              exit code today   asserted receipt catches it
#   stdin race / empty stdin          0  — NO SIGNAL    yes
#   broker unreachable                1                 yes
#   `--command` omitted               1                 yes
#
# Only the first is silent, and it is the one that produced bugs_open/327. That is the
# whole argument for asserting rather than printing.
#
# ⚠ `--command` IS LOAD-BEARING. The edenhill/kcat image's ENTRYPOINT *is* kcat, so
# without `--command` your `sh -c …` arrives as ARGUMENTS TO KCAT: you get kcat's usage
# text and nothing is published — the same silent zero reached a different way.
#
# ⚠ THE PAYLOAD TRAVELS IN THE CONTAINER COMMAND, NOT ON STDIN. That is what makes the
# race structurally impossible rather than merely unlikely. Do not "simplify" this back
# to a heredoc or a pipe.
#
# ⚠ ONE MESSAGE PER LINE. `kcat -P` splits stdin on newlines and publishes each line as a
# SEPARATE message, so a pretty-printed JSON payload becomes N broken messages. This
# library REFUSES a multi-line payload rather than silently mangling it; build yours with
# `jq -c` (097:249-251 already does, for exactly this reason). The base64 hop protects the
# payload from the shell in the same move.
#
# ---------------------------------------------------------------------------
# WHAT THIS DOES NOT DO
# ---------------------------------------------------------------------------
# It does NOT retry. On an indeterminate receipt an automatic retry is a double-publish
# engine: the one case where you cannot tell whether the message landed is exactly the
# case where sending it again is most dangerous. Retry is the caller's decision, taken
# after `kafka_verify_landing` has said which world it is in.
#
# It is also NOT a broker ack. The receipt is scraped from the pod's output via
# `kubectl attach`, which is weaker than the acknowledgement the broker itself gives.
# platform/kafka/producer.go is already `RequiredAcks: kafka.RequireAll, Async: false` —
# an in-cluster Go submit path would make a silent drop UNREPRESENTABLE rather than
# merely detected. That is the right destination and it needs an image build, a council
# round and a fleet roll; this library needs none of those and protects the customer path
# today. See the concept register entry for the deferred design.
#
# ---------------------------------------------------------------------------
# CONVENTIONS FOR CALLERS
# ---------------------------------------------------------------------------
# Every internal grep carries `|| true`: a no-match grep exits 1, and callers run under
# `set -euo pipefail`. No function here exits; they all return. (council-scope.sh records
# what happens when that discipline slips — an unguarded grep killed every submission.)

# The fleet broker. Override per call with --broker if you must.
KAFKA_PUBLISH_BROKER="${KAFKA_PUBLISH_BROKER:-personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092}"
KAFKA_PUBLISH_IMAGE="${KAFKA_PUBLISH_IMAGE:-edenhill/kcat:1.7.1}"
KAFKA_PUBLISH_NAMESPACE="${KAFKA_PUBLISH_NAMESPACE:-kafka}"
# librdkafka's default delivery timeout is ~5 minutes, which turns an unreachable broker
# into a five-minute hang that reads as a slow cluster. Bound it.
KAFKA_PUBLISH_TIMEOUT_MS="${KAFKA_PUBLISH_TIMEOUT_MS:-15000}"
KAFKA_PUBLISH_RECEIPT="PUBLISH_OK"

# Where kafka_verify_landing looks. Both are read-only queries.
KAFKA_VERIFY_NAMESPACE="${KAFKA_VERIFY_NAMESPACE:-ai-persona-system}"
KAFKA_VERIFY_POD="${KAFKA_VERIFY_POD:-postgres-clients-0}"
KAFKA_VERIFY_DB="${KAFKA_VERIFY_DB:-clients_db}"
KAFKA_VERIFY_USER="${KAFKA_VERIFY_USER:-clients_user}"

# Set by kafka_publish_checked so a caller can print/verify the id it actually used.
KAFKA_PUBLISH_CORRELATION=""
KAFKA_PUBLISH_OUTPUT=""

# _kafka_shquote — single-quote one string for safe interpolation into the `sh -c` body.
# The header values reach the container inside a double-quoted shell string, so a space,
# a `$` or a quote in an orchestration_name would otherwise re-split or expand there.
_kafka_shquote() {
  printf "'%s'" "$(printf '%s' "$1" | sed "s/'/'\\\\''/g")"
}

# _kafka_classify — the decision, split out from the doing so it can be tested with no
# cluster. $1 = kubectl's exit status, $2 = its combined output. Echoes the code.
#
# The three arms, and why the middle one is not "assume failure":
#   receipt present            -> 0   it published; this is the only positive evidence
#   receipt absent, status !=0 -> 10  the publisher ran and failed. Nothing landed, so a
#                                     retry is safe AND correct.
#   receipt absent, status ==0 -> 11  we cannot tell. kubectl was happy but the marker
#                                     never reached us — a lost attach stream looks like
#                                     this, and so would a publish we simply did not see.
#                                     Calling this a failure would invite a duplicate
#                                     publish, which is the one thing worse than the bug.
_kafka_classify() {
  local status="$1" output="$2"
  if printf '%s' "$output" | grep -q "$KAFKA_PUBLISH_RECEIPT"; then
    printf '0'
  elif [ "$status" -ne 0 ]; then
    printf '10'
  else
    printf '11'
  fi
}

# kafka_publish_checked --topic T --payload JSON [--correlation ID] [--broker B]
#                       [--timeout-ms N] [--header k=v]...
#
# Returns 0 / 10 / 11 / 2 as documented above. Sets KAFKA_PUBLISH_CORRELATION and
# KAFKA_PUBLISH_OUTPUT. Prints ONE line naming the outcome and the correct next move —
# the operator should never have to infer the response from a stack of pod noise.
kafka_publish_checked() {
  local topic="" payload="" correlation="" broker="$KAFKA_PUBLISH_BROKER"
  local timeout_ms="$KAFKA_PUBLISH_TIMEOUT_MS"
  local -a headers=()

  while [ $# -gt 0 ]; do
    case "$1" in
      --topic)       topic="${2:-}";       shift 2 ;;
      --payload)     payload="${2:-}";     shift 2 ;;
      --correlation) correlation="${2:-}"; shift 2 ;;
      --broker)      broker="${2:-}";      shift 2 ;;
      --timeout-ms)  timeout_ms="${2:-}";  shift 2 ;;
      --header)      headers+=("${2:-}");  shift 2 ;;
      *) echo "kafka_publish_checked: unknown option '$1'" >&2; return 2 ;;
    esac
  done

  if [ -z "$topic" ] || [ -z "$payload" ]; then
    echo "kafka_publish_checked: --topic and --payload are required" >&2
    return 2
  fi

  # REFUSE a multi-line payload rather than publishing N broken messages. This is the
  # trap that hides behind the stdin race: a pretty-printed heredoc walks into both at
  # once, and only one of them is visible afterwards.
  if [ "$(printf '%s' "$payload" | wc -l)" -ne 0 ]; then
    echo "kafka_publish_checked: payload is multi-line — kcat -P publishes ONE MESSAGE PER LINE." >&2
    echo "  Build it single-line (jq -c '.') or this publishes fragments, not your message." >&2
    return 2
  fi

  command -v kubectl >/dev/null 2>&1 || { echo "kafka_publish_checked: kubectl not found" >&2; return 2; }

  if [ -z "$correlation" ]; then
    correlation="$(cat /proc/sys/kernel/random/uuid 2>/dev/null || true)"
  fi
  KAFKA_PUBLISH_CORRELATION="$correlation"

  # base64 does two jobs: it survives the sh -c quoting intact, and it cannot reintroduce
  # a newline into the payload on the way through.
  local b64; b64="$(printf '%s' "$payload" | base64 -w0)"

  local hdr_args="" h
  [ -n "$correlation" ] && hdr_args=" -H $(_kafka_shquote "correlation_id=$correlation")"
  for h in ${headers+"${headers[@]}"}; do
    hdr_args="$hdr_args -H $(_kafka_shquote "$h")"
  done

  local remote
  remote="echo '$b64' | base64 -d | kcat -P"
  remote="$remote -b $(_kafka_shquote "$broker")"
  remote="$remote -t $(_kafka_shquote "$topic")"
  remote="$remote -X $(_kafka_shquote "message.timeout.ms=$timeout_ms")"
  remote="$remote$hdr_args && echo $KAFKA_PUBLISH_RECEIPT"

  local pod="kcat-pub-$(date +%s)-${RANDOM}"
  local out status
  set +e
  out="$(kubectl -n "$KAFKA_PUBLISH_NAMESPACE" run "$pod" --rm --restart=Never \
          --image="$KAFKA_PUBLISH_IMAGE" --attach=true --quiet \
          --command -- sh -c "$remote" 2>&1)"
  status=$?
  set -e
  KAFKA_PUBLISH_OUTPUT="$out"

  local code; code="$(_kafka_classify "$status" "$out")"
  case "$code" in
    0)  echo "PUBLISHED  topic=$topic correlation=$correlation" ;;
    10) echo "NOT PUBLISHED  topic=$topic correlation=$correlation" >&2
        echo "  Nothing landed. RETRY NOW — a retry collides with nothing." >&2
        printf '%s\n' "$out" | sed 's/^/  | /' >&2 ;;
    11) echo "RECEIPT INDETERMINATE  topic=$topic correlation=$correlation" >&2
        echo "  kubectl exited 0 but the receipt never arrived, so it is unknown whether this published." >&2
        echo "  DO NOT blind-retry. Run: kafka_verify_landing $correlation" >&2
        printf '%s\n' "$out" | sed 's/^/  | /' >&2 ;;
  esac
  return "$code"
}

# kafka_verify_landing <correlation_id> [timeout_seconds]
#
# For `orchestrate` envelopes. Anything else must verify at its own artefact — a receipt
# says the broker took the bytes, never that the work happened.
#
# Returns 0 / 12 / 13 / 2.
#
# ⚠ THE RETENTION ASYMMETRY THAT MAKES THE ORDER OF THESE CHECKS MATTER. Measured
# 2026-08-23: `orchestration_states` holds about TWO DAYS; `agent_error_log` holds about
# thirty. So "did it land?" is answerable only in a narrow window, while "was it refused?"
# stays answerable for a month. Run this while the answer still exists — and never read a
# zero-row lookup on an OLD correlation as a drop, because past the window that query
# returns 0 for delivered and dropped messages alike.
kafka_verify_landing() {
  local correlation="${1:-}" timeout="${2:-60}"
  [ -n "$correlation" ] || { echo "kafka_verify_landing: correlation id required" >&2; return 2; }
  command -v kubectl >/dev/null 2>&1 || { echo "kafka_verify_landing: kubectl not found" >&2; return 2; }

  local waited=0 row=""
  while [ "$waited" -lt "$timeout" ]; do
    row="$(_kafka_psql "SELECT status || '|' || COALESCE(current_step,'') FROM orchestration_states WHERE correlation_id = '$correlation'::uuid LIMIT 1;")"
    [ -n "$row" ] && break
    sleep 5
    waited=$(( waited + 5 ))
  done

  if [ -n "$row" ]; then
    echo "LANDED  correlation=$correlation  ${row}"
    # bugs_open/326: create_work_item dedups on item_key in ANY status, including
    # terminal ones. A re-submission therefore reports COMPLETED having queued nothing,
    # and `deduped: true` is buried in collected_data where nobody looks. Surface it —
    # this is 326's own "make the silence loud" candidate, done at the trigger side
    # without touching the platform dedup predicate.
    local deduped
    deduped="$(_kafka_psql "SELECT string_agg(k || '=' || COALESCE(v->>'item_key','?'), ', ') FROM orchestration_states o, jsonb_each(o.collected_data) AS e(k,v) WHERE o.correlation_id = '$correlation'::uuid AND jsonb_typeof(v) = 'object' AND v->>'deduped' = 'true';")"
    if [ -n "$deduped" ]; then
      echo "  ⚠ DEDUPED — this run queued NOTHING NEW: $deduped" >&2
      echo "    The work item already exists in some status, terminal included (bugs_open/326)." >&2
      echo "    Re-submitting will report COMPLETED and do nothing again. See that bug before retrying." >&2
    fi
    return 0
  fi

  # No orchestration row. Before calling this a drop, ask the recorder that keeps a
  # month: a REFUSED spawn creates no orchestration row either, and produces exactly the
  # same three absences. A lane blamed kcat on 2026-08-20 for precisely this, while
  # agent_error_log held the delivery record the whole time.
  local refusal
  refusal="$(_kafka_psql "SELECT error_code || ' :: ' || LEFT(error_message,120) FROM agent_error_log WHERE context::text LIKE '%$correlation%' OR orchestration_id LIKE '%$correlation%' OR error_message LIKE '%$correlation%' ORDER BY occurred_at DESC LIMIT 1;")"
  if [ -n "$refusal" ]; then
    echo "CONSUMED AND REFUSED  correlation=$correlation" >&2
    echo "  $refusal" >&2
    echo "  The message WAS delivered and rejected. Do not resend it unchanged — fix the envelope." >&2
    return 12
  fi

  echo "PUBLISHED, NOT LANDED  correlation=$correlation (waited ${timeout}s)" >&2
  echo "  The broker took it; nothing consumed it yet. This is ALSO the signature of ordinary" >&2
  echo "  queue latency, so WAIT and re-poll rather than re-publishing:" >&2
  echo "    SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='$correlation'::uuid;" >&2
  echo "  One exception: a chassis pod that (re)started within ~300s of your publish DOES drop" >&2
  echo "  spawns silently — there, re-publishing is correct." >&2
  return 13
}

# _kafka_psql — one read-only query, trimmed. Never fails the caller.
_kafka_psql() {
  local sql="$1"
  kubectl -n "$KAFKA_VERIFY_NAMESPACE" exec -i "$KAFKA_VERIFY_POD" -- \
    psql -U "$KAFKA_VERIFY_USER" -d "$KAFKA_VERIFY_DB" -tAc "$sql" 2>/dev/null \
    | head -1 | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' || true
}

# ---------------------------------------------------------------------------
# SELF-TEST — `bash scripts/kafka-publish-lib.sh --self-test`
#
# Runs entirely offline. It exercises the DECISION (_kafka_classify) against fixtures and
# the payload guard, because those are the parts that can be wrong without anything
# looking wrong. It deliberately does NOT publish: a self-test that needs a healthy broker
# would be testing the arm that already works.
# ---------------------------------------------------------------------------
_kafka_selftest() {
  local fails=0 got
  _t() { # _t <label> <expected> <actual>
    if [ "$2" = "$3" ]; then echo "  ok    $1"; else echo "  FAIL  $1 (expected '$2', got '$3')"; fails=$(( fails + 1 )); fi
  }
  echo "kafka-publish-lib self-test"

  _t "receipt present, status 0 -> 0"            0  "$(_kafka_classify 0 "PUBLISH_OK")"
  # The case that matters most: a receipt still counts when the publisher also chattered.
  _t "receipt present among noise -> 0"          0  "$(_kafka_classify 0 "% Delivered\nPUBLISH_OK\npod deleted")"
  # A receipt seen with a non-zero status is still a publish: the marker cannot print
  # unless kcat succeeded, and pod teardown can fail afterwards.
  _t "receipt present, status 1 -> 0"            0  "$(_kafka_classify 1 "PUBLISH_OK")"
  _t "no receipt, status 1 -> 10 NOT PUBLISHED"  10 "$(_kafka_classify 1 "% ERROR: broker down")"
  # The old form's silent zero, and the reason 11 exists rather than folding into 10.
  _t "no receipt, status 0 -> 11 INDETERMINATE"  11 "$(_kafka_classify 0 "")"
  # The --command omission trap: kcat prints usage, nothing publishes, status non-zero.
  _t "kcat usage text -> 10"                     10 "$(_kafka_classify 1 "Usage: kcat <options>")"

  got=0; kafka_publish_checked --topic t --payload "$(printf 'a\nb')" >/dev/null 2>&1 || got=$?
  _t "multi-line payload refused with 2"          2  "$got"
  got=0; kafka_publish_checked --payload '{}' >/dev/null 2>&1 || got=$?
  _t "missing --topic refused with 2"             2  "$got"
  got=0; kafka_publish_checked --topic t --payload '{}' --bogus x >/dev/null 2>&1 || got=$?
  _t "unknown option refused with 2"              2  "$got"

  # Quoting: a header value with a space and a quote must survive into ONE sh word,
  # unchanged. Assert the PROPERTY (it round-trips through the shell), not the exact
  # escaping form — there is more than one correct way to quote this, and pinning the
  # representation would fail a rewrite that is still right.
  local tricky="a b'c \$HOME \"q\""
  _t "shquote round-trips through the shell" "$tricky" "$(eval "printf '%s' $(_kafka_shquote "$tricky")")"
  _t "shquote yields exactly one word"       1        "$(eval "set -- $(_kafka_shquote "$tricky"); echo \$#")"

  if [ "$fails" -eq 0 ]; then echo "  ALL PASS"; return 0; fi
  echo "  $fails FAILED"; return 1
}

# Executed directly rather than sourced? Only --self-test is supported.
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  if [ "${1:-}" = "--self-test" ]; then _kafka_selftest; exit $?; fi
  echo "kafka-publish-lib.sh is SOURCED, not executed:" >&2
  echo "  . \"\$REPO_ROOT/scripts/kafka-publish-lib.sh\"" >&2
  echo "(the only direct invocation is: bash $0 --self-test)" >&2
  exit 2
fi
