#!/usr/bin/env bash
# Publish a provocations feed to the vonc.com deploy repo.
#
# THIS SCRIPT IS ALSO THE ROLLBACK. There is no separate revert path: rolling
# back is running it with a backup file instead of a freshly built one. That is
# deliberate — PLAN §10.4 requires a rollback that exists and is exercised
# before anything publishes automatically, and a revert path that is only ever
# run in an emergency is a revert path nobody knows works. Publishing forward
# tests it.
#
#   ./publish_feed.sh --dry-run out.json
#   ./publish_feed.sh out.json "vonc.com: rotate provocation"
#   ./publish_feed.sh backups/provocations_2026-07-31_pre.json "revert"   # rollback
#
set -euo pipefail

REPO="gqls/sites"
PATH_IN_REPO="vonc.com/data/provocations.json"
SERVED_URL="https://vonc.com/data/provocations.json"

DRY=0
ROLLBACK=0
while [[ "${1:-}" == --* ]]; do
  case "$1" in
    --dry-run) DRY=1 ;;
    --rollback) ROLLBACK=1 ;;
    *) echo "unknown flag: $1" >&2; exit 1 ;;
  esac
  shift
done
FILE="${1:?usage: publish_feed.sh [--dry-run] [--rollback] <feed.json> [commit message]}"
MSG="${2:-vonc.com: update provocations.json (provocation_pipeline)}"

[[ -f "$FILE" ]] || { echo "no such file: $FILE" >&2; exit 1; }

# 1. Preflight, in TWO TIERS. The distinction is not cosmetic:
#
#    SAFETY checks run on every path including rollback. They are the ones
#    whose failure is an OUTAGE — round.go returns 503 when `today` is absent
#    (deliberately, "fail loud"), and the page loader silently leaves the shell
#    empty if the fields it reads are missing.
#
#    QUALITY checks (slug/date present) are a FORWARD standard introduced by
#    Phase 0, and are skipped under --rollback.
#
#    Why the split exists: the first version enforced both on every path, and
#    the very first thing it did was REFUSE to roll back to the currently-live
#    file — because that file has no `today.slug`, which is precisely the defect
#    Phase 0 fixes. A revert must be able to restore a known-good earlier state
#    even when that state predates the standard we have since adopted.
#    An escape hatch gated on the improvement it exists to undo is not an
#    escape hatch. Found by dry-running the rollback, not by reasoning.
python3 - "$FILE" "$ROLLBACK" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
rollback = sys.argv[2] == "1"

missing = [k for k in ("generated_at", "today", "arena", "archive") if k not in d]
if missing:
    sys.exit(f"REFUSING: feed is missing top-level keys: {missing}")
t = d["today"]
for k in ("headline", "body", "eyebrow", "primary_cta", "secondary_cta", "stats"):
    if not t.get(k):
        sys.exit(f"REFUSING (safety): today.{k} is missing or empty — the live loader reads it")

slugs = [e["slug"] for e in d["archive"]["entries"]]
if len(slugs) != len(set(slugs)):
    sys.exit("REFUSING (safety): duplicate slugs in the archive")
if t.get("slug") and t["slug"] in slugs:
    sys.exit(f"REFUSING (safety): today ({t['slug']}) is also in the archive")

if not rollback:
    for k in ("slug", "date"):
        if not t.get(k):
            sys.exit(f"REFUSING (quality): today.{k} is missing — it could never be "
                     f"archived. Use --rollback to restore a pre-Phase-0 file anyway.")

label = t.get("slug") or "<no slug — pre-Phase-0 file>"
print(f"  preflight ok{' (ROLLBACK: quality checks skipped)' if rollback else ''}: "
      f"today={label} ({t.get('date', 'no date')}), archive={len(slugs)}, "
      f"generated_at={d['generated_at']}")
PY

SHA=$(gh api "repos/$REPO/contents/$PATH_IN_REPO" --jq '.sha')
echo "  target : $REPO:$PATH_IN_REPO"
echo "  cur sha: $SHA"
echo "  new    : $(wc -c <"$FILE") bytes"

if [[ $DRY -eq 1 ]]; then echo "  DRY RUN — nothing written"; exit 0; fi

# 2. PUT with the payload on STDIN. argv blows ARG_MAX on a file this size, and
#    the failure mode is a truncated commit rather than an error.
python3 - "$FILE" "$MSG" "$SHA" <<'PY' | gh api --method PUT "repos/$REPO/contents/$PATH_IN_REPO" --input - >/dev/null
import base64, json, sys
content = base64.b64encode(open(sys.argv[1], "rb").read()).decode()
json.dump({"message": sys.argv[2], "content": content, "sha": sys.argv[3]}, sys.stdout)
PY

echo "  committed."

# 3. Verify at the ARTEFACT. A green PUT says GitHub accepted a commit; it says
#    nothing about what the CDN is serving. Poll until the served bytes match
#    what we pushed, and say plainly if they never do.
WANT=$(md5sum <"$FILE" | cut -d' ' -f1)
for i in $(seq 1 40); do
  GOT=$(curl -s -m 20 "$SERVED_URL" | md5sum | cut -d' ' -f1)
  if [[ "$GOT" == "$WANT" ]]; then
    echo "  SERVED and matching after ~$((i * 15))s (md5 $GOT)"
    exit 0
  fi
  sleep 15
done
echo "  *** pushed, but the served file still does not match after 10 minutes ***" >&2
echo "  *** expected md5 $WANT — check the B2 sync workflow before re-pushing ***" >&2
exit 2
