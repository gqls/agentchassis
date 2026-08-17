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
