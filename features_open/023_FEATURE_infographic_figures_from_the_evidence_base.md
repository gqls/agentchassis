# 023 FEATURE — infographic figures must come from the evidence base (and where generated images stop)

**Filed:** 2026-07-25, at the closure of `bugs_closed/011` (hero provider routing), where these
were carried as "R3" and "R4". Split out rather than dropped: 011's routing defect is fixed and
live, but it *unlocked* a capability — publishable infographics with legible text — and the
guard rails for that capability were never built.
**Status:** OPEN — designed, not built.

## Why this exists at all

Before `bugs_closed/011`, the fleet could not put readable text in a generated image, so the
question "where do the numbers in that image come from?" never arose. The routing fix made
publishable infographics real, and the first one produced — `infographic_what_we_build` on
leopardessconsulting.co.uk — was accurate **only because a human hand-wrote audited figures into
the prompt and explicitly forbade any others**.

That is prompt discipline holding a line that should be structural. The platform already learnt
this lesson one layer up: `bugs_open/043` is generated page *copy* inventing quantitative claims,
and it needed a detector plus writer blocks plus a sweep before it stopped. An invented figure
inside a JPEG is strictly worse, because none of that machinery can see it.

## R3 — build infographic prompts from `site_specs.evidence_base`

**The intent:** an infographic should be structurally incapable of stating an unverified number.
Prompts for `kind:"infographic"` are assembled from `evidence_base` facts, and the instruction
that no figure outside the supplied set may appear stops being a sentence a human remembered to
type.

**The substrate exists but is thin** (measured 2026-07-25):

```sql
SELECT count(*) FILTER (WHERE is_current) AS current_rows,
       count(DISTINCT site_id) AS sites
  FROM site_specs WHERE aspect = 'evidence_base';
-- → 8 current rows, 8 sites (17 rows all-versions)
```

**8 sites of the fleet have an evidence base at all.** That is the first thing to confront: a
generator that can only use audited figures will produce *nothing* on a site with no audited
figures. Whether that is the correct behaviour (fail closed, no infographic) or a blocker to
adoption is the design decision this feature turns on — and it should be made deliberately,
because "fall back to the model's own numbers" is the failure mode the whole feature exists to
prevent.

Ties directly into the claims-verification layer, which owns what counts as audited.

## R4 — the boundary: generated images explain, code-rendered SVG states

The other half, and it is a **scoping rule, not a build**. Generated infographics are now good
enough for *explanatory* graphics — how a job runs, what we build, the shape of an architecture.
They remain the wrong tool for any chart whose values must be:

- **exact** — a diffusion model draws a bar of approximately the right height;
- **selectable** — text baked into a JPEG cannot be copied;
- **translatable** — it cannot be re-rendered in another language;
- **screen-reader accessible** — it is, to assistive technology, a blank rectangle.

For those, the answer stays what it was: **Go emits SVG from real values** (the "L7 chart
component" named in 011 R4 — designed there, not built, and not referenced anywhere else in the
repo as of this filing). 011's correction removed the argument that we *needed* SVG for
explanation; it did nothing to the argument that we need it for data.

**The trap this rule exists to prevent:** the infographic lane now works well enough to be
tempting for a chart. The first time someone generates a bar chart because the picture came back
looking convincing, we will have published a graph whose bars mean nothing — and, per
`features_open/022`, nothing in the pipeline will notice.

## Sequencing note

022 (the legibility/number guard) and this file overlap deliberately: 022's *numbers* check —
"flag any number not present in the request" — is the **detector**, and R3 here is the
**source of truth** that makes the request trustworthy in the first place. Built together they
close the loop; built apart, 022 first is the useful order, because a detector with no evidence
base still catches invented figures, while an evidence base with no detector is back to relying
on discipline.
