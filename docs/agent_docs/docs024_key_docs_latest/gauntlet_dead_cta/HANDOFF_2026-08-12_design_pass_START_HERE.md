# HANDOFF — vonc.com design pass, requested 2026-08-12

**Owner request, verbatim:** *"Please initiate a design pass, the design of the site can
still be a lot better."*

This file is a **cold start for a fresh session**. It is deliberately a scoping document,
not a plan: the previous session had almost no context budget left when the request
landed, so what follows is what it established first-hand plus the traps it knows about.
**Nothing has been designed, changed or dispatched.**

---

## 1. THE ROUTE — this design pass runs THROUGH THE PIPELINE (owner, 2026-08-12)

> **EVERY SITE GOES THROUGH THE FRAMEWORK. Never hand-build one (OWNER RULING,
> 2026-08-04).** No hand-authored HTML uploaded to the bucket, however small, however
> temporary, however much faster it would be.

**Restated by the owner on 2026-08-12, specifically about this task:** *"The design pass
should go through the framework."* So this is not a caution attached to a design task —
**it is the shape of the task.** The deliverable is a composition re-resolved and
re-rendered by the pipeline, not a set of improvements a session made and the pipeline
happens not to have overwritten yet.

If you find yourself writing `.hero { padding: 4rem }`, stop and ask which agent should
have decided that. **"The framework cannot do it yet" is a bug to file, not a licence to
hand-roll** — the 2026-08-04 ruling exists because a session hand-built the webdesign.uk
shopfront, on the lane whose entire product is framework-built sites.

### 1.1 The live pipeline, and it is THREE stages (DES-003, `status: deployed`)

1. **`domain-research-classifier`** writes the `design_intent` spec — structured palette
   and typography `reference_values` plus `style_direction`. **This is where taste
   enters the system.** If the site should look different, this is the input that should
   differ; everything downstream is resolution and rendering.
2. **`site-design-planner`** — **deterministic, no LLM** — triggered by a
   **`needs_composition`** work item. Resolves layout (weighted scheme-aware tag match),
   typography (match-or-insert) and a site-specific palette through
   `validate_composition_inputs` → `resolve_composition_layout/typography/palette`, then
   `install_site_composition` writes `css_themes` + `style_collections` +
   `sites.style_collection_id` + a `resolved_composition` spec **in one transaction**.
   **It renders nothing.**
3. **`webdesign-agent`** — triggered by **`needs_design`**, `depends_on` composition —
   produces the LLM design overlay and renders the layout template over the installed
   base. **It is the sole writer and deployer of `styles.css`.** Merge authority is
   fixed: LLM wins core palette slots + typography, composition wins layout, structure
   tokens and specialised slots.

**⚠ A standalone planner run ends at install and emits NO `needs_design`** (DES-049).
The render is a **separate, explicit** `webdesign-agent` orchestration. A session that
re-resolves and then looks at the site will see no change and conclude the pipeline
failed.

### 1.2 ⚠ THE TRAP THAT DECIDES YOUR WHOLE APPROACH: install REFUSES to overwrite

`install_site_composition` sets `sites.style_collection_id` **only if it is NULL**, and
**errors rather than overwrites** an existing composition — *"re-resolve not supported;
clear it manually"* (DES-005).

**vonc.com already has one** [MEASURED 2026-08-12]: `style_collection_id =
e1d23bb0-8ef1-4e7a-9a25-6f4d902459fc`, with `design_intent` and `resolved_composition`
specs both `is_current`, **dated 2026-06-22**. So the design was resolved once, seven
weeks ago, and never revisited. **Firing `needs_composition` naively will fail.**

The supported path is **DES-049 — composition re-resolve procedure (gated, file-based,
backup-first)**, `status: deployed`, proven end-to-end on idea.uk 2026-06-25. Ordered SQL
**files**: backup + inspect (four uniqueness checks that must all come back 0) → gated
detach and clear (NULL the `style_collection_id`; delete the site's own
collection→theme→palette→typography chain **only where `source_site_id` matches**;
supersede the old `resolved_composition` spec) → state check → kcat re-trigger of
`site-design-planner` (`domain` is required by `ensure_site_record`) → verify.

Two caveats from that procedure are doctrine:

- **Run the SQL as FILES, never pasted.** Pasting mangled `\set` and blank lines and left
  an open transaction in a real incident.
- **Do NOT use the adoption teardown** (bulk delete by `source_domain`) — it is a
  different procedure and must not be used on a site like this.

### 1.3 ⚠ The register's own pointer to that procedure is WRONG

**DES-003's `relations` line says "DES-047 (composition re-resolve procedure)". It is
not.** DES-047 is *Computed-styles extraction via browser JS injection*, `status:
aspirational`, and its own stage-2 verification records **0 repo-wide hits** for the Go
action it describes. **The re-resolve procedure is DES-049.** (DES-030's relations line
has it right, which is how the discrepancy surfaced.) Follow DES-003's relations blindly
and you land on an unimplemented entry while looking for the one thing that unblocks you.

### 1.4 Framework-native ways to find out what is wrong — use these before your own eye

Four audit agents are **live in `agent_definitions`** [MEASURED 2026-08-12] and are the
framework's own answer to "is this design any good":

`design-audit-agent` (analyst) · `visual-design-auditor` (analyst) ·
`design-discovery-agent` (specialist) · `visual-designer` (adapter)

**Run them before forming an opinion.** A session's aesthetic judgement is exactly the
input the 2026-08-04 ruling exists to keep out of the pipeline, and an audit finding is
something the owner can be shown.

⚠ **`brand-designer` is `is_active = true` in the live table but the register marks
DES-008 superseded.** The live row wins as a statement of what *exists*; the register
wins as a statement of what is *intended*. Do not drive it without resolving which is
true — *the seed is not the system, and neither is the register.*

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

1. **Run the audit agents (§1.4) and look at the rendered site — in that order.** `curl`
   cannot settle a design question. The lane has a Playwright venv
   (`~/.venvs/vonc_pw/bin/python`) and `round_record/drive_round_record.py` is a worked
   example of driving a real browser, screenshotting, and reading **computed** styles
   (which is the only thing that answers what the page actually paints — a `var(--x,
   #fallback)` literal is in the source and never applied).
2. **Ask the owner what "better" means** before dispatching anything. He has given a
   voice, not a look, and the two are not the same. **This is the one genuinely blocking
   unknown**, and it matters more here than on most tasks because his answer belongs in
   `design_intent` (§1.1 stage 1) — i.e. it is a pipeline input, not a brief for you.
3. **Decide whether this is a re-resolve at all.** If the answer is "the palette and
   layout are right, the execution is not", that is a `webdesign-agent` re-render and
   §1.2's whole procedure is unnecessary. **Re-resolving is the expensive, destructive
   option** — it deletes and rebuilds the site's composition chain. Establish which of
   the three stages is actually wrong before touching any of them.
4. **Do the two mechanical wins in §2** (`needs_rebuild` arena page, the never-built
   provocation blog page) — they need no aesthetic decision, they go through the normal
   rebuild path, and one of them is nine days stale.
5. **Open the standing five for the lane** if this becomes a workstream of its own
   (`PLAN` / `RUNBOOK` / `NOTES` / `README_where_we_are` / `SUMMARY`), per CLAUDE.md.

**What "through the framework" rules OUT, concretely:** hand-editing `styles.css` or any
component's CSS to taste; uploading HTML to the bucket; `apply_section_edit` used as a
styling tool; and "I will just fix this one contrast value while I am here". The last is
the tempting one — `bugfix_122_contrast_ink_slots` is the lane that owns contrast, and it
fixes them at the token level so the fix reaches every page rather than the one you were
looking at.

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
