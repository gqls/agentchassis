# SUMMARY — CTA / link integrity, 2026-07-25: bug 023 is closed

*Milestone read-out. The previous one is `SUMMARY_2026-07-19_cta_link_integrity.md`; read them
as a series — this one is not a replacement for it.*

## What we're trying to do

Make it impossible for one of our sites to ship a button that goes nowhere. Not "fix the ones we
can see" — impossible, at the level of the component library every site is built from.

## Where we've come from

On 19 July the owner looked at leopardessconsulting.co.uk and said of four buttons: *"I don't
understand what these buttons are and what they do, I think they are broken."* They were, and
each of the four was broken by a **different mechanism that defeated a different check** — a
label frozen from a different tool, an empty destination rendered as `href=""`, a hardcoded
`#guide-start` that existed on no page, and a web address the AI had invented by taking the
site's own contact email and swapping the `@` for a dot.

Underneath all four was one structural fact: **a button's label and its destination are two
unrelated schema fields, and nothing in the platform expresses "a label implies a destination".**
The label nearly always renders (its source is `static` with a fallback, which bypasses every
required/on_missing rule). The destination may be absent, empty, unresolvable, or invented.

Since then: the four buttons were made extinct fleet-wide; the hardcoded-fragment class was
removed (migration 179); the four components that were on live pages were gated and their
invented-URL fields retired (migration 181); and a schema-derived CTA pairing shipped in
observe-only mode after a ten-round council trail. Three sibling defects were split out to their
own bugs — the write-only human-review queue (033), the library's only tool hero being hardwired
to a Bayesian ranker (045), and 312 live 404s from stale chrome (049) — so that 023 could close
on its own scope instead of being gated on other people's decisions.

## What we've done

Today we finished 023's own scope: **the whole active library, not just what is on a page now.**

- **Migration 211** — 156 CTA anchors across 31 active components wrapped in `{{if .x_url}}`, so
  an unresolved destination renders **no control** rather than a dead one. Seven more anchors in
  four further components were repaired that no field-based search could see, because their
  destination was a hardcoded `#` or a `{{else}}#` fallback; each got a proper renderer-owned URL
  field and a gate. And fleet-wide, **23 URL fields across 5 components stopped being both
  AI-authored and required** — that combination is the instruction to invent that produced
  `leopardess.contactforsales.com`.
- **Migration 212** — the last hardcoded `#` CTA in the library, found by the new lint on its
  first run.
- **`scripts/check_cta_gates.py`** — a standing check that reports all four fault shapes across
  the library in about a second, with its deliberate exclusions written into the file so nobody
  re-derives them.
- **RUNBOOK R17/R18** — how to run the lint, and the recipe for safely mass-editing live shared
  templates: parse (never regex-count), compute offline, prove the SQL equivalent by hashing
  read-only, then hash-gate the migration before *and* after so a concurrent session's edit
  aborts the transaction rather than corrupting a template.

Measured live after: **ungated CTA anchors 0** (was 156), **URL fields that are AI-authored and
required 0** (was 23).

## Where we are now

`bugs_open/023` is closed and moved to `bugs_closed/`. Its three criteria: the structural one —
no active component can pair a rendered label with an absent destination — is **met and verified
live**; the live-page criteria were already met and re-verified.

Three residuals are recorded rather than quietly dropped, and the lint keeps reporting them: one
card component whose link wraps its own image and text (gating it would delete content, not a
control), and two components whose stored "template" is a saved snapshot of a rendered page — a
different fault, already known to the vonc workstream.

Worth stating plainly, because it constrains what closure means: **gating a template does not
rewrite HTML that is already deployed.** Pages carrying dead controls rendered before today keep
them until their next render. The defect can no longer be *created*; the existing instances drain
as the fleet re-renders.

## Where we're going

Not 023's work any more, but the file's descendants:

- **The flip round** — stage 2 of the schema-derived pairing, where the schema decides CTA writes
  and the hardcoded six-entry map retires. It carries five binding constraints accumulated from
  five council seats and is the change that would give the last un-owned URL field an owner.
- **`bugs_open/033`** — the human-review queue that has never delivered a finding to anyone. The
  platform correctly detected one of the original four buttons two days before the owner clicked
  it. Detection was never the gap.
- **`bugs_open/045` and `039`** — component selection: the wrong component chosen, and no
  component chosen.
- **`bugs_open/049`** — links that resolve to pages that do not exist. Gating cannot help there:
  `/privacy.html` is not empty, it passes every gate, and it 404s.
