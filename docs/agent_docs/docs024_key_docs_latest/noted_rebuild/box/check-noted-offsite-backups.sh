#!/usr/bin/env bash
# check-noted-offsite-backups.sh — assert the off-box backups are actually there.
#
# RUNS OFF THE BOX, ON PURPOSE. A monitor living on the machine it monitors
# reports nothing when that machine is the thing that failed — which is the exact
# scenario the off-box copy exists for.
#
# Needs an ADMIN b2 key (the backup key cannot list or read, by design).
#
# THE REASON THIS LISTS WITH --versions, AND IT IS NOT A STYLE CHOICE:
#   The backup key holds `writeFiles`, which INCLUDES the right to HIDE an
#   object. A hidden object vanishes from an ordinary `b2 ls` completely — so
#   "someone hid every backup" and "the backup job never ran" produce the SAME
#   empty listing. `--versions` shows the hide markers, which is the only way to
#   tell destruction from absence. Verified by probe, 2026-08-10.
set -uo pipefail

BUCKET=${NOTED_BACKUP_BUCKET:-personae-noted-backups}
PREFIX=noted/pg/
EXPECT_DAYS=${EXPECT_DAYS:-7}
fail=0

say()  { printf '%s\n' "$*"; }
bad()  { printf 'FAIL: %s\n' "$*"; fail=1; }

listing=$(b2 ls --recursive --long --versions "b2://${BUCKET}/${PREFIX}" 2>&1) || {
    bad "cannot list b2://${BUCKET}/${PREFIX} — $listing"; exit 1; }

# A hide marker is a 0-byte entry with the same name as a real object.
hidden=$(printf '%s\n' "$listing" | awk '$5 == 0 {print $NF}' | sort -u)
if [ -n "$hidden" ]; then
    bad "HIDE MARKERS PRESENT — backups have been hidden, not lost. Recover with:"
    printf '%s\n' "$hidden" | sed 's|^|      b2 file unhide b2://'"$BUCKET"'/|'
fi

# One dump per day for the last EXPECT_DAYS. Names carry a UTC stamp, so this
# does not depend on B2's own timestamps (which a re-upload would move).
missing=0
for i in $(seq 0 $((EXPECT_DAYS - 1))); do
    day=$(date -u -d "$i days ago" +%Y%m%d)
    if ! printf '%s\n' "$listing" | grep -q "noted-${day}T"; then
        say "  no dump for ${day}"
        missing=$((missing + 1))
    fi
done
[ "$missing" -gt 1 ] && bad "$missing of the last $EXPECT_DAYS days have no dump"

# Size floor: an age-encrypted pg_dump of even an empty database is >1 KB.
# A run of suspiciously tiny objects means the dump broke while still exiting 0.
tiny=$(printf '%s\n' "$listing" | awk '$5 > 0 && $5 < 1000' | wc -l)
[ "$tiny" -gt 0 ] && bad "$tiny object(s) under 1000 bytes — dump may be truncated"

# Everything must be encrypted. An object without .age means a code path
# uploaded plaintext notes, which is a breach and not a warning.
plain=$(printf '%s\n' "$listing" | awk '{print $NF}' | grep -v '\.age$' | grep -c . || true)
[ "$plain" -gt 0 ] && bad "$plain object(s) NOT .age-encrypted"

total=$(printf '%s\n' "$listing" | grep -c 'noted-' || true)
say "checked b2://${BUCKET}/${PREFIX}: ${total} object version(s), ${missing} day(s) missing of ${EXPECT_DAYS}"
[ "$fail" -eq 0 ] && say "OK" || say "PROBLEMS FOUND"
exit $fail
