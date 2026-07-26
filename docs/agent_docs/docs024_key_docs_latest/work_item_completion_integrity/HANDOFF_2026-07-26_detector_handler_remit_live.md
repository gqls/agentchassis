# HANDOFF — bugs_closed/077 is LIVE on v1.0.1171. What is proven, what is not, what is next.

**Written:** 2026-07-26 ~21:10 UTC, by the thread that fixed `077`.
**Read this first if you are picking the work up cold.** It is the state of play,
not a narrative — the narrative is in `NOTES_work_item_completion_integrity.md`
and `README_where_we_are.md`, and the case itself is
`bugs_closed/077_HANDOFF_2026-07-25_detector_predicate_wider_than_its_handler_files_unfixable_items.md`.

---

## One paragraph

Discovery checks used to file work items their handlers provably could not fix,
which the two-strike rule then parked as *"[unresolved after 2 attempts]"* —
blaming a fixer that never had anything to do. Checks now **partition** their
population by the handler's own literal transform: the in-remit part becomes the
normal dispatchable item with an honest count, and the residue becomes a
`capability_gap` (`status='deferred'`, empty `handler_agent`), which is the
platform's pre-existing durable record of *"found work I have no handler for"* and
the intake for the feature builder. Live on **v1.0.1171** since 21:02:56Z.

## Commits (all on `086_experience_loop`)

| commit | what |
|---|---|
| `ce4adfac4` | the code: `remit.go` seam + four checks + the build-enforced handler guard + migration 221 |
| `3c5710ed2` | close 077 → `bugs_closed/`, `016b` §9 entry, §10 index row, `WRONG_CALLS` entry |
| `203e5d7ff` | RUNBOOK entries; the one edge deliberately not closed, recorded in `remit.go` |
| `db35c857a` | a hung council run contributed into `bugs_open/029`; why 077 carries no trailer |

## Proven live — with the evidence, not the claim

**1. The deploy carries it, proven discriminatingly.** The vacuous version of this
check would be "grep the new string, find it". The real one asserts the OLD line
is gone:

```
0   "hardcoded hex colors in inline styles instead of CSS variables"   <- OLD summary, GONE
1   "capability gap, not a handler failure"                            <- new
1   "the colour fixer can replace with CSS variables"                  <- new
1   "PartitionByRemit"                                                 <- new
1   "unresolved after %d attempts"                                     <- positive control
```

**2. Migration 221 applied AFTER the roll, in that order.** `UPDATE 3` — exactly
the three rows predicted (ai-agent-orchestration.com, finetuning.uk,
gaswholesalers.com), each now `wont_fix` and carrying `spec->>'retired_by'`. The
sites with genuinely fixable work (leopardess 1, vonc 1, robot-hands 3) were left
alone, which is the property that mattered. Recorded via `--record-only`.

> **Applied by hand, deliberately.** `run-migrations.sh --apply` applies every
> pending file in order — there were four other threads' migrations pending. The
> runner is right; using it here would have been a cross-thread sweep.

**3. The zero-remit arm, on the real fleet.** Fired `design-discovery-agent` at
finetuning.uk (8 detector matches, 0 in remit). Result — exactly one row, and no
dispatchable item at all:

```
item_type      | status   | handler_agent | gap_kind      | builder              | pop | residue
capability_gap | deferred | (empty)       | handler_remit | color-variable-fixer | 8   | 8
```

**4. The POSITIVE CONTROL — the load-bearing one — passed.**
leopardessconsulting.co.uk (`4851f6fc-71cf-4160-a270-e03d6d3e0732`), population 4.
Both arms fired in one run:

```
item_type                | status   | handler              | found | pop | out_of_remit | gap_kind
hardcoded_section_colors | detected | color-variable-fixer | 1     | 4   | 3            |
capability_gap           | deferred | (EMPTY)              |       | 4   | 3            | handler_remit
```

The arithmetic closes — **1 + 3 = 4**, nothing dropped — and `1 of 4` reproduces
the case file's original table (leopardess: 4 matches, 1 in remit), computed weeks
earlier by a different method (the Go transform over a `row_to_json` dump). Two
independent measurements, same answer.

**Why not robot-hands.com:** its row is `status='detected'`, which is NOT terminal,
so it still holds the `idx_swi_dedup` slot and any insert is silently suppressed by
`ON CONFLICT DO NOTHING`. It would have looked like a failure and proved nothing.
leopardess's rows are `unresolved` — terminal — so its slot was free.

> Firing discovery files items at `status='detected'`, and per `bugs_open/083`
> nothing promotes those today. So this verification spent no dispatch credits —
> the one useful side effect of 083 still being open.

## The verification is COMPLETE. Nothing about 077 is outstanding.

---

## READ THIS BEFORE YOU FIRE ANY kcat DISPATCH — it cost four lost runs

`kubectl -n kafka run -i --rm … kcat -P` **silently produces nothing, most of the
time.** Measured here 2026-07-26: **four of five publishes vanished** — no
orchestration row, no chassis log line for the correlation, `exit 0`, and the
wrapper cheerfully printing a correlation id and "pod deleted" exactly as it does
on success. `kubectl run -i` attaches stdin asynchronously; if the container
reaches `kcat -P -c 1` first it sees EOF, produces nothing and exits clean.

**This affects the shipped triggers**, including
`097_TRIGGER_council_review_v1.sh:121` and
`scripts/initial_messages/290_design_discovery/081_…sh`. It is the likeliest
explanation for this thread's "vanished" council round, which I had wrongly
attributed to the ~300s post-restart window (`WRONG_CALLS.md`, 2026-07-26).

**The fix is one line: put the payload in the container COMMAND, not on stdin, and
make it confirm itself.**

```bash
kubectl -n kafka run "kcat-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '<JSON>' | kcat -P -b <broker> -t <topic> \
  -H correlation_id=<uuid> ... && echo PUBLISH_OK"
```

Every publish after adding `PUBLISH_OK` landed first time. A working copy is at
`fire_discovery2.sh` in this thread's scratchpad; it is worth promoting into
`scripts/` and worth fixing in the triggers themselves.

## Council: NO VERDICT, and no trailer. Do not "fix" this by adding one.

Submission `346500db-89ca-47f3-bc5a-e1c099d6f4f8`, fired twice, never returned a
verdict:

- **Round 1** vanished with no row at all — published 2–4 minutes after a chassis
  restart, inside CLAUDE.md's ~300s drop window. The 097 trigger warns about the
  lane but never looks at the pod.
- **Round 2** landed cleanly, started 19:23:14, and froze **one second later** at
  `review_editquality` with `awaited_steps=[]` and no error. A different
  correlation started four minutes later and COMPLETED TWICE while it sat there,
  so the lane was healthy — filed as a fresh instance on `bugs_open/029`.

A third round was not fired. The trailer is earned by an APPROVED verdict only, so
its absence is correct; the `098` report will list these commits as un-reviewed,
accurately. **If you want review, resubmit under
`RESUBMIT_CORR=346500db-89ca-47f3-bc5a-e1c099d6f4f8`** so the trail accumulates —
and check the chassis pod's `startTime` before publishing.

## What is next, in order

1. **Finish the positive control** (above). Until then the in-remit arm is
   test-proven, not fleet-proven — say so in anything you write.
2. **The capability gaps are now intake, not decoration.** Group them and pick:
   ```sql
   SELECT spec->>'builder_needed' AS builder, spec->>'gap_kind' AS kind,
          count(*) AS items, count(DISTINCT site_id) AS sites
   FROM site_work_items WHERE item_type='capability_gap' AND status='deferred'
   GROUP BY 1,2 ORDER BY 3 DESC;
   ```
   The feature designer's gate needs an `owner_approval` stamp in the spec — a
   human act, deliberately. Full sequence in
   `RUNBOOK_work_item_completion_integrity.md`.
3. **`forced-text-color-fixer` is the cheapest gap and the most dangerous one to
   rush.** Its action `fix_forced_text_colors` is already written and already
   registered; only the `agent_definitions` row is missing. But that action bails
   out entirely below its WCAG contrast floor and only rewrites text-element
   selectors — **seeding it without also partitioning `check_forced_text_colors`
   re-creates 077 under a new item type.**
4. **`site-metadata-fixer` is a real build**, not a seed: no action exists.
5. Unrelated to 077 but now unblocked-ish: `bugs_open/083` (98 rows stuck at
   `detected`). 083 names 077 as a blocker it had to clear first. **That is
   satisfied for these four checks only** — the other ~18 item types in that pile
   are untouched by this work, and enabling its promoter without checking their
   handlers is the thing 083 itself warns against.

## Landmines this thread paid for

- **A superset proves zero; it can never disprove it.** My SQL remit predicate is a
  deliberate over-approximation of the Go transform. I read a `1` from it as
  contradicting a previous thread's `0` and wrote a "CORRECTION" into the approved
  plan and a committed migration header. Both measurements were right. Logged in
  `WRONG_CALLS.md`. Ask: *what would have to be true for both numbers to be right
  at once?*
- **A remit is not always all in Go.** `nav-link-fixer`'s patterns are seeded in
  `agent_definitions` and OVERRIDE the Go defaults — 3 live against 4 in source.
- **The artefact the handler edits is not always the one the detector read.**
  `broken_nav_links` detects on rendered `site_components`; its handler rewrites
  `content_components.html_template`.
- **`detected` is not terminal**, so it holds a dedup slot and suppresses re-files.
  This is why robot-hands.com is useless as a control and why retiring a row to
  `wont_fix` had to wait until after the roll.
- **A guard nobody has watched fail is not known to work.** The unit probes for
  `handler_coverage_test.go` exercise the assertion function, not the source scan.
  It was proven by inserting a bogus agent name into `check_generic_theme.go`
  (red, with the right message) and restoring it (green).
- **Check the chassis pod's `startTime` BEFORE publishing anything that spends
  credits**, not after waiting on it. One `kubectl get pods`.
