#!/usr/bin/env bash
# noted-offsite-backup.sh — encrypt the newest local dumps and push them to B2.
#
# Installed on webdesign.vs.mythic-beasts.com 2026-08-10. Runs after
# noted-pg-backup.sh, which is what produces the local dump this uploads.
#
# WHAT IT COVERS
#   noted/pg/          the noted Postgres database
#   webdesign-chat/    webdesign-chat's own state + transcripts (owner asked for
#                      this on 2026-08-10 — nothing else backs them up, and they
#                      share the box's "lose the box, lose the data" exposure)
#
# THE TWO PROPERTIES THAT MAKE THIS SAFE, AND WHY THEY ARE NOT OPTIONAL
#
#   1. EVERYTHING IS ENCRYPTED BEFORE IT LEAVES THE BOX, to an age public key
#      whose private half is NOT on this machine. Rooting this box gets you the
#      ability to write new backups and NO ability to read a single old one.
#      This is why the recipient is a `-R` file and there is no decrypt path
#      anywhere in this script: there is nothing here to steal.
#
#   2. THE B2 KEY CAN WRITE AND NOTHING ELSE — no readFiles, no listFiles, no
#      deleteFiles, scoped to one bucket. Verified by probe on 2026-08-10.
#
#      BUT: `writeFiles` INCLUDES HIDE. A stolen key cannot delete or read a
#      backup, but it CAN hide one, and a hidden object disappears from ordinary
#      listings completely — `b2 ls` returns nothing at all, so a hidden backup
#      and a missing backup look identical. Recovery is `b2 file unhide` with an
#      admin key and the object is intact. Any monitor for this bucket MUST list
#      with `--versions`, or it cannot tell "destroyed" from "never uploaded".
#      Object Lock (governance, 30 days) is what stops a hide from ever becoming
#      a real deletion inside the window.
set -euo pipefail

ENV_FILE=/etc/noted/b2.env
LOCAL_DUMPS=/var/backups/noted
CHAT_DATA=/var/lib/webdesign-chat
WORK=$(mktemp -d); trap 'rm -rf "$WORK"' EXIT

[ -r "$ENV_FILE" ] || { echo "offsite-backup: $ENV_FILE missing" >&2; exit 1; }
set -a; . "$ENV_FILE"; set +a
: "${B2_APPLICATION_KEY_ID:?}" "${B2_APPLICATION_KEY:?}" "${NOTED_BACKUP_BUCKET:?}" "${NOTED_AGE_RECIPIENT:?}"

# Keep this run's b2 auth out of root's shared account file, so a concurrent
# admin b2 invocation cannot be silently downgraded to the restricted key.
export B2_ACCOUNT_INFO="$WORK/b2_account"
b2 account authorize "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY" >/dev/null

uploaded=0

upload_encrypted() {
    local src="$1" dest="$2"
    local enc="$WORK/$(basename "$dest")"
    age -e -r "$NOTED_AGE_RECIPIENT" -o "$enc" "$src"

    # An age file starts "age-encryption.org/v1". If this ever uploads plaintext
    # because age silently no-opped, that is a data breach, not a bug — so the
    # check is here rather than in a test, and it refuses rather than warns.
    head -c 21 "$enc" | grep -q '^age-encryption.org' \
        || { echo "offsite-backup: REFUSING to upload $dest — not age-encrypted" >&2; exit 1; }

    b2 file upload --no-progress --quiet "$NOTED_BACKUP_BUCKET" "$enc" "$dest" >/dev/null
    echo "offsite-backup: uploaded $dest ($(stat -c %s "$enc") bytes encrypted)"
    uploaded=$((uploaded + 1))
}

# --- the noted database: newest local dump only (older ones are already up) ---
newest=$(find "$LOCAL_DUMPS" -name 'noted-*.dump' -type f -printf '%T@ %p\n' 2>/dev/null \
         | sort -rn | head -1 | cut -d' ' -f2-)
if [ -n "$newest" ]; then
    upload_encrypted "$newest" "noted/pg/$(basename "$newest").age"
else
    echo "offsite-backup: no local dump found in $LOCAL_DUMPS" >&2
    exit 1
fi

# --- webdesign-chat: its own state, tarred first (many small files) ---
if [ -d "$CHAT_DATA" ]; then
    tar -czf "$WORK/chat.tar.gz" -C "$CHAT_DATA" . 2>/dev/null || true
    if [ -s "$WORK/chat.tar.gz" ]; then
        upload_encrypted "$WORK/chat.tar.gz" \
            "webdesign-chat/webdesign-chat-$(date -u +%Y%m%dT%H%M%SZ).tar.gz.age"
    fi
else
    # Not fatal: this script's primary job is the notes database. Say so loudly
    # enough to be noticed, quietly enough not to fail the run.
    echo "offsite-backup: NOTE — $CHAT_DATA not found, webdesign-chat not backed up" >&2
fi

echo "offsite-backup: done, $uploaded object(s) uploaded"
