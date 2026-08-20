# PLAN — editorial design uplift (2026-08-20)

**The ask (owner, 2026-08-20), verbatim in substance:** the editorial feature pages
work but can look far better — *"more imagery and perhaps relevant to different
parts of the text, more graphic treatments, more typography design, more variation
in graphs and charts with perhaps more imagery and so on. And more timelines if we
can find them or even start collecting information for them."* Correspond with the
experience loop, the component loop, the visual designer and other relevant agents.
The owner named it a difficult ask needing its own plan, separate from the news
editorial work — hence this lane.

**Scope boundary.** `news_editorial_features` ships FEATURES (stories, substrate,
pages). This lane changes how the page family LOOKS, and owns the components that
serve it. Where they meet: this lane hands the other a better component set; it
does not author editorial copy.

---

## 1. What we are starting from

One live instance: `robot-hands.com/insights/robot-demand-step-change.html`
(NEWS-020) — six sections, **flat alternation** of hero → prose → time series →
bar charts → prose → CTA. Honest and legible; visually plain. Everything it uses
already existed.

**What is already good and must not be lost.** The chart components make the
unsourced state *unrepresentable*: a plotted point cannot carry its own number
(every value resolves through a `fact_id`), and a series observation carries its
own citation rather than inheriting the parent's. Any redesign inherits that or
it is a regression, not an uplift.

**Three constraints that decide what "better looking" can mean here** — all
verified, all load-bearing:

1. **No arithmetic in the render funcmap, and a missing function is a PARSE
   error** (VIZ-007). The funcmap is `default, eq, ne, lower, upper, isset, safe`.
   A template that computes coordinates renders **nothing** — it does not
   degrade. This is why both live charts pass values to CSS custom properties and
   let the browser divide.
2. **Text inside `<svg>` is invisible to the claims gate** (VIZ-009,
   `claims.go` `nonAssertionElements`). So "just draw it in SVG" is the wrong
   default for anything carrying words or figures: HTML text with CSS-drawn
   furniture stays the rule.
3. **Chart furniture is a graphical object → WCAG 3.0 non-text contrast applies**
   (VIZ-011), and the intuitive choice (`--color-border`) is usually the failing
   one.

## 2. The two mechanisms the owner asked us to find

Both are in **one** document:
`docs024_key_docs_latest/inline_guide_imagery/PLAN_2026-08-14_durable_inline_guide_imagery.md`
(design-only, nothing built).

**(a) Interleaving copy with images/graphs.** The blocker is not layout, it is
**durability**: `article-body` holds a whole article in ONE llm-owned `content`
field, so an in-body `<figure>` works today and is destroyed by the next prose
rewrite (LANDMINE, ~90ms measured loss window on a real page). Remedy:
**plan-as-truth** — a locked `site_plan_imagery` row per figure, consumed by the
already-live IMG-056 resolver, plus a `style_hints.placement` splice that injects
the figure into `rendered_html` only, with a writer marker that is optional and
never load-bearing.

**⚠ Ownership: that plan is the `inline_guide_imagery` / dartsonline lane's, not
ours.** Run `scripts/who-owns.py` and coordinate before touching it. This lane
CONSUMES it; it must not fork it.

**(b) Components comprising other components.** Owner steer, 2026-08-15. Measured
2026-08-20: `page_components.parent_instance_id` exists with FK and index and is
used by **0 of 1580 rows**, with **zero Go references**;
`render_mode='composite'` + `child_components` exist but `deriveRenderMode` can
only emit `agent`|`template`, so composite is **unreachable by construction**;
assembly is flat concatenation and the single template executor has no
`{{template}}` support. **So composition is build-and-prove, not wiring.**

`features_open/035_FEATURE_component_hierarchy.md` is the reserved slot for that
design. **It is Fable's to write** (owner, reaffirmed 2026-08-20) and is
**BLOCKED**: four dispatches now, the latest failing on
*"You've reached your Fable 5 limit"*. **Not substituted.** The brief is ready;
this is a capacity block, not a knowledge one.

**The consequence for sequencing, and it is the main design judgement in this
plan:** the editorial page family does **not** need composition to look
dramatically better. Section-level alternation already interleaves prose and
figures — the live page proves it. So this lane proceeds on the flat mechanism,
and treats composition as an *unlock for later shuffling*, not a prerequisite.
Anything that would only be possible with composition is parked behind 035.

## 3. Correspondence with the loops (what each can actually tell us)

Eleven live design/experience agents [MEASURED 2026-08-20]: `visual-designer`,
`brand-designer`, `feature-designer`, `site-design-planner`, `design-audit-agent`,
`design-discovery-agent`, `visual-design-auditor`, `webdesign-agent`,
`experience-planner`, `experience-approval-council`, `experience-register-writer`.

| correspondent | what it is | how we ask it | what we get back |
|---|---|---|---|
| **design-audit-agent** | a real orchestration: spawns `visual-design-auditor` + a content auditor, runs algorithmic checks AND an LLM visual audit, writes findings | kcat `orchestrate`, input `{site_id, domain}` (site-wide) | findings written by its own `write_findings` step |
| **experience loop** | `experience-planner` + a 4-critic council; **honesty critic holds a hard veto** and has historically been right every time it refused | G17 convention: durable intake item + kcat orchestrate on one correlation | an approved (or refused) experience plan — journeys, promises, where imagery belongs |
| **component loop** | `compute_component_quality` + improvement-loop triage | score the editorial components directly | per-component quality signal |
| **visual-designer / brand-designer** | single-step design agents | direct orchestrate | art direction proposals |

**Dispatched already:** `design-audit-agent` at robot-hands.com, correlation
`51404b33-5287-42cf-b74e-93b5f8d3ea29` — so this plan's next revision is built on
its findings rather than on taste. **Note the honest limit: it audits a SITE, not
a page**, so its findings will cover robot-hands as a whole and must be filtered
to the editorial pages.

**Standing rule for every one of these, from design D (`bugs_open/126`):** an
artefact this lane generates enters **human review; never auto-repair.** The
cited defect is a failing tool acceptance auto-raising a repair job carrying the
failing criteria as its spec — the only way to satisfy it was to delete a legally
load-bearing consent gate. Fine while a human watches one thing at a time;
unacceptable in a lane generating across pages and sites.

## 4. The work, in dependency order

**Phase A — measure before designing (no new components).**
A1 Harvest `design-audit-agent` findings; filter to the editorial pages.
A2 Score the editorial components with `compute_component_quality`.
A3 Run `render_audit` / `render-audit-agent` over the two live editorial pages —
   contrast and overflow at the *served* page, including chart furniture (VIZ-011).
A4 Write the **design brief** for the family from A1–A3 + §1's constraints.
*Falsifier: if A1–A3 surface nothing an ordinary reader would notice, the uplift
is a taste project and should be scoped down and said so.*

**Phase B — typography and graphic treatment (cheapest, no new mechanism).**
Editorial furniture the family lacks: standfirst, drop cap, pull-quote, section
rules, figure captions as a designed object, and a big-number moment (the
brochure library's `stat-band` already exists — reuse before building). Delivered
as `css_snippets` + small template edits. Council-gated; each shared-component
edit counts its blast radius first.

**Phase C — hero and in-body imagery.**
C1 Hero: image + semi-transparent overlay is now the DEFAULT (owner ruling) —
   already satisfied by the live template's image branch; the work is ensuring
   every editorial page has a generated `content_hero_*` asset and per-site art
   direction via `style_hints`.
C2 In-body figures: join the `inline_guide_imagery` phases at the shared-component
   step, **in coordination with that lane**. This is what "imagery relevant to
   different parts of the text" actually requires.

**Phase D — chart variation (substrate first, always).**
The sequencing rule is VIZ-002/VIZ-003's and it is not negotiable: **a renderer
built before its substrate is an invitation for a writer to fill the series from
the model.** So each new chart kind gets its fact shape registered first, with a
worked series, then the renderer.
D1 Line/slope treatment for longer series (the current column form reads as
   comparison, not trajectory).
D2 Small multiples for the same measure across regions/segments.
D3 Annotated series — a point carrying an event label, which is the bridge to E.

**Phase E — timelines (the owner asked for these specifically).**
E1 **Start collecting now, before any renderer exists** — each feature's substrate
   registers the story's dated events, each with its own citation, alongside its
   series. Implemented as evidence facts (an events shape generalising
   `Observation`), **not** a new `event_timeline` table: reuse before build, and
   the claims gate already understands facts. Revisit a table only if the fact
   shape strains.
E2 The timeline component, once real data exists: HTML text + CSS furniture, each
   event carrying its citation, in the `mechanism-flow` idiom (which already
   proves a no-numeric-field component can carry a sequence honestly).

**Phase F — composition (BLOCKED on 035/Fable).** Revisit once the plan exists.
Nothing above depends on it.

## 5. What would make this fail

- **Building renderers before substrate** — the one failure mode this platform has
  already learnt twice (VIZ-002 was deliberately built second; features_open/023's
  charts fail closed on purpose).
- **SVG text** — leaves the verification net entirely (VIZ-009).
- **Forking the inline-imagery plan** instead of coordinating with its lane.
- **Silently substituting a model for Fable** on 035.
- **Treating a site-wide audit's findings as page-specific.**
- **Letting a design change reach `rendered_html` without re-running claimscan** —
  every editorial page is figure-dense by construction, so the claims surface
  scales with the redesign.
