# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-18)

**Supersedes `HANDOFF_2026-08-16b_continue_here.md`.** That file's §0 owner rulings and its §4
(the directory-URL question) still stand and are NOT repeated here — read it second, for those two
sections only; everything else in it is either done or restated below.

**Nothing is blocked and nothing is half-finished.** This lane is in a clean state. What follows
is (1) what is live, (2) the three decisions that are the owner's, (3) what is genuinely next, and
(4) the traps this arc paid for.

---

## 0. The one thing to know before you touch anything

**Two mechanisms that were dormant are now running, and both were switched on deliberately in the
last two days.** If something on this site or the fleet starts filing work you did not expect, it
is probably one of these, and neither is a fault:

1. **`site-discovery-rotation-completeness` is ENABLED** (owner, 2026-08-17 11:31Z). One site per
   hour, oldest-unchecked first, 7-day eligibility. It owns link integrity
   (`phantom_internal_links`, `dead_internal_link_live`) plus 40 other checks. Findings insert at
   `detected` and `detected-item-promoter` (live, 15-min cadence) promotes them into dispatched
   work — so this produces real repair traffic across the estate, by design.
   **Stop it with one statement:** `UPDATE scheduled_tasks SET enabled=false WHERE
   name='site-discovery-rotation-completeness';`
2. **The `stamp-duty` fence declares its 13 SDLT facts** (CLM-022, owner: *"seed it for real"*).
   13 `fact_drift_review` items exist on this site and are **supposed** to be there — the one-time,
   self-quieting burst. Do not "tidy" them; see §3.2.

---

## 1. Live state — measured, with the date each figure was taken

| artefact | state | measured |
|---|---|---|
| all deployed pages' `<title>` | **27/27 match `pages.title`** | 08-16 15:47Z |
| `/guides/mortgage-scorecard/index.html` | 200, "Where you stand before you apply", 526 words, no template leak | 08-16 |
| `/guides/lender-restrictions/index.html` | 200, "What might limit your options", 444 words | 08-16 |
| `/scorecard-simulator.html` | **still 404** — build refused by `bugs_open/260`; item `0c65f9fa` parked | 08-16 |
| `/tools/rate-forecaster/` (directory form) | **still 404** — target exists at `…/index.html`; deliberate, see §3.1 | 08-16 |
| `images/mortgagecalculatormono.xcf` | **GONE** — 404, removed from bucket AND deploy source | 08-17 |
| logo / favicon / og-card / hero / header roundel | all 200, unchanged | 08-16 |
| the 10 tool-page heroes | still paid-and-unconsumed — `bugs_open/114`'s to wire | 08-16 |
| `stamp-duty` fence | `doc_plans` `400657e0…`, 13 facts declared, `no_auto_fix` true | 08-17 |
| evidence register | 1 current row, relief cap **500000**, `pinned` carried, 0 test rows | 08-17 |

Site id `62b5978e-4271-4589-8e00-4baebfc0447c`. `sites.github_repo` empty → **B2 route** (§3.1).
Site unlocked. No work items armed by this lane.

---

## 2. What this arc actually did (08-16 → 08-17), one line each

- **The "30 stale titles" item was already done** and had been carried forward unmeasured for five
  days. Acting on it would have fired 30 rerenders to change nothing.
- **Found 8 live broken internal links**, 4 dead targets, 7 pages. Owner chose to fix by building
  the missing pages rather than re-pointing copy.
- **Built two guides through the framework** by arming two `needs_page` items parked since 07-31.
  Fixed 3 of the 8 links. **Dead links went 8 → 7, not 8 → 1** — see §3.3, it is a finding.
- **Contributed to `bugs_open/260`** (the template-leak class that blocks the third page) with a
  census isolating it from the 8 other issue types sharing its `error_code`.
- **Enabled the completeness rotation** after measuring its shape; proved the checks fire (87
  findings on the first real site, including 4 genuine live 404s).
- **Removed the `.xcf`** from bucket and deploy source, with four byte-identical copies retained.
- **Seeded the CLM-022 `facts` declaration and PROVED IT END TO END** — both arms, plus the
  self-quieting. §3.2.

---

## 3. The open items, in the order they matter

### 3.1 The directory-URL 404 — OWNER'S CALL, three lines of code, 36 zones

On the B2 route **every** directory-form URL 404s: `/guides/`, `/about/`, `/tools/repayment/`,
`/investor/` — while every `…/index.html` serves 200. The git route serves both
(`relojistas.com/noticias/` 200). Anyone who types, shortens or shares an address without the exact
ending gets a 404.

- **Cause and fix located precisely:** `scripts/cloudflare/worker.js:9-12` special-cases only the
  root; an `else if (path.endsWith('/')) path += 'index.html';` closes the class — for human,
  typed and inbound links, not just internal ones.
- **Why it is not a lane's call:** one worker fronts all 36 B2 zones, and
  `scripts/cloudflare/README.md` warns that a deploy without the two B2 bindings strips the
  worker's credentials and takes every site down. Owner + council.
- **Already in `LANDMINES.md`** (search *"A `/section/` URL 404s on every B2-hosted site"*), with
  an addendum this lane added: the platform's own `NormalizePagePath` collapses `/x/` and
  `/x/index.html`, so five DB-derived mechanisms all wave the dead form through. **Do not "fix" it
  by making that helper route-aware.**

### 3.2 CLM-022 — PROVEN, and the 13 items are supposed to be there

Both arms induced on 2026-08-17. The three runs are the proof and no single run shows it:

| run | `results[0].fact_drift` | kind |
|---|---|---|
| dry, pre-baseline | **13** | `unreconciled_declaration` |
| REAL sweep (`dry_run:false`, site-scoped) | **13** | all `outcome: inserted` — the baselines |
| dry, post-baseline, one fact moved | **1** | **`value_drift`**, `old 500000 → new 550000` |

**13 → 13 → 1 is the healthy sequence.** The tool's own question ("does it compute from the
register?") was then answered with `verify_criteria.py`: **4 agree, 0 mismatch, all from REGISTER**,
and the induced red (`--mutate sdlt-ftb-relief-cap=625000`) moves exactly one assertion by exactly
the perturbation.

**The 13 `fact_drift_review` items are left OPEN deliberately.** Closing them is provably safe —
`factDriftLastItemQuery` (`refresh_evidence_fact_drift.go:275-278`) has **no status filter**, so a
resolved item still serves as the baseline. What stopped this lane is that the type is handler-less
with no documented resolution path, and `bugs_open/033` says that queue has no working surface, so
closing would be inventing semantics for tidiness. **Owner's call; the evidence stands either way.**

### 3.3 `bugs_open/260` blocks the third page, and the queue is SELF-FUELLING here

`/scorecard-simulator.html` cannot build: its content comes back with raw Go template directives
(`{{if .eyebrow}}…{{end}}`) and `validate_content` correctly refuses it. 260 is filed, root cause
proven, actively owned elsewhere — **contribute, do not compete, do not re-fire item `0c65f9fa`**
(it is the live end-to-end test case for whoever fixes it; correlation
`8d4467e0-3c1b-4949-88fe-79e4566282a5`).

**The finding worth carrying:** the two pages the framework just wrote **each link to
`/scorecard-simulator.html` themselves**, because the site's `design_intent` names it. Live
instances of that dead href went **4 → 6** while three others were fixed. Every page the framework
writes here adds another link to the one page it cannot build. That is a measured argument for
prioritising 260 and it belongs in front of whoever ranks it.

### 3.4 Unchanged from the previous handoff

Router fleet assignment (IMG-071, and its rule: fresh discovery pass FIRST, then route) · card
icons (114-class, needs a component-field change) · the fleet **design** rotation, which remains
`enabled=false` and remains the owner's separate call — **do not flip it on the strength of the
completeness decision, they were paused for different reasons.**

---

## 4. Traps this arc paid for — read before you repeat any of this

- **A carried-forward "unchanged" is a claim about the past.** Re-measure any inherited status
  before acting on it, especially on a site with a live improvement loop.
- **A post-deploy 404 may be the CDN's cached negative** (`max-age=300` on the worker's 404 arm).
  Probe twice, second one cache-busted.
- **A single 404 in a fast scan is not evidence.** A 28-URL burst reported a healthy page as dead;
  it is 200 on every repeat.
- **Grep prior art by FOOTPRINT SYMBOL, never symptom phrasing.** This lane filed a "new
  fleet-wide class" that LANDMINES had carried for weeks, because it grepped *"trailing slash"* and
  the entry says *"a `/section/` URL"*. (`WRONG_CALLS.md` 2026-08-16.)
- **Rewriting a shared ledger needs a PINNED base.** Replacing one LANDMINES entry with an
  `s[:start]` truncation deleted two other lanes' entries committed in the intervening 20 minutes.
  Recovered from a pinned sha; verify with `--numstat` and a single-hunk check.
- **`fact_drift` is per-site NESTED** — `refresh_result->'results'->N->'fact_drift'`. The top-level
  key does not exist and `total_drifted` counts CITATION drift; together they say "it never fired"
  about a run that fired 13 times. This lane reported exactly that out loud before dumping the
  payload (`WRONG_CALLS.md` 2026-08-17, and a LANDMINES entry).
- **A pinned row id expires** — `mutate.sql` named a `site_specs` row that this lane's own real
  sweep superseded two commands later; the re-run aborted on `idx_site_specs_current` and wrote
  nothing. Resolve the current row dynamically.
- **Verify a build at the DIGEST, before the run, not after.** `docker inspect <img> --format
  '{{json .RepoDigests}}'` vs the pod's `imageID`; and the image label
  `org.opencontainers.image.revision` gives the build commit and **outlives the `build provenance`
  log line**, which had already scrolled on 68-minute-old pods.
- **A dry run's `writer_block_action: "regenerated"` is a PLAN, not pending work.** The real run
  said `unchanged` with an identical `writer_block` md5.

---

## 5. Files of record

- `NOTES_mortgagecalculator_couk.md` — `## 2026-08-16 (afternoon…)`, `## 2026-08-17 (morning)`,
  `## 2026-08-17 (afternoon)`, `## 2026-08-17 (evening)`.
- `README_where_we_are.md` — `2026-08-16 (Sunday afternoon)`, `2026-08-17 (Monday morning)`.
- `SUMMARY_2026-08-18_two_dormant_mechanisms_switched_on_and_proven.md` (the milestone read-out).
- `bugs_open/260` §10 (this lane's contribution) · `LANDMINES.md` (two entries) ·
  `WRONG_CALLS.md` 08-16 and 08-17 ·
  `register_guards_code_phase_b/CONTRIB_REPLY_2026-08-17b_…md` (seeded + both addenda).
- Reusable scripts: `scratchpad/{dryrun_safe,realrun_safe,induce}.sh` + `{mutate,restore}.sql`
  — **scratchpad is session-scoped; copy them into the lane dir if you need them again.**
- Cross-lane: `webdesign_uk_build_service/NOTE_2026-08-18_from_mcalc_lane_…md` (their
  blocker-detail provenance is a month out; the table also prunes).
