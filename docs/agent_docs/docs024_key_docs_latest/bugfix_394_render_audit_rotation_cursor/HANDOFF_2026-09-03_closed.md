# HANDOFF — 2026-09-03 — `bugs_closed/394`: CLOSED. Three open QUESTIONS for the owner.

**Supersedes every earlier handoff in this directory** (08-26, 09-02, 09-02b — all stamped).

**Bug:** `bugs_closed/394_HANDOFF_2026-08-25_...md` (moved 2026-09-03)
**Councils:** `f67593f5` APPROVED (cursor, r3) · `f49da30d` APPROVED (CronJob, 2 advisories, both answered)
**Register:** VIZ-019, status **LIVE AND ACCEPTED**

---

## 1. Nothing is left to DO. Three things are left to DECIDE.

The lane is closed on the estate's own bar — fixed AND live — and both halves are proven at the
artefact, not inferred. What remains are three **product/judgement questions** I deliberately did
not answer alone. They are laid out in §4. None of them blocks anything; all three are cheap to act
on later, and each is cheaper to answer now while the evidence is fresh.

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

## 4. ⚠ THE THREE DECISIONS

### Decision A — should the design critic rotate, or keep looking at the same 8 pages?

**What the thing is.** `design-critique-agent` opens 8 pages of a site in a real browser and asks a
vision model for a design opinion. It picks the first 8 in navigation order — the same 8 every run.

**The rule that governs it.** For a *coverage* sweep, always taking the same pages is precisely the
bug just fixed. For a *sample*, taking the same pages may be the point: you want the pages that
matter most, and you want two critiques to be comparable.

**How this case measures against it.** `[MEASURED 2026-09-03]` it is **manual-only — no schedule**,
so it accrues no debt between runs; at cap 8, **25 sites** truncate; its 8 are the nav pages plus
the top tools alphabetically. It is a taste instrument, not a coverage instrument.

**The options.** (i) Leave it — the machinery exists, flipping `rotate_coverage: true` on its step
is a one-line migration whenever you want it. (ii) Rotate it — sees the whole site eventually, but
each critique sees a different slice and run-to-run comparison gets noisier. (iii) Curate it —
name the pages deliberately, which needs `bugs_open/452` implemented first.

**My recommendation: leave it (i)**, unless you want design critique to become a coverage
instrument. Its prefix is currently acknowledged by name in the reader's acks file, so it is a
reviewed exception rather than an oversight.

### Decision B — accept webdesign's new-defect detection latency, or buy it back?

**What the thing is.** Two different latencies. *Confirmation* latency is "a repair shipped today,
how long until the audit re-measures the page and withdraws the work item". *Detection* latency is
"a NEW defect appeared today, how long until anything looks at that page".

**The rule.** Migration 469 is your instruction of 2026-08-18: seven days to confirm a repair was
unacceptable, cut to three.

**How this case measures against it.** **Confirmation latency is preserved at 3 days** — that is
what the union window is for; every page with an open finding is in every run. **Detection latency
moved**: for webdesign's first 60 pages it went from 3 days to one cycle (~9 days), and for the
other 91 it went from **never** to ~9 days. So it is a strict improvement for 91 pages and a
regression for at most 60, and the regression is in discovery, not confirmation.

**The options.** (i) Accept it. (ii) Raise `max_pages` for `render-audit-agent` — now a pure latency
dial with no coverage cliff behind it — at the cost of render-audit pod minutes, which is the
dedicated pod's whole purpose. (iii) Shorten that site's rotation interval, which takes capacity
from the other 30 sites.

**My recommendation: accept (i) for now**, and let the reader tell you if it stops being true. The
number to watch is site growth, not the cap.

### Decision C — is a 14-day dormancy window right?

**What the thing is.** The new watchdog stops judging a site that has gone quiet. `[MEASURED
2026-09-03]` exactly one site is dormant: `loancalculator.co.uk`, whose only truncation row is from
2026-08-11 and which now has 28 live pages against a cap of 60 — it can never truncate again.

**The rule.** A check that alarms on frozen history is a check people stop reading, and the first
wire test proved this one would have: it reported a config regression that had not happened.

**How this case measures against it.** 14 days ≈ 4 missed opportunities at the 3-day per-site
cadence, so a genuine regression still produces recent rows and still alarms. The stated blind spot:
a regression on a site that *also* stops truncating would go unseen — bounded, because a site stops
writing these rows when it fits inside its cap, and a site inside its cap has no coverage debt.

**The options.** Accept 14, or name another number. It is one constant, one line.

**My recommendation: accept.** This is the lowest-stakes of the three; I raise it only because it is
a judgement wearing the costume of a measurement, and you should know which it is.

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

## 7. Spun out, still open

`bugs_open/452` — `page_names` is declared in `request_render_audit`'s spec and read by nothing.
LOW; no live carrier sets it, so nothing is broken. It costs a false affordance (Decision A option
(iii) would reach for exactly that key) and a slot of the N=10 optional-key budget.

## 8. Commit trail

`95a04168c` cursor · `72b16391b` R1 fix · `41b03241d` reader + registry flip · `ea08c831d` R3
advisories + `Council-Reviewed` · `c71b46be0` migration 660 applied · `faf4872ce` identity fix ·
`a3610ea23` artefact proof · `667973351` acceptance passes · CronJob + dormancy · `3749132e0`
registry `why` · `ad8eb5f4f` pod proof · `a20de1431` **close(394)** · `0d53d2e18` file 452 +
pointer correction · `953733d1c` register LIVE.
