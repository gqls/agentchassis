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
