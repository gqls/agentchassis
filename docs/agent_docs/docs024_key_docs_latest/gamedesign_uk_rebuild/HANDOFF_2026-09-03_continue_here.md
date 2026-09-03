# HANDOFF — gamedesign.uk lane — continue here (2026-09-03 ~09:30Z)

**Written by the session named `gamedesign.uk`.** Supersedes `HANDOFF_2026-09-02_continue_here.md`
(kept as history — it carries three stacked update blocks including a "closeable" verdict that was
WRONG). Read this, then `SUMMARY_2026-09-03_…` for the read-aloud version, `NOTES` for evidence
(newest at bottom), `RUNBOOK` for commands (§7a and §7b are the two traps this lane paid for).
Every figure `[MEASURED]` with its time unless marked.

Lane dir: `docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/`
Site: `gamedesign.uk` = `8f17eb73-fc74-4718-8371-b3125bc4e414` · sibling `gamesdesign.co.uk` = `e33263f4-…`
Bugs this lane filed: `432` (owns fix, built), `439`, `446` (owns instance), `447` (owns instance)
Bugs reopened/contributed: `315` (reopened → AI page 3, fixed same day), `444` (5th site, 4th mechanism)

## 0. STATE RIGHT NOW (09:25Z)

| thing | state |
|---|---|
| the served site | 4 pages live, **wrong for the vertical** — the owner's review stands (§2). Sitemap `/ /about /articles/ /contact`; home/about/contact heroes all `hero-home.jpg`; articles hub still 2,148 chars with "What they avoid"; `[MEASURED 08:28Z]` |
| rebuild #2 (brief v2) | **classifier + research + strategy done 20:16–20:26Z on 09-02; chain then WITHHELD BY DESIGN** (§3). `needs_briefing` `95d834f8…` enqueued by hand **08:31:19Z, still `triaged`/unclaimed at 09:16Z** — starving behind older site backlogs (bug 413's selector mechanism; `dispatch_throughput/RUNBOOK` says hours is normal). Site unlocked, nothing claimed. **First thing to check: has it been claimed?** `SELECT status, claimed_at FROM site_work_items WHERE id='95d834f8-ed14-48a0-8459-3550bc725150';` |
| growth posture | `hold` (SEED e, 22:10Z 09-02) — first of 39 sites; **unexercised** (no `evaluate_tools`/`add_tool` filed since) |
| chassis build | **`7bf1ff674021`** (09-03, rolling 09:15Z: 91 pods new / 322 old). Carries 444's gate `6525b45ae`, WDS-020 `c2349955d`, empty-main `d777cb4d2`, 315's `8eca969cb` — all ancestors, HEAD→stamp NO. ⚠ 444's gate also needs **migration 720** applied — NOT verified (schema_migrations columns differ; check `\d schema_migrations`) |
| components' migration 721 | APPLIED 09-02 ~21:50Z (six hero components declare `background_image` typed `image`); **untested at this artefact** — needs the rebuild's rerender |
| monitors | none armed at handoff (artefact monitor timed out 09:16Z; DB monitor `boprhw0f7` may still be running in the old session — re-arm in the new one) |

## 1. What this lane is

gamedesign.uk served 13 empty pages for 4.5 months with no `sites` row (April adoption wiped and
republished the shells; row later deleted; invisible to every detector — `bugs_open/432`). The
owner ruled: rebuild through the framework, in a DIFFERENT direction to the sibling (practice
seat vs authority seat, positioning GD2). Rebuild #1 went live 09-02 18:00Z and the owner reviewed
it as **wrong for the vertical** — a games site with no games, no game imagery, an articles index
with no articles that explained its own brief, a hero over a 404. Root cause **mine**: my imagery
guide banned game imagery, my brief asked for restraint. `bugs_open/446`. Rebuild #2 is the fix.

## 2. The owner's review (09-02 ~20:30Z, verbatim) and what each point turned out to be

> "this site needs to be seen again by the checkers. please run the improvement loop over it. it
> suffers from the same problems that designblog.co.uk etc suffered with. please correspond with
> that blog to determine the best fixes. We need to change the design and copy. hero images are
> missing e.g. articles/index.html that same page shows an explanation of the brief and so on. It
> is a game design site why isn't it full of games and images and excitement -please add that to
> the errors list it is a major error"

- **checkers:** ran (corr `8b2473ab`). 27 verdicts, ALL `filing_mode=record`, "[verdict, not
  dispatched]" — they saw everything; record mode acts on nothing. 446 §4a.
- **designblog:** same critique, same day. Joined their routes (444 brief-echo; 114/721 imagery;
  theme kits / site design planner / components / copy quality two stage). Migration 718 landed
  19:59Z 09-02: planner now EXPECTS content imagery — rebuild #2 inherits it.
- **hero missing on articles/:** planner requested no site-scope hero and none for the
  section-index; template default `/assets/images/hero.jpg` → 404 on THIS domain only (other
  sites have one — controlled). 446 §3.3.
- **explains the brief:** the hub lists 0 articles because the plan made ONE article page with 0
  sections (parented nowhere, at `/blog/`); the writer wrote about the page. 444 4th mechanism.
- **games/images/excitement:** my `imagery_style_guide` banned game imagery — the hero prompts say
  "no game imagery" verbatim. 446 §3.1. **Errors list = `SITE_DEFECT_CATEGORIES.md` §10** (added:
  10.1 spec bans the vertical's subject · 10.2 index with zero members · 10.3 hero over a CSS-url
  404 · 10.3b the WRONG hero passes presence checks · 10.4 zero interactive elements on an
  interactive vertical · 10.5 no detector).
- **also found, not in the review:** about + contact wear the HOMEPAGE hero; their own generated
  heroes referenced on 0 pages. Fleet-wide: 7 components, 158 instances, 61 of 65 page heroes
  orphaned (inline guide imager, by predicate). Mechanism: site-wide `hero_url` injection, per-page
  aliasing gated on an image-TYPED field, which `about-hero`/`contact-hero` lacked. **Components
  fixed it: migration 721.** 446 §3.6/§4b.

## 3. THE TRAP THAT COST 12 HOURS — read before touching the chain

**082 on a DEPLOYED site is a strategy refresh, not a rebuild.** `domain-strategist`'s
`gate_next_item` step: `site_state.is_deployed == true → complete` — "a deployed site's strategy
refresh must NOT enqueue the briefing→site-plan rebuild chain". It COMPLETED at 20:26:45Z with no
error and no item, and I watched the artefact all night for a plan that was never going to be
written. Brief v2 DID land (classifier's `design_intent` = `bold-creative`, game imagery, gold
accent, "sensibility of a magazine"; strategy v2 current 20:26:40Z). **The chain is enqueued on
purpose** via `SEED_2026-09-03_enqueue_briefing_chain.sql` (needs_briefing in the strategist's own
shape; per-site key `briefing_gamedesign.uk` dedups only against non-terminal rows). RUNBOOK §7a.

**What follows once it is claimed:** build-briefing-agent → `needs_site_plan` → build-site-planner
(post-718) → `reconcile_site_plan` diffs the new plan against the 4 realised pages (twin risk:
`bugs_open/340`) → composition → design → pages → imagery → rerender. Budget: ~50 min once claimed
(rebuild #1 took 50 min dispatch→styled). growth_posture=hold keeps tools out.

## 4. What was HELD / CANCELLED and why (all reversible, all in SEED files)

- `SEED_2026-09-02b`: `article` slot cancelled + page archived (owner: "cancel the article slot").
- `SEED_2026-09-02d`: **the tool-suggester's plants** — 28 items cancelled, 12 `tool-*` pages
  archived (never built). The loop's `evaluate_tools` → tool-suggester (reads `identity`×8 +
  `classification`×2, NEVER the brief) filed 8 `add_tool`, six of them the SIBLING's tools by
  name, via tool-deployer. `bugs_open/447`. Positioning: GD2 now states `hosts_tools=FALSE`;
  cluster-scale instance found 09-03 (marketing siblings). Loop owner: candidate 3 refuted (the
  structure floor is a count, not a checklist — struck); real instrument = WDS-020 hold (§0);
  their question to the owner: **born `hold` rather than `open`?**
- `SEED_2026-09-02e`: `growth_posture='hold'`.

## 5. OWNER DECISIONS OWED (none block the rebuild)

1. **Contact email** — auditors ×3: `gamedesign@contactforsales.com` "signals a placeholder or
   third-party lead-capture service" to senior studio professionals. Keep or replace.
2. **An author / editorial identity** — "no named author, no studio background"; the evidence
   rules forbid inventing one. Supply, or anonymous by design.
3. **Newsletter / RSS** — no repeat-visit mechanism; the feed lane's mechanism exists, undriven.
4. **Born `hold`?** (improvement-loop owner's question, 447 §5a).
5. **447 fix ownership** — tool-suggester reading the brief; deployer sibling check. Unowned.
6. (from 09-02) `bugs_open/432` stays open until `audit-rowless-serving-domains.sh` (IMP-059) is
   scheduled; the 8 rowless domains are the adoption backlog "after this one, with oversight".

## 6. HOW TO VERIFY REBUILD #2 — AS A READER, NOT A CENSUS

Use `SITE_DEFECT_CATEGORIES.md` §10 as the checklist, and RUNBOOK §8 for the mechanics. Then:
- **Plan** (before pages): `SELECT name, role, parent_section, url FROM site_plan_pages WHERE
  plan_id=(SELECT id FROM site_plans WHERE site_id='8f17eb73…' AND is_current)` — expect N
  article-role pages **parented under the articles section**, and `site_plan_imagery` with a
  site-scope hero + per-page heroes + content imagery (718). Zero articles again = tell the 444
  session (their gate `6525b45ae` is in the build; whether 720 is applied decides if it fires).
- **Heroes** (721's first live test, owed to `components`): filename-anchored — about must carry
  `hero-about.jpg`, contact `hero-contact.jpg`, articles NOT `hero.jpg`; control `hero-home.jpg`
  on `/` only. Report the `url()`s to components right or wrong. **⚠ (components, 09-03):** an
  ASSEMBLE-mode rerender (`spec->>'reason'` = none) re-ships stored bytes and CANNOT pick up 721's
  field — 66 of 66 post-721 rerenders they found were that mode. Only a page BUILD or a rerender
  with a qualifying reason re-resolves. The rebuild BUILDS the pages, so it should bind; if
  about/contact still wear `hero-home.jpg`, read the item's reason before calling 721 failed.
  Also: never set `section_shrink_floor`/`section_component_floor` on page-rerender's step
  (fleet-wide); an agent-scoped, time-boxed override with a rollback is the only sanctioned shape.
- **Imagery temperature** (10.1): the prompts — `SELECT result->'response'->'image_result'->>'prompt'
  FROM site_work_items WHERE site_id=… AND item_type='needs_imagery' AND created_at > '2026-09-03'`
  — must NOT say "no game imagery"; must be game scenes, characters, worlds.
- **Hub** (10.2): lists real article links (exclude the hub's own self-links), no "What they
  avoid" / "What the pieces do".
- **Copy**: real games named; no sales/score/internal-decision claims (evidence_base v2 bans as
  shapes); no negative-identity headings. **Read a page aloud.**
- **Held growth items**: `SELECT * FROM site_work_items WHERE site_id=… AND item_type IN
  ('evaluate_tools','add_tool') AND created_at > '2026-09-02 22:10'` — if any, they must be
  `deferred`, handler `''` (WDS-020's record shape). Report to the loop owner.
- **The archived article page must not be re-filed** (356's class): 0 rows with
  `page_id='2ea5d983-b798-4bb2-b30a-5e3047369561'` created after 09-02 19:20Z.

## 7. Cross-lane threads (all live sessions, names exact)

`designblog.co.uk` (critique twin; routing), `bugs_open/444` (gate; has my measurements),
`components` (721; owed the url()s), `theme kits` (438 — the served palette ≠ composed palette,
recorded as the ruling working; `mission` palette is the ONLY durable seed surface and it does not
reach CSS), `Portfolio positioning` (GD2 `hosts_tools=FALSE`; 447 cluster instance), `improvement
loop` (447; born-hold question), `AI page 3` (315 owner), `gamesdesign.co.uk` (renamed to
GamesDesign.co.uk, verified at their artefact; repointed their one inbound link — which I had
broken by retracting the old tree without a census: RUNBOOK §7b), `google` (analytics — added a
GTM key; rerender rode through).

## 8. Landmines this lane hit (all in LANDMINES / WRONG_CALLS / RUNBOOK)

082-on-deployed-site is a refresh (§3) · retract without an inbound-link census (§7b) · a seeded
`design_intent` is superseded by the classifier, pinned or not; a seeded `mission` palette wins
composition and never reaches the served CSS (LANDMINES, corrected in place twice) · `<img>` count
misses CSS-background heroes; `grep hero-about` matches the CSS class not the filename · a
COMPLETED submitter is not a delivered brief (438) · an absent file measured as an empty file ·
a monitor's `none` is its own timeout · CLAUDE.md's provenance log grep is out of range on
agent-chassis — use `service_binary_capabilities` (column `service`) · the sites-repo local master
is 14k commits behind with another session's unpushed commit — use a detached worktree at
`origin/master` · `cd … && cat >` in a persisted cwd silently writes nowhere.

## 9. Commits (this lane, 09-02 → 09-03, all pathspec)

eba9c3bb6 … 749277337 (first build + close) · ad874e303/381529d5a (LANDMINES corrections) ·
f661f3cbc (sibling link miss) · 22d05a59e (446) · 830aef0e8/940a262b7 (444) · 7064e7502 (categories
§10) · a4c0791f9/f8cc139da/156b52baf (446 updates) · e422a1d21 (447) · 0d2feee2f (447 hold) ·
089beb128 (447 §7) · 769d9f410 (token) · cfc3cd01c (chain enqueued) · this handoff.

---

## UPDATE BLOCK 1 — 2026-09-03 ~10:55Z (same lane, next session). §0 IS NOW STALE; read this first.

**The chain unblocked itself ~1 h after the handoff was written. Nothing was hand-spawned.**
`needs_briefing` `95d834f8` claimed 09:33:16Z → complete 09:34:35Z; `needs_site_plan` `173744d7`
claimed 10:38:06Z → complete 10:40:40Z. §0's "still `triaged`/unclaimed, starving" line and its
"First thing to check" are both **spent** — the answer was yes, it was claimed.

**§6's first check FAILS, and this is the headline: the new plan has ZERO article pages.**
Plan `c920da7a` (10:40:21Z) = `index` (landing) · `articles-index` (section-index) · `about` ·
`contact`. Same four as rebuild #1. **The gate did not drop them — the planner never proposed
them** (`llm_call_log` `7b3bffdd`, 4,072/16,000 out, not truncated).

**CAUSE — a planner REFUSAL, not an accident, and not site-specific.** Its own `strategy_notes`:
"The blog-post type is satisfied by the **blog infrastructure**; individual posts are not planned
as static pages here." No such infrastructure exists (`[MEASURED 10:48Z]` active `blog-post` pages
with sections: webdesign.co.uk 52, dartsonline.com 23, finetuning.uk 22, **gamesdesign.co.uk 13** —
all ordinary planned pages). `[MEASURED 10:47Z]` **3 of 32** `plan_site` runs in 30 days reason
this way; the other two are **designblog.co.uk** (09-02) and seotools.co.uk (09-02). Full evidence:
**CONTRIB appended to `bugs_open/444` today** — that is the durable copy, read it there.

**There is NO per-site lever, and this is the load-bearing finding.** Mission v3 (seeded 09:45:50Z,
55 min before the planner ran) says "The site launches with real articles…"; I verified those words
reached the model by reading the RENDERED prompt (line 110), not the seed. It planned none anyway.
`site_plan_directives` is not a lever either: all 1,922 rows are WRITTEN BY the planner, and
"directive" appears 0× in its rendered prompt. **So do not answer this by rewriting the brief
again — that has now failed twice, at increasing explicitness.** The fix lives in
`build-site-planner`'s prompt row (migration; council-in-scope), and belongs to 444's lane, which
holds the owner's `producer=BOTH` ruling. Contributed there, not taken here.

**§0's ⚠ on migration 720 is RESOLVED: it is APPLIED and LIVE** — `enforce_listing_sources: true`
on `validate_plan`, narrowed rule 3 at position 25019 of `plan_site.prompt_template`. **But it is
absent from `schema_migrations`** (which carries 721/723/724/726/727/728). Told 444.

**The gate's first live rebuild run was CORRECT** — 2 `capability_gap` rows,
`gap_kind=producer_missing`: `index`→`blog_posts`, `articles-index`→`section_children:articles-index`.
Neither dropped, rightly (both realised → 001 preserve guard).

**STILL IN FLIGHT at 10:45Z** (all `triaged`, unclaimed): `needs_page` (index) · 5× `needs_imagery`
(`hero_articles`; `illustration_featured_article`; `illustration_article_card_pipeline` /
`_balance` / `_leadership`) · `needs_rerender`. **Let them run** — the realised `index` already
carries `["hero","featured-content","content-listing","generic-text-block"]`, so the article-grid
shape is live from rebuild #1 and finishing adds no new defect while delivering the game imagery,
the 721 hero test and the palette. **§6 is still the right checklist for what lands** — heroes
(721's first live test, owed to `components`), imagery temperature, copy. Two of its expectations
are now known to fail before you look: article pages (zero) and a hub listing real articles.

**One more thing §5 should carry:** the plan keeps a `contact` page with a `contact-form` section
while mission v3 says "There is no contact page, no contact form and no email address anywhere on
the site". The planner said so itself and preserved it because it is already built. So owner
decision #1 (the `contactforsales.com` address) is now a **three-way** choice — keep, replace, or
remove the page entirely as v3 asks — and the preserve rule means it will not go on its own.

---

## UPDATE BLOCK 2 — 2026-09-03 ~11:10Z. The build is BLOCKED on an owner decision, and the class fix went live 20 min AFTER our plan.

**`needs_page` `ac76ec54` (index) is `needs_human_review`, not running.** Corr `afce7c49`, claimed
10:58:44Z, failed `validate_content` at 11:03:32Z. Read the detail row, not the chassis row:
`SELECT jsonb_pretty(context) FROM agent_error_log WHERE work_item_id='ac76ec54-8f04-455f-8018-03bc5834ac96';`
(the chassis row says only "1 blockers, 0 errors"; `error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`
carries the issue list).

**The blocker is a banned claim, and the gate was right:** the writer wrote "…the contact page is
there for it" into the homepage; the registered claim is *"contact page (owner 2026-09-03: no
contact page and no contact form.)"*. **Do not fix this by editing copy or lifting the ban on your
own initiative** — it is the visible edge of a three-way conflict only the owner can settle:

1. **mission v3** says the site has no contact page, no contact form, no email address;
2. **the preserve guard keeps the realised `contact` page live** (planner's own `strategy_notes`:
   "preserved structurally… it exists only because it was already built") — it is in the nav and
   will not leave on its own;
3. **the banned-claim gate** blocks any copy that mentions it.

So §5 decision 1 is **promoted from a preference to a BLOCKER**: no longer "keep or replace
`gamedesign@contactforsales.com`" but **"remove the contact page, or lift the ban"**. The homepage
cannot build until one of the three gives. Nothing else in the chain is blocked by it.

**Second finding, same run — the zero-articles defect at the build layer.** Step `plan_sections`:
`page "index" section "featured-content": 6 required non-llm field(s) resolved from neither their
declared source nor the page's stored content_data — omitted under on_missing=skip_field`
(`featured_image`, `featured_excerpt`, `featured_category`, `featured_author`, `featured_read_time`,
`featured_date`). **The homepage's featured-article slot lost every field because there is no
article to feature.** ⚠ `skip_field` **degrades silently — it does not block**, so a rendered page
is not a page whose sections resolved. Check that `plan_sections` finding row after any build on a
site with an empty content type; nothing else surfaces it.

**THE CLASS FIX IS LIVE, AND OUR PLAN PREDATES IT.** The designblog.co.uk lane shipped migration
**730** (10:59:59Z) and **731** (~11:03Z) off this lane's evidence: `build-site-planner` gains rule
20, *"NO LATER EDITORIAL PASS RUNS — do not defer posts to one. The one mechanism that could create
posts later (blog-content-planner, via create_blog_posts) has been DORMANT since 2026-04-24…"*,
instructing 3–6 launch posts on real subjects from the site's own briefing. **Verified at the live
row by the `bugs_open/427` lane, not just "recorded as applied".** ⚠ **Grepping for 730's original
literal "THERE IS NO LATER EDITORIAL PASS" now returns ZERO rows** — 731 superseded it four minutes
later; that is a false negative, not a failed migration.

**So the next action for this lane is a RE-PLAN, not a rescue of this build.** Our plan `c920da7a`
was written 10:40:21Z, ~20 minutes before rule 20 existed. A re-planned site runs under 718
(content imagery), 720 (listing gate) **and** 730/731 (launch posts) — the article half arrives
free. **Settle the contact-page decision first**, or the same banned claim blocks the rebuilt
homepage too.

---

## UPDATE BLOCK 3 — 2026-09-03 ~11:45Z. Owner reversed the contact ruling; specs v4 live; RE-PLAN ENQUEUED.

**Owner ruling (midday):** the contact page **STAYS**; the address is
**`gamedesignuk@contactforsales.com`**; re-plan. This supersedes the 09:45Z no-contact ruling and
**closes §5 decision 1** and UPDATE BLOCK 2's blocker.

**Applied, both clean, both with ROLLBACK companions:**
- `SEED_2026-09-03c` (11:42:28Z) — `mission_brief` v4 (anchored replace, not re-pasted);
  `evidence_base` v4 with the two contact/email bans removed (**18 → 16**; the AI-masthead ban
  deliberately retained and post-condition-asserted); `writer_block` CONTACT paragraph replaced;
  address updated in `submission` + `briefing`.
- `SEED_2026-09-03d` (11:44:45Z) — cancelled the blocked `needs_page` `ac76ec54`, then enqueued
  `needs_briefing` **`5cce64a6`** (`triaged`, key `briefing_gamedesign.uk`).

⚠ **The trap in the re-plan, worth carrying to any lane that re-plans a site with a blocked item:**
`ac76ec54` held `item_key='needs_page:index'` at status `needs_human_review`, which is **NOT
terminal** (`workItemTerminalStatuses`, `work_items_common.go:42`). `idx_swi_dedup` keys on
non-terminal rows, so **the re-plan's own `needs_page` would have been silently deduped away and
the homepage would never have rebuilt** — no error, no missing-row warning. Cancel the stale row
FIRST, and assert the key is free before enqueuing (09-03d does both).

**What to expect from this plan, and what to check** — §6 remains the checklist, plus:
- **Article pages under `/articles/`.** This plan is the FIRST written under rule 20 (migrations
  730/731, live 10:59:59Z/~11:03Z). The previous plan `c920da7a` (10:40:21Z) predated it by ~20 min
  and had zero. If this one ALSO has zero, that is a rule-20 failure and the `bugs_open/428` /
  `designblog.co.uk` lanes need telling immediately — it would be the first evidence against a fix
  built on this lane's measurements.
- **The contact page must survive and carry the NEW address.** ⚠ `about` and `contact` are
  `deployed` (not `needs_rebuild`) and their stored copy still holds
  `gamedesign@contactforsales.com` in 4 component rows. **The email ban is gone, so nothing will
  flag it.** Grep the served bytes; if stale, file a targeted rewrite. `recompose_pages` on
  `needs_site_plan.spec` is the sanctioned re-composition lever (mig 385) but the briefing agent
  creates that item, so it could not be pre-set here.
- **The homepage's `featured-content` slot** had all 6 fields resolve empty last time for want of an
  article (UPDATE BLOCK 2). If rule 20 delivers articles, it should fill; check the `plan_sections`
  finding row, because `on_missing=skip_field` degrades **silently**.

**New bug filed from this lane's work: `bugs_open/460`** — the OTHER blog-post producer,
`blog-content-planner`, ran 13 times between 2026-03-14 and 2026-04-24 and has been silent for four
months. Unowned, no root cause asserted. It matters here because it is what `bugs_open/444`'s
`builder_needed=blog_posts` gaps name, and because rule 20's text now asserts its dormancy **with a
date** — revive it and that text goes stale.

---

## UPDATE BLOCK 4 — 2026-09-03 ~15:50Z. RE-PLAN RAN. Rule 20 worked; `validate_site_plan` deleted its output. Root cause = `bugs_open/463`.

**The chain completed 13:42Z → 15:02:57Z.** New plan **`005fb393`** (14:15:25Z). **Still four pages,
still zero articles.**

**DO NOT read that as "rule 20 failed" — I nearly reported exactly that and it is wrong.**
`llm_call_log` `00fe50c7` (14:15:16Z): rule 20 present in `prompt_rendered`, and the planner
planned **five** launch articles, saying so in `strategy_notes`. Subjects came from the briefing's
own four strands, with real games named in the section subjects (Hades, Slay the Spire, Dead Cells,
Baldur's Gate 3, Divinity: Original Sin 2, Cyberpunk 2077, Elden Ring, Disco Elysium, The Witcher 3,
God of War Ragnarök, Horizon Forbidden West), each grounded in what a player can observe. **The
planner half of this lane's problem is SOLVED.**

**`plan_site` = 9 pages; `validate_plan` = 4.** Silent: `capability_gaps_emitted: 0`, no
`agent_error_log` row, orchestration COMPLETED. **Pass C** (`v3_site_actions.go:7599`) drops any
LLM page whose `slugOf` matches a realised section stem, and both `slugOf` (`:6467`) and
`sectionStemOf` (`:6447`) reduce to **the first path segment** — so a child at
`/articles/<slug>.html` is indistinguishable from a flat page colliding with `/articles/index.html`.
Live since 2026-05-21, invisible because `Pass A: union` restores **realised** pages afterwards and
a NEW child has nothing to restore it. **So a section index that is empty today can never be
filled.** Full case, controls and fix candidates: **`bugs_open/463`**.

**⇒ DO NOT re-plan again for the articles.** A fourth plan will be deleted by the same pass. The
articles are blocked on 463, which this lane does not own (the `428` session is mid-build inside
that same action and has been told).

**THE CHECK THAT LOCALISES THIS CLASS, and the one I should have run first** — read the STEP
BOUNDARY, never `site_plan_pages` alone:
```sql
SELECT jsonb_array_length(collected_data->'plan_site'->'result'->'pages') AS proposed,
       jsonb_array_length(collected_data->'validate_plan'->'pages')       AS survived
  FROM orchestration_states WHERE correlation_id='<corr>';
```
`proposed > survived` means validation ate pages. The served page cannot tell you whether the
planner never proposed them or validation dropped them, and those have opposite fixes.

**Contact address: SPECS updated, PAGES were not — now delivered by `SEED_2026-09-03e`.**
`about` and `contact` were `deployed` (not `needs_rebuild`), so the re-plan preserved them and their
copy still carried `gamedesign@contactforsales.com`. Confirmed at the served bytes 15:45Z, not
inferred. ⚠ **The address is inside LLM-written PROSE** (`hero-contact.subheadline`,
`generic-text-block.content` ×2), not a resolved field — **no re-render can fix it**, including
post-454. Filed 3 × `section_edit` → `section-editor` (`edit_type=content_edit`, `field_updates`),
which IS a live dispatching path (158 completed fleet-wide in 3 days); `content_rewrite` is NOT
(all 182 recent rows are record-mode under RFC_056 and would have parked silently). **Verify at the
served bytes**, and note `about` carries it too, not just `contact`.

**⚠ MY WATCHER WENT BLIND 13:01–13:57Z — the window the chain ran in.** I re-armed the monitor with
a "better" query carrying a fleet-wide `GROUP BY` for queue position; it consistently exceeded the
60 s timeout. It said so only because I had added a consecutive-failure counter; with the usual
`|| true` + `2>/dev/null` it would have rendered as "no change". **Keep a watcher's probe cheap, and
make its silence prove itself.**

---

## UPDATE BLOCK 5 — 2026-09-03 ~17:30Z. 463 FIXED by another lane (NOT rolled); the contact address is a FOUR-surface change and the last surface is in flight.

**463 — FIXED, NOT LIVE, DO NOT RE-PLAN YET.** Session `463` took the fix this lane filed and declined:
`9b540c2e6` (17:52 BST), both halves — Pass C no longer deletes new section children, AND the write
path no longer relocates `blog-post` pages to `/blog/` (their correction to my §5: `parent_section`
DID matter, on the write path, not in Pass C — 109 of 109 `blog-post` plan rows had it absent). It
is Go, so **inert until an image rolls**: live agent-chassis is `30438851…` and
`git merge-base --is-ancestor 9b540c2e6 <stamp>` says NO. **They will tell this lane when it has
rolled and verified (proposed = survived at the step boundary; children at `/articles/<slug>.html`
not `/blog/`). Only then re-plan.** A re-plan before that is deleted by the same pass. They also
filed **`bugs_open/467`** in passing — a re-plan cannot add ANY new page to a site of 20+ pages
(`truncatePreservingRealised`), 26 of 42 sites; not this site (4 live pages) but the same family.

**The contact address lives on FOUR surfaces. Three fixed and verified; the fourth is queued.**
Each earlier pass verified clean at its own table while the live page still served the old value —
**only the served bytes are ground truth**, and even those need the CDN caveat below.

| # | surface | seed | verified how |
|---|---|---|---|
| 1 | `site_specs` — `submission.email`, `briefing.contact.contact_email` | 09-03c | rows |
| 2 | 3 components' `content_data` (about/generic-text-block, contact/generic-text-block, contact/hero-contact) | 09-03e → `section_edit` | rows 3 new / 0 old |
| 3 | **`sites.email`** — what chrome AND the contact-form's `rendered_html` render from (`render_site_components_action.go:464`, `rerender_pages_actions.go:796`) | 09-03f | row |
| 4 | **rendered chrome** (`site_components.footer`) — regenerated ONLY by `render_site_components`, which `rerender-pages` runs when `needs_rerender.spec.refresh_site_components=true`; a single-page `page_rerender` does NOT touch it | 09-03g | footer row `updated_at 17:13:15`, new addr ✓ old addr ✗ |

**Why the served page STILL shows the old footer at 17:30Z, and why that is not a failure:**
`rerender-pages` regenerates chrome and then **spawns one `page_rerender` child per page**
(`create_rerender_items` step) — it does not deploy itself. Four children were spawned 17:13Z and
are `triaged`, queued behind the fleet selector. **Until they run, no page has been deployed since
chrome was fixed** — `pages.deployed_at` is 16:26–16:28Z for three pages and index is still
`needs_rebuild` from the 11:03 failure guard. A background watcher (this session) fires when all
four complete and then reads deploy dates + cache-busted served bytes. **If you inherit this
before it fires:** `SELECT status FROM site_work_items WHERE site_id='8f17eb73…' AND
item_type='page_rerender' AND created_at > '2026-09-03 17:12';` — then the curl loop in
`SEED_2026-09-03g`'s header. **Pass condition: ONLY `gamedesignuk@` on every page, ZERO bare
`gamedesign@`.**

⚠ **CDN:** `cache-control: max-age=3600`, `cf-cache-status: DYNAMIC`. A cache-busted URL
(`?v=<epoch>`) reaches origin; a `Cache-Control: no-cache` REQUEST header does not. `last-modified`
on the served file is the honest deploy clock — it read **15:12Z** at 17:20Z, i.e. no deploy had
landed since before any fix, which is how "CDN lag" was ruled OUT and "children not yet run" ruled
IN.

⚠ **Two watcher lessons from today, both mine:** a heavier probe (fleet-wide `GROUP BY`) outran its
timeout and blinded the watch for the 56 minutes the chain ran in; and a watcher that greps only
the success marker is silent through a failure. The current one polls a single-row count by key,
exits on any failed/needs_human_review/unresolved child, and caps at 150 min.
