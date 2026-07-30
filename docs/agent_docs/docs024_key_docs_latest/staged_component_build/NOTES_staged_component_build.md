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

## 2026-07-30 (later) — SUBMITTED to the council; and D5 was wrong, twice over

Owner: *"please carry on and take that through the council gate."* Done — correlation
**`e5673868-7c5b-489c-931a-7ba59b959b91`**, commit `c659e312b`. Budget ~30 minutes for a
verdict; the council itself takes 2–5 but the dispatch queues behind the fleet.

**Two assumptions I held at the start of this were both wrong, and checking them is the
whole content of this entry.**

**Wrong assumption 1: "the gate will refuse this, because it is a DDL file under
`docs/`."** I read the trigger script expecting a refusal —
`SCOPE_RE='^(platform|internal|pkg)/'`, and its own comment says docs never spend
council credits. I was ready to report that the honest governing path was owner approval
plus the migration guards, citing migration 270 (which was indeed applied out of band
that way). **That was wrong because my understanding of the CHANGE was incomplete, not
because the scope rule is odd.** The change genuinely touches `platform/`, so it passes
the scope check with no `FORCE`. Also worth recording: the check is
`[.plan.edits[].file | select(test($re))] | length > 0` — **any** in-scope edit qualifies,
not all.

**Wrong assumption 2 — the serious one: "the migration is the whole change."** It is
not. `subject_type` has a **second enforcement point in Go**: `validDocSubjectTypes`
(`platform/orchestration/actions/doc_subjects_common.go:37`), which gates
`write_doc_plan`, `append_doc_note`, `load_doc_context` and `persist_diagnosis_note`.
Shipping the DDL alone would have reproduced **`bugs_open/064` for the third time** —
migration 163 (+`experience`) missed the `persist_diagnosis_note` gate; migration 184
(+`action`) moved the DB CHECKs **only** and left its own seeded action docs unreachable
through every doc action. The file's own header says the rule outright: *a value the DB
accepted but a Go gate rejects, or vice versa, is a split contract; move both together.*

**My RUNBOOK, PLAN, PROPOSAL, NOTES and memory entry all described the incomplete fix**,
and every one of them is now corrected in place rather than quietly edited. This is
`WRONG_CALLS.md` material and is logged there.

**What actually caught it:** not care, and not a review — a **code comment**, which
pointed at `experience_register/design/subject_type_addition.md`, a checklist that
already enumerates all four enforcement points and was written precisely because *"every
addition so far has missed at least one"*. I found the checklist only because I opened
the file I was about to edit. **The cheap check that would have caught it: grep the
value you are adding, not the table you are changing** — `git grep -n
"experience-pattern"` returns the Go list, the migration and the checklist in one shot.

**D5 is REVERSED (→ D5′).** Beyond the Go half, the migration **must** be numbered in
`sql_for_agents/`, because `TestValidDocSubjectTypes_LockstepWithMigrationCheck` parses
the newest numbered `.sql` recreating `doc_plans_subject_type_check` and fails if its
array differs from the Go list. So withholding the number protects nothing and **reddens
HEAD** for every other session the moment the Go edit lands. D5 as written was
unbuildable. The residual risk D5 worried about (another session's `--apply` sweeping it
in early) is real but **inert here** — nothing writes component docs yet, so a widened
CHECK ahead of the image has no effect. That is stated in the migration header so nobody
over-reacts to an early apply.

**Applied this lane's own S2 rule to itself.** The lockstep test passing is not evidence;
it had to be seen red. Hid the migration → **FAIL**, naming 184's exact failure mode
(*"split contract: validDocSubjectTypes = [action component experience
experience-pattern pipeline tool] but 218_experience_register_substrate.sql sets the
CHECK to [action experience experience-pattern pipeline tool] — move both together"*);
restored → passes. `platform/...` and `internal/...` build clean; actions tests pass.

**Blast radius measured, not asserted** (07-29 ruling §3 — name the consumers and tell
them, do not merely measure). Consumers: `write_doc_plan`, `append_doc_note`,
`load_doc_context`, `persist_diagnosis_note`, `check_tool_acceptance`,
`check_tool_acceptance_due`, `rename_tool_identity`/`RekeyTravellingDocs`, and
`scripts/landmines-sync.py`. **Nothing changes for any of them.** The one consumer that
reads `doc_notes` without a `subject_type` filter is `digestGatherImmune`
(`fixloop_digest_action.go:295-299`) and it is narrowed by
`categories ? 'triage'`/`'silent-check'`, which component notes will not carry — they
mirror the tool convention (`acceptance-run`/`acceptance-fail`). So the count does not
move, and that is a query rather than an argument.

**Registered in the shipping commit** as **DOC-068** (CLAUDE.md platform-seams condition
2 — a shared-vocabulary seam must be registered in the same commit that ships it), with
its landmine and its open review question; dropped from `102_coverage_ratchet.txt`
accordingly. Index row added. Noted in passing, **not fixed**: DOC-067 is absent from
`000_concept_index.md` — another thread's omission, and tidying it silently is not mine
to do.

**Unrelated red HEAD found while building, flagged not fixed:** `cmd/reasoningset/main.go`
does not compile at HEAD (`declared and not used`, line 504) — confirmed present in
HEAD's own copy with no local modifications, so it is another lane's (reasoning_dataset).
`platform/...`, `internal/...` and every other `cmd/` build clean.

## 2026-07-30 (evening) — round 1 REVISE; every objection answered; round 2 submitted

Verdict landed **~12 minutes** after submission, not the ~30 the runbook budgets.
`decided_by`: *"gating objection from debug_historian"*. `unreadable: null` — so this was
a real verdict, not a truncated harness (that is the field to read, not `abstained`).
**7 approve / 5 object**, abstained 5.

Approve: `editquality`, `tooling_provenance`, `diagnosis_guardian`, `render_guardian`,
`constitution`, `mission`, **`architecture`** — the last with
`ARCHITECTURE_SIGNAL: point_fix` and an explicit finding that **the RFC carve-out was
invoked correctly**: *"this is a value addition to an existing enumerable vocabulary, not
a new namespace, wire shape, or state machine… no consumer's guarantee changes, all filter
explicitly."* That settles the scoping question the lane had been carrying.

Object: `debug_historian` (gating, HIGH), `bug_historian`, `reuse_agent`, `guardian`,
`prior_art_librarian`.

**THE GATING OBJECTION WAS RIGHT AND I CONCEDED IT WHOLE.** It said risk #1 named the
ordering hazard and then defused it with an **argument** — *"no producer exists yet, so an
early apply is inert"* — where the discipline is **verify against the pod**, never an
inertness argument and never image-tag trust. Its sharpest line: *"the fix should not
repeat the pattern of trusting stated ordering over verified deployment state"* — on the
second attempt at the split that produced `bugs_open/064`. A comment in a SQL header is
not a gate.

So there is now an executable one, `VERIFY_273_before_apply.sh`, and **it correctly
refuses right now**: *"binary built 2026-07-30 17:25:05 UTC, BEFORE the Go half was
committed (19:28:54 UTC) — this pod cannot carry 'component'"*, exit 1, both replicas.

### The gate reported a FALSE PASS on its first run, and that is the entry worth reading

First version printed **`Go half present (new=0 0, control=0 0)`** — a green verdict built
out of four zeroes, on pods that plainly lacked the marker. Cause: **`grep -c` exits 1
when the count is zero**, so my `|| echo 0` fallback printed a *second* zero, giving the
two-line string `"0\n0"`. That is not equal to `"0"`, so `[ "$CTL" = "0" ]` was false, the
refusing branch was skipped, and control fell through to the success branch.

**A gate whose only untested branch is the one that refuses is not a gate.** This is the
same shape as the hazard this lane filed hours earlier (an all-skipped check set reads as
PASS) and the same shape as the carousel bug that started the lane. Third instance of one
class in one day, in my own work each time. Rewritten to normalise to a single integer and
**fail closed**; then re-run to *watch* it refuse.

### All three of my markers were wrong, and measuring is what fixed it

| marker | count | why |
|---|---|---|
| `component carries a PLAN…` | **0** | a Go **comment** — comments never reach a binary |
| `is a split contract; move both together` | **0** | also a comment, so my **control** could never have passed either |
| bare `component` | **761** | matches `content_components`, `component_level`, … |
| `experience-pattern` | **1** | so slice literals **do** reach rodata |
| `actionexperience-pattern` / `experience-patterncomponent` | **0** | literals are not laid out contiguously — do not rely on adjacency |

**The conclusion is a fact about this change, not a preference: it adds no unique greppable
string.** So the sound check is to **date the build** (`stat -c %Y` on the binary vs the
commit's `%ct`), keeping the control grep only to prove the binary is readable at all. The
script states that build-date is **necessary, not sufficient** — a same-tag rebuild from a
stale checkout would also date late — and names the behavioural probe (write a component
`doc_plans` row, read it back *through* `load_doc_context`) as the only definitive check.
Note a control marker must be a **string literal**, never a comment; I got that wrong in
the very act of applying the rule I had written down that morning.

### The objections answered with a query rather than prose

- **`RekeyTravellingDocs` has exactly one caller** — `rename_tool_identity_action.go:115`,
  hardcoding `"tool"`. Absence claim confirmed systematically.
- **Five Go readers of `doc_notes`**, four filtering `subject_type` explicitly
  (`check_tool_acceptance.go:574`, `check_tool_acceptance_due.go:93`,
  `load_doc_context_action.go:79`, `tool_acceptance_actions.go:837`); the fifth,
  `digestGatherImmune`, narrowed by `categories`. Enumerated, not asserted.
- **The migration-number objection was the seat misreading two consistent facts**, and the
  query settles it: the lockstep regex matches `ADD CONSTRAINT
  doc_plans_subject_type_check`, and **270 alters `doc_notes` only** — so 218 genuinely is
  the newest migration recreating the `doc_plans` CHECK. 271 and 272 are claimed
  (`272_feature_designer_plan_repair_loop.sql`), so 273 is right.

### `reuse_agent` asked the best question of the round, and I had not done the check

It asked whether two existing mechanisms already model per-site-per-component verdicts.
Both exist; both ruled out **on evidence**:

- **`content_components.quality_score` / `quality_issues` / `avg_quality_score`** — real
  columns on the exact row being extended, **but the table has no site dimension.** One
  fleet-shared row serves 11 sites, so it structurally cannot express *works on site A,
  collapsed on site B*, which is the whole of what S4–S7 asks. It is a fleet quality
  *score* (smallint 0–100), not a per-site verdict with evidence.
- **`site_work_items`** (`site_id` + `component_id` + `result`) is a genuine candidate —
  but it is the **request** mechanism, not the **record**: `idx_swi_dedup` is UNIQUE
  `(site_id, item_key)` over the non-terminal statuses, so slots are reused and status is
  mutable, whereas `append_doc_note` is *"one INSERT into doc_notes (no
  read-modify-write)"* and `doc_notes` has no status column. **Decisive: the tool ladder
  already made this exact choice in code** — `tool_acceptance_due` files the work item and
  `check_tool_acceptance_due.go:90-97` reads the cooldown **from `doc_notes`**. Verdicts
  live in notes; requests live in work items. Using work items for component verdicts
  would *create* the two-ways drift the seat was warning about.

### Deferrals converted into tracked work rather than defended

`bug_historian` (medium) and `architecture` (low) both said the rename-orphan risk should
be a ticket, not a comment — *"the only defence is a comment nobody is forced to read"*.
Filed **`features_open/028`**, with the one-caller grep as its measurement, `bugs_open/136`
as the live precedent, and candidate 2 (a detector for travelling docs whose `subject_key`
resolves to nothing) flagged as worth building regardless, because **it is the only
candidate that finds the orphans already created by past renames — a count nobody has ever
measured, across all six subject types.**

`guardian` (medium) was right that naming a consumer is not telling it. Stated plainly,
for the record and in DOC-068: **`persist_diagnosis_note` will now accept
`subject_type='component'` diagnosis notes**, as a consequence of single-sourcing, judged
correct on the same argument `subject_type_addition.md` item 3 made for `action`, and
independently approved by `diagnosis_guardian` on the grounds that persistence-never-fails
and skip-never-guess are untouched.

**Round 2 submitted** — trail stays `e5673868-7c5b-489c-931a-7ba59b959b91` (so the
artefacts accumulate in one place), run correlation `82506f25`, orch `e2de1d7f`.
Round-2 artefacts committed `d16447161`.

### Also written this session

`CONSULT_2026-07-30_next_tool_build.md` — the brief for the lane the owner asked to
consult with us on its next tool build. It maps S0–S7 onto a tool, lists the five traps
with the check for each, and asks four genuinely open questions. **It explicitly tells
them to skip a stage that feels like ceremony and report that instead**, because the
ladder has never been run *forwards* by anyone, including me, and compliance would teach
us less than friction does.
