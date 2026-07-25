#!/bin/bash
# ============================================================================
# webdesign_publish_assets.sh — push webdesign.co.uk's static assets to the
# deploy repo (gqls/sites), from which the B2 workflow syncs them.
# ============================================================================
# The chassis publishes PAGES. It does not publish a tool's sibling engine JS,
# the images the articles reference, the compat stylesheet, search.json,
# sitemap.xml, robots.txt or 404.html — none of those are page components. They
# go straight into the deploy repo via the GitHub contents API, which is the
# route the robot-hands lane established (no local checkout of gqls/sites
# exists, and cloning it to add 28 files would be the slower, heavier option).
#
# Idempotent: an existing file's blob sha is fetched first and passed back, so a
# re-run updates rather than colliding. A file whose content already matches is
# skipped without a write.
#
# Reads the asset list from the transform's manifest, so it can never drift from
# what was actually built. Source resolution:
#   - manifest asset with `source`      -> that path under the sites repo
#   - manifest asset marked `generated` -> port/site_assets/<dest>, else
#                                          build/output/<domain>/generated/<dest>
#
# Usage:
#   ./scripts/webdesign_publish_assets.sh [--dry-run]
# ============================================================================
set -euo pipefail

REPO="${REPO:-gqls/sites}"
DOMAIN="${DOMAIN:-webdesign.co.uk}"
SITES_DIR="${SITES_DIR:-$HOME/projects/sites}"
PORT_DIR="${PORT_DIR:-docs/agent_docs/docs024_key_docs_latest/webdesign_couk/port}"
OUT_DIR="${OUT_DIR:-build/output/webdesign_couk}"
MANIFEST="$OUT_DIR/manifest.json"
DRY_RUN=""

[ "${1:-}" = "--dry-run" ] && DRY_RUN=1

[ -f "$MANIFEST" ] || { echo "no manifest at $MANIFEST — run \`webdesignport transform\` first" >&2; exit 1; }

published=0; skipped=0; missing=0

# dest|source|generated, with a literal "-" for an empty field.
#
# The separator is "|" and the placeholder is load-bearing. The first version
# used tabs, and `read` with IFS=$'\t' COLLAPSES runs of tabs — tab is IFS
# whitespace — so an empty middle field silently shifted every later field left.
# Every generated asset then resolved its source to the string "1" and reported
# MISSING, while the image assets (which have a source) worked perfectly. A
# partial failure that looks like a path problem and is really a parsing one.
python3 - "$MANIFEST" <<'PY' | while IFS='|' read -r DEST SRC GEN; do
import json, sys
m = json.load(open(sys.argv[1]))
for a in m.get("assets", []):
    print("|".join([a["dest"], a.get("source") or "-", "1" if a.get("generated") else "-"]))
PY

  # Resolve the local file.
  LOCAL=""
  if [ "$SRC" != "-" ]; then
    LOCAL="$SITES_DIR/$SRC"
  elif [ "$GEN" = "1" ]; then
    if [ -f "$PORT_DIR/site_assets/$DEST" ]; then
      LOCAL="$PORT_DIR/site_assets/$DEST"
    elif [ -f "$OUT_DIR/generated/$DEST" ]; then
      LOCAL="$OUT_DIR/generated/$DEST"
    fi
  fi

  if [ -z "$LOCAL" ] || [ ! -f "$LOCAL" ]; then
    echo "  MISSING  $DEST (no local file)" >&2
    missing=$((missing + 1))
    continue
  fi

  REMOTE="$DOMAIN/$DEST"

  # Existing blob sha, if any. `gh api` exits non-zero on 404, which is the
  # normal first-publish case — hence the `|| true`.
  SHA="$(gh api "repos/$REPO/contents/$REMOTE" --jq '.sha' 2>/dev/null || true)"

  # Skip when the content already matches: git blob sha is
  # sha1("blob <bytes>\0" + content).
  if [ -n "$SHA" ]; then
    LOCAL_SHA="$( { printf 'blob %s\0' "$(wc -c < "$LOCAL")"; cat "$LOCAL"; } | sha1sum | cut -d' ' -f1)"
    if [ "$LOCAL_SHA" = "$SHA" ]; then
      skipped=$((skipped + 1))
      continue
    fi
  fi

  if [ -n "$DRY_RUN" ]; then
    echo "  WOULD PUT $REMOTE ($(wc -c < "$LOCAL") bytes)"
    published=$((published + 1))
    continue
  fi

  # The payload goes in on STDIN, not as an argv. Passing base64 content with
  # `-f content=...` puts the whole file on the command line, which blows
  # ARG_MAX ("Argument list too long") for anything around a megabyte — and it
  # fails per-file, so small assets publish and large ones quietly do not.
  PAYLOAD="$(python3 -c '
import base64, json, sys
body = {"message": "webdesign.co.uk: publish " + sys.argv[1],
        "content": base64.b64encode(open(sys.argv[2], "rb").read()).decode()}
if len(sys.argv) > 3 and sys.argv[3]:
    body["sha"] = sys.argv[3]
print(json.dumps(body))' "$DEST" "$LOCAL" "$SHA")"

  if printf '%s' "$PAYLOAD" | gh api -X PUT "repos/$REPO/contents/$REMOTE" --input - >/dev/null 2>&1; then
    echo "  published $REMOTE"
    published=$((published + 1))
  else
    echo "  FAILED    $REMOTE" >&2
  fi
done

echo
echo "done (counts are per-subshell; see the lines above for the record)"
