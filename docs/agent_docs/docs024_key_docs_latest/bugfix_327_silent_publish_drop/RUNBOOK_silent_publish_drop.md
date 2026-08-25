# RUNBOOK — bug 327, silent publish drop

Every command that was hard to get right, with its gotcha attached.

---

## Disambiguate the two bug 327s

`who-owns.py` CONFLATES them. Always resolve by file path, never by number:

```bash
python3 scripts/who-owns.py 327          # prints the AMBIGUOUS warning — read it
git log --format="%ad %h %s" --date=short -- "bugs_open/327b_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md"
```

⚠ Every "bug 327 CLOSED / round 2 / fix is LIVE" commit in `git log --grep 327` belongs to
the **other** 327 (`a_partial_spec_write_silently_shrinks_the_brief`). The open one has
exactly one commit: its filing.

## Re-verify the bug at the SOURCE (the only check that keeps working)

```bash
grep -n "kubectl -n kafka run\|kcat -P\|<<JSON\|--command\|PUBLISH_OK" \
  scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh
```

Defective if you see `run -i` and `<<JSON` and NO `--command` / `PUBLISH_OK`.

## ⚠ Do NOT re-verify at `orchestration_states` — it retains ~2 days

```sql
SELECT created_at::date, count(*) FROM orchestration_states GROUP BY 1 ORDER BY 1;
```

Run this BEFORE reading anything into a zero-row correlation lookup. As of 2026-08-23 the
table held rows only for 08-22/08-23 plus four July dates. A `count(*)=0` for any
correlation older than ~48h is the retention window, not a drop, **and it looks identical.**

## Rule out the validation-refusal explanation (it is NOT optional)

A refusal produces the same three absences as a drop. It IS durably recorded:

```sql
SELECT occurred_at, agent_type, error_code, LEFT(error_message,90)
FROM agent_error_log
WHERE context::text ILIKE '%<corr-prefix>%'
   OR orchestration_id ILIKE '%<corr-prefix>%'
   OR error_message ILIKE '%<corr-prefix>%';
```

**An empty result needs a positive control** — prove the recorder was alive that day:

```sql
SELECT count(*) AS rows_that_day,
       count(*) FILTER (WHERE error_code='VALIDATION_ERROR_DROPPED') AS refusals
FROM agent_error_log WHERE occurred_at::date='<the date>';
```

`agent_error_log` retains ~30 days (from 2026-07-24), so this check still works long after
`orchestration_states` has forgotten.

## The framework-wide census (re-run before quoting the numbers — they go stale by ADDITION)

```bash
# total publishers
grep -rl "kcat -P" --include="*.sh" . | wc -l
# the racing form
grep -rl "kcat -P" --include="*.sh" . | while read f; do grep -q "run -i" "$f" && echo "$f"; done | wc -l
# print a receipt
grep -rl "PUBLISH_OK" --include="*.sh" . | wc -l
# ASSERT on the receipt — the number that actually matters
grep -rl "PUBLISH_OK" --include="*.sh" . | while read f; do
  grep -qE 'grep .*PUBLISH_OK|if .*PUBLISH_OK|\[\[ .*PUBLISH_OK' "$f" && echo "$f"; done
```

`[MEASURED 2026-08-23]` 218 / 201 / 25 / **2** — but see the corrections below; the racing figure counts comments.

⚠ **TWO corrections make the naive census wrong in the same direction. Use the command below,
not a bare grep.**

1. **A file containing the pattern is not a publisher that can RUN** — 23 are scrapbooks with a
   `.sh` extension (pasted SQL, no shebang, or a syntax error), e.g.
   `020_build_pipeline/076_trigger_build_pipeline.sh` and `077_submit_domain.sh`.
2. **A COMMENT describing the trap matches a grep for the trap.** `[MEASURED 2026-08-24]` **18
   files** carry `run -i` + `kcat -P` only inside comments — warnings *about* this very hazard,
   including the ones this lane wrote into every file it migrated. **A migrated file keeps
   matching a naive census for ever**, so the number stops moving exactly when the work starts
   working. (`pattern-check.py`'s detector was never fooled: it strips comments. Only this
   one-liner was.)

**Strip comments AND require it to parse:**

```bash
grep -rl "kcat -P" --include="*.sh" . \
  | while read f; do sed 's/#.*//' "$f" | grep -q "run -i" && bash -n "$f" 2>/dev/null && echo "$f"; done \
  | wc -l
```

`[MEASURED 2026-08-24]` **219** publishers · **183** racing in code (201 if you count comments)
· **160** racing *and* parsing — **quote 160 as the exposure** · **4** callers on the checked
library.

Per CLAUDE.md's counting rule, quote these with the date; check for additions since with:

```bash
git log --since=2026-08-23 --diff-filter=A --name-only -- '*.sh' | grep '\.sh$' | sort -u
```

## The known-good publish form (and why each part is load-bearing)

```bash
kubectl -n kafka run "kcat-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '<base64>' | base64 -d | kcat -P -b <broker> -t <topic> \
  -H correlation_id=<uuid> ... && echo PUBLISH_OK"
```

- **`--command` is load-bearing**: the image ENTRYPOINT is kcat, so without it your `sh -c`
  arrives as kcat *arguments* — nothing publishes, the same silent zero by another route.
- **payload in the COMMAND, not stdin**: `run -i` attaches stdin asynchronously; lose the
  race and kcat sees EOF, publishes nothing, exits 0.
- **base64 of a SINGLE line**: `kcat -P` publishes one message per LINE, so a
  pretty-printed heredoc is a second, independent trap.
- **`&& echo PUBLISH_OK`**: the receipt. ⚠ **printing it is not enough — the caller must
  assert on it and exit non-zero**, which 23 of the 25 scripts that print it do not do.

## Council admission — test it for free before spending credits

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

⚠ Scope is `^(platform|internal|pkg)/|^cmd/config-key-audit/|^scripts/pattern-check\.py$` plus
appliable migrations (`scripts/council-scope.sh`).

> **UPDATED 2026-08-24 (owner ruling).** `scripts/pattern-check.py` is now IN scope, so this
> lane's submission — refused exit 2 on 08-23 — is admitted exit 0 today. **The rest of
> `scripts/` is still refused**, including `kafka-publish-lib.sh`, so the publisher itself
> remains unreviewable.
>
> ⚠ **Widening scope means editing TWO files.** `098_REPORT` carries `SCOPE_PATHS`, a hand-kept
> pathspec array, as a pre-filter — `git log` takes pathspecs, not regexes. A path in the regex
> but not the array is **invisible** to the coverage report (absent, not unreviewed). That cost
> 22 hidden commits on 2026-08-23.
