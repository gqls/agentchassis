# The architecture-review track (RFC process)

**Created 2026-07-25 by owner decision**, after the council gate's guardian
vetoed the bug-003 delivery-guarantee redesign twice — not on technical grounds
(those it conceded in round 2) but because "a coordinated rewrite of the
fleet's delivery guarantee … across five-plus packages" is, by the gate's own
charter, an architecture change dressed as a fix, and the platform had no
review track sized for one. This document is that track. It is deliberately
lightweight: one file per RFC, the owner as the deciding authority, the
existing council gate reused for what it is good at.

## When a change needs an RFC (the trigger test)

Taken from the guardian's charter language, which two verdicts have now
exercised. A change needs an RFC when ANY of these hold:

- it changes a **shared contract**: a dedupe key, a delivery guarantee, a
  state machine consumed by more than one package, a wire/message shape;
- it changes or removes an **exported symbol** other packages depend on
  (signature changes count);
- it lands coordinated edits across **many packages at once** (rule of thumb:
  three or more of `platform/*`'s top-level packages);
- it needs a **staged or reversible rollout** to be safe (the change and its
  verification cannot both fit in one deploy step).

A point fix that happens to be large does not need an RFC; a small change
that rewrites a contract does. When in doubt, the cost of an RFC is one
document — write it.

## The bar is asymmetric, and that is deliberate (D6, named 2026-07-27)

The trigger test above says when a change *needs* an RFC. It does not say what
an RFC must *prove*, and leaving that unstated is how "we should redesign this"
becomes arguable on confidence alone. So, named explicitly:

**Keeping battle-tested code is the default and needs no evidence.** Code that
has been in production and has not failed carries an argument by survival. The
guardian's stability preference is that argument, and it does not need
restating per submission.

**Replacing it must clear four bars.** All four, not a majority:

1. **A defect the current design cannot express a fix for.** Not a defect that
   is awkward to fix, or that has recurred — a defect where the contained fix
   does not exist. Recurrence alone is *not* sufficient: a bug found at four
   call sites may be four point fixes. The question is whether a fifth site can
   still be created after the fix lands.
2. **Blast radius derived mechanically** (§4 below) — `go list -deps`,
   compile-proof for removals. A qualitative "this is fairly contained" fails
   this bar however true it turns out to be.
3. **Independently-valuable stages.** Each stage must be worth shipping if the
   later ones never ship, and must build and test alone against the tree as it
   exists before it. A plan whose stages must land together is one change
   wearing a staging costume.
4. **A rollback needing no migration.** The previous binary must tolerate the
   new schema. If it cannot, say so loudly and expect that to be the objection
   the owner rules on.

**Why asymmetric rather than "weigh the evidence both ways":** a symmetric bar
sounds fairer and is not, because the two sides have unequal instruments. The
risk of change is measurable before the fact (blast radius, call sites, stages)
while the benefit is mostly forecast. Weighing a measurement against a forecast
as though they were the same kind of quantity reliably favours whichever side
the author already preferred. Making the bar explicitly asymmetric puts the
burden where the uncertainty is.

`RFC_001` already meets bars 1, 2, 3 and 4 — this section does not change it,
it names the standard it was already written to. The bar is stated here so a
*future* RFC cannot be argued through on the strength of its prose.

**The counter-argument, kept visible:** a bar this explicit can be satisfied
formally by a plan that is still wrong, and it says nothing about a change that
is *insufficient* — a point fix that leaves the mechanism exploitable clears
every bar above by never being an RFC at all. That gap is the `review_architecture`
seat's job (`insufficient` is one of its three signals), not this document's.

## What an RFC contains

One markdown file, `RFC_NNN_<slug>.md`, in this directory. Status line at the
top: `DRAFT` → `RATIFIED` (owner) → `IMPLEMENTED` → (possibly) `RETIRED`.
Sections, all required:

1. **Problem + evidence** — the defect or need, with live figures and
   file:line citations. Point at the bug file; don't fork it.
2. **Design** — what changes, per package, with the load-bearing mechanics
   (the two or three queries/orderings everything rests on).
3. **Alternatives considered** — including the do-less option, each with the
   evidence that ruled it out. An alternative dismissed without evidence is
   an objection waiting to happen.
4. **Blast radius, named** — derived mechanically (`go list -deps` per cmd
   target, compile-proof for removals), not qualitatively. Name the binaries
   whose behaviour changes, and the ones that merely relink.
5. **Staged rollout plan** — the order things ship, what is watched at each
   stage, and the induced-fault tests (not just happy-path greps). Name the
   canary if there is one.
6. **Rollback plan** — what undoes each stage; schema must tolerate the
   previous binary (image-first rollback), or say loudly why not.
7. **Acceptance evidence** — the measurements that will retire the RFC's
   risk: pod-greps of created literals, behavioural probes, week-later stats.

## The flow

1. Write the RFC (status DRAFT). Commit it.
2. **Owner ratifies** — the owner is the architecture authority; a ratified
   RFC records the decision the way PLAN decisions (D1, D2…) already do.
   Objections and revisions happen in the file, visibly, like any working doc.
3. Optionally, exercise the **council gate** on the RFC-shaped plan (097; the
   `plan.edits` name the real files, the rationale points at the RFC). The
   gate remains advisory; its seats are good at catching what one author
   missed. A guardian veto on a ratified RFC is input for the owner, not a
   block.
4. Implement **in the RFC's stages**. Each stage that fits the point-fix
   shape goes through the normal council gate as usual.
5. When acceptance evidence is in, mark the RFC IMPLEMENTED, with the
   evidence inline or pointed at.

## Numbering and index

- `RFC_001_at_least_once_delivery.md` — bug-003 delivery-guarantee redesign
  (the case that created this track).
- `RFC_002_criteria_check_type_vocabulary.md` — who may add a check type to the
  shared Tier 2 criteria vocabulary, and on what terms. **RETROSPECTIVE**: the
  change is already live (v1.0.1197), routed here on the owner's instruction
  after the `review_architecture` seat ruled it architecture-scope while the
  guardian declined to veto in the same round. The first RFC on this track whose
  subject is a precedent rather than a build.

- `RFC_003_fleetwide_banned_claims_at_the_build_gate.md` — the build gate may
  refuse a page on a site that never opted in.
- `RFC_004_a_deploy_action_that_can_refuse.md` — a deploy action that can refuse.
- `RFC_005_targeted_review_for_docs_that_feed_the_fleet.md` — targeted review for
  the two doc classes that feed the fleet back, plus a staleness sweep for
  `bugs_open/`.
- `RFC_006_one_promoter_one_owner_for_shared_promoting_steps.md` — a shared step
  whose effect is "take everything in state X" needs one owner.
- `RFC_007_chrome_eligibility_needs_a_package_both_sides_can_import.md` — the
  guard-scan count is the meter.
- `RFC_008_a_mandatory_write_seam_for_page_components_rendered_html.md` — why an
  advisory lint is the wrong ceiling.
- **`RFC_009` — AMBIGUOUS, two unrelated papers. Resolve by slug, never by number:**
  - `RFC_009_one_derivation_for_a_deployed_assets_path.md` (`bugfix_168`, filed 11:50)
  - `RFC_009_the_renderer_should_enforce_input_schema_on_missing.md` (`bugfix_140`, filed 23:39)
- **`RFC_010` — AMBIGUOUS, two unrelated papers. Resolve by slug, never by number:**
  - `RFC_010_discovery_checks_can_raise_a_finding_but_not_retract_one.md` (`bugfix_168`, filed 19:59)
  - `RFC_010_who_may_answer_a_page_name_collision.md` (`bugfix_175`, filed 23:56) —
    **this is the one CLAUDE.md's OWNER RULING 2026-08-02 cites**, and it is RATIFIED.

> ## THIS LEDGER STOPPED BEING MAINTAINED AT `RFC_002`, AND ON 2026-08-02 IT COLLIDED TWICE IN ONE DAY
>
> Every paper from `RFC_003` to `RFC_010` was filed **without** claiming its number here —
> including both of mine (`bugfix_168`). The rule below was not disputed or overridden; it was
> simply never executed by anybody, for eight consecutive papers, so nothing at all stood
> between two lanes and the same integer. **Restored 2026-08-02 by reading the directory**, so
> the numbers above are observed, not remembered.
>
> The two collisions are left in place rather than renumbered, matching the estate's existing
> convention for `/bugs_*/` numbers — *never reassigned, several numbers name two unrelated
> cases, so resolve by slug*. Renumbering would silently break citations in a RATIFIED paper,
> in `CLAUDE.md`, and in a closed bug file, which is a worse failure than an ambiguous integer
> that is documented as ambiguous.
>
> **So: `grep` this list before claiming a number, and add your line in the same commit as the
> paper.** A ledger nobody appends to is not a ledger — it is a comment, and this is what a
> comment enforces.
>
> **AND IT WENT UNMAINTAINED AGAIN IMMEDIATELY — eight more papers, `RFC_011` to `RFC_018`,
> filed between 2026-08-02 and 2026-08-08 without one of them claiming its number here.**
> Restored below on 2026-08-08 by the same method (`ls` the directory), by the session filing
> `RFC_019`. That is now **sixteen of nineteen papers** self-numbered against a stale line
> reading "the next free number is `RFC_011`" — which is worse than the first lapse, because
> this time the restoration and the rule were both *already in the file*, three paragraphs
> above the wrong number. **The lesson is not "try harder": a ledger that only tells the truth
> when every author remembers to write to it will keep reading `RFC_011` for ever.** The one
> thing that has actually worked, twice, is deriving it from the directory — so **derive it,
> then write what you derived**, and treat the line below as a hint to be re-checked, never as
> the answer:
> ```bash
> ls docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_*.md | sed 's/.*RFC_0*\([0-9]*\)_.*/\1/' | sort -n | uniq | tail -3
> ```
> The obvious fix is to make the check mechanical (a `pattern-check.py` rule, or a
> `PROCESS`-vs-directory drift check in the same shape as `landmines-sync.py --check`).
> Nobody has built it, and until somebody does, the collisions are the expected outcome, not
> the surprise.

- `RFC_011` — **AMBIGUOUS, two unrelated papers** (observed 2026-08-08, both filed 2026-08-02+
  without claiming the number). Resolve by slug:
  - `RFC_011_a_fleet_wide_execution_deadline_on_the_step_seam.md`
  - `RFC_011_git_adapter_action_vocabulary_and_the_unpublish_verb.md`
- `RFC_012_the_await_overwrite_destroys_action_findings.md` — the await overwrite destroys action
  findings; three owner rulings, all delivered (lane: `rfc012_await_findings/`).
- `RFC_013_per_category_provocations_and_a_contract_no_compiler_can_see.md`
- `RFC_014_handleragent_is_a_stringly_typed_routing_contract.md`
- `RFC_015_decision_records_allow_change_forbid_regression.md`
- `RFC_016_section_entry_wire_shape_and_plan_time_fact_assignment.md`
- `RFC_017_verifier_registry_fails_open_on_error.md`
- `RFC_018_reaper_accounting_as_a_shared_mechanism.md`
- `RFC_019_one_ladder_for_which_agent_is_running.md` — one ladder for "which agent is running"
  (`ExecutionContext.ResolvedAgentType`), and where its bottom rungs live.
  **RETROSPECTIVE**, like `RFC_002`: the code shipped in `1bc08d1ce` with register entry
  `RSH-009`, per OWNER RULING 2026-07-29 §2 (review here is after the fact by design; a thread
  cannot hold a change out of a shared HEAD). It asks the owner to draw a line the author is the
  interested party in: whether an exported method on `types.ExecutionContext` is architecture-scope
  as such, given `RSH-008`'s one-week-old `point_fix` precedent was licensed by "stays inside
  `platform/orchestration/actions`" and this does not.

Claim the next number by adding a line here in the same commit as the RFC —
the same collision discipline as migrations, and this list is the ledger.
**The next free number is `RFC_020`** — derived from the directory on 2026-08-08, not carried
forward. Re-derive it before you trust it; twice now this line has been wrong by eight.
