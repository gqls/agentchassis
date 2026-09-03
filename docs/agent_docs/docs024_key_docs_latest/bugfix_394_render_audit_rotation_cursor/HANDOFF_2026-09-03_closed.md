# HANDOFF — 2026-09-03 — `bugs_closed/394`: CLOSED, and all three questions RULED.

**Supersedes every earlier handoff in this directory** (08-26, 09-02, 09-02b — all stamped).

**Bug:** `bugs_closed/394_HANDOFF_2026-08-25_...md` (moved 2026-09-03)
**Councils:** `f67593f5` APPROVED (cursor, r3) · `f49da30d` APPROVED (CronJob, 2 advisories, both answered)
**Register:** VIZ-019, status **LIVE AND ACCEPTED**

---

## 1. Nothing is left. The lane is DONE.

Closed on the estate's own bar — fixed AND live — with both halves proven at the artefact rather
than inferred. The three product/judgement questions I did not answer alone were **put to the owner
and ruled on 2026-09-03** (§4). Nothing on this lane is outstanding.

The one thing that survives it is `bugs_open/452`, spun out and unowned — a dead config key, LOW,
and ruling A narrows which of its fix candidates is live.

## 2. What was wrong, and what is true now

**Was:** `request_render_audit` sorted a site's pages and took the first `max_pages`, every run, for
ever. On webdesign.co.uk that was the same 60 pages — **91 never audited**, including **all 45**
`tool-*-guide` pages, which no cap below 98 could reach. `bugs_closed/242` had made the truncation
loud; nothing read the signal.

**Is:** a keyset coverage cursor (opt-in `rotate_coverage`, migration **660**) whose window is a
**UNION** — the pages carrying an open `contrast_failure` ride EVERY run, so the 3-day
repair-grading latency that migration 469 was an owner instruction to buy is preserved — plus a
scheduled reader that alarms when the rotation regresses.

| claim | evidence |
|---|---|
| **acceptance PASSES** | `[MEASURED 2026-09-02]` union of `audited_paths` over 3 scheduled runs = **151 distinct pages**; live = **151**; **missed = 0**, graded against the site |
| cycle completes | run 3 final window 37 pages → `tool-llm-cost-calculator` (`nav_order` 201), `cursor_cleared=true`, cursor row then deleted |
| a second site too | `loanandmortgagecalculator.co.uk` (61 pages) cycled unaided; started a fresh cycle 09-03 04:17Z |
| identity fix live | discriminating test — the two hypotheses predicted OPPOSITE windows; observed `window_first=index` + a new `render-audit-agent` row |
| reader **runs** | `[MEASURED 2026-09-03]` CronJob fired **on its own schedule** 07:50:00Z; durable row 07:50:14Z: *22 rows / 4 sites / 0 findings / 1 dormant named* |
| registry honest | `RENDER_AUDIT_TRUNCATED` → `consumed`, reader + reader_sink, verified by the checker opening the file |

## 3. Why this took the shape it did — the two decisions already made, for context

- **"Raise the cap" was rejected on a MEASUREMENT.** The tail was a *class*, not a count: all 45
  guide pages sat at `nav_order` 200, past any cap below 98, on a site that grew 131 → 151 pages in
  ten days. A constant cannot chase that. And `max_pages` is per-DISPATCH, so there was no single
  "the cap" to raise.
- **The window is a union because a plain cursor would have broken an owner ruling.** It would have
  taken per-page re-measurement to ~9 days — worse than the 7 that migration 469 was issued to
  remove. The protected population is not the site but the pages awaiting a grade: 3 of 151. Cost:
  3 of 60 slots.

## 4. THE THREE DECISIONS — ALL RULED BY THE OWNER, 2026-09-03

Recorded here AND in `bugs_closed/394`'s close-out, because a ruling that lives only in a chat log
is a ruling the next session re-opens. The durable copy is the bug file; this is the pointer.

### A — `design-critique-agent` KEEPS ITS PREFIX. Ruled: *"leave it"*.

It stays a **taste instrument, not a coverage instrument**. It is manual-only, so it accrues no
debt between runs, and comparability across critiques is worth more than eventual coverage of an
8-page sample.

**Reversible by design:** `rotate_coverage: true` on its `audit` step is a one-line migration.
**The operative control is the acks entry** in
`docs/agent_docs/docs024_key_docs_latest/architecture_review/render_truncation_acks.json`, now
stamped with this ruling — so the watchdog's arm (c) treats its prefix rows as a reviewed exception
rather than an oversight, and any THIRD caller still pages.

⚠ **The ruling does NOT cover the cadence case, and the ack says so.** If this agent ever gains a
schedule, re-open it: a scheduled sampler that never revisits IS a coverage debt, and that is not
what was ruled on.

### B — the detection-latency trade is ACCEPTED. Ruled: *"accept it"*.

Stated precisely, because the two clocks are easy to conflate:

- **Confirmation latency stays at 3 days** — a repair ships, the audit re-measures, the item is
  withdrawn. That is what migration 469's owner instruction bought and what the union window exists
  to protect. **Unchanged.**
- **Detection latency moved** on webdesign.co.uk — a NEW defect on the first 60 pages now waits up
  to one cycle (~9 days) instead of 3, and the other 91 pages went from **never** to one cycle.

Accepted on that basis. `max_pages` is now a pure latency dial with no coverage cliff behind it, so
buying head-page latency back is a config change rather than a redesign — the number to watch is
site growth, and the watchdog reports it.

### C — the 14-day dormancy window is ACCEPTED. Ruled: *"accept it"*.

Accepted **as a judgement rather than a measurement**, which is how it was put: ~4 missed
opportunities at the 3-day per-site cadence. One constant (`dormantAfterDays`,
`cmd/config-key-audit/rendertruncation.go`), one line to change.

Its stated blind spot is unchanged and is not a defect: a regression on a site that ALSO stops
truncating goes unseen — bounded, because a site stops writing these rows when it fits inside its
cap, and a site inside its cap has no coverage debt to detect. Dormant groups are counted and NAMED
in every run, so the blind spot stays legible rather than implicit.

## 5. Traps this lane paid for — read before touching these

- **`snapshot_agent` has TWO overloads writing to DIFFERENT TABLES**; a bare literal is ambiguous
  and aborts the migration; the two-arg form writes `agent_definitions_backup` and returns the
  SOURCE id, so verifying in `agent_definitions` reads 0 and looks like a no-op.
- **NEVER parse a `contrast_failure` `item_key`** — a selector may contain `#` and so may a page URL
  (`idea.uk` has `/tools.html` AND `/tools.html#audience-check` active). Match forward with
  `workItemKey(...)` + `HasPrefix`, longest path first. The council caught this as a LIVE defect.
- **BusyBox `grep` over `/proc/1/exe` reports FALSE ABSENCES while both controls pass.** Use
  `tr '\0' '\n' | grep -Fc`, both controls through the SAME pipeline.
- **`sql_for_agents` numbers are a shared sequence with no reservation** — mine was renumbered twice
  (646 → 649 → **660**). Take `max+1` at the moment you APPLY.
- **A test can only discriminate what its fixture varies** — one shared fixture set two identities
  to one literal and blinded twelve tests to a real defect.
- **Prove a mutation is not vacuous before trusting it**: build the mutated tree first (a build
  failure reads just like the red you wanted), and predict which assertion moves.

## 6. Do NOT chase this in the shared tree

`go test ./cmd/config-key-audit/` may fail on `TestBudgetCronCountsLiteralMatchesTheRegistry`.
Another session has **uncommitted** work adding a 5th optional key (`max_image_dimension`) to
`execute_vision_prompt`, plus untracked `vision_image_downscale.go`/`_test.go`. Their commit will
hit that guard — that is the guard working. Confirm with
`git status --short -- platform/orchestration/actions/ | grep -i vision`.

## 7. Spun out, still open — and ruling A changed its shape

`bugs_open/452` — `page_names` is declared in `request_render_audit`'s spec and read by nothing.
LOW; no live carrier sets it, so nothing is broken today. It costs a false affordance and a slot of
the N=10 optional-key budget (reads as 7, really 6).

**Ruling A narrows it.** The "curate the design critique's sample" motive — which was the only live
reason to IMPLEMENT the key — is off the table. So 452's candidate **1 (delete the declaration)** is
now the live one and candidate 2 (implement it) is parked. Whoever takes 452 should read A first
rather than re-deriving the choice.

## 8. Commit trail

`95a04168c` cursor · `72b16391b` R1 fix · `41b03241d` reader + registry flip · `ea08c831d` R3
advisories + `Council-Reviewed` · `c71b46be0` migration 660 applied · `faf4872ce` identity fix ·
`a3610ea23` artefact proof · `667973351` acceptance passes · CronJob + dormancy · `3749132e0`
registry `why` · `ad8eb5f4f` pod proof · `a20de1431` **close(394)** · `0d53d2e18` file 452 +
pointer correction · `953733d1c` register LIVE.
