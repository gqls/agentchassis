# NOTES — component instance scope (`bugs_open/283`)

Append-only, newest at the bottom. The missteps are not an appendix — they are the point.

---

## 2026-08-15 (session 1, reconstructed from the case file and commits)

Filed out of the LMC Track B2 lane, where the reuse demonstration could not be done on a real page
and had to run on a throwaway one. Owner ruled A+C: *"if we chose to list all the calculators on one
page we'd hope it would work"* — reuse should be a genuine property of the platform.

Built: `{{.InstanceID}}` on three render paths, `DetectInstanceCollisions`, and an opt-in guard
(records everywhere, refuses only when armed). Council returned **REVISE**.

Misstep recorded that day: **"7 live components depend on `{{.ComponentID}}`" was false — it is 5.**
Caught by the `prior_art_librarian` seat objecting that the claim was unevidenced. The seat did not
know the number was wrong, only that it was unchecked, and that was enough.

---

## 2026-08-16 (session 2)

### The deploy question the previous session left open

The handoff said deploy status was unverified and warned against guessing. Both recipes it had
tried were dead ends: the `build provenance` startup line had scrolled out of `--tail=20000` on
both chassis pods, and a binary probe can only confirm a **guess**, because the binary carries its
own build stamp and not its ancestors. (The previous session's first probe used all-zeros as the
absent-control and read PRESENT — binaries are full of zero bytes.)

What worked: **pod `imageID` digest → local image `RepoDigests` → that image's
`org.opencontainers.image.revision` label → `git merge-base --is-ancestor`.** The digest match is
the load-bearing step; without it a local tag is just a local tag, rebuildable by any session.
`v1.0.1303` was built from `5e075a6f9`, of which all four 08-15 commits are ancestors. So the first
half had been live for ~15 hours while the handoff said "unverified".

### Design: settling `data_uuid` vs function+occurrence

The case file left this as "the single most consequential open question". It resolved quickly once
the question was framed as *what does a selector have to know*, rather than *which is more unique*:

- `position` — unique per page, **differs across pages** (LMC tool slot: position 0 on 7 pages, 1 on
  16). Couples every selector to section order.
- `data_uuid` — 1,580/1,580 distinct, uniqueness **provable**. Per-row means per-page, and opaque.
  Strictly worse for selectors.
- **function + occurrence** — `c-mortgages-repayment`. Same on every page for a single-instance
  component; unique within a page; legible. Uniqueness is *derived*, and the detector pays for that.

The check that made the change free: **0 of 243 active templates reference `{{.InstanceID}}`.** It
could have come out non-zero (another session converting a template), which is what makes it
evidence rather than reassurance.

### MISSTEP 1 — I censused "every call site" over the directories I expected

To answer the council's "the mechanism is left generic", I ran
`grep -rn 'RenderTemplate(' --include=*.go platform/ internal/` → 11 sites in 7 files, and wrote
**"eleven non-test call sites"** into three places (two Go comments and the new check's header).

It is **14 across 8 files**. The eighth is `cmd/component-render-check/rendercheck.go`, which
renders every active component through the production entry point — exactly the caller the census
existed to find. I never looked at `cmd/`.

Caught by accident: I broadened the new pattern-check's regex for an *unrelated* reason (it keyed
on the argument's *name*, `htmlTemplate`, so a new call site passing `tpl` would slip), and ran the
broadened version over `git ls-files '*.go'`. The eighth file appeared immediately.

**The cheap check:** census with the tool that has no scope argument. Logged in `WRONG_CALLS.md`.

### MISSTEP 2 — my first version of the durable control had the same defect it was written to catch

`COMPONENT_RENDER_RE` initially required the template argument to be spelled
`HTMLTemplate|htmlTemplate|headTemplate`. That is a lint that catches today's call sites and misses
tomorrow's — the same staleness, one rung down, in the mechanism written *because* enumeration goes
stale. Broadened to match the **call** (`\bRenderTemplate\w*\s*\(`); measured cost: zero, because
exactly eight non-test files call any such helper and all are bound or allow-listed.

### MISSTEP 3 — a structured census answered a narrower question than I asked, in the same format

Answering the council `guardian` seat's "your census is Go-file-level, not pipeline-level", I
queried `agent_definitions` for workflow steps naming each edited action. `render_component`
returned **0** — i.e. a whole render path executed by no live workflow.

That is false. A text control (`default_config::text LIKE '%render_component%'`) found
`page-content-writer`. The action runs inside `process_sections_loop`, at
`config.sub_workflow.steps.render_section.action` — invisible to a `steps.*.action` census. My
first attempt to look inside the loop guessed the key one level too shallow (`config ? 'action'`)
and returned 0, which reads like a completed check.

Fleet-wide the blind spot hides **80 invocations across 19 action names**. Landmine written.

**Without the control I would have published "that render path is dormant" in a council
submission.** The general form: a structured query encodes a belief about the document's shape, and
when the shape is wrong it does not error — it answers a narrower question, correctly, in the
format you asked for.

### MISSTEP 4 — I asserted a follow-up was "filed" when nothing was

Round 2's submission and the register both said the `ComponentID` unification was "filed as the
follow-up the architecture seat asked for". The `reuse_agent` seat objected at medium severity that
no work item, doc_plan or artifact id appeared anywhere — *"a vague 'filed' claim with no locator is
not evidence."*

Correct. What existed was a `verify-later` line and a sentence in a commit message. **`RFC_032` now
exists**, and it carries the architecture seat's own ask: the RFC_022 exception this was approved
under expires the moment a live template references `{{.InstanceID}}`, and the mechanical trigger
for that expiry is ~~**not built** — a commit-time lint cannot see a template stored in a DB
column~~ **BUILT AND RUNNING, later the same day — see the entry at the bottom of this file. The
"not built" was a cost estimate dressed as a decision.**

### Not a misstep, but worth recording: my LANDMINES edit was swept

Another session's commit (`fcb0e93e0`) took my uncommitted `LANDMINES.md` edit along with their
own. Nothing lost — forward-only holds — but it also consumed the "new entry" status, so
`landmines-verify-dispatch.sh` armed *their* entry and not mine. Recovered with
`trigger-landmine-verifier.sh` and the slug read out of `doc_notes`.

### Verified rather than assumed, on the council's request

- `RenderTemplate` is a three-line wrapper over `RenderTemplateReportingMissing`
  (`component_library.go:952`) — so the new report reaches every call site, including the plain
  `RenderTemplate` ones. The seat was right that the sketch never showed this.
- The carried-section branch `continue`s at `rerender_page_sections_action.go:399`, before the
  counter advances at `:511` — so "the counter does not advance for a carried section" holds.
- `slot_name_from` keeps **two pre-existing readers** (`v3_site_actions.go:2119`, `:2177`), so
  deleting the instance-token derivation added yesterday removed no live behaviour. Checked
  because deleting the only reader of a live config key would have been a silent regression.

### Outcome

Council **APPROVED** round 2 (6 advisory objections, none high-severity), and the code went live on
chassis `v1.0.1304` the same morning. Still inert: 0 of 243 templates reference the token.

### Later the same day — building the expiry tripwire, and a reclassification

`RFC_032` §6 and the case file both said the RFC_022 expiry trigger was "not built", and assigned it
to whoever converts the first template. That was a cost estimate dressed as a decision. The estate
already has **twelve** daily check CronJobs, two of them (`single-owner-carriers-check`,
`optional-key-budget-check`) built for exactly this reason — *"the number this check watches changes
by a route no commit can carry"* — so the marginal cost was four files, not a project.

`instance-token-adoption-check` is now live at 07:40 UTC. Three design points worth carrying:

1. **The seat suggested a pattern-check finding, and that could not have worked.** A component's
   `html_template` is a database column written by the component-creator agent, by hand-authored
   SQL, by migrations and by the admin UI — four routes, none through a commit. Taking the seat's
   *intent* (a mechanical trigger) rather than its *example* (a lint) was the whole design.
2. **A check whose healthy answer is ZERO needs a demand control inside the check.** A broken query,
   a mis-escaped `LIKE` and an empty table all return zero. So every run also counts
   `{{.ComponentID}}` through the same `LIKE` in the same statement, and the job **refuses** if that
   returns 0 rather than reporting a zero it has not earned. Live: adopters 0, control 5, active 243.
3. **The polarity is inverted vs every sibling, so the report says so in its own words.** A trip
   means an owed architecture review, not a defect — and the person who reads a failed Job at 08:00
   is not the person who built it. Writing "THIS IS NOT A DEFECT REPORT. Do NOT 'fix' this by
   reverting a conversion" into the output itself is cheaper than hoping they find the docs.

Verified rather than assumed, before trusting it: `kubectl kustomize` builds; **both** report
branches exercised via `--stdin` with exit codes checked (the tripped branch is the one that never
runs in practice, so it is the one that rots); the census SQL run against the live DB returns the
exact shape the script parses; and then an actual Job run in-cluster — `SUCCEEDED 1`, the quiet
branch printed, the `doc_notes` row present at 15:29:24 UTC.

**Not done, and it is a shared-tree consequence rather than an oversight:** no `deploy-…` makefile
target, because `makefile` carried another session's uncommitted change and a pathspec commit takes
same-file passengers. Recorded in the register's `verify-later` and in the CONTINUE_HERE.

---

## 2026-08-17 (session 3) — scoping the conversion on the owner's go-ahead

### State check first, because the owner asked for one and it was the right call

120 commits had landed since my last. Three touched files I own — two on `v3_site_actions.go`
(the 282 lane) and one adding 50 lines to `pattern-check.py`. **Nothing disturbed this lane:** all
four binding sites intact (1/2/2/2), `check_unscoped_component_render` still registered, both 283
files present. `v1.0.1305` was built from `6a782274b`, digest-matched to the running pods, and every
283 commit is an ancestor — so the seam is still live.

The tripwire had run **unattended** at 07:40 UTC and caught real drift: active components
243 → **244**. That is the mechanism doing its job on its own clock, and it is also a reminder that
any figure carried from yesterday is already stale.

### MISSTEP 5 — I measured the blast radius at the wrong UNIT, and it flattered me

Cross-site sharing, measured by `function`: *"4 components shared across up to 5 domains"*, with
`tool-llm-cost-calculator` at 5 domains. I nearly wrote that into `RFC_034`.

It is wrong. Those four functions are **exactly the four that carry forks** — several active
`content_components` rows under one `function` name — so grouping by function merged several
single-domain rows into one apparently-widely-shared component. Re-measured at the ROW level, which
is the unit that actually gets converted: **1 row on 2 domains, 3 rows on 2 pages.**

Two consequences, and the second is the one that would have cost something:

- the blast radius is *smaller* than I was about to publish, so the error was in the direction that
  makes a plan look scarier rather than safer — but it is still a wrong number in an RFC;
- **a `function`-keyed conversion silently skips 9 forked rows.** The wrong unit was not just a
  reporting error; it would have been a conversion bug.

Caught by noticing that two counts of "the same thing" disagreed (83 functions vs 88 vs 91 rows) and
chasing the discrepancy instead of picking one. **The general form: when a census can be taken at
more than one grain, the grain IS the finding — state it in the row label, and re-derive at the
grain the WORK happens at, not the grain the query was easy at.**

### The finding that changes the plan, and why I proved it instead of asserting it

"Namespace the ids" is the obvious half of the conversion, and it is the half a regex can do. I
believed — and could have written — that it was therefore a safe first phase.

It is not, and the reason is collision class 2: both instances still declare `function runCalc()` at
top level, the second declaration replaces the first, and every instance's `onclick="runCalc()"`
resolves to the survivor. So an id-only conversion gives a page with **zero duplicate ids** where
every button runs the last instance's logic. It passes the check while producing the wrong answer.

`TestIDOnlyConversion_readsCleanOnIDsAndIsStillBroken` asserts exactly that (0 duplicates, 2
surviving `window.onload`, 2 surviving global scripts), and mutating the fixture to be script-scoped
fails it, so it is sensitive to what it claims. **This is the difference between a plan that
sequences ids-then-scripts and one that does not — worth the twenty lines.**

And a smaller one found by reading a real template rather than imagining it: `{{.InstanceID}}`
renders as `c-mortgages-repayment`, so the obvious de-collision `function runCalc_{{.InstanceID}}()`
is a **syntax error**. The IIFE route is forced, not chosen. Asserted in a test so a converter author
meets it there rather than on a shipped page.

### Where it stopped, deliberately

`RFC_034` filed; nothing converted. The shape question (deterministic fix_type vs LLM rewrite vs
hybrid) is the owner's, and it decides whether 94 live pages change over days or weeks. Building a
converter before that decision would be building for a shape that may not be chosen.

### MISSTEP 6 — I sized the job with three regexes and was about to publish "~30 of 91"

Asked to choose a conversion shape, I triaged the 91 templates with three patterns —
`window.onload`, inline `on*=`, and a `function` keyword within the first 200 characters of a
`<script>`. That gave **24 need judgement, 67 mechanical**, and I wrote "the ~30 rows whose scripts
need judgement" into `RFC_034` §4 as the basis of the recommended option.

**The real classifier says 88 of 91.**

> **CORRECTED same day (see MISSTEP 7 below): the 88 was the classifier's own defect** — a 70%
> false-flag rate from an anchored wrapper regex that could not see past a leading `/* tool-doc */`
> comment. The corrected figure is **25**, and the regex triage this misstep condemned had been
> within one of the truth. The misstep recorded here is still real — the proxy *was* the wrong
> instrument — but its punchline number was not.

What caught it, before the owner acted on it: the triage query's own last line said all 67
"mechanical" rows still contained an inline `<script>`. A component with inline JavaScript and *no*
scoping problem is possible but unusual, and 67 of them was not credible. **Two numbers I had put
side by side did not fit together, and following that rather than picking one was the whole catch.**

The fix was not a better regex. `DetectInstanceCollisions` — the detector this lane built in
session 1, which parses script bodies and which will gate every conversion — had never been run
over the live corpus, only against test fixtures. `cmd/instanceaudit` now does exactly that:

| | rows |
|---|---|
| already scoped | **3** |
| declare into global scope | **88** |
| assign `window.onload` | 8 |
| duplicate ids if doubled | **91 of 91**, 1,345 ids total |

**The general form, and it is the sharpest version of this lane's recurring lesson:** a
hand-written triage is a *second implementation* of a judgement the estate already implements. It
will disagree with the real one, silently, and in whichever direction its author's patterns happen
to lean — here, the direction that made the plan look cheap and got it as far as a filed RFC.
**When the acceptance gate already exists, size the work with the gate.**

Corroboration worth noting because it was free: the detector counts 1,345 duplicate ids when each
template is doubled; the independent SQL census counts 1,346 literal `id=` attributes. Two code
paths, agreeing within one — neither was written to check the other.

### MISSTEP 7 — "size the work with the gate" was followed, and produced a worse number than the regex it condemned, because the gate itself was broken

The owner asked me to look the analysis over once more before deciding on it. Three things fell out,
in ascending order of consequence.

**The off-by-one was a finding, not noise.** The detector's 1,345 doubled-duplicate ids against the
SQL census's 1,346 attributes resolves to one template — `tool-spawn-rate-balancer` — carrying
`id="chartTitle"` twice within a single copy: once in markup, once in a JS string that rebuilds the
same SVG `<title>`. Nothing binds it (`getElementById` uses: 0); it is an aria-labelledby target on
one gamesdesign.co.uk page. Benign, but "agreeing within one" is now *explained* rather than waved
at, which is the difference between corroboration and coincidence.

**The 88 was the DETECTOR's false-flag rate.** Sampling one flagged template
(`tool-css-unit-converter`) showed a body that visibly ended in `})();` — an IIFE, hidden from the
anchored wrapper regex by the estate's conventional leading `/* tool-doc */` comment. Measured
properly: **62 of the 88 flags were false, a 70% rate.** The corrected corpus is **66 scoped / 25
genuinely global**, and the 25 are the 23 LMC calculators plus two tools — the original "22
templates" scope, rediscovered from the other direction. Detector fixed (`stripLeadingJSComments`,
leading comments only, launder-control in the test, both mutations killed), council round 3
submitted on the same correlation, `5b30a831b`.

**My cross-check lost to the fixed detector exactly as MISSTEP 6 predicted.** The Python depth-walk
I used to expose the 88 said 65/26; the fixed detector says 66/25; the disagreement is
`tool-css-specificity-calculator`, whose JS is full of regex literals that unbalanced the crude
walk. It is a comment-led IIFE; the production detector is right. I built a second implementation
to check the first one, and the second implementation was wrong in precisely the way the lane's own
lesson says second implementations are.

The composed lesson, stated once: **"size with the gate" and "sample the gate's flags" are one
practice, not two.** A gate run over a corpus it has never been sampled against is a first run of
an instrument, not a measurement. The full chain 24 → 88 → 25 is in `WRONG_CALLS.md` (two entries,
the second correcting the first) and `RFC_034` §3a, history kept.

**Why the fix ships now rather than with the programme:** RFC_034's plan gates every conversion on
this detector, and the unfixed gate would have refused 62 correct results mid-programme — the
moment at which someone either "fixes" unbroken components or relaxes the gate in a hurry. It also
matters for the guard: once `enforce_instance_scope` is armed, a false UNSCOPED refuses renders of
correct components.

---

## 2026-08-17 (session 3, continued) — ruling, build, and the fixture that earned its keep

**Owner ruled** after the twice-corrected numbers were in front of them: hybrid, LMC first, and —
the addition that shaped the build — **through the framework**. Recorded in RFC_034's header.

**Round 3 approved clean** ("all reviewers approve", 12:43 UTC) — the detector fix. The lane's
council arc: REVISE → approved with 6 advisories → approved clean.

**The converter was built** (`b7b396cb3`, CLC-017): transform + gate + the `fix_component_template`
seam, with the §2.1 refusal mechanical — `needs_script_scoping`, nothing written. Census updated
(third html_template write in `component_write_guard.go` + `fanOutIntendedWriters`).

### The live fixture caught what no composed test could — before the code ever ran in anger

Pinning `tool-css-unit-converter`'s real bytes as the happy-path fixture immediately failed the
surface-count assertion: 11 literal `getElementById` calls, not 12. The twelfth is
`getElementById(targetId)` — a variable, fed at runtime from `data-target="result-px"` on five
copy buttons. **A conversion that renamed the ids but not the `data-target` values would have
shipped five silently-dangling buttons on every converted page carrying that pattern.** No pass I
had designed touched `data-*`; no fixture I would have composed contained the pattern. The pass
now exists (exact-match values only — a value merely *containing* an id is the documented
concatenation limit), and the fixture asserts `data-target="{{.InstanceID}}-result-px"`.

The general form is the lane's oldest lesson pointed at test design: **a fixture you compose
exercises your own belief about the artefact; only pinned live bytes can disagree with you.** This
is the second time in two days real bytes overturned a design assumption (the tool-doc comment was
the first).

### ⚠ The 14:43 "fresh build" shipped nothing — same-tag cache trap, observed live

Pods restarted 14:43 UTC, still on yesterday's digest `f90a7e88…` (revision `6a782274b`). The
14:30 local rebuild (`89a0cbeb7`) contains the detector fix and converter, was pushed — under the
SAME tag `v1.0.1305` — and the restart served the node's cached image. Another lane measured the
same trap the same day (203 commits unshipped; it is now a MEMORY banner). Nothing to do but say
it: **the fix rides the next roll under a bumped tag.** Digest equality is the only check that
sees this; the tag, the restart time, and even the local label all read "fresh".

---

## 2026-08-18 (session 4) — the canary's product was a fleet-wide finding, not a converted page

Full narrative in case file §13; the technical residue worth keeping:

**The finding.** A perfect conversion + a green rerender + a successful deploy still served the OLD
bytes. `check_rerender_mode` routes to the sections path only for three reasons; "template changed"
was not a reason. Every template fix on this estate has been assemble-only. The precise prediction
is what made it visible — nobody had checked a served page against expected bytes this exactly.

**The fix (460+461, live, round 4 submitted with FORCE=1 — the path filter cannot see that a config
migration is a platform change).** `template_changed` in the vocabulary; the fixer files PAGE-SCOPED
reason-carrying rerenders (site_id from the PAGE, covering the cross-domain row).

**MISSTEP 8 — I shipped a nonexistent column inside a query string, and my probe run could not see
it.** 460's embedded query referenced `p.filename`; `pages` has no such column (Go derives it).
The probe run proved the OUTER update; the inner query is DATA until the step executes. Caught
minutes later running the same shape directly. 461 corrects it and adds the check class 460
lacked: **PREPARE-compile the embedded query in the verify block.** Landmine written (footprint:
every `query_database` config, every `scheduled_tasks.pre_query`).

**Canary end-state, measured at the served artefact:** 34 instance-scoped ids, 0 old, 0 unrendered
tokens, 5/5 `data-target` chains paired. Tripwire tripped 07:40 as designed (adopters 1).

**Batch released:** 70 items (eligible count 66→69→70 over two days — derive, never paste).
Monitor `budmv5g2d` on the drain.

## 2026-08-18 (session 5, evening) — the judged pipeline is DESIGNED; every load-bearing input re-measured

Deliverable: `PLAN_2026-08-18_judged_pipeline.md`. Session opened with RUNBOOK §1 (pods on
`v1.0.1310`, digest-matched, revision `0b185bad2` = ancestor of HEAD — clean). No code written;
this was the design session the continue file asked for.

**The two measurements that changed the design from what §12.5 sketched:**

1. **All 23 LMC calculators sit on `rebuild_policy='owned'` pages** (24 placements;
   `mortgages-repayment` on 2). So the mechanical programme's proven delivery
   (`template_changed` rerender) covers only the 2 generic-page tools — mig 462 excludes owned
   pages BY DESIGN. LMC delivery = section-editor (`apply_section_edit`), which binds the
   instance token (verified at `section_editor_actions.go:850/:948`, occurrence 0 — correct,
   since no page places any of the 25 twice; 26 distinct component×page pairs).
2. **22 of the placed 25 are `component_level='section'`** — so tool-improver's fenced write
   (`sharedComponentWriteCheck`) would REFUSE `mortgages-repayment` (section-level, 2 pages).
   That killed the "reuse tool-improver wholesale" option: the judged writer must be a
   fan-out-intended sibling of `scope_component_instance` inside
   `fix_component_template_action.go`, declared in `component_template_writer_coverage_test.go`.

**Shape chosen:** extend `component-template-fixer` — the mechanical arm's
`needs_script_scoping` refusal IS the router (it fired correctly on all 25 in the batch), same
`instance_scope_conversion` seed, same `component_versions` audit trail
(`change_source='scope_component_instance_judged'`). LLM step gets the IDS-CONVERTED template
and a deliberately narrow brief (IIFE + rewire the inventoried `on*=` handlers + replace
`window.onload`, change nothing else); the gate re-derives the ids-converted baseline itself
and refuses on: GateConvertedTemplate not fully clean, markup-parity-outside-script broken,
id-set drift, or the comparative collapse guard. Refusal = `needs_human_review`, nothing
written, no auto-retry in v1.

**Truncation defence is layered, not single:** `execute_llm_prompt` already hard-refuses capped
completions unless a step opts into `tolerate_truncation` (`ai_actions.go:409–517`) — the judged
step will NOT opt in; the write path's comparative collapse guard is the second layer; id-set
parity is the third (a cut template loses ids).

**[UNVERIFIED], named in the plan and proven by the canary:** that a `section_edit` item with
empty `field_updates` re-renders the slot from the converted template AND deploys, on an LMC
owned page specifically. tool-improver's 42 completes make it likely; nobody has watched it
end-to-end there.

**Cross-lane:** CONTRIB filed in the LMC lane dir (b2_verify byte-identity ends at conversion;
oracle selectors move in lockstep per tool, same commit as each conversion's verification;
canary `loans-standard-calc` with a veto window). webdesign rebuild lane unaffected (none of
the 25 is theirs); copy_quality lane unaffected (links, not ids).

**Out of scope, unchanged:** `ec2` rename and `chartTitle` repair are pre-repairs for the
MECHANICAL path (the judged prepare would hit the same hex refusal — routing them to judged
would not help); forked-function shrink investigation; the companion producers item; RFC_032.

### 2026-08-19 — LMC lane's answer to the veto window (from the D6 planner lane)

**NO VETO on `loans-standard-calc`.** Full reply, with the measurements:
`CONTRIB_2026-08-19_from_LMC_lane_no_veto_plus_one_gap_in_the_assurance_argument.md` in this dir.
Three things from it that bear on your next session, shortest first:

1. **You go first.** The LMC lane is holding its D6 planner round 2 until your canary is landed
   and verified, because `oracle.py` is the pass/fail for both of us and a red oracle with two
   concurrent authors is ambiguous. Say here when the canary is verified and they will pick up
   after. They will not fire a `build-site-planner` run meanwhile — if you see one, it is not them.
2. **⚠ THE ORACLE PROVES 23 OF 24 ARITHMETIC CALCULATORS.** `tool-overpayment-priority`
   (tool-level, built by the improvement loop 08-15) does real amortisation and is **not** in
   `oracle.py`'s hand-authored 18-key dict. If it is in your judged set, the instrument that
   licenses LMC-first cannot cover that page — exclude it, or ask them to extend the oracle
   first (it is on their list, unscheduled; your need would move it up).
3. Census confirmed: **23 section-level owned components on 23 owned pages** — exactly your
   figure. The site also has **3** tool-level components, not 2: censusing with
   `sections::text LIKE '%tool-%'` silently drops `tool-overpayment-priority`.
   `loans-standard-calc` itself: 3 active components, 0 locked, and 9 of the 170 oracle checks.

## 2026-08-19 (session 6) — judged pipeline BUILT; and reading one canary script found bugs_open/324: the batch shipped 32/69 templates with dangling bindings, 14 serving

Deploy verified first (RUNBOOK §1): pods on v1.0.1314, digest-matched, revision d3590ca46 —
carries the design commit; docs were current.

**MISSTEP 9 (mine, logged in WRONG_CALLS): the PLAN's "the refusal fired correctly on all 25
during the batch" was FALSE** — the batch never contained the 25; the classification was the
detector's. Corrected visibly in the PLAN §2.

**The finding.** Reading loans-standard-calc's real bytes to write the LLM prompt: its script
binds ids from `const inputs = ['amount','interest','years']` → `getElementById(id)`. The
converter renames `id="x"` and literal `getElementById('x')` and asserts completeness by
grepping for those SAME forms — so the array literal survives, the lookup dangles, and every
batch check reads green. Censused all 69 converted rows two independent ways (python sweep +
a purpose-built Go detector; agreed 32/32, the Go one also catching a composition hazard):
**32 dirty, 27 mechanically repairable, 5 judged; 14 rows / 15 placements SERVING the broken
bytes** (verified at robot-hands.com's served HTML). Full mechanism + classes: bugs_open/324.
016b §9 pattern written ("a renamer's completeness check that greps for the forms it renames");
LANDMINES entry appended + verifier armed (dispatch needed one retry after a kubectl stream
EOF; second run: Dispatched 1, 0 failed).

**Built (one commit, council round 6 submitted, same correlation):**
- `component_instance_bindings.go`: pass 5 (classes A/B/C) + `UnprefixedBindings` (incl.
  composition hazards) + `RepairConvertedTemplateBindings`. Refuse-contexts (comparison, case
  label, object key, computed access — the last is why automation-savings is judged: a
  `values['staff-count']` read must match the now-prefixed `values[field.id]` write) are
  skipped and REPORTED. Detector wired INTO `GateConvertedTemplate`.
- `component_instance_judged.go`: `JudgedConversionIssues` — two-instance gate fully clean,
  markup parity outside script bodies (expectation derived from the baseline), id-set parity,
  no unprefixed bindings, no surviving inline handlers.
- `fix_component_template_action.go`: arms `scope_component_instance_judged` (gate+write
  fused; converges to mechanical if the row was scoped between steps) and
  `repair_instance_scope_bindings`; mechanical refusal result now carries the ids-converted
  template + handler inventory + unplaced bindings; `writeScopedTemplate` shared.
- `cmd/instanceaudit --bindings`: census + done-check (exit 3 while anything dangles
  post-repair). Baseline run archived: scratchpad/bindings_audit3.txt.
- Migration `486_judged_instance_scope_pipeline_HOLD.sql` (six workflow steps + rewiring +
  PREPARE-compiled delivery query + tolerate_truncation-absent assertion) and seed
  `487_seed_bindings_repair_HOLD.sql` (derived at apply time; serving-broken at priority 30).
  Both _HOLD: they dispatch fix_types that exist only post-roll. ⚠ numbers 484/485 were taken
  by two other lanes BETWEEN listing and writing — renumbered, references fixed.

**Verification discipline:** full actions suite green on git-archive-HEAD + my files only
(the tree carries another session's broken datahelpers WIP — not touched); fixtures = live
bytes for all three classes + one pre-conversion snapshot from component_versions; mutation
controls both directions at transform, detector AND wiring (pass 5 disabled → fresh-conversion
test fails; report silenced → refuse-context test fails). pattern-check: only two other-lane
gofmt nits.

**Register:** CLC-021 (bindings) + CLC-022 (judged pipeline) added; CLC-017 status corrected
(its "failure direction is the design" paragraph believed refusal covered this class — it
did not). Index rows added after CLC-019 (CLC-020's index row is the 311 lane's to add).

**Execution order next session:** roll (digest-verify, ancestry of THIS commit) → rename+apply
486 → rename+apply 487 → drain (27 fixed / 5 needs_human_review expected; monitor) →
`--bindings` exit 0 → per-page spot-checks with a BINDING check this time → then the LMC
judged sequence (owed steps, canary, 22, generic pair + the 5).

## 2026-08-19 (session 6, later) — round 6 REVISE, all objections answered by measurement, round 7 SUBMITTED

Round 6: REVISE (guardian gating; 8 object / 4 approve). Every objection checked against the
live system rather than defended: only ONE workflow dispatches fix_component_template; fixer
has NO root ai_service; 69 pre-images exist (change_source='scope_component_instance' — the
"no backup" concern measured false); 19 section_edit completions on OWNED pages (the refusal
landmine covers save_page_sections' generic save, not the targeted section-editor);
applyContentEdit (section_editor_actions.go:793-850) re-renders even on an EMPTY field_updates
map — the owned-delivery [UNVERIFIED] narrows to runtime-only, canary-first; create_rerender
carries reason='template_changed' with no site filter (cross-site placements covered); LMC
locked sections = 0. Two real hardenings applied to 487: an enforced DO-block precondition
(RAISES unless 486's steps are live) and the /proc/1/exe symbol grep with present+absent
controls in the header. Edit 7 re-filed as config_change naming component-template-fixer.
Round 7 submitted, same correlation. Open architecture note recorded for a human (bug_historian
+ architecture, converging): the judged writer and tool-improver are now TWO paths authorised
to rewrite shared templates — one shared fan-out-safety gate, or two maintained arguments?

## 2026-08-19 (session 6, later still) — roll v1.0.1315 VERIFIED (carries the 324 fix); round 7 REVISE produced a REAL catch; gate hardened; round 8 submitted

Roll verified three ways: digest match, revision `590ca3a20` has `ffef54338` as ancestor, and
the pod binary itself (grep -ac `repair_instance_scope_bindings` /proc/1/exe → 5; absent
control → 0). The repair machinery is LIVE-CAPABLE; 486/487 stay _HOLD pending the verdict
AND the next roll (see below).

**Round 7: REVISE, 10 approve / 3 object — and the gating HIGH was RIGHT.** bug_historian:
markup parity excludes script bodies by design, so what asserts the script still EXISTS?
Answer before fixing: NOTHING — control run on the 22,791 B affordability fixture: every
inline script body replaced by an empty IIFE retains 58% (over the 50% collapse floor) and
drew ZERO issues from the judged gate and the write guard. The exact blindness class 324 is
about, in my own gate, caught by the council before it could ship. Closed (`35d2e0f9c`):
check 5a prefixed script-reference parity (the brief only ADDS references), 5b comparative
script-mass floor 0.7 ([REASONED, not measured] — recalibrate at the canary). Regression pair
pinned: gutted refuses on both; a single dropped lookup refuses on parity; identity trips
neither. Full suite green on the archive overlay.

**The rest:** compositionHazards confirmed BLOCKING at the code (UnprefixedBindings → gate
refusal, both paths; fuel-budget's JUDGED classification is its live proof); 486 gained
double-apply guards BOTH directions (a hand-reapply would snapshot the patched row as a
"pre-image", poisoning newest-first rollback — debug_historian); prior_art's asked-for SQL
evidence attached verbatim to round 8 (19 owned-page section_edit completions; 69 pre-images;
the pod probes).

**⚠ Sequencing consequence: 486 must NOT be applied until a roll carries `35d2e0f9c`.** On
v1.0.1315 the judged gate lacks 5a/5b — un-holding 486 there would let the 5 refusal-routed
rows through an LLM whose gutted output could pass (and fuel-budget/loot-table currently serve
WORKING pre-conversion pages that a bad judged write + owned delivery could break). The
CONTINUE file's step 1 ancestry check must name `35d2e0f9c`, not `ffef54338`.

## 2026-08-19 (session 6, round 8→9) — the REVISE loop keeps finding real things; the shared guard is now stub-aware, CALIBRATED

Round 8 REVISE (gating HIGH, bug_historian, 10 approve / 3 object): my round-7 control proved
BOTH guards blind and I had patched only the judged gate — componentRegressionIssues (which
also guards tool-improver's update_component_html; caller grep: exactly 3 sites) stayed
gut-blind. Closed at the SHARED guard (`79fa79fb4`): `scriptStubRegression` — elements kept,
every inline (non-src) body <200 B while current carries ≥1000 B inline. **Calibrated per the
guard file's own rule, and the calibration EARNED its keep:** the naive <30%-mass variant
flagged two REAL legitimate shapes — the js-extraction pattern (heat-rater: body moved behind
src=, empty by design) and the arena v3→v4 rework (a 7,189 B clean program at 29.7%) — the
refined signature matches 0 of 235 historical transitions and still catches the gutted
control. Both negative shapes pinned in `TestScriptStubRegression_calibratedShapes`.

The 0.7 judged floor is now MEASURED against the pool (all 30 rows): worst script
comment+whitespace fraction 46.5% (mortgages-stamp-duty) — so a brief-violating tidy could
refuse to a human at 0.54 (chosen direction, kept), and the honest statement is in the
constant's comment: no structural check sees partial logic deletion that keeps its lookups;
the ORACLE does, which is why LMC-first. Round 9 submitted (edits consolidated to the 8-cap —
the first two attempts were refused for a 10-edit array and submitted NOTHING; check the
trigger's tail, not just its exit). Fleet state pinned in the submission: v1.0.1315 carries
round-6 code (symbol probe 5/0), NOT `35d2e0f9c`/`79fa79fb4` — 486/487 stay HOLD until a roll
carries BOTH hardening commits.

## 2026-08-19 (session 6, close) — ROUND 9 APPROVED (15 seats, 2 abstained). The trail on `07635a2f…` closes for the second time; advisory triage below

Verdict READ (15:53). The arc this correlation now carries: REVISE → APPROVED → APPROVED →
REVISE → APPROVED (r5, mechanical close) → REVISE → REVISE → REVISE → **APPROVED (r9)** — and
rounds 7 and 8 each found a REAL defect in my own gates (the gutted-script blindness, then the
same blindness in the SHARED guard). The REVISE loop earned its cost twice over.

**Advisory triage (none gating; each checked before recording):**
- *editquality bundling ×2*: artefact of the trigger's 8-edit cap (the 10-edit array was
  refused outright); the ROLLBACK + coverage-test files ARE in the commits. No action.
- *reuse_agent/guardian "create_section_edit_delivery action undefined"*: misreading — it is a
  STEP whose action is `query_database` (486 line 113), the same registered action
  create_rerender uses. No runtime gap. No action.
- *guardian "calibration may not include tool-improver's transitions"*: the 235 pairs are ALL
  of component_versions (checked: components with completed improve_tool items carry versions
  in the set). No action.
- *bug_historian "other html_template writers uncovered"*: REAL follow-up — scriptStubRegression
  guards the 3 componentRegressionIssues call sites; store_generated_component's REGEN path has
  its own birth gates but no comparative stub check. Recorded in `bugs_open/324` as a named
  residual (route: wire the same pure check into the regen compare; separate round).
- *llm_reliability thinking-spend*: refusal-not-silent is the direction; canary refusal rate is
  the observable (already in the floor's comment). No config guess shipped.
- *prior_art "{{.ComponentID}} convention unchecked"*: it is checked — RFC_032 is the standing
  open question, in the LANDMINE and §13's traps. Pointer recorded.
- *debug_historian derived-vs-pinned seed*: deliberate (counts drift daily; documented in 487).

**Execution from here is unchanged and gated:** next chassis BUILD (none yet carries
`35d2e0f9c`/`79fa79fb4` — v1.0.1315 predates both) → digest+ancestry verify BOTH → 486 → 487 →
drain → --bindings exit 0 → LMC judged sequence. Commits applying 486/487 may carry
`Council-Reviewed: 07635a2f` — the verdict has been READ.

## 2026-08-19 (session 7) — GATE CLEARED: v1.0.1316 verified, 486+487 APPLIED, 67-item repair batch draining

Roll verified per RUNBOOK §1: both pods v1.0.1316 at one digest (`2d0d3defc…`), local tag
digest MATCHES, revision `07eeba4a1` — `35d2e0f9c` AND `79fa79fb4` both ancestors. Binary
probes with controls: `scriptStubRegression` 3 / absent-control 0; `repair_instance_scope_bindings`
5 / `no_such_symbol_283_xyzzy` 0. Round-9 verdict RE-READ first-hand from diagnosis_artifacts
before writing the trailer (approved, 2026-08-19 15:53).

**486 applied by hand** (ON_ERROR_STOP): pre-image snapshot `076455bf`, both guarded UPDATEs
hit 1 row, verify DO passed incl. the PREPARE-compile. **487 applied**: enforced 486-first
precondition passed; **seeded 67 items, 42 high**. ⚠ 42 ≠ the 14 "serving-broken" measured at
design time, and that is TWO effects, not one defect: (a) the severity predicate marks
serving-CONVERTED bytes, a SUPERSET of serving-broken (confirmed live: `tool-css-unit-converter`
was high yet repaired as "already sound"); (b) pages keep rerendering onto converted templates,
so the serving set grows hour by hour. Severity is prioritisation only — no action.
Both files renamed away from _HOLD and `--record-only`'d; renames committed `88ad91433` with
`Council-Reviewed: 07635a2f` (verdict read).

**Drain under monitor** (poll 60s, exits when nothing in-flight). Early spot-checks match the
design exactly: fixed:true with named repairs (css-specificity: 2 binding literals;
supplier-comparison: concat prefixes) + "already sound" no-ops + auto-filed `page_rerender`
items. **First live owned-page delivery observed**: `section_edit_tplfix_f99de4cc…`
(tool-fuel-cost-estimator) — 486's create_section_edit_delivery firing in production; its
completion + the page's redeploy is the runtime proof of the empty-field_updates link the
canary was to establish. Watch it to completion before citing the link as verified.

## 2026-08-20 (session 7, cont.) — batch DRAINED and verified at the artefact; the [UNVERIFIED] owned-delivery link is now VERIFIED; oracle baseline PASS 170; judged canary SEEDED

**Drain tally (final):** 67 = 28 fixed + 35 no-ops + 4 needs_human_review. Matches design
(~27/~37/5) with two explained deltas: corpus is 67 rows today not 69 (drift; the seed derives,
so nothing was missed — verified by the not-seeded query returning zero rows), and only 4
parked because **the judged pipeline SUCCEEDED on one judged row**
(automation-savings 16a0bf97, fix_type=scope_component_instance_judged, gate accepted,
snapshot v3) — the judged branch's first production success. The 4 refusals are all
"refused by the judged gate" with the failing checks named (composition hazards; loot-table's
dynamic unprefixable ids) — the designed refusal-not-silent path, exercised in production.

**Done-check:** `instanceaudit --bindings` over a fresh 67-row export: all 67 script bodies
scoped, 0 duplicate ids if doubled, dangling rows = EXACTLY the 4 parked (0 mechanically
repairable) — exit 3 naming them. ⚠ first export TRUNCATED mid-stream (kubectl exec EOF error,
1,052,667 B, JSON unterminated) — parse-validate any exec'd export before trusting it; the
retry parsed clean.

**Artefact verification (not just statuses):** all 7 owned-page section_edit deliveries
checked at stored rendered_html — every id c-prefixed, every literal getElementById lookup
resolves to a declared id, zero bare. One fetched LIVE
(gaswholesalers.com/tools/tool-fuel-cost-estimator.html, deployed_at 21:22): 14 prefixed ids,
0 bare lookups. **That closes §5.2's [UNVERIFIED] link at the artefact — an empty-field_updates
section_edit re-rendered and deployed an owned page end-to-end** (on gaswholesalers, before the
LMC canary even ran). Fleet-wide: 51 placements serve converted bytes, 0 with bare literal
lookups. [SCOPE OF THAT CHECK: literal lookups only — composition-class breakage is invisible
to it, which is exactly the class of the parked rows.]

**Residuals, named:**
- **3 automation-savings placements live-broken** (ai-agent-orchestration.com, finetuning.uk,
  fundamentallyai.com; rows 795c34e6 + c243e0e0) — composition-class, judged gate refused the
  LLM rewrite twice, parked for a human. Contained alternative stays as named in 324: snapshot
  rollback + rerender — **owner call**. fuel-budget + loot-table placements serve
  pre-conversion (WORKING) bytes — no visitor harm.
- **1 page_rerender FAILED (contained):** tool-model-approach-selector on fundamentallyai.com —
  save_page_sections' prune-floor guard refused 3× (plan re-confirmed 1 of 3 stored sections;
  NOTHING deleted). Page serves pre-conversion working bytes; its template is repaired in
  store. The refusal is another lane's protection working; the FAILED orchestrations are on
  record for the immune sweep. Not a 283 defect.
- **Estate drift:** 18 new class-A (already-scoped) tool rows arrived unconverted since the
  census, and the judged pool's "2 generic tools" is now 3 rows (archetype-clash ×2 +
  bayesian-ranking). Follow-on, not this session's scope.

**LMC CONTRIB reply read** (their dir + copy in ours): NO VETO; canary sized (3 components,
9/170 checks on the page); they hold their planner phase 4 until our canary lands — tell them
via these NOTES. **Their §4 gap resolved by measurement:** tool-overpayment-priority is
ALREADY CONVERTED — it was in the MECHANICAL pool (ids-only; clean in today's --bindings
audit), so no LLM touches its arithmetic and it is NOT in the remaining judged set. The oracle
gap on it is their follow-on (c), not a blocker of ours. Their sequencing ask accepted:
no build-site-planner runs from us; any such run is not them.

**Owed steps before canary: DONE.** Oracle baseline on the repaired estate:
**PASS 170 / FAIL 0 / CONVENTION 6 / N/A 0** (2026-08-20). Instrument control in the same
session: `--mutate expectation --tools standard-calc` → 12 FAIL / 0 PASS, CONTROL OK.
b2_verify rebaseline is owed AFTER conversion (baseline must be of the converted estate);
its pin must be IMPORTED, not restated (LMC trap, poisoned-ref history).

**Canary seeded** (item_key instance-scope:b420389f, created_by 283-judged-canary, prio 30)
under a monitor watching item + section_edit delivery. Next after chain completes: verify
served page (prefixed ids, 0 unrendered tokens, binding check), move oracle selectors
#id → #c-loans-standard-calc-<id>, PASS 170 restored + mutate control, one commit.

## 2026-08-20 (session 7, cont.) — canary judged write CLEAN; its delivery was collateral of a cross-lane outage (bugs_open/336), already rolled back by that lane; delivery requeued

**Canary (loans-standard-calc): the judged pipeline half is PROVEN.** Item completed fixed:true
via `scope_component_instance_judged` (2,475 → 2,858 B, 6 ids, snapshot v1). The written
template audited clean the same hour: `instanceaudit --bindings` on the written row = script
scoped, 0 collisions doubled, 0 dangling, exit 0.

**Its section_edit delivery FAILED twice (07:07–07:10) — NOT ours.** WORKFLOW_INVALID:
`update_page_status` refused config key `deploy_result_field`. Cause: migration 494 (315 lane)
armed that key on three agents at 06:49:49Z, but the key is declared on `render_component`'s
spec, NOT `update_page_status`'s — whose spec is CheckConfig:true (strict), so every
page-stamping workflow died at dispatch, fleet-wide (12 failures 06:58–07:29). **Diagnosed and
FIXED by another session while we were diagnosing it**: `bugs_open/336` filed, 494 rolled back
07:22:40Z with its own ROLLBACK, restoration proven WITH DEMAND, four orphaned items requeued —
including our canary's section_edit (triaged 07:27:36). LANDMINE appended by them
(`8e413afae`): a key on the WRONG action's spec passes every instrument — binary literal
present, git log -S names the reader — while the strict spec refuses it at runtime. Our own
07:13 binary probe hit exactly that: 3 hits / 0 control, and it was evidence of nothing.
Delivery re-monitored; oracle lockstep still owed after it lands.

## 2026-08-20 (session 7, evening) — CANARY COMPLETE AND PROVEN: loans-standard-calc converted, delivered, served, oracle PASS 170 restored

**→ LMC lane: the canary has LANDED and is VERIFIED — your phase 4 hold can lift** (per your
CONTRIB §3; this is the NOTES line you asked for).

The delivery stall was a THIRD cross-lane trap, now a LANDMINE ("Requeuing a failed
site_work_items row…"): the 336 requeue reset status but not attempt_count; the claim query
(claim_work_item_action.go:103) requires attempt_count < max_attempts, so the row sat triaged
and unclaimable for 9.5 h while 116 rerenders flowed past it. Reset both fields → claimed
within ONE MINUTE, complete moments later (the reset is the demand-proof of the mechanism).
Fleet zombie census after repair: 0.

**The full §5.2 chain, each link verified:**
- judged write: fixed:true, 2,475→2,858 B; `instanceaudit --bindings` on the written row exit 0.
- delivery: section_edit (empty field_updates) → re-render → deploy; pages.deployed_at
  16:56:45; item complete. (Delivery ran on v1.0.1320 with 494 RE-ARMED post-spec-fix — the
  stamping step traversed cleanly, incidentally re-proving 336's fix under demand.)
- served page (live fetch): all 6 ids prefixed `c-loans-standard-calc-*`, 0 unrendered tokens,
  9/9 getElementById lookups resolve, 0 bare.
- oracle: baseline PASS 170 / FAIL 0 / CONV 6 / N/A 0 (pre-conversion, this morning) →
  selectors moved (21 occurrences, block-scoped — other tools reuse names like
  #total-interest, so a file-wide replace would corrupt their blocks) →
  **PASS 170 / FAIL 0 / CONV 6 / N/A 0 restored** → `--mutate expectation` control:
  12 FAIL / 0 PASS, CONTROL OK (run both before AND after the move).

Remaining §5 sequence: b2_verify rebaseline (owed), then the 22-calculator batch (oracle per
wave, mortgages-repayment's 2 pages get a deliberate look), then the generic pool
(bayesian-ranking + archetype-clash ×2), and the 4 parked judged refusals stay an owner call.

## 2026-08-20 (session 7, night) — THE 22-CALCULATOR BATCH IS DONE: 20 converted+delivered+live, 2 gate-refused to humans, oracle PASS 170 across the whole estate

**Batch tally:** 22 seeded → **20 fixed** (all `scope_component_instance_judged`) + **2 refused**:
`loans-application-tracker` (rewrite touched markup outside script bodies — LLM overreach,
gate right) and `loans-consolidation` (dynamic `${n}` template-literal ids; the LLM's
prefixing broke id-set parity vs baseline — conservative-gate-vs-plausible-fix, designed to
go to a human). Both join the parked pool (now 6 total with the repair batch's 4).

**Verified at every layer:**
- all 20 written templates: `instanceaudit --bindings` exit 0, zero collisions doubled.
- 21 section_edit deliveries complete; 20 published. The 21st (`mortgages-repayment-demo`)
  was SKIPPED by `ARCHIVED_PAGE_GUARD` — page is status=archived; the guard refused to
  re-publish a retired page and named it in the result. Correct behaviour; the PLAN's
  "2 pages" for repayment predates the archiving. The deliberate look §5.3 asked for was
  worth it: `deployed_at` unmoved on that page is the guard, not a failure.
- live spot-checks (repayment, settlement-calculator, stamp-duty): prefixed ids served,
  all lookups resolve, 0 bare, 0 unrendered tokens.
- **oracle: PASS 170 / FAIL 0 / CONV 6 / N/A 0** — full estate, identical to the
  pre-conversion baseline; `--mutate expectation` full-suite control: 161 FAIL / 0 PASS,
  CONTROL OK.

**Selector-move gotchas found and handled (for the next converter):**
- 11 press selectors were `button[onclick='F()']` — the conversion REMOVES inline on*=
  attributes, so every one would have gone N/A. The rewiring pass gives its target buttons
  ids (`{{.InstanceID}}-btn-*`), so each press became `#c-<fn>-btn-*`. Mapping read from the
  converted scripts' own addEventListener bindings, NOT guessed — equity-release's
  `calcCompound` is `btn-project`, not `btn-calculate`.
- the loans/mortgages OVERPAYMENT block covers TWO pages in one block — split at
  `def over_mort` before prefixing, or one tool's prefix corrupts the other's selectors.
- `second_press` (investor LTV) and car-finance's parameterised `setType('%s')` (→ 
  `"#c-…-btn-%s" % key`) don't match the simple press pattern — grep `button[onclick`
  afterwards; the ONE legitimate survivor is consolidation's `addDebtRow` driver line
  (page unconverted, refused).

Remaining: the generic pool (bayesian-ranking, archetype-clash ×2), b2_verify red-by-design
on its 7 seeded pages until the LMC lane recaptures, the 6 parked rows (owner call on the 3
live-broken automation-savings placements), and the 08-15-vintage estate drift (18 new
class-A tools) as follow-on.

## 2026-08-20 (session 7, close) — GENERIC POOL DONE: the planned conversion programme is COMPLETE

All 3 generic rows (bayesian-ranking, archetype-clash ×2) converted via the judged pipeline
(fixed:true), written templates audit clean (exit 0), all 3 page_rerenders complete, live
pages verify (prefixed ids, 0 tokens, 0 bare lookups on gamesdesign.co.uk ×2 + vonc.com).
**Functional click-through PASSED on all 3** (chromium via the LMC lane's Driver — reused, not
forked): archetype-clash's button updates its results panel, bayesian-ranking's live inputs
recompute 8 outputs — the converted scripts BIND and EXECUTE on the prefixed ids, the exact
aliveness 324's dead tools lacked. Script: scratchpad clickthrough.py (session-local; the
pattern — snapshot prefixed non-input elements, drive inputs+buttons, assert text changed —
is worth lifting into a tool if a third session needs it).

**Programme state: 24 judged conversions (canary + 20 LMC + 3 generic) delivered and live;
repair batch done; oracle PASS 170 = baseline; 6 rows parked for humans (4 repair-refusals
incl. the 3 live-broken automation-savings placements = OWNER CALL, + 2 batch refusals).**
Follow-ons (not this lane's session): LMC b2_verify recapture on its 7 seeded pages; the
18 new unconverted class-A arrivals (birth-gate or standing sweep — needs a decision);
monitor's own clock bug noted (a window set in the future reads as 'none' — compare the
filter to now() before believing an empty).

## 2026-08-21 (session 8) — THE FLOW HALF: mistake logged, generator taught, birth guard built+armed-on-roll, sweep LIVE (owner ruling: log/teach/guard/sweep, Go auto-convert preferred)

Owner asked why unconverted tools keep arriving; the answer became WRONG_CALLS 2026-08-21
(committed 62936db59): the programme converted the STOCK and never censused the FLOW —
tool-generator (verified: prompt knew neither InstanceID nor getElementById) minted 23
old-style tools 08-18→08-20. The owner ruled: log it, teach the tool, turn the guard on,
prefer the Go code auto-converting the LLM's ids, sweep as part of the improvement loop.

**Built (all committed e186a2bd3, Council-Submitted 6acf8e4e — READ THE VERDICT):**
- `ScopeToolBirthTemplate` (CLC-023): create_tool_component runs the PROVEN CLC-017
  converter+gate at birth. Armed: convert (occurrence-0-bound rendered bytes — placements
  carry the template VERBATIM) or refuse-to-regenerate (no live behaviour at birth ⇒ refusal
  beats judging). Unarmed: record-only. Self-converted output VERIFIED never trusted. Tests
  green on archive overlay, both-direction controls. ⚠ live wiring unproven until the first
  post-roll generation — stated in the register, not grep-asserted.
- Migration 520 APPLIED: prompt rules 21 (single IIFE) + 22 (id discipline; the platform
  namespaces ids at save) — NOT the placeholder (it would render inside prompt_template: the
  escaping trap; and one implementation of the rename, in Go). enforce_instance_scope=true on
  save_tool: inert pre-roll BY MEASUREMENT of the validator semantics (old spec declares no
  contract ⇒ tolerated; the new ConfigKeys declaration ships WITH the reader — 336 applied).
- `instance-scope-sweep` (CLC-024) LIVE, replacing the retired CLC-016 tripwire (repo dir
  removed; register corrected VISIBLY since another session had kept it as a worked example).
  First run: corpus 28 → filed 26 (2 parked refusals dedup-skipped, reconciles exactly);
  25 drained complete within the hour. Demand control = adopter count (90+), REFUSES on zero.
  ⚠ its filed-count originally counted psql's INSERT tally line (27 vs 26) — fixed+redeployed
  same hour; the first doc_notes row carries the off-by-one.

**The 26th item is a KNOWN owed repair surfacing correctly**: tool-spawn-rate-balancer's
template internally declares `chartTitle` TWICE, so the doubled-instance gate fails the
transform loudly (2 attempts burned; the 3rd will fail → status failed). ⚠ GRIND NOTE: failed
is terminal to the sweep's dedup, so it REFILES daily (3 failing runs/day) — and the `ec2`
hex-id row no-ops daily the same way. Cheap, visible, and correct pressure, but the FIX is
the two small template repairs owed since 08-18 (chartTitle internal dup; ec2 hex-ambiguous
id) — whoever picks those up kills the grind at the source. Do NOT widen the sweep's
exclusion to failed rows: a transient failure would then permanently exempt a row.

Also survived: migration number race (519 taken between listing and writing — renumbered
520); two index.lock collisions (waited, never removed another session's lock; one commit
landed despite 'unable to write new index' — verify with git log before retrying, the
retry would have DOUBLE-COMMITTED).

## 2026-08-21 (session 8, cont.) — round 1 REVISE (gating HIGH was RIGHT again), answered with code; round 2 submitted

Round 1 on `6acf8e4e`: REVISE, gated by bug_historian. **The REVISE loop found real things a
third time**: (1) the HIGH — what does a guard refusal during REGENERATION do to the stored
rendered_html (the bugs_closed/056 blanking class)? Answer was "nothing, by ordering" — and
ordering is a promise, so it is now STRUCTURAL: regen refuses empty bytes as its FIRST
statement (test callable with zero ActionParams — the check precedes every use). (2) the
occurrence-0 single-placement assumption had no runtime assertion — now asserted in the regen
census (same-page duplicate refuses to a human). (3) both objecting seats named the
deploy_tool fork door — MEASURED alive (13 births/30d, latest 08-19) and now guarded with the
SAME helper; this also closes a real blindspot neither round named: an UNPLACED library
source is invisible to the sweep (placed-only corpus), so the fork guard is the only
preventive control at that door. Armed by migration 530.

⚠ Migration number races fired TWICE in one session (519 taken between list and write;
my 521 collided with an existing 521 I listed but didn't see — renumbered 530, which two
OTHER lanes also share). The directory's working invariant is now FILENAME uniqueness, not
number uniqueness; the superseded 521-named ledger row is annotated by 530's record note.
⚠ The 260 lane's uncommitted RenderTemplate 4-value signature adapted my committed files in
the working tree; my round-2 commit ships ONLY files proven green on HEAD+exactly-them
(archive overlay), leaving their adaptation for their commit — a green working-tree build is
not a green HEAD, in both directions.

Round 2 submitted (RESUBMIT_CORR, same correlation; first attempt refused at the 8-edit cap —
consolidated by folding, and the docs edit gave way to the deploy_tool edit per editquality's
own note that bookkeeping needn't count). Commit `Council-Submitted: 6acf8e4e`.

## 2026-08-21 (session 8, close) — ROUND 2 APPROVED (verdict READ 14:03:59; 4 advisories, none high). The flow half is done pending the roll.

Advisory triage (each checked before recording): *editquality "migration 530 has no distinct
edit entry"* — an artefact of consolidating to the 8-edit cap (the trigger refuses 9+); the
FILE is committed, applied and ledger-recorded, so nothing was dropped — the cap forces
bundling and the reviewer is right that bundling obscures; noted for future rounds: prefer
dropping a docs edit entirely over folding two migrations into one entry. *editquality wiring
wording* — the description now matches the code (one guard invocation + a seam-local
empty-bytes refusal in the regen). Remaining advisories in the report body; none change code.

**Commits `e186a2bd3` + round-2 revision may now be credited `Council-Reviewed: 6acf8e4e`**
(verdict read; 098 auto-credits the Council-Submitted trailers regardless).

**Arming state on roll** (nothing more owed from this lane until then): the next chassis roll
carrying the 2026-08-21 commits activates BOTH guards (tool-generator birth via 520,
tool-deployer fork via 530). Post-roll demand checks owed: one natural tool-generator birth
(result carries instance_scope fields; the sweep's next-day count should stop growing) and
one deploy_tool fork (fork row born with InstanceID). The daily sweep needs neither.

## 2026-08-21 (session 8, evening) — OWNER RULED: rebuild-not-repair. 8 rebuilds seeded; sweep gains the escalation arc; LMC pair on a veto window

Owner's three rulings: (1) automation-savings REBUILT from scratch through the full pipeline
(not rolled back, not hand-fixed) — "deconstructed, fully editable, improvable"; (2) same for
the gate-refused pool; (3) repairs authorized — folded into the same route (for chartTitle and
the hex-id row the rebuild IS the repair). Plus: make the whole class self-healing via the loop.

**Seeded 8 `add_tool` items** (`created_by='283-owner-rebuilds'`, replace_existing:true, spec
derived from each incumbent's display_name + description + tool-doc header — the framework's
words): automation-savings ×4 sites (3 broken-serving at prio 30, leopardess at 55),
fuel-budget, loot-table, spawn-rate-balancer (chartTitle), process-automation-scorer (hex id —
identified: leopardessconsulting). ⚠ 795c34e6 is SHARED by ai-agent-orchestration +
fundamentallyai: the regen path REFUSES cross-site rows by design (create_tool_component_
regenerate.go:216 names the per-site-fork route) — those two items are EXPECTED to refuse
loudly; the fork-first choreography follows from the refusal text when it lands. CLC-020
(site-aware regen divert) is LIVE in 1320 but lives on store_generated_component, not this
path. **LMC pair (application-tracker, consolidation) NOT seeded yet** — CONTRIB with 24h veto
window filed in their dir (their D6 phase 4 may be in flight; consolidation is oracle-covered
so its block needs a lockstep REWRITE at delivery, bigger than a selector move: its driver
clicks `addDebtRow` by inline onclick, which a rebuild removes).

**Sweep escalation SHIPPED** (`a332c19a7`): ≥2 FAILED conversion items → ONE rebuild item,
14-day rate limit, spec derived from the incumbent; needs_human_review rows deliberately
NEVER auto-escalate (a rebuild discards working behaviour — human's call; the owner made it
by hand for today's six). ⚠ Council gate REFUSED the submission on SCOPE — deployments/ is
outside its footprint (owner ruling 2026-07-17) — not on merits; FORCE deliberately not spent.
The commit-msg hook also correctly refused a 'Council-Submitted: pending' placeholder trailer:
submit FIRST, then commit, or no trailer at all.

**The loop, closed end-to-end once the rebuilds land**: born-bad or legacy-bad template →
sweep detects → mechanical convert OR (failed ×2) rebuild → fresh generation under the live
prompt rules → converted at birth post-roll (or by next sweep pre-roll) → rerender/deploy.
Parked-for-human stays the one deliberate manual gate.

## 2026-08-21 (session 8, night) — v1.0.1322 ARMS BOTH GUARDS; rebuilds: 6 of 8 SUCCEEDED (chartTitle + hex-id defects GONE), 2 refused by the SHRINK FLOOR; sweep wave 2 converting

**Roll verified**: v1.0.1322 digest-matched, revision `bac189921` carries `e186a2bd3` AND
`df26249e0` — with migrations 520/530 already applied, the birth guard and fork guard are
LIVE-AND-ARMED from this roll. Every post-roll tool birth is converted-or-refused at the seam.

**Rebuild outcomes (all verified at the artefact, not the status):**
- 6 succeeded: finetuning.uk automation-savings (its broken page's fix), leopardess
  automation-savings + process-automation-scorer, gaswholesalers fuel-budget, gamesdesign
  loot-table + spawn-rate-balancer. All audit CLASS A (prompt rules working — scripts born
  IIFE-scoped), `chartTitle` dup count 0, hex id gone. Born pre-roll so unconverted; sweep
  wave 2 (triggered manually, post-roll) filed 14 conversions — the 6 rebuilt + 8 new
  arrivals — now draining. Sweep's filed-count fix confirmed live (14 = 14 keys).
- **2 REFUSED, and the refusal reads as `complete`**: the shared-row pair
  (ai-agent-orchestration + fundamentallyai, row 795c34e6). ⚠ the workflow's error path
  (`complete_error`) completes the ITEM with an empty create_result — read the ORCHESTRATION
  error, not the item status. The refusal was the **visible-text shrink floor** (bugs_open/331
  protection): fresh generations kept 46%/48% of the incumbent's 1,114 visible chars, floor
  50% — "nothing written; the incumbent tool still serves". The floor is right by its own
  lights; behind it ALSO waits the cross-site regen refusal (line :216), so richer prose alone
  does not clear the path. **The shared row needs untangling first** — the framework's own
  prescriptions: retire/deactivate the shared row then per-site add_tool with
  adopt_existing_page (the bugs_open/286 route; step-level key), OR per-site fork — but
  deploy_tool's fork of THIS row is now (correctly) refused by the armed fork guard: the
  template is composition-broken judged-class. **OWNER CALL, options costed in the CONTINUE.**
  The floor override (`sectionShrinkFloorKey`, 0 disables) exists but is STEP-level — using it
  would run every concurrent add_tool floor-less; deliberately not touched.

**Housekeeping**: parked repair items for fuel-budget + loot-table CANCELLED with reason
(their broken templates no longer exist — rebuilt); automation-savings' 2 parked items STAY
(rows still broken pending the untangle). Stale-conversion items for rebuilt rows are
naturally superseded by wave-2 items (same item_key, prior ones terminal).

## 2026-08-21 (session 8, late night) — OWNER RULED the untangle: shared row RETIRED, per-site fresh births seeded; decision 2 confirms the LMC rebuilds

Decision 1 executed (SQL_2026-08-21_retire_shared_row_and_rebirth_per_site.sql, applied with
verify DO): 795c34e6 deactivated (its own row + component_versions are the pre-images), both
slots tombstoned `build_status='removed'` (the 286 route's "retire the ported slot" verb;
archive trigger holds the served bytes), `adopt_existing_page=true` armed on save_tool
(**TEMPORARY — un-arm after the two births land**:
`UPDATE agent_definitions SET default_config = default_config #- '{workflow,steps,save_tool,config,adopt_existing_page}' WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;`).
Two `add_tool` items seeded (`283-owner-rebirths`, prio 30, NO replace_existing — no incumbent
exists now): spec = retired row's description + tool-doc contract + an explicit visible-prose
requirement (the shrink-floor lesson applied at the spec, not by lowering a guard). These run
POST-roll ⇒ **they are the armed birth guard's live demand check** — expect
`instance_scope: mechanically converted at birth` in the results and born-converted templates.

Decision 2: the owner confirms rebuild-and-deconstruct for the remaining pool — the LMC pair
proceeds after their veto window (~2026-08-22 midday; the window is THEIR coordination
courtesy, not approval - consolidation's oracle-block REWRITE still owed in the same commit).

## 2026-08-21 (session 8, close) — THE LAST BROKEN PAGE IS FIXED. Both guard branches proven LIVE with demand. One benign tail (finetuning content refresh) queued.

**The untangle landed end-to-end**: a-a-o rebirth SUCCEEDED first sample (`5e3a4ca5`,
result carries `instance_scope: mechanically converted at birth` — **the armed birth guard's
CONVERT branch, proven live**); fundamentallyai REFUSED twice (id-array iteration — the
guard's REFUSE branch, proven live; ⚠ the generator workflow's error path COMPLETES the item,
no auto-retry — reseed by hand, and consider rerouting save_tool's error_step to
fail_work_item as a follow-on), succeeded on sample 3 after the id rule was made STRUCTURAL
in the spec ("every getElementById takes a quoted literal at the call site; never iterate
arrays of ids"). **Live: a-a-o 28/28 prefixed 0 bare; fundamentallyai 28/28 prefixed 0 bare.
ZERO pages on the estate serve a broken tool.** adopt_existing_page UN-ARMED (verified ABSENT).

**Benign tail**: finetuning serves its rebuilt tool WORKING but unconverted — its rerender is
blocked by the CLAIMS FLOOR ("0% reduction"), which is two guards interlocking correctly:
the missingkey=zero exposure renders an unfilled content field as 0, and the claims floor
refuses to publish it. Root cause = the regen path files NO needs_content_page item (fresh
births do). One seeded by hand (`needs_content_fe2cbe67…`); after the content writer fills
the fields, the rerender delivers the converted tool. Follow-on worth registering: the regen
arm should file the content item itself when the template's field set changes.

Remaining for the NEXT session: LMC pair after the veto window (~midday) with consolidation's
oracle-block rewrite; finetuning tail verification; tomorrow's 07:40 sweep steady-state read
(corpus should be ~0 now the shared pair is resolved); then CLOSE 324 and 283.

> **CORRECTED 2026-08-21 21:0x — the finetuning claims-floor diagnosis above was WRONG.**
> I wrote "the missingkey=zero exposure renders an unfilled content field as 0" [INFERRED, and
> the inference was wrong]. The measured cause: the rebuilt template's own ZERO-STATE copy
> literally contains `<span id="…-error-reduction-percent">0</span>% reduction` — the results
> panel reads "a 0% reduction" before any input. Found by searching the template for
> '% reduction' (my earlier '0% reduction' search missed it because the 0 lives inside a span).
> The content-refresh item I seeded on the wrong theory completed as mark_no_ready_sections /
> needs_human_review — harmless, cancel it when the real fix lands. The real fix: one final
> regen of c243e0e0 with a claim-free zero-state requirement in the spec (seeded, watched).
> Both fresh births are clean of the phrase (checked all 5 active rows: only c243e0e0 carries
> it). Follow-on worth a prompt rule: zero-state result copy must not read as a claim —
> today the claims floor catches it fail-loud at publish, which is correct but costs a regen.

## 2026-08-21 (session 8, final) — THE TAIL IS CLOSED. Every page in the programme serves converted, working, claim-clean bytes.

Finetuning's final regen succeeded (⚠ my watch misread it: a REGEN success carries no
instance_scope field in its result map — only the fresh-birth path merges scopeInfo; a
'complete scope=-' on a replace_existing item can be SUCCESS, read the template not the
result). The armed guard converted the bytes before the regen wrote them (template carries
InstanceID; zero-state claim phrase gone), the regen's own rerender passed the claims floor
and deployed 20:02:44, and the live page reads **20/20 prefixed, 0 bare, 0 tokens, 0 claim
phrases** (first fetch was CDN-stale — cache-bust before believing a post-deploy read).
The moot content item is cancelled with its reason. Follow-ons recorded for the register
pass: (1) regen result map should carry scopeInfo like the birth path; (2) generator prompt
could gain the claim-free zero-state rule (the floor catches it fail-loud today, at the cost
of a regen); (3) the regen path files no needs_content item.

**Estate state at close of session 8: zero broken tools, zero unconverted serving placements
in the programme's scope, both guards proven live in both directions, sweep + escalation
standing. Left for next session: LMC pair (veto ~midday), morning sweep steady-state read,
then CLOSE 324 and 283.**

## 2026-08-22 (session 9, morning) — 324 CLOSED; sweep's first scheduled run + first LIVE escalations; escalation blind spot found and fixed; aria-builder parked deliberately

**324 moved to bugs_closed** (`c427c3108`) — fixed AND live at every layer, each demand-proven.

**Sweep ran unattended at 07:40** (corpus 5, converted 119 — from 91 four days ago). It exposed
my escalation's blind spot: the dynamic-id class refuses as **complete no-ops**, not failures,
so the failed-only predicate never fired — three rows ground 3 rounds each (aria-builder,
economy-flow-modeller, shadow-stacker). Fixed (`74a73db31`): the predicate is now
"still unconverted after ≥2 TERMINAL items" (status-agnostic — a fixed:true item converts the
template in its own transaction, so an unconverted row's terminal items are all
non-conversions), correction visible in the docstring. Triggered post-roll-settle (v1.0.1323,
rev = this lane's own last NOTES commit): **the escalation arc fired live for the first time,
filing 3 rebuilds.**

**First escalated samples all refused by the armed guard** (unprefixed dynamic ids — the
guard's refuse branch, 3 more live proofs) — the escalation-derived spec lacked the
structural id rule that converged fundamentallyai; added permanently to ESCALATE_SQL
(`83987c58d`) + hand-reseeded. **2 of 3 converged** (economy-flow-modeller, shadow-stacker —
rows converted at birth, verified at the TEMPLATE not the item). **aria-builder refused a
3rd sample even with a design instruction** (it renders its GENERATED example ids as live
DOM — the tool's domain fights the rule): **parked deliberately** as needs_human_review with
the evidence, which stops both the daily refile and the escalation. It serves a working
unconverted page; a person redesigns the preview or hand-writes the template.

Moot parked items for the retired/reborn automation-savings rows cancelled (2). Remaining
open: LMC pair (window to ~midday), aria-builder (human), then 283 closes.

## 2026-08-22 (session 9, midday) — LMC pair: a messy but contained hour. Consolidation REBUILT (new row, born converted); my un-arm broke another lane (owned, attributed); tracker rerouted through the section door.

**Round-1 LMC failures, both mine:** consolidation's incumbent had an EMPTY display_name (the
seed mapped spec.name from it → required-field failure); application-tracker is
**SECTION-level** — the add_tool door was wrong (its page is live as page_type=content and the
page-role guard rightly refused to re-type it). Rerouted: tracker → needs_new_component
(component-creator regen, item `rebuild_section_loans-application-tracker`).

**Consolidation round 2 SUCCEEDED but not as a regen**: `sanitiseFunction` prefixed the
function to `tool-loans-consolidation`, the per-site probe missed the incumbent, the CREATE
path ran and — `adopt_existing_page` standing — adopted the live owned page, leaving **TWO
deployed tool slots** (the exact hazard the regen path exists to prevent; the placement
INSERT's ON CONFLICT cannot see a DIFFERENT component's slot at the same position). CONTAINED
same hour: old slot tombstoned, old row 3efd4989 deactivated; ONE slot serves (new row
`aacde020`, function `tool-loans-consolidation`, CONVERTED AT BIRTH by the armed guard).
section_edit delivery seeded. ⚠ the OLD function name dies with the old row — the oracle
rewrite must target the NEW markup AND the sweep never sees the retired row.

**The `adopt_existing_page` resurrection was NOT a mystery — it was my incident** (WRONG_CALLS
2026-08-22 + CONTRIB into webdesign_tool_rebuilds): the flag was STANDING since THEIR
migration 435 (2026-08-16); my 08-21 "arm" was a no-op on an already-true key, my bare-UPDATE
"un-arm" removed THEIR production config (no snapshot, no ledger, no provenance grep), their
Phase C build died on it at 11:28Z today, and their migration 558 restored it blind
("nobody's"). Attribution + apology filed; 558 stands; the SQL_2026-08-21 file's un-arm
instruction is hereby VOID — the flag is theirs.

## 2026-08-22 (session 9, afternoon) — the ORACLE DRIVE caught a real defect in the rebuilt consolidation tool on its FIRST vector: a hallucinated interface to the site's shared engine

Delivered page live-verified structurally clean (prefixed ids, 0 bare, 0 tokens) — and then
the calibration drive found the tool CANNOT CALCULATE: every debt row returns "Could not
calculate this debt". Diagnosis at the code arm (parseRow): the generated script calls
`window.calculateAmortization(principal, rate, term)` — the site's SHARED calculators.js
engine, which EXISTS (typeof function) but whose interface the LLM GUESSED (return
keys/units differ), so totalInterest/monthlyPayment come back non-finite. Birth gates
cannot see a runtime interface contract; the oracle's first drive did. **This is the
LMC-first ruling and RFC_034's oracle-as-witness argument proven in one incident** — a
structurally perfect, gate-clean, id-scoped tool that computes nothing.

Also observed while driving: debt terms are now YEARS (old tool: months) and the tool
initialises with 2 rows whose validation counts filled-in rows only — semantics for the
oracle rewrite. Prompt-rule follow-on (generator): rule 7 ("no external dependencies") did
not stop a window.* call into a sibling script — it names CDN/fetch only; worth an explicit
"no window.* / shared-script calls" clause in a future 520-style migration.

Round 3 seeded (same key): self-containment sharpened — inline annuity arithmetic required,
no window.* calls, years semantics stated. Monitor also watches the engine literal in the
template as the tell.

## 2026-08-22 (session 9, late afternoon) — consolidation DONE end-to-end: round-3 self-contained tool live, penny-identical to the oracle; block rewritten; full suite green at its NEW total

Round 3 delivered (section_edit → deployed; propagation verified live). Calibration drive
BEFORE writing the block: tool vs oracles.py to the penny (existing interest £2,886.99, new
interest £2,174.98, new monthly £169.58 — two-debt vector). Round 3's shape is better than
the old tool for verification: FOUR STATIC debt rows (no dynamic ids at all), per-figure
result ids. ⚠ semantics changed and the block encodes them honestly: debt terms in YEARS
(was months), new-loan principal USER-ENTERED (was derived — cases enter the debt total).

Oracle block REWRITTEN (block-scoped; the consolidation dynamic-row setup arm retired with a
fail-loud stub — consol was its only user). Consolidation alone: 12/12 PASS incl. BOTH 0%
boundaries. **Full suite: PASS 166 / FAIL 0 / CONVENTION 6 / N/A 0 (as of 2026-08-22)** —
the total moved 170→166 because the old block carried 4 checks against a derived debt-total
display the new tool replaces with an input field; 170−16(old block)+12(new)=166, nothing
else moved. Mutate control on the rewritten block: 12 FAIL / 0 PASS, CONTROL OK.

Remaining before 283 closes: tracker retry (static-prose regen queued) + its delivery, and
the close-out sweep read.

## 2026-08-22 (post-close hygiene, 18:00) — parked queue reconciled to the closing documents

The two LMC batch-refusal conversion items (instance-scope:29e63065, :3efd4989) were still
parked from 08-20; both rows were since rebuilt/retired, so they were cancelled with reasons.
The parked pool is now EXACTLY the one genuine human case the closing section names:
aria-builder (instance-scope:b486bb24). Chassis at v1.0.1326; nothing of this lane's is
pending any roll. Next scheduled sweep 07:40 UTC 2026-08-23 — its doc_notes row is the
standing health read.

## 2026-08-22 (session 10, late afternoon) — the bug CLOSED under me mid-plan; the residual it named turned out to be bigger than the close said, and the ruled remedy was not executable

**Started as a fresh session on `bugs_open/283`.** Ownership check at ~15:00 said no live 283
peer, lane files all committed, last lane commit `fa6ed1cac` 13:29 — read as a clean handoff.
**It was not: another session closed the bug at 16:09** (`9223c421d`, `Council-Reviewed: 6acf8e4e`,
co-authored Fable) while I was mid-plan. The close is sound and I did not contest it — the filed
defect class (literal ids on `getElementById` components) is fixed and live. ⚠ **`ListAgents`
showed no 283-named peer the entire time**, so a session-name check is not an ownership check;
only `git log` caught it, and only because I re-read HEAD before editing.

### What I measured before any of that (all 2026-08-22, all controlled)

The tool half is done: 125 placed rows bind by id, **124 converted**, the 1 being aria-builder
(parked, correctly). **0 pages carry 2+ `getElementById` components** — the original fire has
still never happened, and that zero is controlled: the same query shape returns 18 for
`{{.ComponentID}}` components and 29 for any repeated component, against 130 live placements.

**But the same defect is live on the OTHER id mechanism**, the one 283 deliberately left alone.
`{{.ComponentID}}` resolves to the component ROW id on both live render paths:

| | |
|---|---|
| pages with a repeated `{{.ComponentID}}` component | **18** (13 ×2, three ×3, one ×4, one ×6 — all `generic-text-block`) |
| redundant placements | **27** |
| placements serving the row UUID | 249 |
| placements serving `<section id="">` | **11** |
| placements serving the assemble shape `component_<fn>_<n>` | **0** |

Live at the artefact, cache-busted, re-confirmed after the close landed:
`apis.uk/index.html` serves **six** `<section id="8d81e665-…">`; three more pages confirmed;
two single-instance pages read as controls show 1 id, 0 duplicates.

### Three findings that changed the design

1. **The assemble path's substitution is DEAD CODE.** `RenderTemplate` runs first
   (`assemble_from_library.go:303`); `missingkey=zero` resolves `{{.ComponentID}}` to
   `<no value>` and `component_library.go:1170` strips it, so the `ReplaceAll` at `:309` never
   matches. **0 of 270 live placements carry that shape** (regex proved against a synthetic
   positive before the zero was believed). So RFC_032's own §2 table — which called that path
   the reuse-SAFE one, and which framed the RFC's question — was **never true**. Corrected in
   place, dated.
2. **A fourth builder nobody listed**: the section editor (`section_editor_actions.go:1113`,
   `:1249`) binds `ComponentID` at all, which is where the 11 empty ids come from. Both routes
   write `rendered_html` straight to a live page with no downstream gate. The platform already
   *detects* this — `RenderTemplate` logs `Warn "fields rendered empty" fields=[ComponentID]` —
   and **both call sites discard the report** (`rendered, _, _, err :=`).
3. **THE RULED REMEDY WAS NOT EXECUTABLE.** `ConvertTemplateToInstanceScope` harvests ids with
   `\sid="([^"{}]+)"`; the class excludes braces, so `id="{{.ComponentID}}"` never enters
   `seen` and the converter refuses with *"template declares no literal element ids"*. Filing
   the five templates through the proven pipeline would have produced **five polite no-op
   completions** — a queue that drains clean and a census that never moves.

### Two of my own wrong calls (both in WRONG_CALLS 2026-08-22)

- **A control that matched nothing.** My first "no duplicate ids" control was a page with
  **zero** id-bearing sections, so its pass was guaranteed before the fetch. It reads exactly
  like a passing control. Replaced with two single-instance pages, which can fail.
- **I read a copied field as authored intent.** I called the nine content-supplied section ids
  "human-authored anchors that must be preserved" and made protecting them a design
  requirement. `content_data->>'ComponentID'` simply **equals `slot_name`** in 9 of 10 rows —
  visible the moment I printed the neighbouring column. Then, having found slot name, I reached
  for it as the FIX; measuring refuted that too (**20 pages repeat a slot name; 15 of the 18
  colliding pages overlap, so it would fix 3 of 18**) — the same conclusion the council's
  `reuse_agent` seat reached on 2026-08-16.

### What shipped

`67d34e6c1` — converter **pass 0**: an id attribute whose whole value is `{{.ComponentID}}`
becomes `{{.InstanceID}}` (the **bare** token; the wrapper id IS the instance identity), run
before the harvest, which now reads its output. Plus a refusal for the placeholder surviving
*outside* an id attribute — the half-state where literals convert, the templated id is dropped,
and **neither the gate nor `DetectInstanceCollisions` can see the resulting `id=""`** because
`reElementID` requires a non-brace character. 0 live templates are in that state (control: 87
carry a literal id), so it refuses nothing today.

Six tests, **all six mutation-proven**: pass 0's regex made unmatchable → all six failed, the
wrapper-only case with the exact pre-change refusal reason; source restored from a scratch
backup and verified (0 occurrences of the mutant literal); suite re-run green including the
pre-existing Instance/Scope/Conversion/Binding tests; and **built + tested from a clean
`git archive HEAD` extract**, because the working tree carries another session's WIP (their
`store_generated_component_action.go` was momentarily uncompilable mid-run — not mine, and it
settled).

⚠ **Both armed birth guards inherit pass 0**, since `ScopeToolBirthTemplate` calls the same
function — a live behaviour change on `create_tool_component` and `deploy_tool_to_site`, named
rather than buried.

Council: `cd6a5ef6-d530-42c2-81fe-238552eb690d`, submitted before committing,
`Council-Submitted:` trailer (no verdict read). **Verdict still owed a read.**

### What is NOT done, stated so nothing reads as finished

The five templates are **not converted**; `ComponentID`'s two bindings and the dead `ReplaceAll`
are **not deleted**; nothing is re-rendered. **18 pages still serve duplicate section ids and 11
still serve `id=""`.** Pass 0 is inert until the next chassis roll under a bumped tag. Also
untouched: the section-writer birth gate (strand 2, unblocked by today's inline-JS ruling) and
the sweeper's `querySelector` blind spot (strand 3 — **6** actionable rows, the other 25 being
owner-ruled out).

## 2026-08-23 (session 11) — the council's conditional was a fact, my own fix re-opened the thing it closed, and the "18 pages" figure was a proxy

Continues session 10. Chassis rolled to **v1.0.1328 at 11:51Z**, so pass 0 (`67d34e6c1`) is
LIVE — verified at the artefact, not the tag: `grep -ac` on `/proc/1/exe` for the string that
commit introduced → **1**, with a pre-existing string as positive control (**1**) and a
nonsense string as negative control (**0**). The tag alone would not have said this.

### 1. Council round 1: APPROVED, and its conditional objection was true

`cd6a5ef6` r1 (2026-08-22 18:10Z): **approved with 3 advisory objections, none high.**
`bug_historian` (edit 4) and `guardian` (edit 1), both medium, both named the same thing —
pass 0 lives in `ConvertTemplateToInstanceScope`, so **both armed birth guards inherited a
live behaviour change tested only at the converter**. `bug_historian` put the risk as a
conditional: *"if any caller treated the old refusal as an expected outcome…"*.

**I ran the census rather than answering the conditional in prose.**
`grep -rn RefusedReason --include="*.go"` → exactly **ONE** production consumer routes on the
refusal TEXT (`tool_birth_instance_scope.go:111`), and that arm returns the caller's bytes
**verbatim in both armed and unarmed mode**; the three other sites only log it. So the
conditional was real — and worse than the seats could see. A template with **no literal ids**
carrying `{{.ComponentID}}` where pass 0 cannot rewrite it reaches the same empty-harvest
return, matches that arm, and is persisted verbatim under a comment reading "nothing to
collide on" — the exact collision the guards exist to stop, arriving *through* the guard.
Round 1's own refusal test could not catch it: all three fixtures carry a literal `id="wrap"`,
so none reaches that return.

### 2. My fix re-introduced the failure mode it was closing — and a seat caught it

First cut: two guards in series (the reason names the placeholder; the arm excludes any reason
naming ComponentID), both mutation-proven independently necessary. Round 2 approved it with a
**medium** from `bug_historian`: that is still a **string-keyed cross-file contract**, so a
rewording silently re-opens the hole, and *a typed field was available at zero cost* because
this very change had already proved the report struct additive-safe.

The seat was right, and the sting is that **my own LANDMINES entry, written that morning, says
"route on a field, not a string"** — I wrote the trap down and then did not apply it to the
code I was writing at the time. `InstanceConversionReport` now carries `NoLiteralElementIDs`
and `ComponentIDUnswappable`; the guard reads those; `RefusedReason` is prose again. **Zero
text routing on this seam.** Mutations E and F both fail as predicted (`0e6c62168`).

### 3. The conversion is done at the CORPUS, and the framework propagates it itself

4 of the 5 templates converted through the fixer (`SQL_2026-08-23_seed_…`): generic-text-block
(179 placements / 152 pages / 21 sites), faq (82/82/15), mechanism-flow (6/6/3),
evidence-timeseries (3/3/3) — all counts **as of 2026-08-23**. All five were the same shape:
one well-formed templated wrapper id, no script, no lookups.

⚠ **The completed item's result reads `fixed: true` with EVERY counter 0** —
`ids_declared: 0`, `id_attrs_renamed: 0`, `hash_refs: 0`. Without `templated_id_swaps` (added
this round, not yet rolled) a real conversion is **indistinguishable from a no-op at the work
item**; the only reason I know it worked is that I read the template. That is precisely the
argument the round-2 submission made, confirmed live an hour later.

**RESIDUAL: `pricing` (row 6175e049) is NOT converted.** Active, same placeholder, **zero
placements** — and `site_work_items.site_id` is NOT NULL with the site only reachable through
a placement, so there is no honest site to file it against. Inert today (nothing renders it).
**It is a precondition of deleting the ComponentID bindings**: retire them while this row still
spells the placeholder and its first placement renders `id=""`.

**The fixer files its own page-scoped rerenders.** The four conversions produced **219
`page_rerender` work items** plus 2 `section_edit` items for owned pages, draining through the
normal queue — 39 of 270 placements converted in stored HTML within the hour. I did not need
to fire anything by hand.

### 4. Two wrong turns of mine, both instructive

**(a) I hand-fired 11 page-rerenders that computed the work and threw it away.** My Kafka
dispatch carried `spec.reason=template_changed` (correct — it routes to `rerender_sections`)
but no page name, and `save_page_sections` — which `rerender_page_sections_action.go:402` calls
**"the ONLY writer of `rendered_html`"** — skips with `{"reason":"no page name","skipped":true,
"success":true,"sections_saved":0}`. All 11 orchestrations read `COMPLETED`, in section mode,
with no refusal, and **persisted nothing**. Harm check: all 11 pages still `deployed`, so the
cost was only my time. apis.uk appeared to work and misled me for ten minutes — because its
page is literally named `index`, and because a queued rerender (correctly shaped) landed on it
at 12:46 while I was looking. **A COMPLETED orchestration in the right mode is not a write.**

**(b) The "18 pages serve a duplicated section id" figure is a PROXY, and it is wrong.** It
counts pages carrying a repeated *component*, which is not the same as pages *serving*
duplicate ids. Measured at the artefact, all 18, cache-busted, 2026-08-23:
**12 serve duplicates** (one also serving 4 empty ids), **3 are clean** (their `content_data`
supplies its own `ComponentID`, so they already serve distinct slot-name ids), **2 return 404**
and **1 returns 302** (webdesign.uk, the parked-domain redirect). Session 10 confirmed four
pages at the artefact and the rest by inference; I repeated the inference in a plan and a
commit message before checking. Logged in `WRONG_CALLS.md`.

Also caught in passing: a first version of that census dropped `function` from the `GROUP BY`
and returned **45** — a page with one `faq` and one `generic-text-block` counted as a repeat,
which it is not. The 18 was the honest number for the question as posed.

### 5. Verified at the artefact — apis.uk/index.html

Before: six identical `<section id="8d81e665-…">`. After: `c-generic-text-block`, `-2`, `-3`,
`-4`, `-5`, `-6` — **6 distinct, 0 duplicates, 0 empty**. Prose 8,722 → 8,938 chars (intact).
Bytes 13,326 → 65,164, and **that is not my change**: 1 `<style>` block became 5 styles + 3
scripts and the headings changed, because the page was serving copy older than the database.
A stale page holds every improvement since it last rendered — never size a rerender by your own
change. Its `build_status` went to `needs_rebuild` from a **pre-existing** 7-of-8 plan
shortfall (7 `page_components` rows against 8 planned), present in the run *before* the
conversion too.

### 6. What is NOT done

- The other ~230 rerenders are still draining; verify at the artefact, not at the queue.
- `pricing` unconverted (§3) — blocks the binding deletion.
- `{{.ComponentID}}`'s bindings and the dead `assemble_from_library` `ReplaceAll` are **not**
  deleted. Order: corpus converted → pages rerendered → *then* delete.
- **The occurrence-0 weakness is untouched.** `BindSingleSectionInstanceToken` supplies
  occurrence 0 to `RenderComponentAction` and the section editor, correct only while a
  component appears once per page — which these five templates are the ones to violate. A
  single-section edit to one of the 12 re-collides, detectably. RFC_032's next step.

## 2026-08-23 (session 11, later) — `pricing` converted, the bindings retired, and the conversion turns out not to HOLD

Owner asked for both: convert the `pricing` row, then sort out the bindings.

### 7. The census found THREE rows, not one

My earlier "only `pricing` remains" was an `is_active`-filtered query. Scanning **every live table
with a template-or-html column** (a `DO` block over `information_schema`, backup tables excluded by
name — not a guess at which table matters) found the placeholder in exactly two places:
`component_versions.html_template` (history, which is its job) and `content_components.html_template`
— **3 rows**: `pricing` (ACTIVE, 0 placements), `header` and `footer` (both INACTIVE, 0 placements,
same 2025-11-21 library seed). All three the same shape: one well-formed templated wrapper id, no
script, no lookups.

All three converted (`SQL_2026-08-23b_…`). The inactive two deliberately: they cannot render today,
but after the bindings go a reactivated row renders `id=""`, which the detector cannot see. **Count
of rows spelling `{{.ComponentID}}`: 3 → 0.**

Not hand-written: each conversion is the output of the REAL converter and passed the REAL gate, with
a CONTROL proving the gate REFUSES each unconverted original (a gate that accepts everything vouches
for nothing). Only the persistence is by hand, mirroring `writeScopedTemplate` (snapshot at MAX+1,
then UPDATE), each UPDATE guarded on the row still spelling the placeholder.

### 8. A replay hole that would have undone this morning's work

`247_mechanism_flow_component.sql` and `250_evidence_timeseries_component.sql` both create their
component with `id="{{.ComponentID}}"`, both carry `ON CONFLICT (name) DO UPDATE`, and **neither is
recorded in `schema_migrations`** — and the runner's entire pending test is "filename not in the
ledger" (`run-migrations.sh:225`). A migration run would have replayed them over two templates
converted hours earlier. One line changed in each; every other replay hazard those files carry is
left exactly as it was rather than quietly widened.

### 9. Two of three bindings deleted; the third is a HEAD-safety refusal

`rerender_page_sections_action.go:641` (live binding) and `assemble_from_library.go:309` (dead
`ReplaceAll`) are gone. **`v3_site_actions.go:2385` is edited in the working tree and NOT committed**:
that file carries another session's uncommitted work — five further hunks calling three functions in
`platform/orchestration/actions/section_metadata_keys.go`, which is **UNTRACKED and absent from
HEAD**. A pathspec commit takes the working-tree file whole, so committing it would put calls to
undefined symbols on the shared branch and break the build for every session. Inert to leave (zero
templates reference the placeholder), named so it is finished rather than forgotten.

⚠ **The first HEAD extract was stale and that is what surfaced this.** The build failed on symbols I
had never heard of; the instinct "my change broke it" was wrong both ways — it was neither my change
nor a stale extract alone, it was another session's WIP inside a file I had edited. Re-extract from
CURRENT HEAD before believing a build failure, then read `git diff --numstat` on your own files.

### 10. ⚠ THE CONVERSION DOES NOT HOLD — and my §9b trigger claim was too narrow

Final artefact sweep of the 12 pages that served duplicates this morning: **9 FIXED, 3 colliding
again.** All three carry the same signature — two sections both stamped the occurrence-0 token.

The 9 were written 12:50–13:24 by the page-rerender queue, which uses `InstanceCounter` and is
correct. The 3 were rewritten at **17:41–17:51** by `content_rewrite` items from **`backfill-353`**,
an unrelated lane's content backfill running through `page-build-handler`. Not a rebuild I caused,
not a section edit, not the owned-page route (all three are `rebuild_policy='generic'`).

So the defect is **any path that renders a section on its own**, not "a build" as RFC_032 §9b said.
Those are ordinary high-volume operations, so a page fixed today is fixed only until the next
content operation touches it, and nothing warns anyone when it flips back. Corrected in RFC_032 §9d.

**The measurement trap this creates, which is the transferable part:** all four templates are
converted, and 244 of 275 placements carry a `c-` prefix — both numbers stay healthy while pages
re-collide, because `c-generic-text-block` twice IS a `c-` prefix twice. **Only distinctness per
page, at the artefact, can see it.** Any dashboard built on the corpus counts would have reported
this work as finished.

### 11. Queue outcome and the residuals that are NOT mine

325 rerenders complete, **3 failed**, queue drained. The failures are other lanes' territory and the
guards are behaving: `webdesign.uk/index.html` ×2 refused by the claims floor (15 banned claims in
its content), `loancalculator.co.uk/tools/loan-vs-savings.html` with 1 of 6 sections unable to
resolve a component. 31 placements remain unconverted across 8 domains — a mixed bag of those
failures plus pages whose rerender completed without rewriting (the escalate-to-writer path,
including the four idea.uk pages serving `<section id="">` since 2026-08-12). **Fixing the cause did
not repair the casualties**, and the casualties are named rather than assumed to heal.
