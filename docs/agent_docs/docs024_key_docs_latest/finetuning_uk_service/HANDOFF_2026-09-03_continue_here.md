# HANDOFF 2026-09-03 — the site is now the playground; the demo model is live on a Hetzner box; the chat route is LIVE on the island (19:08Z); **the chat widget is LIVE on /playground.html and answers (20:33Z, browser-proven)**. Start here.

**COLD-START for the finetuning.uk service lane.** Supersedes
`HANDOFF_2026-09-02_continue_here.md` (still the reference for the playground booking page's birth,
the 443 diagnosis and the plan-less-site finding; nothing in it is retracted except the rows its
own appended correction already strikes). Technical log: `NOTES_finetuning_uk_service.md`, entries
dated 2026-09-03 (many; read them in order). Owner prose: `README_where_we_are.md`. Milestone
read-out: `SUMMARY_2026-09-03_finetuning_uk_service.md`. Plan: `PLAN_2026-07-31_…` Phase P and the
DIRECTION block at its end.

## The owner's direction (2026-09-03, his words, which the whole lane now serves)

> "As a whole, I'd like the finetuning site to be very much focused around this tool. We can still
> have the other tools, but much of the "what else we do as a company" should now move to leopardess
> consulting or other "me" sites. For finetuning.uk I'd like this tool shown prominently on the home
> page and I want in the future example after real example of what we've done and before and after
> examples. And I'd like to host those same models so they can try them (at maybe a couple of pounds
> for an hour or something that covers our costs say 5x) We can talk details later."

Plus, same day: the playground is **BOTH** a public demo and booked GPU hours; the hosted examples
"could be a big [GPU] so it feels snappy" (a100xl is $1.09/hr real ⇒ ×5 ≈ £4/hr, not £1.40); and
third parties may later submit models with a page of their own ⇒ the catalogue is "model pages"
with an OWNER from day one. **Nothing has been moved or re-briefed on the back of this yet**; the
details conversation has not happened. Do not move company-general copy without telling the
`leopardess` lane first (a live session exists).

## STATE OF THE PLAYGROUND BUILD (PLAN Phase P) — read this table before touching anything

| step | state | where |
|---|---|---|
| 1. demo model server | **DONE, LIVE.** Hetzner `relojistas` 167.233.33.159 (key `~/.ssh/hetzner1`), Ollama 0.33.2, model `finetuning-demo` id `cd4c8ea62f1d`, **38–42 tok/s** (cluster CPU was 14), warm reply 1.3 s, 2.0 GB of 3.8 used. Open to the island's `176.126.243.183` ONLY, on a default-deny ufw box. Reachability proven both ways. | RUNBOOK "The demo model host" |
| 2. tools-api route `POST /api/v1/tools/playground/chat` | **DONE, council APPROVED** (`63be72d1`, round 4, all reviewers; rounds 1–3 revise, each found something real). Committed; HEAD builds with tests. Registered **PUB-006** (route) + **PUB-007** (`mountBrowserGroup`). **NOT SHIPPED to the island.** | `internal/tools-api/handlers/playground.go`, `api/server.go`; RUNBOOK "step 2: ship" |
| 2a. island `sites` allowlist row | **DONE** — migration `737_…_ISLAND.sql` applied + ledgered; island `sites` = finetuning.uk, robot-hands.com, vonc.com | island `island_migrations` |
| 2b. `/opt/island/.env` five `PLAYGROUND_*` keys | **DONE 2026-09-03 ~18:55Z (owner, by hand)** — AND the Tenant 3 `environment:` block in the compose file, without which the keys never reach the container (the first restart booted "NOT mounted"; RUNBOOK step 2 corrected, LANDMINES entry). | island; repo compose copy `gauntlet_dead_cta/infra/island/docker-compose.yml` |
| 2c. island image swap | **DONE 2026-09-03 19:08:49Z** — `aqls/tools-api:v1.0.1359-playground` (from `9b540c2e6`) running; boot log `playground route group mounted (ollama=http://167.233.33.159:11434 …)`; from outside: 200 text/event-stream (53 tokens, cold 3.85 s, warm 0.96 s), 403 wrong Origin, 204 preflight, 400 bad body, gripper/gauntlet alive. Rollback files on the box: `.bak-1343-pre1359pg`, `.bak-1359pg-noenvblock`. | RUNBOOK "step 2: ship" (verify record) |
| 3. the chat widget on `/playground.html` | **LIVE 20:26:37Z and BROWSER-PROVEN 20:33Z** (`cdp_chat_probe.py` PASS: typed, sent, streamed reply in the transcript). Earlier state text follows. GENERATED AND LINKED (path (a), 20:11Z): component `tool-playground-finetuning-uk` (`b19eabe6`) is the SEVENTH slot on the page; the page holds the `tool` role since 20:00Z (`SQL_2026-09-03_playground_page_role_tool.sql`, reversible). SERVED only once `page_rerender` `50c2a394` completes — check `pages.deployed_at > 20:11Z` and `grep -c getReader` on the served page; then run the browser probe (`cdp_chat_probe.py`, in the 09-03 session's scratchpad — copy it into this dir if it proves useful) for a real streamed reply. ⚠ The generator also queued a stock tool-page `needs_content_page` that would have REPLACED the six booking sections — CANCELLED (`19b74d62`); if an `add_tool` is ever re-run on this page, cancel its twin again.** Earlier plan text follows. ~~NOT STARTED — and the plan here is CORRECTED (2026-09-03 17:10Z, measured at the library row).~~ ~~fork `chat-input-box` … repointed at the route~~ The library box is single-turn `{message}`→JSON, same-origin, path a literal, no endpoint field; the route is multi-turn `{messages}`, cross-origin, SSE (`token`/`done{truncated}`/`error`). A fork cannot be "repointed"; the JS must differ. Paths: **(a)** `add_tool` with `library_source: null`, the route contract in the description, function `tool-playground` + the generator's `adopt_existing_page` flag so it lands on `/playground.html` (name match) rather than minting `/tools/…`; **(b)** the estate's two live cross-origin widgets (`gripper-report-intake` mig 651, `gauntlet-round-record-vonc-com`) are hand-written `js_snippets` + section + locked row. **Owner's call; lane recommends (a), (b) as fallback** (README). `deploy_config.capabilities += backend` only if the tool carries `requires-backend`, and NOT before the route is live. | NOTES 16:45–17:15Z; PLAN Phase P |
| 4. booked-hour GPU provisioning as a workflow | not started; thunder actions exist for training runs (`thunder_ssh_exec_dispatch`, `_decommission_dispatch`, …) | PLAN |
| 5. booking → session handoff; the examples catalogue ("model pages" with an owner) | not started; "details later" | PLAN DIRECTION |

**Post-ship verification — RUN 2026-09-03 19:09–19:12Z, all green** (symbol 5 / 3 / 0 on the island's
grep vs 2 / 2 / 0 locally on the same bytes — the assertion is >0 / >0 / 0; mount line present; 200
text/event-stream; 403; plus 204 preflight, 400 validation, sibling routes alive). Recipe + figures in
RUNBOOK. ⚠ The owner's own `tail` of `/opt/island/.env` echoed `GRIPPER_SMTP_PASS` into the
2026-09-03 transcript — rotation recommended; read that file through `grep -v PASS`.

## ⚠ THE THINGS THAT WILL COST YOU TIME IF YOU DO NOT READ THEM

1. **tools-api does not run in the cluster.** It is a docker compose stack on the island VM
   `toolsapisuk.vs.mythic-beasts.com` (1 vCPU / 1 GB), behind a Cloudflare tunnel, with NO cluster
   credentials and NO route into the cluster (`[MEASURED]` cluster DNS does not resolve there;
   `162.209.114.65:11434` refused). I told the owner the demo would "run on the cluster's own CPU"
   before checking this (WRONG_CALLS 2026-09-03(e)). The cluster's `finetuning-demo` model still
   exists on `ollama-adapter` and gave the first speed figure; it is not the serving copy.
2. **An island-targeted migration with a plain name under `sql_for_agents/` IS applied to
   `clients_db` by `run-migrations.sh`.** `198_tools_api_gauntlet_rounds.sql` was (2026-08-08), and
   `gauntlet_rounds` exists in the core because of it. The guard is an UPPERCASE suffix (`_ISLAND`),
   which `SIDECAR_RE` reports and never applies. LANDMINES entry filed. Do not copy 198/276/436's
   plain names.
3. **A `needs_content_page` row can be an audit VERDICT, not a brief.** Rebuilding a page born from
   the site build by "copying its last complete item's spec" copies a `content-quality-audit` row
   with `not_dispatchable` set and a one-line suggestion as the whole brief. I did this to the
   homepage and caught it before claim. LANDMINES entry + the guarded recipe in RUNBOOK.
4. **Only `page-build-handler` reaches the 443 fix** (`load_page_sections_from_spec`). The
   `page-rebuild` agent cannot exercise it, and it takes every `needs_rebuild` page on the domain.
5. **Both finetuning.uk rebuilds today shared ONE correlation** (`6e8eadaa`): the dispatch loop
   takes a site's items in one loop. Key per-page LLM-log reads on the writer's `orchestration_id`.
6. **Never nest psql inside `ssh '…'`** — quotes are eaten and a quoted literal reads as a column.
   Heredoc through `ssh … 'docker exec -i … psql'`.
7. **Bug number 456 is duplicated.** This lane's is
   `456_…_writer_emitted_a_malformed_closing_tag…`; cite by slug.

## 443 (repeated sections) — Stage A CLOSED; **Stage B RAN 19:29Z: half a win, and the mechanism is now known**

- **641/A4 live 19:26:46Z.** Stage B dispatched on both pages (`d630f6df` technical-details, `11e1e8ed`
  your-own-model), acceptance in `stage_b_assert.sh`.
- **technical-details served 19:35Z (orch `89059f29`):** six h2s DISTINCT, no `</strom>`, no family
  listing, em dashes unchanged (4, all chrome), controls byte-identical. **BUT sections 2/3/4 all open on
  "small open-weight model… we choose the model" under three different h2s** — the repetition moved
  from the headings into the bodies. Read at `llm_call_log`: the A4 block rendered CORRECTLY in all six
  prompts (own subject + five siblings, right assignment); what overrides it is (1) the block has no
  instruction (a sentence and a list, no verb; "## What To Write" never names the subject) and (2)
  "## Rewrite Guidance (IMPORTANT: incorporate this into the content)" hands the WHOLE six-section brief
  to every section. Sent to the prompts lane (their edit; owner re-reads bytes) and the 443 lane.
- **your-own-model LANDED on attempt 2 (20:22Z, hero 338 chars = 65% kept; the floor was NOT tuned)**
  and shows the SAME shape: five distinct h2s, hero/FAQ/CTA on their subjects, sections 2/3/4 all
  "how it works, in three steps". Two pages, one structural cause; prompts lane's diagnosis run holds it.
- ~~your-own-model REFUSED at save~~ (attempt 1, 19:40Z, superseded by the line above): the writer followed its A4 hero subject and
  wrote a 212-char hero; the existing is 429 (padded by the 08-26 tool cross-link); `save_page_sections`
  refused the page at 49% kept vs the 50% shrink floor (bugs 178/293) and filed `save_refused_incomplete`
  `3034678c`. **The floor is STEP CONFIG on page-build-handler (`section_shrink_floor`), not per item** —
  do not tune it fleet-wide for one page. Retry 2 of 3 at 20:10:48Z; if all three fail, the choice is the
  owner's (options: accept the old hero by editing the brief to keep its length; or a one-off floor
  override migration with snapshot; or leave the page as is). A4 short opening lines + short heroes vs
  the 178 floor is a class, flagged to the prompts lane.

### (superseded heading) 443 — Stage A CLOSED, Stage B waits on 641

- **Stage A PROVEN** on `technical-details` (corr `6e8eadaa`, writer `ce514ce0`): all six
  `sections_ready[].subject` populated from `pages_table` (tier 3), carried to the writer's row;
  detector quiet with a fleet demand control of 7. `pages.section_subjects` backfilled for
  `playground`, `your-own-model`, `technical-details` (6/6 each; ~~clause form, e.g. "what to have
  ready before the hour"~~ **re-authored to A4 sentences 19:20Z**, e.g. "If you'd like to prepare in
  advance of your hour, you might want to get these things ready."); `our-position-on-ai` still
  null (no brief to derive from).
- **Served h2s still repeat** ("The model, and the licence it comes with" ×3, wording varied):
  the subject reaches the writer's DATA, not yet its PROMPT. That is Stage B = migration 641.
- **641 is now the PROMPTS lane's, end to end, as OPTION A** (one field, the subject authored in
  the voice, printed verbatim; sibling list by subject equality). The "You'll want to know ___"
  rule is RETIRED. **Do NOT carry the seed's old INSERTED TEXT block to the owner** — its bytes
  change. When they apply it, this lane runs Stage B: `your-own-model` + `technical-details`
  through page-build-handler (RUNBOOK recipe; assert `NOT (spec ? 'mode')`), distinct-h2s
  assertion, a tier-1 page as control. `apis.uk`'s own 641 SQL (`Council-Submitted 6c92d154`) is
  the base they edit; it adds `sections_for_render` to the writer's `input_fields` in the same
  UPDATE — that half must survive the rework or the sibling list renders silently empty.

## Copy — the homepage rebuilt, two owner rulings settled, one live defect filed

- Homepage content-rebuilt through the writer (item `1513b86a`) on the owner's verdict; his
  sentence serves as "We're not tied to one provider." (from the brief, not the gate). Gate: 9 in,
  3 out, all three survivors rejections. The shape the page still carries most is the "so" clause.
- **Owner rulings, with the copy lane:** cuts to a complete first clause are ACCEPTED (they replaced
  the 40% length ratio with a 5-word survivor floor, `Council-Submitted b9b5fdf8`, inert until the
  next roll); the "so" clause is a JUDGEMENT class with his three sentences as exemplars, because
  his stated discriminator ("follows a definition") missed one of his own positives.
- `technical-details` rebuilt: gate 9 in / 0 out, 8 of 8 truncations in his ruled form; but it
  shipped `<strong>open-weight model</strom>` live — **`bugs_open/456_…malformed_closing_tag`**
  (the save seam's `sanitize…` is bug 190's envelope guard, not an HTML check); not re-run, Stage B
  rewrites it. The page also came back a third shorter (writer variance, not a rebuild property).
- The technical-details BRIEF asked for the three-model listing the owner called unhelpful;
  rewrite the brief before Stage B rebuilds that page. `your-own-model` now has
  `/playground.html` in `content_direction.required_links` for its next rebuild.

## Reported by peers, not yet actioned

- **`/tools/llm-cost-calculator.html` `tool-cta` block: 5 items, all `image: ""`, array frozen
  since 2026-08-12 15:10** (bugs_open/384 lane, verified here 16:40Z: slot 2 is a `section`
  component, not the tool fork, so the section-editor route is safe from the tool-fork blanking
  landmine on THAT slot). Page is `rebuild_policy='owned'`, so `save_page_sections` refuses it and
  10 `page_rerender` items "completed" by refusing. Class: `bugs_open/389` §2 "Owned page"; remedy
  is migration 486's `section_edit` → `section-editor`. Served page shows one `<img>` (the logo),
  so visitors see cards without images, not broken ones. Left with 389 unless the owner wants it
  sooner.

## Next session, in order

0. ~~If the owner has done 2b + 2c~~ **BOTH DONE 19:08Z, probes green. Owner chose PATH (a)** (generated
   through tool-generator; hand-written is the fallback). Start step 3: an `add_tool` item, function
   `tool-playground`, `library_source: null`, the route contract in the description (multi-turn
   `{messages}`, ≤12×1000, SSE `token`/`done{truncated}`/`error`, 403 on wrong Origin), landing on the
   EXISTING `/playground.html` — check first whether the live tool-generator's create step carries
   `adopt_existing_page` (TL-044; step config, default OFF) or the build mints `/tools/…`. First probe
   the route is still up (`curl … -H 'Origin: https://finetuning.uk'` → 200 text/event-stream) — the
   demo box's ufw rule and the island's compose are the two things that can silently take it down.
0a-00. **STATE AT 22:20 BST — read this first.** The owner, in order tonight: (1) *"it looks ok"* on the
   live page; wants practical steps, an input/output example or two, a clearer explanation of what the
   model does; the language "sounds ok, good in parts even". (2) Cost: none per use (answered by
   measurement). (3) GPU: no, 100% CPU at the box. (4) *"I will need to train it better… someone else's
   [writing] that has a defined character"* — Phase 1 reshaped (PLAN, NOTES 21:05Z). (5) THEN: *"it can
   still go up. We can explain it away for now, this is just 5 articles and a handful of short emails.
   and as you'd expect you get not much change"* → **the widget is being regenerated IN PLACE**
   (`playground_widget_replace_dispatch.sql`, item `e1b2bcf8`, triaged 21:10:27Z, TL-047
   replace_existing) with his framing, three prompts to try, and Pair A verbatim (prompt / untrained /
   fine-tuned / what the person wrote) UNDER the box; the box's contract byte-identical. **When it lands:**
   cancel the stock `needs_content_page` twin the generator queues (watcher armed; last time it appeared
   at completion), wait for the `page_rerender`, then verify: six section hashes unchanged
   (snapshot in NOTES 21:10Z), Pair A text present verbatim in the served page, `cdp_chat_probe.py`
   still PASS. (6) **Homepage design request (22:15 BST, verbatim):** *"the copy on the home page is
   much better now. Can we ask one of the design related or experience or component agents to tidy up
   the components and use more interesting ones for the cards, probably different carousel like
   structures. Please ask them to be imaginative, research good alternatives and apply them."* Routed:
   see NOTES 21:20Z (the editorial_design_uplift lane / design-critique-agent — whichever took it).
   **SPLIT 21:28Z (uplift lane's CONTRIB in this dir, `a85bcedea`): this lane chooses and applies the
   card/carousel components on /index.html — AFTER reading the critique report; resolve by FUNCTION not
   name (section_type ≠ function on all seven); count the three placement rows; canary case-studies-grid
   ALONE with the words byte-identical at the served page; six of seven candidates are
   render_mode='agent'. The uplift lane does the infographics: hand-authored `site_plan_imagery` rows at
   kind='infographic' for this page's concept sections (the planner prompt says "most plans will have
   zero"; fleet census infographic=1; a fleet prompt change is the planner owners' call). Three
   construction constraints (no funcmap arithmetic; no words in SVG; non-text contrast) in NOTES 21:30Z.
   `design_critique_run` `204f1ff7` filed 21:16Z → report in `doc_notes` categories 'design-report'.
   Widget replace `e1b2bcf8` LANDED 21:18Z and is **SERVED 21:32Z** (pair text verbatim on the page, box
   intact, `cdp_chat_probe.py` PASS with a page-suggested prompt). Duplicate heading string "What to try"
   (widget + third booking section) — cosmetic, fold into the merged brief.**
   Homepage today: hero, features, differentiators-section, case-studies-grid (17.7 KB), departments-grid,
   call-to-action; library carousel-like sections: hero-card-carousel, swipeable-insight-carousel,
   image-hover-card-grid, teaser-reveal-panel, info-card-grid (carousel defaults ON at resolution).
0a-0. ~~OWNER DECISION PENDING~~ **DECIDED 22:05 BST: improve the model first** — then REVERSED at 22:10 to "it can still go up, explain it away" (see 0a-00); the training plan stands, the publication happens now with the honest framing. ("I don't have enough of my
   own writing so I will try and find someone else's that has a defined character"). Comparison NOT
   published; page copy untouched until the new model. The next run's checklist (rights, pairs at
   volume, echo fix, held-out eval, THEN publish) is in NOTES 21:05Z and PLAN. Waiting on the corpus. The measured truth
   (`COMPARISONS_2026-09-03_base_vs_finetune_demo_model.md`): the Phase 0 fine-tune barely moved the
   model, echoes on two tasks, degenerates on one. The page's missing "what to type / what you get"
   explanation is written only after his answer, because the framing sentence is his. Base model
   `smollm2:1.7b` is on the in-cluster ollama-adapter for comparisons. The echo behaviour goes to the
   training side before any further run (data-boundary suspect; RESULTS 08-15 never measured held-out).
0c. **A chassis roll was announced by the owner at 23:10 BST 09-03 ("in the next hour").** Before any
   dispatch: pods older than 300 s (`kubectl get pods -l app=agent-chassis`), and read the site's
   `page_rerender` queue (the 20:17Z fan-out) for an item the roll interrupted. Nothing of this lane's
   rides the roll.
0d. **HOMEPAGE CARD CANARY — OWNER SAID GO (23:45 BST: "case studies can be swipeable. For any number
   greater than 3") — fire AFTER the announced chassis roll settles (pods ≥300 s).** The served slot
   renders FIVE cards (5 images, 5 links to /case-studies.html; content_data card1–card5), not the
   critic's four. His RULE (>3 cards → swipeable) also names `features` (6, "solid") and
   `departments-grid` (5): canary first, then ask him before touching those two. Infographics are
   FLEET-WIDE by his decision (prompts lane CONTRIB); no hand rows on this page. Original plan text: (critique in
   `DESIGN_CRITIQUE_2026-09-03_finetuning_uk.md`; mechanism in NOTES 22:05Z). Slot `case-studies-grid`
   (four cards → 3+1 orphan) → `swipeable-insight-carousel` (agent-rendered; contract `cards,
   section_title, section_eyebrow`). Steps: snapshot the slot's `content_data` + rendered_html; write a
   deterministic mapper (flat `cardN_*` → `cards[]`, verbatim); count and update the THREE placement rows
   by FUNCTION; render; assert every title/excerpt byte-identical at the served page + orphan gone +
   swipe works (CDP); rollback = archived row + rerender. Alternative the owner may prefer: three or
   six case studies (content decision), no swap. Do NOT touch `features` (solid) or `differentiators`
   (the site's strongest device; the uplift lane's infographic sits with it). Site-wide monotony and the
   hero image reused on five pages are composition/imagery questions, not this lane's card swap.
0f. **Infographics on the homepage are BLOCKED on a structural fact, not on the prompt (12:40Z 09-04):**
   `site_plan_imagery` hangs off `site_plans`, and finetuning.uk has 0 plan rows (plan-less build) — the
   site cannot hold section imagery at all. Fleet-wide: 21 sites could (plan + facts), 0 of them planned
   imagery since 718; the 7 that did could not — the mechanism is UNTESTED, not broken; no prompt edit is
   indicated. The experiment moves to a capable site (uplift/prompts lanes). **Owner question, when the
   time comes:** a site plan for finetuning.uk (what it would regenerate, what it would protect).
   Tonight's chain of confident wrong causes is in NOTES 11:50Z–12:40Z and WRONG_CALLS 09-04.
0e. **Imagery (uplift lane's half, owner's yes needed):** 35 of 38 heroes are `hero.jpg`; IMG-077 items
   `6db67bde` (4 pages `unwired`: use-cases, case-studies, approach, contact — **do NOT hand-wire**: the
   mechanism `wirePageHeroOnLanding` is BUILT and in v1.0.1359 behind opt-in `wire_hero_on_landing`,
   named by zero live rows; arming it is the `bugfix_114_imagery_wiring` lane's (412 §10); migration 664
   hand-wired these pages 08-26 and decayed 9 → 3 in eight days — their CONTRIB in the 114 lane's dir) and
   `d280a6fd` (6 `no_image_slot` tool/guide pages — leave; the 686 double-image trap). Ten stale
   `image_url_404` rows on the case-study cards CANCELLED after a probe with a control; `empty-src` left.
0a. **Playground follow-ups, in order:** (i) a criteria fence for `tool-playground` so TL-013's
   ladder grades it (the brief's ACCEPTANCE list is the fence's content); (ii) a multi-turn probe
   run (the CDP probe sends one message; the route and widget are multi-turn); (iii) the owner's read
   of the widget's static copy on the served page; (iv) the guide page decision (`/guides/playground-
   guide.html`, deployed, unlinked); (v) the merged playground brief (tool at the centre + booking
   copy) as an owner-read rebuild — NOT the generator's stock brief, which was cancelled.
0b. **your-own-model Stage B**: read item `11e1e8ed` (attempts, error) and `save_refused_incomplete`
   `3034678c`; if it failed 3×, take the choice to the owner (443 section above). Then re-run
   `stage_b_assert.sh <before-dir>` (before-snapshots are in the 2026-09-03 session's scratchpad ONLY —
   re-take them from `pages.content_hash` + served bytes if needed).
0c. **Owner's evening answers** (PLAN "Direction, DECIDED"): catalogue shape drafted
   (`DESIGN_2026-09-03_examples_catalogue_shape.md`) — wants his line-by-line reaction; leopardess told
   (`docs/leopardessconsulting/CONTRIB_2026-09-03_…`); pricing by GPU class as a CHOICE, number open.
1. **641 is OPTION A4 (19:15Z, prompts lane)** — the subject prints verbatim as the section's OPENING
   LINE in page voice. **DONE 19:20:58Z: the three `section_subjects` arrays are RE-AUTHORED and
   APPLIED** against the spec (`CONTRIB_2026-09-03b_…_subject_phrasing_spec.md`) by
   `SQL_2026-09-03_section_subjects_A4_reauthor.sql` (previous values in its header; five spec
   guards mutation-killed; playground §4 is the owner's exemplar verbatim). Stage B's acceptance
   wording under A4 is in the Stage B file's header: subject wording in each opening line is the
   INTENT; a section opening on a SIBLING's subject is the failure. G1's marker survives A4. 641
   itself: edited to A4, rehearsed, council round 2 on `6c92d154`; the owner's read of the bytes
   is the last gate.
1b. **When the prompts lane applies 641:** Stage B (above). ~~Rewrite the technical-details brief first.~~
   **DONE 2026-09-03:** `technical_details_stage_b_dispatch.sql` in this dir carries the rewritten
   brief and REFUSES to run until 641 is on the live writer (G1). Rehearse its post-G1 path under
   ROLLBACK first (header says how — it is unrehearsed as a transaction); then run, then
   `your-own-model` by the RUNBOOK recipe.
2. **The details conversation** the owner deferred: pricing (×5 of which GPU class), the examples
   catalogue's data model (owner, model artefact, before/after, price), what moves to leopardess.
   Bring numbers: a6000 $0.35/hr, a100xl $1.09/hr, ×5 ⇒ £1.40 vs £4/hr; the Hetzner demo box is
   already paid for.
3. Optional: ask 389 for the section-editor pass on `llm-cost-calculator`'s `tool-cta`.
4. `our-position-on-ai` subjects (needs_rebuild page; derive from the served page when rebuilt).

## Still true from 2026-08-30 / 09-02

Terms and privacy locked · 9 hero images wired · copy quality with `copy_quality_two_stage` · the
three voice datasets small (26/13/16) · Stripe his, and last · datasets page unbuilt by choice.
