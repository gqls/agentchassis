#!/usr/bin/env bash
# sedo-api.sh — call the Sedo domain-marketplace API (https://api.sedo.com/api/v1/)
# from an ephemeral in-cluster pod, so credentials NEVER enter a session
# transcript, a pod spec, or container logs (owner ruling 2026-08-23: never read
# a key's value into the session; probe from the pod with its own env).
#
# Credentials live in the Kubernetes secret 'sedo-api-credentials'
# (namespace ai-persona-system) with exactly these keys:
#   SEDO_PARTNERID  SEDO_SIGNKEY  SEDO_USERNAME  SEDO_PASSWORD
# The pod receives them via envFrom:secretRef and expands them INSIDE the
# container into a curl config file (mode 0600, tmpfs, dies with the pod).
# Nothing secret ever appears in argv, the overrides JSON, or kubectl output.
#
# Usage:
#   scripts/domains/sedo-api.sh --self-test              # offline checks, no cluster
#   scripts/domains/sedo-api.sh --probe                  # cluster + API reachability, NO creds
#                                                #   (dummy auth; expects SEDOFAULT E7)
#   scripts/domains/sedo-api.sh --check-secret           # secret exists? prints KEY NAMES only
#   scripts/domains/sedo-api.sh <Function> [k=v ...]     # real call, e.g.:
#   scripts/domains/sedo-api.sh DomainList 'results=100'
#   scripts/domains/sedo-api.sh DomainStatus 'domain[]=example.com'
#
# Function reference: https://api.sedo.com/apidocs/v1/Basic/
# Responses are XML; errors come back IN-BAND as <SEDOFAULT> with HTTP 200
# (verified 2026-09-02 — so check the body for SEDOFAULT, not the exit code).
# Pagination: results max 100 per call; page with startfrom=N.
set -euo pipefail

NS=ai-persona-system
SECRET=sedo-api-credentials
IMAGE=curlimages/curl:8.10.1
API_BASE=https://api.sedo.com/api/v1

# Values are embedded into overrides JSON and, in the pod, into a curl config
# file — so the charset is deliberately tight: no quotes, backslashes, dollar,
# backticks, semicolons or ampersands. Domains, prices, counts all fit.
PARAM_RE='^[][A-Za-z0-9_.-]+=[A-Za-z0-9 ._@:/,+-]*$'
FUNC_RE='^[A-Za-z0-9]+$'

# Runs inside the container. $SEDO_* auth vars come from the secret (or dummy
# env for --probe); $SEDO_FUNC / $SEDO_PARAMS are non-secret call details.
# Params are newline-joined so a value may contain ',' or ' ' safely.
INPOD='
set -eu
umask 077
cfg=/tmp/sedo.cfg
: > "$cfg"
printf "data-urlencode = \"partnerid=%s\"\n" "$SEDO_PARTNERID" >> "$cfg"
printf "data-urlencode = \"signkey=%s\"\n"   "$SEDO_SIGNKEY"   >> "$cfg"
printf "data-urlencode = \"username=%s\"\n"  "$SEDO_USERNAME"  >> "$cfg"
printf "data-urlencode = \"password=%s\"\n"  "$SEDO_PASSWORD"  >> "$cfg"
printf "data-urlencode = \"output_method=xml\"\n" >> "$cfg"
printf "%s\n" "$SEDO_PARAMS" | while IFS= read -r p; do
  [ -n "$p" ] && printf "data-urlencode = \"%s\"\n" "$p" >> "$cfg"
done
curl -sS --max-time 60 -K "$cfg" "'"$API_BASE"'/$SEDO_FUNC"
'

usage() { sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'; exit "${1:-0}"; }

build_overrides() { # $1=function $2=newline-joined params $3=auth mode: secret|dummy
  python3 - "$1" "$2" "$3" "$IMAGE" "$SECRET" "$INPOD" <<'PY'
import json, sys
func, params, auth, image, secret, inpod = sys.argv[1:7]
c = {"name": "sedo", "image": image,
     "command": ["sh", "-c", inpod],
     "env": [{"name": "SEDO_FUNC", "value": func},
             {"name": "SEDO_PARAMS", "value": params}]}
if auth == "secret":
    c["envFrom"] = [{"secretRef": {"name": secret}}]
else:  # --probe: dummy values, drawing the documented E7 fault
    c["env"] += [{"name": k, "value": "0"} for k in
                 ("SEDO_PARTNERID", "SEDO_SIGNKEY", "SEDO_USERNAME", "SEDO_PASSWORD")]
print(json.dumps({"apiVersion": "v1",
                  "spec": {"restartPolicy": "Never", "containers": [c]}}))
PY
}

run_pod() { # $1=overrides JSON; prints the API response, exits nonzero on pod failure
  local pod="sedo-api-$$-$RANDOM" phase
  kubectl -n "$NS" run "$pod" --restart=Never --image="$IMAGE" \
    --overrides="$1" >/dev/null
  # shellcheck disable=SC2064
  trap "kubectl -n $NS delete pod $pod --now >/dev/null 2>&1 || true" EXIT
  kubectl -n "$NS" wait --for=jsonpath='{.status.phase}'=Succeeded \
    "pod/$pod" --timeout=120s >/dev/null 2>&1 || true
  kubectl -n "$NS" logs "$pod"
  phase=$(kubectl -n "$NS" get pod "$pod" -o jsonpath='{.status.phase}')
  [ "$phase" = "Succeeded" ] || { echo "sedo-api: pod phase=$phase" >&2; return 1; }
}

self_test() {
  local ovr fails=0
  echo "self-test (offline):"
  [[ "results=100" =~ $PARAM_RE ]] && echo "  PASS param accepts results=100" || { echo "  FAIL"; fails=1; }
  [[ "domain[]=example-site.co.uk" =~ $PARAM_RE ]] && echo "  PASS param accepts domain[]=..." || { echo "  FAIL"; fails=1; }
  [[ 'x="y"' =~ $PARAM_RE ]] && { echo "  FAIL quote not rejected"; fails=1; } || echo "  PASS param rejects quotes"
  [[ 'x=$(id)' =~ $PARAM_RE ]] && { echo "  FAIL subshell not rejected"; fails=1; } || echo "  PASS param rejects \$( )"
  ovr=$(build_overrides DomainList $'results=100\nstartfrom=0' secret)
  echo "$ovr" | python3 -c 'import json,sys; json.load(sys.stdin)' \
    && echo "  PASS overrides JSON well-formed (secret mode)" || { echo "  FAIL overrides JSON"; fails=1; }
  echo "$ovr" | grep -q '"secretRef"' && echo "  PASS secret referenced, not read" || { echo "  FAIL"; fails=1; }
  build_overrides DomainList "" dummy | grep -q '"SEDO_PARTNERID"' \
    && echo "  PASS probe mode uses dummy env" || { echo "  FAIL probe mode"; fails=1; }
  return "$fails"
}

case "${1:---help}" in
  --help|-h) usage 0 ;;
  --self-test) self_test ;;
  --probe)
    out=$(run_pod "$(build_overrides DomainList 'results=1' dummy)")
    echo "$out"
    if echo "$out" | grep -q "E7"; then
      echo "probe: PASS — cluster reached api.sedo.com, API answered (E7 = dummy creds, expected)" >&2
    else
      echo "probe: UNEXPECTED — wanted SEDOFAULT E7 with dummy creds; read the body above" >&2
      exit 1
    fi ;;
  --check-secret)
    kubectl -n "$NS" get secret "$SECRET" >/dev/null 2>&1 \
      || { echo "secret $SECRET does not exist yet — the owner creates it once Sedo issues the SignKey (see RUNBOOK_sedo_domain_management.md §3)" >&2; exit 1; }
    echo "keys present in $SECRET (values never read):"
    kubectl -n "$NS" get secret "$SECRET" \
      -o go-template='{{range $k, $v := .data}}  {{$k}}{{"\n"}}{{end}}'
    for k in SEDO_PARTNERID SEDO_SIGNKEY SEDO_USERNAME SEDO_PASSWORD; do
      kubectl -n "$NS" get secret "$SECRET" -o go-template='{{range $k, $v := .data}}{{$k}}{{"\n"}}{{end}}' \
        | grep -qx "$k" || { echo "MISSING key: $k" >&2; exit 1; }
    done
    echo "all four required keys present" ;;
  --*) echo "sedo-api: unknown flag $1" >&2; usage 1 ;;
  *)
    FUNC=$1; shift
    [[ "$FUNC" =~ $FUNC_RE ]] || { echo "sedo-api: bad function name" >&2; exit 1; }
    PARAMS=""
    for p in "$@"; do
      [[ "$p" =~ $PARAM_RE ]] || { echo "sedo-api: rejected param (charset): $p" >&2; exit 1; }
      PARAMS+="$p"$'\n'
    done
    kubectl -n "$NS" get secret "$SECRET" >/dev/null 2>&1 \
      || { echo "sedo-api: secret $SECRET not found — see RUNBOOK_sedo_domain_management.md for creating it (owner action)" >&2; exit 1; }
    run_pod "$(build_overrides "$FUNC" "$PARAMS" secret)" ;;
esac
