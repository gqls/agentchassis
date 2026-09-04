# HANDOFF — Imagery best-in-class workstream (start a new chat from here)

**Last updated: 2026-09-04 (imagery lane, resumed). UPDATE THIS DOCUMENT EVERY
WORKING TURN, alongside the running notes — it is the single entry point for a fresh
session.**

## ★ START HERE (2026-09-04) — the lane is IDLE and CLEAN; nothing is blocked, nothing is half-done

**Read this block, then `RUNNING_NOTES` 2026-09-04, then pick from the list at the
bottom. The 2026-07-21 block below is SUPERSEDED — kept for the trail, not for
instructions.** Its landing gates all passed on 2026-07-25 (027 and 011 both closed);
its "next concrete step" is done; its `bugs_open/020` hold was lifted on 2026-07-24.

**What happened while this lane was idle (2026-07-25 → 2026-09-04), by other threads:**

- **2026-08-11, `bugfix_214_imagery_scope_ref`** — `WriteSitePlanAction` now
  canonicalises `site_plan_imagery.scope_ref` at write time (`IMG-070`, live
  `v1.0.1283`). Their note is in `RUNNING_NOTES` under that date. **One item is
  explicitly left as this lane's call and is still OPEN:** whether to re-key, cancel
  or leave `needs_imagery` items whose ItemKey embeds a pre-fix ref when their site
  replans. Nothing is stuck today; the decision is owed before a site with renamed
  pages replans.
- **2026-08-24, `bugfix_382_empty_kind_routing`** — an absent `kind` now routes to
  Banana and raises `MISSING_IMAGE_KIND` (`da21ae20f`, live `v1.0.1334`); migration
  `586` gave `call_variant_gen` its `kind?` and `site_id`. **Consequence for this
  lane's own machinery: hero VARIANTS now reach `getImageryStyleGuideForSite` for the
  first time**, so per-kind overrides, `avoid` terms and reference anchors apply to
  them. **[UNMEASURED] whether variant output actually changed** — nobody has looked,
  and that is arguably the most interesting unclaimed item here.

**What this session did (2026-09-04) — measurement only, no code and no live config
changed:**

The residual 382 left on this lane — *"4 steps with no `kind`, reachability UNMEASURED
beyond 1 day"* — is **answered**. Full working in `RUNNING_NOTES` 2026-09-04 §1–2.

- The four steps still exist `[MEASURED 2026-09-04]`: `pageflow-builder` and
  `site-work-orchestrator`, each with `call_logo_generation` + `generate_hero_image`.
  `image-build-handler`'s four all carry `kind?`.
- **Use `assets.origin_model`, not `orchestration_states`, to ask this question.** It
  is durable; the orchestration table is still a 1-day window. Until 2026-08-24 a
  kind-less request went to Stability, so an SDXL row IS a footprint of this path.
- **SDXL generation ceased 2026-08-11 and has not resumed in 24 days**, against a
  demand control of **1,025** generated assets on **39** sites `[MEASURED 2026-09-04]`
  (~~1,046 on 37~~ — **CORRECTED same day**: I summed weekly display buckets including
  the one containing the last SDXL row, and took the site count from a window bounded at
  the 08-24 roll not the 08-11 stop. Conclusion unchanged; correction block at the end of
  `RUNNING_NOTES`, and `WRONG_CALLS.md`).
  **0** of the **14** sites with an `imagery_style_guide` pin stability, so none of the
  16 SDXL assets was sanctioned.
- **The 382 fix has therefore never fired in production** — the traffic stopped on
  08-11 (migration `390`), thirteen days before the fix rolled on 08-24.

**Two traps this turned up, both now written down:**

1. **`agent_error_log` rows EXPIRE** (resolved >14d, unresolved >30d — `database-cleanup`
   arm 1, `sql_for_agents/465`). **`bugs_closed/011`'s live-fire proof row of
   2026-07-24 20:45:57Z is GONE**; the table's oldest survivor is 23:30:20Z the same
   day. The bug file, this lane's `RUNNING_NOTES` and the auto-memory all still cite
   it. Closure sound, evidence unre-runnable. Landmine appended (footprint
   `agent_error_log`), verifier dispatched, corr `8c5d0f5f`.
2. **Migration `390` is applied but UNRECORDED** in `schema_migrations` — a clean hole
   between `389` and `391`, both applied 2026-08-11. Its effect is live. A runner pass
   would list it as pending; re-applying looks harmless (an `UPDATE` plus a post-state
   assertion that already holds). **[UNVERIFIED]** why the row was never written.

**Open items, in the order I would take them:**

| # | item | state |
|---|---|---|
| 1 | **Did hero-variant output change after `586`?** Variants now see the style guide for the first time. Compare variants generated before/after 2026-08-24 on a site with per-kind overrides. | **[UNMEASURED]**, cheap, and it is this lane's own D14 machinery reaching new ground |
| 2 | **214's re-key question** — `needs_imagery` ItemKeys minted from pre-fix `scope_ref`s | OPEN, explicitly this lane's call, not yet urgent |
| 3 | **F3 remaining card surfaces.** `info-card-grid` had the widest reach — **15 pages / 7 sites as of 2026-07-19, STALE, re-census before quoting**. `featured_article` and `product-card-with-cta` were on ZERO live pages; do not start there. | needs a design call from the owner |
| 4 | `features_open/022` (rendered-text legibility guard / OCR) and `023` (infographic figures from the evidence base) | spun out of 011, unstarted. **NOTE: an active `infographics` lane opened 2026-09-04 and owns the code-rendered-graphic route — talk to it before touching 023** |
| 5 | The leopardess gibberish SDXL hero on `how-it-works.html` | still live; owner was offered the regeneration on 2026-07-25 and has not called it. Another lane's client site |

**Ownership check before you route anything:** `scripts/who-owns.py <n>` — the two
imagery-numbered bugs are owned elsewhere (`114` lane active 2026-09-03, `214` lane
closed 2026-08-11). `bugs_open/384` (a landed card image never invalidates its
listing) is imagery-adjacent and **actively owned** — leave it.

---

> **SUPERSEDED 2026-09-04 — the block below is the 2026-07-21 state. Its gates passed,
> its hold was lifted, and its "next step" is done. Kept for the trail.**

## ★ START HERE (2026-07-21) — v1.0.1144 is live; the palette-truncation fix is IN IT

**Two code fixes that were "inert until a roll" are now LIVE in `v1.0.1144`**, both
verified against the running pod (`agent-chassis-59c675c4f-pxr9f`) by log-message
literal, not symbol grep (A6.3):
- **`bugs_open/027` §4b — palette truncation** (`1191cecdb`): the WARN literal
  `Imagery direction TRUNCATED before generation` is present in the binary (control
  literal also present, so the marker is valid). Palette now composes FIRST; the
  backoff is LastIndex; `datahelpers.SafeCut` is the shared rune-safe cut.
- **`bugs_open/028` — avoid lists inert** (other thread, live since v1.0.1140): every
  site's `avoid` list is now genuinely sent to Banana. Both fixes now shape every
  generated image's prompt.

**Because the fix is live, three things are now TRUE, not pending:**
1. **The `027` landing gate is now ACTIONABLE** (was blocked on the roll). Next
   concrete step for a fresh thread: **regenerate robot-hands' 3 ARTICLE content
   heroes and re-check against the D13 criteria** (distinct, on-style, click-through
   matching, cards ≤60KB). Its **3 TOOL content heroes are EXCLUDED — they wait out
   the `bugs_open/020` owner hold** (the STOP block below). Verify the new code fired
   by the WARN literal in the pod logs, not a symbol grep. Use A6.5 to supersede +
   A6.1 to sweep; nothing has run on robot-hands since the 08:47Z deploy, so the gate
   is untouched.
2. **The base-voice flip is now LIVE and unmitigated.** All four sites' BASE voices
   (304–398 chars) now truncate to **colours-without-prose** instead of
   prose-without-colours. Correct by design, but any hero/illustration/infographic
   regenerated on those sites will now drop its medium/mood prose — the WARN names
   which. **Config remedy (live-immediate, no roll): shorten the base-guide palette
   glosses** — hex codes need no prose. Do it as a backed-up needle-gate migration
   (pattern: `SQL_2026-07-19_*`). robot-hands' base guide is the owner's; the other
   three I authored are mine to shorten.
3. **`027` and `028` themselves stay OPEN** until fixed-AND-live is *proven on a page*
   (the /bugs_closed/ bar). The code is live; the proof (a regenerated, on-style,
   on-palette set) is the landing gate above. `027` shipped on a **REVISE, no
   `Council-Reviewed:` trailer** — 7 rounds, 11 approve / 2 object; the two open
   objections are recorded as follow-ups in the bug file, not silently dropped.

**Cross-thread, NOT imagery (parked, owner-driven):** `bounce.leopardess.uk` SES
custom-MAIL-FROM is failing because its **MX record does not exist at the
authoritative nameservers** (dns1/dns2.uk-noc.com). Owner must add, at the Krystal/
uk-noc DNS panel: `MX 10 feedback-smtp.eu-west-2.amazonses.com` and `TXT "v=spf1
include:amazonses.com ~all"` on `bounce.leopardess.uk`, then retry MAIL FROM in the
SES console (eu-west-2). A persistent monitor is watching the authoritative NS and
will report the moment the record lands. Region value is load-bearing.

---

> **SIBLING THREAD, 2026-07-20 — provider routing changed underneath this document.**
> `bugs_open/011` R1 is **live on v1.0.1139**: the adapter's hand-maintained
> kind→provider `switch` is now an enumerable table
> (`internal/adapters/imagegenerator/routing.go`), **`hero` has joined Banana** (so
> *every* declared kind is now Banana-routed; SDXL is reached only by an empty/legacy
> kind or an explicit per-site `provider:"stability"`), and a site can now pin its
> provider from `imagery_style_guide.provider` as **config, not code**. An unrouted
> kind is now detected and logged by name instead of falling silently to SDXL.
> **Resume point and full detail: `HANDOFF_2026-07-20_provider_routing_011.md`.**
> Two consequences land on this workstream: it completed `bugs_open/028`'s blast
> radius (`avoid` is inert on heroes too — that thread has shipped a fix, inert until
> a roll), and it made `bugs_open/027` §4b's 200-char cap due, since the cap is sized
> for SDXL's CLIP wall and no declared kind uses SDXL any more.

## WHERE WE ARE (2026-07-19) — read this block, then "Mechanisms" below

**The D13 gate failure is resolved on the imagery side. Nothing is blocked, and
no fix chat is outstanding.** Phases **I0, I1, I2 and I3 are complete and live**
on **v1.0.1136**. The two things left in I3 need *your decisions*, not code
(RUNBOOK **B16**).

**What "live" means, verified on robot-hands.com 2026-07-19:**
- `learning-center-hub.html` — 3 listed articles, 3 distinct on-style cards,
  every click-through 200 and showing its own hero.
- `index.html` tool directory — 3 tool cards with matching imagery.
- Card files 22–36KB against the ≤60KB budget (the failed D13 set ran 37–73KB).

**The three decisions that produced this (D14 + F2.1 + F3):**
1. **D14 — content heroes are their own KIND, routed to Banana, in flat duotone
   illustration.** The D13 styles were inconsistent because content heroes were
   emitted as `kind='hero'`, which routes to Stability/SDXL — and *that path
   structurally ignores `ReferenceImageURIs`*, so per-site style anchoring was
   impossible. `content_hero` now routes to Banana, and the style guide takes a
   per-kind **override map** (`kinds.<kind>`) that replaces
   direction/avoid/anchors **wholesale** for its kind (partial merging would let
   the site's photographic base voice contaminate a flat-illustration kind).
2. **F2.1 — listed-page eligibility.** One shared constant,
   `queryresolve.ListedPageEligibilitySQL`, governs BOTH the listing and the
   imagery sweep so they cannot drift; it removed the six 404 links.
3. **F3 — the check now iterates a SURFACE TABLE**
   (`contentImageSurfaces` in `check_content_image_missing.go`) instead of
   hardcoding `blog-post`. Each entry carries page type, the consumer LIKE that
   proves the site lists that type, the eligibility predicate, and the prompt's
   subject noun. **Adding a surface is a data change.** `tool` joined it and
   differs in all four fields.

**Commits:** `4e35c8064` (D14 + F2.1) · `358e14af6` (council fixes) ·
`8b804bc27` (F3 surface table) · migration `170` (tool-list image slot, applied)
· `c0ef457a1` (release the adapter with the chassis).

## ✅ LIFTED 2026-07-24 — tool-imagery hold RELEASED (owner instruction)

> **The hold below is LIFTED.** `/bugs_open/020` is **CLOSED** — fixed, live on chassis
> **v1.0.1150**, and induced-fault-proven (case file: `/bugs_closed/020_HANDOFF_2026-07-18_*`;
> workstream: `docs024_key_docs_latest/bug020_tool_recreation_data_integrity/`). The
> tool-recreation path now carries a live prompt contract against inventing data AND a
> mechanical gate (`check_tool_fabrication`, wired via migration 189) that holds any
> fabricated recreation at `needs_human_review` instead of publishing it — proven by a
> live induced-fault test.
> **Tool imagery may resume:** publish / derive / re-render tool-page imagery and fire the
> finetuning.uk / leopardessconsulting.co.uk tool sweeps.
> **Two things worth a look first (neither blocks):** (1) gamesdesign's 9 stored
> `content_hero` assets — 4 violate their own `avoid` list (see `/bugs_open/028`); (2)
> whether gamesdesign's game calculators ever invented data was never verified (020's
> failure mode is *data-backed* tools; game calculators are formula-based, so likely
> clean, but the check was never run).
>
> _The original 2026-07-20 hold rationale is preserved below, for the record._

## 🛑 (HISTORICAL — hold LIFTED 2026-07-24) tool imagery was held by owner ruling until `/bugs_open/020` was fixed (2026-07-20)

**~~Do not publish, derive, re-render or extend tool-page imagery. Do not fire the
tool sweeps on finetuning.uk or leopardessconsulting.co.uk.~~** _(Hold lifted — see the
banner above.)_ This was an owner instruction, not a technical block — the machinery works.

**Why:** `bugs_open/020` is the tool-recreation path inventing datasets and reporting
`complete` (live fabrication reached a public site on vetcomparison). Everything this
rollout dresses is a **tool page**, and **gamesdesign.co.uk has 11 completed
`needs_tool_recreation` items** — its nine tool pages came through that exact handler.
Publishing card imagery would promote them in the tool directory and make them look
more credible. Whether gamesdesign's specific tools invented anything is **UNVERIFIED**
— 020's failure mode is data-backed tools, and game calculators are formula-based — but
that check has not been done, and the hold does not depend on it.

**State at the hold (nothing to roll back):** 9 `content_hero` assets generated and
stored on gamesdesign. **Not live** — the deployed tool pages serve 200 with zero hero
references. Cards were never derived; the listing was never re-rendered. 7 of the 9 are
good; 4 violate their own `avoid` list (3 white/pale grounds, 1 with numerals) for the
reason in `/bugs_open/028`.

**Safe to continue meanwhile** (not tool-imagery): the `/bugs_open/027` §4b and
`/bugs_open/028` code fixes, both of which want the council gate.

---

**⚠️ 2026-07-19 (Turn 53) — the tool rollout is HELD at the no-spend point, and the
style-guide fix alone is NOT sufficient. See `/bugs_open/027` §4b.**

**State right now:** 9 `needs_imagery` items on gamesdesign.co.uk sit at `detected`
(nothing spends until promoted — that IS the control point). 2 were promoted as a
pilot and **failed**: correct ground colour, invented accent, inconsistent with each
other, text in both. **Root cause is NOT the missing style guide** (that was fixed):
`composeDirection` puts the palette LAST and the direction is capped at 200 chars
(`maxImageryDirectionInPrompt`), so a verbose medium+mood silently truncates the
colour instruction away. **robot-hands composes to 233 chars too** — it only looks
right because its cut lands after its accent colour. Config mitigation applied to the
three sites (139–147 chars, accent first); robot-hands deliberately NOT touched.
Re-pilot of the 2 was queued and had not completed at handoff — **check it before
promoting the other 7.** Owner decisions taken this turn: **B16.3 = write the guides
and run** (done); **B16.1 = leave the card grid as emoji cards, CLOSED** (its chosen
image source turned out not to exist — 0 of 86 cards' destinations have a card asset,
41 of 72 links resolve to no page).

**Unverified and next:** whether the Banana path sends `avoid` as a negative prompt at
all. Both pilot images carried lettering despite `avoid` listing it AND the positive
prompt forbidding it. If it does not, every `avoid` list in the fleet is inert for all
Banana-routed (flat) kinds.

Two live-verified findings changed the position below:
- **`content_hero` is unstyled on every site but robot-hands.** D14 added the kind
  to the style guide's override map but NOT to `directionAppliesToKind`, so a site
  with no `kinds.content_hero` override falls back to the free-text
  `design_intent.imagery_direction` — written for photography, handed to a
  flat-illustration kind on Banana. **Only robot-hands.com has an
  `imagery_style_guide` row at all**, so every other site generates in an
  unspecified style: the F1 class that failed the D13 gate. It reads as done
  because robot-hands, the only site exercising the kind, has the override —
  one branch of a two-branch router again.
- **The rollout figures below were wrong.** Real exposure is **19 generations
  across 3 sites** (gamesdesign 9, finetuning 5, leopardess 5), not 10 across 2:
  finetuning.uk and leopardessconsulting.co.uk both pass the tool-list consumer
  gate and were missed, and **idea.uk spends nothing** (its one tool page has
  `deployed_at IS NULL`). Get this from the check's gate query, not a survey.
- **Armed, not on fire:** `scheduled_tasks` has no discovery entry (passes are
  fired by hand), but the check is registered on `design-discovery-agent` by
  `type`, so *any* session's routine sweep of those three sites trips it.
  Containment if one runs: emitted items sit at `detected` and do not spend until
  triaged — delete or cancel them pre-triage at no cost.

**What is left in I3 — RUNBOOK B16, all three need the owner:**
- **B16.1 `info-card-grid`** — the most-deployed listing we have (15 live pages,
  7 sites), but NOT query-fed and with **no image slot at all**. Whether it
  should carry imagery, and from what source, is a design call. Do not just build it.
- **B16.2** — the I5/I6 volume sign-off (B5). ~33 tool-page generations were
  funded 2026-07-18; **7 spent**. Pending: **19 across gamesdesign.co.uk (9),
  finetuning.uk (5) and leopardessconsulting.co.uk (5)** — corrected above.
- **B16.3 (NEW)** — hold / style / accept those 19, given none of the three sites
  has a style guide. Recommendation: **hold**, write the three style guides
  (config, live immediately, no image roll), then let it drain. The fleet-wide code
  half — give `content_hero` a defined default instead of a per-site lottery —
  is 027 §5(a) and wants the council gate.

**Do NOT start with `featured_article` or `product-card-with-cta`** — both are on
**zero live pages fleet-wide** (the older fix handoff suggested otherwise and is
corrected). `news-listing` is Phase I5's own scope.

**Two corrections a fresh session must not re-inherit:**
- **The "stale adapter" incident did not happen.** I reported the
  image-generator-adapter shipped stale at v1.0.1134 and "proved" it three ways.
  All three used the same invalid marker: `content_hero`/`sprite_sheet` are not
  retained by the Dockerfile build (`-a -installsuffix cgo`, alpine) though a host
  `go build` keeps them. **A pod-grep is a POSITIVE test only** — validate the
  marker against a known-good build before believing a miss, and prefer
  log-message strings. Full evidence + control recipe: `016b` §9; RUNBOOK **A6.3**.
- **A page re-render does NOT pick up a component-template change** unless
  `spec.reason` ∈ (`image_landed`, `section_data_resolved`, `cta_links_stale`);
  otherwise it reassembles the stored `page_components.rendered_html` and reports
  success having changed nothing. RUNBOOK **A6.2**.

**Where to read next:** `SUMMARY_2026-07-19_imagery_i3_card_imagery.md` (plain
prose, current state) · `RUNNING_NOTES_…` Turns 48–52 (the technical log,
including every wrong turn) · `RUNBOOK_…` §A6 (the commands that cost a cycle,
each with its gotcha) · `README_where_we_are.md` (the owner's running log).

---

<details>
<summary>Superseded state, kept for the trail — the 2026-07-17 gate failure and the 2026-07-16 position</summary>


## ⚠️ 2026-07-17 (Turn 47): THE D13 GATE FAILED — work moved to TWO FIX HANDOFFS
The D13 machinery ran end-to-end and is PROVEN (9 heroes generated → 9 cards
auto-re-derived → live on the hub), but the user's gate failed on image
quality/consistency AND surfaced site-level damage. **Start fix chats from:**
- **Imagery fixes:** `HANDOFF_2026-07-17_i3_imagery_gate_fixes.md` (style
  redesigned for the small format, blog_posts eligibility filter, remaining
  card surfaces).
- **Site fixes (not imagery):** `../robot_hands/HANDOFF_2026-07-17_robot_hands_site_fixes.md`
  (blue-brochure regression — NOTE the B7 layout FK is INTACT, damage is
  component-level; nav sprawl; 404 article rows; broken tools ↔ experience_loop;
  dead Load More).
This workstream pauses I3 polish until those land. The state below is
accurate as of the gate.

## WHERE WE ARE (2026-07-16, Turn 46) — start here
- **I0, I1, I2 are ✅ COMPLETE AND LIVE** (incl. B12: served CSS default is the
  ARROW, self-healed on the v1.0.1125 pass). Read-out for a status briefing:
  `READOUT_2026-07-16_imagery_status.md`.
- **I3 mechanism acceptance MET LIVE on v1.0.1125 (Turn 46):** both checks fired
  in one discovery pass; 9 cards derived + entity-linked + committed;
  `query.blog_posts` resolved 9 articles with per-article images; the served
  `learning-center-hub.html` shows all 9 `card-*.jpg` (HTTP 200).
- **D13 (user 2026-07-16): per-article content-hero GENERATION is BUILT, rides
  the next deploy.** All 9 first-run cards were byte-identical (every blog-post
  fell back to `hero_canonical` — planner emits no article heroes). The check is
  now a two-mode emitter (generate content hero via image-build-handler's generic
  needs_imagery path / derive-or-re-derive the card on ORIGIN-STALENESS), and the
  preference order plan-hero → content-hero (`ContentHeroKey`) → site-hero is
  unified across check, deriver, and page renderer. Also riding: card q78 (first
  run hit 64,097B > the ≤60KB budget at q82). **Post-deploy the fleet converges
  itself:** pass 1 generates ~9 article heroes (SDXL — B5 budget note), pass 2
  re-derives the 9 cards from them, pass 3 silent. Then the A3 gate: 9 visually
  DISTINCT cards.
- **Dispatch priority is ASC — lower number = sooner** (`ORDER BY wi.priority ASC`;
  "30 // high" in check comments). The old "needs_page@99" habit meant "run LAST".
  Front-of-queue nudge for a watched run = set priority 5.
- **The image-landing trap is CLOSED and the content-loss thread has RECOVERED
  ALL 17 article-body instances** (004 updated by that thread; root cause was
  writer max_tokens truncation) — D13's image landings are safe fleet-wide.

</details>

## ✅ READ FIRST — image-landing trap CLOSED (guard live); residual notes
**History:** landing an image fired a scoped re-render (`image_landed`) that BLANKED
the article body on any page whose content was a never-parsed JSON envelope. The
**escalate-to-writer guard is now LIVE in prod** (re-verified in the running pod this
turn — the 004 handoff's own criterion), so a scoped re-render on such a page now
escalates to the writer instead of blanking. **Residuals to respect:**
- **The testbed (robot-hands.com) is safe** — its article-body pages are healthy and
  every imagery row is fulfilled, so no image landing will even fire. Proceed with I3
  here.
- **Do not assume every OTHER site is safe to land images on yet:** the writer-side
  envelope repair (`ParseLLMJSON`) still fails on some fixtures, so an escalation may
  partially- or not-regenerate. And **4 pages are still blanked** (finetuning×3,
  gamesdesign×1) awaiting recovery in the separate thread.
- **Detection-query fix (found this turn):** the 004/006 `length=1326` test now
  UNDER-reports — I2.5's `sprite-bullets` class made a fresh blank shell 1341 bytes.
  Use `rendered_html ~ 'article-body__content[^>]*></div>'` instead. 004 §5 corrected.
Full write-up: **`../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`**
Root cause + recovery: **`../HANDOFF_2026-07-14_article_body_json_envelope.md`**

## Turn 38–39: I2.2 + I2.3 are DONE, LIVE AND USER-GATED
Sprite bullets render **four distinct glyphs** (ⓘ info / ✓ check / gauge / ⚠
warning) on a real list — `/guides/tool-grip-force-friction-calculator-guide.html`,
the "Safety Factor Selection Guidance" list. The Turn 36 CSS-specificity bug is
fixed, deployed (verified by grepping the running POD's binary, not git), and
sprites.css re-emitted with scoped overrides. One 75,745B sheet; ≤80KB budget met.
**Hard-refresh before eyeballing — sprites.css is cached `max-age=3600`.**

## ✅ I2 IS COMPLETE (Turn 41)
I2.0–I2.5 all done and live. Sprite sheet, per-page glyph bullets (`ul.sprite-list`),
the `sprite_css_missing` fulfilment check (proven idempotent on the live fleet), and
the D10 **container house-style** (`.sprite-bullets` on the article-body wrapper —
content lists theme themselves with no markup). Proven on the friction-calculator
guide: an unclassed LLM-content list now shows sprite glyphs, the Safety Factor list
keeps its explicit info/gauge/warning, and the old JSON leak on that page is gone.

**Default glyph = arrow (user chose it 2026-07-15).** The container default is a
neutral `arrow`, not `check`; check is explicit-only (`sprite-b-check`). Implemented
in `buildSpriteCSS` (const `spriteDefaultBulletGlyph`), `SpriteCSSFormat` bumped 2→3.
It rides the next deploy and self-heals: `sprite_css_missing` sees the format
mismatch, re-emits, re-stamps. **On the next deploy, verify the gate page's content
lists flipped check→arrow** (CSS-only, no page edit) — that doubles as live proof of
the format-version half of I2.4.

**Next phase: I3** (content-linked card imagery / Lane B). See the ordered actions
at the bottom. First, though, the highest-value open item is NOT imagery — it's the
article-body content loss (below), which this workstream's image-landings trigger.

## What this project is (fresh-reader paragraph)

agentchassis is an autonomous agent platform that plans, builds, and operates
a fleet of content websites (tools, guides, games, news feeds, articles).
One generic Go runtime executes declarative workflows stored as JSONB in
Postgres (`agent_definitions`); agents cooperate over Kafka; all work flows
through `site_work_items` (discovery checks find problems → dispatcher →
handler agents → git-backed deploy). The **imagery best-in-class workstream**
(started 2026-07-08) raises fleet visual quality: brand consistency,
data-accurate infographics, card imagery, sprite-sheet bullets, product
sketches, news imagery, performance budgets, audit loop.
**Testbed: robot-hands.com**, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.
Working branch: `084_site_improvements_local_ai` (user makes bulk commits
that sweep in-progress edits — check `git log` before assuming anything is
uncommitted; forward-only, never reset).

## Document map (all in this directory)
1. **THIS FILE** — state + mechanisms + next actions. Start here.
2. `READOUT_2026-07-16_imagery_status.md` — spoken-word status briefing (what
   we've done / where we are / where we're going). For reading aloud.
3. `PLAN_imagery_best_in_class.md` — goals G1–G9, user-confirmed decisions
   D1–D10, phases I0–I8 with dated status blocks. The map.
4. `RUNNING_NOTES_imagery_best_in_class.md` — Turns 1–43, every diagnosis
   with evidence. Append a turn each session.
5. `RUNBOOK_imagery_best_in_class.md` — the human's task queue (A-rituals,
   B-items). Done: B1–B3, B6–B8, B10–B12. Open: B4 (data-source key, at I4),
   B5 (budget sign-off), B9 (reaper cadence), B12 (verify arrow default after
   deploy), B13 (content-loss fixes — separate threads).
6. `SCOPE_I2_sprite_sheets.md` — I2 implementation scope (png→jpg revised).
7. `SHOWCASE_*.md` (3 files) — shareable summaries (one-pager / narrative /
   technical w/ diagrams). Refresh stats before reuse.
8. `SQL_2026-07-*.sql` — every migration/seed run, each with backup+verify.

## State of the world (historical detail — CURRENT state is at the top of this doc)

**As of 2026-07-16: I0, I1, I2 are all ✅ COMPLETE AND LIVE.** The blocks below
are the detailed record (kept for the hard-won bug-stacks); read the top of the
doc for the current one-screen state. Phase I3 is next; the highest-priority OPEN
item is the article-body content-loss fix, spun out to its own thread (see the
READ FIRST warning and the Spun-out section).

**Phase I0 (testbed rebuild + render acceptance): ✅ COMPLETE.**
33-page rebuild w/ live news (9 sources, latest-news on index); 16 distinct
per-page git-path heroes, zero expiring URLs; layout = tool-portal-dark
(B7: classification industry_tags fix + 025-pattern `css_themes.layout_id`
swap — there is NO runtime re-compose path, by design); corrupted-template
class self-healing (bridge check live; 10/14 healed, 4 remaining fix
themselves on other sites' discovery passes).

**Phase I1 (brand consistency): ✅ COMPLETE, LIVE-VERIFIED on served HTML.**
- `imagery_style_guide` site-spec drives every generation; per-kind gating
  (photographic kinds: medium+mood+palette; icon/sprite_sheet: palette only;
  logo: nothing) — PROVEN on real generations (icons carried palette, not
  medium). `avoid` → negative prompt; `reference_asset_keys` → stable s3://
  anchors for Banana.
- Logo: user-approved, LOCKED (`assets.locked_at`, lock_type=permanent);
  store-guard refuses overwrites (D5). Header resolves it from plan imagery
  (`logo-img` live). Favicon + OG card DERIVED from the logo
  (`derive_brand_head_assets`; favicon.png/og-card.png serve 200; og:image +
  twitter:card injected into every head at render time).

**Phase I2 (sprite-sheet bullets): ✅ COMPLETE AND LIVE (2026-07-15/16).** The
detail below traces how it got there (I2.0→I2.5); the top-of-doc summary is the
current state. Remaining loose end: the arrow-default swap is committed but not
yet live — self-heals on the next discovery pass after the format-3 deploy.
- Decisions locked: first surface = LIST BULLETS; 3×3 vocabulary = check,
  gauge, gripper, cog, chart, download, arrow, info, warning.
- Delivery (verified twice): separate committed `/assets/css/sprites.css` +
  head `<link>` — css_snippets is a GLOBAL library (no site_id), and the
  per-site committed bundle is the house pattern (cf. /assets/js/snippets.js).
- I2.0 ✅ (live): chk_kind + validImageryKinds include `sprite_sheet`;
  adapter routes it → Banana; ImagePurposes `sprite_sheet` = **768×768 JPG
  q88** (revised from PNG — see the three-bug stack below); insert-gate passed.
- I2.1 ✅ DONE + GATE PASSED (Turn 34): the sheet is LIVE at
  `/assets/images/sprite-sheet-main.jpg` (768×768 JPEG, 75,745 bytes, under
  the 80KB budget), the glyphs are perfect, and the USER CONFIRMED the cell
  map at the B11 gate. `style_hints.cell_names_verified=true` written into the
  plan row. **Getting a clean deploy took a THREE-BUG STACK, all now fixed —
  the generation itself was flawless first try (cell-alignment risk never
  materialised):**
    1. purpose→hero via ExtractActionInputs' aggressive recursive search →
       fixed workflow-only (`SQL_2026-07-12_asset_deployer_explicit_paths.sql`,
       explicit Strategy-0 `input_data.*` paths on deploy_asset). LIVE.
    2. re-drive left attempt_count capped (3/3 → item excluded from
       find_dispatchable_site → sits triaged; looked like dead dispatch, was
       correctly idle) + a state-machine completion race re-stamped a reset
       item. Fixed by resetting attempt_count=0.
    3. lossless 768² PNG exceeded the Kafka git-commit message-size limit
       ("Message Size Too Large") AND the ≤80KB budget → switched to JPG q88
       (Go, live). SCOPE_I2 revised png→jpg.
- I2.2 ✅ LIVE (deployed v1.0.1114; sprites.css serves 200, 1,711B. NOTE: the
  per-item override selectors it emits are specificity-broken until the next
  deploy carries the Turn 36 fix — see READ FIRST): `emit_sprite_css` — pure CSS
  background-position slicing from the verified grid at bullet size (T=20px →
  sheet drawn 60×60, cells at reading-order offsets). Emits `.sprite` base +
  `.sprite-<name>` (inline/icon/nav) + themed `ul.sprite-list li::before`
  bullets (default glyph = cell 0 = check) + per-item `li.sprite-b-<name>`
  overrides. Commits `/assets/css/sprites.css` (base64 via sendGitCommitRequest).
  GUARDED on cell_names_verified=true. Geometry unit-tested.
- I2.3 ✅ LIVE (head `<link>` verified on served HTML; one real list wired —
  Turn 36): `injectBrandHeadTags` now adds
  `<link rel=stylesheet href=/assets/css/sprites.css>`, GUARDED on an active
  sprite_sheet asset existing. asset-deployer gained a `sprite_css` mode
  (migration `SQL_2026-07-13_asset_deployer_sprite_css_mode.sql`, LIVE):
  check_mode → check_sprite_mode → emit_sprite_css_step. Reusable fleet-wide
  via a `needs_sprite_css` work item (spec.mode='sprite_css').
- **FINISH SEQUENCE (after the next chassis deploy):** dispatch a
  `needs_sprite_css` item (handler asset-deployer, spec {mode:'sprite_css'},
  status triaged, attempt_count 0) → sprites.css commits → rerender-pages with
  refresh_site_components=true (head gets the link) → wire
  `class="sprite-list"` onto ONE robot-hands section's `<ul>` (section-editor
  or a targeted content edit) → LIVE GATE (bullets readable at ~20px, one
  ≤80KB download). Then I2.4: fulfilment discovery check (sheet planned but no
  asset → needs_imagery; asset but no sprites.css → needs_sprite_css); later
  per-section sheets / vision auto-verify.
- Latent gaps recorded (parked): asset-deployer's explicit deploy paths cover
  `input_data.*` (image-build-handler shape) but NOT `input_data.spec.*`
  (dispatch shape) — fix if the undeployed_asset flow misbehaves; historical
  spawned deploys may have silently used hero dims (check the May icons'
  file dims someday); a stale 900² `sprite-sheet-main.jpg` was overwritten by
  the correct one (no longer an issue).

**Phases I3–I8: not started.** I3 (card imagery / Lane B: generic
entity_type+entity_id on assets — user-confirmed D2) is next after I2.
I4 charts = go-echarts in-chassis, per-domain free data sources. All
decisions D1–D8 user-confirmed (see PLAN §4/§8).

## Mechanisms a fresh session must know (hard-won)

- **DB access:** `kubectl exec -n ai-persona-system postgres-clients-0 --
  psql -U clients_user -d clients_db`. A PreToolUse hook auto-approves
  read-only SELECT/\d through this exact form; mutations prompt the user.
  Auth expires ~daily → runbook A1 (user re-login), symptoms: "server has
  asked for the client to provide credentials" on EVERY call.
- **Manual agent trigger** (`system.intake` is STALE — do not use): kcat pod
  → `system.agent.generic.requests`, header `action=orchestrate`, body
  `{"headers":{...},"config":{"agent_type":"<type>"},"input_data":{...}}`.
  Full working example: notes Turn 18; script precedent
  `docs/agent_docs/sql_for_agents/033_rerender_pages_trigger.sh`. Known-good
  for: improvement-loop (full discovery cycle), webdesign-agent (CSS
  re-render+deploy), rerender-pages (+refresh_site_components:true).
  **Do NOT hand-roll spawn_agent+call_agent inline workflows** — the spawned
  child runs its workflow on INIT and idles before your call arrives
  (Turn 26). Route work to spawned handlers via WORK ITEMS + dispatch.
- **Dispatch input contract** (handlers invoked by build-dispatch-loop
  receive): `input_data.spec` (the item's spec), `input_data.site_id`,
  `input_data.domain`, `input_data.item_type`. Write step conditions against
  these (e.g. asset-deployer's check_mode: `input_data.spec.mode ==
  "brand_head" OR input_data.mode == "brand_head"`).
- **Manually-inserted work items are NOT auto-triaged** — insert with
  status='triaged', triaged_at=now(). Dedup: partial unique (site_id,
  item_key) over non-terminal statuses. `site_plan_imagery.chk_source`
  allows only llm|classifier|manual|adoption → seeds use 'manual'.
- **RE-DRIVING a work item (Turn 32 lesson):** ALWAYS reset
  `attempt_count=0` alongside `status='triaged'` and clearing claim metadata.
  At `attempt_count>=max_attempts` the item is EXCLUDED from
  find_dispatchable_site → sits triaged forever, and if it's the site's only
  candidate the trigger idles (looks like dead dispatch; it isn't). Also
  beware a just-finished orchestration's tail re-stamping a freshly-reset
  item back to complete (state-machine race) — verify no in-flight
  orchestrations for the item before/after resetting.
- **Zombie claims:** a claimed item stuck >~10 min blocks its ENTIRE site
  from dispatch. Standing unstick:
  `UPDATE site_work_items SET status='triaged', claimed_by=NULL,
  claimed_at=NULL, updated_at=now() WHERE status='claimed' AND claimed_at <
  now() - interval '10 minutes';` Real fix = B9 (user's TODO 6/10/11).
- **Page assembly reads `page_components.rendered_html` DIRECTLY** (Turn 36,
  `rerender_single_page_action.go:383` getPageSections) — it does NOT re-render
  from `content_data` + template. To change a section's markup you must write
  `rendered_html` (the artifact that deploys); write `content_data` too so
  source and artifact stay consistent. Re-render by inserting a `page_rerender`
  work item shaped exactly like CreateRerenderItemsAction
  (`create_rerender_items_action.go:192`: handler 'page-rerender', priority 80,
  spec {page_id,page_name,filename,domain}; filename = url minus leading '/').
  Omit `spec.reason` ⇒ unscoped ⇒ plain assemble (no section regeneration).
- **A CSS emitter needs a SPECIFICITY assertion, not just a geometry one**
  (Turn 36): emit_sprite_css's offsets were all correct and unit-tested, yet
  every bullet rendered the default glyph because the override selector was less
  specific than the default rule. Geometry tests can't see a cascade loss —
  only a live render (or a specificity test) can.
- **Component regeneration:** queue `needs_component_regeneration` →
  component-creator; spec {function, component_id, section_type, ...,
  description}. `description` renders into the creator prompt — use it to
  DEMAND exact schema-field preservation if the pre-store guard rejects.
- **Deploys:** user-driven (git → GitHub Actions → docker → k8s). Go changes
  are inert until then. Post-deploy ritual: clear zombies; re-trigger any
  loop the restart interrupted (improvement-loop items strand in 'detected'
  if triage hadn't run — re-trigger the loop, don't hand-promote dozens).
- **Paths/conventions:** deployed asset path = `storage.DeployedWebPath(
  asset_key, purpose)` (NEVER assets.url — expiring presigned);
  `imageryplan.ImageRoleForPath` = shared image-role aliases; kind→provider
  routing in `dynamic_adapter.go` (flat kinds → Banana, photographic →
  SDXL); per-kind prompt gating in `directionAppliesToKind` +
  `styleGuide.directionForKind` — ANY NEW KIND must be added to BOTH gating
  functions AND chk_kind AND validImageryKinds AND (usually) the adapter
  switch + ImagePurposes. That five-place checklist is the I2.0 lesson.
- **git-commit path has a Kafka message-size limit** (Turn 33): images are
  base64'd into a Kafka message. A lossless PNG of a detailed image can
  exceed the broker max ("Message Size Too Large") — prefer JPG for anything
  visually dense, which also serves the ≤80KB budgets. Text files (CSS/JS)
  commit fine as base64.
- **Committing a per-site TEXT file** (sprites.css, snippets.js): action
  returns/commits a `files` map `{path: {content: base64(text),
  encoding: "base64"}}` via `sendGitCommitRequest`; head `<link>`/`<script>`
  is injected by render_site_components. No storage client needed (DB + Kafka
  producer only).
- **B11-style human gates on generated imagery:** agents can Read PNG/JPG
  files directly — DOWNLOAD the asset (curl its assets.url, which is the live
  presigned S3 URL) and inspect it yourself BEFORE asking the user, so the
  gate question is precise. Give the user the deployed web URL to open.

## Open threads (parked, non-blocking)
- 4 corrupted components remain (archetype-taster-quiz, lobby-grid,
  provocation-card, tool-cta) — self-heal on their sites' next discovery
  passes; forceable via improvement-loop trigger per site.
- Kafka per-job response-topic partition race — transient; now surfaces as
  failed items (mark_item_failed fix) instead of silent successes.
- No runtime re-compose path (layout changes = 025 FK-swap pattern);
  build a site-design-planner re-resolve mode if this becomes routine.
- learning-center-index orphan component row (pre-rebuild residue) — clears
  with I3 card imagery or a listing rebuild.
- Old orphan pages (how-it-works, selection-guide, learning-center sprawl)
  not in the current plan — cleanup pass someday.
- image_source_unsatisfiable check live but has produced 0 flags (heroes all
  resolve now) — expected.

## Next actions, in order (updated 2026-07-19 — I3 CLOSED bar two owner decisions)

**Nothing here is blocking and nothing needs a fix chat.** In priority order:

1. **Ask the owner B16.1 and B16.3** (should `info-card-grid` carry imagery, and
   from what source? and: hold / style / accept the 19 pending tool generations?).
   B16.1 is a fleet-wide visual change across 7 sites; B16.3 is real spend in an
   unspecified style — `/bugs_open/027`.
2. **Do NOT let the funded tool rollout drain until B16.3 is answered.** When it
   is: gamesdesign.co.uk (9), finetuning.uk (5) and leopardessconsulting.co.uk (5)
   emit on their next discovery passes, 10/site/pass cap. (idea.uk emits nothing.)
   To watch or nudge one: RUNBOOK **A6.1** (fire a pass) + **A6.4** (promote
   items stranded in `detected` AND `unresolved`). After the cards land, the
   listing needs a re-render with `reason='image_landed'` — **A6.2**, this is the
   step that bites.
3. **Then Phase I4** (data graphics — go-echarts, real series; needs RUNBOOK B4
   data-source key), or extend Lane B to news (I5) / products (I6) on the same
   entity columns. Both are sized in the PLAN.

Dispatch reminders: priority is ASC (5 = front, 99 = LAST) and orders only
*within* a site — dispatch is one site at a time against a fleet-wide pool, so a
`triaged` item waiting ten minutes is usually queued, not broken. Clear zombie
claims >15 min, but never while a `needs_page` build is in flight (they run 20+).

<details>
<summary>Superseded 2026-07-16 sequence (D13 deploy-gated), kept for the trail</summary>

Turn 45's B14 sequence RAN and passed (mechanism acceptance met live). What's
staged behind the NEXT deploy is D13 (per-article generation) + q78 cards:

1. **After the deploy, trigger a discovery pass for robot-hands.com**
   (improvement-loop kcat pattern, notes Turn 18) or let the loop cycle.
   `content_image_missing` now emits ~9 **needs_imagery GENERATION** items
   (content heroes, SDXL — expect real API spend, B5) — watch them complete;
   each landing also re-renders its article page via flag_rebuild (safe: guard
   live, all article bodies healthy fleet-wide).
2. **Second discovery pass** (or next loop cycle): the check sees each card's
   `origin_asset_id` ≠ its new content hero → re-emits 9 DERIVE items → cards
   re-cut at q78 from the per-article heroes. Third pass: silent (convergence
   proof, same shape as I2.4's fire-once-then-quiet).
3. **Re-render `learning-center-hub`** (needs_page, priority 5 — sections must
   RE-RESOLVE; assemble-only won't refresh the items). Precedent item shape:
   Turn 46 used item_key `needs_page:learning-center-hub:i3_cards`.
4. **A3 gate (user eyeball):** `/learning-center-hub.html` shows **9 visually
   DISTINCT per-article cards**; click-through shows the same image family on
   the article page (its content hero); `card-*.jpg` ≤60KB now (q78). The
   `learning-center-index` orphan slot clears with a listing rebuild.
5. **Then Phase I4** (data graphics — go-echarts, real series; needs RUNBOOK B4
   data-source key) or extend Lane B to news (I5) / products (I6) on the same
   entity columns.
Dispatch reminders: priority is ASC (5 = front, 99 = LAST); clear zombie claims
>10 min (they block the whole site); watched runs may need both.

</details>

**How I2 closed (2026-07-15, for the record):** to land I2.5's `sprite-bullets`
class the article-body wrapper had to exist, but robot-hands had no healthy article
page. I repaired ONE page (the friction-calculator gate page, which was also a
JSON-leak page) by extracting its article from the unparsed envelope into
`content_data.content`, added the class to the GLOBAL article-body template, and
did an assemble-only re-render. Result proven live: an unclassed content list themes
itself; the Safety Factor list keeps its explicit glyphs; the leak is gone. The
container opt-in (`.sprite-bullets ul`) exists because content `<ul>`s never carry
classes (LLM HTML into `{{.content}}`), so a per-list class is a hand-edit that
regeneration would wipe.

## Spun out of this workstream — NOT imagery, and the current top priority
Three pre-existing platform bugs surfaced during I2, all the same shape (a
background process reporting success while silently failing / losing content). Each
has a full handoff; the user is driving the article-body/image-landing pair in a
separate chat. **Do not chase these here — but heed the image-landing hazard.**
- **Image landing blanks the article body** (the one you're fixing separately):
  landing an image fires a scoped re-render that renders the never-parsed
  article-body envelope EMPTY and overwrites the good HTML → 9 pages already blanked,
  4 more leaking, across 5 sites. **STANDING HAZARD for THIS workstream: our image
  landings are the trigger — do not land an image on an affected page until the
  guard deploys.** → `../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`
  and root cause `../HANDOFF_2026-07-14_article_body_json_envelope.md`.
- **Product pages ship empty** and the `empty_section` fix-loop marked them
  `complete` without filling them (a loop closing without fixing). →
  `../HANDOFF_2026-07-14_empty_product_sections.md`.
