# PLAN — the in-body image slot AND the visual designer's dispatch path, as one design pass

**Opened 2026-08-25** by `vigilant_designer_offer_analysis`, on the owner's decisions relayed through
the `loanzy_uk_example_site` lane: *"Give the visual designer a dispatch path"* and *"Yes an inbody
image slot can be default."* My user asked for both **in one design pass**, which is correct — an
agent that plans imagery is worth little without a slot to place it in, and a slot nothing fills is
decoration.

**Nothing here is built yet. Nothing live has been changed.** This is the design, with the
measurements it rests on, for a council round.

Canonical owner record: `../loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md`.
My measurements: `../loanzy_uk_example_site/CONTRIB_2026-08-25_two_of_the_three_agents_he_names_could_not_have_run.md`.

---

## 1. What the diagnosis actually is — and it is NOT what the review implies

The review's verdict is *"the visual designer hasn't done its job"*. The measurement says something
different and more fixable.

`[MEASURED 2026-08-25]` homegarden.uk's 45 live components, by the component they instantiate:

| component | instances | can its template render `<img>`? |
|---|---|---|
| **Generic Text Block** | **19** | **NO** |
| **hero** | **18** | **NO** |
| info-card-grid | 2 | yes |
| contact-hero / contact-info / about-content / about-hero / contact-form / period-calendar | 1 each | no |

**37 of 45 instances (82%) are the two components that structurally cannot carry an image.** And
`Generic Text Block` — the body of nearly every page — is, in full:

```html
<section id="{{.InstanceID}}" class="section section--generic">
  <div class="container">
    <h2 class="section__title">{{.heading}}</h2>
    <div class="section__content">{{.content}}</div>
  </div>
</section>
```

Two fields, `heading` and `content`. **There is nowhere to put an image.**

⚠ **So the imagery gap is NOT a taste failure and NOT an absent capability in the library.**
`[MEASURED]` **47 of 369** library components DO render `<img>`. But reading their names, they are
**chrome and index furniture** — headers (`header-*` ×many), heroes, and card/listing thumbnails
(`content-listing`, `category-listing`, `featured_article`, `case-studies-grid`,
`featured-inventory`). **Not one places an image inside a prose section.**

**This is `bugs_open/381`'s shape exactly, one notch narrower:** the planner composes pages from
components that cannot express the page it planned. 381's version was "a page headed *month by
month* with no months"; this one is "an editorial article that cannot hold a picture".

**That also explains the fleet census** — 13 of 27 sites with ZERO inline `<img>`, and the best
anywhere 16%: the 16% is card thumbnails on listing-heavy sites, not editorial imagery. **No site has
ever had an in-body image, because no component has ever offered one.**

## 2. Item 3 — the in-body image slot. Follow 381's proven three-part shape

381 fixed the identical class over 24–25 August and **proved it end to end on a live build**
(0 of 7 pages carrying structure → 19 of 20). Its shape was three parts, and all three are needed
here or the fix is inert:

### 2a. Build the component (the missing capability)

A prose section that carries an optional figure. Sketch, deliberately close to `Generic Text Block`
so the planner sees it as the same kind of thing:

```html
<section id="{{.InstanceID}}" class="section section--generic section--with-figure">
  <div class="container">
    <h2 class="section__title">{{.heading}}</h2>
    <div class="section__content">{{.content}}</div>
    <figure class="section__figure">
      <img src="{{.image_path}}" alt="{{.image_alt}}" loading="lazy" width="{{.image_w}}" height="{{.image_h}}">
      <figcaption>{{.image_caption}}</figcaption>
    </figure>
  </div>
</section>
```

⚠ **`image_path` MUST be the deployed git path (`/assets/images/<key>.jpg`), never `assets.url`** —
that column is a presigned S3 URL with a **7-day expiry**, and a page that stores it serves a broken
image a week later. `webPath()` is what resolves the correct form. This is a standing landmine and it
is the single easiest thing to get wrong here.

⚠ **`alt` is not optional and not decoration.** A `<figure>` with an empty `alt` is worse than no
image for a reader using a screen reader, and it is the kind of thing that ships because nobody
asserted on it. The component's own contract should refuse an empty alt.

### 2b. Teach the planner it exists (the capability token)

381's migrations `592`/`593` put capability tokens into the site-planner and content-gap-planner
menus, and its evidence shows the planner then **chose** `period-calendar` unprompted. Same
mechanism, one more token. Without this the component exists and is never selected — which is 381's
own outstanding negative (its checklist component *was* offered and never chosen, and that is
recorded as unexplained; see §5).

### 2c. Teach the writer to fill it

381's other arm changed the writer's instruction, and that is what produced structure on 20 of 45
sections. The analogue: the writer must know that a figure slot exists, and **which asset to name**.

⚠ **THE HARD PART IS 2c, AND IT IS WHERE I WOULD EXPECT THIS TO FAIL.** The writer does not today
receive a list of the site's available assets. Two routes, and I recommend the first:

1. **Resolve server-side, not in the prompt.** The writer emits an intent (`image: "spring pruning"`)
   and a deterministic step matches it against `assets` for that site and writes the web path. The
   LLM never handles a URL.
2. ~~Give the writer the asset list and let it emit the path~~ — **rejected.** An LLM emitting image
   paths is a phantom-link generator; this estate has a whole bug family about exactly that
   (`bugs_closed/029`, the tool-crosslink phantom links, where the emitter fabricated
   `/tools/{function}.html`). Do not hand a model a URL template.

⚠ **And the supply question, unmeasured:** homegarden generated **13** assets for **21** pages. Even
a perfect slot cannot place an image that does not exist. **Whether the imagery pipeline produces
enough assets per page is a separate measurement nobody has taken**, and it belongs in this plan as a
prerequisite rather than a surprise.

## 3. Item 2 — the visual designer's dispatch path. The placement question comes first

`[MEASURED 2026-08-25]` `visual-designer` is active, storage-granted, has a real LLM step
(`design`: `execute_llm_prompt` with `ai_service` + `prompt_template`), and **is referenced by
nothing**: no `scheduled_tasks` row, no live agent's config, no live script; the single Go hit is a
capability grant inside `isStorageEnabledAgent`. Zero `llm_call_log` rows under its own type across
the log's whole span (2026-03-25 → today, 67,789 rows).

**So "give it a dispatch path" is not a wiring task with an obvious target. It is a placement
decision, and I think the owner has already made it elsewhere.**

### 3a. ⚠ The owner has just placed the checkers POST-HOC — *"the checkers will check after the fact (improvement loop)"*

That ruling (relayed 2026-08-25, recorded in
`../loanzy_uk_example_site/REFERENCE_2026-08-25_site_acceptance_council.md`) is about the site
acceptance council, and **it decides this question too, if we let it.** `improvement-loop` already
carries `spawn_design_audit` → `design-audit-agent` and `spawn_offer_analyser` → `offer-analyser` as
real `spawn_agent` steps. **A visual-designer seat there is one more step in a workflow that already
has seven, on a carrier that already exists.**

**Recommendation: put it in `improvement-loop`, not in the build path.** Reasons, in order:

1. **It is consistent with his own placement ruling**, so it needs no separate argument.
2. **It cannot break a build.** A never-exercised agent inserted into the build path affects every
   site built afterwards; the same agent in a post-hoc loop can only file findings. Given it has
   **never made an LLM call**, we have zero evidence about its output, and putting zero-evidence
   output on the critical path is the trade this estate keeps regretting.
3. **It is measurable before it is trusted** — findings accumulate, and we read them before anything
   acts on them.

⚠ **AND IT INHERITS A DEFECT THAT MUST BE FIXED FIRST.** Seven of the eight `call_agent` seats in
`improvement-loop` declare `error_step` **identical to** `next_step` — a seat that fails routes
exactly where success routes, and `record_audit_pass` then asserts a clean audit. Adding an eighth
fail-open seat makes it worse. **The fail-open fix is a precondition of this item, not a follow-up.**
It is the `loanzy_uk_example_site` lane's RFC and I am not taking it; I am declaring the dependency.

### 3b. The prior question nobody has asked: should it be THIS agent?

Honest option worth putting to the owner rather than deciding here: `site-design-planner` and
`webdesign-agent` are what actually produced homegarden's look, and both ran. `visual-designer` may
be a **superseded** agent that was never deleted — `[MEASURED]` it has never run, and 37 disabled
site-pinned `scheduled_tasks` rows show this estate accumulates spent machinery rather than removing
it. **Wiring up an abandoned agent because it has a promising name is a real risk**, and the cheap
test is to read its prompt against what `site-design-planner` already does before giving it a path.
**That read is the first task of this item and it is not done.**

## 4. How the two interlock, which is why they are one pass

- A **slot with no planner** gets built and never chosen (381's checklist, still unexplained).
- A **designer with no slot** produces recommendations nothing can express — the current state.
- **Both, with 2c's server-side asset resolution**, is the only combination where an image
  recommendation becomes a served image.

Order: **2a+2b+2c first** (the slot is useful on its own, and 381 proved the mechanism), **then** the
designer seat once the fail-opens are closed and its prompt has been read.

## 5. What is NOT decided, and what would disconfirm this plan

- **The supply question (§2c).** 13 assets / 21 pages. Unmeasured fleet-wide. **If sites do not
  generate enough imagery, the slot is the wrong fix and the generator is the right one.** Measure
  before building.
- **381's checklist negative.** It offered a component and the planner declined it, and nobody knows
  why. **That is the same mechanism this plan depends on.** If the planner declines the figure slot
  too, 2b is not sufficient and the whole shape needs rethinking. **Read their finding before
  building, and treat their outstanding negative as this plan's biggest risk.**
- **Whether `visual-designer` is the right agent at all (§3b).**
- **Scope of "default".** The owner said *"unless determined otherwise"*, so the slot should be
  available and preferred, not mandatory — a comparison table page does not need a figure.

## 6. Council and register

Both items are platform/config changes on a shared seam: a new component in the library, planner menu
changes affecting every subsequent build, and an agent entering a shared loop. **Council gate, and a
concept-register entry in the same commit that ships the seam** (2026-07-28 ruling, condition 2 —
which is the whole of the requirement since the 2026-07-29 ruling retired condition 1).

---

## 7. ⚠ §5's FIRST PREREQUISITE IS NOW MEASURED, AND IT PARTLY DISCONFIRMS §2

§5 said the supply question was unmeasured and that if sites do not generate enough imagery, the slot
is the wrong fix. **Measured, same day, before building anything:**

`[MEASURED 2026-08-25]` assets per page, all active/deployed sites with >5 pages (27 sites):

| band | sites | examples |
|---|---|---|
| ≥ 2.0 per page | 4 | remortgagecalculator.uk 2.17 · robot-hands.com 2.11 · fundamentallyai.com 2.00 |
| 1.0 – 2.0 | 3 | agritec.uk 1.31 · idea.uk 1.26 · webdesign.uk 1.25 |
| **0.5 – 1.0** | **13** | leopardess 0.96 · **homegarden.uk 0.62** · finetuning 0.58 |
| **< 0.5** | **4** | relojistas 0.46 · gaswholesalers 0.19 · **webdesign.co.uk 0.10 (149 pages, 15 assets)** |
| **ZERO assets** | **3** | adversecreditmortgage.co.uk (19 pages) · loancash.co.uk (22) · loanandmortgagecalculator.co.uk (47) |

**The median site has fewer than one asset per page, and heroes already consume them as background
images.** On more than half the fleet there is simply nothing left to put in an in-body figure.

### What this changes

**It does NOT kill the slot.** The slot is still a precondition — no amount of supply helps while the
prose component has nowhere to put an image, and 4 sites are at ≥2.0 per page where a figure could be
placed today.

**It DOES kill the claim that the slot alone answers the owner's ask.** He asked for *"much more
imagery, placed between paragraphs"*. Shipping the slot on its own would produce a figure on a
minority of sections, on a minority of sites, and — this is the part that matters — **it would look
like the work was done.** A green build, a new component in the library, a register entry, and pages
still nearly bare.

⚠ **So the honest scope is TWO changes, not one, and the second is the bigger:**

1. **the slot** (§2) — a precondition, cheap, follows 381's proven shape;
2. **the imagery supply** — how many assets a build generates per page, and on what rule. **Three
   sites have ZERO.** That is not a slot problem and no component change touches it. It is a
   different mechanism, probably a different lane, and it is **unowned as far as I can tell.**

**Recommendation, revised:** still build the slot first, because it is a precondition and it is
proven-shape work. **But do not report the slot as answering the review.** State the supply figure
alongside it, and name the second change as outstanding — otherwise this repeats the exact pattern
this lane filed `bugs_open/395` about this morning: a status that records the handler succeeded while
the thing that was asked for did not happen.

**And measure the same census after the first build that uses the slot.** If assets-per-page has not
moved, the second change is the whole remaining job.

---

## 8. ⚠ §2a IS SUPERSEDED — THE COMPONENT ALREADY EXISTS, AND I NEARLY BUILT A DUPLICATE OF IT

*2026-08-26. The reuse check `605`'s house style demands ("REUSE CHECKED FIRST") caught this before a
line was written. Recording it in full because the near-miss is more useful than the fix.*

### 8a. What is actually there

**`Illustrated Text Block`** — `content` category, section level, active. Its template is prose plus a
gated figure:

```html
<h2 class="section__title">{{.heading}}</h2>
{{if .image_url}}<figure class="itb__figure">
  <img class="itb__image" src="{{.image_url}}" alt="{{.image_alt}}" loading="lazy">
  {{if .image_caption}}<figcaption class="itb__caption">{{.image_caption}}</figcaption>{{end}}
</figure>{{end}}
<div class="section__content">{{.content}}</div>
```

**It is not a rough draft — it is better than what §2 proposed.** Lazy loading, a caption, responsive
CSS on theme variables, and the whole figure gated so it degrades to plain prose. And critically:

- **`image_url` / `image_alt` carry `source: "site_assets.image"`, NOT `source: "llm"`.** The paths
  are resolved server-side from the site's own assets. **That is exactly the design §2c argued for
  and marked as the hard part** — *"resolve server-side, not in the prompt… an LLM emitting image
  paths is a phantom-link generator"*. Somebody had already reached the same conclusion and built it.
- `on_missing: "skip_field"` — a section with no matching asset renders as prose, no empty frame.
- The `content` field's own guidance **forbids the writer emitting `<img>`, `<figure>` or `<iframe>`**,
  in terms, *"this component already has its own image fields… an image written into the prose would
  bypass them"*.

**`[MEASURED 2026-08-26]` it has SIX live instances in the entire estate, all on ONE site (apis.uk).**

### 8b. Why nothing chooses it — and it is `bugs_open/381`'s exact mechanism, unfinished

The planner's menu is built from `component_expresses(html_template, input_schema)`, the capability
derivation 381 shipped. Run against the two components in question:

| component | what the planner is told it expresses |
|---|---|
| `Generic Text Block` | `[html-block, list, table]` |
| **`Illustrated Text Block`** | **`[html-block, list, table]`** |

**Identical.** The image capability is invisible. The planner is choosing between two components it
has been told are the same thing, and picks the plain one — 19 times on homegarden.uk alone.

`component_expresses` derives exactly four tokens — `html-block`, `list`, `table`, `items` — and
**there is no image token in the vocabulary at all.** 381 taught the planner to see lists, tables and
item sets. Imagery was never added.

> **So this was never a missing component. It is a missing WORD.** The estate can express an
> illustrated section and cannot say so, which is 381's own sentence — *"the part choosing a page's
> components could not see what any component was capable of"* — one vocabulary item short of done.

### 8c. The fix, and it is one UNION arm

Add a fifth arm to `component_expresses`, derived from the SCHEMA rather than the template, exactly as
`html-block` and `items` already are:

```sql
UNION
SELECT 'image' WHERE EXISTS (
  SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
   WHERE f.value->>'source' = 'site_assets.image')
```

⚠ **Derived from `source`, NOT from `<img` in the template, and the difference is the whole
precision of it.** A template grep reports `image` for **47** components — every header, hero and
card thumbnail, whose pictures are chrome the writer cannot influence. The schema predicate reports
it for components that offer the writer a *server-resolved image slot*:
`[MEASURED 2026-08-26]` **9 components, 8 of them active and section-level** — `Illustrated Text
Block`, `case-studies-grid`, `content-block-about`, `featured-inventory`,
`game-master-explanation`, `hero-headline`, `product-details_pre_037`, `product-specs`,
`tool-guide-intro`. Bounded, inspectable, and every one is a genuine content component.

After it, `Generic Text Block` is unchanged and `Illustrated Text Block` reads
`[html-block, image, list, table]` — distinguishable for the first time.

### 8d. What §2 got right, wrong, and what still stands

- ~~**§2a — build the component**~~ **WRONG, and the most useful thing in this plan.** It exists, it
  is better than my sketch, and building mine would have produced a near-duplicate with a worse asset
  story — the fork-vs-reuse defect this estate keeps filing. **The reuse check is what caught it, and
  only because 605's header made it a required step rather than a good habit.**
- **§2b — teach the planner** — RIGHT, and now the *whole* job rather than a third of it.
- **§2c — resolve asset paths server-side** — RIGHT, and already done by whoever built the component.
  My rejection of the LLM-emits-paths route stands and needs no new work.
- **§7's supply finding is UNAFFECTED and still the bigger half.** Making the component selectable
  does not create assets. A median site under one asset per page still has little to place, three
  sites have none, and `on_missing: skip_field` means those sections quietly render as prose. **The
  word and the supply are two changes, and this only makes the first one cheap.**

⚠ **One honest limit on the owner's actual words.** He asked for imagery *"BETWEEN PARAGRAPHS"*. This
component places its figure **above** the prose, once per section. At page level, alternating
illustrated and plain sections gives interspersed imagery; **within a single section it does not.**
If he meant strictly mid-prose, that is a different component and a different plan — worth asking
rather than assuming, because the cheap fix answers the page-level reading and not the other one.

### 8e. Revised shape of the work

1. **One migration** adding the `image` arm to `component_expresses` — council scope (it changes what
   every planner is shown, fleet-wide, on every subsequent build).
2. **Re-measure `component_expresses` output for the 8 affected components** before and after, and
   confirm `Generic Text Block` is byte-identical — the control that distinguishes a widening from a
   reshuffle.
3. **A writer-side check** that the guidance for `image_url` is reachable — the field is
   `site_assets.image`, so what resolves it is server-side, and I have not yet read that resolver.
4. **Then, separately, the supply question (§7)**, which nothing here touches.
