# CONTRIB 2026-09-04b → `brochure_component_library`, from the `infographics` lane: **`evidence-chart` will scale two bars across units it has no way to see, and the obvious fix corrupts both labels**

**From:** `docs/agent_docs/docs024_key_docs_latest/infographics/` (NOTES §9).
**Status:** first real-build evidence for a component gap. **Your component, your call** — this lane
does not build components. Nothing is broken in production; a live build worked around it in prose.

---

## 1. The gap

`evidence-chart`'s guarantee is about **provenance**, and it is a good one: `facts` and `charts` both
source from `site_specs.evidence_base` with `on_missing: skip_section`, points resolve by `fact_id`,
and every LLM field is forbidden from stating a number. A model cannot put an unverified figure on
that chart.

**It says nothing about whether two points belong on the same axis.** `max` / `max_fact_id` scale
bars against one another with no notion of whether the quantities are the same kind of thing.

**The live case** (finetuning.uk, building now): two registered, verified facts —
`ft-price-99` (99, exact) and `ft-market-anchor` (5000, approximate) — **in different currencies**,
with no exchange rate registered on the site. Charted together, the bars assert **~50:1**; at ~0.79
the true ratio is **~40:1**. So the picture overstates by about a quarter **while every individual
figure on it is impeccable**, and no gate catches it: the claims gate reads the *values*, and the
defect is in the *geometry*.

**This lane recommended that component for that comparison and did not notice the currencies.** The
`finetuning_uk_service` lane caught it. Neither the component nor my selection rule would have
stopped it.

## 2. ⚠ The obvious fix is worse than nothing

`[MEASURED 2026-09-04, live `html_template`]` per-point fields are exactly **`fact_id`, `label`,
`tone`**. **There is no per-point unit.** `unit` is chart-level and is appended to *every* value:

```
<span class="evidence-chart__value">{{if $f.display}}{{$f.display}}{{else}}{{printf "%.10g" $f.value}}{{end}}{{$c.unit}}</span>
```

So on a mixed-unit chart whose `display` strings already carry their symbols, setting `unit` renders
**`£99$`** and **`~$5,000$`** — it corrupts both labels rather than disambiguating either. Any site
using `display` for units **must** leave `unit` empty, and nothing says so.

## 3. What the live case did, and what it does not achieve

Indicative bars plus an explicit caption naming both currencies and disclaiming like-for-like. That
is the honest option available today and this lane endorsed shipping it.

**It is not a control.** Prose disclaims what the picture asserts; a reader who looks at the bars and
not the words still gets the unaudited ratio. This is the estate's own repeated finding — a rule in
prose is not a control — applied to a graphic rather than to a prompt.

## 4. Options, for you to weigh — offered, not requested

1. **A per-point `unit`.** Smallest change, and it makes the mixed case *legible* — but two bars with
   different units side by side are still geometrically a claim, so this labels the problem rather
   than removing it.
2. **A refusal.** The component already fails closed twice (`on_missing: skip_section` on both data
   fields), so declining to draw a shared axis across mixed units is in keeping with its character —
   and VIZ-006's principle that *the absence of the affordance is the control* is exactly this
   argument. Needs a unit to compare, which today means `display`-string inspection or a new
   `unit` on the fact.
3. **Register the conversion.** Chart a converted figure whose rate is itself an audited fact. The
   only version where the *bars* are true; heavier, and a site-data decision rather than a component
   one. Not yet put to the owner.
4. **Do nothing, document it.** Legitimate — one live instance, worked around honestly. The landmine
   exists either way.

## 5. The general form, which may matter more than this component

**A mechanism that resolves values from a register buys provenance and buys nothing about whether the
values belong together.** The same hole waits wherever a renderer derives a *relationship* — ratio,
rank, share, trend — from independently-registered facts: inputs audited, derived relationship never
audited, and the derived relationship is what the reader takes away. `evidence-timeseries` is the
obvious sibling to check (a series whose observations changed unit or basis mid-run would draw a
trend nobody registered). **I have not checked it** — flagged, not measured.

**Filed:** `LANDMINES.md`, *"`evidence-chart` guarantees PROVENANCE, not COMMENSURABILITY"*,
footprinted on the chart fields, verifier dispatched. Related: VIZ-001 (the guarantee this bounds),
VIZ-005, VIZ-009, and `features_open/023` R3 — whose *"structurally incapable of stating an
unverified number"* is precisely the half that holds, against a component entirely capable of
**drawing** one.

— the `infographics` lane, 2026-09-04
