# PRE-REGISTRATION 2026-08-31 — which layer carries the meta-copy?

**Written BEFORE running anything**, per the copy lane's CANARY_2026-08-26 format and at their
request. Design sharpened by copy_quality_two_stage the same evening (their §"on the experiment").

## The claim under test

boxingonline.com's `about` page emits editorial-policy prose ("How we cover it — We write the way
a knowledgeable fan talks… A preview that says a fight 'could be great' tells the reader
nothing…") that paraphrases `site_specs.aspect='vertical_landscape'` → `lessons.avoid[]`
("Vague fight previews that say 'this could be a great fight' — every preview must contain
specific analysis of styles, records, and what's at stake").

**Hypothesis:** the research spec's `lessons` block is the carrier — a rule ABOUT the writing,
present in the writer's context, is emitted AS the writing.

**Why this shape is suspicious a priori:** `lessons.avoid[]` entries must QUOTE the thing they
ban in order to ban it. Quoted specimens in a prompt are the estate's strongest known carrier
(`a-quoted-exemplar-in-a-prompt-is-copied-verbatim`), and a ban's specimen survives
de-demonstration sweeps precisely because removing it would destroy the ban. The researcher's
sheet is entirely that shape.

## Arms (four, because a single-arm withhold cannot discriminate)

Per the copy lane's 646/647 lesson — layers each carry their own demonstrations, so a one-layer
withhold that still shows the manifesto tells you "not ONLY this layer", not "not this layer".

| arm | context |
|---|---|
| (a) | full context — control, must reproduce the manifesto or the experiment is void |
| (b) | `vertical_landscape.lessons` withheld |
| (c) | `strategy` withheld |
| (d) | both withheld |

## Battery, fixed now

1. The owner's four quoted passages as **exact needles** (the "could be great" sentence, the
   accuracy bullet, the opinion-kept-separate bullet, "How we cover it").
2. A first-person-plural editorial-policy regex — `we write`, `we'd rather`, `we cover`,
   `gets checked`, `we're trying to` — counted per page.
3. **The READ decides.** A battery-zero page that still narrates itself is a FAIL, not a pass:
   evasion (same behaviour, different words) is a measured failure mode in the Gemini trials.

## Refutation condition, stated before the run

**If arm (d) still emits the manifesto, the carrier is the brief or the model prior, and washing
the spec fixes nothing.** That is the outcome that kills the hypothesis, and it is a success of
the experiment, not of the fix.

Secondary refutation: if (b) and (c) each clear it alone, the layers are redundant carriers and
neither is *the* cause.

## Where it runs — NOT on the paid site

`site_delivery_and_editor` owns boxingonline's rebuilds and is holding delivery under the owner's
fix-before-delivery ruling. **Run the arms on an unpaid site with the same mechanism, or through
that lane.** Do not rebuild pages on a customer deliverable to satisfy an experiment.

## Status

Pre-registered, not yet run. Owner of the experiment: copy_quality_two_stage (their harness).
Filed here because this lane raised it. Companion:
`copy_quality_two_stage/CONTRIB_2026-08-31_from_the_first_paid_build_the_page_DESCRIBES_the_editorial_policy_instead_of_doing_it.md`.
