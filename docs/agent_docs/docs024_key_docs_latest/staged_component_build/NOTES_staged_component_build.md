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
