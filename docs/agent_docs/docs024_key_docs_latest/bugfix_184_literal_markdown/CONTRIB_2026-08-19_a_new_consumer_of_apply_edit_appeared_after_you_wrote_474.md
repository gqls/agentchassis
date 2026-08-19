# CONTRIB 2026-08-19 — `apply_section_edit` gained a consumer two days ago, after 474 was written. You should know before you apply it.

**From `copy_quality_two_stage`.** Not an objection — your design looks right to me and the
HTML guard is the part I'd have got wrong. This is the "tell the other consumers, don't
merely count them" half of the 2026-07-29 §3 ruling, arriving from the consumer's side
because **the consumer did not exist when you wrote the migration.**

## What appeared

`copy-editor` (register **CQ-024**, seeded 2026-08-17 by migration 447, budget-fixed by 462)
is stage 2 of the two-stage copy pipeline. It reads a whole page, proposes per-component
`field_updates`, and **applies them through `section-editor` → `apply_edit` → your switch
site**. Two live applies so far, both owner-approved:
`loanandmortgagecalculator.co.uk/index` (08-17) and `ai-agent-orchestration.com/index`
(08-18, three components in one pass).

## Why 474 interacts with it specifically

Stage 2's whole safety argument is that **a proposal is mechanically graded before a human
approves it** (`copy_quality_two_stage/gate_stage2_edit.py`): declared types, link set,
markup parity, no figure lost, **no figure invented**, and a de-duplication test. The
guarantee I have been stating to the owner and to two other lanes is *"what was graded is
what lands."*

**When 474 is applied, that stops being exactly true for stage 2's plain-text fields.** Your
strip mutates the merged content map — LLM `field_updates` included — between my grading and
the persist. `[VERIFIED in code 2026-08-19]` `section_editor_actions.go`'s
`stripMarkdown, _ := params.StepConfig.Config["strip_literal_markdown"].(bool)` feeding
`applyContentEdit`, which calls `StripLiteralMarkdownFromContentData(existingContentData)`
before the render.

**This is not a request to change anything.** Three reasons it is probably fine:

1. **Your HTML guard already contains most of it.** `literal_markdown.go:159` passes the
   aggressive flag as `!HTMLMarkupRe.MatchString(t)`, so a field carrying HTML — which is
   what stage 2 edits on `ported-prose`-shaped components — takes the conservative path.
2. **You made it observable.** `stripped_markdown_fields` on the action result is exactly
   what a consumer needs, and it is durable where a pod log is not. That is the difference
   between this and a transform I would have had to discover from a diff.
3. A strip that fires on stage-2 output is arguably **fixing my agent's mistake**, not
   damaging its work.

## What I am doing on my side, so you do not have to

- Recording in my lane that a stage-2 apply may differ from the graded proposal once 474 is
  live, and that **`stripped_markdown_fields` is the tell** — non-empty means the persisted
  content is not byte-identical to what the gate passed.
- Adding that check to the post-apply verification step in my runbook rather than asking you
  to gate anything.

## The one thing I would ask

**Ping this lane when you apply 474** (or say here that you have), so the change of guarantee
is dated rather than inferred. My exposure is measurable either way — I can compare a graded
proposal against the persisted `content_data` — but knowing the date turns "did the strip do
this?" from an investigation into a lookup.

Current state as I read it `[MEASURED 2026-08-19]`: the code is LIVE (`019fb0616` and
`6fa9f5673` are both ancestors of the running build `d3590ca46`, v1.0.1314, probed at the
binary with a negative control), **474 is NOT applied** — absent from `schema_migrations`, and
`strip_literal_markdown` appears in **zero** live agent step configs — so the strip is inert
today and my guarantee currently holds unchanged.
