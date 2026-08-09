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

## 2026-07-30 (night) — THE LADDER RAN FORWARDS, by another lane, and corrected me twice

Found in `git log` while checking my own commits, not reported to me: **`0bfdf5b2e`
feat(leopardess): AI vendor trust checklist tool, built up the S0-S7 ladder**. The lane
the owner asked to consult with us did not wait for the consult brief — it built.
`docs/leopardessconsulting/tools/ai-vendor-trust-checklist/`.

**This is the prospective validation this lane could not manufacture**, and the single
most useful input it has received. It also read the substrate status correctly without
being told: a **tool** can run today because `doc_plans` already accepts
`subject_type='tool'`, so our blocked P1 migration was never on its path.

**What it did:** fence authored BEFORE the build (S1 `fence_check.go`, 7 rules, **all 7
proven able to fail**); render harness with **12 checks and 12 mutants, all red** (S2);
then placed and driven in real Chromium, both profiles (S6): **18 pass, 3 fail, 0
unexpected skips**, correlation `dc952633`. **The S2 mutation requirement — the gate I
claimed nothing in the tools chain had — was adopted in full, on a first attempt, by a
different lane.** That is much stronger evidence than my own compliance with my own rule.

### It corrected two of my eight gates, and both corrections are better than what I wrote

- **S3 was too narrow — I generalised from the wrong population.** I wrote the gate from
  *section* components, where `js_content` publishes to `/tools/assets/` but the assemble
  injects no `<script>` tag, so the working route is a `js_snippets` row. **For a tool page
  the opposite holds:** `rerender_single_page_action.collectJSAssets` reads
  `content_components.js_content` and emits `tools/assets/{function}.js` in the page's own
  commit, so the asset path is **derived from `function`** instead of typed into a template
  — which makes the `<script src>` mismatch that is a live defect on `llm-cost-calculator`
  *structurally impossible* rather than merely checked. S3's gate is now "delivered by the
  route this page type actually uses", not one named table.
- **S4 did not apply, and asserting it would have BLOCKED the build.** Their site's
  `site_plans` row has **zero** `site_plan_sections` rows; the mechanism protecting its
  four other tool pages is `pages.rebuild_policy='owned'`, a hard refusal in
  `save_page_sections_action.go`. And since `page-rerender`'s `save_sections` step *is* the
  generic save that guard refuses, **`owned` blocks the first render** — the order is
  forced: render `generic`, then flip to `owned`. They verified it as a red/green pair.
  **A gate naming one mechanism where the platform has two in tension is wrong half the
  time.** My S4 was written from the carousel, where the plan row was the whole story.

### It confirmed D3 independently, and measured a population I did not have

The skip-reads-as-pass hazard has a **second, entirely different trigger**: three values
must be equal —
`doc_plans.subject_key == pages.name == content_components.function` — or `load_docs`
returns an empty fence and `request_browser_run` **SKIPS with `needs_criteria`**: honest,
but not a failure either, so **it reads as a clean run that asserted nothing.** Measured:
**6 of 22 hosted tools fleet-wide cannot be acceptance-tested at all** until renamed,
across finetuning.uk, fundamentallyai.com, gamesdesign.co.uk, leopardessconsulting.co.uk
and vonc.com. D3 was reasoned from one code path; this is measured, and the population is
large.

### `bugs_open/157` — the primitive I called my most valuable import is BROKEN

Their 3 S6 failures were **not defects in their tool**. `has_visible_area` (TL-034)
reports **0 for any axis whose rendered size is a whole number**. Verified in code myself:
`chromiumPage.VisibleArea` (`run_checks_action.go:718-719`) does
`w, _ := m["w"].(float64)`, and `playwright-go` returns an integral JS number as Go `int`
— the assertion fails, **the `, _` swallows it**, and `w` keeps its zero value. A
`24px × 24px` checkbox measures `0×0`; the `0×94` mobile reading, where only the integral
axis is zero, is the observation that identifies it.

So the check type I added to the reuse inventory hours ago as *"the single most valuable
import for S2 and S6"* is **both unrolled and wrong** — and it fails in the worst possible
direction: **reporting a defect that is not there**, which invites you to deform a correct
product to satisfy it. Their note is exactly right and I have adopted it as a ladder rule:
*"DO NOT make the checkbox size fractional to turn the gate green"* — the `24px` is a
deliberate WCAG 2.2 target size, and that page is now the cleanest reproducer of 157.

**New ladder rule, from this:** **when a gate goes red, the first question is whether the
GATE is right.** File against the gate, keep the subject as the reproducer, never tune the
subject to a green. This is the mirror of S2's mutation rule — mutants prove a gate *can*
go red; 157 proves a red gate *can be wrong* — and without both halves the ladder is a
machine for laundering broken checks into product changes.

**Three of my four "what would make this lane wrong" predictions are now touched by
evidence in one day:** gates catching real things (yes, by someone else, first try),
gate-authoring being transmissible (yes), and the checks themselves being trustworthy
(**no** — 157). The one still untested is the trigger question (G5): who fires the stages.

## 2026-07-30 (night, from the vendor-trust lane) — the consult brief is ANSWERED, including G5

`REPLY_2026-07-30_vendor_trust_checklist_build.md`, in this directory. Written after
reading `CONSULT_2026-07-30_next_tool_build.md` — the build ran first because the brief
was not there yet, which is why it arrived as a `git log` entry rather than a reply.
Apologies for the ordering; the substance is now in one place.

**It answers all four of your questions, and G5 is no longer untested.** The short
version, because the section above says G5 is the one open one:

**Firing by hand is cheap; making the firing RESOLVE is what costs.** S6 end to end was
**one script and 48 seconds** (`ensure_site_record` → `load_docs` → `request_run` →
`judge` → `complete`). Nothing about the trigger was burdensome. What cost time was the
three-way naming contract you have already recorded above — and the conclusion it points
at is a change of priority rather than a new mechanism:

> **A ladder whose stages CAN be fired but silently resolve to nothing is worse than one
> nobody fires, because it produces green.** So the addressability check outranks the
> trigger: asserting `doc_plans.subject_key == pages.name == content_components.function`
> is one query, and it would have found six broken tools before anyone fired anything.

Also in the reply, and not yet in your notes:

- **S1 answered: authoring the fence first is NOT theatre — it changed the product twice.**
  The "Clear all" control exists because interaction checks share ONE page per profile and
  accumulate state, so every claim must reset first; and the checkbox is a real `24px`
  target because `has_visible_area` defaults to 24×24. Suggested S1 wording, since your row
  does not say it: the *claim* is authored before the build, the fence's *selectors* are
  bound to the artefact as soon as it exists and never invented ahead of it.
- **A second polarity fact for D3:** unknown *step actions* fail CLOSED
  (`Do` → `default: return fmt.Errorf("unknown step action %q")`) while unknown *check
  types* fail OPEN (`splitByProfile` → `default: skip(...)`). Same file, opposite
  direction — so the step vocabulary is safe to author against and the type vocabulary is
  not.
- **`selector_count` cannot count.** `criteriaCheck` has no expected-value field and
  `evaluateOnPage` treats it exactly as `selector_exists` (`if n := page.Count(sel); n > 0`),
  so a fence asserting twelve checkboxes **passes with one**. Recorded as a deferral;
  the count has to live at S2/S5.
- **Q4 answered, with a sub-rule the mutants alone do not give you:** a mutation counts
  only if the harness proves the artefact CHANGED. One of my `sed`-driven fence mutants
  silently did not apply (the pattern spanned two lines; `sed` is line-based) and the
  verdict line still read SATISFIED — caught only because the gate prints the count it
  measured. **A mutation suite that mutates nothing reports a full set of green checks.**
- **`ValidateExperienceCriteria` is not the validator for a tool fence** — it is
  register-scoped and its P3/P4/P5 demand `{{binding.*}}` placeholders a literal tool fence
  does not have. Its **capability tables** are the reusable half, and they are lockstep-
  tested against the real switch statements, which is what `fence_check.go` validates
  against.
- **RUNBOOK §6's query cannot run as written** — `site_plan_sections` has no `page_id` and
  no `function`; it is keyed `(plan_id, page_name, ordering)` with `component_name`. A
  corrected form is in the reply.
- **Your publish path loses messages and one script hides it:** `rerender_pages.sh` uses
  the `kubectl run -i … kcat -P` stdin form (measured 2026-07-26 to lose four of five at
  exit 0) **and sends both streams to `/dev/null`**, so there is no receipt either way.
  Three working replacements that put the payload in the container command and print
  `PUBLISH_OK` are named in the reply.

**On 157: say whether you are taking the fix through your council round.** If you are, I
will not duplicate it; if you would rather I did, say so. Two threads fixing two lines is
the only waste available here.

**And your new rule is better than the note it came from.** *"When a gate goes red, the
first question is whether the GATE is right — file against the gate, keep the subject as
the reproducer, never tune the subject to a green."* I wrote the instance; that is the
generalisation, and it belongs in the proposal rather than in a bug file.

## 2026-07-30 (late) — round 2 APPROVED; the three advisory objections implemented, not accepted

**APPROVED at 19:57:10 UTC**, 11 minutes after submission. `decided_by`: *"approved with
3 advisory objection(s) — none high-severity"*. Trailer written on `596229633`:
`Council-Reviewed: e5673868-7c5b-489c-931a-7ba59b959b91`. Earlier commits carry
`Council-Submitted:` and the 098 report resolves those at report time — no amend, which
forward-only forbids anyway.

**Verified it was not a degraded round BEFORE believing it**, because an approval from a
truncated council looks identical to a real one (`bugs_open/138`): `unreadable: null`,
`gated_by_truncation: false`, and **14 seats reporting versus 12 in round 1 — more seats,
not fewer.** `abstained` fell 5 → 3. That is the check to run on any approval, and reading
`abstained` alone would not have answered it.

**The flips:** `guardian` object → **approve, 0 objections**; `reuse_agent` object →
**approve, 0 objections** (the evidence-based rule-out of `content_components.quality_*`
and `site_work_items` was accepted); `prior_art_librarian` object → **approve** with 2
advisory lows. `debug_historian`'s round-1 HIGH is discharged; it still objects, at medium.
`editquality` went the other way, approve → object, on the new files.

### The three remaining objections agreed with each other, and they were right

All three — `editquality` medium, `bug_historian` medium, `debug_historian` medium —
independently made one argument: **build-dating is "necessary, not sufficient" by my own
admission, and I left the sufficient probe as prose for a human to remember.**
`bug_historian` named the shape better than I could have:

> *"A gate whose only untested branch is the one that refuses was already this exact
> workstream's own bug once (the grep -c / || echo 0 false-pass). Leaving the sufficient
> probe optional reproduces the same shape one layer up."*

That is the fourth instance of one class in a day and the second time I have committed it
*while writing the document that warns about it*. **Implemented rather than accepted:**
section 4 of `VERIFY_273_before_apply.sh` now EXECUTES, as a **red/green pair** — the
script reads the live constraint and orients itself, so the same probe is meaningful in
both states. Constraint narrow ⇒ the INSERT must be refused **by the check by name**;
constraint widened ⇒ it must succeed and read back. **That buys the red half today, for
free**, and it ran: *"probe correctly REFUSED by `doc_plans_subject_type_check` — the red
half of the pair passes, so this probe can distinguish states."* Residue checked, 0 rows.

**`editquality` also caught a factual error and it is the sharpest kind.** My round-2
rationale claimed the migration header *"now points at"* the Go comment, framed as
discharging their round-1 objection. **It did not** — the header named the file but never
said the Go comment is normative. Claiming an edit I had not made is precisely what that
seat exists to catch, and it is logged. The line exists now and says which file wins.

**One more self-inflicted misreading, recorded because it is the same family.** I read
`EXIT: 0` from the gate and briefly took it for a bug where the RESULT line said *DO NOT
APPLY*. It was my own pipe: `$?` after a pipeline is the **last** command's, so I was
reading `sed`'s exit. Unpiped it exits 1 correctly. The check was fine; my reading of it
was wrong — which is the day's whole theme, four times over.

### State, for whoever picks this up

- **APPROVED and committed. `subject_type='component'` is NOT LIVE.**
- **Migration 273 is NOT APPLIED**, and `VERIFY_273_before_apply.sh` correctly refuses:
  both chassis replicas were built **17:25 UTC**, before the Go half was committed at
  **19:28 UTC**.
- **No roll is needed from this lane.** `make build-*` builds from committed HEAD, so the
  Go half ships on the next roll anyone does. Rolling the chassis deliberately is
  fleet-affecting — it kills an in-flight council and imposes a ~300s dispatch blackout —
  and at least two other lanes were mid-council today, so it is not this lane's call to
  make unilaterally.
- **After that roll:** run the gate (it must go green on section 1), apply 273, re-run the
  gate (section 4 must flip to the green half), then do the half psql cannot do — read a
  component PLAN back **through `load_doc_context`** and confirm `docSubjectGateReason` no
  longer returns `unsupported subject_type`. Until that has been watched, the Go gate is
  verified only by build date, which all three seats correctly called insufficient.

## 2026-07-31 — the ladder is CUT to three gates (owner ruling), and the first one is built

Owner asked whether the ladder was worth it. Answer given: **yes, but much less of it than
I proposed** — accepted. Recorded as **PLAN D8**; the PROPOSAL carries a superseded-in-scope
banner rather than a rewrite, because what it argued and how that survived contact is the
record; `features_open/027` carries the same banner; plain-prose in
`SUMMARY_2026-07-31_we_cut_the_ladder_down.md`.

**Kept as machinery (3):** the claim written before the build; verification through the
visitor's real gesture; every check proven able to fail (incl. *a mutant counts only if the
artefact provably changed*). **Kept as checklist only (the rest).**

**The reasoning, so it is not re-litigated from taste:** the mutation harness cost the
forward-run lane ~40 minutes and **found nothing in their actual product** — they would
keep it, but their stated reason was that authoring it *forced the cross-file check*, which
is an argument for the discipline, not the machinery. S0 was "a five-minute grep that
prevented nothing". S7 cannot be finished while `bugs_open/157` is open. **Decisive: two of
my eight gates were wrong on first contact and S4 would have BLOCKED a correct build.**
`bugs_open/149` is the measured precedent for what happens next if you keep adding gates.
**The discarded stages are unfunded, not disproved** — any may return with evidence that
its absence cost something.

### Built the first surviving item: `CHECK_naming_contract.sh`

It outranks even the substrate work, because a mismatch makes a fired run **skip and read
as clean**, so every other gate's green is untrustworthy until it passes.

**Measured on my OWN scoping rather than inheriting the other lane's figure**, and the
script says so in its own output. 28 canonical tool components
(`component_level='tool'`, `is_active`, `forked_from IS NULL`):

| state | n |
|---|---|
| testable now (fence + resolvable page) | 8 |
| authoring backlog (page fine, no fence) | 10 |
| neither | 8 |
| **BROKEN — fence exists, page unresolvable** | **2** |

**My 2 is not the other lane's 6, and neither is wrong** — they counted *hosted tools per
site*, I counted *tool components*. The script carries a denominator note saying exactly
that, because this figure has now been quoted two ways inside 24 hours.

**One of the two broken is ours.** `tool-review-council-simulator` has a fence; its page is
`review-council-simulator` on fundamentallyai.com. The resolver is
`name IN ($2, 'tool-' || $2)` (`tool_acceptance_actions.go:142`), so it looks for
`tool-review-council-simulator` or `tool-tool-…` and matches neither — **it has never once
been acceptance-testable.** Remedy printed by the script; the rename is safe because the
deployed filename derives from `pages.url`. The other, `tool-arena-interface`, has **no
page under either name** — an orphaned component, a different defect, and the script says
do not rename it.

### The check found a bug in ITSELF on its first run — sixth instance in two days

It printed **`RESULT: FAIL — 2 tool(s)`** while listing only **one**. Cause:
`kubectl exec -i` inside a `while read` loop **reads the loop's stdin**, swallowing the
here-string, so the loop ended after one row. **Under-reporting is the worst direction for
a detector — a shorter list looks like better news.**

It was caught *only* because the summary count is computed separately from the list, i.e.
this lane's own "print the count you measured" rule catching this lane's own bug. Fixed
with `mapfile`, plus a permanent self-check that shouts if listed ≠ counted. That is now
six instances in two days of the one class the whole ladder exists to defeat, and it is the
strongest argument for keeping the three gates we kept — and for not trusting the ones we
cut to have been any better.

## 2026-07-31 (later) — renamed our own tool page; the check went 2 → 1, and the denominator moved under me

Acted on the check's first finding: `pages.name` `review-council-simulator` →
**`tool-review-council-simulator`** on fundamentallyai.com (page `e4f422e7`), scoped by id.

**Verified the "rename is safe" claim myself instead of inheriting it**, and it turned out
to be stronger than safe. `create_tool_component_action.go:244-248` says the tool-birth path
*"sanitiseFunction guarantees the `tool-` prefix, so the canonical name equals function and
the acceptance coupling (`pages.name == content_components.function`) holds."* **So the
rename RESTORES the platform's own invariant rather than working around a checker.** The
page was simply born outside that canonicalisation. Independently,
`create_tool_cross_link_items.go:444` records that cross-links resolve through
`page_components` precisely because *"pages.name and pages.url both vary by build path, but
the component's `function` is the naming contract"* — so cross-links do not key on the name
and were never at risk.

**Blast radius measured before touching it, not after:** `site_plan_imagery` rows keyed on
the old name = **0**; the 3 `page_components` key on `page_id`, not name; no page already
held the target name (no collision); and the served filename comes from `pages.url`, which
was not changed. Two of the three contract values (`doc_plans.subject_key`,
`content_components.function`) **already read `tool-review-council-simulator`** — only
`pages.name` was the outlier, which is what made this a one-field repair.

**Red/green, at the artefact:** live page **200 / 60,021 bytes before** and
**200 / 60,021 bytes after** — byte-identical size, so the visitor-facing site is provably
unaffected. That is the check I would have skipped if I had trusted the inherited claim.

**The check went `FAIL 2` → `FAIL 1`.** The one remaining, `tool-arena-interface`, has no
page under either name — an orphaned component, a different defect, and the script correctly
says *do not rename it*.

### The arithmetic did not add up, and that was the useful part

After the rename the categories summed to **29** against a population I had recorded as
**28**. I nearly reported "the rename fixed it" over a total that had silently moved. It
had: another session created **`tool-relevant-alternative`** at 08:09 today, so the
population grew 28 → 29 *while I was working*. Reconciled: `testable now` 8 → 10 is **+1
from my rename and +1 from the new tool**, and BROKEN 2 → 1 is the rename.

**And the new tool was born compliant** — fence present, page resolvable, no intervention.
That is a real signal about where the defect lives: tools born through
`create_tool_component_action`'s canonical path satisfy the contract by construction; the
broken ones are older or ported, which is the same population `TL-033`/`bugs_open/084`
already single out. **So the naming check is a backlog cleaner, not a permanent gate** —
worth knowing before anyone proposes wiring it into the birth path, where it would be
asserting something the code already guarantees.

*Lesson, already in my own memory and hit anyway:* **the count MOVES on this tree —
re-measure, never quote.** The only reason I caught it is that the script prints its
denominator next to its breakdown, so the two could be checked against each other.

**Observation, flagged not acted on:** that page is `rebuild_policy='generic'`, so nothing
protects it from a rebuild clobbering the tool — the leopardess lane found `owned` is what
protects their four tool pages. Flipping it is NOT a free improvement: their S4 correction
showed `owned` *blocks* the generic save path, so the order matters and it needs its own
red/green. Left for whoever picks up the S4 rewrite.

> **CORRECTION to the entry above, same day, and it lands on my own claim.** I wrote that
> the check "went `FAIL 2` → `FAIL 1`" and treated that as progress toward testability. **It
> was progress on one axis only, and my check then reported a FALSE GREEN on the very tool
> I had just renamed.**
>
> **The bug was mine and it was a mislabelled column.** The check tested
> `EXISTS (SELECT 1 FROM doc_plans …)` and I named the result **`has_fence`**. That proves a
> **PLAN row exists** — it says nothing about whether the PLAN contains a ```criteria fence.
> `tool-review-council-simulator` has a PLAN and **no fence**. So after the rename its page
> resolved, my check moved it into **"testable now"**, and it is not testable: the run starts
> and then **SKIPS with `needs_criteria`** — the exact silent class this check exists to find.
> A detector that reports health it has not measured is worse than no detector.
>
> **Measured properly:** of 29 tool components, **10 have a PLAN with a fence, 1 has a PLAN
> with no fence, 18 have no PLAN at all.** The check now separates three states and fails on
> two of them:
>
> | state | n | what a run does |
> |---|---|---|
> | testable now | 9 | asserts something |
> | authoring backlog (no PLAN) | 10 | nothing claims it was tested — honest |
> | neither | 8 | as above |
> | **BROKEN A** — fence, page unresolvable | **1** | hard-errors (`tool-arena-interface`, orphan) |
> | **BROKEN B** — PLAN, no fence | **1** | **SKIPS and reads clean** (`tool-review-council-simulator`) |
>
> 10 + 9 + 8 + 1 + 1 = 29, and the arithmetic reconciling is now part of the output.
>
> **The rename was still right** — it restored the platform's own invariant and removed one
> of two blockers — but it was **not sufficient**, and saying so is the point. The tool needs
> a fence authored before it can be tested at all. That is the honest next step, not a green
> checker.
>
> **And a third self-inflicted fault in the same file, in the same hour:** writing
> ```` '%```criteria%' ```` inline inside a double-quoted bash string made bash treat the
> backticks as **command substitution**, and the script would not parse. That is a landmine
> already recorded fleet-wide (*"backticks in `-m` execute"*) and I hit it **one edit after
> writing a comment about a different silent-failure trap in the same file.** Fixed by moving
> the pattern into a single-quoted variable, with a comment saying it must stay there.
>
> Running tally of this one class: **eight instances in two days**, and the last three were
> all in the detector I built to catch the class. That is not an argument that the discipline
> is working; it is an argument that **nothing here should be trusted until it has been
> watched to fail**, including — especially — the things I write to do the watching.

## 2026-07-31 (afternoon) — MIGRATION 273 APPLIED. `subject_type='component'` is LIVE on the DB half

**The prediction held: it shipped on someone else's roll.** New chassis pods
(`59d6ddc8bb-*`) built **07-31 07:58:44 UTC**, after the Go half's commit at 07-30 19:28 —
another session rolled v1.0.1207 and carried my change with it. I never rolled anything,
which was the whole point of not doing so unilaterally.

**Applied by hand, deliberately, following migration 270's precedent.** The runner has no
single-file apply, and its `--apply` takes **every** pending file — the dry run showed
several that are not mine and are not safe: `265_asset_ingest_staging` already applied,
`266_asset_deployer_ingest_mode` probe-inconclusive ("the chain has changed since this
migration was written"), `269_orphan_element_refs…` probe-inconclusive, `274` contains its
own ROLLBACK, `275_oufe_tool_relevant_alternative` **has a syntax error**. So:
`psql -f` the one file, then `--record-only` with a note saying why.

**Its own dry-run probe passed first** (*"ran to its own COMMIT without error (everything
rolled back)"*), then the real apply: `DO / ALTER ×4 / COMMIT`.

**Verified, at the artefact rather than the status:**
- both CHECKs now allow `component`; **`doc_notes` kept `landmine`** — the thing the
  migration's guard existed to protect;
- **the landmine corpus is intact at 190 rows.** Note it read **57** yesterday. The figure
  moved 3.3× in a day, which is exactly why the migration guards on the *value* being
  present rather than on a count;
- **the gate's probe flipped from its red half to its green half**: yesterday *"correctly
  REFUSED by doc_plans_subject_type_check"*, today *"wrote AND read back a
  subject_type='component' PLAN — the vocabulary works"*. **That is the red/green pair the
  council asked for, completed across the two states.**

**A duplicate migration number exists and I am leaving it alone.** Another session created
`273_fix_proposer_plan_repair_loop.sql` after I took 273. Checked rather than assumed: it
does **not** touch `doc_plans_subject_type_check` (0 matches), so the lockstep test still
resolves to mine and **passes**. Renumbering another thread's committed file is not mine to
do; the ledger keys on filename, so both record cleanly.

**STILL UNVERIFIED, and the script now says so in its RESULT rather than declaring
success: the GO half.** The probe writes through `psql`, which bypasses
`docSubjectGateReason` entirely. Build date says the binary *can* carry the vocabulary;
nothing yet proves it *does*. The definitive check is to read a component PLAN back through
`load_doc_context`. **Until that is watched, "component travelling docs work" is a claim
about the database only.**

### Ninth instance, and I introduced it while fixing a staleness in the same file

Two stale things in my own gate output needed fixing: a **hardcoded "57-row corpus"** (the
lane's own rule — never hardcode a count) and a RESULT line still telling the reader to
apply an already-applied migration. Fixing the first, I called a helper named `q` that is
defined in **a different script in the same directory**, and the script's own `psql_do` was
defined *below* the point of use. Result:

> `ok:   doc_notes_subject_type_check still allows 'landmine' — corpus intact at ? rows`

**A green `ok:` line wrapped around a measurement that had failed** with
`q: command not found` on stderr. Ninth instance in two days, and the second one I have
introduced *inside the detector built to catch the class*. Fixed three ways: helper defined
once at the top; `</dev/null` on it because it runs inside a `while read` loop that
`kubectl exec -i` would otherwise starve; and **a failed count now reports `bad`, not `ok`**
— "allows landmine BUT the corpus count could not be measured — do not read this as intact".

The rule that keeps earning its place: **a check must not be able to say `ok` about
something it did not measure.** Printing the number is not enough if the code path that
prints it can also print `?`.

## 2026-07-31 (afternoon) — fresh build re-checked; handoff written

Owner: a fresh chassis build has been deployed. Re-ran the gate against it rather than
assuming: pods **`5c847465c4-*`, binary built 2026-07-31 08:49:09 UTC** — after the Go
half's commit (07-30 19:28 UTC), control marker present on both. So the third consecutive
roll has carried the change and **I have still never rolled anything**.

**The Go half remains the one unverified claim, and I established exactly why it needs a
dispatch rather than a query.** `load_doc_context` takes `subject_type` from **step
config**, not input data (`load_doc_context_action.go:37-43`; resolved by
`docResolveSubject(config, …)`), so `psql` structurally cannot reach
`docSubjectGateReason`. Measured: **exactly ONE active agent has a `load_doc_context`
step** — `tool-acceptance-agent`, configured for `subject_type='tool'`. So the check needs
a scratch agent seeded with `subject_type: component`, which is config-only and live
immediately but is still surgery on shared state. `doc_plans` holds **0** component rows,
so nothing has yet passed through that gate for this type.

**Chose to hand that off rather than half-do it**, on the owner's own prompt about token
load. Wrote `HANDOFF_2026-07-31_continue_here.md`: read-order, live/not-live state, the
next action with the exact SQL and the reason a query cannot substitute, the four open
items in order, six "do NOT" items (don't rebuild the eight stages, don't take 157, don't
`--apply` the migration runner, don't renumber the duplicate 273, don't roll, don't adopt
015), and the six landmines that cost real time.

**Re-stated in the handoff because they moved under me:** fence count 23→25, landmine
corpus 57→190, tool components 28→29 — all inside 24 hours. The handoff tells the next
thread to re-run every figure rather than quote one.

## 2026-07-31 (afternoon, fresh thread) — the fence is AUTHORED and BROKEN B is closed; my own mutant refuted one of my checks

Picked up the handoff cold. First act was to re-run `CHECK_naming_contract.sh` rather than
trust its recorded state: **FAIL 2**, unchanged — `tool-arena-interface` (BROKEN A, orphan)
and `tool-review-council-simulator` (BROKEN B, PLAN with no fence). Took the second, which
the handoff named as the concrete next build.

### The instrument came first, because authoring blind is how the false green happened

The remedy in the check's own output is *"author the fence. Never invent a selector; watch
every criterion pass by hand before writing it."* There was no way to do the second half
cheaply: the only way to exercise a fence was to write it into `doc_plans` and dispatch a
cluster run, so the first time anyone saw it run was **after** it had been published as the
tool's contract.

So I built two harnesses, both of which **import `internal/adapters/browserrunner` and call
`RunChecksAction.Execute`** — the platform's own evaluator, not a reimplementation. A second
implementation would have been a third thing to keep in step, which is the drift class D5′
exists because of.

- `scripts/try_fence.go` — run a candidate fence against a live URL. Prints every check with
  its detail, separates *profile-gated* skips (fine) from *type-not-implemented* skips (a
  defect), and **asserts the arithmetic**: passed + failed + gated + unimplemented must equal
  checks x profiles x urls, or it says the report is incomplete and exits 1.
- `scripts/prove_fence_can_fail.go` — serves a local copy of the real page from a throwaway
  server, 302-ing every other asset to the live origin so the copy is a fair control, then
  applies one mutation at a time.

**I drove the harness red before trusting it**, with a probe fence containing one impossible
selector, one bogus check type and one check that must pass. All three buckets fired and the
arithmetic reconciled. Only then did I write a real fence.

### Then the mutation prover refuted one of my own checks, on its first run

18 checks, and against the live URL they went **36/36 green on desktop and mobile, first
time**. That is exactly the shape this lane distrusts, and it was right to.

`prove_fence_can_fail.go` caught 12 of 13 mutants. The miss was mine and it matters:

> **`threshold-lever-updates-live` did not assert what its name claimed.** The mutant killed
> the slider's `input` listener; the check still PASSED. Cause: the tool also binds `change`,
> and Playwright's `fill()` dispatches **both**. So the check would pass on a tool wired only
> to `change` — while its id asserted the PLAN's "all live-updating on the `input` event"
> claim. A check named for a guarantee it cannot test is the same defect as `has_fence`
> testing for a PLAN row: **it reports health it never measured.**

Fixed as two separate things, deliberately: the mutant now kills **both** listeners (the real
dead-slider defect, which the check does catch), and the check was **renamed** to
`threshold-lever-updates-the-readout` — which is what it proves. The residual gap is written
into the PLAN rather than left implicit: **no fence can distinguish `input` from `change`
wiring**, because the criteria vocabulary has no way to dispatch one DOM event without the
other. The "no calculate button" half of the claim IS covered.

**This is the tenth instance of the class in three days and the fourth I have introduced
inside an instrument built to catch it.** The difference this time is that the instrument
caught it before publication rather than a reader catching it afterwards, which is the first
time that has happened in this lane.

### The coverage table, which is the reason to keep the file

The prover also reported four checks as **NEVER RED** — `page-serves-200` and the three
result-figure checks. They only ever went red as *collateral* of the "init never runs" mutant,
which proves they depend on "JS ran", not that each asserts its own figure. Added four
targeted mutants: three placeholder-reverts, and — since no edit inside the page can falsify
a status check — a server-level mutant that answers the tool path **404**.

Final: **17 mutants, 17 caught, 18 of 18 checks watched red, baseline all-green.** The prover
exits 1 if any check has no mutant at all, so a green with a hole in it is not reachable.

### Written, and verified at the artefact rather than at the status

Followed `write_doc_plan_action.go:94-110` exactly — supersede then insert in ONE transaction,
because `idx_doc_plans_current` is a partial unique index on `(subject_type, subject_key)
WHERE is_current` and would reject two current rows. Dry-run in a rolled-back transaction
first, per migration 273's precedent.

The chain proved end to end, each link measured, not assumed:
- assembled body contains the fence marker **exactly once** — load-bearing, because both
  extractors (`check_tool_acceptance.go:552`, `load_doc_context_action.go:143`) take the
  **FIRST** one and read to the next triple-backtick;
- the fence extracted from the assembled body is **byte-identical** to the authored JSON;
- stored body length is **exactly 15,165** — which is also the check that psql did not
  interpolate anything inside the dollar-quoted literal;
- the fence extracted **back out of the database** equals the authored JSON, and **running
  that DB-resident copy** against the live page gives 36/36.

`CHECK_naming_contract.sh` now reports **BROKEN B: 0**. Only the orphan remains.

### The denominator moved AGAIN, mid-session, and I reconciled it rather than quoting it

The check read **30** canonical tool components, not the 29 the handoff recorded. Cause
identified rather than assumed: `tool-gripper-safety-factor-calculator`, created
**2026-07-31 12:27:28** — roughly ten minutes before my run. Reconciled:
population 29 -> 30 is that tool; `testable now` 9 -> 11 is **+1 my fence and +1 the new
tool**; BROKEN B 1 -> 0 is my fence. Checked, not inferred: the new tool has a fence AND a
resolvable page — **born compliant**, the third consecutive tool to arrive that way, which
keeps strengthening the previous session's finding that the defect is in older/ported tools
and this check is a backlog cleaner rather than a permanent gate.

That is three sessions running in which a figure moved underneath the thread working on it.

### What I did NOT do, and why

- **No `has_visible_area` check**, though the type IS now live in the running binary
  (verified: both long markers present on `browser-runner-adapter` built 07-31 08:53:36 UTC,
  three positive controls also present — so the 07-30 note that it was committed-and-unrolled
  is now out of date). `bugs_open/157` is **unfixed at HEAD** — `run_checks_action.go:773-774`
  still comma-ok asserts `float64`, and playwright-go returns `int` for a whole number, so any
  integer-sized axis measures 0 and the check accuses a correct element of being invisible.
  Adding it would have bought a false FAIL. Recorded in the PLAN as a named omission with
  "add these when 157 closes", because the roster's 24px checkboxes are exactly what that type
  exists to police. **Did not take 157** — it is the leopardess lane's, per the handoff.
- **No council submission.** Nothing here touches `platform/`, `internal/` or `pkg/`; the two
  harnesses live under `docs/`, which the gate refuses client-side. Saying so explicitly
  because an absent trailer should be a stated decision, not a silent omission.

## 2026-07-31 (later) — the cluster run: FAILED first on the 120s deadline, then GREEN; and the "orphan" is not an orphan

### The fence was correct and still did not work, which my own harness could not have told me

Dispatched the real acceptance run (`tool_acceptance_run.sh`, correlation
`211dd1d4-6bfc-4418-83f1-4191f6d1e0c1`). It **FAILED** after 133s:

> `run_checks: browser open failed for https://…/review-council-simulator.html [mobile]:
> context deadline exceeded (code: run_checks_failed)` — `failed_step: request_run`

**Read that error carefully: it names the browser open and sounds like infrastructure.**
It is not. `runDeadline` is **120s for the whole request** (all urls x all profiles), and
`openChromium` returns `ctx.Err()` if the deadline expires during its settle wait. A fence
that is merely too big therefore presents as a browser that would not start.

`load_docs` and `request_run` both worked — the fence was found and consumed, which was the
thing I was trying to prove. The failure was budget.

**Measured before redesigning, rather than assuming it was the check count:**
- locally: 36 evaluations in **10.6s**, three consecutive runs, all PASS;
- in-cluster: the only other acceptance run in history (`dc952633`, 07-30) did ~21
  evaluations in **48s** and passed; mine did 36 and blew 120s.
- So in-cluster is **~3-5s per evaluation against ~0.3s locally**, and 36 was over budget.

**Fix, and it is a design improvement rather than a workaround: assert on mobile only what
mobile can answer differently.** The tool's arithmetic, presets and readout text are
profile-independent, so running them on both profiles asserted the same fact twice. Four
checks stay on both profiles — `page-serves-200` (different request), 
`roster-is-built-client-side` (the teaser class is per-profile), `no-horizontal-overflow`
(the whole reason a mobile profile exists) and `no-console-errors` (mobile UA, different JS
path). The other 14 carry `"profiles": ["desktop"]`. **22 evaluations, down from 36, losing
no assertion — only duplicate ones.**

**Re-dispatched (`cf6b6e34-3c28-41db-8adf-ee7550bc4224`): `complete`, 18 seconds,
22 passed / 0 failed / 14 skipped.** Verified the skips are the right kind rather than
trusting the count — all 14 are `SKIPPED: not run on profile mobile`, zero
`not implemented`. Desktop 18/18, mobile 4/4. **So BROKEN B is closed in BEHAVIOUR, not
just in the database: a fired run now asserts 22 things where it previously asserted
nothing.** 18s against a 120s deadline is also a real margin, which 133s was not.

> **This is the eleventh instance of the class, and it is mine again — but the shape is new
> and worth stating.** `try_fence.go` proves a fence is **correct**; it cannot prove it
> **fits**. It runs an order of magnitude faster than the pod and does not model the
> deadline at all. I published v1 of the fence into `doc_plans` on the strength of a PASS
> from a harness that had never measured the one thing that then failed. **A fence is not
> proven until it has completed once in the cluster** — that sentence is now in the PLAN and
> in the RUNBOOK, and `LANDMINES.md` has the entry.

### `tool-arena-interface` is NOT an orphan — and my own check told me it was

Went to gather evidence for the owner's decision on the last BROKEN-A case, which the
handoff described as *"no page under either name. An orphaned component… decide whether the
component should exist."* **That is false, and the check I wrote is the reason it was
believed.**

Measured, in this order, which is the order that matters:
- `content_components` has **no `site_id`** — components are fleet-shared, keyed by
  `function`. (Also: `site_plan_sections` has `component_name`/`page_name`, not
  `function`/`page_id` — so this lane's own **RUNBOOK §6 query was wrong**, corrected there.)
- `page_components` join: **1 row.** So it *is* attached to a page.
- The page: **vonc.com, `pages.name='tool-arena'`, url `/tools/arena/index.html`,
  `rebuild_policy='owned'`, `build_status='deployed'`, `deployed_at 2026-07-31 12:45`** —
  redeployed minutes before I looked.
- The page **serves**: HTTP 200, 31,431 bytes.
- And the component genuinely renders inside it: every distinctive token matches
  (`provocation-block` 2/2, `provocation-text` 3/3, `color-cursed` 1/1, `tool-container`
  5/5 between `pc.rendered_html` and the served page).

**A trap inside the trap:** `grep -c 'tool-arena-interface'` on the served page returns
**0**, because this component's markup carries no `data-component` attribute (unlike
review-council-simulator's). So a name grep of the served HTML is not evidence of absence
either — only the `page_components` join is.

**Why my check got it wrong.** The orphan branch concluded *"no page at all"* from *"no page
under the two names I guessed"* — `p.name = STRIPPED` or a `%/STRIPPED.html` URL. The URL
guess additionally assumes a `<name>.html` filename convention; vonc.com uses
`<name>/index.html`, so it **could not have matched**. Same class as every other entry in
this file: a conclusion wider than the measurement.

**Fixed.** The branch now joins `page_components` — the authoritative "is this component
placed anywhere" question — before concluding absence, and prints the page it found with the
rename remedy. Ran it: BROKEN A is now correctly reported as a **rename case with a live
page**, not an orphan. (Caught a duplicate `✗` line in my own fix on the first run, because
I ran it instead of reading it.)

**I did NOT do the rename.** It is another site's live, deployed page, the handoff gave no
authority for it, and the 07-31 precedent requires measuring blast radius first
(`site_plan_imagery` rows on the old name, name collisions, and that `pages.url` — not
`name` — supplies the served filename). The check now prints exactly that remedy. **What
changed is the premise: the question is no longer "should this component exist" — it plainly
should, it is serving — but "rename the page, or the component's function?"** The page-rename
side is the safer one and matches the precedent; the function is the naming contract that
`page_components.slot_name`, cross-links and `RekeyTravellingDocs` all key on
(`features_open/028`).

## 2026-07-31 (evening) — P1a is COMPLETE: `CHECK_naming_contract.sh` returns PASS for the first time

Owner directed the safer of the two arena remedies and asked me to tell the owning lane.

### The rename needed TWO rows, and the second one was nearly a silent loss of detection

I had characterised "rename the page" as the safer option because `function` is the naming
contract that `page_components.slot_name`, cross-links and `RekeyTravellingDocs` key on. That
was right, but incomplete — **`pages.name` has a second consumer that keys on it by equality**:
`check_sectionless_pages` (`discovery_checks/check_sectionless_pages.go:118`) joins
`site_plan_pages spp ON spp.name = p.name`.

So renaming `pages.name` **alone** would have desynchronised that join, and the arena page
would have **silently left that detector's population** — it qualifies today (0 sections) and
is currently reported by it (work item `559cb636`, `unresolved`, from 07-15). **Trading a
naming defect for a lost detection would have been the worse deal, and nothing would have
reported the trade.** Both name-side rows moved in one transaction, and I re-ran the
detector's own join afterwards to confirm the page is still in it under the new name.

**Measured before applying — every item, none inherited from the 07-31 precedent:**
no collision on the target name (0); `site_plan_sections.page_name='tool-arena'` (0, so no
sections to re-key); `site_plan_imagery` keys on `scope_ref`, not a page name (0);
`page_components` keys on `page_id`; `pages.status` already `active`, which the lookup
requires and which the earlier rename never had to check; `site_plan_pages` carries its own
`slug`/`url`, so `name` is not the URL source; nav renders `nav_label='Arena'` /
`title='The Arena'`, not `name`. Both UPDATEs scoped **by ID, never by name**, so a
concurrent rename could not make them hit the wrong row.

**Red then green, at the code's own query rather than a paraphrase of it.** The Tier-4 lookup
run verbatim (`name IN ($2, 'tool-'||$2)`, `site_id`-scoped, `status='active'`) returned
**0 rows before** and **`/tools/arena/index.html` after**. Served page **byte-for-byte
identical**: md5 `4a2d2030e2f6d2630f6497f68705a067`, 32,553 bytes on both sides.

**Queue checked first**, per the rule: no open work item on that page, and the only two vonc
orchestrations in the last hour were the gauntlet lane's council runs, neither touching arena.

> **A byte-comparison nearly became worthless without my noticing.** The page measured
> **31,431 bytes** when I first fetched it (~12:50Z) and **32,553** when I took the
> pre-change baseline (~15:00Z) — it had been rebuilt underneath me by its own lane's
> 12:45 redeploy. Had I compared the *old* figure against the post-change fetch I would have
> reported a 1,122-byte "change" caused by my rename, which caused none. **A baseline is only
> a baseline if it was taken immediately before the change** — on this tree that window is
> minutes, not hours.

### And the naming fix immediately exposed a real defect it had been hiding

With the page resolvable I ran the arena's **existing** fence — written by `tool-generator`
on 2026-07-14 and **never once executed** — through `try_fence.go`. 5 checks, both profiles:
`status`, `boots`, `console` and `mobile-fit` all pass; **`take-submit` FAILS on both**, with
`timeout 30000ms exceeded waiting for locator('#take-input')`.

**`#take-input` does not exist.** The served page has **no `id` attributes at all** and
exactly one form control — the site chrome's `.mobile-menu-toggle`. So the fence's only
behavioural assertion has never been satisfiable, and nobody could have known, because the
run died on page resolution before reaching it.

**This is the strongest available argument for P1a's priority.** The naming defect was not
merely cosmetic bookkeeping: it was *masking* a substantive disagreement between a tool's
published acceptance contract and its markup. Fixing the name did not create a defect, it
stopped hiding one.

**I did not choose between the two readings** (fence is stale vs tool is missing its
control), and **I did not fire the cluster run.** On a failing verdict the judge inserts an
`improve_tool` item routed to `handler_agent='tool-improver'`
(`tool_acceptance_actions.go:711`) — an automated fixer, aimed at a page whose
`rebuild_policy` is `owned`. Letting a one-shot fixer guess at a design question about
another lane's tool is the wrong way to settle it. Checked the handler before firing, which
is the point of that rule.

### Communicated, and delivered where it will actually be read

Owner of the arena identified by evidence, not assumption: the `gauntlet_dead_cta` / vonc6
lane holds `p4_sources/backups/backup_arena_html_template_2026-07-27.html` and owns vonc's
tool pages. Wrote `CONTRIB_2026-07-31_arena_tool_is_now_acceptance_testable.md` into **their**
directory (their own `CONTRIB_` convention), **and appended a dated pointer as item 6 of
their `HANDOFF_2026-07-31_continue_here.md` §4 NEXT ACTIONS, their words untouched** — because
a doc dropped in a directory is not delivery, which this fleet has already measured
(`a-quiet-git-log-is-not-silence`; the one time a pointer was appended into a cold-start doc,
the owning thread read it and acted within the hour). Their cold-start had no mention of
arena at all, so without the pointer the CONTRIB would have sat unread.

The CONTRIB tells them three things: what I changed and why the second row mattered; that the
fleet-wide "orphaned component, decide whether it should exist" record about their live tool
was **false and is corrected**; and the fence/markup disagreement with both remedies, the
exact dispatch command, and the 120s-deadline warning (two failing `fill` steps burn ~60s of
it before anything else runs).

### RESULT: PASS

`CHECK_naming_contract.sh` — **BROKEN A: 0, BROKEN B: 0**, first PASS since it was written.
30 canonical tool components: **12 testable now** (was 9 two days ago), 10 authoring backlog,
8 neither; 12+10+8 = 30, reconciled. The 10 with no PLAN at all remain honest rather than
misleading — nothing claims they were tested.

---

## 2026-07-31 (evening, fresh thread) — the Go gate is PROVEN: the running pod printed its own subject_type vocabulary

Picked up `HANDOFF_2026-07-31b_continue_here.md` §3, the single open next action: prove the Go
half of `subject_type='component'` genuinely shipped rather than merely being in a binary built
after the commit. It is now proven, in one dispatch, and the proof is stronger than the
handoff asked for because the failing arm of the probe **prints the vocabulary out of the
running binary** instead of arguing from a build date.

### The route: an inline workflow override, so NO agent row was written

The handoff recommended seeding a scratch `agent_definitions` row with one `load_doc_context`
step. I found a smaller route while reading how the dispatch resolves a workflow:
`selectWorkflow` (`platform/messaging/processor.go:906-1005`) checks **Priority 1: an inline
workflow override in the message** (`config.workflow`, :922-928) *before* any DB lookup, and
returns immediately if it finds one. The comment on it says "(for testing)" — which is exactly
this.

So the whole probe workflow travelled **inside the Kafka message**. Nothing was written to
`agent_definitions`, so there was nothing to snapshot, nothing to deactivate, and nothing to
clean up. Confirmed after the run: `SELECT count(*) FROM agent_definitions WHERE type LIKE
'%probe%'` returns 3, all three `is_active=f` and created 2026-07-24 — another lane's scratch
agents, none of them mine.

**Why the misfire is inert as well as visible.** If Priority 1 ever stopped taking precedence,
the fallthrough is Priority 2 (`FindBestGroup` on `config.agent_type`) — and I deliberately
used `doc-subject-gate-probe`, which does not exist in `agent_definitions` — and then Priority
3, the pod's own definition. The pod is `AGENT_TYPE=generic`, and `generic`'s entire workflow
is a single no-op `complete` step (`"description": "No-op — scheduled task pre_query already
did the work"`, `processing_mode: task`, `timeout_seconds: 10`). So a misfire does nothing at
all, *and* it is detectable, because `current_step` would read `complete` instead of one of my
distinctive step names. I wrote that third outcome into the script as a named verdict, **VOID**,
rather than letting it look like a pass — §7's defect class is a check that reports health it
never measured, and "the probe did not run" is the way this probe would have done it.

### The design: two steps in one dispatch, and the second one is the control

- `probe_subject` — `load_doc_context` with `subject_type: component`. **The test.**
- `probe_vocab` — `load_doc_context` with `subject_type: zzz-probe-invalid-$$`. **The control.**
  It MUST error, and its error message is `docSubjectTypesQuoted()`'s rendering of
  `validDocSubjectTypes` **as compiled into the running pod**.

The control earns its place twice over: it proves `probe_subject`'s route was *capable* of
failing (otherwise a green result proves nothing about the gate), and it turns the probe from
a yes/no into a **read-out** of the live vocabulary. That last part is the thing a grep cannot
do, and the reason is in the code: the vocabulary entries are short string literals, which Go
compiles to immediate comparisons that never reach rodata, but the quoted list in the error is
built **at runtime** by joining the slice.

### RESULT: PASS on every arm

Correlation `8f564028-6fc6-488c-96d2-c2e362b243b2`, pods `v1.0.1215` (both replicas),
`COMPLETED` at `finish`, inside the first 3-second poll.

| observation | value |
|---|---|
| `doc_subject.has_plan` | **true** |
| `length(doc_subject.plan_body)` | **827** — byte-identical to the body I generated |
| `doc_subject.criteria_json <> ''` | **t** — the fence extracted |
| `__step_error.failed_step` | **`probe_vocab`** — the control, not the test |

And the read-out, verbatim from `collected_data->'__step_error'->>'message'`:

```
step probe_vocab failed: failed to execute action load_doc_context: load_doc_context:
subject_type must be one of 'tool', 'pipeline', 'experience', 'action',
'experience-pattern', 'component', 'landmine', got "zzz-probe-invalid-1914600"
```

Seven types, from the pod itself. So:

1. **`component` is in the running binary's vocabulary** — not inferred from a build date.
2. `docResolveSubject` accepted it, the PLAN body travelled back through the Go action intact,
   and `extractCriteriaBlock` found the fence. That is the whole `load_doc_context` path, which
   is what an S6-for-components dispatch will use.
3. **`landmine` is in it too** — which independently corroborates the landmine-verifier lane's
   claim that its dependency (commit `7290433f2`) is live on v1.0.1215. I did not set out to
   check that; the control printed it.
4. Both DB CHECKs and the single Go list now demonstrably agree. The split-contract class
   (`bugs_closed/064`, migrations 163 and 184, and the `landmine` gap found live two days ago)
   has, for the first time, a **runtime** check rather than a build-time one.

### CORRECTION — the handoff predicted the wrong failure string, and it would have read as inconclusive

`HANDOFF_2026-07-31b` §3 (carried forward verbatim from the previous handoff, so this error is
two handoffs old) says the FAIL signal would be `unsupported subject_type "component"`. **It
would not have been.** That wording belongs to `docSubjectGateReason`
(`doc_subjects_common.go:96`), which is called by exactly one action —
`persist_diagnosis_note_action.go:83`. The route the handoff recommends, `load_doc_context`,
goes through `docResolveSubject` (`write_doc_plan_action.go:143-145`), whose message is
`subject_type must be one of …, got "…"`. A session grepping the step output for the predicted
string would have found nothing on either a pass or a fail and had to work out why.

The cheap check that would have caught it: **read the function the recommended route actually
calls, not the one whose name matches the concept.** Logged in `WRONG_CALLS.md`.

**What this does and does not prove about point 4 of the four-enforcement-point checklist.**
Both gates consult the same package-level `validDocSubjectTypes` through
`isValidDocSubjectType` — one slice, one file, no second copy — so proving membership in the
running binary carries to `persist_diagnosis_note` as well. But I did **not** dispatch
`persist_diagnosis_note` with a component subject, and I am not claiming I did: what was
watched is the shared list, via one of its two callers. DOC-068's open review question (no test
asserts that branch specifically) is untouched by this run.

### Cleanup, and the one thing left in the database

The probe PLAN row was written under `source='handoff-goproof'` (dry run first, per RUNBOOK §9:
827 chars stored = 827 generated, exactly one `criteria` fence, no `:name` interpolation), and
deleted after the run — `DELETE 1`, then `count(*)=0`, and `doc_plans` is back to 0 component
rows. **So `SELECT subject_key FROM doc_plans WHERE subject_type='component'` is empty again,
and DOC-068's `verify-later` about a real component PLAN being written and read by a gate still
stands.** The capability is proven; a *use* of it is not. Those are different claims and the
register keeps them separate.

The `orchestration_states` row IS the evidence, and that table is retention-clocked, so the
read-out is pasted above rather than pointed at. Re-running is cheap:
`scripts/PROBE_doc_subject_go_gate.sh component teaser-reveal-panel`.

---

## 2026-07-31, later still — reconnaissance for item 2 (S6 for components), before starting it

The handoff and the PLAN both call this step "wiring, not construction," citing the
`smart-contrast` pilot as proof the mechanism works end to end. Read `request_browser_run`
(`platform/orchestration/actions/tool_acceptance_actions.go:87-152`) before believing that,
because the pilot's own write-ups (`CONSULT_2026-07-30_next_tool_build.md:72`,
`PROPOSAL_2026-07-30…:520`) describe it as "11/11 checks asserting real arithmetic" — that is
a tool, not a component. Nothing in this lane's docs shows a component ever going through this
action. Worth being precise about what "proven end to end" actually covers.

**What the code does, concretely.** `RequestBrowserRunAction` resolves a single URL to test by:
1. reading `function` (a string) from `input_data.spec.function` — hard-errors if empty
   (line 98-100, not config-driven for the fallback path);
2. looking it up directly against `pages`: `SELECT url FROM pages WHERE site_id=$1 AND name IN
   ($2, 'tool-'||$2)` (lines 139-145) — i.e. it assumes **one function ⇒ exactly one page**,
   which is the same three-way identity (`doc_plans.subject_key == pages.name ==
   content_components.function`) P1a's naming check already polices for tools.

**That identity does not hold for a component, and it is not close.** Checked live:
`teaser-reveal-panel` (`content_components.id = '22c12251-73aa-4232-bd67-ef9edcfe8061'`) is
placed via `page_components` on **5 distinct pages across 2 distinct sites**
(`SELECT count(DISTINCT pc.page_id), count(DISTINCT p.site_id) FROM page_components pc JOIN
pages p ON p.id=pc.page_id WHERE pc.component_id=...` → `5|2`). A component is fleet-shared by
design (`content_components` has no `site_id` — RUNBOOK §5 already says this); a tool is not.
So "the page for this subject" is not a well-formed question for a component the way it is for
a tool, and `request_browser_run`'s SQL has no way to express "which of the 5" without new
input.

**What this means for item 2, concretely — not a blocker, but not free either:**
- `tool-acceptance-agent`'s shape (`ensure_site_record → load_docs → request_run → judge →
  complete`) is still the right skeleton to copy — `ensure_site_record` already resolves one
  `site_record` per dispatch, so the existing per-site scoping is compatible with "one site,
  one placement" if the caller supplies which site/page, rather than the action inferring it
  from a function name.
- `request_browser_run` itself needs either (a) a new `page_id_field`/`site_id`-and-`page_id`
  input path that bypasses the `pages.name` lookup entirely when present, resolving instead via
  `page_components.component_id = <uuid> AND page_id = <given>`, or (b) a sibling action. (a) is
  less code and keeps one action's config surface, at the cost of a branch in a function that
  currently has exactly one path; (b) keeps `request_browser_run` untouched (lower blast radius
  on a working tool path) at the cost of near-duplicate plumbing (headers, envelope, profiles).
  Not decided here — a real design choice, not a default.
- Either way this is a change inside `platform/orchestration/actions/`, which is council-gate
  scope per CLAUDE.md (the "platform seams" section) — even though it is additive. Per the
  2026-07-29 owner ruling, it only needs an RFC if it changes what a shared mechanism
  *guarantees*; adding an opt-in `page_id_field` that nothing calls until a component agent
  names it does not change `request_browser_run`'s existing guarantee for tools, so the normal
  council gate (not architecture review) looks like the right one — flagged for whoever builds
  it to confirm against the actual diff, not asserted in advance of one existing.
- `teaser-reveal-panel` is a real, live component (confirmed above), not an invented name — the
  PLAN already picked it as the first real target "because its history is fully written down."
  Its 5 placements mean the first real dispatch should pick ONE (site_id, page_id) pair
  explicitly rather than trying to resolve "the" page for it.

Nothing built yet. This is written down before starting so the next thread does not
re-discover the one-function-one-page assumption the hard way, the way the arena rename and
the P1a naming check both had to re-discover their own version of "this identity doesn't hold
the way it looks like it should."

## 2026-08-02 — P1's tail closed: `teaser-reveal-panel` has a real, proven, persisted fence

Picked the concrete placement to develop against by re-running the 5-placement query above
and taking the most recently rendered row: `leopardessconsulting.co.uk/services.html`
(`page_id=ebc2c413-61e2-465e-b22b-9aab0167abc9`, `site_id=4851f6fc-71cf-4160-a270-e03d6d3e0732`),
confirmed HTTP 200 and a real 6-item `content_data` before writing anything.

Read `page_components.rendered_html` for that exact row before authoring a single check —
this component turned out to need no JS at all for its core reveal (`content_components.js_content
IS NULL`; it is native `<details>`/`<summary>`), but the sibling-close-on-open and the
`?open=<key>` deep-link both live entirely in the fleet-shared `/assets/js/snippets.js`, loaded
via a plain non-deferred `<script src>`. That distinction mattered for the mutation proof below.

Authored `fence_teaser_reveal_panel.json`: 12 checks, profile-gated to 15 evaluations
(9 desktop-only + 3 on both), scoped to what this component actually contracts for (PLAN's own
Behaviour contract) rather than a mechanically copied tool fence — no arithmetic, no
`spec.function`, and no rendered "N cards" text to assert a count through, so two named cards
(`first-card-present`, `last-card-present`) stand in for a count `selector_count` cannot make
(RUNBOOK §8's own point, re-confirmed). `try_fence.go`: **15/15 passed, zero skips, arithmetic
reconciled** (24 = 15 + 9 gated).

**Then the mutation half surfaced a wrong assumption about the lane's own tooling.**
`scripts/prove_fence_can_fail.go` read, from its own doc comment and RUNBOOK §8's phrasing, as
a generic harness driven by the fence file. Ran it as-is against this fence anyway, rather than
trusting that reading: **14 of its 17 hardcoded mutants reported "target string not present."**
Every one of those 14 is a literal string from `tool-review-council-simulator`'s own source
(`el.passN.textContent`, `rcs-bar-row`, `applyPreset('typical')`...) — the file's `mutants`
slice is that ONE tool's mutants, not a generic mechanism, and only the 3 tool-agnostic ones
(404 / wide-div / console-error) ever fired. **This is the same shape of mistake
`HANDOFF_2026-07-31c` §7 already named for `request_browser_run`:** a piece of machinery assumed
reusable because it was built to look generic, that in fact encodes an assumption specific to
the first thing it served. `[VERIFIED by running it, not by reading its header comment]` — the
header comment's own framing ("go run ... <fence.json> <url>") was the misleading part; the
`mutants` variable is what settles it.

Built a sibling, `prove_fence_can_fail_teaser_reveal_panel.go`, copying the proven architecture
(baseline-first; one string-replace per mutant; `page-serves-200` must survive every mutant
unless it IS the mutant; a coverage table asserting every check has a mutant) with this
component's own 12 mutants, one per check. **One genuine extension was needed, not just new
strings:** an optional per-mutant asset override. The sibling-close behaviour lives in
`snippets.js`, and the existing architecture 302-redirects every non-page request straight to
the live origin — there was no way to serve a mutated copy of an external asset. Added a second
throwaway-server path that, for exactly the one mutant that needs it, serves a locally
string-mutated `snippets.js` (`if (other !== card) { other.open = false; }` → commented out)
while every other request still redirects live, matching the page-body mutants' own fairness
rule (baseline gets the real asset via a normal redirect; only the one mutant needing it gets
the override).

**Result: 12/12 mutants caught, 12/12 checks watched red, baseline green.** Two mutants
(`pointer-events:none` on a card's `<summary>`, to prove the "real click" checks actually need a
real click) each cost close to Playwright's 30s action timeout — Chromium's own hit-test
resolves a `pointer-events:none` target's click to its ancestor `<details>` instead
("intercepts pointer events"), so the click call itself times out rather than silently
succeeding. That is the right failure mode (a synthetic `.open = true` cannot produce it), but
it pushed the whole run past the interactive shell's 120s foreground timeout into the
background — confirmed by watching the process stay alive (`ps aux`), not assumed hung.

Wrote the PLAN + two backfilled/new NOTES via RUNBOOK §9's supersede-then-insert, generated by
a Python script (not a shell string — the PLAN body contains triple backticks). Dry-ran with
`ROLLBACK` first, then committed for real: `doc_plans` row (19,953 bytes), two `doc_notes` rows
(a condensed backfill of the convention-037 file's history, 4,374 bytes; a new build/verification
note for this session, 5,340 bytes) — all three lengths asserted equal to what was built,
inside the transaction. **Then read the PLAN back OUT of the DB, extracted its `criteria` block
with the same first-triple-backtick rule the platform's own extractors use, diffed it
byte-for-byte against the authored file (identical), and re-ran `try_fence.go` against that
DB-extracted copy: 15/15 passed again.** Writing the field is not reading the field
(RUNBOOK §9) — this is the check that actually closes that gap, not the write alone.

This closes PLAN `P1`'s stated tail ("one real component gets a PLAN with a criteria fence...
every criterion watched to pass by hand... NOTES backfilled") and DOC-068's own open
verify-later line ("whether a component PLAN is ever actually written and read by a gate" —
still not dispatched through a gate, but no longer merely a deleted probe row either: a real,
persisted, mutation-proven contract now exists). **Still not done: dispatching this fence to
`browser-runner-adapter` in the cluster** — `request_browser_run`'s one-function-one-page
lookup still cannot express "which of this component's 5 pages", and that design choice
(extend vs. sibling action) is exactly where the previous entry above left it. Also not done:
old file-based `PLAN_teaser-reveal-panel.md`/`NOTES_teaser-reveal-panel.md` are marked
superseded with a pointer, per their own stated instruction, but left on disk as history.

## 2026-08-02 (later) — chassis rolled to v1.0.1229; re-checked, unaffected

Owner reported a fresh chassis build. Pod-checked rather than assumed: both `agent-chassis`
and `browser-runner-adapter` replicas are `v1.0.1229` (up from `v1.0.1219`, started
2026-08-02 18:39 UTC). `git log --oneline -- platform/orchestration/actions/tool_acceptance_actions.go`
shows no new commits; the commits between the two tags are unrelated lanes
(`bugs_closed/165`, `168`, `179`, `097`, portfolio work) — none touch
`doc_subjects_common.go`, `tool_acceptance_actions.go`, or any component-fence dispatch
path. Re-verified rather than trusted: `doc_plans` still holds the `teaser-reveal-panel`
row, `is_current=true`, 19,953 bytes, unchanged; a fresh pod-grep of the compiled binary
still shows a vocabulary marker present. Nothing in this lane's state moved.

Handing off to a fresh conversation for Part B (S6 dispatch wiring) rather than continuing
in this one — this thread has already been through the Go-gate proof, the fence build, and
two doc passes, and Part B is a genuine platform-code design decision plus implementation
(`platform/orchestration/actions/`, council-gate scope), better started with full budget
than squeezed into an already-long session. New handoff:
`HANDOFF_2026-08-02_continue_here.md`.

## 2026-08-02 (fresh session) — Part B decided and built, not yet dispatched

Read the standing five in the order `HANDOFF_2026-08-02` §1 names, then the actual
`request_browser_run` code (`tool_acceptance_actions.go:87-270` as it stood before this
session) before deciding anything — confirmed by reading, not by trusting the handoff's own
paraphrase, that the page lookup really is `pages.name IN (function, 'tool-'||function)`
with nothing that could take a `page_id`.

**Decided (a) vs (b): sibling action, not a branch.** Written up as D9 in the PLAN before
writing any Go, per the handoff's own instruction. Reasons in full there; short version —
`request_browser_run` is the one path every existing tool-acceptance run depends on, and
this lane has already shown once (D5′) that "the smallest possible platform change" is easy
to under-count. A sibling action makes the tool path's blast radius provably zero (nothing
in its source changes) rather than argued zero.

Confirmed two things by research before writing code, not assumed from the handoff's own
phrasing:
- **The registry**: actions dispatch through `GlobalActionRegistry` in `registry.go`
  (`GetAction`, ~line 1987), a flat `map[string]ActionDefinition`. `request_browser_run`'s
  entry sits at line 1166. Adding a new action needs one more map entry there, nothing else.
- **`content_components.function` is NOT reliably `name`** — checked the aggregate, not just
  the one row: only 64% of `section`-level components and 7% of `tool`-level rows have
  `name = function` (tools especially get a site-slug suffix on `name` at fork time). For
  `teaser-reveal-panel` specifically the two happen to be equal, but the new action resolves
  by `function` throughout, never by `name`, so this doesn't matter for correctness — noted
  here so the next session doesn't generalise from the one row that happens to match.

**Built:** `dispatchBrowserRun` — the shared tail (profile resolution through envelope
build, marshal, produce, and the awaited-response result) extracted out of
`RequestBrowserRunAction` into its own function, called by both actions. Only the two log/
error strings that literally said "request_browser_run" changed text (to
"dispatchBrowserRun") — grepped the repo first for anything that parses those exact
strings (nothing does; they're log lines, not a wire contract). `RequestComponentBrowserRunAction`
resolves its page via `page_components JOIN content_components ON content_components.function
JOIN pages`, scoped by the explicit `page_id` and `pages.status='active'`, keeping the same
"no fake pass" empty-criteria skip the tool action has. Registered in `registry.go` right
after the tool entry, and its `ActionInputSpec` registered in `tool_acceptance_actions.go`'s
existing `init()`, both by the exact pattern the tool action already uses.

**Verified, not assumed:** `go build ./...`, `go vet ./platform/orchestration/actions/...`
(one pre-existing `unreachable code` warning in `load_component_library_actions.go`,
confirmed via `git status` to be untouched by this change — not mine to fix here), and
`go test ./platform/orchestration/actions/...` all clean. Diffed the change against
`RequestBrowserRunAction`'s pre-change body line by line — its control flow for every
existing caller is byte-identical except the extraction itself; only the two message
strings noted above changed, and nothing parses them.

Registered the new action in the concept register as **DOC-072** (own entry — a new
dispatch path is exactly the "another workstream could call this and wouldn't know it
exists" bar CLAUDE.md sets), and updated DOC-068's own verify-later line to point at it,
same commit.

**Not done yet, and this is the whole of what remains:** council submission (this touches
`platform/orchestration/actions/`, in scope for the advisory gate), an image build/roll,
pod-grep proof it shipped, and the actual S6 dispatch against `teaser-reveal-panel`'s page
with a negative control in the same run. None of that can happen before the commit exists.

**A mistake, recorded per this file's own rule:** committed (`f6bfb7e6e`) with
`Council-Submitted: pending` — a placeholder, not a real correlation, written before actually
submitting. That trailer's whole design is a correlation ID 098 can resolve at report time;
"pending" resolves to nothing, and forward-only means this specific commit's trailer cannot
be corrected. **The real submission, done immediately after and for the record:**
`SUBMISSION_CORR=33d00513-2fd8-4872-ad5a-a19c24a1ae0b`,
`RUN_ORCH_ID=f458ee92-215b-43d5-84a0-644024a8a4c5`, submitted 2026-08-02. Anyone auditing
`098`'s coverage report and finding `f6bfb7e6e` unresolved: this is why, and this is the
correlation that actually answers for it. The check that would have caught this before
committing: run the trigger script FIRST, capture the real correlation, THEN commit — never
write the trailer as a promise to submit later.

**Stopping point, 2026-08-02: code committed and submitted, deliberately NOT rolled.**
`f6bfb7e6e` (+ correction `195a169ff`) is on HEAD. Council verdict pending
(`33d00513-2fd8-4872-ad5a-a19c24a1ae0b`). Per this handoff's own §5 ("do not roll the
chassis to ship anything — your commit ships on anyone's next roll regardless"), did not
build/deploy. Baseline recorded so the next roll is detectable: both `agent-chassis`
replicas on `v1.0.1229`, started `2026-08-02T18:39` — a pod-grep for
`RequestComponentBrowserRunAction` or `request_component_browser_run` returning empty
against that image is expected and NOT evidence the code is broken; it just predates the
commit. Next session (or this one, if the owner wants a roll triggered explicitly): confirm
a newer image, pod-grep for the marker, then run the S6 dispatch per §3 of the handoff, with
a negative control (wrong `page_id`) in the same run.

**2026-08-02, later — owner reported a fresh chassis build; checked rather than trusted it,
and it has NOT reached this deployment.** Both `agent-chassis` pods are still `v1.0.1229`,
still the same two pod names, **158 minutes old** — identical to the baseline above, not
restarted. `rollout status` reports the existing rollout already settled (`Progressing=True,
NewReplicaSetAvailable`), no second ReplicaSet, no in-progress rollout. So this is not even
the familiar "rolled but pod-grep the string, don't trust the roll" case — there is no
evidence of a roll reaching this namespace at all yet.

What IS true: `makefile`'s `IMAGE_TAG` is `v1.0.1230` on HEAD (`21defe33d`, another lane —
"v1.0.1229 was built without 2da3e08e5 (shrink guard)"), one ahead of what's running, so a
**build** may well have happened. But the `agent-chassis` production kustomization overlay
(`deployments/kustomize/services/agent-chassis/overlays/production/uk_001/kustomization.yaml`)
has an UNCOMMITTED working-tree diff bumping `newTag` from `v1.0.1222` to `v1.0.1229` — i.e.
even that file doesn't yet name `v1.0.1230`, and it's not this lane's change to touch or
commit. Read, not acted on. Asked the owner rather than guessing which of "build not pushed
yet" / "deploy not applied yet" / "wrong environment" was true.

## 2026-08-02 — P2 CLOSED: real S6 dispatch through `browser-runner-adapter`, with a negative control that proves the placement check, not just routing

Owner rolled `v1.0.1231` and said so. Checked before acting on it, per the same discipline
as the entry above: `kubectl get pods -l app=agent-chassis` — both replicas on `v1.0.1231`,
different pod names, started `21:39`, i.e. genuinely new pods this time, not a re-read of the
same ones. **Pod-grepped both replicas before dispatching anything**, per the fleet-wide
practice (a roll is not evidence a specific commit shipped): `strings /app/agent-chassis |
grep -c` for two markers — the exact registry key `request_component_browser_run` (6 hits:
map key, `RegisterActionInputSpec` call, logger field, several `fmt.Errorf` prefixes — a
LONG marker, not the short-literal trap D7 warns about) and the distinctive error string `"a
component can be placed on more than one page"` (1 hit) — both replicas, both positive. A
sanity negative control (`grep -c` for a nonsense string) returned 0 on both, confirming the
grep pipeline itself can actually distinguish presence from absence.

Re-verified the placement row before dispatching (the standing landmine: placements move) —
`teaser-reveal-panel` is still on the same 5 pages, `ebc2c413-...`/`services.html` still
active. Re-verified the `doc_plans` fence row is still current (19,953 bytes, unchanged).

**Built the dispatch by reading the LIVE `tool-acceptance-agent` workflow from
`agent_definitions`** (not by guessing from the PROBE script's simpler shape, which uses
literal `subject_type`/`subject_key` in config rather than the real workflow's
`subject_key_field` pointer) — `ensure_site_record → load_docs → request_run → judge →
complete`, `timeout_seconds: 600`. Copied it with two changes: `load_docs.config.subject_type
= "component"`, and `request_run.action = "request_component_browser_run"` with
`page_id_field: "input_data.spec.page_id"` added. Sent inline via `config.workflow`
(`selectWorkflow` Priority 1, no `agent_definitions` row — same technique as
`PROBE_doc_subject_go_gate.sh`), `agent_type` deliberately nonexistent so a misfire would
fall through to `generic`'s no-op `complete` rather than silently running the real tool
workflow.

**Added a `neg_control` step AFTER `judge`, in the SAME dispatch**, per this lane's own
standing rule that a green run and a run that skipped silently look identical unless
something is watched to fail. Deliberately did NOT reuse "some UUID that doesn't exist" —
that would only prove the query returns no rows for a garbage key, which `sql.ErrNoRows`
handles trivially. Instead picked a REAL, active page on the SAME site
(`fc505ab2-...`/`faq.html`) that genuinely does not carry `teaser-reveal-panel`, so the test
actually exercises the `page_components`/`content_components.function` JOIN failing to match
a row, not merely a page lookup failing to find a page. `page_id_field` for this step points
at a second, separate input field (`input_data.spec.bad_page_id`) so the real run and the
negative control use independent inputs in one message. `error_step` on `neg_control` points
at a step named `neg_control_confirmed_red` — the SAME "the must-fail arm's error IS the
pass" shape `PROBE_doc_subject_go_gate.sh` already established for the Go-gate probe.

**Result, correlation `e6a258eb-6ba1-44df-b344-16e42443975f`, `COMPLETED` in 31s (well under
the 120s per-request adapter deadline the fence was already sized for):**
- `current_step = neg_control_confirmed_red` — the negative control fired and was caught.
- Real run: `collected_data->'browser_run'->'response'->'summary'` = `{"passed":15,
  "failed":0,"skipped":9}`. Read the skip reasons, not just the count (D3's own rule): all 9
  are `"SKIPPED: not run on profile mobile"` — the fence's own intentional gating, none are
  `"<type> not implemented"` (the defect class that reads as PASS and suppresses re-checks
  for 7 days). 15+9=24=15 checks × profiles, arithmetic reconciled, matching `try_fence.go`'s
  offline 15/15 exactly — same evaluator (`RunChecksAction.Execute`), now proven reachable
  through the cluster dispatch path too.
- `judge`'s own verdict (`acceptance_verdict`): `all_passed: true`, `failed: 0`,
  `site_chrome_failures: 0` — confirms `judge_acceptance_results` needed no changes, exactly
  as predicted (it keys off `function`, never off how the page was resolved).
- Negative control's actual error, read from `collected_data->'__step_error'` (a FAILED step
  reports `status=COMPLETED` with the real message here, never in `error` — RUNBOOK §10):
  `"request_component_browser_run: component \"teaser-reveal-panel\" is not placed on page
  fc505ab2-a991-4421-85e1-fa856f5b7a39 (or that page is inactive)"` — the EXACT message the
  new code was written to produce, not a generic SQL or adapter error. This is what makes the
  control mean something: it proves the JOIN predicate is what's doing the rejecting, not
  some unrelated failure that happened to land on the same step.

**What this does and does not close.** DOC-068's own verify-later line — "an S6 run citing
its fence" — is now genuinely closed: a component fence has been dispatched through
`browser-runner-adapter` in the cluster and passed. **What it does NOT re-prove:** that every
individual check in the fence CAN fail — that was already established offline, mutation-proven,
12/12, by `prove_fence_can_fail_teaser_reveal_panel.go` (TL-036), calling the *same*
`RunChecksAction.Execute` the live dispatch also calls. Re-running that same mutation set
through the cluster would corroborate, not newly prove, the checks' own falsifiability, and
would cost real Playwright time (two of those twelve mutants took ~30s each per the earlier
entry) for no new information. **What today's negative control proves that the offline
harness could not** is the one thing that WAS new in this session: that
`RequestComponentBrowserRunAction`'s placement resolution — the part with no offline
equivalent, since `try_fence.go` never resolves a page ID, it's handed a URL directly — fails
closed in the real cluster, against the real DB, through the real dispatch path. That is the
gap this whole P2 phase existed to close.

**P2 is done.** Nothing outstanding in this lane except reading the council verdict
(`33d00513-2fd8-4872-ad5a-a19c24a1ae0b`) when it lands and acting on it if it's REVISE/REJECTED.

## 2026-08-02 — re-checked P3 before building it; it was stale, and measured what it collapses into

Owner asked to carry on to the next items, asking specifically for a re-check of P3 first.
Good instinct: **P3 ("the remaining gates, cheapest first") was written 2026-07-30, and D8 —
which retired exactly that approach — landed the next day, 2026-07-31.** Nobody reconciled
the two. Read D8's own words again rather than trust my memory of them: *"the discarded
stages are not disproved, they are unfunded. Any of them may return with evidence."* No such
evidence has surfaced since. So P3 as written does not get built. Corrected in the PLAN
(`6944902b1`) and re-labelled: P3 actually collapses into the authoring backlog (apply the
three FUNDED gate-types — S1 fence, S2 mutation-proof, S6 dispatch — to more subjects), which
is a real next step under its honest name rather than a fourth invented one.

**Measured the backlog rather than guessed at its size**, joining `content_components` against
`doc_plans` on `subject_key = function`, split by the subject_type each level actually maps to
(`tool` for `component_level='tool'`, `component` for everything else this ladder has touched,
i.e. `section` — DOC-068's own scope, not the other levels, see caveat below):

| population | active | with a current PLAN | with none |
|---|---|---|---|
| tools (`component_level='tool'`) | 49 | 13 | **36** |
| section components (`component_level='section'`) | 112 | 1 (`teaser-reveal-panel`) | **111** |

(Inactive rows exist too — 24 tools, 17 sections — deliberately excluded from the headline:
authoring a contract for a component nobody serves isn't backlog, it's waste.)

**Two things this measurement surfaced that are NOT part of the answer, flagged rather than
folded in:**
1. **15 `doc_plans` rows with `subject_type='tool'` have NO matching `content_components.function`
   at all** (`animated-favicon`, `smart-contrast`, `mind-map`, ... — full list in the query
   above). Checked two by name-similarity in case this was rename drift (features_open/028's
   own failure shape) — `smart-contrast` matches nothing even partially, `favicon` only matches
   `tool-favicon-generator`, not `animated-favicon`. **[UNVERIFIED]** whether these are
   framework-level dev tools tracked outside `content_components` entirely (plausible — the
   portfolio lane's memory mentions a separate "framework tool-build pipeline") or something
   else. Not chased further — out of scope for "measure the backlog," and worth a deliberate
   look later rather than a guess now.
2. **`component_level` has three other values this ladder has never addressed** (`header`,
   `footer`, `site`, `head`, `element` — 19 rows total, active+inactive). DOC-068's own `what:`
   line says it extends the substrate to *section* components specifically; whether `header`/
   `footer`/`site`-level components should also get `subject_type='component'` PLANs is an open
   question this measurement surfaces but does not answer. Excluded from the backlog table
   above rather than silently assumed in or out.

**Not proceeding to author fences for 111+36 subjects** — that's a real, large undertaking
or a decision to phase it, not a "carry on" default. Reported the size; next call on scope is
the owner's.

## 2026-08-02 — checked ownership of the two side items before touching either

**`features_open/028` (rename orphaning).** `who-owns.py` doesn't cover `features_open/` —
it's scoped to `bugs_open/`/`bugs_closed/` only, confirmed by reading its own output for
`028` (which resolved to two unrelated `bugs_closed/028` files by the SAME number, a known
trap, neither relevant). Read the actual file instead: filed 2026-07-30 by this exact lane,
at the council gate's instruction (`bug_historian` + `architecture` seats, correlation
`e5673868-...`, same round DOC-068 went through). **Status: FILED, unowned, not designed.**
It names its own first candidate — a cheap detector query counting orphaned docs per subject
type — as "start here if only one gets built," but that's a new build, not a check, and
wasn't what was asked (ownership, not implementation). Left filed and unowned; not picked up.

**`has_visible_area` backfill.** `who-owns.py 157` (the gating bug) resolves cleanly and
names `staged_component_build` itself as the likely owner — unsurprising, since this lane
found and fixed 157. But the ITEM in question ("checks owed to every EXISTING fence") is a
different population than the bug: queried every current fence-carrying PLAN fleet-wide
(`doc_plans WHERE is_current AND body ~ '```criteria'`) and cross-referenced for
`has_visible_area` — **30 of 33 fences fleet-wide don't use it**, and the 3 that do are all
dated 07-30 or later (i.e. authored after the fix, not backfilled into anything older). Those
30 span many other lanes' own subjects (`gauntlet-round-record`, `vonc-spark-game`,
`tool-arena-interface`, ...) — editing them unilaterally would be exactly the "compete rather
than contribute" mistake the ownership-check norm exists to prevent, even though the BUG that
unblocks this is this lane's own. No existing ticket tracked this (checked `bugs_open/`,
`bugs_closed/`, `features_open/` by grep first). **Filed as `features_open/029`** rather than
fixed — same shape as 028 (measured gap, ranked fix candidates, deliberately not designed or
owned), same precedent this lane already set the same week.

**Neither item picked up for implementation.** Both were genuinely "check before touching,"
and both checks concluded the honest next step is visibility (filing/leaving-filed), not
code — consistent with `HANDOFF_2026-08-02`'s own instruction for both.

## 2026-08-03 — council verdict read: APPROVED, and one objection is a real design lesson

Owner reported a fresh chassis build (`v1.0.1231` → **`v1.0.1238`**, new pod names, confirmed
by `startTime` before trusting it — same discipline as every prior roll-check in this file).
Then read the council verdict on `f6bfb7e6e` properly rather than just the headline.

**First mistake of this entry: the standard "latest council-gate note" query answers the
WRONG question.** `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
created_at DESC LIMIT 1` (the exact query CLAUDE.md itself prints) returns the fleet's most
recent council note **fleet-wide**, not mine — it came back naming a totally different
correlation (`e78c62e3-...`, someone else's submission). Caught by checking the printed
correlation against the one I actually submitted, not by trusting the row. Re-ran scoped to
`body LIKE '%33d00513-2fd8-4872-ad5a-a19c24a1ae0b%'` and got the right one.

**Verdict: APPROVED — 11 reviewers, 3 advisory objections (2 medium, 1 low-severity each of
the rest), 0 unreadable, `gated_by_truncation: false`** (checked per `bugfix_138`'s own
landmine — a truncated review gates identically to a real one, so this is a genuine, undegraded
result, not a false pass). No `Council-Reviewed` trailer needed on a new commit — the design
is explicit that `Council-Submitted` on the earlier commit (`195a169ff`) resolves automatically
at 098's report time, and forward-only forbids amending `f6bfb7e6e` to add one anyway.

**Read every objection rather than the headline, and verified the factual claims rather than
argued them:**

- **`editquality`, medium — "this ships inert... until a workflow names this action, the
  cited blocking mechanism is unresolved."** True, and already the exact wording of DOC-072's
  own verify-later ("whether a real `agent_definitions` row... ever gets seeded"). Not a defect
  in what shipped — P2's own gate (PLAN) was proving the DISPATCH MECHANISM works, which the
  live run did; a real automatic trigger (a due-sweep or scheduled discovery equivalent to
  whatever fires `tool-acceptance-agent` for tools) is genuinely separate, larger, unfunded
  work, not something P2 was ever scoped to include.
- **`reuse_agent`, low — possible duplicate resolution path.** Checked:
  `grep -rn "page_components" --include=*.go` fleet-wide. Every other hit is a different
  purpose (deploy, render, repair, admin-dashboard CRUD) — nothing else resolves "one page for
  a given component, for browser-run dispatch." No duplicate.
- **`tooling_provenance`, low — no `doc_notes`/`doc_plans` entry for the action itself.**
  Checked whether `subject_type='action'` is actually a general convention for orchestration
  actions: `SELECT subject_key FROM doc_plans WHERE subject_type='action'` returns exactly 4
  rows, all diagnosis-loop internals (`diagnose_build_gate`, `diagnose_prepare_fix_commit`,
  `diagnose_read_repo_files`, `spawn_agent`). **`request_browser_run` itself — the action this
  one sits beside, shipped long before this session — has no such entry either.** Not a gap
  this change introduced; consistent with actual (not aspirational) practice.
- **`guardian`, low — registry treated as a closed/counted set?** Checked: `ListActions` /
  `ListAllActions` / `ListActionsByCategory` / `ListDeprecatedActions` all iterate the live map
  dynamically; nothing hardcodes a count or an expected member list. No risk.
- **`guardian`, low — does the new `p.status='active'` predicate match the existing one?**
  Checked: byte-identical to `RequestBrowserRunAction`'s own existing tool-lookup predicate.
  Confirmed consistent, not a second, divergent definition of "deployable."
- **`debug_historian`, low — the `pages.status` landmine (naive filtering under-counts).**
  Checked: that landmine is about `linkablePageStatusPredicate`
  (`prepare_link_context_action.go:35`, `status NOT IN (...)`), a DIFFERENT predicate for a
  DIFFERENT purpose (can content link to this page) — not the browser-run-resolution predicate
  either action uses. No divergence; different question entirely.
- **`debug_historian`, low — no pod-grep mentioned in the plan.** True of the SUBMITTED plan
  (written before the roll existed to grep); already done in fact, the same day, and recorded
  above in this file. Timing artefact of round 1, not a gap in what happened.
- **`guardian`, medium — "'no behaviour change' rests on manual diff alone, not an existing
  test."** Checked, and this one is factually WRONG, with better evidence than argument:
  `TestRequestBrowserRunPayloadCarriesCaptureRenders` and
  `TestRequestBrowserRunCaptureRendersDefaultsOff` (`tool_acceptance_actions_test.go:411,427`)
  call `RequestBrowserRunAction` directly against a `capturingProducer` mock and assert on the
  ACTUAL produced Kafka payload — i.e. they exercise `RequestBrowserRunAction → dispatchBrowserRun
  → Producer.ProduceWithValidation` end to end, the exact call path in question. Both existed
  before this session and both passed after the extraction (`go test` output already recorded
  above). The manual diff was corroboration, not the only evidence — this objection undersold
  what was actually checked.
- **`prior_art_librarian`, medium — THE REAL ONE. "The absence claim may be false: `url_field`
  already exists on `RequestBrowserRunAction` as an override that bypasses the `pages.name`
  lookup entirely."** Checked and **this is factually correct**:
  `grep -n "url_field" tool_acceptance_actions.go` shows it at line 160 (original code,
  predating this session) and line 370 (mine). So a caller COULD already have fed
  `RequestBrowserRunAction` an explicit URL — resolved by ANY mechanism — via `url_field`,
  bypassing the name lookup, with zero new code in the tool action at all. **This means a
  smaller design existed: a standalone resolver action doing ONLY the
  `page_components`/`content_components` placement-JOIN, writing a URL into `collected_data`,
  then calling the EXISTING `request_browser_run` unchanged via its `url_field`/`function_field`
  overrides.** That design needs no `dispatchBrowserRun` extraction and touches
  `RequestBrowserRunAction`'s own file not at all — a strictly smaller blast radius than what
  shipped, on the exact axis D9 itself argued from. **This is the "I costed a change by reading
  one enforcement point... there were two" lesson again (D5′), and I made the same class of
  miss: read `RequestBrowserRunAction`'s page-resolution branch closely, but not its own
  existing escape hatch one field below it.**

**What this does and does not change.** The code that shipped is correct, tested (including
by the two end-to-end tests above, which the guardian objection wrongly assumed didn't exist),
and proven live in the cluster with a working negative control — reverting proven, working
code to rebuild a marginally smaller equivalent is a real cost with no functional benefit, and
not something to do unilaterally on an advisory, non-blocking objection. **Not silently
dismissed either**: recorded here in full, and in DOC-072's own register entry, as a real
design lesson for whoever builds the next sibling action — check the target action's own
`ActionInputSpec` for an existing override field before concluding "nothing can express X,"
not just its main resolution branch. Owner's call whether this is worth a follow-up
refactor; not undertaken here without that call, given the cost/benefit above.

## 2026-08-04 — cross-lane CONTRIB parked with the vigilant-designer/offer-analyser lane

Owner asked whether tool/component testing could integrate with the decisions the visual
designer and offer/benefit analysers make (that lane = `vigilant_designer_offer_analysis`,
session `e0b18fd9-...`). Researched their actual state before answering: the live design
pipeline persists decisions to `site_specs` (`design_intent`, `resolved_composition`), the
A2 critic and B4 offer-analyser (benefit analysis inside it) are planned-not-built and will
write findings via `write_findings` into the findings machinery — nothing on either side
touches `doc_plans`/fences today, so any integration is new wiring, not extending a link.

**The seam identified:** their findings assert claims about the SERVED page ("benefit
surfaced?", "CTA present"), and their finding contract already carries an `acceptance_test`
field — if page-shaped finding types populate that field in the browser-runner's existing
check vocabulary (selector/type/text) instead of free prose, their verifiers gain the
real-browser instrument (their planned NEW verifiers for exactly this class are static —
lexicon re-scan, DB query) and our lane gets fence checks derived from decisions, which is
P4's "dynamic generation of gates" arrived at bottom-up. The known design problem, named
not solved: scope asymmetry — our fences are fleet-wide per subject (D4), their decisions
are per-site, so per-site checks likely live on the finding, not in a fleet fence.

**Owner ruling: wait until their thread matures, then suggest coordination.** So: CONTRIB
parked in their directory now (`CONTRIB_2026-08-04_your_decisions_could_be_fence_checkable_
when_your_vocabulary_settles.md`) — explicitly staged for their A2/B4 vocabulary-authoring
moment, nothing asked of them today — plus a dated, attributed pointer appended to their
current cold-start handoff so the note is found at the right moment rather than becoming
folklore (same pattern as the gauntlet-lane CONTRIB precedent, and the reciprocal
brochure-lane CONTRIBs already flowing both ways between these directories). The note
hands them this lane's own council-caught lesson at the cheap moment: check whether the
browser-runner + criteria vocabulary already expresses "verify the served page" before
their verifiers grow a new mechanism for it — the `url_field` override + inline-workflow
recipe means a page-scoped claim can be driven with zero new platform code already.

Nothing built, nothing seeded, nothing routed at their queues. Next move is theirs, when
they reach vocabulary authoring; ours is P3-backlog scope (owner call, still open).

## 2026-08-05 — owner call landed: TAKE ON the backlog; handoff written

The P3-backlog scope question is answered — owner: *"let's take on this backlog."* Written
up as `HANDOFF_2026-08-05_continue_here.md` (supersedes the 08-02 handoff), whose two
load-bearing choices are: (1) the per-subject recipe is P1+P2's own proven sequence
generalised (read the real artefact → author → `try_fence.go` → per-subject mutation prover
→ RUNBOOK §9 persist+readback → S6 dispatch with negative control), and (2) **calibration
before commitment** — a first tranche of ~5 subjects (2–3 naming-clean tools, 1–2 section
components picked by placement count and simplicity), timed, then a cost-per-subject report
back to the owner to set pacing, because 147 fences is ~10 sessions and pace is an owner
call, not a default.

**A loss caught and repaired while writing it:** the S6 component-dispatch script that
closed P2 existed only in this session's scratchpad — and the scratchpad had already been
wiped by tmp-cleanup. Rebuilt from this conversation's own record, parameterised (5
positional args; the proven teaser-reveal-panel invocation in the header), and committed as
`scripts/DISPATCH_s6_component_run.sh` with the preconditions and the negative-control-page
rule in its header. The check that would have caught it sooner: an instrument proven in a
run belongs in the workstream's `scripts/` in the SAME session that proves it — scratchpad
is for drafts, not for the only copy of anything cited by a handoff. Also added RUNBOOK §13
(the backlog census query with its three reading-traps, and the dispatch pointer).

## 2026-08-05 — calibration tranche begins: preconditions re-run, subject 1 dropped for a real defect, subject 1(v2) complete to persist

**Preconditions (11:12–11:16 UTC).** Census re-run per RUNBOOK §13: active tools moved
49→58 total, with-PLAN 13→19 (six new PLANs all `source='tool-generator'`, dated 08-02
onward — tools born through the canonical path are born compliant, exactly as the handoff
predicted; the backlog is pre-existing stock only). Sections unchanged: 111 of 112 active
with no PLAN. `CHECK_naming_contract.sh`: **PASS, 0 broken** — 49 canonical tool
components, 18 testable now, 15 page-fine-no-PLAN (the cheap candidates), 16 neither.
Browser-runner pod `v1.0.1252` (started 09:10 UTC today) carries the vocabulary: three
long positive markers →1, negative control →0, one exec.

**Tranche picked:** tools `tool-gas-unit-converter`, `tool-fuel-cost-estimator` (same
site, warmed-path datum), `tool-loan-vs-savings`; sections `hero` (213 placements/16
sites) and `call-to-action` (192/16), both JS-free. `tool-equity-release` excluded at
selection: page row active but the URL 404s — the build_status-is-history landmine, live.

**SUBJECT 1 (`tool-gas-unit-converter`) DROPPED at S1, ~15 min.** Reading the real thing
killed it: the served page renders the full converter STRUCTURE with every text slot
empty — `<h2></h2>`, blank labels, blank `<option>`s, blank button — because the template
carries 28 `{{.placeholders}}` and the placement's `content_data` is NULL. The external
JS wires conversion logic but hydrates no text, so visitors see an unlabelled form.
**Breadth measured before concluding:** the placeholder-template+empty-data class is 91
active placements fleet-wide, but only the 2 TOOL placements are live-broken
(`tool-gas-unit-converter`@gaswholesalers rendered-empty; `tool-ab-test-calculator`@idea.uk
serves RAW `{{.section_heading}}` etc — never rendered at all). The 89 SECTION placements
(hero/call-to-action across 4+ sites) SERVE FINE — spot-checked finetuning.uk/pricing,
full copy present — so for sections, empty `content_data` does not mean broken serving
(the served render sourced its copy elsewhere; at worst a latent rerender risk, a
different fact, not folded into this one). **The queue already knows the two broken
tools**: `required_fields_missing` + `empty_section` for gas-unit-converter sit in
`needs_human_review`, its full-rebuild `needs_page` was closed `wont_fix`, and
ab-test-calculator's content items are `failed`. Detected-but-parked, not undetected — no
new filing (grep of bugs_open/closed + 016b found the pattern already recorded, §"zero
shared keys"); goes in the owner report instead. S1 doing exactly what D8 funds it for.

**SUBJECT 1 v2 (`tool-fuel-cost-estimator`) COMPLETE to persist, S6 DEFERRED (~11:16–12:00).**
- S1: served page healthy (no empty tags, no raw placeholders); behaviour read from the
  inline script — live recompute on input, no submit button, unit/period toggles,
  wholesale≥retail guard, per-field validation, aria-live results. Golden captured:
  5000 gal/wk @ $3.85/$3.45 → $2,000.00 weekly, $0.40/gal, $104,000.00 annual;
  monthly → $24,000.00.
- Fence: 13 checks (`fence_tool_fuel_cost_estimator.json`), mobile gated to
  status/overflow/console, `computed_values` golden for the arithmetic, default-state
  checks first, the reload-using guard check late per §8's ordering rule.
- `try_fence.go`: every tool-scoped check green, arithmetic reconciled (26=13×2). The ONLY
  failures are `no-console-errors` both profiles, and they are REAL: the live origin 404s
  `/assets/images/logo.png` on every gaswholesalers.com page (the `assets` row exists,
  active, since 03-05; the file is absent at the served path — an asset-deploy gap, site
  chrome, not the tool). Fence NOT weakened to match the broken site (that is the
  fixing-the-checker landmine); defect goes to the owner report.
- S2: sibling prover `prove_fence_can_fail_tool_fuel_cost_estimator.go` (teaser
  architecture; asset-override machinery dropped — all tool JS is inline; ONE declared
  deviation: serves a local 1×1 PNG at the logo path in EVERY run incl. baseline, else the
  control aborts red on the chrome defect and proves nothing). **Baseline green, 13/13
  mutants caught, 13/13 checks watched red, first run**, 4m26s. The 10×10-overflow-hidden
  collapse mutant (instead of display:none) kept inputs actionable — no 30s timeouts.
  Both arms of the annualisation constant proven independently (52→50 caught only by the
  golden; 12→10 caught only by the monthly-toggle check).
- §9 persist: generated by `gen_fce_plan_sql.py` (scratchpad), dry-run ROLLBACK then
  COMMIT; length asserted INSIDE the transaction via DO/RAISE (7,280 bytes, exact);
  read back out, fence extracted from the DB copy → byte-identical to the authored file,
  re-run through the evaluator → identical result. Marker count 1.
- **S6 DEFERRED with stated reason, in the PLAN body itself**: a failing verdict raises
  `improve_tool` and routes tool-improver at the tool, but console failures carry NO
  chrome attribution (only overflow does — `judge_acceptance_results` reads), so a red
  run here routes a fixer at a cause no tool edit can reach. Checked first: NO scheduled
  acceptance sweep exists (`scheduled_tasks`: 0 rows matching), so persisting the fence
  arms nothing automatic; the dispatch decision stays manual. Fire
  `tool_acceptance_run.sh` for this subject the moment the logo serves.

**Wall-clock so far: preconditions 4 min; subject 1 (dropped) ~15 min; subject 1v2
~45 min end-to-end including building its prover.**

## 2026-08-05 (contd) — subjects 2–4 complete end-to-end; one false red became a LANDMINES entry; tranche DONE

**SUBJECT 2 (`tool-loan-vs-savings`) COMPLETE, ~12:02–13:05, iteration count 2 + one
cluster round-trip.** S1 was a gift: the tool's own inline comments record its golden
vector (1000 @ 7.5%/5.0% → £75.00/£50.00, computed ON LOAD), the copy-bank badge idiom
(badges start empty; JS fills exactly one — a JS-dead page asserts nothing false), and the
deliberate strict-`>` tie rule (tie → savings). Fence: 14 checks pinning exactly those.
Two authoring lessons the instruments caught, not me:
- `try_fence` iteration 1: I asserted the copy-bank casing ("Better option"); CSS
  uppercases the badge and innerText reads the RENDERED text → "BETTER OPTION". Fixed
  case-insensitive — assert what the visitor reads, not what the source holds.
- **The cluster round-trip caught a class the offline harness structurally cannot: fence
  vocabulary NEWER than the deployed binary.** v1's tie check used the `reload` step
  action. S6 (corr `3874c8b5-63bb-44d5-93ec-f2086f63567c`): 16 passed, 1 failed —
  `unknown step action "reload"`. Timeline proven at the artefact: browser-runner
  v1.0.1252 built 09:10 UTC; `reload` landed at HEAD **11:05 UTC the same morning**
  (`67a4c50bd`, bugs_open/126); I authored at ~12:10 with HEAD's evaluator, which knows
  it. Pod-grep: `"reload navigation failed"` → 0, control → 1. Unknown check TYPES skip;
  unknown STEP ACTIONS **fail the check** — so this raised a real `improve_tool` item
  (`6c06b0ad`) routing tool-improver at a healthy tool. **Cancelled it with the full
  reason written into `result`** (status was `detected`, unclaimed). De-reloaded BOTH
  fences (lvs: the tie is now built from accumulated state — with the 40% band selected,
  12.5% × (1−0.4) is bit-identical in IEEE doubles to 7.5%, so the tie is exact;
  fce: the guard check now raises wholesale above the retail already on the page),
  re-proved both (14/14, 13/13), superseded both PLANs to v2, re-dispatched.
  **Re-run corr `ab6434ee-a5d9-411f-8626-2846500f32f7`: 17/17 PASSED in-cluster, 18s,
  0 chrome items, every skip 'not run on profile mobile'.**
  → LANDMINES entry appended (offline fence harnesses run HEAD's evaluator; prove STEP
  ACTIONS against the pod too, or `git log -S` the word vs the image build date), synced
  to doc_notes (`landmines-sync.py --apply`, 1149 rows; entry flagged NEEDS_VERIFICATION
  by the frozen code index, expected per MEMORY's 108 landmine).

**SUBJECT 3 (`hero`) COMPLETE, ~20 min.** 213 placements/16 sites, no JS. Fence is the
small static shape the handoff predicted: exists / visible (BOTH profiles — a hero can
collapse on mobile alone) / h1-non-empty / 200 / overflow / console. The h1 check models
the gas-unit-converter defect class (`{{.headline}}` is unconditional in the template, so
empty data serves an empty h1). try_fence 10/10 first try (1366×630 desktop, 390×526
mobile — the 70vh/60vh contract visible in the numbers). Prover 6/6 first run, **26s**.
Persisted (`subject_type='component'`, second ever after teaser-reveal-panel). S6 via
`DISPATCH_s6_component_run.sh` at the finetuning.uk/pricing placement with about.html as
the negative control: corr `141d4fb9-837f-4e1e-b128-5af3abb9445e`, landed
`neg_control_confirmed_red`/COMPLETED — **10/10 passed, control red, 0 chrome items**.

**SUBJECT 4 (`call-to-action`) COMPLETE, ~15 min.** 192 placements/16 sites, sibling of
hero in every respect (unconditional h2, 023-gated optional buttons, no JS). try_fence
10/10 first try, prover 6/6 first run. Persisted. S6 at the same pricing placement,
contact.html as control: corr `b8858007-3b05-4272-aa33-54554ba717af` →
`neg_control_confirmed_red`, **10/10 passed, control red**.

**Tranche wall-clock (11:12–13:20 UTC ≈ 2h10m):** preconditions 5m · subject dropped at
S1 15m · fce ~45m (new prover architecture built here) · lvs ~65m (~25m of that the
reload incident, a one-off that is now a landmine) · hero ~20m · cta ~15m. Marginal cost
estimate once instruments are warm: **static section component ~15m; interactive tool
~30–45m**; a subject with a live defect costs a finding instead of a fence (that is S1
working, not overhead).

**Open ends left deliberately:**
- fce S6 deferred until gaswholesalers' logo serves (stated in its PLAN's Known state).
- The two live-broken tools (gas-unit-converter@gaswholesalers rendered-empty,
  ab-test-calculator@idea.uk raw placeholders) are in the queue already
  (needs_human_review / failed / wont_fix) — owner report, not new filings.
- `tool-equity-release`: page row active, URL 404s — one more artefact-vs-row drift datum
  for the report.

## 2026-08-05 (contd) — D10 exhaustive clearance begins: production line built, batch 1 (5 sections) done end-to-end

**Owner ruling D10 (recorded in PLAN): option (c), exhaustive.** Scope guards: no fence
for a tool with no serving page (creates CHECK_naming_contract's BROKEN A); component
levels beyond section/tool stay out (DOC-068 boundary).

**Line instrument: `prove_fence_mutants_file.go`** — the S2 architecture with the mutant
list as per-subject JSON (nothing hardcodable; the trap was hardcoded lists READING as
generic). Validated by reproducing call-to-action's exact 6/6 from
`mutants_component_call_to_action.json` before first use. Line rules learned/encoded:
prefer SINGLE-INSTANCE placements (a rename-first mutant is uncaught when the selector
matches a second instance — section--generic had 3 on the first-choice page); verify
every `from` string count in the SERVED page (scripted into the batch generator).

**Backlog truths the census surfaced:** 35 of 109 "active" sections have ZERO active
placements — nothing to dispatch against; listed, not fenced. `gauntlet-round-record` is
a section component whose PLAN sits under subject_type='tool' (gauntlet lane's; flagged,
skipped). `ported-page` placements on loanandmortgagecalculator + loancash (58 rows)
point at pages whose SERVED HTML carries no component markup at all — placement-row vs
artefact drift; webdesign.co.uk's 97 are real.

**Batch 1 — five subjects, all end-to-end (fence → try → mutants-file prove → §9 persist
→ readback → S6 dispatch with neg control):**
| subject | checks | prover | S6 |
|---|---|---|---|
| generic-text-block (99 placements/15 sites; NO data-component attr — root is `section.section--generic`) | 6 | 6/6 | 10/10 + control red (`b7dca437`) |
| article-body (49/10) | 6 | 6/6 | 10/10 + control red (`5a2fe125`) |
| features (33/6; asserts ≥1 `.feature-item` — empty-grid class) | 7 | 7/7 | 11/11 + control red (`885efc90`) |
| ported-prose (28/2; existence/visibility only — opaque ported HTML, stated in PLAN) | 5 | 5/5 | 9/9 + control red (`d50de1a0`) |
| hero-about (28/13) | 6 | 6/6 | 10/10 + control red (`72b82464`) |

**Two real finds, honestly routed:**
1. **`article-body` ships no `pre`/`code` overflow handling** — the first-choice placement
   (blog/multi-agent-failure-isolation…) genuinely scrolls horizontally on mobile (798px
   `code` in a 390px viewport), caught by the fence's own trial. Template greps `pre`→f,
   `overflow`→f. Proof moved to a clean placement; the defect is recorded in the PLAN
   body itself and here. One open work item already touches that page.
2. **ported-page deferred:** its only markup-carrying placements are on webdesign.co.uk,
   where the LOCAL prover harness cannot get a green baseline for harness reasons (the
   Cloudflare RUM beacon fails CORS from a localhost origin; the page's own
   `/search.json` fetch turns cross-origin under the 302 redirect). The live page passes
   try_fence 9/9 clean. Needs a dedicated prover with declared deviations (fce
   precedent): serve /search.json locally + strip the beacon uniformly. NOT persisted.

**Batch-1 wall clock ≈ 45 min for 5 subjects (~9 min/subject)** — the line beats the
calibration estimate once S1s are batched and mutant files are generated with verified
counts.

## 2026-08-05 (contd) — batch 2: nine more sections end-to-end, all S6-green with controls red

| subject | checks | prover | S6 CID |
|---|---|---|---|
| info-card-grid (22/11) | 6 | 6/6 | 561b98ff 10/10 |
| content-block-about (15/7) | 6 | 6/6 | 6f2ef61a 10/10 |
| tool-cta (10/4) | 6 | 6/6 | 8fcea50f 10/10 |
| hero-contact (12/11) | 6 | 6/6 | ac7cbce2 10/10 |
| differentiators (21/12; ≥1 item asserted) | 7 | 7/7 | c5b3c9e9 11/11 |
| faq (17/7; real-click gesture: summary → [open], pointer-events mutant) | 7 | 7/7 | 5f82e62b 11/11 |
| about-content (16/9) | 6 | 6/6 | abdc65fb 10/10 |
| contact-form (13/12; form.contact-form asserted) | 7 | 7/7 | 993bf362 11/11 |
| contact-info (9/9; heading has designed fallback so asserted; grid NOT asserted per bugs_closed/140) | 6 | 6/6 | 10b550c1 10/10 |

All negative controls landed `neg_control_confirmed_red`. Conditional headings
(faq/about-content `{{if .section_title}}`) deliberately NOT asserted; unconditional ones
are. **One more instance of the known-defect class:** about-content's first-choice
placement (dartsonline.com/about.html) fails no-console-errors on the site's
`/assets/images/hero.jpg` 404 — the EXACT defect bugs_closed/128 measured on 2026-07-31,
still serving 5 days later (detection flag-only; repair never dispatched). Proof placement
moved to finetuning.uk/approach.html; recorded in the PLAN body.

**Running D10 tally: 16 sections + 2 tools done end-to-end; deferred: ported-page
(harness CORS), fce S6 (gaswholesalers logo), gas-unit-converter + ab-test-calculator
(broken pages, queue-parked), equity-release (row-vs-URL drift).**

## 2026-08-05 (contd) — batch 3: eleven more sections end-to-end, incl. ported-page rescued by two declared harness accommodations

**Instrument evolution, forced by evidence (2 subjects), validated before trust:**
`prove_fence_mutants_file.go` gains two OPTIONAL, DECLARED, uniform-across-all-runs
accommodations, set in the mutants file where a reviewer reads the mutants: `serve_local`
(paths the page's own JS fetches same-origin — the redirect harness turns them
cross-origin and CORS fails no-console-errors on pages that are CLEAN live; robot-hands
`/data/latest-news.json`, webdesign `/search.json`) and `strip` (third-party beacon tags
that POST to an external origin and can never pass CORS from localhost; Cloudflare RUM).
Backward-compatible (bare-array mutant files unchanged); re-validated by reproducing
call-to-action's exact 6/6 AFTER the change. Neither accommodation can affect a mutant's
verdict (uniform incl. baseline), and no-console-errors keeps its own dedicated mutant.

| subject | prover | S6 CID (all `neg_control_confirmed_red`, all_passed) |
|---|---|---|
| brief-explanation (6/5; serve_local) | 6/6 | 11504241 10/10 |
| hero-case-studies (4/4) | 6/6 | 03e0acb0 10/10 |
| hero-services (6/6) | 6/6 | 88e2affd 10/10 |
| hero-tool (6/3) | 6/6 | cf96ff9e 10/10 |
| services-grid (5/4; root class ≠ function name — resolve by attribute) | 7/7 | 785895cf 11/11 |
| system-stats (5/4) | 6/6 | ed533c8b 10/10 |
| testimonials (6/2; root class social-proof-section; ≥1 item; headline conditional) | 6/6 | 02208554 10/10 |
| tool-guide-intro (5/4) | 6/6 | cd5f6cbe 10/10 |
| tool-list (6/4) | 6/6 | f71a2cff 10/10 |
| use-cases-list (5/2; ≥1 article) | 7/7 | 8288e2b1 11/11 |
| ported-page (97 real placements on webdesign; 58 drift rows on lmc/loancash) | 5/5 | 66fad769 9/9 |

Line refinements this batch: heading extraction tolerates inline markup (`<em>` inside
robot-hands' h2 broke the `[^<]+` regex — non-greedy `.*?</h2>` now); CSS
`url()` background references probed per page after darts' hero.jpg lesson;
lendzy.co.uk dropped as a proof site (origin flaky behind Cloudflare, 522s).

**Running D10 tally: 27 sections + 2 tools end-to-end.**

## 2026-08-05 (contd) — batch 4: six more sections; running tally 33 sections + 2 tools

| subject | prover | S6 CID (all controls red, all_passed) |
|---|---|---|
| content-listing (3/2; NO data-component — root `section.section--articles`) | 5/5 | d45f261d 9/9 |
| departments-grid (4/2; class `team-section` SHARED with leadership-team — attribute-resolve) | 5/5 | 7c520ef2 9/9 |
| evidence-chart (3/2; header fully conditional — existence/visibility only) | 5/5 | 73b8ab53 9/9 |
| guide-list (4/3; unconditional heading asserted) | 6/6 | 53533939 10/10 |
| leadership-team (3/3; same-class twin of departments-grid, BOTH on aao/about.html) | 5/5 | 46922650 9/9 |
| mechanism-flow (4/2) | 5/5 | bcee65e8 9/9 |

Line rule from the twins: when two components share a root CLASS and can share a PAGE
(team-section), every check resolves by the data-component attribute and grid-child
assertions are dropped (a string-replace mutant cannot scope to the first `.team-grid`).

## 2026-08-08 — resumed after 3 days; batch 5 closes the uncontested static stock (9 more end-to-end, 8 blocked on chrome, 2 more drift rows)

**State re-established before acting** (3-day gap): no other session touched the lane;
census 78 sections + 37 tools no-PLAN (consistent with 08-05 + 2 new tools born);
browser-runner now v1.0.1267 (rolled today 16:27 UTC) and **pod-grep confirms `reload`
is IN the deployed binary now** ("reload navigation failed" → 1) — the 08-05 vocabulary
skew is closed at the fleet end. gaswholesalers logo **still 404 (3+ days)** — fce S6
stays deferred. Scratchpad had been tmp-wiped (as the 08-05 handoff predicted for
anything left there; all instruments were committed, nothing lost).

**Batch 5 — 19 candidates, split by evidence:**
- **9 done end-to-end** (all S6 all_passed, controls red):
  about-commercial-block (3d5d46cd 9/9) · archetype-combinations (11403df8 10/10) ·
  case-studies-list (196c63c1 10/10) · gripper-spec-sheet (cfb009ae 9/9) ·
  hero-use-cases (ef852c33 10/10) · image-hover-card-grid (dcac64db 9/9, darts INDEX is
  clean unlike its about page) · intent-probe (21d25046 10/10, asserts the POST form;
  serve_local latest-news.json) · portfolio-showcase (e371e7f0 9/9) · stat-band
  (39ac13ba 9/9; shares its proof page with portfolio-showcase, attribute-scoped).
- **8 BLOCKED on site chrome 404s** — fences authored + committed, NOT proven/persisted
  (baseline cannot go green; the bar stays every-check-watched-red): archetype-grid +
  intent... no — archetype-grid (relojistas glosario hero.jpg), directory-listing
  (**vetcomparison.uk hero.jpg — a NEW member of the bugs_closed/128 family, not in its
  07-31 list**), funding-fit + patent-check (idea.uk favicon+hero), game-master-explanation
  + platform-comparison (vonc about hero.jpg), people-feature-block (fundamentallyai
  about hero.jpg), social-proof (gaswholesalers logo). One asset fix per site unblocks
  its subjects; the hero.jpg family now measures at LEAST 7 sites.
- **2 more PLACEMENT-DRIFT rows**: featured-content (finetuning/ai-guides) and pricing
  (gaswholesalers/how-pricing-works) have placement rows whose served pages contain no
  such component at all. Effectively unplaced; listed, not fenced.

**Running D10 tally: 42 sections + 2 tools end-to-end.** Remaining: ~16 interactive
sections (JS read each), ~10 ready tools, 8+3 chrome-blocked, 7 lane-owned
(coordination), ~35 unplaced + drift rows (listings).

## 2026-08-08 (later, fresh session) — batch 6 opens the INTERACTIVE stock: 5 subjects end-to-end, and one of them lies to the visitor

**State re-established before acting.** Fresh chassis roll landed while this session
started: chassis + browser-runner both **v1.0.1269** (chassis pods up 22:01–22:02 UTC,
browser-runner 22:02). No other session had touched the lane since `5c5ab9256`. Census
re-run: **112 active sections, 43 with a PLAN, 69 without; 60 active tools, 23 with,
37 without** — consistent with 08-08 morning plus batch 5's nine.

**Vocabulary pod-grepped in the DEPLOYED binary before authoring anything** (the 08-05
skew landmine). On `browser-runner-adapter-c6ccdf86c-ndlft`: `"reload navigation failed"`
→ 1, `"non-numeric w/h in result"` → 1, `"computed_values"` → 1, `"no element matches"`
→ 1. **`"has_visible_area"` → 0 and that is NOT a miss** — it is the short-literal
artefact `bugs_closed/157` §3 warns about; the same binary carries that check's own
error strings, which is the long-marker proof. HEAD's switch (`run_checks_action.go:554`)
admits exactly `page_status_ok, selector_exists, selector_count, no_console_errors,
no_horizontal_overflow, interaction, has_visible_area, computed_values`, and this batch
uses no vocabulary outside it.

**Candidate set re-measured, not carried forward.** 17 active section components with
`length(js_content) > 0` and no PLAN (the 08-05 handoff said "~16"). Batch 6 took the
five most-placed of them.

| subject | placements/sites | checks | prover | proof page | S6 CID |
|---|---|---|---|---|---|
| news-listing | 9 active / 8 | 8 | 8/8 | robot-hands.com/news/index.html | `5cb3f14d` 12/12 |
| latest-news | 6 / 6 | 8 | 8/8 | robot-hands.com/index.html | `55f5c2cf` 12/12 |
| case-studies-grid | 4 / 3 | 8 | 8/8 | leopardessconsulting.co.uk/who-we-help.html | `10651039` 12/12 |
| contact-block | 3 / 3 | 8 | 8/8 | robot-hands.com/contact.html | `57d41cf0` 12/12 |
| blog-listing | 1 real / 1 | 8 | 8/8 | fundamentallyai.com/platform-log/index.html | `9f5d8639` 12/12 |

All five landed on `current_step='neg_control_confirmed_red'`, `all_passed: true`, and
**every skip is a profile gate** (`…@mobile` on a desktop-only check) — not one
`not implemented`, which is the skip that reads as a PASS.

### The line rule this batch adds: an interactive fence needs one check a STATIC render cannot satisfy

Batches 1–5 asserted structure. For a JS-driven component that is not enough: every
check can pass with the component's script deleted, and the fence would then certify a
dead panel. So each of these five carries exactly one assertion reachable **only** if the
script ran, chosen from what the JS itself observably does:

- `news-listing` — `#news-listing-count` must read `\d+ item`. The footer is server-rendered
  `style="display:none"` and the count element is empty; the script un-hides it and writes
  the text after `/data/news-archive.json` resolves. Read with `InnerText`, so a hidden
  footer reads `""` — the hiding does half the discrimination for free.
- `latest-news` — `#news-footer a.news-more-link` must exist. The server renders
  `<div id="news-footer"></div>`, literally empty; that anchor exists in no server render.
- `case-studies-grid` — click the `strategy` filter, expect
  `.csg-filter-btn[data-filter="strategy"].active[aria-current="true"]`. The served page
  contains **zero** `aria-current` attributes (grepped), so nothing but the handler can
  produce one.
- `blog-listing` — same shape, but its script sets **no** `aria-current`, so the assertion
  is `.active` alone — and it clicks `cat1`, **never `all`**, because `all` carries
  `.active` in the server render and asserting it would pass with the script deleted. That
  distinction is the whole difference between a driven check and a decorative one.
- `contact-block` — click submit on an EMPTY form, expect `#cb-status.cb-error` to carry
  text.

Each was mutation-proven by replacing the component's own `<script src=…>` with an inert
`<script>/* … */</script>` — no 404, so no console-error collateral, and the only check
that goes red is the driven one.

### Two authoring traps found the hard way, both about the component re-rendering over your mutant

1. **You cannot kill a JS-rendered list by deleting the server-rendered items.** The
   obvious mutant for `at-least-one-item` — remove or rename the `<article>` markup — is
   defeated twice over: the `from` string occurs 20 times (the prover demands exactly
   one), and the script re-renders the list from the feed anyway. What works is renaming
   the **container class** the fence's descendant selector goes through
   (`class="news-listing-items"` → `…-renamed`), leaving the `id` the script binds to
   untouched: the items are still created, they are just no longer inside the listing the
   contract names.
2. **`serve_local` for the feed is mandatory AND it disarms feed-side mutants.** Without
   it the same-origin `fetch("/data/…")` turns cross-origin under the redirect harness and
   CORS reds `no-console-errors` on a page that is clean live. With it the prover serves
   the real feed verbatim in every run — so no mutant can change the DATA, only the page.
   That is why every driven mutant here attacks the SCRIPT or the SLOT, never the feed.

### Three finds, routed rather than fixed

1. **`bugs_open/228` FILED — `contact-block` tells the visitor "Your message has been
   sent" and there is no transport in the component at all.** No `action`, no `method`,
   and `grep -cE 'fetch\(|XMLHttpRequest|sendBeacon|form\.submit\('` over the SERVED
   2,100-byte `/tools/assets/contact-block.js` returns **0**. The success message comes
   from a 1,200 ms `setTimeout`, then `form.reset()` wipes what the visitor typed. Live on
   three pages, one of them robot-hands' contact page. Blast radius measured, not
   estimated: a census over all 30 active form-bearing components asked three questions
   (`action` present / JS transport present / claims-sent) and **contact-block is the only
   one false on both of the first two while true on the third**.
   **The fence deliberately asserts the VALIDATION path and NOT the success message** —
   asserting it would have made this lane's own contract vouch for a false claim to a
   visitor, which is `bugs_open/161`'s failure applied before the fact rather than after.
   Landmine appended + verification dispatched (`0b5aceb7`).
   `[UNVERIFIED, deliberately not claimed]` the sibling `contact-form` (13 pages) resolves
   its action to a **`mailto:` URL with `method="POST"`**; what a current browser does with
   that is a separate question and is recorded in 228 as an adjacent observation with the
   check that would settle it, not as a finding.
2. **`finetuning.uk/index.html` 404s FIVE `case-studies-grid` card images** —
   `/assets/images/case-study-{facilities,financial-data,legal-rag,logistics-strategy,private-ai}.jpg`.
   Same detected-but-never-repaired class as the `hero.jpg` family (`bugs_closed/128`).
   It is why the proof placement is leopardess, not finetuning.
3. **A third placement-drift row**: `leopardessconsulting.co.uk/blog.html` has a
   `blog-listing` placement row and its served HTML carries **no `data-component`
   attribute at all** — not this component's, not any. Joins `featured-content`,
   `pricing` and the `ported-page` 58.

### Instrument committed that batches 1–5 kept losing

`scripts/gen_component_plan_sql.py` — the persist generator (RUNBOOK §9's supersede-then-
insert, dollar-quoted, driven from a per-batch manifest). Batches 1–5 hand-rolled this in
a session scratchpad every time and lost it to tmp-cleanup every time. Its length assert
is a `DO`/`RAISE`, **not** a `SELECT`: `ON_ERROR_STOP` ignores a non-empty result set, so
a verify block made of SELECTs cannot stop the COMMIT (RFC_006's landmine). Dry-run
(ROLLBACK) first, then `--apply`; all five bodies asserted to the byte, then **read back
out of `doc_plans` and diffed byte-identical against the proven files**, and two were
re-run through the evaluator from the DB copy — writing the field is not reading it.

**Running D10 tally: 47 sections + 2 tools end-to-end.** Remaining: ~12 interactive
sections, ~10 ready tools, 8+3 chrome-blocked, 7 lane-owned (coordination), ~35 unplaced
+ 3 drift rows.

## 2026-08-09 (same session, past midnight) — batch 6b: two more interactive subjects, one of which is not interactive at all

Fleet rolled again mid-session: chassis + browser-runner **v1.0.1270** (chassis pods up
08:49:38Z). Re-grepped before dispatching, not assumed: `request_component_browser_run`
→ 6 with a negative control → 0; browser-runner's `"non-numeric w/h in result"`,
`"no element matches"`, `"computed_values"` → 1 each. Dispatched at 13 min past the
restart, clear of the ~300 s window.

| subject | checks | prover | proof page | S6 CID |
|---|---|---|---|---|
| game-list | 7 | 7/7 | gamesdesign.co.uk/index.html | `67eba6b8` 11/11 |
| ai-readiness-quiz | 9 | 9/9 | leopardessconsulting.co.uk/ai-readiness-quiz.html | `fa948522` 13/13 |

Both landed `neg_control_confirmed_red`; every skip a profile gate.

### `game-list` is a FALSE MEMBER of the interactive pile — its JS is dead twice over

The census puts a component in the interactive pile when `length(js_content) > 0`. For
`game-list` (963 bytes) that signal is wrong on both counts:

1. **It binds nothing that exists.** The script queries `.gl-filter-btn` and
   `#gl-load-more-btn`. Neither string occurs in the component's own `html_template`
   (grep: **0**) nor in either served page (**0**). There is no filter bar and no
   load-more button in this template at all; both handlers are no-ops.
2. **It is not delivered anyway.** Neither placement page emits
   `<script src="/tools/assets/game-list.js">` — neither references `tools/assets` at
   ALL — while `curl https://gamesdesign.co.uk/tools/assets/game-list.js` returns **200
   with the real code**. The only script either page loads is `/assets/js/snippets.js`,
   334 bytes, containing no `game-list` code.

So the fence is deliberately static, and the PLAN says so rather than leaving a reader to
wonder why an "interactive" subject has no driven check.

**The delivery gap was then MEASURED rather than generalised** — this is the part worth
copying. My first pass sampled ONE page per JS-bearing component, and that sample
disagreed with itself: `contact-block` came back script-not-loaded on finetuning.uk and
loaded on robot-hands. So I probed **all 38 active placements** of JS-bearing section
components against their served pages:

- **2 placements render the component without loading its script** — `game-list` on both
  gamesdesign pages, and nothing else. A two-page anomaly, not a class.
- **2 placements are drift** (component absent from the served HTML entirely):
  `blog-listing` on leopardess (already known) and — **new** —
  `contact-block` on `finetuning.uk/case-studies.html`.
- The other 34 load their script correctly.

Not filed as a bug: visitor impact of the `game-list` case today is **zero** (the
component renders correctly as a static list and the missing script would add nothing).
What it costs is a false signal in the backlog — it read as a 30–45-minute interactive
subject and was a 9-minute static one.

### CORRECTION to `bugs_open/228`, made the same day, and it is my error not the system's

228 first said **"three live pages"**. It is **two**. `finetuning.uk/case-studies.html`
has the placement row and serves no `contact-block` markup at all. Corrected in place in
the bug file with a struck-through row and a dated note; logged in `WRONG_CALLS.md`.

The uncomfortable part, recorded because it is the transferable bit: **the rest of 228 was
verified at the artefact on purpose** — I curled the page and the 2,100-byte script rather
than trust `content_components`. Then for the blast radius, the one number a reader uses
to judge urgency, I asked `page_components` and wrote down its answer, three paragraphs
below my own explanation of why that table is not evidence. And this lane had already
found four drift rows and written "re-verify the row AND the served markup" into its own
handoff — I re-verified the row I was going to TEST on and trusted the rows I was only
going to COUNT, which is exactly backwards: a row you test on announces its own absence,
a row you count never does. **A placement-row join is an UPPER BOUND on live pages. Quote
it as one or spend the curls.**

### A fence-authoring rule the quiz forced: check order is load-bearing, and a later step cannot re-do an earlier one

`answering-a-question-enables-next` was first authored self-contained — click Start, then
click an option. **It failed on first trial**: the evaluator drives every check against
ONE shared page in declaration order, so by the time that check ran, the earlier
`real-click-starts-the-quiz` had already advanced past the start screen, the Start button
was hidden, and the click sat there for the full 30 s timeout. The fix is not to make the
check self-contained — it cannot be — but to let it continue from the state its
predecessor left, which is also what a visitor does. Recorded in the PLAN body so nobody
"fixes" it back.

**Running D10 tally: 49 sections + 2 tools end-to-end.** Remaining: ~10 interactive
sections, ~10 ready tools, 8+3 chrome-blocked, 7 lane-owned, ~35 unplaced + 4 drift rows.

## 2026-08-09 (later) — the contact forms deliver; and I duplicated an owning lane, which is the more important entry

**Owner instruction: "enable the contact forms end to end."** Done for `contact-block`
(2 pages, the bug) and `contact-form` (13 pages, 12 live). Full state, and the coordination
note that matters, are in `bugs_open/228` § "CONTRIBUTION 2026-08-09". Handoff:
`HANDOFF_2026-08-09_continue_here.md`.

### THE MISSTEP, first, because it is the transferable part

`bugfix_228_contact_block_transport` owned this bug: standing five, PLAN diagnosing the
cause **one level below my bug file** (the sanitiser's *presence gate*, not the missing
action), council rounds 1–2, Go fix `85390ee33` committed at **09:39Z — seven minutes
before my re-render**, image built and pushed, apply script prepared and **deliberately
gated** on the roll. I ran `who-owns.py` before FILING (08-08) and not before FIXING
(08-09). Their `NEW_FORM_TAG` and my template edit are byte-identical — two independent
designs converged, which is the good news and the waste.

I also broke their stated gate and produced exactly the `action=""` they predicted. What
caught it: reading `RenderTemplateReportingMissing` to find out why the sanitiser had not
fired, and finding a comment **citing my own bug number**, written by someone else minutes
earlier. `WRONG_CALLS.md` has the entry. **The rule: an ownership answer ages in HOURS
here; re-run it at the point of WRITE, and especially for a bug you filed yourself, because
filing a good bug file is what causes someone to pick it up.**

### The measurement that unblocked it without the roll

`sanitiseFormAction`'s gate is `present`, **not non-empty**. So seeding
`content_data.form_action = ''` on the placement makes the **currently deployed** binary do
the repair — `''` is already in `nonDeliveringFormActions`. Applied to both served
placements; both then rendered the correct mailto on **v1.0.1270**, which pod-grep confirms
does NOT carry `85390ee33` (`"seeded empty form_action for sanitiser"` → 0 on both
replicas, positive control `form_action` → 2). This does not make their fix redundant: it
makes theirs the class fix and mine the two-row special case.

### Measured, not reasoned: a `mailto:` FORM does not reliably carry the message

`probe_mailto_form_encoding.go`, Chromium, 08-09. `method=GET` **replaces the action's
query** — the platform's own `?subject=` is destroyed and every field becomes a mail
header. `method=POST` (either enctype) hands the text to a request **body**, which a
`mailto:` URL cannot carry. This settles the `[UNVERIFIED]` note I left in 228 about
`contact-form`'s 13 pages: they were losing the message to the browser's discretion. Hence
both new scripts BUILD the URL with explicit `subject=`/`body=` and navigate.

### The fix, and why the success string is now unreachable by accident

One attribute is load-bearing: `contact-block`'s form now carries
`action="{{.form_action}}"`, because the sanitiser only engages when the **template
mentions the field** — the mechanical reason the platform's per-site address repair had
been fixing the sibling for a fortnight and never touched this one. Three destination
shapes, three honest outcomes (http → report the server's status; `mailto:` → "opening your
email app", never "sent"; nothing → refuse and say so). `prove_contact_delivery.go` drives
all five branches in a real browser for both components; both PASS, and both live pages
were then driven as a visitor against the **served** page.

**Two harness bugs found en route, both worth more than the fix:** a `<script src>` placed
before its own markup silently self-disables (the script hits `if (!form) return`, the
browser does a native submit, and all five cases fail looking like a component defect); and
`contact-form` has no `novalidate`, so the **browser** refuses an invalid submit first —
asserting only our own status text scored a correct refusal as a defect.

### Two silent platform traps, now encoded in `RERENDER_page.sh`

1. **`page-rerender` has two paths.** Without `input_data.spec.reason` ∈
   (`image_landed`|`section_data_resolved`|`cta_links_stale`) it assembles from STORED
   `rendered_html`, so a TEMPLATE change never appears — **while still republishing
   `/tools/assets/*.js`**. New script, old markup, `COMPLETED`, green asset check. Measured:
   asset 2,100 → 7,345 bytes with the form tag untouched.
2. **`page_name` must be at `input_data.spec.page_name`** (read out of the live
   `save_sections` config, not guessed). Elsewhere → `{"skipped":true,"success":true,
   "sections_saved":0,"reason":"no page name"}`: three sections re-rendered and discarded,
   reported as success. `bugs_open/095`'s family.

Also: **`kubectl run -i` inside a `while read` loop eats the loop's stdin** — the first
rollout dispatched exactly one of ten pages and exited silently. Use an array or
`< /dev/null`.

### Blast radius, measured on a canary before the other twelve

Whole-page diff on `leopardessconsulting.co.uk/contact.html`: **17 lines changed, every one
of them the intended edit** (form id, status div, script ref, status CSS). Nothing else
moved. Only then did the remaining pages go.

**Owed:** `idea.uk/contact.html` — re-rendered and committed to `gqls/vm-sites` at 10:01Z,
served object still 05 Aug; a different repo and host from the other twelve.
**Also owed:** `contact-block`'s fence deliberately asserted the validation path only, so as
not to ratify the fake success. That reason has now gone — the fence should gain a check
that the success state is downstream of a destination.

## 2026-08-09 (afternoon) — the roll landed; the class fix is PROVEN OPERATING, and contact-block's fence now carries the check whose absence let 228 ship

**`85390ee33` live on v1.0.1274** (chassis + browser-runner, pods up 12:23Z). Pod-grepped
both replicas with controls in both directions: `"seeded empty form_action for sanitiser"`
→ **1** each, against **0** on v1.0.1270 three hours earlier — a measured transition, not a
green reading. Positive controls `form_action` → 3 and `request_component_browser_run` → 6;
negative control → 0.

**Then the part a pod-grep cannot do.** A binary containing a string is not a mechanism
operating. So I DELETED my own morning workaround — the hand-seeded
`content_data.form_action = ''` on both served placements — and re-rendered each page:

| | hand_seeded | rendered |
|---|---|---|
| leopardessconsulting.co.uk | false | `action="mailto:leopardess@contactforsales.com?subject=…"` |
| robot-hands.com | false | `action="mailto:robot-hands@contactforsales.com?subject=…"` |

**The check could have come out otherwise, which is the whole point** — without the class
fix reaching that path both would have rendered `action=""`, which is exactly what they did
on v1.0.1270 this morning. `count(*) WHERE content_data ? 'form_action'` over contact-block
placements is now **0**: no per-row special case left to rot. Live end-to-end re-run on the
class mechanism, still PASS.

### The fence edit I owed, and why it is the one check that matters here

`contact-block`'s fence deliberately asserted the validation path only, so this lane's own
contract would not vouch for a claim that was false to the visitor (`bugs_open/161`'s class,
applied before the fact). **That reason has gone**, and leaving the fence as it was would
have meant the component's contract still could not tell a delivering form from a dead one.

Added `form-has-a-real-destination`: `form.cb-form[action]:not([action=""])`. Deliberately
scheme-agnostic, so it stays true when this is pointed at a real receipt endpoint, and
deliberately not asserting the VALUE — that is derived per site from `sites.email` by
`sanitiseFormAction`, and pinning it would make the contract site-specific and duplicate a
rule that lives in one place.

**It is mutation-proven against BOTH states this bug actually passed through**, which is
what makes it a regression test rather than a decoration:

- *the form loses its destination entirely* (the original 228 defect) → caught
- *the form keeps the attribute but it is empty* (the v1.0.1270 intermediate state) → caught

Fence now 9 checks; **10/10 mutants caught, 9/9 watched red, baseline green**. Persisted,
read back byte-identical, S6 re-dispatched: `f3cd89a2` **13/13 passed**, landed
`neg_control_confirmed_red`, every skip a profile gate. PLAN body carries a visible
SUPERSEDED note over the old "deliberately does not assert" paragraph rather than quietly
deleting it.

**Tally unchanged at 49 sections + 2 tools** — this was a contract strengthened, not a new
subject.
