# NOTES — gripper dossier pilot

*Technical log. Append-only, newest at the bottom. Missteps included.
Design of record: `DESIGN_2026-07-24_gripper_dossier_pilot.md` (consolidated
from two grounded design passes; §5 seam reconciliation wins over either
half where they disagreed).*

---

## 2026-07-24 — workstream opened; scoring action built

Parent: per_site_ai (PLAN D15/D17), features_open/013. Owner unpaused after
robot-hands-site-fixes closed (R1–R9). Pilot scope locked: chat intake
(island) → pull → report-builder workflow → deployed report page + status
sidecar → island emails link.

### Owner decisions (2026-07-24, AskUserQuestion)
Shared sender (robot-hands@contactforsales.com) · prod E2E fixtures approved
(clean up after) · soft-launch unlinked · $50/mo cap on the new Anthropic
key. Defaults accepted: §5.3 mapping table, UUID+noindex, 24h/90d retention,
max_attempts=1. Tunnel VERIFIED LIVE (tools.apis.uk → 404 from island Caddy;
/api/v1/tools/health → 502 = no engine container yet, correct). Remaining
owner action: issue the $50-capped key.

### Built: score_grippers (first task)
- `platform/orchestration/actions/score_grippers_action.go` + registry entry
  + `score_grippers_action_test.go` (all green) +
  `sql_for_agents/204_robot_hands_matchmatrix_normalized_specs.sql`
  (NOT yet applied; pre-image safe — names no actions).
- Port source: dumped the live tool template (32,228 chars) to scratchpad;
  physics ported line-for-line (MATERIALS μ/ferrous map; dyn=m·a·S;
  fJaw=dyn/(μ·n); fDir=dyn; mEq=dyn/9.81; ipRow only under a requirement;
  unknown-never-pass; conflict note; verdict thresholds incl. the 1.25×
  marginal band; rank/headroom sort; the tool's exact no-match sentence).
- Input hardening beyond the tool (server context): material accepts μ key
  OR name alias; ip accepts "IP54" form; cycle_rate→safety tier (≤10→2,
  ≤30→3, >30→4, explicit safety_factor wins) per DESIGN §5.3; malformed
  spec = hard error (routes to error_step, never a guessed default).
- **Misstep, caught by the test run**: my conflict-note test asserted
  impliedMu = (mass·a·S)/(force·n) = 0.15; the tool actually derives it from
  the PUBLISHED PAYLOAD, (payload·a·S)/(force·n) = 0.26. The code was right
  (ported faithfully); the test expectation was wrong. Lesson: when porting,
  test expectations must be computed from the reference implementation, not
  re-derived from what the formula "should" be.
- fact_block contract: opens with the mandatory sentence when match_count=0;
  marks unpublished figures "NOT PUBLISHED by the manufacturer — say so if
  mentioned; never estimate it"; carries per-candidate source_url +
  verified_date; carries the substituted formula strings (these double as
  the discriminating E2E artefact — no other page carries them).

### Verified this session (live)
- 10 products rows, all with source_url + verified_date, none with
  matchmatrix block yet (seed 204 will add).
- Island tunnel public; API path answers 502 (nothing behind Caddy yet).

### Committed + submitted (same day)
- `e19aa5d10` score_grippers + tests + registry + seed 204 (pathspec commit).
- `ede694ef2` gofmt fix (pre-commit pattern check caught the test file).
- `12fa24e6b` this workstream's DESIGN + standing docs.
- `5229d7fa1` per_site_ai session docs + features_open/013 + WRONG_CALLS +
  CAPABILITIES correction.
- Council gate: **SUBMISSION_CORR ffccb83c-1833-45cb-bb0f-8edcc874699e**
  (score_grippers change; first submission bounced client-side — `plan` must
  be an OBJECT {summary, edits[], grounded_in, risks}, not a bare edits
  array; header of 097 has the schema). If APPROVED, later platform commits
  in this lane carry `Council-Reviewed: ffccb83c-…`.

### Missteps this session (append-only, the point of this file)
1. Conflict-note test expectation derived from intuition instead of the
   reference JS (impliedMu uses published payload, not request mass) —
   caught by the first test run; code was right.
2. **A `| tail -2` pipe swallowed a `go test` build failure**, so the gofmt
   commit chained through on a red build. Lucky twice over: the failure was
   another session's uncommitted vet_med WIP (shared-tree-wont-compile
   class, verified by building a clean `git archive HEAD` — my code green),
   and the commit was formatting-only. Practice: never gate a commit on a
   piped test command; check `go build` exit directly or use pipefail.
3. First council submission rejected client-side on schema (see above) —
   cheap, by design (no credits spent).

## 2026-07-24b — batch 2: the four pipeline actions (all built, tested, committed)

- **Council verdict on batch 1 (score_grippers): APPROVED** (corr ffccb83c,
  17:07Z). Commits predate the verdict (advisory-first flow), so no trailer
  on them; verdict recorded here and resolvable by corr.
- **Batch 2 committed `2849564ec`** (pull_report_requests +
  emit_report_status_files / verify_report_prose / create_report_page /
  report_charts.go + registry). Council submission:
  **SUBMISSION_CORR 7ed137d1-361c-4f69-9361-9e4ba1dfa6bf** (pending).
- All tests green against clean `git archive HEAD` overlay (shared tree
  still carries the other session's vet_med WIP — never gate on the dirty
  tree).
- **Design deltas, declared**: ONE chart (headroom) — the payload scatter
  dropped: published payload is only comparable within payload-rated
  technologies, a cross-tech scatter would mislead. verify_report_prose
  gained a `context_field` (request strings like the mounting standard are
  legitimate prose context score_grippers never sees) and a SKU-shape check.
- **Misstep 4 (kept me honest twice)**: my "clean prose" test fixtures used
  the 2.5 kg + IP54 scenario — which is a GENUINE zero-match application
  (every candidate fails or goes unknown on unpublished IP). The gate
  rejected my own fixtures for missing the mandatory no-match sentence
  before it ever saw an LLM. Fixture bug, not gate bug — but also the first
  live proof the gate bites. Second: the SKU regex missed digit-leading
  models ("2F-140"); caught by its own test, tightened.
- **Accepted pilot residual (documented in the council submission)**: a
  fabricated plain-word product name with no digits passes the deterministic
  gate — left to the writer prompt + validate_page_content contamination
  checks. Revisit if observed.

### Next
1. Await council verdict corr 7ed137d1 (~30 min; missing row = queued).
2. Seeds 205 (report-dossier component), 206 (island config placeholder),
   207 (three agent_definitions), 208 (two scheduled_tasks, disabled) —
   205 pre-image OK; 207/208 strictly post-image.
3. Island service `cmd/gripper-intake/` (code+tests offline; live turns
   blocked only on the owner's $50-capped key).
4. Image roll (IMAGE_TAG bump; discriminating pod-grep: 'report-dossier' +
   'pull_report_requests' + positive control), then 206–208, then induced
   E2E fixtures (success / no-match / failure) per DESIGN §6.

## 2026-07-25 — batch 2 REVISE answered; truncation guard; a false trailer

### CORRECTION — the trailer on `8e8b55818` claims a review that did not happen

> **CORRECTED 2026-07-25:** commit `8e8b55818` carries
> `Council-Reviewed: 7ed137d1`. **7ed137d1's verdict was REVISE, not APPROVED**
> (`complete_revise | COMPLETED | 2026-07-24 20:30:31Z`) — and it had been
> decided the previous evening, before the commit was written. I appended the
> correlation id from these NOTES without querying the verdict, treating
> "submitted" as "reviewed". Forward-only: the trailer stands in history and
> `098` will bucket it as MISMATCH, which is the report working correctly.
> Full entry in `WRONG_CALLS.md` (2026-07-25). What caught it: writing the
> next commit message and re-reading the trailer-discipline rule.
> **The check is one query, and it belongs BEFORE the trailer, not after.**

### Built: truncation guard on the prose gate (`8e8b55818`)
`execute_llm_prompt` tolerates a cut response and stamps `__truncated` beside
`.result` — the step then SUCCEEDS with a fragment. A partial that PARSES is
still a partial: prose cut mid-section can close into valid JSON with every key
present and the honest caveats gone, and the numeric/SKU checks would pass it.
Marker path is **derived** from `prose_field` via `markerFieldFor` (the
`diagnose_council_decide` precedent) rather than a second config key, so it
cannot drift when `prose_field` changes; `truncation_field` overrides.

### Council REVISE on batch 2 (corr 7ed137d1) — answered in `b7fd2ef8b`
Decided by a gating objection from **compliance**. Seats: editquality approve,
tooling_provenance approve, guardian approve, render_guardian approve,
bug_historian object, reuse_agent object, compliance object. 5 abstained.

- **compliance HIGH** — a fabricated *plain-word* vendor passed the gate:
  `modelNumberRe` needs digits, and `check_claims:false` means
  `validate_page_content` is not backstopping report pages either. So "we also
  considered Piab" sailed through. Fixed with a **trace requirement over the
  vertical's real vendor field**: a name on that list is a violation UNLESS the
  scored candidate set / fact block contains it — which means it **relaxes on
  its own as the product index grows** (6 of 24 indexed today). This was
  logged in batch 2's own submission as an "accepted pilot residual"; the seat
  was right that accepting it silently was the wrong call for a feature whose
  whole claim is that names trace. Remaining residual (a wholly invented vendor
  on no list) is now named in the code and pushed onto the writer prompt.
- **compliance MEDIUM** — no proximate disclaimer on a new public page type.
  Added at the top of the dossier, not a footer: machine-generated, information
  not engineering advice, unpublished figures never estimated, verify against
  the datasheet before buying.
- **reuse_agent MEDIUM** — the pull was "mirrored line for line" from
  `intent_collector`'s `collectOneSite`; they asked for evidence that
  *extending* it had been considered, not just that it had been found. It
  hadn't. Extracted the genuinely identical half (GET + `X-Internal-Key` +
  NDJSON scan + raised buffer) to `scanInternalNDJSONFeed` and moved **both**
  call sites onto it — including the live collector, since leaving it
  duplicated is the objection. New `partialStreamError` keeps the load-bearing
  distinction the inline code had: mid-stream death = partial success (lines
  stand, resume from checkpoint); transport/status failure = fatal. Tested
  with `httptest`, incl. that an oversized line cannot masquerade as a
  complete feed.
- **bug_historian MEDIUM** — asked whether the new renderers expose Go's
  `missingkey=zero` (the platform's most recent unpatched root cause). They do
  not: pure `strings.Builder`/`Fprintf`, no template engine anywhere in
  `create_report_page_action.go` or `report_charts.go`. Now stated in the file
  header as a property to preserve, which is what they asked for.
- **bug_historian / render_guardian LOW** — the chart silently dropped
  candidates with no capacity figure, so a figure lost to an upstream bug
  looked identical to one never published. `renderHeadroomChart` now returns
  the omissions and the page **names them** ("Not plotted, because no
  comparable capacity figure is published for them: …").
- **guardian LOW** — asked for two collision checks before this lands. Run and
  clean: no existing `awaiting_report`/`reporting` status, no `report*`
  `handler_agent`, no `%report%` `item_type` in `site_work_items`; no registry
  key collisions. No namespacing needed.

### Seed renumbering — 205–208 are TAKEN, this lane uses 207–210
Other sessions filed `205_bug003_*`, `206_content_gap_planner_*` and
`206_planner_news_index_*` (206 is itself doubled). This lane's seeds are now:
**207** report-dossier component · **208** island config placeholder ·
**209** three agent_definitions · **210** two scheduled_tasks (disabled).
Seed **204** (matchmatrix normalized specs) is unchanged and still unapplied.
`create_report_page_action.go`'s comment still says "seed 205" — corrected when
the seed lands.

### Resolved: how the report-builder loads its work item
`query_database` takes `$N` placeholders plus `config.params` — an array of
dotted paths into collected_data, each tried again with an `input_data.`
prefix if it resolves nil. Live precedent: `fix-implementer.load_plan` uses
`params: ["input_data.fix_correlation_id"]` with `output_format: "object"`
(which also flattens row 1 to the top level, so later steps can path at
`loaded.field` without array indexing). Seed 209's `load_request` copies that
shape with `input_data.work_item_id`.

### Also confirmed live (for seed 209)
- `execute_llm_prompt` config shape, from the live `diagnose-agent`:
  `ai_service{model, provider, max_tokens, api_key_env_var}`, `temperature`,
  `input_fields`, `output_format`, `prompt_template` with `{{.step.field}}`
  interpolation. Returns `{result: <parsed>, type: "json"}` — so the prose path
  is **`report_prose.result`**, and the truncation marker is its sibling.
- The scratchpad was reaped between sessions; `head_archive` (clean
  `git archive HEAD` overlay) had to be rebuilt. All tests this session ran
  against that overlay, not the shared tree.

### Next
1. Resubmit batch 2 to council with `RESUBMIT_CORR=7ed137d1-…` so the trail
   accumulates. Only a genuine APPROVED verdict earns a trailer.
2. Seeds 207–210 (207 pre-image OK; 209/210 strictly post-image).
3. Island service `cmd/gripper-intake/`.
4. Image roll, then 208–210, then the three induced E2E fixtures per DESIGN §6.

## 2026-07-25b — council round 2 (REVISE again), seeds 207–210, fail_workflow, bug 076

### Round 2 verdict: REVISE (corr 7ed137d1, 17:09Z, ~7 min end to end)
Round 1's objections were all answered — none of them came back. The new
objections are on the *shape of the submission* and on going further:

- **editquality MEDIUM — minimality.** The truncation guard was bundled into
  the round-2 plan while being, in my own words, "not raised by the council but
  in the same honesty family". Correct objection: it is a separate change and
  is already a separate commit (`8e8b55818`). It comes OUT of the next
  submission.
- **bug_historian MEDIUM — one call site, generic defect.** The truncation
  guard patches the consumer, not the producer. Asked how many other
  `execute_llm_prompt` call sites exist and whether they were audited. **I had
  not asked.** Measured: **118 LLM steps across 58 active agents; 5 agent
  definitions reference `__truncated` at all.** Filed as `bugs_open/076`
  rather than fixed here — the blast radius of the obvious fix (fail closed by
  default) is unknown until someone sizes it. This is the most valuable thing
  the council has produced in this lane.
- **guidelines MEDIUM — declared contracts.** Claims a workflow step reading a
  derived field (`<prose_field>.__truncated`) needs a declared `input_contract`,
  else the guard silently never fires. [UNVERIFIED — to check before the roll.]
  Evidence against: `diagnose_council_decide` reads exactly this shape from
  collected_data and works live. Evidence for: I have not read the validator.
  **Do not treat the guard as proven until an induced-truncation run fires it.**
- **tooling_provenance MEDIUM** — `intent_collector` is a live standing pipeline
  and should carry a travelling doc_notes row recording why `collectOneSite`
  now delegates. Not yet done.
- **tooling_provenance + editquality LOW** — the vendor list hardcoded names a
  data table already backs. **Fixed** (`82b34564d`): `loadKnownVendors` unions
  every manufacturer in `products` with the curated seed list. The two halves
  cover opposite gaps — live names become checkable with no code change; the
  seed half covers vendors we have NOT indexed, which is the actual fabrication
  risk and cannot come from a query.

### Built this session
- **`fail_workflow`** (`6b14055d7`), a new core primitive. Found while writing
  seed 209 by reading what `complete_work_item` actually does rather than
  assuming: its guard reads `response.status`, so a handler that catches its
  own failure, tidies up and calls `complete_workflow` reports SUCCESS and gets
  stamped 'complete' beside the evidence it failed. Cleanup paths had a choice
  between lying and skipping the cleanup. The report-builder's failure path
  needs to publish the island sidecar BEFORE ending in failure, so it needed
  this.
- **Report CSS inlined** (`5eb433e47`). rerender concatenates stored
  `rendered_html` and collects no component stylesheets; robot-hands.com's
  site_specs contain **zero** `report-*` classes, so the dossier — the paid
  deliverable — would have shipped as unstyled text (bugs_open/027 shape).
  Two drift guards added because the failure is SILENT: a test that renders a
  report and asserts every emitted class has a rule (**it caught two real
  misses on its first run, after I had already eyeballed the list**), and one
  pinning the verdict→class slugification. My first CSS draft styled
  `.report-card` while the renderer emits `.match-card`.
- **Seeds 207–210** (`4b752f5bc`, `034a3eade`), all dry-run against the live DB
  inside a rolled-back transaction.

### Missteps this session
5. **Wrote `Council-Reviewed: 7ed137d1` on `8e8b55818` without reading the
   verdict** — it was REVISE, decided the previous evening. Correction recorded
   above and in `WRONG_CALLS.md`. The check is one query and belongs BEFORE the
   trailer.
6. **Seed 208's first guard was silently broken, twice over.** I asserted that
   psql renders an unset `:'pull_key'` as a literal string and that the guard
   would catch it inside a `DO $$ ... $$` block. Both false: psql does **not**
   interpolate inside dollar-quoted strings (syntax error), and does not
   interpolate with `-c` at all. Verified against the live DB, then rebuilt on
   `set_config`/`current_setting` with the echo sent to `/dev/null` so the
   secret never reaches the terminal or pod logs. **Both failure branches were
   then tested rather than assumed.**
7. **Seed 207 used `created_from = 'seed_207'`**, which a CHECK constraint
   rejects (`manual|generated|adopted|tool|forked`). Caught by the rolled-back
   dry run — the reason to dry-run every seed rather than reading it over.
8. Put a `Council-Reviewed: not applicable (docs)` line on the 076 bug commit.
   Harmless (docs-only, and it does not claim approval) but it is trailer-shaped
   noise on a report that joins on that key. Don't decorate docs commits with it.

### Next
1. Resubmit round 3 with the truncation guard SPLIT OUT (editquality), the
   vendor fix included, and 076 cited as the answer to bug_historian.
2. Add the intent_collector travelling doc_notes row (tooling_provenance).
3. Check the `input_contract` question before the roll — if guidelines is right,
   the truncation guard is a silent no-op.
4. Island service `cmd/gripper-intake/`; then the image roll, then 207→209→210,
   then the three induced E2E fixtures (DESIGN §6).

## 2026-07-25c — batch 2 APPROVED (round 3), and what that verdict does NOT cover

### APPROVED — corr `7ed137d1-361c-4f69-9361-9e4ba1dfa6bf`, 18:03Z
`approved with 3 advisory objection(s) — none high-severity`. 14 seats ran:
**11 approve, 3 object (all advisory)**, `abstained: 2`, `unreadable: null`.
Objecting seats were editquality (2), guardian (2), prior_art_librarian (3) —
all advisory, none gating. Round 1's compliance HIGH did not return, and
neither did any other round-1 or round-2 objection.

Three rounds, and the trail is worth reading in order: round 1 found a real
hole in the honesty gate (plain-word vendor fabrication), round 2 found a real
hole in my *process* (bundling an unrelated change) and a real hole in the
*platform* (`bugs_open/076`), round 3 approved. **Every round paid for itself.**

### The trailer on `8e8b55818` is STILL wrong, and now wrong in a subtler way
> **NOTE 2026-07-25:** `8e8b55818` (the truncation guard) carries
> `Council-Reviewed: 7ed137d1`. That correlation is now APPROVED — but round 3's
> approved plan is **precisely the plan the truncation guard was removed from**,
> at editquality's request. So the trailer now points at a real approval that
> explicitly **excludes the change it is attached to**. That is worse than a
> stale claim: it reads as reviewed, resolves to an APPROVED verdict, and the
> verdict is for other code.
>
> Fixed the only way forward-only allows: the truncation guard was submitted on
> its OWN, **corr `37a32e02-19a7-409a-a74f-9363556bb39e`**, with the whole
> situation stated in its rationale. Whatever that verdict says is the real
> status of `8e8b55818`; resolve it by THAT correlation, not the one in the
> commit.
>
> **The transferable rule is sharper than "check the verdict first":** a
> correlation id identifies a *submission*, and a submission's contents change
> between rounds. A trailer is only true if the approved round still contained
> your change.

### Council-Reviewed trailer, going forward in this lane
Commits `2849564ec`, `b7fd2ef8b`, `5eb433e47`, `82b34564d` predate the verdict
(advisory-first flow) and carry no trailer — resolvable by corr 7ed137d1, which
is APPROVED. Later platform commits in this lane may carry
`Council-Reviewed: 7ed137d1-361c-4f69-9361-9e4ba1dfa6bf`, **but only for code
the approved round-3 plan actually contained** — which excludes the truncation
guard, and excludes `fail_workflow` (built after the submission; it needs its
own round).

### Not covered by any approval yet
- `fail_workflow` (`6b14055d7`) — a new CORE primitive, built after round 3 was
  submitted. A core workflow-control action deserves its own review; do not
  attach 7ed137d1 to it.
- The report CSS + drift guards (`5eb433e47`) — shipped after the round-3 plan
  was written, so also outside it.
- Seeds 207–210 — config, not `platform/`; out of council scope by the
  client-side rule.

### The approved round's ADVISORY objections — three checkable claims, now checked
The council approved, but `prior_art_librarian` objected that I had asserted
three codebase-shape claims **without attaching evidence**. That is a fair hit
and exactly the discipline in the working-docs rules ("a verified fact needs
its evidence inline"). All three run below. **One of my claims was wrong.**

**1. "input_contracts is consumed ONLY by call_agent" — OVERSTATED, correct me.**
> **CORRECTED 2026-07-25:** there are two further references. The claim as
> written was wrong.
```
grep -rn "input_contracts\.\|InputContract\b" --include=*.go platform/ internal/ cmd/ pkg/ | grep -v _test | grep -v ^platform/orchestration/input_contracts/
  platform/orchestration/actions/call_agent.go
  platform/orchestration/datahelpers/action_inputs.go
grep -rln "input_contract" scripts/ .githooks/
  scripts/goscripts/workflow_validator/main.go   (+ run/, docs)
```
- `datahelpers/action_inputs.go` → `GenerateInputContract` is a **producer**
  (builds contract JSON from an ActionInputSpec). It validates nothing.
- `scripts/goscripts/workflow_validator/main.go` →
  `validateInputContract` (line 608) checks in the **opposite direction** to the
  guidelines seat's concern: for each field the contract *expects*, it warns if
  **no step uses it** (`category: "unused_input"`, severity `warning`). There is
  no check that a field a step READS must be declared, and it is an offline
  script, not a runtime gate.

**The conclusion survives with better evidence than it had.** Nothing declared-
or-not affects whether `report_prose.__truncated` is present in collected_data:
it is a previous step's `output_field` in the same workflow, the shape every
live workflow already uses (`rendered_page.skipped`, `claimed.count`,
`deploy_result.commit_sha` — none declared anywhere). But **I asserted it from
one grep and called it settled**, and the seat was right that one grep is not a
codebase check. [STILL UNVERIFIED: the guard has never actually fired. Only an
induced-truncation run proves it — a green report proves only that no
truncation occurred.]

**2. "no third implementation of the pull shape" — CONFIRMED.**
```
grep -rln "bufio.NewScanner(resp.Body)" --include=*.go .
  platform/orchestration/actions/ndjson_feed.go        (only)
grep -rln "X-Internal-Key" --include=*.go .
  ndjson_feed.go · report_request_pull_action.go · intent_collector_actions.go
  + docs/**/traffic_probe, idea.uk, content_quality (VM SERVER code — the other
    end of the wire, not cluster pulls)
```
The extraction consolidated the only two cluster-side instances; there is no
third to fold in. (The doc hits are the servers being pulled FROM.)

**3. "no Go template engine in the renderers" — CONFIRMED.**
```
grep -n "text/template|html/template|template\.(New|Must|Parse)" \
  create_report_page_action.go report_charts.go
  → the ONLY match is my own comment at create_report_page_action.go:250
```

### The one advisory objection NOT acted on, and why
`guardian` (medium) says the `intent_collector` refactor is an unrelated live
pipeline bundled into a report-feature batch, names the contained alternative
(use the helper in the NEW action only), and notes there is no pre-existing
test on `collectOneSite` so behaviour-preservation rests on diff-reading.

**Two seats want opposite things**: round 1's `reuse_agent` objected precisely
because the duplication was left in place, and round 3's `guardian` objects
because removing it touches production. Both are right about their own risk.
The verdict is APPROVED, so the refactor stands — but guardian's real point is
the untested live path, and that is worth carrying: **if intent events stop
arriving from a VM-hosted site, this delegation is the first suspect.** That
sentence is now in the travelling `doc_notes` row for
`collect_intent_events`, where the next thread will find it.
`editquality` (medium) is also right that the doc_notes row was claimed in prose
but was not an edit in the plan — it was real (inserted before submission), but
a reviewer could not see it from the plan.

### RESOLVED — the truncation guard is APPROVED on its OWN correlation
**`37a32e02-19a7-409a-a74f-9363556bb39e`**, 2026-07-25 18:33Z:
`approved with 5 advisory objection(s) — none high-severity`. 12 seats ran
(7 approve, 5 object, all advisory), `abstained: 4`, `unreadable: null`.

So the change in `8e8b55818` **is** genuinely council-approved — just not by the
correlation its own trailer names.

> **HOW TO RESOLVE `8e8b55818`'s REVIEW STATUS (read this before trusting its
> trailer):** the commit says `Council-Reviewed: 7ed137d1`. That correlation is
> approved but its approved round **excludes** this change. The correlation that
> actually reviewed and approved it is **`37a32e02-19a7-409a-a74f-9363556bb39e`**.
> Forward-only: the commit cannot be corrected, so this note is the correction.
> `098` will join on the trailer and reach the wrong verdict for the right
> commit — a MISMATCH it cannot detect, because both correlations say
> `approved`.

**Both halves of the batch are now approved:** the pipeline under `7ed137d1`
(round 3), the truncation guard under `37a32e02`. Splitting it, which is what
editquality asked for and what I initially resisted by bundling, produced a
cleaner record than the bundle would have — each change is now resolvable to
the verdict that actually judged it.

**Still not approved by anything:** `fail_workflow` (`6b14055d7`, a new CORE
primitive) and the report CSS + drift guards (`5eb433e47`). Both were built
after the round-3 plan was written. A core workflow-control action in
particular should get its own round before anyone attaches a trailer to it.

## 2026-07-26 — coherence check (owner asked): the island half was about to fork the estate

Owner asked how this integrates with the other site tools and the plans for
them. Checked rather than reasoned about, and it caught a live collision.

**The finding.** The DESIGN's island half is 48 hours stale. It was written
when the island held only Postgres + Caddy; **tools-api shipped 2026-07-25**
and is live as `docker.io/aqls/tools-api:v1.0.1163`.
- `internal/tools-api/api/server.go:23` → routes mount at
  **`/api/v1/tools/<tool>`** — a namespace built for MULTIPLE tools.
- `internal/tools-api/store/sites.go` → CORS resolves per request against the
  island's own `sites` table: **multi-site by design**, not gauntlet-specific.
- Shared `RateLimitMiddleware` + `InputCapMiddleware` + one pgx pool + one
  Anthropic key already exist there.
- Island `Caddyfile` forwards **`/api/v1/tools/*` and nothing else**.

**So `cmd/gripper-intake/` as a second service was wrong**, and would have been
the *fourth* VM fork — the `vm_estate` PLAN independently warns: *"A third
divergence is being born in the island's compose/Caddyfile. Left alone this
becomes four."* Two threads reached that conclusion from opposite directions.

**Already-committed defect, found by this check:** seed 208 sets
`base_url = https://tools.apis.uk/api/gripper/v1`. The island Caddy allowlist
would **404** that path, so `pull_report_requests` would fail every tick with a
`returned 404` per site. Not applied yet → nothing broken in production, but
the committed value is wrong. Corrected target:
`https://tools.apis.uk/api/v1/tools/gripper`.
[UNVERIFIED: not re-tested end to end — the island route does not exist yet.]

**What survives unchanged:** everything in DESIGN §2 that describes BEHAVIOUR
(chat contract, honeypot + timing gates, status-sidecar polling, degraded mode,
retention, "email never leaves the island"). Behaviour ports into a route
group; only the packaging was wrong.

**What tools-api genuinely lacks** and must GAIN rather than have forked around
it: an SMTP mailer (it has none; only idea.uk's VM app has a working one in the
whole estate), the report-status poller, and the `/requests` pull endpoint.

**The wider coherence picture** (for `README_where_we_are`): 30 Tier-1 tools
live across 10 sites, all client-side, no backend. The dossier is the first
Tier-3 and composes robot-hands' own Tier-1 physics server-side — that is the
funnel working as designed. The real scaling risk is `score_grippers`: it is
**site-specific Go**, and a Tier-3 per site that each needs bespoke Go does not
reach 1,000 sites. The request→work→email-a-link journey now exists in idea.uk's
VM app and is being rebuilt here — the third implementation — which is exactly
what the **experience register** exists to stop, and the gauntlet became its
first harvest today. The dossier should be its second.

### Next (revised)
1. **Do not write `cmd/gripper-intake/`.** Build `/api/v1/tools/gripper` in
   `internal/tools-api/handlers/` instead; re-seed 208's base_url.
2. Coordinate with the gauntlet thread before touching tools-api — they own it
   and have `bugs_open/083` open against its error handling.
3. Everything cluster-side is unaffected and still stands (approved, committed,
   inert until the roll).

---

## 2026-07-27 — chasing WHY nothing caught the near-miss; three verified findings

The owner asked for the missteps to be recorded and for a plan to regulate this
class of divergence. Chasing the cause produced more than the near-miss did.

**The failure class, stated precisely.** Not "I failed to check for prior art."
I checked on 2026-07-24 and the check was **exhaustive and correct** —
`cmd/tools-api` did not exist in the tree that day. Had the design gone to the
council on the 24th, `reuse_agent` and `prior_art_librarian` would have found
nothing and approved, correctly, on the evidence. **The class is a fact that was
true at review time and false at build time**, and nothing in the platform
re-validates a decision after the world moves.

**Finding 1 — the decision lived in a medium no mechanism reads.**
`097_TRIGGER_council_review_v1.sh:53` sets `SCOPE_RE='^(platform|internal|pkg)/'`
and refuses docs client-side. Sound for credits (72 DESIGN/PLAN/SPEC docs were
created in `docs024_key_docs_latest` in July alone). Consequence: *a design
document that decides to build a new service is refused by the only mechanism
that would object to it.*

**Finding 2 — `prior_art` asks code questions into a void on the gate.** The seat
emits `code_checks` and its prompt promises they are *"answered from the
code_symbols index next round"*
(`0NN_fix_proposer_v20_prior_art_librarian.sql:51,61,73`). But
`0NN_council_gate.sql:40-45` records that the `code_lookup` step is **deliberately
not mirrored** onto the gate, justified as *"its authors are code-capable sessions
who read the objections themselves."* Confirmed in `099_SYNC_gate_roster.py:28`.
That justification assumes the author will look — and the reason the seat exists
is that authors don't. In my case I *had* looked, two days early.

**Finding 3 — the index behind that seat manufactures false absence.** This is
the real defect and it is filed separately.

```go
// composeSymbolContent builds the searchable text (embedded AND trigram-matched):
// kind + symbol + signature + doc + path.
```
`platform/orchestration/actions/code_symbols_actions.go:336-352`. **Function
bodies are never indexed** — they are read on demand by `ReadSymbolBody`
(`internal/analysis/symbolbody.go:31`), which the indexer never calls. So a
`content` check for any route, registry key, table name or string literal returns
zero rows, and the seat reads zero rows as absence. The contract documented at
`diagnose_code_lookup_action.go:29-31` says `content` matches *"symbol source
bodies"*; it does not. Its own example, `"%stop_reason%"`, is a literal that only
ever appears in a body — **the documented example cannot work.**

### The check, measured before proposing it

Predicate: a **staged** `.md` adds a line naming a `cmd/<x>/` that does not exist
in `cmd/`. Simulated over real history, added lines only:

```
git show --format= --unified=0 --diff-filter=AM <sha> -- '*.md'   # added lines
commits scanned: 1500  firing: 10  rate: 0.67%   (window 2026-07-19 → 07-27)
  12fa24e6b 07-24 docs(gripper-dossier): pilot workstream opened -> gripper-intake
  e9fb8a174 07-25 · ce97c8bca 07-25 · 79fd07caa 07-26           -> gripper-intake
  9658d3921 / af07067df / fc0652ce8 / d7b8f34d9  07-20          -> assembler
```
0.67% is inside `pattern-check.py`'s accepted band (SUMMARY 2.0%, README 0.7%).

**The property that matters is not the first fire.** On 07-24 the peer list would
not have contained `tools-api` and I would have been right to proceed. It is that
the check is free and idempotent, so it **re-fires on 07-25 and 07-26 with
`cmd/tools-api` newly present in the peer list**. No council re-runs itself for
free two days later. Hence: print peers annotated by recency, because the failure
mode is a peer that arrived *after* you looked.

> **CORRECTED — two sub-agent figures I nearly wrote down.** A sweep reported "8
> byte-identical `StartHealthServer` copies"; hashing the bodies gives **8 distinct
> hashes serving 1–3 endpoints each**, which flips that item from cheapest win to
> active trap (eight behavioural migrations on live liveness probes for zero
> benefit at any domain count). A second reported the doc detector firing "4 times
> in 1,614 docs"; a whole-tree scan actually fires on **~190**, almost all archived
> copies naming the retired `cmd/bundle` — the staged-diff figure above is the real
> one. Caught by running both commands. **A sweep's figure carries no measurement
> date and no method; treat it exactly like a figure copied from a sibling doc.**

### Answering the owner's actual question

*"Is it a council member, the diagnosis loop, or the architecture council?"* —
**none of the three.**

- **A council seat is the wrong instrument.** "Does this already exist?" is a
  factual question, and `pattern-check.py`'s founding doctrine is *"spend the LLM
  council on judgement, not on what a string comparison can settle."* Two seats
  already hold this remit; they need their instrument fixed, not company.
- **The diagnosis loop structurally cannot.** Verdicts are only
  CONFIRMED/REFUTED/UNVERIFIABLE; a CONFIRMED needs both a static *and* a runtime
  citation; the loop halts on `scope-not-narrowing`. A survey for duplicates
  widens scope by definition and can never confirm.
- **The architecture council already exists and is mid-flight in another thread.**
  `architecture_review/PROCESS_architecture_review.md` (RFC track, owner as
  authority, one RFC ratified) plus
  `DECISIONS_open_for_owner_2026-07-26_architecture_seat.md`, updated **today**
  with the owner's D7(a) ruling and a new D8. **Do not fork it.** Our measurements
  settle their D4 (`[UNMEASURED]`) and extend their D8: the corpus problem is
  worse than stated, because the Go index is itself body-blind.

Owner ruling this session: this thread builds the check and feeds them the
evidence; they keep the seat question.

### Doctrine proposed (one sentence, for the owner to ratify)

> **Divergence is allowed when it is parameterised and forbidden when it is
> copied.** A second implementation is fine as a row in a table or a profile; it
> is not fine as a second copy of the code.

Generalises `vm_estate`'s *"merge the generator, not the trust boundary"*. Worked
example, and it is a good one: `med_export_json` (`registry.go:1691`) sits **ten
lines above** `directory_export_json` (`:1701`), whose own header reads *"nothing
site-specific may be hardcoded here."* Both registered, both live, nobody saw it.

### The scale arithmetic, for the record

296 registry entries, 25 category strings against 10 declared, `site` alone = 107.
**9 of the 296 exist for 2 of ~1,000 sites, and 5 of those shipped in one week.**
A per-site action is a per-site entry in Go source needing a rebuild and a
redeploy — at 1,600 domains that couples site count to binary size. The pattern
that already does this correctly is `CHVerticalProfile`
(`companies_house_vertical_profiles.go`): a config table, not Go per vertical.

Owner ruling this session: **finish the pilot as-is, generalise after.** Prove the
Tier-3 shape end to end on one site before paying to abstract it from one example.

---

## 2026-07-27 evening — THE LANE RUNS END TO END. Fixtures 1 and 2 pass live.

The chassis rolled to **v1.0.1173 at 13:45** and again to **v1.0.1175 at 18:00** (both
other sessions), so the six actions went live without me doing the roll. Verified by
pod-grep against the *running* pod with a negative control (`gripper_intake_nonexistent` → 0)
— and re-verified after the second roll, because the pod I first checked was gone.

**Seeds applied** (dry-run in a rolled-back transaction first, all clean): 204 (gripper
matchmatrix spec blocks, 10 products), 207 (component), 209 (three agents), 210 (two tasks,
both seeded disabled). **208 deliberately NOT applied** — its `base_url` points at the
island route that no longer exists in that form.

### FIXTURE 1 — success. PASSES, live on the public internet.

`/reports/d1a371be-04a5-4ee6-b744-d64c6fd9e7c4.html` — **HTTP 200, 43,049 bytes**. The
discriminating string, the substituted formula literal **`(2.5 × 12 × 2) ÷ (0.15 × 2)`**, is
present on the live page; the negative control (`(9.9 × 99 × 9)`) is absent. Sidecar
`/reports/<uuid>.json` → `{"status":"ready","url":"…"}` HTTP 200. Zero references to
`reports/` on the homepage, so it is correctly unlinked.

### FIXTURE 2 — honest no-match. PASSES, live.

500 kg / IP67 / glass. Work item **`complete`** — an honest no-match deploys as SUCCESS, as
designed. `/reports/29c3f8aa-…html` **HTTP 200, 41,670 bytes**, carrying the mandatory
sentence *"No gripper in this index meets the requirement"* verbatim, and zero
Match/Marginal verdicts.

> **`deployed_at` is not fetchability** (`bugs_open/098`). Fixture 2 was **404 for ~2
> minutes** after the work item said `complete`, then 200. I nearly recorded that 404 as a
> failure. The git → Action → B2 → Cloudflare leg is real latency; poll it, don't sample it.

### FIXTURE 3 — induced failure. Observed TWICE, unplanned, and it behaved correctly.

Both of my own mistakes below drove the failure path, so it is exercised rather than
theorised: `handle_failure` → `publish_failed` → `fail_out` → **`fail_workflow`**, the work
item ending **`failed`** (never `complete`), with a precise `agent_error_log` row. That is
`fail_workflow` — the new core action — doing exactly the job it was written for. Not yet
observed: the `failed` **sidecar** live, because both failures happened before the publish
step. That is the remaining gap before this branch is "verified".

### Three defects found, all mine, all fixed

1. **`target_topic` was a topic nothing consumes.** Seed 210 named
   `system.agent.generic.requests` (the column DEFAULT). Live: **18 of 18 enabled tasks use
   `system.agent.scheduled.requests`**; the `.generic.` topic held 7 tasks of which the only
   enabled one was this bug. Fixed in the seed and live. It fails **silently** and looks
   healthy from the producer: the scheduler logged "Successfully produced message" and
   "Triggered task". The only discriminating evidence is downstream — zero
   `orchestration_states` rows for the agent type, zero mention of the correlation_id in the
   chassis log.
2. **Seed 204 was never applied.** `score_grippers` failed with
   *"no active grippers with matchmatrix spec blocks … (seed 204 applied?)"* — an error that
   named its own fix. The build order lists 204 at step 3; I jumped to 207.
3. **My fixture's `request_id` was not a UUID.** `create_report_page` refused
   `'fixture-1-success'` (17 chars). Correct: the id becomes the page's public URL.

> **CORRECTED — I misread an in-progress run as a hang and put it in another thread's bug
> file.** I recorded "2 of 2 hung at `spawn_handler`" in `bugs_open/029-hung-spawns`. Run 1
> did hang (4m45s, no `handler_spawned`, cleared manually). **Run 2 did not** — it completed
> in 92s and failed later at `score`. I sampled `current_step` ~20 s in and generalised from
> one reading. The tell was in the signature table I had *just written into that file*: a
> hang has `handler_spawned` ABSENT; run 2 had it present. Corrected there the same session
> (`f3cdc3377`) before anyone acted on it, with the load-bearing evidence moved to the four
> `build-pipeline-trigger` `spawn_dispatch` rows that carry a real error.
> **Cheap check skipped: wait for a terminal state before calling something stuck.**

### State

Lane **parked** (`report-dispatch` disabled) now the fixtures have run; `report-request-pull`
never enabled. Both fixture pages left live and unlinked for owner inspection — cleanup
(`source='manual-test'`) is owed once they have been seen. Owner has issued an Anthropic key
(capped per project, not per key — acceptable for now); it gates only the island half.

### FIXTURE 3 proper — the failing branch is now VERIFIED LIVE (19:18Z)

The earlier two failures never published a `failed` sidecar, and the reason was
mechanical rather than mysterious: `handle_failure` builds the sidecar from
`request.request_id` (`emit_report_status_files`, `status: failed`), so the
invalid 17-char id errored straight to `fail_out` and skipped `publish_failed`.
**An invalid request_id silently disables the failure-notification path** — worth
knowing, because the island's whole apology flow depends on it.

Re-run with a valid UUID and the designed fault (`mass_kg="not-a-number"`):

```
19:18:19  report-builder  score      score_grippers  field mass_kg: invalid value "not-a-number"
19:18:25  report-builder  fail_out   fail_workflow   gripper dossier build failed
19:18:26  report-dispatch-loop call_handler call_agent  CHILD_ORCHESTRATION_FAILED
```
- work item → **`failed`** (never `complete`)
- `https://robot-hands.com/reports/edd863e8-…json` → **`{"status":"failed"}` HTTP 200**
- **0 pages** created for that request id
- no error rows for `handle_failure` or `publish_failed` — they succeeded quietly

So `score_grippers` treats a malformed spec as a **hard error, never a guessed
default**, exactly as its header claims, and the failure path publishes.
**All three DESIGN §6 fixtures now pass, the failing branch included.**

Lane parked again: both `report-*` tasks disabled. Three `manual-test` work items
and two live report pages left in place for owner inspection; cleanup owed.
`HANDOFF_RESUME_gripper_dossier.md` written as the cold-start entry point.

---

## 2026-07-28 — a LIVE honesty defect, found by a test written to look for it

Went to plan the A1 generalisation (`score_grippers` → config-driven engine) and
came back with the opposite conclusion plus a customer-visible bug. Both are
worth the space.

### The defect: an unpublished cup range scored as `Match`

```go
if spec.Tech == "soft" && spec.GripMinMM != nil && spec.GripMaxMM != nil {
    // compare part size against the published range
} else {
    // "Not applicable — surface hold, no jaws"
}
```

Correct for vacuum and magnetic grippers, which have no jaws. **Wrong for a soft
gripper, which has a size window we merely do not know.** The unpublished case
fell through to "not applicable", set no `unknown` flag, and the candidate scored
`Match` on its remaining criteria — so the report would tell a paying customer
the part fits a cup whose size nobody has published. Six of seven criteria in
that file handled absence correctly; this one diverged, which is exactly why a
reviewer skims past it.

**Fixed** (`7f87c0afa`) to flag `unknown` like its siblings. **INERT until the
next roll** — the lane is parked, so no report can be generated meanwhile.
Transferable pattern written up in `016b` §9.

### The test that found it, and how it corrected me

`TestUnknownNeverPasses`: for every subset of the six nullable figures, across
every fixture and four request shapes, removing a published figure must never
turn a non-passing candidate into a passing one, nor raise headroom. 2^6 × 10 ×
4, under a second. It found the defect **on its first real run**.

> **My first version of the property was wrong and the code was right.** I
> asserted the rank could only ever *worsen* when a figure was removed. It failed
> at once: OnRobot 2FG7 moves from `No match` to `Insufficient data`. That is
> honest — without the figure we cannot assert failure either. **Losing
> information moves a candidate toward uncertainty in BOTH directions.** The
> guarantee is not "uncertainty is bad", it is "uncertainty is never mistaken for
> success". Asserting the stronger version would have been permanently red.

Also fixed in the same commit: the gate's contract (`no_match_sentence`,
`prose_sections`) now travels with the scoring output instead of being a package
const — a second report type would otherwise have been checked against this
one's sentence, silently — and `prose["summary_html"].(string)`, the file's only
unchecked assertion, reachable only on the `match_count==0` path and therefore
able to panic exactly where the gate matters most.

Council corr `721ac4f7` submitted; **no trailer until `decided_by` is read.**

### A1 is a WON'T-DO, and the number that justified it was wrong

Recorded in full in `features_open/024`. Short version: `CHVerticalProfile`, the
cited "config table" exemplar, is a **Go map** with one populated entry. The
"9 of 296 single-site actions" recounts to **1** — `pull_report_requests` already
selects sites by `deploy_config ? 'report_island'`, `emit_report_status_files` is
plumbing, and code-literal gripper mentions are 41 / 3 / 1 / **0** across the
four others. And N=2 already exists: idea.uk's Tier-3 scorer is an LLM 1–5 rubric
whose intersection with the proposed config table is **the empty set**.

**What generalises is the pipeline, not the scorer** — and four of the five
actions are already in that layer.

### Round 2 of council `721ac4f7` — and a correction to the handoff that set it up

**2026-07-28, later.** Resubmitted on the same correlation (`RESUBMIT_CORR`), six
edits unchanged, everything new in the rationale and evidence.

> **CORRECTED: the handoff's §4 named the wrong objection as gating.**
> `HANDOFF_2026-07-28_continue_here.md` said the gating objection was the
> DECLARED CONTRACTS one — `report-builder` having `input_contract`/
> `output_contract` NULL — and framed the whole next task around declaring them.
> It is not. `council_decide.decided_by` reads *"gating objection from
> bug_historian"*, and bug_historian's only **high** is on edit 3 and is about
> **enforcement**: *"the plan never establishes what the CALLER does with those
> violations… if logged/recorded but not used to block report delivery, this is
> exactly the documented shape in bugs_open/079 and bugs_open/083."*
> The contracts concern is real but **medium**, raised by `guidelines` (×2),
> `prior_art` and `guardian`. Caught by reading `decided_by` and then every
> seat's own objection severities instead of trusting the handoff's prose.
> **The cheap check that would have caught it at write time:** one query —
> `SELECT k, collected_data->k->'result'->>'verdict', <severities> FROM …
> jsonb_object_keys(collected_data) k WHERE k LIKE 'review\_%'` — which prints
> all ten seats in one row each. Had I resubmitted to the handoff's brief I
> would have answered four mediums and left the high untouched.

**Seat map, round 1** (10 reviewers, 6 abstained): bug_historian `object`
high:edit3 + low:edit4 · guardian `object` medium:edit2 + low:edit3 ·
guidelines `object` medium:edit2 + medium:edit3 · prior_art `object`
medium:edit2 · debug_historian `approve` medium:edit0 · editquality `approve`
low:edit1 + low:edit6 · constitution / diagnosis_guardian / mission /
reuse_agent `approve`, clean.

**The enforcement answer (bug_historian, high).** Three parts, all read rather
than argued, and the conclusion is that enforcement already existed:

1. `verify_report_prose_action.go:135-139` returns `(nil, error)` — the
   violations are the error text. The `logger.Warn` on :136 is incidental to
   the `return` on :137.
2. `coordinator.go:3350-3363` `routeToErrorStepOrFail`: step-level `error_step`
   first, config-level second, **and `failWorkflow` if neither**. There is no
   branch on which a failed step proceeds to its `next_step` — the engine is
   fail-CLOSED by default.
3. Live wiring: `verify_prose` has `config.error_step=handle_failure`,
   `next_step=compose_page`, and `compose_page` is the **only**
   `create_report_page` step in the fleet. A violation goes
   `handle_failure → publish_failed → fail_out (fail_workflow)`; no page.

So this is the **inverse** of `bugs_open/079` (repair computed, then discarded
at save): here the return value *is* the control-flow signal.

**LANDMINE, and it nearly reversed my own reading.** My first pass queried
`s.value->>'error_step'` and got NULL for **every** step in report-builder — it
reads exactly like "no error routing configured anywhere", which would have made
bug_historian right. `error_step` lives at `s.value->'config'->>'error_step'`.
This is already in **016b** (§ at line ~663, and the census at ~4890: *0 of
14,209 persisted plan steps carry the step-level twin vs 1,828 carrying
`config.error_step`*) — so the trap was written down and I walked into it anyway
by querying the field name I expected rather than the one that persists. Not
filed again; cite 016b.

**The contracts answer (guidelines/prior_art/guardian, medium).** prior_art put
it best — *"a shape-check that was skipped, not merely an omission"* — so I did
the shape-check, and the named mechanism does not fit this boundary:

- `input_contract` has **one** runtime reader: `call_agent.go:1005-1011`, under
  *"PRIORITY 1: Check for new explicit input_mapping"*. It validates the payload
  sent **to a target agent type** — an agent-to-agent call boundary.
- `score` / `write_prose` / `verify_prose` are plain actions with **no**
  `input_mapping` and no target agent; they pass data via `output_field:
  scoring` → dotted `scoring_field: "scoring"` in `collected_data`.
  `ValidateInputContract` is **structurally unreachable** on this path.
- Declaring it there would also be **false**: report-builder's `input_contract`
  says what a *caller of report-builder* supplies, and `prose_sections` is
  produced *inside* the workflow by `score`.
- **`output_contract` has zero readers, runtime or offline.**
  `workflow_validator/main.go:31` unmarshals it and nothing reads it;
  `validateOutputContract` (:637) never mentions `agent.OutputContract`,
  checking only `complete_workflow`'s `output_fields` against produced fields.

**Measured, not assumed** (the "~0% adoption measures the MECHANISM" habit):
of **182** active non-snapshot `agent_definitions`, **95** have `input_contract`
and **94** `output_contract` — so the columns are in real use and *"nobody uses
it"* was **not** available as a defence. Of those 95, **91** carry the runtime
`{required:[…]}` shape; the offline validator's own `InputContract.Expects` map
(`main.go:35`) appears in **none** of them, so `validateInputContract`
early-returns at `:611` for every live row. Two shapes, one column.

**[GAP — recorded, not closed]** The platform has a declaration point for
agent-call inputs (`input_contract`) and one for an action's step-config keys
(`datahelpers.ActionInputSpec`), and **neither covers keys inside a payload
handed from one action to another inside a single workflow** — which is exactly
what `prose_sections`/`no_match_sentence` are. I did **not** invent a third
mechanism: that is an architecture-scope change and this is a bug patch, which
is precisely the precedent CLAUDE.md now records against `bugs_closed/124`.
Named in the submission's `risks` so no reviewer has to infer it. Belongs at
architecture review on its own merits.

**editquality's edit-1 symbol.** Confirmed **correct and kept**, not changed:
`scoreGrippers` (:636) → `assessGripper` (:652 call, :580 decl) → `assessPayloadRated`
(:588 call, :505 decl), where the guard lives. A two-hop delegation was the
"structural relationship" the seat asked to see; the handoff's instruction to
"fix the symbol field" would have introduced an error.

**debug_historian's post-roll objection: discharged.** Re-verified myself rather
than repeating the handoff's table — v1.0.1194, **both** replicas
(`…-7p6d8`, `…-rxb52`, started 20:48Z): `carries no prose_sections` 1,
`carries no no_match_sentence` 1, positive control `No gripper in this index` 3,
negative control `nonexistent_marker_xyz` 0. Identical on both. The round-1
`risks` text calling the change "INERT until a chassis roll" is now stale and was
**withdrawn in the resubmission** rather than left standing.

Fleet-wide sole-consumer claim re-verified without a `LIKE` match this time:
a `jsonb_each` over every live workflow returns exactly three step rows for the
three actions, all `report-builder`.

### Round 2 verdict: **APPROVED** — and what actually moved

**2026-07-28 21:43Z**, run `75787940`, ~4 minutes end to end (the queue was clear
by the time it fired; round 1's 30-minute budget did not apply).
`decided_by`: *"approved with 1 advisory objection(s) — none high-severity"*.
Trailer id is the **submission correlation**, `721ac4f7-2076-4fea-9242-b234cfe648d6`.

Seat movement, round 1 → round 2 (same 10 seats, 6 abstained both rounds):

| seat | round 1 | round 2 |
|---|---|---|
| `bug_historian` | **object** — high:edit3 (enforcement), low:edit4 | **object** — medium:edit4 only |
| `guidelines` | object — medium:edit2, medium:edit3 | **approve, clean** |
| `guardian` | object — medium:edit2, low:edit3 | **approve, clean** |
| `prior_art` | object — medium:edit2 | **approve** — low:edit3, low:edit0 |
| `editquality` | approve — low:edit1, low:edit6 | **approve, clean** |
| `debug_historian` | approve — medium:edit0 | **approve, clean** |
| constitution / diagnosis_guardian / mission / reuse_agent | approve, clean | approve, clean |

**Nothing in the plan changed except its evidence.** Same six edits, same
sketches. Four seats moved to clean on reading — which is the honest read of what
round 1 was actually objecting to: not the change, the *unproven claims around
it*. The high did not survive contact with `coordinator.go:3350-3363`.

**The three residual advisories, none blocking, two of them owed:**

1. **`bug_historian` medium:edit4 — [ANSWERED HERE, worth carrying to any resubmit
   that cites it].** It asks whether this file was one of `bugs_closed/076`'s
   *"113 unguarded call sites"* and whether 076 produced a shared safe-extraction
   helper this fix should reuse. Checked, and the premise does not hold two ways:
   - **076 is a different mechanism.** It is about steps that *tolerate a
     truncated LLM response* (`tolerate_truncation`, opt-in, default false), not
     about bare type assertions on an LLM-parsed map. There is **no shared
     safe-extraction helper** to reuse — `grep` for `safeString|SafeString|
     safeExtract|SafeExtract` across `platform/orchestration/` returns nothing.
     076's fix was `truncation_guard.go` + `diagnose_council_decide_action.go`.
   - **The "113" is a figure 076 itself struck through.** Its own correction:
     *"CORRECTED 2026-07-26 — the title and the headline measurement are both
     wrong… There are not 113 unguarded call sites. Those 113 steps do not
     tolerate truncation at all"*, restated as **"37 of 118, not 113 of 118."**
     The seat quoted the file's *title*, which the file's body retracts.
   - And this file is already an **adopter** of 076's actual mechanism —
     `verify_report_prose_action.go:101-117` is the truncation guard, and
     `truncation_guard.go:38` lists `verify_report_prose` explicitly.

   So edit 4 is a genuinely separate class from 076. **What the seat is still
   right about**, and this part is not answered: whether bare `.(string)` on
   LLM-parsed maps recurs elsewhere unaudited is an open question nobody has
   measured. **[UNMEASURED]** — not filed as a bug, because a count is needed
   before it is a claim.

2. **`prior_art` low:edit3 — OWED, and it is the right ask.** The gap claim (no
   mechanism covers action-to-action `collected_data` contracts) is *"load-bearing
   for the decision not to declare prose_sections/no_match_sentence anywhere…
   evidenced with file:line citations rather than bare assertion, which is the
   right shape, but I have not independently confirmed it and it deserves one
   round of verification before it hardens into precedent for future plans that
   cite this one."* Exactly right, and precisely the failure mode this estate keeps
   paying for. **Nobody should cite this round's contracts argument as settled
   precedent until a second reader has walked `call_agent.go:1005-1011` and the
   `output_contract` grep.**

3. **`prior_art` low:edit0 — a correct piece of epistemic hygiene, no action.**
   It notes the council has **no check tier that can reproduce a pod-grep**, so
   its approval must not be read as independent confirmation that v1.0.1194 is
   live. It isn't; mine is the only evidence, and it is in the NOTES above with
   both controls and both replicas.

**Method note worth keeping.** Round 1 read as "the council disliked the change".
It did not — it disliked four *unevidenced assertions*, three of which I could
settle by reading code I had never opened, and one of which (enforcement) I had
asserted in a **file header comment** without checking the engine underneath it.
`verify_report_prose_action.go:20-22` still says violations route *"so the
workflow's error_step routes to the failure path"*, which is true for this
workflow and under-states the engine — absent an `error_step` it fails outright.
A comment describing the configured case as if it were the only one.

### Post-roll re-verification (v1.0.1196) and the httpguard adoption opened

**2026-07-29.** A roll I did not do took the fleet to **v1.0.1196** (both
replicas, 22:37/22:38Z). Re-grepped rather than assumed: `carries no
prose_sections` 1, `carries no no_match_sentence` 1, positive control `No gripper
in this index` 3, negative control `nonexistent_marker_xyz` 0 — identical on
both pods. The fix survives the roll.

**Adoption item opened as `bugs_open/139`** rather than as a recommendation.
Re-measured, not inherited: `platform/httpguard` and `platform/mailer` both exist
and `grep -rl 'agentchassis/platform/httpguard'` over every `.go` in the repo
returns **nothing** — still zero importers. And the exposure the handoff asserted
is real at **two** sites, not one: `internal/tools-api/middleware/ratelimit.go:30`
(`getLimiter(c.ClientIP())`) and `internal/tools-api/handlers/round.go:109`
(`ipHash := hashIP(c.ClientIP())`, which is **persisted** — the worse of the two,
because a poisoned identity column is a wrong answer nobody thinks to distrust).
`internal/tools-api/api/server.go:14` is `gin.New()` with **no
`SetTrustedProxies`**.

**[UNMEASURED], stated in the bug file too:** I did not fire the
`curl -H 'X-Forwarded-For: …'` probe at tools-api and did not read gin's
`ClientIP()` source. `bugs_closed/090` proved the mechanism against production on
idea.uk with exactly that one command, so it is cheap — but until someone runs it
here, 139's headline is an inference from two call sites plus a missing
`SetTrustedProxies`, not a demonstration. Marked as such rather than left to look
checked.

Filed as a NEW case rather than reopening 090: 090 is correctly closed (fixed,
deployed, proven live) on **idea.uk**, a different service on a different box.
Reopening it would have made its closure dishonest. Grepped both `/bugs_open/`
and `/bugs_closed/` first — nothing covered tools-api.

**Not taken on.** `tools-api` is the gauntlet_dead_cta thread's, with `083`
(slug `gauntlet_engine_503_discards_the_error`) open against it. 139 is the
evidence half of a conversation; `who-owns.py` reads commits and cannot see a
session mid-fix, so check the tree too before anyone acts on it.

## 2026-07-29 — the probe was run, and it refuted 139's headline

> **CORRECTION to the entry immediately above.** Its `[UNMEASURED]` marker was
> right that the probe was owed. It was wrong to expect the probe to confirm.
> **139's headline claim is REFUTED.** Full corrected record in the bug file;
> what follows is the evidence and the order it arrived in, missteps included.

**The two probes.** `POST https://tools.apis.uk/api/v1/tools/gauntlet/round`,
`Origin: https://vonc.com`, browser UA (a `Python-urllib`/curl fingerprint draws
Cloudflare's `error code: 1010`, per the gauntlet RUNBOOK). 07:46:07Z forging
`X-Forwarded-For: 203.0.113.77`; 07:53:26Z forging `X-Real-IP: 203.0.113.77`.
Both **200**. Both stored `client_ip_hash = 245c0ffc0f6a0215471542b9add1fa53`;
the gin log recorded `172.18.0.1` for both. `sha256("203.0.113.77")[:16] =
0c25434b09c62046f88142b1412b949e` appears nowhere in the table.

**What the constant is.** `sha256("172.18.0.1")[:16] = 245c0ffc…` — the docker
bridge gateway. Census:

```
SELECT count(*), count(DISTINCT client_ip_hash) FROM gauntlet_rounds;
 -- 83 rows, 1 distinct, 2026-07-25 → 2026-07-29
```

So the real defect is a **degenerate identity**, not a spoofable one: the "per-IP"
limiter is one global bucket for all visitors, and `client_ip_hash` has never
distinguished anybody. Worse than what was filed, and it needs no attacker.

**Wrong turn #1 — I had the mechanism backwards twice before measuring it.**
Reading `gin.New()` + no `SetTrustedProxies` + `gin.go:474` `validateHeader`
(walks right-to-left, every entry "trusted", returns `items[0]`) I predicted the
forged value would win. It did not, and the prediction was confidently derived
from correct source. The missing half was two hops I had not looked at, one of
which is not in this repo at all (`/opt/island/Caddyfile`).

**Wrong turn #2 — and I nearly wrote it into the bug file.** On seeing the
constant, my next theory was that adopting `httpguard` would *create* the spoof,
because `httpguard.ClientIP` prefers `X-Real-IP` and my local Caddy repro showed
Caddy forwarding a client-supplied `X-Real-IP` **verbatim**. Plausible, and
wrong: **Cloudflare strips `X-Real-IP` at the edge.** Caught only by firing it at
the `020` probe vhost *with an arbitrary `X-Zzz-Control` header alongside* — the
control arrived, `X-Real-IP` did not, so "absent" was distinguishable from "never
sent". Without that control the result would have been unreadable.

**The measured hop table** (each row is an observation, not a reading of docs):

| hop | instrument | result |
|---|---|---|
| CF edge → origin | `020` probe vhost access log (logs all headers) | forged XFF **arrives** as `203.0.113.77,2a02:c7c:…` — CF appends, same shape as 090's nginx |
| CF edge → origin | same + `X-Zzz-Control` control header | `X-Real-IP` **stripped**; control header arrived |
| CF edge → origin | forged `CF-Connecting-IP` | **403, `error code: 1000`**; control without it → 404 from origin |
| Caddy → app | local repro: `caddy:2.11.4` + the island's own Caddyfile → echo upstream | XFF **overwritten** with Caddy's peer in every case; `X-Real-IP` and `CF-Connecting-IP` forwarded verbatim |

The local repro is worth keeping in mind as a technique: the island's Caddyfile
plus the pinned image reproduced the hop exactly (same `172.18.0.1` peer shape)
on this machine, in about two minutes, with no risk to the live service.

**Consequence for our own NEXT action, which is the part that matters to this
lane.** The handoff's item 1 was "adopt `platform/mailer` + `platform/httpguard`
into tools-api". **`httpguard.ClientIP` would not have fixed this**: its peer gate
passes (`172.18.0.1` is RFC1918), `X-Real-IP` is absent (CF strips it), so it
falls to the **rightmost** XFF entry — which is `172.18.0.1`, the same constant,
now reached through a shared helper and therefore *reading* as fixed. And its
docstring's justification for preferring `X-Real-IP` ("set with
`proxy_set_header`, so a client-supplied one is replaced") describes **nginx on
idea.uk**, not Caddy here. That is a genuine design input for `features_open/024`
A3 — the package should name the front-end its rules assume, or take the trusted
header set as config — and it is **not** a reason to unpick A2/A3, which were
approved on their merits.

**Three rows left in their table**, ids and times listed in 139. Not deleted:
they are the gauntlet thread's data and the evidence, and tidying up after myself
inside another lane's production table is not my call.

---

## 2026-07-29 afternoon — the design input became the fix, and it was approved first time

The A3 design input recorded in the entry above is now shipped. `httpguard.ClientIP`
takes a **required** `FrontEnd` argument naming which headers *this deployment's*
proxy is known to **write** (as against merely forward). Three pre-declared:
`Nginx()` — byte-identical behaviour to the old hard-coded rules, so
`bugs_closed/090`'s regression test is untouched and still proven to fail on
reintroduction — `CloudflareTunnel()` (`CF-Connecting-IP`), `Direct()` (trusts
nothing). **Required, not defaulted**, deliberately: the defect was never the rules
themselves, it was that a caller could inherit an assumption it had never stated.
Commit `31c684124`, which carries the register entries PUB-002 + PUB-003 **in the
same commit** — the condition the 07-28 owner ruling makes non-negotiable for a
platform seam.

**Council `49392838-5ada-4c8e-baeb-94b01e5855b4`, round 1: APPROVED** — *"approved
with 1 advisory objection(s) — none high-severity"*. 9 seats fired, 8 abstained on
relevance, none unreadable. The `architecture` seat answered the venue question the
submission raised against itself, returning `ARCHITECTURE_SIGNAL: point_fix`:
*"hardening a shared mechanism's own contract before its first real consumer, which
is the cheapest point in its life to do it, not a shared-mechanism change smuggled in
via a symptom fix."* Worth keeping as precedent — it is the **inverse** of
`bugs_open/124`, and the discriminating test is whether any consumer's stated
guarantee changes. Both mediums asked for confirmation rather than argument, so both
were answered with evidence, not prose: `guardian` wanted the register entry
*confirmed* to ship in the same commit (it did), `prior_art_librarian` wanted the
load-bearing zero-importer claim re-grepped **after** the verdict rather than trusted
from the submission (re-run; still zero).

**The architecture seat's carried objection, discharged the same day** rather than
left to become folklore. Its wording: *"Fine to ship, but the open question should not
go stale - the next thing to land against this package should close it, not add a
fourth FrontEnd."* The open question is the peer gate — trusting `CF-Connecting-IP`
is trusting Cloudflare to be in front, and the gate is the only thing that makes the
header revert to being ignored if the origin is ever reachable directly. It was
implied, not stated. Now: `TestPeerGateBoundaryIsStatedNotAssumed` pins where the gate
falls across ten realistic address forms, and `TestClientIPParsesTheRemoteAddrARealSocketProduces`
drives a real `httptest` socket so the peer is parsed from what the **runtime** puts
in `RemoteAddr` rather than a hand-typed string — the case where a bug would actually
hide, since a leaked port or bracket fragments every limiter bucket per-connection.
Commit `df7f918b8`.

Measured while doing it, and it is a real deployment constraint rather than trivia:
Go's `net.IP.IsPrivate` covers **RFC1918 and RFC4193 only**. A proxy behind CGNAT
(`100.64.0.0/10`) or on a link-local address is **NOT** trusted and its headers are
ignored. Fails in the safe direction — coarse key, not a spoofable one — but it
decides where this package can be deployed, and it is now pinned by test instead of
discovered in production.

**What remains structurally unclosable locally, stated in PUB-002 rather than left
implied:** a connection from a genuine *public* peer. Every address a dev machine can
bind is loopback or RFC1918 and therefore lands on the trusted side of the gate by
construction. Only a real direct-exposure deployment closes it. **Do not let a future
thread "close" it with another unit test.**

### MISSTEP — my first mutation proof was worthless and reported success

I "verified" the two new tests could fail by mutating the peer gate with
`sed -i 's|peer.IsLoopback() || peer.IsPrivate()|peer.IsLoopback()|'`. The `||` in
the Go source **is the sed delimiter**, so sed printed `unknown option to 's'`, left
the file untouched, and the suite then passed honestly — a green `ok` that I had
briefly read as "the mutation did not break anything", which is the worst possible
misreading available. Redone with a Python script that **asserts the anchor string
was found** before substituting; both tests then FAILED on all four RFC1918/RFC4193
rows as intended.

**The transferable check:** a mutation test that passes is either evidence the guard
is redundant or evidence the mutation never happened, and those look identical from
the exit code. Make the mutator assert its own anchor, and read the mutated file
before trusting the run. Same family as *"a quiet test passes when the RULE is gone,
not when the guard works"*.

---

## 2026-07-30 — the patch was filed correctly and still did not arrive

Checked what became of the adoption patch left in the gauntlet lane's directory on
07-29 at 13:34 (`CONTRIB_2026-07-29_tools_api_client_identity_is_a_constant.md`,
commit `171ff677c`). Nothing became of it, and the *reason* is worth more than the
fact.

Measured, 2026-07-30:

| check | result |
|---|---|
| `grep -rn 'agentchassis/platform/httpguard' --include=*.go .` (excl. self) | NONE |
| `grep -rn 'agentchassis/platform/mailer' --include=*.go .` (excl. self) | NONE |
| `go test -count=1 ./platform/httpguard/ ./platform/mailer/` | `ok 0.003s` / `ok 0.002s` |
| `ls internal/tools-api/` | `api config db handlers httperr middleware store` — no `clientip` |
| last commit touching `internal/tools-api/` | `a9a1b3556`, 07-29 **09:34** — before the CONTRIB |
| gauntlet commits after the CONTRIB landed | **six** (07-29 14:24 → 18:27), none touching that service or the CONTRIB |

`-count=1` matters: without it both packages answer `(cached)`, which is not a run.

**The decisive measurement.** Their live cold-start doc carries a "Consolidation ping"
item ending *"Nothing owed yet."* I dated the line rather than reading it as current:

```
git log -S'Nothing owed yet' --format='%h %ad %s' --date=format:'%m-%d %H:%M' \
  -- docs/…/gauntlet_dead_cta/HANDOFF_2026-07-29_continue_here.md
# e304e3955 07-29 08:22 …
```

**Written 08:22 — five hours BEFORE the patch arrived — and unchanged through four
later edits of that same file.** So their next cold-start reads "nothing owed" with a
finished patch two files away. Not a judgement they made about the patch; a line they
never had cause to revisit.

**Landmine, and it is the LANDMINES/D10 authoring-vs-delivery gap on a new corpus:**
**a CONTRIB in the right directory is AUTHORING, not DELIVERY.** Nothing tells a lane
that a new file in its own directory applies to it. I filed the evidence exactly where
the convention says to and it still did not reach anybody.

**Also the general form of a check I keep needing:** *date a line in someone else's
document before treating it as their current position.* `git log -S'<line>'` costs one
command. Prose in a live doc carries no timestamp, and a stale line in a cold-start
path is indistinguishable from a considered decision.

**Acted on it, minimally and additively:** appended a dated note under their item 4 in
`gauntlet_dead_cta/HANDOFF_2026-07-29_continue_here.md` — appended, **none of their
words edited** — stating that the contact has arrived, that the finding is about their
service and does not depend on this programme (83 of 83 rows, one distinct
`client_ip_hash`, so the "per-IP" limiter is one global bucket), what the patch is, and
which part is `[INFERRED]` and better settled by them (that `CF-Connecting-IP` reaches
the app process — measured at Caddy, not at the app, and I would not add a header-echo
endpoint to their service to find out). Did **not** apply it: `tools-api` is theirs,
`bugs_open/083` (by slug) is open against it, and reaching in is what the
contribute-don't-fix convention exists to prevent.

> **SHARPENED the same hour, and it corrects the paragraph above — "it did not arrive"
> was too strong.** Running `./scripts/landmines-sync.py` (the obligation after any
> append, now that D10 is ruled and `LANDMINES.md` is authoritative) listed an entry I
> had not written: *"A 'per-IP' limiter behind Cloudflare is probably one global bucket
> — and `httpguard` reads as the fix"*, footprinted on
> **`internal/tools-api/middleware/ratelimit.go`** and **`platform/httpguard`**.
> `git log -S'global bucket'` attributes it to **`11654d102`, 07-30 13:28** — the D10
> lane, who read `bugs_open/139` and filed its substance themselves. With D10's
> session-start hook matching entries against the dirty working tree, a gauntlet
> session that opens that limiter file is now told automatically.
>
> **So the residue is narrower and more interesting than what I wrote:** the **warning**
> was delivered (by a third lane, through a mechanism built that morning, better than my
> appended note). The **patch** was not, and has no delivery path at all. **A landmine
> can say the ground is mined; it cannot hand you the three edits.** If D10 is ever
> extended, that is the gap to name — and note that my own check nearly missed this,
> because I measured *uptake by the owning lane* and the pickup came from a lane I had
> not thought to look at. `git log -S` over the corpus, not just over their directory.

> **CORRECTED AGAIN ~16:30Z, and this one retires the "not delivered" claim entirely.**
> The owner asked which thread owns the gauntlet lane. It is the session titled **`vonc 6`**
> (`c4daed6f-5514-49f1-be6a-7bbf6bbd3c98`, last active 07-30 15:19Z), and reading its
> transcript settled the delivery question:
>
> | 07-30 | what its own record shows |
> |---|---|
> | ~13:40Z | my note appended to their `HANDOFF_2026-07-29_continue_here.md` |
> | **14:12Z** | it reads that file, note inline (attachment record) |
> | **15:14Z** | reads it again |
> | **15:15:39Z** | runs a Bash call described *"has anyone acted on the CONTRIB?"* |
> | **15:16:48Z** | puts it to the owner as **item 3 of 4** — the 83-of-83 measurement, the ready patch, their own `httperr` precedent, the `[INFERRED]` last hop as theirs to settle, the acceptance check **with** the presence-check warning — and **recommends it before the distribution leg** because *"the rate limiter is the one thing that behaves differently once more than one person is arguing"* |
>
> **Read at 14:12, in front of the owner at 15:16.** So the landmine is *better* evidenced
> (right directory: nothing for a day; their cold-start path: under an hour) but the state
> claim above is dead — **they have picked it up, and the next move is the owner's
> sequencing call. Do not re-notify them.**
>
> **The methodological miss is the transferable part, and it is mine.** I measured uptake by
> **commits and their directory**, and a session that reads, decides and reports to the owner
> produces *neither*. **A quiet `git log` is not silence** — for "has another lane acted?",
> the artefact is their transcript (`~/.claude/projects/<proj>/<id>.jsonl`; `customTitle`
> sits on line 1), not the repo. Same family as *"your measurement answers the question you
> ENCODED"*: I encoded "did they commit?" and read the answer as "did they engage?".

---

## 2026-07-30 evening — the owner read the fixture pages and found two rendering defects nothing here could see

He looked at the served report and sent two screenshots. Both defects were real, both were
in `platform/orchestration/actions/report_charts.go`, and **both had passed every check the
pipeline has** — because an SVG `viewBox` **clips** rather than overflows, so neither
presents as a broken layout. They present as corrupted *content*, which in a report whose
header comment says *"every bar is drawn from a real figure the scoring action computed or a
manufacturer published"* is the worst possible disguise: the natural diagnosis is a scoring
bug.

| defect | mechanism | what the page served |
|---|---|---|
| value label runs off the right edge | `plotW = width - labelW - 90.0` reserved a **fixed** 90 units; `"6.42× (Insufficient data)"` needs ~150 at font-size 11 | `6.42× (Insufficien` |
| reference captions overprint | both captions drawn `text-anchor="middle"` on **one** baseline; 1.0× and 1.25× are 0.25 apart on a 3× axis ≈ 32 user units, captions ~110 and ~140 wide | `reqmiaegimealttufh1r.0ex/f)old (1.25×)` |

Fixed by deriving both from content instead of constants: the gutter is sized from the widest
label (plot shrinks, text is never lost), and captions are assigned to the first lane that
clears the previous caption's right edge, the SVG growing one 12-unit lane per extra lane.
Byte-stability is preserved — refs are still iterated in sorted key order, because an
unstable chart would re-diff every committed report page.

**Third defect, found while fixing the first two and worth more than either:** bars clipped
at the 3× cap were drawn *identical lengths* while the figures printed beside them differed
(6.42× and 7.60×). A chart whose doctrine is "no invented data" was drawing two different
numbers as the same bar. Capped bars now end in a point. Commit `f8e7c31ce`; council
submitted, corr `60d05267-a671-4b98-9b87-6a97e16d78a0`.

**Verified by LOOKING, because that is how it was found.** Rendered the exact fixture shape
through headless chromium and read the PNG. Two things to know before repeating it:
chromium here is a **snap**, so it cannot write a screenshot into `/tmp/claude-*` (fails
`No such file or directory`) or into any **dot-directory** under `$HOME` (fails
`Permission denied`) — a plain `~/chartcheck/` works. Both failures look like "the tool
isn't available" rather than "the path is refused".

### MISSTEP — my first clipping test passed under mutation, and the reason is a new trap

I reverted the gutter fix expecting the "nothing is clipped" assertion to fail. **It passed.**
Not because the mutation failed to apply (that is yesterday's landmine, and the mutator now
asserts its own anchor — it did apply) but because a **second guard absorbed it**: the
`fitText` fallback silently truncated the label to fit. Nothing was clipped. Content was
lost. The assertion was *true and useless*.

**The fix is to assert the outcome the caller asked for, not the absence of the symptom** —
the test now requires the emitted label to *equal* the requested label. Both traps are filed
in `LANDMINES.md`; they are companions and it is worth knowing which is which:
a mutation that never **applied**, versus a mutation that applied and was **compensated**.

Left undone deliberately: candidates at `0.00×` render no visible bar at all (zero-width
rect). It looks empty next to its own `0.00× (No match)` label, and a minimum bar width
would misrepresent the magnitude — which is the one thing this chart exists not to do. The
label carries the figure; the bar should not lie about it.

**The two live fixture pages are unchanged by any of this** — they are stored artefacts, and
`report-dispatch` / `report-request-pull` are both still disabled, so nothing regenerates
unasked. The fix reaches the next report generated.

> **SUPERSEDED within hours, deliberately, on owner instruction:** dispatch was enabled
> 2026-07-30 22:13Z and FIXTURE 4 queued 07-31 08:15Z. See the entries below.

---

## 2026-07-30 22:13Z — `report-dispatch` ENABLED on owner instruction, and it is correctly self-gating

`UPDATE scheduled_tasks SET enabled = true WHERE name='report-dispatch'` (only that one;
`report-request-pull` left off — enable order is dispatch first). Fired at 22:14:04 and
22:16:04, changed nothing, cost nothing.

**Why an idle ON is free, which I got wrong first:** the scheduler evaluates the task's
`pre_query` and, finding no rows, logs
`"Pre-query found no rows — task ran with nothing to do"` and **publishes no message at
all**. The query ends `HAVING count(*) > 0` precisely so that zero work returns zero rows.

> **WRONG CALL, logged in `WRONG_CALLS.md`.** I told the owner that `pre_query` "counts
> every `report_request` with no status filter" and would "read 3 queued forever" — offered
> as a platform defect. **False.** It filters on `status='awaiting_report'`, guards
> `attempt_count < max_attempts`, includes a stuck-claim reaper clause, and ends with that
> `HAVING`. I had displayed the column as `left(pre_query,120)` for table width and then
> reasoned about the truncation; the clause that refutes the whole claim sits at ~char 250.
> **If you truncate a field for display you may quote what you SAW, not what it MEANS.**
> Caught by the scheduler log contradicting my own prediction — I predicted "fires and
> claims nothing", it did not fire a message at all, and chasing *why not* is what made me
> fetch the full query.

Plumbing proven rather than assumed: `system.agent.scheduled.requests` appears in the
**live** deployment's `EXTRA_REQUEST_TOPICS` (not just the repo overlay), and a sibling
scheduled task (`endpoint-health-checker`) flowed over the same topic at 22:18:07 and
processed successfully — an untouched peer in the same window is the cheapest proof the
path works.

---

## 2026-07-31 08:07Z — the chart fix is LIVE, and the council approved it

**Chassis `v1.0.1213`**, both pods, started 08:07Z (a roll the owner ran, not me).
Pod-grepped both replicas:

| marker | count |
|---|---|
| `estTextWidth` (a symbol the fix ADDED) | 1 |
| `Capacity headroom against your requirement` (positive control) | 1 |
| `nonexistent_marker_xyz` (negative control) | 0 |

**Marker-choice note worth keeping:** my second candidate marker, the comment *"a clipped bar
must not read"*, greps **0** in the binary — comments do not survive compilation. A **symbol
name** does. Picking a comment as your deploy marker manufactures a false negative and would
have read exactly like "the fix did not ship".

**Council `60d05267-a671-4b98-9b87-6a97e16d78a0`: APPROVED round 1**, *"approved with 2
advisory objection(s) — none high-severity"*, 9 seats abstained on relevance, `architecture`
returned `point_fix` (*"architecture-scope only if a second caller appears; it doesn't yet"*).
All four checkable items discharged the same morning:

- `prior_art` (low) — *single-caller claim, method not shown, and the code index may lag
  HEAD*: re-grepped at HEAD, not via the index. `renderBarChartSVG` has **exactly one**
  caller (`report_charts.go:239`, inside `renderHeadroomChart`).
- `architecture` (missing) — *are there other SVG generators needing the same logic?*
  `git grep -l '<svg viewBox' -- '*.go'` returns **report_charts.go and nothing else.** The
  pattern is contained, measured rather than asserted, so lifting it into a shared helper
  now would be speculative — as the seat itself said.
- `debug_historian` (missing) — *how will you verify at the POD, not the tag?* The table
  above, both replicas, added-symbol plus two controls.
- `debug_historian` (missing) — *the "mutation-verified" claim is the exact shape of the
  mutation landmine; show it ran.* It is in the entry above: three mutants, each paired with
  its test, and the run that mattered is the one that came back **PASS (BAD)** and forced the
  test to be strengthened. A mutation harness whose output is all-green proves nothing.

**`prior_art`'s medium objection was PRESCIENT, and I am the one who falsified it.** It
flagged that my risks section asserted *"report-dispatch is currently disabled, so nothing
regenerates unasked"* — a **live-state** claim with no check behind it, load-bearing for the
whole containment argument, and vulnerable to *"another lane re-enabling it"*. True when
submitted at 19:30Z; false by 22:13Z, enabled **by me**, on owner instruction, two hours
later. **A live-state claim in a submission is perishable, and the seat that says so is not
being pedantic.** Date such claims, or express them as a condition the reader can re-check.

**`bug_historian`'s medium objection is the one that needed an action, not an argument:** the
generator was fixed while two known-corrupted pages stayed served, with nothing queued to
force regeneration — the shape of `bugs_closed/046`. **Answered by FIXTURE 4** (below), which
is also what the owner asked for independently.

---

## 2026-07-31 08:15Z — FIXTURE 4: regenerating the worst case, and the column I did not look at

Queued `4ccc73d7-c467-480f-9a39-0b327b383870`, `request_id`
`bf3765d6-befe-43a8-b1cd-ca5c210f39e9`, `source='manual-test'`,
`item_key='manual-test-fixture-4'`. It re-runs **fixture 1's exact spec** — 2.5 kg steel,
a=12, S=2, IP54 — so the before/after is a comparison of the same inputs rather than of two
different reports. That spec is the one that produced the 6.42× and 7.60× capped bars with
`Insufficient data` verdicts, i.e. **the exact case that clipped**.

**It failed on the first attempt and the cause is a landmine now filed.** `claim_item` picked
it up correctly within 50s, then `spawn_handler` died:

```
failed to execute action spawn_agent: configuration extraction failed in spawn actions:
agent_type is required (provide 'agent_type' or 'agent_type_field')
```

That message names neither the row nor the column, and reads as a defect in the agent's
config. The config is fine — `agent_type_field: "claimed.handler_agent"` resolves against the
**claimed row**, and my hand-built row had `handler_agent` empty. I had copied fixture 1's
shape from a `SELECT` of the columns I *thought* mattered, so the row was missing precisely
what I never looked at. **The check that found it in one query** (now in `LANDMINES.md`):
diff your row against a working one over **all** columns via `to_jsonb(w)` +
`jsonb_object_keys`, rather than reading the error. It named `handler_agent` immediately, and
also that fixture 1 carries an `item_key`.

Re-armed with `handler_agent='report-builder'` (verified `is_active`), `attempt_count` back to
0, error/claim cleared.

**Second attempt: rejected by the honesty gate — and the gate was right to.**
`verify_prose` failed with *"summary_html names model-like token \"IP54-or-better\" not in
the candidate set or fact block"*. `modelNumberRe` classifies the phrase as SKU-shaped, and
the clearance test is verbatim containment in the fact block, which carries `IP54` and not the
writer's composed phrase. The step is **fail-closed**, so no page was composed at all — the
URL 404'd. **Filed as `bugs_open/160`** with four ordered fix candidates; the gate itself must
stay strict, the *classifier inside it* is what cannot tell a fabricated sibling SKU from a
recombined fact. Worst property: **intermittent**, because the trigger is the writer's
phrasing — the identical spec passed this same gate on 07-27.

**Third attempt (max_attempts raised to 2): COMPLETE at 08:25:01Z, 175s.** Justification for
retrying rather than fixing first: the same spec had passed on 07-27, so the violation was
known to be phrasing-dependent rather than a deterministic block.

### The chart fix, verified on the SERVED artefact

`https://robot-hands.com/reports/bf3765d6-befe-43a8-b1cd-ca5c210f39e9.html` — **HTTP 200,
43,546 bytes.** Checks run against the page's own inline SVG, pulled with curl, not against a
local re-render:

| check | result |
|---|---|
| value labels emitted | 10, longest `6.42× (Insufficient data)` |
| labels truncated (ending `…`) | **NONE** — the gutter fitted them |
| reference captions | `requirement (1.0×)` at y=364, `marginal threshold (1.25×)` at y=376 |
| captions on distinct baselines | **true** (the lane stacking fired) |
| capped-bar tips | **2** — the 6.42× and 7.60× bars, the two over the 3× cap |

Then rendered and **looked at** it (`~/chartcheck/fixture4_chart.png`): labels whole,
captions legible on two lines, capped bars visibly pointed. Both defects the owner
screenshotted are gone on a page the pipeline produced by itself.

**Note what the numbers say that the old chart could not:** `7.60×` and `6.42×` are both
clipped to the cap, and the pointed tips now say so. Before, they were two identical bars with
different numbers beside them.

---

## 2026-07-31 — the OWED council debt, discharged: the contracts-gap claim is CONFIRMED, and sharper than I stated it

`prior_art` approved council `721ac4f7` (2026-07-28) but asked that this claim be
**independently verified before anyone cites it as precedent** — *no mechanism declares
action-to-action `collected_data` fields; `input_contract` only fires at the `call_agent`
boundary; `output_contract` has zero readers.* I was the only reader. Measured now, and it
holds — with three corrections that change what the precedent may be used to argue.

| part of the claim | verdict | evidence |
|---|---|---|
| `input_contract` only fires at the `call_agent` boundary | **CONFIRMED** | the sole runtime read is `call_agent.go:988–1011`, and it is gated behind `ParseInputMapping(config)` — so it fires only when a step declares an input mapping |
| `output_contract` has zero readers | **CONFIRMED at runtime, but the wording was too strong** | the `OutputContract` type at `input_contracts/input_mapping.go:52-53` has **no uses anywhere** in `platform/`, `internal/` or `cmd/` — it is a declared type with zero references. But there ARE readers: `validateOutputContract` in the offline `scripts/goscripts/workflow_validator` |
| no mechanism declares action-to-action fields | **CONFIRMED, and stronger than claimed** | **0 of 184** active agents declare `input_contract` *or* `output_contract` (`default_config ? 'input_contract'`) |

**The three sharpenings, which matter more than the confirmation:**

1. **"Zero readers" is false as stated and true as meant.** Say *"no runtime reader"*. The
   offline validator would catch output-contract violations — if anyone ran it.
2. **And nobody does: the validator is wired into nothing.** No makefile target, no
   `.githooks/` reference, no CI. It runs only when a human runs it by hand. It also exists as
   **two near-identical copies** (`workflow_validator/main.go` and `workflow_validator/run/main.go`,
   both carrying `validateOutputContract` at the same line number), which is its own
   duplication smell and not something I chased.
3. **The whole contracts layer is undeclared fleet-wide, not merely narrow.** `input_contract`
   *works* — it validates, at one boundary — but with 0 of 184 agents declaring one, it
   validates nothing in production today. **So this may be cited as "the mechanism exists and
   is inert", and must NOT be cited as "contracts are enforced narrowly"** — there is no
   narrow enforcement, there is none.

Same family as [[zero-adoption-means-read-the-mechanism]]: ~0% adoption measured the
*mechanism*, and reading the gate is what turned "contracts fire narrowly" into "contracts are
declared by nobody and their only checker is unwired". **`prior_art` was right to refuse the
claim as precedent on one reader** — the corrected version supports a different argument than
the original did.

---

## 2026-07-31 — the last `[UNMEASURED]` marker resolved: the bare `.(string)` hunch is REFUTED as stated

`bug_historian`'s surviving advisory (council `721ac4f7`) suspected that bare `.(string)`
assertions **on LLM-parsed maps** recur elsewhere unaudited. I had already shown its *premise*
was wrong (`bugs_closed/076` is a truncation-tolerance mechanism with no shared safe-extraction
helper, and its headline "113 call sites" is retracted by that file itself to "37 of 118"), but
left the underlying question marked `[UNMEASURED]` rather than guess. Counted now.

**1,734** `.(string)` occurrences across `platform/`, `internal/`, `cmd/`. Of those, **40**
survived a two-value-form filter, and of those 40:

| category | n | why it is not the claim |
|---|---|---|
| `x, _ = f.(string)` — the SAFE comma-blank form | 12 | my filter missed them because an index expression (`inputFields[i]`) precedes the comma |
| comments, not code | 4 | and **two of them document a bare `.(string)` being REMOVED** — `thunder_ssh_exec_dispatch.go:234`, `verify_report_prose_action.go:357`. The codebase already hardened this pattern where it bit |
| **genuinely bare and panicking** | **24** | of which… |
| ↳ `r.Context().Value("user_id").(string)` in `auth-service/project/handlers.go` | 10 | middleware-injected, not LLM |
| ↳ startup/DB/step config (`cmd/agent-chassis`, `evolution.go:152`, `ai_actions.go:1099`, `core-manager`, `contentcreator`) | ~10 | config and DB rows, not LLM |
| ↳ plausibly agent/LLM-derived | **2** | `v3_site_actions.go:3424` (`result["review_mode"]`) and `:3721` (`comp["name"]`) — **both inside `zap.String(...)`**, so a panic would happen while *logging* |

**Verdict: REFUTED as stated.** Bare assertions do recur (24), but the specific claim — on
**LLM-parsed** maps — does not hold. The dominant cluster is request-context reads in one
service, and only 2 of 24 arguably touch model output. **No bug filed**, per the standing rule
that a count comes before a claim.

**Residue worth someone's attention, and it is a different risk class:** the 10
`r.Context().Value(...).(string)` assertions in `auth-service/project/handlers.go` panic the
handler if the auth middleware is ever reordered or bypassed on that route. That is a real
fragility, it is **not** what `bug_historian` asked about, and I am not filing it as one on the
strength of a grep.

**Method limits, stated because the number is the point:** this is line-based, so it cannot see
a multi-line assertion or one made inside a helper. My **first** count was also 40 but for the
wrong reasons — the exclusion list was a guess at variable names (`, ok`, `, _`, `, found`) and
let `loopVarName, hasVarName := …` through; re-run with a general two-value pattern. Two
filter mistakes, same total, different membership — [[two-blind-checks-agree-with-each-other]]
nearly happened to me here.
