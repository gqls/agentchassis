# HANDOFF — Track A is DONE. Owner ruled on the decisions; four are actioned. START HERE.

> **⚠ SUPERSEDED 2026-08-14 — START AT `HANDOFF_2026-08-14_continue_here.md`.**
> Everything ruled here is executed or carried forward there. This file remains the
> record of the D1–D6 rulings and their outcomes; it is no longer the entry point.

> **UPDATED 2026-08-11 (afternoon), after the owner's rulings and a fresh chassis
> build (v1.0.1286).** The decisions below are kept with their reasoning because the
> reasoning is what a later session needs; each now carries its OUTCOME. Read §D1
> first — its answer changed the shape of Track B.

**Written 2026-08-11 (afternoon)**, revalidated against chassis **v1.0.1286**.
Supersedes `HANDOFF_2026-08-10d_track_a_prose_decomposition.md` (that brief is
**executed**, not pending) and is the entry point after `HANDOFF_2026-08-10c`.

> ~~**Do not start Track B on the strength of this file.**~~ **SUPERSEDED — OWNER
> RULINGS, 2026-08-11 (evening), given to the lane session in chat:**
> 1. **Track B: GO.** One page at a time, checks between, per §D1's procedure.
> 2. **Bug 251 (canonical): FIX NOW.** Before 252's og: half, per the dependency.
> 3. **Bug 252: lane session's discretion.** Plan of record: og: half via the
>    `spliceMetaDescription` placeholder seam after 251 lands; lang half deferred
>    (needs a per-site language field decision).
> 4. **Site plans (D6): SEED.** Seed the site spec, let the planner plan, reseed
>    until the plan is *reasonably close to where we are*. Owner's constraints,
>    verbatim in effect: the exact combination/makeup of calculators and guides is
>    NOT important; **keeping the overall size, density and complexity IS — the
>    site must not shrink on rebuild**; the improvement loop growing it over time
>    is welcome. (Site is new; no visitor-facing risk from URL churn.)
> 5. **Complaint-deadline oracle + Track C: as recommended** — the oracle as its
>    own small job, Track C after Track B.
> 6. **Index rewrite: option B.** The agent's first output read like *a different
>    site* — seed a `content_direction` (the cards are good and STAY), re-run,
>    compare again.
>
> The lane session (this one's owner-facing thread) is executing; check
> `TaskList`/NOTES before duplicating any of it.

---

## 0. State in one paragraph

**All 17 LMC prose pages are decomposed and live, each byte-identical to its offline
prediction.** Zero generic verbatim pages remain on this site: 18 pages on
`["prose-0"]`, 22 owned verbatim (Track B), 1 owned decomposed (`loans-consolidation`).
Arithmetic untouched (`oracle.py` 170 PASS / 0 FAIL, both controls fired in-session).
Byte gate passes. **Three** bugs were filed on the way — `250` (both halves now
fixed), `251` and `252`, the last two both fleet-wide platform defects. Nothing is
half-applied and nothing is waiting on a queue.

## 1. Verified on v1.0.1286, this afternoon — not inherited

| check | result |
|---|---|
| `189`/`204` pod-grep, **both** replicas | `1 / 1 / 0` (negative control clean) |
| 17 pages served == `predicted/` | **17 / 17**, 0 differ |
| **offline mirror still valid under the new binary** | re-rendered `legal` on v1.0.1286 → **byte-identical** to the prediction |
| `gate_component_bytes.py` | GATE PASSES (22 verbatim exact, 21 assembled skipped) |
| `oracle.py` + controls, same session | 170 PASS / 0 FAIL / 6 CONVENTION; `--selftest-parse` OK; `--mutate expectation` → 4 FAIL, 0 passed |
| `verify_site.py` | 1 FAIL — the canonical, `bugs_open/251`, **not ours** |

**⚠ Re-do the mirror check after every chassis roll.** The lane's whole safety model
is that `assemble_mirror` + `inject_canonical` predict offline exactly what the
chassis renders. A binary that changes head assembly, JSON-LD or canonical injection
invalidates every `predicted/` file **silently** — the diff only runs when a human
runs it. It is one rerender and one diff.

---

## 2. THE DECISIONS — and what was ruled

### D1 · Track B — go, or hold? **(the big one)**

22 owned verbatim pages, every one a live consumer-finance calculator. Different risk
class from Track A: the failure mode is silent and the blast radius is a tool people
act on.

**Do not go wide until the re-slot question is measured.** `HANDOFF_2026-08-10c` §6
marks it `[INFERRED]` and Track A **could not** settle it — none of the 17 has a
locked row, so nothing exercised `matchLockedRow`. The claim needing a measurement:
an incoming composition that omits the tool slot causes the locked row to be moved to
`position = len(sections)+1`, i.e. **the calculator lands at the bottom of the page**.

- The cheapest subject is **`loans-consolidation`**, already decomposed with a locked
  `tool-1`. Measuring it means firing a build-path run at a live calculator — so it
  is itself a decision, not a free check.
- Watch for two things: the locked row's `position`, and a `lock_blocked` work item,
  **which is the only non-silent signal**.

**Options:** (a) authorise the one-page measurement, then decide Track B on the
result; (b) authorise Track B outright; (c) hold Track B entirely.

> ### ✅ RULED (a) — MEASURED, and the answer REFRAMES the trap. Track B is not blocked by it.
> Settled without firing at a live calculator, by test:
> `platform/orchestration/actions/save_sections_positional_tool_slot_test.go`.
>
> **The original framing was wrong.** `matchLockedRow` compares the locked row's slot
> against the **incoming section name**, and a positional name is a string like any
> other. A composition built from `pages.sections` — which for `loans-consolidation`
> **is** `["prose-0","tool-1","prose-2"]` — matches `tool-1` **exactly, on the first
> branch**, is consumed, and stays put. Probed the normaliser too: `tool-1` → `tool-1`
> and `tool_1` → `tool-1`, so a positional slot is matched **twice over**.
>
> **So the trap fires only when the incoming composition OMITS the tool slot** — which
> is what a **seeded site plan** would produce if it names sections semantically.
> **The dangerous act is D6 (seeding a plan), not Track B's per-page rerenders.**
>
> ⚠ **The induction failed honestly first and that matters:** disabling the exact-match
> branch alone still PASSED, because the kebab branch catches it — two guards in
> series. Only disabling **both** fails. A mutation that passes has usually hit a
> guard in series; do not read it as "the test is fine".
>
> **Still genuinely unmeasured:** the live end-to-end path. No writer run has been
> driven against a decomposed page holding a locked row. The matching *rule* is now
> pinned; the pipeline around it is not.

### D2 · `bugs_open/251` — the canonical, fleet-wide. Fix at the platform, or accept?

Every assembled homepage declares `https://<domain>/index.html` as canonical instead
of the bare `/`. **Measured: 9 of 10 live fleet homepages already do this**; 23 sites
are on that path. Long-standing platform behaviour that Track A *surfaced*, not
caused. Cause is one line —
`rerender_single_page_action.go:1074`, `canonical := "https://" + page.Domain + page.URL`,
no directory-index normalisation.

Not a lane fix: it is shared render-path Go, so it is **council-gate scope** and
architecture-adjacent (the guarantee "an assembled page declares its own preferred
URL" changes for every site at once). `injectPageJSONLD` is documented in-source as
byte-identical in its URL construction and must move with it.

> ### ⏸ SCHEDULED, not fixed. Filed as `bugs_open/251` with cause, blast radius and
> three costed fix candidates. **`bugs_open/252` (D4) depends on it** — `og:url` must
> agree with the canonical, so fixing 252 first would reproduce the `/index.html`
> error into `og:url` as well. **Do 251 before 252.**

### D3 · `bugs_open/250` — the sibling lane's copy of the broken rollback

`loancalculator_couk/decompose/load_decomposition.py` still carries **both** backup
defects, and that lane's backup table already holds **one poisoned rollback**
(28 rows over 27 pages) plus a 27-vs-28 column count that makes its backup
inoperative. The LMC half is fixed and proven. It is another lane's tool, so:
port it, or hand it to that lane's owner?

> ### ✅ RULED: port it. DONE 2026-08-11 — and the exposure was **63 rows, not 1**.
> Ported into `load_decomposition.py`; table repaired (28/27 → **27/27**, column
> parity 28 == live, stray preserved to `page_components_bak_strays_20260811_loancalc`,
> `DO`/`RAISE` verify block). Proved without touching a live page: both guards run
> read-only as `SELECT count(*)` — **OLD per-ROW guard would sweep in 63 live rows at
> the next `--apply`; NEW per-PAGE guard inserts 0.** That lane is far more decomposed
> than LMC was, so the single stray in its table was the residue of one earlier run,
> **not the size of the exposure**. `bugs_open/250` understated it and is corrected.
>
> **Still owed:** a full apply → restore → re-apply round trip on a live loancalculator
> page. The probes cover the mechanism, not that lane's end-to-end transaction shape.

### D4 · The og:* / lang accepted loss — still accepted at this scale?

Per-page `og:title`/`og:description`/`og:url` and `lang="en-GB"` → `lang="en"` were
accepted on 2026-08-05 (`PLAN_2026-08-05` §6) when they applied to **2** pages. They
now apply to **19**, and to all **59** when both sites finish. Nothing a reader sees
on the page changes — it is the link-preview card and the language tag.

Restoring per-page `og:*` is **not** a lane fix: the shared `<head>` has nowhere to
put them, so assembly would have to splice them per page the way it already splices
`<title>` and `<meta description>` — platform work.

> ### ✅ RULED: escalate both. Filed as `bugs_open/252`, and the scale is larger than
> the lane figure: **503 assembled pages fleet-wide**, not 19.
> **A (og:)** has a ready-made seam — `spliceMetaDescription` already fills a *blank
> placeholder* left in the site head and removes it when unfilled. Copy that shape;
> a site with no placeholder then gets no tags, i.e. unchanged behaviour, opt-in per
> site, which per the 2026-07-29 ruling §1 keeps it a normal gate change, not an RFC.
> **B (lang)** is hardcoded in **four** places across three files, and `sites` has
> **no** language/locale/country/region column — so it is not "wire up the existing
> field". Cheapest option may be to stop emitting `<html lang>` from Go and let the
> per-site head carry it. **Do not fix B in one lane's chrome.**

### D5 · 40 dormant `page_rerender:detected` rows on this site

From `discovery`, each carrying a `spec.reason` — which selects **REBUILD** mode, not
assemble. Inert today because nothing promotes `detected`. But they now point at 17
pages that would **rebuild from components** rather than hit the old refusal, which is
a different exposure from the one they were filed under. Cancel them, or leave them?

> ### ✅ RULED: cancel. DONE 2026-08-11 — 40 rows cancelled in one transaction,
> narrowed on site + type + status + `source='discovery'` + `spec ? 'reason'`, with a
> `DO`/`RAISE` guard asserting **exactly 40** before the update and **0 detected rows
> remaining** after. Each carries a `cancelled_reason` in its spec saying why, so a
> later reader is not left guessing. Refile deliberately if a rebuild is ever wanted.

### D6 · Site plans — the thing decomposition still does not buy

`site_plans` is **0 rows** for both sites. Decomposition bought per-component editing
and genuine rebuildability; wholesale rebuild-from-plan needs a plan, and seeding one
is a separate decision with its own risks (see 10c §6: a seeded plan that omits the
tool slot is exactly what triggers D1's trap).

---

## 3. Also owed, no decision needed — just unstarted

- ~~**loancash.co.uk has no arithmetic oracle**, and two of its three tools hardcode
  dated FCA caps…~~ **STARTED AND THE PREMISE REFUTED, 2026-08-11.** Verified against
  the FCA Handbook, not the page: 0.8%/day = **CONC 5A.2.3R**, £15 default =
  **5A.2.14R**, 100% total = **5A.2.2R**, last amended **02/01/2015**, all three
  correctly implemented, arithmetic sound. The SDLT analogy misleads — SDLT moves with
  Budgets, this cap has not moved in eleven years. The *monitoring* gap survives and
  is real but not urgent. **The worry moves to `complaint-deadline-calculator.html`**,
  unchecked, running off limitation periods and the FOS six-month deadline — a
  different legal source that does move. Workstream:
  `docs024_key_docs_latest/loancash_couk_fca_validation/` (its PLAN argues against
  cloning `oracle.py` here, and says why).
- Fleet sweep for the `bugs_open/224` class (a guard that leaves a handler without
  writing the DOM) on `mortgagecalculator.co.uk` and `loancash.co.uk`.
- Six LMC pages are fence-eligible with no fence (class C — they want `invariants.py`,
  not arithmetic).

## 4. Commands (all verified today)

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
export DECOMP_WORK=<your own scratch dir>          # must be YOURS

# after ANY chassis roll — revalidate the mirror before trusting predicted/
python3 $LANE/deploy_pages.py --tag <tag> legal
diff <(curl -s -A Mozilla/5.0 https://loanandmortgagecalculator.co.uk/legal.html) \
     $DECOMP_WORK/predicted/legal.html

python3 $LANE/load_lmc.py --check --all            # prediction only
python3 $LANE/load_lmc.py --apply   <page>         # CHANGES A LIVE PAGE
python3 $LANE/load_lmc.py --restore <page>         # per-page rollback (PROVEN 08-11)
python3 $LANE/gate_component_bytes.py
python3 $LANE/verify_site.py
cd $LANE && python3 oracle.py --selftest-parse && python3 oracle.py --mutate expectation --tools simple
cd $LANE && python3 oracle.py                      # controls in the SAME session or it is not evidence
```

## 5. Traps found running Track A that are NOT in the older briefs

1. **`deploy_pages.py` files at `priority 90`; the selector is `priority ASC`**
   (`load_work_item_actions.go:681`), so it sorts **behind** the routine fleet
   batches at 80. `legal` completing in 5 minutes was an empty queue, not the work.
   ⚠ I then predicted an ~85-minute wait from that and **was wrong** — they started
   in 4 minutes. The ordering is read from code and stands; the throughput estimate
   on top of it was inferred from one sample. Don't repeat it.
2. **A poller that times out and then grades is indistinguishable from a real
   failure.** Mine fell through after 10 minutes and diffed a page whose rerender had
   not run; the "DIFFERS" output looked exactly like a broken decomposition. The tell
   was that the served page still had `og:5` / `lang="en-GB"` — i.e. it was the
   hand-built page. Gate on `pending == 0` *before* grading, never on a timeout.
3. **`visible()` does not strip HTML comments**, and the chrome components carry long
   authoring comments the hand-built pages never had. A before/after visible-text
   comparison reports all 17 pages DIFFER until you strip `<!--…-->`. Sixth time on
   this estate that a red from a same-day checker was the checker.
4. **`--check` is a DESTINATION guard, not a content guard** (10d §4.3, induced). But
   also: **`P-visible` cannot fire on a Track A-shaped page at all** — with no
   script-addressed ids the whole `#content` inner becomes one prose block by
   identity, so the assertion compares a value with itself. The real content
   guarantee is stronger and was proved directly: manifest prose == source `#content`
   inner, byte for byte, and those bytes are a substring of the live stored row.
5. **`verify_site.py`'s repo-byte check reads your LOCAL sites clone.** Mine was
   **355 commits behind** and reported a false "live bytes differ from repo". `git
   fetch && git merge --ff-only origin/master` first — and note the remote branch is
   **`master`**, not `main`; `origin/main` fails with `bad revision` and, if you pipe
   it into a `comm`, produces a clean empty pass against a 0-line file.

## 6. Read in this order if starting fresh

1. this file
2. `SUMMARY_2026-08-11_track_a_complete_….md` — the plain-prose read-out
3. `NOTES_…md`, the 2026-08-11 entries — full working, every induction, every misstep
4. `HANDOFF_2026-08-10c_continue_here.md` §2b, §6, §7 — still current for Track B
5. `bugs_open/250` and `bugs_open/251`
