# NOTES — contact-block transport (bugs_open/228)

Append-only, newest at the bottom. Technical log: what was tried, what the
system actually said, missteps included.

---

**2026-08-08, session start.** Surveyed `bugs_open/` for an unowned bug via
`scripts/who-owns.py` against every numbered file. First pass showed almost
every bug as "OWNED or recently active" — turned out to be a **false-positive
trap in the script itself**: its `subject_commits` check does a bare substring
match of the bug number against commit subjects in the window, so e.g. bug
`228` matched commits about unrelated things that happened to contain the
digits `228` (a site count, a line number). Re-ran the scan filtering on
`"(none identified)"` in the "likely OWNING workstream(s)" section instead of
trusting the verdict line — only `161` (already fixed-at-source, awaiting
redeploy per memory) and `228` came back genuinely unowned.

`228` (`contact-block` fabricates "message sent" with zero transport) was
filed same-day by the `staged_component_build` lane, found while writing that
component's acceptance fence. Re-verified live myself before starting: curl of
`robot-hands.com/contact.html` (form has no `action`/`method`) and of the
served `/tools/assets/contact-block.js` (2100 bytes, zero matches for
`fetch\(|XMLHttpRequest|sendBeacon|form\.submit\(`). `git log --since="3 hours
ago" --all` showed nobody had touched contact-block since filing.

Traced the root cause one level below the bug file: the chassis already has a
proven mechanism for "dead contact form, no server backend" —
`sanitiseFormAction`/`deliverableFormAction` in `component_library.go`, which
rewrites a known-dead `form_action` value into a real `mailto:`. It only fires
if `ctx.ContentData["form_action"]` is *present* — contact-block's
content-generation schema never asks the LLM to author that field, so it's
absent, not empty, and the sanitiser's own presence-gate silently declines.

Traced the deployment mechanics for the JS asset too:
`RerenderSinglePageAction`'s `collectJSAssets` reads `content_components.
js_content` per component used on a page and writes `tools/assets/{function}.
js` into the same git commit as the re-rendered HTML — the same page-rerender
handler `check_contact_form_undeliverable.go`'s own auto-remediation already
dispatches for the analogous `contact-form` case. No new deploy plumbing
needed once the DB row is fixed.

**Design delegated to Fable** (per the task's own instruction) to draft
`PLAN_2026-08-08_contact_block_transport.md`, given the full research above.
Fable's draft caught two things I'd fed it wrong, both independently
re-verified before accepting:

1. **Three live pages was wrong — it's two.** `finetuning.uk/case-studies.html`
   is a placement-drift row (the `page_components` row is active but the
   served page carries no contact-block markup at all).
   `[VERIFIED]` `curl -s https://finetuning.uk/case-studies.html | grep -c
   'data-component="contact-block"\|cb-contact-form'` → `0`. This matched a
   same-day self-correction that landed in the bug file concurrently from the
   filing lane — no conflict, independent confirmation.
2. **The old-binary regression shape is `action=""`, not literal
   `<no value>`.** `RenderTemplateReportingMissing` strips the `<no value>`
   artefact on purpose (dead-control logging, bugs_open/018).
   `[VERIFIED]` read `component_library.go:1019-1021` directly — confirmed the
   `strings.ReplaceAll(result, "<no value>", "")` call. Same ordering
   constraint either way, but the plan's own claim needed correcting before I
   trusted it.

**Consumer enumeration run before submitting to council** (2026-07-29 owner
ruling #3 — tell consumers, don't just measure): `SELECT function FROM
content_components WHERE is_active AND html_template LIKE '%form_action%';` →
exactly one row, `contact-form`. Also confirmed no existing concept-register
entry for `form_action` (`grep -rl form_action
docs/agent_docs/docs026_concept_register/register/` → nothing) — the
mechanism postdates the 2026-07-13 extraction freeze and had never been
registered.

**Implementation.** Added the seeding block to
`RenderTemplateReportingMissing` (before `contextToInterfaceMap`/`
contextToMap` run): if the template string contains `"form_action"` and
`ctx.ContentData["form_action"]` isn't already present, seed `""` (already a
recognised `nonDeliveringFormActions` value) with a Debug log line as the
pod-grep target. `go build` clean. Existing test suite green, including the
one invariant that most needed to survive:
`TestSanitiseFormAction/a_component_with_no_form_does_not_acquire_a_form_action`.

Added three new tests exercising the seeding at the `RenderTemplateReportingMissing`
entry point (not just the lower-level `sanitiseFormAction`): seeds-and-sanitises
when the template references the field; stays honest (`action=""`, no
fabricated mailto) when no address is resolvable; does NOT seed for a template
that never references the field. **Mutation-checked by hand**: `git stash`'d
just the production file, re-ran the three new tests — the "seeds and
sanitises" test failed exactly as expected (`action=""` instead of a real
mailto), the other two still passed. `git stash pop` restored the fix. This is
the load-bearing proof the tests actually guard the change, not just green by
construction.

Registered the widened mechanism as `LNK-031` in
`register/link-management.md` (count bumped 30→31) — covers what it is, the
gap this closes, the fix, the measured single consumer, and the landmine
(old-binary + new-template-reference ships `action=""` silently, not an
error).

**Committed** `component_library.go` + `component_library_form_action_test.go`
+ `link-management.md` together (one commit, per the ordering-exemption's
surviving condition: register in the SAME commit that ships the seam) as
`85390ee33`.

**Council.** Submitted before committing (`097_TRIGGER_council_review_v1.sh`),
correlation `46f87e4c-05fc-4a5c-bd6a-93a073b63253`. Queue was clear (LAG 0) at
submission. Committed with `Council-Submitted: 46f87e4c-...` trailer — first
attempt used the literal string `Council-Submitted: pending`, which the
`commit-msg` trailer gate correctly rejected (not a UUID/hex prefix, resolves
to nothing at report time) before any commit was made. No harm done — the gate
blocked *before* creating the commit, not after. Logged as a misstep below and
in `WRONG_CALLS.md`.

**Build.** `makefile`'s `IMAGE_TAG` had an uncommitted local bump to
`v1.0.1270` from another session already in the tree (visible in `git diff`
before I touched it) — did not reuse it, to avoid any collision with whatever
that session intended. Checked the actually-*deployed* tag first
(`v1.0.1269`), bumped to `v1.0.1271` (one past the highest number seen
anywhere, deployed or staged), built (`make build-agent-chassis`, printed
`Building agent-chassis from committed ref HEAD = 85390ee33`, confirming the
right commit), pushed (`docker push docker.io/aqls/agent-chassis:v1.0.1271`),
committed the makefile bump separately (`10877c525`).

**Deploy — asked rather than assumed.** Memory
([[releases-are-whole-fleet-make-release]]) records that a single-service
`kubectl apply -k` was blocked by a permission classifier on 2026-08-03, and
that the owner runs the whole-fleet `make release redeploy-agents` himself.
Asked the user directly rather than attempt a deploy that precedent says is
blocked. User's answer: "A fresh chassis build has been deployed."

**Pod-verify — the fresh build does NOT carry the fix.** `[VERIFIED]`
`kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis
| grep -c "seeded empty form_action for sanitiser"'` → **0 on both replicas**
(`agent-chassis-5c8776654c-wml6w`, `agent-chassis-5c8776654c-zhz2g`). Negative
control (a deliberate misspelling, expect 0) also read 0 on both, as it
should. The deployed tag is `v1.0.1270` — NOT the `v1.0.1271` I built from
`85390ee33`. Pod start time `2026-08-09T08:49:38Z`/`08:49:59Z`, my commit
timestamp `08:39:51Z` — the pod started 10 minutes after my commit landed,
so on timing alone it's ambiguous, but the binary evidence is decisive
regardless of timing theory: **the fix is not live.** Likely explanation:
that release build's `git archive HEAD` ran before my commit landed, and
build+push+rollout latency pushed the pod's *start* time past my commit
timestamp without the *checkout* being past it. `git log --all --oneline
--since="2 hours ago"` shows another session's commit (`fe202d7ea`, "225:
FIXED + LIVE...") landed AFTER mine — so HEAD now unambiguously contains my
fix; a release run from current HEAD, whenever it next happens, will pick it
up. Reported this to the user plainly (not guessed at) and asked for one more
release run. **Do NOT apply the `content_components` row change for
contact-block until this re-verifies positive** — that's the hard ordering
constraint the plan names, and it isn't satisfied yet.

**Council round 1 (correlation `46f87e4c-05fc-4a5c-bd6a-93a073b63253`): REVISE.**
Gating objection from `editquality`, and it was correct: the submitted plan
only included edit 1 (the Go seeding change), and its own risk section
admitted contact-block's template doesn't reference `form_action` yet — so as
submitted, the change was a no-op for the actual diagnosed bug, benefiting
only `contact-form`. Resubmitted (`RESUBMIT_CORR`) with the DB-side edit
included as edit 4, describing the pending `content_components` change and
naming the ordering constraint explicitly.

**First resubmission attempt failed client-side-passed-but-server-rejected**
(`complete_invalid`, no verdict row — indistinguishable from "still queued"
until checked): edit 4's `file` field was a descriptive string
(`"content_components (function='contact-block', ...)"`, with spaces), and
`diagnose_persist_fix_plan` requires a real repo-relative path with no
whitespace. Not one of the 097 script's documented "three type traps" —
logged as a new RUNBOOK gotcha. Fixed by pointing `file` at real repo paths
already written to disk (`RUNBOOK_228_contact_block_transport.md` for the
html_template edit, `js_content_after_228_fix.js` for the JS edit) instead of
a label.

**Council round 2: REVISE**, gating objection from `render_guardian`. This
round's feedback was unusually substantive — 14 reviewers, several real
catches:

- `render_guardian` (**HIGH**): editing `content_components.html_template`/
  `js_content` does not itself propagate to already-*rendered* pages —
  assemble-mode rerender redeploys stored HTML unchanged. My plan named the
  rerender dispatch only in RUNBOOK prose, never as a submission edit with its
  own verification. Exactly the `bugs_open/024` false-green class: DB row
  fixed, live page still serving the old markup, nothing in my own plan would
  have caught it.
- `debug_historian` (**HIGH**): the html_template/js_content mutation was
  "surgical replace()" in prose only — no needle-occurrence guard, no backup,
  no `RETURNING` postcondition, no separate rollback file. Named this the
  **NEEDLE-GATE SQL SURGERY** shape: `replace()` silently no-ops on a missed
  anchor while still reporting `UPDATE 1`.
- `prior_art_librarian` (medium): my "confirmed not live" claim rested on a
  2-replica `-l app=agent-chassis` grep — and there's a standing LANDMINES
  entry saying that selector undercounts (41 pods run the same binary
  fleet-wide, measured 2026-08-05: 7 pods on a new tag, 34 stale, invisible to
  the label-scoped check). I had read `who-owns.py`'s landmine class earlier
  in this same session and still walked into this one — the platform's own
  documented traps are numerous enough that reading the index doesn't
  guarantee recalling the specific one that applies.
- `guardian` (medium): blast radius of the `form_action` seeding change
  needed bounding without the `is_active` filter too, and the config_change
  edits needed their owning pipeline named explicitly (which handler actually
  turns the DB row into a live page).
- `tooling_provenance` (medium): never checked `doc_notes`/`doc_plans` — the
  platform's own travelling-docs mechanism — for prior context on this exact
  chokepoint before planning.
- `bug_historian` (medium, architectural, correctly not gating): the generic
  `<no value>`-stripped-to-empty-string behaviour in
  `RenderTemplateReportingMissing` is the same *shape* as the platform's
  worst-documented recurring bug class (Go template `missingkey=zero`
  rendering required-but-missing fields empty with no error) — this fix
  extends that mechanism narrowly for one field rather than closing the
  generic exposure. Correct observation; accepted as a named, deliberately
  out-of-scope follow-on rather than folded in.

**What I did in response, all verified before resubmitting, not just
described:**
- Re-enumerated chassis-binary pods by **image**, not label, per the
  landmine's own remedy: 54 total (49 Job + 4 ReplicaSet on `v1.0.1270`, 1
  stale Job on `v1.0.1269`, **zero** on `v1.0.1271`) — same conclusion as the
  2-pod check, now actually evidenced against the documented undercount risk.
- Re-ran the `form_action` consumer query **without** the `is_active` filter:
  still exactly one row, `contact-form`.
- Queried the full blast radius of `contact-block` itself directly against
  `page_components`: exactly 3 rows, none `locked_at`-frozen (the 2 live pages
  + the already-known `finetuning.uk` drift row).
- Queried `content_data IS NULL` for the contact-block section on both target
  pages (the `049b_deploy_single_page.sh` script's own documented escalation
  gotcha) — both `false`, confirming the rerender stays in the light
  `section_data_resolved` path, no LLM regeneration risk.
- Queried `doc_plans` for `subject_key='contact-block'`: found the component's
  own acceptance-fence PLAN row (authored by `staged_component_build`,
  2026-08-08), which states it "deliberately does NOT assert the success
  message" — confirms my planned JS change doesn't conflict with that
  contract.
- Wrote two real scripts (`apply_228_contact_block_fix.sh`,
  `dispatch_228_rerenders.sh`) implementing the disciplined procedure instead
  of prose: needle-count guard, fresh backup, auto-generated rollback file,
  `RETURNING`-gated manual commit, and a dispatch+deployed-page-verification
  step. Neither has been run.

Resubmitted (round 3) with these two scripts as new edits (`add`, not
`modify`) and the above evidence folded into `grounded_in`/`risks`. Verdict
pending as of this note.
