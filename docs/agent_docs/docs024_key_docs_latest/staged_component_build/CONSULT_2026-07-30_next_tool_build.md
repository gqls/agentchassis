# Consultation brief — running the next tool build through the stage ladder

**For the lane building the next tool** (session `6eaa3e23-ffe3-4b7a-9957-121b43c87c54`),
at the owner's request: *"I have asked lane … to consult with you on its next tool build
that we can use to follow through with this plan."*
**From:** `staged_component_build` (`features_open/027`), 2026-07-30.

**Read this first, then talk to me.** It is deliberately short on theory and long on what
I need from you and what you get back. The full argument is in
`PROPOSAL_2026-07-30_step_by_step_build_with_stage_gates.md`; the decisions are `PLAN`
D1–D7; the commands are the `RUNBOOK`.

---

## Why your build is the right first case, and what I am NOT asking for

Everything this lane has learned so far was learned **backwards**, from a carousel that
took five rounds and whose fifth round found a defect present since the first. That is a
post-mortem. **Your build is the first chance to use the ladder forwards**, which is the
only way to find out whether it actually helps or just adds ceremony.

So the ask is *not* "adopt eight stages and file paperwork". It is:

> **Author the claim before you build, and let each stage's check exist before the thing
> it checks.** Then tell me where that was useful and where it was friction.

If a stage feels like ceremony on your build, **say so and skip it** — that is a finding
about the ladder, not a failure by you, and it is more valuable to me than compliance.
The ladder has never been run prospectively by anyone, including me.

---

## What I need from you (small, and mostly things you would produce anyway)

1. **The tool's claim, in one sentence, before any code.** Not "a contrast checker" but
   *"given two hex colours it reports their WCAG ratio, and #767676 on #ffffff is
   4.54:1"*. A claim with a known answer in it is worth ten without.
2. **Its entry point, named as a visitor gesture.** "Type in the field and the number
   updates", "paste an image and a sticker appears". **Not** a function name.
3. **Which of your checks you have watched go red.** Any of them. One is fine.
4. **Any point where the ladder had no vocabulary for what you needed to assert.** These
   are the most useful thing you can give me — they are missing check types, and the rule
   is to record them as deferrals rather than substitute something weaker.

That is it. Send them however is convenient.

---

## What you get back

- **The stage ladder mapped onto a tool** (below), so you are not deriving it.
- **The traps that have already cost this fleet real time**, each with the check that
  catches it — the section after next. Three of them cost me hours today.
- **A read on your criteria fence** before you commit it, if you want one.

---

## The ladder, mapped onto a tool build

The tools lane sketched a four-stage version (skeleton → one real behaviour → the rest →
polish) and then deferred to this numbering so the two lanes do not fork a vocabulary.
Its four map onto S1/S2/S6/S7.

| stage | the question | for a tool, concretely |
|---|---|---|
| **S0 shape** | does this shape already exist? | check `experience_patterns` and the 63 existing tools before inventing. A tool that is a near-clone of a live one should say so |
| **S1 contract** | is the claim stated, and are the hazards answered? | the claim sentence + the ```criteria fence **drafted before the build**, with a known-answer pair in it |
| **S2 template** | does it render, and are the checks real? | harness green **and ≥1 mutant red per assertion class**. This is the gate nothing in the tools chain currently has, and it is the cheapest one to adopt |
| **S3 register** | is it reachable? | the row exists, the page resolves, and JS (if any) is in the live bundle — not merely published |
| **S4 place** | is it durable? | it survives one re-render. For a tool page this is the `rebuild_policy='owned'` question |
| **S5 serve** | does the visitor get it? | fetched page, 0 unrendered `{{`, and **slice the `<style>` block away before counting elements** |
| **S6 operate** | does it *work* when driven? | **real Chromium, real gestures, desktop + mobile.** `interaction` + `text_matches` via `browser-runner-adapter`. Proven end to end on 2026-07-29 by the `smart-contrast` pilot: 11/11 checks asserting real arithmetic |
| **S7 regress** | does it still work? | re-run S5+S6 after any roll or rebundle |

**S6 is the one that matters most and costs least**, because the machinery already exists
and has been proven — it has simply never been pointed at anything but tools by the lane
that built it. If you do one stage properly, do this one.

---

## The traps, with the check for each. These are not hypothetical

**1. An unknown check type is SKIPPED, not failed — and an all-skipped fence PASSES.**
The Tier-4 runner ends its type switch with
`default: skip(ch.ID, ch.Type+" not implemented")`, and the judge's `len(Failed)==0`
then writes a **PASS note plus a 7-day cooldown**. So a fence written against a check
type the running binary lacks records a green verdict on a tool nobody asserted anything
about, and suppresses the re-check for a week.
**Live right now:** `has_visible_area` (TL-034) — the newest and most useful check type,
which tells "on the page" from "big enough to see and click" — was committed 2026-07-30
15:19 and is **not in the running `browser-runner-adapter` pod**.
**The check:** before authoring a fence with any recently-added type, prove it is in the
*running* pod. Not the repo, not the image tag.

**2. Prove it with a LONG marker, or your negative is worthless.**
Go compiles short string literals to immediate comparisons that never reach rodata.
Measured today: `grep -ac "selector_count"` returns **0** on a binary that fully supports
`selector_count`. Also: these images have **no `strings` binary** — use `grep -ac` on the
binary directly, and always pair the new marker with a long pre-existing **control** in
the same exec, so a broken grep cannot read as a clean absence.
**And if your change adds no unique string, grep cannot answer at all** — date the build
instead (`stat -c %Y` on the binary vs the commit's `%ct`). I hit this today: all three
markers I picked were wrong, two because they were **comments**, which never reach a
binary, and one (`component`) because it matched **761** times.

**3. Verify through the visitor's gesture, never the tool's internal functions.**
The tools lane verified `pasteboard` by calling `addItem()` and `logic-architect` by
calling `loadTemplate()`. Both returned the right answer, so both "passed" — while a
visitor could reach neither, because the work areas measured **1146×0**. I did the same
thing in a different costume: forced `.open = true` on DOM nodes and called it verified
for four rounds. **These are one defect class — verifying through a privileged path the
visitor does not have — arrived at independently by two lanes on the same day.**
If the vocabulary cannot express the gesture, that is a **missing check type to record as
a deferral**, not a licence to substitute a function call.

**4. A fence must assert the TERMINAL value, not the first observable state change.**
*"Status reads LIVE EDITING"* is a waypoint; *"text can be edited and emphasised"* is the
point. A fence asserting the waypoint passed while `micro-cms` was unusable.

**5. Never author a criterion you have not watched pass by hand.** This is why the
`smart-contrast` pilot passed on its first complete run. And **never trust a check you
have not watched fail** — my own pre-apply gate reported a false pass on its first run
today, because `grep -c` exits 1 on zero matches and my fallback produced `"0\n0"`, which
failed a string comparison and skipped the refusing branch. A gate whose only untested
branch is the one that refuses is not a gate.

---

## What I would like to learn from your build, stated as questions

These are genuinely open. I do not know the answers and the proposal marks them so.

1. **Does authoring the fence first actually change the build**, or do you end up
   rewriting it once the tool exists? If the latter, S1 is theatre and I want to know.
2. **Which stage caught something?** If none did, that is the most important result
   available and it should go in the record.
3. **Who should fire the stages?** This is the ladder's biggest open risk (`G5`):
   discovery passes are manual-fire and the improvement loop is stopped by owner ruling.
   A ladder with no trigger is a mechanism rotting unexercised. Your build is a chance to
   see what firing them by hand actually costs.
4. **Is the mutation requirement (S2) worth it on a tool**, where the harness is a fence
   rather than a Go file?

---

## Status of the substrate, honestly

The thing that would let a *component* carry travelling docs like a tool does is
**submitted and not live**: council `e5673868-7c5b-489c-931a-7ba59b959b91`, round 1
REVISE (7 approve / 5 object), round 2 submitted with every objection answered. Migration
273 is **deliberately not applied** — image first.

**None of that blocks you.** Tools already have travelling docs (TL-017) and already have
the whole ladder from birth to browser verdict. Your build needs nothing from this lane
except the traps above. The substrate work is so *components* can join the same scheme.
