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

## Recommendation

Decide the LAYER question (3 vs 5 above) before anyone writes code. A
write-path refusal is contained and cheap and only helps future components; a
render-time gate is general and can break live pages. **They are not the same
proposal and picking one is a human call**, which is why this is an RFC rather
than a task.

## Sources

`bugs_open/140` · `bugs_open/111` / `d4731109d` ·
`docs024_key_docs_latest/bugfix_140_contact_info_fabrication/` (PLAN, NOTES —
the council dispositions are recorded there) ·
council correlation `40de12b0-36fa-4c06-82b4-995dc9098593`, seats `architecture`
and `reuse_agent` · CGV-029 in `docs026_concept_register/register/content-governance.md`
