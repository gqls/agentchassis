# NOTES — web admin console

Running record, append-only, newest at the bottom. Missteps included on purpose.

---

## 2026-08-23 — a build-break I attributed wrongly, twice, in opposite directions

While committing the `/c/` prefetch guard, `go build ./internal/core-manager/...` failed with
three `not enough arguments in call to emitRequiredFieldsMissing` errors in
`render_site_components_action.go` and `section_editor_actions.go`. I put a line in the commit
message: *"do NOT build/roll core-manager until HEAD compiles again."*

**That line is now FALSE, and a stale blocker is worse than no note** — this estate has already
been bitten by an "inert until the roll" line that left a detector switched off for nine days
after its blocker cleared (`LANDMINES.md`). So, plainly: **the build is clean. core-manager can
be built and rolled.**

### What actually happened, and both of my readings were wrong

1. **First I said it was "another session's in-flight work".** Reasonable, but unchecked.
2. **Then I checked and said the opposite** — the two failing callers were not dirty and neither
   was the definition file, which pointed at committed breakage on HEAD. Also wrong.
3. **The decisive check settled it: `git archive HEAD | tar -x` into a temp dir and build there.**
   HEAD compiled fine. So the fault was in the working tree after all, and reading (1) was
   right — but I only knew that after building a tree with no working-tree changes in it, which
   is the only way to separate the two on a shared checkout.

The owner is the **`bugs_open/342`** lane (an absent required field rendering empty and silent
at 13 of 15 render call sites). It was mid-refactor: `emitRequiredFieldsMissing` had gained a
`pageContext` parameter in `work_items_common.go` while its callers had not caught up in that
session's tree. **It fixed itself within the hour** — `eb918bd58` and the commits after it — so
there was nothing to chase and nobody to nudge.

### The transferable bit

**On a shared working tree, "the build is broken" is not a fact about the repository until you
have built a tree with no working-tree changes in it.** `git status` on the failing files is not
enough: the file whose *signature* moved can be committed while the callers that need updating
sit in someone else's uncommitted edit, or the reverse, and both look identical from a status
line. The one-liner:

```bash
T=$(mktemp -d); git archive HEAD | tar -x -C "$T"; (cd "$T" && go build ./...) ; rm -rf "$T"
```

**And a transient break needs a re-check before it goes in a commit message**, because the
message cannot be amended (forward-only) and the claim outlives the condition by days.

---

## 2026-08-24 — sent to correspond, not to build; and a pattern language that ate my evidence

The owner was asked (by me, from a two-day-stale context) whether to expose the console as
`admin.apis.uk`, keep it VPN-only, or gate it. He answered neither: *"Please correspond with the
webdesign.uk live webdesign thread that has built a web facing console."* **The console was
already live.** My session had been handed `PLAN_2026-08-22` and had not re-read the lane before
forming a question — the lane had produced six commits and a new handoff in between.

**The check that would have caught it before I asked: `ls -lt` the lane directory.** It takes
one second and would have shown `HANDOFF_2026-08-24_continue_here.md` written 20 minutes
earlier. I ran it only after the owner's answer. On this tree a plan handed to you names the
state it was written in, not the state you are in.

### Re-measured three of the 08-24 handoff's own falsifiers (2026-08-24 ~11:15Z)

- `https://admin.apis.uk/` → **302** to `billowing-smoke-5ed4.cloudflareaccess.com`; the meta
  JWT decodes to `auth_status: NONE`, `hostname: admin.apis.uk`. Live and gated. CONFIRMED.
- ~~`https://www.apis.uk/` → **301**, `location: https://apis.uk/`. §3's redirect rule **is
  applied** — the handoff still lists it as owner-pending. One falsifier closed.~~
  > **CORRECTED 2026-08-24, same session, by the `apis_uk_bees_homepage` lane within the hour.**
  > The 301 is real; **the attribution was invented.** It is served by the
  > `portfolio-sites-router` **Worker** — `scripts/cloudflare/worker.js:23`,
  > `if (hostname.startsWith('www.'))` → `Response.redirect(url, 301)` — reached because the
  > `apis.uk` zone carries a `www.apis.uk/*` **worker route** (LANDMINES 2026-08-23: the zone has
  > `apis.uk/*` and `www.apis.uk/*`, **2** routes as of 2026-08-22). **A worker route intercepts
  > before DNS is consulted**, so §3's dashboard Redirect Rule is not what fires, and this
  > measurement says **nothing** about whether §3 was applied. That falsifier is NOT closed.
  >
  > **The mechanism of my error, because it is not "stale data" and will recur:** I observed the
  > outcome §3 predicted, found §3 listed as pending, and joined them — without asking what else
  > could produce a `www`→apex 301. Something else did, and it predates the handoff. This is the
  > memory index's *"a believable cause explaining an observation is when to doubt the
  > instrument"* — a pending instruction that would explain what you see is the **weakest**
  > moment to credit it, because it is exactly when you stop looking.
  >
  > **The check: a redirect is not evidence of the rule you have in mind unless you can name the
  > thing serving it.** For a Cloudflare zone that means enumerating worker routes before
  > crediting a DNS or Rules change — routes win, silently.
- `https://links.webdesign.uk/c/x` → **could not resolve host**. §2's box steps are **not**
  applied, so `/c/` has not moved off the shopfront and the parking-page-rule landmine is still
  the only thing holding it. Still owner-pending, correctly listed.

### The misstep worth the entry: `LIKE '%__step_error%'` is not a literal search

Sizing the `bugs_open/099` landmine for the build-steps screen, I counted COMPLETED rows whose
`collected_data` mentions `__step_error` with `LIKE '%__step_error%'` and got **315**.

**In SQL `LIKE`, `_` is a single-character wildcard.** The pattern actually asked for "any two
characters followed by `step_error`". The honest count via `strpos` is **176**. The key whose
distinguishing feature is a double underscore is exactly the key `LIKE` cannot be trusted with —
the wildcards sit precisely where the evidence is.

Then I compounded it: I assumed the 176 − 67 gap between "literal anywhere" and "top-level key"
was **nesting**, and wrote that the top-level test misses 109 real errors. It does not. One query
extracting 320 characters around the literal showed the gap is **workflow configuration naming
the field** — `"note_body_field": "__step_error.message"` inside an `append_doc_note` step. The
top-level jsonb test is exact: 67 real `"__step_error":` keys, all top-level, 67 = 67.

So I twice reported a fabricated defect in someone else's correct design. **Both errors share one
cause: I believed a count produced by a pattern without reading a single row it matched.** That
is now the check — read one matching row before quoting any pattern-derived count. Recorded in
`WRONG_CALLS.md`; the finding itself is in `PLAN_2026-08-24_build_steps_screen.md` §6b, written
up in the direction that survived.

### What that measuring pass did produce, all in `PLAN_2026-08-24_build_steps_screen.md` §6

The plan's §2 states `orchestration_states` has no site column. It has one, with **three**
indexes on it, populated on 2,136 of 4,410 rows — and of the two JSON paths §2 proposed instead,
one matches **zero** rows. That is its only backend change, made smaller and safer. §6d is the
one that may reshape the screen: `execution_path` is empty on **100%** of rows, so there is no
recorded step sequence, and a site's `site_id`-tagged orchestrations are mostly its periodic
sweeps rather than its build — the thing the owner calls a build is a `site_work_items` chain
(`082_submit_domain_unified.sh`). I did **not** act on that; it is this lane's call.

---

## 2026-08-24 (later) — the build-steps screen is BUILT and COMMITTED; verdicts, falsifiers, and one near-miss

Session picked up `HANDOFF_2026-08-24_continue_here.md` and built §4. State changes, evidence inline:

- **Council verdict on the prefetch guard (`6b1726ab…`): APPROVED, round 1**, 2026-08-23 12:07Z,
  one advisory objection, none high-severity (doc_notes, categories ? 'council-gate'). Commit
  `0e9cb31ee` carries `Council-Submitted:`, so 098 credits it automatically. §5.2 CLOSED.
- **Falsifiers re-run:** `links.webdesign.uk/c/x` → no connection (curl 000) — §2 still
  owner-pending. `customer_access_tokens` = **0 rows**. `www.apis.uk` → 301 to apex — and I
  **nearly re-made the exact attribution error this file documents two entries up**: my first
  in-session note said "§3 is APPLIED". It is served by the `portfolio-sites-router` Worker;
  the 301 says nothing about the dashboard Redirect Rule. Caught by re-reading these NOTES
  before writing the summary — which is what the read-order exists for.
- **Owner FYI mid-session: a fresh chassis build/deploy is in flight.** Consequences honoured:
  council submission timed before it (dispatch may still race a chassis restart's ~300s dead
  window — verify the orchestration row by payload before assuming it ran), and the transient
  `psql … unexpected EOF` during `landmines-verify-dispatch.sh` (17:23Z) was retried clean.

### What was built (commits `e6350e74b` backend, `b3fbfdd02` frontend)

**Backend** (`internal/core-manager/admin/`, council `Council-Submitted:
45b3c93f-7937-474d-8234-31c39bab033b`, read the verdict when it lands):
1. `GET /admin/workflows` gains `site_id` filter (validated UUID; plain `WHERE site_id = $n`
   against the indexed column, per PLAN §6a — NOT the JSON extraction §2 proposed) and per-row
   `has_step_error` = `collected_data ? '__step_error'` (exact per §6b: 67 = 67), plus
   `site_id` in each row.
2. `updateWorkflowStatus` table name fixed `orchestrator_state` → `orchestration_states`
   (ADM-002 B2; register entry updated with the commit and the fixed-AND-live caveat). A
   0-rows UPDATE now returns `sql.ErrNoRows` instead of silent success.
3. `HandleUpdateSiteSpec` evidence_base guard (§6h hazard): parse incoming through
   `datahelpers.ParseEvidenceBase` before writing — unparseable → 400; parses-nil over a
   register that parses non-nil → 409 `EMPTY_EVIDENCE_BASE` unless `confirm_empty:true`;
   response returns stored counts (facts/banned_claims/allowed_entities/regulated) and a
   `superseded` flag (false = the save CREATED an aspect — the no-allow-list typo tell).
4. `spec_update_guard_test.go` — six sqlmock tests. **Worth recording: the wrong-shape test
   failed on first run (500, "Begin was not expected") because my `goodRegister` fixture used
   a string `value` where `EvidenceFact.Value` is `*float64` — the CURRENT register failed to
   parse, so the guard correctly stood down. The test design (arm ONLY the guard's SELECT)
   caught a non-firing guard by construction — an accidental live mutation test.**

**Frontend** (`frontends/admin-dashboard/src/App.tsx`, vite build proven in the Docker
builder stage): `BuildsView` — per-site stage timeline over `site_work_items` in the 082
cascade order with durations (PLAN §6d shape: the build IS the work-item chain; explicit
`BUILD_STAGES` vocabulary; if the 1000-row window is truncated, per-stage-type refetch so the
oldest rows — the build — cannot silently fall out), other-activity rollup, divergence-count
warning, and an orchestrations drill-down whose detail panel surfaces `__step_error` entries
in red above whatever `status` says, marks steps as "reconstructed from outputs", and gates
resume/terminate behind confirms naming the correlation id. Plus §6g: `⚠ overwritten ×N`
badges on components with `page_divergence_overwritten` history (keyed
`page_id|slot_name|position`, slot fallback), red + "unlocked" when the next rebuild would
eat a hand edit again. Plus §6h SPA half: ENFORCED/advisory chips, counts echoed on
evidence_base saves ("REGISTER IS EMPTY: claims checking is now OFF" on 0/0), the 409 →
explicit confirm → `confirm_empty` resend flow, and a prohibition-worded-save nudge toward
`banned_claims`. One self-caught bug: the Save button passed the click EVENT as
`confirmEmpty` (truthy) — fixed to `() => handleSaveSpec()` plus `=== true` coercion.

### Deploy ordering — DO NOT deploy the dashboard image first

**New SPA + old core-manager is the misleading combo**: old gin ignores the unknown
`site_id` query param, so BuildsView would render the WHOLE FLEET's workflows labelled as
the site's, with no `has_step_error`; and `confirm_empty` would be ignored while the counts
read `undefined`. Old SPA + new backend is harmless. So: **core-manager must roll a build
carrying `e6350e74b` BEFORE `make admin-dashboard`** (build/push/deploy — dashboard builds
from the working tree, so make sure the tree's App.tsx is at/after `b3fbfdd02` when
building). Check with the provenance stamp:
`kubectl -n ai-persona-system logs -l app=core-manager --tail=300 | grep -m1 'build provenance'`
then `git merge-base --is-ancestor e6350e74b <stamp>`.

### Also this session

- **LANDMINES entry appended** (uncommitted — the file carries another lane's WIP entries;
  whoever commits takes both, append-only makes that safe): the evidence_base wrong-shape
  silent-off trap, footprinted on `site_specs`/`ParseEvidenceBase`/SQL seeds, with the
  read-back-counts check. Synced + verifier dispatched (2 runs published; `landmines-sync.py
  --check` clean).
- **ADM-002 register entry**: B2 bullet updated (fix committed, open-until-roll).
- HEAD verified building after the backend commit (`verify-head-builds.sh` OK at `e4d20d97a`
  — HEAD had already moved past my commits; shared tree as usual).

### 2026-08-24 ~16:35Z — verdict on `45b3c93f`: APPROVED round 1; the three advisory objections answered

**APPROVED, 3 advisory objections, none high-severity, 8 seats abstained** (report:
`diagnosis_artifacts` kind=`council_report`, correlation `45b3c93f…`). Objections, each
re-checked against the code/DB rather than argued with:

1. **editquality (medium): does the INSERT write raw bytes or a re-marshalled struct?**
   (A landmine exists on round-tripping ParseEvidenceBase — the typed struct drops fields
   it does not model.) **Answered at the code: raw bytes.** `site_admin_handlers.go:282`
   inserts `body.Data` (the original `json.RawMessage`); the parsed struct is used ONLY for
   validation and counts. The landmine is not reproduced. The seat was right that the plan
   never said so — this note is the statement.
2. **bug_historian (medium): `WriteSiteSpecAction` (platform path, higher volume) has no
   shape guard.** **Checked, and the hazard is REAL but NARROWER there** `[MEASURED
   2026-08-24]`: the action deep-MERGES the partial over the current doc, so a wrong-shape
   KEY (`bannedClaims`) is additive — the existing `banned_claims` survives and the register
   still parses non-nil. Whole-register silent-disarm needs the admin door's REPLACE
   semantics, which is now guarded. BUT `siteSpecDeepMerge` overwrites non-map values
   wholesale (`site_spec_actions.go:554`), so a partial carrying `"banned_claims": []`
   DOES empty the array — and the highest-volume evidence_base writer is `source='scheduled'`
   (214 of 319 all-history rows; automated). **Follow-up owed, own council round:** decide
   whether a merge whose RESULT shrinks banned_claims below current needs a flag, remembering
   the scheduled refresher may legitimately shrink registers — census the shrink history
   before designing the guard (memory: census-the-write-history rule).
3. **guardian (medium): terminate is DB-label-only — it does not interrupt a running step.**
   True and pre-existing (it was the endpoint's semantics before the table-name fix made it
   reachable). **Closed the honest half in `1a8db99f9`:** the SPA terminate confirm now says
   exactly that. The status written is `FAILED`, the same literal the platform's own
   updateWorkflowStatus path has always used. Whether a real cancel/interrupt mechanism is
   wanted is an owner-scoped question, not smuggled in here.
4. Lows recorded, not acted on: `aspect` still has no allow-list (named by me in the plan;
   architecture seat wants a follow-up ticket, and agrees fixing it here would be
   scope-creep); constitution notes ParseEvidenceBase itself could distinguish
   "genuinely empty" from "wrong shape" for ALL callers — deferred, stated.
5. **debug_historian's gap, now an owed verification step:** after core-manager rolls a
   build carrying `e6350e74b`, smoke-test terminate against a real non-terminal correlation
   (expect 200 + row status FAILED, not 500) — sqlmock proves the SQL shape, not the table.

### 2026-08-24 evening — owner ruled the residual, the route census, and a token expiry

- **Owner RULED the mail-scanner residual: second click required** ("We can't have email
  scanners clicking the accept button so we'll need a separate page"). Decision doc:
  `../webdesign_uk_build_service/DECISION_2026-08-24_confirmation_needs_a_second_click.md`.
  Blocks the first delivery email; GET /c/ becomes render-only, confirm moves to POST.
- **Deploy check: core-manager does NOT carry `e6350e74b`** — both pods stamp `70fd163c2`
  (15:37Z), ancestor check fails. Dashboard deploy stays blocked.
- **Route census** (owner challenged "second public cluster route" — right to): portfolio
  domains (noted.co.uk, idea.uk, robot-hands.com — live, apis.uk, ~39 zones) are
  Cloudflare-fronted static sites (B2 bucket via the Worker / git route) — NOT cluster
  routes. Cluster-reaching paths: admin.apis.uk (Access-gated, live);
  webdesign.uk/{c/,stripe/webhook} — **both measured 302-parked to webdesign.co.uk right
  now**, so today zero ungated cluster paths are reachable; links.webdesign.uk/c/ will be
  the first. ⚠ flagged to webdesign lane: Stripe webhook events would 302-bounce today.
  One instrument note: `?debug=1` on noted.co.uk returned site HTML, not the Worker's
  debug JSON — the deployed Worker may differ from the repo copy (live-and-committed are
  independent facts), so "Worker or git route" is stated, not which.
- **kubectl token expired ~16:50Z** (fleet-wide Unauthorized, the 3-day cycle; owner
  refreshes). All figures above predate it.

### 2026-08-24 ~20:55 — edge rate limit DEPLOYED by the owner

The Cloudflare rate-limiting rule for `links.webdesign.uk` is created and deployed
(owner, at the dashboard, via the Edit-expression route — the free-plan builder has no
Hostname field; handoff §3.7 carries the corrected recipe). `(http.host eq
"links.webdesign.uk")`, IP characteristic, 10-per-10s → Block. **Its 429 verify CANNOT
run yet** — `links.webdesign.uk` still does not resolve (curl exit 6, re-checked
immediately after), so the rule sits armed and matching nothing until the owner's box
steps + the DNS record land (morning handoff §2). Run the 40-curl loop from handoff
§3.7 as part of the §2 verify when that happens.

### 2026-08-24 ~21:20 — links.webdesign.uk LIVE, all arms verified from this machine

Owner applied the runbook (`../webdesign_uk_build_service/RUNBOOK_links_host_box_steps.md`)
and reported clean; re-verified here rather than trusted: 404 (/other), 404 (/c/x — the
hardening), 200 (43-char token-shape — proves the full path to core-manager:8088), 302→
Access (admin.apis.uk healthy post-restart), 302→webdesign.co.uk (apex still parked),
and the 40-loop returned 8×404 then 32×429 — the edge rate-limit rule blocks at the
threshold. Morning §2 CLOSED. Boundary-review condition MET; council round owed, blocked
on the kubectl token refresh. Second public-facing milestone of the day after the Builds
screen; SUMMARY_2026-08-24 already covers both.

### 2026-08-24 ~22:15 — the roll shipped BOTH halves; everything this lane had in flight is LIVE

Owner's fresh roll (~21:00, pods 21:03): core-manager `635f2d32f` — `git merge-base
--is-ancestor e6350e74b` PASSES, backend live. Dashboard `v1.0.1336` built from this
shared tree AFTER my commits — proven at the ARTEFACT: fetched `assets/index-Bqjp4Gs8.js`
from the pod (127.0.0.1:8080 — `localhost` resolves to ::1 in that busybox and refuses;
worth remembering) and grepped all five markers incl. `1a8db99f9`'s "does NOT interrupt".
So the deploy-ordering constraint dissolved — another session's roll swept both halves in
the correct order. Deploy pair CLOSED without this lane deploying anything. kubectl token
refreshed by the owner. Falsifiers re-run: tokens 0, handed_over 0/0. RFC_054 filed (the
boundary review the architecture seat's condition owed — census reused from tonight, not
re-measured). New consolidated handoff: HANDOFF_2026-08-24c_continue_here.md. Remaining
code task: the second-click page (DECISION doc has the spec; POST same-path pinned).
