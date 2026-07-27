# 121 — The house voice was duplicated across two substrates, and the architecture seat that exists to catch exactly this has never reviewed anything

**Filed** 2026-07-27 by the `gemini_content_provider` workstream ·
**Status** **FIXED for content-creator; the second half is INERT until the chassis
roll — so OPEN** · **Introduced by** `d39995125` (my own change, same day) ·
**Supersedes** `features_open/026`, which filed this as a feature request; the owner
ruled it is a bug, and the reason is below

---

## The defect

Applying the owner's "house voice is the default for all content" directive, I put
the rules in a Go `const` in `internal/agents/contentcreator/agent.go` while an
identical copy already lived inside `page-content-writer`'s `prompt_template` in the
database.

Two hand-maintained copies of one contract, in two substrates, **with different
change latencies**: the DB row changes instantly, the Go const needs a build and a
roll. That asymmetry is what makes it a defect rather than untidiness — the two
cannot even be changed together, so any edit guarantees a window where site copy and
blog copy are written to different rules. The drift leaves no error and no failing
status. It surfaces only when a human reads both and notices they no longer sound
like the same company.

This is the class CLAUDE.md already names: *"Two hand-maintained rosters that must
stay identical is exactly the drift class this council reviews for"* — solved there
by a mirror script, not by discipline.

## Why this is filed as an architecture-review miss

**The architecture seat exists to catch precisely this shape and has never reviewed
anything.** Measured 2026-07-27:

```sql
SELECT count(*) FILTER (WHERE body ILIKE '%review_architecture%')            AS mentions,
       count(*) FILTER (WHERE body ILIKE '%"reviewer":"architecture%')       AS actual_reviews
FROM diagnosis_artifacts WHERE kind='council_report';
--  mentions | actual_reviews
--         0 | 0
```

Zero and zero, across every council report ever written. The seat is seeded and
live (`architecture_review/HANDOFF_2026-07-27_continue_here.md` records it as
"FULLY LIVE 07-27, NOTHING OWED"), but its own handoff also records why it is
silent: **it is rate-limited on owner-approved specs, and both of the specs it
could review are owned by other threads.** So the seat has a footprint, a prompt,
and a roster slot, and has never once fired.

The consequence is concrete. My duplication went through the council on
`a1a5cf20` — ten reviewers, `unreadable: 0` — and none of them was the seat whose
remit is "two things that must stay identical". The **reuse** seat is the closest
match and it looks for *duplicated code*, not for *one contract stored twice in
different substrates*. So the gate did its job and still could not have caught this.

**A seat that has never fired is not coverage.** It reads as coverage on the roster,
in the mirror script, and in the 16-of-16 seat count — which is worse than an
absent seat, because it is counted.

## Missteps, mine, in order

1. **I created the duplicate while writing the commit message that warned about
   duplication.** `d39995125` says *"I CREATED A SECOND COPY OF THE RULES AND FILED
   IT RATHER THAN HIDING IT"* — I saw the defect, described it accurately, and
   shipped it anyway with a `[KNOWN DUPLICATION]` comment. **A comment naming a
   defect is not a mitigation.** The honest options were: don't ship the second
   copy, or fix the source. I took a third one that felt responsible and wasn't.
2. **I filed it as a feature, which downgraded it.** `/features_open/` is for
   things we want to build; this is a defect with a wrong output (two voices on one
   estate). Filing it as a feature put it in the queue nobody treats as urgent. The
   owner corrected this.
3. **I invented an override mechanism the owner had not asked for.** I built
   `core_logic.voice_style_block` — a config switch with a present-but-empty
   opt-out — and wrote a paragraph justifying the empty-vs-absent distinction. The
   owner meant something simpler and different: *"a request has its own prompt in
   the request — not that we have a switch."* I designed against an imagined
   requirement, then defended the design. **Cheap check: when a directive contains
   a word like "override", ask what it means before building the mechanism.**
4. **I put the prompt in Go.** The owner's correction — *"probably not in go by
   choice"* — names a property I should have reasoned to unaided: prose that a
   non-engineer may want to edit does not belong somewhere that needs a compile
   and a fleet roll to change.
5. **The example sentence I wrote into the rule was vague.** *"getting them to
   recover from errors"* does not say what is actually being done. Corrected to
   *"building them to recover"*, and the underlying point promoted to its own rule:
   *"Name the action, not a vague gesture at it."* Worth noting the rule's own
   example is the strongest signal in a style prompt, because it is the thing the
   model imitates most directly.

## Fix

**One row, read by everything, none of it in Go.**

- **Migration 240 (APPLIED)** seeds the canonical block into
  `agent_default_configs` (`config_name='voice_style_block'`, 2,499 chars). It
  guards its own content: refuses if the block is implausibly short, or contains an
  em dash, which would be the rule teaching the opposite of itself.
- **`platform/voicestyle`** holds the reader and a 60s cache — *no text*. It takes
  a fetcher closure because the chassis holds `*sql.DB` and content-creator holds a
  `pgxpool.Pool`; rather than force an adapter on one, each supplies three lines.
  An unavailable block degrades to "no house voice", never to a failed generation.
- **The chassis** injects the block as `{{.voice_style}}` into the template data of
  **every** `execute_llm_prompt`, so any prompt opts in by naming it. It does not
  overwrite a `voice_style` a step already supplied — that is the request-level
  override the owner actually meant.
- **content-creator** reads the same row; the Go const is deleted.
- **Migration 241 (WRITTEN, NOT APPLIED)** swaps page-content-writer's literal for
  the placeholder. **Gated on a pod-grep**, and the gate is not ceremonial: the
  renderer is `missingkey=zero`, so applying it against a chassis without the
  injection renders the placeholder as **nothing**, silently deletes the house voice
  from every page build, and the only symptom is that the writing gets worse.

## Why it stays OPEN

Until 241 is applied, `page-content-writer` still holds a literal copy — so the
duplication is halved, not closed. It cannot be applied until the chassis roll, and
`bugs_open/112` (spawned pods get no `GEMINI_API_KEY`) is waiting on the same roll.

## The commit hook caught what the seat did not

Committing the fix printed:

```
── architecture signal ──
   • migration + platform code in one commit — needs a staged rollout order
   ↳ this meets the RFC trigger test. If it is a point fix, carry on.
     If it changes a shared contract, write an RFC first
```

It is right, and I should answer it rather than wave it through: injecting
`{{.voice_style}}` into the template data of **every** `execute_llm_prompt` **does**
change a shared contract. Every prompt in the platform gains a name it did not have,
and any template that already used `voice_style` for something else would now be
shadowed (none does — checked). The staged rollout it warns about is exactly the
240-now / 241-after-the-roll split, which is why 241 is written and unapplied.

**Note what this says about the seat.** A ~20-line advisory shell hook, running
locally in under a second, produced the correct architectural observation about this
change. The 16-seat council did not, because the seat that would have has never
fired. **Deterministic triggers caught what a rate-limited LLM seat could not** — and
the hook fired on *every* commit, needing no owner-approved spec and no roster slot.
Worth weighing before adding more seats.

## For the architecture-seat thread

This is a live worked example for that workstream, not a complaint. The seat's own
handoff says it is rate-limited on owner-approved specs owned by other threads.
**This bug is a spec it could review that no other thread owns** — a genuine
two-substrate contract duplication, caught by a human after a ten-seat council
passed it. If the seat needs a first case, this is one.
