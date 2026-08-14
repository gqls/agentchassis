# PLAN 2026-08-14 — durable in-body imagery for generated guide/blog articles

**Status: design, nothing implemented.** No DB writes, no code edits, no commits made by
this planning session. Everything marked `[MEASURED]` was verified against the live DB or
the tree on 2026-08-14 by this session; the query or file:line is inline.

**The ask (owner, 2026-08-13):** guide/blog articles should carry explanatory imagery
inside the article body — between paragraphs and beside them — not just the header image.
Motivating case: `dartsonline.com/blog/flight-shapes.html`.

**The problem as correctly framed by the brief:** capability exists; **durability** does
not. In-body `<figure>` markup lives inside `article-body`'s single `content` field
(`source: "llm"`), so any wholesale content regeneration silently destroys it. The
flight-shapes figure died 90ms into a body rewrite on 2026-08-05 and nothing reported it
for nine days (`bugs_open/114`, foot correction of 2026-08-14).

---

## 0. Corrections to the brief (recorded per the working-docs rules)

The brief was right about the mechanism, the class (238/226/229, not 114), and the crux.
Four factual points have moved or were incomplete:

1. **Blast radius is 93 instances / 18 sites, not 91 / 17.** `[MEASURED]`
   ```sql
   SELECT s.domain, count(pc.id) FROM page_components pc
   JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE pc.component_id='5835b2e1-50d7-4f20-8a9c-8da4d270ae3d' GROUP BY 1;
   ```
   18 rows; full site list in §5. The tree moves daily — re-run before acting.

2. **The damage is wider than the one flight-shapes figure. Three MORE guides lost their
   2026-08-05 figures and are still missing them today:** `beginners`,
   `steel-tip-vs-soft-tip`, `tungsten-guide` — all with `article-body` last rewritten
   2026-08-09, no `<img` in `content_data` or `rendered_html` now. `[MEASURED]` (census
   query in §7.3). The 08-05 lane's README says all 8 built guides were illustrated and
   live-verified that day; today 4 of those 8 carry their figure (`barrel-weight`,
   `board-setup`, `brand-comparison`, `shaft-length`), 1 is restored-but-not-live
   (flight-shapes, next point), 3 are gone. All three are recoverable:
   `page_component_history` holds archived `content` containing `<img` for each
   (latest with-img rows: beginners 08-09 10:00Z, steel-tip 08-05 11:57Z, tungsten
   08-05 11:05Z `[MEASURED]`).

3. **The brief's own stopgap restore is NOT serving.** The guarded UPDATE restored the
   figure into `content_data` only (6706 chars, `<img` present `[MEASURED]`), but the
   stored `rendered_html` does not contain it, and the served page's only `<img>` is the
   site logo (`curl … | grep -o '<img[^>]*>'` → `logo.png` alone; control: barrel-weight
   serves 2 `[MEASURED]`). The rerender item filed to publish it
   (`page_rerender_flight-shapes_…_section_data_resolved`, plus work item
   `aaabecaa-b69c-472d-8a3a-e004d1215b05`) is still `triaged`, undrained `[MEASURED]` —
   the drain gap the brief itself flagged (`detected`/queued items do not drain where
   `improvement-sweep` is disabled). So the one-page stopgap is currently a DB-only
   restore. Phase 0 finishes it.

4. **Why `divergence` was NULL on the 08-05 rows — two reasons, not one.** (a) The rows
   predate the archive trigger (loss 08-05; trigger live on v1.0.1276, 2026-08-09,
   `bugs_open/229` header). (b) They are the wrong row *family*: `source =
   'save_page_sections_overwrite'` rows are the save path's own pre-DELETE `content_data`
   snapshot and never carry `divergence`; the stamped rows are the trigger's separate
   `source='artefact_archive_trigger'` rows. `[MEASURED]`:
   ```
   SELECT source, op, divergence, count(*) FROM page_component_history GROUP BY 1,2,3;
   -- save_page_sections_overwrite · NULL · NULL · 16708
   -- artefact_archive_trigger · delete/overwrite · machine_made|unstamped|hand_patched · 2173
   ```
   Consequence for §4 (detection): the hook exists but does not cover this loss class.

Minor: `grip-styles` — "the one unbuilt guide" in the 08-05 log — has since been built
(body written 08-12, `content_hero_grip_styles` asset created 08-11 `[MEASURED]`), so the
site now has 9 guide pages, none figure-illustrated at birth.

---

## 1. Prior art (the job is mostly wiring, and I can show it)

**The brief's hypothesis is confirmed: section-scope plan imagery + the `ensureAssets`
kind-alias resolver (IMG-056 "Edit B") is live, in production code, and is the intended
durable home for exactly this.** `[MEASURED]` at the tree, not just the register:

- `site_plan_imagery` (`scope='section'`, `scope_ref '<page>:<ordinal>'` enforced by
  `chk_scope_ref_consistency`; kinds incl. `illustration`/`icon`/`infographic`; `source
  IN (llm|classifier|manual|adoption)`; `locked_at`/`locked_by` with lock transfer across
  replans matched on `(scope, scope_ref, key)` — IMG-012/013). Schema read live (`\d`),
  148 section-scope rows exist fleet-wide `[MEASURED]`.
- `ensureAssets`, `platform/orchestration/actions/plan_sections_action.go:389-563`. The
  Edit B block at **:475-518** queries section-scope rows for the current page
  (`spi.scope='section' AND spi.scope_ref LIKE $page || ':%' AND spi.kind IN
  ('illustration','icon','infographic')`), joins active `assets`, resolves each to
  `storage.DeployedWebPath(asset_key, purpose)` — the stable deployed path, **not** the
  expiring presigned URL — and maps it **by key** and **by kind first-wins alias**.
- Component fields consume it via `source: "site_assets.<x>"`
  (`plan_sections_action.go:652`, image-role alias fallback at :658-666, IMG-058's
  `ImageRoleForPath`).
- **Both render paths re-resolve through the same code.** The build path resolves via
  `planSection`; the re-render path *reuses* `planSection` ("side-effect-free",
  `rerender_page_sections_action.go:398`) and merges `stored content_data ⊕ fresh
  resolved_data`, resolved-last (:482-496, :526). Since 2026-08-11 the build path also
  carries non-llm keys forward from the page's own deployed row (PBP-039;
  `LANDMINES.md` "save REPLACES / rerender MERGES" entry), and the 2026-08-14 commit
  `8f899cc8d` (bug 268) extended that carry to renderer/static fields.
- The guide pages are eligible: all 9 `dartsonline.com` `/blog/*` pages are plan pages
  (`in_current_plan = t`, `page_type='blog-post'`, `rebuild_policy='generic'`
  `[MEASURED]`), and dartsonline's current plan already carries 7 *section*-scope icon
  rows (`about:2`, `index:2`) — the shape is exercised on this exact site `[MEASURED]`.
- **Every needed image already exists as an active asset.** The 08-05 illustration set
  (`illustration_flight_shapes_comparison_lg`, `illustration_board_setup_diagram_lg`,
  `illustration_brass_vs_tungsten_barrel_lg`, `content_hero_*_grip_texture` etc., 32
  guide-related active assets, `origin_model` mostly `banana/gemini-3-pro-image-preview`)
  `[MEASURED]`. Slice 1 needs **zero** new generations.

**What does NOT exist** (checked, not assumed): no figure/diagram component anywhere in
`content_components` (`name/function ILIKE '%figure%'|'%diagram%'` → 0 rows
`[MEASURED]`); no prior bug or plan for in-body article imagery (`grep -ril
'article.figure|inline.figure|in-body imag|inline imag' bugs_open/ bugs_closed/` → only
114's own correction `[MEASURED]`); `article-body`'s schema is the single `content` llm
field, template is `{{.content}}` in a styled wrapper `[MEASURED]`. The register's
covers-through-2026-07-13 freeze was heeded: everything above was re-verified at code or
DB, not taken from the register.

So the durable home exists end-to-end: **plan row → asset → resolver → render, all
outside the regenerable blob.** The missing 10% is that `article-body` has no field that
looks at the resolver, and no mechanism can place an image *inside* the prose flow.

---

## 2. The crux — a representation that survives regeneration

The failure to design against, stated precisely: `save_page_sections` **DELETE+INSERTs**
a page's agent-writable rows; the writer legitimately owns the `content` field and
rewrites it wholesale; anything living only inside that field dies with the rewrite
(`bugs_open/238` class). `rerender_page_sections` MERGES and cannot lose a key — so a
representation is durable iff it lives in something the regeneration path *re-derives*
(plan + assets + resolver) rather than something it *replaces* (the llm blob).

### Options weighed

| | survives wholesale rewrite? | mid-prose placement? | constrains the writer? |
|---|---|---|---|
| **(a) figure as its own page section** | yes (plan-owned) | **no** — `pages.sections` is a flat list; `article-body` is one section holding the whole article, so a sibling section sits only above/below it | none |
| **(b) plan-declared section-scope imagery, resolver-sourced fields on the component** | **yes** — URL is resolver-owned, re-resolved on every build and rerender | coarse: where the template puts it (floated aside beside opening prose, or after the content block) | none — writer never sees the URL; it may caption via optional llm fields |
| **(c) marker convention in prose, resolved by a post-pass** | only if every rewrite prompt re-supplies the marker instruction — a prompt convention, i.e. **a comment, not a control** (CLAUDE.md 2026-08-02 §2's exact language) | **yes** — precise | yes — writer must emit/preserve markers |
| **(d) make the regenerator merge, not replace** | n/a — merge of two versions of *prose* is ill-defined; the writer legitimately restructures articles (238's card rewrite is the worked proof: old figures no longer mapped onto new content). Also `bugs_open/178` is a standing stop sign against new floors in `save_page_sections`, and PBP-039 already merges everything merge *can* protect (non-llm keys) | — | — |

### Recommendation: **(b) now, (b)+(c) as the full mechanism** — plan-as-truth, with an optional writer-marker refinement whose fallback never depends on the writer

- **The plan row is the durable representation.** One `site_plan_imagery` row per in-body
  figure: `scope='section'`, `scope_ref='<page>:<article-body-ordinal>'`,
  `kind='illustration'|'infographic'`, `key=<asset stem>`, `source='manual'` initially,
  **`locked_at` set** so replans carry it (IMG-013 lock transfer). It survives content
  regeneration *by construction* — different table, different writer.
- **The component consumes it through the resolver**, exactly as heroes do: optional
  fields sourced `site_assets.illustration` (kind alias, IMG-056) resolve to the deployed
  path at every build and every rerender. A rewrite replaces the prose; the next render
  re-attaches the figure. Loss becomes a render-path bug rather than a data race.
- **Caption/alt are optional llm fields** beside it — the writer captions the figure in
  the article's own voice and re-captions on rewrite (that is *good* churn), guided by
  `llm_guidance`; the URL is never the writer's to lose. They must be `required: false`
  — a missing **required** llm field makes the `section_data_resolved` rerender pre-check
  escalate the whole page to the content-writer (`LANDMINES.md`, leopardess entry).
- **Phase 3 adds mid-prose placement** without changing the durability story: a
  placement declaration on the plan row (`style_hints.placement`, e.g.
  `{"mode":"inline","after":"h2:2","float":"right"}`) drives a small Go splice helper at
  render time that injects the resolved `<figure>` into the rendered content — at a
  writer-emitted marker (`<!--figure:<key>-->`) when present, at the deterministic
  fallback position (`after h2:2` etc.) when not. The writer is *invited* to place
  markers (plan_sections already emits per-field `LLMFieldSpecs` the writer prompt
  iterates — `plan_sections_action.go:885-891, 2064-2275` — so the planned-figure list
  can travel to the writer the same way), but **nothing depends on its cooperation**: a
  rewrite that drops the marker degrades placement, never presence. This is why (c) is
  safe as a refinement and unsafe as the foundation.
- Stored `content_data.content` stays the writer's pure prose; the splice exists only in
  `rendered_html` (digest-stamped machine_made by the save/rerender paths, consistent
  with STY-056). No new stored state to drift.

**Why not (a):** cannot answer the ask (mid-prose/beside-prose); and repointing section
membership walks straight into the `pages.sections`-cache / section-resurrection
landmine family for zero placement gain.

**Why the naive "nicer figure component in the blob" is rejected** (the brief's own
point, endorsed after verification): it rebuilds the 90ms trap with better styling —
nothing in `save_page_sections` distinguishes figure markup from prose, and per
`bugs_open/178` it should not learn to.

---

## 3. Bug 274 bearing (checked as directed)

`bugs_open/274`: ~15,000 completed child workflows across 60 agent types failing to
deliver results to parents since 08-03, including `page-rerender` (4,794) and
`asset-deployer` (439) `[MEASURED in the bug file]`. Bearing on this design:

- Slice 1 **avoids the exposed surface**: no new image generation (assets exist), no
  child-result dependency — the resolver reads tables, not workflow payloads.
- Rerenders persist their work server-side before the parent handoff (the failure is
  result *delivery*, and rerender writes `page_components` directly), so a drained
  rerender still lands. But any Phase-4+ automation that *chains* on a child's returned
  payload (e.g. "generate figure, then rerender with its key from the child result")
  must be work-item-mediated (asset lands → `flag_page_image_rebuild` emits the rerender
  item, IMG-051), never return-value-mediated, until 274 is fixed.

---

## 4. Detection — how the next loss becomes visible

Silence is what cost nine days. Three layers, ordered by what they actually cover:

1. **Recovery is already solved — do not rebuild it.** Since v1.0.1276 every destruction
   of a page component's artefact archives (`artefact_archive_trigger`, fail-closed,
   migration 357; 2,173 stamped rows already `[MEASURED]`). The lost figures were
   recoverable in one query. Nothing to build here.

2. **`divergence` is assessed and is NOT the hook for this class.** It classifies the
   *destroyed* row: `hand_patched` destruction WARNs and files an item (proven e2e in
   229's protocol); `machine_made`-over-`machine_made` — which is what a figure-carrying
   machine save being replaced by a figure-less machine save is — stays silent by
   design, and must (16,708 routine overwrites `[MEASURED]`; flagging them all is
   noise). It *does* usefully cover the interim: the hand-restored flight-shapes row is
   digest-unstamped, so its next destruction archives visibly. Keep it as the
   hand-patch tripwire it was built to be.

3. **The real hook: a plan-vs-rendered discovery check** — the one absence nothing
   currently tests. `check_unfulfilled_imagery_plan` (IMG-015) fires when a plan row has
   **no asset**; `image_url_404` (reworked, `bugs_closed/128`) fires when a rendered
   path has **no asset**; nothing fires when **plan row + active asset exist and the
   section's `rendered_html` does not contain the asset's `DeployedWebPath`** — bug
   114's blind spot, restated at render level. New check
   (`check_unrendered_section_imagery`, same `DiscoveryCheck` family, no LLM): for each
   current-plan section-scope row of an inline-declared kind with an active asset,
   assert the deployed path appears in that page's stored section `rendered_html`;
   emit a `page_rerender` item (`reason='image_landed'` — the merging, LLM-free path)
   with the component id so the rerender scopes (`create_rerender_items_action.go:219`).
   With the §2 design this check should never fire except on a genuine render-path
   regression — it is the tripwire that proves the guarantee, not the mechanism.
   **Stated dependency:** discovery has no recurring driver (`bugs_open/230`;
   `improvement-sweep` disabled, and its cap skips busy sites even when enabled —
   `LANDMINES.md`). The check is only as live as its driver; until 230 lands, the
   RUNBOOK for this workstream must carry the hand-fire
   (`run_improvement_sweep_once.sh` / direct check dispatch) as an explicit step, not
   an assumption.

---

## 5. Blast radius — named, per CLAUDE.md 2026-07-29 §3

`article-body` consumers (93 instances, 18 sites `[MEASURED]`, query in §0.1):
**finetuning.uk (19), ai-agent-orchestration.com (12), dartsonline.com (11),
leopardessconsulting.co.uk (11), fundamentallyai.com (8), gamesdesign.co.uk (7),
robot-hands.com (5), mortgagecalculator.co.uk (3), vonc.com (3), vetcomparison.uk (3),
relojistas.com (2), loancash.co.uk (2), gaswholesalers.com (2), noted.co.uk (1),
oufe.com (1), lendzy.co.uk (1), cookly.uk (1), webdesign.uk (1).**

Consumers to *tell* (what changed about their guarantee, not a key list):

- **Phase 1 tells nobody's guarantee anything** — it touches only dartsonline-scoped
  rows and a new component row nothing else names.
- **Phase 2** (shared `article-body` gains optional resolver fields): the message to the
  17 other sites' lanes is "your article sections can now carry a plan-declared figure;
  absent a locked section-scope `illustration` plan row, rendering is byte-identical."
- **Phase 3** (render splice): the message to the render path's consumers — the
  `brochure_component_library` lane (component library owner), the 238/268 carry lane
  (whose `content_rewrite` canary `canary_268_beginners` is CLAIMED on this site right
  now `[MEASURED]`), and the `dartsonline_traffic` lane — is "rendered article HTML may
  contain figure markup that is not in `content_data.content`, gated on a
  `style_hints.placement` declaration that zero rows currently carry."

**The opt-in field, concretely (CLAUDE.md 2026-08-02 §2):** `style_hints.placement` on
the `site_plan_imagery` row. The unsafe side (render-path content mutation) is reachable
only when a row declares it; no row does — `SELECT count(*) FROM site_plan_imagery WHERE
style_hints ? 'placement' OR constraints ? 'placement'` → **0** `[MEASURED 2026-08-14]`
(the enumeration RFC_022 demands, run, not asserted). Phase 2's component fields are
independently default-OFF: optional, `{{if}}`-guarded, `missingkey=zero` renders them
away when unresolved.

**Freeze compliance:** commit `75ebe4229` freezes `platform/colour` + `palette_*` to the
`bugfix_122` lane. This plan touches neither; the figure template uses only existing CSS
tokens (`--color-surface`, `--color-primary`) already present in `article-body`'s style
block. Stated explicitly because contrast items are open on the same site.

---

## 6. Architecture-scope call: **council gate, no RFC** — with one register obligation

Against CLAUDE.md's narrowings, in order:

- **2026-07-29 §1 (guarantee-changing vs additive-and-inert):** nothing here changes
  what a shared mechanism guarantees to existing consumers. The resolver already
  resolves section-scope imagery (IMG-056 shipped it); the component fields are
  optional and unresolvable on 17 sites (no plan rows → `{{if}}` suppresses); the
  splice is unreachable without a declaration nothing makes. Additive-and-inert.
- **RFC_022 (all three conditions, each measured):** (1) opt-in —
  `style_hints.placement` / optional schema fields; (2) unsafe default OFF — absent
  declaration, rendering is byte-identical; (3) zero live consumers name it — the 0-row
  query above, plus 0 schema fields referencing the new names (none exist yet). Not
  architecture-scope. **And the honest counterweight the same ruling demands:** this
  adds another optional key to a shared surface, which is precisely the accumulation
  RFC_022's unbuilt counter exists to notice. The concept-register entry (below) must
  say so, so the tenth-key reviewer finds it.
- **2026-07-28 condition (2), which is the whole surviving requirement:** the splice
  helper and the placement vocabulary are a **shared seam** and get a concept-register
  entry (`register/imagery.md`, new IMG number: what it is, the landmine — "figure
  markup in `rendered_html` is deliberately absent from `content_data`" — and the open
  review question) **in the same commit that ships them**.
- Per current norm: each code phase goes through the council gate before/alongside its
  commit (`097` trigger, ~30 min budget), `Council-Submitted:` trailer if committing
  ahead of the verdict.

Phase 1 is DB config only (no `platform/` diff) — the gate refuses non-platform
submissions client-side, so Phase 1 ships on the normal commit rules; Phases 2–4 are Go +
seed changes and go through the gate.

---

## 7. Phasing

### Phase 0 — finish the stopgap, honestly (minutes, no design)
1. Drain the already-queued flight-shapes rerender (item
   `aaabecaa-b69c-472d-8a3a-e004d1215b05` / `page_rerender…section_data_resolved`,
   `triaged` `[MEASURED]`) — the queue does not drain itself on this site; use the
   proven direct path (`brochure_component_library/scripts/rerender_page_sections_direct.sh`,
   per LANDMINES). **Coordination gate first:** the brochure lane's standing warning —
   *do not rebuild any page on any site until 090 corr `b885a92e` reports* (a fleet-wide
   `{{if}}/{{end}}` copy leak). A sections rerender is LLM-free and regenerates no copy,
   but check that 090's status before dispatching anything, and say so in NOTES.
2. Do **not** hand-restore the other three lost figures the same way. The stopgap path
   (blob-splice by hand) is the anti-pattern this plan exists to retire; their archived
   markup (§0.2) becomes caption/alt seed material for Phase 1, which re-places them
   durably within days.

### Phase 1 — the minimal durable slice: dartsonline only, shared component untouched
No Go changes. No new images. One site.

1. **Fork, don't edit:** new `content_components` row `article-body-illustrated`
   (component_level/section semantics identical to `article-body`), template =
   `article-body`'s + one guarded block:
   `{{if .figure_url}}<figure class="article-figure">…<img src="{{.figure_url}}" alt="{{.figure_alt}}"…><figcaption>{{.figure_caption}}</figcaption></figure>{{end}}`
   (floated/aside styling at desktop, full-width between hero and prose on mobile —
   "beside them" delivered; "between two specific paragraphs" is Phase 3's).
   Schema: `content` unchanged; + `figure_url` (`source: "site_assets.illustration"`,
   optional), `figure_alt`, `figure_caption` (`source: "llm"`, **optional** — see the
   escalation landmine in §2), with `llm_guidance` telling the writer what the figure
   shows (from the plan row's `prompt`).
2. **Repoint the 9 guide pages** to it — in the PLAN, not just the cache:
   `site_plan_sections.component_name` (current plan), `pages.sections` entry, and
   `page_components.component_id`, consistently, because `pages.sections` is a
   materialised cache of `site_plan_sections` and a name/id disagreement is decided
   id-wins with an observe-only log (`plan_sections_action.go:1188`). This triple-write
   is the riskiest step of the slice; it gets the RUNBOOK treatment (and the loading of
   the template via psql follows the `md5()` byte-fidelity landmine, never bare
   `\set`+INSERT).
3. **Seed and LOCK the plan rows** (`source='manual'`, `locked_at=now()`,
   `locked_by='inline_guide_imagery'`): one section-scope row per guide,
   `scope_ref='<page>:2'` (article-body is position 2 on these pages `[MEASURED]`),
   `key` = the existing asset (e.g. `illustration_flight_shapes_comparison_lg`),
   `kind='illustration'`. Lock transfer (IMG-013) carries them across replans.
4. **Strip the embedded figures from the 5 blobs that still carry them** (flight-shapes'
   restored one, and barrel-weight / board-setup / brand-comparison / shaft-length) in
   the same pass that seeds their plan rows — otherwise the next render shows the figure
   twice. Their `<figcaption>` text seeds `figure_caption`.
5. **Rerender the 9 pages** (sections path, LLM-free) and verify at the served bytes:
   every guide serves exactly one in-body `<figure>` whose src is the deployed asset
   path; then the acceptance test that defines the whole workstream — **fire a real
   `content_rewrite` at one guide and watch the figure survive the rewrite.**
   Coordinate with the 268 canary owner before using `beginners` for this (§5).

Value delivered at the end of the slice: all 8 owner-visible guide figures back, durable
against the exact event that destroyed 4 of them, on the motivating site, with the other
17 sites byte-identical.

### Phase 2 — promote to the shared component (council-gated)
Fold the optional fields + guarded template block into `article-body` itself; retire the
fork (repoint the 9 pages back; delete the fork row only after `git log`-able evidence
nothing else adopted it). Tell the 17 lanes (§5 message). Sites opt in per page by
seeding a locked section-scope plan row — no code, no roll. Extend the planner prompt
(IMG-014's `## Imagery Block`) so NEW guide builds can declare an in-body illustration at
plan time, behind the same kind gate — that makes the planner a second producer of these
rows, which the register entry names.

### Phase 3 — mid-prose placement (council-gated; the splice)
`style_hints.placement` vocabulary + the splice helper on both render paths
(`RenderComponentAction` + `rerender_page_sections`), marker-first,
deterministic-fallback, as §2. Writer prompt learns the planned-figure list via
`LLMFieldSpecs`. This is the phase that fully answers "between paragraphs"; it ships
with the register entry and the 0-consumer measurement re-run at commit time.

### Phase 4 — detection + generation for new content
`check_unrendered_section_imagery` (§4.3) with its drain dependency stated; and only
here, new-figure *generation* for future articles (needs_imagery already routes
section-scope rows — IMG-015/016/020 chain), with the 274 caveat (§3) and per-image
spend measured before fleet-wide enablement.

---

## 8. Honest costs and risks

- **Per-image generation spend: zero for Phases 0–3** on dartsonline — all figures exist
  as active assets `[MEASURED]`. For new articles (Phase 4): images route through
  `banana/gemini-3-pro-image-preview` (the fleet's current image model per the assets
  table `[MEASURED]`); per-image £ is `[UNMEASURED]` here — image calls do not land in
  `llm_call_log`, so measure at the provider bill before enabling any per-article
  auto-generation, and note the account hit its monthly token cap on 08-14 (LANDMINES) —
  a bad week to add spend without the owner's nod.
- **The 93 existing instances:** byte-identical through Phase 1 (fork touches 9 rows on
  one site). Phase 2's shared edit is guarded + optional; the negative control in its
  verification is a figure-less site's article page rendering byte-identical pre/post.
  Risk that survives review: a future session editing `article-body`'s template can
  still break 18 sites at once — that is today's standing exposure (LANDMINES, chrome
  `<head>` entry names the same class), not one this plan adds.
- **Placements already lost:** 3 restorable as Phase 1 captions + plan rows; their exact
  08-05 caption prose is in `page_component_history` if wanted. Nothing unrecoverable.
- **Coordination risks, live today `[MEASURED]`:** `canary_268_beginners`
  (`content_rewrite`, **claimed**) — do not race it on that page;
  `chrome_divergence_overwritten` needs_human_review on dartsonline's header (226's
  first real catch — different mechanism, same site, do not confuse the signals); six
  unresolved `misdirected_cta` rerender items on these very guide pages — a Phase-1
  rerender may satisfy or collide, check before firing; the brochure lane's
  do-not-rebuild hold (§7 Phase 0). All four go in the RUNBOOK as pre-flight checks.
- **Design risk owned:** the fork's triple-repoint (Phase 1.2) is hand-config on a
  cache-plus-plan pair with a documented resurrection landmine; it is the price of not
  touching the shared component, it is reversible, and Phase 2 deletes it.
- **What this plan does not fix:** the general 238 class (any llm-owned field carrying
  structure) — this plan removes *imagery* from the blob; prose-embedded links, tables,
  and markers stay exposed; that class belongs to the 238/268 lane, and this design
  deliberately adds nothing to `save_page_sections` (per `bugs_open/178`).
