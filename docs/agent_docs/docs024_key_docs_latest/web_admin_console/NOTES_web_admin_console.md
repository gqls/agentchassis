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

### 2026-08-25 morning — re-verified yesterday's deploy, then BUILT the second-click page

**First: yesterday's proof had expired and I re-ran it rather than repeat it.** A roll landed
~09:25. Dashboard is `v1.0.1337`; served bundle is STILL `assets/index-Bqjp4Gs8.js` — the same
hash as 08-24, because the name is content-derived and the SPA did not change in that roll.
⚠ **So neither direction of that reading is evidence on its own**: the tag moved without the
bundle changing, and an unchanged bundle name does not mean a deploy failed. All five markers
grepped present in it, with `ZZZ_control_must_be_absent` returning 0 in the same command.
core-manager stamps `4c996e1b5`; `git merge-base --is-ancestor e6350e74b 4c996e1b5` passes and
the REVERSED test fails, as it must.

**Then item 2 of the handoff: `/c/<token>` splits by method.** GET → `HandleConfirmPage`
(renders one button, **no database access on any arm**), POST → the existing
`HandleConfirmTransfer`, same path. Owner settled the two open forks in-session (2026-08-25):
**no lookup at all on GET** (rather than a read-only peek — it keeps the change inside
`internal/`, makes "the link click mutates nothing" structural, and denies a guessed token a
free validity oracle), and **the page names nothing** (preserving `delivery_deps.go:36-38`).
No vhost change ships: `location ~ "^/c/[A-Za-z0-9_-]{20,128}$"` already carries
`limit_except GET POST`, and the regex is anchored, which is exactly why the POST had to be
same-path. Council submitted, `SUBMISSION_CORR=ea99befa-ec62-4f61-b052-c3af3d003d55`.

> **MISSTEP, and it is the reason to keep running mutations rather than trusting a green
> suite.** My first version of the GET test asserted only what the page said: "the body does
> not contain *that is recorded*". That follows this file's own rule — assert the EFFECT,
> never the absence of a call — and **it was wrong here.** The mutation I ran to prove the
> test (`HandleConfirmPage` calls `ConfirmTransfer`, then renders the button page anyway)
> **PASSED**: the database would have been mutated and the rendered page is byte-identical,
> so nothing in the response witnesses it. The rule does not apply to this property, because
> *"the link click reaches no database"* IS an absence and no response can witness one. The
> direct assertion (`len(f.gotTokens) != 0`) went in and the same mutation now fails.
> **The check: when the property under test is an absence, ask what in the response could
> possibly differ — if nothing can, the effect-based assertion is vacuous no matter how
> correct the rule that prescribed it.** A second mutation (routing GET back to the confirm
> handler) failed correctly from the start. Both restored; clean suite green;
> `verify-head-builds.sh --test --with` OK against HEAD `021843119`.

**Still owed before the first delivery email:** this must be LIVE, not merely committed —
it rides the next core-manager roll and gets proven at the pod's provenance stamp.

### 2026-08-25 12:36Z — council APPROVED r1, and what each objection cost

`ea99befa-ec62-4f61-b052-c3af3d003d55` — **all reviewers approve**, 10 seats fired, 7
abstained, `gated_by_truncation: false`. Submitted 12:31, verdict 12:36: **five minutes**, not
the ~30 the handbook budgets — worth recording because the queue depth, not the council, is
what makes that number vary.

Four `low` objections and three "missing" flags. **Two were answerable with a note; two were
not, and those two are the value of the round:**

| seat | objection | what it cost |
|---|---|---|
| guardian, edit 5 | `delivery_test.go` mirrors `api/server.go` "by discipline only" — no structural guard | **REAL. Fixed in `d1a4bdcdf`**: `RegisterRoutes` now owns the table and both callers use it. Fixing it exposed a divergence that already existed — the test router registered HEAD and production does not |
| guardian, edit 4 | confirm nothing else GETs `/c/:token` expecting the old mutating behaviour | **REAL as a check.** Enumerated, not asserted: every repo-wide hit is inside this change; live config gives `agent_definitions` naming `/c/` = 0 and current `site_specs` naming `links.webdesign.uk` = 0, **against a control of 21 current specs mentioning webdesign at all** |
| editquality, edits 3 and 6 | comment-only edit is not coverage; the self-correction could have folded into edit 5 | Accepted, no change. Edit 3 was never offered as mechanism coverage — it deletes claims the change falsifies |
| prior_art_librarian + bug_historian | cannot see `customer_access_tokens` (different database), asked a human to re-run | Run here [MEASURED 2026-08-25]: **0** tokens, **0** of purpose `confirm_transfer`, sites handed_over **0**, transfer_confirmed **0** |
| debug_historian | no post-deploy pod verification in the plan | Accepted into the close-out below. Note it asked for a symbol grep; **the estate's own landmine forbids that** (`strings` is absent from the images and a discovery grep matches Go's digit table). Doing it as a capability probe plus provenance ancestry instead |
| architecture | `ARCHITECTURE_SIGNAL: point_fix`, DEFLECTIONS unknown | No trigger. "No new shared contract… blast radius is genuinely one endpoint" |

> **The mutation proof got stronger as a side effect of answering the guardian, which is not
> what I expected from a low-severity objection.** Before `d1a4bdcdf`, "route GET back to the
> confirm handler" mutated the TEST's private copy of the route table. Now it mutates
> `RegisterRoutes`, i.e. the code production actually runs. A mutation that can only be
> applied to a test fixture proves less than one that can be applied to the shipped path, and
> nothing in the suite would have told me the difference.

**CLOSE-OUT, still owed — this is committed and NOT live.** `24b63120d` + `d1a4bdcdf` ride the
next core-manager roll. Then, in this order:
1. `kubectl -n ai-persona-system logs -l app=core-manager --tail=2000 | grep -m1 'build provenance'`,
   then `git merge-base --is-ancestor d1a4bdcdf <the stamp>`, plus the reversed test as a control.
2. **Probe the capability, not the commit**, from OUTSIDE: `GET https://links.webdesign.uk/c/<43-char>`
   must return the button page, `POST` the same path must return a page, and
   `POST …/c/<token>/confirm` must 404 **at the box** (that last one is the control — see the
   new LANDMINES entry).
3. Only then is the webdesign lane's delivery email unblocked.

### 2026-08-25 afternoon — handoff item 3 (the `WriteSiteSpecAction` deep-merge) is REFUTED by its own census. Do not build the guard.

The 08-24c handoff carried this as an owed follow-up: *"`WriteSiteSpecAction`'s deep-merge lets
`"banned_claims": []` empty a register wholesale (`site_spec_actions.go:554`) … census the
legitimate-shrink history first — the scheduled refresher may shrink registers by design."*
The census was the right instruction and it killed the item. **All figures [MEASURED 2026-08-25]
against `clients_db`, `site_specs` where `aspect='evidence_base'`: 336 rows all-history,
19 sites, 19 current, earliest 2026-07-16.**

**1. The code property is real and UNREACHABLE for this aspect.** `siteSpecDeepMerge` takes the
scalar-overwrite arm for anything that is not a map, so an incoming `[]` does replace an array
wholesale — that half of the claim is true. But **no agent writes `evidence_base` through this
action.** 20 live steps call `write_site_spec`; their aspects are `identity`, `strategy`,
`briefing`, `content_direction`, `site_plan`, `mission_brief`, `roadmap`, `tools`,
`design_intent`, `offer_ordering`, `vertical_landscape`, `classification`, `submission`,
`mission`, `roadmap_brief`. **None is `evidence_base`, and across ALL definitions including
snapshots and inactive rows it is still zero.**
**Control, because a literal-only enumeration is exactly how you miss a templated value:**
`{{`-templated aspects on `write_site_spec` = **0**, against **157** steps elsewhere in the same
configs that do use templates. So the zero is a real zero, not a pattern that cannot match.

**2. The refresher does not shrink registers — the hypothesis is refuted, not confirmed.**
Successive-write deltas per writer, over the 336-row history:

| writer | writes after the first | shrank bans | emptied bans | shrank facts | grew bans |
|---|---|---|---|---|---|
| `evidence-refresher` (`source='scheduled'`) | **222** | **0** | **0** | **0** | 1 |
| `evidence-researcher` (`source='research'`) | 9 | 0 | 0 | 0 | 1 |
| every manual/lane writer combined | ~105 | 2 | 1 | 6 | ~15 |

That matches the refresher's own design statement (`refresh_evidence_base_action.go` header:
*"never invents a fact, never removes one"*), and it reads the document as a generic map
specifically to preserve keys it does not own. **The guard would have had nothing to protect
against in 231 machine writes.**

**3. The one emptying in all history lasted 59 SECONDS.** `vonc.com`, 2026-07-24 14:22:41,
`session-2026-07-24-043-treatment`, bans 9 → 0 and top-level keys 6 → 1 — the exact wholesale
-replacement shape. The **same** `created_by` restored it at 14:23:40. It was one session's
two-part write, not damage. ⚠ **I nearly wrote it up as the motivating case before reading the
next row** — a single flagged transition is not an incident until you look at what follows it.

**4. The six facts-shrinks are legitimate curation and say so in their own `notes`**:
de-duplicating one repeated cap, dropping a source dated 2017 found during vetting, a run that
"returned zero usable facts". **So a guard on ANY shrink would have produced six false refusals
and one true one. A guard on shrink-to-EMPTY would have fired once, on the transient.** That is
the threshold the console door already uses, and it is now evidence-backed rather than chosen.

**5. No live register is disarmed.** All 19 current registers carry bans or facts, so none
parses to nil: claims checking is on everywhere today.

**The residual, which is NOT what the item described.** **8 of 19 sites have zero refresher
writes ever** — `apis.uk` (40 bans), `webdesign.uk` (33), `relojistas.com` (9), `noted.co.uk`
(7), `adversecreditmortgage.co.uk` (6), `remortgagecalculator.uk` (6), `finetuning.uk` (3),
`webdesign.co.uk` (1). Nothing re-derives those, so the self-healing that saved `vonc.com`
does not exist for them — and the only door that has ever emptied a register is **hand-written
SQL**, which no Go guard can reach. The two doors code CAN guard are the admin API (guarded
2026-08-24, shipped) and the agent action (no caller). **So the honest next step is not a
council round; it is that a migration touching `evidence_base` on one of those 8 sites deserves
the `DO`/`RAISE` verify block, and that is a review habit, not a mechanism.**

### 2026-08-25 late — owner feedback on the Builds screen, and a stale instruction in our own handoff

Owner used the screen on `agritec.uk` and reported two things: the orchestration list runs long
and the older rows are not interesting, and **"I don't see a terminate button anywhere"**.

The second is not a bug and the screen was wrong to leave it implicit. Resume and Terminate
render only on a **non-terminal** workflow; every row on that site is `Complete`, so no button
can appear. Fixed by saying so, and by pinning anything still running above the fold — a
running row hidden behind a truncation is precisely how an operator concludes the button does
not exist. The list and the detail view now share ONE terminal test rather than two copies of
the same status array, so what gets pinned is by construction what has the buttons.

**And our own handoff was telling the owner to do something impossible.** 08-24c said to press
Terminate on "some months-old EXECUTING_STEP orchestration". `[MEASURED 2026-08-25]`
`orchestration_states` holds **nothing older than 2026-08-24** — about two days — so there are
no months-old rows to find. Fleet-wide there are **7** non-terminal orchestrations and **2** are
tagged to a site, both on `webdesign.co.uk`, started today, in the other thread's live lane.
**So the terminate fix stays proven at the statement only, and that is a data limit, not a
missing action** — terminating another lane's live run to exercise a button would be the wrong
trade. Handoff item 4 corrected in place.

> **The transferable bit: a screen that renders an action conditionally owes the reader the
> condition.** Our own PLAN said "Resume button on non-terminal rows" and the code did exactly
> that — correct, and still a support question within a day of the owner using it. The absence
> of a control is indistinguishable from a broken one unless the screen says which.

Also corrected: the list heading printed `({workflows.length})` and the API's `count` field is
`len(workflows)` — both the PAGE SIZE. "(50)" read as a total when it was the window. Now
"newest 50+", with the `+` only when the window is full.

### 2026-08-25 evening — the roll landed: second-click page LIVE in the cluster, box step still outstanding

Release carried it. core-manager stamp `a7459a44b`, image `v1.0.1339`, **one distinct stamp
across all pods** (checked — a release can straddle and ship several revisions under one tag).

**(a) Ancestry, all three, with the reversed control failing as it must:** `24b63120d` (method
split), `d1a4bdcdf` (shared route table), `d30917150` (the other lane's delivery listener). One
without the others would have been a half-shipped state, which is why all three were checked.

**(c) Capability probes, in-cluster, run from the dashboard pod** (core-manager's own image has
**no** `curl`, `wget` or busybox — worth knowing before planning a probe):

| probe | result | meaning |
|---|---|---|
| `:8090/api/v1/admin/work-items` | **404** | admin API absent from the delivery listener |
| `:8088/api/v1/admin/work-items` | **401** | …and the PAIR is what makes that mean something: the path exists on the admin port |
| `:8090/c/<token>` GET | **200**, `<h1>Confirm you have moved your site</h1>` + `<form method="post">` | the button page, no `action=`, no success copy |
| `:8088/c/<token>` GET | **404** | the routes were MOVED, not copied |
| `:8090/health` | **404** | delivery-only really is delivery-only |
| `:8090/c/<token>` POST, unmatched token | **"That link is no longer active."** | routing, handler and DB lookup all exercised, nothing mutated |

`customer_access_tokens` = **0** and `sites.transfer_confirmed_at` = **0** after all of it.

**(b) The box has NOT been applied yet, and I can prove that rather than assume it.** From
outside, all three external probes 404 — including the control, so status codes alone could not
tell "box not applied" from "box broken". **The BODIES discriminate:** the token-shaped path
returns `404 page not found` as `text/plain` — that is **gin's** 404, so the request crossed
WireGuard and reached core-manager, on the port where `/c/` no longer lives (`:8088`). The
suffix path returns nginx's HTML 404 page, i.e. it died at the box, so the anchored regex is
intact. That is exactly the intermediate state `links.webdesign.uk.nginx`'s header predicts,
and it is harmless: zero tokens exist.

> **The check worth keeping: when every arm of a probe returns the same status, the status is
> not the instrument.** Two 404s from different sources are different facts, and the body,
> the content-type or the `Server` header will usually say which. I nearly recorded "the box is
> not applied" as an inference from the vhost header saying it would be — which would have been
> a guess dressed as a result, and would have read identically if the box had been applied and
> the regex broken.

**What is still NOT proven, stated so nobody reads this as complete:** the SUCCESS arm of the
POST — a real token being redeemed and `transfer_confirmed_at` being stamped. It needs a minted
token against a real site, and confirming stamps that site as customer-confirmed, which feeds
the retract-on-schedule path. **That is a production mutation for a test and it is the owner's
call, not a session's.** Everything up to it is proven at the live endpoint; the redemption SQL
itself was verified against real Postgres in a rolled-back transaction on 2026-08-20.

---

## 2026-08-25 — post-roll review of the Builds screen (owner asked for a Fable pass; the pass was Opus)

The owner asked to look the Builds-screen work over again **with Fable**. Two Fable review
agents were launched with full hunt lists and died on launch — the Fable session limit was hit
(resets 12:10am London). Rather than stall, this session (Opus) ran the identical review itself,
stated plainly here so the record does not claim a second model's eyes it never had. A genuine
Fable re-run is available after reset if the owner still wants it.

Scope: the LIVE code at HEAD — `e6350e74b` (backend), `b3fbfdd02` + `1a8db99f9` + `8e5a35ef9`
(SPA) — plus the council's three advisory objections on `45b3c93f` (APPROVED r1). Full findings
with evidence: `PLAN_2026-08-24_build_steps_screen.md` §7 (this entry is the pointer, not a fork).

Headline: the two real findings are both **pre-existing defects the new code armed or made
visible**, not defects in the new code — terminate's blast radius is the whole correlation
(`correlation_id` is NOT unique: 6,936 rows / 3,102 distinct, up to 19 rows share one, **5**
terminate-eligible multi-row correlations live right now), and the drill-down `QueryRowContext`
with no `ORDER BY` returns an arbitrary sibling. The council's sharpest open question
(editquality: raw bytes vs re-marshaled struct) resolves CLEAN — the INSERT stores `body.Data`
verbatim — but the sqlmock tests carry no `WithArgs`, so the exact landmine mutation would
survive the suite. bug_historian's WriteSiteSpecAction objection: the 8d134735d refutation
HOLDS at HEAD (aspect is a config literal; 0 live agent_definitions pair `write_site_spec`
with `evidence_base`).

### 2026-08-25 evening (later) — the box was applied; `/c/` is LIVE ON THE PUBLIC INTERNET and proven

Owner applied `links.webdesign.uk.nginx` (`nginx -t` ok, reload clean) after I applied the
egress fence's `8090` line. Box-local three arms: **200 / 404 / 404**, as the runbook expects.

**From the public internet, this session:**

| probe | result |
|---|---|
| `GET https://links.webdesign.uk/c/<43-char>` | **200**, `content-type: text/html`, `cache-control: no-store`, `referrer-policy: no-referrer`, `x-frame-options: DENY`, body `<h1>Confirm you have moved your site</h1>` + `<form method="post">` + the button |
| `POST` same path, unmatched token | **200**, `<h1>That link is no longer active.</h1>` — routing, handler and DB lookup all exercised |
| **CONTROL** `POST …/c/<token>/confirm` | **404 with nginx's HTML body** — died at the box, never reached the service; the anchored regex is intact |
| **CONTROL** `GET /other` | **404** |
| after all of it | `customer_access_tokens` **0**, `sites.transfer_confirmed_at` **0** |

The `<form method="post">` carries **no `action` attribute**, so the token stays in the request
line and never enters the HTML — verified at the served page, not only in the unit test.

**⚠ THE FIND OF THE DAY, and it was one command from being missed.** The other lane's commit
`d30917150` correctly added `port: 8090` to `networkpolicy-wireguard-egress.yaml` **and nobody
applied it.** The live policy still allowed `8088` only, so the repo read correct while the
cluster refused the exact port the new vhost proxies to. Applying the box config first would
have produced a **502 on every customer link**, on the box, after the owner had already made the
change — with the repo, the commit and the runbook all saying it should work.
Proven before touching anything, from the WireGuard pod, with controls: 8090 blocked, 8088 open
(the probe works), 9999 blocked (the fence discriminates). After applying: 8090 open, 8088 open,
9999 blocked, **and postgres re-checked still blocked** — the containment the fence exists for.
`configured` not `unchanged`, and the pod never restarted (same start time, 0 restarts).

> **The transferable check: `committed` and `applied` are independent facts for CONFIG too, and
> config has no image tag to give it away.** For Go we have the provenance stamp and the whole
> estate knows to ask the pod. A NetworkPolicy, a ConfigMap or an overlay edit has no equivalent
> tell: `git log` shows it landed, `kubectl get` shows the old value, and nothing reconciles
> them. **Read the LIVE object, and prove reachability from the pod that must do the reaching,
> with a must-fail control.** Runbook corrected in place (`37f49291d`).

**Now unproven, and only this:** the SUCCESS arm — a real token redeemed, `transfer_confirmed_at`
stamped. Offered to the owner; it mutates a real site row, so it waits on their say-so and their
choice of site.

### 2026-08-26 — fresh build `v1.0.1341`: everything re-verified, and a defect of mine found by another lane's measurement

**Re-verified rather than repeated** (yesterday's proofs expired with the build). core-manager
stamp **`2fb40a960`**, one distinct stamp across all pods; all three of this lane's commits still
ancestor-proven with the reversed control failing. `SERVICE_SERVER_DELIVERY_PORT` = 8090 survived
the roll, the live egress policy still allows `8088 8090`, and from the public internet `GET`
returns the button page with its headers while `POST` returns the spent-link page. Controls: the
suffix path still dies at the box with nginx's body, `/other` 404s. Tokens 0, confirmed 0,
handed_over 0.

**`48e75aad2` (another session, overnight, now live) found what I had assumed away.**
`correlation_id` is **NOT unique** — one correlation is a TREE of orchestration rows.
`[MEASURED 2026-08-26]` 6,031 site-tagged rows; a single site has one correlation shared by **27**
rows, then 25, 25, 22, 21. They fixed the two backend consequences (terminate was relabelling
finished siblings FAILED; the detail SELECT answered with an arbitrary sibling, so the step-error
panel appeared and disappeared between two clicks on the same row).

**The third consequence was mine and they could not see it from the backend:** the list rendered
those sibling rows with `key={wf.correlation_id}` — a React key on a value that repeats, on a list
that re-polls every 10 seconds. Fixed in `c016b3fb4` by grouping on correlation and showing `×N`
in the screen's existing idiom, which is also the honest unit: every sibling row already opened the
same detail once they pinned it root-first.

> **The check this earns: when a JOIN or a list keys on an id, ask whether that id is UNIQUE IN
> THE TABLE YOU ARE READING, not whether it is unique in the concept.** "Correlation id" reads as
> one-per-workflow and is one-per-TREE. I wrote the list, the paging, the truncation and the
> pinning on top of it and never asked; the count that would have told me is one `GROUP BY … HAVING
> count(*) > 1`, and it returns 27. The cost of not asking was invisible — the screen renders,
> nothing errors, and duplicate keys only misbehave while polling.

Handoff for the next session: `HANDOFF_2026-08-26_continue_here.md`.
