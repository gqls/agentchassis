# RFC 009 — the renderer should enforce `input_schema.on_missing`, instead of trusting every template author to

**Filed** 2026-08-02 by `bugfix_140_contact_info_fabrication`, at the direction of
the council gate's **architecture** and **reuse_agent** seats, which reached this
independently in the same round (`40de12b0-36fa-4c06-82b4-995dc9098593`,
APPROVED with 7 advisory objections).
**Status** OPEN — raised, not decided. **Not blocking anything.**

> The seat's own verdict on the change that prompted this was **`point_fix`** —
> *"no new namespace, wire shape, or shared runtime contract is introduced … does
> not meet the needs_rfc trigger test on its own terms"* — while recording
> `ARCHITECTURE_SIGNAL: insufficient`. So this RFC is not an appeal against that
> plan and does not ask for it to be revisited. It exists because a **scope
> observation is answered by routing it, not by resubmitting the same plan with
> better measurements** (CLAUDE.md, owner ruling 2026-07-28).

## The claim

`content_components.input_schema` declares, per field, what to do when the datum
is absent:

```json
"hours":   { "source": "site_specs.identity.hours", "on_missing": "skip_field" },
"section_title": { "source": "llm", "fallback": "Contact Us", "on_missing": "use_fallback" }
```

**Nothing reads `on_missing` at render time.** It is documentation. Obedience is
hand-implemented by whoever writes the Go template, in `{{if}}` branches, per
component, per field, for ever.

## The evidence that this does not generalise

It has now failed **twice**, and both times the repair was a hand-rewritten
template rather than a mechanism:

| | what happened | fix |
|---|---|---|
| `bugs_open/111` (2026-07-28) | the fallback footer rendered an `<h4>Contact</h4>` over an empty mailto | `RenderFallbackFooter` gated by hand (`d4731109d`) |
| `bugs_open/140` (2026-08-02) | `contact-info` substituted `+1234567890`, `Monday – Friday, 9am – 6pm`, `info@example.com` for absent data — **8 live sites served the invented hours**, one served the invented phone as a `tel:` link | migration 287 gated each card by hand |

In the second case the schema said `skip_field` for exactly the three fields that
fabricated. The contract was correct and the template ignored it, and **no layer
between them could tell**.

Two costs are measured rather than asserted:

- `hours` is supplied by **0 of 1,089** `page_components` fleet-wide, so every
  Hours card the platform has ever served was fabricated.
- The desync in that same row was detected by `compute_component_quality.go` on
  **2026-05-18** (`schema_template_synced = false`) and consumed by nothing for
  eleven weeks.

## Why the current defences do not cover it

- **`scoreComponent`** (`compute_component_quality.go`) is absolute and
  single-artifact, and does not inspect `{{else}}` branch CONTENTS at all. It
  scored this component 80 while it was fabricating.
- **`component_write_guard.go`** is COMPARATIVE by design — "is this replacement
  worse than the row it replaces" — so a birth write carrying a fabricated
  fallback passes cleanly. Its header says as much and says why.
- **`check_placeholder_contact`** matches a roster of literals against rendered
  HTML. It contained not one literal our own library ships; across every unlocked
  row its nine patterns matched **1** while missing **9** fabrications.
- **`scripts/check_placeholder_fallbacks.py`** (CGV-029, shipped with 140) reads
  the live library and separates a fact default from a label default. **This is
  the closest thing we have, and it is advisory, by-hand and post-hoc.** It
  catches the third instance after it exists, not before.

Every one of those is downstream of the template author's decision. None is the
enforcement point.

## SIZING, measured 2026-08-02 after the RFC was first written — and it changes the picture

The RFC originally argued from two incidents. Here is the corpus, which is a
better basis for the decision and which nobody had measured.

**1. The contract is mostly ABSENT, not mostly disobeyed.** Across every active
component's `input_schema.fields`:

| `on_missing` | fields | components |
|---|---|---|
| **(not declared at all)** | **1,938** | **116** |
| `skip_field` | 181 | 56 |
| `use_fallback` | 21 | 9 |
| `skip_section` | 15 | 14 |
| `needs_human_review` | 8 | 6 |

**~90% of fields (1,938 of 2,163) declare no `on_missing` whatsoever.** So a
render-time gate driven by `on_missing` would be **inert for nine fields in ten**
on day one. Its reach is not a function of how well it is built; it is a function
of a declaration most component authors do not make. That is the single most
important number here and it cuts against option A being a general answer.

**2. The live violation surface is 68 fields across 20 components** — declared
`skip_field`, referenced by their template, with no `{{if .field}}` gate anywhere:

```
platform-comparison   15   (row1_platform1_value … row5_spark_value)
product-specs          8   (spec_1_value … spec_8_value)
system-stats           8   (stat1_label … stat4_description)
featured_article       6   hero-tool 5   case-studies-grid 5   Pricing Tiers 4
archetype-result-card 3   bayesian-ranking-hero-tool_pre_037 3
about-hero / case-studies-hero / contact-hero / hero / services-hero  1 each (subheadline)
content-listing 1   portfolio-showcase 1   social_proof 1   about-commercial-block 1
```

*[APPROXIMATE — this tests for a gate anywhere in the template, not a block-stack
scope check, so it over-counts a field gated in one place and used bare in
another. It does not under-count.]*

**3. And these are a DIFFERENT, milder class than `140`.** This is the
distinction the decision turns on:

- **`140`'s class — ungated AND substituting a literal.** The platform *asserts a
  false fact*: an invented phone number, invented opening hours. **Fleet-wide
  count today: 0.** `check_placeholder_fallbacks.py` (CGV-029) covers exactly this
  and reports clean across 173 active components.
- **The 68 above — ungated with NO fallback.** The platform renders a *blank*: an
  empty table cell in `platform-comparison`, an empty spec row in
  `product-specs`, a missing subheadline. Bad, sometimes visibly so on a spec
  sheet, but it **asserts nothing untrue**. Nothing currently detects it.

So the harm this RFC would prevent is mostly **blanks, not fabrications** — the
fabrication class is already closed by a lint that needs no roll and no schema
declaration. That materially weakens the urgency argument in the section above,
and it is why the recommendation below has changed to prefer the cheap options.

## The shape being proposed (not yet designed)

A render-time gate driven by `input_schema.on_missing`, applied **once** at the
`executeGoTemplate` / `RenderContext` layer: a field declared `skip_field` that
is absent yields nothing regardless of what the template says, and a field
declared `use_fallback` yields its declared `fallback` value.

## The hard questions, stated so nobody thinks this is easy

1. **Where does the schema reach the renderer?** `executeGoTemplate` receives a
   `map[string]interface{}` from `contextToInterfaceMap`, not the component row.
   The schema is not currently in scope at the point of execution, so this is a
   plumbing change before it is a policy change.
2. **`RenderContext` can legitimately supply what `content_data` lacks.** It
   carries top-level `Email`/`Phone` whose json tags reach the template contract,
   so "absent" is not simply "missing from `content_data`" — `idea.uk` renders a
   phone from site identity. Any enforcement must define absence across every
   path a value can arrive by. This is the same unsoundness that killed the
   roster-free version of the 140 detector.
3. **A template can already contradict the schema in the other direction**, and
   some of those are correct today. An enforcement layer that is stricter than
   the fleet's live templates would break working pages on the roll — the classic
   over-fire that gets a guard switched off.
4. **Who owns the migration path?** 172 active components exist; the ones whose
   `{{else}}` branches are legitimate labels must keep working unchanged.
5. **Is the right layer the RENDERER or the WRITE PATH?** Refusing a fact-shaped
   `{{else}}` at `store_generated_component` closes the door at birth and cannot
   break an existing page, but does nothing about the components already there.
   These are different trades, not the same fix at different depths.

## What is NOT being claimed

- Not that the 140 fix was wrong or should be redone. It is the correct point
  fix, it is live, and it was approved.
- Not that a third instance exists. `check_placeholder_fallbacks.py` reports
  **clean across all 172 active components** as of 2026-08-02. The exposure is
  prospective.
- Not that this is urgent. Two instances in eleven weeks, one now closed at
  source, with a standing lint that would surface a third.

## THE DECISION — four options, costed

Not "should we do this" but "at which layer, and is it worth it". The sizing above
should be read first; it moved my own recommendation.

**A — Render-time gate.** `executeGoTemplate` reads `on_missing` and enforces it:
`skip_field` + absent ⇒ nothing, whatever the template says.
*Reach:* every render, all 173 components. *Fixes:* all 68 blank-field violations
at once, and any future one, without touching a template.
*Cost:* the schema is not in scope at execution — plumbing first. Can BREAK LIVE
PAGES: any of the 173 templates whose current output depends on a bare
`{{.field}}` rendering empty changes behaviour on the roll, and the ones most
likely to are the 20 already in violation. *And it is inert for ~90% of fields*,
which do not declare `on_missing` at all. Highest power, highest risk, worst
reach-per-unit-effort.

**B — Write-path refusal.** `store_generated_component` refuses a new/updated
template that leaves a declared `skip_field` ungated, or that gives a fact-shaped
`{{else}}` literal.
*Reach:* new and rewritten components only. *Fixes:* nothing that exists today.
*Cost:* low, contained, cannot break a live page — it only ever refuses a write.
Needs calibration against the real corpus or it will refuse legitimate work (the
lesson `component_write_guard.go`'s header is built around).
Closes the door at birth; leaves the 68 where they are.

**C — Promote the lint (CGV-029) from advisory to routine.** It already exists,
needs no roll, no schema declaration, and is the only defence that does not depend
on `on_missing` being declared. Extend it to also report the 68 ungated
`skip_field` fields, and run it on a schedule or from the existing sweep rather
than by hand.
*Reach:* the whole live library, both classes. *Fixes:* nothing automatically —
it reports. *Cost:* hours, not days. Cannot break anything.

**D — Do nothing further.** The fabrication class is closed and measured at 0
fleet-wide. The remaining 68 are blanks, not false claims. Accept that a third
incident, if it comes, is caught by the lint and fixed per-component as the first
two were.

### Recommended: **C now, B next, A only on evidence**

The sizing inverted my original instinct. **A is the architecturally satisfying
answer and the poorest value**: it is the only option that can break live pages, it
requires plumbing the schema into the renderer, and after all that it is inert for
nine fields in ten because the declaration it depends on is usually missing. **The
`on_missing` contract is too sparsely populated to be load-bearing**, and making it
load-bearing is a bigger cultural change (get 1,938 fields declared) than a
technical one.

C is cheap, safe, already built, and covers the class A cannot (undeclared
fields). B closes the door for new components at low risk. A becomes worth
revisiting **if** the declaration rate rises, or **if** a third *fabrication*
incident occurs despite the lint — which would be evidence the lint is the wrong
layer. Neither is true today.

**What I would want from you:** just the C/B/A/D call. If C, I would also want to
know whether "routine" means scheduled or wired into the existing discovery sweep
— the second is more useful and slightly more invasive.

## Sources

`bugs_open/140` · `bugs_open/111` / `d4731109d` ·
`docs024_key_docs_latest/bugfix_140_contact_info_fabrication/` (PLAN, NOTES —
the council dispositions are recorded there) ·
council correlation `40de12b0-36fa-4c06-82b4-995dc9098593`, seats `architecture`
and `reuse_agent` · CGV-029 in `docs026_concept_register/register/content-governance.md`
