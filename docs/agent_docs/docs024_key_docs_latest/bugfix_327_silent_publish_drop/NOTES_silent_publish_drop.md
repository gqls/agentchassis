# NOTES — bug 327, the build trigger can publish nothing and exit 0

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-23 — session start, evidence gathering

**Disambiguation first.** There are TWO bug 327s and `scripts/who-owns.py 327` says so
explicitly. Resolving by slug, as CLAUDE.md requires:

- `bugs_open/327_..._the_build_trigger_can_publish_nothing_and_exit_zero.md` — THIS one.
- `bugs_closed/327_..._a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads.md` — closed 08-23 by the `copy_quality_two_stage` lane, unrelated.

`who-owns.py` conflates them: it reports 24 mentions in `copy_quality_two_stage` and a
"bug 327 CLOSED" commit, **all of which belong to the other 327.** `git log` on the FILE
PATH separates them cleanly:

```
git log --format="%ad %h %s" --date=short -- "bugs_open/327_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md"
# 2026-08-19 db375212c  bug 327: the build trigger can publish nothing and exit 0 ...
```

**One commit, ever — the filing.** The open 327 is UNOWNED and untouched. Confirmed no
dirty files under `scripts/` in the working tree, and no open `site_work_items` naming
this mechanism.

### The bug is still valid — established at the SOURCE, not at the DB

`scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh` is unchanged
since **2026-07-30** (`95639d4f6`), i.e. before the bug was filed on 08-19. Lines 161-178
still carry every element of the defect:

```
161: kubectl -n kafka run -i --rm kcat-submit-$(date +%s) \
164:   kcat -P \
178:   -H responses_topic=system.agent.generic.responses <<JSON
```

`run -i` + payload on stdin, no `--command`, no `PUBLISH_OK` receipt, no landing check.

### MISSTEP 1 — my first re-verification query could not have come out otherwise

I ran this to re-confirm the drop:

```sql
SELECT count(*) FROM orchestration_states
WHERE correlation_id = '8fa2a4a6-2af1-4675-bae9-bbd59b702160';
-- 0
```

Zero. It would have been easy to write "re-verified 2026-08-23, still zero rows". **That
would have been worthless**, and I nearly had it. The control that killed it:

```sql
SELECT created_at::date, count(*) FROM orchestration_states GROUP BY 1 ORDER BY 1;
-- 2026-07-19 |    6
-- 2026-07-20 |    8
-- 2026-07-21 |    6
-- 2026-07-24 |    4
-- 2026-08-22 | 1417
-- 2026-08-23 | 3140
```

**`orchestration_states` holds ~2 days.** There are ZERO rows for the whole of 2026-08-18.
The query returns 0 whether the message landed or not. This is exactly the shape CLAUDE.md
warns about — *"a `[MEASURED]` figure is only evidence if the measurement could have come
out otherwise"* — and the cheap check is one `GROUP BY` on the date. Logged in
`WRONG_CALLS.md`. **The validity of the bug rests on the SOURCE (above), not on this query.**

### The competing explanation, ruled out properly

On 2026-08-20 another lane (`41c06f1d1`) blamed kcat for what turned out to be a
**validation REFUSAL recorded in `agent_error_log` all along**. A refusal produces the same
three absences as a drop, so it must be excluded, not assumed away. It was not checked in
the original bug file.

```sql
SELECT occurred_at, agent_type, error_code, LEFT(error_message,90)
FROM agent_error_log
WHERE context::text ILIKE '%8fa2a4a6%' OR orchestration_id ILIKE '%8fa2a4a6%'
   OR error_message ILIKE '%8fa2a4a6%';
-- (0 rows)
```

An absence, so it needs a positive control — is the instrument even alive on that date?

```sql
SELECT count(*) AS rows_on_18th,
       count(*) FILTER (WHERE error_code='VALIDATION_ERROR_DROPPED') AS refusals_on_18th
FROM agent_error_log WHERE occurred_at::date='2026-08-18';
-- 3761 | 1
```

**3,761 rows including one real `VALIDATION_ERROR_DROPPED` on the very day.** The recorder
was live and the refusal path demonstrably fires. Its silence about `8fa2a4a6` is therefore
evidence. **The refusal explanation is ruled out; the drop stands.**

### Retention asymmetry — and it is an argument, not a footnote

| table | retained from | rows |
|---|---|---|
| `orchestration_states` | ~2 days (08-22 onward) | 4,580 |
| `agent_error_log` | 2026-07-24 (~30 days) | 46,141 |

So **a dropped submission is forensically unrecoverable after ~48h.** You can always ask
later "was it refused?"; you can never ask later "did it land?". That asymmetry is why the
fix has to be a receipt at publish time and not better retrospective diagnosis.

### Framework-wide census `[MEASURED 2026-08-23]`

| | count |
|---|---|
| shell scripts publishing via `kcat -P` | **218** |
| using the racing `kubectl run -i` + stdin form | **200** |
| printing a `PUBLISH_OK` receipt | **25** |
| **actually asserting on that receipt** (grep/if → non-zero exit) | **2** |

The two that assert: `analytics_gtm/scripts/fire_reassemble_site.sh`,
`idea_uk_vm_site/scripts/fire_reassemble_idea_uk.sh`.

**The finding that matters most: the documented remedy is itself only half-applied.** 23 of
the 25 "fixed" scripts print a receipt for a human to read and still exit 0 when it is
absent. A receipt nobody asserts on is not a control — it is a log line.

### The class is live, not historical

2026-08-22 — one day before this session — a landmine was filed for a `097` council
dispatch that exited 0 having published nothing: 90 minutes with no row while other lanes'
runs executed; re-dispatch produced a row in 3 seconds (`7760963cf`). Same class, a
different trigger, current.
