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
