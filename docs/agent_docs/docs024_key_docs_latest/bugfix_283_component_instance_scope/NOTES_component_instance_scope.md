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
