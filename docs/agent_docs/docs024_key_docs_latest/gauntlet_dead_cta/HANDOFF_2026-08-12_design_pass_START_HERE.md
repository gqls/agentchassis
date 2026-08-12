# HANDOFF — vonc.com design pass, requested 2026-08-12

**Owner request, verbatim:** *"Please initiate a design pass, the design of the site can
still be a lot better."*

This file is a **cold start for a fresh session**. It is deliberately a scoping document,
not a plan: the previous session had almost no context budget left when the request
landed, so what follows is what it established first-hand plus the traps it knows about.
**Nothing has been designed, changed or dispatched.**

---

## 1. THE RULE THAT GOVERNS THIS ENTIRE TASK — read it before anything else

> **EVERY SITE GOES THROUGH THE FRAMEWORK. Never hand-build one (OWNER RULING,
> 2026-08-04).** No hand-authored HTML uploaded to the bucket, however small, however
> temporary, however much faster it would be.

**This ruling was written because a session hand-built the webdesign.uk shopfront** — on
the lane whose entire product is framework-built sites. Two reasons it is a rule and not
a preference: a hand-built page **demonstrates nothing** on a site selling this
capability, and it silently opts out of every control the pipeline applies. **"The
framework cannot do it yet" is a bug to file, not a licence to hand-roll.**

A reinforcing ruling from the same owner, 2026-08-06: **the framework writes the
content, not you.** That one was earned three times over on the provocation lane in the
week to 08-12 — a session's hardcoded example paragraph turned out to describe the
corpus wrongly, and a session's hand-drafted provocations were the ones the owner later
rejected as unreadable.

**So: a design pass here means driving the design pipeline, not opening a CSS file.**
If you find yourself writing a rule like `.hero { padding: 4rem }`, stop and ask which
agent should have decided that.

## 2. What the site actually is [MEASURED 2026-08-12]

**23 pages**, nearly all redeployed 2026-08-12:

| kind | count | notes |
|---|---|---|
| `entity-page` | 8 | `/archetypes/*.html` — catalyst, judge, maker, mentor, oracle, scout, surgeon, wildcard |
| `blog-post` | 4 | three tool guides + `/blog/provocation.html`, which is **`planned` and has never been built** |
| `tool` | 3 | gauntlet round record, take-strength-scorer, archetype-clash-calculator; **`tool-arena-interface` is `needs_rebuild`, last deployed 2026-08-03** |
| `content` | rest | home, about, contact, … |

**Two things are already wrong before any aesthetic judgement is made**, and both are
cheap wins a design pass should not step over:

1. **`tool-arena-interface` has sat at `needs_rebuild` since 2026-08-03.** ⚠ Do not size
   this by your own change — a stale page holds *every* improvement made since it last
   rendered, so rebuilding it may change far more than expected. Canary two pages; they
   will disagree.
2. **`/blog/provocation.html` is `planned` and was never built.** Decide whether it
   should exist at all before designing anything around it.

## 3. Where the design machinery lives — start here, do not re-derive it

`docs026_concept_register/register/design-composition.md` is the index of what exists and
what is **superseded**. Read the status lines: DES-007, DES-008, DES-009, DES-011 are all
marked superseded and DES-010 abandoned, so a naive grep will lead you to dead agents.

The live shape is **DES-003: direction → composition → execution**
(`site-design-planner` + `webdesign-agent`), with **DES-001**'s three layers
(`content_components` / `css_themes` / `style_collections`) underneath and **DES-005**'s
`resolved_composition` pointer spec and `install_site_composition` semantics as the
install path. **DES-012** records the pipeline's guiding mottos — read those before
proposing a direction, they are the closest thing to a stated house style.

Also relevant: `styling-render-pipeline.md` and `visualisation-and-charts.md` in the same
register.

## 4. Landmines that will bite a design pass specifically

Grep `LANDMINES.md` for each footprint before touching it — these are the ones already
known to fire on this kind of work:

- **`generic_theme` misfires fleet-wide (colour churn).** Pin the palette via
  `design_intent.palette.reference_values`; see the `webdesign colour-churn` entry.
- **A CSS `var(--x, #fallback)` literal is in the source and NEVER applied.** Grepping
  the stylesheet proves nothing about what the browser paints — ask `getComputedStyle`,
  and measure **contrast in the same run**.
- **A token's measured contrast ratio does not travel into a tinted box.** Worked
  example from this very site on 08-10: `#ffc9d6` measures 4.9:1 on the bare purple and
  drops to **4.42:1** — under the floor — inside `.gr-ruling`, because that box paints
  `--gr-surface` on top. `bugfix_122_contrast_ink_slots` is the lane that owns this.
- **`deploy_image_asset` resolves its source image by PURPOSE, not by the `asset_id`
  you pass** — the second same-purpose asset on a site silently deploys as the first.
  `sha256sum` the downloads and confirm they differ; do not stop at `success:true`.
- **The `assets` table records no served path.** It is derived —
  `storage.DeployedWebPath(asset_key, purpose)`, branched through
  `storage.BrandHeadAssetPaths` for brand-head purposes.
- **Two page-head producers exist and only one honours `pages.noindex`**
  (`bugs_open/232`). A rebuild through the other path silently drops the tag — relevant
  the moment you rebuild anything under `/tools/gauntlet/`.
- **`apply_section_edit` does NOT republish `js_content`** — use the assemble-only
  rerender (`gauntlet_dead_cta/scripts/rerender_page_vonc.sh <page_id>`), and note it
  maintains `pages.deployed_at`/`build_status` where the edit path does not.

## 5. What the owner has actually said about this site's voice

Not design instructions as such, but they bound the aesthetic and they are recent and
strongly held:

- **Plain over clever.** 2026-08-11, rejecting the site's own provocation prose:
  *"almost unreadable … readable by a 5 year old or something like that … Cut out the
  long words perhaps use ASD-STE100."*
- **No specialist register.** 2026-08-10, rejecting a headline: *"noone knows what
  scales means - that's a techie term."*
- **British English**, spelling and idiom (platform convention; the generator now states
  it explicitly because nothing else did).
- A design that reads as consultancy-brochure default is a **named failure mode** in the
  platform's own mission statement (`business-strategy.md`), not a neutral fallback.

**Inference, not instruction `[UNVERIFIED]`:** a site whose product is blunt, plain
argument probably wants a design with the same quality. Do not treat that as a brief —
get the owner to say what he wants it to feel like, because nothing in the record does.

## 6. Suggested first moves for the fresh session

1. **Look at the site before reading more about it.** `curl` is not enough for a design
   judgement — render it. The lane has a Playwright venv (`~/.venvs/vonc_pw/bin/python`)
   and `round_record/drive_round_record.py` is a worked example of driving a real
   browser, taking a screenshot and measuring computed styles.
2. **Ask the owner what "better" means to him** before dispatching anything. He has
   given a voice, not a look, and the two are not the same. This is the one genuinely
   blocking unknown.
3. **Read `design-composition.md` DES-003, DES-005, DES-012** and establish which agents
   are live rather than superseded — then drive those, per §1.
4. **Do the two mechanical wins in §2** (`needs_rebuild` arena page, the never-built
   provocation blog page) — they need no aesthetic decision and one of them is nine days
   stale.
5. **Open the standing five for the lane** if this becomes a workstream of its own
   (`PLAN` / `RUNBOOK` / `NOTES` / `README_where_we_are` / `SUMMARY`), per CLAUDE.md.

## 7. What is NOT part of this, and who owns it

- **The provocation pipeline** (content, gate, generator, readability) is a different
  lane and is in good order: cold start is
  `provocation_pipeline/HANDOFF_2026-08-10b_the_generator_works.md`. Site content through
  **2026-08-22** is booked. **Do not regenerate or re-date provocations from a design
  session.**
- **Owed to the council** (`65d153f0`, REVISE): extract `ai_actions.go:351-372`'s
  option-building so it and `llmOptionsFromConfig` share one implementation. Touches 127
  live steps across 55 agents; the seat that raised it asked that a human require it
  before merge. **Deliberately not rushed into a roll. It wants its own round, not a
  design session.**
- **RFC_020 §5.2** (`namecheck`) is built, council-approved and **not live** — it ships
  from the island VM, and the guardian seat asked the owning lane to schedule that,
  not us.
