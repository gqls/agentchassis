# PLAN — staged component build with a gate per stage

**Lane adopted 2026-07-30** by `brochure_component_library` on owner direction:
*"This provenance and ladder project is now this lane's project."* Anchor feature:
`features_open/027`. The design argument and the evidence live in
`PROPOSAL_2026-07-30_step_by_step_build_with_stage_gates.md`; **this file holds
decisions and their reasons**, and corrections to the proposal are recorded here
rather than silently edited into it.

## What we are trying to do

Make a component build a sequence of small, individually-verified steps instead of
one leap, so that a more complicated component is *more stages of the same size*
rather than a bigger risk. Each small part carries its own travelling doc — the same
`doc_plans`/`doc_notes` machinery tools already use — and each stage has one question
and one gate that is capable of failing.

The originating evidence: `teaser-reveal-panel` took five rounds, and Round 5 found a
bug present since Round 1 — JavaScript that had never once run client-side. Four
rounds of honest, non-vacuous, served-artefact verification passed straight through
it, because every check read static markup or forced DOM state and none ever fired a
real click. **The gap was a missing stage, not missing rigour.**

## Decisions taken, and why

**D1 — Adopt 027's S0–S7 numbering; do not invent a parallel vocabulary.**
The tools lane independently sketched a four-stage version (skeleton → one real
behaviour → the rest → polish) and then deferred to S0–S7 as a superset. Two lanes
using one numbering is worth more than either lane's preferred names. *Reason:* a
forked vocabulary is the exact drift class the council reviews for.

**D2 — Every gate is validation; none is judgement.**
Forced by the owner's validation-versus-judgement correction to the tools lane. A gate
answers a closed question with a fixed rule and the same answer every time. *"Is this
component any good?"* is judgement and belongs to a reviewer seat. **The ladder has no
aesthetic gate and that is deliberate**, not an omission — conflating the two yields a
gate that drifts per component and a judgement boxed into a checklist.

**D3 — `skip` is not a pass, and this is the ladder's load-bearing correctness
requirement.** The Tier-4 runner skips an unknown check type
(`default: skip(ch.ID, ch.Type+" not implemented")`), and G4 means an all-skipped
result set reads as PASS plus a 7-day cooldown. For a *ladder* that is worse than for
a checklist, because stage N's pass licenses stage N+1. A stage that cannot evaluate
its question is **inconclusive**. *Reason:* discovered live — `has_visible_area` is
committed and not rolled, so the newest and most useful check type is currently the
one that would silently skip.

**D4 — The PLAN is the fleet-wide contract; the NOTES are the per-site verdicts.**
Not a preference — the schema already says so. `doc_notes` has `site_id`, `doc_plans`
does not. A component's template is fleet-shared (one `content_components` row serves
11 sites for `info-card-grid`) so a site-less PLAN is correct; S4–S7 are per-site
facts and land in NOTES.

**D5 — ~~Stage the enabling migration in the RUNBOOK, not in `sql_for_agents/`.~~**
~~The migration runner takes *every* pending file in a directory, so an unreviewed
`272_*.sql` could be swept in by an unrelated session's `--apply`. It gets a number
when it goes to the council gate.~~

> **CORRECTED 2026-07-30, same session — D5 WAS WRONG, and what caught it was reading
> the enforcement points instead of trusting my own read of the schema.**
> Two things I had not looked at made it wrong:
>
> 1. **The migration was never the whole change.** `subject_type` has a **second
>    enforcement point in Go** — `validDocSubjectTypes`
>    (`platform/orchestration/actions/doc_subjects_common.go`), which gates
>    `write_doc_plan`, `append_doc_note`, `load_doc_context` and
>    `persist_diagnosis_note`. Shipping the DDL alone would have reproduced
>    **`bugs_open/064` for the third time**: migration 163 missed the
>    `persist_diagnosis_note` gate, migration 184 moved the DB CHECKs *only* and left
>    its own seeded action docs unreachable through every doc action. The file's own
>    comment states the rule — *a value the DB accepts but a Go gate rejects is a split
>    contract; move both together.*
> 2. **The migration MUST be numbered, because a test parses it.**
>    `TestValidDocSubjectTypes_LockstepWithMigrationCheck` finds the newest **numbered**
>    `.sql` under `sql_for_agents/` that recreates `doc_plans_subject_type_check` and
>    fails if its ARRAY differs from the Go list. So withholding the number does not
>    protect anything — it **reddens HEAD** for every other session the moment the Go
>    edit lands. D5 as written was unbuildable.
>
> **Replaced by D5′: the Go edit and the numbered migration land in ONE commit, and the
> migration is not applied until an image carries the Go half.** The residual risk D5
> was worried about — another session's `--apply` sweeping the file in early — is real
> but *inert here*, because nothing writes component docs yet, so a widened CHECK ahead
> of the image has no effect. Shipped as `273_doc_subjects_component.sql` +
> `doc_subjects_common.go`, commit `c659e312b`, council correlation
> `e5673868-7c5b-489c-931a-7ba59b959b91`. **The lesson is the one this lane exists to
> make mechanical: I costed a change by reading one enforcement point and calling it
> "the smallest possible platform change". There were two.**

**D6 — Do not take ownership of `features_open/015`.**
The tools lane's decomposition (015 = rung vocabulary, 027 = gate mechanism, 026 =
missing instrument) makes the three composable rather than merged, which means this
lane can proceed without owning the site-scale ladder. Recorded as PROPOSED — whether
015 stays a separate thread is the owner's call.

**D7 — Prove a check type in the running binary before authoring a gate against it,
using a LONG marker.** Go compiles short string literals to immediate comparisons that
never reach rodata, so `grep -ac "selector_count"` returns 0 on a binary that fully
supports it. A negative from a short marker is worthless.

**D8 — OWNER RULING 2026-07-31: THE LADDER IS CUT FROM EIGHT GATES TO THREE.**
The owner asked whether the ladder was worth it. The honest answer was *yes, but much
less of it than proposed*, and he accepted that. **S0–S7 remain as a written checklist a
builder reads and ignores where it does not fit** — which is exactly what the leopardess
lane did, correctly, twice. **Machinery gets built for three things only:**

1. **the claim written before the build** (S1) — both lanes found this changes what gets
   built, and it is the cheapest of the three;
2. **verification through the visitor's real gesture** (S6) — the only stage that catches
   a class nothing else does, and it costs 48 seconds to fire;
3. **every check proven able to fail** (the S2 discipline, not the S2 harness) — including
   the sub-rule that a mutant counts only if the artefact provably changed.

*Reason, and it is evidence not taste:* the mutation harness cost the forward-run lane
~40 minutes and **found nothing in their actual product**; S0 was "a five-minute grep that
prevented nothing"; S7 cannot be completed while `bugs_open/157` is open. Decisive:
**two of my eight gates were wrong on first contact with reality, and S4 would have
BLOCKED a correct build.** Eight gates is eight chances to be confidently wrong at
someone else's expense, and `bugs_open/149` is the measured precedent for checkers
multiplying until 22 agents are configured and 2 run anything.

*What this does NOT mean:* the discarded stages are not disproved, they are unfunded.
Any of them may return **with evidence** — a case where its absence cost something real.

**D9 — 2026-08-02: P2's dispatch gap closes with a SIBLING action
(`request_component_browser_run`), not a branch on `request_browser_run`.**
`HANDOFF_2026-08-02_continue_here.md` §3 left this open as a real tradeoff between
(a) an opt-in `page_id_field` branch on the existing tool action and (b) a sibling
action that never touches the tool path. Decided (b), for three reasons:

1. **Blast radius.** `request_browser_run` is the one path every one of the fleet's
   22+ hosted tools' acceptance runs already depends on (D-something above; measured
   6/22 unresolvable by name alone, so the other 16 lean on this exact function
   working unchanged). A branch — even opt-in, even currently dead for every existing
   caller — adds a second resolution strategy to a function this lane has already
   shown is easy to misjudge the size of (D5′: "I costed a change by reading one
   enforcement point... there were two"). A sibling action makes the blast radius
   *provably* zero rather than *argued* zero: nothing about the tool path's source
   changes, so there is nothing there to review for regression.
2. **The two lookups are genuinely different shapes, not a parameterisation of one
   query.** Tool: `pages.name IN (function, 'tool-'||function)` — the page names
   itself after the function. Component: `page_components.component_id` (joined via
   `content_components.function`) `AND page_components.page_id = <given>` — the page
   doesn't name itself after anything; placement is a many-to-many row, and the given
   `page_id` is *asserted*, not derived, then checked against that row. Forcing both
   into one function's control flow would have made the shared function's contract
   harder to state, not easier — this is not the "three identical copies of
   `envelopePaths`" case the codebase already has a rule against; it's two distinct
   resolution predicates that happen to feed the same downstream envelope.
3. **The council's own precedent runs this way.** CLAUDE.md "Platform seams" and the
   `bugs_closed/124` guardian veto both read as: a shared mechanism's *existing*
   guarantee is the thing to protect, and the cheapest way to protect it is not to be
   the commit that has to prove you didn't change it.

**What IS shared, because duplicating it would recreate exactly the drift class this
codebase already paid for once (`envelopePaths`'s own doc comment):** everything from
profile resolution through envelope-build, marshal and produce (existing lines
~158-270) moves into one unexported helper both actions call. Only the page/URL
resolution (lines ~120-152, structurally different per above) and each action's own
`ActionInputSpec` stay separate. This is a behaviour-preserving refactor of the
existing action — `RequestBrowserRunAction`'s output for every existing caller is
unchanged; verified by `go build`/`go vet`/existing tests before commit, not assumed
from the diff being mechanical.

`judge_acceptance_results` needs **no change**: it already keys off
`input_data.spec.function` via config default, and the S6 dispatch sets that to the
component's `function` (== `subject_key`) the same way a tool dispatch does — the gap
was only ever in resolving *which page*, never in judging the results once they
arrive.

Council-gate read (restated from the handoff, now against the actual diff): a new
sibling action that nothing calls until a component-fence dispatch names it does not
change what `request_browser_run` guarantees existing tool callers — normal council
gate, not an RFC, per the 2026-07-29 owner ruling. Register in the concept register
(DOC-068 area) in the same commit.

## Phasing

**P0 — adoption and design (this session).** Standing five created; the blocking
unknown resolved; the proposal updated with what the review found. **Done.**

**P1 — make a component documentable. SUBMITTED 2026-07-30, awaiting verdict.**
Both halves written, tested and committed (`c659e312b`); council correlation
`e5673868-7c5b-489c-931a-7ba59b959b91`; **migration 273 NOT applied** — image first.
Mutation-proven rather than merely green: with the Go half alone the lockstep test
fails naming 184's exact failure mode, and with both it passes.
Remaining in P1, in order: verdict → build/roll an image carrying the Go half →
pod-grep to prove it shipped (a roll is not evidence) → apply 273 → then one real
component (`teaser-reveal-panel`, because its history is fully written down, so nothing
has to be reconstructed) gets a PLAN with a criteria fence and its NOTES backfilled
from `NOTES_brochure_component_library.md`. Gate: the fence exists, passes the ten-rule
validator, and every criterion has been watched to pass by hand.

> **P1 DONE, 2026-08-02.** Verdict/image/pod-grep/273 were already done (see NOTES,
> entry `8f564028-6fc6-488c-96d2-c2e362b243b2`). The remaining item — a real fence for
> `teaser-reveal-panel` — is done too: 12 checks, `try_fence.go` 15/15 against the live
> URL, every check watched to FAIL under its own mutant (not just watched to pass — a
> stronger bar than this line asked for), written into `doc_plans`/`doc_notes` as a real
> row (not the throwaway probe row this PLAN's own D-something worried would be mistaken
> for proof), and read back out of the DB to confirm the write round-trips. Detail:
> NOTES entry "2026-08-02 — P1's tail closed". **The backfill source was actually
> `NOTES_teaser-reveal-panel.md`** (the component's own file), not
> `NOTES_brochure_component_library.md` (the lane-wide file this line named) — the
> component-specific file is the more precise source and itself instructed exactly this
> port once 273 landed.

**P1a — the three-way naming contract check. NEW, and it jumps the queue** (2026-07-30,
on the first forward run's measured recommendation). Assert
`doc_plans.subject_key == pages.name == content_components.function` for every subject
that has a fence, and report the mismatches. **One query.** It is first because a
mismatch makes a fired run *skip and read as clean* — so every other stage's verdict is
untrustworthy until this passes, and it already has a known population of **6 of 22
hosted tools**. It also needs nothing from the blocked migration.

**P2 — make S6 real.** Dispatch a component's fence to `browser-runner-adapter` the
way `tool-acceptance-agent` does for tools. This is the stage whose absence cost five
rounds, and it is wiring rather than construction — the mechanism was proven end to
end on 2026-07-29 by the `smart-contrast` pilot. Gate: a deliberately broken
component makes it go red.

> **P2 DONE, 2026-08-02.** D9 decided sibling action over branch; built as
> `request_component_browser_run` (DOC-072), `go build`/`vet`/`test` clean, committed
> `f6bfb7e6e`, council-submitted `33d00513-2fd8-4872-ad5a-a19c24a1ae0b` (verdict pending —
> read it and act if REVISE/REJECTED). Live on `v1.0.1231`, pod-verified both replicas
> (positive + sanity-negative controls). **Dispatched for real**, correlation
> `e6a258eb-6ba1-44df-b344-16e42443975f`: `teaser-reveal-panel`'s fence ran through
> `browser-runner-adapter` — 15/15 passed, 9 legitimately skipped (mobile gating, not
> "not implemented"). Gate's own wording ("a deliberately broken component makes it go
> red") is satisfied by the NEGATIVE CONTROL on page resolution, not a fresh live mutation
> run: a real, active, wrong page on the same site correctly produced
> `component "teaser-reveal-panel" is not placed on page ... (or that page is inactive)` —
> proving the new placement JOIN fails closed. Re-running the checks' own mutation proof
> through the cluster was judged redundant with TL-036's already-complete offline proof
> (same `RunChecksAction.Execute`, 12/12 mutants caught) — see NOTES for the full reasoning.
> Full detail: NOTES "2026-08-02 — P2 CLOSED"; `SUMMARY_2026-08-02b_the_dispatch_gap_is_closed.md`.
>
> **COUNCIL VERDICT 2026-08-03: APPROVED**, 11 reviewers, 3 advisory objections (2 medium, 6
> low), none blocking, `gated_by_truncation: false`. Two medium objections checked against the
> actual code rather than argued: `guardian`'s "no test covers the extraction" is **factually
> wrong** — `TestRequestBrowserRunPayloadCarriesCaptureRenders` +
> `TestRequestBrowserRunCaptureRendersDefaultsOff` already exercise
> `RequestBrowserRunAction → dispatchBrowserRun → Producer` end to end and both passed.
> **`prior_art_librarian`'s is factually RIGHT and a real lesson**: `request_browser_run`
> already had a `url_field` override that bypasses the `pages.name` lookup entirely, so a
> smaller design existed — a resolver action doing only the placement JOIN, feeding
> `request_browser_run` unchanged via `url_field`, needing no `dispatchBrowserRun` extraction
> at all. Not reverted (shipped, tested, proven live; reverting working code for a marginally
> smaller equivalent is a real cost the owner hasn't asked for) but recorded plainly as a
> "smallest possible platform change" miss, same class as D5′. Full read-out, all 9 objections
> individually verified: NOTES "2026-08-03 — council verdict read".

**P3 — the remaining gates**, cheapest first, each with its mutation.

> **CORRECTED 2026-08-02 — P3 as written is STALE, and re-checking it before building
> anything from it is what caught this.** This Phasing section was written 2026-07-30. **D8
> landed the next day, 2026-07-31**, and cut the ladder from eight gate-types to three funded
> ones (S1 claim-before-build, S6 visitor-gesture, S2 mutation discipline) — explicitly
> retiring "build the remaining gates" as a plan: *"the discarded stages are not disproved,
> they are unfunded. Any of them may return with evidence — a case where its absence cost
> something real."* P3's own wording ("the remaining gates, cheapest first") is exactly the
> approach D8 overruled the day after this line was written, and nobody reconciled the two
> until now — the same class of drift this whole file exists to prevent (D5 got corrected the
> same way, same session it was written).
>
> **No new evidence has surfaced since D8 that would fund S0/S3/S4/S5/S7 for components.**
> So P3 does not get built as stated. **What P3 actually collapses into: applying the THREE
> already-funded gate-types to more subjects** — more components/tools getting an authored,
> mutation-proven fence (S1+S2 discipline) and a dispatched S6 run, the same shape P1+P2 just
> proved once for `teaser-reveal-panel`. That is the honest next step, and it is the authoring
> backlog named in `HANDOFF_2026-08-02_continue_here.md` §4 item 3 — not a new phase, the same
> one under its real name. Re-labelled below.

**P3 (re-labelled 2026-08-02) — the authoring backlog: measure it, then close it one subject
at a time with S1+S2+S6, cheapest/most-evidenced first.** Concretely: count components/tools
with no PLAN at all (the honest backlog, not a defect), then for each, author a fence backed
by real behaviour (read the component before writing checks, per this session's own practice),
prove it mutation-hard (own harness per component if `prove_fence_can_fail.go`'s hardcoded
mutant list doesn't fit — check before assuming reuse, twice burned already in this lane), and
dispatch S6 for real. **Do not wire a PLAN-required check into any birth path as a side
effect** — this is backlog clearance, not a new gate, per the same reasoning that kept
`CHECK_naming_contract.sh` out of the birth path after P1a closed.

**P4 — only then** stage-scoped dynamic generation of gates, and the anti-vacuous
verdict rule (D3), which changes a shared guarantee and therefore goes to the gate and
plausibly an RFC.

## What would make this lane wrong

Stated up front so it is falsifiable rather than defended later:

- ~~**If nothing fires the stages, the ladder is worthless.** This is the tools lane's
  G5, and it is the most likely way this fails.~~
  > **ANSWERED BY MEASUREMENT 2026-07-30, and the answer reframes the risk rather than
  > clearing it** (`REPLY_2026-07-30_vendor_trust_checklist_build.md` Q3). Firing by hand
  > is **cheap**: S6 end to end was **one script, 48 seconds** wall-clock, correlation
  > `dc952633`. So the trigger was never the binding constraint. **Addressability is.**
  >
  > Three values must be equal or a fired run quietly does nothing —
  > `doc_plans.subject_key == pages.name == content_components.function`. `load_docs`
  > keys on `spec.function`; a mismatch yields an empty fence and `request_browser_run`
  > **SKIPS with `needs_criteria`**: honest, but not a failure either, so **it reads as a
  > clean run that asserted nothing.** Measured fleet-wide: **6 of 22 hosted tools cannot
  > be acceptance-tested at all** until renamed, across five sites — with an honest
  > denominator note that including the three non-tool riders would read 9 of 25 and
  > "flatters the problem".
  >
  > Their conclusion, adopted: *"a ladder whose stages CAN be fired but silently resolve
  > to nothing is worse than one nobody fires, because it produces green."* **So the
  > highest-value single thing to build is not a trigger — it is the check that asserts
  > the three-way naming contract.** One query, and it would have found six broken tools
  > before anyone fired anything. It is now P1's first item.
- **If gates proliferate into dead config, it is a net negative.** `bugs_open/149` is
  the measured precedent: 22 discovery handler agents, only 2 running
  `validate_page_content`, six registered checks in no agent and zero items ever.
- **If the claim that stages would have caught Round 5 is wrong**, the whole argument
  weakens. It is marked `[INFERRED]` in the proposal and stays marked until a
  deliberately broken component is caught by an S6 gate nobody tuned for it.

**D10 — OWNER RULING 2026-08-05: the backlog is cleared EXHAUSTIVELY (option c).**
Posed with the calibration tranche's measured costs (~15 min/static section, ~30–45
min/interactive tool); the owner chose full clearance over the focused-push
recommendation. Scope as measured, with two standing exclusions:
1. **A tool with no resolvable, serving page gets NO fence** — a fence without a page is
   `CHECK_naming_contract.sh`'s BROKEN A class (the run hard-errors; a claimed subject
   that can never be asserted). Those tools are LISTED for page repair instead, and get
   fences when their pages serve.
2. **`component_level` beyond `section`/`tool` stays out of scope** (header/footer/site/
   head/element — DOC-068's boundary, unchanged; extending it is a separate owner call).
Order: sections by placement count (static first), then remaining ready tools. Every
subject keeps the full recipe — no shortcuts licensed by volume; the per-subject mutation
prover moves from hand-written siblings to a DATA-DRIVEN mutant file per subject (same
architecture, mutants explicit in JSON, validated by reproducing call-to-action's 6/6
before first use) so the sibling-file boilerplate stops scaling with the backlog.

**D11 — OWNER RULING 2026-08-09: the defects the contract work surfaces are ADDRESSED,
from WITHIN the framework, and PREVENTED in initial builds.** The standing defect list
stops being a report and becomes a work programme. Three binding clauses, in the owner's
own framing: (1) fix the CHECKERS and HANDLERS — the repair goes through the framework's
own detection→repair machinery, never a hand-edit to an artefact (the 2026-08-04
every-site-through-the-framework ruling extends naturally to repairs); (2) where
detection exists but repair never dispatches, the dispatch gap IS the defect to fix;
(3) prevention lands in the initial build path, so the same classes cannot be born again
— a build must not report success with schema-required content empty or brand assets
that do not serve. Process unchanged: cross-cutting causes through the 090 loop before
assertion, platform code through the council gate, `who-owns.py` before touching another
lane's filed bug (155/168/201/149 all have prior claims). Routed work programme:
`HANDOFF_2026-08-09_continue_here.md` §6 (appended the same day).
