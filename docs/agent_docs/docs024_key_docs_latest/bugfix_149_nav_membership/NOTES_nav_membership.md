# NOTES — nav membership (bugs_open/149 Group A). Append-only, newest at the bottom.

## 2026-07-31 — session 1: A2 + A6 + A4's recordable half

### Picking the lane, and who I stayed off

`scripts/who-owns.py 149` named `oufe` as the workstream with most mentions — but
that lane's 149 work was C1/B4 (done and live on v1.0.1211). The live risk was
elsewhere: `docs024_key_docs_latest/robot_hands_checker_gaps/` was **untracked and
had a file written at 09:17 the same morning**, i.e. a session actively in the tree.
Reading its NOTES showed it working links 2 and 3 of the checker chain — SCHEDULING
and DISPATCH — which is 149's B1/B2/B3 and the 263-detected/0-triaged fact. So I
took **Group A (routing)**, which is next in 149's own suggested order and does not
touch theirs.

`who-owns.py` reads COMMITS. That lane had committed nothing at the time, so it was
invisible to the tool and visible only in `git status`. **Check the tree as well as
the log** — the file's own advice, and it was load-bearing today.

### What the queue said, and where it pointed wrong

149 A2 said `nav_drift` for a `/tools/` URL is "structurally unfixable by its own
handler", and A6 said fix at creation time because "unrepresentable beats
detectable". Both were nearly right and the second was **wrong in its remedy**, for
a reason that only shows up when you read what makes a page reachable:

- A2's mechanism is real and I reproduced it (below).
- A6 proposes writing the nav row at creation. But chrome is a **stored artefact**
  (`bugs_open/117`/`118`), so a nav row ships nothing on its own — while
  `check_orphan_pages` treats the presence of a nav row as reachability. A
  creation-time row would therefore have left the page exactly as unreachable and
  **silenced the only check that would have noticed**. My own write blinding my own
  detector, which is a shape this fleet has recorded before.

So the creators now **request** the rebuild (`nav-updater`, which derives nav AND
re-renders chrome AND propagates), rather than write the row.

### The evidence, and the control that made it a mechanism

The useful thing was not the code read — it was finding a **completed work item
that had named its own targets**, so the failure is on the record rather than
argued:

```
gamesdesign.co.uk | 2026-07-29 | complete 17:27:50 |
  ["bayesian-ranking","tool-drop-rate-tuner","tool-loot-table-balancer","tool-xp-curve-designer"]
```

All four still absent from `site_nav_items` on 07-31. Same shape on
ai-agent-orchestration.com (`tool-ai-agent-roi-estimator`, complete 07-25).

Then the control, which is what stopped this being a cadence story: robot-hands.com's
`nav_drift` item named `["learning-center","news"]` — `/learning-center.html` and
`/news.html`, no child prefix — and **both are in nav today**. Same check, same
handler, same action. The only difference is the URL prefix.

I looked for that control on purpose. "The handler cannot repair X" is much weaker
than "the handler repaired Y and could not repair X, and Y differs from X only in
the predicate I am accusing".

### The fix turned out to be an ORDERING, not a policy

`classifyPagesForNav` already contained the right answer and never reached it: the
`page_type`-keyed never-primary branch sends such a page to `utility` **if it has
declared a flag**, and the URL-keyed branch `continue`d out above it. So the
platform had already decided what to do with "must not be in the main menu but has
declared membership" — and a blunter rule shadowed it. That reframing is what made
the change small and made "am I inventing a placement policy?" answerable with the
function's own doc comment (*"Tier 4 (never primary): individual tool pages…"*).

### Missteps, in the order I made them

1. **I replicated the Go classifier in SQL by listing its predicates instead of
   reading its branch order, and got a wrong answer.** `legalNames`/`isLegalPage` is
   tested **before** any flag, so a legal page needs no `in_header`/`in_footer`. My
   first regression census therefore reported `loancalculator.co.uk /legal.html` as
   a row the fix would stop reproducing. It is reproduced, by the legal branch, and
   always was. Caught by re-reading the function, not by anything the query said —
   **a replica that errs in the same direction as your hypothesis returns a
   plausible number.** Cheap check: replicate top-to-bottom including the `continue`
   after each branch. Logged in `WRONG_CALLS.md`.
2. **I invented a filter and got a one-row answer that looked like a finding.** A
   per-site nav-group census filtered `sites.status='active'` and returned exactly
   one site. That is not how site liveness is recorded here; a plain `GROUP BY` over
   the join answered it. I noticed because "one site has nav groups" is absurd, not
   because I checked the filter — the absurdity was luck, and a plausible number
   would have gone through.
3. **`grep -o -E '<nav[^>]*>.*</nav>'` on a served page returns nothing even when
   the nav is present**, because the markup spans lines. My first check for "is this
   tool link in the live header?" came back empty and I nearly read that as "no tool
   links", which was the answer I expected — the same failure as (2) with the same
   near-miss. The `sed '/<header/,/<\/header>/p'` range gave the real answer, which
   happened to agree, but for the right reason this time.
4. **I nearly skipped the diagnosis loop on a defensible argument.** I had the code
   path, a completed item, four still-absent pages, a positive control and an
   independent second source in `LANDMINES.md` — and 149 A2 had already asserted the
   structural claim two days earlier, so I was not filing a new one. Then I noticed
   HEAD had moved 26 minutes earlier: **owner ruling 2026-07-31 (`f75432a39`)**
   makes the loop a stated norm for exactly this class, with a declared escape
   hatch. I ran it rather than use the hatch, because the council submission was
   already queued so it cost no serial time. The reason this is a misstep and not a
   judgement call: **I read CLAUDE.md at session start and acted on it 90 minutes
   later without re-reading it**, on a tree where the rules are co-edited by other
   sessions. Cheap check: `git log --oneline -5 -- CLAUDE.md` before acting on a
   multi-session rule.

### Verification done before rolling

- Package green: `ok github.com/gqls/agentchassis/platform/orchestration/actions`.
- `git archive HEAD` builds clean (the shared-tree check — another session's commit
  plus mine must compile together).
- **The guard test was watched failing.** With the old `continue` temporarily
  restored, 4 of 5 subtests fail with `in utility = false, want true ... utility=[]`,
  and the two control tests still pass. A guard test never seen to fail is
  indistinguishable from one asserting nothing.
- **The image was proven before the roll, not after.** `docker run --entrypoint sh
  aqls/agent-chassis:v1.0.1215` + `strings /app/agent-chassis`:
  `declares no nav membership` = 1, `nav_membership_declared` = 1,
  `classifyPagesForNav: skipping child page` = **0** (the string this fix REMOVED),
  `CONTENT_CLAIMS_FLOOR_DETAIL` = 1 (positive control from 149 C1's fix, proving the
  binary is a real chassis build and my grep works).

### Two ordering facts worth inheriting

- **`IMAGE_TAG` is `?=`, so `make build-agent-chassis IMAGE_TAG=v1.0.1215` needs no
  file edit.** The makefile was dirty with another session's bump (tree `v1.0.1214`,
  HEAD `v1.0.1206`); editing it would have put my change inside their uncommitted
  diff.
- **A chassis roll kills an in-flight council run and an in-flight diagnosis run** —
  both execute in those pods. So: build immediately (touches nothing), roll only
  after the verdicts land. That is why this session built at 09:11 and rolled later.

## 2026-07-31 (afternoon) — the fix is LIVE and PROVEN, and post-roll verification found a defect the diff could not show

### The proof, end to end

Pod-verified on **both** replicas of `v1.0.1215`: `declares no nav membership` 1,
`nav_membership_declared` 1, `nav_rebuild:%s` 1, and the string the fix **removed**
(`classifyPagesForNav: skipping child page`) **0**, with `CONTENT_CLAIMS_FLOOR_DETAIL`
1 as a positive control and a nonsense needle at 0 as a negative one.

Then the behaviour, on the site the evidence came from. gamesdesign.co.uk before:
`primary` 5 · `tools` 1 · `utility` 1, with five flagged `/tools/` pages absent from
nav. After one `nav-updater` run: `primary` **5** (unchanged — the control),
`utility` **7**, the `tools` group **gone**, and all six flagged tool pages placed in
`utility` — including the exact four that the 07-29 `nav_drift` item had completed
without placing. The bespoke `tools`/`primary` group self-healed with no migration,
because `populate_nav_tables` deletes a site's groups before rebuilding, exactly as
predicted.

### The defect that verification found, and the diff could not

**Six live footer labels read `Tools/Damage Formula Designer/Index`.**
`navSimplifyLabel`'s URL fallback derived its name from the **whole path** and
title-cased it — correct for as long as only FLAT pages reached it (`/about.html` →
"about"), and this change routed child pages there for the first time. **Nothing in
the fix touches that function**, so no amount of re-reading the diff would have found
it. Only the rendered rows did.

Fixed in `c053bb31f` (`navLabelSegmentFromURL`: last non-empty segment, ignoring a
trailing `index`; leading `tool-` stripped because it is an internal convention that
the directory-style URL already omits). **And the control test caught my first
attempt**: skipping `index` unconditionally made `/index.html` return `""`, so the
homepage lost its "Home" label — a worse nav regression than the one I was fixing.
`TestNavLabelFlatPagesUnchanged` is why that is in a commit and not in production.

**The transferable lesson: a change that widens what reaches a function can break the
function without touching it.** The blast-radius query I ran before submitting counted
ROWS the derivation would add. It could not have told me those rows would be badly
LABELLED, because labelling is a different function and the diff is silent about it.
Ask additionally: *what code now receives inputs it never saw before?*

### Two things I got wrong about the cluster, both by reading a stamp

1. **I told the council the dispatch lane was alive, on a 7-day aggregate, and it was
   dead at that moment.** Answering the `bug_historian` seat's medium objection I
   measured `nav_drift` 17/17/17 all-history, and 1,580 of 1,664 `build` items claimed
   over 7 days, and `build-pipeline-trigger` `enabled` with a recent
   `last_triggered_at`. All true. But **claiming had stopped FLEET-WIDE at 13:21** —
   two hours before I quoted those numbers — with 9 items stalled and the oldest from
   13:56. My work item sat `triaged` and unclaimed for 15 minutes, which is how I
   found out. The seat's objection was **better than my rebuttal**: I answered "is
   this lane alive?" with a window that could not see "right now", which is the third
   time today I have been caught by the shape of my own measurement.
   `last_triggered_at`/`last_completed_at` advancing is a **fire-and-forget stamp** —
   it proves the scheduler fired, never that a dispatch-loop orchestration was
   created. The tell was `complete_idle`, whose newest run was 13:18.
2. **I waited ten minutes for two councils that were already corpses.** Before rolling
   I polled for `review_*` steps and watched two sit unchanged, then clear, and read
   that as "they finished". Another session's commit (`99f387fd3`) shows what really
   happened: their round 3 was **killed at 15:00:20 by another session's chassis
   roll** — the pods I later found already on `v1.0.1215` were that roll, not mine.
   So my own `make deploy` at 15:10 changed no pods and killed nothing, and my polite
   wait was watching a dead run. **The tell is pod `startTime` against the run's last
   `updated_at`** — they wrote that landmine the same afternoon, so it is covered
   there rather than duplicated here.

That second one has a corollary I should have seen sooner: **another session rolled
the image I had built and pushed**, so my change went live at 15:00 without my doing
anything — which is exactly what CLAUDE.md says happens on this tree, and it is why
"hold the deploy pending review" is not available here.

## 2026-07-31 (evening) — all three fixes verified on v1.0.1218, and a near-miss about somebody else's closed bugs

**Verified live on `v1.0.1218`, both replicas, 18:07 UTC** — A2's marker 1, A6's marker 1,
`navLabelSegmentFromURL` 2, the removed string **0**, positive control 1. So the label fix
(`c053bb31f`) is now live too, shipped by another session's roll rather than mine, which
is the third time today that has happened and is simply how this tree works.

**Dispatch is still down.** Newest claim anywhere `15:52:35` (136 min before the check),
gamesdesign still holding 34 `page_rerender` items at `triaged`, served footer unchanged.
Note the newest claim moved from 13:21 (afternoon reading) to 15:52 — so the lane
recovered briefly and stopped again, which is worth knowing for whoever diagnoses it and
is more useful than "it is dead".

**The near-miss, and it is the day's fourth instance of one shape.** At ~17:46 I found the
pods running **`v1.0.1216`** — nine minutes after another session committed
`close(145): LIVE and pod-verified on both replicas of v1.0.1217`. `git merge-base
--is-ancestor` settled that 1216 genuinely predates their fixes by **65 commits**, so on
that evidence two just-closed bugs had been rolled backwards into being inert again, and
I started drafting correction banners for their `bugs_closed/` files.

Then I checked the replicasets before writing: **`v1.0.1218` had been rolling since
17:58**, and by 18:07 both pods were on it. The 1216 window was **thirteen minutes** and
had already closed. Nothing was wrong; my alarm was an artefact of when I happened to
look.

Two things I want to keep from that:

- **A cluster snapshot can go stale inside a single investigation on this tree.** I had
  measured carefully (ancestry, not a needle) and would still have filed a false
  correction against another lane's closed work, because the *measurement* was right and
  the *world* moved. Re-read pod state immediately before asserting anything about it —
  not at the start of the reasoning that leads to the assertion.
- **My first needle for that check was from a `_test.go` file**, which is never compiled
  into the binary, so its `grep -c` = 0 meant nothing at all. I caught it because a zero
  from an unvalidated needle is now the thing I distrust most (three entries in
  `WRONG_CALLS.md` today), and switching to `git merge-base --is-ancestor` answered the
  question without a binary marker. **Ancestry beats a string grep for "is this commit in
  that image", and it cannot be fooled by a comment or a test file.**

## 2026-07-31 (18:00–18:30 UTC) — second-site proof, and it found the half of the invariant that was never covered

Picked up cold from `HANDOFF_2026-07-31_continue_here.md` §4 item 1: prove the A6 label
fix on a site that actually exercises it, which the `guardian` seat had asked for before
any fleet-wide dispatch.

**Re-measured first, and the handoff's own headline had already moved.** §3 said the build
dispatch loop was dead with the newest claim anywhere at `15:52` (136 minutes). At 18:13
`max(claimed_at)` fleet-wide read **1 minute ago** — which looks like recovery and is not.
Grouping by `claimed_by`:

```
diagnose-dispatch-loop   | 18:14:07 |    6 min | 4 claims/6h
build-dispatch-loop      | 15:44:10 |  156 min | 16 claims/6h
```

The only lane claiming is **another lane's diagnosis run investigating the outage itself**.
So RUNBOOK **R12 query 4** — the query this lane added that morning, precisely to stop a
7-day aggregate hiding a dead lane — has the **same defect one level down**: a fleet-wide
`max()` cannot tell "the queue recovered" from "somebody is standing next to the corpse".
R12 updated to require `GROUP BY claimed_by`. Third time this lane has been bitten by an
aggregate that answers a narrower question than the one asked.

**Choosing the site — the handoff's candidate list was wrong, and wrong in an instructive
way.** It listed candidates by *count of flagged tool pages* (`gaswholesalers (4),
finetuning (4), ai-agent-orchestration (4), vonc (3), oufe (2), fundamentallyai (2)`). The
property that decides whether the code runs is `nav_label IS NULL` — an authored label
≤30 chars is returned verbatim and `navSimplifyLabel` is never called. Measured:

| domain | flagged child pages | of which `nav_label` NULL |
|---|---|---|
| gaswholesalers.com | 4 | **4** |
| finetuning.uk | 4 | 2 |
| ai-agent-orchestration.com | 4 | 2 |
| leopardessconsulting.co.uk | 5 | 1 |
| fundamentallyai.com | 3 | 1 |
| **vonc.com** | 3 | **0** |
| **oufe.com** | 2 | **0** |
| gamesdesign.co.uk | 6 | 0 |

> **CORRECTED 2026-07-31:** vonc and oufe were named as candidates in the handoff and
> **would not have exercised the fix at all** — every one of their tool labels is authored.
> They would have produced a clean-looking result that proved nothing, which is the same
> failure as gamesdesign and the reason a second site was asked for. A candidate list
> filtered on the wrong column is worse than no list, because it looks like it was checked.

**The run.** `./TRIGGER_nav_rebuild.sh gaswholesalers.com`, corr
`fd85fc84-f021-4e47-801a-db4c90174127`, 18:16 UTC, chassis v1.0.1218, pre-flight
`unreproducible_rows=0`. COMPLETED; no `__step_error` key in `collected_data`.

Prediction written down before the run, and it held exactly: all four land in **utility**
(never-primary by URL, flag present), 19 rows → **23**, existing labels unchanged.

```
utility |  9 | Fuel Cost Estimator         | /tools/tool-fuel-cost-estimator.html
utility | 10 | Gas Unit Converter          | /tools/tool-gas-unit-converter.html
utility | 11 | Breakeven Volume Calculator | /tools/tool-breakeven-volume-calculator.html
utility | 12 | Fuel Budget Forecaster      | /tools/tool-fuel-budget-forecaster.html
```

**The decisive fingerprint, and it is better evidence than the count.** The page title is
`Wholesale Break-Even Volume Calculator | Tools`; the derived label is
`Breakeven Volume Calculator`. **`Break-Even` vs `Breakeven`** — the label cannot have come
from the title, only from the URL slug, and nothing but `navLabelSegmentFromURL` strips the
`tool-` prefix. This site also uses the flat `/tools/tool-x.html` convention, so the prefix
strip ran in production for the first time; gamesdesign is all directory-style.

Chrome re-rendered: the stored footer carries all four links, header carries none (correct
— utility is a footer group), and every anchor text is clean.

**A false positive of my own, caught within one query.** My first chrome check was
`rendered_html ILIKE '%Tools/%'` for a mangled label, and it returned **true**. It was
matching the `href` — `/tools/tool-fuel-cost-estimator.html` — case-insensitively. A needle
that matches the thing you are looking for AND the thing you are looking at is not a check.
Fixed by extracting anchor *text* rather than scanning the whole document. Logged in
`WRONG_CALLS.md`.

Served pages unchanged as expected: 0 `/tools/` links on the live homepage, 31
`page_rerender` items filed at 18:16:30, all sitting `triaged`. Nav tables ✅ → stored
chrome ✅ → served pages ⛔, exactly as §3 predicted, for exactly the reason §3 gives.

### The finding: the invariant only ever covered one of two paths

While measuring which sites had authored labels I read the authored values, and three sites
carry labels shaped `Tools / X`. `navSimplifyLabel`'s own doc comment says **"a label
containing a slash is never right"** — and A6 made that true of every label the derivation
*computes*. `navLabelForPage` returns an authored `nav_label` **verbatim** when it is ≤30
chars, so the guarantee stopped at the function boundary beside the one that was reviewed.

Not hypothetical — one had already reached a live nav row:

```sql
SELECT s.domain, ni.label, ng.group_type FROM site_nav_items ni ... WHERE ni.label LIKE '%/%';
 ai-agent-orchestration.com | Tools / AI Readiness Quiz | utility
```

8 pages carry a slash-shaped `nav_label`, 7 flagged for nav, exactly **2** short enough to
be trusted verbatim (25 and 30 chars).

**Siblings checked before fixing only mine**, per the `bugfix_072` landmine. There are
**four** label paths and the picture is not what `nav_tables.go`'s header comment claims:

- `navSimplifyLabel` — fixed by A6. Live.
- `navLabelForPage` — the gap. Live. **Fixed here.**
- `rerenderSimplifyNavLabel` (`rerender_pages_actions.go:440`) — **still contains the exact
  pre-A6 defect** (`name := strings.TrimPrefix(url, "/")`, the whole path). Reads like a
  live bug. It is **dead**: its only callers are `rerenderGetHeaderNavFromDB` /
  `rerenderGetFooterNavFromDB`, whose sole call site is **commented out** at
  `rerender_pages_actions.go:168`. Deliberately left alone — deleting reachable-looking
  code is a separate decision from fixing a live label.
- `datahelpers.SimplifyNavLabel` — takes a page **name**, not a URL, so it cannot
  title-case a path. Not at risk.

`nav_tables.go:554` says `navSimplifyLabel` "consolidates" the other two. **That is the
intent, not the state** — both still exist. A comment describing a consolidation that never
happened is indistinguishable from one describing a real one.

**The fix** (`6fc1ff3ed`): `navLabelDropCategoryPrefix` takes the last `/`-separated segment
of an authored label **before** the length test. Ordering is the load-bearing part — a label
still oversized once its prefix is gone (`Tools / Agent Architecture Complexity Estimator`
→ 39 chars) still falls through to `navSimplifyLabel`, so shedding a prefix cannot smuggle a
39-character item into a footer column. Last-segment rather than URL-derivation because the
URL flattens the planner's capitalisation: `ai-readiness-quiz` → `Ai Readiness Quiz`, where
the planner wrote `AI Readiness Quiz`.

Guard proven the R9 way — reverting the one line gives 4 FAILs naming the served label,
restored gives `ok`, window a few seconds. `git archive HEAD` + these two files only:
`go build ./platform/...` and the package tests both green, so the commit cannot break HEAD
through another session's WIP in the same package.

Council `11c5c813-dfad-437e-b4a9-09c56475e8d2` submitted; committed with
`Council-Submitted:` rather than holding code for the verdict.
