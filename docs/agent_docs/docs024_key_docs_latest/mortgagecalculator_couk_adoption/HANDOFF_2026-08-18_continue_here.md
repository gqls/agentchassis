# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-18)

**Supersedes `HANDOFF_2026-08-16b_continue_here.md`.** Only that file's **§0 owner rulings** still
need reading — its §4 (the directory-URL question) is now **fixed and shipped**, see §3.1.

**Nothing is blocked and nothing is half-finished.** This lane is in a clean state. What follows
is (1) what is live, (2) what this arc did, (3) the open items — **one platform bug owned
elsewhere, one queue decision, and the parked backlog** — and (4) the traps this arc paid for.
**Updated 2026-08-18 evening: the directory-URL fix is LIVE (§3.1).**

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
| `/tools/rate-forecaster/` (directory form) | **200** — fixed by DGH-012, no copy touched | 08-18 |
| `images/mortgagecalculatormono.xcf` | **GONE** — 404, removed from bucket AND deploy source | 08-17 |
| logo / favicon / og-card / hero / header roundel | all 200, unchanged | 08-16 |
| the 10 tool-page heroes | still paid-and-unconsumed — `bugs_open/114`'s to wire | 08-16 |
| `stamp-duty` fence | `doc_plans` `400657e0…`, 13 facts declared, `no_auto_fix` true | 08-17 |
| evidence register | 1 current row, relief cap **500000**, `pinned` carried, 0 test rows | 08-17 |

Site id `62b5978e-4271-4589-8e00-4baebfc0447c`. `sites.github_repo` empty → **B2 route** — which since
DGH-012 no longer means directory URLs 404 (§3.1).
Site unlocked. No work items armed by this lane.

---

## 2. What this arc actually did (08-16 → 08-18), one line each

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
- **Shipped the directory-URL fix to the Cloudflare worker (DGH-012)** after the owner approved
  it — `/guides/` was a 404 on every B2-hosted site and is now 200, fleet-wide. §3.1.

---

## 3. The open items, in the order they matter

### 3.1 ~~The directory-URL 404~~ **FIXED AND LIVE 2026-08-18 12:23Z — register DGH-012**

Owner approved it; it is deployed and verified. `scripts/cloudflare/worker.js:9-27` now appends
`index.html` to any path ending in `/`.

**Verified at the artefact, same probe both sides:** `mortgagecalculator.co.uk/guides/` and
`/tools/repayment/` **404→200**; `gaswholesalers.com/tools/supplier-comparison-calculator/` and
`leopardessconsulting.co.uk/tools/automation-savings-estimator/` **404→200**. Unchanged: every
`…/index.html`, the site roots, `/worker-health`, and a genuine miss still **404 with the site's
own 404 page and no bucket internals**. Git-route control (`relojistas.com/noticias/`) unaffected.
`/guides/` and `/guides/index.html` return the same `<title>`, so it is the real page, not a
soft-404. The repo copy was re-exported and diffed after the PUT: **deployed == repo**.

**It closed this site's last non-260 broken link without touching a word of copy** —
`/tools/rate-forecaster/` now resolves. A full re-audit of all 30 internal links leaves exactly
**one** dead target: `/scorecard-simulator.html`, which is §3.3's bug.

⚠ **Two deploy traps, now in `scripts/cloudflare/README.md` and DGH-012, because either could
cost an outage or a false alarm:**
- **The PUT response returns `result.bindings: []` on a completely successful deploy.** That is
  the API not echoing them, and it is indistinguishable from the credential-stripping outage the
  README warns about. Confirm from the `/settings` endpoint and a live fetch, never from the PUT.
- **`node --check` PASSES a syntactically broken `worker.js`.** ESM syntax in a `.js` file makes
  the check a no-op — a copy with one `)` removed exited **0**; as `.mjs` it correctly failed.
  Check as `.mjs` **and prove the checker fails on a broken copy in the same session.**
- Build the PUT metadata from `~/.cloudflare/portfolio-sites-router.settings.json`, not by hand:
  the README's minimal recipe silently resets `observability` and `compatibility_flags`.

**Not council-reviewed, and that is not an omission:** the gate refuses paths outside
`platform/`/`internal/`/`pkg/` client-side, so `scripts/cloudflare/` cannot be submitted. The
ordering-exemption's condition (2) was met instead — registered in the same commit as the ship,
with its landmines and one open review question (canonicals emit `…/index.html` and now both
forms serve; nobody has ruled on the duplicate-content question).

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
  ⚠ **Written before the worker fix shipped, so its "where we're going" is already overtaken** —
  it names the trailing-slash question as open and the rotation as still running; both are done.
  Left unedited (the series is the record); the corrections are in NOTES 2026-08-18.
- `bugs_open/260` §10 (this lane's contribution) · concept register **DGH-012** (+ its index row) ·
  `scripts/cloudflare/{worker.js,README.md}` · `LANDMINES.md` (two entries, both now CORRECTED in
  place because DGH-012 killed the behaviour they described) ·
  `WRONG_CALLS.md` 08-16 and 08-17 ·
  `register_guards_code_phase_b/CONTRIB_REPLY_2026-08-17b_…md` (seeded + both addenda).
- Reusable scripts: `scratchpad/{dryrun_safe,realrun_safe,induce}.sh` + `{mutate,restore}.sql`
  — **scratchpad is session-scoped; copy them into the lane dir if you need them again.**
- Cross-lane: `webdesign_uk_build_service/NOTE_2026-08-18_from_mcalc_lane_…md` (their
  blocker-detail provenance is a month out; the table also prunes).
