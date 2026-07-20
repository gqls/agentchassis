# NOTES — work-item completion integrity (append-only, newest at the bottom)

---

## 2026-07-18 — session "bugfix thread2"

Picked `bugs_open/017` off the queue. Chose it over 013 (implementer gofmt) because leg 2
is a fleet-wide correctness lie rather than a loop-yield cost.

**Verified the filed claims before working from them — and one was wrong.** The report
said Defect 1 was drift between two hand-maintained rosters. Read
`actioncheck/actioncheck.go:20`: `IsLocalAction` delegates to a checker `registry.go`
installs at `init`. Read `local_actions.go:185-188`: the map's own lookup is commented out.
Grepped the symbol repo-wide: zero live references. So the "second roster" was dead code,
not a drifting roster. The misinformation source is the `batch_webscrape_action.go` header
comment telling authors to "register in TWO places" — which had also reached two live guide
docs. Deleted all of it.

**Scale was wrong too.** The report recorded 2 affected items; the sweep found 54 across
6 sites and 4 item types, back to May. One of the 54 is a *different cause* with the same
symptom — a seed naming `render_js_snippets` where the registry has
`render_js_snippets_for_site`.

**Chose the guard predicate against live data rather than intuition.** `SELECT DISTINCT
response.status` over the whole table returns exactly one value: `'failed'`. 2905 completed
items carry no `response.status` at all. There is no ambiguous middle population, so keying
on an explicit failure verdict cannot mis-fire. Over 30 days the guard would have blocked
6 of 1662 completions, all genuine.

**Negative control on the new parity test.** Removed the registry entry → test fails and
names the action; restored → passes. A passing test proves nothing until you have watched
it fail.

### MISSTEP 1 — I called a queued orchestration a dropped one, and it cost three council runs

Council round 1 returned REVISE in ~9 minutes. Round 2's dispatch produced **no**
`orchestration_state_audit` rows after two minutes, where round 1 had produced its first
within ~10 seconds. I concluded the spawn had been silently dropped — CLAUDE.md documents a
real drop mechanism, which made it feel confirmed.

It was **queued**. Submitted 16:41, first audit row 16:57 — ~16 minutes under backlog. In
between I resubmitted three times, twice shipping a "fix" for a hypothesis I had not
tested: first that the ~27KB payload exceeded what `kubectl run -i` stdin carries (I even
rebuilt the submission smaller), then that `RESUBMIT_CORR` was broken (I ran a
fresh-correlation control). Both wrong. All four submissions were queued and all ran.

The lesson is not "be patient" — it is that **"no rows yet" is consistent with every
hypothesis**, so it cannot discriminate between them. The discriminating query asks when
*other* orchestrations started. Mine was sitting in that list, 16 minutes late, while I was
proving it had never arrived. Filed as 016b §9 + memory.

### MISSTEP 2 — I asserted a structural claim from filenames, and dismissed the council for catching it

I claimed the delivery-vs-verdict conflation is "structurally unique to
`CompleteWorkItemAction`", based on a regex sweep finding 8 paths that write
`status='complete'` plus reading ~18 lines of four of them. **I never opened the three
admin paths.** I inferred "human HITL decisions" from their filenames.

`bug_historian` (low) and `guardian` (medium) both objected that this rested on an
author-run prose audit rather than an independent check. I recorded that in the handoff as
*"verification-of-my-audit asks, not defects"* and declined to spend another run. That
characterisation was wrong: an unverified structural claim IS the defect, and it had
already reached a commit message, a bug handoff, a §9 guide entry and two `doc_notes`.

> **CORRECTED 2026-07-19:** opened all three admin paths. The claim **holds** —
> `confirm_work_item_handler.go:212`, `site_admin_handlers.go:793` and `:987` each build
> their result with `jsonb_build_object` from human input (`'resolved_by','admin'` /
> `'approved_by','admin'`) and never read or store a `response` envelope, so none can carry
> a failed verdict. **Caught by:** re-reading CLAUDE.md at the owner's prompt, whose
> diagnosis section had been inverted that day with a correction describing this exact
> failure mode — "a confident structural claim built from grep hits whose functions it had
> never opened". The conclusion survived; the method did not. Being right by luck is not
> the same as having checked.

### Process failure — I did not create these docs until prompted

The standing-five directive (owner, 2026-07-18; cadence 2026-07-19) says to create them at
the START and update as you go. I wrote none until the owner asked me to re-read CLAUDE.md
at the end of the session. Everything above was reconstructed from scrollback, which is
precisely the failure the cadence rule exists to prevent — a doc written at the end is a
report, and reports lose the wrong turns.

### Also worth knowing

- My `registry.go` edit was swept into another session's commit `06376bcbf` mid-task. The
  git rules cover this: nothing lost, finish and commit the remainder, say so.
- Two different cases share the number `017` (one in this dir's index). Resolve by slug.

---

## 2026-07-20 — deploy verification and closure

Owner shipped the chassis image: **v1.0.1139**, pod `agent-chassis-645674b498-rndg9`.

### MISSTEP 3 (caught before sign-off) — my first pod-grep was worthless

Ran the CLAUDE.md deploy check and got `fix_forced_text_colors` → **1**. Nearly signed off
on it. Then realised: that string was **already in the binary before the fix**, emitted by
the action file's own `RegisterActionInputSpec("fix_forced_text_colors", …)` call, which has
existed since the action was written. A pod running the OLD image passes that grep
identically. The check confirmed the subject existed, not that my registry entry shipped —
which was the entire fix for leg 1.

Re-verified with **discriminating** strings, ones that cannot exist unless the change
shipped:

| string | count | proves |
|---|---|---|
| `Strip forced child-text colours that override the --section-*…` | 1 | the registry entry itself (leg 1) |
| `completion blocked: handler saga reported failure` | 1 | the guard's blocking path (leg 2) |
| `unrecognised handler verdict` | 1 | the round-2 `agent_error_log` follow-up |
| `verifyBeforeComplete` 5, `CompleteWorkItemAction` 6 (controls) | >0 | `strings` works — a 0 above would be real |

Note the negative control `LocalActions` → 0 is *weak* evidence and I am not leaning on it:
a Go map's variable name need not appear in the binary at all, so 0 would be expected
either way. The Description string is the load-bearing proof.

Filed as 016b §9 "The pod-grep passes even when nothing shipped".

### Live state

Sweep = **0**. 11 completions through the new path since deploy, zero false-positives —
which is the regression check passing. 0 `WORKFLOW_INVALID` anywhere.

**Deliberately NOT claimed:** the guard's *blocking* path has not fired in production. No
saga has reported failure since deploy — expected, since registering the action removed the
generator of 49 of the 54. Blocking logic rests on 11 unit subtests with a negative control,
not on a live firing. Said so in the case file rather than implying live proof.

**Decided not to force a dispatch to manufacture proof.** Three stale `detected`
`hardcoded_section_colors` items exist (robot-hands, vonc, gamesdesign), and dispatching one
would exercise leg 1 end-to-end — but it would also run a colour-stripping action at a live
site, which 017 itself judged misconceived on robot-hands, and the owner's ruling was
"mark them failed and start fresh", not "re-run them". The registry entry being provably in
the binary makes the "requires a topic" failure structurally impossible; that is enough
without editing a live site to prove it.

Moved to `/bugs_closed/017_…` (number and filename preserved, per that dir's rules).

---

## 2026-07-20 (later) — handed off

Wrote `HANDOFF_2026-07-20_start_here.md` as the cold-start entry point and pointed PLAN,
SUMMARY and README_where_we_are at it. Nothing is in flight: no pending dispatches, no
uncommitted work of mine, no background jobs.

**Two inbound assignments discovered in this directory during closure**, both placed by the
reasoning-dataset thread while phase 1 was running, neither started:

- `HANDOFF_2026-07-19_verifier_absent_row_defect_and_coverage.md` → `bugs_open/032` +
  `bugs_open/021` §INSTANCE 2 (one `RegisterVerifier` call for ~50 item types). Plan
  `submission_B`, 2 council rounds, both REVISE, objections enumerated.

  > **CHECKED BEFORE WRITING THE HANDOFF — and the inbound doc was already stale.** It
  > describes 032 as an open defect with a fix "drafted". The fix is **written and
  > committed** (`a467baa11`), in the conservative shape it recommended. I nearly told the
  > next thread to go and implement it. Applied the discriminating-symbol rule from misstep
  > 3 in the other direction, to prove something is NOT live: pod-grep for
  > `"genuinely fixed or silently deleted"` → **0**, with my own 017 guard string → **1**
  > as the positive control, and the commit (10:33 UTC) postdating the pod start (07:35
  > UTC). So 032 is fixed-but-INERT — the state 017 was in yesterday — and correctly stays
  > in `/bugs_open/`. Next action is an image roll, owned by
  > `empty_sections_loop_integrity`. The lesson generalises: **an inbound handoff is a
  > claim about the past; verify its state before forwarding it.** What actually remains of
  > that handoff for this thread is the 021 coverage policy, re-verified today
  > (`RegisterVerifier` still called exactly once).
- `HANDOFF_2026-07-20_submission_A_work_item_origin_provenance.md` → **owner-assigned to
  this thread on 2026-07-20**. `site_work_items.origin_correlation_id`, 3 council rounds,
  all REVISE, "two small answers away". **Verified genuinely unstarted:** the column does
  not exist on `site_work_items` and the identifier appears nowhere in `platform/`.

Note the shape of 032 relative to our own work: our PLAN had already named the gap it
exploits — *"the verifier is opt-in per item_type"* — and the council's `bug_historian`
seat found the defect while reviewing a proposal to COPY that verifier's behaviour to two
more item types. The blind spot propagates by being reused. Worth remembering when
implementing: the conservative fix (return an error, not a verdict) relies on our gate
already failing OPEN on verifier error, which is a property of the code this thread owns.

**Verified before handing over:** defining sweep = 0; no work items carry
`error LIKE 'completion blocked%'` (the guard has not yet blocked in production);
no `UNKNOWN_HANDLER_VERDICT` rows in `agent_error_log`. All three are the expected values,
and the second two are *absence of evidence*, recorded as such rather than as proof.

---

## 2026-07-20 (later) — the verifier coverage gap (bugs_open/021 §INSTANCE 2)

Owner asked to fix the gap and chose the full scope: contract widening + a
page_rerender verifier + the coverage guard. Shipped two of the three; the third is
held, and that is the substance of this entry.

**Corrected the filed diagnosis.** 021 and `verifiers.go` both attribute "one
verifier for 69 item types" to the mechanism being opt-in — *"stays at one unless an
author remembers"*. Incomplete, and it points the fix at discipline instead of code.
The contract passed only `(ctx, db, spec, logger)`. Measured over all 5,514 live
items: 2,370 specs carry `page_id`, 310 carry `component_id`, **9** carry `site_id`.
So for a site-aggregate type — `hardcoded_section_colors` files ONE item per site and
its predicate needs the site_id — a verifier was **unwritable**, however willing the
author. submission_B proposed exactly that verifier, so it was unimplementable as
specced. `VerifyTarget` fixes the real blocker.

### MISSTEP 4 — I recommended page_rerender as the easy win. It was the trap.

Reasoned from volume + identity: 1,849 of 4,644 completions, `page_id` on 1,914 of
1,929, per-target items, predicate already in `check_misdirected_cta`. I put it to
the owner as the recommended scope on that basis, and used it to argue
council-reviewed submission_B had chosen badly.

Wrote it. Tested it — six cases, all passing. Then, writing the verifier's own
scope-guard comment (explaining why an unrecognised `spec.reason` must refuse), I
had to state what the handler is responsible for — which sent me to
`check_misdirected_cta`'s header:

> *"the recompute only rewrites the CTA url fields of components in the actions
> package's ctaFieldNames set ... a misdirected link inside any other component
> (e.g. prose) ... is re-detected on the next discovery pass and escalates via the
> two-strike rule to human review — loud, not silent."*

My verifier checked **every** anchor on the page. It was **stricter than the
handler's remit**: a correctly-handled rerender with an out-of-remit prose misdirect
would verify unresolved, burn attempts, and land in `failed` — destroying the
designed escalation across 1,849 items. A regression wearing a fix's clothing.

**My tests all passed because they tested the predicate I chose, not the one the
handler implements.** That is the transferable bit: test-green says your code does
what you meant, never that you meant the right thing.

Held it on the owner's call. Removed the verifier and reverted the
`loadCTAMatchIndexFor` split with it — an uncalled helper is dead code, and dead code
misleading a future diagnosis is precisely what bugs_open/017 was. Kept
`ctaClassifyAnchor` (Run uses it) and its 9 test cases, which the check's core
classification logic had never had. Logged in `WRONG_CALLS.md`; tally row *"read the
code before asserting a mechanism"* 4 → 5.

**Also caught while writing my own tests:** one case I asserted FAILED —
`NormalizePagePath` does not insert a leading slash, so a relative href would read as
a misdirect. Checked before calling it a bug: 0 components carry relative internal
hrefs against 169 absolute. Latent asymmetry, not a live defect; documented in the
test, not filed.

**Shipped:** `VerifyTarget` contract, the predicate extraction + its tests, and the
coverage guard (69 item types classified: 1 verified, 35 mechanical, 2 no_target,
15 creation, 17 judgement). Net new verifiers: **zero**. The gain is that the gap is
now a checked, categorised decision that breaks the build when a new item type
appears unclassified — not an invisible default. Categories are marked [INFERRED]
where I did not open the check, per CLAUDE.md's new marker rule, because misstep 4
was that exact error one level down.

