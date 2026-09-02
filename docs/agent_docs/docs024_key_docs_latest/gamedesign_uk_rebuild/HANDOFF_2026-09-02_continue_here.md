# HANDOFF — gamedesign.uk lane — continue here

**Written 2026-09-02 (evening)** by the session named `gamedesign.uk`. Cold-start document: read
this first, then the PLAN for decisions, NOTES for evidence, RUNBOOK for commands. Every figure is
`[MEASURED 2026-09-02]` unless marked.

Directory: `docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/`
Bug filed: `bugs_open/432_HANDOFF_2026-09-02_a_site_whose_db_rows_were_deleted_keeps_serving_and_no_detector_can_enumerate_it.md`

> ## ⚡ UPDATE 2026-09-02 ~17:10Z — ALL FOUR OWNER DECISIONS LANDED AND EXECUTED; BUILD DISPATCHED
>
> Read §4/§5 below as history. Current state:
> - **A1 email** `gamedesign@contactforsales.com` — on the row. **A2 look** — theme-kits lane
>   supplied a bespoke "practice journal" palette (warm paper `#F4F1EA`, ink `#23211E`, earth
>   accent `#A6521F`, `serif-editorial` typography, `soft-editorial` layout); seeded via
>   `mission.preferred_palette` (rung 1 of the cascade) + `design_intent`. ⚠ **Kits are NOT
>   live** (`theme_kits` unapplied) — values seeded directly. ⚠ **Owner ruling (via theme kits):
>   `reference_values` is NOT a pin**; the overlay may move off it — check the served
>   stylesheet after the build, do not clamp. **A3 old files** — CLEARED: sites-repo commit
>   `40bd35f19` removed 58 tracked files, kept `404.html`; verified at the artefact (`/`,
>   `/tools.html`, `/about.html` → 404 fleet page, control 404, bucket = 1 object). **A4 brief**
>   — one sentence added on look (the sibling's actual values as the referent), 2,891 chars.
> - **SEED applied 17:04Z** (`SEED_2026-09-02_gamedesign_uk_site_and_specs.sql`): site row
>   `8f17eb73-fc74-4718-8371-b3125bc4e414`; `mission`, `design_intent`, `evidence_base`,
>   `imagery_style_guide` all current+pinned. Also `oxenunity.com` row (owner: "should have a
>   row, created here"; hand-built; email NULL, not invented; `settings.managed_by=hand`).
> - **DISPATCHED 17:07:55Z** — correlation `f07313f6-976c-4593-9e5e-44892008fb74`,
>   orchestration `2069ee9e-c626-4b98-9faf-1553b1e3376a`, submitter COMPLETED;
>   `needs_domain_research` triaged to `domain-research-classifier`. The brief landed in
>   `mission_brief.text` (which is what the classifier reads — verified in its definition), NOT
>   in `mission`, so the pre-seeded palette row was never even merged over.
> - **432 fix BUILT + LIVE** (`scripts/audit-rowless-serving-domains.sh`, IMP-059): first run
>   found **10** NO-ROW-AND-SERVING (bucket 55 prefixes vs repo 36 — three domains a repo census
>   cannot see). After the seed: gamedesign.uk + oxenunity.com → ROW_NO_PAGES, control stayed
>   NO_ROW. Owner: the remaining 8 are the **adoption backlog**, after this lane, with oversight.
> - **315 REOPENED** → `bugs_open/`, handed to `site_ai_agent_orchestration` ("AI page 3").
> - **Brand leak** routed: a dedicated `gamesdesign.co.uk` session [783baf] now exists and has
>   the owner's instruction + evidence; positioning recommends "GamesDesign.co.uk". The CLASS
>   bug (adoption copies `company_name` verbatim to a different domain) is still MINE to file —
>   not yet filed; coordinate with that session so it is filed once.
>
> **WATCH (17:20Z):** `bugs_open/438` (theme kits) — aspect `mission` is never populated on the
> FRESH path, so my `mission.preferred_palette` seed is rung 1 by ACCIDENT and the only
> populated one on the estate. Fix = repoint `persist_mission` (config, live on apply). Narrowed:
> `write_site_spec` deep-merges, so the fix ADDS `text` beside the palette; and my submitter has
> already COMPLETED, so it cannot re-run without a re-submission. Classifier superseded my
> `design_intent` (pinned did not hold) but in the SAME direction (`#F5F0E8`/`#9B4E2A`/light,
> Playfair headings). ~~Read `palette_source` on composition~~ **RESOLVED 17:38:00Z:
> `mission_hint` — rung 1 fired for the first time in production; palette byte-for-byte the seed; 438's diagnosis HOLDS.** Layout `magazine-grid` (not soft-editorial), typography Playfair + Libre Baskerville. Site plan = 5 pages. Design/pages/rerender/deploy still to land; then read the served `--color-*`.
>
> **WHAT IS LEFT (was §5 steps 5–9):** wait for the cascade; verify at the artefact (RUNBOOK §8)
> — every file non-empty `<main>`, legal pages + sitemap 200, no empty `mailto:`, control 404;
> read the served stylesheet for the palette; read the served copy against the positioning
> constraints; tell `Portfolio positioning` if it landed elsewhere; first SUMMARY; file the
> class bug; close. **432 stays open** until the reconciler is scheduled (IMP-059 gap 1) or the
> owner says keyboard-run is enough.

> ## ✅ UPDATE 2026-09-02 ~18:10Z — SITE LIVE AND VERIFIED AT THE ARTEFACT
>
> Dispatch 17:07:55 → homepage live ~17:56 → styled ~18:00. Four pages serve with content
> (`/` 2,118 chars, `/about` 5,335, `/contact` 1,984, `/articles/` 2,148), sitemap 200, control
> 404, no empty `mailto:`, no forbidden phrases, sibling cross-linked, every internal link 200
> except `/assets/images/favicon.png`. Full table + the palette reading: NOTES 18:05Z.
> First SUMMARY written: `SUMMARY_2026-09-02_gamedesign_uk_rebuild.md`.
> **Served palette ≠ composed palette on all 8 slots** (overlay took the classifier's values) —
> permitted by the owner's ruling; reported to theme kits as values. `audit-rowless-serving-domains`:
> gamedesign.uk now OK (absent), oxenunity.com ROW_NO_PAGES.
>
> **STILL OWED / OWNER DECISIONS:** (1) `article` slot parked at `needs_human_review` — leave or
> cancel; (2) no privacy/terms pages — not planned, not linked; want them?; (3) favicon 404;
> (4) llm-cost-calculator.html retract-vs-rebuild (from 315's lane); (5) 432 stays open until the
> reconciler is scheduled (IMP-059 gap 1) or keyboard-run is ruled enough. Lane closes on (1)+(2).

> ## ✅✅ UPDATE 2026-09-02 ~20:00Z — OWNER'S LIST EMPTY; LANE CLOSEABLE
>
> All four rulings executed and verified at the artefact (NOTES 19:25Z, 19:55Z): article slot
> cancelled + archived (`SEED_2026-09-02b`); no legal pages; **favicon LIVE** (64×64 PNG + OG
> card, head-referenced); llm-cost-calculator is ai-agent-orchestration's and MOOT (tool exists).
> Sibling renamed to GamesDesign.co.uk and verified at their artefact; their one inbound link
> repointed. Second SUMMARY: `SUMMARY_2026-09-02b_…`.
> **If you are picking this up cold:** the only thing owed by THIS lane is the one-query watch in
> NOTES 19:55Z (no work re-filed at the archived `article` page after the next rotation). Then
> close. `bugs_open/432` (scheduling IMP-059) and the 8-domain adoption backlog are platform /
> owner items, not this lane's.

## 1. What this lane is, in one paragraph

The owner asked for gamedesign.uk to be fixed — "it is in a bad way". It is: five of its linked
pages (13 of 47 files) serve a header, a footer and an empty `<main>` to the public, and have
since **2026-04-16**. The platform has **no record the site exists** — no `sites` row, no `pages`
rows — so nothing could repair it and nothing could see it. The owner then redirected: first
establish **why the April adoption broke it**. That is done and independently re-verified. Then
the site gets rebuilt through the framework, in a **different direction** to its sibling
`gamesdesign.co.uk`, on positioning agreed with the `Portfolio positioning` lane. The rebuild is
prepared and **not dispatched** — it waits on four owner decisions (§4).

## 2. What is DONE and proven

| item | state | evidence |
|---|---|---|
| Live damage measured, with a fabricated-URL control | DONE | PLAN §2, 432 §2 |
| Root cause of the April break | DONE, re-verified on a second model | NOTES; 432 §3 |
| Discriminating control (this site vs the fleet that day) | DONE, reproduced exactly | 4/11 emptied on gamedesign.uk, 0/139 on six other sites, 2026-04-16 |
| Is the publish defect still live? | NO for new publishes — two real guards (`d777cb4d2` 05-12, `6579e9ae1` 07-27) + one build-side rescue (`856fc4a51` 06-08) | verified in the RUNNING binary: stamp `ebf27c60377f` (fresh build 2026-09-02 ~16:19, all three ancestors, both controls pass); previous stamp `a2732c72` likewise |
| Detector gap filed | `bugs_open/432`, UNOWNED | absence verified across `cmd/`, `scripts/`, in-chassis Go; the only bucket-listers are the WRITE path |
| Corrections after the Fable re-investigation | 7 details fixed in place, 6 `CORRECTED` blocks in 432, NOTES + 016b + WRONG_CALLS updated | commit `e7fa8f738`, `054020429` |
| Positioning agreed | DONE — practice seat vs the sibling's authority seat; register rows GD1/GD2 written by that lane | PLAN D4 |
| Mission brief | DRAFTED, committed, owner has seen it, **not approved yet** | `MISSION_2026-09-02_gamedesign_uk.txt` |
| Standing five | PLAN, NOTES, README_where_we_are, RUNBOOK (this session), no SUMMARY yet — first milestone is the site live | this dir |

## 3. The mechanism, for whoever has to explain it again

The April adoption was run with gamedesign.uk as **both source and destination** (the 2026-03-30
trigger took one domain by construction; source/destination separation arrived 04-21). It **wiped
and recreated** the page rows as placeholders. The rerender of the day refused only when
`len(siteComponents)==0 && len(sections)==0` — the site had chrome, so it wrote header + empty
`<main>` + footer, `git-adapter` committed "Rerender: X", the Action `b2 sync --delete`d it to
the bucket. The content cascade never completed (`validate_page_content` blockers). The 04-16
handoff recorded the empty pages as "P3 — work items not processed yet" and expected them to fill.
Later the site row was deleted (undatable; alive 04-20, gone by 09-02), taking every repair handle
with it — except **1,147 `site_work_items_archive` rows** that still carry the id and domain.

## 4. DECISIONS THE OWNER OWES — nothing below dispatches until these land

**A. The rebuild inputs** (all three needed before §5 step 1):
- **A1 — contact email** for the `sites` row. Fails-open check without one (`bugs_open/063`).
  House pattern looks like `<site>@contactforsales.com`.
- **A2 — look and feel.** Old site and sibling are both dark + cyan (`#121212`/`#00bcd4`). Say
  "visibly different from the tools site" and one sentence goes in the brief; say nothing and the
  classifier picks.
- **A3 — the old files.** 47 HTML files in `~/projects/sites/gamedesign.uk/`, including five tool
  pages and `tools/ guides/ games/ blog/` — the content kind the brief forbids. A FRESH build only
  overwrites paths it produces; the rest survive as stale orphans. Recommendation: clear the
  directory in the sites repo before the build lands (one pathspec commit, `b2 sync --delete`
  does the rest). Keep `index.html-gamedesign.uk`? It is untracked, the original hand-written
  homepage — reference material, never deployed.
- **A4 — approve the brief** as written, or edit it. It fixes the site's whole direction at the
  classifier.

**B. The LIVE 315-shaped defect on another lane's site.** `ai-agent-orchestration.com/roi-estimator.html`
serves empty; row `active`/`deployed`/**0 component rows**; **eight `page_rerender` items
`complete` 08-26→09-02**. The guard skips, `check_skipped` closes the item as complete, the artefact
is never touched. `bugs_closed/315`'s profile, live. Two more from the 04-18 born-empty wave still
serve (`llm-cost-calculator.html` archived-and-serving; `robot-hands.com/learning-center-article.html`).
**Options:** reopen 315 → `bugs_open/`; file new; or hand to that site's lane. Not this lane's
site — recorded in 432 §3a, not filed.

**C. Who owns `bugs_open/432`.** Filed UNOWNED. Fix candidate 1 (enumerate the bucket / sites
repo, reconcile against `sites`) is small and self-contained; the Cloudflare `zones?name=` call
the deploy Action already makes is a ready enumerator. This lane can take it after the rebuild, or
it sits in the queue.

**D. The seven other row-less serving domains** (`nanangmrk.com`, `oxenunity.com` [known
hand-built], `puzzlegenerators.com`, `testllmlog.example.com`, `website-design.com`,
`websitedesign.com`, `wykefarm.uk`). Four serve 200 with a 404 control. Are they meant to? A
one-line answer each, or a census for 432's fixing thread.

**E. Sibling brand leak** — 23 of 49 `gamesdesign.co.uk` titles read "… | GameDesign.uk". That
lane's site. Tell them, or not.

## 5. What is LEFT before this lane can close

The lane closes when **gamedesign.uk serves a complete, framework-built site in the agreed
direction, verified at the artefact**. Steps, in order, once §4 A1–A4 land:

1. Pre-seed the site row + `evidence_base` + `imagery_style_guide` — RUNBOOK §6. Set `name` AND
   `network_id` explicitly (`ensure_site_record` NULL trap).
2. Clear the old artefacts per A3 — pathspec commit in `~/projects/sites`, push to `master`
   (not `main`; the Action fires on `master` only). Verify `b2 ls b2://portfolio-sites/gamedesign.uk/`.
3. Check the fleet is not mid-roll (RUNBOOK §5) — no dispatch within ~300 s of a chassis restart.
4. Dispatch — RUNBOOK §7. Save the printed correlation; find the run by payload.
5. Wait for the cascade (`needs_domain_research → strategy → briefing → site_plan → composition →
   design → content_page ×N → rerender`). Budget hours, not minutes; the classifier queues behind
   the fleet.
6. Verify at the artefact — RUNBOOK §8. Every HTML file non-empty `<main>`; `/privacy.html`,
   `/terms.html`, `/sitemap.xml` 200; zero `href="mailto:"`; the control still 404s. **A
   `complete` work item is not a rendered page** (432 §3a).
7. Read the served copy for the positioning constraints: no calculators/tool pages/guide library;
   no "game room"; no paid-product claims; no negative-identity copy; links to the sibling where
   tools are referenced. If it landed elsewhere, tell `Portfolio positioning` so GD2 is revised.
8. Write `SUMMARY_<date>_gamedesign_uk_rebuild.md` — first milestone. Append README.
9. Close the lane. **`bugs_open/432` stays open regardless** — it is a platform defect, not this
   site's; it needs its own owner (§4 C).

Not in scope for closing: B, D, E above are other lanes' or the owner's.

## 6. Landmines this lane hit or confirmed

- **An absent file measures as an empty file** in a `git show $sha:$path` loop when `$sha` is
  empty. Guard the sha. (WRONG_CALLS 2026-09-02 #1.)
- **A bare line number on `rerender_single_page_action.go` is stale within hours** —
  `improvement_loop` edits it daily (581 → 652 → 680 in one afternoon). Cite sha + code string.
- **CLAUDE.md's `build provenance` log grep is out of range on agent-chassis** even at
  `--tail=20000`. `service_binary_capabilities` (column `service`, not `service_name`) is the
  reliable route. During a roll, two stamps coexist.
- **A CSS rule is not damage until the markup instantiates it.** Enumerate the markup's classes
  before computing a contrast ratio. (WRONG_CALLS.)
- **A second investigator's refutation is a claim too** — Fable's "19 directories" was wrong
  (36). Re-measure before replacing your number.
- **`audit-archived-still-serving.sh` cannot see a domain with no `pages` rows** — 432's whole
  point; do not read its clean pass as covering orphaned sites.
- The `cd … && cat > file` idiom **silently writes nothing** if the `cd` fails because the shell
  cwd persisted from a previous call. Use absolute paths in heredocs.

## 7. Commits this session (all pathspec, all this lane's files only)

`eba9c3bb6` standing docs + root cause · `422e83351` ratchet · `d77fb709d` FILE 432 ·
`9cba99118` 016b §9 + WRONG_CALLS · `f772ef668` mission brief + PLAN D4 · `e7fa8f738` 432
corrected after Fable · `054020429` 016b figure · (this handoff + RUNBOOK: see `git log`).

## 8. Threads open to other lanes

- `Portfolio positioning [b9957b]` — GD2 stands; tell them if the brief lands elsewhere after
  owner review. They now carry 432's finding as a stakes-upgrade on their "21 domains with no
  register row" debt.
- `bugs_open/429` lane (delivery) — adjacent, not overlapping; nothing owed either way.
- gamesdesign.co.uk's owning lane (commit `f691db16b`) — the brand leak (§4 E) is theirs if the
  owner says so.
