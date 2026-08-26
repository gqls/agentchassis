# HANDOFF — bug sweep, 2026-08-26 (session "bugsweep2")

**Read this first, then the three lane directories it names.** Everything below is either
`[MEASURED <date>]` with the query, or marked as unverified. Where I got something wrong it is
recorded rather than quietly corrected — four entries went into `WRONG_CALLS.md` today and the
pattern across them is the most transferable thing in this handoff (§6).

---

## 1. STATE IN ONE TABLE

Three bugs taken from `bugs_open/`, all previously **unowned**. All three fixes are **committed,
LIVE on chassis `v1.0.1345`, and their config applied**.

| bug | what it is | council | live? | what remains |
|---|---|---|---|---|
| **359** archived page still serving | nothing detected a RETIRED page still answering the public | **APPROVED r1** `6ce98a66` | ✅ **CLOSED — proven in production** | nothing. Moved to `bugs_closed/` — §3 |
| **407** site can't set its own header | header membership was a hardcoded fleet-wide page-NAME list | **APPROVED r1** `cb67cc71` | ✅ code + migration 654 applied | **ONE OWNER DECISION** (§4) + a nav rebuild |
| **404** re-render reason vocabulary | 4 readers, and every one that doesn't know a value fails toward the SILENT branch | **REVISE r1 AND r2** `f2e4ac2a` | ✅ code + migration 656 applied | **ROUND 3 OWED** — §5 |

Commits: `36b51a51b` + `f5108dd47` (359) · `74e92e961` (407) · `ef4236b4d` (404) · plus docs,
registers and the HOLD discharges. `scripts/verify-head-builds.sh` was green at each stage.

**Lane docs** (always give the path — owner ruling 2026-08-19):
- `docs/agent_docs/docs024_key_docs_latest/bugfix_359_archived_page_still_serving/`
- `docs/agent_docs/docs024_key_docs_latest/bugfix_407_site_declares_its_own_header/`
- `docs/agent_docs/docs024_key_docs_latest/bugfix_404_rerender_reason_vocabulary/`

---

## 2. WHAT I DID AFTER THE ROLL, AND THE TRAP IN IT

The `v1.0.1345` roll landed, so all three holds became dischargeable. **Every one was probed at the
artefact before its config was applied**, with a positive AND a negative control in the same query:

```
service_binary_capabilities, kind='discovery_check', BOTH live chassis pods:
  archived_page_still_serving   present      site_unreachable  present (+ctl)   …_NOTREAL 0 (−ctl)
binary grep, BOTH live pods:
  header_slots 3    RerenderSectionReasons 2    site_unreachable 7 (+ctl)   zzz_not_a_real_symbol_404 0 (−ctl)
```

⚠ **AND THE FIRST READING WAS A TRAP WORTH THE PARAGRAPH.** Over a 45-minute window the capability
query said **95 of 428 pods** carried the new check — which reads as a half-finished rollout and
would have stopped me applying 648. It is wrong: `agent-chassis` **spawns short-lived pods**, so a
wide window mixes pre- and post-roll ones. There are only **two** long-lived deployment pods, both
had it, and narrowing to a 10-minute window gave **153 of 153, zero without**.

**Ask the WINDOW, not just the count, on any service that spawns.** A wide window on a spawning
service manufactures a mixed-image reading out of pods that no longer exist.

Applied, in the order each file's header requires, each verified at the LIVE OBJECT afterwards:

- **648** → `availability-discovery-agent` checks array is now 3:
  `["site_unreachable","page_content_divergence","archived_page_still_serving"]`
- **656** → the fixer's query now carries `p.status = 'active'`, reason stamp intact
- **654** → `ai-agent-orchestration.com :: slots=9 cap=9 siblings_kept=2` (the sibling count is the
  load-bearing half: `locale` and `analytics` survived, so the write set ONE path)

**"What did I break?" was asked before "did it work?"** — `DISCOVERY_CHECK_ERROR` rows in the 40
minutes after: **0**. Both HOLD files are renamed (suffix off, discharge banner added, 541's
convention) and all three are recorded in the migration ledger.

---

## 3. 359 — ✅ CLOSED, PROVEN IN PRODUCTION ON THE DISCONFIRMING PAIR

Done and moved to `bugs_closed/`. Recorded here because HOW it was proven is the reusable part.

The rotation takes one site per 300s with a 4-hour floor, so it had not reached an affected site
and the detector had filed nothing — and **a zero at that point is not a pass**, which the bug file
says in terms. So `robot-hands.com` was brought forward one tick (it carries the pair on one site)
and run `9ca6a795` followed at 22:32Z:

```
gripper-catalog                /gripper-catalog.html                 FLAGGED      (200)
news                           /news.html                            FLAGGED      (200)
gripper-cycle-time-estimator   /gripper-cycle-time-estimator.html    not flagged
gripper-payload-calculator     /gripper-payload-calculator.html      not flagged
learning-center-article        /blog/learning-center-article.html    not flagged
learning-center-index          /learning-center/index.html           not flagged
```

**Both arms.** The two the census said were serving are the two flagged — `gripper-catalog` being
the page first recorded serving on 2026-08-14, so twelve days on something has finally said so.
The four that correctly 404 were **not** flagged, which is the arm that rules out a detector that
flags everything. `checks_failed: []`, so both instrument controls held; items are `medium`/50,
`detected`, `handler_agent` EMPTY.

⚠ **Nothing needs restoring from the rotation nudge** — the site was then legitimately swept, so
its new stamp is the correct one.

**Still expected, and it is not a defect:** the other four sites' serving pages
(`ai-agent-orchestration.com`, `finetuning.uk`, `fundamentallyai.com` ×2,
`leopardessconsulting.co.uk`) will be flagged as the rotation reaches them, ~4–8h.
`scripts/audit-archived-still-serving.sh` is the expected set **on the day** — re-run it rather
than quoting any list, because the population moves.

**And the items are flag-only, so draining them is a human act.** Each carries a `triage_hint`
naming both remedies — retract, or **un-archive** if the page was retired by mistake — and the
`agent_error_log` query to read first.

## 4. 407 — ONE OWNER DECISION IS OWED, AND A NAV REBUILD

**The decision** (`bugs_open/407` §B, with both sides written out). The fix lets a site's declared
header override **three** membership rules at once — `pages.in_header`, `neverPrimaryTypes`
(blog-post / tool / entity-page) and the child-URL bar. The council's `guardian` seat objected,
fairly, that this is a widening beyond "fix the tier order" and is the owner's call, not mine.

- **For:** all three are fleet-wide DEFAULTS, and the point of the change is that a default may not
  outrank a site's own word. `idea.uk/report.html` is the worked case — the site set `in_header`,
  set `nav_order` 3, and gets nothing because the platform has decided pages of type `tool` are
  never in a header.
- **Against / cost of saying no:** the fix still solves finetuning.uk (tier) and gaswholesalers.com
  (same-tier tie) either way; only `idea.uk/report.html` stays unpromotable.
- Two rules are **not** overridable in either version — system pages and legal pages. Those are
  correctness, not preference, and a site declaring one is TOLD so rather than obeyed.

**The nav rebuild.** The declaration is applied but **does nothing until `populate_nav_tables` next
runs for the site**, and the served header stays stale until the re-render items it files drain.

```
{"action":"process","agent_type":"nav-updater","data":{"domain":"ai-agent-orchestration.com"}}
```

Then read the step's OWN result before inferring anything:
`nav_declaration_source` (`default|site_config|invalid`), `declared_slots`, `declared_missing`,
`declared_ineligible`, `declared_flag_disagreed`, `max_header_items_effective`.
`invalid` means the spec holds a shape the reader could not use — fix the spec, don't leave it
half-read. Full queries: the lane RUNBOOK.

⚠ **VERIFY AT THE SERVED PAGE, AND ANCHOR ON `<nav>` — NEVER ON A BARE `<header>` TAG.** A rendered
page carries several `<header>` elements because components carry their own; matching the wrong one
gave me a 420-byte "header" with no links in it during this investigation. The tell: a site chrome
with zero `href` is not a chrome.

---

## 5. 404 — ROUND 3 IS OWED, AND MOST OF IT IS ALREADY BUILT

Round 2 came back **REVISE** again (gating HIGH from `debug_historian`). **The gating objection is
already fixed in the tree** — I have not yet resubmitted.

| objection | state |
|---|---|
| `debug_historian` [HIGH] — 656's `SELECT ... INTO` had no `STRICT` and no row-count guard; a `type`-scoped read can take an arbitrary row of a duplicate pair while the UPDATE writes to ALL | **FIXED** in `656` (count guard + `INTO STRICT`, both reads). `[MEASURED 2026-08-26]` all three affected agents have exactly 1 active row, so it was inert — shipped because inertness is not safety |
| `editquality` [medium] ×2 — the r2 SUBMISSION's sketches did not show the named constants or `TestReasonConstantsAreExactlyTheVocabulary`, which the narrative claimed | **A SUBMISSION-ACCURACY FAILURE, NOT A CODE ONE.** Both exist in the tree and are mutation-proven; my sketches were stale. Fix the sketches in r3 |
| `debug_historian` [medium] — no pod-grep deploy verification | **NOW DONE**, §2: `RerenderSectionReasons` = 2 on both live pods with controls |
| `bug_historian` [medium] — the WARN lives only in `create_rerender_items`; other writers stamp `spec.reason` directly and get no loudness | **REAL AND UNFIXED.** This is the gate-side loudness deferred as RFC-scope. State it explicitly in r3 rather than leaving it implicit |
| `guardian` [medium] — enumerate callers already passing the two new reasons through this action who would see a BEHAVIOUR change | **MEASURED: ZERO**, over full live+archive history. Say so as an explicit claim in r3 |

**To resubmit:** `RESUBMIT_CORR=f2e4ac2a-2bfc-4c82-ac99-d5fd7616edef` with the sketches corrected.
Draft to start from: `scratchpad/submission_404_r2.json` (session scratch — re-create if gone).

---

## 6. THE THING I MOST WANT THE NEXT SESSION TO TAKE

I wrote **three** "break this line and that test fails" mutation tables today. **Every one had rows
that were false when I wrote them** — and each for a *different* reason:

| lane | why the guard was green and inert |
|---|---|
| 359 | the tests were passing on a **guard in series** — a second failure downstream, and the assertion was only `err != nil` |
| 407 | the **fixtures could not produce the failure** — a declared order that was already alphabetical; a corpus whose utility group was empty |
| 404 | the **discriminator could not see the real shape of its input** — the lint scanned migrations for a literal and matched ONLY the prose comments, because executable SQL doubles its quotes. It reported "12 checked" and passed |

None was visible by reading. All three were found by breaking the code and watching whether the
test noticed. The general statement, which is now in `WRONG_CALLS.md` three ways:

> **A passing test proves the current code is acceptable to that test, and nothing else** — not
> that any line is load-bearing, not that the fixture reaches it, not that the scan can see it.

Two habits came out of it and both are cheap:
1. **Never write a mutation row you have not executed.** A table written from the design is a
   design document wearing the costume of evidence.
2. **Pin every scan to a KNOWN POSITIVE, never to a non-zero count.** "It found 12 things" answers
   *did it run*, never *can it see what it is for*. 404's lint now carries positive controls naming
   migrations 460 and 473.

A fourth entry records asserting an absence — *"there is NO `site_specs` table"* — in a **council
submission**, having read it off a `\dt site*` listing **I had truncated at twenty rows**. It sorts
on line 21. Never conclude absence from a command you truncated; an estate-level absence claim
deserves a second instrument.

---

## 7. WORK DONE ALONG THE WAY THAT IS NOT ANY OF THE THREE BUGS

- **`scripts/audit-archived-still-serving.sh`** — the census as a command, both controls built into
  its verdict logic, `--self-test`, and a control that refuses when its own loop is truncated.
- **`liveConfiguredChecks` refreshed by union** — the guard against a fleet-wide discovery outage
  was asserting **63 of the 82** names live agents configure, blind to a whole fifth agent.
  Mutation-proven load-bearing.
- **`LANDMINES.md` contribution** — the "refuses when already applied" guard is **defeated by
  ordering**, and migrations **526 and 541 both have it that way**: they snapshot before the
  idempotency check, so a replay writes a "pre-update" row holding post-update state. Verifier
  dispatched.
- **Registers**: `IMP-058` (359), `NAV-014` (407), `REB-002` extended (404). **016b §9** gained the
  transferable pattern from 407 — *when two causes wear one symptom and a timestamp cannot separate
  them, REPLAY the deterministic classifier as a query*. **016b §10** gained rows for all three.

---

## 8. PICKING THE NEXT BUG — THE SWEEP, SO NOBODY RE-WALKS IT

`[MEASURED 2026-08-26]` of 39 candidate numbers checked with `scripts/who-owns.py`, only **three**
had no ACTIVE owning workstream: **338**, **356**, **404** (taken).

- **`338`** — the voice gate's DENSITY rules applied to a single sentence, where they are not
  measurements. **Genuinely open and unowned. This is the next one to take.**
- **`356`** — fixed in the tree, awaiting a roll; its remaining work is **17 separate routing gaps,
  each needing its own judgement**. That is a programme, not a bug fix.
- **`348`** is superseded by its own banner (*"Do not fix from this file — `bugs_open/344` OWNS
  THIS"*, and 344 is closed). Don't re-pick it.

⚠ **Most low-numbered "OPEN" bugs are FIXED AND LIVE** and stay in `bugs_open/` only by the owner's
2026-08-06 ruling — 181, 207, 210, 217, 220, 221, 222, 230, 233, 244, 247, 249 all read that way.
**"OPEN" in the index is not a work-available signal; open the status header.**

---

## 9. THE FIRST FIVE COMMANDS FOR WHOEVER PICKS THIS UP

```bash
cd /home/ant/projects/agentchassis
git log --oneline -25                      # what has landed since; the tree is shared
# 1. 359's acceptance — THE PAIR, not one arm (§3)
scripts/audit-archived-still-serving.sh    # expected set, on the day
# 2. did anything break?
#    agent_error_log error_code='DISCOVERY_CHECK_ERROR', last hour
# 3. 404 round 3 — the gating fix is already in the tree, only the sketches are stale (§5)
# 4. 407 — the owner decision (§4), then the nav rebuild
# 5. next bug: 338, voice-gate density rules (§8)
```
