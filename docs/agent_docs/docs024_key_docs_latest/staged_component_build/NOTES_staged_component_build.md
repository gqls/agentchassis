# NOTES — staged component build

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep. **The missteps are the point, not an appendix.**

---

## 2026-07-30 — lane adopted; the blocking unknown resolved in two queries

Owner: *"This provenance and ladder project is now this lane's project. Please adopt
it and update the document and write a summary of where we are."* So
`staged_component_build` moves from "proposal awaiting a separate thread" to an owned
lane of `brochure_component_library`. `features_open/027` updated accordingly; the
standing five created here.

**Resolved the `[UNVERIFIED]` the proposal named as the first action.** The question
was whether `doc_plans` fits a component without schema change. Answer: **no, and the
obstruction is one line.**

```
doc_plans_subject_type_check  CHECK (subject_type = ANY (ARRAY[
  'tool','pipeline','experience','action','experience-pattern']))
doc_notes_subject_type_check  CHECK (subject_type = ANY (ARRAY[
  'tool','pipeline','experience','action','experience-pattern','landmine']))
```

**Neither allows `component`.** An insert fails outright. Extending it is additive and
has a **four-times precedent** — migrations `163_doc_subjects_experience.sql`,
`184_travelling_action_subjects.sql`, `218_experience_register_substrate.sql`,
`270_doc_notes_landmine_subject_type.sql`. 270 is the template and is well written:
it guards that the constraint is the shape it expects, refuses if already applied, and
is re-run safe.

**The better half, and it improves the design rather than merely unblocking it:**
`doc_notes` has a `site_id` column and `doc_plans` does not. That is exactly the split
a component needs and it was already there — a component's template is fleet-shared
(one `content_components` row serves 11 sites for `info-card-grid`), so a site-less
PLAN is *correct*, while S4–S7 are per-site facts that land in NOTES. **The PLAN is the
fleet-wide contract; the NOTES are the per-site verdicts.** No column has to be added;
only the constraint changes. Mild evidence these tables were designed for something of
this shape.

Live population when read (**re-run, do not quote**): doc_notes pipeline 362 / tool 73
/ landmine 57 / experience 56 / action 16; doc_plans experience 59 / tool 35 /
experience-pattern 24 / action 4.

**A trap spotted before it fired, worth more than the resolution.** The `doc_notes`
constraint must be re-added **with `'landmine'` kept**. Copying `doc_plans`' array
across — the obvious way to make two tables agree — would drop it and orphan **57 live
rows** written by two other threads. The two tables do not have the same constraint and
should not be assumed to. Written into the RUNBOOK beside the DDL.

**Decided NOT to create `272_*.sql` in `sql_for_agents/`**, and this is a real decision
rather than caution. The migration runner takes *every* pending file in a directory, so
an unreviewed numbered file sitting there could be applied by an unrelated session's
`--apply`. **A staged artefact another session can apply by accident is not staged, it
is armed.** The DDL lives in the RUNBOOK and gets a number when it goes to the council
gate.

## 2026-07-30 (same session) — what the review of the tools-chain report contributed

Reviewing `webdesign_tools_repair`'s rewritten report (asked for separately) produced
three things this lane now depends on, so they are recorded here rather than only in
`brochure_component_library`'s notes.

**1. A live hazard that threatens every gate in the ladder.** An unknown check type is
**skipped, not failed** — `run_checks_action.go`'s type switch ends
`default: skip(ch.ID, ch.Type+" not implemented")` — and that report's own **G4**
(verified in code) says an all-skipped result set yields `len(Failed)==0` → **PASS note
plus a 7-day cooldown**. So a gate authored against a check type the running binary
does not carry **passes vacuously and suppresses its own re-check for a week**.

For a *ladder* this is worse than for a checklist, because stage N's pass is what
licenses stage N+1. Hence PLAN D3: **a stage that cannot evaluate its question is
inconclusive, not passed.** Filed to `LANDMINES.md` (footprinted on the `default:`
skip arm) rather than as a bug, because nothing is broken — the trap is authoring
against it.

It is live right now, not hypothetical: `has_visible_area` (TL-034) was committed
07-30 15:19 (`1850acb07`) and the running `browser-runner-adapter` pod predates it. So
the newest and most useful check type for this ladder is currently **exactly the one
that would silently skip.**

**2. A method misstep of mine, caught by a control — this is the transferable part.**
My first pod-grep used the check-type names themselves: `has_visible_area` **0**, but
also `selector_count` **0** on a binary that demonstrably supports `selector_count`.
**Go compiles short string literals to immediate comparisons that never reach rodata**,
so a short marker returns 0 on a fully-supporting binary. Had I stopped at the first
result I would have reported the pod as missing a check type it has. Re-ran with long
markers unique to the change (`"too small to see or click"`,
`"A collapsed flex/grid child is the usual cause"`) against three long pre-existing
controls (`"page overflows horizontally"`, `"but a parent CLIPS it"`,
`"in the live DOM after settle"`): new **0/0**, controls **1/1/1**. *A negative from a
short marker is worthless.* Also: that image has **no `strings` binary** — use
`grep -ac` on the binary directly.

**3. Two gate-authoring rules, better stated by the other lane than by me**, adopted
into the proposal: verify through the visitor's gesture and never the subject's
internal functions (and if the vocabulary cannot express the gesture, that is a missing
check type to record as a deferral, **not** a licence to substitute a function call);
and a gate must assert the **terminal value**, not the first observable state change.

**The convergence is the strongest evidence this lane has.** That lane verified
`pasteboard` by calling `addItem()` and `logic-architect` by calling `loadTemplate()`;
both returned the right answer so both "passed", while a visitor could reach neither.
This lane forced `.open = true` on DOM nodes and called it verified for four rounds.
**Those are one defect class — verifying through a privileged path the visitor does not
have — not two similar mistakes.** Two lanes, two subject types, one day, arrived at
independently. Its generalisation is the best single line for the browser tier:
**a property of the composition can only be checked in the composition.**

**Also corrected in the proposal**, from the same review: two figures in the tools
report had drifted inside its own editing window (fence count 23 → **25**, re-run
here), which is a live demonstration of its own *"never quote a count you didn't
re-run"* rule and the reason this lane's RUNBOOK marks every figure as re-runnable.

**Not adopted, deliberately:** `features_open/015`, the site maturity ladder. The tools
lane's decomposition — **015 is the rung vocabulary, 027 is the gate mechanism, 026 is
the missing instrument** — makes the three composable rather than merged, so this lane
can proceed without owning the site-scale question. Recorded as PROPOSED; whether 015
stays a separate thread is the owner's call, not ours to take by adopting it.
