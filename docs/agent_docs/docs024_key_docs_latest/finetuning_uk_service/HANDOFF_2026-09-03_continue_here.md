# HANDOFF 2026-09-03 — the site is now the playground; the demo model is live on a Hetzner box; the chat route is approved; two hand steps and a page stand between here and a working demo. Start here.

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
| 2b. `/opt/island/.env` five `PLAYGROUND_*` keys | **NOT DONE — operator hand edit.** My session's classifier refused to edit a live box's env; that is correct. Values in RUNBOOK. Inert until 2c. | island |
| 2c. island image swap (`aqls/tools-api:v1.0.1343` → a tag carrying the route) | **NOT DONE — OWNER'S CALL. Image BUILT + PROVEN 2026-09-03 16:52Z: `aqls/tools-api:v1.0.1359-playground`** (from `9b540c2e6`; symbol 2/2/0 on the extracted binary). `docker load` onto the island was refused by the session classifier; the owner's three lines are in RUNBOOK step 2. `make build-tools-api-ref IMAGE_TAG=…` is the target (the `build-%-ref` pattern rule — "no makefile target" above was wrong). **The restart also serves robot-hands.com + vonc.com.** | island; RUNBOOK "step 2: ship" |
| 3. the chat widget on `/playground.html` | **NOT STARTED — and the plan here is CORRECTED (2026-09-03 17:10Z, measured at the library row).** ~~fork `chat-input-box` … repointed at the route~~ The library box is single-turn `{message}`→JSON, same-origin, path a literal, no endpoint field; the route is multi-turn `{messages}`, cross-origin, SSE (`token`/`done{truncated}`/`error`). A fork cannot be "repointed"; the JS must differ. Paths: **(a)** `add_tool` with `library_source: null`, the route contract in the description, function `tool-playground` + the generator's `adopt_existing_page` flag so it lands on `/playground.html` (name match) rather than minting `/tools/…`; **(b)** the estate's two live cross-origin widgets (`gripper-report-intake` mig 651, `gauntlet-round-record-vonc-com`) are hand-written `js_snippets` + section + locked row. **Owner's call; lane recommends (a), (b) as fallback** (README). `deploy_config.capabilities += backend` only if the tool carries `requires-backend`, and NOT before the route is live. | NOTES 16:45–17:15Z; PLAN Phase P |
| 4. booked-hour GPU provisioning as a workflow | not started; thunder actions exist for training runs (`thunder_ssh_exec_dispatch`, `_decommission_dispatch`, …) | PLAN |
| 5. booking → session handoff; the examples catalogue ("model pages" with an owner) | not started; "details later" | PLAN DIRECTION |

**Post-ship verification when 2c happens** (council round 4, debug_historian): probe the island
container for the SYMBOL `PlaygroundChatHandler` with `GripperChatHandler` as present-control and
an impossible string as absent-control (locally 2 / 2 / 0); `docker logs` for "playground route
group mounted"; then curl the route with `Origin: https://finetuning.uk` (200 text/event-stream)
and a wrong Origin (403). Recipe in RUNBOOK.

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

## 443 (repeated sections) — Stage A CLOSED, Stage B waits on 641

- **Stage A PROVEN** on `technical-details` (corr `6e8eadaa`, writer `ce514ce0`): all six
  `sections_ready[].subject` populated from `pages_table` (tier 3), carried to the writer's row;
  detector quiet with a fleet demand control of 7. `pages.section_subjects` backfilled for
  `playground`, `your-own-model`, `technical-details` (6/6 each; clause form, e.g. "what to have
  ready before the hour"); `our-position-on-ai` still null (no brief to derive from).
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

0. **If the owner has done 2b + 2c:** run the post-ship probes (RUNBOOK), then start step 3 on
   whichever path he chose (README 2026-09-03 evening; default (a) if he has not said). Checked
   2026-09-03 16:48Z: **neither done** (0 env keys, compose 1343, route 404 with controls).
   If still not: nothing on the playground can move; say so plainly.
1. **When the prompts lane applies 641:** Stage B (above). ~~Rewrite the technical-details brief first.~~
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
