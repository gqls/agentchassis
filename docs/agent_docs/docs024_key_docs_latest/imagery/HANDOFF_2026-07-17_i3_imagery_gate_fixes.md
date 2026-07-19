# HANDOFF — I3/D13 imagery gate FAILED: fix card imagery (start a new chat here)

> # ⛔ THIS FIX CHAT IS FINISHED — do not start a new chat here.
> **Start from `HANDOFF_imagery_best_in_class.md`** (the workstream entry point,
> current as of 2026-07-19). This document is kept for its evidence and its
> record of what failed; its "suggested fix order" and F3 ordering are **spent or
> superseded**. Everything below is history unless the banner says otherwise.
>
> **STATUS 2026-07-19 — F1 and F2 CLOSED; F3's first two surfaces LIVE.**
> The article listing and the tool directory both carry on-style card imagery on
> robot-hands (v1.0.1136). What remains of F3 is two OWNER DECISIONS, not work —
> RUNBOOK B16. See the entry-point handoff.
>
> **STATUS 2026-07-18 — F1 and F2 are CLOSED on learning-center-hub.**
> D14 (flat duotone illustration; `content_hero` as its own kind routed to Banana;
> per-kind style-guide overrides) and the F2.1 eligibility filter shipped in
> `4e35c8064`, with council-gate fixes in `358e14af6`, live and pod-verified on
> **v1.0.1135**. Live result on the hub: 3 listed articles, 3 distinct on-style
> cards at 22–26KB (budget ≤60KB), every click-through 200 and showing its own
> hero. Full account: `RUNNING_NOTES` Turns 48–49; decision text: PLAN §D14.
>
> **Two traps this fix chat hit, worth reading before you start:**
> 1. **A pod-grep marker the build does not retain reads exactly like a stale
>    deploy.** I greped `content_hero`/`sprite_sheet` on the image-generator-adapter,
>    got 0 (twice, plus an "old symbol" control that was also 0), and concluded the
>    adapter had shipped stale. **That was wrong** — the Dockerfile build
>    (`-a -installsuffix cgo`, alpine) does not retain those literals, though a host
>    `go build` does; the binary was current all along. A pod-grep is a **positive
>    test only**: a miss proves nothing until you show the marker survives a
>    known-good build. Use **log-message strings** as markers, never `case` values.
>    Full measured evidence + control recipe: `016b` §9 "A pod-grep marker that the
>    build does not retain…"; retraction in RUNNING_NOTES Turn 51.
> 2. **Ground-colour drift is fixed via `avoid`, not `medium`** — Banana put one of
>    three heroes on a white ground despite "deep charcoal ground" in the medium;
>    adding explicit light-ground terms to the override's `avoid` fixed it on re-roll.
>
> Start at **F3** below. The 6 excluded blog-post rows remain R6 (build-or-retire)
> in the sibling site handoff.

**Filed:** 2026-07-17, immediately after the user's A3 gate on the D13 card run.
**Scope: IMAGERY ONLY.** The same gate surfaced serious SITE-level defects
(theme regression, broken tools, 404 pages, nav sprawl, dead load-more) — those
are the SIBLING handoff: `../robot_hands/HANDOFF_2026-07-17_robot_hands_site_fixes.md`.
Some items interlock (marked ⇄); read both before starting either.

## What this project is (fresh-reader paragraph)

agentchassis is an autonomous agent platform operating a fleet of content
websites. One Go runtime executes declarative workflows (JSONB in
`agent_definitions`); agents cooperate over Kafka; work flows through
`site_work_items` (discovery checks → dispatch → handlers → git deploy).
The **imagery best-in-class workstream** (docs in this directory — PLAN,
RUNNING_NOTES Turns 1–47, RUNBOOK) has shipped: per-page heroes, brand layer
(locked logo/favicon/OG), sprite bullets (I0–I2), and **Phase I3 Lane B card
imagery** — entity-linked 800×450 card crops derived from per-article
"content heroes" generated from article content (decisions D11–D13).
Testbed: **robot-hands.com**, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.
DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
(read-only auto-approved; A1 re-login if credentials expired).

## The mechanism as built (all LIVE in prod v1.0.1128 — it WORKS; the output quality doesn't)

1. `content_image_missing` (discovery check, registered on
   design-discovery-agent) — two-mode emitter per listed article
   (`page_type='blog-post'`, gated on a `query.blog_posts` consumer existing):
   - no plan hero AND no content hero → **GENERATE**: `needs_imagery` item,
     `imageryplan.ContentHeroKey(page)` = `content_hero_<page_underscored>`,
     kind `hero`, prompt composed from page title+meta_description
     (`contentHeroPrompt`), routed down image-build-handler's generic path;
     the imagery style guide prepends medium/mood/palette at generation.
   - source exists but card missing OR **card.origin_asset_id ≠ current
     preferred source** → **DERIVE**: `needs_content_image` → asset-deployer
     `content_card` mode → `derive_card_asset` (cover-crop 800×450 JPG q78,
     entity-linked `entity_type='page'/entity_id`, lineage in
     `origin_asset_id`).
2. Preference order (unified in check, `derive_card_asset`,
   `plan_sections.ensureAssets`): plan page hero → content hero → site hero.
3. `queryresolve` `blog_posts` base + `image` projection feeds
   `content-listing`'s per-item `{{.image}}` (card first, plan hero fallback).
4. **Origin-staleness auto-heals**: REGENERATING a content hero automatically
   re-derives its card on the next discovery pass. This is the lever the fix
   chat should lean on — fix the hero, the card follows free.

Proven live 2026-07-17: 9 SDXL content heroes generated → 9 cards re-derived
from them (distinct sizes 37–73KB) → served learning-center-hub shows 9
distinct card refs.

## GATE VERDICT (user, 2026-07-17): FAILED on quality + consistency

### F1. Style-guide adherence failures on the generated heroes/cards
- **Not all monochrome — at least one card is in COLOUR** (user observation;
  I verified 3 of 9 as monochrome, the colour one is among the other 6).
- **Styles vary card-to-card**: verified — `card-news-post` photographic,
  `card-gripper-payload-calculator-guide` studio-photo **with a garbled
  pseudo-text/logo artefact** on the arm (despite "no text or lettering" in
  the prompt), `card-grip-force-friction-calculator-guide` came out as
  **line-art/comic illustration** (also the 73,046B budget-buster — dense
  linework compresses poorly at q78; budget is ≤60KB, D8).
- Root cause direction: SDXL adherence to the free-text style direction is
  weak, and `hero` kind routes to Stability (no reference-image anchoring —
  `ReferenceImageURIs` only flow on the **Banana** path,
  `dynamic_adapter.go` kind switch).

**USER DIRECTION for the fix: change the style to fit the limitations of the
image size we've made.** I.e. don't chase photorealism at 800×450-rendered-
small; pick a style that (a) reads at card size, (b) compresses inside 60KB,
(c) a model can hit CONSISTENTLY. Candidates to put to the user in the fix
chat: flat/duotone illustration in the site palette routed to **Banana with
the style guide's `reference_asset_keys` as anchors** (Banana honours
references; SDXL does not in our adapter), or a heavily-constrained SDXL
recipe. Consider whether content heroes should be their own KIND (e.g.
`content_hero`) so gating/routing can differ from `hero` — if so, the
**five-place kind checklist applies** (chk_kind is NOT needed — no plan rows —
but `directionAppliesToKind`, `styleGuide.directionForKind`, adapter switch,
`ImagePurposes` and the check's Row.Kind are; see PLAN D3/D12 reasoning).

### F2. Card ↔ article click-through inconsistent (verified state, 2026-07-17 ~15:50Z)
| listed article (all status=active in DB) | live URL | article hero shown |
|---|---|---|
| tool-grip-force-friction-calculator-guide | 200 `/guides/...` | its content hero ✅ MATCHES card |
| tool-gripper-payload-calculator-guide | 200 `/blog/...` | its content hero ✅ MATCHES card |
| tool-gripper-cycle-time-estimator-guide | 200 `/guides/...` | **hero-canonical ❌ MISMATCH** — its `flag_rebuild` re-render never landed (likely stranded `detected`, see Landmines) |
| grip-force-friction-calculator-guide | **404** `/blog/...` | n/a — page never deployed |
| gripper-cycle-time-estimator-guide | **404** `/blog/...` | n/a |
| gripper-payload-calculator-guide | **404** `/blog/...` | n/a |
| learning-center-article | **404** `/blog/...` | n/a — scaffold/template page |
| learning-center-post | **404** `/blog/...` | n/a — scaffold/template page |
| news-post | **404** `/blog/...` | n/a — scaffold/template page |

Two imagery-side fixes fall out:
1. **`query.blog_posts` needs an eligibility filter.** It currently lists any
   `page_type='blog-post'` row with status active/deployed — which includes 3
   scaffold/template pages (news-post, learning-center-article/-post) and 3
   never-deployed /blog/ duplicates of the /guides/ articles. Options: filter
   to pages whose file actually deployed (status='deployed'? verify what the
   9 rows' statuses mean — all read 'active'), and/or exclude scaffold pages
   (they have placeholder titles "Learning Center Article | Robot-Hands.com").
   ⇄ The pages themselves (build-or-retire the 404s, dedupe /blog/ vs
   /guides/) are the SITE handoff's item R6 — coordinate: if those rows get
   retired, the filter question shrinks.
2. **The cycle-time article needs its content-hero re-render** (one page):
   after the style fix regenerates heroes this happens anyway via
   flag_rebuild; if doing it alone, a `needs_page`@5 item (shape precedent:
   item_key `needs_page:learning-center-hub:d13_cards`, spec
   {reason,page_id,page_name}) or re-drive the stranded rerender item.

### F3. Rollout gap: only ONE surface has cards
"None of the other pages have imagery above their cards yet" (user). Current
coverage: `content-listing` on learning-center-hub ONLY. Remaining I3
surfaces (PLAN Phase I3 scope): `featured_article` (`query.featured_post` —
base NOT implemented in queryresolve), `product-card-with-cta`
(`query.affiliate_products` — NOT implemented), news-listing/Latest News Feed
(client-JS, populated from /tools/assets/*.js + data/news-archive.json — I5
territory), `info-card-grid` (no image slot in template — decide whether it
gets one), tool directory `tool-list` (has `tl-card` markup; check its image
slot). Each needs: resolver base (if query-fed) + image projection + the
component's template actually rendering `{{.image}}`.

**SURVEYED 2026-07-18 — F3 IS NOT SHOVEL-READY, and the order above is wrong.**
Live fleet usage and the real blocker per surface:

| surface | live pages | sites | state / blocker |
|---|---|---|---|
| `info-card-grid` | **15** | **7** | the most-deployed of all of them. NOT query-fed, no `<img>` at all. Needs a **design decision**: do category/info cards want imagery, and from what entity? |
| `news-listing` | 7 | 5 | client-JS populated from `/tools/assets/*.js` — **I5 territory**, not I3. |
| `tool-list` | 5 | 3 | query-fed by `pages_where_type:tool`, and **the resolver ALREADY returns `image`** (shared `pageImageProjection`). But: **0 of 38 tool pages fleet-wide have any image** — no card, no plan hero — so a template slot renders nothing. Real cost: ~38 generations + extending `content_image_missing` past `page_type='blog-post'` (its consumer gate is `query.blog_posts`-specific, so that needs per-page-type logic). **Blocked on the B5 budget call.** |
| `content-listing` | 3 | 3 | **DONE** — the reference implementation. |
| `featured_article` | **0** | 0 | **unused fleet-wide.** `query.featured_post` unimplemented. Building it is speculative work for a component on no page. |
| `product-card-with-cta` | **0** | 0 | **unused fleet-wide.** `query.affiliate_products` unimplemented; products are **I6**. |

**So there is no decision-free F3 code left.** Every remaining surface is
blocked on a budget call (tool-list), a design call (info-card-grid), another
phase (news-listing → I5, product cards → I6), or is unused (featured_article,
product-card-with-cta). Do **not** start with `featured_article` as the
original text suggests — it is on zero pages.

Also note `content_components` is a **global library** (no `site_id`;
`forked_from` for forks), so editing any of these templates is a fleet-wide
change to every site that adopts the component — not a per-site tweak.

## What NOT to redo (working, verified)
- The pipeline plumbing (check → generate → derive → resolve → render) is
  proven end-to-end; do not rebuild it — fix the STYLE and the FILTER.
- `storage.CoverCropResize`, q78 `card` purpose, entity columns, asset-deployer
  `content_card` mode, `sprite`/brand layers — all live and healthy.
- Fixing a hero auto-fixes its card (origin-staleness). Regeneration recipe:
  supersede/deactivate the content-hero asset (or bump its content) → next
  discovery pass re-generates? NO — the check only generates when NO content
  hero exists. To force regeneration: set the old content-hero asset
  `status='superseded'` → check sees no source → emits generation → landing
  re-derives card next pass. (Or re-drive the original needs_imagery item:
  reset status='triaged', attempt_count=0, priority=5.)

## Landmines (hard-won, verified this week)
- **Dispatch priority is ASC** — LOWER number = sooner; 5 = front, 99 = last.
- **Improvement-loop triage strands discovery items in `detected`** (twice
  this week). If emitted items sit `detected`, hand-promote:
  `UPDATE site_work_items SET status='triaged', triaged_at=now(), priority=5 WHERE ...`.
- **Zombie claims** block the whole site; clear claims >10 min — BUT page
  builds legitimately run 20+ min; don't auto-clear on a 10-min timer while a
  needs_page build is in flight (I did; it restart-looped the build).
- **Verify deploys by grepping the RUNNING POD's binary** (tags are rebuilt in
  place): `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<marker>"'`.
- Banana honours `ReferenceImageURIs`; the Stability path does NOT.
  Kind→provider switch: `dynamic_adapter.go` (~:534).
- Style guide: `imagery_style_guide` site-spec aspect;
  `directionForKind` gates per kind (`imagery_style_guide.go:104`);
  robot-hands' guide seeded 2026-07-10 (charcoal/electric-blue, industrial
  photography) — **the fix likely EDITS this guide** (supersede-row pattern,
  not in-place).
- The image-landing/article-blank trap is CLOSED (guard live; all 17
  article bodies recovered) — image landings are safe fleet-wide.
- Cost: generations are real API spend (SDXL today; Banana if rerouted).
  RUNBOOK B5 (budget sign-off) is still formally open; cap is 10/site/pass.

## Suggested fix order
1. Decide the card style WITH the user (small-format-first; D-decision it).
2. Implement: style-guide edit ± kind/routing change ± prompt template change
   in `contentHeroPrompt`; add a negative-prompt/`avoid` update if staying SDXL.
3. Force-regenerate the 9 content heroes (supersede old ones) → watch cards
   auto-re-derive → re-render hub + the cycle-time article.
4. Add the blog_posts eligibility filter (⇄ site handoff R6 on what survives).
5. Gate again on learning-center-hub (9 consistent, on-style, ≤60KB cards,
   all click-throughs matching on non-404 pages).
6. Then extend to the next surface (featured_article or tool-list).

## Key files/docs
- Code: `check_content_image_missing.go` (+`content_image_helpers*.go`),
  `derive_card_asset_action.go`, `imageryplan.go` (`ContentHeroKey`,
  `SpriteCSSFormat`...), `queryresolve.go`, `plan_sections_action.go`
  (ensureAssets), `imagery_style_guide.go`, `generate_image_actions.go`,
  `dynamic_adapter.go`, `url_helpers.go` (`ImagePurposes`),
  `image_processing.go` (`CoverCropResize`).
- Docs: `PLAN_imagery_best_in_class.md` (D1–D13, phase statuses),
  `RUNNING_NOTES_imagery_best_in_class.md` Turns 44–47,
  `HANDOFF_imagery_best_in_class.md` (workstream entry point),
  `RUNBOOK_imagery_best_in_class.md` (B5, B15),
  SQL artifacts `SQL_2026-07-16_*.sql` (entity link, content_card mode,
  check registration).
