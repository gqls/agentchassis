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
  on `/` only. Report the `url()`s to components right or wrong.
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
