# LANDMINES — traps that mislead you BEFORE you have a symptom

**Fleet-wide, append-only, newest at the bottom. Any thread may append.**
Created 2026-07-29 at the owner's direction, alongside `WRONG_CALLS.md`.

A landmine is **a thing that will quietly mislead you the moment you touch a
particular file, table, command or service** — where the wrong result looks
exactly like the right one, so nothing prompts you to check. You do not arrive
here because something broke. You arrive here **before** you act, because you are
about to touch the thing it guards.

---

## The test for an entry, and why it is strict

> **Would a session touching this, with no symptom and no suspicion, get it wrong
> without this entry?**

If it needs a symptom first, it is not a landmine — it is a failure pattern, and
it belongs in `016b` §9. If it is the story of a claim that turned out false, it
belongs in `WRONG_CALLS.md`. **This file is the middle rung: the distilled check,
attached to the thing it guards.**

~~That rung has never had a home.~~ **CORRECTED 2026-07-29, hours after this file
was created — it had one, in Postgres.** `doc_notes` carries a **`landmine`
category with 7 rows**, written 2026-07-27/28 by two other threads (*architecture
council 2*, *bugsearch-thread*), footprinted by action name
(`subject_type='action'`, `subject_key='index_code_symbols'`, `scrape_web`,
`store_business_verification`, …). **So this file was created alongside an
existing store, not into a vacuum**, and which one is the system of record is
exactly what D10 has to decide. My prior-art search grepped the filesystem; the
prior art was in the database. Same shape as `cmd/contrastscan` below — a search
whose *filter* defined its answer.

`PROPOSAL_D9_landmines_as_a_footprinted_corpus.md` (open as **D10**) names the
gap and measures its cost: with no *agreed* home, authors put landmines in the
auto-loaded `MEMORY.md`, because that is the only file guaranteed to be read —
which is why the index was compacted twice inside one hour on 2026-07-26 and
re-inflated past its starting size in 46 minutes. Measured 2026-07-29: **18
LANDMINE lines in `MEMORY.md` (21.8KB against a 25,000-byte cap) and 22 more in
`MEMORY_workstreams.md`.**

> **✅ D10 RULED (owner, 2026-07-29): THIS FILE IS AUTHORITATIVE.**
> Landmines are written here, in markdown, and **synced into `doc_notes`** by
> `scripts/landmines-sync.py` so that machines — council seats above all — can
> read them. That inverts the proposal's rows-first preference deliberately: a
> markdown append costs ten seconds, needs no cluster, and works in a fresh clone
> or with an expired kubeconfig, and **an unadopted ledger is the exact failure
> this item exists to fix.**
>
> Consequences you need to know:
> - **Append here. Never hand-write a landmine row into `doc_notes`** — the sync
>   owns those rows and will replace them.
> - The **7 pre-existing `doc_notes` landmine rows** (written 07-27/28 by two other
>   threads, before there was a decision) are being folded into this file. They are
>   not discarded, and they are not yours to delete.
> - **The 40 `MEMORY*.md` LANDMINE lines stay put** until delivery is working —
>   owner ruling, following the proposal's §5 staging rule. Duplication is correct
>   for now; do not "tidy" it.

**Routing, so this does not become a fifth pile:**

| you have… | it goes… |
|---|---|
| a symptom, and a mechanism that failed | `016b` §9 |
| a claim you wrote that turned out false | `WRONG_CALLS.md` |
| a trap that fires on *touching* something | **here** |
| a durable defect in production | `/bugs_open/` |
| a mechanism another workstream could call | the concept register |

## How a landmine reaches you — all three routes, live 2026-07-29

D10(c)+(d) ruled: **all three**, because the proposal recommended against
discipline standing alone and this is what stops it standing alone.

1. **Automatically, at session start.** `scripts/landmines-session-start.py`
   (wired in `.claude/settings.json`) matches entries against the files already
   dirty in the working tree and injects them. Path-shaped footprints only, capped
   at 6, and it exits silently on any error — a startup hook that can break a
   session would be worse than any landmine it reports. **New sessions only:** an
   already-running session needs `/hooks` or a restart.
2. **To every council seat**, in the `schema_hint` they already consume —
   migration **271**, live 2026-07-30. The seats get `doc_notes`' columns plus a
   compact index (one line per `subject_key`) and the query to pull a full body.
   **Index, not bodies**, and it costs +5,558 bytes per seat prompt (measured):
   fine at ~16 entries, **not fine at 100**, so relevance-gating is the next
   increment — proven read-only (16 → 2 on a real submission) but unshipped
   because `load_schema_hint` has no `error_step`. Full working in the migration
   header.
3. **By grep, deliberately** — still the route for table, command and symbol
   footprints, which cannot match a file path:
   `grep -n "<path-or-table>" docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`

**Do NOT drain `MEMORY.md`'s landmine lines into here.** Owner ruling, following
D10 §5: moving them before delivery is proven removes protection we have today.
**Both run in parallel** — duplication is correct, not untidiness.

### Keeping the rows in step

The markdown is authoritative; the rows are derived. After appending, run:

```bash
./scripts/landmines-sync.py            # dry run — what would change
./scripts/landmines-sync.py --apply    # write
./scripts/landmines-sync.py --check    # exit 1 if the DB has drifted
```

It owns only rows whose `source` starts `LANDMINES.md#` and replaces those; the
7 landmine rows written by other threads before D10 was ruled are structurally out
of its reach. **Consumers should filter on `categories ? 'landmine'`, not on
`subject_type`** — the category spans both conventions.

## Entry format — and why it looks like this

Each entry carries an explicit **footprint**: the path, table, symbol or command
it guards. That is not decoration. If D10 is decided in favour of `doc_notes`
rows, every entry here converts mechanically — `footprint` → `subject_key`,
`subject_type='landmine'`, the body → `body`, tags → `categories`. **This file is
written to be a staging format for that corpus, not a competitor to it.**

```
### <imperative one-liner: what will mislead you>
- **footprint:** `the/path`, `the_table`, `TheSymbol`   ← what you must be touching
- **fires when:** the action that walks into it
- **the tell:** what you will see (usually: something that looks correct)
- **the check:** the discriminating command or query
- **source:** bug file / WRONG_CALLS date / workstream that learned it
- **added:** YYYY-MM-DD, <lane>
```

Keep entries short. If it needs three paragraphs, the paragraphs belong in the
source document and the entry points at it.

---

# Entries

### `agent_definitions` has no `name` column — it is `display_name`

- **footprint:** `agent_definitions`
- **fires when:** writing any ad-hoc SQL against the agent table
- **the tell:** `ERROR: column "name" does not exist` — cheap, but it costs a round
  trip against a cluster every time, and the query *reads* perfectly
- **the check:** `\d agent_definitions` first. CLAUDE.md already says "schema
  first"; this is the table people skip it on, because the query feels too small
- **source:** `webdesign_uk_build_service/NOTES` 2026-07-28
- **added:** 2026-07-29, webdesign.uk lane

### There is no `postgres` role on `postgres-clients-0`

- **footprint:** `postgres-clients-0`, `clients_db`
- **fires when:** reaching for the conventional superuser to run `\l`, `\du`, or
  anything administrative
- **the tell:** `FATAL: role "postgres" does not exist` — reads like a broken pod
  or a permissions problem rather than a wrong username
- **the check:** use `-U clients_user` for everything, including `\l`
- **source:** `webdesign_uk_build_service/RUNBOOK`, 2026-07-28
- **added:** 2026-07-29, webdesign.uk lane

### The Bash tool's working directory persists between calls — a `cd` outlives the command that made it

- **footprint:** `Bash`, `git commit`
- **fires when:** you `cd` into a subdirectory to run one convenient command, then
  later run `git commit <repo-relative-path>` in a *separate* call
- **the tell:** `error: pathspec '<path>' did not match any file(s) known to git`
  on a path you can see with your own eyes. It reads as "my file is untracked" or
  "another session moved it" — both wrong, and both send you looking at git
- **the check:** `pwd` before any command that takes a repo-relative path; or
  simply always pass absolute paths. Note this bites *hardest* with the
  commit-by-pathspec rule, which is mandatory here, so the two interact
- **source:** hit directly, `webdesign_uk_build_service` session 2026-07-29
- **added:** 2026-07-29, webdesign.uk lane

### `cmd/contrastscan` does not exist — the contrast tool is Python

- **footprint:** `cmd/`, `scripts/render_audit.py`
- **fires when:** planning any contrast/accessibility check, or reading a memory
  line or doc that names `contrastscan`
- **the tell:** nothing — the name appears in several documents and in the auto
  memory index, and reads as settled fact. It was built and **deleted the same
  day** (2026-07-28) as a duplicate
- **the check:** `ls cmd/`. More generally: a path recalled from memory is an
  `[UNVERIFIED]` claim wearing settled clothes. The register entry is VIZ-010
- **source:** `WRONG_CALLS.md` 2026-07-28; register `visualisation-and-charts.md`
- **added:** 2026-07-29, webdesign.uk lane

### A "per-IP" limiter behind Cloudflare is probably one global bucket — and `httpguard` reads as the fix

- **footprint:** `internal/tools-api/middleware/ratelimit.go`, `platform/httpguard`,
  any service behind Caddy/nginx + Cloudflare
- **fires when:** relying on a per-IP rate limit to bound spend or abuse; or
  "fixing" one by adopting httpguard
- **the tell:** it works perfectly from your machine. The island's stored
  `client_ip_hash` was `sha256("172.18.0.1")` — the docker gateway — in **83 of 83
  rows**. Caddy overwrites `X-Forwarded-For`; Cloudflare strips `X-Real-IP`;
  httpguard's rightmost-XFF fallback lands on the same constant
- **the check:** `count(DISTINCT <ip key>) > 1` **from two different networks**.
  One test machine cannot distinguish a constant from a working key. The real
  address is in `CF-Connecting-IP` only
- **source:** `bugs_open/139` (gauntlet/island lane — theirs, not mine; do not fix
  it here, contribute there)
- **added:** 2026-07-29, webdesign.uk lane — imported, not independently reproduced

### Switching a lane to `claude-fable-5` is not a config-only change

- **footprint:** `agent_definitions.default_config`, `claude-fable-5`, `platform/aiservice`
- **fires when:** pointing any agent at `claude-fable-5`. Note DB model config is
  **live on write** — there is no build step to slow a mistake down
- **the tell:** `400 invalid_request_error` on a payload that is obviously valid.
  Fable rejects `temperature`, `top_p`, `top_k`, and both forms of `thinking`
  (it is always on, and explicitly disabling it also 400s). Separately, an org on
  **zero data retention gets 400 on every Fable request** — and the error names
  the request, not the org setting, which is how it costs an afternoon
- **the check:** before switching — (1) confirm org retention ≥30 days;
  (2) `grep -rn "temperature\|top_p\|top_k\|budget_tokens\|\"thinking\"" --include="*.go" platform/ internal/`
  **and** check `agent_definitions.default_config`, since params live in DB rows
  too (all 16 council seats set `max_tokens` from config)
- **source:** `claude-api` skill, read 2026-07-29;
  `webdesign_uk_build_service/PLAN` §7b
- **added:** 2026-07-29, webdesign.uk lane — from documentation, not yet hit in
  production, because **nothing on the fleet runs Fable yet** (0 calls in 4 days)

### `git checkout -- <path>` restores from the INDEX, not from HEAD

- **footprint:** `git checkout`, `git checkout HEAD --`
- **fires when:** you `git add` a file, then try to undo the change with
  `git checkout -- <path>`. It reports success and restores **the staged
  mutation**, not the committed content
- **the tell:** none at the point of failure — the command is silent and exits 0.
  It surfaces later as a *different* test giving the wrong answer, because the
  file you believed was pristine is not. Hit directly 2026-07-29: a negative
  control reported the guard "wrongly fired on an append" when what it had
  actually seen was the un-restored deletion from the previous test
- **the check:** `git checkout HEAD -- <path>` when you mean HEAD, and prove the
  restore rather than assuming it: `git diff HEAD --stat -- <path>` should be
  empty, or compare `git show HEAD:<path> | md5sum` against the file
- **the wider point:** *a contaminated control does not announce itself — it
  announces a bug in the thing you were testing.* Restore, then verify the
  restore, then run the next control
- **source:** hit directly while positive/negative-controlling the
  `shared-ledger-not-appended` guard, 2026-07-29
- **added:** 2026-07-29, webdesign.uk lane

### `pages.sections` is a materialised CACHE — the build reads `site_plan_sections`

- **footprint:** `pages.sections`, `site_plan_sections`, `pages.title`,
  `site_plan_pages.title`, `load_page_sections_from_spec_action.go`
- **fires when:** changing which sections a page has, or a page's title, by
  updating the `pages` row — the obvious place, and the one the page itself reads
  back to you
- **the tell:** the `UPDATE` reports `1`, the row shows your value, the rebuild
  reports `complete`, and the page comes back with the OLD section list. Nothing
  errors. I swapped a homepage section this way and the rebuild silently restored
  the previous component
- **the check:** after any `pages` edit, ask what regenerates it:
  `SELECT sps.page_name, sps.ordering, sps.component_name FROM site_plan_sections sps
   JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.site_id='<id>' AND sp.is_current;`
  Fix the plan table too, or instead
- **the wider point:** the chain is `site_plans` → `site_plan_pages` /
  `site_plan_sections` / `site_plan_imagery` → `pages` → `page_components`.
  **Fixing every reader of a value is not fixing it.** Ask which table
  *regenerates* the one you just corrected, and follow it until the answer is
  "nothing"
- **source:** `dartsonline_traffic/SQL_2026-07-29s`, WRONG_CALLS 2026-07-29
- **added:** 2026-07-30, dartsonline_traffic lane

### A hand-made `page_rerender` item needs `page_id` — in the spec AND the column

- **footprint:** `site_work_items` where `item_type='page_rerender'`,
  `rerender_single_page`
- **fires when:** raising a rerender by hand, having just raised a `needs_page`
  and copied its spec shape. `needs_page` resolves by `page_name`; this does not
- **the tell:** `step render_page failed: failed to execute action
  rerender_single_page: page_id not found in input` — and it retries to
  `attempt_count = 3` before it stops, so it reads like a flaky handler rather
  than a malformed item
- **the check:** copy the spec shape from a **completed row of the same
  item_type**, not from the action's source or a sibling type:
  `SELECT jsonb_pretty(spec), page_id FROM site_work_items
   WHERE item_type='page_rerender' AND status='complete' LIMIT 1;`
  It wants `domain`, `page_id`, `page_name`, `filename` — and `page_id` in the
  column as well
- **source:** hit directly, `dartsonline_traffic` 2026-07-29 (2 items, 3/3 attempts each)
- **added:** 2026-07-30, dartsonline_traffic lane

### `snapshot_agent` has TWO overloads writing to TWO different tables

- **footprint:** `snapshot_agent(text)`, `snapshot_agent(text, text)`,
  `agent_definitions`, `agent_definitions_backup`
- **fires when:** snapshotting an agent before a config change, then checking the
  snapshot exists
- **the tell:** the two-arg form prints `NOTICE: Snapshot captured: type=…,
  source_id=…` and returns a uuid, so it looks unambiguous — but it writes to
  **`agent_definitions_backup`**. The one-arg form inserts an `is_snapshot = true`
  row into `agent_definitions`. Check the wrong table and a successful snapshot
  looks like a silent no-op: I found **0 `is_snapshot` rows fleet-wide** for the
  agent and was one step from filing it
- **the check:** don't ask whether a snapshot exists, ask whether it holds the
  **pre-change** config — a row carrying the post-change value restores nothing:
  `SELECT snapshot_taken_at, snapshot_reason,
          (default_config #>> '{workflow,steps,…}') LIKE '%<the old literal>%' AS has_old
   FROM agent_definitions_backup WHERE type='<agent>' ORDER BY snapshot_taken_at DESC LIMIT 1;`
- **source:** hit directly, `dartsonline_traffic/SQL_2026-07-30t` 2026-07-30
- **added:** 2026-07-30, dartsonline_traffic lane

### A trailing `ROLLBACK;` protects nothing unless the `BEGIN;` at the head is live

- **footprint:** any `docs/agent_docs/sql_for_agents/*.sql` in the house
  "dry-run-by-default" style; `psql -f`
- **fires when:** you follow the convention of commenting out `BEGIN;`/`COMMIT;`
  and leaving `ROLLBACK;` as the safe default — or you inherit a file written that
  way and run it expecting a dry run
- **the tell:** psql prints your verification output and then
  `WARNING: there is no transaction in progress` — one line, at the end, after
  everything you were reading. Every statement has already autocommitted. **The
  "safe default" committed the change while appearing not to.** I wrote a file in
  exactly this shape and only got a real dry run because I happened to uncomment
  BEGIN for it
- **the check:** if the first non-comment statement is not `BEGIN;`, there is no
  transaction. Either uncomment it, or wrap:
  `(echo 'BEGIN;'; cat file.sql; echo 'ROLLBACK;') | psql …`
- **source:** `dartsonline_traffic/SQL_2026-07-30t` 2026-07-30
- **added:** 2026-07-30, dartsonline_traffic lane

### A palette slot no LAYOUT declares is never emitted — deriving it is dead config

- **footprint:** `palette_specialised_slots.go` (`darkSchemeDerivations`),
  `layouts.css_template`, `render_css_from_spec_action.go`
- **fires when:** fixing a light-literal-on-a-dark-site defect by adding the
  offending slot to the derivation list. It is the obvious fix and it compiles,
  tests, and changes nothing
- **the tell:** none at all. The palette gains the slot, the log line says it was
  filled, and the rendered CSS still carries the component's literal — because the
  palette reaches the stylesheet **only** through `{{palette "X" "literal"}}` calls
  in a layout template. I added `icon_chip_bg` and measured afterwards: `card_bg`
  is declared by **18 of 18** layouts, `surface_alt` by 3, `icon_chip_bg` by **0**
- **the check:** before adding a derivation, confirm a layout asks for it:
  `SELECT count(*) FROM layouts WHERE css_template LIKE '%palette "<slot>"%';`
  And check how narrow the literal really is —
  `SELECT name FROM content_components WHERE html_template LIKE '%<var-name>%' AND is_active;`
  turned out to be **one** component, whose replacement already used a derived slot
- **source:** `dartsonline_traffic` 2026-07-29; council `bf208075`; `bugs_open/122`
- **added:** 2026-07-30, dartsonline_traffic lane

### A verification `ILIKE` over a config blob matches the prohibition you just wrote

- **footprint:** `site_specs.data`, `site_plan_imagery.prompt`, `pages.page_spec`,
  any `data::text ILIKE '%word%'` audit
- **fires when:** you fix a site by writing rules that NAME the forbidden words —
  `honesty_rails: "Never claim to stock…"`, `cta_style.never_use: ["Add to Bag"]`,
  an image prompt ending `"No packaging, no price tags"` — and then grep for those
  words to prove they are gone
- **the tell:** the count never reaches zero, and each remaining hit is your own
  corrective text. Reads as "the fix did not take". Hit **three times** on one
  site in one day
- **the check:** go per-key rather than over the blob —
  `SELECT key, value FROM site_specs, jsonb_each(data) WHERE …` — and **read the
  matching rows instead of counting them**. A count cannot tell a violation from
  its own prohibition
- **source:** `dartsonline_traffic/SQL_2026-07-29k`, `…p` 2026-07-29
- **added:** 2026-07-30, dartsonline_traffic lane

### `assets.url` is a presigned S3 URL with a 7-day expiry — never store or serve it

- **footprint:** `assets.url`, `queryresolve.go` (`pageImageCols.webPath`)
- **fires when:** wiring a generated image into a page, or writing a check that
  fetches an asset to confirm it exists
- **the tell:** it works today. The URL carries `X-Amz-Expires=604800`, so an
  asset row created more than a week ago returns 403 while the row still says
  `status='active'`. Half the rows on the site I looked at were already past it
- **the check:** use the deployed git path (`/assets/images/<key>.jpg`), which is
  what `webPath()` resolves and what pages must reference. If you fetch
  `assets.url` for inspection, do it immediately and do not record the result as
  the asset's address
- **source:** `queryresolve.go:292` comment ("Never `assets.url`"); confirmed while
  inspecting dartsonline's 33 assets 2026-07-29
- **added:** 2026-07-30, dartsonline_traffic lane

### A mutation that never happened is indistinguishable from a guard that works

- **footprint:** `sed -i 's|…|…|'` against Go source; mutation testing / "prove
  this test can fail" of any kind
- **fires when:** you break a guard on purpose to show a new test would catch it —
  and the guard's own source contains your sed delimiter. `s|peer.IsLoopback() ||
  peer.IsPrivate()|…|` has **four** `|`, so sed parses garbage, prints
  `unknown option to 's'`, and **leaves the file untouched**
- **the tell:** there isn't one. `go test` then prints a perfectly honest `ok`,
  and the reading available to you is *"the mutation did not break anything"* —
  i.e. exactly the conclusion "this guard is redundant", drawn from a file that was
  never mutated. The sed error scrolls past above the result you are actually
  looking at
- **the check:** make the mutator **assert its own anchor was found** before
  substituting (a 5-line Python script beats `sed -i` for anything containing
  operators), and **diff the file** before believing the run. Generally: a mutation
  test that passes is either evidence of a redundant guard or evidence of a failed
  mutation, and nothing in the exit code separates those two
- **source:** `robot_hands_gripper_dossier/NOTES_…` 2026-07-29 afternoon, while
  discharging a council objection on `platform/httpguard`. Same family as *"a quiet
  test passes when the RULE is gone, not when the guard works"*
- **added:** 2026-07-30, consolidation lane (`features_open/024`)

### Prose in another lane's live HANDOFF carries no timestamp — date the line, don't read it

- **footprint:** `docs/agent_docs/docs024_key_docs_latest/*/HANDOFF_*.md` belonging
  to another thread; the CONTRIB-into-their-directory convention
- **fires when:** you read another lane's cold-start doc to learn their current
  position on something — or you file a CONTRIB into their directory and assume
  landing it in the right place means it reached them
- **the tell:** none, and this is the whole problem. A live doc's prose reads as
  current and considered. Measured 2026-07-30: a gauntlet handoff item said *"the
  consolidation people may be in touch — nothing owed yet"*, and
  `git log -S` dated that line to **five hours before** the finished patch arrived
  in that same directory; it then survived **four** later edits of the file
  untouched. A stale line sitting in a cold-start path is indistinguishable from a
  decision to decline
- **the check:** `git log -S'<the exact line>' --date=format:'%m-%d %H:%M' -- <their
  file>` before quoting it as their position — one command. And when you file into
  another lane, **append a dated pointer where their next session actually reads**
  (their cold-start doc, their words untouched), then record that you did. Filing in
  the right directory is **authoring**; nothing in the system performs **delivery**
  — the same gap D10 was opened for, on a different corpus
- **source:** `consolidation/HANDOFF_2026-07-30_continue_here.md` §4; the patch is
  `gauntlet_dead_cta/CONTRIB_2026-07-29_tools_api_client_identity_is_a_constant.md`
- **added:** 2026-07-30, consolidation lane (`features_open/024`)
- **UPDATE same day, ~16:30 — both halves now measured, and the second is a second
  landmine.** (a) The fix works: appended to their cold-start doc ~13:40, **read 14:12,
  in front of the owner 15:16** — versus a day of nothing for the same content sitting
  in their directory. (b) **Do not measure another lane's uptake by `git log` or by
  their directory.** I did, concluded "not delivered", and was wrong: a session that
  reads, decides and reports to the owner commits **nothing**. The artefact is their
  **transcript** — `~/.claude/projects/<project>/<session-id>.jsonl`, with
  `customTitle` on line 1 — where a read shows as an `attachment`/`user` record and the
  decision shows as assistant text. Grep it for your own filename before claiming
  silence

### `agent_definitions_backup` keeps the SOURCE row's `id` and `created_at` — order by `snapshot_taken_at`

- **footprint:** `agent_definitions_backup`, `snapshot_agent`
- **fires when:** snapshotting an agent before a config change, then trying to find
  "the snapshot I just took" — to diff against it, or to restore it
- **the tell:** **every backup row for one agent shares the same `id` and the same
  `created_at`**, because both are copied from the source row. `ORDER BY created_at
  DESC LIMIT 1` therefore returns an *arbitrary* snapshot — for `council-gate` on
  2026-07-30 it returned a 17 July one. The uuid `snapshot_agent()` prints back is
  the SOURCE row's id too, so it does not identify your snapshot either. The
  failure is silent and it lies in the worst direction: the diff came back showing
  **16 steps changed** when exactly one had, which reads as "I have broken the
  council gate"
- **the check:** order by **`snapshot_taken_at DESC`**, and pass a distinctive
  second argument to `snapshot_agent(type, reason)` so you can find yours by
  `snapshot_reason`. There is an index for precisely this:
  `(type, snapshot_taken_at DESC) WHERE snapshot_taken_at IS NOT NULL AND restored_at IS NULL`.
  Then diff step-by-step (`jsonb_each` + `IS DISTINCT FROM`) rather than trusting a
  whole-blob comparison
- **source:** hit directly while applying migration 271, 2026-07-29/30. Sibling of
  the `snapshot_agent` two-overloads entry above — same table, different trap
- **added:** 2026-07-30, webdesign.uk lane

### `strings` splits a Go literal at every non-ASCII byte, so an em-dash marker greps to 0 in a binary that contains it
- **footprint:** `strings /app/agent-chassis`, any pod-verification of a deploy, `bugs_open/153`
- **fires when:** you pick a marker string from your own Go source to prove a build shipped, and the literal contains an em dash, a curly quote or any other non-ASCII character — which house style makes likely, because our error messages use em dashes
- **the tell:** `grep -c "<your marker>"` returns **0** on an image you just built from the commit that added it. Indistinguishable from the 153 case (tag bumped, never rebuilt) — you will conclude your fix did not ship
- **the check:** pick an **ASCII-only** fragment of the literal, and prove the method rather than the result: run a positive control (a string you did NOT add, e.g. `CONTENT_LINK_REPAIR_DETAIL`) and a negative control (a string that exists nowhere) in the **same** exec. Measured 2026-07-30 on `v1.0.1208`: the em-dash form 0, the ASCII fragment `silently omitted it` 1, positive control 1, negative control 0
- **source:** hit while verifying `bugs_open/149` C1+B4 into `v1.0.1208`, 2026-07-30
- **added:** 2026-07-30, oufe lane

### `grep` in the Claude Code shell is a ugrep wrapper with `-I`, so one binary byte makes a whole file return zero matches and print nothing
- **footprint:** `grep` (the shell function, not `/bin/grep`), any scan output containing site copy — `cmd/claimscan` output especially
- **fires when:** you grep a file that contains non-UTF-8 bytes anywhere in it. Site copy routinely does
- **the tell:** **`grep -c` prints NOTHING AT ALL** — not `0` — and exits 1. Real grep always prints a number, so a blank where a count should be IS the diagnosis. `LC_ALL=C` does **not** fix it; the `-I` (skip-binary) flag in the wrapper is what does it
- **the check:** `type grep` to confirm the function, then use `command grep -a` to bypass it. Re-run any "clean" scan you got from a bare `grep` on machine output
- **source:** hit 2026-07-30 scanning the fleet claims corpus; refines concept-register `CLM-014` landmine (3), whose stated fix (`LC_ALL=C grep -ac`) is insufficient in this shell
- **added:** 2026-07-30, oufe lane

### `kubectl exec` truncates a large export mid-stream, and the short file looks complete
- **footprint:** `kubectl exec -i postgres-clients-0 -- psql`, any bulk export to a local file
- **fires when:** you export more than a few hundred rows (base64 HTML blobs especially). Two of fourteen sites came back short on one run — 58 of 139 rows, 30 of 67
- **the tell:** `Copying stdout failed: read message: unexpected EOF` on **stderr**, easy to miss if you tail stdout, and the file is **non-empty and well-formed** — every downstream tool accepts it and reports a clean result on a partial corpus
- **the check:** count the rows in the DB with the same predicate first, assert the export matches, and **retry until it does** — one site needed three attempts (77, 113, 139). Never scan an export you have not counted
- **source:** hit 2026-07-30 exporting `page_components` for the claims blast-radius measurement; new mode, not in `CLM-014`'s existing list
- **added:** 2026-07-30, oufe lane

### A `jsonb_path_query($.**.checks)` over `agent_definitions` returns another action's vocabulary too
- **footprint:** `agent_definitions.default_config`, `run_discovery_checks`, `scan_sites_for_maintenance`
- **fires when:** you enumerate configured discovery-check names across the fleet to find dead or unregistered ones
- **the tell:** three names (`stale_pages`, `missing_content`, `orphan_nav`) resolve in no discovery-check registry and look like dead config. They belong to **`maintenance-triage`**, which has no `run_discovery_checks` step at all — the array is `scan_sites_for_maintenance`'s. Filing them would be a fabricated defect, and it would make a hard-fail-on-unregistered fix look unshippable
- **the check:** filter to agents that actually have a `run_discovery_checks` step before reading their `checks` array — there are three (`completeness-`, `design-`, `quality-discovery-agent`), not four. Same family as the existing trap that `default_config::text LIKE '%action%'` matches **prompt text**: `council-gate` and `fix-proposer` both "contain" `save_page_sections` and neither has the step
- **source:** `bugs_open/149` B4, 2026-07-30
- **added:** 2026-07-30, oufe lane

### `save_page_sections` can now REFUSE a save, so a green orchestration status no longer means the sections were written
- **footprint:** `platform/orchestration/actions/save_page_sections_action.go`, `save_sections_claims_guard.go`, `page_components`
- **fires when:** you rerender or rebuild a page whose stored copy carries a banned claim. Four agents that never had any claims check now have one: `pageflow-builder`, `page-rebuild`, `page-rerender`, `site-work-orchestrator`
- **the tell:** the step fails with `claims floor blocked: N banned claim(s)…` and the page keeps its OLD content — which is the correct outcome, but reads as a build regression if you do not know the floor exists. There were **four** refusing guards in this function before this one (ownership, content-regression, interactivity, locked-slot); it is now five
- **the check:** `SELECT context FROM agent_error_log WHERE error_code='CONTENT_CLAIMS_FLOOR_DETAIL'` names the page, the section and the matched text. Withdraw fleet-wide in seconds with `check_claims:false` (or `check_claims_fleet_wide:false`) on the step — DB config, live immediately, no roll. Measured population that can hit this: 3 of 949 live components, 2026-07-30
- **source:** `bugs_open/149` C1, concept register `CLM-018`, commit `f61dce806`
- **added:** 2026-07-30, oufe lane

### A criteria check type the running browser-runner binary does not know is SKIPPED, and an all-skipped fence PASSES
- **footprint:** `internal/adapters/browserrunner/run_checks_action.go` (the `default: skip(ch.ID, ch.Type+" not implemented")` arm), `platform/orchestration/actions/experience_criteria.go`, any ```criteria fence in `doc_plans`, `has_visible_area` specifically
- **fires when:** you author a fence using a check type that was added to the repo but whose image has not rolled — which is the NORMAL state for hours after a new check type lands, and the report that introduces one will read as if it is usable ("what a claim can say today")
- **the tell:** there isn't one, and that is the whole entry. The check does not error and does not fail — it is **skipped**. Then the Tier-4 judge's `len(Failed)==0` yields a **PASS note + a 7-day cooldown** (the report's own G4, verified in code). So a fence whose checks all skip records a green acceptance verdict on a tool nobody has asserted anything about, and the cooldown then suppresses re-checking for a week. A new check type is at its most dangerous in exactly the window when a fresh report is urging people to author fences with it
- **the check:** before authoring a fence with any recently-added type, **grep the running pod, not the repo** — and pick a LONG marker, because Go compiles short string literals (`selector_count`, 14 bytes) to immediate comparisons that never reach rodata, so `grep -ac "selector_count"` returns **0 on a binary that fully supports it**. Use a distinctive sentence from the new arm plus a long pre-existing control in the same exec. Measured 2026-07-30 on `browser-runner-adapter-8646cddb79-qfcmr` (16h old, predating commit `1850acb07` at 15:19): `"too small to see or click"` **0**, `"A collapsed flex/grid child is the usual cause"` **0**, controls `"page overflows horizontally"` / `"but a parent CLIPS it"` / `"in the live DOM after settle"` **1 each**. Note the image has no `strings` binary — use `grep -ac` directly
- **source:** found reviewing `webdesign_tools_repair/REPORT_2026-07-29_concepts_for_a_working_tools_chain.md` §3/§7b against the live pod, 2026-07-30. The register entry (TL-034) correctly says status **`built`**; it is the report's §3 table ("today") and §7b ("enforced and not merely written") that read as deployed. The durable fix is the same one G4 already names: an unknown or wholly-skipped check set must be `inconclusive`, never a pass
- **added:** 2026-07-30, brochure_component_library lane (session `provenance step by step build tools`)

### `diagnosis_artifacts.kind` is CHECK-constrained to five values, so a new artefact kind fails at runtime and `go build` cannot see it
- **footprint:** `diagnosis_artifacts` (the `kind` column), `platform/orchestration/actions/diagnose_persist_fix_plan_action.go`, `diagnose_council_decide_action.go`, any new writer of a loop artefact
- **fires when:** you add an artefact kind for a new loop signal — a refusal note, a repair record, a per-iteration note. The obvious move is `INSERT … kind = 'my_new_kind'`, and it compiles
- **the tell:** there is none until the row is written, and then it is a constraint violation at the moment the new path first fires — i.e. on the failing branch, which is the branch nobody exercises before shipping. The constraint is `kind = ANY (ARRAY['bundle','iteration_note','fix_plan','council_report','escalation'])`; `\d diagnosis_artifacts` shows it, `go build` and every test with a mocked DB do not
- **the check:** `\d diagnosis_artifacts` before choosing a kind. Prefer the allowed-but-unused **`iteration_note`** slot with a `metadata->>'note_kind'` discriminator over DDL on a shared table — `iteration_note` was in the constraint with **0 rows and no Go reader** as of 2026-07-30, which is why `plan_validation_refusal` uses it. **Corollary, and the reason this is a landmine and not a note: any reader of `iteration_note` must now filter on `note_kind`,** or it counts refusal notes as whatever it assumed `iteration_note` meant. The one existing reader of arbitrary kinds (`fixloop_digest_action.go:229-245`) aggregates `kind || ':' || count` and is kind-agnostic, so it is safe — but it will start showing an `iteration_note:N` label that nothing wrote before
- **source:** `bugs_open/099` candidate 2, concept register `FIX-057`, 2026-07-30
- **added:** 2026-07-30, bugfix_099 lane

### A NULL `orchestration_id` never satisfies `= $2`, so a run-scoped artefact count returns 0 for ever and cannot bound a loop
- **footprint:** `diagnosis_artifacts.orchestration_id`, `nullIfEmpty(params.ExecutionContext.OrchestrationID)`, `diagnose_council_decide_action.go:514-517`, `diagnose_persist_fix_plan_action.go`
- **fires when:** you bound a retry/revise/repair loop by counting durable artefacts for this run — the established idiom here, and the right one — using `WHERE correlation_id = $1 AND orchestration_id = $2` with the id passed through `nullIfEmpty`
- **the tell:** none. SQL `NULL = NULL` is NULL, not true, so the predicate matches nothing, the count comes back **0**, and 0 always reads as "first attempt". The loop does not error, does not warn, and does not terminate — it just keeps granting rounds. A count of 0 is indistinguishable from a genuinely first attempt
- **the check:** assert the run id is non-empty **before** relying on the count, and fail closed if it is absent (a bookkeeping failure must not hand out extra LLM rounds). `diagnose_persist_fix_plan` does this as of 2026-07-30. **`diagnose_council_decide_action.go:514-517` has the same `orchestration_id = $2` shape and no such guard — [UNMEASURED] whether its `OrchestrationID` is ever empty in practice, so this is a shape to check, not a filed defect.** Scoping by correlation alone is not the answer: the correlation belongs to the DIAGNOSIS and accumulates across re-runs, which is the bug `FIX-033` already fixed once
- **source:** found while implementing `bugs_open/099` candidate 2, 2026-07-30
- **added:** 2026-07-30, bugfix_099 lane

### `llm_call_log.agent_type` was RELABELLED on 2026-07-26, so any per-agent measurement across that date silently splits one population into two
- **footprint:** `llm_call_log` (the `agent_type` column), any `WHERE agent_type = '…'` or `GROUP BY agent_type` over a window spanning 2026-07-26, `step_name LIKE 'review_%'`
- **fires when:** you measure anything per agent from the LLM log over the last week or two — cost, latency, truncation rate, token headroom, model mix. The natural query is `WHERE agent_type='<the agent you care about>'`, and it returns rows, and they look fine
- **the tell:** **there is none.** Calls made through the generic chassis worker logged `agent_type='generic'` until 2026-07-26 ~14:54; the same calls log their resolved agent type from ~15:03. **1,836 rows in the 16 days to 07-30 carry `generic`, and the last one is 07-27.** So a 14-day query filtered to a real agent name quietly drops everything before the cutover and reports a fortnight's figure computed from four days. For the council seats specifically it is worse: `fix-proposer` has **never** appeared in this column at all, so `WHERE agent_type='fix-proposer'` returns zero rows for a council that has run hundreds of reviews, and `agent_type='council-gate'` discards 1,798 review rows of the same population
- **the check:** before trusting a per-agent aggregate, run `SELECT date_trunc('day',created_at)::date, count(*) FILTER (WHERE agent_type='generic'), count(*) FROM llm_call_log WHERE created_at > now() - interval '16 days' GROUP BY 1 ORDER BY 1;` — if `generic` is non-zero anywhere in your window, your filter is not selecting the population you think. Key on `step_name` (which did not change) plus whatever else identifies the config, and report how much of the population is actually attributable rather than assuming all of it is
- **why it bites twice:** the same column is how you would check whether a seat is "noisy" before retiring it — and `bugs_open/138` is precisely about a seat looking noisy when it was being truncated. A measurement that silently loses the pre-cutover history is the wrong input to that decision
- **source:** `bugs_open/138` candidate 2, concept register `FIX-058`, 2026-07-30. `104_REPORT_seat_token_pressure_v1.sh` handles it by keying on (seat, cap) and printing `n_holder`
- **added:** 2026-07-30, bugfix_138 lane

### In `agent_definitions.default_config`, a step's prompt and its token cap sit at DIFFERENT depths — and both wrong paths return a confident uniform answer instead of erroring
- **footprint:** `agent_definitions.default_config`, `->'workflow'->'steps'-><step>->'config'`, `prompt_template`, `ai_service.max_tokens`, `099_SYNC_gate_roster.py`, `102_LINT_council_seat_parity.py`
- **fires when:** you read or patch a step's LLM config — auditing caps, surveying prompts, applying a prompt block, checking whether a fix reached a seat
- **the tell:** none, and that is the entry. The cap is **nested**: `config.ai_service.max_tokens`. The prompt is **not**: `config.prompt_template`, a SIBLING of `ai_service`. Ask for either at the other's depth and jsonb returns NULL for **every row** — so `config.max_tokens` reports `(unset→default)` across the whole roster, which reads as "nobody has ever right-sized these", and `config.ai_service.prompt_template` reports NULL for all 51 live `review_*` seats, which reads as "these seats have no prompts". Both are clean, uniform, plausible and completely false. `jsonb_set(..., create_if_missing := false)` fails the same way on a write: a wrong path is a **silent no-op**, not an error, so the affected row count is the only check
- **the check:** name both paths together whenever you write one of them — `SELECT s.key, s.value->'config'->'ai_service'->>'max_tokens' AS cap, length(s.value->'config'->>'prompt_template') AS prompt_chars FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s WHERE a.type='<agent>' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;` — if a whole column comes back uniformly NULL or uniformly unset, suspect the path before believing the finding. On a write, assert the row count
- **why it is a landmine and not a note:** the cap half was already written down (RUNBOOK, `WRONG_CALLS` 2026-07-29) and the prompt half was walked into **the next day by a thread that had read it** — because "watch the depth" does not tell you which keys are nested. Only the pair does
- **source:** `bugs_open/138` candidates 3 and 4, concept register `FIX-059`, 2026-07-29 and 2026-07-30
- **added:** 2026-07-30, bugfix_138 lane

### `deploy_image_asset` resolves its source image by PURPOSE, not by the `asset_id` you passed it — the second same-purpose asset on a site silently deploys as the first one
- **footprint:** `platform/orchestration/actions/deploy_image_asset_action.go` (`resolveStorageURIFromAsset`, Priority-1 query), `platform/orchestration/actions/v3_site_actions.go:2730` (`StoreAssetAction` → `updateContentDataField(..., purpose+"_uri", ...)`), `sites.content_data->>'{purpose}_uri'`
- **fires when:** you deploy an asset via `asset-deployer`/`deploy_image_asset` passing only `asset_id` + `purpose` (no `s3_uri`) — the natural, documented-looking way to trigger a deploy from a hand-triaged `site_work_items` row — on ANY site that has **2 or more** active assets sharing one `purpose` (any icon set, any multi-hero site). One asset per purpose is the case this path was built for and is silently fine; a second one is not
- **the tell:** none per-call. Every deploy reports `response.data.success: true`, writes to its own correctly-named destination path, and the work item completes. Only comparing the DOWNLOADED BYTES across the several "successful" deploys shows they are byte-identical — all of them fetched whichever same-purpose asset was generated **last** (`StoreAssetAction` last-write-wins into the single site-wide `content_data->>'{purpose}_uri'` cache, keyed by purpose alone). Confirmed live 2026-07-30: 6 distinct icon deploys (6 distinct `asset_id`, 6 distinct destination paths) all downloaded the same sha256, a 1408×768 hero photo belonging to none of the 6
- **the check:** before trusting a multi-asset-per-purpose deploy batch, `sha256sum` the downloaded files and confirm they differ, then look at (Read) at least one — do not stop at `success:true`. To deploy correctly today, supply `spec.s3_uri` explicitly as a genuine `s3://bucket/key` (derived from that asset's own `storage_path`, NOT its `url` column — `bugs_open/152` covers why: `url` may already be a stale presigned link or, post-deploy, an overwritten local path). A non-empty `s3_uri` is used before the buggy site-wide cache is ever consulted
- **source:** `bugs_open/155`, 2026-07-30, dartsonline_traffic lane
- **added:** 2026-07-30, dartsonline_traffic lane
  > **STILL LIVE as of 2026-08-06 — the fix is COMMITTED but NOT SHIPPED, so keep using the check above until the pod says otherwise.** `1d11827c1` deletes the Priority-1 purpose-cache read outright (all source resolution goes through the new `storage.AssetSourceRef`, which reads the asset row's own `storage_path`/`url`), and `StoreAssetAction` no longer writes the `{purpose}_uri` key at all. Pod-verified 2026-08-06 on `v1.0.1257`, both replicas: `AssetSourceRef` → **0**, `"Resolved s3_uri from site content_data via asset_id"` → **1** — i.e. the fix is absent and the buggy branch is present, in the same command. **That string pair is the retirement test for this entry**: post-roll it must read `≥1` / `0` on every replica, and then the `s3_uri`-workaround half of "the check" above becomes unnecessary (harmless, but no longer needed — an explicit `s3_uri` still wins). Do not retire this entry on the commit, on the tag, or on a roll — only on that pair. Migration `323` (applied 2026-08-06) backfilled `storage_path` on 205 presigned-only rows, so the row-side answer already exists for every generated asset.
  > **RETIRED 2026-08-06 — the retirement test above PASSED, so this trap no longer fires.** Chassis `v1.0.1259`, both replicas: `AssetSourceRef` → **2**, `"Resolved s3_uri from site content_data via asset_id"` → **0** (plus a nonsense control at 0, so the grep discriminates). The purpose-keyed branch is gone from the binary and `StoreAssetAction` no longer writes the key. **What replaces this entry as the thing to know:** `deploy_image_asset` now resolves an asset's source from that asset's OWN row via `storage.AssetSourceRef(storage_path, url)`, and a row that names no stored object is a LOUD SKIP (`deployed:false, skipped:true, reason:…`) rather than a silent fetch of a neighbour's file — so the new failure mode is a visible refusal, not wrong bytes. The `s3_uri` workaround in "the check" above still works and is no longer necessary. **The one thing this proof does NOT cover:** no real multi-same-purpose deploy has been run since the fix, so the sha256-differ check remains the outstanding behavioural test (`bugs_open/155`, still OPEN for that reason).

### A page whose `url` is `/` deploys a file with NO NAME — `getPageInfo` derives the filename by trimming the leading slash
- **footprint:** `pages.url`, `platform/orchestration/actions/rerender_single_page_action.go` (`getPageInfo`, `PageInfo.Filename`), `platform/orchestration/actions/adopt_verbatim.go` (`urlToDeployPath`), any hand-written `INSERT INTO pages` or `UPDATE pages SET url`
- **fires when:** you set a homepage's `url` by hand, or write any code that derives `pages.url` from a crawled/collected URL. A crawler reports the site root as `"/"` or `"https://host/"`, and `"/"` is the obvious, correct-looking value to store for a homepage
- **the tell:** none at write time — the row is valid, the constraint passes, the page looks right in every `SELECT`. The filename is computed downstream as `strings.TrimPrefix(p.URL, "/")`, so `"/"` yields `""`, and the deploy commits a file with an empty name into the site's directory. `getPageInfo` has a guard (`p.URL == "/" || p.Name == "index"` → `index.html`) which is why this has not bitten in the normal path — but the guard keys on the page being NAMED `index`, so an adopted or hand-made homepage under any other name (`home`, `landing`, `main`) falls straight through it
- **the check:** never store a bare directory URL. Normalise `"/"` and any trailing-slash or extensionless path to its `index.html` **before** it reaches `pages.url`, and assert the invariant rather than the string: `strings.TrimPrefix(url, "/") != ""`. In SQL: `SELECT name, url FROM pages WHERE url = '/' OR url = '' OR url LIKE '%/';` should return no rows
- **why it is a landmine and not a bug report:** the value is not wrong, it is wrong *for one downstream consumer three files away*, and the failure is a silently misnamed artefact rather than an error. `urlToDeployPath` now does this normalisation for verbatim adoption and its test asserts the empty-filename invariant directly — but anything writing `pages.url` by another route still needs the check
- **source:** built during `fidelity=locked` verbatim adoption, concept register `ADO-037`, 2026-07-30. Mutation-verified: forcing the root case to return `"/"` fails `TestURLToDeployPath` on the invariant, not just the expected value
- **added:** 2026-07-30, loancalculator_couk lane
- **UPDATED 2026-07-31 (bugfix_125 lane) — the CONSUMER side is now structurally closed, and one sentence above is imprecise.** `getPageInfo` no longer derives the filename itself: it calls `datahelpers.PageDeployFilename(p.URL, p.Name)`, where `"/"` and `""` both resolve to `index.html` and **no input can produce an empty filename** (pinned by `TestPageDeployFilenameNeverEmpty`, plus the four named cases from this entry — `home`/`landing`/`main` at the root). So the downstream half of this trap is gone, and the entry is kept because the **write** side still matters: a bare `/` in `pages.url` is still wrong for the nav, the sitemap and every link resolver, which is a different consumer than the one this entry was written about.
  - **The imprecise sentence, corrected:** *"the guard keys on the page being NAMED `index`, so a homepage under any other name falls straight through it"* — it did not. The pre-change guard was `p.URL == "/" || p.URL == "" || p.Name == "index"`, and the **first clause already caught the root case whatever the page was called**; the name clause was a second, redundant catch. So the NO-NAME outcome was not reachable through `getPageInfo` by the route described. The entry's *conclusion* (normalise before it reaches `pages.url`) was right and its footprint was right; the mechanism was not. Recorded rather than deleted because five council seats spent a round on this entry (corr `758f6e62`) and the round was worth it — it is what forced the check that produced this correction.
  - **`urlToDeployPath` is NOT a copy of the deploy-path classifier and must not be converted to one.** It is the **inverse** function: it produces the value STORED in `pages.url` and deliberately keeps the leading slash (`/index.html`), whereas `PageFilePathFromURL` produces a repo-relative path and deliberately strips it. They are the two halves of one invariant, with opposite output contracts. Anyone tidying these into "one URL helper" will silently break one side.

### The adoption crawl index stores ONE page under SEVERAL keys — dedupe it by content pointer, never by URL string
- **footprint:** `platform/orchestration/actions/apply_adoption_plan_action.go` (`buildCrawlPageIndex`, `matchCrawlContent`), `platform/orchestration/actions/adopt_verbatim.go` (`crawlPathIndex`), `research_results` rows with `result_type='adoption_crawl_page'`
- **fires when:** you iterate the crawl index to enumerate a crawled site's pages — the natural move when you want "the list of pages we found", rather than looking a single page up by URL
- **the tell:** a plausible-looking page count that is a small multiple of the truth. `buildCrawlPageIndex` deliberately registers the SAME `*crawlPageContent` under every alias it can (absolute URL, path-only form, and `metadata.sourceURL` when it differs) so `matchCrawlContent` finds the page however the LLM plan spells it. Iterating those keys therefore yields 2–3 entries per real page, each with identical content — and since every entry is genuinely valid, nothing errors. A 27-page site reads as 60–80 pages, and if you create a row per entry you get duplicate pages that differ only in URL form
- **the check:** dedupe by the map VALUE's pointer identity, not the key — `seen := map[*crawlPageContent]string{}`. Then pick the surviving alias **deterministically** (sort the keys first): Go map iteration order is randomised, so without a sort a re-run legitimately picks a different alias and writes a different `pages.url` for the same page. Sanity-check the reduced count against the site's own `sitemap.xml` before creating anything
- **source:** built during `fidelity=locked` verbatim adoption, concept register `ADO-037`, 2026-07-30
- **added:** 2026-07-30, loancalculator_couk lane

### ~~`snapshot_agent` has two overloads…~~ **DUPLICATE — cut back 2026-07-31 by its own author; read the three originals instead**
> **This entry should not have been written.** Everything it claimed was ALREADY in
> this file, in three entries I did not grep for before appending: the overload split
> (§`snapshot_agent has TWO overloads writing to TWO different tables`), the wrong
> table (§`snapshot_agent() writes to agent_definitions_backup, NOT an is_snapshot
> row…`), and the ordering tie (§`agent_definitions_backup keeps the SOURCE row's id
> and created_at — order by snapshot_taken_at`, which already states that *every*
> backup row for one agent shares both). My "measurement" of three `feature-designer`
> backups sharing `created_at 2026-07-17 18:06:05` is a confirming instance of a
> documented fact, not a finding. **Kept as a stub rather than deleted because this
> file is append-only, and because a duplicate that is visibly marked is more useful
> than one silently removed.** Logged in `WRONG_CALLS.md`.
>
> The irony worth keeping: the council seat that caught my wrong rollback block cited
> one of these existing entries *by name*. The information was already delivered to the
> agents; I just had not read it. CLAUDE.md says the SessionStart hook only matches
> **path**-shaped footprints and that you must "still grep it yourself for table,
> command and symbol footprints" — `snapshot_agent` is a symbol, so it was mine to grep.
- **the two things here that were NOT already recorded**, and the only reason to read on:
  - **`222_feature_designer_one_edit_per_file_per_stage.sql` still carries the wrong rollback instruction** ("the newest `is_snapshot` row"). Not edited — another lane's applied migration. `272` was corrected 2026-07-31.
  - **A dry run leaves NO snapshot.** `snapshot_agent` inside a transaction you `ROLLBACK` is rolled back with everything else, so a clean dry run is not a backup. Easy to assume otherwise when the dry run prints `Snapshot captured: …`.
- **footprint:** `snapshot_agent`, `agent_definitions_backup`, `222_feature_designer_one_edit_per_file_per_stage.sql`
- **added:** 2026-07-30, bugfix_099 lane · **cut back to a stub 2026-07-31**

### An SVG viewBox CLIPS, so a label that does not fit reads as CORRUPTED CONTENT, not as a layout bug

- **footprint:** `platform/orchestration/actions/report_charts.go`, `renderBarChartSVG`, any inline-SVG chart whose plot geometry is a fixed constant
- **fires when:** you add or lengthen any text in an inline SVG chart — a value label, an axis caption — while the plot width reserves a **fixed** gutter for it. Also fires with no edit at all, the first time real data produces a longer label than the sample did
- **the tell:** the page renders, nothing errors, nothing logs. The text is simply cut off mid-word (`6.42× (Insufficien`) or two captions overprint into mush (`reqmiaegimealttufh1r.0ex/f)old (1.25×)`). **In a report of computed figures this reads as corrupted DATA and gets diagnosed as a scoring bug**, which is the expensive part. Measured 2026-07-30: the gutter was 90 units, the real label needs ~150, so EVERY capped bar was clipped — on pages that had passed every check the pipeline has
- **the check:** assert geometry against **content**, not against a golden file — `x + estTextWidth(text) <= viewBoxWidth`, and no two captions sharing a baseline whose spans overlap. Then **render it and LOOK**: `chromium --headless=new --disable-gpu --no-sandbox --screenshot=out.png "file://.../page.html"`. ⚠ **chromium here is a SNAP**: it cannot write into `/tmp/claude-*` (fails `No such file or directory`) or into any **dot-directory** under `$HOME` (fails `Permission denied`) — use a plain `~/dir`, or the screenshot silently never appears and you conclude the tool is unavailable
- **source:** owner screenshots of the 2026-07-26 gripper fixture pages, 2026-07-30; fix `f8e7c31ce` with geometric tests
- **added:** 2026-07-30, robot_hands_gripper_dossier lane

### A passing mutation test may mean a SECOND guard absorbed the mutation, not that the guard is redundant

- **footprint:** mutation testing where two mechanisms protect the same property (a computed size **and** a truncation fallback; a validator **and** a default)
- **fires when:** you revert fix A to prove test A fails — and fallback B quietly compensates, so the property still holds by a different route and the test stays green
- **the tell:** a clean `PASS` that you are primed to read as *"this guard is redundant"*. Worked case 2026-07-30: reverting a chart's computed label gutter left the "nothing is clipped" assertion green, because a `fitText` fallback silently **truncated** the label instead. Nothing was clipped. Content was lost. The assertion was true and useless
- **the check:** assert the **outcome the caller asked for**, not the absence of the symptom — here, that the emitted label *equals* the requested label, rather than merely that it fits. If a mutation passes, look for the compensating path before concluding redundancy, and mutate **that** too
- **source:** `robot_hands_gripper_dossier/NOTES_…` 2026-07-30. Companion to *"a mutation that never happened is indistinguishable from a guard that works"* above — that one is a mutation that did not apply, this one is a mutation that applied and was absorbed
- **added:** 2026-07-30, robot_hands_gripper_dossier lane

### `platform/httpguard` is INBOUND-abuse only — it has no SSRF/outbound-fetch guard

> **CLOSED 2026-07-31 for the one call site that mattered most.**
> `platform/fetchguard` (register `DBI-025`) is the outbound sibling this entry
> called for; wired into `internal/adapters/webscrape/adapter.go`'s
> `downloadImage`, which was a real, live SSRF hole (`bugs_open/159`) — image
> URLs taken from scraped page content, fetched with no checks at all. **Still
> open:** `browser-runner-adapter`'s Playwright `page.Goto` navigation is a
> different fetch surface a Go `http.Transport` guard cannot see — not fixed,
> flagged. Any *new* code fetching a URL the platform did not choose should
> use `fetchguard.NewClient`, not reach for `httpguard` or a bare
> `&http.Client{}`.

- **footprint:** `platform/httpguard`, `internal/adapters/webscrape`
- **fires when:** building any feature that fetches a user-supplied URL (domain
  intake, scrape-on-demand, "check this site") and reaching for `httpguard`
  because the name sounds general-purpose
- **the tell:** nothing stops you reading the package and walking away satisfied
  — it genuinely has rate limiting, a trustworthy client-IP resolver, and a
  form honeypot/timing gate, all real and well-built. The gap is a *missing*
  file, which greps for `IsPrivate`/`IsLoopback`/`169.254`/SSRF turn up empty
  everywhere in the repo, not a bug in an existing one
- **the check:** read the package doc comment at the top of `httpguard`'s
  `clientip.go` — it says outright *"the platform's ONE set of INBOUND-abuse
  primitives"*. Inbound (who is hitting us) and outbound (what we fetch on a
  visitor's behalf) are different trust directions and this package is only
  the first. `internal/adapters/webscrape/adapter.go` — the live scrape path —
  has no SSRF check either
- **source:** `webdesign_uk_build_service/PLAN` §8/§9, checked 2026-07-29
- **added:** 2026-07-29, webdesign.uk lane

### `doc_plans`/`doc_notes` `subject_type` has TWO enforcement points, and `\d <table>` only shows you one
- **footprint:** `doc_plans`, `doc_notes`, `subject_type`, `platform/orchestration/actions/doc_subjects_common.go`, `validDocSubjectTypes`, `sql_for_agents/*_doc_subjects_*.sql`
- **fires when:** you add a new kind of thing that should carry travelling docs (a component, a site, a seat, anything). You read the table, find a CHECK constraint listing the allowed values, and correctly conclude a migration widens it. **That is half the change.**
- **the tell:** there is none at write time — the migration applies cleanly and the constraint reads correctly afterwards. The failure surfaces later and looks like something else entirely: every doc action (`write_doc_plan`, `append_doc_note`, `load_doc_context`, `persist_diagnosis_note`) refuses the new subject, because `validDocSubjectTypes` in Go is a **second, independent** allow-list. The DB accepts what the whole application refuses. **This has already happened twice: migration 163 (+`experience`) missed the `persist_diagnosis_note` gate; migration 184 (+`action`) moved the DB CHECKs only and left its own seeded action docs unreachable through every doc action — that is `bugs_open/064`.** A third session (mine, 2026-07-30) planned the same mistake and was saved only by a code comment.
- **the check:** **grep the VALUE you are adding, not the table you are changing.** `git grep -n "experience-pattern"` returns the Go list, the migration that last set the CHECK, and the four-enforcement-point checklist (`experience_register/design/subject_type_addition.md`) in one command. `\d <table>` tells you what the DATABASE enforces and is structurally silent about every gate in front of it. Then: **both halves in ONE commit** — `TestValidDocSubjectTypes_LockstepWithMigrationCheck` parses the newest *numbered* migration recreating `doc_plans_subject_type_check` and fails on drift, so landing either alone reddens HEAD for every session on this shared tree. Order of application is **image first, then migration**.
- **second trap in the same footprint, which the obvious fix walks straight into:** the two tables do **not** carry the same vocabulary. `doc_notes` also allows `'landmine'` (migration 270); `doc_plans` does not. Re-adding `doc_notes`' constraint from `doc_plans`' array — the natural way to make them agree — **drops `'landmine'` and orphans the live landmine corpus** (57 rows measured 2026-07-30, owned by other threads, synced from this very file). Read the constraint you are replacing; migration 273 refuses to run if `'landmine'` is missing, for this reason.
- **source:** planned wrong then corrected 2026-07-30 while adding `subject_type='component'` (`c659e312b`, DOC-068, council `e5673868`); the two prior instances are `bugs_open/064` and the history comment in `doc_subjects_common.go`
- **added:** 2026-07-30, staged_component_build lane

### The council-gate runbook's own verdict query returns the most recent note FLEET-WIDE, not yours
- **footprint:** `doc_notes` where `categories ? 'council-gate'`, `097_TRIGGER_council_review_v1.sh`, `RUNBOOK_council_gate.md`
- **fires when:** you read your council verdict with the query the runbook and the trigger script both print: `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`
- **the tell:** **there isn't one.** You get a complete, well-formed REVISE note with real objections and real reviewer checks — belonging to whichever session submitted most recently. Several councils run per hour on this fleet. On 2026-07-30 this returned a verdict about `internal/tools-api/clientip` and httpguard to a session whose submission was about the claims floor; the only reason it was caught is that the plan summary described someone else's change. **Revising against another thread's objections is the failure this produces**, and it would look like diligence
- **the check:** always key on YOUR correlation — `WHERE body LIKE '%<your-corr-prefix>%'`, or read `diagnosis_artifacts WHERE correlation_id=... AND kind='council_report'` and parse `body` as JSON (the `reviews` array with per-seat verdicts and objections is there; `metadata->'reviewers'` is only a COUNT, and `metadata->>'decided_by'` can be null on the artifact while the note names the gating seat)
- **source:** hit 2026-07-30 reading round 1 of council `2d0dbc2e`; the trigger script prints the correlation-keyed query for the artifact but the *human-readable note* query it prints alongside is the unkeyed one
- **added:** 2026-07-30, oufe lane

### Deduplicating `page_components` with a unique index on `(page_id, slot_name)` breaks 11 legitimate pages
- **footprint:** `page_components`, `save_page_sections_action.go`, any migration adding a uniqueness constraint on a page's slots
- **fires when:** you find a page holding the same `slot_name` twice and reach for the obvious structural fix — "a page can only have one of each slot, so make that an invariant". It reads like exactly the kind of make-the-bad-state-unrepresentable change this platform prefers
- **the tell:** **there is none at write time.** The index builds, or it fails with a bare conflict naming one arbitrary page and reads like a data-cleanup problem to clear first. The premise is what is wrong: a repeated slot name is **normal**. Measured fleet-wide 2026-07-30 — 17 duplicate `(page_id, slot_name)` groups, and **11 are legitimate**: `generic-text-block` used 2–3× on one page across ai-agent-orchestration, leopardess, gaswholesalers, finetuning and idea.uk, plus `info-card-grid` ×2 on webdesign.co.uk, all with **differing** content. Only 6 were true duplication (`vonc.com/about`, `bugs_open/156`)
- **the check:** the discriminator is **content identity, not slot repetition** — always add `count(DISTINCT md5(content_data::text))` to the census and act only on the groups where it is 1:
  `WITH dups AS (SELECT page_id, slot_name FROM page_components GROUP BY 1,2 HAVING count(*)>1) SELECT s.domain, p.name, pc.slot_name, count(*), count(DISTINCT md5(pc.content_data::text)) FROM page_components pc JOIN dups d USING (page_id, slot_name) JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id GROUP BY 1,2,3;`
  Two traps inside that query: a group can report `distinct_content = 0` — that is **NULL** `content_data` on every row, not agreement (`finetuning.uk/our-position-on-ai`), and a census filtering `content_data IS NOT NULL` cannot see rows of that shape at all. `content_hash` looks like the column for this and is **empty**, so it cannot stand in for the md5
- **the wider point:** before hand-deleting duplicate rows, check what **regenerates** them — `site_plans` → `site_plan_sections` → `pages.sections` → `page_components` (see the dartsonline entry above). On vonc/about all three upper sources correctly said 6, which is what made the cleanup safe; that is a fact to establish, not assume
- **source:** `bugs_open/156`, 2026-07-30, while answering HANDOFF D's first-move question in the gauntlet_dead_cta lane
- **added:** 2026-07-30, gauntlet_dead_cta lane

### `has_visible_area` reports 0 for any axis whose rendered size is a WHOLE NUMBER, and blames a collapsed flex child
- **footprint:** `internal/adapters/browserrunner/run_checks_action.go` (`VisibleArea`, lines ~703-721), the `has_visible_area` criteria check type, any ```criteria fence asserting a control is big enough to click
- **fires when:** you author a `has_visible_area` check against a control you sized in whole pixels — an icon button, a checkbox, an avatar, a `width:24px` touch target. Which is to say: exactly the elements the check was added to police
- **the tell:** the check fails with `renders 0x0 … present in the DOM but too small to see or click. A collapsed flex/grid child is the usual cause: check that its parent establishes a height` — **and the named cause is not there.** Two things in the same run contradict it: the `interaction` checks that CLICK that element PASS (Playwright's `Click()` enforces visibility, so it cannot have been 0x0), and a screenshot shows the control rendering normally. The discriminating observation is an element with ONE integral axis and one fractional axis: `#vtc-verdict` measured **0x94** on mobile, i.e. only the integral axis read 0. No collapse theory predicts that
- **the cause:** `m["w"].(float64)` — but `mxschmitt/playwright-go@v0.6100.0/js_handle.go:109-114` returns **`int` for an integral number** and `float64` only for a fractional one (`if math.Ceil(v)-v == 0 { return int(v) }`). The comma-ok assertion discards the value and leaves **0**, which is then compared against the threshold. `go build` cannot see it and no test with a fake page can either, because the fake returns float64
- **the check:** never accept a `0x0` from this check type on its own. **Screenshot the live URL** (`chromium --headless=new --window-size=1366,1500 --virtual-time-budget=8000 --screenshot=$HOME/x.png <url>` — snap chromium cannot write outside `$HOME`), and look for an `interaction` check in the same run that successfully clicked the same selector. `min_width`/`min_height` cannot be tuned around it: the measured value is 0, so every positive threshold fails
- **do NOT "fix" the page** by making the element's size fractional. That silences the gate and leaves the platform bug in place for everyone else
- **source:** `bugs_open/157`, found 2026-07-30 building `tool-ai-vendor-trust-checklist` as the first end-to-end run of the S0-S7 ladder. Blast radius is exactly two lines — `grep -n '\.(float64)' internal/adapters/browserrunner/*.go` returns only those. Concept register `TL-034` says `has_visible_area` is **built**, which is true; this is why built and trustworthy are different rows
- **added:** 2026-07-30, leopardess vendor-trust-checklist build (with staged_component_build)
- > **CORRECTED 2026-07-31 (157 fix session) — FIXED AT HEAD (`71680ad513`), STILL LIVE IN THE POD.** `VisibleArea` now coerces via `evalNumber` (`float64 | int | int64 | int32 | json.Number`) and **errors** on an undecodable value instead of reporting 0, so the trap above is closed in git. **It is NOT closed in production:** a Go change is inert until a roll, and on 2026-07-31 15:26 UTC pod `browser-runner-adapter-78f467dbb7-52bx2` (`v1.0.1215`) returned the new marker `non-numeric w/h in result` → **0** with three positive controls → **1**. **So keep reading this entry as live until you have pod-grepped otherwise** — and note the check the sibling entry below gives you (`grep -n 'm["w"]'` over the repo) now answers "is it fixed in git", which is a DIFFERENT question from "is it fixed in the binary I am about to author a fence against"
- > **RESOLVED 2026-07-31 17:47 UTC — THIS TRAP IS RETIRED. The fix is LIVE and proven at the artefact.** Pod `browser-runner-adapter-fbb78fbbb-rpjl8` (`v1.0.1216`) returns the marker `non-numeric w/h in result` → **1** (it was **0** on `v1.0.1215` two hours earlier) alongside five positive controls → **1** each in the same exec. `has_visible_area` now measures whole-number axes correctly. **Two corrections to the line above, which was written before the roll:** (1) it says the accepted set is `float64 | int | int64 | int32 | json.Number` — that was the first draft; round 2 (`f15e00a47`) **narrowed it to exactly `float64 | int`**, which is all `parseValue` emits, after the council's edit-quality seat correctly called the wider set dead code. (2) "STILL LIVE IN THE POD" was true when written and is now false. **Keep both lines:** the pair is the record of a fix crossing from git into production, which is the gap this entry existed to make visible. Closing evidence in `bugs_closed/157`

### idea.uk's static `privacy/terms/refund-policy` pages 301 to a DIFFERENT application, so tagging the static site misses the conversion pages
- **footprint:** `idea.uk`, `sites` row `1244516d-014d-421c-88c6-090bb1e9552a`, `docs024_key_docs_latest/idea.uk/golang_files/service.go` (`App.page`), `/etc/nginx` on 116.203.204.115, any fleet-wide sweep that asserts something about "every page" of a domain
- **fires when:** you apply anything site-wide to idea.uk from the chassis side — an analytics tag, a meta tag, a consent banner, a footer link, a JSON-LD block — and verify it by re-rendering the `pages` rows and curling them
- **the tell:** **almost none, and the one you get is easy to explain away.** 19 of 20 URLs pass; `/privacy.html` returns **301** and scores 0 hits, which reads as "redirect, fine, skip it". It is not fine: nginx proxies a **16-route reserved set** to a Go binary on the VM, and `/privacy`, `/terms`, `/refund-policy` are served by **that binary**, with the static `.html` copies 301'ing to them (owner decision 2026-07-18, deliberate — the legal text is tied to the £29 purchase terms). Eleven of those routes render HTML through one wrapper, `App.page()`, and none exist in the static build. **The two that matter are `"Request received"` and `"Payment received"` (`/order/success`) — the only pages that can evidence a sale.** A tag applied to the static site alone therefore reports traffic and **zero conversions**, which reads as "the site doesn't convert" rather than "the tag isn't there"
- **the check:** record the **status code alongside the assertion, per URL**, and look at any row that differs — `curl -s -o f -w '%{http_code}'` WITHOUT `-L` (with `-L` you score the redirect *target's* content as if it were your page, which hides the same gap the other way). Then enumerate the other application: `grep -n 'HandleFunc' docs024_key_docs_latest/idea.uk/golang_files/service.go` and `grep -rn 'a\.page(' ` on the same dir. `curl -s https://idea.uk/order/success | grep -c <your marker>` is the honest verification
- **source:** found 2026-07-30 applying GTM `GTM-PQ3WCTBD`; `analytics_gtm/NOTES_analytics_gtm.md`, `idea_uk_vm_site/RUNBOOK` §3b. The reserved-path set is documented there but nothing connects it to "site-wide change" until you have already shipped one
- **added:** 2026-07-30, analytics_gtm lane

### Editing a site's `head`/`header` chrome edits every site that shares the component — 3 head components cover all 14 live domains
- **footprint:** `site_components`, `content_components`, `slot_name IN ('head','header')`, `render_site_components_action.go`, `Document Head` (`116c5f91-bc0d-439d-9e13-a3ba2d145571`)
- **fires when:** you change one site's `<head>` or header — a tracking tag, a font, a meta tag, a favicon — by editing the `content_components.html_template` its `site_components` row points at
- **the tell:** **none at write time.** The UPDATE reports `UPDATE 1` and the site you were asked about renders correctly. Measured 2026-07-30 across the 14 `status='deployed'` sites: the `head` slot resolves to just **3** components (`Document Head` **×9**, `head-seo-standard` ×4, one fork) and `header` to **6**. So a per-site value written into `Document Head` silently ships to **eight other domains** — and for something like an analytics container that is not a cosmetic bug, it is other people's traffic arriving in the wrong account
- **the check:** before editing any chrome template, `SELECT count(*) FROM site_components WHERE component_id='<id>'`. If `>1`, do NOT hardcode: put the value in `site_specs` (aspect `site_config`) and **gate** the template with `{{if .field}}…{{end}}`, so unset sites render byte-identically. The resolver already supports it — `input_schema` field with `source: config.<dotted.path>` → `resolveConfigPath` (`plan_sections_action.go:516-527`)
- **second trap inside the same footprint:** the `input_schema` entry must be **map-valued**. `render_site_components_action.go:612-615` skips any entry that is not a map as "not a field descriptor" — which is why `Document Head`'s existing flat `{"title": "...", "description": "..."}` scalars have never resolved and never could. A scalar you add will be silently ignored and you will debug the resolver instead
- **third trap:** writing only the template changes **nothing on any page** — chrome is a stored artefact (`bugs_open/117`), and pages assemble from `site_components.rendered_html`. Writing only the artefact gets it reverted by the next chrome rebuild. **Both, or it is inert or temporary**
- **source:** 2026-07-30, applying GTM to idea.uk; `analytics_gtm/RUNBOOK_analytics_gtm.md` §0/§2
- **added:** 2026-07-30, analytics_gtm lane

### `hero-tool` / `tool-cta` render NO buttons unless you set the `*_url` fields — the `*_label` fields alone are dead data
- **footprint:** `content_components` functions `hero-tool` and `tool-cta`, any `page_components.content_data` for a `page_type='tool'` page, the keys `cta_primary_label` / `cta_secondary_label` / `primary_cta_label` / `secondary_cta_label`
- **fires when:** you build or copy a tool page and fill in the CTA labels, reasonably assuming a field called `cta_primary_label` puts a button on the page. It does not. Both templates gate the entire button row on the URL, not the label: `hero-tool` on `{{if or .cta_primary_url .cta_secondary_url}}` and `tool-cta` on `{{if .primary_cta_url}}` / `{{if .secondary_cta_url}}`. Note the two components spell the same idea in **opposite orders** (`cta_primary_url` vs `primary_cta_url`), so copying a key from one to the other also silently produces nothing
- **the tell:** **none whatsoever.** No error, no empty box, no gap — the row is omitted entirely, so the page looks like a deliberate design with no call to action. `fundamentallyai.com/tools/llm-cost-calculator.html` has shipped this way: its stored `content_data` carries `"cta_primary_label": "Run the calculator"` and no URL, and the live hero has **zero** buttons
- **the check:** grep the SERVED page for the rendered anchor, not the class name — `grep -c 'htl-btn-primary"' page.html`. **A grep for the class alone always returns at least 1 because the class is defined in the component's own inline `<style>` block**, which is how I first talked myself out of this entry: `grep -c htl-cta-row` returned 1 on a page with no CTA at all. Match the markup (`class="htl-btn-primary"`) or extract the `<div class="htl-cta-row">` element and confirm it exists as an element rather than a selector
- **source:** 2026-07-30, building `tool-review-council-simulator` on fundamentallyai.com; verified by extracting the element from both the sibling page (0 anchors) and the new page (2 anchors) after supplying both labels and URLs. The sibling's dead labels are NOT fixed by that build
- **added:** 2026-07-30, brochure_component_library lane

### A browser probe injected before `</body>` runs BEFORE the component's `DOMContentLoaded` init, and reports the exact bug it exists to catch
- **footprint:** `docs024_key_docs_latest/brochure_component_library/scripts/probe_*.py`, any headless-Chromium check built by appending a `<script>` to fetched HTML and reading it back with `--dump-dom`
- **fires when:** you write an interaction probe the proven way (fetch the page, splice a driver script in before `</body>`, run `chromium --headless=new --dump-dom`, parse a `<pre>` of results) against any component whose own init is gated on `document.readyState` / `DOMContentLoaded` — which is the convention 6 of 7 active `js_snippets` follow, and the fix `bugs_open/138`-era work applied to `teaser-reveal-panel`
- **the tell:** **it looks precisely like a real, serious bug, and it is the specific bug this class of probe was invented to find.** Your injected script is also inline, so it executes during parsing, strictly before `DOMContentLoaded` fires. You therefore measure the pre-init page: placeholder text still in the headline, a data-driven roster with zero rows, every "does the control change the output" assertion failing with `null -> null`. Read at face value it says "the JS never ran", i.e. the teaser-reveal-panel silent no-op. A thread that "fixes" the component in response breaks a working one
- **the check:** make the driver wait, then confirm it can still fail. Wrap the whole driver body in a deferral — `if (document.readyState === 'complete') run(); else window.addEventListener('load', run);` — and then **run it against a deliberately broken copy**. If a mutant with `init()` deleted, and a mutant with the `<script>` moved ahead of its markup, both still fail, the probe distinguishes the two states. If they pass, the probe is measuring nothing
- **source:** 2026-07-30, first run of `probe_council_simulator.py` reported 7 failures against a component that was correct. Same lane's `probe_reveal_open_state.py` never hit this because it forces DOM state directly instead of waiting for init
- **added:** 2026-07-30, brochure_component_library lane

### Verifying a page straight after firing its rerender shows a 404 or stale copy, and both look like you broke the site
- **footprint:** `rerender_page_vonc.sh`, `rerender_arena_vonc.sh`, `rerender_single_page`, any `curl` verification following a `page_rerender`/`deploy_page`
- **fires when:** you fire N rerenders and verify immediately — the natural thing to do, and the more pages you fire the worse it reads. Deploys land seconds apart, and the serving layer is briefly mid-write per page
- **the tell:** a **404 with a ~290-byte body** on some pages and **old content at HTTP 200** on others, in the same sweep, while the DB says every page is `status='active'` with a fresh `deployed_at`. Measured 2026-07-30 on vonc: 8 archetype pages rerendered, verified at once — 2 served new copy, **2 returned 404**, 4 served stale. Ten minutes later all 8 were correct with no further action. The 404 is the file being replaced, not deleted; the stale 200 is propagation. **Both are transient and neither is a defect** — but a 404 on a page you just touched reads as destruction, and it is very tempting to "fix" it by re-firing or rolling back, which is how you turn a non-event into an incident
- **the check:** read `pages.deployed_at` **before** believing a fetch — if it is within a minute or two of now, you are inside the deploy window and your fetch proves nothing. Then poll until the condition holds rather than sampling once: `until [ "$(curl -s "<url>?cb=$(date +%s%N)" | grep -c "<old string>")" = 0 ]; do sleep 20; done`. Assert on a string the change **removed** as well as one it **added**, plus a control you never touched — "new string absent" and "page broken" are indistinguishable without the control. Cache-bust every request; a bare URL can serve you the pre-deploy copy long after propagation finishes
- **source:** hit 2026-07-30 rewriting vonc's 8 archetype hero/CTA sections (`bugs_open/156` lane work); the first verification sweep looked like two destroyed pages and was simply early
- **added:** 2026-07-30, gauntlet_dead_cta lane

### Firecrawl's `rawHtml` is the POST-JAVASCRIPT DOM, not the bytes the origin served — so it cannot be used to preserve a page verbatim
- **footprint:** `firecrawl_crawl` / `firecrawl_scrape` steps, `scrape_config.formats` containing `rawHtml`, `buildCrawlPageIndex`, `crawlPageContent.RawHTML`, `platform/orchestration/actions/adopt_verbatim.go`, `research_results` rows with `result_type='adoption_crawl_page'`
- **fires when:** you treat crawled `rawHtml` as "the page as served" — verbatim adoption, a byte-diff against a live site, archiving a page, or any before/after comparison where the crawl is one side
- **the tell:** the name says raw and the content is valid HTML for the right URL, so nothing looks wrong. Measured on loancalculator.co.uk, 2026-07-30: **27 of 27 pages differed from the served file, every one LARGER by ~8,900–9,060 bytes** — a suspiciously CONSTANT delta, which is the signature. The delta was that site's `nav.js` having already executed: ~9KB of generated nav markup baked into `<div id="nav-placeholder">`, its `<style>` block hoisted into `<head>`. Also, silently: every relative URL rewritten absolute (`/assets/x.css` → `https://host/assets/x.css`), `href="#"` rewritten to a self-link (`https://host/page.html#`, which RELOADS the page on click instead of doing nothing), `<meta charset>` replaced by `<meta http-equiv="Content-Type">`, whitespace collapsed, `&` normalised to `&amp;`
- **the check:** never assume; measure one page before trusting the set. `md5` the stored `rendered_html` against the byte source you believe it equals (the deploy-repo file, or `curl` with a cache-buster). **A constant-ish size delta across every page means a script ran, not that the crawl is noisy.** If you need served bytes, fetch them without a browser (plain HTTP GET) or read the deploy repo — `formats: ["rawHtml"]` will not give you them
- **why it bites hardest where it matters most:** the whole point of a verbatim/preservation path is that the bytes do not change, and this is the one input that silently changes them. Three pages of a working live site were redeployed from this DOM before the checksum gate caught it (~2.2KB → ~11.2KB each); they were restored from the repo
- **source:** `docs024_key_docs_latest/loancalculator_couk/` (G10), concept register `ADO-037`, 2026-07-30, loancalculator_couk lane
- **added:** 2026-07-30, loancalculator_couk lane

### `b2 sync --skip-newer` silently skips a file whose bucket copy is NEWER — so a REVERT can fail to propagate, and the deploy still reports success
- **footprint:** `.github/workflows/deploy-to-b2.yml` in the `gqls/sites` repo (`b2 sync --delete --skip-newer`), any revert/restore of a file under `<domain>/` in that repo
- **fires when:** you restore a file that something else (an automated `Rerender:` commit, another lane) wrote to the bucket moments earlier — i.e. exactly the revert case, because the thing you are undoing is by definition the most recent write
- **the tell:** the workflow run is GREEN, the "Sync to B2" step prints no error, and the reverted file is simply absent from the `upload` lines. Measured 2026-07-30: a 3-file restore printed **2** `upload` lines; the third file stayed mutated at the origin (`last-modified` and `x-amz-version-id` still showing the earlier bad write) while the repo held the correct bytes. `curl` with a cache-buster confirmed the ORIGIN, not the CDN, was stale — so waiting for the cache to expire would never have fixed it
- **the check:** after any revert, verify at the ORIGIN with a cache-buster (`curl -s "https://host/path?cb=$RANDOM" | wc -c`) and compare to the repo file — do not trust a green run, and do not conclude "stale cache". Count the `upload` lines against the number of files you changed: `gh run view <id> --log | grep -c 'upload '`
- **the fix:** `gh run rerun <run-id>`. A fresh checkout stamps every file with a current mtime, which beats the bucket copy, and the sync then uploads. (Re-committing does not work when the content is already correct in git — there is nothing to commit, so no run fires.)
- **source:** hit directly while restoring 3 pages, `docs024_key_docs_latest/loancalculator_couk/` 2026-07-30
- **added:** 2026-07-30, loancalculator_couk lane

### `nav-updater` DELETES every nav link whose URL sits under /tools/, /blog/, /guides/ … and puts none back
- **⚠ STATUS FIRST, BECAUSE READERS TRUNCATE — NARROWED 2026-07-31, chassis v1.0.1215+.** A rebuild now **KEEPS** a child-path link whose page declares `in_header` or `in_footer`; only a row whose page declares **neither** flag is still lost (one exists: leopardess `/tools/password-entropy.html`). Fixed by `bugs_open/149` A2, register NAV-013. Full narrowing at the `⚠ NARROWED` bullet below — **read it before acting on the headline.** *(This line was moved to the top on 2026-07-31 after the council gate's landmine lookup returned a body truncated at "the tell: there is none at the time. The run COMPLETES, th…" — i.e. it cut off ~4 bullets ABOVE the correction. Five seats, including a HIGH gating objection, then objected to a change on this file for a hazard that had already been fixed by the same lane. **A correction placed below the fold is invisible to every truncating reader, and the machines truncate.**)*
- **footprint:** `nav-updater` (agent), `populate_nav_tables` (its first step), `platform/orchestration/actions/populate_nav_tables_action.go` (`classifyPagesForNav`), `site_nav_items`
- **fires when:** you want a nav or footer change to reach the site and reach for the agent whose name says navigation. It is the obvious choice, it completes cleanly, and it also re-assembles every deployed page so the damage ships immediately
- **the tell:** there is none at the time. The run COMPLETES, the nav still looks populated (primary nav is rebuilt correctly), and only the child-path links are gone — a footer that quietly stops listing the site's tools. `populate_nav_tables` does `DELETE FROM site_nav_items WHERE site_id = $1` and rebuilds from `pages`, and `classifyPagesForNav` **skips a page entirely** when `isChildPageURL(url)` matches (`/tools/`, `/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`, `/resources/`, `/insights/`) unless `isSectionIndexType(page_type)` — which is only `blog-index`, `entity-directory`, `section-index`, `news-index`. **`tool` is not one of them**, so the skip happens BEFORE the `neverPrimaryTypes` → utility branch that would otherwise have kept it
- **the check:** before running it, count what you are about to lose —
  `SELECT s.domain, count(*) FROM site_nav_items ni JOIN sites s ON s.id=ni.site_id WHERE ni.status='active' AND (ni.url LIKE '/tools/%' OR ni.url LIKE '/blog/%' OR ni.url LIKE '/guides/%' OR ni.url LIKE '/articles/%' OR ni.url LIKE '/case-studies/%' OR ni.url LIKE '/news/%' OR ni.url LIKE '/resources/%' OR ni.url LIKE '/insights/%') GROUP BY 1;`
  Measured 2026-07-30: **16 rows across 7 domains** would be deleted — leopardess 5, gamesdesign 3, idea.uk 2, dartsonline 2, webdesign 2, fundamentallyai 1, robot-hands 1. Those rows cannot have been created by `populate_nav_tables`, so nothing recreates them
- **⚠ NARROWED 2026-07-31 (chassis v1.0.1215, council `4486f1a9` APPROVED) — STILL TRUE, but only for a page that declares NO nav flag.** `classifyPagesForNav` now demotes a child-path page to the `utility` group instead of dropping it, **provided `in_header` or `in_footer` is set** — so a rebuild KEEPS such a link where it used to delete it (6 of the 7 affected rows fleet-wide). The trap survives for a child-path nav row whose page has **both flags false**: nothing derives it, so a rebuild still removes it and puts none back. One such row exists today — leopardess `/tools/password-entropy.html`, hand-written into `utility` against a page declaring nothing — and the repair is to set the flag, not to re-add the row. **So the pre-run count below is still the right check; just read a non-zero result as "which of these are unflagged?" rather than "all of these will be lost".** The wider defect this entry recorded (a `nav_drift` item completing having placed nothing) is fixed: `bugs_open/149` A2, concept register NAV-013.
- **what to use instead:** `nav-link-fixer` refreshes `site_components.rendered_html` from the EXISTING nav tables and has no populate step; then propagate to deployed pages in **assemble mode** (`page-rerender` with NO `spec.reason`). Worked example: `docs/leopardessconsulting/scripts/reconcile_footer_nav.sh` (29/29 pages) and `refresh_owned_page_chrome.sh` for `rebuild_policy='owned'` pages, which the ordinary path cannot touch
- **two traps inside the replacement:** the assemble branch needs **`page_id`**, not `page_name` (`rerender_single_page` errors `page_id not found in input`; only the section branch resolves a name — `reconcile_headers.sh` sends `page_name` alone and fails on every page, 29/29 here). And do NOT send `reason=section_data_resolved` for a chrome change: that runs `rerender_page_sections`, whose pre-check escalates the WHOLE page to the content-writer on a missing required `source:"llm"` field — **5 of 34 active leopardess pages would escalate**
- **source:** found 2026-07-30 wiring `tool-ai-vendor-trust-checklist` into the footer; read before running, so nothing was lost
- **added:** 2026-07-30, leopardess vendor-trust-checklist build

### `platform/kafka.DeliverReply` returns the outcome that stops the silent starve — and a caller that ignores it compiles, passes platform/kafka's tests, and is unfixed
- **footprint:** `platform/kafka/reply_delivery.go` (`DeliverReply`, `DeliveryOutcome`, `FailedUndeliverable`, `IsMessageTooLarge`), and any `sendSuccessResponse`-shaped function in `internal/adapters/*` or `internal/agents/*` you are adopting it into
- **fires when:** you adopt the shared reply-delivery policy at one of the 8 remaining log-and-return sites (`bugs_open/158` item 1: reasoning, contentcreator ×2, websearch ×2, thunder ×2). The whole mechanism is a **return value**, not a side effect — `DeliverReply` cannot send your error response for you, because the error envelope differs per service
- **the tell:** there is none. `outcome, _ := kafka.DeliverReply(...)` with the outcome unused compiles cleanly, every test in `platform/kafka` still passes (they test the policy, not your call site), the log lines about degrading and undeliverability all still appear — and the caller waiting on the reply topic still gets nothing. The pod looks like it is doing the right thing because it is *saying* the right thing
- **the check:** the adoption test must live in **your** package and assert a message with `status=error_recoverable` was actually produced — not that `DeliverReply` classified correctly. Copy the shape from `internal/adapters/webscrape/reply_delivery_test.go` (`recordingProducer` counting produced messages by status header): `TestSingleScrapeOversizeReplyDegradesThenSendsErrorIfStillRefused` asserts **3** produces — full, smaller, error — and `len(sent[1]) < len(sent[0])` so a degrader that does not actually shrink is caught
- **also:** `DeliverReply` must NOT be given a degrader for a transient failure to chew on — it deliberately returns `FailedTransient` unchanged so the coordinator's retry stays in charge. If you "helpfully" degrade on every error you lose payload on failures that would have succeeded on retry
- **and:** widening adoption past webscrape **is an RFC moment**, ruled by the council's architecture seat on `7478233b` — ADP-017's registration explicitly does not pre-clear it, because it changes 4 services' caller-observable failure behaviour from timeout to error response
- **source:** `bugs_closed/133` fix + council round `7478233b`; mechanism registered as ADP-017; remaining work in `bugs_open/158`
- **added:** 2026-07-30, bugfix_133 lane

### A truncation marker in a scrape reply may only claim a stored copy by NAMING a URI — and the URI must be resolved per FIELD, not per request
- **footprint:** `internal/adapters/webscrape/truncation.go` (`transportTruncationMarker`, `topLevelFieldURIKeys`, `pageFieldURIKeys`, `pageStorageURIFor`), `truncatableTopLevelFields`, `truncatablePageFields`
- **fires when:** you add a field to the truncation list, change an upload key, or "simplify" the marker by passing a boolean instead of a URI. The last one is the tempting edit and it silently restores `bugs_closed/133`: the defect was that `"full version in S3"` was a **string literal reachable with no URI in scope**
- **the tell:** none at runtime. The reply is well-formed, the status is `complete`, and the marker reads plausibly. The only signal is `stored_copy=false` in the `Truncating large field for Kafka` log line — which exists *because* of this fix and did not exist before it
- **the check:** `go test ./internal/adapters/webscrape/ -run "TestEveryTruncatedFieldHasAURIRuling|TestURIKeysAreOnesTheUploaderActuallyWrites|TestTheOldUnconditionalClaimIsUnreachable"`. The first fails if a truncated field has neither a URI mapping nor an explicit never-uploaded ruling; the second greps `adapter.go` for `uploadInfo["<key>"]` so a mapping to a key the uploader does not write fails loudly (it `t.Fatal`s if its own needle stops matching, so it cannot pass vacuously); the third parses **string literals** with `go/ast` — deliberately not a grep, because `truncation.go` quotes the old marker in a comment
- **the per-field trap specifically:** `uploadScrapingResults` uploads each field independently and best-effort (`logger.Warn` then carry on), so a single "did we upload?" flag lies about whichever field's upload failed — **even with `upload_results: true`**. There is no request-level answer to "is there a copy?"
- **the page trap:** `storage.pages` is COMPACTED (`if len(pageInfo) > 0`), so `storage.pages[i] != result.pages[i]`. Resolve a page's URI by the index embedded IN the URI (`/page_<i>.`, trailing dot — `page_1` must not match `page_10`), never by list position, or you attach another page's URI to this page's marker. Do not delete that lookup without replacing it: it fails to the honest "discarded" form, never to a wrong URI (`bugs_open/158` item 2)
- **source:** `bugs_closed/133`; pattern in `016b` §9 "A false claim in a message is a STRING LITERAL reachable without its evidence"
- **added:** 2026-07-30, bugfix_133 lane

### A component's `var(--spacing-*)` computes to NOTHING and the whole declaration is DROPPED — the theme defines no spacing scale, only `--spacing-section`
- **footprint:** `content_components.html_template` (any component's own `<style>` block), `css_themes`, `/assets/css/styles.css`, the names `--spacing-xl` / `--spacing-lg` / `--spacing-md` / `--spacing-sm`
- **fires when:** you write CSS for a component using the spacing-scale names every design system has — `padding: var(--spacing-xl) var(--spacing-lg)` — because the theme *looks* like it has one (it defines `--spacing-section`, plus a full colour scale, `--radius*`, `--shadow*`, `--card-pad`, `--grid-gap`, `--container-pad-x`, `--section-pad-y`). It does not. Measured 2026-07-30 against the live stylesheet: **`--spacing-section` is the ONLY `--spacing-*` custom property that exists fleet-wide.**
- **the tell:** **there is none in the source, and the failure is not the one you would guess.** An undefined `var()` *with no fallback* does not fall back to the initial value of the property — it makes the declaration **invalid at computed-value time**, so the browser discards it **entirely**. `padding: var(--spacing-xl) var(--spacing-lg)` therefore yields `padding: 0`, not `padding: <something small>`. In the rendered page the text sits flush against the card border and reads as a deliberate edge-to-edge design. The stored CSS still says `padding: …`, so reading the template tells you nothing, and grepping the served HTML finds the rule present and looks fine
- **the check:** read the COMPUTED value in a browser, never the declaration — `getComputedStyle(document.querySelector('.your__el')).padding`. `0px 0px 0px 0px` on an element whose CSS clearly sets padding **is this bug**. To audit a whole template in one command, diff the names it uses against the names the theme defines:
  `curl -s https://<domain>/assets/css/styles.css | grep -oE '^\s*--[a-z0-9-]+\s*:' | tr -d ' :' | sort -u > /tmp/have; grep -o 'var(--[a-z0-9-]*' template.html | sed 's/var(--//' | sort -u | comm -23 - /tmp/have`
  Anything printed is undefined and, if it has no fallback, silently zero
- **do NOT "verify the vocabulary" and stop at colours.** The live instance is `teaser-reveal-panel`, whose style block opened with a comment stating every variable had been *confirmed present in an active theme* — true of its colours, never checked for its spacing, and **eight declarations were dead**, including the open-state body padding, the section padding, the track gap and the title margin. A partial audit reads exactly like a complete one
- **the fix that survives a theme change:** name the scale locally on the component root with a literal fallback to a variable the theme really defines — `--trp-card-x: var(--card-pad, 1.5rem)` — so a missing theme name degrades to the literal instead of to nothing
- **source:** 2026-07-30, owner reported "no space between the edge of the card and the lines" on fundamentallyai.com; computed padding measured `0px 0px 0px 0px`. `brochure_component_library/NOTES` 2026-07-30 (evening, second entry)
- **added:** 2026-07-30, brochure_component_library lane

### `\set var `backtick`` in psql runs the command INSIDE THE POD, and the empty result still COMMITs
- **footprint:** `kubectl exec -i postgres-clients-0 -- psql`, any `sql_for_agents/*.sql` using `\set tmpl \`cat …\``, `content_components.html_template`
- **fires when:** you load a file into a column with psql's backtick interpolation while piping the script through `kubectl exec`. The documented house pattern for tool components (`oufe/PREPARED_tool_insert.sql`) uses exactly this form, and it works **only** when psql runs on a machine holding the repo
- **the tell:** one line of `cat: <path>: No such file or directory` on stderr, then `BEGIN / INSERT 0 1 / … / COMMIT` — **every statement reports success** and the row is created with an **empty string, not NULL**, so a `IS NOT NULL` check passes too. `\set ON_ERROR_STOP on` does NOT help: a failing shell command is not a psql error. A tool component inserted this way is a live page bound to a zero-byte template
- **the check:** build the statement locally (dollar-quoted literal via python) and pipe the finished SQL; then **assert `length(col)` against `wc -c` of the local file** rather than trusting the `INSERT 0 1`. Byte counts must match exactly
- **source:** hit 2026-07-31 applying `sql_for_agents/275` (oufe tool 2); the corrected loading procedure is in that file's header
- **added:** 2026-07-31, oufe lane

### A hand-made `report_request` needs `handler_agent` on the ROW — the loop reads the handler from a COLUMN, not from its own config

- **footprint:** `site_work_items.handler_agent`, `report-dispatch-loop` (step `spawn_handler`), any `spawn_agent` step configured with `agent_type_field`
- **fires when:** you insert a work item by hand to exercise a pipeline — regenerating a fixture, reproducing a bug — and copy an existing row's shape by `SELECT`ing the columns you thought mattered
- **the tell:** the loop claims your item **correctly** (so the queue looks healthy), then dies at `spawn_handler` with `agent_type is required (provide 'agent_type' or 'agent_type_field')`. That message names **neither the column nor the row**, so it reads as a defect in the agent's own configuration — and the configuration is fine: `agent_type_field: "claimed.handler_agent"` resolves against the **claimed row**, so an empty column surfaces as "missing config"
- **the check:** diff your row against one that worked over **every** column, not the ones you happened to select —
  `WITH a AS (SELECT to_jsonb(w) j FROM site_work_items w WHERE id='<worked>'), b AS (SELECT to_jsonb(w) j FROM site_work_items w WHERE id='<yours>') SELECT k, a.j->>k, b.j->>k FROM a, b, jsonb_object_keys(a.j) k WHERE a.j->>k IS DISTINCT FROM b.j->>k;`
  One query named `handler_agent`. **Reach for that diff before reading the error message** — a row you built from a partial `SELECT` is missing exactly what you did not look at
- **source:** FIXTURE 4 regeneration, 2026-07-31. Same family as the existing entry for a hand-made `page_rerender` needing `page_id` in the spec **and** the column
- **added:** 2026-07-31, robot_hands_gripper_dossier lane

### `curl | grep` twice against the same URL during a deploy reports a regression that never happened
- **footprint:** any `for u in …; do curl -s "$url" | grep …; curl -s "$url" | grep …; done` verification of a live site, `049b_deploy_single_page.sh`, chrome/footer changes
- **fires when:** you verify two properties of a page by fetching it twice in the same loop iteration, while a deploy of that page is in flight
- **the tell:** one property reads present and the other reads **absent on the same page in the same second** — and re-checking by hand shows both present. On 2026-07-31 this reported oufe.com's footer honesty note as MISSING from `/about.html` immediately after a redeploy. It was there. The two `curl`s hit either side of the artefact being replaced
- **the check:** **fetch ONCE to a file, then grep the file** for every property. Costs nothing and removes the whole class. If you must re-fetch, treat a disagreement between two fetches as a timing artefact until a single saved response contradicts it
- **why it matters more than it looks:** the property it falsely reported missing was the one that `sql_for_agents/268` exists to protect, and the documented remedy for a missing footer note is a chrome regeneration — which would have *actually* deleted it. A false positive here routes you to a destructive fix
- **source:** hit 2026-07-31 verifying the oufe tool-2 footer rollout
- **added:** 2026-07-31, oufe lane

### `net/netip`'s classifier methods already handle IPv4-in-IPv6 mapped addresses — don't assume `Unmap()` is required

- **footprint:** `net/netip`, `netip.Addr.IsPrivate`, `IsLinkLocalUnicast`, `IsLoopback`
- **fires when:** writing an IP-classification check (SSRF guard, allowlist,
  anything judging whether an address is private/public) and reaching for
  `.Unmap()` before calling a classifier method, on the assumption that a
  4-in-6 address like `::ffff:169.254.169.254` needs unwrapping first to be
  judged correctly
- **the tell:** none, if you don't test it — it *reads* like the kind of thing
  that would need explicit handling, and a comment asserting so sounds
  authoritative. `IsPrivate()`, `IsLinkLocalUnicast()`, `IsLoopback()` etc. all
  already resolve the wrapped v4 address correctly with **no** unmap step:
  verified empirically, `IsPubliclyRoutable(mapped) == IsPubliclyRoutable(mapped.Unmap())`
  for every case tried
- **the check:** don't take a stdlib method's coverage on faith — write the
  positive and negative case and run it, exactly as for any other security
  claim. `Unmap()` still has a real, smaller use: `ip.String()` reads
  `"169.254.169.254"` rather than `"::ffff:169.254.169.254"` in a log or error
  message
- **source:** hit directly writing `platform/fetchguard`, 2026-07-31 — an
  earlier draft of that package's own code comment claimed unmapping was "the
  exact bypass" this needed to close; a test proved that false before the
  comment shipped. `WRONG_CALLS.md` carries the full account
- **added:** 2026-07-31, webdesign.uk lane

### `provocations.json` hardcodes its own `generated_at` — the freshness check reads a literal

- **footprint:** `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/p4_sources/build_provocations.py`,
  `https://vonc.com/data/provocations.json`, the field `generated_at`
- **fires when:** verifying that vonc's daily provocation feed is fresh, or that
  a regeneration actually regenerated anything. The natural check is
  `curl -s .../provocations.json | jq .generated_at` — and a handoff in the
  gauntlet lane recommends exactly that as its verification step 3
- **the tell:** none from outside. `generated_at` is a **Python string literal**
  (`"2026-07-26T00:00:00Z"`, `build_provocations.py:226`), not a computed value.
  Re-run the builder today, tomorrow or next year and it emits that same
  timestamp. **A file regenerated one minute ago and a file untouched for a month
  are byte-identical in the field you are using to tell them apart** — the check
  reports the stale one as fresh, for ever, and it looks like it passed
- **the check:** date the feed from git, not from its own payload —
  `gh api "repos/gqls/sites/commits?path=vonc.com/data/provocations.json"` — until
  the field is made real. More generally: **before trusting any `generated_at` /
  `updated_at` / `last_run` field, grep the producer for the literal.** A
  self-reported timestamp is only evidence if something computes it
- **source:** found 2026-07-31 picking up `gauntlet_dead_cta/HANDOFF_2026-07-30_B`,
  whose own verification step this defeats. Fixing it is Phase 0 item 1 of
  `provocation_pipeline/PLAN_2026-07-31_provocation_pipeline.md`
- **UPDATED 2026-07-31 (same day) — HALF FIXED, and the half that remains is the
  dangerous half.** The successor builder
  (`provocation_pipeline/builder/build_provocations.py`) computes the field, and
  the live feed now carries a real stamp. **The original at
  `gauntlet_dead_cta/p4_sources/build_provocations.py` still hardcodes the
  literal and is still runnable** — so running the old builder now *reverts* the
  feed's timestamp to 26 Jul while appearing to publish successfully. Two
  builders, one contract, and the wrong one silently un-fixes it. Check which
  path you are running before you publish.
- **and a NEW trap the fix creates:** a real `generated_at` proves the file was
  **rebuilt**, not that the provocation **changed**. Once a daily job exists, the
  timestamp will advance every day whether or not the schedule moved, so
  "generated_at is today" will read as "rotation works" while the site says the
  same thing for a month — the original defect wearing the fix as a disguise.
  **For "did it rotate", diff `today.slug`. Never `generated_at`.**
- **added:** 2026-07-31, provocation_pipeline lane

### vonc's provocation feed is read by the SERVER too — a client-side selector desyncs the game from the page

- **footprint:** `internal/tools-api/handlers/round.go` (`FetchProvocation`,
  `provocTTL`), `vonc.com/data/provocations.json`, the `today` key
- **fires when:** adding rotation, A/B variants, personalisation or any
  date-based selection to the provocation feed. Both vonc pages fetch this file
  **client-side**, so every visible sign says it is a browser-only asset, and the
  cheapest design — ship a pool in the JSON, pick by date in JS — looks obviously
  correct
- **the tell:** none at build time and none in the browser. `round.go:44`
  independently fetches the same URL server-side, requires the single `today`
  key, and caches it per-domain for 5 minutes. So a client-side selector makes
  the page display provocation N while the Gauntlet engine argues provocation M.
  **It only misfires on days when the selector's answer differs from `today`** —
  so it will pass every test written on the day you build it, and the bug reads
  as "the AI is answering the wrong question", pointing at the engine rather than
  at the feed
- **the check:** before treating any `/data/*.json` on a site as browser-only,
  `grep -rn "<filename>" --include="*.go"`. If a Go service reads it, the file is
  a **contract between two consumers**, and whatever chooses must run at
  generation time and write the choice into the key the server reads
- **source:** found 2026-07-31 planning provocation rotation; I had already
  half-committed to the client-side design before grepping for consumers
- **added:** 2026-07-31, provocation_pipeline lane

### `gauntlet_rounds` is not in `clients_db` — the standard psql command says it does not exist

- **footprint:** table `gauntlet_rounds`; the `kubectl -n ai-persona-system exec
  -i postgres-clients-0 -- psql -U clients_user -d clients_db` command in
  CLAUDE.md; `internal/tools-api/store/rounds.go`;
  `sql_for_agents/198_tools_api_gauntlet_rounds.sql`
- **fires when:** asking anything about vonc.com Gauntlet rounds — how many
  visitors played, what a real round contains, whether a feature has any traffic
- **the tell:** none. The repo-documented DB command returns
  `relation "gauntlet_rounds" does not exist` / `Did not find any relation`,
  which is **exactly** what an unused feature looks like. The migration file sits
  in `sql_for_agents/` like every other migration, so nothing hints that it was
  applied somewhere else. The table lives on the **island** VM in its own
  postgres container (db `tools_api`, user `tools_api`), because tools-api runs
  there and not in the cluster. A session that stops at the error will conclude
  the round store is empty or the migration never ran, and both are wrong
- **the check:** `ssh root@toolsapisuk.vs.mythic-beasts.com "docker exec \$(docker
  ps --format '{{.Names}}' | grep -i postgres | head -1) psql -U tools_api -d
  tools_api -c 'SELECT count(*) FROM gauntlet_rounds;'"` — and before believing
  *any* "table does not exist" for a tools-api table, check whether the service
  is a cluster deployment at all (`kubectl get pods | grep tools-api` returns
  nothing)
- **source:** 2026-07-31, gauntlet_dead_cta lane; cost two commands and could
  have become a false "no visitor has ever completed a round" in a handoff
- **added:** 2026-07-31, gauntlet_dead_cta lane

### `chromium` here is a SNAP and cannot write to `/tmp` — `--screenshot` fails at exit 0 through a pipe

- **footprint:** `chromium` / `/snap/bin/chromium`, `--screenshot`, `--dump-dom`,
  any headless render into the session scratchpad under `/tmp/claude-*`
- **fires when:** rendering a page or canvas to PNG for verification or for a
  mock, writing into the scratchpad directory the harness tells you to use
- **the tell:** `Failed to write file <name>: Permission denied (13)` on stderr —
  but piped through `| tail` or `| grep` the shell reports **exit 0**, and the
  PNG simply is not there. Reads as "the render produced nothing", so the
  instinct is to debug the page or the canvas code, which is fine
- **the check:** render from a `$HOME` directory and copy the result where you
  need it; assert the file exists and is plausibly sized (`> 20000` bytes for a
  1200×630 card) rather than trusting the exit status
- **source:** 2026-07-31, gauntlet_dead_cta lane, building share-card mocks
- **added:** 2026-07-31, gauntlet_dead_cta lane

### A source that stores non-ASCII as literal `\uXXXX` cannot be edited through an escape-decoding channel

- **footprint:** `p4_sources/gauntlet_js_*.js` and any file whose string
  literals use `—`, `“`, `·` rather than the characters; the Edit
  tool's `old_string`/`new_string`
- **fires when:** editing such a file — matching existing text OR writing new
  text. Not just emitting: **matching**, which is the direction that looks like
  a mystery
- **the tell:** the escape-decoding trap already recorded fleet-wide covers
  writing; the additional fact is that **no form works.** A typed `·`
  decodes to `·` (so it never matches the file's 7 literal characters) and a
  typed `\\u00B7` lands as **two** backslashes (so it never matches either). Two
  Edit attempts fail with *different* diagnostics, and the second failure's
  "mismatch is likely elsewhere" note sends you hunting the wrong part of the
  string
- **the check:** confirm the convention first —
  `python3 -c "print(repr(open(F).read()[a:b]))"` — then splice with a script
  that builds the escape from a **language** literal (in Python, `"\\u00B7"`
  evaluates to backslash + `u00B7`). Afterwards assert both directions: the
  escape is still present in the file **and** no raw character leaked in
- **source:** 2026-07-31, gauntlet_dead_cta lane, replacing the share-card
  renderer; `p4_sources/apply_card_edit.py` is the working pattern
- **added:** 2026-07-31, gauntlet_dead_cta lane

### `length()` on stored HTML is CHARACTERS; a file's size is BYTES

- **footprint:** `page_components.rendered_html`, `site_components.rendered_html`,
  `content_components.html_template`; any byte-fidelity gate comparing stored markup
  against a deployed file
- **fires when:** you compare a stored page against the file it was loaded from, or
  build the "did the bytes survive" check that verbatim/adoption work depends on
- **the tell:** a small, unexplained deficit — stored *smaller* than the file by a
  handful. `standard-calc` on loancalculator.co.uk reports `length()` = **5,730**
  against a **5,734**-byte file. The gap is exactly the **four `£` signs**, 2 bytes
  each in UTF-8. Any page with `£`, `—`, `“`, `é` has one. It reads as a truncation
  and is not
- **why it is a landmine, not a bug:** it is wrong in **both** directions and both
  look right. It fails a byte-exact page (so a faithful adoption is reported broken),
  and it can **offset a real difference against a multi-byte character** (so a
  corrupted page is reported exact). A gate that is wrong in the safe-looking
  direction is the worse half
- **the check:** compare `md5(rendered_html)` / `sha256(...)` against `md5sum` of the
  file, or `octet_length()` against `wc -c`. **Never `length()`.** Both sides of the
  loancalculator check reconcile at `14643b1f76ba4ee333d39a2ecfdf4352`
- **source:** 2026-07-31, loancalculator_couk lane, re-verifying the 27/27 byte-exact
  claim before building the render_guardian's requested in-pipeline fidelity gate
- **added:** 2026-07-31, loancalculator_couk lane
- **the repo tells you to do the wrong thing** (added 2026-07-31, gauntlet_dead_cta
  lane, hit independently — I filed a duplicate entry before finding this one, which
  is itself the argument for grepping LANDMINES.md and not just the runbooks).
  `sql_for_agents/275_oufe_tool_relevant_alternative.sql:29-30` instructs: *"ASSERT
  THE LENGTH, do not assume it: compare `length(html_template)` to `wc -c` on the
  local file. **They must match exactly.**"* For any component whose comment banner
  carries an em dash or a box-drawing rule — which is every hand-written one here —
  they cannot. Measured on `gauntlet-round-record`: `wc -c` **15,985**, `wc -m`
  **15,758**, `length()` **15,758**. A 227-character phantom deficit reported by the
  one check meant to catch truncation
- **second valid pairing:** `wc -m` (characters) *does* equal `length()`, just as
  `octet_length()` equals `wc -c`. Pair like with like or use md5; the failure is
  only ever mixing the two units
- **generators get this right for free:** Python `len(s)` on a `str` already agrees
  with `length()`, and `hashlib.md5(s.encode())` already agrees with `md5()`. A
  build script that asserts on those two cannot make this mistake — the trap is
  specific to reaching for a shell `wc`

### A tool-audit verdict is only meaningful with the harness version that produced it

- **footprint:** `docs024_key_docs_latest/webdesign_tools_repair/toolaudit.py` +
  `toolprobe.py`; any before/after tool comparison, and any `RESPONDS`/`DEAD` verdict
  quoted from a doc
- **fires when:** you compare a new audit run against a recorded baseline, or cite
  another lane's verdict. No symptom — both runs succeed and print a clean table
- **the tell:** this harness is **edited most days, by whichever lane is using it**,
  and its fixes are written against whatever site exposed the blindness. On
  2026-07-31 the committed version (`f38f5bf7f`) scored **two working
  loancalculator.co.uk calculators DEAD**: `damage-checker`'s only controls are
  checkboxes, and the driver assigned `.value` (a no-op on a checkbox — a tick is a
  `click()`); `credit-health-check` is a wizard that responds by moving a class so a
  different `<div id="step-N">` becomes visible, which `innerHTML` diffing cannot
  see. The fixes for both sat **uncommitted in the working tree** while another lane
  ran it against the same site
- **the consequence:** a baseline taken with harness A and compared against harness B
  attributes the harness's drift **to the site**, in whichever direction happens to
  flatter or condemn it. This is the prospective form of
  `a-pass-from-a-blind-check-outlives-the-blindness`
- **the check:** `sha256sum` **both** files and record it beside the verdict; re-pin
  the identical pair before the after-run. If the version you used is uncommitted,
  save `git diff` of it alongside the results (worked example:
  `loancalculator_couk/acceptance/harness_wip_vs_f38f5bf7f.diff`). Ports and profile
  dirs are randomised per run, so concurrent audits are safe — it is the *file* that
  is shared, not the browser
- **source:** 2026-07-31, loancalculator_couk lane, taking the acceptance baseline
  while the loanandmortgagecalculator lane was mid-edit on the same harness
- **added:** 2026-07-31, loancalculator_couk lane

### A page under `/tools/` need not be a tool, and `NO-CONTROL` is often the RIGHT answer

- **footprint:** `pages` rows with `url LIKE '/tools/%'`; tool acceptance gates;
  `discovery_checks/tool_eligibility.go`, `check_missing_tools.go`; any count of "how
  many tools does this site have"
- **fires when:** you size a tool inventory from the URL prefix or the page type, or
  write a gate of the form "every tool page must respond"
- **the tell:** the count is inherited from a handoff and has never been measured.
  loancalculator.co.uk was described as having **12 calculators in five separate
  documents**; it has **11**. `tools/credit-roadmap.html` is static prose that lives
  in the tools folder — zero `<input>`/`<button>`/`<select>`/`onclick`/
  `addEventListener`, and its only `<script>` is the shared `nav.js`. The same shape
  has now bitten three lanes: webdesign and mortgagecalculator each have a `/tools/`
  **hub page** whose "buttons" are all `<a>` navigation
- **why it is a landmine:** the gate is **unpassable**, and it fails in the direction
  that looks like a site defect rather than a harness fault. "10 of 11 respond" reads
  as one broken calculator for ever, and an always-failing gate gets ignored — which
  costs you the gate, not just the count
- **the check:** derive the inventory from **controls, not the path** —
  `grep -c '<input\|<button\|<select\|onclick\|addEventListener'` per page, and
  corroborate at runtime (a real-browser audit scores a genuinely static page
  `NO-CONTROL — nothing a visitor can touch`). Record the expected `NO-CONTROL` set
  **as part of the gate**, so the pass condition is `RESPONDS=11 NO-CONTROL=1` rather
  than "all respond"
- **source:** 2026-07-31, loancalculator_couk lane, building the acceptance bar for
  the `fidelity=high` decomposition
- **added:** 2026-07-31, loancalculator_couk lane

### GitHub answers "you may not see this repo" with **404**, never 403 — so token scope looks like a wrong path

- **footprint:** `GITHUB_READ_TOKEN` and the `personae-platform-secrets` secret;
  `diagnose_read_repo_files_action.go`; any new action reading a repo other than
  `gqls/agentchassis` (notably `gqls/sites`); `isRepoCloningAgent`
  (`spawn_actions.go:3066`)
- **fires when:** you point existing repo-read machinery at a **different repository**
  than the one it was built for. The code is correct, the token is valid, the request
  is well-formed, and the answer is `404 Not Found`
- **the tell:** the platform's read token is a **fine-grained** PAT scoped to selected
  repositories — measured 2026-07-31: `gqls/agentchassis` → **200**, `gqls/sites` →
  **404 while authenticated**. It is genuinely authenticated (`x-ratelimit-limit:
  5000`, not the anonymous 60), so nothing in the response says "permission". The
  message is literally `Not Found`
- **why it is a landmine:** 404 is the same answer you get for a typo, a wrong `ref`,
  a renamed directory or a case-sensitivity slip — so the diagnosis goes to the path,
  which is the one thing that is fine. The site directory in question
  (`loancalculator.co.uk/`) **does exist** on `main`
- **the check:** always fetch a **known-good positive control through the identical
  code path** — same token, same `Accept: application/vnd.github.raw`, different repo
  — and compare. One 200 and one 404 from the same credential is scope, full stop.
  Then check whether the calling agent is even on the `isRepoCloningAgent` gate, which
  is what injects the token; a non-member gets an empty env var and a different,
  louder error
- **the fix is not yours:** widening a fine-grained token's repo list needs GitHub
  admin, which is not on this machine. Route it to the owner rather than working around
  it — and note the bytes may already be reachable another way (for adoption they were
  already in `page_components.rendered_html`, verified byte-exact)
- **source:** 2026-07-31, loancalculator_couk lane, costing the "adopt from our own
  files" step before building it
- **added:** 2026-07-31, loancalculator_couk lane

### The `http_request` action is a STUB — it returns a hardcoded mock and never makes a request

- **footprint:** `http_request`, `HTTPRequestAction`, `platform/orchestration/actions/generic_actions.go`
- **fires when:** you need an orchestration step to call an external HTTP
  endpoint, grep the action registry, and find `http_request` already there —
  registered, categorised `external`, described as *"Make an HTTP request to an
  external endpoint"*
- **the tell:** there is none at dispatch time. The registry entry is real and
  reads exactly like a working action. The handler
  (`generic_actions.go:130-155`) ignores everything but `url`/`method`,
  **makes no network call at all**, and returns a literal
  `{"status": 200, "body": {"success": true, "data": "mock response"}}` — with
  a comment saying *"In a real implementation, make actual HTTP request / For
  now, return mock response"*. A step using it therefore **succeeds every
  time**, and a workflow reading `.status == 200` sees the happy path forever
- **the check:** read the handler before trusting any registry entry —
  `grep -A20 "func HTTPRequestAction" platform/orchestration/actions/generic_actions.go`.
  More generally on this platform: **a registry entry is a declaration, not an
  implementation.** Confirmed 2026-07-31 that **0 live agent definitions
  reference it**, so nothing is currently being fooled — which is also why this
  is a landmine and not a filed bug
- **why it matters now:** `webdesign_uk_build_service` P4 (the outbound
  order-pull) is precisely the shape of work that would reach for this first.
  A P4 built on it would report healthy runs while collecting nothing, forever
- **source:** found 2026-07-31 while grounding P4's plan, `webdesign_uk_build_service`
- **added:** 2026-07-31, webdesign.uk lane

### `encode(bytea,'base64')` in psql WRAPS at 76 chars — a per-line parser reads a stub and reports success

- **footprint:** `psql -t -A -c "SELECT encode(...,'base64')"`; any export of
  `rendered_html`, `html_template`, `content_data` or a bytea column out of
  `postgres-clients-0` for offline work
- **fires when:** you dump a column wider than 57 bytes and parse the output
  line-by-line (`for line in f`, `while read`, `cut -d'|'`, `awk -F'|'`)
- **the tell:** **every** decoded row is exactly **57 bytes** — 76 base64
  characters is 57 bytes of payload — and the row count is right. No error, exit 0,
  the expected number of files written. I exported 27 adopted pages this way and got
  27 plausible-looking files
- **why it is a landmine:** the failure is uniform, so nothing looks anomalous
  relative to anything else, and the next step (parse the HTML) fails in a way that
  reads as *"the pages are empty"* — a conclusion about the DATA. A whole
  decomposition was one step from being built on 57-byte fragments
- **the check:** `translate(encode(...), E'\n', '')` in the SQL, and then **verify
  each decoded row against a known-good copy** (`md5` against the deploy-repo file).
  The md5 comparison is the only reason this was caught; a size sanity check
  (`> 1000 bytes`) would also have done it. Do not trust the row count — it is right
- **source:** 2026-07-31, loancalculator_couk lane, exporting stored components to
  prove the decomposition rule offline
- **added:** 2026-07-31, loancalculator_couk lane

### A registered fact makes a GREEN claims gate meaningless as evidence of truth — the register is the authority, so it disarms every gate at once

- **footprint:** `site_specs` where `aspect='evidence_base'`, `facts[]`,
  `writer_block`; `platform/orchestration/datahelpers/claims.go` `numberSupported`;
  `claims_stats.go` `ScanStatClaims`; `save_sections_claims_guard.go`;
  `cmd/claimscan`; `docs/agent_docs/sql_for_agents/218_evidence_facts_for_043_sites.sql`
- **fires when:** you read "the claims scan returned 0 findings" as "the copy is
  true", or you conclude from a silent scan that the scan has a blind spot
- **the tell:** there is none. A number that is registered and a number that is
  unscannable both produce **exactly** silence, and the register is consulted by
  every consumer of `numberSupported` — the prose scan, the stat-field audit, and
  the persistence floor — so one wrong row disarms all three simultaneously
- **why it is a landmine:** the register is also the **writer whitelist**
  (`writer_block`, composed from `facts[].writer_line`, injected into the
  page-content-writer prompt under "NUMBERS (state only these…)"). So a false fact
  is *self-ratifying*: the platform instructs the writer to state it, then vouches
  for it. `bugs_open/161` is the live case — gamesdesign.co.uk asserts "10,000
  Monte Carlo trials per query" attributed to shipped tool JavaScript that contains
  **no randomness of any kind**, and every gate passes it, correctly
- **the check:** never stop at the verdict — ask **which fact matched**. Print the
  register alongside the scan (`SELECT jsonb_pretty(data->'facts') FROM site_specs
  WHERE aspect='evidence_base' AND is_current AND site_id=…`) and read the
  `context_terms`/`claim` wording against the artefact it cites. `refresh_evidence_base`
  re-checks a value only when `source` carries a `query`/`sql`; a
  `source.attested_by` or `source.artifact` fact is **never** machine-verified,
  and no mechanism anywhere checks a fact's *wording* against its artefact
- **source:** 2026-07-31, working the checker-layer handoff §1 — the section had
  read the same silence as a `businessClaimContextRe` vocabulary gap
- **added:** 2026-07-31, bugfix lane

### `page_component_history.source` does NOT give you the provenance `page_components` lacks

- **footprint:** `page_component_history.source`, `page_components` (no provenance
  column), `ApplySectionEditAction`, `save_page_sections`
- **fires when:** you need to know *which action or agent* wrote a component —
  e.g. bounding `ApplySectionEditAction`'s surface — spot the `source` column in
  the history table, and take it for the answer
- **the tell:** it looks like a provenance column and is populated on every row.
  It is a **write-mode** label, not a writer: `save_page_sections_overwrite` on
  **12,386 of 12,416 rows** fleet-wide, every pipeline emitting the same literal.
  The remaining 30 are hand-typed operator strings
  (`operator_copy_anchors_2026-07-29`), which is what makes the column look
  discriminating
- **the check:** `SELECT source, count(*) FROM page_component_history GROUP BY 1
  ORDER BY 2 DESC;` — if one value is ~99.8% of rows it is a mode, not a source.
  So the standing claim that `ApplySectionEditAction` cannot be bounded from
  `page_components` **stands**, and history does not rescue it
- **source:** 2026-07-31, bugfix lane, while bounding `bugs_open/161`'s witnessed case
- **added:** 2026-07-31, bugfix lane

### Grepping shipped tool code for a figure like `10000` matches `100000` — and confirms the fact you were testing

- **footprint:** any `command grep` for a bare number in `page_components.rendered_html`,
  tool JS, or an evidence-register verification ("the figure is hard-coded in the code")
- **fires when:** you verify a registered numeric fact by checking the number appears
  in the artefact its source cites
- **the tell:** none from a count. `grep -c 10000` returns 1 and you conclude the
  figure is present. In `bugs_open/161` **both** apparent hits were something else:
  a `if (pity <= 0 || pity > 100000)` bound in one tool and a
  `return Math.min(val, 10000)` **input clamp** in the other — the clamp is a real
  10000 but means "maximum attempts", not the "trials per query" the register claimed
- **why it is a landmine:** it fails in the direction of **agreeing with you**. A bare
  count would have ratified a false fact, and the fact was the thing under test
- **the check:** print the match in context (`grep -o -E '.{0,60}10,?000.{0,60}'`) and
  read what the number *does*. Also grep the mechanism, not just the magnitude — a
  "Monte Carlo trials" claim needs `Math.random` (count was **0**), so the absent
  symbol was stronger evidence than the present number
- **source:** 2026-07-31, bugfix lane, `bugs_open/161`
- **added:** 2026-07-31, bugfix lane

### `selector_count` does not assert a count — it is `selector_exists` with a friendlier detail line

- **footprint:** any criteria fence in `doc_plans.body`; `internal/adapters/browserrunner/run_checks_action.go`
- **fires when:** you author a fence and want to assert "there are 26 seats" / "there are 3 sliders"
- **the tell:** the detail line reads `26 element(s) match .rcs-seat in the live DOM` — which
  looks exactly like a count assertion and is not one. `run_checks_action.go:497` handles
  `selector_exists` and `selector_count` in **the same case arm** and passes on `n > 0`;
  `criteriaCheck` has **no expected-count field** at all. So a page that renders **1** of an
  expected 26 passes, and prints a number next to the pass that a reader will take for the check
- **why it is a landmine:** it fails in the direction of agreeing with you, and the evidence it
  prints is the thing that makes it convincing
- **the check:** assert counts through text the tool itself renders — `interaction` with
  `expect.text_matches: "^26 seats on the panel$"`. Bonus: that also proves the tool can *state*
  what it did, which a DOM count never does
- **source:** 2026-07-31, staged_component_build, authoring `tool-review-council-simulator`'s fence
- **added:** 2026-07-31, staged_component_build

### Prose naming the criteria fence in backticks silently hijacks fence extraction

- **footprint:** `doc_plans.body` for any subject; `platform/orchestration/actions/load_doc_context_action.go:143`;
  `platform/orchestration/actions/discovery_checks/check_tool_acceptance.go:552`
- **fires when:** you write or edit a PLAN that both *discusses* its acceptance criteria and
  *contains* them — i.e. any PLAN written carefully
- **the tell:** none at write time, and the failure surfaces as an unparseable-JSON error from
  `run_checks` that names neither your prose nor your fence. Both extractors are
  `strings.Index(body, "```criteria")` -> read to the next triple-backtick. They take the
  **FIRST** occurrence. A sentence like *"this PLAN previously had no ```criteria fence"* placed
  above the real fence therefore becomes the fence, and its "JSON" is a paragraph of English
- **why it is a landmine:** the PLAN is more correct-looking for having explained itself, and the
  document renders perfectly in every markdown viewer
- **the check:** the marker must appear **exactly once** in the body —
  `(length(body) - length(replace(body, '```criteria',''))) / length('```criteria')` must be 1 —
  and extract it the way the Go does, then `json.loads` the result, before you write the row.
  In prose, name it in plain words
- **source:** 2026-07-31, staged_component_build, writing the fence into `tool-review-council-simulator`'s PLAN
- **added:** 2026-07-31, staged_component_build

### `has_visible_area` is now LIVE in the running binary — and that makes it more dangerous, not less

- **footprint:** `has_visible_area` in any criteria fence; `/app/browser-runner-adapter`;
  `internal/adapters/browserrunner/run_checks_action.go:773-774`
- **fires when:** you author a gate against the newest and most useful check type, having
  correctly verified it is in the running pod
- **the tell:** the 07-30 note that this type was "committed but not rolled" is **out of date** —
  verified 2026-07-31 on `browser-runner-adapter` built 08:53:36 UTC, both long markers present
  with three positive controls. So the D7 check now says GO. **But `bugs_open/157` is unfixed at
  HEAD:** the decode comma-ok asserts `float64` while playwright-go returns `int` for a whole
  number, so **any axis whose rendered size is an integer measures 0** and the check reports
  "present in the DOM but too small to see or click". Deliberately-sized controls — 24px
  checkboxes, icon buttons, avatars — are precisely what lands on an integer
- **why it is a landmine:** D7 ("prove the type is in the running binary") passes, so the usual
  guard waves it through. Presence in the binary is necessary, not sufficient — the second
  question is whether the type is *correct*, and only a `/bugs_open/` grep answers that
- **the check:** `grep -n 'm\["w"\]' internal/adapters/browserrunner/run_checks_action.go` — if
  it still reads `.(float64)`, 157 is open. Then grep `/bugs_open/` and `/bugs_closed/` for the
  check type by name before authoring against it, not just the binary
- **source:** 2026-07-31, staged_component_build, omitting the type from a fence it belonged in
- **added:** 2026-07-31, staged_component_build
- > **CORRECTED 2026-07-31 (157 fix session) — THE CHECK ABOVE NOW GIVES THE WRONG ANSWER, and the way it is wrong is the durable lesson.** `m["w"]` no longer reads `.(float64)` (it goes through `evalNumber` as of `71680ad513`), so the stated check concludes "157 is closed" — while the running pod, measured the same afternoon, **still has the bug** (marker `non-numeric w/h in result` → 0 on `v1.0.1215`, three controls → 1). **A landmine check that reads the REPO answers "is it fixed in git", never "is it fixed in the binary my fence will run against"**, and for a Go change those two answers differ for as long as the roll takes. This entry's own headline — presence in the binary is necessary, not sufficient — has a mirror image: **absence from the repo is not evidence of presence in the binary.** So the corrected check is BOTH, in this order: grep `/bugs_open/` + `/bugs_closed/` for the check type by name, read the STATUS banner at the top of the bug file (157's names the commit and says plainly that it is not live), and only then pod-grep a long marker from the fix with a positive control in the same exec
- > **RESOLVED 2026-07-31 17:47 UTC — the repo check and the pod now AGREE, so this entry is retired as a live trap and kept as a lesson.** `v1.0.1216` carries the fix (marker → 1, five controls → 1). The window in which "the bad spelling is gone from the repo" and "the pod still has the bug" disagreed lasted **~2 hours** — 15:26 UTC (fix committed, pod on `v1.0.1215`) to 17:47 UTC. **That duration is the entry's real content:** it is long enough for a session to author a fence against a check it believes is trustworthy, and nothing in a repo grep would have warned it. The durable rule is unchanged and now recorded in `WRONG_CALLS.md` and `016b` §9: **absence from the repo is not evidence of presence in the binary** — pod-grep a long marker plus a positive control, or read the bug file's STATUS banner, which is why 157 carried one

### `page_components.content_data` often holds the SITE-WIDE BOILERPLATE, not section content — so any content-identity rule can collapse two unrelated components

- **footprint:** `page_components.content_data`,
  `platform/orchestration/datahelpers/section_text.go`
  (`NormaliseSectionText`, `SectionIdentityKey`),
  `platform/orchestration/actions/discovery_checks/check_content_duplication.go`,
  `platform/orchestration/actions/remove_duplicate_page_sections_action.go`,
  any new comparison, dedupe, similarity or diff over section content
- **fires when:** you compare two `content_data` blobs to decide whether two sections
  "say the same thing" — for a dedupe, a drift check, a similarity screen, a cache key
- **the tell:** none. The blobs are byte-identical, so every safety property built on
  "identical means interchangeable" reports green, and a deterministic no-LLM repair
  looks like the *safest* possible design
- **what is actually in there:** on `vonc.com/index.html` (measured 2026-07-31) the
  `provocation-card` and `lobby-grid` rows — **different components, different
  `component_id`** — both carry the identical site-context blob and nothing else:
  `{"tone":"","year":"2026","email":"vonc@contactforsales.com","domain":"vonc.com",`
  `"industry":"","_built_at":"2026-07-25T09:30:41Z","nav_items":[{"label":"Home",...}]}`.
  1,093 bytes each, no section content at all
- **why it is a landmine:** the pre-fix in-remit rule (same page + identical normalised
  text) found **exactly one group fleet-wide and it was this one**, and would have
  DELETED a live section from a home page. A rule can be deterministic, well tested,
  refuse-to-delete-everything guarded, and still be reading the wrong field.
  `section_text.go`'s own header credits its asset-key filter with having fixed this
  ("two unrelated sections matched at 1.00 similarity purely on captured footer/nav
  text") — it did not: the filter strips `url|href|src|image|icon|slug|id|class|target|colour|color`
  but **not** `email`, `year`, `domain`, `tone` or nav `label`s
- **the check:** require **slot equality as a NECESSARY condition** and compare the
  **canonical blob**, not the normalised prose (`SectionIdentityKey` does both). Note
  slot equality as *sufficient* is the separate, opposite error that breaks 10 real
  pages — see the `(page_id, slot_name)` unique-index entry above. And before believing
  any figure about what such a rule will do, **compile the shipped function against
  live data** (`gauntlet_dead_cta/RUNBOOK` §16b) — a SQL reimplementation of the rule
  is a second definition of "identical", which is the drift the shared helper exists
  to prevent
- **source:** 2026-07-31, gauntlet_dead_cta lane; fix `43492ec94`, council round 2 on
  corr `da3f2d9b-ae6f-492d-ad3b-748323b66367`; `WRONG_CALLS.md` same date
- **added:** 2026-07-31, gauntlet_dead_cta lane

### The diagnosis loop's `data_request` truncates a large text column at ~10.7KB — so an ABSENCE claim about a big artefact is unconfirmable by it

- **footprint:** `090_TRIGGER_needs_diagnosis_v1.sh`, the diagnose-agent `data_requests`
  path, any hypothesis citing `page_components.rendered_html`, `html_template`,
  `content_data` or another wide text column as its artefact
- **fires when:** your hypothesis turns on something NOT being in a large artefact — no
  randomness in a tool, no call to a function, no reference to a key — and you ask the loop
  to verify it
- **the tell:** the loop reports the column "truncates at the identical point" on repeated
  fetches and concludes it cannot close the gap. Measured 2026-07-31 on `bugs_open/161`:
  it read **10,704 of 21,527 characters (49.7%)** of one component, twice, cut at the same
  offset. Nothing errors; the verdict comes back `UNVERIFIABLE` / `stopped: iteration-cap`,
  which reads like the loop ran out of thinking when it actually ran out of *bytes*
- **why it is a landmine:** the truncation lands mid-artefact, so the loop sees enough to
  cite plausibly and not enough to settle the question. In 161 the tool's doc comment
  ("geometric distribution…") sat 55 chars INSIDE the cut and `Math.pow` sat 2,343 chars
  BEYOND it — so the loop could quote the artefact's *description* while never reading its
  *computation*. A partial read that quotes confidently is worse than a failed one
- **the check:** for an absence hypothesis, export the column yourself and **assert the
  exported byte count against the row's own `length()`** before reasoning from it
  (`SELECT length(rendered_html) …` then `wc -c` the export; watch stderr for
  `unexpected EOF`). Give the loop the artefact's decisive fragment IN the symptom text
  rather than pointing at the column, or expect UNVERIFIABLE. And read a
  `stopped: iteration-cap` verdict for WHICH evidence it kept asking for — the loop names
  it in `NeededEvidence`, and in 161 every substantive element was corroborated in its
  first four iterations
- **source:** 2026-07-31, bugfix lane, running the loop on `bugs_open/161` under the
  2026-07-31 owner ruling
- **added:** 2026-07-31, bugfix lane

### A criteria fence can be correct, fast locally, and still fail in the cluster on the 120s run deadline — reporting "browser open failed"

- **footprint:** any criteria fence in `doc_plans.body`; `internal/adapters/browserrunner/run_checks_action.go` (`runDeadline`, `openChromium`); `tool_acceptance_run.sh`
- **fires when:** you author or extend a fence, verify it offline, and publish it as a tool's contract
- **the tell:** the error names the **browser**, not your fence —
  `run_checks: browser open failed for <url> [mobile]: context deadline exceeded
  (code: run_checks_failed)` — so it reads as infrastructure and invites a retry. It is not.
  `runDeadline` is **120s for the WHOLE request** (every url x every profile), and
  `openChromium` returns `ctx.Err()` if the deadline expires during its settle wait, so an
  oversized fence surfaces as a browser that would not start
- **why it is a landmine:** offline verification passes with a huge margin and tells you
  nothing. Measured 2026-07-31 on the same fence: **36 evaluations = 10.6s locally (x3, stable)
  but FAILED at 133s in-cluster**; profile-gated to 22 evaluations it ran in **18s, 22 passed**.
  Budget **~3-5s per evaluation in-cluster against ~0.3s locally**. The only other acceptance
  run in history did ~21 evaluations in 48s
- **the check:** gate to desktop every check whose answer is profile-independent; keep on
  mobile only what mobile can answer differently (status, did-the-JS-run, horizontal overflow,
  console errors). Then **run it once in the cluster before believing it** — an offline PASS
  proves a fence is CORRECT, never that it FITS. Read the outcome with
  `collected_data->'request_run'->'response'->'summary'`, and note a failed run reports
  `status=COMPLETED` with `current_step='complete_error'` and the real message in `__step_error`
- **source:** 2026-07-31, staged_component_build, first cluster dispatch of `tool-review-council-simulator`'s fence
- **added:** 2026-07-31, staged_component_build

### "No page found under the name I expected" is not "no page" — and a component's own name may be absent from the page it renders on

- **footprint:** `content_components.function`; `pages.name`; `page_components`; `CHECK_naming_contract.sh`
- **fires when:** you decide whether a tool/component is an ORPHAN, or whether a component is
  present on a served page
- **the tell:** two independent blind spots that both fail towards "absent".
  (1) A name search tries a guessed convention — `pages.name = <stripped>` or a
  `%/<stripped>.html` URL — and misses a page named anything else; the URL guess also assumes
  a `<name>.html` filename, while **vonc.com uses `<name>/index.html`**, so it cannot match.
  (2) `grep`ping the SERVED html for the component's `function` returns **0** for any
  component that emits no `data-component` attribute — which is most of them
- **why it is a landmine:** `tool-arena-interface` was recorded fleet-wide as "an ORPHANED
  tool component with no page at all — decide whether the component should exist". It is
  **live, deployed and serving** on vonc.com under a page named `tool-arena`
  (`/tools/arena/index.html`, `build_status=deployed`). A decision to delete a working tool
  was one step away, taken on a measurement that never asked the question
- **the check:** ask PLACEMENT, not naming, and ask it first:
  `SELECT p.name, s.domain, p.url, pc.slot_name, pc.build_status FROM page_components pc
   JOIN content_components cc ON cc.id=pc.component_id JOIN pages p ON p.id=pc.page_id
   JOIN sites s ON s.id=p.site_id WHERE cc.function='<fn>';`
  To confirm it really renders, diff distinctive tokens from `pc.rendered_html` against the
  served page — not the function name. (Also: `content_components` has **no `site_id`**;
  `site_plan_sections` keys on `component_name`/`page_name`, not `function`/`page_id`.)
- **source:** 2026-07-31, staged_component_build, refuting its own check's orphan verdict
- **added:** 2026-07-31, staged_component_build

### `strip_comments()` in pattern-check.py is only suppress-only for checks that search for the OFFENCE — for one searching for a GUARD it INVENTS findings

- **footprint:** `scripts/pattern-check.py` (`strip_comments`, `COMMENT`, `check_stdin_eater`); any new check added to that script
- **fires when:** you add or fix a rule in `pattern-check.py` and reach for the shared `strip_comments()` helper so a comment ABOUT an invariant is not read as a violation of it
- **the tell:** the helper's own docstring states the safety property — *"it can only ever
  suppress a finding, never invent one"* — and that is **true only for a check that searches
  the stripped text for the thing it is complaining about.** `check_stdin_eater` searches the
  body for the **guard** (`</dev/null`), so stripping can delete the guard and manufacture a
  finding. Worse, `COMMENT` treats **`--`** as a comment start, and `--` is **kubectl's
  argument separator**, so the extremely common
  `kubectl exec -i pod -- psql -c '…' </dev/null` strips to `kubectl exec -i pod` and a
  correctly-guarded loop is flagged
- **why it is a landmine:** the docstring reads as a proof of safety, so you stop thinking.
  Both directions were hit within minutes on 2026-07-31 — first the rule fired on a COMMENT
  warning about the trap (`CHECK_naming_contract.sh`, whose code uses the correct
  `mapfile`+`for`), then the "fix" using `strip_comments` flagged a properly guarded loop
- **the check:** strip with a language-appropriate stripper (`#`-only for shell, and only where
  `#` starts a word so `${row#tool-}` survives), search the OFFENCE in the stripped body and
  the GUARD in the RAW body — then **run all four controls, because two of them are the ones
  that catch this**: (a) a genuine unguarded eater must fire, (b) a guarded one must not,
  (c) a genuine eater whose body carries a comment mentioning the trap must still fire,
  (d) the file that motivated the fix must be clean. Narrowing a detector until only (d)
  passes makes it inert
- **source:** 2026-07-31, staged_component_build, fixing a false positive on its own check script
- **added:** 2026-07-31, staged_component_build

### A test asserting a query is NOT issued passes VACUOUSLY against `insertWorkItem` — it swallows the error the mock raises
- **footprint:** `platform/orchestration/actions/load_work_item_actions.go` (`insertWorkItem`, the two-strike block at :1082-1118, `recurrenceExpected` at :1069), `withWorkItemTx`, and any sqlmock test in `platform/orchestration/actions` covering a work-item insert
- **fires when:** you pin a NEGATIVE about work-item insertion — "this caller sets `recurrenceExpected`, so the two-strike COUNT must never run", "this path does no lookup first" — by registering only the expectations you do want and letting sqlmock reject the rest. It is the obvious way to test a negative and it is the reason my own guard test shipped green while asserting nothing (2026-07-31)
- **the tell:** **there is none — the test is GREEN both ways.** sqlmock does raise an error for the unexpected query, and `insertWorkItem` then does `if err == nil && terminalCount > 0 { … }`, so the error is discarded, the branding it guards never happens, execution falls through to the INSERT, and every registered expectation is satisfied. `ExpectationsWereMet()` reports *missing* expectations, not *extra* calls, so it passes too. I set `recurrenceExpected: false` expecting a failure and got `ok … 0.036s`. **The mock environment masks exactly the difference the test exists to detect**, and it does so by way of the production code's own error tolerance
- **the check:** assert the mechanism's EFFECT, never the absence of a call. Register the query so it SUCCEEDS and returns state that *would* change the outcome (for the two-strike rule: `AddRow(2, 100.0)` — two prior terminal items, newest 100h old so the <3h within-cycle suppression does not also apply), then pin the INSERT's own argument: `ExpectExec(...).WithArgs(args...)` with `sqlmock.AnyArg()` in all 16 positions except `$12` (`status`), which must equal `'triaged'`. A mismatch fails the Exec, which fails the call, which fails the test. Worked example: `nav_rebuild_request_test.go:TestNavRebuildRequestSkipsTheTwoStrikeRule`, with the reasoning in its own doc comment
- **and:** do NOT then assert `ExpectationsWereMet()` in that test — on the correct path the COUNT query is legitimately never issued, so the expectation is legitimately unused, and requiring it inverts the test. **Verify every such test by breaking the thing it guards and watching it fail.** On this seam green is the default, not the finding
- **source:** hit 2026-07-31 answering the council's `guardian` objection on `bugs_open/149` A6 (round `4486f1a9`); `WRONG_CALLS.md` same date; concept register NAV-013
- **added:** 2026-07-31, bugfix_149_nav_membership lane

### `recurrenceExpected` is load-bearing on any repeated-`item_key` REQUEST — drop it and the THIRD one per site is born terminal, looking exactly like a dedup
- **footprint:** `workItem.recurrenceExpected` (`load_work_item_actions.go:1069`), `insertWorkItem`'s two-strike block (:1082-1118), `RequestNavRebuild` (`nav_rebuild_request.go`), and any new emitter that reuses a per-site or per-page `item_key`
- **fires when:** you add a work-item emitter with a stable `item_key` — "re-render this page", "rebuild this site's nav", "refresh this feed" — or you copy an existing emitter and trim what looks like an optional field. The flag reads as a tuning knob and is not one
- **the tell:** **none for the first two items.** `insertWorkItem` brands the THIRD item on a repeated `item_key` as `status = 'unresolved'` and prefixes the summary `[unresolved after N attempts]`. `unresolved` is TERMINAL (`work_items_common.go:29-35`), so the item is never claimed and never dispatched — and the emitter's return value is indistinguishable from the ordinary, correct "an open request already covers this site". So the failure appears only on the third tool (or third rerender) added to a given site, months after the code was written, and reads as successful coalescing
- **the check:** decide which of the two things you have and say so at the call site. A **DETECTED DEFECT** recurring means the fix is not working — the two-strike branding is right, leave the flag off. An **ACTION REQUEST** recurring means the previous one SUCCEEDED — set `recurrenceExpected: true`. Then pin it with a test that supplies a two-strike history and asserts the INSERT's `status` column (see the landmine directly above; the obvious test does not work). Also give a request its OWN `item_key` prefix rather than reusing a detector's, or it inherits that detector's strike history
- **why it is worth an entry rather than a comment:** this is `bugs_open/024`'s mechanism — two SUCCESSFUL re-renders poisoned a shared `item_key`, every later re-render on that site was born `unresolved`, and durable template fixes silently stopped reaching the live page for three cycles. The flag exists *because* of that, so a new emitter that omits it re-opens a closed bug rather than writing a new one. It is also `bugs_open/149` A1's "20 of 24 repeat detections born `unresolved`" seen from the writing side
- **source:** `bugs_open/024`; applied 2026-07-31 in `RequestNavRebuild` (`bugs_open/149` A6, council `4486f1a9` APPROVED — the `guardian` seat raised exactly this as a medium objection); concept register NAV-013
- **added:** 2026-07-31, bugfix_149_nav_membership lane

### Your pod-grep positive control can be INVALIDATED BY YOUR OWN CHANGE, and then a live deploy reads as a failed one
- **footprint:** any `kubectl exec … grep -ac "<marker>" /app/<binary>` deploy proof; `bugs_open/153`'s pod-grep discipline; refactors that move or reword a string literal
- **fires when:** you follow the standing rule — grep a string your change ADDED *plus* a positive control in the same exec — and you pick the control from the same function you just edited. Especially likely when the control is an error-message literal, because those are exactly what a refactor consolidates
- **the tell:** the control returns **0** on a binary that is unambiguously correct. Measured 2026-07-31 on `v1.0.1214`, both chassis replicas: the two new markers returned 2 and 1, and the chosen control `"staged plan failed validation"` returned **0** — because the change under test had replaced that literal with a runtime-assembled `"%s failed validation: %s"` where `%s` is `"staged plan"`. There is no contiguous string left to match. A single-control check here says "the grep is broken" or "nothing shipped", and both readings are wrong
- **the check:** pick a control that is **invariant under your own diff** — ideally a symbol or message in a *different* file, or one you can see unchanged in `git diff`. Cheaper still: use **two** controls, so a disagreement between them localises the fault (here `"diagnose_persist_fix_plan"` returned 11 and settled it in one exec). And remember the sibling trap already recorded: Go compiles SHORT string literals to immediate comparisons that never reach rodata, so a short control returns 0 on a binary that fully supports it — a control must be both long and untouched
- **source:** proving `bugs_open/099` candidate 2 live on `v1.0.1214`, 2026-07-31. Caught by the second control, not by care
- **added:** 2026-07-31, bugfix_099 lane


### Renaming `pages.name` silently removes the page from `check_sectionless_pages` — because that detector joins `site_plan_pages.name = pages.name`

- **footprint:** `pages.name`; `site_plan_pages.name`; `platform/orchestration/actions/discovery_checks/check_sectionless_pages.go:118`; any tool-page rename for acceptance addressability
- **fires when:** you rename a page so its tool becomes acceptance-testable (`pages.name` must
  equal the component's `function`, or `'tool-'||function`), having correctly measured that
  the served filename comes from `pages.url` and that `page_components` keys on `page_id`
- **the tell:** **there is none.** Every check you would think to run passes: the page still
  serves byte-identically, nav is unaffected (`nav_label`/`title`, not `name`), no collision,
  no `site_plan_sections` rows to re-key, and the acceptance lookup now resolves. What is
  gone is a **detection**: `check_sectionless_pages` joins
  `site_plan_pages spp ON spp.name = p.name`, so a page renamed on one side of that join
  drops out of the detector's population entirely — no error, no report, and the page then
  looks healthy precisely because nothing is examining it. On the arena rename (07-31) that
  page qualified (0 sections) and was actively reported as item `559cb636`
- **why it is a landmine:** you are *fixing* an addressability defect, so the frame is
  "make the checker able to see this tool" — and the same edit makes a different checker
  unable to see it. Trading a naming defect for a lost detection is the worse deal, and
  nothing announces the trade
- **the check:** move **both** name-side rows in one transaction, scoped **by ID** not by
  name, then **re-run the detector's own join afterwards** and confirm the page is still in
  its population under the new name:
  `SELECT p.name FROM pages p JOIN site_plans sp ON sp.site_id=p.site_id AND sp.is_current
   JOIN site_plan_pages spp ON spp.plan_id=sp.id AND spp.name=p.name
   WHERE p.site_id=$1 AND (p.sections IS NULL OR p.sections='[]'::jsonb)
     AND COALESCE(p.status,'')<>'deleted';`
  Generally: **`git grep` the COLUMN as a join key (`= p.name`, `spp.name`), not just the
  table** — a rename's blast radius is every equality join on the renamed value.
- **source:** 2026-07-31, staged_component_build, renaming vonc.com's arena page (worked example: `staged_component_build/scripts/RENAME_arena_page_to_function.sql`)
- **added:** 2026-07-31, staged_component_build

### A served-page byte baseline goes stale in minutes on this tree — take it immediately before the change, or it attributes someone else's rebuild to you

- **footprint:** any `curl`-and-compare verification of a live page; `pages.rendered_*`; page rerender / rename / republish work
- **fires when:** you prove a change is visitor-safe by diffing the served page before and after
- **the tell:** the two figures differ and the difference looks like yours. vonc.com's arena
  page measured **31,431 bytes** at ~12:50Z and **32,553** at ~15:00Z on 2026-07-31 — its own
  lane redeployed it at 12:45 between the two fetches. A rename that changed nothing would
  have "caused" 1,122 bytes if the earlier figure had been used as the baseline
- **why it is a landmine:** it fails in **both** directions. A stale baseline invents a change
  you did not make (and sends you hunting), or — worse — a real change of yours lands inside
  someone else's rebuild and reads as theirs
- **the check:** fetch the baseline **in the same minute** as the change, keep the **md5** and
  not just the size, and if they differ afterwards **diff before attributing** — several lanes
  rebuild the same sites concurrently. Sizes alone cannot distinguish "unchanged" from
  "changed by the same number of bytes". (`4a2d2030e2f6d2630f6497f68705a067`, identical both
  sides, is what made the arena rename's claim actually load-bearing.)
- **source:** 2026-07-31, staged_component_build, arena page rename
- **added:** 2026-07-31, staged_component_build

### A form control does not inherit `color`, so `background: var(--color-surface)` alone paints UA-default dark text on a dark theme
- **footprint:** any `<input>`/`<select>`/`<textarea>` in a tool component or generated CSS, `--color-surface`, `--color-text`
- **fires when:** you style a control with a themed `background` and leave `color` unset, assuming it inherits from the page. It does not — the UA stylesheet supplies a near-black default
- **the tell:** perfect in a light theme and **unreadable in a dark one**, which is the fleet default. On oufe (`--color-surface: #1B2A3B`, `--color-text: #E8E2D9`) the numbers rendered near-black on dark navy. Reported by a human looking at the live page on 2026-07-31
- **the check:** **never set one of the pair without the other.** Grep your component for `background:` and assert a `color:` in the same rule. The existing sibling tool got this right (`tool-recovery-waterfall.html` sets `color: var(--color-text, #0f172a)` on its inputs) and the line was simply dropped when copying its conventions — so *copy the whole rule, not the shape of it*
- **why no check caught it:** the tool's acceptance criteria are `selector_exists` / `no_console_errors` / `page_status_ok` / `no_horizontal_overflow` / interactions. **Not one of them can see contrast**, and the component renders, boots and computes correctly. Related: `bugs_open/122` (generated CSS fails WCAG on four live sites) and the standing rule that CSS cannot say what is PAINTED
- **source:** hit 2026-07-31 on oufe tool 2 (`tool-relevant-alternative`); fixed in the template and redeployed
- **added:** 2026-07-31, oufe lane

### `snapshot_agent()` writes to `agent_definitions_backup`, NOT an `is_snapshot` row in `agent_definitions` — so the obvious "did my snapshot happen?" check says no
- **footprint:** `snapshot_agent(text[,text])`, `agent_definitions_backup`, `agent_definitions.is_snapshot`, `snapshot_taken_at`, `snapshot_reason`, `099_SYNC_gate_roster.py`, `102_LINT_council_seat_parity.py`, `scripts/apply-seat-length-budget.py`
- **fires when:** you take a snapshot before patching a live agent definition (the practice every config-writing script here follows) and then check that it worked
- **the tell:** none, and it points the wrong way. Every query in this repo that reads live agent config filters `COALESCE(is_snapshot,false)=false`, which strongly implies snapshot rows sit in `agent_definitions` alongside the live ones. **They do not.** `snapshot_agent()` copies the live row into the separate `agent_definitions_backup` table, stamping `snapshot_taken_at` and `snapshot_reason`; `created_at`/`updated_at` are copied **verbatim from the source row**, so ordering backups by `created_at` also finds nothing recent. `SELECT count(*) FROM agent_definitions WHERE is_snapshot AND created_at > now() - interval '20 minutes'` returns **0** immediately after three successful snapshots. Two conventions coexist (the `is_snapshot` column exists and is filtered for), which is exactly why the wrong one looks right
- **the check:** `SELECT type, snapshot_taken_at, snapshot_reason FROM agent_definitions_backup WHERE snapshot_taken_at > now() - interval '30 minutes';` — and **prove it is a PRE-update copy, not a post-update one**, by asserting the backup does NOT contain your change: `… (default_config #>> '{workflow,steps,<step>,config,prompt_template}' LIKE '%<your marker>%') AS has_change` must be **false** on the backup and true on the live row. A snapshot taken after the write is not a rollback
- **source:** hit while verifying `bugs_open/138` candidate 4's rollback safety, 2026-07-31. The alarm was a wrong check, not a broken snapshot — but a thread that stopped at the wrong check would have concluded it had no rollback and either taken a redundant snapshot or abandoned the write
- **added:** 2026-07-31, bugfix_138 lane

### A chassis roll makes a scheduled task look BROKEN for ~5 minutes — check pod age before diagnosing

- **footprint:** `scheduled_tasks`, `build-pipeline-trigger`, `last_triggered_at`, `agent-chassis`
- **fires when:** you insert work for a scheduled task and watch for it to be
  picked up — a queue row, a work item, a dispatch — and nothing happens
- **the tell:** `last_triggered_at` simply stops advancing. There is no error,
  no failed row, no log line saying why. A 120s-interval task sitting still for
  5 minutes reads exactly like "this pipeline is dead" or "my row is malformed",
  and both are far more interesting hypotheses than the true one
- **the check:** `kubectl -n ai-persona-system get pods -l app=agent-chassis`
  and look at **AGE**. Anything under ~5 minutes means a roll just landed, and
  CLAUDE.md's rule applies: **no orchestration dispatch survives within ~300s of
  a chassis pod (re)start — the spawn is silently dropped.** Wait it out and
  re-check before concluding anything about the thing you were testing. Compare
  the deployment's image tag too: a changed tag confirms a roll rather than a
  crash-restart
- **worked example:** 2026-07-31, testing whether `build_queue` still seeds
  after four months dormant. Pod age was checked FIRST (5h58m — safe) and the
  test began; a roll to `v1.0.1215` landed mid-test and the trigger went quiet
  for ~5 min with the row stuck `queued`. Re-checking pod age (2m50s) diagnosed
  it in seconds. Waited out the window; the trigger fired and the test passed.
  **Without the pod-age check this would have been written up as "build_queue
  is broken" — a false negative that would have invalidated a whole plan**
- **source:** `webdesign_uk_build_service/PLAN_2026-07-31_p4_order_intake.md` §5
- **added:** 2026-07-31, webdesign.uk lane
### A chassis roll KILLS an in-flight council, and the tell is a step stuck on its OWN completed output

- **footprint:** `orchestration_states`, `currently_executing`, `execution_path`,
  `council-gate`, `council-gate-orchestrator`,
  `fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh`, `make deploy-agent-chassis`,
  any long multi-step orchestration on the shared chassis
- **fires when:** you submit a council review (or any multi-minute orchestration) and
  **another session rolls the chassis while it runs.** You cannot see their roll coming
  and nothing warns you; on this tree that is a normal Tuesday
- **the tell, and this is the part worth knowing:** the run does not error and does not
  move. `status` stays `EXECUTING_STEP`, `error` is NULL, and **`currently_executing`
  names a step whose output is ALREADY PRESENT in `collected_data`** — the step
  finished, the response came back to a consumer that no longer exists, so nothing
  advanced the state machine. `jsonb_array_length(execution_path)` reads **0** even
  though a dozen steps' results are sitting in `collected_data`. Measured 2026-07-31:
  council round 3 on corr `da3f2d9b` had 8 of 12 reviews landed plus 5 gates, last
  `updated_at` **14:59:59**, both chassis pods `startTime` **15:00:20 / 15:00:42**
- **why it is a landmine:** a stalled council is indistinguishable from a SLOW one, and
  the two demand opposite actions. CLAUDE.md correctly warns that a missing
  orchestration row is latency and that retrying costs a duplicate round — so the
  default posture is to wait, and waiting on a roll-killed run waits for ever
- **the check:** when a council has not advanced for longer than a whole prior round
  took (rounds 1 and 2 on this trail: **11 and 13 minutes** end to end), compare
  `updated_at` against pod age — `kubectl -n ai-persona-system get pods -l app=agent-chassis -o custom-columns='NAME:.metadata.name,START:.status.startTime'`.
  A pod younger than the stall means the run is dead, not slow. **Resubmit on the same
  trail** (`RESUBMIT_CORR=`), unchanged: this is an infra death, not a judgement, the
  same class the runbook already records for `complete_invalid` with no
  `council_report`. Wait out the ~300s post-restart window first, or the new spawn is
  silently dropped too
- **updated 2026-08-01 — the zombie shape EXPIRES after ~4h:** a reaper sweeps stale
  `EXECUTING_STEP` rows to `status=FAILED` with `error='reaper: stale EXECUTING_STEP
  for >4h; step=<step>'`. So the SAME kill has two appearances depending on when you
  look: within 4h it is the zombie described above (EXECUTING_STEP, NULL error);
  after, it is a FAILED row whose error names the reaper, not the cause. Both mean
  the same thing — check pod `startTime` against the last `updated_at` before the
  stall. Also: this fired **three times in 24 hours** (runs `45d143e0`, a 5-min-old
  run killed by the 23:10 roll, plus the 15:00 kill) — on this tree a >10-minute
  council has roughly coin-flip odds of surviving any given evening; budget the
  resubmit as normal cost, not an anomaly
- **source:** 2026-07-31, gauntlet_dead_cta lane; killed runs `45d143e0` (re-fired as
  `e4f81e61`, approved) and `0be8c542` (re-fired as `afd50fd4`)
- **added:** 2026-07-31, gauntlet_dead_cta lane

### `orchestration_states` has NO `id` column — it is `orchestration_id`, and the wrong name reads as "the dispatch was dropped"

- **footprint:** `orchestration_states`, any polling loop or watch script over it,
  `psql -t -A ... 2>/dev/null`
- **fires when:** you write the obvious `WHERE id = '<the id the trigger printed>'`.
  The trigger prints `RUN_ORCH_ID=`, so `id` is the natural guess; the column is
  `orchestration_id` (there is no `id` at all — also `correlation_id`,
  `parent_orchestration_id`, `currently_executing`)
- **the tell:** none, if you did what everyone does in a watch loop and sent stderr to
  `/dev/null`. `psql` raises `ERROR: column "id" does not exist`, prints it to
  **stderr**, exits, and your `$(...)` capture is an **empty string** — which prints as
  a blank field and reads as *"no row exists"*
- **why it is a landmine:** an empty result here has a specific, documented, and WRONG
  meaning ready to hand. CLAUDE.md says "a missing orchestration row is almost always
  latency, not a dropped dispatch — do not retry on that evidence." So a typo in a
  column name is laundered into a confident diagnosis of queue latency. I read a blank
  field as "no row" for **31 minutes** on 2026-07-31 while the run was in fact alive,
  then dead, and the query had never once succeeded
- **the check:** **never `2>/dev/null` a query whose emptiness you intend to interpret.**
  Drop the redirect, or assert non-empty explicitly:
  `if [ -z "$ST" ]; then echo "WARNING: no row -- query or dispatch problem"; fi`.
  Run the SELECT once by hand before putting it in a loop, and `\d <table>` first —
  schema first is in CLAUDE.md for exactly this
- **source:** 2026-07-31, gauntlet_dead_cta lane, polling council round 3
- **added:** 2026-07-31, gauntlet_dead_cta lane

### `toolaudit.py`'s `RESPONDS` cannot distinguish a working calculator from a dead one

- **footprint:** `docs024_key_docs_latest/webdesign_tools_repair/toolaudit.py` (the
  `DRIVE_JS` `snap()`/`changed` pair); any `RESPONDS`/`DEAD` verdict quoted as evidence
  that a tool WORKS; tool acceptance baselines and before/after tool comparisons
- **fires when:** you read `RESPONDS` as "this tool computes correctly" — the natural
  reading of the word, and the reading every baseline table invites
- **the tell:** the reactivity test is `changed = snap() !== before`, and `snap()`
  includes `'##' + all.map(e => e.value ?? '').join('|')` — **the value of every
  control, including the one the driver has just assigned.** So for any tool driven by
  typing (number, range, text, textarea, select) `changed` is true *by construction*,
  whatever the page does. **Proven by construction 2026-07-31:** a page with one
  `<input type=number>`, no script, no listener and nothing that could possibly
  compute scores `RESPONDS`
- **what it DOES certify, being fair to it:** no console errors, no failed
  subresources, every id-reference resolving, a non-empty main region. Real signal —
  just not arithmetic. The vacuity is confined to the reactivity component, and only
  for typed controls: for checkbox/radio tools a click does not change `.value`, which
  is exactly why `damage-checker` scored DEAD until the visibility fingerprint was
  added
- **why it is a landmine:** it fails in the flattering direction. A rewritten,
  restyled or re-ported calculator that returns wrong money — or no money — passes,
  and the verdict table reads all-green. A lane can carry "11/11 RESPONDS" into a
  handoff as proof the tools survived a port when nothing about the numbers was
  checked. Two lanes did quote it that way this week
- **the check:** compare the fingerprint **before** driving with after, excluding
  control values, and — for anything numeric — assert the output CHANGES when the
  inputs change, then record the actual values.
  `loancalculator_couk/toolgolden.py` does this (input vectors derived from each
  field's own default ×1/×2/×0.5, fingerprint = every id-bearing element's text +
  `display`) and refuses to write a golden file when a tool is inert, input-independent
  or partially captured. `check_tool_acceptance.go` will not cover you either — it
  validates that selectors EXIST and its header says static checks *"CONFIRM, never
  refute"*
- **source:** 2026-07-31, loancalculator_couk lane, building the equivalence gate for
  a tool rewrite
- **added:** 2026-07-31, loancalculator_couk lane

### Neither a grep nor a build date can prove a doc `subject_type` reached the Go gate — and the failure string you would grep for belongs to the OTHER gate
- **footprint:** `platform/orchestration/actions/doc_subjects_common.go`, `validDocSubjectTypes`, `docSubjectGateReason`, `docResolveSubject`, `platform/orchestration/actions/load_doc_context_action.go`, `platform/orchestration/actions/persist_diagnosis_note_action.go`, `doc_plans`, `doc_notes`, `experience_register/design/subject_type_addition.md`
- **fires when:** you have moved both halves of the `subject_type` contract (the DB CHECKs and the Go list — see the sibling entry above) and now want to confirm the Go half is *actually in the running fleet*. Every obvious way to check is silently inconclusive, and the third one is worse than inconclusive.
- **the tell:** none. All three wrong checks return a confident-looking answer. (1) **`psql` cannot reach the gate at all**: `load_doc_context` takes `subject_type` from **step config only** (`docResolveSubject`, `write_doc_plan_action.go:136-145`, reads `config`, never input data), so hand-writing a row proves the DB CHECK and nothing in Go — and it *feels* like an end-to-end test because the row is really there afterwards. (2) **`grep -ac` on the pod binary is meaningless either way**: the vocabulary entries are short literals (`'component'`, `'landmine'`) that Go compiles to immediate comparisons which never reach rodata, and words like `component` appear in the binary for a dozen unrelated reasons, so you get a nonzero count on a binary that does *not* support the type. (3) **A build date is necessary, not sufficient** — and it is what DOC-068 rested on for a day.
- **the third trap, which costs a stall rather than a wrong answer:** there are **two Go gates over the one shared list, with different messages**, and it is easy to quote the wrong pair. `docResolveSubject` says `subject_type must be one of …, got "…"` and is what `write_doc_plan` / `append_doc_note` / **`load_doc_context`** hit. `docSubjectGateReason` (`doc_subjects_common.go:96`) says `unsupported subject_type "…" (valid: …)` and has **exactly one caller**, `persist_diagnosis_note_action.go:83`. Two of this lane's handoffs told the next session to expect the second message from a `load_doc_context` probe — a string that appears on neither a pass nor a fail of that route (`WRONG_CALLS.md`, 2026-07-31).
- **the check:** **make the binary print its own vocabulary.** `docSubjectTypesQuoted()` joins the live slice **at runtime**, so a deliberately invalid `subject_type` returns the complete list as compiled into the running pod — the one thing a grep cannot get. Run `docs024_key_docs_latest/staged_component_build/scripts/PROBE_doc_subject_go_gate.sh <type> <key>`: two `load_doc_context` steps in one dispatch, the type under test then an invalid control that MUST error. Verified 2026-07-31 on v1.0.1215 (correlation `8f564028`), which printed `'tool', 'pipeline', 'experience', 'action', 'experience-pattern', 'component', 'landmine'`. **Do not skip the control** — without it a green run cannot be told apart from a probe that never ran, which is why the script names that outcome VOID rather than letting it read as a pass. No `agent_definitions` row is needed: the workflow travels inline in the message (`selectWorkflow` Priority 1, `platform/messaging/processor.go:922-928`), and the fallthrough is `generic`'s no-op `complete` step, so a misfire is inert *and* visible.
- **source:** 2026-07-31, staged_component_build lane, closing `HANDOFF_2026-07-31b` §3 (DOC-068's Go half); the two-message confusion is logged in `WRONG_CALLS.md`
- **added:** 2026-07-31, staged_component_build lane

### Correcting a false register fact does NOT arm detection against the copy it already caused — and arming BEFORE repairing the copy strands the pages

- **footprint:** `site_specs` `aspect='evidence_base'`, `facts[]`, `banned_claims[]`;
  `datahelpers/claims.go` `editorialPageTypes` / `ProseNumbersAreClaims` /
  `ScanBannedClaims`; `save_sections_claims_guard.go`
- **fires when:** you find a false registered fact, correct it, and reasonably assume the
  gates will now catch the published copy that fact was licensing
- **the tell:** none — the scan stays silent and silence looks like success. Measured on
  `bugs_open/161`: **0 findings before the register fix and 0 after**, with 10 live false
  components sitting there the whole time. `ScanUnregisteredNumbers` is skipped entirely on
  `editorialPageTypes` (**guide, blog-post, news-index, tool, game** — five of the fleet's
  commonest types), and marketing-shaped prose lives on those types constantly
- **the check:** correcting the fact stops the writer being INSTRUCTED and stops the engine
  VOUCHING; only a **`banned_claims` pattern** detects what is already published, because
  `claims.go` states that `ScanBannedClaims` "runs on every surface". **And the order is
  load-bearing: repair the copy FIRST, then arm.** A banned claim is BLOCKER severity, so
  arming it while the false copy is still stored makes every affected page refuse to save —
  the falsehood stays published *and* becomes unfixable through the normal path. Put a guard
  in the arming SQL that counts still-offending components and refuses on non-zero
- **also:** make the pattern require the NUMBER (or the attribution), not the bare phrase.
  `/Monte Carlo/` would have blocked the honest guide that *teaches* the technique;
  `10[,.]?000\s*Monte\s*Carlo` caught 10/10 false components with 0 false positives. Dry-run
  with `cmd/claimscan` over the whole corpus before arming, and check what it SPARES
- **source:** 2026-07-31, bugfix lane, fixing `bugs_open/161`
- **added:** 2026-07-31, bugfix lane

### A data repair RACES the sweep that publishes it — and the DB is not the website

- **footprint:** `page_components.content_data` / `rendered_html`, `pages.deployed_at`,
  `pages.build_status`, `site_work_items` `item_type='page_rerender'`
- **fires when:** you repair stored copy with SQL and treat the committed transaction as the
  fix being live
- **the tell:** `pages.deployed_at` is OLDER than your `page_components.updated_at`. On
  `bugs_open/161` the repair landed 15:28 and the 6 pages had last deployed 12:51 — the
  served HTML still asserted the falsehood while every DB check said clean
- **two traps, not one:**
  1. **`build_status='needs_rebuild'` is a DEAD queue** — 44 pages, oldest stuck since
     **2026-04-23**. Setting it strands your repair indefinitely. Measure the age of the
     oldest stuck row before trusting any status-flag route.
  2. **A concurrent sweep can publish your half-finished repair.** `rerender-pages` queued a
     full-site rerender at 15:28:33, *inside* the 15:28:04–15:28:34 window in which I was
     rewriting 10 components. Finishing seconds later would have deployed the FALSE copy and
     recorded a successful rerender. On a shared estate a data repair is not atomic with
     respect to the publishers
- **the check:** after repairing, assert `deployed_at > <your repair time>` and then read the
  **served URL**, not `rendered_html`. If you must force it, fire `page-rerender` with a
  `reason` **not** in (`image_landed`, `section_data_resolved`, `cta_links_stale`) — those
  route to `rerender_sections` (regenerate), anything else routes to `render_page`
  ("assemble stored HTML"), which republishes bytes you have already fixed and cannot
  regress an interactive tool
- **source:** 2026-07-31, bugfix lane, `bugs_open/161` step 2
- **added:** 2026-07-31, bugfix lane

### Template syntax written inside a COMMENT in `html_template` fails the parse, and the renderer silently falls back to a regex engine

- **footprint:** `content_components.html_template`, `platform/orchestration/actions/component_library.go` (`RenderTemplateReportingMissing`, `executeGoTemplate`), any `<style>` or `<!-- -->` comment in a component template
- **fires when:** you document a template's own logic in a comment beside it — the natural, conscientious thing to do when you add a conditional branch to a shared component. Write `{{if $.carousel}}` inside a CSS comment to explain the guard and Go's parser reads it as a real action. Comments are a lexer concern for CSS and HTML; `text/template` has no notion of either and scans the whole string
- **the tell:** **none on the page, and the error you get locally is not the error production gets.** `template.Parse` returns `unexpected EOF` (the unmatched `{{if}}` never closes), but `RenderTemplateReportingMissing` does **not** propagate that: it logs at Warn and drops to `renderEachBlocks`/`renderIfBlocks`/`renderGoStyleSubstitutions`, a regex fallback that handles a different, smaller dialect. So the component still renders, still reports success, and ships **mangled** markup — unrendered `{{range}}` bodies, missing sections — rather than failing. A `deployed` build_status and a non-empty `rendered_html` are both consistent with this
- **the check:** parse the template before you install it, and assert on the parse rather than on the page. `template.New("c").Parse(string(tpl))` in a throwaway Go program is enough, and it takes seconds. Do it against **every** live instance's real `content_data`, because a parse error surfaces on execute for some inputs and on parse for others. Cheap grep as a first pass: `grep -n '{{' template.html | grep -E '/\*|\*/|<!--'`
- **source:** 2026-07-31, adding the opt-in carousel branch to `info-card-grid` (leopardess lane). The byte-identity harness caught it on its first run — 18 of 18 instances returned `parse: template: c:320: unexpected EOF`. The offending text was a comment explaining that everything below it sat behind a conditional guard. Had I installed it and looked at the page, the regex fallback would have rendered *something*
- **added:** 2026-07-31, leopardess lane

### `length()` on a loaded template counts CHARACTERS, so the documented byte-count assertion FAILS on any template containing an em dash

- **footprint:** the `\set`/dollar-quote loading procedure for `content_components.html_template`, `js_snippets.js_content`, `oufe/PREPARED_tool_insert.sql`, `sql_for_agents/275`, and the LANDMINES entry above titled "`\set var `backtick`` in psql runs the command INSIDE THE POD"
- **fires when:** you follow that entry's advice exactly — *"assert `length(col)` against `wc -c` of the local file. Byte counts must match exactly"* — on a template written in this house's prose style. Every em dash, curly quote, `‹`, `›` and non-breaking space is one character and 2–3 bytes
- **the tell:** **it looks precisely like the truncation the check exists to catch.** Loading an 11,851-byte template reports `db_bytes 11841` — ten short, a plausible tail-truncation, and the natural response is to re-load or start hunting a quoting bug in a load that was byte-perfect. The gap scales with how much prose the template carries, so a heavily-commented component looks worse than a bare one
- **the check:** compare **`octet_length()`** against `wc -c`, or better, compare **`md5()`** against `md5sum` — a hash catches transposition and re-encoding, which either length check misses. `SELECT octet_length(html_template), md5(html_template) FROM content_components WHERE …` beside `wc -c file && md5sum file`. The original entry's *reasoning* stands entirely (never trust `INSERT 0 1`); only its comparison operator was wrong
- **source:** 2026-07-31, installing the `info-card-grid` carousel template: `length`=11,841 / `wc -c`=11,851 / `octet_length`=11,851 / md5 identical both sides
- **added:** 2026-07-31, leopardess lane

### A dead internal link is REPAIRED into orphaned prose, so the defect you are shown is "text that should be a link", not a 404

- **footprint:** `platform/orchestration/datahelpers/link_repair.go` (`RepairPageLinks`), `rerender_link_repair.go` (`repairOutboundPageLinks`), `validate_page_content.go`, any `content_data.link_url` / `cta_url` field, `page_components.rendered_html`
- **fires when:** you go looking for a reported broken link. `RepairPageLinks` unlinks an internal `<a href>` whose target is not a real `pages.url` and **keeps the inner content**. Where the anchor's inner content is the link *label* — which is every card, CTA and "read more" control on this platform — the served page shows the label and the arrow glyph as bare text in the middle of the card
- **the tell:** **the stored `rendered_html` and the served HTML disagree, and the stored one looks correct.** `page_components.rendered_html` holds a perfectly-formed `<a class="…" href="/services/monitoring.html">`; the deployed page has the same words with no `<a>` anywhere. Diffing DB against wire is the only way to see it, and a `curl | grep -c 'href="/services/'` returning **0** reads as "that link isn't on the page" rather than "that link was removed on the way out". Nothing is logged on the page itself; the account is in `agent_error_log` / the link-repair log, keyed by page, not by the label you are searching for
- **the second half, which is the part that bites twice:** the regex is anchored on `<a …>…</a>`, so a dead URL written as **prose** — `"use our Gap Finder at /tools/foo.html"` inside a `body` or `subheadline` field — has no anchor to unlink and ships **verbatim**. The phantom-link detector has nothing to detect either. Cleaning a page's `link_url` fields therefore does **not** clean the page: measured 2026-07-31 on one leopardess page, one invented tool URL survived a 2026-07-18 clean-up in **three** separate prose fields
- **the check:** grep the URL, not the anchor, and grep `content_data` as well as `rendered_html` —
  `SELECT slot_name FROM page_components WHERE page_id=… AND content_data::text LIKE '%<the-url>%';`
  then confirm against the wire. To find the class fleet-wide, join every `content_data` URL-shaped string against `pages.url` rather than trusting that the repair covered it
- **source:** 2026-07-31, leopardess `/services.html`: six card links to `/services/*.html` pages that never existed, all six unlinked-to-prose on the wire while the DB held valid anchors; plus three prose instances of a phantom tool URL
- **added:** 2026-07-31, leopardess lane

### A headless-Chromium gesture probe reports a WORKING carousel as dead, because `--virtual-time-budget` never advances a smooth scroll

- **footprint:** `chromium --headless=new --virtual-time-budget`, any probe asserting on `scrollLeft`/`scrollTop` after a synthetic `.click()`, `js_snippets` using `scrollBy({behavior:"smooth"})` (`hero-card-carousel`, `teaser-reveal-panel`), `brochure_component_library/scripts/probe_*.py`
- **fires when:** you do the right thing — defer the driver to `window.load` so you are not measuring the pre-init page (the landmine above about `DOMContentLoaded`), click the arrow, then read the scroll offset. Virtual time fast-forwards timers but does not drive the compositor's smooth-scroll animation, so the offset either has not moved yet or is caught mid-flight
- **the tell:** **it is indistinguishable from the real bug, and it is not even consistent.** One carousel reports `before=25 after=34` — a 9px twitch that satisfies `after > before` and *passes* — while its sibling on the same page reports `0 → 0` and *fails*, from identical code. Polling until the value changes makes it worse: you return on the first frame of the animation, so the follow-up "does prev bring it back" assertion measures from a moving baseline and fails on a working component
- **the check:** **force reduced motion** — `--force-prefers-reduced-motion`. Both snippets read `matchMedia('(prefers-reduced-motion: reduce)')` and switch to `behavior:"auto"`, which is synchronous, so the new offset is readable on the very next statement with no animation and no polling. Then prove the probe can fail, and prove it fails *for the right reason*, with **two** mutants: one that disables the snippet's `initAll` (everything JS-driven must die) and one that neutralises the CSS so the track cannot overflow (must die differently). Assert `scrollWidth > clientWidth` separately — "the handler never ran" and "this box was never scrollable" produce the same `scrollLeft` and want opposite fixes
- **also worth asserting separately:** behaviours riding the async `toggle` event (`<details>` sibling-close, URL deep-linking) need a turn of the event loop between clicks. Firing two `summary.click()` calls back to back reports sibling-close broken on a component where it works — same session, same page, second false alarm
- **source:** 2026-07-31, verifying both carousels on leopardess `/services.html`. First run reported the info-card arrows dead; the mutant control was what exposed the probe, because with `initAll` deleted the *other* carousel still reported `NEXT_SCROLLED=true` — an assertion that passes on deliberately broken code is measuring noise
- **added:** 2026-07-31, leopardess lane

### A public site data-file is read by the SERVER too, so "no page fetches this" does not license changing its shape

- **footprint:** `internal/tools-api/handlers/round.go` (`FetchProvocation`, `RoundHandler`), `https://<domain>/data/provocations.json`, `provocation_pipeline/builder/build_provocations.py`, `js_snippets` loaders, any `data-runtime-fill` shell fed by a published JSON file
- **fires when:** you are removing or renaming a field in a site's published data file and you establish, correctly, that the page which must not show it never requests it. On vonc that is verifiable and true: the Gauntlet page never fetches `provocations.json`, and today's provocation is not in its DOM hidden or otherwise. **The engine fetches the same URL server-side** — `FetchProvocation(domain)` pulls `https://{domain}/data/provocations.json`, takes the whole `today` object, and `RoundHandler` persists it as the round's provocation and returns it to the browser. Delete `today.headline`/`today.body` and every round opens blank
- **the tell — there isn't one at the page, which is the whole problem.** The site renders perfectly; the failure is at `POST /api/v1/tools/gauntlet/round`, on a **different host** (`tools.apis.uk`, per `var API` in `/tools/assets/gauntlet-interface.js` — a POST to `<site>/api/v1/...` 404s, and that 404 is your path being wrong, not the engine being down). Nothing about the published file, the git commit, or the served bytes distinguishes a shape that works from one that has just killed the core feature
- **the check, before changing ANY field of a published data file:**
  `grep -rn "<the-filename>" --include=*.go internal/ platform/` — the question is *which code reads this artefact*, never *which page fetches it*. Then exercise the server path with the right Origin: `curl -X POST https://tools.apis.uk/api/v1/tools/gauntlet/round -H 'Origin: https://<domain>' -d '{}'` and assert the payload is populated, not merely that it is 200
- **the second half, and it is the one that nearly landed:** the lane's RUNBOOK named a generator that had been **superseded hours earlier** (`gauntlet_dead_cta/p4_sources/build_provocations.py` → `provocation_pipeline/builder/build_provocations.py`). Before overwriting a published artefact, **diff the live file against what your generator produces** — the mismatch (two fields mine could not emit) is what exposed both the newer pipeline and this landmine. `gh api repos/gqls/sites/contents/<path> --jq .content | base64 -d` then compare against `curl <served-url>`; identical bytes also prove you have the right deploy repo, which `sites.github_repo` may not tell you (empty for vonc)
- **and:** three independent guards existed that would each have refused the bad shape — the owning lane's `publish_feed.sh` safety preflight, its `verify_rotation.py`, and its builder — and **none fired, because the publish was about to go through a different path than the one they guard.** "There are checks on this file" is not a reason to skip your own
- **source:** 2026-07-31, vonc seal (`gauntlet_dead_cta`, HANDOFF C + `WRONG_CALLS.md`). Committed as `bb9719d3c`, corrected in `f331dcf9d` before anything published; the shipped design keeps `today` intact and carries display copy in sibling keys (`seal`, `sample`) the engine never reads
- **added:** 2026-07-31, gauntlet_dead_cta lane

### `output_contract` is a REQUIRED field that nothing reads, and it describes output_field NAMES, not the fields inside them
- **footprint:** `agent_definitions.output_contract`, `input_contract`, `platform/orchestration/input_contracts/input_mapping.go`, `003_contracts_and_standards.md` §"Required fields"
- **fires when:** a reviewer or a checklist tells you a new agent output must be "declared in the output_contract", and you go to add it — or you add one believing it will be validated
- **the tell:** none at all, in either direction. Adding a declaration succeeds and changes nothing; omitting one also changes nothing. `003_contracts_and_standards.md:905` states plainly that *every* agent definition includes `input_contract` and `output_contract`, which reads as a hard requirement. Measured 2026-07-31: **95 of 185** live agents have one, so ~half the fleet does not — including all three consumers of the fix-loop's `diagnose_persist_fix_plan` (`feature-designer`, `fix-proposer`, `council-gate`, all NULL). And in Go, `OutputContract` (`input_mapping.go:52-53`) is a **type declaration with zero consumers**: `grep -rn OutputContract platform/ internal/ pkg/ cmd/ --include=*.go` returns only the definition. Nothing validates it, nothing fails on a mismatch
- **CORRECTED 2026-07-31, by the landmine-verifier — the first version of this entry said `OutputContract` has "zero consumers", and that was WRONG as written.** It was measured with `grep -rn OutputContract platform/ internal/ pkg/ cmd/ --include=*.go`, a scope that **excludes `scripts/`** — where two files declare the struct and bind it to the JSON (`scripts/goscripts/workflow_validator/main.go:31,39` and its `run/main.go` twin). A filter chosen from the question defined the answer. **But the underlying claim survives, and is stronger than I knew:** the function actually named `validateOutputContract` (`main.go:637`) **never reads the contract**. It walks the workflow for `complete_workflow` steps and checks their `config.output_fields` against fields other steps produce — a different thing entirely, wearing the contract's name. And `grep -rn workflow_validator makefile .githooks/ .github/ scripts/*.sh` returns **nothing**, so no target, hook or CI job invokes it at all. So: the struct is declared and parsed; **nothing validates the contract, and nothing runs the validator.** The trap is now double — a field nothing enforces, guarded by a function whose NAME says it does.
- **the check:** before "declaring a field in the output_contract", read one. The format is a flat list of **step `output_field` names** — `{"produces": ["monitor_loop"]}` — **not** the individual keys inside that output. So a field like `plan_valid` sitting inside an existing `plan_persisted` output **cannot be expressed in it at all**, and adding a block that omits it while looking thorough is worse than adding nothing. If you genuinely need a downstream-read field to be declared somewhere enforceable, the contract is not that place today — say so rather than writing a decorative block
- **source:** the council gate's `guidelines` seat raised this as a HIGH gating objection on corr `f4a4628f`, 2026-07-31; the distinction it drew (a control-flow field is not telemetry) is right, but the mechanism it pointed at does not operate at that granularity and is unenforced. Same family as "a dead config key looks exactly like a live one"
- **added:** 2026-07-31, bugfix_099 lane

### `ReadSymbolBody`'s whole-file branch is ADVERTISED TO AN LLM, so "no producer can reach it" is wrong and "tidy it away" breaks a feature
- **footprint:** `internal/analysis/symbolbody.go`, `ReadSymbolBody`, `findFile`, `platform/orchestration/actions/diagnose_assemble_bundle_action.go`, `scopeFromCodeResults`, `pkg/diagnose/loop.go`, `namedScope`, `nextScope`, `pkg/diagnose/verdict_wire.go`, `route.scope`, `input_data.seed_scope`
- **fires when:** you reason about who can put a **bare file path** (no `:Symbol`) into the diagnosis loop's scope, in order to bound what `ReadSymbolBody` may read. The obvious producer is `scopeFromCodeResults`, and it looks safe — `code_symbols.symbol` is `NOT NULL`, so its bare-path branch never fires. **That is the wrong producer.** The scope fallback chain's FIRST and highest-priority source is `route.scope` (`:139-140`), which is `EncodeScope(decision.NextScope)`, which is the **LLM verdicter's own `next_scope` JSON array** — and `namedScope` (`loop.go:409-415`) only trims and dedupes those strings. Nothing validates that an entry is a known path, or a path at all
- **the tell:** **there isn't one, and the code reads as though the branch were dead.** `bugs_open/145` was filed by the architecture seat — correctly — and still concluded the branch was "currently unreachable by accident of the writer", because it followed the DB-backed producer and stopped. The bundle text at `:597` settles it in the other direction: it prints *"put the bare file path in next_scope to see it whole"* **to the model**. So the branch is a live, advertised capability driven by the least trustworthy producer in the loop
- **the second trap, which looks like the safety net and is not:** the §7D fuzzy resolver (`diagnose_route_action.go:512-600`) classifies an unknown path-shaped entry as FUZZY and sends it to embedding search — but when nothing clears `resolver_min_similarity` it **fails open to the original string** by explicit contract ("no worse than not resolving"). It is enrichment. A `resolver_budget_seconds` timeout fails open too
- **the check:** to bound this function, gate on **membership in the analyser `Output`** (`findFile`), never on the `.go` extension and never on which producer you think is live. `Output` IS the analyser's inclusion rule (`analyse.go:80-99`: Go, non-test, minus `vendor/`, `testdata/`, download-duplicates, `excluded()`), so an extension test is a second drifting copy of it that still admits a vendored or excluded file — measured 2026-07-31, `f_test.go`, `vendor/dep/dep.go` and `testdata/sample.go` all leaked pre-fix. And **do not remove the branch**: `145`'s own fix candidate 1 (drop it / move it to an explicit `ReadWholeFile`) would break what `:597` promises the model
- **and the reason to keep the boundary IN the helper rather than in its caller:** the invariant already existed once. contextkit's `cmd/assembler` resolves `byPath[path]` and skips with "path not found in analysis" **before** calling `ReadSymbolBody`; the chassis port calls it straight off the scope list. A precondition parked in a caller is one port away from gone — which is exactly what the architecture seat meant by "a generic hazard reachable by any future scope-entry producer"
- **source:** 2026-07-31, `bugs_closed/145` (`docs024/bugfix_145_symbolbody_boundary/`). Pre-fix, a test naming files that exist on disk got all five unanalysed ones back, **plus a file from outside the checkout** (`filepath.Join(root, "../outside.go")` resolves) — so the leak was never bounded by the repo, which the filing did not claim. Consumer renders the result to the verdicter inside a ```go fence under "## In-scope code"
- **added:** 2026-07-31, bugfix_145 lane

### Guarding an asset's provenance UPSERT is not guarding the asset — the git commit already ran

- **footprint:** `platform/orchestration/actions/asset_lock_guard.go` (`lockedAssetKeys`, `assetAgentWritableSQL`), `derive_card_asset_action.go`, `derive_brand_head_assets_action.go`, `deploy_image_asset_action.go` (`sendGitCommitRequest`), `emit_sprite_css_action.go`, the `assets` table (`locked_at`, `lock_type`, `lock_expires_at`), `WHERE assets.locked_at IS NULL`
- **fires when:** you write or review an action that regenerates a site asset (a card, a favicon, an OG image) and you put the lock guard where it obviously belongs — in the `ON CONFLICT DO UPDATE ... WHERE assets.locked_at IS NULL` of the provenance upsert. That guard is real and it works. It also protects **only the DB row**, and in every one of these actions the `sendGitCommitRequest` that replaces the file in the site repo runs *first*. An owner-approved artefact is destroyed and its approving row survives to describe the file that is no longer there
- **the tell:** **the action reports `derived: true` and nothing anywhere disagrees.** `ExecContext` returns a `sql.Result` that is trivially discarded (`_, err = ...`), so a `DO UPDATE` suppressed by the lock is indistinguishable from one that wrote — no error, no row, no log. The DB looks consistent because the row you would compare against is the stale one that was protected. `bugs_open/143` sat latent for two days for exactly this reason: there was no failure to notice
- **the check, before adding or reviewing any asset writer:** call `lockedAssetKeys(ctx, db, siteID, keys...)` **before** the S3 read / image work / git commit, and keep the predicate on the write as well — the pre-check avoids the work, the `WHERE` clause is the enforcement (TOCTOU). Then read `RowsAffected`: 0 means the lock fired between the two and the artefact and its row now disagree, which must be reported and not swallowed. `go test ./platform/orchestration/actions/ -run 'AssetLock|AssetDeriv'` pins the shape and fails when a NEW `sendGitCommitRequest` producer appears
- **the trap in the other direction, and it is easy to walk into while fixing this one:** **`deploy_image_asset` has no asset-lock check anywhere and must NOT be given one.** It looks like the same defect — it commits bytes and its `UPDATE assets SET url = ... WHERE id = $1` carries no `locked_at` guard. But it commits bytes *the named row already points at*, so deploying a locked asset is publication of the approved artefact, not replacement of it; guarding that `UPDATE` would leave a locked row pointing at an expiring presigned URL. `emit_sprite_css` commits CSS and only reads the asset. The 143 class is narrower than "commits to a site repo": **regenerates an artefact from a source AND upserts the row describing it**
- **and one that will bite whoever does LOCK-004:** the asset predicate is bare `locked_at IS NULL` while the component predicate `pageComponentAgentWritableSQL` **is expiry-aware**. They are not the same rule, and unifying them WEAKENS the asset guard (an expired timed lock stops protecting the artefact). Measured 2026-07-31: 5 locked `assets` rows fleet-wide, `lock_expires_at` NULL on all 5 — so a test of the two forms against live data shows **no difference**, and the data cannot tell you which is right. Decide it on the guarantee. `asset_lock_guard_test.go` pins the current answer so a change is a decision, not a side effect
- **source:** 2026-07-31, `bugs_closed/143` (`docs024/bugfix_143_asset_lock_before_commit/`). Filed by the `bugfix_131_og_card` lane after its own `bug_historian` seat asked "does any other action share this shape?" of the brand-head fix (`e9e345464`) — the answer was yes, and the sibling had had the guard in the wrong place all along
- **added:** 2026-07-31, bugfix_143 lane

### A site's contact email lives in THREE stores and `plan_sections` could only see the emptiest one — so "the writer nests what the schema reads flat" is true and fixes nothing

- **footprint:** `platform/orchestration/actions/plan_sections_action.go` (`sourceResolver.resolve` `site_specs` branch, `resolveSpecPath`, `navigateMap`, `ensureSiteRow`, `resolveSpecAlias`), `site_specs` `aspect='identity'` (`data->'contact'`), the `sites` table's `email`/`phone`/`contact_address`/`company_name`/`tagline`/`logo_text`/`logo_url` columns, `content_components.input_schema` `source: site_specs.identity.*`
- **fires when:** you investigate why a component sourced at `site_specs.identity.<field>` does not render, find that the component asks for a **flat** path while `domain-research-classifier` writes a **nested** one (`identity.contact.email`), and conclude that repointing the path — or teaching the resolver a nested fallback — will fix it. The mismatch is real. It is also **not why the sites are broken**
- **the tell:** none from the code, and the discriminator in the data actively misleads. `bugs_open/072` measured that `contact-info` rendered on exactly the sites carrying a flat `identity.email` — no exceptions in either direction — which reads as proof that the path is the cause. It is a **confound**: those are simply the sites that have contact data *at all*. Measured 2026-07-31, the nested `contact` sub-object exists on **14 of 15** sites but its **values are null/empty on exactly the 8 that fail**, so both filed remedies resolve on **0 of 8**. Checking that the nested KEY exists is not checking that it holds anything — `jsonb ? 'contact'` is true for `{"contact":{"email":null}}`
- **the check, before proposing any fix to an identity-sourced field:** put the three stores side by side in one query and read the row, not the shape — `sites.email` / `site_specs.identity->>'email'` / `site_specs.identity->'contact'->>'email'`, per site. `sites.email` is populated on **12 of 15** real sites and covers **5 of the 8** failures; that is where the owner's data actually goes. **And filter `domain NOT LIKE 'pool-%'`** — `sites` holds 14 `pool-*.internal` rows with no content, so an unfiltered count turns "12 of 15" into "12 of 29" and makes the defect look three times worse than it is
- **the trap on the fix side:** the obvious tidy-up is to make the fallback `COALESCE` across columns the way `loadSiteDataFull` does (`company_name → name → domain`). **Do not.** That function needs a non-empty string for a template; the resolver's question is whether the field resolved *at all*, and substituting a domain for a missing company name satisfies an `on_missing: needs_human_review` field with a value nobody supplied — a silently suppressed HITL request that looks exactly like success. Missing must stay missing. The negative control is a site with no contact fact in any store (`gamesdesign.co.uk`): if it starts rendering a contact block, the fallback is fabricating
- **and the reason it took three goes fleet-wide:** this is the **third** path to need the same fix. `loadSiteDataFull` reads the columns; `buildRerenderBaseData` was changed to prefer them (`bugs_open/006` §B — *"making both render paths agree"*); `plan_sections` was missed both times. When you find a path that cannot see a canonical store, **grep for its siblings before fixing just yours**
- **source:** 2026-07-31, **`bugs_open/072`** (contact-info / identity source resolver — **not** the other `072`, component-markup-without-CSS, which is a different, closed case). Fixed and council-APPROVED, but it stays in `bugs_open/` until the chassis rolls, per this repo's fixed-AND-live bar. *(This line first said `bugs_closed/072`, corrected within the hour — writing the destination path before the file has moved is apparently easy to do: two other landmines added today cite `bugs_closed/143` and `bugs_closed/145` while both files are still in `bugs_open/`. **Cite the path the file is at, not the one you expect it to reach.**)* Root cause CONFIRMED by the diagnosis loop first iteration, corr `0f76987c`, independently citing a live `vetcomparison` row with `sites.email` populated and `site_specs.identity` all NULL. The bug file had the answer in a passing sentence — *"written only to `sites.phone`, which no component reads"* — and drew no conclusion from it
- **added:** 2026-07-31, bugfix_072 identity-source-resolver lane

### A green indexing run proves NOTHING about the prune floor — it is inert by design on a healthy repo, and its guard predicate is the DELETE's own predicate inverted

- **footprint:** `platform/orchestration/actions/prune_floor.go` (`evaluatePruneFloor`, `pruneFloorFromConfig`), `platform/orchestration/actions/code_symbols_actions.go` (`IndexCodeSymbolsAction` prune block, `codeSymbolPruneCohorts`, `recordPruneRefusal`), the `code_symbols` table (`commit_sha`, `kind`, `path`), `prune_floor_ratio`, `DELETE FROM code_symbols`
- **fires when:** you touch, review or "verify" the code-index prune. Three separate ways to get the same wrong answer:
- **1. Verifying it.** You roll the image, run the indexer, see `indexed: true` and 4,992 rows still there, and record the guard as working. **It was never asked a question.** The floor only speaks when a cohort falls below it, so a healthy repo produces a run that is byte-for-byte indistinguishable from the un-guarded behaviour. `bugs_open/135` says this in its own verification section and it is still the easiest thing in this file to get wrong. To see it fire: `INSERT` ~400 synthetic rows of one small kind (`interface` has 33) at a bogus `commit_sha`, run the indexer, and require `prune_status='refused_floor'` with `pruned=0` — then delete them and require a normal prune. Full SQL in `docs024/bugfix_135_prune_floor/RUNBOOK_prune_floor.md`
- **2. Tidying its SQL.** The DELETE is `commit_sha IS DISTINCT FROM $2`; the cohort measurement is `commit_sha IS NOT DISTINCT FROM $2`, the **exact complement**, written that way so it cannot drift. Rewriting either to `= $2`, a `COALESCE`, or any other NULL-safe spelling makes the guard judge a population that is not the one being deleted — and the prune then proceeds normally, so **the wrong result looks exactly like the right one**. Same for the measurement's position: it runs AFTER the upserts and immediately BEFORE the delete, so a symbol whose upsert failed counts as unconfirmed. Move it earlier and you lose precisely the partial-run signal
- **3. Reading `pruned: 0`.** It means one of *six* things, which is why `prune_status` exists beside it: `pruned` · `pruned_floor_disabled` · `refused_floor` · `refused_unmeasurable` · `failed` · `skipped_no_commit`. "Nothing to delete" and "we refused to delete" want opposite responses, and before this change the result could not tell them apart — a run that deleted 4,000 rows because it only saw 900 files reported `pruned: 4000` and `indexed: true`, i.e. **the alarm presented as output**
- **the tell:** for the failure the guard exists to catch, *there is none* — a partial tarball, a moved directory or a short analyser `Output` all produce a small-but-nonzero result with **no error**, and the index afterwards is not corrupt, just thin. And a thin index answers "absent" *more* confidently than it used to, because `bugs_closed/108`'s fix deliberately strengthened the empty-answer wording ("the query was RUN and matched none")
- **the check when the floor DOES refuse:** the retained rows are stale by construction, so the index now spans more than one `commit_sha` while the freshness banner (`ORDER BY updated_at DESC LIMIT 1`) names only the newest. `codeIndexScope.mixedCommitNote()` is what contradicts it, and it is rendered by **both** readers (`diagnose_code_lookup`, `diagnose_load_runtime`). If you add a third reader of `code_symbols` that draws a freshness conclusion, it needs that note too — `SELECT commit_sha, count(*) FROM code_symbols GROUP BY 1` is the one-line check for whether you are in that state
- **and one for whoever generalises it:** three other live actions delete-what-they-did-not-see with no floor — `populate_nav_tables_action.go:147` (whole-site `site_nav_items`/`site_nav_groups`), `site_db_actions.go:1474` (`link_registry`), `save_page_sections_action.go:532` (`page_components`). The rule is reusable (pure counts, no SQL) but **the cohorts are not**: each site must choose classes it can lose independently plus one signal in a *different unit* (paths seen, not rows written). Do not port the code_symbols cohorts across
- **source:** 2026-07-31, `bugs_closed/135` (`docs024/bugfix_135_prune_floor/`, register **CTXA-025**). Filed by the council's `bug_historian` seat reading the call site during review of an unrelated plan, before anything failed — the round-4 objection was *"this plan had the context and the call site open in front of it and chose to gate only the new addition, leaving the older, larger, already-populated half of the same table exposed to the identical mechanism"*
- **added:** 2026-07-31, bugfix_135 lane

### Two functions eleven characters apart decide a page's deploy path — one reads `pages.url`, one never did, and the wrong one is what the build pipelines reach

- **footprint:** `platform/orchestration/actions/git_deployer_actions.go` (`determinePageFilename`, `pageFilenameFromIdentifiers`, `ensureHTMLExtension`), `platform/orchestration/datahelpers/file_extractor.go` (`determineFilename`), `platform/orchestration/datahelpers/page_file_path.go` (`PageFilePathFromURL`, `PageDeployFilename`), `platform/orchestration/actions/rerender_single_page_action.go`, `get_pages_for_rerender_action.go`, `rerender_pages_actions.go`, the `pages.url` column, any `git_commit` step carrying `page_field`
- **fires when:** you touch anything that turns a page into a file path, or you read one of these functions to learn "how does the platform decide where a page is deployed?" — because **there were five answers and they disagreed**. `determineFilename` (datahelpers) and `determinePageFilename` (actions) differ by eleven characters in one repo; the first consults `pages.url` and the second, until 2026-07-31, never did at any priority. The one that ignored the url is the one the three BUILD agents reach (`pageflow-builder`, `page-rebuild`, `site-work-orchestrator`), so **the correct implementation being right there was no protection at all**
- **the tell:** none at the failure. The wrong path deploys **successfully** — `success: true`, a real commit, a green workflow — and the original page is left untouched and still correct, so every check that looks at the *intended* page passes. What you get is a SECOND live page at the wrong address with no `pages` row, nothing linking to it, and no sweep that removes it. It is only visible if you fetch the URL the deploy actually wrote to. `bugs_open/125` was found this way: 65% wrong at filing, **316 of 472 (67%) re-measured 2026-07-31**
- **the check, before changing any of them:** `SELECT count(*) FILTER (WHERE url <> '/'||name||'.html'), count(*) FROM pages WHERE url IS NOT NULL AND url <> ''` — if those two numbers differ, name-derived and url-derived paths are not interchangeable and you must know which one your call site produces. And grep the *derivation*, not the symptom: `grep -rn 'TrimPrefix(.*[Uu][Rr][Ll], "/")' platform/ internal/` finds the copies; grepping the bug's symptom finds one
- **the trap on the fix side — a fragment url must be DECLINED, never stripped:** exactly one live row carries a fragment (`idea.uk` / `tool-audience-check` → `/tools.html#audience-check`) and the obvious sanitisation gives `/tools.html`, **which is a different page's canonical url** (`idea.uk`/`tools`). Stripping it aims one page's rebuild at another page's file — strictly worse than the bug. The prior investigation's written pre-work recommended stripping. Generally: when a sanitiser's output lands in a namespace someone else occupies, making an input *valid* and making it *correct* are different operations
- **and the leading slash:** `pages.url` is site-absolute on **472/472** rows, and `CommitToRepo` builds `data.Domain + "/" + path` (`internal/adapters/git/github_client.go:69`), so passing a url through unchanged yields `example.com//tools/x.html` — a `//` and an empty segment in the GitHub tree. Every path in this pipeline is repo-relative and unprefixed (`assets/css/styles.css`)
- **do NOT reuse `ensureHTMLExtension` for a url:** it REPLACES an existing extension (`foo.php` → `foo.html`). That is right for a page *name* and wrong for a *url*, where the extension is authoritative and rewriting it 404s the canonical address
- **why the path was quiet for so long:** `page-rebuild` never reached `deploy_page` — it died upstream at `bugs_open/087`. Fixing 087 is what armed this. **"This path never runs, so its bugs are low priority" is often backwards**: it may never run because something upstream is broken, and your fix is what starts the traffic
- **source:** 2026-07-31, `bugs_open/125` (`docs024/bugfix_125_deploy_path_from_url/`). Filed 07-28 by the bug-sweep thread after its own 087 acceptance test published a live orphan on finetuning.uk, which the owner then had to delete by hand — the platform has no unpublish verb (`delete_repo` returns "not yet implemented"; same gap as `bugs_open/098`)
- **added:** 2026-07-31, bugfix_125 lane

### `git mv` + a pathspec commit silently ships a COPY — the rule that protects you from other sessions drops half of your own move

- **footprint:** `git mv`, `git commit <path>`, `bugs_open/`, `bugs_closed/`, `git status --porcelain`
- **fires when:** you move a file and commit it with an explicit pathspec, which CLAUDE.md requires on this tree. `git mv` stages **two** changes — a delete at the old path and an add at the new one — and a pathspec naming only the new path commits only the add. **The file then exists at BOTH paths in HEAD**, and your working tree looks perfect because on disk the move really happened
- **the tell:** none in the commit output, which reports the new file created exactly as expected, and none in your own tree. `git status --porcelain` shows a **staged `D`** for the old path afterwards — and on this tree that line is easy to dismiss, because other sessions' staged deletions sit in the same list and you have been trained to leave those alone
- **why it matters more than it sounds:** the two things most often moved here are bug case files. Closing a bug means moving it to `bugs_closed/`; a half-committed move leaves it in `bugs_open/` **as well**, so the case reads as still open — the exact and only thing that closing it was for. 2026-07-31: closing `135` and renumbering a duplicate-numbered file BOTH shipped as copies, and the renumber therefore delivered into HEAD the very number collision it had just been made to avoid
- **the check:** name **both** paths on the commit — `git commit bugs_open/OLD.md bugs_closed/NEW.md -m "..."` — and verify at HEAD, not at the tree: `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep <number>` should return exactly one line. `ls` cannot tell you this; the file is gone from disk either way
- **and the trap in avoiding the trap:** do NOT reach for `git add -A` or a bare `git commit` to sweep the deletion in. That takes every other session's staged work with it, which is the damage the pathspec rule exists to prevent. Two paths, named explicitly, is the whole fix
- **source:** 2026-07-31, bugfix_135 lane (`bugs_closed/135`, `bugs_open/165`). Caught by reading `git status` after the final commit rather than assuming the moves were complete
- **added:** 2026-07-31, bugfix_135 lane

### A work item's first-class columns are invisible to its handler — the creator that uses the schema properly is the one whose items cannot be dispatched

- **footprint:** `platform/orchestration/actions/load_work_item_actions.go` (`LoadWorkItemsAction`, `setRoutingField`), `site_work_items` columns `component_id` / `entity_id` / `affected_url` / `page_id`, `agent_definitions` rows `build-dispatch-loop` (`process_item.sub_workflow.call_handler.input_mapping`) and any handler's `query_database` `params`, `platform/orchestration/input_contracts/input_mapping.go` (`ResolveInputMapping`), `platform/orchestration/actions/create_rerender_items_action.go`
- **fires when:** you write a new work-item creator, or add a field to an existing one, and put the value in the **column** `site_work_items` provides for it. Until 2026-07-31 `LoadWorkItemsAction` selected only `page_id` of the four, so `current_item` never carried the others — the only reachable path was `current_item.spec.<key>`, a copy the creator had to remember to duplicate into the `spec` JSONB. **Using the column instead of the blob produced a structurally undispatchable item.**
- **the tell: the failure names the field as MISSING at the exact moment the row has it SET.** `tool-auditor` items died with `query param path 'input_data.component_id' resolved to nil` — and those were precisely the rows whose `component_id` column was populated, while the rows that succeeded had it NULL. Nothing warns in between: the dispatch mapping marks the field optional (`"component_id?"`), and `ResolveInputMapping` **silently skips** an unresolved optional path (`input_mapping.go:122-129`), so the value is dropped without a log line and the handler's first `query_database` is where it surfaces. Census 2026-07-31: 4 of 4 `tool-auditor` `improve_tool` items column-only and all failed; 16 of 16 items from three other creators spec-only and fine.
- **the check, before adding or reading a work-item field:** ask which of the two stores your value is in, and confirm the dispatcher can reach it — `SELECT count(*) FILTER (WHERE <col> IS NOT NULL AND NOT (spec ? '<col>')) AS unreachable_shape, count(*) FILTER (WHERE spec ? '<col>') AS spec_copies FROM site_work_items;`. A non-zero first number with agents reading `spec.<col>` is this bug. Since 2026-07-31 `component_id`/`entity_id`/`affected_url` resolve **column first, then `spec.<key>`**, so one path works either way — but `page_id` is still **column-only** by deliberate choice.
- **the trap on the fix side — the obvious simplification is the dangerous one:** backfilling the resolved value into `spec` needs no config change at all and is therefore very tempting. Do not. `rerender-pages` reads `input_data.spec.component_id` and `create_rerender_items_action.go:219` gates `scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""` on it, so a write into `spec` can **flip a site-wide rerender into a component-scoped one** in a path unrelated to whatever you were fixing. `TestSetRoutingField_NeverMutatesSpec` fails if someone tidies it up.
- **and do NOT extend the spec fallback to `page_id` "for symmetry":** measured 2026-07-31, **218** rows have a NULL `page_id` column and a `spec.page_id`. Every one would newly gain `current_item.page_id`, changing what reaches those handlers without editing them. Symmetry is not a reason to widen a handler's inputs.
- **also:** a routing key with neither source must stay **ABSENT**, never `""`. An optional mapping path that *resolves* is forwarded; one that is *missing* is skipped — so materialising an empty string converts "not supplied" into "supplied as empty" for every handler that gates on presence.
- **and the class is closed for `component_id` ONLY — do not read the entry above as covering its siblings** (narrowed 2026-07-31 by the council's `bug_historian` seat). The Go change exposes three columns; one dispatcher mapping was rewired. Measured that day: `entity_id` column set on **0** rows and read by exactly one agent (`asset-deployer`, via `input_data.spec.entity_id` — through the `spec` passthrough, **not** a dispatcher mapping, so the column-first coalesce cannot reach it); `affected_url` column set on 0 rows, read by nothing. So **the first creator to write `entity_id` on the COLUMN hits this bug again**, and fixing it then needs TWO edits, not one: add an `entity_id?` mapping to `build-dispatch-loop` AND repoint `asset-deployer` at it.
- **source:** 2026-07-31, `bugs_open/154` (`docs024_key_docs_latest/bugfix_154_work_item_routing_columns/`). Filed 07-30 by the lane that first got `improve_tool` items dispatched at all (`bugs_open/083`); the filing marked its own explanation `[INFERRED — not yet read in the code]` and named the two configs to read, which is exactly what turned it into a framework fix rather than a patch on `tool-auditor`
- **added:** 2026-07-31, bugfix 19 lane

### `max(claimed_at)` fleet-wide cannot tell "the queue recovered" from "somebody is standing next to the corpse" — and an outage ATTRACTS the traffic that poisons it

- **footprint:** `site_work_items` (`claimed_at`, `claimed_by`, `status`), `orchestration_states.current_step = 'complete_idle'`, `scheduled_tasks.last_triggered_at` / `last_completed_at`, `build-dispatch-loop`, `build-pipeline-trigger`, any question of the form "is the dispatch loop alive?"
- **fires when:** you check whether work you just filed will actually be picked up — before firing a work item, before believing a handoff that says the loop is dead, or before telling a council seat that a queue is healthy. There is no symptom: the queue looks identical whether it is running or stopped, because **the items you filed sit at `triaged` exactly as they would while waiting their turn.**
- **the trap, in three layers, each of which was written down as the fix for the one above it:**
  - **1. `scheduled_tasks.last_triggered_at` / `last_completed_at` keep advancing while nothing runs.** Fire-and-forget stamps: they prove the scheduler fired, never that a dispatch-loop orchestration was created. `build-pipeline-trigger` read `enabled`, last fired minutes ago, throughout a multi-hour outage.
  - **2. A 7-day claim rate cannot answer "will this be claimed now."** On 2026-07-31 a thread told the council gate's `bug_historian` seat that the lane was healthy, citing **1,580 of 1,664 items claimed over 7 days, with claims on every day**. Every figure was true; the lane had died at 13:21 that day and 9 items were stalled. **A near-perfect 7-day rate is precisely what a lane that died two hours ago looks like.** The seat was more right than the rebuttal.
  - **3. And the fix for (2) has the same shape one level down.** `SELECT max(claimed_at) FROM site_work_items` was added to the runbook that afternoon *as the correction*. Asked cold at 18:13 the same evening it returned **1 minute ago** — because the only lane claiming was `diagnose-dispatch-loop`, **another session's diagnosis run investigating this very outage**, while `build-dispatch-loop` had last claimed 156 minutes earlier. A stopped queue provokes exactly the kind of activity that makes the aggregate look alive, so the measure degrades *when the fleet responds correctly*.
- **the check:** never ask a fleet-wide aggregate; make the answer name the lane.
  `SELECT claimed_by, max(claimed_at), round(EXTRACT(EPOCH FROM (NOW()-max(claimed_at)))/60) AS mins_since, count(*) FILTER (WHERE claimed_at > NOW() - INTERVAL '6 hours') AS claims_6h FROM site_work_items WHERE claimed_by IS NOT NULL GROUP BY 1 ORDER BY 2 DESC;`
  Corroborate with the loop's own terminal step — `SELECT max(updated_at) FROM orchestration_states WHERE current_step='complete_idle'` — and with the backlog itself (`page_rerender` items at `triaged` with an `oldest` hours back is the tell that costs nothing to read).
- **do NOT read `claimed` grouped by STATUS instead:** `triaged | 98 | 0` reads as "nothing claims triaged items" and is definitional — a claimed item has by then MOVED to `complete`. Group by `item_type` or `pipeline` over a window, never by current status.
- **and the seam this sits on:** `detected` → `triaged` and `triaged` → claimed are **two different queues**. The fleet's "263 detected / 0 triaged" figure is about the first one and tells you nothing about the second. `RequestNavRebuild` is born `triaged` deliberately, to start on the side that works.
- **the bypass, when the lane IS stopped:** `docs024/bugfix_149_nav_membership/TRIGGER_nav_rebuild.sh <domain>` publishes the orchestrate envelope directly; it completed in ~30s with the loop dead. It only makes nav tables and stored chrome correct — **served pages still need the queue**, so say which of the two you have.
- **source:** 2026-07-31, `bugs_open/149` Group A lane (`docs024/bugfix_149_nav_membership/`, RUNBOOK **R12** queries 1–5). Layers 2 and 3 are both logged in `WRONG_CALLS.md` as separate entries by the same thread on the same day, which is the point: this is not a fact you learn once.
- **added:** 2026-07-31, bugfix_149_nav_membership lane

### Four functions compute a nav label; the one that still contains the ORIGINAL bug is the dead one, and it reads exactly like a live defect

- **footprint:** `platform/orchestration/actions/nav_tables.go` (`navSimplifyLabel`, `navLabelSegmentFromURL`), `platform/orchestration/actions/populate_nav_tables_action.go` (`navLabelForPage`, `navLabelDropCategoryPrefix`), `platform/orchestration/actions/rerender_pages_actions.go` (`rerenderSimplifyNavLabel`, `rerenderGetHeaderNavFromDB`, `rerenderGetFooterNavFromDB`), `platform/orchestration/datahelpers/nav_labels.go` (`SimplifyNavLabel`), `pages.nav_label`, `site_nav_items.label`
- **fires when:** you touch nav label behaviour, or you read one of these to learn "how does a page get its menu label?" — because there are **four** answers, `nav_tables.go:554` claims they were consolidated into one, and **they were not**.
- **1. The dead one still holds the pre-fix defect.** `rerenderSimplifyNavLabel` derives its name with `strings.TrimPrefix(url, "/")` — the **whole path** — which is exactly the bug that put `Tools/Damage Formula Designer/Index` into gamesdesign's live footer and was fixed in `navSimplifyLabel` on 2026-07-31. It looks like an urgent live bug. Its only callers are `rerenderGetHeaderNavFromDB` / `rerenderGetFooterNavFromDB`, whose **sole call site is commented out** (`rerender_pages_actions.go:168`: `/*dbNav := rerenderGetHeaderNavFromDB(...)*/`). **Do not fix it on sight** — repairing dead code spends a council round and teaches the next reader that it matters.
- **2. `datahelpers.SimplifyNavLabel(label, pageName)` takes a page NAME, not a URL.** Same-looking signature, different second argument, so it cannot title-case a path and is not at risk from anything path-shaped. Check which you have before porting a fix across.
- **3. The invariant covered only the DERIVED path for a day.** `navSimplifyLabel`'s own comment states *"a label containing a slash is never right"*, and the fix + tests made that true of every label the derivation **computes**. `navLabelForPage` returns a planner-**authored** `pages.nav_label` **verbatim** when it is ≤30 chars — slash and all — so the guarantee stopped at the function boundary beside the reviewed one. Measured 2026-07-31: 8 live pages with a `Tools / …` nav_label, 2 trusted verbatim, and **ai-agent-orchestration.com was already serving `Tools / AI Readiness Quiz`**. Fixed by `navLabelDropCategoryPrefix` (commit `6fc1ff3ed`, council `11c5c813`).
- **the check, before believing a label fix works:** the property that decides whether the derivation runs at all is **`nav_label IS NULL`**, not page count. A site whose labels were typed in by hand gives a clean, meaningless pass — which is how a one-site verification on gamesdesign (0 NULL labels of 6 tool pages) proved nothing about the code. `SELECT s.domain, count(*) FILTER (WHERE p.nav_label IS NULL OR p.nav_label='') AS null_label, count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE p.status IN ('active','deployed','pending') AND p.url ILIKE '/tools/%' AND (COALESCE(p.in_header,false) OR COALESCE(p.in_footer,false)) GROUP BY 1 ORDER BY 2 DESC;`
- **and the trick that makes a label proof decisive:** pick a page whose title and URL slug **disagree in spelling**. gaswholesalers' title is `Wholesale Break-Even Volume Calculator`; the derived label is `Breakeven Volume Calculator`. `Break-Even` vs `Breakeven` proves the label came from the URL and not from the title by coincidence. A label that merely looks right identifies no code path.
- **and if you verify against rendered chrome, match the FIELD not the document:** `rendered_html ILIKE '%Tools/%'` returns **true on a healthy footer**, matching the `href` `/tools/…`. Extract anchor text. `site_components` uses **`slot_name`** (`header`/`footer`/`head`), not `component_type`.
- **source:** 2026-07-31, `bugs_open/149` A6 (`docs024/bugfix_149_nav_membership/`, RUNBOOK **R14–R16**), concept register **NAV-013**
- **added:** 2026-07-31, bugfix_149_nav_membership lane

### `data-runtime-fill` is tested against WHATEVER YOU PASS — hand it a whole page and you switch the check off for every section on it

- **footprint:** `data-runtime-fill`, `platform/orchestration/datahelpers/runtime_fill.go`, `platform/orchestration/datahelpers/link_repair.go`, `platform/orchestration/actions/render_site_components_action.go`, `platform/orchestration/actions/rerender_single_page_action.go`, `page_components.rendered_html`, `scripts/pattern-check.py`
- *(footprint line added 2026-08-01 by the gauntlet_dead_cta lane: the custom label below does not match `FIELD_RE`'s `[a-z ]+`, so this entry had NEVER synced to `doc_notes` and its SessionStart hook never fired. The author's full list is preserved verbatim underneath.)*
- **footprint — SITES THAT TEST THE MARKER** (verified 2026-07-31 with the gate's own regex over the whole tree; **7** of them, all allow-listed in `scripts/pattern-check.py`): `data-runtime-fill`, `platform/orchestration/datahelpers/runtime_fill.go` (`RuntimeFillSpans`, `HasRuntimeFillMarker`, `InRuntimeFillShell`, `DeadControlAnchorsOutsideRuntimeFill` — the owner), `datahelpers/link_repair.go` (`RepairPageLinks`), `discovery_checks/check_tool_acceptance.go`, `check_dead_controls.go`, `check_phantom_internal_links.go`, `check_backend_entry_orphaned.go`, `check_empty_sections.go`, `static_attribute_checks.go`, `check_required_fields_missing.go`, `check_component_standards.go`, `check_component_template_corrupted.go`, `actions/render_site_components_action.go`, `actions/rerender_single_page_action.go`, `page_components.rendered_html`
- **NOT footprint — files that only CALL a tester** (`rerender_link_repair.go`, `validate_page_content.go`, `save_sections_link_repair.go`): they contain **zero** marker tests; `RepairPageLinks` owns the test on their behalf. ⚠ **an earlier version of this entry listed them in the footprint and a council reviewer read that as an uncovered call site** — a footprint that mixes testers with callers manufactures exactly the false positive it exists to prevent. If you are hunting instances, grep the marker, not this list
- **fires when:** you call any link/control check, or `RepairPageLinks`, with **more than one section** of HTML — an assembled page, a `string_agg` of components, or a fetched URL. Until 2026-07-31 the exemption was `strings.Contains(html, "data-runtime-fill")` at eight sites, so **one hydrating section anywhere in the input exempted every unrelated section**. Measured on assembled `vonc.com/index`: the whole-input test excuses **100%** of 48,956 bytes where the element-scoped one excuses **12.6%**.
- **the tell: there is none — the check returns cleanly and reports nothing.** A masked page is indistinguishable from a clean one in every output the check produces. This is why it survived at eight call sites: read at any one of them the line is obviously correct, because the reader supplies section-shaped input in their head. It cost two live dead controls ("Get Started", "Learn More") and two unrepaired empty hrefs ("Enter the Gauntlet", "Find Your Archetype") — **the last two being the exact controls `check_dead_controls.go` names in its own header as the case it was built for**.
- **the gate:** `scripts/pattern-check.py` `check_runtime_fill_marker` fires at COMMIT time on any raw marker test — the Go string forms, the `regexp.MustCompile` form and the SQL `LIKE '%data-runtime-fill%'` form — unless the file is in `RUNTIME_FILL_ALLOWED` with a reason. Advisory, never blocks. It lives there rather than in a Go test deliberately: a second enforcement path with its own allow-list is the duplication this entry is about
- **the check, before you call one of these with page-level HTML:** ask what your input is, then use the predicate that matches the QUESTION — `RuntimeFillSpans`/`InRuntimeFillShell` for *"is this CONTROL alive?"*, `HasRuntimeFillMarker` for *"is this SECTION a shell?"*. Confirm your own input scope with `grep -n 'rendered_html\|string_agg\|fetchDeployedPage' <your file>`: a row loop over one component is section-scoped and safe; anything else is not.
- **the trap on the fix side — do NOT redirect the section-scoped callers at the element-scoped predicate "for consistency".** `check_empty_sections`, `check_component_standards`, `check_component_template_corrupted` and `sectionHasVisibleContent` genuinely ask the whole-section question, and switching them changes a different question's answer. `TestHasRuntimeFillMarkerIsStillWholeInput` exists to fail if someone tidies this up.
- **and a second copy the Go grep cannot see:** three checks test the marker in **SQL** (`cc.html_template LIKE '%data-runtime-fill%'`, `COALESCE(pc.rendered_html,'') LIKE '%data-runtime-fill%'`). Grepping `--include=*.go` reports them as absent. Grep the marker string across the repo, not the helper name — an inlined predicate has no name to find.
- **CORRECTED 2026-07-31 (same day, council round 1) — the rule is NOT "element-scope everything".** The first version of this entry implied the element-scoped predicate is always the right answer. It is not: **narrowing a JUDGE is safe, narrowing a WRITER is not.** A judge (a discovery check) that sees more markup can only raise more findings, and each is escalated to a human. A writer (`RepairPageLinks`, `DropDeadURLControls`) that sees more markup **acts** on it — and `RepairPageLinks`' action is to strip the `<a>` and keep the text, which is the *"REPAIRED into orphaned prose"* landmine two entries below this one. So the writers keep whole-input scope deliberately, because for a writer the wide exemption UNDER-repairs and under-repair is fail-safe: an unrepaired control stays visible and stays flagged, a repaired one becomes prose nobody can find. **Before you narrow this exemption at a new call site, ask whether that site JUDGES or WRITES.**
- **and the gate has an allow-list, which is not a loophole:** `rerender_single_page_action.go` keeps a raw `regexp.MustCompile("(?i)data-runtime-fill")` on purpose — it asks the section question, and its `(?i)` makes it **the only case-insensitive marker test in the tree**. Converting it to the case-sensitive `HasRuntimeFillMarker` would be a silent behaviour change to the page assembler. If you are unifying these, that divergence is a decision someone must make, not a tidy-up.
- **source:** 2026-07-31, `bugs_closed/137` (`docs024_key_docs_latest/bugfix_137_control_liveness/`, 016b §9, LNK-025) — **CLOSED + LIVE on v1.0.1223**, both branches induced. Filed 07-28 as "two disagreeing judges of control liveness" at the council `reuse_agent` seat's request; the disagreement turned out to be a **symptom** of the scope, not a dispute about the rule
- **added:** 2026-07-31, bugfix 137 lane

### Editing the ACTIVE chrome component changes nothing, and correcting the SELECTION does not repoint the 11 sites that already point at the wrong one

- **footprint:** `content_components` (`function` = `site-header`/`site-footer`/`head`, `is_active`, `forked_from`, `component_level`), `site_components` (`component_id`, `slot_name`), `platform/orchestration/actions/component_library.go` (`ResolveChromeComponent`, `GetComponentByFunction`, `RenderHeader`, `RenderFooter`, `RenderHead`), `platform/orchestration/actions/render_site_components_action.go`, `platform/orchestration/actions/link_site_components_action.go`, work item type `deactivated_component`
- **fires when:** you fix anything in a site header or footer — a heading, a gate, a link, a GTM tag — and reach for "the active component" as the obvious target.
- **1. The active component is very likely not the one that renders.** Until `bugs_open/118` shipped, chrome selection had **three** predicates and all three were wrong differently: `render_site_components` had **none at all** (picked `footer-4-column`, `is_active=false`), `link_site_components` filtered on `is_active` alone (picks `header-leopardess` — an **active FORK** of one client's header, because a fork carries its parent's `function`), and `GetComponentByFunction` filtered correctly but unordered, and its winner `site-header` is `component_level='section'` — a page-section component serving as chrome. The **build path still does the third one**, deliberately: switching it changes chrome markup fleet-wide and is an owner call, not a bug fix.
- **2. The fix to the selection is NOT a fix to your site.** `site_components.component_id` pins the choice per site, and nothing in the 118 fix repoints an existing row. **11 of 14 sites are pinned to `footer-4-column`, 7 to `header-bold-gradient`, 9 to `Document Head` — all `is_active=false`.** A correct selection only decides what an *unassigned* slot gets.
- **3. The platform detects this and its repair cannot repair it.** `deactivated_site_components` raises `deactivated_component` items (since 2026-07-17) routed to `rerender-pages`, which re-renders **the component the row already points at** — the deactivated one. Two live items read `[unresolved after 2 attempts]`. **A `complete` deactivated_component item is not a repointed slot.**
- **the check, before you edit a chrome component:** ask the ROW, not the library. `SELECT sc.slot_name, cc.name, cc.is_active, cc.component_level FROM site_components sc JOIN content_components cc ON cc.id=sc.component_id JOIN sites s ON s.id=sc.site_id WHERE s.domain='<yours>';` Then confirm at the artefact with a string only that component carries — `curl -s https://<domain>/index.html | grep -o '<h[34]>[^<]*</h[34]>'` (`Our Services` = `footer-4-column`; `Explore` = `footer-theme-chrome`).
- **and remember 117:** a chrome change is invisible until `render_site_components` runs. Editing the right component and seeing no change may mean nothing has re-rendered yet, which looks exactly like editing the wrong one.
- **source:** 2026-07-31, `bugs_open/118` (`docs024_key_docs_latest/bugfix_118_chrome_selection/`), concept register **CLC-013**, council `5bc232d6-590a-4476-a6b1-4fb6f61751c6`. The bug was found because an owner-approved gate applied to the *active* footer had no effect on any page after a full chrome rebuild.
- **added:** 2026-07-31, bugfix_118_chrome_selection lane

### `LIKE '…' || purpose || '…'` matches the HYPHENATED filename only because `_` is a SQL wildcard — "escaping" it manufactures 38 false findings

- **footprint:** `platform/orchestration/actions/discovery_checks/check_undeployed_assets.go` (`findUndeployedAssets`), `assets.purpose`, `page_components.rendered_html`, and any SQL that builds a path pattern by concatenating a `purpose`/`asset_key`/`slot_name` value into a `LIKE`
- **fires when:** you read that query and notice the unescaped `_`. Every instinct says a value interpolated into a `LIKE` pattern should have its wildcards escaped, and it is the kind of one-line "correctness" fix that looks unarguable in review.
- **the tell: there is none, and the sign is inverted from what you expect.** Escaping makes the predicate STRICTER, so it does not silence findings — it creates them, and they look like a detector that has just started working. The published filenames are hyphenated (`content-hero…`, `og-card.png`) while the purposes are underscored (`content_hero`, `og_card`), and `_` matching any character is the ONLY reason the pattern ever matches. Measured 2026-07-31: for `content_hero`, `deployed_UNESCAPED` = **38 of 38**, `deployed_ESCAPED` = **0 of 38**. Thirty-eight brand-new false findings, on a check whose output nobody drains (`bugs_open/083`), so nothing would contradict them.
- **the check, before you touch any interpolated `LIKE`:** run both forms side by side and diff the verdicts, not the syntax — the query is in `bugfix_142_undeployed_asset_detector/RUNBOOK…md` §R4. Then ask the separate question the accident is hiding: the two vocabularies disagree, and the right long-run fix is to compare against the path the writer RECORDS, not one reconstructed from a purpose.
- **why it is a landmine and not a bug report:** the code is wrong-looking and right, so it invites a fix from someone with no symptom and no suspicion — the exact profile this file exists for. `TestUnderscoreWildcardIsLoadBearing` fails the build on the escape, but only a person reads the reason.
- **source:** 2026-07-31, `bugs_open/142` (`docs024_key_docs_latest/bugfix_142_undeployed_asset_detector/`), council `35d88a60-ec1c-4cd3-b69c-f2813c3e837f`
- **added:** 2026-07-31, bugfix_142 lane

### `site_components.build_status` is `'rendered'` and NEVER `'deployed'` — copying the page_components predicate blinds your query silently

- **footprint:** `site_components.build_status`, `page_components.build_status`, `platform/orchestration/actions/render_site_components_action.go` (`renderAndStoreSiteComponent`), any check reading site chrome (`slot_name` in `head`/`header`/`footer`)
- **fires when:** you write a query against `site_components` next to one against `page_components` — auditing chrome, checking whether a head tag shipped, asking "is this site component live?". `AND build_status = 'deployed'` is right there in the sibling query and reads as the obvious parallel.
- **the tell: none — you get zero rows and a clean pass.** `renderAndStoreSiteComponent` writes `SET rendered_html = $1, build_status = 'rendered'` and nothing ever advances it; all 42 rows fleet-wide (head/header/footer × 14 sites) are `'rendered'`, measured 2026-07-31. So the filter is always false, the check finds nothing, and "nothing found" is indistinguishable from "nothing wrong" — while the page_components version of the same line is correct and necessary.
- **the check:** `SELECT slot_name, build_status, count(*) FROM site_components GROUP BY 1,2;` before filtering on it. If you need "is this chrome live", the answer is the rendered HTML's content, not a status column. And do not "make the two consistent" — the two tables use the column differently, which is the actual defect underneath, and normalising it is a fleet-wide change, not a tidy-up.
- **the second half, easy to miss:** the chrome is where `injectBrandHeadTags` puts the favicon/og:image references, so a query looking for a brand-head asset in `page_components` finds nothing on any site — not because it is undeployed but because it was never there. `bugs_open/142` is that mistake shipped.
- **source:** 2026-07-31, `bugs_open/142` (`docs024_key_docs_latest/bugfix_142_undeployed_asset_detector/` §R3), pinned by `TestSiteComponentsAreNotFilteredOnDeployedStatus`
- **added:** 2026-07-31, bugfix_142 lane

### `storage.DeployedWebPath` — "the single source of truth for the variant filename convention" — is silently WRONG for `og_card`, and right for `favicon` by coincidence

> **STATUS 2026-08-02 (updated, second time) — FIXED AND LIVE on chassis `v1.0.1229`, pod-verified on both replicas (`bugs_closed/168`, IMG-067). THE TRAP BELOW IS HISTORY, not current behaviour.** The entry is kept rather than deleted because two of its clauses survive the fix permanently (last paragraph) and because a reader on an older binary still needs it — check with the command below before trusting either half.
> **Reading the SOURCE at HEAD:** the trap is gone. `storage.DeployedAssetPath` consults `BrandHeadAssetPaths` first and `DeployedWebPath("og_card","og_card")` now returns `/assets/images/og-card.png`. You no longer need `IsBrandHeadPurpose` to ask *"where is this served from?"* — keep it only for the different question *"which table holds the evidence it is deployed?"*, which is unchanged. `check_image_url_404`'s branch was removed as redundant; do not re-add one.
> **Reasoning about a binary older than `v1.0.1229`:** everything below is still exactly true of it. Go changes are inert until then. Confirm which you are in before trusting either: `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "Phase 2E: derived variant deploy path"'` — **non-zero means the OLD binary and the trap below is live.**
> **Still true in BOTH worlds, and the reason this entry is not deleted:** `og-card.png` is not derivable from `og_card`, so `BrandHeadAssetPaths` remains the one declaration — it became the derivation's INPUT, it was not collapsed away. ~~And `deploy_path` still overrides everything and is invisible from `(asset_key, purpose)`.~~
>
> **CORRECTED 2026-08-04 (`bugs_open/179` finding A): `deploy_path` no longer overrides anything.** The override is deleted — `deploy_image_asset` constructs no `AssetPaths` of its own, so every path it commits is `DeployedAssetPath`'s output. An **explicit** `deploy_path` (step config, the deprecated `deploy_path_field`, or `input_data.deploy_path`) draws a refusal before the download and the git commit. A value reachable only by `ExtractActionInputs`' depth-20 recursive search of `collected_data` is deliberately **ignored** rather than refused — that search is how the override used to be armed by a caller who never asked for it, but refusing on it would turn a stray nested key into a false denial of a legitimate deploy. **New trap in its place:** a step that sets `deploy_path` now **completes GREEN** — `deployed:false, skipped:true` plus a `reason`, not an error, because a refusal must let the work item resolve rather than retry for ever. **Read the result's `reason`; the orchestration status says COMPLETED either way.** Tree-wide, `TestAssetPathsAreOnlyConstructedInStorage` now fails the build on any hand-built `AssetPaths` outside `platform/storage`.

- **footprint:** `platform/storage/url_helpers.go` (`DeployedWebPath`, `DeployedAssetPath`, `AssetKeyFilename`, `BuildAssetPaths`, `BrandHeadAssetPaths`), `platform/orchestration/actions/deploy_image_asset_action.go`, `platform/orchestration/actions/derive_brand_head_assets_action.go` (`recordDerivedAsset`), `platform/orchestration/actions/render_site_components_action.go` (`injectBrandHeadTags`)
- **fires when:** you need the web path a generated asset is served from and reach for the helper that documents itself as exactly that. It is the correct answer for every purpose except the two brand-head artefacts.
- **the tell:** a path with an underscore where the served file has a hyphen. `AssetKeyFilename` does the `_`→`-` swap, but `DeployedWebPath` only reaches it when `assetKey != purpose` — and for these two they are equal. Measured 2026-07-31: `DeployedWebPath("og_card","og_card")` and `DeployedWebPath("","og_card")` both return `/assets/images/og_card.png`; `DeployedWebPath("og_card","")` returns `/assets/images/og-card.jpg` (right name, wrong extension). **No argument pair returns the real `/assets/images/og-card.png`.** `favicon` agrees only because it has no underscore to disagree about — so testing your call on favicon proves nothing about og_card.
- **the check:** for `favicon`/`og_card` use `storage.BrandHeadAssetPaths[purpose]`, which carries the literal `recordDerivedAsset` writes after the git commit; `storage.IsBrandHeadPurpose(p)` is the branch. For anything else `DeployedWebPath` is correct. If you are about to compare a path against `rendered_html`, remember the brand-head references live in `site_components`, not `page_components`.
- **why it is a landmine:** the helper's own doc comment asserts it is the single source of truth, so the natural move is to trust it rather than test it — and the one purpose it gets wrong is the one whose artefact is invisible on the page (it only shows in a browser tab and in social previews), so a wrong path does not look broken to anyone reading the site.
- **source:** 2026-07-31, `bugs_open/142`, commit `d671fb2b2`, pinned by `TestDeployedWebPathCannotExpressBrandHeadPaths` — which did exactly what it was written to do: it **fired**, and its failure message named its own remedy, so it was **inverted rather than deleted** on 2026-08-02 (now `TestDeployedWebPathExpressesBrandHeadPaths`). A tripwire that tells the next person what to do when it trips is worth copying.
- **added:** 2026-07-31, bugfix_142 lane
- **updated:** 2026-08-02, bugfix_168 lane — fixed at HEAD, see the status banner above; the entry stays because the running fleet is still on the old binary and the two "still true in both worlds" clauses never went away

### `extractSiteID` cannot see `input_data.site_id`, which is the ONLY place page-content-writer keeps it — so a DB read wired to it fails exactly like the bug it was fixing

- **footprint:** `platform/orchestration/actions/site_db_actions.go` (`extractSiteID`), `platform/orchestration/actions/prepare_link_context_action.go` (`resolveLinkContextSiteID`), `orchestration_states.collected_data`, agent type `page-content-writer`, and any new action added to the writer's workflow
- **fires when:** you add a database read to an action on the page-content-writer path and reach for the package's shared site-id extractor, which is the obviously correct reuse. It is the right helper on five other paths and it resolves **nothing** on this one.
- **the tell: none — you get an empty string, then an empty result set, then a shorter prompt.** `extractSiteID` looks in `site_record.site_id`, top-level `site_id` and `db_sync.site_id`. Measured 2026-07-31 over all 26 `page-content-writer` orchestrations: `input_data.site_id` present on **26**, `site_record` on **0**, top-level `site_id` on **0**, `db_sync` on **0**. So the helper returns `""` on every real run, your guarded DB read is skipped, and the action degrades silently — which is indistinguishable from "this site has nothing", and is precisely the shape of `bugs_open/092` itself.
- **the check, before you rely on any site-id extractor on a path you have not measured:** ask the runs, not the helper — `SELECT count(*) AS runs, count(*) FILTER (WHERE collected_data->'input_data' ? 'site_id') AS input, count(*) FILTER (WHERE collected_data ? 'site_record') AS record, count(*) FILTER (WHERE collected_data ? 'db_sync') AS dbsync FROM orchestration_states WHERE owner_agent_type='<your agent>';` A query shaped "is the field I expected missing?" only confirms your suspicion; this one also tells you **where the identity IS**, which is what you actually need.
- **do NOT fix it by widening `extractSiteID`.** It has five other callers, several of which treat `""` as "skip this work", so teaching it a new location silently changes what five unrelated actions do — a shared-semantics change smuggled inside a bug patch, which is the scope the guardian seat vetoed `bugs_closed/124` for. Resolve the extra locations locally, as `resolveLinkContextSiteID` does (it calls the shared extractor first, then adds `input_data.site_id`).
- **source:** 2026-07-31, `bugs_open/092` (`docs024_key_docs_latest/bugfix_092_writer_link_constraints/`), commit `2e1bfb39e`, council `4b8c5e21-011b-40f0-819a-3dfa4b4c7b1d`, pinned by `TestPrepareLinkContextResolvesSiteIDFromInputData`
- **added:** 2026-07-31, bugfix_092 lane

### `git archive HEAD` into your scratchpad fills a tmpfs SHARED by ~30 sessions — and a full `/tmp` makes commands look failed when they succeeded

- **footprint:** `/tmp/claude-1000/`, any scratchpad `git archive HEAD | tar -x`, `go build`/`go test` run inside such a checkout
- **fires when:** you follow the standing (and correct) practice of verifying a change against committed `HEAD` rather than against the shared working tree, because the tree carries other sessions' in-flight edits. Nothing warns you, and a single checkout looks trivially small.
- **the tell, and it is misleading:** Bash starts returning `the temp filesystem … is full … the child process's stdout/stderr writes failed with ENOSPC`. That is an **output-capture** failure, not necessarily a command failure — measured 2026-07-31, a `cat >> file` heredoc that reported this error had in fact **appended correctly**. So the natural response, re-running it, double-applies the write. Check the file before retrying anything non-idempotent.
- **the numbers:** `/tmp` is a **16G tmpfs**, shared by every concurrent session on the box. One `git archive` checkout plus its build cache is ~220MB; two took this lane to 440MB. Sessions were holding 800MB–1.7GB each, including one finished the previous afternoon still holding 1.7GB. It reached 100% mid-session.
- **the check:** `df -h /tmp` before extracting, and `rm -rf <checkout>` in the same breath as the test run that needed it — not "later". `du -sh /tmp/claude-1000/*/*/ | sort -rh | head` shows who is holding what; map a directory to its session's liveness with `stat -c %y ~/.claude/projects/<project>/<session-id>.jsonl` before concluding it is abandoned, and do not delete another session's scratchpad on your own authority.
- **source:** 2026-07-31, bugfix_092 lane (`docs024_key_docs_latest/bugfix_092_writer_link_constraints/RUNBOOK…md` § Testing against HEAD)
- **added:** 2026-07-31, bugfix_092 lane

### The `assets` table records NO served path — `purpose` is a role name, `url` is an expiring S3 link, and `filename`/`storage_path` are mostly empty

- **footprint:** table `assets` (columns `purpose`, `url`, `filename`, `storage_path`, `asset_key`), `platform/orchestration/actions/discovery_checks/` (any check asking "does this rendered path resolve?"), `platform/storage/url_helpers.go` (`DeployedWebPath`)
- **fires when:** you need to know whether a `/assets/images/…` path a page renders will actually serve, and you look for the answer in the obvious column. Every candidate is wrong in a different way, and each is wrong *quietly*.
- **the tell: none — every column returns a plausible-looking string.** Measured 2026-07-31 over all 267 active asset rows: `url` is a presigned `https://s3.…` link on **152** (it expires, and it is not where the site serves the file from); of the 115 whose `url` does start `/assets/images/`, **47 are the unresolved template literal** `/assets/images/input-data.asset-key.jpg`; `filename` is empty on **191** and `storage_path` on **189**. And `purpose` is a ROLE — `hero`, `icon`, `logo` — never a path: comparing it to a filename is what made `image_url_404` blind to 83% of its own surface for months (`bugs_closed/128`), because owning one hero asset anywhere made every rendered `hero*` path unreportable.
- **the check:** the served path is **derived, not stored** — `storage.DeployedWebPath(asset_key, purpose)`, which is what `deploy_image_asset` commits to (via the shared `storage.AssetKeyFilename`) and what all five writers render. Branch brand-head purposes through `storage.BrandHeadAssetPaths` first: see the separate `DeployedWebPath` entry above, which is the trap inside this one. If you want ground truth rather than intent, only HTTP has it — and `verifier_coverage_test.go:171` records the standing objection to putting an outbound probe on the completion path, so derive it instead.
- **what this does NOT tell you:** whether the file was actually committed and deployed. A row can derive a perfectly correct path for a file that was never pushed (`gaswholesalers.com` `/assets/images/logo.png`, 404 with a healthy `logo` asset row, live 2026-07-31). That question belongs to `check_undeployed_assets`, and conflating the two is how one check ends up asserting the other's finding.
- **why it is a landmine:** the natural query — `SELECT url FROM assets WHERE purpose='hero'` — returns a real URL that really works when you paste it in a browser, so the column looks authoritative right up until you compare it with what the page renders. Nothing errors; the answer is simply about a different file in a different place.
- **source:** 2026-07-31, `bugs_closed/128` (`docs024_key_docs_latest/bugfix_128_image_url_404/`), commit `beff42809`, council `99dca96a-413a-4bcb-b278-9577f920786d`, pinned by `TestImageURL404_SilentOnEveryPathAnAssetActuallyDeploysTo`
- **added:** 2026-07-31, bugfix_128 lane

### Two discovery checks already own "a page renders the fallback image path" — extending one silently competes with the other

- **footprint:** `platform/orchestration/actions/discovery_checks/check_image_url_404.go`, `platform/orchestration/actions/discovery_checks/check_placeholder_image_in_use.go`, item types `needs_hero_image` / `needs_logo`, agent `design-discovery-agent`
- **fires when:** you add or widen image-reference detection and reach for whichever of the two files you found first. They overlap by construction and neither file mentioned the other until 2026-07-31.
- **the tell:** two work items for one repair, under two different `item_key` prefixes — so `idx_swi_dedup` cannot collapse them and the queue reads as two problems. Until this fix, `image_url_404`'s recognised-purpose branch and `check_placeholder_image_in_use` shared the same two paths (`/assets/images/hero.jpg`, `/assets/images/logo.png`), the same purposes, the same item types, the same `image-build-handler`, the same severity and the same precondition, and **both are enabled on the same agent**. Neither had ever fired, which is why the duplication was invisible: `SELECT count(*) FROM site_work_items WHERE item_key LIKE 'placeholder_image_in_use:%'` returned **0**, and 0 of the 13 `image_url_404:%` rows carried either routed type.
- **the check:** before adding image detection, `grep -l "assets/images" platform/orchestration/actions/discovery_checks/*.go` and read the headers — there are now four checks in this space with deliberately different questions: `image_url_404` (a rendered path no active asset deploys to — flag-only), `placeholder_image_in_use` (the documented fallback path with no asset of that purpose — routes a regeneration), `undeployed_asset` (the asset row exists, the file was never pushed), `image_source_unsatisfiable` (a component's `input_schema` asks for an image nothing can supply). If your new predicate does not differ from all four, you are widening one of them, not adding a fifth.
- **why it is a landmine:** a check that has never fired looks like dead code to a reader and like available space to an author. It is neither — it is a live predicate whose precondition is rare, and duplicating it costs a second unclearable work item on the first site where the condition finally holds.
- **ONE SHAPE IS NOW SETTLED, so the next author does not re-derive it (2026-08-03).** A third kind, `bare_token_src` (`<img src="cpu">` — an icon name rendered into an image slot), was added to `image_url_404`. **It cannot collide with `placeholder_image_in_use`, and that is a property of the pattern rather than a judgement:** the sibling matches two LITERAL paths, `/assets/images/hero.jpg` and `/assets/images/logo.png`, and the new predicate's character class is `[^"/.:#\s]+` — it excludes `/` and `.`, which both of those paths contain. **No input fires both.** The four-way split this entry's "the check" line prescribes still holds and is worth restating with the new kind in place: `image_url_404` now owns three shapes (a path no asset deploys to · empty/`#` src · a bare word), all flag-only; `placeholder_image_in_use` owns the two documented fallback paths and ROUTES a regeneration; `undeployed_asset` and `image_source_unsatisfiable` do not read rendered HTML at all. Council `cfc94d91-3d17-4f29-a370-2b91d1a59a6f`: round 1 REVISE, gated high on exactly this overlap not having been checked; round 2 **APPROVED** once it had been.
- **and note HOW that landmine was missed**, because it generalises: the author grepped this file by the path and symbol they were EDITING, and this entry's discriminating text is about the OTHER file. **An overlap landmine is keyed to a PAIR, and is often findable only by the half you are not touching** — so when you are widening any detector, grep the whole family's footprint, not just your own file.
- **source:** 2026-07-31, `bugs_closed/128`, commit `beff42809`
- **added:** 2026-07-31, bugfix_128 lane
- **updated:** 2026-08-03, finetuning.uk repair lane — the `bare_token_src` shape resolved against this entry, at the council's request that the analysis live somewhere agents can read rather than only in a Go comment

### An API USAGE-LIMIT death looks exactly like a transient seat fault, and the runbook's advice for the transient case — "resubmit unchanged" — is actively wrong for it

- **footprint:** `orchestration_states.current_step = 'complete_invalid'`, `collected_data->'__step_error'`, `execute_llm_prompt`, any council-gate round (`097_TRIGGER_council_review_v1.sh`), any diagnosis-loop run (`090`), any agent whose workflow contains an LLM step
- **fires when:** a council round or diagnosis run ends `COMPLETED / complete_invalid` after several seats have already run. `RUNBOOK_council_gate.md` teaches — correctly, for its case — that `complete_invalid` with reviewers already fired means a **transient** seat fault (a TCP reset), and that you should **resubmit unchanged under `RESUBMIT_CORR`** rather than edit your JSON. There is a second cause with the same signature and the opposite remedy.
- **the tell, and it is only in the message body:** `API request failed with status 400: {"type":"invalid_request_error","message":"You have reached your specified API usage limits. You will regain access on 2026-08-01 at 00:00 UTC."}`. Note **400 `invalid_request_error`**, not 429 and not a connection error — so it reads like a malformed request, and the step name (`review_tooling_provenance`, `review_editquality` — whichever seat drew the short straw) reads like that seat's problem. It is neither. It is an account-level cap with a stated reset time, and **every** subsequent LLM call fails identically until then.
- **why the wrong remedy is expensive:** resubmitting burns a round against a wall — each retry re-runs the seats that already succeeded, spends nothing useful, and produces another `complete_invalid` that looks like a new fault. Editing the JSON is worse: you will "fix" a submission that was never broken, and the next round's diff then carries changes nobody asked for.
- **the check, before you resubmit ANY `complete_invalid`:**
  `SELECT collected_data->'__step_error'->>'failed_step', left(collected_data->'__step_error'->>'message',300) FROM orchestration_states WHERE correlation_id='<run corr>'::uuid;`
  Then distinguish three cases by what the message says, NOT by the step name: `persist_submission` + no reviewers fired ⇒ your JSON really is invalid (schema); "connection reset"/"endpoint unavailable" on one seat ⇒ transient, resubmit unchanged; **"usage limits"/"regain access at <time>" ⇒ STOP and wait for the stated time.**
- **and measure the blast radius with the RIGHT NEEDLE:** the phrase is **`usage limit`**, not `spending limit`. Grepping `ILIKE '%spending limit%'` returns **0 rows** during a live outage and reads as "it's just me" — the mistake was made and caught within one query on 2026-07-31. `ILIKE '%usage limit%'` over `orchestration_states` in the last few hours tells you whether the fleet is down or your run is unlucky.
- **source:** 2026-07-31 18:58:41 UTC, first hit fleet-wide, killing a `bugs_open/149` A6 council round at `review_tooling_provenance` and another session's round at `review_editquality` a minute later; stated reset 2026-08-01 00:00 UTC. Distinct from — and a counter-example to — the transient-fault guidance at `RUNBOOK_council_gate.md`'s `complete_invalid` note, which should be read together with this entry
- **added:** 2026-07-31, bugfix_149_nav_membership lane

### A completeness floor on `page_components` must exclude LOCKED slots from its target, or it refuses perfect rebuilds

- **footprint:** `platform/orchestration/actions/save_sections_prune_floor.go`, `pages.sections`, `page_components.locked_at`

A guard that asks "did this rebuild produce enough?" needs a target to compare
against, and `pages.sections` (the planned composition) is the obvious one — it is
written by seven other actions, so it is a genuine second opinion rather than an
echo. **It is also wrong on its own.** An actively-locked slot is not deleted and
the incoming section that matches it is discarded, so a locked slot can never be
part of what the save writes. Count it in the target and a *flawless* rebuild
scores below the floor and is refused — on exactly the pages a human cared enough
about to lock. `idea.uk/index.html` plans 6 sections and holds 4 locks, so a
perfect rebuild writes 2 and scores 2/6 = 33%.

**the check:** before trusting any per-page count that carries the agent-writable
predicate, remember it answers "rows this save may TOUCH", not "rows the page
HAS". Read the rows, not the count:

```sql
SELECT pc.slot_name, pc.build_status, pc.locked_at IS NOT NULL AS locked
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='<domain>' AND p.url='<url>' ORDER BY pc.position;
```

Then build the target as `planned − suppressed − locked`, and simulate it across
every page the code can actually REACH — split by `rebuild_policy`, because
`owned` pages are refused ~370 lines earlier and counting them inflates the
false-positive rate with pages that never arrive. Measured: raw planned → 3 trips;
locked excluded → **0 of 238 reachable pages**. Applies to `site_nav_items` too,
which carries the same lock columns.

**and do not identify a page by `url` alone** — `/index.html` exists once per
site, so `WHERE url='/index.html'` returns 14 unrelated pages and reads as
duplicate rows. Join `sites`, or key on `pages.id`.

---

### idea.uk is NOT behind Cloudflare — its own runbook says it is, and builds a security plan on that

- **footprint:** `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/RUNBOOK_idea_uk_vm_site.md`, `idea.uk`, `116.203.204.115`, `set_real_ip_from`, `CF-Connecting-IP`

`RUNBOOK_idea_uk_vm_site.md:12` states `FRONT nginx + Let's Encrypt, DNS
(Cloudflare) → the VM`, and its §4a opens *"idea.uk is behind Cloudflare"* under
the heading **"Restore the real client IP — nothing else works until this is
done"**. Measured 2026-07-31: **it is not, and never appears in the request
path.** `dig +short NS idea.uk` returns three **Hetzner** nameservers, the A
record is the bare VM (`116.203.204.115`), and the live response header is
`server: nginx/1.28.3 (Ubuntu)` with **no `cf-ray`**. `relojistas.com`, the other
VM site, **is** proxied — so the two boxes do not share a front-end shape even
though the estate docs treat them as one pattern.

This misleads in **both** directions, which is why it is a landmine and not just a
stale line. Believe the doc and you will (a) do §4a's real-IP work, which is
unnecessary — nginx already sees true client addresses, so `limit_req` on
`$binary_remote_addr` is *already correct*; (b) assume a **WAF, Turnstile and DDoS
layer sit in front of a live money-taking box**, and they do not; (c) follow
`RUNBOOK:435`'s *"purge the Cloudflare cache for idea.uk"*, a **no-op** — debugging
a stale page you would purge a cache that isn't there and conclude the purge
failed. Disbelieve it and copy `setup.sh` to a box you *do* proxy, and you inherit
a limiter that buckets every visitor into one global counter.

**the check:** never infer the front end from `dig` — an A record cannot tell you
whether a proxy terminates the request, and Cloudflare NS with a non-Cloudflare A
record is **grey/DNS-only**: delegated but unproxied, with no WAF and a public
origin IP. Ask the response:

```bash
curl -sS -o /dev/null -D- -m 12 https://<domain> 2>&1 | grep -iE '^server|cf-ray'
# cf-ray present            => proxied, Cloudflare is in the path
# server: nginx/... , no cf-ray => you are talking to the origin directly
```

Run it **per domain**. Sharing an owner, a provider or a runbook does not mean
sharing an ingress: measured the same minute, `idea.uk` had none and
`relojistas.com` was proxied.

---

### The domain→B2 object-key mapping is in a Cloudflare Worker that is NOT in this repo — grep finds nothing and it looks like config

- **footprint:** `sites.domain`, `sites.github_repo`, B2 static deploy path, `webdesign.co.uk`

Ask "how does a domain become a file in the bucket?" and every instinct points at
the repo or the DB: `sites.deploy_config`, `resolveGitRepoNameDB`, the GitHub
Action. **The answer is in none of them.** A Cloudflare Worker fronts B2 and
derives the object key from the request **host, verbatim** — and its source is not
in this repository (`grep -rn 'objectKey\|B2 returned error'` → no match). So the
search that should find it returns nothing, and the honest conclusion from an
empty grep — "this must be configured somewhere else, probably data" — is wrong in
a way that costs an afternoon. PLAN §6(i) of the webdesign.uk lane carried
`[UNVERIFIED — I have not read how a domain maps to a bucket prefix]` for three
days on exactly this basis.

**the check:** ask the live 404, which states the key it looked for. One second,
no credentials, no source:

```bash
curl -sS https://<any-domain-on-the-account>
# {"error":"B2 returned error","objectKey":"<host>/index.html","status":404, ...}
```

**Two consequences worth knowing before you touch the static path.** (1) The
mapping is host-derived, not allow-listed — proven by `webdesign.uk`, which has
**no row in `sites`** at all and still gets a key built for it (one sample; the
source is unread, so treat as strong inference, not fact). Any hostname you point
at this path becomes a served prefix with no per-site configuration, which is
either a free wildcard-hosting mechanism or an accident waiting, depending on what
you point at it. (2) That JSON is served to anyone visiting **any unpopulated
domain on the account**, naming the origin technology and the key convention.
Bucket keys are not credentials, so it is a disclosure rather than a hole — but it
is free to close with a holding-page object or a redirect rule, and worth doing
before a customer-facing domain is the one printing it.

### A mutant that BREAKS THE BUILD prints the same `FAIL` as a mutant that was caught

- **footprint:** mutation testing / "prove this test can fail" against any package
  other sessions also edit — `platform/orchestration/actions/`, `internal/adapters/…`,
  anything with more than one thread in it; any loop that edits a file, runs
  `go test`, restores, and reads the result
- **fires when:** the package stops compiling **for a reason that is not your
  mutation** — most often another session saving a half-finished edit to a
  different file in the same package, which on this tree happens *between* two
  iterations of the same loop
- **the tell:** none in the place you are looking. A caught mutant prints
  `FAIL <pkg> 0.05s`. A build break prints `FAIL <pkg> [build failed]`, and the
  `--- FAIL: TestX` lines you were grepping for are simply **absent** — so a
  `grep -E '^(--- FAIL|FAIL)'` summary shows `FAIL` for both and the mutant reads
  as caught. Measured 2026-07-31: four of six mutants in one run were invalidated
  this way by another session's edit to `save_sections_prune_floor.go`, while my
  own file was untouched and correct
- **the check:** two things, and the second is the one that actually saves you.
  (1) Make the harness **build before it tests** and say so out loud —
  `go build ./pkg/ || echo "!! DID NOT COMPILE — this mutant proves nothing"` —
  because a mutant must compile before its result means anything.
  (2) **Run the whole proof against a clean export, not the working tree:**
  ```bash
  git archive HEAD | tar -x -C "$SCRATCH/headtree"
  cp <your changed files> "$SCRATCH/headtree/<same paths>"
  cd "$SCRATCH/headtree" && go test ./<pkg>/ -count=1   # then mutate HERE
  ```
  An untracked or uncommitted file belonging to another session cannot follow you
  into `git archive HEAD`, so the only uncommitted code in the run is your own —
  which is the property a mutation proof needs and the shared tree cannot give you.
  This also re-verifies your change against what HEAD will actually build.
- **source:** `brochure_component_library/NOTES_…` 2026-07-31 evening, proving the
  TL-035 caller half. Third member of the family with *"a mutation that never
  happened is indistinguishable from a guard that works"* (mutation did not apply)
  and *"a passing mutation test may mean a SECOND guard absorbed the mutation"*
  (mutation applied, absorbed) — this one is **mutation applied, result unreadable**
- **added:** 2026-07-31, brochure lane

### `modelNumberRe` only sees tokens whose FIRST hyphen segment mixes letters and digits — and it matches SUB-tokens, so your counterexample may never reach the check you are testing

- **footprint:** `platform/orchestration/actions/verify_report_prose_action.go`, `modelNumberRe`, `skuTokenTraces`, `verifyReportProse`, the `report-builder` agent
- **fires when:** you change, test or reason about the report prose gate's SKU check, and pick example tokens from the surrounding material — a real candidate name, a bug report's illustration — assuming the check sees them
- **the tell:** none at all, and that is the trap: an example the regex never matches produces **no violation**, which reads exactly like "the gate correctly allowed it". `bugs_open/160` proposed `EGP-50-X` (against the indexed `Schunk EGP 40-N-S-B`) as the fabricated sibling a careless fix would let through. It is not matched by `modelNumberRe` at all — `EGP` carries no digit, and the regex needs the letter-digit adjacency in the segment **before the first hyphen**. The same is true of the real candidate `Festo EHPS-20-A-LK`. A test built on it passes before *and* after any change, proving nothing. The mirror image: for a paraphrase of `ISO 9409-1-50-4-M6` the regex does **not** match the whole token, it extracts the sub-token **`M6-compatible`** — `\b` starts a fresh match after each hyphen — so the string in the violation message is not the string in the prose
- **the check:** before using a token as evidence about this gate, confirm the classifier actually returns it: `modelNumberRe.FindAllString(text, -1)`. Every counterexample must begin with a mixed letter-digit segment (`2F-…`, `IP54-…`, `GEP5010IO-…`). Then assert on the **violation text** (`model-like token` plus the token), never on `len(violations)` — the numeric gate at `:307` rejects many of the same strings for a different reason and will absorb your mutation
- **why it matters:** the gate is fail-closed — a false violation destroys the whole report and 404s its URL — so both directions of a wrong test are expensive: a fix that looks proven and is not, or a guard quietly disarmed by a fix "verified" on examples it never judged
- **source:** `bugs_open/160`, fixing lane 2026-07-31; both corrections are recorded in the bug file
- **added:** 2026-07-31, bugfix_160_prose_gate_recombination lane

---

### `browser-runner-adapter`'s image has NO `strings` binary — the standard pod-grep returns 0 for everything and reads as "your change did not ship"
- **footprint:** `internal/adapters/browserrunner`, `browser-runner-adapter` (pod / deployment)
- **the check:** CLAUDE.md's deploy-verification recipe is
  `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<sym>"'`.
  On this service the pipeline fails silently — no `strings` in the image — and
  prints `0`, which is indistinguishable from "the symbol is absent". Use
  `grep -ac '<sym>' /app/browser-runner-adapter` instead; `-a` treats the binary as
  text and needs nothing installed. **Always pair it with a positive control in the
  SAME exec** (e.g. `no_horizontal_overflow`, present in every build): if the control
  is also 0, the check is broken and you must fix the CHECK, not rebuild the image.
  This was found exactly that way, by a control firing on the guard's first run —
  without it the reading was "roll again", costing a needless build and roll.
  Worked example: `docs024_key_docs_latest/loancalculator_couk/acceptance/criteria/INSTALL_GATE.sh`.

### A schema value containing a QUOTE, rendered into a `<script>`, kills the whole tool — and it still passes every structural check
- **footprint:** `content_components.html_template`, `input_schema`, `RenderTemplateReportingMissing`
- **the check:** `text/template` escapes NOTHING, by design. A fallback like
  `in "58-day" interest charges.` interpolated into `var S = "{{.field}}";` renders as
  `var S = " in "58-day" interest charges.";` — a syntax error that kills every
  statement in the block. The calculator then shows `£0.00` for every input while
  **still shipping a `<script>` tag (so `tool_health` passes), still matching every
  selector (so Tier 2 passes) and still rendering normally.** Only a check that reads
  the computed VALUES catches it. There is no context-aware fix available, so the rule
  is structural: **put copy in the MARKUP and let the script write only the number** —
  a hidden `<div>` of `data-copy` spans read with `textContent` costs nothing and has
  no quoting hazard at all. Grep before you install:
  `grep -o '"{{\.[a-z_]*}}"' <template>` inside any `<script>` region is the shape to
  refuse. `loancalculator_couk/rewrite/render_tool.go` enforces it offline.

### `display:flex` on a control's parent BLOCKIFIES the control — and computed `display` is part of a tool's equivalence fingerprint
- **footprint:** `docs024_key_docs_latest/loancalculator_couk/toolgolden.py`, `content_components.html_template`
- **the check:** a flex (or grid) container forces every child's computed `display` to
  the blockified form, so an `<input type=checkbox>` that computed `inline-block`
  computes `block` the moment you modernise its row to flex. Nothing looks different
  and nothing computes differently, but `toolgolden`/`computed_values` record display
  because a tool that responds by REVEALING a region changes nothing else. Two tools on
  this one site needed **opposite** layouts (`damage-checker` non-flex,
  `application-tracker` flex) because their original stylesheets disagreed — unknowable
  from reading either page, they look identical. So: never harmonise the layout of two
  ported tools without re-baselining both, and expect the failure to point at the row,
  not at the control.

### Two local copies of a site can both look right — `~/projects/sites2/<domain>` is NOT what is served
- **footprint:** `~/projects/sites`, `~/projects/sites2`
- **the check:** both trees hold a full, plausible copy of the same domain. For
  loancalculator.co.uk, `~/projects/sites/` matches the served bytes exactly and
  `~/projects/sites2/` differs on **every** page checked. Nothing in either directory
  says which is authoritative, and a port built from the wrong one produces components
  faithful to the wrong original that pass their own review. Cost is one command:
  `diff <(curl -s https://<domain>/<path>) ~/projects/sites/<domain>/<path>` — or md5
  both — **before** reading a single file as source of truth.

### A `/section/` URL 404s on every B2-hosted site, and a local server hides it
- **footprint:** `~/projects/sites`, `scripts/cloudflare/worker.js`, `python3 -m http.server`
- **the check:** the worker maps `{hostname}{path}` straight to a B2 object key and
  rewrites **only** `/` → `/index.html` (`worker.js:8-11,27`). An object store has no
  directory index, so `href="/loans/"` asks B2 for an object literally named `loans/`
  and returns **404**. Confirmed live on every site in the fleet:
  `loancalculator.co.uk/tools/`, `mortgagecalculator.co.uk/guides/`,
  `webdesign.co.uk/tools/`, `loanandmortgagecalculator.co.uk/loans/`.
  **`python3 -m http.server` DOES resolve directory indexes, so it is a MORE FORGIVING
  server than production and will pass a site that is broken live.** Never verify a link
  graph, a sitemap or a canonical against it. Resolve every reference against the live
  origin:
  `curl -sS -o /dev/null -w "%{http_code} $u\n" "https://<domain>$u"`.
  This shipped on loanandmortgagecalculator.co.uk: 3 dead references × 42 pages, 3
  sitemap entries, and **3 canonicals naming a URL that did not exist** — the worst
  form, because a canonical in a sitemap tells Google the authoritative URL is a 404.
  Related and separate: `bugs_open/116` (the same root property) and `bugs_open/132`
  (the JSON error blob served instead of `404.html`).
  **And beware the second-order trap that actually did the damage:** when a link checker
  flags `/loans/`, the tempting "fix" is to teach the checker to resolve it to
  `loans/index.html`. That converts a TRUE POSITIVE into permanent silence. If a checker
  disagrees with production, change the checker to match production — never the reverse.

### After a locked adoption, `page_components` and the deploy repo are TWO writers for the same file
- **footprint:** `page_components`, `pages.rebuild_policy`, `~/projects/sites`, `deploy_page`
- **the check:** a `page_rerender` on an `owned`/`verbatim` page renders the stored
  `rendered_html` and **commits it to the shared sites repo** (`deploy_result.repo_url =
  github.com/gqls/sites`, message `Rerender: <file>`), which then deploys to B2. So once
  a repo-built site is adopted, the same bytes have two independent writers: your builder
  → your commit → Actions, and `page_components` → rerender → git-adapter commit →
  Actions. They agree only while something keeps them in step. **So: after ANY change to
  a builder that emits an adopted page, re-run the byte gate with `--repair`** — otherwise
  the DB still holds the old bytes and the next rerender silently reverts your change.
  Gate: `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/gate_component_bytes.py`.
  Do **not** trust `content_data.sha256` as the fidelity check — it is computed over the
  stored bytes, so it is self-consistent by construction and says nothing about whether
  they match the origin.

### The `site_specs` aspect named `audience` is read by NOTHING — and it is the most populated one
- **footprint:** `site_specs`, `site_specs.audience`, `content_direction`, `apply_gap_plan_action.go`
- **the check:** you are asked to steer what a site writes, so you set the aspect named
  for it. `audience` is present on **29 of 33 sites** — the most widely-populated aspect in
  the database — and **no agent, prompt or Go path reads it.** The row appears, the edit
  looks applied, and nothing changes. Same for `editorial`, `voice`, `voice_and_tone`,
  `content_standards`, `terminology_and_positioning`, `voice_and_audience` — all created ad
  hoc by the gap-planner, which takes `aspect` straight from an LLM plan with no allow-list
  (`apply_gap_plan_action.go:744`). Before writing any aspect, prove something reads it:
  `SELECT type FROM agent_definitions WHERE deleted_at IS NULL
   AND default_config::text LIKE '%site_specs.specs.<aspect>%';`
  The three seams that DO reach `page-content-writer`'s prompt are
  `identity.target_audience`, `identity.key_differentiators`, and
  **`content_direction.formatted` — the ONLY field of `content_direction` the writer
  reads.** So new steering goes as a KEY INSIDE `content_direction`, never as a new aspect.
- **and the second half, which is worse:** `formatted` is generated by
  `datahelpers.FormatContentDirection` (via `write_site_spec`). **A hand-written
  `content_direction` that does not regenerate `formatted` is invisible to the writer.**
  If you regenerate it yourself, GATE the reimplementation: reproduce the CURRENT spec and
  require a match as a multiset of **lines**, not as a string — Go map iteration is random,
  so section and sub-key order in the stored value is arbitrary and carries no meaning.
  Worked example: `loanandmortgagecalculator_couk/set_divergence_specs.py`.

### A TRUNCATED LLM call has `output_tokens = NULL`, so any truncation count or token statistic written against `output_tokens` silently excludes exactly the calls you care about
- **footprint:** `llm_call_log` (`output_tokens`, `error_message`, `success`), any `output_tokens >= max_tokens` predicate, any p95/max/mean of `output_tokens`, `104_REPORT_seat_token_pressure_v1.sh`, `scheduled_tasks.council-seat-token-pressure`
- **fires when:** you measure token usage, headroom, or truncation rate from the LLM call log — cost analysis, cap right-sizing, "is this agent about to get cut off", "how often does this truncate"
- **the tell:** none, and it is doubly blind. A call cut off at the cap is written `success=false`, **`output_tokens` NULL**, with the cut stated only in `error_message` (`response truncated: stop_reason=max_tokens (output_tokens=8000 reached the configured cap)`, sometimes prefixed `TOLERATED (step continued on the partial)`). So (1) the natural truncation count `count(*) FILTER (WHERE output_tokens >= max_tokens)` **can never match one** — measured 2026-07-31, it returned **4** where the true count was **94**, and the 4 were successful calls that happened to land exactly on the cap; and (2) the same `output_tokens IS NOT NULL` filter that every aggregate carries implicitly **drops those rows from your p95/mean/max**, removing the most extreme calls from precisely the agents that truncate most. Six council seats read "0 truncations" while having truncated, one of them 51 times
- **the check:** count truncations from the message, never the number — `count(*) FILTER (WHERE error_message ILIKE '%stop_reason=max_tokens%')`. For headroom, score a truncated call as 1.0 of its cap rather than excluding it: `COALESCE(output_tokens::numeric/max_tokens, CASE WHEN error_message ILIKE '%stop_reason=max_tokens%' THEN 1.0 END)`. And note `success=false` covers more than truncation (API 400s, context-cancelled, dial timeouts), so `NOT success` is **not** a truncation filter either
- **why it is a landmine and not a note:** the wrong query returns a plausible small number rather than an error or an empty set, and a low truncation count is exactly the answer that stops you looking further. It also let a correct conclusion rest on a false measurement: `bugs_open/138`'s instrument argued "truncations read ~0 because the caps were raised" when they read ~0 because the query was blind
- **source:** `bugs_open/138` FIX-058, found and fixed 2026-07-31 — one day after shipping the instrument with the defect
- **added:** 2026-07-31, bugfix_138 lane

### CORRECTION 2026-07-31 (evening) to "Editing the ACTIVE chrome component changes nothing…" — point 2 is now half-superseded, and point 3 has a SECOND mechanism nobody had seen

- **footprint:** same as the entry above (`site_components`, `content_components`, `render_site_components_action.go`, `deactivated_component`), plus **`site_components.build_status`** and `renderAndStoreSiteComponent`'s `!force` idempotence exit.
- **Point 2 ("the fix to the selection is NOT a fix to your site") — superseded for header/footer, permanent for head.** The standing fleet was repointed by hand on the owner's ruling: **21 assignments moved, 28/28 header+footer slots now render from an ACTIVE component**. And `bugs_open/166`'s fix (committed `39afbf697`, council `e242e9d3`, inert until the next roll) makes `render_site_components` repoint an ineligible assignment automatically. **The `head` half stays true indefinitely** — 13 slots still point at deactivated components because NO active `head` component exists fleet-wide, and the repointer correctly declines rather than churn them.
- **Point 3 had a second mechanism, and this is the one that will waste your afternoon.** It is not enough that the repair re-renders the wrong component. **Correct the assignment yourself and the slot is STILL skipped**, because the `!force` exit tests `rendered_html IS NOT NULL AND != ''` — whether the slot holds bytes, never whether it holds the RIGHT component's bytes. A repointed slot holds the OLD component's HTML, so it reads as already-rendered and the render returns `true` having done nothing.
- **the tell: `rerender-pages` reports COMPLETED and the row does not move.** `status=COMPLETED, current_step=complete` with `site_components.updated_at` unchanged. There is no error anywhere. I hit this twice in ten minutes for two different reasons and both looked identical.
- **the check, before you conclude a chrome repair "did not dispatch":** read the ROW, not the run. `SELECT slot_name, build_status, length(rendered_html), updated_at FROM site_components WHERE site_id=…;` — an unchanged `updated_at` after a COMPLETED run means it was SKIPPED, not lost. And on a chassis older than `39afbf697` the way past it is `force_rerender: true` in the step config; `input_data` does not carry it.
- **and the OTHER gate on the same path:** `rerender-pages` renders chrome only when **`refresh_site_components: true`** is in `input_data` — there is a `check_refresh_components` conditional in front of the step. The detector sets it; a hand-run does not, and gets a silent skip with a COMPLETED status.
- **do NOT clear `rendered_html` to force it.** That is what I did, and I had no copy — the artefact regenerates from the template so it recovered, but a failed render would have left the slot permanently empty. `build_status='pending'` is the supported signal (that is why the fix uses it).
- **source:** 2026-07-31, `bugs_open/166` + `bugs_closed/118` (`docs024_key_docs_latest/bugfix_118_chrome_selection/`), concept register **CLC-013**, `WRONG_CALLS.md` same date
- **added:** 2026-07-31, bugfix_118_chrome_selection lane

### `strings` output is NOT one-constant-per-line — the Go linker packs unrelated literals into one blob, so `grep -x` / `^` / `$` return 0 for a string the binary DOES contain

- **footprint:** `strings /app/agent-chassis`, any `kubectl exec … strings … | grep` deploy proof, `grep -x`, `grep -cx`, any anchored (`^`/`$`) pod-grep pattern, `bugs_open/153`'s pod-grep discipline
- **fires when:** you follow the standing rule (grep a literal your change ADDED, with a positive control in the same exec) and you anchor the pattern — which is the natural thing to do when the same literal exists in BOTH the old and new builds and you are trying to distinguish *where* it sits, e.g. "is this div still inline in the template, or is it now its own constant?"
- **the tell:** **both** the old-form and new-form patterns return **0** while a substring positive control returns non-zero. That result is self-contradictory — the code must be one of the two — and the contradiction is the only warning you get. Measured 2026-07-31 on `v1.0.1223`: `grep -cx '<div class="footer-contact"><h4>Contact</h4><p>%s</p></div>'` → **0**, and `grep -cE '^[[:space:]]+<div class="footer-contact">…'` → **0**, on a binary where the literal is plainly present — it sits mid-line between `check adoption_plan config path` and `NormalizeSectionNames: Converted section name…`
- **the check:** pod-grep with **plain substring** matching (`grep -c`), never `-x` and never `^`/`$`. `strings` emits maximal runs of printable bytes, so a "line" is an artefact of where the blob happens to contain a non-printable byte — it has no relationship to your literal's boundaries. When the question is genuinely *where* a constant lives, print structure instead of counting: `strings /app/<binary> | grep -A 6 "<a neighbouring literal>"`. That is what settled the case here — the multi-line footer template rendered as its own lines and showed `%s` on the line the contact div used to occupy, which no count could have shown
- **why it is a landmine and not a note:** a 0 from an anchored grep is indistinguishable from the `bugs_open/153` failure it is *designed* to detect (tag bumped, image never rebuilt), so the wrong reading is "my fix did not ship" — and the remedy for that is to rebuild and roll, which wastes a cycle and still returns 0
- **sibling entries, distinct mechanisms:** non-ASCII bytes splitting a literal (em dashes, 2026-07-30) and a positive control invalidated by your own diff (2026-07-31). Those make a *present* string unmatchable by content; this one makes it unmatchable by *position*
- **source:** pod-verifying `d4731109d` while closing `bugs_closed/111`, 2026-07-31; `WRONG_CALLS.md` same date
- **added:** 2026-07-31, bugfix_111 lane

### The `{{if or .email .phone}}` wrapper in `footer-theme-chrome` is a filed bug fix, not formatting — drop it in a redesign and 12 sites silently regain a "Contact" heading over empty space

- **footprint:** `content_components.html_template` where `name='footer-theme-chrome'` (and `footer-4-column`), the `footer-contact` div, `RenderFallbackFooter` in `platform/orchestration/actions/component_library.go`, `bugs_closed/111`
- **fires when:** you edit, regenerate, restyle or LLM-rewrite the footer chrome component — the component serving **12 of 14** chrome rows fleet-wide since `bugs_closed/118`'s repoint. The gate looks like ordinary conditional markup, sits three lines below two *other* gated columns that make it look like a house pattern rather than a fix, and nothing in the template says a bug was filed about it
- **the tell: there is none at edit time, and none at render time on most sites.** The gate is only load-bearing where a site has neither `sites.email` nor `sites.phone` — today `gamesdesign.co.uk` (no data) and `relojistas.com` (owner ruling: no contact route). Every other site renders identically with or without the gate, so a rewrite that drops it passes every eyeball check you would think to do, and the defect appears on exactly the two sites you were not looking at
- **the check, after ANY footer-chrome edit:** confirm the wrapper survived — `SELECT html_template ~ '\{\{ *if +or +\.email' FROM content_components WHERE name='footer-theme-chrome';` — and then prove it at the bytes on a site with no contact data: `curl -s https://gamesdesign.co.uk/ | grep -c 'class="footer-contact"'` must be **0** (its 4 CSS-rule hits for `footer-contact` are in the `<style>` block and do not use `class=`). **Do not grep for `mailto:`** — Cloudflare rewrites every one into `/cdn-cgi/l/email-protection#<hex>`, so it returns 0 regardless of the truth, which is what hid this bug for a day
- **the gate is `or .email .phone`, deliberately, not `.email`** — narrowing it to email alone would drop a phone-only site's phone number. No live site is phone-only today, so that regression would also ship invisibly
- **there is no automated guard.** The Go half (`RenderFallbackFooter`) has a regression test; the DB half has none, and a component-wide detector was considered and rejected while closing 111 — it fires on 15 legitimate tool components whose bodies are filled by JavaScript at runtime. This entry IS the guard
- **source:** `bugs_closed/111`, closed 2026-07-31 after fleet-wide verification at the served bytes (0 of 16 live domains render a bare heading)
- **added:** 2026-07-31, bugfix_111 lane

### An ordering-critical migration left unapplied in `sql_for_agents/` is a loaded gun for the next `--apply` — the banner asking you not to apply it is read by nobody but you

- **footprint:** `scripts/migration/run-migrations.sh --apply`, `docs/agent_docs/sql_for_agents/NNN_*.sql`, `schema_migrations`, any change split into a Go half and a config half
- **fires when:** you write the config half of a two-part change and deliberately do not apply it because the binary has to ship first — the house pattern, correctly followed (`sql_for_agents/278`, `281`). The trap is that "unapplied" is not a state the directory can hold: `--apply` takes **EVERY** pending file in number order, and it is another session that runs it, hours later, with no idea your file is waiting on an image. Measured 2026-07-31: **67 pending files**, several of them ordering-critical halves belonging to threads that had gone quiet
- **the tell: none, in either direction.** Your migration applies cleanly and its own guard passes — the guard checks for DRIFT (is the row still the shape I was written against), not for ORDER (is the binary that reads this key live yet). The applying session sees a normal successful run. Your change then fails in the direction the pre-image behaves, which for a config path reading a key the old binary does not emit is nil → false → the workflow's else branch, silently, on every run. For `bugs_open/150` that meant every improvement sweep reporting "site is clean" — strictly worse than the bug being fixed
- **the check, and it is a rename:** end the filename with an UPPERCASE suffix — `NNN_name_HOLD.sql`. The runner's `SIDECAR_RE` (`_[A-Z][A-Z0-9_]*\.sql$`) excludes it from `--apply` while still **listing** it under *"Sidecars (hand-run only, NOT applied by this runner)"*, so it is held back visibly rather than hidden. Verify with `./scripts/migration/run-migrations.sh --no-probe` and confirm your file appears under Sidecars and NOT in the Pending list. Then write the two commands to apply it by hand into the file's own header, because the person who runs them will not be you
- **do NOT rely on the guard, the banner, or the number.** A `WHERE` clause that refuses a drifted row still applies happily at the wrong time; a banner is advisory to a human who is not reading your file; and taking the "next free" number from an `ls` you ran earlier is its own trap (two numbers were claimed under me inside forty minutes — `WRONG_CALLS.md`, 2026-07-31)
- **also worth knowing:** a *dry run* of the runner EXECUTES each pending file inside a doomed transaction (the already-applied probe), so a pending file's SQL runs and rolls back. Harmless for a guarded `UPDATE`; think twice if your file has side effects a rollback cannot undo
- **source:** `bugs_open/150`, 2026-07-31 — the hazard was created by this lane and caught by reading the runner's header before leaving the file in place. `sql_for_agents/278` (bugs_open/154's config half) carries the same exposure today and is in the pending 67
- **added:** 2026-07-31, bugfix_150 lane

### Chrome selection now has TWO guard scans with disjoint blind spots and a shared vocabulary — add a chrome function to one and you have silently halved the coverage

- **footprint:** `platform/orchestration/actions/chrome_selection_test.go` (`chromeFunctionLiteral`, `TestNoChromeSelectionHandTypesItsOwnLookup`), `platform/orchestration/actions/chrome_build_path_test.go` (`chromeFunctionGoLookup`, `TestNoBuildPathResolvesChromeByPlainFunctionLookup`), `ResolveChromeComponent` / `chromeEligibleSQL` / `ChromeSlotFunction` in `component_library.go`, `content_components.function` values `site-header`/`site-footer`/`head`
- **fires when:** you add a chrome slot or function (a `site-nav`, a `site-banner`, a second head), or you widen `ChromeSlotFunction`. Both scans hard-code the same three function names in **separate regexes in separate files**, and neither knows the other exists at runtime. Adding a function to one leaves the other blind to it — and each is already blind to the *other's* form by design, so the gap is the union of two silences
- **why there are two, and why they must NOT be merged:** the 118 scan matches hand-typed **SQL** and **skips `component_library.go`** (that file legitimately contains the predicate). The 167 scan matches the **Go call** form — `GetComponentByFunction(ctx, db, "site-header", …)` — and **covers** that file, because all three of `bugs_closed/167`'s defects lived there and contained no SQL at all. Each exemption is correct; each is exactly where the next instance hides
- **the tell: none. Both scans pass, loudly and greenly, on a codebase they cannot see the defect in.** A green guard reads as "this class is handled" and says nothing about its exemption list. That is how 167 existed for four days inside a class 118 had already "closed"
- **the check, before you trust either scan:** induce the defect. Drop a throwaway `.go` file in the package containing the form you care about, run the test, and confirm it FAILS naming your line — then delete it. Both scans were proven this way rather than assumed:
  ```
  --- FAIL: TestNoBuildPathResolvesChromeByPlainFunctionLookup
      chrome function resolved through a section-shaped lookup at [zz_induce_167_tmp.go:11]
  ```
- **the other half of the same trap:** `ResolveChromeComponent` **ALWAYS returns a row**, deliberately — so `err == nil` tells you nothing about whether the component is fit to be chrome. Live today, `head` has NO eligible component and the resolver's answer is `Document Head`, an 8,523-char `component_level='section'` component. **Any new call site must gate on the `eligible` return**, not on the error; using the row because a row came back renders a page section as site chrome, which is the whole of `bugs_closed/167`
- **and a third door, still open:** the style-collection **pin** path (`GetComponentByID`) applies **no eligibility predicate at all** and neither scan looks at it — `bugs_open/170`, three deployed sites pinned to an `is_active=false` header. Note the asymmetry if you fix it: `forked_from IS NULL` is right for pool *selection* and **wrong** for a pin, because pinning a site to its own fork is the intended use
- **source:** `bugs_closed/167`, closed 2026-07-31; `bugs_closed/118` for the first scan; 016b §9 "A guard's exemptions are the exact shape of its blind spot"
- **added:** 2026-07-31, bugfix_167_chrome_build_path lane

### `site_nav_groups` holds group_keys NO Go code writes, and every nav rebuild deletes them and recreates only three

- **footprint:** `platform/orchestration/actions/populate_nav_tables_action.go` (`PopulateNavTablesAction`, `classifyPagesForNav`, `upsertNavGroup`), tables `site_nav_groups` / `site_nav_items`, `group_key` values other than `primary`/`legal`/`utility`
- **fires when:** you reason about nav from the action's source, or you build any check, cohort or report that partitions nav by `group_key`. Reading the code tells you a site's nav groups are exactly `primary`, `legal` and `utility` — those are the only three `upsertNavGroup` is ever called with. **The live data disagrees.** Measured 2026-07-31: `robot-hands.com` and `leopardessconsulting.co.uk` both carry a `tools` group, created 2026-07-18 and 2026-07-31 12:27 respectively. `grep -rn site_nav_groups --include=*.go platform/ internal/ pkg/` returns **only** `populate_nav_tables_action.go` and `nav_tables.go`, and neither writes that key — so they arrived by hand SQL, which is routine on these tables (there are eight `bak_*nav*` backup tables in the schema from previous hand surgery)
- **the two-part trap.** (1) `PopulateNavTablesAction` runs `DELETE FROM site_nav_items WHERE site_id = $1` then `DELETE FROM site_nav_groups WHERE site_id = $1` — **unscoped by group** — and then upserts only the three it knows. So any hand-created group and its items are destroyed by the next nav rebuild and never recreated, silently. (2) The pages in such a group are **not** orphaned by that: the current classifier re-homes them, so `robot-hands.com`'s `/tools/gripper-safety-factor-calculator` moves from `tools` into `utility` and the site's total item count is unchanged
- **the tell: none, and the totals actively hide it.** Item counts before and after are identical (robot-hands: 17 = 17), so any check comparing totals passes. Only a per-`group_key` comparison sees it, and that comparison is itself wrong for a different reason — see below
- **the check, before you partition nav by group:** run the group-level comparison and expect disagreement to be NORMAL, not a defect:
  ```sql
  SELECT s.domain, g.group_key, count(i.id) AS items
  FROM site_nav_groups g JOIN sites s ON s.id=g.site_id
  LEFT JOIN site_nav_items i ON i.group_id = g.id
  GROUP BY 1,2 HAVING g.group_key NOT IN ('primary','legal','utility');
  ```
  (Note the join is on `i.group_id = g.id`, not `i.site_id = s.id` — joining both children off `sites` multiplies items by groups; `WRONG_CALLS.md`, 2026-07-31.)
- **do NOT build a per-`group_key` completeness cohort, guard or drift check.** `classifyPagesForNav` re-homes pages between groups as a matter of course, so a per-group comparison scores a legitimate re-homing as a 100% loss of one class. `bugs_open/165` asked for exactly that shape for its site B and the data refused it: a per-group cohort reads `tools` 0/1 = 0% and would refuse robot-hands.com's nav rebuild **for ever**. Group membership is a classifier OUTPUT, not an independently-losable class. Pinned by `TestNavFloorAllowsAPageReHomedBetweenGroups` in `nav_prune_floor_test.go`, which fails if anyone adds one back
- **source:** `bugs_open/165` site B, 2026-07-31 — found while measuring cohorts against production rather than copying the ones the bug file proposed. Related: the `nav-updater` landmine above (URL-prefix deletion), and `bugs_open/149` A2 for the classifier's membership contract
- **added:** 2026-07-31, bugfix_165 sites B+C contribution lane

---

### Pointing a new domain at an existing single-vhost box silently serves the OLD site under the NEW name — and idea.uk's engine will take money on it

- **footprint:** `116.203.204.115`, `idea.uk`, `docs024_key_docs_latest/idea.uk/golang_files/service.go`, `setup.sh`, nginx `default_server`

Adding an A record for a new domain that points at a box already serving another
site does **not** produce a 404, a holding page, or any other signal that the new
domain is unconfigured. nginx hands the request to its default server and the
**existing site is served in full under the new hostname**. Measured 2026-07-31:
`webdesign.uk`, `ugg2.com` and `idea.uk` returned **byte-identical** bodies
(`md5 cf4c46c2b4e0`, `<title>idea.uk — …</title>`) within minutes of the records
being added.

**On the idea.uk box specifically this is a money bug, not a cosmetic one.** The
engine never inspects the request host — `grep -n 'r\.Host' main.go service.go
billing.go` returns **no match**; redirect targets come from the configured
`PublicBaseURL`. So the shop is fully functional on the wrong domain and **a real,
payable order can be created there**, with the buyer bounced mid-checkout to a
domain they never visited. The "does it work?" check (`curl` → `200`) is the check
that *hides* this: 200 is the symptom.

**the check:** after pointing any domain at a shared or pre-existing box, compare
the body against the site you did **not** intend to serve, not against nothing:

```bash
for d in <new-domain> <the-site-already-on-that-box>; do
  echo -n "$d : "; curl -sS -m 15 "https://$d" | md5sum | cut -c1-12
done
# identical hashes = you are serving the old site under the new name
```

**Fix it at the edge, never on the box** — a Cloudflare Redirect Rule or removing
the A record touches nothing on a live earning machine. Use a **302**: a 301 is
cached near-permanently by browsers and you will be undoing it when the real site
ships.

**Two more traps in the same move.** (1) The origin's certificate almost certainly
does not cover the new name — `openssl s_client -servername <new-domain>` on that
box returns `subject=CN=idea.uk` with SAN `DNS:idea.uk` only. If Cloudflare still
returns 200, that **proves** the zone is on SSL mode "Full", not "Full (strict)",
because strict would fail the handshake — a free way to read a dashboard setting
you cannot see. (2) HSTS leaks across the borrowed vhost:
`Strict-Transport-Security: max-age=31536000; includeSubDomains` served under the
new hostname pins **every future subdomain of it** to HTTPS-only for a year, which
matters if that domain was chosen to host wildcard subdomains later.

### The scratchpad is a 16G tmpfs shared by ~80 sessions — and when it fills, a SUCCESSFUL command is reported to you as an ERROR

- **footprint:** `/tmp/claude-1000/...`, `CLAUDE_CODE_TMPDIR`, any `git archive HEAD` / `git worktree` extraction, `git clone`, `cp -r` of this repo, large `psql`/`kubectl` output redirects, the Bash tool's `tasks/*.output` files
- **fires when:** you extract a copy of this repo into your scratchpad — which is the *recommended* practice for testing against committed HEAD, and nothing tells you where to put it. **One checkout is ~440M**, `/tmp` is a **16G tmpfs** (not disk), and ~80 session directories share it. Seven such copies in finished sessions were holding **3.0G** on 2026-07-31; live sessions held another 4.5G, so it reached **100% twice in one evening**
- **the tell, and it is INVERTED from the usual one:** at 0MB free the Bash tool returns
  `"<session>/tasks is full (0MB free). The child process's stdout/stderr writes failed with ENOSPC"` — **an `is_error: true` about the OUTPUT PIPE, not about your command.** The command frequently **ran and succeeded**; only the capture of its stdout failed. So the danger is not a lost command, it is **re-running a non-idempotent one** — a commit, an INSERT, a deploy, a Kafka publish — because the tool told you it failed. **On any ENOSPC error, VERIFY THE EFFECT before retrying** (`git log`, the row, the pod), exactly as you would for `kcat` publishing at exit 0
- **the check, before you extract anything repo-sized:** `df -h /tmp`. Under ~3G free, put the tree on real disk instead — `/` had **327G** free the day this was written — or set `CLAUDE_CODE_TMPDIR`. Your scratchpad is for notes and small artefacts; a 440M checkout is not a temp file
- **cleaning up safely, because the dirs are OTHER SESSIONS' state:** delete only the regenerable repo copies, never whole scratchpads — a scratchpad also holds hand-written work (this session's council submissions lived in one). Classify dead as "transcript `<uuid>.jsonl` untouched for >2h", confirm `lsof +D` is empty, and only remove subdirectories that still look like a checkout (`-d "$d/platform/orchestration" || -f "$d/go.mod"`), re-asserting that test immediately before each `rm`. On 2026-07-31 that reclaimed **3.0G of the 3.3G** reclaimable while leaving all 214M of real work untouched: 80% → 61%
- **the trap in the cleanup:** "quiet for 2h" is not "finished" — a session waiting on a long job looks identical. That is *why* the rule is repo-copies-only: losing a regenerable checkout costs a live session a re-extract, losing its scratchpad costs it work
- **source:** 2026-07-31 — hit independently by two lanes the same evening (this one, and the lane that recorded the ENOSPC-hides-success tell). Cleanup performed and measured; the durable fix is not putting checkouts in tmpfs
- **added:** 2026-07-31, bugfix 137 lane

### `has_items` on a PROMOTING step is call-scoped, and two other agents run the same step — the copy you branch on may honestly report zero for work that exists

- **footprint:** `platform/orchestration/actions/triage_detect_items_action.go` (`TriageDetectedItemsAction`), the `triage_detected_items` action in `agent_definitions` (`improvement-loop.triage_findings`, `design-audit-agent.triage`, `site-review-agent.triage`), any `conditional` step whose condition names `has_items`, `site_work_items.status='detected'`
- **fires when:** you add or read a branch on this action's output — the natural thing to do, since `has_items` is a fleet-wide convention across actions and reads correctly in the three places that already use it (`build-dispatch-loop.check_has_items`, `site-work-orchestrator.check_has_items` and `.check_has_fix_items`). Those three read their OWN loader. This one does not: the promotion is unconditional over the site (`WHERE site_id = $1 AND status = 'detected'`, no type filter), so **the first copy to run takes every row and every later copy honestly reports `promoted: 0`** — and the improvement loop calls both children before running its own
- **the tell: none, and the shape is why.** The action succeeds, logs success, and returns a *correct* value; the defect is a correct value routed to the wrong branch. `orchestration_states` shows `COMPLETED`. It is also **redundant with something that works** — `build-pipeline-trigger` dispatches `triaged`+`build` on its own 120s tick — so the fixes still happen and only the closing rerender goes missing. Measured twice: 67 findings promoted then "No issues found — site is clean" (`30692439`), and 27 findings the same way (`911ecdd8`)
- **the check, before branching on this action:** read `site_dispatchable` (bool) or `site_dispatchable_count` (int), **not** `has_items` — they count the site's work in a dispatchable status for the target pipeline whoever promoted it. If you must confirm the race itself, read the children's copies in the same row, never the parent's alone: `SELECT collected_data->'triage_result', collected_data->'call_design_audit'->'response'->'triage_result', collected_data->'call_site_review'->'response'->'triage_result' FROM orchestration_states WHERE orchestration_id='<id>';` — the parent's `promoted: 0` beside a child's `promoted: 24` is the signature
- **do NOT "fix" it by redefining `has_items`,** and do not add a fourth `triage_detected_items` caller without reading `RFC_006`: the first breaks three correct consumers, the second reproduces this verbatim
- **ordering, if you wire a condition to the new key:** it is emitted by the BINARY and read by config that is live IMMEDIATELY. On a chassis that predates the key it resolves to nil → false → the else branch, on **every** run. `sql_for_agents/281_..._HOLD.sql` is held back for exactly this reason
- **source:** `bugs_open/150`, fixed in code 2026-07-31 (`337fdd9af`, council `757cc7be` APPROVED); register `WDS-015`; 016b §9 + its 2026-07-31 addendum
- **added:** 2026-07-31, bugfix_150 lane

### A Cloudflare-fronted site's `robots.txt` is REWRITTEN at the edge — `curl` gives you the origin file plus text that is not in it, with the `x-amz-*` headers still present to reassure you

- **footprint:** `robots.txt` on any B2+Cloudflare domain (all 16 framework-managed domains), `~/projects/sites/<domain>/`, `b2 sync --delete`, any task that reconciles a local tree against what a site "actually serves"
- **fires when:** you fill a gap in a local tree from the live site — the obvious move when a file is in the bucket but not in your checkout. On mortgagecalculator.co.uk the served file is **2,327 bytes**; the object in B2 is **491**. The difference is a `# BEGIN Cloudflare Managed content` … `# END Cloudflare Managed Content` block (Content-Signal directives, `Disallow` for GPTBot/ClaudeBot/CCBot/…) that **Cloudflare injects on the way out**
- **the tell, and it points the WRONG WAY:** the response still carries `x-amz-id-2`, `x-amz-request-id`, `x-amz-version-id`, which is the usual proof that an object came from B2 rather than being synthesised by the edge. Those headers are honest about *where the object was fetched*; they say nothing about whether the body was rewritten in transit. **`cf-cache-status: DYNAMIC` is present too and also does not distinguish this**
- **why a careful session still misses it:** the injected block is at the **TOP**, above the origin's own rules. A `curl … | tail -5` — a reasonable way to confirm a file is non-empty and looks right — lands entirely inside the origin's content and shows nothing unusual. That is exactly how it was mischaracterised as "a real origin file, not Cloudflare's Managed robots.txt block" in `mortgagecalculator_couk_adoption/HANDOFF_2026-07-31`
- **the cost if you commit it:** the origin permanently carries a hardcoded copy of the managed block, which Cloudflare then injects **again** on every request — a duplicated directive set in the one file crawlers parse strictly. And it is invisible afterwards, because the served file still "looks like" the file you committed
- **the check, whenever bytes are the deliverable:** take them from the origin store, never through the CDN — `b2 sync b2://portfolio-sites/<domain>/ ./bucket` (no `--delete` = download only), then `grep -c "Cloudflare Managed"` the downloaded copy, which must be **0**. The positive control that makes this specific rather than paranoid: on the same domain **28 of 28** non-robots files were sha256-identical live-vs-local, so the edge rewrites `robots.txt` and not HTML generally
- **source:** 2026-07-31, mortgagecalculator.co.uk adoption lane — caught one command before the file was committed into the deploy repo. `mortgagecalculator_couk_adoption/NOTES` + `RUNBOOK` §2
- **added:** 2026-07-31, mortgagecalculator adoption lane

### `b2 sync --dryRun` is the v3 spelling and EXITS 2 on the v4 CLI — and the usual way of summarising its output turns that failure into "nothing will be deleted"

- **footprint:** `b2 sync`, `b2` CLI ≥4 (this machine: 4.7.0 / b2sdk 2.12.0), `~/projects/sites/.github/workflows/deploy-to-b2.yml`, any pre-flight for a `--delete` sync
- **fires when:** you simulate a destructive sync before running it — i.e. precisely the moment you are being careful. The flag is **`--dry-run`**, not `--dryRun`; the camelCase form exits **2** with a usage dump rather than an unrecognised-option error you would notice
- **the tell is ABSENT BY CONSTRUCTION, and this is the real trap:** the idiom `b2 sync … | tee out.txt; grep -i '^delete' out.txt || echo "(none)"` prints

  ```
  === DELETIONS the sync would perform ===
  (none)
  === UPLOADS the sync would perform ===
  (none)
  ```

  over the *usage dump*. **A failed dry run and a perfectly safe no-op produce identical output**, because `grep … || echo "(none)"` cannot distinguish "zero matches" from "the command never ran". You then push, having "verified" nothing
- **the check:** print `${PIPESTATUS[0]}` in the same block and require **0** — a pipeline hides the exit status of everything but the last stage, so `$?` after a `| tee` is `tee`'s and is always 0. More generally: **any "no findings" print needs a positive control emitted by the same run**; for a subprocess the exit status is the cheapest one there is
- **what a REAL dry run of this workflow looks like, so a correct result is not mistaken for a bad one:** for a domain already fully mirrored, expect **29 uploads + 35 deletes**, not silence. 30 deletes are `(old version)` B2 version-pruning each paired with a re-upload of byte-identical content; 5 are `.bzEmpty` folder placeholders. **The property to assert is not "no deletes" but "no live content file deleted without replacement"**: `comm -23 <(bucket listing) <(repo listing)` must be empty. Uploads happen even for identical bytes because `b2 sync` compares mtime and `--skip-newer` skips only when the **destination** is newer
- **source:** 2026-07-31, mortgagecalculator.co.uk adoption lane — the false-clean run was the gate on a `--delete` sync against a live site. `mortgagecalculator_couk_adoption/NOTES` + `RUNBOOK` §1/§4
- **added:** 2026-07-31, mortgagecalculator adoption lane

### `--fidelity high` is not a milder `locked` — it is the ABSENCE of a setting, and it silently selects the full recreate path that renames every URL and has an LLM rewrite every page

- **footprint:** `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh`, `platform/orchestration/actions/apply_adoption_plan_action.go:426`, `platform/orchestration/actions/adopt_verbatim.go`, `input_data.fidelity`, `datahelpers.CanonicalisePage`, doc 028's "fidelity dial"
- **fires when:** you adopt a site you intend to KEEP and reach for a middle setting, because doc 028 and the script's own `--fidelity <level>` usage line both present a five-position dial — `locked | high | medium | low | new`. **Only `locked` exists in code.** The comparison is a strict binary, `if fidelity := adoptionFidelity(...); fidelity == fidelityLocked`, and `082`'s NOTE says the rest are "recorded in `input_data` … but **modulating nothing**"
- **the tell: none — and every observable says it worked.** `high` is accepted by the script, stored in `input_data`, echoed back by the orchestration row, and the run completes successfully. It behaves *identically to passing no fidelity at all*
- **what it actually does:** `CanonicalisePage` synthesises a fresh URL for every page and discards the crawled one (`/repayment.html` → `/tools/repayment/index.html`), pages land `build_status='planned'` with `content_data.mode='recreate'`, and `page-build-handler` / `tool-recreation-handler` regenerate each page with an LLM — working hand-built tools included. Every indexed URL on the site changes
- **the check, before submitting:** decide by the *behaviour* you want, not the word. Byte-and-URL preserving is `locked` and **nothing else**. Editability is not the reason to avoid `locked`: `rebuild_policy='owned'` is a **per-page** flag with real readers (`rerender_single_page_action.go:310`, `save_page_sections_action.go:149-156`, `reconcile_site_plan_action.go:233`), so a locked site can be opened to the pipeline one page at a time. After submitting, confirm the branch actually taken — `SELECT collected_data->'input_data'->>'fidelity' FROM orchestration_states WHERE owner_agent_type='site-adoption-agent' ORDER BY created_at DESC LIMIT 1;` — since `call_agent`'s `input_mapping` is an ALLOW-LIST and dropped `fidelity` entirely until migration 274
- **the assertion that INVERTS between the two paths:** `HANDOFF_2026-07-31_adopt_mortgagecalculator.md` §5d says `needs_content_page` + `needs_tool_recreation` must be **0** and to stop if either appears. That is the **`locked`** assertion. Under `high` those work items are the *intended* output and their absence means the run did nothing — copy the check across and you halt a correct run at its first correct step
- **source:** 2026-07-31, mortgagecalculator.co.uk adoption lane (owner decision D1 taken with the code quoted); `mortgagecalculator_couk_adoption/PLAN` D1 + `RUNBOOK` §6
- **added:** 2026-07-31, mortgagecalculator adoption lane

### The diagnosis bundle's two new omission counts are NULL on every bundle written before v1.0.12xx — so "no omissions" and "no data" look identical, and the honest reading is the pessimistic one

- **footprint:** `diagnosis_artifacts.metadata` keys `symbols_omitted_size` / `symbols_unreadable` / `truncated`, `platform/orchestration/actions/diagnose_assemble_bundle_action.go` (`bundleArtifact`, `persistBundleArtifact`), any query measuring diagnosis-bundle coverage
- **fires when:** you measure how often the bundle drops in-scope code — i.e. when you verify `bugs_open/164`'s fix, or size the problem again later. `(metadata->>'symbols_omitted_size')::int` returns **NULL, not 0**, for every bundle written before the fix shipped (2026-07-31), and `sum()`/`avg()` skip NULLs silently while `count(*)` does not
- **the tell is absent, and it points the wrong way:** a `GROUP BY` over the whole table shows a large clean-looking cohort with no omissions, which reads as "this was rare all along" — the exact opposite of the truth, since those are precisely the rows written by the version that *could not* report. The pre-fix rate was **18 of 254 bundles (7.1%)**, measured off `truncated`, which is the only key that exists on both sides
- **the check:** filter on `created_at >= '<the roll that shipped it>'` before quoting either new key, or `COALESCE(...,0)` **and** say in the same sentence which window the denominator covers. To compare across the boundary use `metadata->>'truncated'` — it predates the fix. ⚠ Its MEANING shifted though: pre-fix it meant "the loop stopped early", post-fix "at least one body did not fit". Verified against the live DB (not a source grep) that **no** live agent reads it at any depth, so nothing broke — but a trend line spanning the roll is comparing two definitions
- **also, and this is the one that bites hardest:** `diagnosis_artifacts` is retention-clocked (`bundle_retention_days` default **30**), and `orchestration_states` retains barely a **day** here. So `count(*)` is "still retained", never a census — always select `min/max(created_at)` in the same query and print the window beside the figure. A rate quoted without its window is the failure mode, not a style preference
- **do NOT use `length(body) >= 59000` as a truncation proxy** (164's own filing suggested it): `body` is the WHOLE bundle including runtime evidence and schema, so a fat untruncated bundle scores as truncated — and the three *worst* real cases have `body_chars = 0`, so it misses them entirely. Wrong in both directions
- **source:** 2026-07-31, `bugfix_164` lane. `bugs_open/164` § MEASURED + § VERIFY, `bugfix_164_bundle_body_cap/RUNBOOK`. Council `75f3cd52` APPROVED; the NULL-vs-0 hazard was raised as a `bug_historian` advisory and this entry is where it was discharged
- **added:** 2026-07-31, bugfix_164 lane

### Cloudflare refuses in front of the island, and its refusals are indistinguishable from ours by status code

- **footprint:** `tools.apis.uk`, `internal/tools-api/middleware/ratelimit.go`,
  `internal/tools-api/middleware/cors.go`, `/opt/island/docker-compose.yml`,
  any curl/urllib check against a Cloudflare-fronted origin
- **fires when:** you verify an endpoint's behaviour from a script, or measure a
  rate limit, or conclude "my CORS/limit config did not take effect". No symptom
  beforehand — the status code is exactly the one your own code would return.
- **the tell:** the BODY, and the `server:` header. Ours are JSON from
  `httperr.JSONError` (`{"error":"..."}`, `content-type: application/json`).
  Cloudflare's are `text/plain`, carry `server: cloudflare` and a `cf-ray`, and
  say `error code: NNNN`:

  | you see | who refused | body | what it means |
  |---|---|---|---|
  | **403** | our CORS middleware | `{"error":"origin not allowed"}` | no `Origin` header sent |
  | **403** | **Cloudflare** | `error code: 1010` | browser-integrity check — python-urllib's default User-Agent is rejected. **Never reached the island.** |
  | **429** | our limiter | `{"error":"rate limit exceeded"}` | past `RATE_LIMIT_BURST` |
  | **429** | **Cloudflare** | `error code: 1015`, `retry-after: 9` | ~10 rapid requests from one IP. **Never reached the island.** |

- **why it is a landmine:** it fails in the direction that reads as *"my change
  did not ship"*, and it fabricates evidence about our own configuration. Hit
  three times in one session on 2026-07-31: a 403 that looked like our CORS
  refusing a driver; a 429 measured as "our ceiling is burst 5" that was partly
  Cloudflare's; and a 20-request burst that returned exactly **10** successes
  against a limiter configured for **20** — the app was right and the check was
  measuring the wrong thing.
- **the check:** to measure OUR behaviour, take Cloudflare out of the path and
  hit the container directly. This is the only way to see the app's real ceiling:
  ```bash
  ssh root@toolsapisuk.vs.mythic-beasts.com "docker exec island-tools-api-1 sh -c '
    for i in \$(seq 1 25); do wget -q -O /dev/null -S \
      --header=\"Origin: https://vonc.com\" \
      http://127.0.0.1:8080/api/v1/tools/gauntlet/round/<slug> 2>&1 \
      | grep -o \"HTTP/1.1 [0-9]*\" | tail -1; done' " | sort | uniq -c
  # 20 x 200 then 429 == RATE_LIMIT_BURST=20 exactly. Over the internet the same
  # test returns 10 x 200 because Cloudflare bites first.
  ```
  From outside, always send **both** `Origin` and a browser `User-Agent`, and
  **read the body before believing the status**.
- **the consequence worth knowing:** for internet traffic the binding rate limit
  is **Cloudflare's ~10 rapid requests**, not `RATE_LIMIT_*`. Raising the app's
  ceiling above that changes nothing a visitor can observe — it is defence in
  depth, not the effective ceiling. Do not "fix" a 1015 by raising `RATE_LIMIT_BURST`.
- **added:** 2026-07-31, gauntlet_dead_cta lane, retuning the limit after the
  record page went live.

### `mock.ExpectationsWereMet()` is NOT "no database call happened" — with no expectations registered it is trivially TRUE, so an assertion written that way passes while the code writes to the DB

- **footprint:** `sqlmock.New`, `mock.ExpectationsWereMet()`, `github.com/DATA-DOG/go-sqlmock`, any `*_test.go` asserting a NEGATIVE ("must not touch the DB", "never writes", "no side effect") — 33 files in this tree call it; `platform/orchestration/actions/diagnose_plan_refusal_test.go` (`dbTouchWatcher`)
- **why it is a landmine:** the API reads like English and means something narrower.
  `ExpectationsWereMet` reports expectations that were **registered and not consumed**.
  Register none and it is satisfied by definition — it never sees an *unexpected* call.
  So the idiom "register nothing, then assert ExpectationsWereMet" is a guard that
  cannot fail, sitting in the file looking exactly like coverage, and it is counted as
  coverage by the next reader. It fires with no symptom: the test is green, the comment
  above it states the contract in strong terms, and nothing is checking it.
- **measured, not reasoned (2026-07-31, bugs_open/162):** four assertions in
  `diagnose_plan_refusal_test.go` said "must not touch the DB". `recordPlanRefusal` was
  moved above the opt-in check — the one edit that would silently change a
  non-opted-in consumer's contract — and **all four still passed**. One of them had been
  written ten minutes earlier, by this lane, in direct answer to a council objection
  asking for exactly that guarantee to be pinned by a test.
- **the check:** never assert a negative through the mock's own bookkeeping. Observe the
  call instead. Where the code logs on DB failure, an observed logger is enough, and
  sqlmock with no expectations makes every call fail, so every attempt logs:
  ```go
  core, logs := observer.New(zapcore.DebugLevel)
  // ... call the function with zap.New(core) ...
  // then assert no entry matches the messages the DB paths emit on failure
  ```
  See `dbTouchWatcher` in `diagnose_plan_refusal_test.go` for the worked version.
  **And whichever way you write it, mutate the code and watch the test fail.** That is
  the only thing that distinguishes a guard from a decoration, it costs about a minute,
  and it is what caught this.
- **scope, honestly:** none of the 33 files is *wholly* inert (each registers
  expectations somewhere), so the failure is per-TEST, not per-file. 4 found, 4 fixed,
  all in one file. The other 32 files are **[UNMEASURED]** — a per-test audit is a
  reasonable sweep for someone, and the grep to start from is
  `grep -rn -B15 "ExpectationsWereMet" --include=*_test.go | grep -iE "must not|never|no db"`.
- **added:** 2026-07-31, bugfix_162 lane.

### An empty `agent_error_log` for `FIX_PLAN_VALIDATION_REFUSED` does NOT mean no fix plan was refused — two of the five terminal exits write nothing, deliberately

- **footprint:** `agent_error_log` where `error_code='FIX_PLAN_VALIDATION_REFUSED'`, `platform/orchestration/actions/diagnose_persist_fix_plan_action.go` (`planValidationRefusal`, `recordPlanRefusal`, `planRefusalErrorCode`), `diagnosis_artifacts` `kind='iteration_note'` + `metadata->>'note_kind'='plan_validation_refusal'`, the `diagnose_persist_fix_plan` action in `agent_definitions`
- **why it is a landmine:** the action's own comment used to say the row's absence meant
  "no refusal happened", so the misreading was *documented*. It is false for the exits
  that write nothing: a consumer with **no `repair_step`** (opt-in is the whole design of
  `bugs_open/099` candidate 2 — `council-gate` is deliberately out) refuses and leaves no
  row, no artefact, and `orchestration_states.error` **NULL**. The only trace is
  `collected_data->>'__step_error'`. So the query comes back clean and the run reads
  clean, which is precisely the silent-loss complaint 099 was filed about.
- **the check:** before reading an empty result as an absence, get the population —
  which consumers can even reach the recording path:
  ```sql
  SELECT a.type, s.key, s.value->'config'->>'repair_step' AS repair_step
    FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
   WHERE a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
     AND s.value->>'action' = 'diagnose_persist_fix_plan';
  ```
  A NULL `repair_step` means that consumer's refusals are invisible to your query by
  design. 2026-07-31: 4 consumers — `feature-designer` and `fix-proposer` opted in,
  `council-gate` deliberately out, `council-gate-036scratch` `is_active=false`.
- **also:** the table's timestamp column is **`occurred_at`**, not `created_at`. The
  wrong name errors loudly here, but the same guess against a table carrying both is how
  a window silently slips.
- **added:** 2026-07-31, bugfix_162 lane (comment corrected in the same commit, `417d6fd87`).

### `mistyped_deployed_page` is a DECISION, not a job — it has no handler by design, and it deliberately occupies a dedup slot for ever

- **footprint:** `mistyped_deployed_page`, `site_work_items.status='needs_human_review'`,
  **`site_work_items.handler_agent`** (the column-level trap in the fourth bullet
  applies to EVERY row in the table, not just this item type — added to the footprint
  2026-08-02 after it bit a second lane verifying an unrelated no-handler refusal),
  `platform/orchestration/actions/apply_gap_plan_action.go`
  (`applyNewPage`, `refuseDeployedPageTypeConflict`), the `apply_gap_plan` action's
  `applied` / `page_created` result fields, any drain over the human-review queue
  (`bugs_open/033`)
- **the trap:** every other `site_work_items` row names a `handler_agent` that can
  resolve it. This one **cannot be resolved by any agent**, and that is the fix, not
  an omission. `bugs_closed/081` established that no predicate distinguishes a real
  news listing from a catalog index that embeds one — on robot-hands.com both carry
  `sections=["news-listing"]` byte-identically and a third page of that shape is
  archived — so the platform refuses to guess and asks a person. A queue drain that
  assumes "needs_human_review + no handler ⇒ misfiled, route it somewhere" will
  route this to an LLM, which will guess, and a wrong guess **re-types a live page
  and breaks it**. `handler_agent` is `NOT NULL DEFAULT ''::text`, so it reads as
  empty string, not NULL — a `handler_agent IS NULL` filter will not even find it.
- **the check:** before draining or re-routing a `needs_human_review` row, read
  `spec->>'bug'`. This type carries `bugs_open/081` there, plus `existing_type`,
  `wanted_type` and a ready `resolution` UPDATE. The only correct actions are: a
  human runs that UPDATE, or the item is closed as "this page should not hold that
  role". **Do not hand it to a model.**
  ```sql
  SELECT s.domain, swi.status, swi.spec->>'page_name', swi.spec->>'existing_type',
         swi.spec->>'wanted_type', swi.spec->>'resolution'
  FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
  WHERE swi.item_type='mistyped_deployed_page';
  ```
- **the second half, and it is deliberate:** the row holds a **non-terminal** dedup
  slot (`idx_swi_dedup`, WII-005), keyed per `(site, page, wanted_type)`. So for as
  long as the decision is outstanding the originating check **cannot** file a
  duplicate — which is the point, and also means the queue depth will not fall on
  its own. It is not a leak. Its sibling `missing_news_page` item is set to
  **`blocked`, not `complete`**, for the same reason: `complete` over an untouched
  defect is what let this loop run from 2026-05-01 to 2026-07-31 unnoticed.
- **and the caller-visible change:** `apply_gap_plan` now returns `applied:false`
  with `reason:"deployed_page_type_conflict"` where it previously returned
  `applied:true`. Checked before shipping: no active `agent_definition` branches on
  that field — **re-check before writing a `conditional` step that does**, because
  a plan that legitimately refused now looks like a plan that failed. Use the new
  `page_created` field to tell a created page from a refreshed one.
- **added:** 2026-07-31, bugfix_081 lane. Committed but **NOT LIVE** — inert until
  the next chassis roll, so the type does not exist in production yet and a query
  for it returns zero rows today. That zero is not evidence of anything.

### `grep` in a Claude Code session is a FUNCTION wrapping ugrep, not GNU grep — and it can return exit 1 / zero matches on an ERE that GNU grep matches

- **footprint:** every `grep` call in a Bash tool block on this machine, and especially any attempt to reproduce locally what a CI job, a `.sh` script or a pod does with `grep`; `~/projects/sites/.github/workflows/deploy-to-b2.yml`, `scripts/*.sh`
- **fires when:** you re-run a pipeline from a workflow or script to preview what it will do. `type grep` shows a shell function that execs `ugrep 7.5.0` with `-G --ignore-files --hidden -I --exclude-dir=…`. **It is a different regex engine**, and greedy negated classes that must backtrack across a following literal do not behave the same:

  ```
  printf 'a.b/c\n'                              | grep -E '^[^/]+\.[^/]+/'  -> exit 1, NO MATCH
  printf 'mortgagecalculator.co.uk/index.html\n' | grep -E '^[^/]+\.[^/]+/' -> exit 1, NO MATCH
  printf 'mortgagecalculator.co.uk/index.html\n' | grep -E '^.+\.[^/]+/'    -> MATCHES
  command grep -E '^[^/]+\.[^/]+/'  (real GNU grep)                         -> MATCHES
  ```

- **the tell: none, and it is INDISTINGUISHABLE FROM A TRUE NEGATIVE.** Exit 1, no output, nothing on stderr — exactly what "the thing genuinely isn't there" looks like. The `--ignore-files`/`--exclude-dir` flags add a second silent-miss mode when grepping trees: a file matched by an ignore rule is skipped without comment
- **what it nearly cost:** previewing `deploy-to-b2.yml`'s changed-domain computation locally printed **nothing**, which in that workflow means `CHANGED` is empty and it falls through to `ls -d */` — *sync every domain in the repo*. That reads as a serious deploy-pipeline defect. The runner uses real GNU grep and had computed `Changed domains: mortgagecalculator.co.uk` correctly on both runs; **the workflow was never at fault, the local reproduction was**
- **the check:** use **`command grep`** whenever you are reproducing behaviour that executes somewhere else (CI, a pod, a shell script), and then confirm against **that system's own log** rather than your local re-implementation of its logic. If a local preview of remote behaviour disagrees with the remote, suspect the preview first
- **source:** 2026-07-31, mortgagecalculator.co.uk adoption lane. `mortgagecalculator_couk_adoption/NOTES`
- **added:** 2026-07-31, mortgagecalculator adoption lane

### Three of this platform's domains NEST inside each other's names, so `ILIKE '%domain%'` silently returns another site's rows — populated and plausible, not empty

- **footprint:** `sites.domain`, `site_work_items.spec::text`, `orchestration_states.collected_data->'input_data'->>'destination_domain'`, any pre-flight "is another lane working this domain?" query; the domains `loancalculator.co.uk`, `mortgagecalculator.co.uk`, `loanandmortgagecalculator.co.uk`
- **fires when:** you scope a check by domain substring — the natural first phrasing, and the one the adoption runbooks use. **`loanandmortgagecalculator.co.uk` CONTAINS `mortgagecalculator.co.uk`**, and it also contains `loancalculator`… no it does not, but `loancalculator.co.uk` is a substring of nothing here — the trap is specifically the `%mortgagecalculator.co.uk%` pattern, which matches both the short and the long domain
- **the tell: the wrong answer is POPULATED, which is worse than empty.** Asking whether another lane was mid-adoption on `mortgagecalculator.co.uk` returned **41 `page_rerender` rows in `triaged`** — a plausible page count for that site, and exactly the sort of result that stops you adopting. Every row belonged to the sibling; 41 is the count the sibling's own handoff reports. The same flaw made two `COMPLETED` adoption runs look like ours, and reading `input_data.fidelity` off them returns `locked`, which could easily be mistaken for evidence about *your* run's settings
- **an absence would have made you look harder; a confident number invites you to act on it.** That asymmetry is the whole reason this is a landmine and not a typo
- **the check:** join and match exactly, never `ILIKE '%…%'` —

  ```sql
  SELECT s.domain, w.item_type, w.status, count(*)
    FROM site_work_items w JOIN sites s ON s.id = w.site_id
   WHERE s.domain = 'mortgagecalculator.co.uk' AND w.status NOT IN ('complete','cancelled','rejected')
   GROUP BY 1,2,3;
  ```

  If you must pattern-match, anchor it: `s.domain = $1`, or at minimum verify by grouping **by `s.domain`** so a foreign site announces itself in the output instead of hiding in a total
- **source:** 2026-07-31, mortgagecalculator.co.uk adoption pre-flight. `mortgagecalculator_couk_adoption/NOTES` + `RUNBOOK` §6
- **added:** 2026-07-31, mortgagecalculator adoption lane

### A relative path is resolved against a cwd an earlier `cd` left behind — and `find .`, `git log --`, `git ls-tree` and `ls` are ALL blind the same way, so they corroborate each other

- **footprint:** any `Bash` call using a repo-relative path; `cd <dir> && <cmd>` anywhere earlier in the session; `find .`, `git ls-tree -r HEAD --name-only`, `git log -- <relpath>`, `ls <relpath>`
- **fires when:** you run `cd docs/x/y && python3 thing.py` for convenience, and then — several tool calls later — check whether some unrelated file exists. The working directory **persists between calls**. A "Shell cwd was reset" notice appears for *some* calls and not others, so its absence is not a guarantee, and nothing in the later command looks wrong.
- **the tell: you conclude a file, a script, or a whole directory tree DOES NOT EXIST** — and the conclusion arrives with what feels like overwhelming corroboration. Worked case (2026-08-01): `ls docs/.../LANDMINES.md` → no such file; `find . -name "LANDMINES*"` → nothing; `git log --all -- docs/.../LANDMINES.md` → nothing; `git ls-tree -r HEAD --name-only | grep -i landmine` → nothing; `ls scripts/landmines-sync.py` → no such file. **Five checks, unanimous, all wrong.** The file is 342KB and tracked. The cwd was a `builder/` subdirectory from a `cd` five calls earlier, so every relative path resolved beneath it — `find .` searched only that subtree, `git ls-tree` listed only that subtree, and `git log -- <relpath>` scoped to a path that does not exist.
- **why it is worse than a plain typo:** the checks are *methodologically different* (filesystem, index, history) which reads as independent confirmation, but they share one input — the cwd — so they fail identically. This is [[two-blind-checks-agree-with-each-other]] with the blindness supplied by the shell rather than by the query. It cost a false statement to the user and a **false claim inside a live council submission**, which is the expensive kind.
- **the check:** for any existence or absence claim, use an **absolute path**, or `cd` to the repo root in the same command:

  ```bash
  ls -la /home/ant/projects/agentchassis/docs/.../LANDMINES.md          # absolute
  git ls-files | grep -i landmine  # anchored
  ```

  `git ls-files` (whole index regardless of cwd) is safer than `git ls-tree -r HEAD` (subtree-scoped). **And print `pwd` in the same call as any absence check** — one extra word, and it makes the failure mode visible instead of invisible.
- **source:** 2026-08-01, provocation_pipeline lane. `provocation_pipeline/NOTES_provocation_pipeline.md`; `WRONG_CALLS.md`
- **added:** 2026-08-01, provocation_pipeline lane

### A pod-grep POSITIVE control cannot prove your pattern is spelled right — only a NEGATIVE control can, and case is the trap

- **footprint:** `kubectl exec … strings /app/agent-chassis | grep -c`, any deploy proof, `bugs_open/153`, the fleet rule "grep a string your change ADDED plus a positive control in the same exec"
- **fires when:** you verify a Go change at the pod and your change added **more than one** log line. Go log messages are sentence-cased at the START of the message, so `logger.Warn("No eligible header component…")` is capitalised while `logger.Warn("RenderHead: no eligible head component…")` is not — the same change, two spellings, and a single retyped grep pattern matches one of them. Measured 2026-07-31 on `v1.0.1225`: `header: 0, footer: 0, head: 1` on a binary containing **all three**
- **the tell: none, and the standing fleet rule actively reassures you.** The positive control returns 1, so `strings | grep` is proven to work on that binary, and a `0` therefore reads as a *genuine absence* — "the roll did not carry my fix". That is exactly the conclusion the control was supposed to prevent, arrived at with the control in place. **A positive control tests the PIPELINE; it is by construction a DIFFERENT string, so it can never test your pattern's spelling**
- **the check: add a NEGATIVE control — grep for a string your change REMOVED.** It must return **0**, and a 0 there is only possible if the old code is gone, which means the new code is present whatever your other patterns say. Positive control proves the fix ARRIVED; negative control proves the old code LEFT. Run both in the same exec, on every replica:
  ```
  strings /app/agent-chassis | grep -c "<literal your change ADDED>"      # expect >0
  strings /app/agent-chassis | grep -c "<literal your change REMOVED>"    # expect 0
  strings /app/agent-chassis | grep -c "<unrelated literal known present>" # expect >0
  ```
- **and: use `grep -ic`, or paste the literal out of the source rather than retyping it.** Retyping is where the case is lost. Sibling of the two pod-grep entries already here (non-ASCII splitting a literal; the linker packing constants so anchors misfire) — same command, third distinct way to get a false 0
- **source:** `bugs_closed/167`, 2026-07-31 — I published the mis-cased command in a RUNBOOK and a closed bug file before catching it; `WRONG_CALLS.md` same date. The lesson "a grep proves an absence only for the SPELLING it searches" was already recorded and did not prevent it
- **added:** 2026-07-31, bugfix_167_chrome_build_path lane

---

### The idea.uk box has NO catch-all vhost, so it serves the whole shop to any hostname — and the obvious fix (enable Ubuntu's stock `default`) takes the site down at the next reload

- **footprint:** `116.203.204.115`, `/etc/nginx/sites-enabled`, `/etc/nginx/sites-available/default`, `idea.conf`, `default_server`, `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/box/`, `setup.sh`

> **UPDATED 2026-08-02 (same lane): the catch-all is now LIVE on the box** —
> applied with owner authorisation (`000-default-deny.conf`, `nginx -t` clean),
> all three probes below flipped to denial, the 16-route baseline re-ran
> identical, and `setup.sh` now writes and links the catch-all itself, so a
> fresh provision no longer reopens the hole. **The trap inside the fix is
> UNCHANGED and now fires faster:** the live `000-default-deny.conf` claims
> `default_server`, so symlinking Ubuntu's stock `default` breaks `nginx -t`
> immediately — everything below stands as the evidence and the per-box checks.

Measured 2026-08-01: `sites-enabled/` on the idea.uk box holds **exactly one
file** (`idea.conf`) and **nothing anywhere claims `default_server`** — Ubuntu's
stock `default` vhost is present in `sites-available/` but is **not symlinked**.
So idea.conf's blocks are the de-facto default for `:80` and `:443`, and the
complete site — **money path included** — is served to every unmatched hostname
and to requests carrying no name at all:

```bash
curl -sk -o /dev/null -w '%{http_code}\n' https://116.203.204.115/                         # 200  no SNI
curl -sk -o /dev/null -w '%{http_code}\n' --resolve fake.example:443:116.203.204.115 \
     https://fake.example/                                                                 # 200  foreign SNI
curl -s  -o /dev/null -w '%{http_code}\n' -H 'Host: fake.example' http://116.203.204.115/   # 301  foreign Host
```

**Why it misleads:** on 2026-07-31 two unrelated domains (`webdesign.uk`,
`ugg2.com`) briefly pointed here and served **byte-identical** idea.uk content,
and a real payable order was creatable from the wrong one. That reads as a DNS or
a CDN problem and it is neither — it is a missing catch-all, and **any** hostname
pointed at this address reproduces it. The application cannot save you: the Go
tool never reads `r.Host` (`grep -n 'r\.Host' main.go service.go billing.go` →
no match) and builds every redirect from `PublicBaseURL`, so a buyer who starts
checkout under the wrong name is bounced mid-flow to a domain they have never
visited.

**the check, and the trap inside the fix:** add a catch-all — `box/default-deny.nginx`
in the idea.uk lane is written for exactly this (`server_name _`, `return 444` on
:80, `ssl_reject_handshake on` on :443, installed as a NEW file so `idea.conf` is
never edited and rollback is deleting a symlink). **Do NOT instead symlink
Ubuntu's stock `sites-available/default`**: it claims `listen 80 default_server`
too, and two `default_server`s on one address:port is an `nginx -t` **error**, so
the box keeps serving until someone reloads and then does not come back. Verify
which files are actually live before adding one:
`ls -la /etc/nginx/sites-enabled/ && grep -rn default_server /etc/nginx/sites-enabled/`.

Two facts decide whether a catch-all is safe on a box like this, and both must be
re-checked per box rather than carried across: **(1)** how ACME renews — this box
is `authenticator = webroot` over port 80 with `Host: idea.uk`
(`/etc/letsencrypt/renewal/idea.uk.conf`), so renewal matches the named vhost and
can never fall into the catch-all; a **TLS-ALPN** authenticator would be a
different answer. **(2)** whether rejecting no-SNI clients costs anything — here
the only certificate is `SAN: DNS:idea.uk`, so a client arriving without SNI
*already* fails validation and cannot be a working visitor.

**And you cannot settle it from the logs:** nginx here writes the stock
`combined` format, which carries **no `$host`**, so the access log cannot tell
you whether real traffic ever arrives under a foreign hostname. Add `$host` to
`log_format` first if you need traffic evidence rather than the certificate
argument — but that is itself a production reload.

⚠ **`setup.sh` re-provisioning removes any fix here** — it rewrites `idea.conf`
from its own template and runs `ufw --force reset`. A catch-all must go into its
stage-2 template as well, or the exposure returns silently on the next rebuild.

- **source:** 2026-08-01, `idea_uk_vm_site/RUNNING_NOTES` §X.34 + RUNBOOK §4a-bis; confirms and explains §5 of `idea_uk_vm_site/HANDOFF_2026-07-31_cloudflare_decision.md`

### `output_format` means two different things, and on an LLM step it now BUYS A SECOND LLM CALL

- **footprint:** `output_format`, `output_type`, `getOutputType`, `llmOutputVocabulary`, `platform/orchestration/actions/ai_actions.go`, `platform/orchestration/actions/database_actions.go`
- **fires when:** you add or change `output_format` on a workflow step, or you widen the LLM output vocabulary because "a value is being ignored".
- **the tell:** everything looks right. The key was ignored for months and the steps still worked — 782 of 785 JSON-declared outputs already parsed — so a wrong value here produces no error, just a wrong instruction set or a silent doubling of that step's LLM spend on its failure path.
- **the check:** **one key name, two vocabularies, two actions.** On `execute_llm_prompt` it means `json|text|html|markdown` (`getOutputType`, allow-listed). On `query_database` it means `array|object` (`database_actions.go:26`). They are unrelated and must never be merged — widening `llmOutputVocabulary` to accept `array` would let a database step's declaration select an LLM instruction set. Before touching either, confirm which action the step runs:
  ```sql
  SELECT ad.type, s.key, s.value->>'action', s.value->'config'->>'output_format'
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key,value)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->'config' ? 'output_format';
  ```
  **And know what you are buying:** since `4c1a97874` (WFA-005), `output_format: json` on an LLM step means BOTH "append the JSON instructions" AND "re-ask once if the answer will not parse". That is correct for a step that genuinely needs JSON and wrong for one that declared it loosely and is happy with prose — the second case pays for a retry it does not want, and nothing will tell you. `output_type` takes precedence over `output_format` when both are non-empty.
- **the sibling trap:** the reverse of how this was found. `output_format` was written by **100** live steps and read by **nothing** in the LLM path, while `output_type` — the key the code actually read — was written by **6**. A declared-and-inert key is indistinguishable from a live one by inspection, so `grep -rn '"<the key>"' --include=*.go .` before believing any config key does what its name says (`bugs_open/134` is the same class).
- **source:** `bugs_open/119` §3, `bugfix_119_seat_retry/NOTES_seat_retry.md`, register WFA-005
- **added:** 2026-08-01, bugs_open/119 seat-retry lane

### A `090` diagnosis run on a symbol in a file over ~60KB returns bundles and NO verdict — and that looks exactly like a run still in progress

- **footprint:** `docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh`, `diagnosis_artifacts` (`kind='bundle'`, `body`, `iteration`), `diagnose-agent`, `diagnose-dispatch-loop`, and any Go file large enough to matter — measured targets today include `platform/orchestration/actions/component_library.go` (**93,905 bytes**), `platform/orchestration/actions/v3_site_actions.go`, `platform/orchestration/actions/rerender_pages_actions.go`
- **fires when:** you follow the standing rule — file a `090` before asserting a cross-cutting or structural root cause (CLAUDE.md "Diagnosis before debugging", and the owner ruling of 2026-07-31 that makes it the default for a `bugs_open/` file) — and the mechanism you are asking about lives in one of the platform's big action files. Which, for anything cross-cutting, is most of the time.
- **the tell, and it is inverted from what you expect:** the run does not error, does not stall, and does not report a gap. `orchestration_states` goes `COMPLETED`, `diagnosis_artifacts` fills with `bundle` rows — and there is **no `iteration_note`, no `council_report`, no `doc_note`**. Several bundles and no verdict is indistinguishable from a run that has not finished yet, so the natural reading is "give it another twenty minutes", and twenty minutes later it looks the same. The bundle itself states the problem, but in the middle of a 22–88KB document: `_(body omitted — 75174 chars, and 0 of the 60000-char body budget is already spent. It was found; it did not fit. Put THIS SYMBOL ALONE in next_scope to read it whole.)_` followed by `**This section is INCOMPLETE.** 0 of 1 in-scope symbol(s) rendered with a body.`
- **why it matters more than "one run wasted":** the diagnoser never saw the function under test, in **any** iteration, so whatever it did or did not conclude rests on the schema dump and the runtime evidence alone. A verdict from such a run would be worth less than it looks; an *absent* verdict at least fails honestly. And the owner ruling names the loop as the route by which a structural claim becomes "filed" — so for a large file, that route is not actually available and the ruling's stated escape hatch (declare that you substituted equivalent first-hand verification) is the only one open. Take it explicitly; do not quietly skip the run and do not quietly claim it confirmed you.
- **the check, before you spend 30 minutes waiting:** `wc -c` the file your hypothesis names. Over ~60,000 bytes, expect no verdict. Then either (a) name a NARROWER symbol so its body fits — the bundle's own advice, "put THIS SYMBOL ALONE in next_scope" — or (b) plan on first-hand verification from the start and say so in the bug file. To read the outcome: `SELECT kind, iteration, length(body), source_agent FROM diagnosis_artifacts WHERE correlation_id LIKE '<RUN_CORRELATION_ID prefix>%' ORDER BY created_at;` — the column is **`body`**, not `content`, and the correlation is the **run** one the trigger prints second, not the intake one it prints first.
- **not the same as `bugs_closed/164`:** that was the bundle body cap *breaking* and dropping the rest of scope; this is the cap behaving as designed and the run completing anyway, with the omission recorded truthfully and the conclusion simply never written.
- **source:** 2026-08-01, `bugs_open/170` (`docs024_key_docs_latest/bugfix_170_chrome_pin_eligibility/`), run `ce9bcd92-7be7-4819-bdf8-f8a57622128f`, `WRONG_CALLS.md` same date
- **added:** 2026-08-01, bugfix_170_chrome_pin_eligibility lane

### A `style_collections` chrome PIN is not a per-site assignment, and the predicate that guards it must NOT be the one that guards the pool

- **footprint:** `style_collections` (`header_component_id`, `footer_component_id`, `header_home_component_id`), `platform/orchestration/actions/component_library.go` (`chromePinEligibleSQL`, `chromeEligibleSQL`, `GetChromePinComponent`, `RenderHeader`, `RenderFooter`), `platform/orchestration/actions/link_site_components_action.go`, `platform/orchestration/actions/fork_theme_from_site_action.go`, `platform/orchestration/actions/discovery_checks/check_integrity.go` (`DeactivatedSiteComponentsCheck`)
- **fires when:** you touch chrome eligibility anywhere and reach for `chromeEligibleSQL` as "the" predicate — the whole point of `bugs_closed/118` was that there should be only one, so reusing it reads as exactly right.
- **the trap:** there are **two** questions with two correct answers, and applying the pool predicate to a pin inverts its meaning. `chromeEligibleSQL` carries `forked_from IS NULL` because a fork inherits its parent's `function`, so an active fork of one client's header would otherwise become every other site's default. A **pin** is the supported way to give one site its own chrome, so naming a fork is the intended use. Measured over all four live pins (2026-08-01) the two predicates disagree on **exactly one row** — `leopardessconsulting.co.uk` → `header-leopardess`, active and forked. So copying the pool predicate into the pin path **catches the three wrong pins and deletes the one right one**, and its first live action is destroying a client's bespoke header. This is the same asymmetry as retirement-vs-eligibility one level up (`chrome_selection_test.go` `TestRepointAsksAboutRetirementNotEligibility`): the same columns answer different questions depending on whether you are CHOOSING a default or judging an existing CHOICE.
- **the second trap, and it is the one that bites hardest:** a pin is not only read at render — **`link_site_components` WRITES it** into `site_components.component_id` (via `relinkSiteComponent`, which also sets `rendered_html = NULL, build_status = 'pending'`) whenever it is non-NULL, and `fork_theme_from_site` **copies** it into every new collection. So "the pin is just a display preference" is false in both directions: an unguarded pin reverts the repoint `bugs_closed/166` installed, and propagates itself to collections that did not have it. Fixing the render path alone leaves both.
- **the check, before you change any chrome predicate:** run the two side by side over the live pins and confirm they disagree on the fork row — RUNBOOK §R2 in `bugfix_170_chrome_pin_eligibility/`. If they agree everywhere, you have collapsed the asymmetry. `TestChromePinPredicateDoesNotExcludeForks` fails on `got == chromeEligibleSQL("")`, but only a person reads the reason.
- **and a dormant third column:** `style_collections.header_home_component_id` exists, is populated on **0 of 14** collections, and is read by **no** Go consumer — the `StyleCollection` struct does not model it. Do not "complete" the set by wiring it up; there is no consumer to be consistent with.
- **source:** 2026-08-01, `bugs_open/170` (`docs024_key_docs_latest/bugfix_170_chrome_pin_eligibility/`), concept register **CLC-013**, council `21bac2a2-2b46-4883-894f-19d7ec5e5b45`
- **added:** 2026-08-01, bugfix_170_chrome_pin_eligibility lane

### `llm_call_log.prompt_rendered` contains the *submission you are measuring* — a phrase you quoted while explaining a bug scores as that bug being absent

- **footprint:** `llm_call_log` (`prompt_rendered`, `prompt_template`, `agent_type`, `step_name`), `diagnosis_artifacts` (`kind='council_report'`), `council-gate` and its `review_*` steps, `platform/orchestration/actions/ai_actions.go` (`appendOutputInstructions`, `getJSONOutputInstructions`, `getDefaultOutputInstructions`)
- **fires when:** you verify a prompt-construction change by counting how many logged prompts contain some expected sentence — the obvious and otherwise correct way to check that an instruction block is being appended. It needs no symptom and looks like a clean before/after.
- **the trap:** a council submission's `rationale` and `plan` are **rendered into every seat's prompt**, so any text your submission quotes is *in the corpus you are measuring*. Explaining a bug precisely — which the gate requires — plants the bug's own vocabulary in `prompt_rendered`. Measured 2026-08-02: `prompt_rendered LIKE '%Ensure valid JSON syntax%'` over 9,063 calls to `output_format: json` steps returned **31**, every one of them the author's own two council rounds; the honest figure was **2**, and 9,061 calls had received the wrong instruction block. The false positives land precisely on the rows you care most about, and they inflate the "already working" side, so the error always points toward *do nothing*.
- **the tell:** count per `step_name` and look at the **shape**, not the total. **Exactly N per seat** across a roster (2 per seat, 16 seats) is two runs, not traffic; production never distributes that evenly. A four-month window whose only hits fall on one day, and that day is yours, is the same signal.
- **the check, before you believe any `prompt_rendered` measurement:** (1) detect on a **block header** the appended text owns (`CRITICAL OUTPUT FORMAT - JSON:`), never a prose sentence a rationale would quote in passing; (2) cross-check the same pattern against **`prompt_template`**, which no appended block ever touches — if the template matches 0 times but rendered matches N, the N came from *rendered data*, not from your code; (3) ask what else is interpolated into that prompt before trusting the count.
- **and a join gotcha in the same query:** joining `llm_call_log` to `agent_definitions` on `(agent_type, step_name)` silently drops pre-2026-07-26 seat calls — council seats logged under `agent_type='generic'` until the 07-26 relabel and `'council-gate'` after, stranding 1,798 rows against no definition. Conservative for "was the block appended"; wrong for any absolute historical volume.
- **source:** 2026-08-02, `bugs_closed/119` (`docs024_key_docs_latest/bugfix_119_seat_retry/`, RUNBOOK R11), `WRONG_CALLS.md` same date. Sibling of the same lane's finding that a landmine written *by* a change scores as six council objections *against* it (`016b` §9)
- **added:** 2026-08-02, bugfix_119_seat_retry lane

### `git diff | grep '^-[^-]'` cannot see a deleted markdown BULLET — the "did I remove anything?" check reads clean while lines are going

- **footprint:** `git diff`, `git commit` pathspec review, `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`, `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md`, `bugs_open/`, `bugs_closed/`, `docs/agent_docs/docs026_concept_register/`
- **fires when:** you append to a contended markdown ledger and want to prove your diff only ADDS before committing it by pathspec — the right instinct, because a pathspec commit still takes a **same-file** passenger and these files take appends from several sessions a day.
- **the trap:** a markdown list item is `- **thing**`, so a *removed* one appears in a unified diff as `-- **thing**`. Two leading hyphens, which `^-[^-]` excludes by construction. The same blindness covers `--- ` underlines and any content line starting with `-`.
- **the tell — there isn't one, and that is the point:** `0` is exactly what a clean pure-append diff prints, so a false all-clear is indistinguishable from a true one. It fails OPEN, silent on the case it exists to catch. Hit 2026-08-02 on `docs026_concept_register/register/context-assembly.md`: `--stat` said **1 deletion**, the grep said **0 removals**, and the grep was the one being trusted.
- **the check:** gate on the COUNT, which no content can fool, and only read lines once it is non-zero — `git diff --numstat <file>` prints `added deleted path` and is trivially assertable; to see what went, `git diff <file> | grep '^-' | grep -v '^---'`, because dropping the diff header by prefix is what `[^-]` gets wrong.
- **the same trap in a second spelling, ten minutes later:** grepping the diff for a foreign symbol to prove none rode along — `git diff <f> | grep -c "<symbol>"` — counts **context** lines too, so an untouched neighbour three lines above your append reads as a passenger. It fails the other way (false ALARM, not false all-clear), which is why it is easy to talk yourself out of. Restrict to the side you mean: `git diff <f> | grep '^+' | grep -v '^+++'`. **Generally: every one of these checks must name which SIDE of the diff it is asking about, or it is answering a different question than you asked.**
- **and a THIRD, in `git log` rather than `git diff`: `-S` is occurrence-COUNT based, so it misses any edit that preserves the count.** Asked "has this function signature ever been changed before?", `git log -S "func (v pruneFloorVerdict) Reason("` returned exactly ONE commit — the one that created it — and I had changed that very line an hour earlier. Adding a parameter kept the searched string present, so the count did not move and the pickaxe stayed silent. It fails OPEN and supports a confident *"nothing has ever touched this"*, which is precisely the claim people reach for `-S` to make. **Use `-G <regex>`, which matches the diff TEXT, whenever the question is "who changed this line"; keep `-S` for "who added or removed this string".**
- **source:** 2026-08-02, `bugs_open/165` site B round (`docs024_key_docs_latest/bugfix_165_reconciliation_deletes/`). Nothing had gone wrong yet — the check was being run *because* the work was careful, on exactly the contended ledgers where the passenger risk is real. Sibling of the `MEMORY.md` entry on `${#var}` counting characters where the file counts bytes: both are a measurement whose spelling quietly answers a different question.
- **added:** 2026-08-02, bugfix_165_reconciliation_deletes lane

### A writer of `page_components.rendered_html` that does not repair its links looks exactly like one that cannot introduce a link — and the SET of writers grows while you are not looking

- **footprint:** `page_components` (`rendered_html`), `platform/orchestration/actions/section_editor_actions.go` (`ApplySectionEditAction`, `applyComponentSwap`, `updatePageComponentAfterEdit`, `updatePageComponentSwap`), `create_report_page_action.go`, `create_tool_component_action.go`, `deploy_tool_action.go`, `adopt_verbatim.go`, `rebuild_blog_listing_action.go`, `repairComponentHTMLBeforePersist`, `repairSectionsBeforePersist`, `RepairPageLinks`
- **fires when:** you add or edit anything that persists a component's HTML, or you satisfy yourself that dead-link repair "is handled" because you can see it in the section-save path. It needs no symptom: the unrepaired write succeeds, reports `success:true`, and the 404 only exists for the visitor who clicks.
- **the trap:** the repair is wired at ONE writer (`SavePageSectionsAction`, `bugs_open/079`) and there are **ten** SQL sites that set that column. Nothing about a writer's code says whether it is in the class. `RebuildBlogListingAction` looks unguarded and is safe (its hrefs come from `pages.url`, the same table the repair index is built from — measured: the live `content-listing` template has exactly one anchor, `href="{{.url}}"`). `CreateReportPageAction` looks safe and is NOT: `renderReportSection` runs `html.EscapeString` over ~25 deterministic fields, which reads as total, and then embeds four LLM-authored prose fields with a bare `%s`. **Escaping density is not evidence; find the one `%s` that is not wrapped.**
- **the second trap, which is the durable one:** the writer SET is not stable. `adopt_verbatim.go` became a writer *between* `bugs_open/136` being filed (2026-07-28) and fixed (2026-08-02), and no reader of the bug file would have known. Any conclusion of the form "these are the writers" is true only on the day it was measured.
- **the third trap, inside one action:** `ApplySectionEditAction` had two persist sites — `content_edit` returned its HTML for the caller to write while `component_swap` wrote its own row *inside* `applyComponentSwap`. A guard added before either one is silently bypassed by the other, and both branches return the same `success:true` shape. **Before guarding an action, count its writes, do not count its branches.**
- **the check:** `grep -rlE "(INSERT INTO|UPDATE) +page_components" --include=*.go .` then narrow to the ones that actually SET `rendered_html`, and for each ask the only question that matters — *can an LLM-authored string reach this column through here?* Then confirm mechanically rather than by memory: `scripts/pattern-check.py`'s `check_unrepaired_component_write` fires on any changed `.go` writer that neither calls a repair seam nor is named in `COMPONENT_WRITE_ALLOWED` with a reason. ⚠ **Two writers are deliberately NOT allow-listed** (`create_tool_component_action.go`, `deploy_tool_action.go`) — they are a known open gap, so the check firing on them is TRUE. Adding them to silence it converts a live debt into a false all-clear.
- **and the direction repair fails in:** `RepairPageLinks` UNLINKS a target it cannot find, keeping the text. On a shared tree that is fail-safe for prose and destructive for a listing whose targets legitimately live outside `pages` — so before wiring a new caller, prove where its hrefs COME FROM, not just that they look internal. A no-op guard is not free either: it costs a `pages` query per write and adds a failure surface to a path chosen for having none.
- **source:** 2026-08-02, `bugs_open/136` section-editor slug (`docs024_key_docs_latest/bugfix_136_sibling_link_repair/`), LNK-027. The parent objection came from the council's `bug_historian` seat against `079`'s fix: *"does not establish that SavePageSectionsAction … are the ONLY writers … only that they are the only ones with a save_page_sections STEP NAME"*.
- **added:** 2026-08-02, bugfix_136_sibling_link_repair lane

### Two queries decide "is this work dispatchable" and they DISAGREE — the loser is a dispatch loop that runs, succeeds and does nothing

- **footprint:** `site_work_items` (`depends_on`, `approval_mode`), `agent_definitions` `build-pipeline-trigger` (`find_dispatchable_site`), `build-dispatch-loop` (`load_items`), `platform/orchestration/actions/load_work_item_actions.go` (`LoadWorkItemsAction`), `docs/agent_docs/sql_for_agents/052_build_pipeline_trigger.sql`, `sql_for_agents/284`, `sql_for_agents/285`
- **fires when:** you reason about why a site is or is not being dispatched, change either the site selector or the item loader, or read one of them to learn "what counts as dispatchable". No symptom is needed and none is produced — this is why it is here rather than in 016b §9.
- **the trap:** the SELECTOR (`find_dispatchable_site`, picks the site) tests three predicates; the LOADER (`LoadWorkItemsAction`, picks the items on that site) tests those three **plus two more** — `(COALESCE(approval_mode,'auto')='auto' OR status='approved')` and the `depends_on` unresolved-dependency clause. So the selector can hand the loop a site whose only "eligible" item the loader refuses. The loop loads **zero** items, reports `has_items:false` with **`rows_dropped:0`**, notifies the scheduler and completes **COMPLETED with no error** — having done nothing. No claim is made, so the site is still eligible next tick, and it is picked again. **The queue does not advance and nothing anywhere records a failure.**
- **the tell, and why it is easy to misread:** `rows_dropped:0` alongside `item_count:0` is the signature. It reads as "there was nothing to do" and means "the SQL never matched a row that another query calls eligible". A site that visibly has queued work while its dispatch loops keep completing is the same fact from outside. **Do not conclude anything from `orchestration_states` status here — every one of these runs is COMPLETED.**
- **the check:** run the loader's predicates against the site the selector chose, not the selector's:
  ```sql
  SELECT id, status, approval_mode, depends_on FROM site_work_items
  WHERE site_id = '<the site>' AND status IN ('triaged','approved') AND attempt_count < max_attempts;
  ```
  then test each survivor against `(COALESCE(approval_mode,'auto')='auto' OR status='approved')` and its `depends_on` targets' statuses. Or read the loop directly — `collected_data->'load_items'->>'item_count'` on recent `build-dispatch-loop` rows; a run of `0` on the same site is the defect.
- **the dependency clause has a second trap inside it:** the loader's subquery is **site-scoped** (`WHERE site_id = $1`), so a `depends_on` pointing at another site's item can NEVER resolve and blocks its item for ever. `285` copies that behaviour deliberately — the selector's job is to AGREE with the loader, not to be independently correct — so **fixing the cross-site case means changing the Go loader, and both queries then have to move together.**
- **a dependency in `needs_human_review` is a permanent block, not a slow one.** Nothing automated moves an item to `complete`/`verified` from there, so an item depending on one is undispatchable until a person acts. Measured 2026-08-02: **one** such row (`93f2a3b7`, robot-hands.com) out of 366 selector-eligible items across 17 sites — and because it was at the head of the queue it stalled **the entire fleet**: 0 claims anywhere for 89 minutes (08:06→09:36Z), then again for 68 minutes, while `build-dispatch-loop` ran 16 times and completed cleanly every time.
- **⚠ FAIRNESS ORDERING CONVERTS THIS FROM INTERMITTENT TO PERMANENT — the two changes are only safe together.** Under the old lowest-UUID selector a blocked site held the head only while it happened to sort lowest. Under `284`'s oldest-waiting-first the key is `created_at`, which never changes and only ages, so **an unloadable item, once at the head, is at the head for ever.** If you are reasoning about either change alone you will get its blast radius wrong: `284` without `285` is a permanent fleet stall behind one row.
- **source:** 2026-08-02, found while verifying `bugs_closed/154`'s fix, fixed by `sql_for_agents/285`, proven at the artefact (0 claims in the 68 min before; relojistas 5 / vetcomparison 2 / webdesign 1 in the 8 min after, in exact FIFO order, with `93f2a3b7` still `triaged` as the negative control). Register WDS-002. It retires a wrong call of mine recorded the same week — twice I logged ~90-minute fleet quiet spells as "comparable to known behaviour, not yet outside it"; **that range WAS this mechanism, and a recurring gap matching a known range is not thereby explained.**
- **added:** 2026-08-02, bugfix_154_work_item_routing_columns lane

---

### A migration's verify block made of `SELECT`s cannot stop the `COMMIT` — and `ON_ERROR_STOP=1` will not save you

**footprint:** `docs/agent_docs/sql_for_agents/` · `scripts/migration/run-migrations.sh` · any migration with a "VERIFY BEFORE COMMIT" section

The house style for a migration is pre-flight assertion → snapshot → change →
**verify before commit** → `COMMIT`. The verify step is almost always written as
`SELECT`s with an expectation in a comment above them (`-- expect ZERO rows`,
`-- expect exactly one row`). **Those cannot fail.** psql prints the result and
proceeds to `COMMIT`; a non-empty result set is not an error, so
`-v ON_ERROR_STOP=1` does not trigger. The transaction commits with the defect
your own check just found and printed on screen.

This is not hypothetical: migration 286 (2026-08-02) deleted a workflow step, its
check (iii) correctly reported one surviving reference to that step, and it
committed regardless — shipping a dangling `error_step` to live config. Fixed
forward by 288. The wrong result and the right result look identical here: a
successful-looking run either way, with the finding buried in the middle of the
output above `COMMIT`.

**the check:** where the assertion must actually hold, make it RAISE, not print:

```sql
DO $$
DECLARE bad text;
BEGIN
    SELECT string_agg(...) INTO bad FROM ( <the check> ) t;
    IF bad IS NOT NULL THEN RAISE EXCEPTION 'still wrong: %', bad; END IF;
END $$;
```

Then **prove it can fail** before trusting it: induce the defect inside a
transaction and confirm the block aborts, then `ROLLBACK`. A guard that has only
ever been run against a clean database is indistinguishable from no guard.
Keep the `SELECT`s too — they are useful evidence in the output — but do not let
them carry the weight of a decision.

---

### Deleting a workflow step: the SUCCESS edge is the one you remember, `error_step` is the one that strands a run

**footprint:** `agent_definitions.default_config->workflow->steps` · any migration deleting a step

Repointing a deleted step's predecessor is obvious for `next_step` and easy to
forget for the other three pointer fields — `error_step`, and a conditional's
`config.then_step` / `config.else_step`. An error edge is the worst case: it may
never fire, so the definition carries a stranding bug for weeks with nothing
reporting it, and a green test run proves nothing because the test took the
success path.

**the check:** before believing any step deletion is finished, run the
dangling-reference census over ALL FOUR pointer fields, and fleet-wide rather
than at the agent you edited (measured 2026-08-02: 0 rows fleet-wide, so a hit is
real):

```sql
SELECT ad.type, step.key AS step_name, tgt.target AS points_at
FROM agent_definitions ad,
     jsonb_each(ad.default_config->'workflow'->'steps') AS step,
     LATERAL (VALUES (step.value->>'next_step'), (step.value->>'error_step'),
                     (step.value->'config'->>'then_step'),
                     (step.value->'config'->>'else_step')) AS tgt(target)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND tgt.target IS NOT NULL AND tgt.target <> ''
  AND NOT (ad.default_config #> '{workflow,steps}') ? tgt.target;
```

Reading the step's own JSON is not enough — the reference that strands you lives
in a DIFFERENT step, pointing in. Full form with the enforcing wrapper:
`docs/agent_docs/sql_for_agents/288_repoint_the_error_step_286_left_dangling.sql`.

### `query_database` STRINGIFIES every jsonb column, so a projected JSON value arrives as text and every shape check on it fails quietly

- **footprint:** `platform/orchestration/actions/database_actions.go` (`QueryDatabaseAction`), `platform/orchestration/datahelpers/data_helpers.go` (`ExtractStringListHelper`, `ExtractNestedField`), `platform/orchestration/input_contracts/input_mapping.go` (`ResolveInputMapping`), any workflow step with `"action": "query_database"` projecting `col->'key'` / `jsonb_agg(...)` / `to_jsonb(...)`, `site_work_items.spec`
- **fires when:** you add a jsonb column or expression to a `query_database` step's `SELECT`/`RETURNING` so a later step can read structure out of it — a list, an object, anything that is not a scalar. The natural assumption is that `claimed.my_thing` is then a list or a map in `collected_data`. **It is a string.**
- **the mechanism:** `QueryDatabaseAction` scans every column into `interface{}` and converts any `[]byte` it gets back to `string` — `if b, ok := values[i].([]byte); ok { row[col] = string(b) }`. `database/sql` requires a driver to hand back one of a small set of types, and for a jsonb column pgx hands back bytes, so **every** json/jsonb projection becomes the JSON *text*. Nothing downstream re-types it: `ResolveInputMapping` stores values verbatim (`result[actualDestField] = value`), and the Kafka envelope then marshals a Go string as a JSON string.
- **the tell: there isn't one, and every observable says it worked.** The column is present. `input_mapping` resolves it and logs "Resolved input mapping". The child receives the key. Only the consumer disagrees, and it does so by returning an empty result that is indistinguishable from "the caller supplied nothing" — which is how `bugs_open/174` cost three lanes' diagnoses their chosen scope with no error anywhere. A type assertion (`val.([]interface{})`), a `len() == 0` guard, and a `ExtractNestedField(x, "a.b")` traversal all fail the same silent way.
- **the check, before you rely on a projected jsonb value:** do not reason about it — read it back off a real run. `SELECT jsonb_typeof(collected_data->'<output_field>'-><'key'>) FROM orchestration_states WHERE ...` tells you what the DB holds, which is NOT the question; the question is what the Go map holds, so assert on the **effect** (did the consumer use the value?) rather than on the field being present. If the consumer takes a list, `datahelpers.ExtractStringListHelper` handles the JSON-text form since 2026-08-02 (`bugs_open/174`) — for any other shape you must decode it yourself, e.g. through `datahelpers.SafeUnmarshal`.
- **and the reason this is a landmine rather than a bug to fix:** fixing `QueryDatabaseAction` to decode json/jsonb columns would change the shape every consumer receives. **Measured 2026-08-02: exactly ONE live `query_database` step projects a JSON-TYPED value** (`diagnose-dispatch-loop.claim_item`, added by 174's own fix), so the blast radius today is nil — but the measurement is the point. A loose grep for `->|jsonb_|json_` over the query text returns **14** steps, and 13 of those are `->>` **text casts** (whose consumers already expect text) or arrows inside **WHERE predicates** (not output at all). 174's council submission quoted the 14 as the blast radius; that figure was wrong. **Classify by whether the arrow lands in a projection, not by whether the query mentions json.**
- **source:** 2026-08-02, `bugs_open/174` (`docs024_key_docs_latest/bugfix_174_seed_scope_relay/`). The `090_TRIGGER_needs_diagnosis_v1.sh:345` comment had recorded the consequence — *"ExtractStringListHelper takes []interface{} or []string only; a bare "a,b" string yields nil and the seed is ignored"* — the whole time, next to the code that wrote the value. Raised as a gating-adjacent objection by the council's `bug_historian` seat on corr `081d98b3`, which required this entry as the minimum acceptable mitigation for the deferred root-cause fix.
- **added:** 2026-08-02, bugfix_174 lane

### `input_mapping` is an ALLOW-LIST and so is a claim query's RETURNING — a dispatcher has TWO gates, and fixing the one you can see leaves the key dropped

- **footprint:** `agent_definitions` rows `diagnose-dispatch-loop` / `report-dispatch-loop` / `build-pipeline-trigger`, their `claim_item`/`call_handler` steps, `platform/orchestration/input_contracts/input_mapping.go` (`ResolveInputMapping`), `agent_definitions.input_contract`, `cmd/config-key-audit/relaygaps.go`, `scripts/audit-relay-gaps.sh`
- **fires when:** an optional field an operator sets is not reaching the agent that consumes it, on any pipeline where a dispatch loop stands in front of a handler. You find the `call_agent` `input_mapping`, see the key missing, and add it. **That is half the fix, and the other half is invisible from where you are standing.**
- **the tell:** none — the added mapping key resolves to nothing and, being optional (`key?`), `ResolveInputMapping` skips it at **Info** level and continues. You get a successful run, a normal-looking result, and the same missing field. `bugs_open/174`'s own filing proposed exactly this fix, having read the mapping and not the claim query.
- **the check:** read **both** allow-lists and the callee's declared contract, which is the authority neither of them is checked against:
  ```sql
  SELECT substring(default_config #>> '{workflow,steps,claim_item,config,query}'
         from position('RETURNING' in default_config #>> '{workflow,steps,claim_item,config,query}')) AS projected,
         default_config #>> '{workflow,steps,call_handler,config,input_mapping}' AS forwarded
    FROM agent_definitions WHERE type='<the dispatch loop>'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  -- then the authority both must satisfy:
  SELECT input_contract FROM agent_definitions WHERE type='<the handler>' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
  Or run it fleet-wide: `./scripts/audit-relay-gaps.sh`.
- **the second trap, for anyone WRITING a check for this class:** the obvious general rule — "every `call_agent` must forward every key its callee declares" — is **not sound**, and measuring it is how you find that out. Live on 2026-08-02 it gave **31** findings from 75 resolvable call sites, and the spot-checked ones were legitimate (`webdesign-agent` carries an `else_step: load_site_context` and loads the key it was not passed). Tightening to "the callee also READS `input_data.<key>`" gave 3 and still could not separate "the caller dropped it" from "the caller never had it". **Worse, both versions were blind to 174 itself**, because `call_handler` resolves its callee at runtime via `agent_type_field: claimed.handler_agent` and a static resolver skips it. A dispatcher is the one case where the question is answerable — its envelope IS the work item spec — which is why the shipped check is a declared registry over 3 relays rather than a fleet-wide rule.
- **source:** 2026-08-02, `bugs_open/174`. Same mechanism as the `--fidelity` landmine (`input_mapping` dropped `fidelity` until migration 274), one hop further back: 274 only had to fix the mapping because that path had no claim-query projection in front of it.
- **added:** 2026-08-02, bugfix_174 lane

---

### Cloudflare answers `Python-urllib` with **403**, so a Python health check reports a healthy site as broken

**footprint:** `urllib.request` against any Cloudflare-fronted site in this estate · any Python asset/link/health checker

Measured 2026-08-02 against `https://loancalculator.co.uk/assets/css/style.css`:

| client | result |
|---|---|
| `curl` GET, HEAD, with or without a UA | **200** |
| `urllib` GET and HEAD, default UA | **403** |
| `urllib` GET with a browser UA | **200** |

So it is the **agent string**, not the method and not the asset. A checker
written the obvious way (`urllib.request.urlopen`, no headers) marks every asset
on a perfectly healthy site as unreachable. That is worse than having no checker:
the output is indistinguishable from the real 404s it was written to catch, and
the honest human response to "everything is broken" is to stop believing the
check. It fired on the first run of `decompose/load_chrome.py` and reported the
two assets that had just been confirmed 200 by curl.

**the check:** set a User-Agent on every `urllib` request, and prove the checker
can distinguish a good asset from a bad one before trusting a clean run —
`/assets/css/style.css` must be 200 and `/assets/css/styles.css` (the plural
typo that was live in this site's chrome) must be 404, from the SAME code path.
A checker that returns the same status for both has told you nothing. Verifying
with `curl` and then implementing with `urllib` is exactly how this is missed.

---

### `site_components` having rows does NOT mean the chrome works — and on a verbatim site nothing will ever tell you

**footprint:** `site_components` · `rerender_single_page_action.go (loadVerbatimPageHTML` · any `rebuild_policy='owned'` adopted site

`assemblePage` reads chrome from `site_components`, but it is never reached for a
page that ships verbatim — `loadVerbatimPageHTML` returns first. So on an adopted
site the chrome can be written, be completely broken, have a full-site rerender
run against it, and report 27 successes while changing nothing.

Exactly that happened to loancalculator.co.uk: rows written 2026-08-01 08:02,
27 `page_rerender` items five seconds later, all `complete`, site byte-identical
— and the chrome pointed at `/assets/css/styles.css` (**404**; the real file is
`style.css`), carried a nav `<ul>` with **zero** `<li>`, and linked a 404 favicon
and a 404 `og:image`. The first page decomposed would have shipped unstyled and
unnavigable.

**the check:** never ask "are there rows". Ask whether the chrome RESOLVES:
fetch every `href`/`src` it references and require 200 (with a UA — see the
landmine above), count the nav links and join them against `pages.url`, and
confirm the head still contains the two LITERAL strings assembly rewrites by
exact match — `<title></title>` and `content=""`. Reorder the head so another
tag holds the first `content=""` and the page's meta description silently lands
in the wrong tag. Implemented as a refusal in
`loancalculator_couk/decompose/load_chrome.py`.

---

### A verbatim page is defined by its ROW COUNT, so ADDING a section silently switches the page to assembly

**footprint:** `page_components` on any `rebuild_policy='owned'` site · `loadVerbatimPageHTML`

The condition is `rebuild_policy='owned'` AND **exactly one** component row AND
that row's `content_data.deploy_mode='verbatim'`. There is no flag to clear.

So attaching a second component to a verbatim page does not "add a section to the
page" — it flips the whole page from shipping its stored bytes to being
assembled, and the old row, which holds a COMPLETE document with its own
`<!DOCTYPE>`, `<head>` and `<body>`, becomes one of the sections. The result is a
document nested inside a document. The action logs a warning and proceeds.

**the check:** to decompose, REPLACE the verbatim row in the same transaction —
`DELETE FROM page_components WHERE page_id=…` then insert the new rows — never
insert alongside. Before believing a page is still verbatim, count:
`SELECT count(*), count(*) FILTER (WHERE content_data->>'deploy_mode'='verbatim')
FROM page_components WHERE page_id='…';` — anything other than `1|1` means the
page assembles, whatever the row's own `deploy_mode` still says.

---

### A splitting/extraction rule that reads only INLINE scripts is blind on any page whose logic is in a `<script src>`

**footprint:** `loancalculator_couk/decompose_prover.py (script_ids` · any tool/widget extraction keyed on `getElementById`

The honest definition of "what a widget needs" is the set of element ids its
scripts address. The trap is the word *its*: read only the inline `<script>`
blocks and a page that keeps its arithmetic in a shared `.js` file has script
targets the rule cannot see, and they get classified as ordinary prose — free for
a writer agent to rewrite or delete.

On `/index.html` the whole of `calculateLoan()` lives in `/assets/js/global.js`,
so `#monthly-display`, `#total-interest` and `#total-cost` — the calculator's
entire results box — decomposed as editable prose. **The rule's own safety proof
passed and could not have failed**: it asks whether every id an INLINE script
addresses travels with the tool, and that page's inline script addresses only the
three inputs, all of which did travel. The proof was narrower than the risk.

**the check:** fold every `<script src>` the page loads into the same id set
before splitting, and make an unreadable referenced script a HARD FAILURE rather
than a warning — proceeding with the ids you could see is the silent-success
shape that causes this. Then assert the narrower property that actually matters:
no script-addressed id may land in a block classified as prose
(`stranded_script_targets` in `decompose_pages.py`).

### `orchestration_states` keeps terminal rows ~24 HOURS — and `min(created_at)` says 20 days, because the statuses it reaps are not the ones that set the floor

- **footprint:** `orchestration_states`, `scheduled_tasks` (`database-cleanup`), any "has agent X ever run / 0 orchestrations / never dispatched" claim, `owner_agent_type`, `collected_data`
- **fires when:** you establish that something never happens by counting `orchestration_states` — "this agent has 0 orchestrations", "the action has never executed", "no run has ever recorded field F" — and then bound it by reading the table's own oldest row.
- **the trap, and it is TWO traps stacked:** (1) **COMPLETED and FAILED rows are reaped after ~24h.** Measured 2026-08-02: 2,504 COMPLETED rows whose oldest is **24.7h** old, FAILED oldest **25.4h**. (2) **`min(created_at)` over the whole table reads 2026-07-13 — twenty days — because `CANCELLED` (24 rows), `RUNNING` (4) and `INITIALIZED` (2) are NOT reaped.** A handful of stragglers in statuses nobody cleans up set a "retention floor" twenty times longer than the real one, and it is the *successful* runs — the ones your census is about — that vanish.
- **the tell:** none at query time. I watched a specific COMPLETED row (`dcf88c1c…`, `site-adoption-agent`, created 2026-08-01 09:04) that I had read and quoted at ~09:40 **disappear by ~10:40 the same morning**. The table had grown 2,454 → 2,546 in that hour while its oldest row never moved — growth and reaping at once, which is exactly what a stable-looking table does.
- **the check, and do it BEFORE you write "never":** never bound retention with a whole-table `min()`. Bound it **per status**, because the reaper is status-selective: `SELECT status, count(*), min(created_at), round(extract(epoch from (now()-min(created_at)))/3600,1) AS oldest_hours FROM orchestration_states GROUP BY status;`. Then, if the claim is durable, **re-source it from a table with no retention job** — `site_specs` goes back to 2026-02-25 (1,874 rows, 36 sites), `site_work_items` and the target table itself persist. A 5-month `site_specs` census answers "which builder was ever chosen" flatly; a 24-hour orchestration census cannot.
- **why it bites hardest:** an absence claim is exactly what people reach for this table to prove, and the failure is silent and *directional* — short retention can only ever manufacture absence, never presence. So it fails toward the conclusion you were already testing for. Hit 2026-08-02 while retiring an agent: "0 orchestrations in the retained window" was written into a bug file, an 016b index row and the concept register meaning *~20 days*, when it meant *~24 hours*. The conclusion happened to survive on other evidence; the stated reason did not.
- **`database-cleanup`** (`scheduled_tasks`, hourly, enabled) is the likely reaper — named here as the place to look, **not** as a verified mechanism: no Go implementation of it was found by grep, so treat the *behaviour* above as measured and the *cause* as unconfirmed.
- **source:** 2026-08-02, `bugs_closed/165` / `bugs_closed/092`, retiring `multipage-website-builder`. Sibling of the `bugfix 003` note that history tables are retention-clocked — record a RATE; this is the same family with a specific, much shorter number and a booby-trapped floor.
- **added:** 2026-08-02, bugfix_165_reconciliation_deletes lane

---

### Two migrations can carry the SAME number — creating the file does not reserve it, and "migration 286" then names two unrelated changes

**footprint:** `docs/agent_docs/sql_for_agents/` · `schema_migrations` · any commit message or doc citing a migration by number

The advice "re-check the highest number immediately before writing the file" is
necessary and **not sufficient**. Measured 2026-08-02:
`286_one_owner_for_the_site_wide_promoter.sql` was created at 11:21:29 — the
number taken by `ls`, then claimed with `touch` in the same command — and
`286_triage_robot_hands_blocked_crosslink.sql` was created at **11:38:15**, 17
minutes later, by another session, with the first file already sitting on disk.
Claiming the name reserves nothing if the other session's number came from an
earlier read, a plan written an hour ago, or a hardcoded value.

Nothing breaks: `schema_migrations` keys on **filename**, so both record and both
apply independently. What breaks is the assumption that a number identifies a
change. It is the same trap `CLAUDE.md` already documents for `/bugs_open/`
numbers ("several numbers name two unrelated cases … resolve by slug").

**the check:** cite migrations by **slug**, never by bare number —
`286_one_owner_for_the_site_wide_promoter`, not "286". Before acting on a
reference to "migration NNN", run `ls docs/agent_docs/sql_for_agents/NNN_*` and
expect more than one line. And when your own file's number matters to a later
reader (a rollback sidecar, a correction migration), put the **slug** in that
file's header too — `288_repoint_the_error_step_286_left_dangling.sql` is
ambiguous by exactly this trap, and says which 286 it means in its first
paragraph for that reason.

### A component template's `{{else}}` can INVENT a business fact, and every render then "succeeds" — the schema already forbids it and nothing enforces the schema

**footprint:** `content_components.html_template` · `input_schema.fields.*.on_missing` · `scripts/check_placeholder_fallbacks.py` · any component seed or `store_generated_component` write

A template fallback is normally furniture — `{{else}}Read more{{end}}`,
`{{else}}Get Started{{end}}` — and the library is full of them, correctly. But the
same construct can substitute a **fact about the business**, and then the platform
publishes an unsourced claim with no error, no failing status and no truncation:

```
{{if .phone}}…{{else}}+1234567890{{end}}                        <- a tel: link that is not a phone number
{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}
```

`bugs_open/140`: **8 live sites served those invented hours**, and vetcomparison.uk
served `tel:+1234567890` on the wire. It rendered in the same style as the real
details, so nothing distinguishes the invented line from the true one by eye.

**The trap that makes this landmine rather than a bug report:** the component's own
`input_schema` ALREADY says `"on_missing": "skip_field"` for those fields, so a
session reading the schema concludes the behaviour is correct and never opens the
template. **The schema is documentation; nothing enforces it at render time.**
`compute_component_quality.go`'s `scoreComponent` does not look at `{{else}}`
contents, and `component_write_guard.go` is COMPARATIVE by design (is this
replacement worse than what it replaces) so a birth write carrying a fabricated
fallback passes it cleanly.

**the check:** before trusting that a component degrades safely, run
`python3 scripts/check_placeholder_fallbacks.py` (advisory; 0 clean / 1 findings /
2 no DB) — it reads the LIVE library and separates a fact default from a label
default, because `content_components` exists only in the database and no
file-diff linter can see it. Then read the `{{if}}` branch of anything it reports:
a fallback whose text also appears in the branch it replaces invents nothing (a
builder attribution rendered as link-or-text is the common shape, and was this
script's first false positive).

**And do NOT "improve" `check_placeholder_contact` into the roster-free version.**
The obvious upgrade — flag any rendered contact fact absent from the component's
`content_data` — is UNSOUND: `RenderContext` carries top-level
`Email string \`json:"email"\`` and `Phone string \`json:"phone"\``, and
`contextToInterfaceMap` derives the template contract from those json tags, so a
component can legitimately render a phone its `content_data` does not hold because
the value arrived from site identity. `idea.uk` is exactly that shape and would be
flagged as fabricated.

**Reading the STORED render will also mislead you after a fix.** `page_components.
rendered_html` still contains the fabrication until the page rerenders — fixing the
template does not regenerate anything. Judge the template with a query against
`content_components`, and the artefact separately.

### An offline link census built from `pages` MUST use `linkablePageStatusPredicate`, or an ARCHIVED page turns a correct rewrite into an apparent false positive
- **footprint:** `platform/orchestration/datahelpers/content_data_links.go`, `platform/orchestration/datahelpers/link_repair.go`, `loadValidPagePaths`, `linkablePageStatusPredicate`, `pages`
- **fires when:** you re-measure the fleet's dead internal links outside the chassis — a SQL census, an offline Go harness, or any "let me check this fix was right" query that builds its own page set.
- **the tell:** a rewrite that looks WRONG. `robot-hands.com` has `/learning-center.html` (active) **and** `/learning-center/index.html` (**archived**). `NormalizePagePath` maps the second to `/learning-center` — so an index that includes archived pages makes `/learning-center` *resolve*, the finding disappears, and the live fix looks like it is rewriting a link that was already fine. The reverse mistake is worse: a census that is stricter than the runtime invents phantoms that the running code never reports.
- **the check:** copy the predicate, never retype it — `status NOT IN ('deleted', 'archived')`, the shared constant at `prepare_link_context_action.go:54`, which `loadValidPagePaths` uses and every link consumer inherits. Then confirm your harness and the binary agree on ONE case before trusting the totals. **And prefer running the shipping function over a dump to re-implementing it in SQL** — the SQL has to hand-copy `NormalizePagePath` and `ClassifyLinkScope`, and the two answers differed (51 vs 52) the first time they were compared on 2026-08-02.
- **source:** `bugs_open/097` / `docs024_key_docs_latest/bugfix_097_content_data_links/RUNBOOK` R1–R2; concept register LNK-028
- **added:** 2026-08-02, bugfix_097_content_data_links

### `CONTENT_LINK_REPAIR_DETAIL` does NOT include the content_data audit — three codes now answer three different questions
- **footprint:** `agent_error_log`, `CONTENT_LINK_REPAIR_DETAIL`, `CONTENT_LINK_REPAIR_SKIPPED`, `CONTENT_DATA_LINK_AUDIT`, `writeLinkRepairLog`, `writeContentDataLinkLog`
- **fires when:** you ask "has the link machinery seen this page?" or "is the repair still running?" with a query written before 2026-08-02.
- **the tell:** a confident empty result. The old query returns exactly what it always returned — which is now a *subset* of what the platform knows, and nothing in the output says so. A page whose only defect is a phantom link stored in `content_data` produces **no** `CONTENT_LINK_REPAIR_DETAIL` row at all.
- **the check:** query all three codes, and read the split as three answers, not one: `CONTENT_LINK_REPAIR_DETAIL` = what the pass CHANGED in the markup that shipped · `CONTENT_LINK_REPAIR_SKIPPED` = a pass that declined because the page index was untrustworthy · `CONTENT_DATA_LINK_AUDIT` = what the page's SOURCE still holds (component, field path like `cards[2].link_url`, href, and which arm). The codes are deliberately distinct for the reason `linkRepairErrorCode` is distinct from `validationDetailErrorCode`: merging them would silently change the population every existing query counts.
- **source:** `bugs_open/097`; concept register LNK-028, LNK-023/024/027
- **added:** 2026-08-02, bugfix_097_content_data_links

### `grep -c` exits 1 and prints NOTHING on a zero count — so the NEGATIVE control in a pod-verification loop captures an empty string, not `0`

- **footprint:** `kubectl exec … -- sh -c 'strings /app/agent-chassis | grep -c "…"'`, any post-roll verification loop, `bugs_open/153`, `docs024_key_docs_latest/*/RUNBOOK*` pod-grep recipes
- **fires when:** you follow the fleet's standing verification rule — prove a deploy with a positive control **and a negative one** (a string your change REMOVED, expected to read 0) — and you collect the counts in a shell loop, e.g. `N=$(kubectl exec $POD -- sh -c '… | grep -c "old string"' 2>/dev/null | tail -1)`.
- **the tell:** **there isn't one, and that is the whole problem.** `grep -c` writing `0` also *exits 1*, and with `2>/dev/null` the command substitution yields an **empty string**. Printed in a table it renders as a blank cell or as `negative=` — which reads as "the exec failed / no data for this pod", the same as a genuinely broken command, a `NotFound` pod, or an expired token. The one control whose zero value is your proof is the one control whose zero value is invisible. Observed 2026-08-02 verifying `bugs_closed/168`: `negative(expect 0)= command terminated with exit code 1`.
- **the check:** default the capture — `N=${N:-0}` — or use `grep -c … || true`, or invert to `grep -q … && echo 1 || echo 0`. Then read the table and confirm the negative control shows a literal `0`, not a blank. **And run the positive control in the same loop**: if it is also blank, your exec is broken and neither number means anything.
- **also, on the same recipe:** verify **after** `kubectl rollout status`, never during. A mid-roll enumeration returns pod names that are already gone — 3 of 4 came back `NotFound` while one answered, and one replica answering during a roll is not "both replicas verified" (`bugs_open/153` requires every replica, same exec).
- **why it is a landmine:** the negative control exists *because* a positive one cannot prove your binary shipped; this trap silently disarms the half that does the work, and it disarms it into a value that looks like an infrastructure problem rather than a result. You will re-run the command, get the same blank, and conclude the exec is flaky.
- **source:** 2026-08-02, `bugs_closed/168` verification on chassis `v1.0.1229`, both replicas
- **added:** 2026-08-02, bugfix_168 lane

### `RepairPageLinks` cannot tell an anchor from JAVASCRIPT THAT BUILDS ONE — it deletes the link from the program, and the output is still valid JS

- **footprint:** `platform/orchestration/datahelpers/link_repair.go` (`RepairPageLinks`, `repairAnchorRe`, `LinkScopeEmpty`), `repairOutboundPageLinks`, `repairSectionsBeforePersist`, `repairComponentHTMLBeforePersist`, `create_tool_component_action.go`, `deploy_tool_action.go`, any `page_components.rendered_html` containing a `<script>` block
- **fires when:** you wire the link repair into a new writer — especially a TOOL writer, which is exactly where markup is assembled in JS — or you audit stored HTML for dead links with a regex over `href="…"`. No symptom precedes it: the write succeeds, the page reads correctly, and only a click fails.
- **the trap:** `repairAnchorRe` is byte-level over the whole document and does not skip `<script>`. Against `'<a href="' + q.link + '" target="_blank">See guide section</a>'` the href capture `[^"']*` cannot cross the `'` that follows `href="`, so it captures the **empty string**, `ClassifyLinkScope` returns `LinkScopeEmpty`, and the unlink arm strips the `<a>` while keeping the text. **The link is deleted from the source of a working runtime feature.** Verified 2026-08-02 by running the shipping function over the exact live bytes — `repairs=1 action=unlink href=""`, output `' See guide section.</p>'`.
- **the tell — there is none, in three directions at once:** the output is still valid JavaScript so nothing throws; unlink keeps the anchor text so the prose still reads; and `success:true` is reported. A `data-runtime-fill` marker would exempt it, but the measured instance has no marker on the component OR anywhere on its page.
- **and the template-literal form is WORSE, not better:** `href="${url}"` is a non-empty href, so it takes the `LinkScopePage` arm instead, fails the index lookup and is unlinked as a *phantom* — same destruction, different arm, and invisible to a grep written for the `' +` spelling.
- **the check, before wiring the repair into any writer:** ask whether that writer's markup can contain a `<script>`, and if so run the real function over a real sample before shipping — three lines in a throwaway `_test.go` against `datahelpers`, which is decisive where reading the regex is not. For an AUDIT, remember the mirror-image version of the same error: a regex over `href="…"` counts href-shaped byte sequences, not links, so any such census is an **upper bound** (mine included one JS fragment in 35).
- **source:** 2026-08-02, `bugs_open/180`, found by the `bugfix_136_sibling_link_repair` lane while re-running its own census after the v1.0.1229 roll. It reordered that lane's stated next step: wiring the seam into the tool-markup writers must wait for 180, or it will delete working buttons from tools
- **added:** 2026-08-02, bugfix_136_sibling_link_repair lane

---

### Decomposing an adopted site with NO `site_components` head ships every page unstyled — the fallback hardcodes the wrong stylesheet name

**footprint:** `rerender_single_page_action.go (buildDefaultHead` · `site_components` on any `fidelity=locked` adopted site · `loanandmortgagecalculator.co.uk` · `loancash.co.uk`

`assemblePage` falls back to `buildDefaultHead` when a site has no stored head, and
that fallback is five lines ending in:

```go
<link rel="stylesheet" href="/assets/css/styles.css">
```

**`styles.css`, plural.** That is right for a platform-BUILT site — measured
2026-08-02, it returns 200 on 15 of 16 deployed sites. It is wrong for an ADOPTED
one, which brought its own asset names with it. So the fallback is most likely to
fire exactly where it is most likely to be wrong: an adopted site is the kind that
has no chrome rows.

Measured the same day, both remaining verbatim-adopted sites:

| site | verbatim rows | `site_components` head | `style.css` | `styles.css` |
|---|---|---|---|---|
| loanandmortgagecalculator.co.uk | 41 | **none** | 200 | **404** |
| loancash.co.uk | 18 | **none** | 200 | **404** |

Both link `style.css` (singular) from every live page. So the first page either one
decomposes gets a head pointing at a 404, **plus no header and no footer at all**
(`resolveComponent` returns empty strings), and it will do so silently — the render
succeeds, the deploy succeeds, and the page is unstyled and unnavigable.

None of this is visible beforehand, because assembly is never reached while a page
ships verbatim (see the sibling landmine on `site_components` rows not meaning
working chrome).

**the check:** before decomposing ANY adopted page, confirm the site has all three
chrome rows AND that every asset they reference resolves — do not accept the
fallback:
```sql
SELECT slot_name, octet_length(rendered_html) FROM site_components
WHERE site_id='<id>' ORDER BY slot_name;   -- expect head, header, footer
```
then fetch every `href`/`src` in them and require 200 **with a browser
User-Agent** (Cloudflare 403s `Python-urllib`). Working implementation, including
the refusal: `loancalculator_couk/decompose/load_chrome.py`. Take the head from
what the site's OWN pages link, not from the platform default — on
loancalculator.co.uk the one-letter difference was the whole of the failure.

---

### A structural parity test cannot see an ENCODING — two JSON writers agree while the bytes diverge

**footprint:** `platform/orchestration/actions/provocation_feed_action.go` ·
`provocation_feed_parity_test.go` · `encoding/json` · any Go action that writes a
JSON artefact previously written by another tool

Go's `encoding/json` **HTML-escapes `<`, `>` and `&` by default**, and marshals a
map with its **keys sorted**. Neither is visible to a test that unmarshals before
comparing — `<em>` and its escaped form are the same value once parsed, and key
order does not survive a decode at all. So a golden-fixture parity test can pass,
indefinitely, while the artefact you actually publish differs from the oracle on
every line.

Measured 2026-08-02: the first live commit of `provocations.json` by
`render_provocation_feed` was **+119 / −119 lines** — the entire file rewritten to
publish one timestamp. The parity test was green throughout.

**the check:** when a Go action takes over writing a file some other tool used to
write, diff the **bytes** of the first artefact it produces against the previous
version, before you trust any test:

```bash
gh api "repos/<o>/<r>/commits/<sha>" --jq '.files[] | "\(.filename) +\(.additions) -\(.deletions)"'
```

A `+N/−N` where N is the whole file means **the writer changed, not the content**.
Then fix what matters and decide the rest explicitly: `enc.SetEscapeHTML(false)`
restores literal markup; key order needs a hand-maintained field list per object,
which is usually a worse trade than one messy commit per change of writer — but
decide it, do not inherit it. Note the ping-pong if both writers are retained:
each rewrites the whole file after the other.

**Related, and it fires while you write the fix:** a literal backslash-u-0-0-3-c
typed into Go source is decoded by the compiler inside a double-quoted string,
and by several tool channels in transit — so an assertion written to prove that
sequence ABSENT silently becomes a search for the character it decodes to, which
is present. Build the needle by concatenation (`` `\` + "u003c" ``) and spell it
in words in prose. It fired four times in one session, including inside the
paragraph warning about it.

### A migration number is not yours because you named a file — it is yours when the LEDGER says so, and applying by hand never claims it

- **footprint:** `docs/agent_docs/sql_for_agents/NNN_*.sql` · `schema_migrations` ·
  `scripts/migration/run-migrations.sh` · `--record-only` · `kubectl exec … psql -f`
- **fires when:** you pick the next number with `ls docs/agent_docs/sql_for_agents/ | sort -n | tail`,
  write `NNN_your_thing.sql`, and apply it the fast way — piping it straight into
  `psql` — because `--apply` takes **every** pending file and other threads' halves
  are sitting in there (67 of them, measured 2026-07-31). Both halves of that are
  the house-recommended reflex, and together they lose you the number.
- **why the wrong result looks exactly like the right one:** your SQL committed.
  The guards passed. The config is correct. Nothing anywhere says *taken*, because
  the only record of a number being claimed is a `schema_migrations` row keyed on
  **filename** — and a hand `psql` run writes none. The directory listing you read
  is shared mutable state with no reservation step, so it answers "what existed a
  moment ago", never "what is free". Measured 2026-08-02: I created `291_` at
  18:58, another session created a **different** `291_` at 19:03 and put it through
  the runner at 19:04:24. Theirs is recorded; mine was pending the whole time. It
  is not a race I lost by five minutes — **I never entered it.**
- **the check, before you type the number:** ask the ledger, not the directory, and
  ask it for the numbers that are *spoken for* rather than the ones on disk —
  `SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 5;` — then
  reconcile against `ls`. A number present on disk but absent from the ledger is
  **unclaimed**, and that includes yours.
- **the check, after you apply by hand:** record it the same minute, or it stays
  pending for ever and the next `--apply` replays it:
  `./scripts/migration/run-migrations.sh --record-only NNN_your_file.sql --note '<what you verified>'`.
  The `--note` is the point — it is the only place the by-hand provenance survives.
  A dry run (`./scripts/migration/run-migrations.sh`, no flags) lists your file as
  pending and probes it; "pending" there means *either* not applied *or* applied
  and never recorded, and the runner's own footer says so.
- **do not renumber a file that is already recorded, and do not rewrite transcripts
  when you renumber one that is not.** Renumbering an unrecorded file is free
  (`mv` + fix its internal `RAISE`/`snapshot_agent` strings — grep the number, there
  were five in mine). But error output you already pasted into NOTES was printed
  under the old number: leave it, and say why. A tidier record that disagrees with
  what the system actually said is worse than a visible discrepancy.
- **source:** 2026-08-02, brochure lane, arming TL-035 (`sql_for_agents/292`).
  Nothing needed undoing — the write was idempotent — which is precisely why this
  is a landmine and not an incident: the cost is silent and lands on whoever reads
  two files with the same number later.
- **added:** 2026-08-02, brochure_component_library lane

### A `page_rerender` regenerates SECTIONS only when `spec.reason` is set — the default re-staples the page from STORED html, so a template/config fix never reaches the artefact

**footprint:** `site_work_items` `item_type='page_rerender'` · `spec.reason` · `check_rerender_mode` on the `page-rerender` agent · `create_rerender_items_action.go` · `rerender_single_page` · `rerender_page_sections`

The live `page-rerender` agent routes on **`spec.reason` alone**
(`check_rerender_mode`):

```
condition: spec.reason == 'image_landed' OR 'section_data_resolved' OR 'cta_links_stale'
then_step: rerender_sections -> rerender_page_sections   REGENERATES each section from its template
else_step: render_page       -> rerender_single_page     re-staples the page from section HTML ALREADY STORED
```

**A reason-less item is the fleet's default and it faithfully preserves whatever
is wrong.** `tool_render_path_test.go:127` names it exactly: *"a reason-less item
takes render_page and **deploys stale HTML**"*.

**What this costs you, measured (`bugs_open/140`, 2026-08-02).** A shared
component template was fixed at source, live immediately (DB config). The
`page_rerender` backlog then drained **294 → 0**; **six affected pages ran a
COMPLETED `page_rerender` and not one section updated** — one still stamped
`2026-05-02`. The fix was correct and live the whole time; nothing reached the
artefact. Seven items carrying `reason='section_data_resolved'` repaired all
seven in twenty minutes.

**the check:** after a template, component or render-config fix, do NOT assume
queued rerenders will propagate it. Read the pending items' reasons —
`SELECT spec->>'reason', count(*) FROM site_work_items WHERE item_type='page_rerender' AND status NOT IN ('complete','cancelled') GROUP BY 1;`
— and if they are NULL, the queue will complete without repairing anything. Queue
your own with `reason='section_data_resolved'`, which is the narrowest repair
available (blast radius: one page's sections).

**Do NOT read `create_rerender_items_action.go:219`'s
`scoped := (reason == … ) && componentIDStr != ""` as the consumer's rule.** It is
a PRODUCER-side gate deciding when *that creator* stamps a reason. The agent
itself requires only the reason — an item with a reason and no `component_id`
still takes the section-regenerating branch, and believing otherwise will stop you
queuing a repair you can in fact queue.

**And `page_component_history` will not tell you a rerender happened.** Its only
observed `source` fleet-wide is `save_page_sections_overwrite` (the content-writer
path), so a section regenerated by `rerender_page_sections` leaves no history row.
Use it to learn what WROTE a section, never to conclude nothing did — and never
infer the writer from whatever work item completed nearby, which is how a
content-writer repair gets mis-credited to a rerender.

### Any regex you write against HTML will also match inside `<script>`/`<style>`/`<textarea>`/comments — and if the regex REWRITES, that is silent content destruction that still parses

**footprint:** `platform/orchestration/datahelpers/link_repair.go (repairAnchorRe` · `platform/orchestration/datahelpers/links.go (anchorRe` · `platform/orchestration/actions/drop_dead_url_controls.go (deadAnchorRe` · `platform/orchestration/datahelpers/markup_spans.go` · `page_components.rendered_html`

Every anchor regex in this estate is a byte-level match over the whole input, and
each one's header carefully explains its `\b` and its non-greedy tail while none of
them can explain this. Stored `rendered_html` contains **JavaScript that builds
markup**, so:

```
<script>h = '<p>' + t + ' <a href="' + q.link + '">See guide</a>.</p>';</script>
```

reads as an anchor whose href is **EMPTY** — `[^"']*` cannot cross the `'` that
immediately follows `href="` — and a repair pass deletes the anchor from the
program. The output is still valid JavaScript and the prose still reads, so
nothing throws, no test fails, and the visitor simply cannot click.

**the check:** if your pattern REWRITES markup, go through
`datahelpers.MarkupMatches` / `ReplaceAllInMarkup` (LNK-029) instead of
`FindAll…`/`ReplaceAllString`. They mask the non-markup regions first. If you are
only COUNTING, say what you counted — "href-shaped byte sequences" is an upper
bound, not a number of links — and never quote one spelling's exposure as the size
of the class: `href="' + x` and `href="${x}"` take **different arms** of the same
repair (empty-href vs phantom), so a census for the first is blind to the second.

**⚠ the second trap, for anyone writing this guard somewhere else.** Do not filter
matches after the fact — MASK the region before matching, and pick the filler byte
deliberately:

- **filtering leaves a second defect.** `<script>var t='<a href="/gone">';</script><a href="/gone">Pricing</a>`
  — the non-greedy `</a>` closes at the REAL anchor, so ONE match spans both.
  Dropping it drops the genuine defect with the decoy, and `FindAll` never
  revisits those bytes.
- **a WHITESPACE filler manufactures matches.** `\ssrc\s*=\s*""` cannot cross
  `<style>…</style>`, but it crosses the spaces that replaced it — and that match
  **begins outside the span**, where an offset check has no view, so it deletes the
  style element and the attribute together. Use NUL. A same-length mask preserves
  offsets; it does **not** preserve the absence of matches.

- **source:** 2026-08-02, `bugs_open/180`, found by the `bugfix_136_sibling_link_repair`
  lane when a census row turned out not to be an anchor. Live damage at the time:
  1 page of 509 fleet-wide (`vetcomparison.uk/tools/cma-obligation-checker`), with
  `success:true` on every path. Fixed in `07576d4e1` for the two WRITERS; the
  DETECTORS still have the blind spot deliberately (a false finding costs
  attention, a false repair costs content).
- **added:** 2026-08-02, bugfix_136_sibling_link_repair lane

### `pages.build_status = 'deployed'` is NOT "is this page live" — 35 of 46 `needs_rebuild` rows have already shipped and are still being served

- **footprint:** `pages.build_status`, `pages.deployed_at`, any guard spelled
  `build_status = 'deployed'` / `COALESCE(build_status,'') <> 'deployed'`,
  `realisedPageIsBuilt`, `findStrandedNavPages`, `NeverDeployedPagePredicateFor`,
  `PageHasShippedPredicateFor`
  > **CORRECTED 2026-08-03 — this footprint used to name `pageHasBeenLive`, which no
  > longer exists** (it WAS the hand-rolled predicate this entry was written about, and
  > it was deleted the same day in favour of the shared one). Four council seats read
  > the stale name and asked whether that symbol still carried the defect — a footprint
  > pointing at a deleted symbol costs a reviewer a search for a bug that is gone.
  > **When you fix the thing a landmine guards, re-read the landmine.** The reusable
  > builders are named instead: reach for those, never a fresh spelling.
- **fires when:** you write a predicate meaning "don't touch a page a visitor can
  see" — a replan guard, a re-type guard, an overwrite refusal, a repair
  candidate set
- **the tell:** nothing. The guard reads perfectly, passes review, and protects
  the 491 `deployed` rows exactly as intended. The 46 `needs_rebuild` rows walk
  straight through it, and **35 of them carry a non-null `deployed_at`** — they
  were deployed, they are still being served, and they are merely waiting to be
  rebuilt. Your "safe" branch then mutates a live page, and the artefact on the
  CDN does not change, so the damage is invisible until someone loads the page.
- **the check:** ask the column that cannot lie about it, and check both:
  ```sql
  SELECT COALESCE(build_status,'(null)') AS bs, count(*),
         count(*) FILTER (WHERE deployed_at IS NOT NULL) AS ever_deployed
  FROM pages GROUP BY 1 ORDER BY 2 DESC;
  -- 2026-08-02: deployed 491/490 · needs_rebuild 46/35 · planned 42/0
  ```
  **Do NOT hand-roll it, and do NOT name the status** — use the estate's one
  definition, `datahelpers.NeverDeployedPagePredicate`
  (`deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed'`), negated for
  "has shipped". **CORRECTED 2026-08-02, hours after this entry was written: its
  first version said to write `build_status IN ('deployed','needs_rebuild') OR
  deployed_at IS NOT NULL`, and that is WRONG on 11 live rows** — `needs_rebuild`
  with no `deployed_at`, no `last_built_at`, mostly zero components, never built.
  Naming the status is itself the trap: `datahelpers/links_deployment_test.go`
  forbids it, because singling it out produced a 34-page false-positive class for
  the nav lane. `deployed_at IS NULL` already catches the 35 that HAVE shipped.
  And note the converse trap on the same columns: `deployed_at IS
  NOT NULL` means "deployed ONCE", **not** "fetchable now" (`bugs_open/098` —
  archiving does not retract the live page), so it is the right test for *may I
  overwrite this* and the wrong one for *is this reachable today*.
- **source:** `bugs_closed/037` is this exact miss, filed as its own case ("`needs_rebuild`
  pages are unprotected by the replan guard" — `realisedPageIsBuilt` returns
  `status == "deployed"`); re-measured and widened in `bugs_open/175` / PBP-027,
  where `bugs_closed/081`'s refusal guard had inherited the narrow form
- **added:** 2026-08-02, bugfix_175_page_role_upsert lane

### There are THREE `pages` upsert helpers and they have OPPOSITE collision policies — picking the familiar name re-types a live page

- **footprint:** `UpsertPageForRole` (`platform/orchestration/actions/page_role_upsert.go`),
  `upsertPage` (`platform/orchestration/actions/site_db_actions.go:1090`), `upsertPage`
  (`cmd/webdesignport/import.go:163`), `resolveNewPageConflict`
  (`apply_gap_plan_action.go`), the `pages` table
- **fires when:** you are writing a new action that creates a page and you reach for
  "the page upsert helper"
- **the tell:** both compile, both return a page id, and both look like the idempotent
  thing to call. They answer a name collision in **opposite** ways:
  `site_db_actions.upsertPage` carries `page_type = EXCLUDED.page_type` and will
  **re-type whatever it collides with** (correct for it — it is the plan-sync path,
  where the plan is the authority on what a page is);
  `UpsertPageForRole` **refuses** a live row of another role and files a human decision
  (correct for it — a constant-role arm has no authority to re-type a served page).
  Call the wrong one and nothing errors; a live page silently changes what it serves,
  or a legitimate plan sync starts filing human decisions it should not.
- **the check:** ask **where your `page_type` comes from**, not which helper is nearest.
  A role that is a **compile-time constant of your arm** (`'tool'`, `'report'`,
  `'blog-post'`) → `UpsertPageForRole`. A role that arrives from a **plan, an LLM or a
  crawl** → the plan-sync path, and do NOT use `UpsertPageForRole`: its ADOPT branch would
  hand a model-steered arm the authority `bugs_closed/081` deliberately declined.
  `UpsertPageForRole` rejects a caller that passes its own `page_type` column, so that
  half is enforced; the LLM-role half is enforced by review only, which is
  `architecture_review/RFC_010`'s open question.
- **source:** `bugs_open/175` / PBP-027, 2026-08-02. The third helper was found by the
  council's `prior_art_librarian` seat asking whether a page-upsert helper already existed
  before a new one was built — it did, and it is the one with the opposite policy
- **added:** 2026-08-02, bugfix_175_page_role_upsert lane

### A diagnosis bundle's "agent state" section lists an agent with NO llm_call_log lines — that is NOT evidence the agent made no calls

- **footprint:** `diagnosis_artifacts` (`kind='bundle'`), the
  `### agent state (auto-gathered` section, `llm_call_log`,
  `platform/orchestration/actions/diagnose_load_runtime_action.go` (`gatherAgentState`)
- **fires when:** you are reading a diagnosis bundle — your own run's or a retained one —
  and reasoning from which agents do or do not appear in its `- llm_call_log [...]` lines.
  No symptom precedes this: the section looks complete and its heading says
  "agent types named in the symptom/hypothesis".
- **the tell:** there isn't one, and that is the trap. Until `v1.0.1232` the gather asked
  for **one shared `LIMIT`** across every named agent type
  (`WHERE agent_type = ANY(...) ORDER BY created_at DESC LIMIT n`), so rows went to
  whichever agent had been busiest — the others rendered **zero lines, with no marker**.
  Measured over the retained corpus: of 23 bundles naming more than one agent type and
  returning any rows, **all 23** showed exactly ONE type. A bundle listing four agents
  under `agent_definitions[...]` and log lines for only one is the DEFAULT shape, not a
  finding about those agents.
- **the check:** **every bundle written before `v1.0.1232` is affected, and they are
  retained ~30 days — so this trap is live in the corpus until ~2026-09.** Do not cite a
  bundle's agent-state section as evidence an agent was uninvolved. Go to the table:
  ```sql
  SELECT agent_type, count(*), max(created_at) FROM llm_call_log
  WHERE agent_type = ANY('{a,b,c}'::text[]) GROUP BY 1;
  ```
  On a POST-`1232` bundle the section states it: `> no llm_call_log rows exist for: X`
  means the table is genuinely empty for X, and
  `> the per-type llm_call_log cap of N was reached for: Y` means Y has older calls that
  were not gathered. **Absent both lines on an old bundle, you know nothing either way.**
  The same bundles' `### agent state` heading may also over-claim coverage if the symptom
  named more than `agent_state_cap` (5) types — pre-`1232` that truncation is silent AND
  the survivors are non-deterministic, so two runs on one symptom can disagree.
- **source:** `bugs_open/172`, fixed `3761a04ca`, 2026-08-02. Found while sizing the
  count-based cap the ticket names — which was and remains latent.
- **added:** 2026-08-02, bugfix_172_agent_state_cap lane

### An ENTRY-POINT agent reads as an orphan on every database axis whether it is dead or in daily use — its caller is a person with a script

- **footprint:** `agent_definitions` (`is_active`, `deleted_at`),
  `scripts/initial_messages/`, Kafka topic `system.agent.generic.requests`,
  `{"action":"orchestrate","config":{"agent_type":...}}`,
  `docs/agent_docs/docs024_key_docs_latest/retired_agents/`
- **fires when:** you are deciding whether an agent is safe to retire, deprecate or
  delete, and you are building the case out of queries. No symptom precedes this —
  the checks are the *right* checks and they all come back clean.
- **the tell:** there is none, and that is the whole trap. The standard blast-radius
  battery — `site_work_items.handler_agent`, `site_specs.created_by`,
  `scheduled_tasks`, `agent_instances.template_id` FK, other agents'
  `default_config` — asks **which agent references this agent**. An entry point has
  no such referrer *by definition*: it is spawned by an operator publishing to
  Kafka, so a live one and a dead one produce **identical zeros**. This is the third
  spelling of one error in this directory: `report-builder` was saved by a
  `scheduled_tasks` row (absence of WORK ≠ absence of WIRING); this one has no row
  to find at all (**absence of WIRING ≠ absence of a CALLER**).
- **the check:** grep `scripts/` for the agent type and **compare file dates against
  the script that supersedes it** — the question is not "does anything reference it"
  but "**has the operator habit moved**":
  ```bash
  grep -rIl '<agent-type>' scripts/ | xargs -r ls -la
  grep -rn 'agent_type' scripts/initial_messages/*/ | grep orchestrate   # who else is an entry point
  ```
  A superseded trigger script is positive evidence; a *current* one is a veto. For
  `intake-orchestrator` the old script was last touched 2026-06-21 and
  `020_build_pipeline/082_submit_domain_unified.sh` (2026-07-30) routes to
  `domain-submitter` instead — that date gap, not the zero rows, is what made the
  retirement safe. Note the corollary when RESTORING: putting the rows back does not
  put the habit back, and the menu the agent reads may have been gutted meanwhile.
- **source:** `retired_agents/README_multipage-website-builder.md` § "Third
  retirement", 2026-08-02. Found while retiring `intake-orchestrator` +
  `site-classifier`, after the same session's `report-builder` near-miss made the
  DB-only battery look sufficient.
- **added:** 2026-08-02, retired_agents lane

---

## A component's `input_schema` fallback is NEVER consulted at render time — a new field renders EMPTY and the deploy reports success

- **footprint:** `content_components.input_schema`, `page_components.content_data`,
  `platform/orchestration/actions/rerender_page_sections_action.go`,
  `RenderTemplate` / `executeGoTemplate`
- **fires when:** you add a field to a component's `input_schema` and a
  `{{.your_field}}` to its `html_template`, then re-render the pages that use it.
  No symptom precedes this: the schema validates, the template parses, the loader
  accepts it, and the page renders.
- **the tell:** there is none on the page, and the one marker that WOULD have made
  it visible is deliberately removed. Three citations, read at HEAD on 2026-08-03:
  - `rerender_page_sections_action.go:393-396` — the render context is built by
    exactly three `mergeIntoRenderContext` calls: `baseData`, `s.contentData`,
    `plan.ResolvedData`. **`input_schema` is not one of them**, and neither file
    mentions `fallback` at all.
  - `call_agent.go:1172` — `Option("missingkey=zero")`. On a
    `map[string]interface{}` that yields `nil`, which `text/template` renders as
    the literal `<no value>`.
  - `component_library.go` (`RenderTemplateReportingMissing`) — then
    `strings.ReplaceAll(result, "<no value>", "")`. **The artefact is stripped on
    purpose**, leaving a clean empty string, with a `Warn` naming the fields
    (`Error`, if the blank landed in an `href=`/`src=` — a dead control).

  So the placeholder vanishes silently and everything around it is correct. It is worse than a visible break, because the
  page still passes structural and numeric acceptance checks: the element is
  present, and no number moved. **Adding a schema field + a template placeholder
  is only TWO THIRDS of a change.**
- **the check:** before re-rendering, diff the schema's field list against every
  consuming row's `content_data` keys — a field missing from ANY row is a field
  that will render empty on that page:
  ```sql
  SELECT p.name,
         (SELECT string_agg(k, ', ' ORDER BY k)
            FROM jsonb_object_keys(cc.input_schema->'fields') k
           WHERE NOT (pc.content_data ? k)) AS will_render_empty
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE cc.function = '<the component>';
  ```
  Backfill before you render — and merge `patch || stored`, **not**
  `stored || patch`, so a live value edited by the writer loop or a human beats a
  fallback. `loancalculator_couk/decompose/backfill_content_data.py` does exactly
  this with `--check`/`--apply`.
- **source:** loancalculator.co.uk, 2026-08-03, caught one step before it shipped.
  It would have served the `tool-loan-vs-savings` accessibility badge as an EMPTY
  element — the whole of that fix, absent, on a page whose acceptance check would
  still have passed — and `tool-consolidation-risk`'s withheld-comparison notice as
  invisible text in two empty inline colours.
- **evidential basis, stated because the two differ:** the render path above is
  READ, at HEAD, and quoted line-for-line. The blank page was never OBSERVED,
  because the whole point was to backfill before rendering rather than ship one to
  find out. What IS measured is the gap itself: `backfill_content_data.py --check`
  found exactly the 4 fields the two schemas had gained and 0 others, on rows whose
  pages render correctly today — so the schema and the row demonstrably disagree,
  and only the row is consulted.
- **added:** 2026-08-03, loancalculator lane. `landmine-verifier` returned
  NEEDS_HUMAN_REVIEW on the first pass — **not a challenge to the claim**: the code
  index was last built at `d98010e8` (2026-07-28), so every footprint file returned
  zero rows and it could not confirm or deny mechanically. It asked a human to
  confirm no fallback-merging had been added since; the three citations above are
  that confirmation, taken from the working tree rather than the index.

---

## `toolgolden.py` only ever drives NEIGHBOURHOODS OF THE SHIPPED DEFAULTS, so a boundary defect is invisible to a green equivalence gate

- **footprint:** `loancalculator_couk/toolgolden.py` (`VECTORS`),
  `loancalculator_couk/rewrite/verify_rewrite.py`,
  `webdesign_tools_repair/toolprobe.py`
- **fires when:** you change a calculator and read a green `verify_rewrite.py` /
  `toolgolden.py --compare` as evidence that the change works. It is strong
  evidence that nothing *regressed*; it is frequently **no evidence at all** that
  your fix is present.
- **the tell:** none — the gate prints `MATCHES` and an id-field count, which reads
  as coverage. `VECTORS = [("defaults", 1.0), ("double", 2.0), ("half", 0.5)]`
  scales each numeric field's **own default**. That is a good policy (it keeps
  every value in its intended domain for any tool, with no per-tool config) and it
  means the driver never leaves the neighbourhood of the shipped defaults. Any
  defect at a boundary is unreachable: **0** is not a scaling of an APR default of
  8.9, and a **blank** field is not a scaling of anything, because the driver fills
  every numeric input it can find. Measured on this lane: two fixes reported
  `MATCHES` both **before and after**, having asserted nothing about either.
- **the check:** write a case that drives the defect condition itself, and prove
  the case can FAIL — render the component from a pinned pre-fix sha
  (`git show <sha>:<path>`, never `HEAD`, which stops being a control the moment
  you commit) and require the same case to READ DIFFERENTLY, not merely to fail.
  Scoring on pass/fail is weaker and partly wrong: a case asserting "£448.024
  before, £448.02 after" passes on both sides and is the most exactly specified
  case you have. `loancalculator_couk/rewrite/defect_vectors.py --both` is the
  worked implementation; it scores each case PROVEN / CONTROL / VACUOUS.
- **source:** loancalculator.co.uk, 2026-08-03, fixing 0%-APR car finance and a
  blank-rate consolidation row. The harness then reproduced the same failure in
  itself within the hour — see its `PRE_FIX_REF` note.
- **added:** 2026-08-03, loancalculator lane

### A scalar written into an array-declared chrome schema field silently destroys the WHOLE slot's render — until v1.0.1231+1 rolls, and loudly-absent after
- **footprint:** `site_specs` aspect `site_config` values consumed by `content_components.input_schema` fields with `type: array` (e.g. `config.chrome.compliance_lines`, STY-051); `render_site_components_action.go` gap-fill
- **fires when:** you hand-seed a per-site chrome config value (the STY-050/STY-051 pattern) and write a STRING where the schema declares an ARRAY — e.g. `"chrome": {"compliance_lines": "one line"}` instead of `["one line"]`. Nothing errors: the fill succeeds, `{{range}}` over the string errors the ENTIRE template at exec, and the renderer silently drops to the regex-fallback path — the footer/head/header renders degraded fleet-shape-wise on that site, with only a `logger.Warn` anywhere.
- **the tell:** the stored `rendered_html` loses its Go-template-rendered structure (fallback output) while the render step reports success. `<!-- FOOTER SOURCE: ... -->` markers may vanish or the gated blocks half-render.
- **the check:** `jsonb_typeof(data#>'{chrome,compliance_lines}')` must read `array` BEFORE dispatching a chrome re-render. From the first chassis roll after commit `2046b6975`, the gap-fill REFUSES the mismatched fill (named Warn, block renders absent, rest of chrome normal) — but the seed being wrong is still your bug; the guard only contains it.
- **source:** council corr 56ab6e23 (bug_historian advisory objection on the STY-051 seam), 2026-08-02; guard measured safe against all 69 array/list-declared fields (53 ranged, 16 unreferenced, 0 bare-output)
- **added:** 2026-08-02, portfolio_positioning lane

### The idea.uk box now answers 80/443 ONLY to Cloudflare — a timeout curling the IP is the FIREWALL, not an outage; and grey/DNS-only records would take the site down AND silently kill cert renewal

- **footprint:** `116.203.204.115`, `2a01:4f8:1c18:7c31::1`, `idea.uk`, `ufw-cloudflare-lockdown.sh`, `/etc/nginx/conf.d/cloudflare-realip.conf`, `proxied`, `59aded94c550f4b20c462bb7619e70c8`

Since 2026-08-03 (Option B complete): idea.uk is orange behind Cloudflare (SSL
Full strict), nginx restores real client IPs from `CF-Connecting-IP`, and ufw
allows 80/443 **only from the 22 published CF ranges** (v4+v6; SSH open as
before). Verified: direct curls to the IP time out on 80, 443 and the v6
address, while the site serves 200 with a `cf-ray` via the edge; all 16 routes
identical; two-network real-IP proof passed (5.65.164.9 + 116.203.204.115
distinct in access.log, zero CF-range client IPs).

**Why it misleads, twice:**
1. **A timeout curling `116.203.204.115` now looks like the box is down.** It
   is the firewall doing its job. Test via `https://idea.uk` (expect `cf-ray`),
   or from the box itself; `ssh` still works and is the real liveness check.
2. **Flipping the records grey/DNS-only no longer "just bypasses Cloudflare" —
   it takes the site DOWN**, because visitors then dial the origin directly
   and the firewall refuses them. **And certbot renewal dies with it** (LE's
   validators are not in CF ranges; while orange they reach the origin via the
   edge, which IS allowed). Grey was the safe rollback before the lockdown;
   now rollback is `ufw allow 80/tcp && ufw allow 443/tcp` FIRST, grey second.

Also: `setup.sh`'s full provision runs `ufw --force reset` — after any
re-provision, re-run `/root/ufw-cloudflare-lockdown.sh` (copy also in the repo
at `idea_uk_vm_site/box/`), or the origin is world-open again while DNS still
says orange, which works and hides it.

- **source:** 2026-08-03, `idea_uk_vm_site/RUNNING_NOTES` §X.40; RUNBOOK §4a "Progress" block
- **added:** 2026-08-03, idea_uk_vm_site lane ("ideauk sec" session)

---

### A keyed work item whose KEY is coarser than its FINDING drops every finding after the first — and the happy path is identical either way, so nothing tells you

- **footprint:** `platform/orchestration/actions/load_work_item_actions.go` (`insertWorkItem`, `writeWorkItem`), `site_work_items`, `idx_swi_dedup`, `item_key`, `refreshOnConflict`, `platform/orchestration/actions/refresh_evidence_base_action.go`, `platform/orchestration/actions/evidence_citations.go`, `platform/orchestration/actions/directory_claims.go`

`insertWorkItem`'s dedup is correct and does what it says: one OPEN row per
`(site_id, item_key)`. What it does NOT do is keep the FINDING. When the key is
per SITE and the finding is per FACT/PAGE/CITATION, the second, **different**
finding hits `ON CONFLICT DO NOTHING` and is gone, and the open row goes on
describing the first thing it ever saw — for as long as a human leaves it open.

**Why nothing prompts you to check:** the happy path is byte-identical. An insert
that succeeds behaves the same whether or not the key is granular enough, so the
defect is invisible in every test, every green build, and every log line except
one `inserted=false` in a pod that will be replaced. Measured 2026-08-02 on
`stale_evidence`: **four of five open items named the wrong facts** — one naming a
completely different fact from the one that had moved, one describing drift that
no longer existed. Nothing had errored, ever.

**the check:** for any keyed detector, ask **"is this item's `item_key` as
granular as its finding?"** If the `spec` carries a LIST that will differ next
run while the key does not, it is dropping findings today. Prove it against the
live DB — compare what the open item SAYS against what the last run FOUND:

```sql
SELECT s.domain, swi.spec->'<the list>' AS what_the_record_says, swi.created_at
  FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
 WHERE swi.item_type='<your type>' AND swi.status NOT IN
       ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
```

From 2026-08-03 the remedy exists and is opt-in per caller:
`writeWorkItem(ctx, tx, item, refreshOnConflict, logger)` refreshes the open row's
`summary`/`spec` instead of discarding the finding (BATCH-005). **It is OFF by
default and switching a caller on is a judgement, not a formality** —
`needs_human_review` is deliberately not a "held" status, so a refresh CAN change
an item under a human who is reading it. Three call sites are in this shape and
have NOT been switched: `evidence_citations.go` (`citation_unverified:<site>`),
and `directory_claims.go` twice (`directory_citation_unverified`,
`stale_directory_claim`).

- **source:** `bugs_open/091`; `docs024_key_docs_latest/bugfix_091_workitem_conflict_refresh/`
- **added:** 2026-08-03, bugs-sweep lane (bugfix_091)

---

### The anti-churn probe SWALLOWS its own error, so no sqlmock test in `platform/orchestration/actions` can detect a change to `recurrenceExpected`

- **footprint:** `platform/orchestration/actions/load_work_item_actions.go` (`insertWorkItem` two-strike block), `recurrenceExpected`, `platform/orchestration/actions/work_item_recurrence_test.go`, `github.com/DATA-DOG/go-sqlmock`

The two-strike block runs `SELECT COUNT(*) …` and then reads it as
`if err == nil && terminalCount > 0` — **deliberately**, so a probe failure
degrades to "no history" rather than blocking the insert. The side effect is that
a sqlmock test which does not `ExpectQuery` that probe **still passes**: the
unexpected query returns an error, the error is discarded, and the insert
proceeds. `ExpectationsWereMet()` does not save you either — it reports
*unfulfilled* expectations, not *unexpected* calls.

**Why it misleads:** flipping `recurrenceExpected` on a caller changes whether the
probe runs at all, i.e. whether that caller's items can be suppressed within 3h or
branded `unresolved` after two strikes — the exact regression `bugs_open/024`
recorded on the tool fix loop. A full green suite is **not evidence** that you
have not done it. Found on 2026-08-03 by mutation: clearing the flag failed one
direct unit test and left three behavioural tests passing.

**the check:** cover `recurrenceExpected` with a **direct assertion on the built
`workItem`**, never with a behavioural test that drives the action. And when you
mutate-test anything in this package, remember the observer may be deaf: if a
mutation does not fail a test, find out whether the guard held or whether nothing
was listening ([[a-mutation-that-passes-may-have-hit-a-guard-in-series]]).

- **source:** `bugs_open/091`; `docs024_key_docs_latest/bugfix_091_workitem_conflict_refresh/NOTES` 01:10
- **added:** 2026-08-03, bugs-sweep lane (bugfix_091)

### `agent_definitions.updated_at` is bumped by BULK SWEEPS — a fresh timestamp is not another session working on your agent

- **footprint:** `agent_definitions`, `agent_definitions.updated_at`
- **fires when:** you check "is anyone else in this agent?" before editing a
  definition, and read a recent `updated_at` as a person. It reads exactly like
  an active lane, and the correct response to that (back off, go and find the
  owner) is expensive and wrong
- **the tell:** none on the row itself. `version` does not move either — bulk
  updates are in-place, so a swept row and a hand-edited row are byte-identical
  in their metadata
- **the check:** ask whether the timestamp is *shared*. A sweep stamps the same
  minute across the fleet:
  ```sql
  SELECT date_trunc('minute', updated_at) AS minute, count(*)
    FROM agent_definitions WHERE updated_at > now() - interval '24 hours'
   GROUP BY 1 ORDER BY 1;
  ```
  Measured 2026-08-02: **184 rows at 22:08**, 3 at 21:16, 2 at 22:37. A count in
  the dozens-to-hundreds is a sweep; a count of 1–3 is someone's hands. Then
  confirm against the thing that actually matters — diff the **field you care
  about** against its seed, don't infer from the timestamp. And grep live
  `.jsonl` transcripts for the **step name**, not the agent type: agent types
  appear in every fleet census, so they return hits from sessions doing nothing
  to you (9 sessions matched `domain-research-classifier`; **0** matched
  `classify_and_extract`, which was the truth)
- **source:** `bugs_open/183`, filed 2026-08-03
- **added:** 2026-08-03, mortgagecalculator.co.uk adoption lane

### A `PostToolUse` hook that writes to **stderr and exits 0** reaches NOBODY — and the transcript records it in a way that greps like delivery

- **footprint:** `.claude/settings.json`, `.claude/settings.local.json`,
  `PostToolUse`, `hookSpecificOutput`, `scripts/memory-index.py`,
  `scripts/landmines-session-start.py`, `.claude/hooks/psql_readonly_gate.py`
- **fires when:** you write (or trust) a hook that reports something — a budget
  breach, a lint, a warning — using `sys.stderr.write(...)` + `sys.exit(0)`,
  usually with a comment saying "advisory: never block". It is wired, it fires on
  every matching tool call, its logic is correct, and its message is delivered to
  no one. Nothing anywhere reports an error
- **the tell:** none from outside. Worse, there is an ANTI-tell: the text IS in
  the session `.jsonl`, as an `attachment.stderr` record — one per invocation
  (measured: 65 in a single transcript) — so `grep`ping the transcript for your
  banner **finds it**, and that reads exactly like proof it was delivered. It was
  not. Running the hook by hand is also not a test: with no stdin it parses no
  `file_path`, exits silently, and reads as "not firing at all"
- **the check:** two facts settle it. (1) On **exit 0** Claude Code parses
  *stdout* for JSON and, for every event except `UserPromptSubmit`,
  `UserPromptExpansion` and `SessionStart`, routes it to the debug log; only
  **exit 2** "shows stderr to Claude". So stderr+exit 0 is the single combination
  that reaches neither the model nor the transcript. (2) For `PostToolUse` the
  tool has **already run** — exit 2 cannot block anything, so the safety that
  motivated exit 0 never existed. Deliver via
  `{"hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": "…"}}`
  on stdout. Verify at the ARTEFACT, never at the hook: feed realistic stdin
  (`echo '{"tool_input":{"file_path":"…"}}' | python3 hook.py --hook`), then make
  a real edit and confirm the text arrives in your own context as a system
  reminder. Use a sibling hook you demonstrably DO receive as the positive
  control — in this repo, `landmines-session-start.py`
- **source:** `scripts/memory-index.py` was mute from 2026-07-28 to 2026-08-03;
  **15 hand-compactions of `MEMORY.md` happened underneath it**, every one a
  session hand-rolling `wc -c` while the tool computed the exact number for them.
  Fixed in `3137554ff`. (The harness's own 0.8× nag was NOT mute — what was lost
  was the earlier 17.1KB trigger, the per-write delta and the section breakdown)
- **added:** 2026-08-03, auto-memory index lane

### A stale stub-resolver cache makes a freshly-migrated domain look DOWN — and `dig @1.1.1.1` corroborates the wrong answer

- **footprint:** `curl` / `dig` against any domain whose **nameservers changed
  recently** · `systemd-resolved` (`127.0.0.53`) · `/etc/resolv.conf` ·
  `idea.uk` · any origin firewalled to Cloudflare ranges as part of a migration
- **fires when:** a domain moves nameservers (Hetzner → Cloudflare, registrar →
  anywhere) **and** the old origin is firewalled in the same change — which is
  what a *well-executed* migration does. Your resolver still hands out the
  pre-migration address; that host now silently DROPs packets; every request
  times out with **no HTTP status, no TLS error, nothing to read**. The site is
  serving 200 to the rest of the world the entire time
- **the tell:** none — and there is an **anti-tell that actively confirms the
  error.** The instinctive cross-check, `dig +short A <host> @1.1.1.1`,
  **bypasses the system resolver**, so it returns the correct new records. You
  then hold "DNS is fine" and "the site times out" simultaneously and conclude
  the server is down. Two tools, two different resolvers, and nothing announces
  that they disagree. A Cloudflare-fronted host that hangs for **90s** is itself
  a signal — a genuinely unreachable origin behind Cloudflare returns **522 in
  ~15s**, so a total hang means you never reached Cloudflare at all
- **the check:** ask the two resolvers separately and compare, then pin the
  address before believing any outage:
  ```bash
  getent ahosts <host> | awk '{print $1}' | sort -u   # what curl ACTUALLY uses
  dig +short A <host>                                 # system resolver
  dig +short A <host> @1.1.1.1                        # differs => you are stale
  curl --resolve <host>:443:<ip-from-1.1.1.1> https://<host>/   # 200 => site is UP
  sudo resolvectl flush-caches
  ```
  The discriminating signature is **`curl` hanging where
  `openssl s_client -connect <ip> -servername <host>` succeeds** — those two
  differ in exactly one step, name resolution. If openssl-by-IP returns 200, the
  fault is on your side of the wire. **Generalisation: a timeout is evidence
  about the PATH, not about the server, and the path starts at your own
  resolver.**
- **source:** 2026-08-02. I reported `idea.uk` — a live, card-taking site — as
  down, having run `dig @1.1.1.1` and read its agreement as corroboration. It was
  serving 200 throughout; `getent` still held the Hetzner address from before the
  Cloudflare migration. Logged in `WRONG_CALLS.md`;
  `idea_uk_vm_site/HANDOFF_2026-07-31_cloudflare_decision.md` §7
- **added:** 2026-08-02, webdesign.uk build-service lane

### `dig` cannot distinguish a missing DNS record from a missing Cloudflare Worker route — both return an empty answer

- **footprint:** `dig +short A <sub>.<zone>` · Cloudflare Worker routes ·
  `*.ugg2.com/*` · `portfolio-sites-router` · any wildcard-preview scheme
- **fires when:** you probe a subdomain that should be served by a Worker, get
  nothing back, and name a cause. A wildcard preview needs **two independent
  things** — a proxied `*` DNS record **and** a `*.<zone>/*` Worker route — and
  the absence of *either* produces the identical empty `dig`. So the measurement
  you just ran cannot support a claim about which one is missing, and it feels
  like it can
- **the check:** ask the API, which answers exactly this:
  `GET /zones/{zone}/workers/routes` (list) — one call, and it names the script
  bound to each pattern. Then state only what you looked at
- **source:** 2026-08-02. I measured "no A record for `test.ugg2.com`" and told
  the owner **both** halves were missing. The route `*.ugg2.com/*` →
  `portfolio-sites-router` had existed all along; only the DNS record was absent
- **added:** 2026-08-02, webdesign.uk build-service lane

### The Cloudflare API token reaches all 36 zones, cannot write Redirect Rules, and silently rejects DNS comments over 100 chars

- **footprint:** `~/.config/cloudflare/token` · `api.cloudflare.com/client/v4` ·
  `/zones/{z}/rulesets` · `/zones/{z}/pagerules` · `/zones/{z}/dns_records`
- **fires when:** you script a Cloudflare change. Three separate traps.
  (1) **Scope:** the token is account-wide — **36 zones**, the whole estate, with
  no per-zone fence. A loop over `/zones` will happily edit someone else's site;
  name the zone id explicitly.
  (2) **Rulesets are denied:** `/zones/{z}/rulesets*` returns
  `code 10000 Authentication error`, so the **modern Redirect Rules API is not
  available** — while `/pagerules` works fine. A "Forwarding URL" Page Rule
  delivers the same 301/302 and is the workaround (3 per zone on Free).
  `code 10000` reads like a bad token; the token is fine, that one product is not
  in its policy.
  (3) **DNS `comment` is capped at 100 characters** — over it you get
  `code 9313` and **the record is NOT created**. It reads like a warning about
  metadata; it is a hard failure of the whole write
- **the check:** `GET /user/tokens/verify` proves the token is live but tells you
  **nothing about which products it can reach** — probe the specific endpoint
  with a GET before scripting a POST against it. And re-read the response
  `success` field rather than the HTTP status; Cloudflare returns 200 with
  `success:false` for permission and validation failures alike
- **source:** 2026-08-02, applying the webdesign.uk holding redirect
- **added:** 2026-08-02, webdesign.uk build-service lane

### Pinning a `site_specs` row does NOT protect it — `write_site_spec` ignores `pinned` and drops it

- **footprint:** `site_specs.pinned`, `write_site_spec`,
  `platform/orchestration/actions/site_spec_actions.go`
- **fires when:** you hand-write a spec (positioning, an owner decision, a
  correction) and set `pinned = true` believing that stops an agent superseding
  it. The column exists, the write succeeds, and it buys you **nothing** on the
  path that actually overwrites specs
- **the tell:** none. The pin is silently ineffective *and* silently lost — the
  supersede `UPDATE ... SET is_current=false` has no `pinned` check, and the
  follow-on `INSERT` does not name the column at all, so the replacement row
  defaults to `pinned=false`. Nothing errors, nothing logs
- **the check:** `grep -c pinned platform/orchestration/actions/site_spec_actions.go`
  → **0**. Only two writers honour it (`evidence_citations.go:374`,
  `refresh_evidence_base_action.go:719`), and both are evidence-base paths — so
  `pinned` works for `evidence_base` and is inert for `identity`,
  `content_direction`, `design_intent` and every other aspect. To find a pin that
  was actually lost, do **not** count `pinned AND NOT is_current` (that is what
  ordinary versioning looks like when the pin IS carried forward — it returned 35,
  all but one benign). Ask whether the CURRENT row of a once-pinned stream is
  still pinned:
  ```sql
  WITH ever AS (SELECT DISTINCT site_id, aspect FROM site_specs WHERE pinned)
  SELECT c.pinned, count(*) FROM ever e
    JOIN site_specs c USING (site_id, aspect) WHERE c.is_current GROUP BY 1;
  ```
  Measured 2026-08-03: 9 streams still pinned, **1 lost** (`vonc.com`/`evidence_base`)
- **so:** if a hand-written spec must survive, the durable protection is
  `sites.locked_at` (a real dispatch gate — `load_work_item_actions.go:126-138`
  returns zero items for a locked site), not `pinned`. Re-check the spec after any
  agent run that writes the same aspect
- **evidence strength:** the code gap is verified directly; the empirical harm is
  **one** case, on `evidence_base`, via a session write — stated here so nobody
  quotes this entry as "pins get destroyed constantly"
- **source:** `bugs_open/183` lane (mortgagecalculator.co.uk adoption), 2026-08-03
- **added:** 2026-08-03, mortgagecalculator.co.uk adoption lane

### `sites.locked_at` does NOT hold a site — the live dispatch gate never looks at it

> **CORRECTED 2026-08-03, hours after filing, by the lane that filed it — the
> lock clause is now LIVE, so the headline above is HISTORY, not current state.**
> On the owner's instruction ("fix broken things") the missing predicate was added
> to `find_dispatchable_site`: `WHERE s.locked_at IS NULL AND wi.status IN
> ('triaged','approved')`. Verified by INDUCTION, not by reading the config back —
> one item armed `triaged` against a locked site, held across firing ticks (see
> below for why "it did not dispatch" alone proves nothing).
> **This entry stays because the SHAPE is what protects you**: a control that
> reads back exactly as written and does nothing. It is also only the ONE clause —
> the rest of `213`'s reconciliation (pipeline scope, `approved` status, the
> claimed-mutex in the `pre_query`) is still undone and still belongs to the
> `bugs_open/029` lane. **Do not read "locked_at works now" as "the three
> predicates agree now." They do not.**

- **footprint:** `sites.locked_at`, `build-pipeline-trigger`, `find_dispatchable_site`,
  `scripts/../213_dispatch_gate_matches_dispatcher.sql`
- **fires when:** you lock a site to stop automated work — before a risky build,
  during a review, while an owner decides. The `UPDATE` succeeds, `locked_by` reads
  back exactly as you wrote it, and **the site carries on dispatching**
- **the tell:** none at the lock. The tell is downstream and easy to misread as
  something else: new `orchestration_states` rows for `build-dispatch-loop` on that
  site, minutes AFTER the lock. Measured 2026-08-02: locked at 23:21:35, fresh
  dispatch loops at 23:23:13, 23:25:44, 23:28:13, and a chain ran four handlers deep
- **why:** three predicates in one chain disagree.
  `scheduled_tasks.build-pipeline-trigger.pre_query` (**does** check
  `s.locked_at IS NULL`, but it is a fleet-wide `HAVING COUNT(*)>0` existence test —
  it decides only *whether to fire at all*, never *which site*), then
  `agent_definitions.build-pipeline-trigger.workflow.steps.find_dispatchable_site`
  (**no lock clause** — this is the one that picks the site), then `load_work_items`
  in Go, which *does* honour it (`load_work_item_actions.go:126-138`) but is reached
  too late to stop the dispatch
- **the check:** ask the gate that actually chooses, not the one that fires:
  ```sql
  SELECT CASE WHEN default_config->'workflow'->'steps'->'find_dispatchable_site'
                     ->'config'->>'query' LIKE '%locked_at%'
              THEN 'HONOURS' ELSE 'IGNORES' END
    FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active;
  -- 2026-08-03: IGNORES
  ```
- **what to use instead:** the predicate the gate *does* read — item `status`.
  `deferred` is not in `workItemTerminalStatuses`, so deferring holds the row without
  freeing its `idx_swi_dedup` slot and release is one `UPDATE`. **But note this only
  holds items that already EXIST**: in a chain (`needs_domain_research` →
  `needs_vertical_research` → `needs_strategy` → `needs_briefing` →
  `needs_site_plan`) each handler creates the next item, so you are racing a
  120-second tick. What buys you the time is the gate's own claimed-mutex — while
  any item on the site is `claimed`, the site is not selected
- **already written, never applied:** `213_dispatch_gate_matches_dispatcher.sql` adds
  exactly this clause and names the divergence in its header ("A honours
  sites.locked_at and B does not, so B can dispatch a locked site"). It assessed the
  gap as **"Inert today (0 of 32 sites locked, ever)"** — true when written, and the
  first session to actually use the lock makes it live. **A dormant gap's inertness
  is a property of nobody having used the feature, not of the feature being safe.**
  Migration is unapplied (`schema_migrations` has no 213 row) and belongs to the
  active `bugs_open/029` dispatch-gate lane — contribute there, do not apply it as a
  side effect of your own task
- **how to prove a lock actually holds** (this is the part people get wrong): arm
  exactly ONE item `triaged` against the locked site and watch it NOT dispatch —
  but **a quiet queue has two causes with opposite meanings**, so in the same window
  you must also show the gate *looked*:
  ```sql
  SELECT last_triggered_at FROM scheduled_tasks WHERE name='build-pipeline-trigger';
  ```
  Sample it at the start and end. If that timestamp did not move, your item sat
  still because **nothing ran**, not because the lock held, and you have proved
  nothing. Then release the lock and confirm the same item DOES dispatch — a guard
  that never lets anything through is indistinguishable from a broken pipeline
- **source:** mortgagecalculator.co.uk adoption lane, 2026-08-03 — locked a site,
  watched it build anyway
- **added:** 2026-08-03, mortgagecalculator.co.uk adoption lane

---

## A Go regex carried into Postgres matches NOTHING at exit 0 — `\b` is a BACKSPACE there, and an empty census reads as a clean fleet

**footprint:** any `regexp_matches` / `~` / `!~` over `page_components.rendered_html` · `site_components.rendered_html` · `content_components.html_template` · any fleet census SQL copied from a Go checker in `platform/orchestration/actions/discovery_checks/`

- **the trap:** the discovery checks are Go, and Go's RE2 spells word-boundary `\b`. When
  you audit one of those checks the natural next move is to run its pattern as SQL to size
  the defect fleet-wide — and **Postgres POSIX regex has no `\b` word boundary at all.**
  There `\b` is the **backspace character** (`\y` is the word boundary). A pattern like
  `'<img\b[^>]*\bsrc\s*=\s*"([^"]*)"'` asks for a literal backspace, matches nothing, and
  returns **0 rows with no error and exit 0**
- **why it is worse than an ordinary typo:** the failure is not "my query errored", it is
  **"the fleet is clean"** — the single most reassuring result the query can produce, and
  the one nobody re-checks. It arrives in exactly the shape of good news, at the exact
  moment you are deciding whether a defect is site-specific or systemic. Measured
  2026-08-03: the broken pattern said 0, the correct pattern said **31 occurrences across
  2 sites from 1 component**. A session that measured before reading the page would have
  filed "site-specific", fixed one site, and left the other broken
- **the check, and it costs one line:** **run the census against a row you already know
  matches, FIRST.** Add `AND s.domain = '<the site you are looking at>'` and confirm it
  returns non-zero before you trust the fleet-wide number. A census whose known-positive is
  absent is a broken census, not a clean estate. This is the only check that works, because
  the correct and incorrect queries are visually near-identical and both succeed
- **when writing SQL, not porting it:** use `[[:space:]]` and explicit character classes.
  If you want a word boundary in Postgres it is `\y`. Never paste a Go pattern into psql
  unedited — also note Go's `(?i)`/`(?s)` inline flags are not Postgres syntax (the flag
  is the `'gi'` argument to `regexp_matches`), so a ported pattern can fail two ways at once
- **related:** the existing "a grep proves absence only for the SPELLING it searches" rule
  is adjacent but different — there the spelling is plausible-but-narrow, here the pattern
  is **silently in a different language**, so re-reading it in your own head does not help
- **source:** `docs024_key_docs_latest/finetuning_uk_repair/` (broken-image census,
  2026-08-03); WRONG_CALLS.md same date
- **added:** 2026-08-03, finetuning.uk repair lane

### A Cloudflare token with Client-IP filtering reports the SAME failure two different ways — one of them sends you hunting a permissions bug that does not exist

- **footprint:** `~/.config/cloudflare/token` · `api.cloudflare.com/client/v4` ·
  `code 9109` · `code 10000` · any dual-stack machine calling the CF API
- **fires when:** the token has **Client IP Address Filtering** set and your public
  address changes — a reconnect, a different network, or the machine simply
  preferring IPv6 on one call and IPv4 on the next. The token is valid,
  `/user/tokens/verify` still returns `status: active`, and calls that worked an
  hour ago start failing
- **the tell:** `code 9109 Cannot use the access token from location: <ip>` names
  the address and is unmistakable — **but the very next call can return the generic
  `code 10000 Authentication error` for the identical cause.** Measured 2026-08-03,
  two calls seconds apart: DELETE → 9109, PATCH → 10000. If you happen to see only
  the 10000 you will conclude the token lacks a scope and go re-issue it, which
  fixes nothing.
- **⚠ THE HEALTH CHECK IS EXEMPT FROM THE FILTER — measured, not inferred.**
  2026-08-03, same machine, same minute: `GET /user/tokens/verify` returned
  `success:true, status:active` over **both IPv4 and IPv6**, while
  `GET /zones/{z}/pagerules` from **that same IPv4 address** returned
  `9109 … from location: 5.65.164.9`. So the endpoint everyone reaches for to
  answer "is my token OK?" **answers a different question** — it reports the
  token's own lifecycle, never your right to use it from here. A green `verify`
  is worth nothing as evidence that the next call will work
- **both address families are filtered**, also measured: 9109 named an **IPv6**
  address (`2a02:c7e:3066:5400:…`) on one call and an **IPv4** address on another,
  so allow-listing one family leaves a dual-stack machine failing intermittently
  as it flips between them. Pin the family with `curl -4` so only one address has
  to be listed
- **the check:** read the **`errors[0].message`, never the HTTP status** (Cloudflare
  returns 200 with `success:false`), and on any 10000 immediately re-run one call
  with `curl -4` and one with `curl -6` — if either reports 9109, it is the
  allow-list and **both families need listing**, not one. Fix in Cloudflare → My
  Profile → API Tokens → the token → Client IP Address Filtering
- **source:** 2026-08-03, mid-way through publishing the webdesign.uk shopfront —
  the page reached the bucket but the DNS/page-rule change could not be applied.
  `webdesign_uk_build_service/HANDOFF_2026-08-03_P1_shopfront.md` §3
- **added:** 2026-08-03, webdesign.uk build-service lane

### Two acceptance callers file notes under ONE category with the SAME `created_by` — so "the feature stopped working" and "this row came from the caller you never armed" are the identical query result

- **footprint:** `platform/orchestration/actions/tool_acceptance_actions.go`,
  `request_browser_run`, `request_component_browser_run`, `dispatchBrowserRun`,
  `capture_renders`, `doc_notes`, `tool-acceptance-agent`, `acceptance-run`
- **fires when:** you re-check a per-step opt-in (TL-035's `capture_renders`) by
  selecting over the note category and reading a derived column — `body LIKE
  '%Rendered:%'`. A render-less note appears **newer than** your armed one and you
  conclude the flag was reverted or the seed lost. It was not: **two different actions
  file into that one category**, `request_browser_run` (tool pages) and
  `request_component_browser_run` (components), and arming one leaves the other dark
- **the tell:** there isn't one in the note, and that is the whole trap. **Both callers
  write `created_by='tool-acceptance-agent'` and `source='tool-acceptance'`**, so every
  column the note carries is identical; the discriminator lives on the *orchestration
  row*, which the note does not reference (`source_item_id` is NULL). The two exits the
  brochure lane's own handoff offered both dead-end here: the run **PASSED** 15/15 (so
  "a failing run legitimately has no render" is closed) and the flag was **still `true`**
  in `agent_definitions` (so "the seed was lost" is closed)
- **the check:** never conclude from the note — **join to the run and read the ACTION**:
  ```sql
  SELECT o.orchestration_id, o.workflow_plan #>> '{steps,request_run,action}' AS action,
         o.workflow_plan #>> '{steps,request_run,config,capture_renders}' AS flag
    FROM orchestration_states o
   WHERE o.created_at BETWEEN '<note ts>'::timestamptz - interval '5 min' AND '<note ts>'::timestamptz;
  ```
  A NULL `flag` with a non-null `action` is an **unarmed caller**, not a regression.
  Then ask the prior question the category query cannot: **how many callers are there?**
  `grep -n "dispatchBrowserRun(" platform/orchestration/actions/tool_acceptance_actions.go`
  → `:184` and `:390`, two call sites, one shared helper reading the flag at `:220`
- **the wider rule:** when you arm a shared helper by config, **the blast radius is the
  helper's CALL SITES, not the config rows you edited** — count them in the code before
  claiming coverage. Here the second caller has **no `agent_definitions` row at all**
  (0 rows fleet-wide for `%component_browser_run%`, including snapshots and soft-deleted):
  it runs from an **inline `workflow_plan`**, so a census of live agent configs reports
  full coverage while a live caller sits unarmed and invisible to it
- **related:** this is the config-side twin of the memory index's `input_mapping` /
  `RETURNING` allow-list entry — there two gates on one path, here one gate on two paths
- **source:** 2026-08-03, brochure lane's first re-check after TL-035 went live;
  `brochure_component_library/NOTES_brochure_component_library.md` (2026-08-03 entry) and
  `HANDOFF_2026-08-02_continue_here.md` §2, corrected the same day
- **added:** 2026-08-03, brochure component library lane

---

### A pod-grep marker taken from a COMMENT can never match — and a change that is only identifiers has no marker at all, so "0 on both replicas" reads as "not shipped" when it means "unmeasurable"

- **footprint:** `strings /app/agent-chassis`, `kubectl exec`, pod-grep, deploy verification, `bugs_open/153`, `HANDOFF_*_continue_here.md`

The deploy-verification ritual on this platform is `strings /app/agent-chassis |
grep -c '<marker>'`, and it only works on **string literals the compiler emits**. Two
ways it silently doesn't:

1. **The marker is a Go comment.** Comments are not in the binary. The grep returns 0
   for ever, on a perfectly good image, and reads exactly like a failed deploy. A
   marker lifted from the change's own explanatory comment is the natural mistake,
   because that is the most quotable sentence in the diff.
2. **The change contains no literal at all.** A fix that is *identifiers* —
   `insertWorkItem` → `writeWorkItem`, a flag flipped, a call site rerouted — adds
   nothing greppable. There is no marker to pick, and the honest answer to "how will
   we confirm this shipped?" is **"we cannot, as written"**.

Both were shipped into a handoff on 2026-08-03 (`bugs_open/184`) and caught by a
council seat, not by running it: the file told the next session to grep
`'the bugs_closed/091 class'`, which is a comment.

**the check, before you write a marker into a runbook or handoff:**

```bash
# 1. it must be a LITERAL in production code, not a comment and not a _test.go file
grep -rn "<marker>" --include=*.go platform/ internal/ | grep -v '_test.go' | grep -v '^\s*//'
# 2. prove it COMPILES IN — build a real binary and grep that, do not assume
go build -o /tmp/probe ./cmd/agent-chassis && strings /tmp/probe | grep -c "<marker>"
# 3. and verify the NEGATIVE control is absent from your own source first
#    (a string your change KEEPS greps 1 before and after, and reads as "not shipped")
```

**If step 2 returns 0, the change has no marker — add one rather than pretend.** The
useful form is a log line the code should arguably have had anyway: on 184 the three
emitters made *no record of their own execution* (the DB row was the only evidence
they ran), so giving each an outcome log fixed the observability gap and produced the
marker in the same stroke. A marker invented purely to be grepped is a smell; a marker
that is also the thing you would want in the logs is not.

- **source:** `bugs_open/184`; `docs024_key_docs_latest/bugfix_091_workitem_conflict_refresh/HANDOFF_2026-08-03_continue_here.md` (the corrected block); council `d6cda33d`
- **added:** 2026-08-03, bugs-sweep lane (091/184)

---

## `~/projects/sites` is written by the platform's OWN deploys — it is not a record of what a site used to look like, and a harness baselined on it silently starts comparing your output against itself

- **footprint:** `~/projects/sites`, `~/projects/sites2`,
  `loancalculator_couk/rewrite/verify_rewrite.py` (`SITE_SRC`), any lane script
  taking "the original page" from a sites working copy, `git pull --rebase` in that
  repo
- **fires when:** you build a before/after comparison — "splice my rewrite into the
  REAL page and check the numbers still match" — and take the real page from the
  sites checkout. It is the obvious source: it is on disk, it is the site, and it is
  what the deploy publishes.
- **the tell:** none for as long as your checkout happens to be STALE, which is why
  this is not obvious. Every `Rerender:` commit the platform makes lands in that
  repo, so each page your lane successfully changes **replaces its own baseline**.
  Measured 2026-08-03: `verify_rewrite.py` passed all morning against a checkout
  sitting on a 2026-08-01 commit — genuinely valid runs. Another session ran
  `git pull --rebase` at 10:19 and the next run broke, because
  `tools/standard-calc.html` at HEAD was now the DECOMPOSED page (5 `ported-prose`
  sections) rather than the hand-built original.
  **It broke loudly only by luck of an unrelated rule.** The harness requires each
  cut pattern to match exactly once, and those patterns anchor on original markup
  an assembled page no longer has, so it refused with "opening-div regex matched 0
  times". With looser patterns it would have spliced the component into a page that
  **already contained it** and compared the rewrite against itself — passing, while
  proving nothing.
- **the check:** never read a baseline from the working copy. Export it from a
  pinned ref that predates your own writes, and record in the file WHY that ref:
  ```bash
  cd ~/projects/sites && git log --oneline -- <domain>/ | head        # find a pre-adoption commit
  git archive <sha> <domain>/ | tar -x -C "$(mktemp -d)"              # and use THAT
  ```
  Confirm the ref really is pre-change before trusting it — grep the exported page
  for a marker only the NEW shape has (`class="ported-prose"`, `data-component`)
  and require 0. ⚠ **The repo root is `~/projects/sites`, so paths are
  `<domain>/tools/x.html`**; `git show <ref>:tools/x.html` fails with a `fatal:`
  that `2>/dev/null` will hide, and `grep -c` on the empty result returns 0 — which
  reads exactly like "the marker is absent". That misfire happened while
  diagnosing this entry.
- **source:** loancalculator.co.uk, 2026-08-03. Same class as `defect_vectors.py`'s
  `PRE_FIX_REF` three hours earlier, in a different file: a baseline that names a
  MOVING thing stops being a control, and the expiry is silent.
- **added:** 2026-08-03, loancalculator lane

---

## Obeying `on_missing:"skip_field"` is NOT "wrap the field in `{{if}}`" — for most of the library the field sits in a cell of a fixed-arity row, where that edit is either a NO-OP or emits malformed HTML

- **footprint:** `content_components.html_template`, `input_schema.fields.*.on_missing`,
  `scripts/check_placeholder_fallbacks.py` (its UNGATED class),
  `docs/agent_docs/sql_for_agents/295_twenty_components_gate_their_declared_skip_fields.sql`,
  and any lane clearing an "ungated skip_field" finding
- **fires when:** the lint hands you a list of fields "declared skip_field, referenced,
  never gated" and you do the obvious thing — put `{{if .field}}…{{end}}` around the
  field. No symptom precedes this: the template parses, the lint goes quiet, and the
  page renders.
- **the tell:** there is none, and the two failure modes look opposite but both pass
  the lint, because **the lint tests for a gate ANYWHERE in the template — it cannot
  see what the gate encloses.**
  1. **The NO-OP.** `<td>{{.spec_1_value}}</td>` becomes
     `<td>{{if .spec_1_value}}{{.spec_1_value}}{{end}}</td>`. When the datum is absent
     that renders `<td></td>` — byte-for-byte what it rendered before. The finding
     clears, the blank cell is still there, and the lint will now never mention it
     again. This is the more dangerous half: you have removed the detector, not the
     defect.
  2. **The MALFORMED ROW.** Gate the `<td>` itself and a 4-column
     `platform-comparison` row emits three cells. Every column after the gap shifts
     left, which is worse than the blank it replaced and appears only when a site
     happens to omit that datum.
- **why it is nearly always a cell:** measured 2026-08-03 across the live library —
  **62 of the 68 ungated fields are the ungated PARTNER of a gated field**:
  `{{if .spec_1_name}}<tr><th>{{.spec_1_name}}</th><td>{{.spec_1_value}}</td></tr>{{end}}`.
  The row is gated on the NAME; the VALUE is not. So the element that must disappear
  is almost never the one the field sits in.
- **the check:** before editing, read what ENCLOSES the field and ask what the smallest
  independently-removable element is. Then prove the edit at the render, both ways:
  ```bash
  # the element must VANISH when the datum is absent …
  # … and STILL RENDER when it is present — the positive control is the half that
  # catches an over-firing gate, which "did it disappear?" passes perfectly
  ```
  Render through `actions.RenderTemplate` (the production entry point), not a replica
  of its `text/template` config — 295's harness does exactly this for 20 components.
  Where the enclosing element cannot be dropped, **widen the row's EXISTING gate**
  (`{{if .row1_feature}}` → `{{if and .row1_feature .row1_platform1_value …}}`) rather
  than adding a second one.
- **related:** the entry above on `input_schema` fallbacks never being consulted at
  render time — same seam, opposite direction. That one is why the blank is silent;
  this one is why the obvious repair does not fix it.
- **source:** RFC_009 / `bugs_closed/140`, migration 295, 2026-08-03. All four
  treatments and the per-field reasoning are in that migration's header.
- **added:** 2026-08-03, bugfix_140 lane
- **UPDATED 2026-08-04 — the pressure to make this exact mistake just went up, on
  purpose.** By owner ruling the ungated class now **exits 1**, so a new ungated field
  turns the daily `component-fallback-check` Job RED. The fastest way to green is the
  NO-OP gate above, and it still works on this lint — it clears the finding and leaves
  the identical blank cell. What is different from 08-03 is that it no longer gets you
  away with it: `component-render-check` (CGV-030) renders the component and reports
  the hole from the output side, so the no-op edit turns one red job green and a second
  one red. **If you are clearing this finding under time pressure, run
  `/tmp/rck --compare` before and after your edit** — a fix that leaves NEW findings
  there did not fix anything.

---

### The thunder reaper's own smoke test supplies the one field whose absence breaks it

- **footprint:** `thunder_instances`, `internal/adapters/thunder/store/instances.go`, `docs/agent_docs/sql_for_agents/114_thunder_reaper.sql`, `thunder-reaper`
- **fires when:** proving the reaper works, or hand-inserting a row to clean up an
  instance seen at Thunder
- **the tell:** the drill passes. `store.Instance.InstanceIP` is a plain Go `string`
  against a **nullable** column, so `lookupOne` dies with *"converting NULL to string is
  unsupported"* — but `114_thunder_reaper.sql:186-199` builds its synthetic row **with
  `instance_ip = '10.0.0.42'`**, so the only documented proof can never reach that branch
- **the check:** run the drill **omitting `instance_ip`** — the omission *is* the test.
  Then read `thunder_instances.status`, not the clock: `last_triggered_at` advancing
  proves the SCHEDULER fired, never that anything was reaped. Runs are in
  `orchestration_states` under **`owner_agent_type`** (`agent_type` does not exist and
  errors — which reads as "no rows" if you suppressed stderr)
- ⚠ **the drill id is a safety control, not a label.** Real `thunder_instance_id`s are
  bare integers (`0`, `1`); use a **non-numeric** one, because
  `decommission_action.go:123-129` refuses an unparseable id *before* calling Thunder.
  The seed's `999999` instead relies on Thunder 404ing — weaker. And DELETE the row after
  (match `id` AND `thunder_instance_id`): one left in `decommissioning` is re-selected
  every 900s for ever by the widened selector (`sql_for_agents/280`)
- **a state the schema models is not one the code writes:** `042_thunder.sql` references
  `status='provisioning'`; nothing writes it (`provision_action.go:413` hardcodes
  `'running'` and INSERTs only *after* the box is up) — which is also why the real
  exposure is an instance **billing at Thunder with no row at all**, invisible to every
  check we own. The vendor API is truth; our table is a cache
- **source:** `bugs_open/186`, `finetuning_uk_service/NOTES` 2026-08-03
- **added:** 2026-08-03, finetuning_uk_service lane

## A concept-register STATUS line is a snapshot that outlives its truth — and council seats read it as ground truth

- **footprint:** `docs/agent_docs/docs026_concept_register/register/`, `000_concept_index.md`, `MDL-038`
- **fires when:** citing a register entry's status ("deployed as a bug", "unfixed",
  "nothing consumes X") as authority — in a submission, a review seat, or a plan —
  without grepping the source for the mechanism the status describes
- **the tell:** the citation is precise, dated and CONFIRMED-graded, so it reads as
  stronger evidence than a fresh look. But the register froze 2026-07-13 (bugs_open/106)
  and hand-patched entries do not update when fixes land. Proven cost 2026-08-03:
  MDL-038 said "GenerateText never decodes stop_reason — confirmed present and unfixed"
  while the fix had been live since 2026-07-20 (a3b606798, typed TruncatedError, every
  provider, tests); council fee9d810's llm_reliability seat read the entry and GATED a
  correct submission on the missing mechanism. The wrong result (a REVISE) looks exactly
  like the right one (a caught defect)
- **the check:** before repeating any register status as evidence, grep the cited
  source for the mechanism itself (`grep -n "stop_reason\|StopReason" platform/aiservice/*.go`
  beat the entry in one command). A status older than the file it describes
  (`git log -1 -- <source>` newer than the entry's date) is a stale-status suspect.
  When you catch one, correct the ENTRY visibly (strike-through + date + cost), not
  just your own doc — the next reader inherits whichever you fixed
- **source:** council fee9d810 r1→r2, `vigilant_designer_offer_analysis/NOTES` 2026-08-03
- **added:** 2026-08-03, vigilant_designer_offer_analysis lane

### The documented scratchpad is a SHARED 16G tmpfs — a git-archive build context there fills /tmp for every session at once

- **footprint:** `/tmp/claude-1000/`, `git archive HEAD | tar -x`, `go build` (GOTMPDIR), any "verify against committed HEAD" check
- **fires when:** following the harness's own instruction ("always use this scratchpad
  directory for temporary files") for a HEAD-check build context — `git archive HEAD`
  extracted + `go build` is ~450MB per iteration, iterations accumulate (one session held
  five at 3.5GB), and /tmp is one 16GB tmpfs shared by EVERY session on the machine
- **the tell:** `no space left on device` from a go build whose code is fine — or, subtler,
  a harness tool-result reading "Command output was lost … ENOSPC", which looks like a
  tooling glitch and is actually the machine-wide symptom. Measured 2026-08-03 11:55:
  /tmp 100% full, 456K free, top consumers five sessions' scratchpad headcheck dirs
  (447MB each) while every concurrent session's builds were failing
- **the check:** `df -h /tmp | tail -1` BEFORE creating any multi-hundred-MB context; put
  HEAD-check contexts on disk instead (`mkdir -p /home/ant/tmp/headcheck-<lane>`;
  `GOTMPDIR=/home/ant/tmp go build ...` — / has ~235G free), and `rm -rf` the context in
  the same session that made it. Your own session dir being small proves nothing — the
  tmpfs is shared, so the polluter and the victim are usually different sessions
- **source:** bugfix_177_tool_content_items/NOTES 2026-08-03 (hit mid-implementation; the
  fix agent's builds were pre-emptively redirected and succeeded)
- **added:** 2026-08-03, bugfix_177 lane

### Adopting the retraction seam on a CO-DEDUP'd `item_type` closes the OTHER producer's finding — and the two producers' evidence can be positively correlated
- **footprint:** `CheckResult.Resolved`, `resolveWorkItems` (`work_items_common.go`), `undeployed_asset`, `write_render_audit_findings_action.go`, any `item_type` filed from more than one place
- **fires when:** you adopt RFC_010's retraction seam (WII-009) on a check and reason, correctly, that "my check raised these items, so my check may close them". The seam addresses rows by `(site_id, item_type, item_key)` — which is **coarser than the producer set that shares that key**. The owner's ruling of 2026-08-02 (RFC_010 *who may answer a page name collision*) explicitly **blesses and encourages** converging N producers onto one `item_type`/`item_key`, so the shape is spreading, not shrinking: **13 item types have ≥2 Go producers today**
- **the tell:** **none at the call site.** Your check's own code, tests and item_key all look self-consistent; the other producer is in a different package and never mentions you. The retraction UPDATE names no producer, so a wrongly-closed row is indistinguishable from a rightly-closed one afterwards
- **the worked case, which is worse than "they might disagree" — their evidence is POSITIVELY CORRELATED.** `undeployed_asset` is filed by BOTH `check_undeployed_assets` and `write_render_audit_findings_action` ("same item_type, same key namespace, same handler — deliberate co-dedup"). The render audit's finding is *"this image serves broken on a real page"*; `check_undeployed_assets` treats *"a deployed page_component's rendered_html references the filename"* as healthy. **You cannot have a broken `<img>` unless the HTML references its src** — so every render-audit 404 finding sits on an asset that the other check reads as healthy. Adopting retraction there would have retracted **every** render-audit 404 on the next sweep, silently, fleet-wide
- **the check:** before adopting, count the producers — do NOT assume your check is the only one:
  `grep -rn --include=*.go -E '(ItemType|itemType):[[:space:]]*"<your_type>"' platform/ internal/ | grep -v _test`
  If it returns more than one file, ask the harder question: **is my positive observation a REFUTATION of the other producer's finding, or merely unrelated to it?** Unrelated is not good enough — retraction closes the row either way. Prefer a single-producer type for a first adoption; if you must adopt on a shared one, the observation has to refute *every* producer's predicate, not just your own
- **source:** 2026-08-03, `bugfix_168_deployed_asset_path` lane, choosing the first real adopter of WII-009. `check_undeployed_assets` was the obvious candidate (95 open items, the largest stale population in the queue) and was **rejected on this ground**; `check_empty_sections` was adopted instead because it has exactly one producer
- **added:** 2026-08-03, bugfix_168_deployed_asset_path lane

### A monotonic check's `if len(findings) == 0 { return }` early return makes its new retraction INERT on exactly the sites that need it
- **footprint:** `CheckResult.Resolved`, any `discovery_checks/check_*.go` `Run` adopting retraction, `check_empty_sections.go`
- **fires when:** you add retraction to an existing discovery check by appending it to the end of `Run` — the natural, minimal, diff-friendly edit. Most checks open with a guard of the shape `if len(findings) == 0 { return &CheckResult{}, nil }`, which is **correct** while a check can only FILE (nothing found, nothing to say) and **exactly backwards** once it can also RETRACT
- **the tell:** **every test passes and the mechanism does nothing.** A site with zero findings is the ONLY site the early return fires on, and it is precisely the site whose stale items need closing — so the retraction works in every test that has a finding, works in every test you would naturally write, and is inert in production wherever it matters. `items_resolved` stays 0 and reads as "nobody adopted it" rather than "adopted and unreachable"
- **the check:** write the zero-findings retraction test FIRST, and prove it by mutation — reinstate the early return and require the test to fail. `TestRetractionRunsWhenThereAreNoFindingsAtAll` in `check_empty_sections_test.go` is that test; mutation M1 in the 2026-08-03 round is that proof
- **source:** 2026-08-03, found while adopting WII-009 on `check_empty_sections` — the guard was in the file already and the append-to-the-end edit would have shipped a mechanism that was green everywhere and inert everywhere
- **added:** 2026-08-03, bugfix_168_deployed_asset_path lane

### The comment explaining a removal makes the removed symbol's negative control non-zero — your own fix documents itself into looking unshipped
- **footprint:** any "expect 0 occurrences of the string my change REMOVED" check; `strings /app/agent-chassis | grep -c`, `curl <asset> | grep -c`, `scripts/rerender_*`, `assets/js/snippets.js`, vonc seal verification
- **fires when:** you verify a deploy the way this estate tells you to — grep the artefact for a string the change ADDED (expect >0) and one it REMOVED (expect 0) — and the change is one a future reader will need explained, so you wrote a comment saying *"this used to paint `today.headline` + `today.body`, which is why it no longer does"*. The comment ships inside the same artefact. The negative control now counts your own prose and returns non-zero
- **the tell:** **the count looks exactly like an unshipped fix, and the better-documented the change, the worse it reads.** Measured 2026-08-03 on vonc's served `snippets.js`: `today.headline` = 2 and `today.body` = 2, all four inside the seal's explanatory block; the correct probe is the minified local `t.headline`, which is 0. A thread reading `2` and stopping would conclude the seal never shipped and "re-fix" a live, working feature
- **the check:** **print the context, never the count alone** — `grep -o -E '.{140}<symbol>.{140}'`, or in Python a `re.finditer` window; then classify each hit as code or prose before reading the number. And pick the probe from the SHIPPED form: after minification the live identifier is `t.headline`, not `today.headline`, so the source-level spelling is the wrong string in both directions
- **the sibling, same class and different cause:** a negative control's expected value is not always 0 because **other files may legitimately carry the spelling** — the 2026-08-02 chassis verification read `2` on a removed query spelling where 2 *was* the pass value and 3 would have meant stale (`gauntlet_dead_cta/NOTES` 2026-08-02). Between them the rule is: **derive the expected count from HEAD before you run the grep, and never default it to 0**
- **source:** 2026-08-03, `gauntlet_dead_cta` lane, re-verifying HANDOFF C's seal three days after it shipped. Caught because the count contradicted a render-level sweep that had just reported 0 of 20 pages leaking — two instruments disagreeing is what forced the context print
- **added:** 2026-08-03, gauntlet_dead_cta lane

## `page-build-handler`'s content writer never sees a page's OWN stored prose unless `spec.mode="recreate"` — and even then it loads the ORIGINAL adoption crawl, not the current content

- **footprint:** `platform/orchestration/actions/load_existing_content_action.go`, `create_tool_cross_link_items.go`, `apply_gap_plan_action.go`, any emitter of a `content_rewrite`/`needs_content_page` item, `agent_definitions` step `call_content_writer` (`page-build-handler`)
- **fires when:** an item asks `page-build-handler` to edit or extend a section on a page that ALREADY has content, and the emitting action never sets `spec.mode`
- **the tell:** the item reports `complete`, the new content is well-formed and does contain what was asked for (a link, a fact), but a same-named prose slot has been silently rewritten from scratch — shorter, restructured, a changed heading. Looks like a regeneration bug in the writer; it is a missing input. `load_existing_content_action.go:64-69` no-ops (`{"has_existing": false, "reason": "not_recreate"}`) unless `mode=="recreate"`; `call_content_writer`'s `input_mapping` passes that no-op result as `existing_content?` and `current_page: page_record`, and `load_page_record`'s own description says it carries only "sections, title, page_type" — no prose. So the writer receives the item's guidance text and NOTHING to edit, and must fabricate a replacement that satisfies the instruction's shape. Confirmed on `bugs_open/178` (item `93f2a3b7`: `create_tool_cross_link_items.go` never sets `mode`, `apply_gap_plan_action.go`'s content_rewrite emission unchecked — flag for whoever touches it)
- **the second trap, for anyone tempted to "fix" this by setting `mode=recreate`:** that gate, even satisfied, sources `research_results` — the ORIGINAL adoption-crawl snapshot — never the page's current `page_components.content_data`. Setting it feeds the writer stale pre-edit content, not none. There is today no workflow channel that passes a page's LIVE stored section content to its own writer for editing
- **the check:** before raising any `content_rewrite`/`needs_content_page` item against a page you know has content, either confirm the emitter sets `spec.mode` (and accept that "recreate" means adoption-crawl content, not current), or treat the item as writing a section FROM SCRATCH and expect prior text to be lost. A 090 diagnosis of "why did my edit-item overwrite the section" will very likely land here — check this entry before re-running one
- **source:** `bugs_open/178`, root cause confirmed 2026-08-03 (`bugfix_154_work_item_routing_columns` lane), completing a 5-iteration automated diagnosis (run `aece2920`) that stopped UNVERIFIABLE naming this exact config as the missing evidence
- **added:** 2026-08-03, bugfix_154_work_item_routing_columns lane

### A claims finding printed by `cmd/claimscan` is judged from a 100-char snippet — and a citation for a figure lives just outside that window
- **footprint:** `cmd/claimscan` output, `claimSnippet` (`platform/orchestration/datahelpers/claims.go:993`), any dry run over `page_components.rendered_html` used to judge a candidate pattern
- **fires when:** you dry-run a new or widened claims pattern over the live corpus and read the printed findings to decide which are real. It fires hardest on the pattern class you most want to add — anything about **sourcing** (unsourced figures, uncited statistics, unsupported claims)
- **the tell:** **there is none, and the output looks authoritative.** `claimSnippet` prints 60 bytes either side of the match. A figure's citation sits in the sentence BEFORE it, or in the component before it — i.e. **just outside that window, in the direction the window is narrowest**. On 2026-08-03 this produced 9 confident "real unsourced statistics" out of 14, of which at least 3 were cited (a University of Melbourne study named in the preceding sentence), a statutory definition ("at least 51% of people accepted" — what a representative APR legally means), or anaphoric to a figure cited earlier on the page. Reading the snippet is reading the instrument
- **the check:** re-read every candidate finding in **≥300 characters of context**, or in its full component, before any count goes into a doc. The corpus is already exported as a TSV, so it is one loop:
  `python3 -c "import base64,re;[print(re.sub(r'\s+',' ',re.sub(r'<[^>]*>',' ',base64.b64decode(l.split(chr(9))[2]).decode('utf8','replace')))[max(0,i-250):i+120]) for l in open('corpus.tsv') for i in [re.sub(r'\s+',' ',re.sub(r'<[^>]*>',' ',base64.b64decode(l.split(chr(9))[2]).decode('utf8','replace'))).find('<your needle>')] if i>=0]"`
- **and the structural consequence, which is the durable half:** because the exonerating evidence routinely sits in a **different block or a different component**, a citation-style check must be **DOCUMENT-scoped, not block-scoped**. Measured the same day: block-scoped scored **0 true / 4 false** over 1,130 live components; document-scoped scored **0**, while still firing on the motivating fabrication. `ScanAttributedUncitedStats` asks `DocumentCarriesCitations` first for exactly this reason
- **source:** `bugfix_123_content_creator_claims` lane; `WRONG_CALLS.md` 2026-08-03; concept register CLM-019
- **added:** 2026-08-03, bugfix_123 lane

### `negatedClaimMatch` scans BACKWARDS from the match start — a pattern that spans subject-to-verb swallows its own negation
- **footprint:** `platform/orchestration/datahelpers/claims.go` `negatedClaimMatch`, `negationCueRe`, and any new scan in this layer that calls it
- **fires when:** you add a claims pattern whose match begins at the SUBJECT and ends at the VERB — the natural shape for an attribution cue ("industry data … shows", "studies … find"). `negatedClaimMatch(block, loc[0])` then reads only the text **before the subject**, while the negation sits **between** subject and verb, inside the matched span
- **the tell:** **the verdict inverts silently and in the dangerous direction** — every DENIAL is reported as an assertion. "Industry data **does not** show that 40% of teams do this" becomes a finding. Nothing errors; the guard is called, appears in the code, and returns false every time. CLM-017 recorded the same geometry for the fleet-wide patterns, where the negation also sits inside the match — **there it is harmless** (those patterns' negations are part of what is banned), which is why the entry reads as reassurance and does not warn you
- **the check:** call it at **`loc[1]`** (the end of the match, i.e. the verb), not `loc[0]`. It stays clause-local either way — the backwards scan still stops at the first `.!?;:,<>` — so nothing else changes. Prove it with a denial fixture per cue verb, and mutation-neuter the guard to confirm the fixtures fail without it: three unit tests caught this; re-reading the code did not, twice
- **also:** bare `no` is **deliberately excluded** from `negationCueRe` (it appears as an intensifier — "There are no exceptions: every claim is verified"), so "No industry data shows…" IS still reported. Do not add a local fix for that: a second negation implementation in this layer is what CLM-004 exists to prevent. Pin it as an inherited residual instead
- **source:** `bugfix_123_content_creator_claims`; concept register CLM-019, CLM-017
- **added:** 2026-08-03, bugfix_123 lane

### `sites.github_branch` says `main`; the deploy repo has no `main` — deploys work only because nothing passes the column
- **footprint:** `sites.github_branch`, `internal/adapters/git/github_client.go` `CommitToRepo`, `GitCommitData.Branch`, `GitDeleteFileData.Branch`, `platform/orchestration/actions/git_deployer_actions.go`, `deploy_image_asset_action.go`, any new caller assembling a git-adapter payload
- **fires when:** you write a new action that commits to (or deletes from) a site repo and helpfully populate `branch` from the site record, because the column is right there and named exactly what you need
- **the tell:** **there is none at the payload — it fails at GitHub, and looks like an auth or a repo problem.** `gqls/sites` carries `master` (its default) and `750start`; there is **no `main`**. `sites.github_branch` is `'main'` for most rows. Every deploy that works today works because `GitCommitAction` leaves `Branch` empty and `CommitToRepo` falls back to `repo.DefaultBranch`. Pass the column and you target a branch that does not exist — and on a repo where it DOES exist but is not the default, you get the worse outcome: a successful commit to a branch the B2 workflow does not watch (`on: push: branches: [master]`), so the site never updates and nothing errors
- **the check:** leave `Branch` unset for site content, and let `CommitToRepo` resolve the repo default. Before trusting the column anywhere, ask GitHub rather than the DB: `gh api repos/<org>/<repo>/branches --jq '.[].name'` and `gh api repos/<org>/<repo> --jq '.default_branch'`. A branch-targeted commit is for PLATFORM-repo work (the fix-implementer's fix branch), which is the only case that legitimately sets it
- **source:** `bugfix_098_unpublish_primitive` lane, 2026-08-03; concept register DGH-006
- **added:** 2026-08-03, bugfix_098 lane

### `TreeEntry.SHA` is a `*string` on purpose — "tidying" it to a plain string silently disables every file deletion
- **footprint:** `internal/adapters/git/interface.go` `TreeEntry`, `internal/adapters/git/github_client.go` `CommitToRepo`
- **fires when:** anyone reads `SHA *string` as an un-idiomatic pointer on a small struct and simplifies it, or adds `omitempty` while formatting the struct tags
- **the tell:** **writes keep working, so every existing test and every deploy passes.** GitHub's tree API expresses "remove this path" as an entry whose `sha` is JSON `null`. A plain `string` marshals an unset SHA to `""`, which the API rejects; `omitempty` DROPS the key, which is not a deletion but a malformed entry. Nothing in the codebase except the null-sha test would notice, and the failure surfaces only on the unpublish path — which has no live caller yet, so it can rot broken for months
- **the check:** `TestDeletionIsSentAsNullSHA` (`github_client_delete_test.go`) decodes the tree payload as **raw maps**, not into `[]TreeEntry` — decoding into the struct turns a missing key into a zero value and hides the exact null-vs-`""`-vs-absent distinction under test. Keep it that way. Probed on the live repo 2026-08-03: null sha on an existing path → 201; on an absent path → **422 `GitRPC::BadObjectState`** (which is also why the existence filter is required, not defensive)
- **source:** `bugfix_098_unpublish_primitive` lane; concept register DGH-006
- **added:** 2026-08-03, bugfix_098 lane

### `build_status='deployed'` is not "the platform still wants this page served" — a selector keyed on it alone RESURRECTS retired pages
- **footprint:** `platform/orchestration/actions/render_news_section_html.go` `queueNewsPageRerenders`, `render_directory_action.go` `queueDirectoryPageRerenders`, any query choosing pages to re-render/re-publish/re-list, `pages.build_status`, `pages.status`
- **fires when:** you write (or review) a page-selection query and reach for `build_status = 'deployed'` as the liveness test, which reads as "this page is live" and is not
- **the tell:** **none at the query, and the damage is invisible in the DB** — it shows up as a deploy-repo commit history. `build_status` records whether the page ever shipped; `status` records whether it should still be served. **Archiving sets `status` and leaves `build_status` untouched**, so an archived page stays selected for ever. Live instance: `robot-hands.com/learning-center-index` (archived) was re-rendered and re-committed **twice a day** from 07-31 to 08-03, six work items, while `bugs_open/098` described archived pages as "frozen". It also makes any retraction **self-undoing** — delete the file, the next refresh republishes it, and a post-delete `curl` still shows 404 at the moment you look
- **the check:** pair them — `AND p.build_status = 'deployed' AND p.status = 'active'` — as `load_site_pages_action.go:80`, `plan_sections_action.go:247`, `maintenance_actions.go:751`, `request_render_audit_action.go:110` and `store_generated_component_action.go:873` already do. Census the whole class rather than the one call site you were sent to: the two functions that had the defect are the two that call each other "cousin" in their own comments, which is exactly how a reader concludes they behave alike. To see it in the estate: `SELECT status, count(*) FROM pages GROUP BY 1` (only `active` and `archived` exist) then run your predicate with and without the `status` arm and diff the row sets
- **source:** `bugs_open/098` correction 2026-08-03; commits `5b66615d4`, `8f73e7279`; `bugfix_098_unpublish_primitive` lane
- **added:** 2026-08-03, bugfix_098 lane

### `go run` collapses the child's exit status — your carefully-chosen exit 2 arrives as exit 1
- **footprint:** `go run` inside any `scripts/*.sh` wrapper, `cmd/config-key-audit`, exit-code branching on a Go tool run from source
- **fires when:** a wrapper script branches on the exit code of `go run ./cmd/<tool>` — e.g. distinguishing a tool's "findings, exit 1" from its "refusing to report, exit 2" — because the tool's own os.Exit codes are documented and you reasonably branch on them
- **the tell:** there is none at the call site: `go run` prints `exit status 2` to stderr and then itself exits **1**, so every non-zero child status ≥ 1 arrives as 1 (measured 2026-08-03, bugfix_134 lane). A branch on `rc > 1` is dead code that never fires and looks exactly like "the refusal case just never happens"
- **the check:** discriminate on the tool's OUTPUT, not its code — an empty stdout where JSON findings belong is the refusal (`scripts/audit-config-keys.sh` does this for all three of its `go run` captures, with the reason in a comment). If the code itself matters, `go build -o` to a temp path and run the binary, whose status arrives intact
- **source:** `bugfix_134_optional_marker` lane, 2026-08-03 — found writing the `--suspicious-keys` capture; the first draft branched on `rc > 1` and would have silently never taken the refusal path
- **added:** 2026-08-03, bugfix_134 lane

### Firing `section_data_resolved` on a LOCKED, positionally-named section duplicates it, not protects it
- **footprint:** `platform/orchestration/actions/save_page_sections_action.go` (`extractSectionsFromMetadata:896-902`, `matchLockedRow:586`), `page_components.locked_at`, `rerender_page_sections` any site with slot_names that are not component identities (loancalculator.co.uk, oufe.com)
- **fires when:** you fire a `section_data_resolved`/`image_landed` re-render on a page where the target slot is BOTH locked AND previously unresolvable by name (fixed to resolve via `component_id` by `bugs_open/182`) — 14 such sections exist right now: loancalculator.co.uk (12), oufe.com (2)
- **the tell:** none at fire time — the work item goes `complete`. The page silently gains a duplicate `<section>`: the locked row's own `slot_name` (e.g. `tool-2`) never matches the incoming section's post-resolution name (e.g. `tool-loan-vs-savings`, since `extractSectionsFromMetadata` prefers `component_function` over `component_name` once a component resolves), so the 058 lock guard's name match misses, the fresh copy gets INSERTED instead of discarded, and the pre-existing locked row (excluded from the DELETE-all by its own protection) survives alongside it — same `component_id` twice, byte-near-identical `content_data`, on one page
- **the check:** before firing on any of the 14 armed sections, `SELECT count(*) FROM page_components WHERE page_id='<id>' AND component_id='<id>';` before AND after — expect it unchanged. If it goes up by one, you just hit this; remediate by deleting the fresh unlocked duplicate and repositioning the locked row back to its old slot (`bugs_open/189` records the exact SQL used to reverse it live on loancalculator.co.uk 2026-08-03)
- **source:** `bugs_open/189`, found inducing the `bugs_open/182` verification, 2026-08-03 — reproduced once, remediated live same session
- **added:** 2026-08-03, bugfix 182 lane

### The retraction-killing early exit is not always the LEADING guard — a mid-loop `return` (a per-pass cap) hides from the grep the existing entry teaches
- **footprint:** `CheckResult.Resolved`, any `discovery_checks/check_*.go` `Run` adopting retraction, `maxRequiredFieldFlagsPerPass` and any other per-pass cap / `if emitted >= N`, `check_required_fields_missing.go`
- **fires when:** you follow the entry above ("a monotonic check's `if len(findings) == 0 { return }` makes retraction inert") to the letter, check the top of `Run`, find no leading guard, and conclude the check is safe to adopt on
- **the tell:** **none, and it is worse than the leading-guard case.** `check_required_fields_missing` has no `len(findings) == 0` guard at all — its retraction-skipping `return result, nil` sits **in the middle of the row loop**, fired by a noise cap at 25 findings. So it is inert on exactly the badly-shaped sites that carry the most stale items, green in every test that stays under the cap (all of them), and invisible to a grep for the documented shape. The cap's `return` was *correct* while the check could only file — a noise bound should stop early — and became a bug the instant the check could also retract
- **the check:** read EVERY exit between the scan and the retraction, not just the leading guard: `awk '/^func .*Run\(dctx/,/^}/' <check>.go | grep -n 'return'` and ask of each whether the retraction below it still runs. Fix by `break`, not by moving the retraction earlier. Prove it: build a filing result LARGER than the cap and assert a retraction still comes back — `TestRequiredFieldsRetractionSurvivesThePerPassCap`, mutation M1 of the 2026-08-03 round
- **the armed sites, CENSUSED 2026-08-03 so you do not have to** (asked for by council `64430363`'s `bug_historian` seat — "this is the second per-call-site fix of the same shape and nothing audits the rest"). Five `discovery_checks/check_*.go` carry a per-pass cap. Three already `break` and are safe: `check_componentless_pages.go`, `check_component_template_corrupted.go`, `check_section_source_drift.go`. ~~`check_required_fields_missing.go` was the `return` and is fixed.~~ **CORRECTED 2026-08-04: that fix was REVERTED** (owner option A — the adoption it enabled duplicated `revalidate_review_queue`; see `WRONG_CALLS.md` 2026-08-04). That file is back to `return result, nil` and is **armed again** — correctly so, because it no longer retracts. **So TWO sites are armed-but-inert: `check_required_fields_missing.go` and `check_image_source_unsatisfiable.go` (`if emitted >= maxUnsatisfiableFlagsPerPass { return result, nil }`).** It is INERT today — that check does not populate `Resolved`, and its `return` is *correct* while a check can only file. **The trap fires the moment someone adopts the seam there, and that adoption is the commit that must also change it to `break`.** Re-run the census with: `for f in check_*.go; do awk '/^func \(c \*.*\) Run\(dctx/,/^}$/' "$f" | grep -qE 'if +[a-zA-Z]+ +>=? +[a-zA-Z]*[mM]ax' && echo "$f: $(awk '/^func \(c \*.*\) Run\(dctx/,/^}$/' "$f" | grep -A6 -E '>=? +[a-zA-Z]*[mM]ax' | grep -cE 'return result')"; done`
- **source:** `bugfix_168_deployed_asset_path` lane, second RFC_010 adopter, 2026-08-03; commit trailer `Council-Submitted`
- **added:** 2026-08-03, bugfix_168 lane

### When your predicate is "the CONFIG declares X and the DATA lacks X", an unreadable config computes to HEALTHY — and retraction then closes the item
- **footprint:** `findResolvedRequiredFields` (`check_required_fields_missing.go`), `datahelpers.SchemaContentFields`, `content_components.input_schema`, `CheckResult.Resolved`, any check whose defect is "declared-but-absent" (required fields, declared slots, declared assets, `on_missing` enforcement)
- **fires when:** you adopt the retraction seam on a check whose predicate derives its EXPECTED set from a schema/config row rather than from the observation itself. The filing half almost certainly `continue`s when that row is NULL, unparseable, or in an unrecognised dialect — which is correct there (no schema, nothing to assert)
- **the tell:** **the arithmetic silently inverts.** No schema → no required fields → the "missing" list is empty → the slot reads as *filled*. The component is present, deployed and looks perfectly healthy; only the thing that says what it OUGHT to contain has gone. So the retraction fires hardest exactly when a component's schema was dropped — the same silent-loss class as a deleted component (`bugs_open/012`, `/021`, `/032`), arriving by a route that reads as success. Copying the filing half's `continue` straight across is what does it, because there `continue` means "do not file" and here it must mean "do not count as observed"
- **the check:** make the healthy counter reachable from ONE place — the line where the predicate actually ran and returned nothing — and let every other path fall through to a `healthy != deployed` gate, so a new refusal is added by NOT counting rather than by remembering to veto. Then mutate: make the unreadable-config branch increment `healthy` and require a test to fail (M3/M4 of the 2026-08-03 round). ⚠ **and check your mutation is not being caught by a guard in SERIES** — deleting the "no deployed component" refusal left every test green, because today's join makes a LEFT JOIN miss always carry a NULL schema too, so the schema guard shadowed it. It took a synthetic row (miss + healthy-looking columns) to pin that guard on its own
- **source:** `bugfix_168_deployed_asset_path` lane, second RFC_010 adopter, 2026-08-03
- **added:** 2026-08-03, bugfix_168 lane

### `pages.rendered_header` / `rendered_footer` / `rendered_head` are VESTIGIAL — empty on all 562 pages fleet-wide, so a site with empty ones tells you nothing
- **footprint:** `pages.rendered_header`, `pages.rendered_footer`, `pages.rendered_head`, `site_components`, `discovery_checks/check_missing_structure.go`, any "this page has no header/nav/footer" investigation
- **fires when:** a rebuilt page ships with no `<header>`/`<nav>`/`<footer>`, you run `\d pages` to find where chrome is stored, and three columns with exactly the right names come back empty for the site you are looking at. Nothing about that reads as a dead end — it reads as the answer
- **the tell:** **there is none for a single site, and that is the whole trap** — a scoped query returns "empty" and a fleet query returns "empty", and only the second one is informative. Chrome actually lives in **`site_components`** (`slot_name` in header/footer/head), written by `render_site_components`. The three `pages` columns are read by exactly one caller left in the tree — `check_missing_structure.go:94-100`, which flags `rendered_header IS NULL` and therefore flags *every page on every site*. Confirmed 2026-08-03: `loanandmortgagecalculator.co.uk` serves full chrome on 41 pages with all three columns empty AND zero `site_components` rows (chrome is baked into the deployed artefact at render time, so deleting the source rows does not un-ship it — which is why "the sibling works" cannot be read backwards to "the sibling has rows")
- **the check:** never scope the census to your own site — the `WHERE domain=` is what manufactures the false positive. `SELECT s.domain, count(*) FILTER (WHERE coalesce(length(p.rendered_header),0)>0) FROM pages p JOIN sites s ON s.id=p.site_id GROUP BY 1;` → 0 for all 19 domains. Then ask `site_components` instead: `SELECT slot_name, build_status, length(rendered_html) FROM site_components WHERE site_id=…;` — **zero rows there is the real "no chrome" signature**, and the fix is `nav-updater` (`populate_nav_tables → render_site_components`), not a column
- **source:** `mortgagecalculator_couk_adoption` lane, 2026-08-03 (canary rebuild; NOTES + HANDOFF §12)
- **added:** 2026-08-03, mortgagecalculator adoption lane

### `include_statuses: ["deployed","active"]` filters `pages.status`, where `'deployed'` NEVER OCCURS — the config reads "deployed pages only" and selects every non-archived page
- **footprint:** `get_pages_for_rerender_action.go:96,153`, `nav-updater` step `get_pages` (`config.include_statuses`), `pages.status` vs `pages.build_status`, `create_rerender_items`
- **fires when:** you reason about the blast radius of a nav rebuild or a site-wide reassemble — "it only touches deployed pages, so an unbuilt homepage is out of scope". The step's own `description` is *"Get all deployed pages for reassembly"*, the config value is the literal string `deployed`, and both are wrong about which column they mean
- **the tell:** **`pages` has BOTH a `status` and a `build_status`, and `'deployed'` is a value of the second one only.** Line 153 emits `AND p.status IN (…)`. Fleet-wide `pages.status` takes exactly two values — `active` (561) and `archived` (27) — so the `"deployed"` element matches **nothing** and `"active"` matches **everything not archived**. A site with 1 deployed page and 25 `build_status='planned'` ones gets **all 26** fanned out into `page_rerender` items, homepage included. Nothing errors and no count looks wrong, because the number you expected was never printed
- **the check:** `SELECT status, count(*) FROM pages GROUP BY 1;` before you trust any `include_statuses` reasoning — if `deployed` is absent from that list, the filter is inert. What actually protects an unbuilt page is one branch further downstream, not this filter: `rerender_single_page_action.go:565` returns empty HTML for a page with **zero `page_components`** rows and `:168-209` converts that to `skipped:true` with no deploy. So the safety property is "has no component rows", **not** "is not deployed" — and a page that HAS component rows and is not deployed will deploy. Confirm per site with `SELECT p.name, count(pc.id) FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id WHERE p.site_id=… GROUP BY 1;` and defer the items anyway rather than resting an invariant on a downstream branch
- **source:** `mortgagecalculator_couk_adoption` lane, 2026-08-03 (pre-flight before a nav rebuild on a live production site; HANDOFF §12)
- **added:** 2026-08-03, mortgagecalculator adoption lane

### The code index is SINGLE-COMMIT and days behind HEAD — a fix to the LOOKUP verified against a HEAD-only symbol looks exactly like a broken fix
- **footprint:** `code_symbols` (`commit_sha`), `platform/orchestration/actions/diagnose_code_lookup_action.go` (`answerCodeCheck`, `parseSymbolQuery`, `symbolClauseFor`), `scripts/trigger-landmine-verifier.sh`, `index_code_symbols`
- **fires when:** you change or verify anything that READS the code index — the diagnose loop's code tier, the council's `code_checks`, the landmine-verifier — and pick a probe symbol from the code you just wrote
- **the tell:** the whole index sits at ONE `commit_sha` (`SELECT commit_sha, count(*) FROM code_symbols GROUP BY 1` → a single row, 2026-08-03: `d98010e8b…`, 4,992 rows, indexed 07-28, ~1,600 commits behind HEAD). Your new symbol is legitimately absent, the lookup correctly returns 0 rows, and a CORRECT fix reads as broken — the same 0 you were trying to eliminate, from a cause the answer's freshness banner names but a hurried reader skips. The inverse fires too: a symbol you DELETED at HEAD still resolves
- **the check:** before using any symbol as a lookup probe, `SELECT path, symbol FROM code_symbols WHERE symbol ILIKE '%<name>%';` — probe with one that is IN the index (`symbolTokenClause` and `ReadSymbolBody` both are, at d98010e8b). For a fix-is-live proof, assert the PREDICATE shape changed (0→1 rows on the same indexed symbol), never the presence of new code in the index
- **source:** `bugfix_163_symbol_lookup` lane, 2026-08-03 — the 163 fix's own verification had to route around this twice
- **added:** 2026-08-03, bugfix_163_symbol_lookup lane

### After the scratch tmpdir moved to disk, TWO scratch roots are live at once — `df -h /tmp` answers about a tree you may not be in
- **footprint:** `CLAUDE_CODE_TMPDIR`, `/tmp/claude-1000`, `~/.claude-scratch`, `scripts/scratch-report.py`, `scripts/scratch-git-snapshot.py`, any `df -h /tmp` capacity check
- **fires when:** you check scratch capacity, hunt for another session's working files, or clean up — any time after 2026-08-03, when `CLAUDE_CODE_TMPDIR` was set to `/home/ant/.claude-scratch` in `~/.claude/settings.json` to stop the 16 GB tmpfs filling. **A running session keeps the tmpdir it launched with**, so sessions started before the change keep writing to `/tmp` while new ones write to disk, and both trees stay populated for as long as the old sessions live
- **the tell:** **none — both answers look authoritative and neither is wrong, just partial.** `df -h /tmp` can report healthy while the tree you are actually in is full, or report 100% full while your own writes are going to a disk with 212 GB free. Your own session dir being small proves nothing either way: the tmpfs is shared, so the polluter and the victim are usually different sessions
- **the check:** ask where YOUR scratchpad actually is rather than assuming — `echo "${CLAUDE_CODE_TMPDIR:-/tmp}"` — and when measuring the estate, read **both** roots. `scripts/scratch-report.py` does this by design (its `ROOTS` list carries the legacy `/tmp` entry precisely for this window); a hand-rolled check that inspects one root will be confidently wrong until the pre-move sessions die
- **the related trap, same change:** a full tmpfs surfaces as a Bash call returning **no output at all** — ENOSPC on the harness's own stdout capture — which reads like "the command found nothing", not like a disk error. If a command that should print something prints nothing, check `df` before believing the empty result
- **source:** 2026-08-03, gauntlet_dead_cta lane, moving scratch to disk after `/tmp` hit 100% (12.2 GB of scratchpads, 99.3% of it reproducible `git archive` extractions). Registered as OPP-005
- **added:** 2026-08-03, gauntlet_dead_cta lane

### A queue `page_rerender` item with no `reason` in its spec is ASSEMBLE-ONLY — the item completes, the page deploys, and your `content_data` edit is not on it
- **footprint:** `site_work_items` `item_type='page_rerender'`, handler `page-rerender`, `rerender_page_sections`, `page_components.rendered_html`, `RUNBOOK_brochure_component_library.md` §queue-recipe
- **fires when:** you edit a section's `content_data` (or a component's `html_template`) and file the standard queue item to "republish" the page — the item goes `complete`, the page's `last-modified` moves, and you verify by the item status
- **the tell:** the served page still carries the OLD markup while the work item says `complete`. Measured 2026-08-03 on fundamentallyai: content committed 10:58, queue item complete 11:00:33, the section's `rendered_html` untouched until a direct `rerender_sections` dispatch at 11:04:39. Nothing errors — assemble-only stitches the STORED section html, which is exactly what it is for
- **the check:** after any rerender meant to pick up a content/template change, compare `page_components.rendered_html` (`updated_at`, or grep the new string) BEFORE trusting the item status — then verify at the served page. To actually re-render sections: `brochure_component_library/scripts/rerender_page_sections_direct.sh` (proven; passes `reason=section_data_resolved`), or put `'reason','section_data_resolved'` into the queue item's spec jsonb [UNVERIFIED]
- **source:** brochure_component_library lane, 2026-08-03 (086b: three pages "republished", zero of the edits served)
- **added:** 2026-08-03, brochure lane 2

### A `static`-source field in a component's `input_schema` OVERWRITES your stored `content_data` on every section resolve — authored copy comes back as the schema's fallback and nothing reports the swap
- **footprint:** `content_components.input_schema` (`fields.*.source='static'`), `rerender_page_sections` / the section-planner resolve pass, `page_components.content_data`
- **fires when:** you hand-author `content_data` for a section and any resolve pass later runs over the page — your value for a static-source key is replaced by the schema's `fallback` (and a `query.*`-source field, e.g. tool-cta's `items`, is regenerated from its query), while `llm`-source fields you wrote are left alone
- **the tell:** one field of your authored block reverts while its neighbours survive. Measured 2026-08-03: tool-cta `secondary_cta_label` authored as "Talk to us about AI tooling", back as "Learn how it works" (`source: static, fallback: "Learn how it works"`) in the same resolve that kept the authored `headline` and `description`
- **the check:** before authoring `content_data`, read the component's schema: `SELECT jsonb_pretty(input_schema) FROM content_components WHERE name='<component>' AND is_active;` — any key with `source: static` or `source: query.*` is the SCHEMA's to write, not yours. Author only the `llm`/unsourced keys and let the resolver own the rest
- **source:** brochure_component_library lane, 2026-08-03 (/tools.html build)
- **added:** 2026-08-03, brochure lane 2

### Chassis pod logs rotate away within MINUTES — a clean grep proves the last few minutes, not the incident

- **footprint:** `kubectl logs` on `agent-chassis` (and any chattier chassis-image service), `ORCHESTRATION_TAKEN_OVER`, any "did mechanism X fire?" question answered from pod logs
- **fires when:** you grep a chassis pod's logs for an event a few minutes old, get 0, and conclude the event did not happen. The coordinator emits dozens of debug lines per action, the container log rotates on size, and `kubectl logs` returns only the current file — the observable window can be under 5 minutes at busy times (measured 2026-08-03: a 21:08 grep still saw 21:03 lines; a 21:15 grep of the same pod saw nothing before 21:07)
- **the tell:** your grep matches SOME instances of the pattern (other orchestrations' takeovers) but not the one you induced — which reads as "the mechanism skipped mine" and is actually "mine was earlier than the rotation floor"
- **the check:** `kubectl logs <pod> | head -1` FIRST — if the oldest surviving timestamp is later than your event, the log cannot answer your question. For ownership takeovers specifically, the durable witness is `orchestration_state_audit`: `TakeOverOrchestration` is the only platform writer that preserves `version`, so `SELECT * FROM orchestration_state_audit WHERE orchestration_id='…' AND old_version=new_version` lists every takeover (and every manual psql UPDATE — your own stamp shows here too), outliving both rotation and pod deletion
- **source:** bugs_closed/075 close-out 2026-08-03; WRONG_CALLS same date (two wrong calls from this exact trap in one hour)
- **added:** 2026-08-03, 075 pickup session

## A fix under `internal/adapters/` or `internal/agents/` does NOT ship in the chassis — "the chassis rolled" verifies the wrong binary

- **footprint:** `internal/adapters/`, `internal/agents/`, `make build-agent-chassis`,
  `make deploy-agent-chassis`, any handoff's "roll and pod-verify" step
- **fires when:** you fixed code under `internal/adapters/<x>/` or
  `internal/agents/<x>/` and are about to name the roll that makes it live, or to
  pod-grep a binary to prove it shipped. No symptom precedes this: a chassis roll
  after your commit LOOKS like your deploy, and a pod-grep of `/app/agent-chassis`
  returning 0 reads as "not shipped yet" (or worse, as "another session's build
  missed my commit" — bugs_open/153's exact shape, entered from the other side).
- **the tell:** the positive-control grep reads 0 on a binary that genuinely rolled
  after your commit. Before concluding the pipeline dropped your change, ask
  whether that binary ever CONTAINED your package.
- **the check:** map the package to its image first:
  `grep -l "internal/adapters/webscrape" build/docker/backend/*.dockerfile`
  (or the equivalent for your path). This repo builds FOURTEEN backend binaries
  from one tree; `platform/` and orchestration code ship in `agent-chassis`, but
  each adapter/agent under `internal/` ships in its OWN image with its own
  `make build-<service>` / `make deploy-<service>` and its own pod label.
  Verify at `/app/<service>` on `-l app=<service>` pods — with the same
  positive + negative controls, which is what caught this.
- **source:** bugs_open/158 item 3, 2026-08-03 — the lane wrote "needs a chassis
  roll" into its own committed handoff for a `web-scrape-adapter` fix; the wrong
  target was caught by a 0 on the chassis grep, corrected same session
  (WRONG_CALLS has the fuller account).
- **added:** 2026-08-03, bugfix_158_reply_drop_sizing lane

## A DISABLED `scheduled_tasks` row shows a fresh `last_completed_at` — the column that looks like proof it is running is written by the agent, not the scheduler

- **footprint:** `scheduled_tasks.last_completed_at`, `scheduled_tasks.enabled`,
  `improvement-sweep`, `cmd/scheduler/main.go` (`loadDueTasks`, `stampCompleted`),
  any `notify_scheduler` step in an `agent_definitions` workflow
- **fires when:** you ask "is this scheduled job actually running?" and read
  `last_completed_at`. On `improvement-sweep` that column is minutes old while
  `enabled = false` and `last_triggered_at` is **2026-05-02** — three months stale.
  No symptom precedes this: the row reads as a healthy, currently-running job, and
  the owner's 2026-07-29 ruling that the improvement loop is stopped DELIBERATELY
  reads as though it had been quietly reversed.
- **the tell:** `last_completed_at` is recent while `last_triggered_at` is ancient.
  Those two columns can only diverge if something other than the scheduler wrote
  one of them — `stampCompleted` (`cmd/scheduler/main.go:343-348`) always advances
  **both** in one statement, so the scheduler can never produce that shape.
- **the check:** read `enabled` and `last_triggered_at`, and ignore
  `last_completed_at` for the "is it running?" question. The writer is the agent's
  own workflow step: `improvement-loop`'s `notify_scheduler` runs
  `UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'improvement-sweep'`
  on every completion, **including a hand-fired run**, keyed by NAME with no check
  that the scheduler dispatched it. Find every such writer with
  `SELECT type FROM agent_definitions WHERE default_config::text LIKE '%scheduled_tasks SET last_completed_at%' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;`
  To answer "did the SCHEDULER fire it", use `enabled` + `last_triggered_at`; to
  answer "did the agent run", count `orchestration_states` rows instead.
- **not a functional defect, and do not file it as one:** the redundant stamp cannot
  cause a double-dispatch. `cmd/scheduler/main.go:287` calls `stampCompleted`
  immediately after `fireTrigger`, so for a message-firing task the in-flight guard
  (`last_completed_at >= last_triggered_at`) is already satisfied at fire time by
  design (`:337-342`). A session tonight built exactly that double-dispatch theory
  and reading `stampCompleted` killed it — the damage here is to the reader, not the
  scheduler.
- **source:** bugs_open/116 pickup, 2026-08-03 — the row's freshness was the first
  thing that made a deliberately-stopped loop look live
- **added:** 2026-08-03, bugfix_116_link_check_coverage lane

## An action that RETURNS findings and AWAITS a response loses the findings — the reply overwrites the step's record, and the green result reads as "everything was reported"

- **footprint:** `platform/orchestration/coordinator.go` (`applyResponseToState`,
  `storeActionResult`), `await_response`, `output_field`, any action that both
  computes findings and dispatches to an adapter (`retract_page_deployment` was
  the case that found it)
- **fires when:** you write (or read the record of) an action that returns a rich
  result — refusals, audits, counts — with `await_response: true`. The result IS
  stored (`storeActionResult` writes it under the step name and `output_field`),
  your immediate read-back confirms it, the step completes green. Then the
  adapter's reply lands and `applyResponseToState`'s default branch REPLACES both
  keys wholesale: `state.CollectedData[stepName] = normalisedData`, same for
  `output_field`. Everything the action computed is gone from the durable record;
  it survives only in pod logs, which do not survive a rollout. No symptom: the
  status is `complete`, the reply looks like the step's output, and a refusal or
  warning the action "returned" was never recorded anywhere a reader will look.
- **the check:** for any action with `await_response: true`, ask where its
  findings live AFTER the await, not before it. If the answer is "in its return
  value", they live nowhere. ~~Persist them before dispatching: a top-level SIBLING
  collected_data key survives (the response handler writes ONLY the step-name and
  `output_field` keys)~~ — **CORRECTED 2026-08-04, by the first live run of the fix
  that trusted it: the sibling key does NOT survive an awaited step either.**
  `persistAwaitingStateWithRetry` (coordinator.go:2052) parks the step by loading
  FRESH state from the DB and copying onto it only the awaited-request entries and
  status — every CollectedData mutation from the step's execution is discarded at
  park time (run `e23b7257…`: binary carries the key-writing code, strings-proven;
  persisted record has no key). Sibling keys (`image_result`, `final_html`) work
  ONLY on the non-awaited path, whose completion persist saves the live map. **The
  only check that holds for an awaited action is a DIRECT DB WRITE before
  dispatch** — `agent_error_log` via the package's INSERT shape (proven durable:
  `recordRetractionConditions`, `retract_page_deployment_action.go`) — plus
  RFC_012 addendum 2 for the class. And when you fix a detector or a sink, RE-RUN
  it on the motivating case: the in-memory tests passed while the DB round-trip
  lost everything.
- **source:** bugs_open/098 debt 5, 2026-08-03 — the one real retraction's record
  held only the adapter's `{paths, success, …}`; candidates, refusals and the
  whole graph audit were discarded by the await, contradicting the action's own
  "refusals are RETURNED, not swallowed" comment (071/083/091's
  detected-then-discarded class)
- **added:** 2026-08-03, bugfix_098_unpublish_primitive lane

## A `strings | grep` pod probe returns 0 for a marker the binary CONTAINS if the marker spans a non-ASCII byte

- **footprint:** `scripts/pick-pod-marker.py`, `strings /app/`, pod-grep deploy
  verification (the "prove a deploy at the artefact" practice), any marker chosen
  from Go source containing `…`, `—`, `£`, `⚠` or any other multibyte rune
- **fires when:** you verify a deploy by grepping `strings <binary>` for a source
  string literal that contains a non-ASCII character. `strings` emits printable
  ASCII runs and SPLITS its output at every multibyte byte sequence, so a marker
  crossing one is cut across two output lines and `grep -cF` returns 0 — against a
  binary whose bytes verifiably contain the full string. The failure is
  indistinguishable from "the fix never shipped", and a negative control does NOT
  catch it: the control is 0 for the right reason while your positive is 0 for the
  wrong one. Found live 2026-08-03: pick-pod-marker's own first suggested marker
  spanned a U+2026 and probed 0/0 on every replica of a binary that contained it.
- **the check:** markers must be printable ASCII end to end — `python3 -c
  "print(open('m').read().isascii())"` or just run `scripts/pick-pod-marker.py
  <commit>`, which now enforces ASCII-only nomination and prints an every-replica
  probe with a negative control. If you must probe for a non-ASCII string, grep
  the binary directly with `grep -acF` (byte match, no `strings` in the pipe) —
  and prove the probe on a marker you KNOW is compiled first.
- **source:** bugfix_140 plan item 5 build, 2026-08-03 — caught by running the
  tool's own suggested command against live pods before shipping it
- **added:** 2026-08-03, bugfix_140_contact_info_fabrication lane

## `site_work_items.page_id` is NULL on most rows even when the page exists — a page_id join reports "page missing" while the page is there

- **footprint:** `site_work_items.page_id`, `site_work_items` triage joins,
  `needs_page` / `needs_content_page` / `page_rerender` item populations
- **fires when:** you triage work items by `LEFT JOIN pages p ON p.id =
  w.page_id` and read a NULL as "the target page does not exist". Emitters
  mint items from (site_id, page_name) and mostly never set `page_id` —
  measured 2026-08-03: 27 of 28 parked `needs_page` rows carried NULL
  `page_id` while EVERY one of their target pages existed by name. A
  page-existence conclusion drawn from that join inverts the truth for the
  whole population, with no error anywhere.
- **the check:** join by name, taking the name from the spec: `JOIN pages p
  ON p.site_id = w.site_id AND p.name = COALESCE(w.spec->>'page_name',
  split_part(w.item_key,':',2))` — and prefer `spec->>'page_name'`: item_key
  prefixes differ per producer (`needs_page:` vs `page_rerender:`, the
  WII-004 drift), so key-parsing is a second trap inside the first.
- **source:** bugs_open/187 triage, 2026-08-03 — the lane's first triage query
  said every target page was missing; the by-name re-join said every one
  existed. WII-010 carries the same warning for the resolver's consumers.
- **added:** 2026-08-04, bugfix_187_sectionless_needs_page lane

### Hand-placing an object in `portfolio-sites/` works perfectly and opts the site out of every control the pipeline applies

- **footprint:** `b2 file upload … portfolio-sites` · `portfolio-sites/<host>/index.html` ·
  `portfolio-sites-router` · any `.html` you are about to write for a domain we own ·
  `082_submit_domain_unified.sh` · `evidence_base` · `imagery_style_guide`
- **fires when:** you need a page up quickly. Hand-writing the HTML and uploading it
  to the bucket **works on the first try** — the Worker serves it, the bytes match,
  the page renders, and every check you think to run passes. There is no error, no
  warning, and no trace anywhere that the site skipped the build pipeline
- **the tell:** none at the artefact, which is the whole problem. The tells are all
  *absences* in the database, and you only see them if you go looking:
  `SELECT * FROM sites WHERE domain='<host>'` returns **no row**, and therefore no
  `site_specs`, no `evidence_base`, no `pages`. **A missing `evidence_base` does not
  fail loudly — `loadEvidenceBase` returns nil and every claims lane silently
  no-ops** (`validate_page_content.go:727-746`), so the claims layer is not lenient,
  it is *absent*. Same for the hallucinated-email check, which **fails open** with no
  site email (`bugs_open/063`). A hand-built page is therefore the one kind of page
  on which a clean claims report means nothing at all
- **the check:** before writing any page content, ask **"does the framework already
  produce this?"** — for a domain we own the answer is yes. Then verify a page's
  provenance from the DB, never from the artefact:
  ```sql
  SELECT s.domain, count(p.id) AS pages,
         bool_or(ss.aspect='evidence_base') AS has_evidence_base
    FROM sites s LEFT JOIN pages p ON p.site_id=s.id
    LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.is_current
   WHERE s.domain='<host>' GROUP BY s.domain;
  ```
  No row, or `has_evidence_base=false`, means whatever is being served was not built
  by the pipeline regardless of how good it looks. **Correct path:** seed
  (`SEED_*.sql`, `oufe`'s is the worked example) then
  `082_submit_domain_unified.sh <domain> --email … --mission-file …`
- **source:** 2026-08-04. The webdesign.uk shopfront was hand-written and uploaded
  on 08-03, verified thoroughly, and reported as done — on the lane whose product is
  framework-built sites. Caught by the owner asking, not by any check.
  **OWNER RULING, now in `CLAUDE.md` Platform conventions: every site goes through
  the framework; the framework being slower is not a reason.**
  `WRONG_CALLS.md` 2026-08-04
- **added:** 2026-08-04, webdesign.uk build-service lane

## A loop substep's `config.continue_on_error` was accepted, audited as legal, and then silently overwritten — the knob you just set does nothing

- **footprint:** `platform/orchestration/loop_expansion_handler.go`
  (`handleLoopExpansion`, `resolveSubstepContinueOnError`) ·
  `platform/orchestration/loop_error_handler.go` (`shouldContinueLoopOnError`) ·
  `platform/orchestration/actions/loop_actions.go:66` ·
  `agent_definitions.default_config->workflow->steps->*->config->{substeps,sub_workflow.steps}->*->config->continue_on_error`
- **fires when:** you want ONE substep inside a loop to tolerate failure (or one
  substep inside a tolerant loop to stay strict) and you do the obvious thing —
  write `continue_on_error` into that substep's own `config`. **Before chassis
  v1.0.125x this had no effect whatsoever.** `handleLoopExpansion` deep-clones the
  substep config (so your value is genuinely there), then three lines later stamps
  the **loop's** value over it for every injected iteration step. `continue_on_error`
  is in `datahelpers.frameworkStepConfigKeys`, so `config-key-audit` reports it as
  legitimate framework vocabulary and never flags it as unknown
- **the tell:** none — that is what makes it a landmine rather than a bug you
  notice. No error, no warning, no audit finding, and the config reads correctly to
  every human and every checker. The only way to see it was to read the expander.
  **Any substep config predating the fix that declares this key has been inert since
  it was written**, and whoever wrote it probably believed otherwise — so if you find
  one, do not assume the loop has been behaving that way
- **the check:** confirm the resolution exists in the binary you are actually running
  before trusting a per-substep declaration, on **every** replica:
  ```bash
  POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'resolveSubstepContinueOnError'"   # 1 = per-substep honoured; 0 = your key is being overwritten
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'continue_on_error is true for this loop iteration step'"  # positive control, >=1 either way — proves the probe reads the binary
  ```
  and find any existing declarations fleet-wide (**0 rows as of 2026-08-04**, so a
  row here is new adoption, or an inert one somebody wrote in hope):
  ```sql
  SELECT a.type, s.key AS loop_step, b.key AS substep,
         COALESCE(s.value->'config'->>'continue_on_error','(unset)') AS loop_value,
         b.value->'config'->>'continue_on_error' AS substep_value
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s,
       LATERAL jsonb_each(COALESCE(s.value->'config'->'substeps', s.value->'config'->'sub_workflow'->'steps')) b
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.value->>'action'='loop' AND b.value->'config' ? 'continue_on_error';
  ```
  **A loop's body lives under `sub_workflow.steps` for most loops and `substeps` for a
  few — COALESCE both or you will get a confident 0 from the wrong key.**
- **the SECOND trap, for anyone editing the resolution:** `if v, ok :=
  cfg["continue_on_error"].(bool); ok && v` is **wrong**. Folding the type assertion
  into the truth test reads a declared **`false`** as *no declaration*, which silently
  destroys the strict-substep-inside-a-tolerant-loop direction — the one that stops a
  fan-out loop swallowing a failure that should have been loud. **Presence and truth
  must be tested separately.** A test suite that only covers the `true` direction
  passes against that bug; the guard here is mutation-proven against both
- **the THIRD trap:** `loop_metadata["continue_on_error"]` (same file, ~line 85) is
  **loop-level and must stay so** — `skipToNextLoopIteration` deliberately does not
  read it, because that shared key is overwritten when a second loop expands in the
  same orchestration. The per-substep value belongs on the **step**, which is where
  `shouldContinueLoopOnError` looks
- **source:** `bugs_open/173` (the missing degree of freedom, filed at the direction
  of the council's `architecture` and `constitution` seats after four seats rejected
  the workaround it forced in `bugs_closed/165`); fixed 2026-08-04, WFA-008.
  The clobber was found by reading the expander while verifying the bug was still
  valid — it is not stated in `173` itself
- **added:** 2026-08-04, bugfix_173_substep_error_tolerance lane

### A step whose `output_field` names a key an earlier step wrote REPLACES it — so a "pass-through plus bookkeeping" return shape silently demotes the real value one level

- **footprint:** `output_field` in any `agent_definitions` workflow step ·
  `platform/orchestration/coordinator.go` (`storeActionResult`, ~line 1858) ·
  `platform/orchestration/actions/load_current_section_content_action.go` ·
  `platform/orchestration/actions/plan_sections_action.go` (`section_plan`) ·
  any action returning `{"<the_thing>": …, "applied": …, "reason": …}`
- **fires when:** you add a step that *refines* a value an earlier step produced, and
  reuse that value's key as your `output_field` so downstream `input_mapping`s need no
  change. This is a **documented, sensible-looking pattern** — seed `299`'s header
  recommends it in those words — and it is safe **only if your return value is the
  refined value itself.** Return a wrapper around it plus some bookkeeping and you have
  replaced the key, not annotated it: `storeActionResult` stores an action's return
  value **wholesale** under `output_field`
  (`state.CollectedData[step.OutputField] = result`, no merge, no unwrap).
- **the tell — and why there isn't one:** **the wrong result looks exactly like the
  right one.** All the data is still present, one level down, at
  `<key>.<key>.<field>`. Nothing errors at the producing step; it reports success.
  The failure surfaces **two steps later**, in whichever consumer first walks the old
  path, under an error naming *its* missing key — so every reader is sent to the wrong
  file. In `bugs_open/192` that error (`key 'sections_ready' not found`, raised in
  `loop_actions.go`) was the whole visible symptom while the fault was an
  `output_field` declaration in a seed, and it broke **every page build in the fleet**.
- **the check, and it is one line:** never verify such a value by reading the path you
  expect — **enumerate the keys**:
  `SELECT (SELECT string_agg(k, ',' ORDER BY k) FROM jsonb_object_keys(collected_data->'<key>') k) FROM orchestration_states WHERE …;`
  A `SELECT collected_data->'x'->'y'` returns a quiet NULL for "absent" and for "you
  are one level too high" alike, so it can only confirm what you already believed.
  Compare a failing row against a healthy one: two different key sets is the answer.
  Before writing the step: if your action declares `output_field: X` and X already
  exists, its return value must **be** an X. Bookkeeping goes to the log, or to a
  namespaced key **inside** X — never around it.
- **the second trap, for whoever fixes one of these:** a compatibility shim is only
  safe where the reader has an **ordered fallback**. `select_sections`' fallback list
  tolerates both shapes and retires itself because the flat path sits ahead of the
  wrapper path. `input_mapping` has **no** such ordering — repointing one at the
  wrapper path fixes it today and silently re-breaks it on the roll, with no error.
  Leave that one broken and say so.
- **source:** `bugs_open/192` (filed 2026-08-04 by the `154` lane as undiagnosed;
  diagnosed and fixed the same day). Introduced by `bugs_open/178`'s fix
  (`08d0515f3`) + seed `299`; the action's own header promised the plan came back
  "byte-for-byte unchanged" and its unit test asserted the **wrapper**, so the test
  passed on the code that caused the outage. WFA-009.
- **added:** 2026-08-04, bugfix_192_select_sections_wrapper lane

### A "FIXED AND LIVE on v1.0.NNNN" close-out EXPIRES — the fleet rolls past your image within hours, and the sentence still reads as current
- **footprint:** `bugs_closed/` (any close-out naming an image tag), `makefile` (`IMAGE_TAG`), `make build-agent-chassis`, `kubectl get pods -l app=agent-chassis`
- **fires when:** you rely on a closed bug's "LIVE on v1.0.NNNN" line — reading it as a reader, quoting it in a handoff, or (worst) as the AUTHOR hours after you wrote it. This tree rolls several images a day from many lanes: 1245→1248 inside eleven hours on 2026-08-04, four replicasets inside 90 seconds at one point
- **the tell:** there isn't one. The sentence is correctly formed, was true when written, cites a real tag and a real pod grep — and says nothing about the image actually running now. `bugs_open/153`'s landmine is the INVERSE case (the image may PREDATE your commit); this is the one where your proof predates the image. Both produce a confident, dated, wrong "it is live"
- **the check:** re-grep the CURRENT pods, not the tag you proved on: `kubectl get pods -n ai-persona-system -l app=agent-chassis -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image --no-headers`, then `kubectl exec … -- sh -c 'strings /app/agent-chassis | grep -c "<a string YOUR change added>"'` on **every** replica, with a positive and a negative control. Because `make build-*` builds from committed HEAD, a later lane's build normally carries your fix for free — "normally" is the trap, and one exec settles it. Cite the claim as "live on `<tag>` as at `<date>`", never bare "live"
- **source:** `bugfix_163_symbol_lookup` lane, 2026-08-04 — proved on v1.0.1245, re-proved on v1.0.1248 rather than inheriting the claim
- **added:** 2026-08-04, bugfix_163_symbol_lookup lane

### `make build-*` fails with a LINKER error naming the Go toolchain, and the cause is a full 16G `/tmp` tmpfs — while `df /` shows hundreds of GB free

- **footprint:** `make build-<service>` · `go build` · `/tmp` · `/tmp/claude-1000` ·
  `TMPDIR` · `no space left on device` · `mapping output file failed`
- **fires when:** you build any service image, or run `go build ./...`, on the dev box.
  Compilation succeeds; the **link** step fails. The message points at the toolchain and
  looks like a Go or module problem:
  `.../pkg/tool/linux_amd64/link: mapping output file failed: no space left on device`
- **the tell, and why it misleads:** `df -h /` shows the root disk healthy (185G free,
  80% used, measured 2026-08-04). **`/tmp` is a separate 16G tmpfs**, and Go links
  through it. Check the right filesystem: `df -h /tmp`. On 2026-08-04 it was at **100%**
  with 384K free, so *every* concurrent session's builds were failing and none of them
  could see why from `df` alone.
- **what fills it:** `/tmp/claude-1000` — per-session Claude scratchpads — held **14G
  across 92 session directories**, the largest single one 3.5G. They are not reaped when
  a session ends, so on a box running dozens of concurrent sessions this fills steadily
  and silently. A failed `go build` also leaves its `/tmp/go-build*` behind (Go removes
  these only on success), so a full tmpfs is self-perpetuating: ~760MB of orphans were
  sitting there from earlier failed builds.
- **the check, before you conclude the build is broken:**
  `df -h /tmp` and `du -sh /tmp/claude-1000/*/ | sort -rh | head`. Then
  `pgrep -a "go |compile|link"` — with **no** live go process, any `/tmp/go-build*` is
  orphaned scratch and safe to remove; with one, it is not.
- **the trap inside the fix:** **do not sweep `/tmp/claude-1000/*` blind.** Many of those
  directories belong to sessions that are **live right now** (cross-check the ids against
  recently-modified `~/.claude/projects/*/…/*.jsonl`). Deleting an active session's
  scratchpad destroys work in flight that no git hook is protecting, because scratchpad
  files are deliberately outside the repo. Clear your **own** session's orphans, and
  escalate the rest rather than guessing. Alternatively set `TMPDIR` to a path on the
  root disk for the build only — it needs no cleanup decision at all.
- **source:** hit directly 2026-08-04 by the `bugfix_192_select_sections_wrapper` lane
  while running `go build ./...` as a final check; `./platform/...` had compiled clean
  moments earlier because that package set never reaches a link step.
- **added:** 2026-08-04, bugfix_192_select_sections_wrapper lane

---

### The concept index's own row count cannot detect a MISSING row — and the whole series of ~20 re-measurements was blind the same way

- **footprint:** `docs/agent_docs/docs026_concept_register/register/000_concept_index.md` ·
  `grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|' 000_concept_index.md` · adding any concept-register entry
- **the trap:** the index header asks each thread to re-measure the headline after adding a
  row, and every thread has done exactly that: count the rows, compare to the previous row
  count, confirm "1,720 → 1,721, exactly my row". That check answers *did my own row land,
  and did another thread's arrive alongside it* — and it is **structurally blind to a row
  nobody ever wrote**, which is the failure it looks like it is guarding. Measured
  2026-08-04: the headline said **1,722** while the category files held **1,756** entries.
  **34 concepts had a register entry and no index row**, including all of `CLM-001`…`012`
  — the first half of the claims-verification layer — plus `IMG-067`, `LNK-029`, `DBI-025`,
  `PLAN-043`…`046`, `PUB-002`…`004`, `WII-009`. The index is what a session or a council
  seat consults to learn whether a mechanism exists, so a missing row is invisible in
  precisely the lookup the file exists for, and it reads as "this does not exist".
- **why it drifts one way only:** adding a concept is two edits in two files, and only the
  first — the entry itself — is load-bearing for the author. The index row is the half that
  gets skipped. The reverse list (a row pointing at no entry) has always been empty.
- **the check, and it is the pair, not either half:**
  ```
  cd docs/agent_docs/docs026_concept_register/register
  cat *.md | grep -oE '^### [A-Z]{2,4}-[0-9]{3}' | sed 's/^### //' | sort -u > /tmp/h.txt
  grep -oE '^\| [A-Z]{2,4}-[0-9]{3} ' 000_concept_index.md | tr -d '| ' | sort -u > /tmp/r.txt
  comm -23 /tmp/h.txt /tmp/r.txt   # entries with NO index row — must be empty
  comm -13 /tmp/h.txt /tmp/r.txt   # rows with NO entry — must be empty
  ```
  Backfilled 2026-08-04 and both lists are empty at 1,756 ids each way. **Run the pair, not
  the row count**, whenever you touch the index — the row count agrees with itself no
  matter what is missing.
- **source:** found 2026-08-04 by the "concept register" session while bringing the register
  up to date; commit `8f998e86b`. Same session, same shape as the ratchet defect fixed in
  `102_CHECK_register_coverage.py`: a check whose result could not have come out otherwise.
- **added:** 2026-08-04, concept-register lane

---

### A doc citing `SOMETHING(12).md` now finds nothing on disk — 1,339 superseded version-family members were deleted on 2026-08-04, and an absent citation is not a fabricated one

- **footprint:** any `sources:` / `see` citation naming a `NAME(N).md` path ·
  `docs/agent_docs/docs024_key_docs_latest/travelling_docs/` ·
  `docs/social001_vonc_tiktok_social/` · `docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/` ·
  `docs/agent_docs/docs024_key_docs_latest/idea.uk/` · `docs/_archive/`
- **the trap:** the estate used to keep every save of a living doc as its own file —
  `RUNBOOK_travelling_docs.md`, `(1)`, `(2)` … `(39)`; up to **57 members** in one family.
  On 2026-08-04, at the owner's instruction, the superseded members were deleted (commit
  `e1fe8765e`: 441 families, 1,973 members, **1,339 deleted**, newest of each kept). Older
  documents — including **43 concept-register `sources:` lines** — still cite the deleted
  members by path. A session following one finds nothing, and the honest-looking conclusion
  is that the citation was invented or the evidence never existed. **It existed; it is in
  git.** This is the reverse of the usual danger: here the absence is the artefact, not the
  claim.
- **the check:** before doubting a citation to a `(N)` path, ask git, which still has all of it:
  ```
  git log --oneline --diff-filter=D -- '<path>'      # the commit that removed it
  git show <sha>^:'<path>' | head -50                 # its content at deletion
  git log --all --oneline -- '<path>' | tail -3       # its whole history
  ```
  Quote the path single-quoted — the parentheses are shell metacharacters, and an unquoted
  `X(3).md` fails with a syntax error that looks nothing like "file not found".
- **what is still on disk, so you know what a live citation looks like:** the newest member
  of every family (by mtime, which is **not always the highest N**), plus the unnumbered
  base name in every family — deliberately kept because code, bundle scripts and prose
  reference it, and in some families it is the *newest* member (`docs024/005_tool_pipeline.md`
  is 2026-07-26; its `(1)` is 2026-06-22).
- **added:** 2026-08-04, concept-register lane

### A `LIKE '%plan_sections%'` census returns FALSE POSITIVES, because `_` is a wildcard — and the step you are hunting is usually nested where `jsonb_each` cannot see it

- **footprint:** `agent_definitions.default_config` · any `psql` census of workflow steps ·
  `LIKE '%<name>_<name>%'` over jsonb-as-text · `jsonb_each(default_config->'workflow'->'steps')`
- **the trap:** two independent ways to get a confident wrong answer about "which agents run
  step X", both of which look exactly like a correct one.
  **(1)** In SQL `LIKE`, `_` is a **single-character wildcard**, not a literal. So
  `default_config::text LIKE '%plan_sections%'` also matches the substring `plan.sections`
  inside `section_plan.sections_ready` — and on 2026-08-04 `page-content-writer`, which had
  **no planning step whatsoever**, came back as a positive. The pattern reads as a literal to
  every human who writes it.
  **(2)** Most workflow steps that matter live inside a `loop`'s `sub_workflow`, not at the
  top level. `jsonb_each(default_config->'workflow'->'steps')` enumerates only the outer map,
  so four of the six `save_page_sections` call sites were invisible to it (`bugs_open/194`) —
  a clean empty-ish result with no error and no hint that half the population was never read.
- **the check:** read the step **keys**, and descend:
  ```sql
  -- every step at any depth, with its owning agent
  SELECT ad.type, j.value->>'action' AS action
  FROM agent_definitions ad,
       LATERAL jsonb_path_query(ad.default_config, '$.**.steps.*') j(value)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND j.value->>'action' = '<the action>';
  ```
  If you must use `LIKE` on the serialised text, escape the underscore —
  `LIKE '%plan\_sections%'` (or add `ESCAPE '\'`) — and treat the result as a superset to be
  confirmed against the keys, never as the answer. **A negative under the unescaped (more
  permissive) pattern is still sound; a positive is not.**
- **added:** 2026-08-04, `bugs_closed/087` lane

### A migration verify block comparing a jsonb path with `<>` sits GREEN for ever when the key does not exist — `NULL <> 'x'` is NULL, not TRUE

- **footprint:** `docs/agent_docs/sql_for_agents/*.sql` verify blocks · `DO $$ … RAISE EXCEPTION` ·
  any `#>>` / `->>` compared with `<>` or `=` · `@>` containment tests in PL/pgSQL
- **the trap:** the estate's house style is a `DO` block that `RAISE EXCEPTION`s if the config
  is not the shape you intended — and it is the right style, because a verify block made of
  `SELECT`s cannot stop a `COMMIT` (`ON_ERROR_STOP` ignores a non-empty result). But
  `#>>` on a **missing** path yields NULL, and `NULL <> 'expected'` evaluates to **NULL**,
  which is not TRUE, so the `RAISE` never fires. The check that exists to catch a wrong key
  name is disabled by exactly the wrong key name. Seed `309`'s first draft asserted
  `process_sections_loop.config.items_field` — the real key is `iterate_over` (the
  neighbouring `page-rebuild` loop uses `items_field`, which is where the wrong name came
  from) — and it passed, silently, against NULL.
- **the check:** use **`IS DISTINCT FROM`** for every scalar comparison in a verify block, and
  wrap every containment test in `COALESCE(<expr> @> '…'::jsonb, false)`. Then **induce it**:
  extract the block and run it alone against the *unmodified* row, and require an exception —
  ```bash
  awk '/^DO \$\$/,/^END \$\$;/' docs/agent_docs/sql_for_agents/<seed>.sql \
    | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
  # must ERROR. If it says COMMIT-clean, the block is inert and you have no verification.
  ```
  A verify block you have not watched fail is a claim, not a check. Related: the RFC_006
  lesson already in CLAUDE.md (a verify block of `SELECT`s cannot stop the `COMMIT`) — this is
  the next failure along, where the block *is* a `DO` and still cannot fire.
- **added:** 2026-08-04, `bugs_closed/087` lane

### `agent_definitions.updated_at` has NO trigger — a config row you just rewrote still reads as another session's write, and "nobody has touched it" is the conclusion it invites

- **footprint:** `agent_definitions.updated_at` · `docs/agent_docs/sql_for_agents/*.sql`
  seeds that `jsonb_set` a live agent · any "has this row changed under me?" check on a
  contended tree
- **the trap:** the house pattern for a config change is a seed of targeted `jsonb_set`
  UPDATEs, and the house pattern for *safety* on a tree ~30 sessions share is to read
  `updated_at` first to see whether anyone else has been here. **There is no trigger
  maintaining that column** — verified 2026-08-04, `pg_trigger` has no non-internal row
  for the table — so it is current only when a seed sets it by hand. Seeds `246` and
  `308` do; seeds `309` and `310` as first written did not. Result: after six applied
  UPDATEs the `page-content-writer` row still read `09:01:35Z`, the *previous* lane's
  write, while carrying changes from a lane an hour later. The failure is silent and
  points the wrong way — it manufactures **absence** of activity, which is the reading
  nobody questions, and it mislabels your own change as someone else's.
- **the check:** treat `updated_at` as one-way — recent means someone was here, stale
  means **nothing**. For "did this row change under me?", diff the content across your
  read-to-write window, which no missing trigger can fool:
  ```sql
  SELECT md5(default_config::text) FROM agent_definitions
  WHERE type='<agent>' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
  And put `updated_at = NOW()` in your own seed's `SET` (or as a final stamping UPDATE
  inside the same transaction) so the next session's check is not lied to by yours.
  Same family as the fleet landmine on `scheduled_tasks.last_completed_at`: **a column
  that looks like proof of a write is only as good as whoever remembered to write it.**
- **added:** 2026-08-04, `bugs_closed/087` lane

### Before adopting the retraction seam, count the CLOSERS of your `item_type` — the mandatory producer check cannot see them, and a second closer is the same hazard from the other end
- **footprint:** `CheckResult.Resolved`, `resolveWorkItems`, `platform/orchestration/actions/revalidate_review_queue_action.go` (`reviewRevalidators` map, ~line 161), `site_work_items.result`, any `discovery_checks/check_*.go` adopting RFC_010
- **fires when:** you follow `HANDOFF_2026-08-03b` §3 — *"count the producers of your item_type before you write a line of code"* — do it thoroughly, two independent ways (grep the Go producers, corroborate with `created_by` over every row), get a clean single-producer answer, and adopt. **The check is mandatory, it is correct, and it has a producer-shaped hole.** Nothing in it asks what already CLOSES these items
- **the tell:** **none while you are writing the code, and the live evidence reads as history.** `revalidate_review_queue` closes items by writing `status='complete'` with a `result.revalidation` block; it is not a discovery check, it registers no `Resolved`, and it names your item type in a map entry one line long. A `SELECT status, count(*)` shows you some rows already `complete` — which looks like old hand-closed work, not like a mechanism running weekly. The two closers do not even corrupt each other (`resolveWorkItems` skips rows already in `workItemClosedStatuses`), so nothing ever fails; you simply build a duplicate and, if yours happens to run first, you credit yourself with its work
- **the check:** ask what closed the ones that are ALREADY closed, before claiming nothing can close them: ```sql SELECT status, count(*), left(result::text,120) FROM site_work_items WHERE item_type='<your type>' AND status IN ('complete','verified') GROUP BY 1,3 ORDER BY 2 DESC; ``` A `result.revalidation` block means `revalidate_review_queue` owns this type — read its `reviewRevalidators` map and its selection predicate (`status='needs_human_review'`) before writing a second closer. Then grep the code side too: `grep -rn --include=*.go "\"<your type>\"" platform/ internal/ | grep -v _test` **without** filtering to `ItemType:`, which is what confines the producer check to producers
- **worked case:** `required_fields_missing`, 2026-08-04. Adopted the seam on the stated ground that the type was flag-only so *nothing* could close it. `revalidate_review_queue` had been closing it since 2026-07-27 with the same predicate, covering 100% of the population (every one of these items is born `needs_human_review`, which is exactly what it selects on), plus a refusal the new code lacked. Caught only by re-running the sizing after the roll and finding the 6 target items had closed two hours before the new pods started. 14 council seats accepted the absence claim; none could check it either. Full account: `WRONG_CALLS.md` 2026-08-04
- **added:** 2026-08-04, bugfix_168 lane

## `jsonb::text LIKE '%"key":"value"%'` matches NOTHING — Postgres renders jsonb with a SPACE after the colon, so the census returns 0 whatever the data holds

- **footprint:** any `<jsonb_column>::text LIKE '%"…":"…%'` census (`orchestration_states.collected_data`, `site_work_items.spec`, `agent_definitions.default_config`, `sites.settings`, `pages.page_spec`), and the "is this config key ever actually SET?" question generally
- **fires when:** you census a key's real usage and want to exclude the bare-word false positives (docs, descriptions, council submissions that quote the key). The natural tightening is to match the JSON shape — `LIKE '%"deploy_path":"%'` — and it is **structurally incapable of matching a jsonb column**: `jsonb::text` normalises to `{"key": "value"}`, **with a space**. `json` (not `jsonb`) preserves the input's spacing, so the same pattern may work on a `json` column and silently fail on the `jsonb` one beside it.
- **the tell:** there is none. You get `0`, which is the answer you were hoping for, and it is indistinguishable from a true zero. It will agree with itself across every table you try, because they are all jsonb — **two blind checks agree with each other.**
- **the check:** **induce a non-zero before you trust a zero.** Write the value (a probe row, or point the query at a row you know carries it) and confirm the query finds it. Then use a spacing-tolerant regexp, and require a non-empty value:
  ```sql
  WHERE collected_data::text ~ '"deploy_path"\s*:\s*"[^"]+"'
  ```
  Or ask jsonb itself rather than its text rendering — `jsonb_path_exists(collected_data, '$.**.deploy_path')` for "anywhere", or `collected_data #>> '{input_data,deploy_path}'` when you know the path. **A key nested at unknown depth is exactly when people reach for `::text LIKE`, which is exactly when it fails.**
- **why it is a landmine:** the broken pattern is *recommended in writing* in this repo — `bugs_open/179`'s Evidence section says "⚠ Match the JSON shape, not the bare word … Use `LIKE '%\"deploy_path\":\"%'`" — because the bare word really does return a pile of false positives. So the advice is right about the problem and wrong about the remedy, and it has been copied into at least the `bugfix_168` handoff and one migration. Following the repo's own guidance is what produces the artefact.
- **source:** 2026-08-04, `bugs_open/179` finding A. The zero rode into a council submission (APPROVED), the IMG-067 register entry, migration 307 and four commit messages before it was caught — by accident, when a post-roll re-run of the same census still said `0` minutes after an induced probe had deliberately written a `deploy_path`. Full account in WRONG_CALLS.md, same date.
- **added:** 2026-08-04, bugfix_179_deploy_path_override lane

## `ImagePullBackOff` reports a Job as still RUNNING, never FAILED — so a CronJob copied from a sibling that pulls a PUBLIC image goes silent while looking slow

- **footprint:** `deployments/kustomize/services/*/base/cronjob.yaml`, `imagePullSecrets`,
  `docker-hub-creds`, any new CronJob modelled on `component-fallback-check` /
  `database-backup` / `agent-job-cleanup`
- **fires when:** you add a CronJob that runs an image from the private `aqls/` repo and
  copy its manifest from an existing job. The three obvious models all pull PUBLIC images
  (`postgres:16-alpine`, `bitnami/kubectl`) and therefore carry **no `imagePullSecrets`**,
  so the omission is invisible — the manifest looks complete and `kubectl kustomize`
  renders it happily. In the cluster the pod enters `ImagePullBackOff`, and the Job's
  status stays **`Running` with `0/1` completions until `activeDeadlineSeconds` expires**.
  It is never reported as Failed, no alert distinguishes it from a slow run, and any
  `doc_notes`/report row the job exists to write is simply never written — so a check
  whose whole purpose is "a silent check is indistinguishable from a dead one" dies
  exactly that way.
- **the check:** after deploying ANY new CronJob, trigger one run
  (`kubectl create job --from=cronjob/<name> <name>-manual-<ts>`) and confirm the **pod**
  reaches `Completed` — `kubectl get pods -l job-name=<job>`, not `get job`, because the
  Job object is the thing that lies here. Then verify the job's ARTEFACT (its row, its
  file), not its exit status. Deployments in this namespace pull with
  `imagePullSecrets: [{name: docker-hub-creds}]`; copy that whenever `image:` starts
  `docker.io/aqls/`.
- **source:** bugfix_140 item 1 carrier, 2026-08-04 — caught on the first manual run of
  `component-render-check`, which is the reason to do one
- **added:** 2026-08-04, bugfix_140_contact_info_fabrication lane

---

## A chrome link validated against `loadResolverPageSet` ships a 404 the nav beside it already refused
- **footprint:** `platform/orchestration/actions/render_site_components_action.go`,
  `platform/orchestration/actions/resolve_internal_links_action.go` (`loadResolverPageSet`),
  `platform/orchestration/actions/nav_tables.go` (`applyNavVisibility`,
  `loadFetchablePageSet`), `site_components.rendered_html`, any new chrome slot that
  emits an `<a href>`
- **fires when:** you add or edit anything in chrome that names a page — a CTA, a
  footer column, a new slot — and reach for the nearest page-set helper. There are two
  and they are one word apart. `loadResolverPageSet` is the page-CONTENT set: status
  floor only, **no deployment predicate at all**. `loadFetchablePageSet` (via
  `LoadChromeLinkPolicy`) is the chrome set. Picking the wrong one is invisible in
  review, invisible in the DB, and invisible on a mature site — it only bites on a site
  with pages planned and not yet built, which is **every adoption, at exactly one stage**.
  Chrome then ships on every page, so it is one 404 button per page, and the chrome
  render is **idempotence-gated**, so nothing re-renders it when the target finally
  deploys. Measured 2026-08-04: `mortgagecalculator.co.uk`'s header nav was filtered to
  its one deployed page while its CTA button, in the same component from the same run,
  pointed at a `build_status='planned'` page returning **HTTP 404**.
- **the check:** chrome link targets go through
  `LoadChromeLinkPolicy(ctx, db, siteID, logger)` + `.Allows(url)` (LNK-030) — never
  `loadResolverPageSet`, whose doc comment now says so, and whose caller allow-list is
  enforced by `chrome_link_policy_test.go`. ⚠ **And when you measure the damage, do not
  trust the obvious SQL.** The query in `bugs_open/191` over-reports twice: its
  `LEFT JOIN` on a regex-extracted href turns every header with **no** CTA into a
  NULL-join row that satisfies `p.deployed_at IS NULL` (4 of its 6 rows today have an
  empty `cta_href` — add `AND substring(...) IS NOT NULL`), and of the 2 real rows one
  (`lendzy.co.uk`) **serves HTTP 200**, because `deployed_at IS NULL` means "no recorded
  deploy", not "does not serve". **Curl every surviving href. The confirmed live 404 was
  1, not 6 and not 2.**
- **source:** `bugs_open/191`, fixed 2026-08-04; the component-eligibility sibling is
  `bugs_closed/118` / CLC-013, which fixed the same shape one layer up and explicitly did
  not touch link targets
- **added:** 2026-08-04, bugfix_191_chrome_link_policy lane

### A council run killed by a chassis roll looks exactly like a slow one for FOUR HOURS

- **footprint:** `097_TRIGGER_council_review_v1.sh`, `orchestration_states.current_step`
  where the row is a `council-gate` run, `diagnosis_artifacts` kind `council_report`,
  any wait-for-verdict poll
- **fires when:** you submit to the council gate and wait on the VERDICT ARTIFACT —
  the documented advice, and correct in general ("a missing orchestration row is
  almost always latency, not a dropped dispatch — do not retry on that evidence")
- **the tell:** none, and that is the whole problem. A roll kills the pod mid-seat;
  the orchestration row stays `EXECUTING_STEP` on whichever `review_*` step it had
  reached, with no error, until the reaper marks it FAILED after **>4h** with
  `reaper: stale EXECUTING_STEP for >4h; step=review_<seat>`. Until then it is
  indistinguishable from a seat that is simply thinking, so a patient thread waits
  four hours for a verdict that can never arrive. Measured 2026-08-03/04: **two of
  five runs in one lane** died this way (`review_prior_art`, `review_editquality`),
  on an estate that rolled `v1.0.1243 → 1247 → 1250 → 1251` inside a day.
- **the check:** poll the ROW, not the artifact, and treat a stalled step as the
  signal rather than waiting for the reaper —
  ```sql
  SELECT current_step, status, NOW()-updated_at AS since_update
  FROM orchestration_states
  WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
  ```
  A council seat is an LLM call of 2–5 minutes. `EXECUTING_STEP` with `since_update`
  past ~20 minutes on ONE step means the pod is gone; resubmit with
  `RESUBMIT_CORR=<corr>` (the trail accumulates, and round counting is
  orchestration-scoped so a resubmission is judged fresh). Do NOT read this as
  licence to retry on a MISSING row — that really is dispatch latency, ~29 min
  measured; the discriminator is a row that EXISTS and has stopped moving.
- **source:** `bugs_open/185`'s council trail, 2026-08-03/04. Extends the existing
  "a roll KILLS an in-flight council" entry with how to TELL and what the wait costs
- **added:** 2026-08-04, bugfix_175_page_role_upsert lane

---

### A daily check's `doc_notes` source is the SCRIPT's name, not the CronJob's — querying by the name you deployed returns 0 rows, which is indistinguishable from a check that has stopped running

- **footprint:** `doc_notes.source`, `component-fallback-check`, `component-render-check`,
  `check_placeholder_fallbacks`, `component_render_check`, any CronJob whose whole purpose
  is to write a row saying it looked
- **fires when:** confirming a scheduled check actually ran — the routine health question,
  with no symptom and no suspicion. You know the CronJob's name because you just read it
  out of `kubectl get cronjob`, so you query `WHERE source='<that name>'`.
- **the tell:** there is none. `SELECT ... WHERE source='component_fallback_check'` returns
  `(0 rows)` — the exact result a dead check produces. **The two checks in this family do
  not even agree with each other:** `component-render-check` writes under
  `component_render_check` (the CronJob's name, underscored), while
  `component-fallback-check` writes under **`check_placeholder_fallbacks`** (the SCRIPT's
  name, which is not the CronJob's name at all). A session that verifies one by the
  obvious rule and then applies the same rule to the other gets a false negative and a
  frightening one, because these rows exist precisely to distinguish "looked and found
  nothing" from "stopped running".
- **the check:** never guess the literal — read it out of the writer, or list what is
  actually there and match by eye:
  ```sql
  SELECT source, count(*), max(created_at) FROM doc_notes
   WHERE created_at > now() - interval '3 days' GROUP BY source ORDER BY max(created_at) DESC;
  ```
  ```bash
  grep -n "source" <the script or cmd/ dir that writes the row>   # the INSERT is the authority
  ```
  And before concluding a check is silent, ask the Job whether it ran at all —
  `kubectl -n ai-persona-system get pods -l app=<name>` — because a Completed pod plus no
  row is a genuine defect, while no pod at all is a different one entirely.
- **source:** hit on 2026-08-04 in the bugfix_140 lane while verifying the daily
  fallback check. Caught before it was asserted, by reading `write_doc_note()` — but the
  0-row result had already been read as "the daily check writes nothing", and the runbook
  was open at the page that quotes the correct literal. Same class as the standing
  "a grep proves absence only for the SPELLING it searches" rule; this is the instance
  where the wrong spelling is the name the cluster itself shows you.
- **added:** 2026-08-04, bugfix_140 lane

### `save_page_sections` now picks its own `sections_metadata` path when the step does not name one — an "absent" key is no longer inert
- **footprint:** `platform/orchestration/actions/save_page_sections_action.go`, `save_sections_metadata_source.go`, `page_components.content_data`, the config keys `sections_metadata_field` / `expects_no_sections_metadata` / `refuse_save_without_sections_metadata`
- **fires when:** you add a `save_page_sections` step to a workflow, or read an existing one, and reason about it from its config alone. Until 2026-08-04 a step with no `sections_metadata_field` took the HTML-parse path and wrote `content_data = NULL`; from the next chassis roll the same config resolves `page_content.response.sections_metadata` by DEFAULT. **The absence of the key no longer tells you which path a save takes** — and if your step happens to carry a `page_content`-shaped reply for a DIFFERENT page or purpose, the default will find it
- **the tell:** none in the config; the save reports success either way. Read the result map instead — every save now returns `sections_source` (`metadata`/`html_parse`), `metadata_field_origin` (`configured`/`default`/`declared_absent`) and `sections_with_content_data`. That triple is the only thing that distinguishes "structured content persisted" from "the page kept its HTML and lost its regeneration source", which used to be indistinguishable from the outside
- **the check:** before adding a step, decide which of the three states it is in and SAY so: name `sections_metadata_field` if the caller has structured content somewhere non-default; set `expects_no_sections_metadata: true` if it genuinely has none (a whole-page tool blob — `tool-recreation-handler` is the live example, and its NULL is correct); set `refuse_save_without_sections_metadata: true` if a missing one should refuse rather than save (**RENAMED 2026-08-05, code-review F7** — it was `require_sections_metadata`, which is a LIVE key of a DIFFERENT, warning-level meaning on `validate_page_content` steps; `page-build-handler` carries both steps, so one word meant two things in one definition. Verified with the nested walk below: 0 of the 6 save callers carried it, so the rename broke nothing). To find out what a live step actually does, do NOT trust a top-level `jsonb_each` — the step is nested in a loop `sub_workflow` in four of the six callers and that query finds only two: `SELECT ad.type, s.value->'config' FROM agent_definitions ad, LATERAL jsonb_path_query(ad.default_config,'$.**.steps') AS steps, LATERAL jsonb_each(steps) AS s(key,value) WHERE s.value->>'action'='save_page_sections' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;`
- **the sixth refusal, and it is OPT-IN:** the entry above records five refusing guards in this function. `refuse_save_without_sections_metadata` is a sixth, but unlike the other five it is reachable by nothing until a step names it (RFC_010, default OFF). A save that fails with `this step declares refuse_save_without_sections_metadata…` is a step that asked for it
- **source:** `bugs_open/194`, concept register `PBP-031`, seed `312`
- **added:** 2026-08-04, bugfix_194 lane

### A page that has not re-rendered is holding EVERY platform improvement since it last did — you cannot size the job by the change you authored
- **footprint:** `platform/orchestration/actions/rerender_single_page_action.go` (`assemblePage`, `injectCanonicalLink`, `injectPageJSONLD`, `injectComponentCSS`, `spliceMetaDescription`), `pages.deployed_at`, any `page_rerender` work item filed with **no** `reason`
- **fires when:** you decide whether re-rendering some stale pages is *worth it*, and price it by the change **you** made — "it's only a footer comment", "it's only a copy tweak", "it self-heals on next rebuild". The reasoning feels like a judgement about your own work, whose inputs you obviously know, so it does not trip the `[INFERRED]` habit at all. **It is a claim about a SET — everything queued behind that page's next render — and most of that set was authored by other lanes.** `assemblePage` alone injects canonical, JSON-LD, per-component CSS and the meta-description splice, each added by a different workstream on a different day
- **the tell:** there isn't one, and that is the point. The stale page renders fine, returns 200 and looks correct. The missing thing is in the `<head>` of a page nobody is looking at, and it stays missing for as long as nobody re-renders it
- **the check:** **re-render ONE page and diff it, before you decide.** ~2 minutes, and it is the only thing that enumerates the set:
  ```bash
  curl -s "https://<domain>/<path>" -o /tmp/p.pre     # then file an assemble-only page_rerender (spec with NO `reason`)
  curl -s "https://<domain>/<path>" -o /tmp/p.post && diff /tmp/p.pre /tmp/p.post
  ```
  ⚠ **Use TWO canaries, of different page types, and expect them to DISAGREE.** Stale pages are **not** a homogeneous population — each holds whatever shipped since *it* last rendered, so two pages last built a week apart hold different sets. A single canary produces a confident, internally consistent, wrong prediction, and the one that agrees with you is the likelier pick.
  For a baseline-free acceptance test afterwards, census the SERVED pages for what the platform currently emits (`<link rel="canonical">`, a non-empty `<meta name="description">`) rather than diffing — and **induce a non-zero first**, because a grep that matches nothing looks identical whether the site is clean or your probe is broken.
  ⚠ **THIRD cause of a zero, and it bit me an hour after writing this entry: the artefact you fetched is not a page.** Curl a URL inside its own deploy window and B2 hands you a 7-line `{"error":"B2 returned error"…"NoSuchKey"}` blob with HTTP 200. **Every** grep against it returns 0 — including the ones you *want* at 0 — so a corrupt fetch reads as a clean partial success. Two independent checks over that one blob agreed with each other and nearly bought a fleet-wide "PageInfo.Domain is unpopulated" finding. **Establish the artefact's KIND before grepping it** (`wc -c`, and `head -c 15 | grep DOCTYPE`), and do not fetch during a deploy — `complete` is the work item's status, not the CDN's; allow ~2 min.
- **source:** loancalculator.co.uk lane, 2026-08-04. The item was recorded on 08-03 as "an invisible comment; 26 forced deploys is disproportionate" and left. Canarying before the bulk showed **20 of 26 served pages had NO `<link rel="canonical">` and carried `<meta name="description" content="">`** — from `9c7a8e9e4` (2026-08-02), live and unnoticed for two days on a site whose product is being found. Full write-up in `WRONG_CALLS.md` (same date) and the lane's `NOTES`
- **added:** 2026-08-04, loancalculator_couk lane

### A pathspec commit can put a call to an UNDEFINED symbol into HEAD — your working tree builds green while HEAD does not compile at all

- **footprint:** `git commit <path>`, `platform/orchestration/actions/save_page_sections_action.go` and any other file several sessions edit at once, `make build-<service>`, `git ls-files`, `git diff --cached --stat`
- **fires when:** you do exactly what CLAUDE.md mandates — commit your task with an explicit pathspec — on a file that another session is also editing. It needs no symptom and gives no warning: `go build ./...` is green, `go test` is green, the commit succeeds, the yellow commit-scope block lists your file and nothing looks wrong.
- **the trap:** a pathspec commit takes the **working-tree** version of the named file, so it takes the other session's hunk in that file too (documented — that is the known same-file passenger). What is NOT documented is the second-order case: **their hunk can CALL a function whose defining file is `git add`ed but not yet committed.** You commit their call site; their definition stays in the index; HEAD now contains a call to an undefined symbol. Your tree still builds because the definition exists **on disk**. `make build-<service>` builds from **committed HEAD**, so every backend service build fails fleet-wide, for a reason that is in nobody's diff. Hit 2026-08-04 on `save_page_sections_action.go`: the `bugs_open/190` lane's call to `sanitizeSectionsContentData` sat in the shared tree with `content_data_envelope_guard.go` staged (`A` in `git status`) and uncommitted.
- **the tell — there isn't one locally, which is the whole entry.** Both halves of the usual reassurance are true and useless here: the build is green (it reads the tree) and the tests pass (same). The only place the breakage exists is HEAD, which nothing you ran looked at.
- **the check, before any pathspec commit on a shared file — read the HUNKS, not the count:**
  ```bash
  git diff <file> | grep -E '^@@|^\+' | grep -v '^\+\+\+'   # what is actually in there besides yours
  git diff --cached --stat                                  # what has SOMEONE ELSE staged
  git ls-files <file-defining-any-symbol-a-foreign-hunk-calls>   # empty output = HEAD will not compile
  ```
  and if you want certainty rather than inference, build the thing you are about to create:
  `git stash create`-free and side-effect-free — `git archive HEAD | tar -x -C <scratch>/headtree`, apply your intended commit there, `go build ./...`.
- **what to do when you find one:** **not** `git add -A` to sweep their definition in — that takes every other session's staged work with it, which is the damage the pathspec rule exists to prevent. Either wait for their commit (it is usually minutes; theirs landed 7 minutes later) and then commit yours, or commit your OTHER files first and hold only the shared one. Whichever you choose, **name the passenger in your commit message** — a hunk that is not yours, shipped under your subject, is invisible to review and to `git log` for ever otherwise.
- **the direction that is safe to get wrong:** committing your file WITHOUT their hunk is impossible with a pathspec, but committing LATER than them costs nothing (their commit makes your file-diff smaller). Waiting is always cheap; the broken-HEAD case is not.
- **source:** 2026-08-04, `bugs_open/156` lane (`docs024_key_docs_latest/bugfix_156_duplicate_sections/`, RUNBOOK R7). Caught before committing, by checking `git ls-files` on the callee — not by anything that failed.
- **added:** 2026-08-04, bugfix_156_duplicate_sections lane

### `page_component_history` archives the state being REPLACED — every count over it is a count of overwrites, never of writes
- **footprint:** `page_component_history` (all columns, especially `content_data` and `source`), `platform/orchestration/actions/save_page_sections_action.go` (the history INSERT immediately before the DELETE), the source literal `save_page_sections_overwrite`
- **fires when:** you use this table to answer "how often did the system WRITE X?" — the natural question when tracing a bad `content_data` shape to its producer, and the table looks purpose-built for it: it has a `source` column, timestamps, and one row per event. Nothing in the schema, the column names or the row contents says the payload is the OLD value.
- **the tell:** none, in either direction. The query returns a confident, internally consistent number, and every follow-up you reach for — group by source, by date, by site — stays consistent under the wrong reading. On 2026-08-04 the count was 65, all from one source, most recent the previous evening, and it read exactly like "the seam has written 65 of these, and it did so again yesterday" — a far more urgent bug than the real one. The correct reading is "65 rebuilds of pages that ALREADY carried one".
- **the check:** **read the INSERT, not more data.** It is `INSERT INTO page_component_history (...) SELECT pc.content_data ... FROM page_components pc WHERE pc.page_id = $1`, executed *before* the DELETE — so it snapshots the pre-existing row. The timestamps confirm it: the history row precedes its replacement `page_components` row by ~80ms. **No amount of further querying disambiguates this**, which is why it is a landmine rather than a §9 pattern: the disambiguator is four lines of SQL inside a Go action, and it takes less time to read than the queries you would run instead.
  - **Second trap in the same query:** `count(DISTINCT component_id)` returns **0**, and 0 here means **NULL**, not "no components". `page_component_history_component_id_fkey` is `ON DELETE SET NULL`, so any archived row whose component was later deleted has the id nulled — and these rows are, by definition, ones whose components got replaced. **Group by `page_id`.** (Same zero-that-means-NULL shape as the `distinct_content = 0` trap already in this file; second sighting in this table family.)
  - What the table *can* honestly tell you is the historical **blast radius**: 65 events → 25 distinct pages across 6 sites, where the live tables showed only 2. And with a date filter it separates a defect still being CREATED from one merely being carried forward — the single most useful cut, and the one nobody runs by default.
- **source:** `bugs_open/190` lane, 2026-08-04. Drafted "the save seam has written 65 envelopes, most recently yesterday" and caught it by reading the action. `WRONG_CALLS.md` same date; full working in `docs024/bugfix_190_content_data_envelope/RUNBOOK` §3.
- **added:** 2026-08-04, bugfix_190_content_data_envelope lane

### The `content_data` envelope guard REFUSES one live page on purpose, for ever — and the one-line "fix" for that noise destroys the page
- **footprint:** `platform/orchestration/actions/content_data_envelope_guard.go` (`normalizeContentDataEnvelope`, the `ProvenanceClean`/`ProvenanceRepaired` test), `agent_error_log.error_code = 'CONTENT_DATA_ENVELOPE'`, gaswholesalers.com `how-pricing-works`
- **fires when:** you are drawn to this guard by NOISE rather than by the bug — a rebuild that keeps failing, a recurring `error`-severity row, a work item that will not clear. The obvious remedy is to accept whatever the parser managed to recover, since it plainly *did* recover something. That single change is the destructive one.
- **the tell:** the refusal looks like over-caution, because the payload genuinely parses. `ParseLLMJSONWithProvenance` returns a real object for it. What the verdict is telling you is *how* it got there: provenance `prose_around` means bytes outside the recovered value were discarded — and for this row those discarded bytes are the entire page copy, sitting in a markdown tail after a small JSON fragment. Accepting it stores **131 characters** over a live page. The recovered fragment is not obviously short unless you measure it.
- **the check:** before loosening anything, run the payload through the parser yourself and read the PROVENANCE, not the success:
  ```sql
  SELECT id, left(content_data->>'result', 200) FROM page_components
  WHERE content_data->>'type' = 'text' AND jsonb_typeof(content_data->'result') = 'string';
  ```
  then compare the recovered object's total size against the raw string's. `TestLossyProvenanceIsRefusedNotDecoded` exists to fail on exactly this loosening, and its fixture asserts a *precondition* that the payload really is mechanically recoverable — so a test author cannot accidentally make it pass for the wrong reason. **The right response to the noise is to repair the page, not the guard**; the row already carries a `needs_human_review` item.
- **also on this file:** its header names `__truncated` as an example of the `__`-prefixed transport keys a decode drops, and `truncation_guard_test.go` scans package **sources** for that literal — so **the comments in this file are load-bearing**. It is exempted in `truncationMarkerExemptions` rather than registered, deliberately: registering `save_page_sections` in `truncationAwareActions` would widen what counts as a truncation-aware consumer for every workflow containing that action.
- **source:** `bugs_open/190`, concept register **PBP-032**, council correlation `09bc4b3d-6721-4479-85b8-b5b56bf9b5d7`
- **added:** 2026-08-04, bugfix_190_content_data_envelope lane

### An IP-allowlisted credential's health checks pass WITHOUT exercising the allowlist — while the office line's IPs rotate under you

- **footprint:** `~/.config/cloudflare/token`, `api.cloudflare.com` (`/user/tokens/verify`, error `9109`), `epp.nominet.org.uk:700`, `docs024_key_docs_latest/domains_cloudflare_rollout/`
- **fires when:** you sanity-check credentials for the domain-portfolio rollout (or any Cloudflare-token / Nominet-EPP work). Both standard checks are counterfeit: Cloudflare's `/user/tokens/verify` returns 200 for a token whose IP filter 403s every real endpoint, and Nominet serves its EPP greeting to ANY connecting IP (proven 2026-08-04: full greeting from an address that was never allowlisted). Each looks exactly like a working credential, so you debug your script instead of the network.
- **the tell:** real calls fail `{"code":9109,"message":"Cannot use the access token from location: <ip>"}` while verify stays green — and the 9109 message names the address you are ACTUALLY egressing from; read it, it is the diagnosis. On a dual-stack host the failures can be intermittent (address family chosen per connection), which reads as flakiness.
- **the check:** prove the Cloudflare token with `GET /zones?per_page=1`, never verify alone; prove Nominet at LOGIN, never at the greeting. The office line rotates BOTH its IPv4 and its whole IPv6 prefix between days (measured 08-02→08-04: `151.226.83.138`→`5.65.164.9`, `2a02:c7c:f61f:ac00::/64`→`2a02:c7e:3066:5400::/64`), so any filter pinned to it WILL break again; the stable addresses are the five k8s node IPs (`kubectl get nodes -o wide`: 134.213.168.26/.37/.44/.54/.56). Pin EPP to IPv4 (`-4`) regardless — the two families get different treatment.
- **source:** `domains_cloudflare_rollout/` RUNBOOK + NOTES 2026-08-04; WRONG_CALLS 2026-08-04 (two rows)
- **added:** 2026-08-04, domains_cloudflare_rollout lane

### `sites.locked_at` holds the QUEUE, not the site — a locked site with 0 armed items can still have a page overwritten, and the gate query will keep saying "held" while it happens
- **footprint:** `sites.locked_at`, `sites.locked_by`, `find_dispatchable_site` (`agent_definitions` type `build-pipeline-trigger`), `TRIGGER_nav_rebuild.sh`, `cta_link_integrity/scripts/049b_deploy_single_page.sh`, any "hold this production site" plan, any handoff asserting a site is safe
- **fires when:** you lock a live site before doing something risky, prove the hold with the gate's own SQL (it answers `NOT SELECTABLE — held`), confirm 0 armed items, and conclude the site cannot change. Every one of those readings is correct and re-checkable, and the conclusion is still false
- **the tell:** **there is none in the lock, because the lock is genuinely working** — the clause `s.locked_at IS NULL` lives in `find_dispatchable_site`'s query and gates **work-item dispatch only**. A direct `orchestrate` publish to `system.agent.generic.requests` never reads `sites`, so `page-rerender` and `nav-updater` fired by hand bypass it completely. That is not an exotic path: **it is the documented bypass this estate reaches for whenever the dispatch queue is slow**, so the sessions most likely to fire it are the ones being careful. Proven 2026-08-04: `mortgagecalculator.co.uk` had `locked_at` set continuously for 33 h, 0 armed items, and its live `/index.html` was still rerendered and deployed over the owner's original (11,125 B → 27,546 B) by another lane verifying `bugs_open/191`'s fix against the very site the bug file named as the reproduction
- **the check:** do not ask the control, ask the artefact. **Before** anything risky, record a restore point (`git log -1 --format=%h -- <domain>/<page>`); **after**, run `git log --format='%h %ci %s' -- <domain>/<page>` — one line names any deploy that touched it, where `locked_at` and the item census both stay reassuring. For a page that must not change, treat the lock as necessary and never sufficient, and say so in the handoff: **"locked + 0 armed" means "nothing can change via the queue"**, which is a strictly weaker claim than the one it reads as. Restoring is cheap if the restore point exists — `git show <sha>:<path> > <path>`, commit by pathspec, push **rebased not merged** (a merge makes the deploy's `git diff HEAD~1 HEAD` drop the domain while the run still goes green)
- **source:** `mortgagecalculator_couk_adoption` lane, 2026-08-04; incident + owner-directed restore in that lane's NOTES and `WRONG_CALLS.md`
- **added:** 2026-08-04, mortgagecalculator adoption lane

### The layout seed driver's "re-running is safe" header is FALSE for six of eighteen layouts — a refresh silently reverts July's live-only changes
- **footprint:** `docs/agent_docs/docs024_key_docs_latest/layouts/003_layouts_seed_driver.sql`, `layouts` table (`css_template`, `structure_tokens`), `layout_*.sql`
- **fires when:** you touch the layout library — a CSS fix, a new layout, a "refresh the templates" tidy — and reach for the driver because its own header says re-running is idempotent-safe (`ON CONFLICT DO UPDATE`). The apply succeeds, the census you wrote passes, and five layouts quietly lose their 2026-07-02 live-only changes: `brochure-bold`, `media-grid`, `docs-sidebar`, `high-energy`, `affiliate-hub` all have live `css_template` ≠ seed file (measured 2026-08-05, md5 of the dollar-quoted block vs `md5(css_template)`). A sixth, `tool-portal-light`, has **no seed in this directory at all** (its truth lives in `idea_uk_section_data_missing/`), so the driver neither refreshes nor protects it and any census keyed on "files in this dir" undercounts the live table (17 files ≠ 18 rows).
- **the tell:** none at apply time — `ON CONFLICT DO UPDATE` reports success identically for a no-op refresh and a clobber, and `updated_at = NOW()` overwrites the one column that would have dated the loss.
- **the check:** before ANY driver run, diff live against file per layout: extract each `$LAYOUT$…$LAYOUT$` block, `md5` it locally, compare with `SELECT md5(css_template) FROM layouts WHERE name=…` (worked script: bugs_open/200's trail). Zero drift → driver is safe. Any drift → fix the live rows surgically (`replace()` with a pre-measured exact-hit count and a DO/RAISE verify — seed `314` is the worked example) or backport the live change into the seed FIRST. Note `layout_16_17_vonc_gamesdesign.sql` does not use `$LAYOUT$` quoting — a block-extraction script parses 0 of its 2 layouts and must say so rather than pass.
- **source:** `bugs_open/200` (found during its fix, 2026-08-05); seed `314_hamburger_toggle_gets_flex_direction_column.sql`
- **added:** 2026-08-05, bugfix_200 thread

---

### `make deploy-component-render-check` ships NOTHING on its own — the overlay pins the tag, and both the make target and kubectl report success anyway

- **footprint:** `deployments/kustomize/services/component-render-check/overlays/production/uk_001/kustomization.yaml`,
  `make deploy-component-render-check`, `make build-component-render-check`,
  `component-render-check`, and any CronJob service whose overlay carries an
  `images:`/`newTag:` block
- **fires when:** shipping a change to this check. You bump `IMAGE_TAG`, build, push and
  deploy — the documented sequence, in order, all four succeeding.
- **the tell:** it looks like a clean deploy and it is a no-op. `make deploy-*` prints
  **`CronJob deployed. Next run:`** unconditionally, and `kubectl apply -k` prints
  **`cronjob.batch/component-render-check unchanged`**. Both statements are TRUE. The
  reason there is nothing to change is that the overlay pins `newTag: v1.0.1250` in the
  file, so the manifest kustomize renders is identical to what is already live — your new
  image is in the registry and the cluster keeps running the old binary. Nothing in the
  output says "the image you just built was not deployed". Measured 2026-08-05: build,
  push and deploy all succeeded and the CronJob was still on `v1.0.1250`.
- **why the fleet habit does not protect you here:** the chassis and the other 13 backend
  services take their tag from `IMAGE_TAG` at apply time, so bumping the makefile IS the
  deploy. This service does not — its tag lives in the overlay. The muscle memory that is
  correct everywhere else is exactly wrong here.
- **the check:** bump `newTag` **in the same commit as the rebuild**, then read the
  ARTEFACT, never the make target:
  ```bash
  kubectl -n ai-persona-system get cronjob component-render-check \
    -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'
  ```
  `configured` rather than `unchanged` in the apply output is the cheap positive signal;
  the jsonpath above is the one that cannot be misread. Then trigger a run and read the
  POD's `state.terminated.exitCode` — a Job is not a pod and a log line is not an exit
  code.
- **the same trap, the other direction:** grep the image you built for a string your
  change ADDED, and grep the **currently deployed tag** for the same string expecting 0.
  Where a change removes no literal there is no removed-string control available, and the
  old image is the honest substitute — a synthetic control only proves the grep works.
- **source:** 2026-08-05, bugfix_140 lane, shipping the clone-identity fix. Sibling of the
  existing `imagePullSecrets` entry on the same CronJob: that one is how the check dies
  silently at RUN time, this one is how it never arrives at DEPLOY time.
- **added:** 2026-08-05, bugfix_140 lane

### A git-adapter publish with `repo_name: "sites"` SUCCEEDS for a domain that serves from vm-sites — every log line green, nothing deployed
- **footprint:** `sites.github_repo`, `system.adapter.git.requests` (`action: commit`), `brochure_component_library/scripts/deploy_stylesheet_direct.sh`, `gqls/sites`, `gqls/vm-sites`
- **fires when:** you publish a file to a site by hand via the git-adapter — typically by reusing the brochure lane's proven `deploy_stylesheet_direct.sh`, which hardcodes `"repo_name": "sites"`. That default is correct for every domain whose `sites.github_repo` is NULL (13 of 15 in bug 200's fleet publish) and silently wrong for the ones that name `vm-sites` (`idea.uk`, `relojistas.com`, measured 2026-08-05). The commit lands cleanly in the WRONG repo: kcat exits 0, the adapter logs `HandleCommitAction` with no error, GitHub shows the commit — and the served file never changes, because the serving pipeline reads the other repo.
- **the tell:** the served file is byte-for-byte the ORIGINAL after the settle wait, while your correlation id appears error-free in every git-adapter replica's logs. Failure-shaped absence with success-shaped evidence.
- **the check:** before publishing, read the row: `SELECT domain, COALESCE(github_repo,'(null=sites)') FROM sites WHERE domain='<d>';` and set `repo_name` accordingly (parameterized variant: `deploy_file_direct_v2.sh` in bug 200's trail, `REPO_NAME=vm-sites`). After publishing, verify at the SERVED file, never at the adapter logs or the repo — `curl -s https://<d>/<path> | diff - <local>`. Note the stray commit in the wrong repo is INERT but real: bug 200's first batch left `idea.uk/` and `relojistas.com/` css commits in `gqls/sites`, recorded not reverted (forward-only).
- **source:** `bugs_open/200` §8 (the two-straggler diagnosis, 2026-08-05); `bugfix-131`'s memory line was the same column from the reading side
- **added:** 2026-08-05, bugfix_200 thread

### A bug file's FIX CANDIDATE can be refuted by that same file's own MEASUREMENT NOTES — they are written for different readers and nothing joins them up

- **footprint:** `bugs_open/`, `bugs_closed/`, any `NNN_HANDOFF_*.md` § "Fix candidates" / "How to verify", `docs/agent_docs/docs024_key_docs_latest/*/PLAN_*.md`
- **fires when:** you pick up a well-written bug file and implement the candidate it recommends. It needs no symptom and gives no warning — the file is *correct*, thorough, and internally contradictory, and the contradiction sits in the half you did not need to read to start work.
- **the trap:** a handoff has two audiences. The measurement/census sections are written for a future person **measuring** the problem, and carry the traps in the data. The fix-candidate section is written for a future person **fixing** it, and carries a proposed key or predicate. Nobody re-reads the first against the second, so a candidate can propose exactly the predicate the census footnote has already shown to be unsafe. **Worked case, `bugs_open/156`:** candidate 1 says dedup on `(slot_name, md5(content_data))`; the census footnote **forty lines above it** records finetuning.uk/our-position-on-ai holding two rows with **NULL `content_data` on both**. Under that key they are "identical" — so implementing the file's own recommendation would have DELETED a live section, on a shape the same file flags as a trap. Shipping the corrected key (`+ rendered_html`) is what made the fix safe.
- **the tell — there isn't one, which is the point.** A good bug file reads as authoritative precisely where it is most dangerous: the candidate is confident, specific, and written by someone who had just done the measurement. Its being *nearly* right is what gets it implemented.
- **the check, and it costs one pass:** before implementing any candidate, **read the file's measurement/census/landmine sections AGAINST the candidate**, and for each column or field the candidate keys on, ask *"what does this file itself say about the values in that column?"* Specifically hunt for the degenerate rows — NULLs, empties, zero-counts — because a census footnote about them is the commonest form this takes. Then state the corrected predicate and **why it differs from the file's**, so the next reader inherits the correction rather than the original.
- **the same shape one level up:** a `count(DISTINCT md5(col))` of **0** means `col` is NULL on every row, not that the rows agree. Any candidate keyed on that column is unsafe for exactly those groups, and a census filtering `col IS NOT NULL` cannot see them at all.
- **source:** 2026-08-05, `bugs_closed/156` (`docs024_key_docs_latest/bugfix_156_duplicate_sections/`). Caught by reading the whole file before writing code — not by anything that failed, and the tests for the wrong key would all have passed.
- **added:** 2026-08-05, bugfix_156_duplicate_sections lane

### A verification run that dispatches work at a site has THREE silent ways to complete green having never executed the code under test — and the run's status distinguishes none of them

- **footprint:** `platform/orchestration/actions/load_work_item_actions.go` (`LoadWorkItemsAction`, :127-139 lock check, :623-661 predicate), `site-work-orchestrator` (`load_work_items` → `check_has_items` → `build_items_loop`), `sites.locked_at`, `site_work_items.status`/`handler_agent`/`pipeline`, any `mode=maintenance` dispatch used as an acceptance test
- **fires when:** you have shipped a fix inside a work-item loop and want to prove it live, so you fire the orchestrator at a site you believe has suitable work queued. It needs no symptom: the run completes, status `COMPLETED`, `current_step: complete`, no error, and the pages it did touch look right.
- **the trap — all three return SUCCESS with zero items, and `has_items == false` routes past the code you are testing:**
  1. **The item predicate is far narrower than "open work".** `load_work_items` requires `status IN ('triaged','approved')`; the orchestrator step adds `pipeline='build'` **and `handler_agent='page-content-writer'`**. A site can hold 40+ non-terminal build-pipeline items and still load **zero**. Measured 2026-08-05: seven sites listed as candidates by their "open build-routed item" counts (6–17 each) were **all 0**; `page-content-writer` has held **14 items fleet-wide in all of history**.
  2. **The site lock short-circuits before every filter** (:127-139) and returns `{"items": [], "count": 0, "skipped_reason": "site_locked"}`. **This is the deceptive one** — the site genuinely has qualifying work, so every count you measured beforehand was right, and only `skipped_reason` says why nothing ran.
  3. **An idle site** — the honest zero, and the only one most people think to guard against.
- **the tell:** there is none in the status, which is the entire problem. The discriminator is **`skipped_reason` in the step result**, plus whether the loop's own steps appear in `execution_path` at all. A run whose `execution_path` goes `load_work_items → check_has_items → load_fix_items` never entered the build loop, whatever it reports.
- **the check, BEFORE dispatching — the lock and the real predicate in one query:**
```sql
SELECT s.domain,
       CASE WHEN s.locked_at IS NULL THEN 'UNLOCKED' ELSE 'LOCKED by '||COALESCE(s.locked_by,'?') END AS lock_state,
       count(wi.id) AS loadable
FROM sites s
LEFT JOIN site_work_items wi ON wi.site_id = s.id
     AND wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
     AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')
     AND wi.pipeline='build' AND wi.handler_agent='page-content-writer'
WHERE s.domain = '<target>' GROUP BY 1,2;
```
  **A non-zero `loadable` on an UNLOCKED site is the only combination that can prove anything.** Induce a positive control first: run it fleet-wide and confirm it returns *some* non-zero row, or you cannot tell a real 0 from a mistyped predicate.
- **and the trap in the other direction:** when the query says the only qualifying site is LOCKED, **the answer is not to unlock it.** `aee11cb90` is the incident where a live homepage was rebuilt under a held lock on that very site (`mortgagecalculator.co.uk`), and the lock is the control added afterwards. A lock reading *"held pending owner decision"* is a decision upstream of your check, not an obstacle to it.
- **source:** `bugfix_194` lane, 2026-08-05 — check 3b, which had **no runnable target on any site in the estate** once all three were applied. Related but distinct: the `sites.locked_at` entry above covers a gate that *failed to read* the lock; this one covers gates that read it correctly and report success anyway.
- **added:** 2026-08-05, bugfix_194_sections_metadata_mapping lane

### Marking a page section `build_status='removed'` does NOT remove it from single-page rerenders — the section RESURRECTS, and the "successful" deploy is an EMPTY git commit

- **footprint:** `page_components.build_status`, `rerender_single_page_action.go`, `getPageSections`, `page_rerender`, `pages.sections`

Hit 2026-08-05 removing idea.uk's home tool-list section. The full-build
assembler excludes `build_status='removed'` (`v3_site_actions.go:3919`), so
"set removed + rerender" looks like the complete recipe. It is not:
**`getPageSections` (`rerender_single_page_action.go:777-780`) selects EVERY
`page_components` row for the page — no `build_status` filter, and it does
not intersect with `pages.sections` either.** So a single-page rerender
re-assembles the removed section from its stored `rendered_html`, byte-identical
output, and the git-adapter dutifully reports `success:true` on an **empty
commit** — the work item completes, the deploy_result lists the files, and
nothing changed. Two wrong conclusions offered themselves before the code
read: "the deploy skipped" (it didn't) and "the sections prune failed" (it
hadn't).

**the check:** to remove a section so BOTH assembly paths agree: delete the
`site_plan_sections` row + prune `pages.sections` + set
`build_status='removed'` **+ empty the row's `rendered_html`** (tombstone —
keep the row so history/dedup survive), then one `page_rerender` (spec needs
`domain`/`page_id`/`page_name`/`filename` — a bare page_name fails
resolution). Verify at the SERVED page and at a NON-EMPTY vm-sites commit
(`git show <sha> --name-only` — subject-only output = empty commit = nothing
deployed). And note `pages.sections - 'name'` only works because sections is
a STRING array — check the shape before trusting the prune.

- **source:** 2026-08-05, idea_uk_vm_site lane; RUNNING_NOTES §X.44-45
- **added:** 2026-08-05, ideauk-sec session

---

## `page_component_history.component_id` points at `page_components`, NOT the component library — and it is NULL on every row
- **footprint:** `page_component_history`, `page_component_history.component_id`,
  `content_components`, any join written to answer "which COMPONENT did this
  historical `content_data` belong to", `save_page_sections_overwrite`
- **fires when:** you use the history table as evidence about components — sizing a
  defect class, attributing a bad `content_data` shape, asking which component
  minted something. The column is named exactly like `page_components.component_id`,
  which *is* a FK to `content_components`, so the join writes itself and returns
  rows.
- **why the wrong answer looks right:** the FK is
  `page_component_history_component_id_fkey ... REFERENCES page_components(id) ON
  DELETE SET NULL`, and `save_page_sections` **DELETEs and re-INSERTs** its rows on
  every save — so the parent is gone and the column is NULL. Measured 2026-08-05 on
  the envelope-shaped subset: **67 of 67 rows NULL, 0 resolvable.** A `LEFT JOIN`
  therefore yields NULL columns, not zero rows, and a `CASE` that tests
  `c.input_schema IS NULL` before `c.id IS NULL` reports every one of them as "a
  component with no schema" — a real, plausible class, and the one you were probably
  counting.
- **the check:** `\d page_component_history` before writing the join — the FK target
  and the `ON DELETE SET NULL` are both printed. Then gate on resolvability before
  classifying anything:
  ```sql
  SELECT count(*) FILTER (WHERE component_id IS NULL) AS fk_nulled,
         count(*) FILTER (WHERE component_id IS NOT NULL) AS fk_live
  FROM page_component_history WHERE <your predicate>;
  ```
  If `fk_live` is 0, the table cannot answer a component-level question at all and no
  arm ordering saves you. Reach for `page_id` + `site_id` (both real, both populated)
  and accept that you are measuring PAGES, or take the question to
  `page_components` live. **Put the `IS NULL` arm first in any `CASE` over a
  `LEFT JOIN`** — that is the general form, and it is one line.
- **source:** 2026-08-05, `bugs_open/199` lane; `WRONG_CALLS.md` same date

---

## `extractContentWithFallbacks` leaks a transport envelope by TWO branches, and the one every doc names is the wrong one
- **footprint:** `platform/orchestration/actions/v3_site_actions.go`
  (`extractContentWithFallbacks` :4476-4528, `hasContentFields` :4531-4546,
  `RenderComponentAction` :1768), `content_from` / `content_field` in any
  `render_component` step config
- **fires when:** you set out to stop an LLM transport envelope
  (`{"type":"text","result":"<string>"}`) reaching a component's content, and you go
  to the branch `bugs_open/199` and the `016b` §10 row both name — the "last resort"
  branch, the one guarded by `hasContentFields`. Fixing it alone leaves the live path
  wide open and every test you write will still pass.
- **why the wrong answer looks right:** for the live config
  (`page-content-writer`'s `render_section`, `content_from: "generated_content.result"`)
  the leak is the **FIRST fallback loop** (:4494-4503), which has **no content check
  at all** — just `len(m) > 0`. `pathsToTry[0]` = `generated_content.result` resolves
  to the envelope's `result` **string**, the map assertion fails and the loop
  *continues*; `pathsToTry[1]` = `generated_content` resolves to the envelope **map**
  and is returned. `hasContentFields` is never consulted. Meanwhile the last-resort
  branch is not dead either: a superset envelope like `{content,result,type}` passes
  `hasContentFields` precisely on its `content` key. **Two doors, and the documented
  one is the quieter of them.**
- **the check:** guard the **caller**, not the resolver — one call at
  `RenderComponentAction` covers both branches and cannot be reopened by someone
  "fixing" the first loop (this is what `normalizeRenderContentEnvelope` in
  `render_content_envelope_guard.go` does, PBP-032's third seam). If you must reason
  about the resolver itself, trace it against a REAL `content_from` from live config,
  not from the function in isolation:
  ```sql
  SELECT s.key, s.value->'config'->>'content_from'
  FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
  WHERE d.type='page-content-writer' AND d.is_active
    AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL;
  -- NB workflow.steps is an OBJECT keyed by step name, not an array; and the
  -- render steps live inside process_sections_loop.config.sub_workflow.steps
  ```
- **source:** 2026-08-05, `bugs_open/199` lane; correction recorded in the bug file,
  the `016b` §10 row and PBP-032

---

## The save seam's identity paths resolve to NOTHING inside the page writer's own run
- **footprint:** `writeContentDataEnvelopeLog`, `renderEnvelopeIdentity`,
  `site_record.site_id`, `current_page.name`, `agent_error_log.site_id`, any new
  durable record written from an action inside `page-content-writer`
- **fires when:** you add an `agent_error_log` (or any attributed) record to an
  action that runs inside the page writer, and you copy the identity paths from the
  nearest existing writer — `save_page_sections`' guards all use
  `site_record.site_id` and `current_page.name`.
- **why the wrong answer looks right:** the guard fires correctly, the INSERT
  succeeds, the row appears. It is just **unattributable** — `site_id` NULL, page
  name empty — and you will not notice, because you are checking that the guard
  fired. Measured 2026-08-05 across every stored `page-content-writer` run (n=110):
  `site_record.site_id` **0/110**, `current_page.name` **0/110**. The writer receives
  its work under `input_data`, and rebuilds a `render_context` beside it.
- **the check:** measure the paths against real runs before choosing them —
  ```sql
  SELECT count(*) AS runs,
    count(*) FILTER (WHERE collected_data #> '{site_record,site_id}' IS NOT NULL) AS save_seam_path,
    count(*) FILTER (WHERE collected_data #> '{input_data,site_id}' IS NOT NULL)   AS writer_path
  FROM orchestration_states WHERE owner_agent_type='<your agent>';
  ```
  What works at this seam: site = `input_data.site_id` → `render_context.site_id`
  (110/110); page = `input_data.current_page.name` → `input_data.page_name` →
  `render_context.current_page` (110/110 union — **and the last is a plain STRING,
  not a map**, so a `#>'{render_context,current_page,name}'` read finds nothing).
- **source:** 2026-08-05, `bugs_open/199` lane; `render_content_envelope_guard.go`

- A Tier-4 acceptance run opens ONE browser page per (url, profile) and runs
  EVERY check against it with no reset — harmless for stateless checks, and
  unsatisfiable for a one-shot consent/disclaimer gate (hides itself after the
  first click, deliberately, because the tool carries a legally load-bearing
  disclaimer). The second interaction check that must also click it fails
  Playwright's "element is not visible", which reads exactly like a broken
  button and is not one. The failing run then dispatches `tool-improver`
  straight at the fence, and the only ways to make it pass weaken or delete the
  disclaimer — a human cancelled the one real occurrence by hand before it
  dispatched; luck, not a guard.
    footprint `internal/adapters/browserrunner/run_checks_action.go` (interaction
    checks, `chromiumPage.Do`), any criteria fence with more than one
    `interaction` check on a page carrying a one-shot/self-hiding control
    the check: a later interaction check that must re-click a one-shot control
    needs `{"action":"reload"}` as its FIRST step (fixed 2026-08-05,
    `bugs_open/126` — TL-040, commit `67a4c50bd`), which resets the shared
    page to its landing state before the check's own steps run. A fence whose
    failure must never be auto-repaired (a real disclaimer/consent gate) should
    also carry top-level `"no_auto_fix": true` (+ `"no_auto_fix_reason"`), which
    routes a genuine failure to the existing `acceptance_stuck` human-review
    escalation instead of `tool-improver` — an automated rewriter can only turn
    such a fence green by weakening the markup it exists to protect. Both keys
    are opt-in and inert until a fence names them, and both are unshipped to
    production until the next `browser-runner-adapter` + `agent-chassis` roll —
    verify at the pod (`grep -ac` two distinct literals + a positive control;
    the fleet's images carry no `strings` binary), never at the tag.

### A component field whose `source` is not `"llm"` makes the content writer skip the LLM ENTIRELY and re-render from template — the run reports success and the copy never changes
- **footprint:** `content_components.input_schema.fields.*.source`, `platform/orchestration/actions/plan_sections_action.go` (`llmFieldSpecs`, `LLMFieldSpecs`), the live `page-content-writer` step `check_render_mode`, work item types `content_rewrite` / `needs_content_page`
- **fires when:** you ask the framework to rewrite a page's copy — a `content_rewrite` item, a voice seeded into `content_direction`, a rewrite guidance — and the component holding that copy declares its field with any `source` other than `"llm"`. **Nothing fails.** The work item completes, the page re-renders, the bytes are identical, and the natural reading is "the model ignored my guidance" or "the seed didn't reach the prompt". Both are wrong: **no model was ever called.**
- **the mechanism, three links, each verifiable on its own:** `plan_sections_action.go:1708` appends to `llmFieldSpecs` only `if source == "llm"`. The struct field is `json:"llm_field_specs,omitempty"` (`:777`), so an empty list serialises as **absent, not `[]`**. The live writer then branches on exactly that — `check_render_mode`: `condition: "current_section.llm_field_specs != null"`, `else_step: "render_from_template"`. So the section takes the no-LLM path.
- **the tell:** the section renders with its OLD content and `llm_call_log` has no row for the run. Check the schema before blaming the prompt:
  ```sql
  SELECT cc.function, f.key, f.value->>'source', f.value->>'type'
  FROM content_components cc, LATERAL jsonb_each(cc.input_schema->'fields') f
  WHERE cc.function = '<the component>';
  ```
- **the check, before you conclude a seed is inert:** the writer reads exactly ONE field of `content_direction` — `{{.site_specs.specs.content_direction.formatted}}`, live in `page-content-writer.prompt_template`. Every other key reaches the prompt only by being serialised INTO `formatted` by `datahelpers.FormatContentDirection`. **A hand-written `content_direction` that does not regenerate `formatted` is invisible to the writer, and looks applied.** So there are TWO independent ways for a voice change to be a silent no-op, and they present identically. Rule them out in this order: is `formatted` regenerated, then is the field `source: "llm"`.
- ⚠ **`source: "authored"` is not a safety guard, it is a factual claim** — "a human supplied this, do not regenerate". Before changing one, ask whether the claim is TRUE. On loancalculator.co.uk it was not: the prose was another model's output lifted byte-for-byte by the decomposer, so `authored` was a mislabel and correcting it was the fix (owner ruling 2026-08-05). **But `ported-page.body` IS genuinely authored** — it is the byte-preserving `--fidelity locked` adoption path, and flipping it would let a writer rewrite a site adopted precisely to be preserved. Two fields fleet-wide carried `source: "authored"`; exactly one was wrong.
- **source:** loancalculator.co.uk lane, 2026-08-05, found while trying to run the owner's "rerun it through the framework" instruction. Measured before changing anything: `ported-prose` exists on loancalculator and **no other site** (51 rows), and `authored` fields fleet-wide went 2 → 1, which is the negative control that exactly one field moved
- **added:** 2026-08-05, loancalculator_couk lane

### The Gauntlet engine validates the provocation feed's `today` key for PRESENCE only — change its SHAPE and every check still passes, then the blob goes into the AI prompt
- **footprint:** `internal/tools-api/handlers/round.go` (`FetchProvocation`, `provocStore`), `platform/orchestration/actions/provocation_feed_action.go` (`checkFeed`, `asToday`), the served artefact `https://<domain>/data/provocations.json`, `gauntlet_rounds.provocation`
- **fires when:** you change what `today` holds — adding categories (a map of category → provocation), nesting it, renaming its fields, or "tidying" the shape. **Nothing errors.** A round is created, a 200 is returned, and the site looks fine.
- **the mechanism:** `FetchProvocation` makes exactly THREE checks and no more — the key exists (`round.go:73-76`), the value is not `null`, the value is not zero-length (both `round.go:78`). It then returns `json.RawMessage` **without ever parsing it**. `store.CreateRound` writes those raw bytes to a `jsonb` column, and `position.go:67` / `defend.go:67` interpolate `string(round.Provocation)` **straight into the model's prompt**. No struct, no field access, no schema — `headline`, `teaser`, `slug` and `detail_body` appear NOWHERE in `internal/tools-api/**.go` outside tests. So a shape change is not caught, it is *served*: the AI argues against whatever JSON you put there.
- **why the obvious reassurance is wrong:** the function's own doc comment says it fails loud "per bug_historian pattern #7 rather than returning a blank provocation", and it does — for an ABSENT or NULL `today`. That designed-loud path reads like shape validation and is not. Presence is not shape.
- **the check, before you change the feed's shape:** the only enforcement of `today`'s fields lives in the WRITER (`checkFeed`, which does demand non-empty headline/body/slug/date). A writer-side invariant cannot protect a reader in a different binary on a different host — and `go list -deps ./platform/orchestration/actions | grep tools-api` returns **no rows**, so the two sides share no Go type and no compiler will ever compare them. Verify a shape change by fetching the artefact and asserting the fields a round actually needs, then run one real round end to end; do not infer it from a 200.
- ⚠ **`provocStore` is keyed by DOMAIN alone** with a 5-minute TTL (`round.go:25-29` — `provocTTL` at `:25`, `provocStore` at `:29`). Any per-category feed must add a category dimension to that key or categories serve each other's provocations for up to five minutes — a staleness bug that only appears under concurrent categories and vanishes while you debug it.
- **source:** `provocation_pipeline` lane, 2026-08-05, deriving the engine contract for `RFC_013` (per-category provocations). Read first-hand end to end; the two negatives (nothing unmarshals it, no field is named server-side) were grepped rather than assumed, because the whole of RFC_013's recommendation turns on them
- **added:** 2026-08-05, provocation_pipeline lane

### `banned_claims` validation sweeps CONTENT PROSE only — the `<title>`, JSON-LD and head escape it, and a clean validation reads as full coverage

- **footprint:** `site_specs` `evidence_base.banned_claims` · `validate_page_content.go` ·
  `pages.title` · `pages.meta_description` · JSON-LD/structured data · any site seed
  copying `oufe`'s or `webdesign.uk`'s pattern
- **fires when:** you seed `banned_claims` to enforce a copy rule (a banned word, a
  style rule like an em-dash ban) and trust a PASSING validation to mean the served
  page is clean. The sweep runs over the writer's content; **`pages.title` is
  rendered into `<title>` and mirrored into JSON-LD by the head builder without
  passing through it.** Measured 2026-08-05 on webdesign.uk's first passing build:
  validation PASSED while the served page carried the banned em dash in the title
  and its JSON-LD mirror — exactly the phrasing class the ban existed to stop
- **the tell:** none at the pipeline — the page completes and validation is green.
  Only the ARTEFACT shows it
- **the check:** after the first passing build of any site with seeded bans, fetch
  the SERVED page and run every ban over it yourself — then **triage the hits,
  because raw HTML false-positives**: CSS (`grid-template-columns` hits a
  `template` ban; comments carry em dashes) and quote-shaped regexes match
  minified head content. Prose-context bans only bind prose. And fix title/meta at
  their SOURCE (`pages.title`, `pages.meta_description`) — they are data, not
  writer output, so no re-roll fixes them
- **source:** webdesign.uk 2026-08-05; the seed's own `[UNVERIFIED]` note flagged
  this exact question at write time. `webdesign_uk_build_service/NOTES` 08-05
- **added:** 2026-08-05, webdesign.uk build-service lane

### The offline fence harnesses run HEAD's evaluator — vocabulary newer than the deployed browser-runner passes offline and FAILS (or skips) in-cluster

- **footprint:** `staged_component_build/scripts/try_fence.go` · `prove_fence_can_fail*.go` ·
  `internal/adapters/browserrunner/run_checks_action.go` (`criteriaStep.Action`, the
  `default:` arms) · any ```criteria fence in `doc_plans` · `reload` step action specifically
- **fires when:** you author a fence using ANY check type or step action and prove it with
  the offline harnesses. They import `internal/adapters/browserrunner` **at HEAD**, so a
  vocabulary word another session landed an hour ago evaluates perfectly on your machine —
  the trial passes, the mutation prover passes, the persist round-trips — and the deployed
  pod has never heard of it. Measured 2026-08-05: `reload` landed at HEAD 11:05 UTC
  (67a4c50bd, bugs_open/126); the running browser-runner v1.0.1252 was built 09:10 UTC; a
  fence authored 12:10 UTC passed 17/17 offline three separate ways, then failed its S6 run
  with `unknown step action "reload"` — and the failing verdict raised a live `improve_tool`
  item routing tool-improver at a healthy tool. Two failure shapes, one cause: an unknown
  CHECK TYPE **skips** (reads as PASS upstream, the older sibling entry above); an unknown
  STEP ACTION **fails the check** (reads as a tool defect and dispatches a fixer)
- **the check:** before persisting a fence, grep the RUNNING pod for a long runtime string
  from every vocabulary word the fence uses — step actions too, not just check types
  (`"reload navigation failed"` → 0 on v1.0.1252, positive control → 1, same exec). Faster
  triage: `git log -1 --format=%ci -S '"<word>"' -- internal/adapters/browserrunner/` —
  a date newer than the pod's image build is a red flag by itself. And when a dispatched
  run fails on this, the `improve_tool` item is a false positive: cancel it with the reason
  written into `result`, or a fixer rewrites a healthy tool
- **source:** staged_component_build calibration tranche, 2026-08-05 — S6 correlation
  3874c8b5-63bb-44d5-93ec-f2086f63567c, 16 passed / 1 failed on vocabulary alone. NOTES
  entry the same day carries the full timeline
- **added:** 2026-08-05, staged_component_build lane

### `params.StorageClient` is nil on the agent-chassis, and every storage CREDENTIAL being present is what makes that invisible

- **footprint:** `params.StorageClient`, `platform/agentbase/agent.go`, `IMAGE_BUCKET`,
  `S3_ENDPOINT`, `execute_vision_prompt`, `platform/orchestration/actions/execute_vision_prompt_action.go`,
  `deployments/kustomize/services/agent-chassis`, `storage.NewS3Client`
- **fires when:** you seed a chassis agent to call an action that takes
  `params.StorageClient`. It fails at runtime with **"no storage client — cannot
  download screenshots"**, and every pre-flight check you would naturally run says
  storage is fine
- **the tell — and why the obvious check LIES:** `env | grep -iE 's3|storage|bucket|b2_|aws'`
  on the chassis returns `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`,
  `B2_APPLICATION_KEY_ID`, `B2_APPLICATION_KEY`, `AGENT_STORAGE_SECRET`,
  `AGENT_STORAGE_CONFIGMAP` — six storage variables, all set, and it reads as
  "storage is configured". **Credentials are not the gate.** `agent.go:308-330`
  builds the client only `if storageConfig.Bucket != ""`, and the bucket comes from
  **`IMAGE_BUCKET`**, with the endpoint from **`S3_ENDPOINT`** — neither of which was
  set on the chassis (measured 2026-08-05). Nil client, silently, for every action
- **why nobody had hit it:** `execute_vision_prompt` is the **only** chassis action that
  takes `params.StorageClient`. Every other storage-using action builds its own client
  from agent/step config (`storage_actions.go:95` and `:612`) or from the service config
  (`internal/adapters/browserrunner/screenshots.go:66`, `cfg.Infrastructure.ObjectStorage`
  — which is why the browser-runner uploads renders perfectly well **while also having no
  `IMAGE_BUCKET`**). Three storage-config paths in one estate, and only one of them is
  wired on the chassis
- **the check:** never infer the client from the credentials. Ask for the **bucket**:
  `kubectl exec -n ai-persona-system <chassis-pod> -- printenv IMAGE_BUCKET` — empty means
  `params.StorageClient` is nil no matter what else is set. Then confirm at the log line
  the code actually emits: `agent.go` logs `"Storage client not configured (IMAGE_BUCKET not set)"`
  at startup, which is the honest signal and is easy to miss among startup noise
- **the deeper trap, which is the transferable part:** **"built + wire-shape tested" cannot
  discover this class.** MDL-040's tests assert request BODIES for both providers — good
  tests, and they pass in a world where the action can never obtain a client. A capability
  with no live caller has an untested dependency on its ENVIRONMENT, and the first real
  call is what finds it. Treat "no live call yet" in the register as "the deployment
  contract is unverified", not merely "unused"
- **the fix:** the chassis overlay now sets `IMAGE_BUCKET`/`S3_ENDPOINT`/`S3_REGION`
  (`820a033c0`), matching the `business-intel` overlay, which carries the same comment
  because that lane hit the same wall earlier. **Hardcode the values; do NOT use
  `configMapKeyRef`** — a wrong key name is `CreateContainerConfigError` and the entire
  chassis stops rather than one action failing. **Requires a chassis ROLL to take effect,
  and a roll kills any in-flight council** — wait for a clear window
- **source:** 2026-08-05, first live call of `execute_vision_prompt` (TL-035 (e), seed 317)
  on `tool-acceptance-agent`; run `25fee04c-6cc8-40b4-92af-da81fa3f8b16`, which reached
  `complete_no_look` with the message above in `__step_error`
- **added:** 2026-08-05, brochure component library lane

### `git_commit`'s `commit_message` template resolves ONLY `{domain, file_count, filename}` — any other key renders `<no value>`, and the commit succeeds anyway
- **footprint:** `platform/orchestration/actions/git_deployer_actions.go` (`buildCommitMessage`, `resolveCommitMessage`), any `agent_definitions` workflow step with `action: git_commit` and a `commit_message` template
- **fires when:** you author a workflow step whose `commit_message` names step data — `{{.input_data.spec.category}}`, `{{.css_fix.result.changes_summary}}` — because the *same syntax resolves correctly* in an `execute_llm_prompt` step's `prompt_template` two lines up. The two templates look identical and are executed against different worlds: the prompt gets CollectedData, the commit message gets a fixed three-key map (`git_deployer_actions.go` `buildCommitMessage`).
- **the tell:** none at config time and none at run time — the commit lands, the step and item read `complete`. The damage surfaces at forensics time, when the audit trail of what each commit was reads `CSS fix: <no value> — <no value>` (all four of `bugs_open/198`'s incident commits; the register's DGH-002 records the same class on the rerender template).
- **the check:** before shipping a `git_commit` step, grep your template for anything beyond `{{.domain}}`, `{{.file_count}}`, `{{.filename}}` — anything else needs the message composed UPSTREAM (a `query_database` step's RETURNING is the proven spot, params resolve there) and `commit_message_field` pointed at it (DGH-007, shipped with migration 318). That field needs a binary carrying `resolveCommitMessage` — inert until the roll that ships it; the template stays as the fallback either way. Do NOT "fix" this by handing CollectedData to the template: that silently changes what every existing fleet template renders.
- **source:** `bugs_open/198` secondary defect; fix = migration 318 + DGH-007
- **added:** 2026-08-05, bugfix-198 session (dispatched at the bug by the owner)

---

## `-l app=agent-chassis` returns 2 pods; **41** run that binary — and a post-roll "both replicas verified" can be true and still not mean live
- **footprint:** `kubectl get pods -l app=agent-chassis`, `kubectl logs deploy/agent-chassis`,
  any post-roll pod-grep, `strings /app/agent-chassis`, `make release redeploy-agents`,
  `docker.io/aqls/agent-chassis`
- **fires when:** you verify a chassis fix after a roll. You grep the two `agent-chassis`
  replicas, both carry your symbol, and you write "LIVE, both replicas pod-verified". That
  sentence is true and it is **not** the claim you need.
- **why the wrong answer looks right:** the chassis binary runs under many names. Measured
  2026-08-05, minutes after a fresh release: **7 pods on the new `v1.0.1254`, 34 still on
  `v1.0.1252`** — `agent-feed-ingester` ×22, `agent-feed-triage` ×5,
  `agent-content-feed-orchestrator` ×5, `agent-model-directory-publisher`,
  `agent-vet-practice-verifier`. Nothing about the label-scoped check reveals them, and both
  available conclusions are wrong by default: "the release fragmented the fleet" (it did not)
  and "they don't matter" (they might).
- **the check:** enumerate by IMAGE, not by label, then ask the only question that decides it —
  can the stale pods reach your code?
  ```bash
  kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.ownerReferences[0].kind}{"\t"}{.spec.containers[0].image}{"\n"}{end}' | grep agent-chassis
  ```
  `ownerReferences.kind` is the discriminator: **`Job`** = a per-work-item pod pinned to the tag
  current when it spawned, correctly not restarted by `redeploy-agents`, ages out on its own.
  **`Deployment`** on an old tag after a release is a real problem. Then, for the Job ones, prove
  reachability rather than assuming it — **with a positive control, or the query cannot fail**:
  ```sql
  SELECT type, (default_config::text LIKE '%<your action>%') AS can_reach
  FROM agent_definitions
  WHERE type IN (<the stale pods' agent types>, '<the type you KNOW uses it>')
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
  The claim you want is **"live everywhere it can be reached"**, which needs both halves. A Go
  action reachable only through a registry entry (no Go call site) makes the second half exact:
  the action name is the whole attack surface.
- **source:** 2026-08-05, `bugs_closed/199` lane, verifying PBP-032's render seam on v1.0.1254

---

## `orchestration_states` has **no `agent_type` column** — and a watch loop that silences its SQL error reports a COMPLETED run as a lost dispatch
- **footprint:** `orchestration_states`, `run_improvement_sweep_once.sh`, any `Monitor`/`while`/`until` poll loop wrapping `kubectl ... psql` with `2>/dev/null` or `|| true`, `correlation_id` progress checks
- **fires when:** you fire anything asynchronous at the fleet (a sweep, a council submission, a diagnosis run) and arm a loop to watch its `correlation_id` land. The natural query is `SELECT agent_type, current_step, status FROM orchestration_states WHERE correlation_id=...` — because every *other* async table in this estate has an agent column, and because `agent_definitions.type` is what you have in hand. That column does not exist here; the table identifies work by `workflow_plan`, `client_id` and `collected_data`, never by agent type.
- **why the wrong answer looks right:** the loop is built to tolerate transients (`2>/dev/null || true`), so the `ERROR: column "agent_type" does not exist` is swallowed on **every** iteration and the variable is empty every time. Empty is exactly what "the row has not appeared yet" looks like. The terminal-state `case` never matches, so the watch runs to its full timeout and reports silence. Two true facts then compound it: publish→run start on this fleet has been measured at up to **~29 minutes**, and a chassis roll really does kill in-flight orchestrations — so "no row after 20 minutes, and a roll happened" is a *fully coherent* wrong story. Measured 2026-08-05: a 20-minute watch saw nothing while the sweep had already finished **14 orchestrations, all COMPLETED, `error` NULL**, inside 11 minutes.
- **the tell:** none from the loop — it looks like a patient watch. The one asymmetry is that a genuinely-waiting loop and a broken one are byte-identical in output, so treat total silence as *unproven*, never as evidence.
- **the check:** run the query **once by hand** before wrapping it in anything, and never send stderr to `/dev/null` in a poll loop — discriminate on OUTPUT and let errors be loud. ~~The columns that exist: `orchestration_id, correlation_id, client_id, status, current_step, collected_data, initial_request_data, final_result, error, workflow_plan, execution_path, created_at, updated_at, parent_orchestration_id, currently_executing, processing_node`. Find a run by payload, not by agent~~: `WHERE collected_data->'input_data'->>'site_id' = '<id>'`. And before concluding a dispatch was dropped, `date -u` — hours pass between turns on this tree, and the timestamps you are calling impossible are usually just older than you think.

> **CORRECTED 2026-08-06 — this entry's column list is TRUNCATED, and "never by agent type" is FALSE.**
> `orchestration_states` has **36 columns**; the struck list is the first 16, i.e. exactly what
> `\d orchestration_states | head -25` prints. Among the omitted ones is **`owner_agent_type`
> (character varying(50))** — plus `owner_agent_id`, `owner_agent_role`, `site_id`,
> `orchestration_name`, `awaited_requests`, `processing_history`, `subtree_agents`,
> `requests_topic`, `responses_topic`, `execution_metadata`, `fuel_budget`.
> **So the table DOES identify work by agent type, and `GROUP BY owner_agent_type` is the right
> query** — four other entries in this same file already rely on it (see the `site-id extractor`,
> `fidelity`/`site-adoption-agent`, `~24h retention`, and `code-index` entries).
> **What stays true:** there is no column named literally `agent_type`, and the silenced-stderr
> lesson — the entry's actual subject — is untouched and still correct.
> **What this cost:** on 2026-08-06 I needed the caller of each `save_page_sections` run,
> followed "find a run by payload, not by agent", and built a fingerprinting scheme over
> `workflow_plan` config values. It was ambiguous (the value I keyed on is shared by four agent
> definitions) and it counted step occurrences rather than runs. One `GROUP BY owner_agent_type`
> replaced all of it.
> **The transferable bit, and it is why this correction is worth its space:** the original
> entry and I made the *same* mistake independently — a `\d` read only as far as the first
> screen. **Never publish a "the columns that exist" list from a truncated `\d`**; a partial
> schema quoted as complete converts one session's impatience into every later session's
> false premise. Read the whole thing, or say which part you read.
- **source:** WRONG_CALLS 2026-08-05 (brochure_component_library, fundamentallyai improvement sweep); RUNBOOK_brochure_component_library.md §"improvement sweep" step 2
- **corrected:** 2026-08-06, code_review_triage_2026-08-05 lane (NOTES §15b, RUNBOOK R10)
- **added:** 2026-08-05, fundamentallyai sweep session

- `kubectl exec ... grep -ac <literal> /app/<binary>` on `agent-chassis` /
  `browser-runner-adapter` (~95MB Go binaries) can take **60-90 seconds per
  invocation** — this looks EXACTLY like a hung exec, and a short/default
  timeout kills it before it returns, which then reads as "the string is
  absent" (a silent `RC=124`/exit-143, not a `0` count) — the false-negative
  shape the pod-grep recipe exists to avoid.
    footprint `kubectl exec ... grep -ac`, `/app/agent-chassis`,
    `/app/browser-runner-adapter`, any pod-verify recipe on a chassis-family
    binary
    the check: budget at least 100-120s per grep, not the tool's default
    2-minute command timeout shared across several chained greps — run ONE
    grep alone first to see its real latency before chaining a positive +
    target + negative control in one `sh -c`. Cause (not confirmed, plausible):
    BusyBox `grep` on a binary with very few newlines effectively scans one
    enormous "line", which is slow by construction; `wc -c` on the same file
    returns instantly, so it is grep-specific, not a slow read. Measured
    2026-08-05, `bugs_closed/126` post-roll verification: a single `grep -ac`
    timed out twice at 20s and 40s, then returned correctly at 100s.

## A call_agent child on a generic-consumed topic cannot select a workflow by agent_type — it silently runs the no-op fallback and COMPLETES

- **footprint:** `platform/orchestration/actions/call_agent.go` (`buildCallRequestMessage`), `platform/messaging/processor.go` (`extractGroupInfo`, `selectWorkflow`), `system.agent.generic.requests`
- **fires when:** you dispatch (or probe) a child via `call_agent` expecting the
  child to run a seeded/named agent definition's workflow. The child request is a
  nested RequestMessage envelope `{headers, body}`; `extractGroupInfo` reads only
  the msgBody TOP level, so `config.agent_type` / `input_data.agent_type` under
  `body` are invisible, Priority 2 group resolution never fires (no "Looking for
  agent group" log line), and selectWorkflow falls to the consumer's own
  definition — on the generic consumer that is the no-op complete step. The child
  answers success and the parent proceeds: the wrong result is indistinguishable
  from the right one unless you read the child's step data (an ECHO of its input
  is the tell). CLI/kcat-published flat bodies (`{action, config, input_data}`)
  DO resolve — 195's probe worked; a probe modelled on it inside call_agent does
  not. Real dedicated-consumer targets (page-content-writer etc.) are unaffected:
  their Priority-3 definition IS their workflow.
- **the check:** before relying on a call_agent child's workflow selection, grep
  the chassis log for `Looking for agent group` with your correlation id — its
  ABSENCE means the fallback ran; and read `collected_data-><step>->response`:
  an echo of the request body means no real workflow executed.
- **source:** bugfix_196 induction attempt 1, 2026-08-05 (NOTES in
  docs024_key_docs_latest/bugfix_196_failure_stamped_complete/)
- **added:** 2026-08-05, bugfix_196 lane

---

### `execution_metadata.completed_steps` is a COUNT, not a list of step names — and it is dead (0 on runs that are COMPLETED), so "did step X run?" answers a confident, permanent NO

- **footprint:** `orchestration_states.execution_metadata`, `completed_steps`, `total_steps`,
  `failed_steps`, `skipped_steps`, `retry_count`, `checkpoints`; any question of the form
  "did step X execute / how far did this run get"; any `@>` / `?` / `jsonb_path_exists`
  containment test against a jsonb field whose type you have not read back
- **fires when:** you want a per-step execution signal out of a run and reach for the
  field whose name says exactly that. The name is the trap: `completed_steps` reads as
  "the steps that completed".
- **the trap, and it is two stacked:** (1) **It is a NUMBER.** `jsonb_typeof` returns
  `number`. So a containment test — `execution_metadata->'completed_steps' @> to_jsonb(step_name)`
  — is comparing a number to a string. Postgres does **not** error on this; jsonb
  containment between mismatched types is simply **false**, for every row, always. The
  query cannot come out any other way. (2) **The counter is dead anyway.** Measured
  2026-08-06: **0 on all 23 runs whose `status` is `COMPLETED`**. Even read correctly as an
  integer it tells you nothing, so "fix the type and re-run" is not the remedy either.
- **the tell:** none, and the direction is what makes it dangerous. You get a clean `0`
  with no error, no NULL and no warning — and `0` is *exactly* what a genuinely
  unexercised path looks like. It therefore fails **toward** the absence you were
  testing for. I caught it only because the 0 contradicted a measurement already taken
  from another table (35 `page_components` rows inserted by those same 23 runs, ending
  2 seconds after the last one). Run the orchestration query first or alone and it is
  believed.
- **the check, before trusting ANY jsonb predicate:** read the type back, in one
  command, before the predicate — `SELECT jsonb_typeof(<expr>), <expr> FROM <table> LIMIT 3;`
  Type first, value second. For the actual question, **use `status`**, plus the step's own
  side effect in the table it writes (the durable witness — a status says the run ended, a
  row says the work happened).
- **the family it belongs to:** a jsonb type mismatch degrades to `false`, never to an
  error — the same silent shape as `jsonb::text LIKE '%"k":"v"%'` matching nothing (jsonb
  renders a space after the colon) and a jsonb PATH read not seeing the shape change
  underneath it. Three spellings, one failure: **the predicate did not look, and reported
  that it found nothing.**
- **source:** 2026-08-06, code-review triage lane, preparing the `CONTENT_DATA_REGRESSION`
  24h re-check (`code_review_triage_2026-08-05/` NOTES §14c, RUNBOOK R10; `WRONG_CALLS.md`)
- **added:** 2026-08-06, code_review_triage_2026-08-05 lane

### `revalidate_review_queue`'s `max_items` selects the OLDEST N across the WHOLE parked queue — and most of that queue is types it can never resolve, so a small N starves the types it CAN
- **footprint:** `platform/orchestration/actions/revalidate_review_queue_action.go` (`loadParkedReviewItems`, `reviewRevalidators`), `scheduled_tasks` row `review-queue-revalidate-daily`, `RevalidateReviewQueueInputSpec` (`max_items` default **50**), `site_work_items.status IN ('needs_human_review','unresolved')`
- **fires when:** you schedule the sweep, tune `max_items` down "to be gentle", or read a run's `scanned: 50` and conclude it looked at a representative 50
- **the tell:** **the sweep reports success every time and the same rows are re-swept for ever.** `loadParkedReviewItems` has NO item_type filter unless one is passed — it takes `ORDER BY created_at ASC LIMIT max_items` over **every** parked row. Measured 2026-08-06: **779 parked, but only 168 in the four types `reviewRevalidators` covers.** The other ~611 return `unknown` ("no revalidator for the type"), which is **not** a terminal state — they stay parked, stay oldest, and get re-selected on every subsequent run. So with the default `max_items = 50` the sweep re-judges the same uncovered head daily, closes nothing, reports `scanned: 50, resolved: 0`, and **never reaches the covered items it exists to drain**. Nothing errors and no counter looks wrong
- **the check:** compare the cap against the WHOLE parked queue, not the covered subset — ```sql SELECT count(*) FILTER (WHERE item_type IN ('required_fields_missing','needs_section_data','unresolved_cta','needs_page')) AS covered, count(*) AS total_parked FROM site_work_items WHERE status IN ('needs_human_review','unresolved'); ``` `max_items` must exceed **total_parked**, not `covered`. ⚠ **CORRECTED 2026-08-06, same day, by measuring the first scheduled run: YOU CANNOT SET IT FROM `scheduled_tasks.input_data`.** The action reads it from the STEP CONFIG (`datahelpers.GetIntField(config, "max_items", 50)`) and the `sweep` step has **no `input_mapping`**, so an `input_data` value is silently inert — I set 1000, the run reported `capped_at: 500`, which is the step config's value. **Two gates again, and the one you can reach is not the one that decides.** To change it, edit `agent_definitions` for `diagnosis-review-queue-revalidator`. Live measurement of that first run: `scanned 500, unknown 469, still_holds 31, resolved 0` — i.e. **94% of the swept batch was uncovered types**, and 279 of the 779 parked rows were never reached at all. The starvation is REAL AND PRESENT, not hypothetical and not fixed by the schedule. An `item_type` filter is the other honest way out
- **why it is not just "raise the number":** `unknown` is deliberately non-terminal (an ambiguity must stay queued for a human), so the uncovered head is *permanent* until each type gets a revalidator or the items are triaged. Any FIFO cap over a queue whose head never drains has this shape; the sweep is only the instance we hit
- ⚠ **UPDATED 2026-08-06 (later, same day) — the harm was WORSE than "never reached", and the fix named above ("an `item_type` filter is the other honest way out") HAS A TRAP OF ITS OWN.**
  - **What the original entry did not measure: how many of the unreached rows were rows the sweep could actually JUDGE.** Ranking the parked set by `created_at` and splitting on both axes at once gives it: **64 judgeable rows sat beyond the 500 head** — `required_fields_missing` 48, `needs_page` 8, `needs_section_data` 7, `unresolved_cta` 1, oldest filed 2026-07-24. That is **38% of the 168** the sweep exists to drain, and because 396 of the 500 head slots hold rows that can never leave, only ~104 slots ever turn over — so those 64 were **never** reached, not reached slowly. The query, which is the whole check: ```sql WITH parked AS (SELECT item_type, row_number() OVER (ORDER BY created_at ASC) AS rn FROM site_work_items WHERE status IN ('needs_human_review','unresolved')) SELECT (rn<=500) AS in_head, item_type IN ('required_fields_missing','needs_section_data','unresolved_cta','needs_page') AS covered, count(*) FROM parked GROUP BY 1,2; ``` **`total_parked` vs `covered` (the check above) tells you the cap is wrong; only this tells you whether it is currently costing you anything.**
  - **The starved rows are the NEWEST covered ones** — `rn > cap` under `ORDER BY created_at ASC` is by definition the young tail. This **inverts the selection's own stated rationale** ("the oldest items are the ones most likely to be describing a page state that no longer exists"): the finding filed last week is the one a recent re-render is likeliest to have already fixed, and it was the one guaranteed never to be looked at. If you tune a FIFO sweep, ask which end of the queue your cap is actually protecting.
  - ⚠ **AND: "one scheduled row per covered type" CANNOT WORK, for the same reason `max_items` could not.** `typeFilter, _ := config["item_type"].(string)` — the **step config**, one line below the `max_items` read this entry already warns about. With no `input_mapping` on the `sweep` step, *n* scheduled rows all read the one step config and run **identically**. It needs four near-duplicate agent definitions, or an `input_mapping` added first. **The trap in this very entry did not protect the fix recommended in the handoff two paragraphs after it** — `WRONG_CALLS.md` 2026-08-06.
  - **FIXED, both halves.** Stopgap **LIVE**: migration `323`, `max_items` 500→1500, commit `b14609e05` (config — live immediately, no roll; its guard asserts the cap still exceeds live `total_parked` **and** that no covered row remains beyond it, so a later apply into an outgrown queue fails loudly). Durable fix **committed, inert until a chassis roll**: `0e4e79124`, council `f64da546` — the selection filters to the types `reviewRevalidators` covers, **derived from the map** so registering a revalidator widens the selection in the same edit; the coverage gap becomes one `GROUP BY` over the whole parked set (the old shape reported the gap as *smallest* exactly when the backlog was worst); and `cap_binding` is logged at **WARN** when a pass fills its cap. **`capped_at` was always in the payload and it still took a fortnight to notice — a number in a blob nobody reads is not a signal.**
  - ✅ **BOTH HALVES NOW LIVE AND PROVEN — `v1.0.1257`, 2026-08-06 ~10:00Z, so THIS LANDMINE IS DISARMED for `revalidate_review_queue` specifically.** Pod-grep 0→1 on both replicas against a pre-roll baseline (positive control non-zero on both binaries, ruling out a broken probe). Proven by effect on the first sweep after it (`267fe850`): `scanned 168 · capped_at 1500 · cap_binding false · resolved 20 · unknown 112 · uncovered_backlog 611` — `scanned` is the **exact judgeable count**, not the 500 head and not the 779 parked, and **all 20 closed rows were created 2026-08-03..05, i.e. entirely inside the tail the old selection could never reach**, against **0** closed by the last pre-fix run. Fleet `auto:revalidated` 34 → 54.
  - **What still generalises, and why this entry stays:** the SHAPE is not specific to this action. **Any FIFO `LIMIT` over a queue whose head contains rows that can never leave starves the tail, reports success every pass, and shows no wrong counter.** `unknown`/`skipped`/`deferred` verdicts that are deliberately non-terminal are the usual cause. Read the entry for the shape; the specific instance is fixed. **The residual coverage gap is unchanged: ~611 parked rows are in types nothing revalidates** — now reported in full every run (`uncovered_backlog`), which the old code was structurally incapable of doing, since it could only count the unjudgeable rows that fell *inside* the cap.
  - ⚠ **A corollary worth its own check, learned closing this:** the same starvation makes a row look **individually skipped**. `HANDOFF_2026-08-04` §0b flagged a `needs_page` row with no `result.revalidation` stamp while its siblings had one, and suspected an `item_key` prefix drift. **Refuted** — it was never *loaded*, and the siblings it was compared against were two weeks older, so the comparison was an age comparison wearing a predicate's clothes. **From a single row, "skipped" and "never loaded" are indistinguishable; check reachability before theorising about a predicate.**
- **source:** `bugfix_168_deployed_asset_path` lane, 2026-08-06, sizing the owner-approved schedule; updated same day by the next session that tried to act on it
- **added:** 2026-08-06, bugfix_168 lane

### `toolgolden` CANNOT certify a ratio-only calculator — its own inert-tool guard refuses the page, and the arithmetic is fine
- **footprint:** `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/toolgolden.py` (`VECTORS`, the `INPUT-INDEPENDENT` refusal), any tool whose outputs are RATIOS of its inputs (yield, LTV, percentage-of, ratio-to-income)
- **fires when:** you capture a golden over a set of calculators and one refuses with *"reacts, but output is identical for every input value — arithmetic ignores its inputs"*
- **the tell:** the refusal names the page, so it reads as a broken tool — and the refusal text itself proposes "the tool is genuinely inert" as a live option. It is a property of the VECTOR SCHEME: toolgolden scales **every** numeric field by the **same** factor (x1, x2, x0.5) deliberately, so values stay in-domain for any tool with no per-tool config. **A ratio is invariant under uniform scaling.** Doubling price and rent together leaves gross yield exactly where it was, so the harness cannot distinguish a correct ratio calculator from a dead one. Measured 2026-08-05 on `loanandmortgagecalculator.co.uk/mortgages/investor.html`: 1 of 23 refused; reading the two functions showed `rent*12/price` and `loan/price`, both correct, and hand-driving reproduced 5.76% on the shipped defaults
- **the check:** before believing the page, **stagger the vectors — move ONE field at a time** and require the outputs to change. Rent x2 → yield doubles; loan x2 → LTV doubles and crosses the commentary branch. Worked implementation: `loanandmortgagecalculator_couk/investor_golden.py` (it also inherits toolgolden's `settle()` and storage-clearing reload, both of which exist for their own reasons). **And keep the refusal**: a staggered capture that STILL shows no variation is the real inert-tool finding
- **the family it belongs to:** the instrument answers the question you encoded. A vector scheme chosen to be safe for every tool is, for one class of tool, a scheme under which no wrong answer can appear
- **source:** 2026-08-05, loanandmortgagecalculator voice-rebuild lane, baselining 23 calculators before decomposition
- **added:** 2026-08-06, loanandmortgagecalculator_couk lane

### psql's DEFAULT `|` SEPARATOR SPLITS INSIDE the data on any site whose titles carry " | Brand" — and the parse looks exactly like truncation
- **footprint:** `psql -tA` with no `-F`, `pages.title`, `sites.domain`-suffixed titles, any script parsing `line.split("|")`
- **fires when:** you read a text column that can legitimately contain a pipe — page titles are the common one, because " | SiteName" is the standard title suffix across this estate
- **the tell:** **there is none in the psql output**, and the downstream symptom accuses the wrong system. A `--check` comparing manifest titles against `pages.title` reported *every* title as truncated at the pipe, which reads precisely like "adoption stored a truncated title" — a plausible, filable platform bug. The truncation was in my own `split("|")`, one line later. Re-measured with `-F '\t'`: **0 of 41 titles or descriptions differ.** Note the direction: extra fields are silently dropped by an index-based parse, so the row still parses and still looks well-formed
- **the check:** pass an explicit separator that cannot occur in the data — `psql -tA -F '\t'` — whenever a selected column is free text. If you must keep `|`, assert the field count per row (`len(parts) == n`) and fail loudly rather than indexing; a row that splits into 6 fields when you expected 5 is the whole signal, and index access throws it away
- **source:** 2026-08-05, loanandmortgagecalculator voice-rebuild lane (`load_lmc.py` page-identity join)
- **added:** 2026-08-06, loanandmortgagecalculator_couk lane

### `DISTINCT ON (agent_type, step_name)` over `llm_call_log` RESURRECTS retired configs — the 2026-07-26 relabel gives every step a second, permanently-stale "latest row"
- **footprint:** `llm_call_log`, `agent_type`, any `DISTINCT ON`/`ORDER BY created_at DESC LIMIT 1` used to answer "what is this step's CURRENT model/cap/temperature", any window longer than ~2 weeks
- **fires when:** you derive a step's live configuration from its most recent call rather than from `agent_definitions` — a legitimate and often BETTER technique here, because it sees what the code actually resolved through the fallback chain (LCO-002) and cannot be fooled by root-block shadowing (MDL-039 / `bugs_open/009`)
- **the tell:** **there is no error and the row looks current.** `agent_type` recorded `generic` for every chassis call until 2026-07-26 ~14:54 and the resolved type after. So a per-agent "latest" produces TWO groups for one step — the real one, and a `generic` group frozen at 07-26 whose latest row is however many months old. Both are returned, both look like current config, and the stale one carries whatever cap was live *then*. Measured 2026-08-06 building `fleet-step-token-pressure`: `classify_and_extract@6000` — a cap retired on 08-02 and raised twice since — came back as a CURRENT pair in the live run, dragging its entire pre-raise truncation history with it and flagging a step that is now running at 13% of cap
- **the check:** key the "latest" on **`step_name` alone** (or on a column that did not get relabelled), and treat `agent_type` as a display label, never as part of a grouping key. Then verify with the case you expect to be **ABSENT** — a cap you know was raised must NOT appear. That negative case is what exposed this; every positive result looked perfect. If you genuinely need per-agent resolution, bound the window to start after 2026-07-26 and say so, accepting that you cannot see further back
- **the family it belongs to:** a column whose MEANING changed mid-history partitions your data along a line that is invisible in every result set. Same root as the two `llm_call_log` traps already in this file (`output_tokens` NULL on truncation; per-agent filters silently dropping the pre-relabel era)
- **source:** 2026-08-06, `bugfix_183_step_token_pressure` lane; NOTES misstep 3
- **added:** 2026-08-06, bugfix_183 lane

### A truncation COUNT does not tell you the SHAPE — the fleet's worst-looking cap problem was two records on an infinite retry loop
- **footprint:** `llm_call_log`, `doc_notes` rows from `fleet-step-token-pressure` / `council-seat-token-pressure`, any decision to raise a `max_tokens` cap
- **fires when:** a headroom or truncation report flags a step and you reach for the obvious remedy — raise the cap
- **the tell:** the number is real and the diagnosis it suggests is wrong. `extract_and_reconcile@2048` topped the first fleet run with **64 truncations**, the largest in the estate, reading as the worst-sized cap we have. Grouping the same population by `md5(prompt_rendered)` showed **two byte-identical prompts** accounting for all 64 — 46 of one, 18 of another, **zero successes on either** — while every other prompt in the window succeeded at a quarter of the cap. Distinct `correlation_id` per call, so these are fresh dispatches, not one orchestration's retries: a 5-minute batch sweep re-selecting records whose verification can never complete (`bugs_open/205`)
- **the check:** before acting on any truncation flag, run `SELECT md5(prompt_rendered), count(*), count(*) FILTER (WHERE success) FROM llm_call_log WHERE step_name='X' AND created_at > … GROUP BY 1 ORDER BY 2 DESC`. **Many distinct prompts near the cap = genuine cap drift, raise it. One or two prompts repeating = a stuck item, and raising the cap treats the symptom** while the loop keeps burning credits. The two cases produce the same headline number and want opposite fixes
- **source:** 2026-08-06, `bugfix_183_step_token_pressure` lane, first live run of the new check
- **added:** 2026-08-06, bugfix_183 lane

### `--color-accent-text` is promised by a code comment, derived by the platform, and resolves to NOTHING on every site
- **footprint:** `--color-accent-text`, `platform/orchestration/actions/palette_specialised_slots.go` (`darkSchemeDerivations`, the `accent_text` entry at ~:112), `layouts.css_template`, any component template hard-coding `color: white` / `color: #fff` over an accent fill
- **fires when:** you fix a white-ink-on-a-themed-fill contrast defect (`features_open/026` family 3, `bugs_open/122` sub-shape C) the way the platform's own comment tells you to — by pointing the component at `var(--color-accent-text)`. The comment reads: *"accent_text has no layout consuming it yet. **It is emitted** so a component can stop hard-coding white over an accent fill: `.form-submit` does exactly that on 12 sites and scores 2.81:1"*. It says **is emitted**, present tense, and it is wrong
- **the tell:** none in the source, and the *page gets worse rather than better*. The palette genuinely gains the slot, `fillDarkSchemeSpecialisedSlots` genuinely logs it as derived, and the variable is never written into any stylesheet — so `color: var(--color-accent-text)` with no fallback resolves to an empty value and the browser falls back to the inherited colour, which on an accent-filled button is usually *less* legible than the white you removed. A `grep` of the Go source confirms the derivation exists and tells you nothing about whether it ships
- **the check:** two queries, both cheap, before you consume ANY `--color-*` variable you have not personally seen in served CSS:
  ```sql
  SELECT count(*) FROM layouts WHERE css_template LIKE '%palette "accent_text"%';        -- 0 of 18
  SELECT count(*) FROM site_components WHERE rendered_html LIKE '%--color-accent-text:%'; -- 0
  ```
  Measured 2026-08-06. Compare against `primary_text`, which is declared by **18 of 18** layouts and does land — so the two siblings, adjacent lines in the same derivation list, behave oppositely. **A palette reaches the stylesheet ONLY through `{{palette "X" "literal"}}` in a layout template**, so the derivation list is necessary and nowhere near sufficient. And regardless: **always write the fallback** — `var(--color-accent-text, #fff)` — so the un-emitted case degrades to today's behaviour instead of to nothing
- **the relationship to the existing entry:** this is the CONSUMING half of *"A palette slot no LAYOUT declares is never emitted — deriving it is dead config"*, which fires from the DERIVING side ("fires when: … adding the offending slot to the derivation list"). That one cannot save a session that never touches the derivation list and is simply trusting a documented variable. Same root, opposite approach, and the consumer is the likelier visitor — the comment actively recruits them
- **source:** measured 2026-08-06, `bugfix_122_contrast_ink_slots` lane; `bugs_open/122` appended §2026-08-06; the sibling entry above from `dartsonline_traffic` 2026-07-29
- **added:** 2026-08-06, bugfix_122 lane

### A discovery check's NAME is not its `item_type`, and querying by the name returns 0 rows — which reads exactly like "this check has never fired"
- **footprint:** `site_work_items.item_type`, `platform/orchestration/actions/discovery_checks/*.go` (`ItemType:` literals), `agent_definitions … run_checks.config.checks`, any "has this check ever fired / is it inert?" question
- **fires when:** you size a check's real-world output before extending or replacing it. The check registers as `phantom_internal_links` (**plural**); the item type it files is `phantom_internal_link` (**singular**) — and that one check files **three** types (`phantom_internal_link`, `empty_internal_href`, `unbuilt_internal_link`), so no single name could serve as both. `SELECT … WHERE item_type='phantom_internal_links'` returns **0 rows and no error**. The true answer on 2026-08-06 was **119 items, 55 complete, newest 08-04**. The zero does not look like a typo: it looks like a finding about the mechanism, and it is the exact premise that justifies building a second, separate check — which is how `bugs_open/093`'s fix reached production correct and never executed
- **the check:** take the spelling from the producing check's SOURCE, never from the check name or the `checks` array — `grep -n 'ItemType:' platform/orchestration/actions/discovery_checks/<check>.go`. From the DB side, never conclude "never fired" from a zero on a name you typed: `SELECT DISTINCT item_type FROM site_work_items WHERE item_type LIKE '%phantom%'` costs the same query and cannot miss by a plural. Same family as *a grep proves absence only for the SPELLING it searches*, made worse because the wrong spelling is the name of the very thing you are asking about
- **and the same error can already be sitting in the concept register:** LNK-009's status line read *"NOT YET ENABLED (deliberate, observe-only later)"* while the check had been enabled and filing for weeks, and council seats read those status lines as ground truth. Corrected in `af2667453`; re-derive with `SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;`
- **source:** 2026-08-06, `bugfix_071_fragment_blindspot` lane, caught while reading the check's source to write a new arm into it; logged in `WRONG_CALLS.md` the same day

### Registering an `ItemVerifier` obliges TWO more edits, and the build tells you about only one of them at a time
- **footprint:** `discovery_checks/verifiers.go` (`RegisterVerifier`), `verifier_coverage_test.go` (`TestRegisteredVerifiersMatchClaimTimeoutExclusion`), `docs/agent_docs/sql_for_agents/220_claimed_item_timeout_generic_evidence.sql`, `scheduled_tasks.pre_query` for `claimed-item-timeout`
- **fires when:** you add a discovery check with a verifier, having reasonably assumed that registering it is what makes it gate. It is not: the `claimed-item-timeout` sweep auto-completes a stuck item at 15 minutes on the handler orchestration's own evidence, walking straight past the verifier — unless the item type is in that sweep's `item_type NOT IN (...)` list. **The lockstep test fails the moment you register**, naming the obligation, which is the designed catch and works. The half it CANNOT check is that the file it reads (`220`) is only the DECLARED list: **the LIVE `pre_query` column is a separate edit**, and a lane has already declared an entry in `220` that never reached the live column, leaving their verifier bypassable for two days (`305`'s header records it)
- **the check:** read the live column BEFORE writing the replace, because a `replace()` must name the exact current string — if it has drifted from `220`, applying only your entry re-encodes someone else's gap into the new string: `SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)') FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';`. Then assert BOTH directions in `DO`/`RAISE` blocks — new list present (1 row) **and old list consumed** (0 rows); a verify block of plain `SELECT`s cannot stop a `COMMIT`. Precedent: `269`, then `305`, then `322`
- **source:** 2026-08-06, `bugfix_071_fragment_blindspot` lane (`dead_fragment_link`); the live list was read first and matched `220` exactly, so `322` carried nobody else's entry

### `075_trigger_discovery.sh` ends by TRIAGING every detected work item on a hardcoded `finetuning.uk` — whatever domain you passed it

- **footprint:** `scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh`, `site_work_items.status`, `finetuning.uk`, and by family `scripts/initial_messages/**/07*_trigger_*.sh`
- **fires when:** you reach for the obvious committed script to fire a discovery sweep at a site. It takes `<domain>` as `$1`, prints your domain, and publishes a correct kcat message for it — so the part you watch is right. **Two blocks after the publish, past the "CORRELATION_ID=" echo where reading usually stops, it runs against a different site:**
```sql
UPDATE site_work_items SET status = 'triaged', updated_at = NOW()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk') AND status = 'detected';
```
  Unconditional, hardcoded, no reference to `$DOMAIN`. The `# See what was found` SELECT above it is hardcoded the same way.
- **why it is worse than a stray query:** `detected` is the status that keeps a finding INERT — `load_work_items` only loads `triaged`/`approved` (`load_work_item_actions.go:633`). Flipping a site's whole detected backlog to `triaged` is precisely the act that makes it **dispatchable**, so the next `build-dispatch-loop` pass can start real LLM rebuilds on someone else's customer site. And a page-build repair **regenerates a section rather than editing it** (`LANDMINES.md`, `spec.mode="recreate"`), so the damage is replaced prose, not a retry.
- **the tell:** none at run time — it prints your domain throughout and exits 0. The tell is afterwards and on a site you were not looking at: a batch of `finetuning.uk` items moving `detected → triaged` with `updated_at` all equal to your run.
- **the check:** **read to the END of any `initial_messages` trigger before running it — the hazard is below the echo that looks like the end.** `grep -n "UPDATE\|DELETE\|finetuning\|hardcode" <script>` before the first run. If you only need the dispatch, copy the kcat envelope into your own script and leave the tail behind (that is what the `201` lane did on 2026-08-06 to fire `quality-discovery-agent`, which this script also cannot do — its `case "$2"` accepts only `design|completeness` and exits 1 on anything else).
- **not an isolated file:** `075d_simple_maintain_trigger.sh` in the same directory **cannot execute at all** — line 9 is a bare `-------------------` under its own `set -euo pipefail`, and line 11 hardcodes `DOMAIN="finetuning.uk"` over the argument line 7 demands (committed that way in `5345ad7e2`). **Treat this directory as unreviewed, not as tooling.**
- **source:** 2026-08-06, `bugfix_201_page_content_writer_dispatch` lane, found by reading the script before running it while firing a sweep at `gaswholesalers.com`; `bugfix_194` lane found the `075d` half on 2026-08-05
- **added:** 2026-08-06, bugfix_201_page_content_writer_dispatch lane

---

## `WHERE domain IS NULL` on `agent_error_log` sees 1.3% of the rows that have no domain — three of the twenty writers store `''`
- **footprint:** `agent_error_log.domain`; `count(domain)`; `domain IS NULL`;
  `domain IS NOT NULL`; `platform/orchestration/agenterrors/agenterrors.go`;
  `platform/orchestration/actions/store_generated_component_action.go`;
  `platform/orchestration/actions/component_write_guard.go`;
  `platform/orchestration/actions/diagnose_load_runtime_action.go`;
  `agenterrors.Write`; `orchestration.LogAgentError`
- **which is which:** the first three files are the clone writers that store `''`; the
  fourth is the reader that filters `domain = $2::text` and so silently drops them.
  (Footprints are deliberately bare tokens with no prose and no bracketed asides —
  `landmines-sync.py` splits this line on `,` and `;`, so
  "`the three clone writers x.go (Write, formerly y)`" parses into three junk keys and
  loses the path. My first version of this entry did exactly that. Check yours with
  `SELECT subject_key FROM doc_notes WHERE subject_type='landmine' AND body LIKE '%<phrase>%'`.)
- **fires when:** you count, filter or group `agent_error_log` by `domain` — including
  "how many rows are unattributed?", any per-domain error dashboard, and any
  site-scoped or domain-scoped diagnosis load. No symptom is needed and none appears.
- **why the wrong answer looks right:** the query is well-formed, returns promptly and
  returns a plausible small number. **`domain IS NULL` matched 128 rows of the 10,077
  that have no domain** [MEASURED 2026-08-06 11:2xZ — **the retained window, NOT all
  history**: this table is reaped by `scheduled_tasks.database-cleanup` at 14 days
  resolved / 30 days unresolved, so no census of it is ever "all time". 128 NULL,
  **9,949 `''`**, 4,688 real, 14,765 total] — an under-report of **79×**, and `count(domain)`
  is worse because it counts `''` as present. The ratio is **not stable**: it was 26×
  on 2026-08-05, because the writers producing `''` are the high-volume ones. Do not
  quote a factor without its date.
- **⚠ THE CENSUS BELOW WAS SUPERSEDED WITHIN A DAY — re-measure before quoting it.**
  `f930de86b` (2026-08-07, RFC_012 B) retired **18 hand-copied INSERTs into one writer**, so
  the 20 sites are now **3**: `agenterrors/agenterrors.go:89`,
  `store_generated_component_action.go:1353`, `internal/agents/contentcreator/claims_guard.go:184`.
  Files listed below as `NULLIF` writers now build an `agenterrors.Entry` and call
  `LogActionEntry` (`actions/log_action_error.go`), **which never touches `entry.Domain`** and
  forwards to the bare-`$2` INSERT. **Net effect: the convention was consolidated ONTO the broken
  shape**, so `''` is now fleet-wide by construction — measured 08-08, `domain IS NULL` returns
  **0 on every error_code** written since the consolidation, where the guard codes used to produce
  NULL. **The upside is bigger than the downside: the fix is now ONE LINE** — `NULLIF($2,'')` at
  `agenterrors.go:94` covers every consolidated caller (plus one line in
  `store_generated_component_action.go`; `contentcreator/claims_guard.go` omits the column and
  already writes NULL).
- **the mechanism as it stood 2026-08-06, kept because it is what the shapes below still mean:**
  of **20** non-test
  writers, 10 write `NULLIF($n,'')`, 7 omit the column (both → NULL), and **3 write a
  bare `$2`** (→ `''`). Those three share a byte-identical 13-column `VALUES` block —
  clone-and-drift, not a missing convention. One of the three is the coordinator's
  **generic** writer, so the classifier-generated codes (`LLM_API_ERROR`,
  `PROCESSING_FAILED`, `TIMEOUT`, `UNKNOWN`, `PARSE_ERROR`,
  `CHILD_ORCHESTRATION_FAILED`, `VALIDATION_ERROR_DROPPED`, `INCOMING_MESSAGE_REJECTED`,
  `UNROUTED_IMAGE_KIND`) are `''` and carry **zero** NULLs, while every guard's own code
  is NULL or a real domain. The partition is exact — no code appears in both sets.
- **the check:** never ask for NULL; ask for either shape, and show all three so a
  future divergence cannot hide —
  ```sql
  SELECT count(*) FILTER (WHERE domain IS NULL)  AS is_null,
         count(*) FILTER (WHERE domain = '')     AS empty_str,
         count(*) FILTER (WHERE domain <> '')    AS real_domain
  FROM agent_error_log;
  -- "no domain" is COALESCE(domain,'') = '' — never `domain IS NULL`, never count(domain)
  ```
  And before citing any file:line for this table's writers, **re-locate them**:
  `grep -rn "INSERT INTO agent_error_log" --include=*.go platform/ internal/ | grep -v _test.go`.
  RFC_012 (owner ruling 2026-08-06) moved the generic INSERT into a leaf package **and
  carried the defect over verbatim** — a refactor is not a review of what it moves, and
  every published file:line for this writer went stale the same day.
- **two adjacent traps worth knowing while you are here:** (1) `bugs_closed/034` closed
  partly on the claim of "**One** `agent_error_log` writer" — true for the two sites it
  was about, and there are now twenty, so do not treat that consolidation as current.
  (2) The `NULLIF($n,'')::uuid` on `site_id`/`work_item_id` is **compelled** by the cast
  (`SELECT ''::uuid` → `ERROR: invalid input syntax for type uuid: ""`), so its presence
  beside a bare `domain` is *not* evidence anyone intended the difference — reasoning from
  that asymmetry sends you at the wrong target. Complements the "save seam's identity
  paths" entry above, which is about *writing* an attributable row; this one is about
  *reading* the column afterwards.
- **source:** 2026-08-06, code-review triage lane (`code_review_triage_2026-08-05/`
  NOTES §16, RUNBOOK R13); the over-read that produced it is in `WRONG_CALLS.md`.
  **NOT fixed** — a stored-shape change on the fleet's highest-volume error writer,
  wants a council round; `090` has not been run on it.

### The owned-page guard belongs on `assemble_page`, NOT on `git_commit` — the wider net breaks the only paths owned pages deploy by

- **footprint:** `platform/orchestration/actions/owned_page_guard.go`, `multipage_actions.go` (`AssemblePageAction`), `get_pages_to_build_actions.go` (`queryPagesForBuild`, `ownedPageExclusionSQL`, config key `include_owned`), `git_deployer_actions.go` (`GitCommitAction`, `checkUpstreamSkipped`), `save_page_sections_action.go` (the ownership refusal), `v3_site_actions.go` (`UpdatePageStatusAction`), the `pages.rebuild_policy` column, agents `page-rebuild` / `pageflow-builder` / `site-work-orchestrator` / `page-rerender` / `section-editor`
- **fires when:** you review or extend the `bugs_open/208` guard and notice it is enforced in two places, one of them an assembly step. The obvious tidy-up is to move it down to `git_commit`, which is where the damage actually happens and which would cover every pipeline at once. **Do not.** `page-rerender` (`rerender_single_page`) and `section-editor` (`apply_section_edit`) also `git_commit` pages, and those are precisely how an owned page legitimately reaches the web — migration 164 says it in terms: *"page_rerender / assemble (re-assembly of EXISTING page_components) is deliberately NOT gated — it is how owned pages deploy."* A guard on `git_commit` would stop tool pages deploying at all, and the failure would look like "tools mysteriously stopped updating", not like a guard misfiring
- **the tell that makes `assemble_page` the right seam, and it is a measurement, not a judgement:** `assemble_page` has **exactly three live consumers** — `page-rebuild`, `pageflow-builder`, `site-work-orchestrator` — and they are **exactly** the three agents whose step order is `assemble_page → deploy_page (git_commit) → save_sections`, i.e. commit-before-guard. All three feed it `content_field: page_content.response.page_html` (freshly LLM-written HTML, never re-assembled components). Re-measure with the nested walk before trusting this, because a top-level `jsonb_each` misses steps inside a loop `sub_workflow`: `FROM agent_definitions ad, LATERAL jsonb_path_query(ad.default_config,'$.**.steps') AS steps, LATERAL jsonb_each(steps) AS s(key,value) WHERE s.value->>'action'='assemble_page'`
- **the second trap, which looks like the same defect and is not:** `needs_page → page-build-handler` is SAFE and needs no guard. Its live order is `save_sections → update_status → deploy_page`, so the ownership refusal fires before anything is committed — and all **158** `needs_page` rows fleet-wide (all history, measured 2026-08-06) route there. If you "fix" that path too you are adding a refusal where one already works. The live exposure of the work-item loop comes from a different door: discovery-check fix items (`literal_markdown`, `placeholder_contact`) with `handler_agent='page-content-writer'`, of which **11 of 14 targeted owned pages** on 2026-08-04
- **the check before changing the refusal's SHAPE:** it must stay a `{skipped:true, skip_reason}` return, never an error. `continue_on_error` is unset on all four build loops and `shouldContinueLoopOnError` (`loop_error_handler.go:70-90`) requires it, so an error inside an iteration fails the whole workflow — and pages are selected `ORDER BY nav_order, name`, so one owned page strands every page after it. That was the pre-fix behaviour of the `save_page_sections` refusal, and it is why the fix threads the skip through `save_page_sections` and `update_page_status` as well
- **and the one that will bite whoever writes the next test:** asserting `res["skipped"] == true` does **not** prove the skip guard works. `SavePageSectionsAction` has a pre-existing "page not found, skipping" branch returning the identical shape, so a test written that way passes with the guard **deleted** (mutation M3, 2026-08-06, `WRONG_CALLS.md`). Assert the discriminator — the `reason` string, which carries `OWNED_PAGE_GUARD` only on the intended path
- **source:** 2026-08-06, `bugs_open/208` (`docs024_key_docs_latest/bugfix_208_owned_page_commit_before_guard/`). Same class as `bugs_closed/143`'s asset-lock entry above — a DB-row guard sitting behind a git commit — but the resolution is the opposite shape: 143 moved the check EARLIER in the same action, 208 could not, because the committing action is shared with the paths that must stay ungated
- **added:** 2026-08-06, bugfix_208 lane

### `--color-<x>-text` and `--color-<x>-ink` are OPPOSITE questions — reaching for the wrong one fixes nothing and looks like a fix

- **footprint:** `platform/orchestration/actions/palette_specialised_slots.go` (`legibleInkFor`, `buildLegibleInkDefaults`, `darkSchemeDerivations`, `pickInkOn`), `platform/orchestration/actions/render_css_from_spec_action.go` step 12, the CSS custom properties `--color-accent-text` / `--color-primary-ink` / `--color-accent-ink`, `content_components.html_template`, `layouts.css_template`
- **fires when:** you are repointing a component whose text fails contrast and you reach for the derived slot whose name looks right. The two directions are **not** interchangeable and the platform now emits both:
  - `--color-<x>-text` — the ink that goes **ON an `<x>` fill**. Use when the element's *background* is `var(--color-<x>)`.
  - `--color-<x>-ink` — `<x>` **itself made legible as an ink**. Use when the element's *colour* is `var(--color-<x>)` and it lands on the page or a card.
  Pick the wrong one and the declaration is valid, the build is green, the variable resolves, and the contrast is unchanged.
- **the worked case, because it is what proved the trap is real:** a council round for `bugs_open/122` described three failures as one shape ("a component hard-coding an ink over a themed fill") and proposed making `--color-accent-text` reachable. **Not one of the three hard-codes anything, and only one of them wanted that slot.** finetuning.uk `.csg-cta-btn` fills with `var(--color-accent)` (#C8873A) and inks with `var(--color-primary-text)` — the ink for a *primary* fill — at 3.01:1: that one wants `accent-text`. gaswholesalers.com's base `a { color: var(--color-accent) }` at 2.22:1 and gamesdesign.co.uk `.stats-eyebrow` at 1.44:1 use accent **as** an ink: `accent-text` is the wrong direction and would have changed nothing on either site. The `editquality` seat caught it as a HIGH objection ("ships no edit that actually consumes it for any of the cited failures").
- **the check, and it is two lines of the template, not a theory:** read the rule and ask which property names the palette colour.
  `SELECT substring(html_template from '\.<selector>\s*\{[^}]*\}') FROM content_components WHERE name='<component>';`
  `background:` names it → you want `--color-<x>-text`. `color:` names it → you want `--color-<x>-ink`. If both appear in one rule, they are different colours and you need both.
- **the second trap, which hides the first:** a `var(--color-primary-text, #fff)` that renders white does **not** mean the fallback fired. Check the served stylesheet before concluding the literal is the culprit — on finetuning `--color-primary-text` **is** defined, as `#ffffff`, and it is *correct for its own slot* (primary is #1A1A2E). The value is right, the slot is wrong, and a grep for hard-coded whites finds nothing.
  `curl -fsS -A "Mozilla/5.0 … Chrome/126" https://<domain>/assets/css/styles.css | grep -- '--color-primary-text:'`
  (a bare curl gets 403 on every site — that is a user-agent rejection, not a routing fault)
- **and the constraint that will be lost first:** `legibleInkFor` takes a **slice** of grounds and requires the candidate to clear AA against **every** one, because a component may place the same ink on the page and on a card (dartsonline: `__eyebrow` 1.04 on `background`, `.tl-card-link` 1.07 on the derived `card_bg`). Simplifying it to one ground is a silent half-regression that reads as a working fix on whichever page you happen to open. `TestLegibleInkFor_TwoGroundsDisagree` is the only guard, and it is the only test that fails under that mutation.
- **while you are in there:** the fixture for such a test must be **satisfiable**. A first version used grounds `#101010` and `#E9E9E9`, for which AA against both is arithmetically impossible (the darker demands relative luminance ≥ 0.200, the lighter ≤ 0.140), so every candidate correctly fell through to the achromatic fallback and the test failed while the code was right. A trap no value can escape does not test preference — it tests the fallback.
- **source:** 2026-08-06, `bugfix_122_contrast_ink_slots` lane; council `c4d9c841-3658-4742-85b5-961e062ecad2` (round 1 REVISE on exactly this, round 2 APPROVED). Register entry VIZ-014.
- **added:** 2026-08-06, bugfix_122 lane

### `collection_tasks.retry_count` is written ONLY by the reaper's pre_query — a repo grep says the column is dead, and the parking behaviour it drives is invisible in Go

- **footprint:** `business_intel.collection_tasks` (`retry_count`, `status='failed'`, `scheduled_for`), `scheduled_tasks` row `stale-orchestration-reaper` (`pre_query`, `reset_tasks` CTE), `platform/orchestration/actions/business_intel_actions.go` (`LoadBusinessBatchAction`), `platform/orchestration/actions/ensure_collection_tasks.go`, `vet-batch-verify` / `vet-batch-processor` / `vet-practice-verifier`
- **fires when:** you read the vet-collection Go code and conclude (a) `retry_count` is never incremented so it can be repurposed or dropped, (b) a task failure ends the matter because nothing re-queues it, or (c) `status='failed'` rows are written by some failure handler you should go looking for. All three are wrong the same way: **the lifecycle's third writer is SQL inside the reaper's `pre_query`** (every 180 s), which no repo grep can see — no seed file defines that row, the live row is the source. Since 2026-08-07 it counts each stale-claim reset in `retry_count`, backs off re-eligibility via `scheduled_for`, and PARKS the task as `'failed'` on what would be the 5th reset (`bugs_open/205`: unconditional resurrection burned 1,575 failed dispatches/day across 33 doomed tasks, invisible until one of them started buying LLM calls).
- **the check:** before reasoning about this table's lifecycle, read the live row: `SELECT pre_query FROM scheduled_tasks WHERE name='stale-orchestration-reaper';` — and treat "quiet vet pipeline" as TWO hypotheses, parked-vs-dead: `SELECT status, count(*) FROM business_intel.collection_tasks GROUP BY 1;` distinguishes them (parked = rows at `'failed'` with `error_message` naming 205; dead = `pending` piling up with `last_triggered_at` stale on `vet-batch-verify`). Un-parking is a deliberate operator UPDATE (RUNBOOK in `bugfix_205_poison_pill_reaper/`), and `ensure_collection_tasks` refuses to re-task a business whose task is `'failed'` — recreating the task would silently zero the counter.
- **source:** 2026-08-07, bugfix_205 lane (`bugs_open/205`, `docs024_key_docs_latest/bugfix_205_poison_pill_reaper/`). Same class as the `stale-work-item-reaper` row-age entry (016b §9, 2026-07-25): a reaper that does not annotate its own work.
- **UPDATED 2026-08-08 (same lane, owner decisions 3+4):** the logic moved again, one hop further from any grep of the pre_query — the reaper's CTE is now `SELECT * FROM business_intel.reap_stale_collection_tasks()`, the numbers live in the **`reaper_policies` table** (per-task_type `park_after`/`backoff_minutes`/`stale_after_minutes`; undeclared types default 5/20m/20m), and a task type declares its own ceiling by INSERT. So the check grows one step: after reading the pre_query, read the FUNCTION (`\sf business_intel.reap_stale_collection_tasks`) and `SELECT * FROM reaper_policies;`. Migration `sql_for_agents/335` (+ROLLBACK); RFC_018.
- **added:** 2026-08-07, bugfix_205 lane

---
> **CORRECTED 2026-08-07 (RFC_012 option B, `f930de86b`) — the headline count is now WRONG in
> the direction that makes this trap WORSE, and the title is left unchanged only because it is
> the `landmines-sync.py` key.** "Three of the twenty writers store `''`" was true when this was
> filed. Since `f930de86b`, **every** converted writer in `platform/orchestration/actions/` goes
> through `agenterrors.Write`, which passes `domain` as `$2` with no `NULLIF` — so nine sites
> that used to store NULL now store `''` too, and the only writer still hand-copying its INSERT
> (`store_generated_component_action.go:1353`) already stored `''`. **`domain IS NULL` is
> therefore an even smaller and even less representative slice than the 1.3% measured here.**
> Re-measure before quoting any figure on this line; do not carry the 1.3%.
>
> Recorded because the RFC_012 lane made exactly the mistake this entry exists to prevent: it
> argued "NULL→`''` is measured inert" from ONE reader without grepping this file first (the
> `SessionStart` hook only surfaces landmines for files already DIRTY, and that lane was
> *calling* `agenterrors.go`, not editing it). The council gate caught it, not the lane. See
> `WRONG_CALLS.md` 2026-08-07 and register entry **RSH-008**, whose stated open review question
> — *should the shared writer `NULLIF` `domain`?* — is the real fix and is still open.
>
> **CORRECTED AGAIN 2026-08-07 (morning), and this reverses the sentence immediately above:
> the `NULLIF` is NOT the real fix — adding it would be a REGRESSION. Question CLOSED, measured.**
> `f930de86b` went live on v1.0.1262 at 05:47Z and the table has **converged on `''`**:
> ```sql
> SELECT CASE WHEN occurred_at >= '2026-08-07T05:47:39Z' THEN 'post-roll' ELSE 'pre-roll' END,
>        count(*) FILTER (WHERE domain IS NULL), count(*) FILTER (WHERE domain = ''),
>        count(*) FILTER (WHERE domain <> ''), count(*)
> FROM agent_error_log GROUP BY 1;
> -- post-roll:   0 NULL | 29 '' | 16 real |    45      [MEASURED 2026-08-07]
> -- pre-roll:  128 NULL | 13,885 '' | 4,762 real | 18,775
> ```
> **Zero NULL rows since the roll.** And every one of the 128 NULL rows was written by a site
> this conversion changed — `GROUP BY agent_type, action` over `domain IS NULL` returns nine
> groups, all of them converted sites, newest `2026-08-05 20:14Z`, i.e. all pre-roll. **The NULL
> bucket is closed and the 14/30-day reaper will empty it.** Adding a `NULLIF` now would send
> 100% of new rows into the shape **0.9%** of rows use, re-splitting a table that has just
> converged, and stranding 13,885 `''` rows against a `domain IS NULL` query that would newly
> appear to work. The remedy this entry already prescribes — `COALESCE(domain,'') = ''` — is
> unaffected and remains correct for both eras.
>
> **The "be consistent with `site_id`" argument does not survive contact with the schema
> either:** `site_id` and `work_item_id` get `NULLIF` because they are **uuid** columns and
> `''::uuid` raises — it is a type necessity, not a null-discipline choice. `domain` is `text`.
>
> **Two real findings this measurement turned up, which the NULLIF framing was hiding:**
> 1. **The census that produced "twenty writers" grepped `platform/` only.** There is a third
>    live INSERT site in `internal/`: `internal/agents/contentcreator/claims_guard.go:184`. It
>    **omits the `domain` column entirely** and the column has **no DEFAULT**, so it is a latent
>    NULL producer — the only one left. It has written **0 rows in the entire retained window**
>    (oldest row in the table is 2026-07-08), so it is dormant, not benign. Left unchanged
>    deliberately: dormant, another lane's package (`fa3b5207a`, bug 123), and see (2).
> 2. **It CANNOT use the shared door, structurally.** `contentcreator` holds a `*pgxpool.Pool`;
>    `agenterrors.Write` takes a `*sql.DB`. So RSH-008's "the ONE writer" is true of the
>    `database/sql` half of the estate only. Any future "convert the last writer" task must
>    solve the driver split first — it is not a copy-paste job.
>
> **What would have falsified this:** post-roll NULL rows still arriving, or a NULL-domain group
> attributable to an *unconverted* writer. Both were checked; both came back empty.

## `LogActionEntry`'s merge fills a provenance you meant to set — and every test in the package stays green

- **footprint:** `platform/orchestration/actions/log_action_error.go` · `LogActionError` · `LogActionEntry` · `LogActionEntryInheritingProvenance` · `LogActionFindings` · `LogActionEntryFindings` · `platform/orchestration/agenterrors` · `agent_error_log`
- **added:** 2026-08-07 (RFC_012 option B, commits `5f49b4cfd` + `f930de86b`)
- ✅ **FIXED AND NOW LIVE — v1.0.1268, pod-proven on BOTH chassis replicas 2026-08-08 evening.** The silent merge is gone from the fleet; **read the new-API paragraph at the bottom, not the pre-roll one above it.** Both commits are in the binary, checked with a discriminating set on `-jvfmc` and `-dwsdl` independently: control `Failed to write to agent_error_log` = 1 (proves the pipeline), `recorded as unattributed rather than credited to the running step` = 1 (the split), `names an agent_type but no step_name` = 1 (the step_name symmetry — this one separates the second commit from the first), `provenance_running_agent_type` = 1, and a needle present in no version = 0 (proves the grep can still return zero).
  - ⚠ **DO NOT verify this with an ANCHORED needle.** `strings /app/agent-chassis | grep -c "^unattributed$"` returns **0 on a binary that carries it** — the Go linker packs string constants into contiguous blobs, so `strings` emits dozens of them concatenated on one line and `^…$` can never match a constant. My own check returned that 0 and it is a needle artefact, not evidence. Use a **distinctive full phrase** from a log message or a context key; those are long enough to be unique without anchoring. General class filed separately below.
  - ⚠ **the `landmine-verifier` verdict on this entry is `NEEDS_HUMAN_REVIEW`, and it is a FALSE alarm — dispositioned here so nobody re-opens it.** It reported `LogActionEntryInheritingProvenance` returning 0 hits and correctly said it could not tell "never merged, removed, or stale". It is stale: the code index it reads is pinned at `93c576963` (08-07 09:31), which `git merge-base --is-ancestor` confirms is an **ancestor of the commit that created the symbol**, so the symbol could not possibly be in it. `git grep` finds it at HEAD and both pods carry the enclosing code. **The general lesson, because it will recur for every lane:** a landmine describing a NEW symbol is *guaranteed* to come back `NEEDS_HUMAN_REVIEW` when filed the same day, so that verdict on a fresh entry carries no information — settle it with `git grep` at HEAD plus a pod-grep, and write the disposition into the entry as here.

**The trap.** `LogActionError(ctx, params, siteID, domain, action, code, severity, msg, ctxPayload, logger)` resolves `agent_type` and `step_name` from `ActionParams` — i.e. **from whatever step is executing**. Several recorders in this package deliberately file under a DIFFERENT provenance: `component_link_repair` and `validate_page_content`'s link recorder file under the **origin of the content they repaired**; `store_generated_component` files as `component-creator`/`store_component`; `discovery_checks` and both envelope guards use their own helpers. Reach for the obvious helper at one of those sites and the row is silently misattributed — it names the wrong agent and the wrong step, so the next investigation opens the wrong file. `component_write_guard.go:276-279` states the cost: *"a row in agent_error_log that misattributes the writer is worse than no row."*

**Why you will not notice.** Every sqlmock test in this package pins the **error_code, the action and the message** — not `agent_type` and not `step_name`. A provenance slip is **green on the full suite**. A council round's edit-quality and guardian seats flagged exactly this risk on `store_generated_component_action.go` and it is quoted at `:1343-1351`; that objection is why the site is still unconverted.

**The check, ON A PRE-ROLL POD (still every pod today).** Use **`LogActionEntry`** (or `LogActionEntryFindings` for a loop) and set `AgentType` / `StepName` / `Action` **explicitly** whenever the row belongs to anything other than the running step. The merge only fills fields left ZERO, so a named field can never be overwritten — but a field you *forgot* to name is filled silently. Before converting or adding a site, read what the OLD code bound to `agent_type`/`step_name`: if it was anything other than `params.ExecutionContext.Sender.AgentType` / `.StepName`, it is a provenance site.

**The check, ON THE NEW API (at HEAD from 2026-08-08; the one to WRITE against).** The forgotten field is no longer silent, so the discipline changes from "remember to name it" to "say which one you mean" — and there is no longer a default that guesses:
- the row belongs to something OTHER than the running step → `LogActionEntry` / `LogActionEntryFindings` and name **`AgentType` AND `StepName`**. ⚠ **naming only `AgentType` no longer borrows the running step's `StepName`** — you get an EMPTY `step_name` and a `logger.Warn`, not a mixed row. That asymmetry was in the first version and two seats of the approving council round objected that a row carrying the caller's agent beside the runtime's step is a claim neither made; the door is strict on both columns from 2026-08-08.
- the row belongs to the RUNNING STEP and you need the Entry door anyway (an explicit `work_item_id` or `site_id` that is not the running step's) → **`LogActionEntryInheritingProvenance`** / **`LogActionEntryFindingsInheritingProvenance`**. Declaring it is the whole point: `grep -rn LogActionEntryInheritingProvenance` is the census of every site trusting the running step, and a reviewer of the CALLER can see the decision without opening the helper.
- the row belongs to the running step and you need nothing else → `LogActionError` / `LogActionFindings`, exactly as before. Their contract already *is* running-step identity, so they declare inheritance for you.
- **forget entirely and the row still lands — as `agent_type = 'unattributed'`**, with `context.provenance_missing`, the running step demoted into `context.provenance_running_agent_type` / `_step_name`, and a `logger.Error`. It is not silent and it is not lost, but it is not attributed either, so it will not answer an investigation. Standing detector: `SELECT action, error_code, count(*) FROM agent_error_log WHERE agent_type = 'unattributed' GROUP BY 1,2;` (0 rows as of 2026-08-08 — a non-zero result is a code defect, not traffic).

⚠ **AND WHAT THE INHERITING DOOR GIVES YOU IS OFTEN THE FILLER `generic`, WHICH IS THE TRAP INSIDE THIS ONE.** Declaring inheritance makes the decision visible; it does not make the value good. The ladder is `ExecutionContext.Sender.AgentType` → `params.AgentType`, and the dispatch-path sender is *usually* `generic` — all **25** live `REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows carry it, and `generic` holds **559 rows across 25 distinct `step_name`s**, the widest spread of any `agent_type` on a table whose main investigation index is `(agent_type, occurred_at DESC)`. `types.ExecutionContext.RunAgentType` is the estate's existing answer (`bugs_open/060`, and its own doc comment says the sender "is often 'generic'"), but **only `coordinator.determineOwnerAgentType` consults it** and `actions` cannot reach that ladder across the very import edge that created `agenterrors`. So: if the row matters for attribution, do not settle for inheritance — read `RunAgentType` yourself or name the provenance. Hoisting the ladder onto `*types.ExecutionContext` so both packages share ONE copy is the structural fix and is NOT done.

Two more in the same family, both cheap to get wrong:
- **`agent_type` is NOT NULL.** Three sites carry their own `"unknown"` fallback. Delegating that to the merge yields `''` when both params sources are empty — a constraint violation that the best-effort writer swallows as a warn, so the row just vanishes. Keep the fallback explicit.
- **A converted site's `domain` goes `''` where it was `NULL`** (the shared writer does not `NULLIF` it). Measured inert 2026-08-06 — 9,964 `''` against 128 NULL already live, and the only domain-filtering reader (`diagnose_load_runtime_action.go:267`) treats them identically — but if you add a reader doing `WHERE domain IS NULL`, it will not see these rows.

**Proving a change to this seam is live** — the SQL text is byte-identical before and after by design, so grepping the statement proves nothing. Count the copies instead:
`kubectl exec -n ai-persona-system <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "INSERT INTO agent_error_log"'` — **14** is the pre-conversion binary, **2** is a build from `f930de86b` or later. Run it on **every** replica.

> **CORRECTED 2026-08-07 — do NOT run the count above; it reports a shipped binary as unshipped.**
> A correct post-conversion binary reads **1**, not 2. The two surviving sites hold
> **byte-identical** SQL and the **Go linker deduplicates identical string constants**, so two
> sites contribute one string. The pre-conversion 14 was 14 only because the hand-copies had
> drifted to 8/9/10/11/13 columns — **converging them is what collapsed the count**, so this
> needle measures the symptom of success and reports it as failure. Measured 1 on both replicas
> of v1.0.1262, which does carry `f930de86b`.
> **Use a discriminating pair instead** — `f930de86b` reworded one log line singular→plural, so
> both halves must hold in the same binary:
> ```bash
> strings /app/agent-chassis | grep -cF "failed to write some discovery check error records"  # POS -> 1
> strings /app/agent-chassis | grep -cF "failed to write discovery check error record"        # NEG -> 0
> ```
> The general class — *a `strings | grep -c` counts distinct SPELLINGS, never SITES* — is its own
> entry at the end of this file. See `WRONG_CALLS.md` 2026-08-07.
>
> **AND A SECOND WAY THE SAME COMMAND LIES, found 2026-08-08 verifying the provenance fix:
> an ANCHORED needle can never match a Go string constant.** `strings /app/agent-chassis | grep -c
> "^unattributed$"` returns **0 on a binary that contains it**, because the linker packs constants
> into contiguous blobs — `strings` emits them concatenated dozens-to-a-line, so `^…$` has nothing
> to anchor to. Confirmed in place: unanchored, the same binary returns 4, and the surrounding line
> reads `…conditions_recordededitorial_referrerscomponents_examined…`. So: **needle on a
> distinctive full PHRASE** from a log message or a context key (long enough to be unique without
> anchoring), never on a short bare constant, and never with `^`/`$`. A 0 from an anchored needle is
> not a negative control — it is a broken instrument, and it fails in the direction that reads as
> "the fix did not ship".

---

## A `090` diagnosis can finish `complete` with its verdict COMPUTED AND THROWN AWAY — the only copy is in `collected_data` and it reaps in 24h
- **footprint:** `diagnosis_artifacts`; `site_work_items.item_type='needs_diagnosis'`;
  `090_TRIGGER_needs_diagnosis_v1.sh`; `diagnose-agent`; `diagnose-orchestrator`;
  `diagnose-dispatch-loop`; `platform/orchestration/coordinator.go`;
  `platform/kafka/reply_delivery.go`; `notifyParentOfFailure`; `DeliverReply`
- **fires when:** you run `090` (or any spawned child workflow) and go looking for its
  verdict. No symptom, no error surfaced to you, and every status you would naturally
  check says the run succeeded.
- **why the wrong answer looks right:** measured 2026-08-07 on run
  `a7b1e113-8857-4161-ad2b-f3b7387e33e9`. `site_work_items.status` = **`complete`**; all
  three `orchestration_states` rows = **`COMPLETED`**; `diagnosis_artifacts` holds **5
  rows, every one `kind='bundle'`** and **zero** report/verdict/diagnosis rows. A reader
  checks the status, looks for the report, finds none, and has nothing to tell them
  whether the loop declined to conclude or lost its conclusion. **It had concluded** — a
  full verdict with citations and named evidence gaps.
- **where the verdict actually is, and how long you have:**
  ```sql
  -- enumerate first: the keys are not documented anywhere
  SELECT string_agg(k, ', ') FROM jsonb_object_keys(
    (SELECT collected_data FROM orchestration_states
     WHERE owner_agent_type='diagnose-agent'
       AND collected_data::text LIKE '%<RUN_CORRELATION>%')) k;
  -- then: 'verdict', 'diagnosis', 'diagnosis_note' are the ones that matter
  SELECT jsonb_pretty(collected_data->'verdict') FROM orchestration_states
  WHERE owner_agent_type='diagnose-agent' AND collected_data::text LIKE '%<RUN_CORRELATION>%';
  ```
  **`COMPLETED` rows are deleted 24h after `updated_at`** by
  `scheduled_tasks.database-cleanup`, so the only copy of the verdict has about a day to
  live. **Recover it before you do anything else**, and paste it into your lane's notes —
  an artifact table with no report is not going to grow one later.
- **the check, to tell "lost" from "declined":** ask the rows what happened, not the status —
  ```sql
  SELECT agent_type, pod_name, step_name, occurred_at, error_message
  FROM agent_error_log
  WHERE error_message LIKE '%could not be delivered to the parent%'
    AND occurred_at > now() - interval '2 hours';
  ```
  Two rows (`diagnose-agent` + `diagnose-orchestrator`, `step_name='complete'`) mean the
  verdict was produced and rejected in transit. `coordinator.go` is behaving correctly here —
  it deliberately converts an undeliverable success into a parent-facing failure rather than
  leaving silence — but **nothing propagates that into the work item or the artifact table.**
- **three sub-traps, each of which cost me a step:**
  1. **`error_code` on those rows is `UNKNOWN`**, the classifier's fallback — so you cannot
     find this class by filtering `error_code`. The informative code
     (`CHILD_ORCHESTRATION_FAILED`) exists only inside the Kafka message body. Filter on
     `error_message LIKE`.
  2. **Do not grep the chassis pods.** `diagnose-agent`/`diagnose-orchestrator` run in their
     own `agent-diagnose-*` pods. I grepped both chassis replicas, got a clean zero, and it
     was the wrong-service trap — a positive control showed 219 log lines in the window and
     **0** containing "diagnose". `agent_error_log.pod_name` names the right pod outright.
  3. **A `bundle` artifact count is not progress toward a report.** Five bundles looks like a
     loop that worked. Bundles are *input* assembly; the report is a different `kind` that
     may never arrive.
- **adjacent, and the reason a re-run will not help:** if the verdict says **UNVERIFIABLE**
  with gaps like "zero hits for `package <x>`" or zero-match symbol searches, that is the
  frozen **code index** (`bugs_open/108`; it was pinned at `d98010e8b`/2026-07-28 until
  migration 332 repointed it on 2026-08-07 — **do not read that sha as current, re-check the
  pin**), not a weak
  hypothesis — anything added since reads as ABSENT. Re-submitting spends another run for the
  same answer. Fix the index ref first. ~~(migration 252 pins `086_experience_loop`; it wants
  `'main'`)~~ **CORRECTED 2026-08-07 08:2xZ, and this matters because the old advice made it
  WORSE: `main` was 4,594 commits behind HEAD while the stale pin was only 2,365 behind, so
  "change it to `'main'`" — which is what 252's own reversal trigger and the memory index both
  said — would have roughly doubled the staleness.** The ref is the *constant* in
  `scheduled_tasks.pre_query` for `code-index-refresh` (NOT `input_data->>'ref'`, which the
  scheduler's pre-query overlay makes inert), and it must name **the live working branch**.
  Repointed to `087_towards_multiple_domains` by **migration 332**.
  **⚠ THE PIN ROTS EVERY TIME A BRANCH IS CUT, silently and while still reporting a real recent
  commit** — so before believing any "no hits" answer, check the pin against the tree:
  `git ls-remote --heads origin | grep -E '<pinned-ref>|<your-branch>'` and
  `SELECT ref, commit_time FROM code_symbols GROUP BY 1,2`. A daily job that succeeds against a
  dead branch looks exactly like a healthy index.
- **source:** 2026-08-07, code-review triage lane (`code_review_triage_2026-08-05/` NOTES §18).
  Root cause of the validation rejection **NOT diagnosed** — `reply_to_request_id` is empty in
  the *failure* message and `DeliverReply` keys on it, but that is a **[HYPOTHESIS]**, not a
  finding: the inspected message is not the rejected one.

## `strings <binary> | grep -c` counts DISTINCT strings, not SITES — the Go linker dedupes identical literals, so a de-duplication refactor moves the number you are using to prove it shipped
- **footprint:** `strings /app/agent-chassis`; `grep -c`; `strings`;
  `platform/orchestration/agenterrors/agenterrors.go`;
  `platform/orchestration/actions/store_generated_component_action.go`;
  `INSERT INTO agent_error_log`
- **fires when:** you prove a deploy the way this estate correctly insists you prove one —
  grep the running binary for a string your change touched — **and your needle is a COUNT
  of a literal that appears at more than one site.** No symptom, and the check itself is
  the recommended practice: the trap is in the choice of needle, not in the method.
- **why the wrong answer looks right:** the Go linker **deduplicates identical string
  constants**, so N sites holding byte-identical text contribute **one** string to
  `.rodata`. The count therefore tracks *how many distinct spellings exist*, not *how many
  call sites exist* — and those two numbers diverge exactly when a refactor makes copies
  identical, which is what most de-duplication work is FOR. The failure is silent and
  points the wrong way: you get a number lower than predicted and read it as "my change is
  missing".
- **the case, 2026-08-07 (RFC_012 lane, RSH-008).** The `agenterrors` conversion collapsed
  19 hand-copied `INSERT INTO agent_error_log` statements onto one shared writer. The
  pre-conversion binary carried **14** distinct statements (14 because the copies had
  drifted to 8/9/10/11/13 columns); two INSERT sites remained in the tree afterwards, so
  the published acceptance test — in the lane HANDOFF, in NOTES **and** in the concept
  register — said the binary "must read **2** on every replica". A correct binary reads
  **1**: the two survivors (`agenterrors.go:89` and the deliberately-unconverted
  `store_generated_component_action.go:1353`) hold byte-identical SQL. **Converging the
  copies is what collapsed the count**, so the needle measured the symptom of success and
  reported it as failure. Caught on first use by the author, who could not explain the
  number; a later session would have concluded the work had not shipped and possibly
  rebuilt or reverted it.
- **the check — use a DISCRIMINATING PAIR, never a count.** Pick one string your change
  **added** and one it **removed**, and require both in the same exec on every replica.
  Best of all is a reworded line, because the two halves are near-twins and no stale image
  or lucky substring can satisfy both:
  ```bash
  kubectl exec -n ai-persona-system <pod> -- sh -c '
    strings /app/agent-chassis | grep -cF "failed to write some discovery check error records"  # POS -> 1
    strings /app/agent-chassis | grep -cF "failed to write discovery check error record"'       # NEG -> 0
  ```
  Derive the pair mechanically from your own commit rather than from memory:
  `git show <sha> --unified=0 -- '**/*.go' | grep '^+' | grep -oP '"[^"]{20,80}"' | sort -u`
  (and the same with `^-`), then **confirm each candidate negative is actually gone from
  the tree** — many "removed" literals are merely moved from SQL text to a bind parameter
  and are still in the binary.
- **if you must use a count, state what it counts.** "14 distinct statements" is a fact
  about spellings. Never write "N sites" over a `grep -c` of a literal, and never predict
  a post-refactor count without asking whether the refactor makes any two sites textually
  identical.
- **source:** 2026-08-07, `docs024_key_docs_latest/rfc012_await_findings/` — NOTES 08-07
  (morning) misstep 8; corrections in place in that lane's HANDOFF and in RSH-008.

### A `complete` work item was graded by whichever producer OWNS the `item_type` — not by the producer that filed it, and the grader is right about the wrong question

- **footprint:** `site_work_items` (the `status` and `item_type` columns, and `spec->>'audit_source'`), `platform/orchestration/actions/write_audit_findings_action.go` (`designItemTypes`, `designRouting`), `platform/orchestration/actions/discovery_checks/` (`RegisterVerifier`, every `Verify*Resolved`), `platform/orchestration/actions/complete_work_item*`
- **fires when:** you read a work item's `status` to decide whether a defect was repaired, or you close/query a route and conclude from `complete` that its sites are clean. Two or more producers may file under one `item_type`; the completion verifier is registered **per `item_type`**, by whichever check happened to define that name first. An item from any other producer is then re-checked against a predicate its author never meant.
- **why the usual defences miss it:** this is **not** the fail-open policy of `RFC_017` — nothing errors, so no error is recorded. And it is not a buggy verifier: reading the verifier's source looking for a wrong predicate finds a *correct* one, with an honest doc comment describing the question it answers. The mismatch is invisible in the item_type, invisible in the verifier, and invisible in the status. It is visible in exactly one place: **which producer filed the row.**
- **the worked case:** design-audit category `dark_section` maps to item_type `hardcoded_section_colors` (`write_audit_findings_action.go:117`). That item_type's verifier re-runs the *discovery check's* population and filters it by `ReplaceHardcodedColors`' remit — "are there hardcoded dark hex literals left?". gamesdesign's defect was `background: var(--color-primary, #1a1a2e)`, **already a `var()`**, so the population was empty and the verifier returned `Resolved: true` with the detail *"no unlocked component carries a colour within the fixer's remit"*. Item `complete` in 3m17s, nothing written, defect still live four days later. `bugs_open/213`.
- **the check — an asymmetry, not a number.** Split the route by producer and look at which one ever fails:
  ```sql
  SELECT status,
         count(*) FILTER (WHERE spec->>'audit_source' IS NOT NULL) AS from_audit,
         count(*) FILTER (WHERE spec->>'audit_source' IS NULL)     AS from_discovery,
         count(*) AS total
  FROM site_work_items WHERE handler_agent = '<agent>' GROUP BY 1;
  ```
  If one producer's items are **never** `unresolved`/`failed` while the other's routinely are, the grader almost certainly cannot see the first producer's defect. On this route it was 7 of 7 `complete` for the audit producer and 6 of 6 not-complete for the discovery producer.
- **then prove it at the target, and use the CLOCK, not the content:** compare the target's `updated_at` against the item's `created_at`. A component last written **10.5 hours before the item existed** is proof no handler wrote it — re-reading the template only tells you it looks unfixed, which a re-render can muddy. (On gamesdesign the page *did* re-render 16 minutes after the close, which makes the row look more trustworthy, not less.)
- **and check whether the item graded itself:** `spec->>'acceptance_test'` is written by the audit producer and, on this route, **read by nothing** — every consumer in the tree belongs to the `improve_tool` / tool-acceptance family. An item carrying its own pass condition and closed without it being read is the shape to distrust.
- **source:** 2026-08-07, `bugs_open/213`, found from `bugs_open/212` by the `bugfix_122_contrast_ink_slots` lane. Related: `bugs_closed/077` (same detector, remit split), `RFC_017` (the other hole in this registry), owner ruling 2026-08-02 / RFC_010 narrowing 1 (N producers on one `item_type` is permitted **provided** the producer set and key shape are written down — this is the cost when they are not).

### `count(DISTINCT created_by)` cannot detect a second producer of an `item_type` — it is a free-text config field with a `generic` fallback, and for 8 live item_types EVERY row says `generic`

- **footprint:** `site_work_items.created_by`, `platform/orchestration/actions/create_work_item_action.go` (`source`, ~lines 129-131 and 283), `platform/agentbase/agent.go` (`agentType` fallback), `platform/orchestration/coordinator.go`, `platform/messaging/processor.go`, and the mandatory producer count in `HANDOFF_2026-08-03b` §3 / any `discovery_checks/check_*.go` adopting RFC_010 or registering a verifier
- **fires when:** you do the producer check this repo mandates in writing — *"grep the Go producers, corroborate with `created_by` over every row"* — before registering a verifier, adopting the retraction seam, or trusting a route's `complete` rows. The grep half is sound. **The corroboration half is not a second, independent way of asking; on some types it cannot return anything but 1.**
- **the tell:** **there is none, and the answer is the reassuring one.** `created_by` is set to `config["source"]` — a free-text string any agent definition may supply — falling back to `params.AgentType`, which itself bottoms out at the literal `"generic"` (`agentbase/agent.go:158`, `coordinator.go:3482`, `processor.go:909`). So distinct real producers collapse into one label, and two producers that both go unresolved are **indistinguishable from one producer**. Measured 2026-08-07: `generic` carries **20+ item_types** (`phantom_internal_link` 45 rows, `undeployed_asset` 98, `page_rerender` 54), and **8 item_types have `generic` as their ONLY label** — `voice_tells` (25 rows), `placeholder_contact`, `asset_reference_404`, `stale_directory_claim`, `directory_citation_unverified`, and three `missing_*_section` types. On any of those the corroboration is guaranteed to say "1 producer" whatever the truth is: **it could not have come out otherwise.**
- **and it is worse than a blind check — the producer set is not enumerable from CODE at all.** `CreateWorkItemAction` is generic and config-driven: any `agent_definitions` row can file any `item_type` under any `source`, with no code change, no registration and nothing to grep. Measured the same day: **11 live definitions** file work items this way (improvement-loop, claims-auditor, domain-strategist, deduplicate-sections, tool-improver, …). A clean Go grep therefore proves only that no *compiled* second producer exists today.
- **the check:** never let `created_by` corroborate a grep — ask it what it can actually answer, and show the `generic` share so a collapsed label cannot hide:
  ```sql
  SELECT count(DISTINCT created_by) AS labels,
         count(*) FILTER (WHERE created_by = 'generic') AS unattributed,
         count(*) AS total,
         string_agg(DISTINCT created_by, ' | ') AS which
  FROM site_work_items WHERE item_type = '<your type>';
  ```
  **If `unattributed > 0`, the label column is not evidence about producers** — fall through to the config side, which the grep cannot see either:
  ```sql
  SELECT ad.type, s->'config'->>'source', s->'config'->>'item_type'
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS st(k,s)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s->>'action' = 'create_work_item';
  ```
  **Run that one WITHOUT your item_type filter first** — it must return the 11-row population, or your query shape is wrong and its zero means nothing. A `total` of 0 (a type that has never been filed) is equally uninformative: it is not a single-producer finding, it is no finding.
- **worked case:** `dead_fragment_link`, 2026-08-07, checked against `bugs_open/213`. The answer came out genuinely clear — one Go producer, no `designItemTypes` route, no config producer — **but only the first and third checks carried any information.** The `created_by` corroboration returned zero rows (the type has never been filed) and would have been read as agreement. The arm is safe by circumstance: nothing structurally stops a definition filing `dead_fragment_link` from config tomorrow, whereupon `VerifyDeadFragmentLinkResolved` grades it against the fragment predicate regardless of what it described.
- **why it is a landmine:** the unsound half is **prescribed in this repo's own guidance**, and the entry directly above it ("count the CLOSERS…", 2026-08-04) certifies the producer check as *"mandatory, correct, and holed only on the closer side"*. It has a second hole, on the producer side, and following the instruction exactly is what produces the false clean answer. Same shape as the jsonb `::text LIKE` entry: right about the problem, wrong about the remedy.
- **relations:** `bugs_open/213` (the entry above — a verifier grading the wrong producer's item; its `spec->>'audit_source'` split is the right discriminator *for the design-audit producer specifically*, and there is no general producer field to generalise it with), owner ruling 2026-08-02 / RFC_010 narrowing 1 (converging N producers on one `item_type` is permitted **provided the producer set is written down** — this entry is why measuring it is not a substitute for writing it down)
- **added:** 2026-08-07, `bugfix_071_fragment_blindspot` lane

### `site_plan_imagery.scope_ref` looks like a foreign key and is LLM-minted free text — 5 of 131 section-scope refs do not resolve, and the consumers that "work" are tolerating, not validating

- **footprint:** `site_plan_imagery.scope_ref`, `flattenImageryBlock` / `write_site_plan_action.go:827-897`, the lock carry-forward `write_site_plan_action.go:651-720`, `plan_sections_action.go:362-386`, `flag_page_image_rebuild_action.go`
- **fires when:** you write ANYTHING that joins, parses or carries `scope_ref` — a consumer resolving the `"<page>:<ordinal>"` to a section, an extension of the lock carry-forward, a census keyed on it, a repair that trusts the ordinal to find the right section.
- **the tell:** there is none at write time. The key is copied verbatim from the planner LLM's `imagery` block; nothing validates the page name or the ordinal against the sections that survived `validate_plan`. Existing consumers LOOK like precedent for trusting it, but the build join is `LIKE <page> || ':%'` (ignores the ordinal entirely) and the carry-forward only ever compares refs to each other — neither ever resolves one, which is why 5 orphans (2 sites, one minted 2026-08-07 by the newest prompt) sat unnoticed with 4 paid-for active assets unreachable behind them.
- **the check:** before consuming scope_ref as a reference, run the orphan census in `bugs_open/214` (page+ordinal resolved against `site_plan_sections`, per plan). If you are writing rows: resolve against the post-validate sections array in the same `CollectedData` — fix candidate 1 in the bug file, ~30 lines at the single door.
- **relations:** `bugs_open/214` (mechanism + census + fixes), `bugs_open/114` (the unreachable-asset symptom), RFC_016 §1 (the contract rule that forbids this keying for new fields; imagery is its named counter-example), 016b §9 2026-08-07 (the transferable pattern).
- **added:** 2026-08-07, brochure_component_library lane

### `stale_site_components` fires on chrome that is not stale and stays silent on chrome that is — 36 of its 39 live firings are unrelated to its own subject

- **footprint:** `platform/orchestration/actions/discovery_checks/check_integrity.go` (`StaleSiteComponentsCheck`, the `stale_site_components` check), `site_components`, `site_work_items` rows with `item_key LIKE 'stale\_sc\_%'`, `needs_rerender` items handled by `rerender-pages`
- **fires when:** you trust a `stale_sc_*` work item, or the absence of one, as evidence about whether a site's header/footer/head is out of date — e.g. verifying a chrome fix reached the fleet, sizing a chrome re-render, or concluding a site is clean because no item was filed.
- **the tell:** there is none from the item. Its summary reads *"Site component <slot> is stale — last rendered before page content was updated"*, and the second clause is the whole defect stated out loud: the check compares `site_components.updated_at` against `MAX(page_components.updated_at)` for the site (`check_integrity.go:320-331`). Chrome is rendered from `content_components.html_template`, `site_nav_items` and site identity — **never from `page_components`.** The reference is independent of the subject, so the item's presence tracks recent page-content activity, not chrome drift. It is a live, draining check (7 complete per slot, latest 2026-08-06), which is what makes it dangerous: its output is believed and it consumes real rebuild capacity.
- **the check:** never read the item; run the cross-tab and read both off-diagonals. This is the query, and it must reproduce the production predicate exactly — including `- INTERVAL '24 hours'` and the `pc.rendered_html IS NOT NULL` filter, or the disagreement you measure is your own:
  ```sql
  WITH real_stale AS (
    SELECT sc.site_id, sc.slot_name, (cc.updated_at > sc.updated_at) AS truly_stale
    FROM site_components sc JOIN content_components cc ON cc.id = sc.component_id
    WHERE sc.rendered_html IS NOT NULL AND sc.rendered_html <> ''
  ), detector AS (
    SELECT sc.site_id, sc.slot_name,
           (sc.updated_at < mx.latest - INTERVAL '24 hours') AS detector_fires
    FROM site_components sc
    CROSS JOIN LATERAL (
      SELECT MAX(pc.updated_at) AS latest FROM page_components pc
      JOIN pages p ON p.id = pc.page_id
      WHERE p.site_id = sc.site_id AND pc.rendered_html IS NOT NULL) mx
    WHERE sc.rendered_html IS NOT NULL AND sc.rendered_html <> '' AND mx.latest IS NOT NULL
  )
  SELECT r.truly_stale, d.detector_fires, count(*)
  FROM real_stale r JOIN detector d USING (site_id, slot_name) GROUP BY 1,2;
  ```
  Measured 2026-08-07 over 53 rows: **t/t 3 · t/f 1 · f/t 36 · f/f 14.** All four cells populated, so it could have come out otherwise.
- **and the corrected comparison is NOT "widen the timestamps".** `GREATEST(cc.updated_at, nav.updated_at, sites.updated_at)` marks ~every row stale — `sites.updated_at` churns for unrelated reasons. Worse, **no timestamp can be complete here:** `fixTemplateColors` (`fix_harcoded_colours_action.go:180`) does `UPDATE content_components SET html_template = $1 WHERE id = $2` with **no `updated_at`**, and its selection query joins through `site_components`, so it edits chrome templates invisibly to every timestamp-based detector. It is the only one of ~9 `UPDATE content_components` writers that omits the stamp.
- **and a timestamp answer is unfalsifiable anyway:** "stale" by timestamp does not mean the output would change. On the 4 rows flagged 2026-08-07, **no** `class="…"` literal in `footer-theme-chrome` splits the 16 stored footers by the template-change time — so a re-render might be a no-op. The question worth asking is *"would a re-render change anything?"*, which is a render-input fingerprint, not a comparison of two `updated_at`s.
- **do NOT duplicate the sibling check:** `deactivated_site_components`, same file, already covers "`component_id` points at an `is_active=false` component" (extended by `bugs_open/170` to the `style_collections` pin). 17 of 19 `head` rows are in that state — it is a known, covered condition, not a new finding.
- **relations:** `bugs_open/117` (chrome is a stored artefact no page re-render regenerates — this is the detector half of it), `bugs_open/170` (the pin half of the sibling check), `bugs_closed/049` (stale chrome fleet-wide), 016b §9 *"Light site renders dark chrome"* (the two-assembly-paths prior art), `docs024_key_docs_latest/bugfix_117_chrome_staleness_reference/` (measurements, runbook, handoff)
- **added:** 2026-08-08, `bugfix_117_chrome_staleness_reference` lane

## A decomposed site's `prose-N` slot may hold the page's `<style>` block, and rewriting it deletes the CSS while every guard still passes

**footprint:** `page_components.slot_name LIKE 'prose-%'` · `content_rewrite` /
`page-build-handler` on any decomposed site · `ported-prose`

A decomposer that splits a hand-built page into positional slots classifies by
POSITION, not by content. On loancalculator.co.uk **8 of 51 `prose-*` rows are not
prose at all** — they hold the `<style>` block that styles the page's calculator
(`.comparison-wrapper`, `.loan-column`, `.stat-label`, `.stat-value`). They look
exactly like prose rows in every listing: same slot name, same component
(`ported-prose`), same `content.source = "llm"`, so a voice rewrite sends them to
the writer like any other block.

**Every protection in place says yes to this.** The component's own `llm_guidance`
promises *"this block contains NO form control and NO element addressed by any
script, so rewriting this prose cannot break a calculator"* — which is **true**, and
says nothing about CSS. The locked-row guard protects the tool row, not the style
row. `validate_page_content` did not object. The tool's arithmetic still computes
perfectly; only its layout collapses.

**And it is a coin flip, not a rule:** on the same run the writer PRESERVED the
`<style>` block on 3 of these pages and DROPPED it on the 4th. A single spot-check
of one page would have cleared the whole class.

**the check:** before rewriting, list the slots that carry CSS —
```sql
SELECT p.name, pc.slot_name, length(pc.content_data->>'content')
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='<site>' AND pc.slot_name LIKE 'prose%'
  AND pc.content_data->>'content' ILIKE '%<style%';
```
— then after each rewrite assert every selector survived, **per page, never sampled**:
extract `([.#][\w-]+)\s*\{` from the baseline row and require each one to still match
`selector\s*\{` in the new row. Checking for the literal `<style` tag is NOT enough —
what breaks the page is a lost selector, and a rewrite can keep the tag while dropping
rules. Recovery is an exact row restore (`content_data`, `rendered_html`,
`content_hash`) from a pre-run backup plus an assemble-only rerender; confirm at the
SERVED page, because the stored row and the deployed file are independent facts.

---

## A new pattern in `validate_page_content` is a BLOCKER by default, and a blocker there means "this page can never be rebuilt again"

- **footprint:** `platform/orchestration/actions/validate_page_content.go`
  (`placeholderPatterns`, `metaCommentaryPatterns`, `checkPlaceholderPatterns`,
  `checkMetaCommentary`, `checkUnrenderedTemplates`, `checkDomainContamination`),
  `content_components.html_template`

**The trap.** Adding a string to one of those pattern lists looks like adding a
warning. It is not. `Severity: "blocker"` reaches a categorise block that does
`return nil, fmt.Errorf("content validation failed: …")` — the **action errors**, the
step fails, and the page is never saved. There is no review queue, no partial save,
no retry that behaves differently: the page becomes **permanently un-rebuildable**
until someone edits whatever the scanner mis-read. Until 2026-08-08 the file's own
doc comment said the opposite in as many words — *"A false positive here routes to
needs_human_review, not silent breakage"* — and that sentence is why the pattern
lists were treated as safe to extend (now corrected in place).

**It fires with no symptom and no suspicion**, because the page keeps serving. The
failure only appears the next time someone asks for a content change, possibly weeks
later, in a different lane — and the error blames the model. Both live instances were
found by someone doing unrelated work: `bugs_open/219` (three pages; the string was
in a developer's `/* … */` note inside the tool's `<script>`) and `bugs_open/221`
(webdesign.co.uk `tools-index`; the copy *"LocalBusiness schema, as an AI-builder
prompt"* matches the pattern `as an ai`).

**Two prose scans have already been re-scoped and a third written the old way
inherits the defect.** `checkPlaceholderPatterns` (218) and `checkMetaCommentary`
(219) now scan `datahelpers.ExtractAssertionText(html)` + `headProseBlocks(html)` —
a real HTML parse that excludes `script/style/code/pre/head/template/noscript/svg/
iframe/textarea/select/option`, attributes, comments and doctypes. Reaching for
`strings.Index(strings.ToLower(html), pattern)` is the idiom the file used to use and
it is still what the surrounding code looks like. **Do not add a second stripper**
either — a council REVISE refused exactly that on 218.

**the check:** before adding or widening a pattern, ask what ELSE in an assembled
page contains that string — the answer is usually "a comment a conscientious human
wrote about the thing your pattern names". Census it, and note the query must scan
the whole template, because the string will not be in an HTML comment:

```sql
-- how many components would this pattern convict, and in what?
SELECT function, substring(html_template from greatest(position('<pattern>' in html_template)-120,1) for 300)
FROM content_components WHERE html_template ILIKE '%<pattern>%';
```

Then decide severity deliberately: `blocker` is correct only if a page serving that
string is worse than a page nobody can ever rebuild.

**And for verifying a fix to any of these checks, the artefact you need is usually
still there.** `orchestration_states.collected_data->'page_content'->'response'->>'page_html'`
holds the **exact bytes** that failed, for ~24h before pruning. Run the changed
function over those, not over a fixture you composed — a fixture you write to
exercise a rule will exercise it. The old-code half of the control pair is the
production row's own `error` text, which is stronger than any local replica.

**ADDENDUM 2026-08-08 (`bugfix_221_ai_disclosure_precision`, fix `61c8cc6ff`) — an entry may now carry a REGEX, and the default is still the dangerous one.**
`metaCommentaryPatterns` entries gained an optional `Re *regexp.Regexp`. **`Re: nil`
means substring — the status quo, NOT the cautious direction.** Substring
over-matching is precisely what 221 was, and the compiler will not remind you: a new
entry whose needle can sit inside a longer legitimate phrase (`as an ai` inside `as
an AI-builder`) must set `Re` and match the **construction**, not the noun phrase.
⚠ **A word boundary is not a substitute** — `\bas an ai\b` still matches `as an
AI-builder`, because `-` IS a boundary. That is the single fact that made 221 survive
the obvious fix.

**Two measurement traps on top of the census above, both of which hand you a
confident zero** (each cost this lane a wrong reading on 2026-08-08):

1. **A SQL `LIKE` is not this check, and they disagree BY DESIGN.** `LIKE` over
   `rendered_html` sees `<script>`/`<style>`/attribute bodies; since 218/219 the
   checker sees only extracted prose. Use SQL to *locate* rows, then run the real
   function over those rows' stored bytes — otherwise you will "reproduce" 219's
   already-fixed scope bug and chase it again. Four `input_schema`/`on_missing` rows
   are live right now that `LIKE` flags and the checker correctly ignores; they are a
   ready-made **negative control** for any harness you build here.
2. **Past firings are NOT in `agent_error_log.error_message`.** That column holds only
   *"content validation failed: 1 blockers, 0 errors"* / *"see context.issues for
   detail"*, so every spelling of your pattern against it returns **0 fleet-wide while
   the check is actively blocking builds**. Enumerate the jsonb:
   `SELECT iss->>'category', iss->>'value', count(*) FROM agent_error_log ael,
   jsonb_array_elements(ael.context->'issues') iss GROUP BY 1,2;`
   ⚠ and not `context::text LIKE '%"category":"meta_commentary"%'` either — jsonb
   renders a **space** after the colon, so that form matches nothing.

**Same class, different checker, separately owned:** `bugs_open/222` —
`check_tool_fabrication_action.go`'s `fabQualifierNearData` regex has no negation
awareness, so a comment *denying* fabrication convicts. Do not fix it from here.

---

### The work-item dispatcher hands EVERY handler `spec.page_name` — an item filed against a different page (its `page_id` column) is acted on at the WRONG page, successfully

- **footprint:** `platform/orchestration/actions/discovery_checks/`, `build-dispatch-loop`, `site_work_items.page_id`, `input_mapping`
- **added:** 2026-08-08, from `bugs_open/220` (found by the `bugs_open/206` lane)

If you write or modify a discovery check whose remediation should act on a page OTHER
than the one the finding lives on (an unbuilt link's TARGET, a sibling, an index over
its members), setting `WorkItemSpec.PageID` to that other page does **nothing at
dispatch time**: `build-dispatch-loop`'s `call_handler` maps `"page_name?":
"current_item.spec.page_name"` and reads no routing column. Your handler will receive
the FINDING's page, rebuild it, succeed, and `mark_complete` will close the item.
The wrong result is a green row plus a real (irrelevant) deploy — nothing errors,
and dedup's terminal-status rule re-mints the item next pass, forever.

**the check:** before trusting a cross-page item type end-to-end, read the live
mapping —

```sql
SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}'
FROM agent_definitions WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

— and if your actionable page ≠ `spec.page_name`, either set `spec.page_name` to the
TARGET's name in the check (the working workaround), or fix the mapping to honour
`page_id` (the class fix, `bugs_open/220` candidates 1/3). Verify at the TARGET
page's `pages.deployed_at`, never at the item's `status` or the handler's success
payload.

### A hand-written walk into `sub_workflow` cannot see `substeps` — and `substeps` is the half that RUNS, so on a step carrying both you audit the inert copy

- **footprint:** `platform/validation/subworkflow.go`, `platform/validation.WalkSteps`,
  `subWorkflowsOf`, `pkg/models/substep_decode.go`, `pkg/models.DecodeSubWorkflowStep`,
  `platform/orchestration/actions/loop_actions.go`,
  `cmd/config-key-audit/sharedoutputs.go`, `cmd/config-key-audit/relaygaps.go`,
  `step.Config["sub_workflow"]`, `substeps`
- **fires when:** you write anything that walks workflow steps in Go — a detector, an
  audit, a census, a "which agents carry action X" answer — and you descend into
  nested steps yourself. The obvious descent reads `step.Config["sub_workflow"]`,
  because that is the shape 17 of the 18 live nested steps use and the only one you
  will see while reading definitions. **A loop's body may instead be declared under
  `substeps`, and at execution `substeps` WINS** — `loop_actions.go:91` reads it first
  and consults `sub_workflow` only when it is absent or empty.
- **the tell:** there is none, and the failure is worse than a miss. A `sub_workflow`-only
  descent returns **0 findings** over a `substeps` body, which is indistinguishable from
  a clean report. On a step carrying **both** shapes it is not silent but WRONG: it walks
  the `sub_workflow` half, which the executor ignores, so it reports a hazard in config
  that never runs — and a reader chasing that finding will not find the behaviour.
  Neither direction produces an error, a warning, or an empty result.
- **the check:** do not write the descent. `validation.WalkSteps` is exported for exactly
  this and is the step set the runtime validator enforces against, so it cannot disagree
  with the executor about what exists; it also decodes each nested step with
  `models.DecodeSubWorkflowStep`, the loop action's own decoder, rather than a JSON
  round-trip that populates fields the executor drops. **If you think you cannot use it
  because you need containment, you can:** `WalkSteps` hands over a qualified path
  (`steps.<container>.<shape>.<name>`) and the container is its third-from-last segment —
  that objection is written into the header of `sharedoutputs.go` as the reason it wrote
  its own descent, and it was simply false. If you must write one anyway, prove it with a
  fixture in **both** shapes plus one carrying both, asserting the `substeps` half wins;
  a live run cannot prove it (see below).
- **why a green live run is no evidence here, measured 2026-08-08:** ZERO live definitions
  carry `substeps` at any depth, so every fleet run of a blind descent is green and stays
  green. The only two rows that carry it are soft-deleted `multipage-website-builder`
  definitions — i.e. the shape has been used on this fleet before and retired, not never
  used, and nothing stops the next author preferring it.
  ```sql
  SELECT count(*) FROM agent_definitions
  WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config->'workflow' @? '$.** ? (@.substeps != null)';   -- 0 on 2026-08-08
  ```
  The SQL twin of this trap is already recorded under the `continue_on_error` entry
  above ("COALESCE both or you will get a confident 0 from the wrong key") — this entry
  is the Go side of the same fact, plus the precedence, which a COALESCE in the wrong
  order also gets wrong.
- **source:** RFC_012 (d), 2026-08-08. Found only because the council gate's `reuse_agent`
  seat asked whether `sharedoutputs.go`'s walk had been checked against `relaygaps.go` in
  the same package — which walks the same structure through `WalkSteps` on purpose, citing
  `bugs_open/144`'s rule that "a second hand-written descent goes blind in its own
  direction". It had. Fixed in `867037f5a`, where the swap is also shown to be a
  byte-identical no-op over the live fleet (both binaries, one 177-agent export), so the
  fix is provably free and the gap was provably invisible.
- **added:** 2026-08-08, rfc012_await_findings lane

## A `_verification.status='error'` row means the OPPOSITE side of 2026-08-08 — it was a COMPLETED item, it is now a BLOCKED one, and nothing in the row says which era it came from

**footprint:** `site_work_items.result->'_verification'` · `platform/orchestration/actions/complete_work_item_verification.go` (`verifyBeforeComplete`) · `platform/orchestration/actions/discovery_checks/` (every `Verify*Resolved`, `RegisterVerifier`) · `check_empty_sections.go:412` · `check_truncated_component.go:272`

> **⚠ CORRECTED 2026-08-08, hours after this entry was written, by the owner ruling on
> `RFC_017` that the entry itself prompted. The behaviour it describes is now the OLD one.**
> Fail-closed is the default from chassis `v1.0.126x` (the next roll): an error **blocks**
> the completion and routes the item into the attempt machinery. **Both readings are
> therefore live in the data at once** — rows written before the roll mean "completed", rows
> after mean "blocked", and the `status` value is identical. This is the worst shape a fact
> can change in, which is why the entry is corrected in place rather than deleted.
> **Tell them apart by the payload's `fail_open` key** (added by the same change; absent on
> every pre-roll row) **and by the item's own `status`** — an error row that is `complete`
> is the old era, one that is `triaged`/`failed` is the new. Never infer the outcome from
> `status='error'` alone, in either direction.
> **⚠ `fail_open` dates ERROR rows ONLY.** It is written on the error branch and nowhere
> else, so a `verified` or `defect_persists` payload has no era marker in EITHER era — the
> first post-roll row (a `verified`, 18:58:44Z on `v1.0.1268`) duly carries none, and reading
> that absence as "pre-roll" is the second-order version of this same trap. For a non-error
> row, date it by `updated_at` against the roll, never by the payload's shape.
> **LIVE from `v1.0.1268`** (pod-verified both replicas 2026-08-08: `fails closed (RFC_017)`
> = 1, negative `failing open` = 0). The fail-closed branch itself has **not yet executed in
> production** — deployment proven, behaviour not.

**The original trap, which is still exactly why this matters.** Reading a `_verification`
payload, `status='error'` looks like the gate having stopped something: the item is
annotated, the failure is named, the record is honest. **Before the flip the item completed
anyway.** `verifyBeforeComplete` returned `(payload, true)` on the error path by documented
design — only `Resolved:false` blocked a stamp. So an error row was indistinguishable *in
effect* from a `verified` row; the difference was only in what the record said, and
`bugs_closed/032` says so explicitly: *"Item flow is unchanged — the item completes either
way."* Every such row already in the table is one of those.

**Why this fires without a symptom.** ~~Two~~ **FOUR** verifiers return an error on their ambiguous
"target row is absent" case — `check_empty_sections.go:411` (the `page_components` row),
`check_truncated_component.go:271` (the `content_components` row),
`check_content_duplication.go:631` (*"page … no longer exists — cannot distinguish a fix
from content loss"*) and `check_page_canonical_collision.go:382` (*"site … no longer
exists"*) — every one citing the fail-open policy correctly, and that
branch reads as the *cautious* choice — which is exactly how it shipped here twice.
**(Count corrected 2026-08-08 after a council seat asked for the enumeration: the first
version said "two", because I grepped the one spelling I had already read instead of the
class. A fifth, `check_phantom_internal_links_fragments.go:284`, is adjacent.)**
Measured 2026-08-08: those are the **only two verifier errors on the platform, ever** (2 of
11 consultations), and on **both**, the page still declared the slot in `pages.sections`,
so `Resolved:false` was the honest verdict and the defect is still live on two production
pages, stamped `complete` at `attempt_count` 0.

**the check:** never read `status='error'` as an outcome in EITHER direction — ask the item's
own status, and ask whether the page still wants the thing before believing a "legitimately
removed" story:

```sql
-- pre-roll: an error row that is 'complete' is a fail-open, not a catch.
-- post-roll: the same v_status BLOCKS, and fail_open says so on the row itself.
SELECT id, status, attempt_count,
       result->'_verification'->>'status'    AS v_status,
       result->'_verification'->>'fail_open' AS fail_open  -- NULL on every pre-roll row
FROM site_work_items WHERE result ? '_verification';
-- then, for an "absent target" error, 032's disambiguator. sections is an array of BARE
-- STRINGS (jsonb_object_keys errors on it) and is snake_case where slot_name is kebab:
SELECT p.sections::text FROM pages p WHERE p.id = '<page_id>';
```

Two adjacent traps in the same payload. **The denominator is tiny and silence is not
health** — 5 of the 8 registered verifiers have never been consulted, so a low error count
measures how rarely verifiers run, not how reliably. And **`result` is OVERWRITTEN on every
completion attempt** (the complete path and `failUnverifiedCompletion` both write it), so
any count off this column is a floor, not a history — there is no history table.

- **source:** `architecture_review/RFC_017` § "The missing number — MEASURED 2026-08-08" and
  its ✅ DECIDED banner (owner ruling the same day, taken on that measurement); register
  entry `WII-011`; the old fail-open policy is `bugs_closed/032`'s accepted fix.
  `RUNBOOK_page_content_writer_dispatch.md` R8 has the census plus the three ways it lies to
  you.
- **added:** 2026-08-08, bugfix_201_page_content_writer_dispatch lane. **Corrected the same
  day**, by the ruling this entry's own measurement produced — the entry is ~6 hours old and
  its headline fact has already inverted, which is the point rather than an embarrassment:
  the trap is now that BOTH readings are true, of different rows, in one table.

### `ActionInputSpec.Deprecated` cannot carry a renamed SETTING — it resolves the old key's VALUE as a data path, so the alias reads as wired and silently takes the default

**footprint:** `platform/orchestration/datahelpers/action_inputs.go` (`ActionInputSpec.Deprecated`, `ExtractActionInputs` Strategy 3) · `platform/orchestration/datahelpers/config_key_aliases.go` (`ResolveConfigSetting`, `DeprecatedConfigKeys`) · `ActionInputSpec.ConfigKeys`

You have renamed a step-config key and you want the old name to keep working. The spec has
a field called `Deprecated`, documented as *"old config key -> new field name"* with a
deprecation warning. It is the obvious thing to reach for and **for a settings key it is
the wrong one, in the direction that looks right.**

`Deprecated` is honoured in exactly one place — `ExtractActionInputs` Strategy 3 — and what
it does there is `ExtractNestedField(collectedData, config[oldKey])`. It treats the old
key's **value as a dot-path into `collected_data`**. That is correct for the shape it was
built for, a *reference* alias like `"site_id_field": "site_record.site_id"`.

Put a **setting** there — `"check_domain": "content"` — and the runtime looks for a
`collected_data` key called `content`, finds nothing, sets nothing, and the action falls
through to its hardcoded default. No error. No warning (Strategy 3 warns only when the
lookup *succeeds*). `UnknownConfigKeys` reports the key as **recognised**, because
`Deprecated` keys are recognised on purpose. So you end up strictly worse off than before
you declared it: the behaviour is unchanged and the detector has gone quiet about it.

**the check:** ask which side of the extractor the key is on, and pick the matching field.

- Value is a **dot-path resolved against `collected_data`**, action reads it via
  `inputs.Get(...)` → `Deprecated`.
- Value **IS the value**, action reads it via `config["k"].(string)` → **`DeprecatedConfigKeys`**
  (SCR-006), honoured by `datahelpers.ResolveConfigSetting` at the read site. This is also
  the tell for which field the key belongs in generally: `ConfigKeys` and
  `DeprecatedConfigKeys` are the settings pair, `Required`/`Optional` and `Deprecated` are
  the reference pair.

```bash
# does the action read this key straight from config, or through the extractor?
grep -n '"<key>"' platform/orchestration/actions/<action>.go
# ^ grep the KEY NAME, not config["  — a key can reach the action through a helper
#   (GetIntField, resolveAIServiceConfig) and a config[" grep will not see it. That
#   mistake was made in this lane on 2026-08-08 and is logged in WRONG_CALLS.md.
```

**Why this is a landmine and not a §9 pattern:** there is no symptom. The half-landed
`domain` -> `pipeline` rename sat on three actions for months, correct on all three, because
each hardcoded default happened to equal what its config asked for. It only stopped being
correct on 2026-08-04, when two checks that propagate `dctx.Pipeline` joined an agent whose
config asks for `content`, and four work items were filed under `design`. Nothing failed.
Nothing logged. A test asserting current behaviour passes either way.

- **source:** `bugs_open/136`, register SCR-006, docs024_key_docs_latest/bugfix_136_config_key_aliases/
- **added:** 2026-08-08, bugfix_136_config_key_aliases lane

---

### `nav-link-fixer` ALWAYS commits `<domain>/assets/js/snippets.js` to the shared `gqls/sites` repo — even for a site with zero snippets, even for a domain that does not exist
- **footprint:** `nav-link-fixer` (agent), `render_js_snippets_for_site_action.go` (`RenderJSSnippetsForSiteAction`), the `deploy_js_snippets` step, `git_commit` / `GitCommitAction`, repo `gqls/sites`
- **fires when:** you dispatch `nav-link-fixer` at a site — including deliberately, at a scratch/pool/internal site, to exercise something else. It is also reached indirectly: `routeBySurface` gives every `site_component`-surface finding `handler_agent='nav-link-fixer'`, so **any** chrome-surface work item you let reach `build-dispatch-loop` will run it
- **the tell:** none in the handler's own result. The workflow's `complete` step declares `output_fields: ["template_fix_result","rerender_result"]`, so the git result is not in the response the work item stores. The item reads `complete`, the response reads `success:true`, and nothing in it mentions a commit
- **why it is unconditional:** `RenderJSSnippetsForSiteAction` returns `{"files": {"assets/js/snippets.js": …}}` for a site with **no** components (`render_js_snippets_for_site_action.go:86-94` — an empty bundle, deliberately, so the head's `<script src>` does not 404). So there is no zero-snippet path that skips the commit, and `git_commit` defaults `repo_name` to `"sites"` (`git_deployer_actions.go:99-102`)
- **the check:** before dispatching, know what domain will be written and confirm afterwards — the push also triggers the repo's deploy workflow, so it reaches B2:
  ```bash
  gh api repos/gqls/sites/contents/<domain> --jq '.[].path'      # exists? then you wrote there
  gh run list --repo gqls/sites --limit 5 --json displayTitle,createdAt,conclusion
  ```
  Measured 2026-08-08: dispatching it at `pool-ai-agents.internal` (a `status='pool'` site nothing serves) created `pool-ai-agents.internal/assets/js/snippets.js` in `gqls/sites`, commit "Update JS snippets bundle" 15:36:28Z, Actions run 31264883288 green 2 seconds later
- **and, separately:** do **not** predict a no-op from "this site has no templates". In the same run `fix_nav_link_templates` reported `"no header/footer component templates assigned to site"` and updated 0, and `render_site_components` then rendered **both** slots from generic templates anyway — wiping hand-planted chrome HTML. Chrome is every page of the site
- **⚠ `deploy_js_snippets` is a WORKFLOW STEP NAME IN THE DATABASE, not a Go symbol** — `grep -rn deploy_js_snippets --include=*.go` returns **zero**, and that is correct, not evidence the step is gone. Read it where it lives:
  ```sql
  SELECT jsonb_pretty(default_config->'workflow'->'steps'->'deploy_js_snippets')
  FROM agent_definitions WHERE type='nav-link-fixer'
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
  (Noted because the automated landmine-verifier returned NEEDS_HUMAN_REVIEW on exactly this, 2026-08-08: it code-searched a name that only ever existed in config. Line numbers re-checked by hand the same day and are exact — `render_js_snippets_for_site_action.go:86-94` is the `len(components) == 0` early return; `git_deployer_actions.go:~102` is `resolveGitRepoNameDB`, whose documented fallback chain ends at `"sites"`.)
- **source:** 2026-08-08, `bugfix_071_fragment_blindspot` lane, exercising the `dead_fragment_link` verifier; the owner had accepted the side effect in advance, which is the only reason it was harmless
- **added:** 2026-08-08, bugfix_071_fragment_blindspot lane

---

### A literal in a workflow step's `config` is NOT a value — `ExtractActionInputs` resolves multi-segment dot-paths and nothing else, and the error says the field is MISSING
- **footprint:** `platform/orchestration/datahelpers/action_inputs.go` (`ExtractActionInputs`, "Strategy 0"), any `agent_definitions.default_config->workflow->steps->*->config` you hand-author, every action with a registered `ActionInputSpec`
- **fires when:** you write a one-shot or ad-hoc workflow and put a concrete value where the shipped definitions put a path — `"work_item_id": "<a real uuid>"`, `"site_id": "<a real uuid>"`. It is the obvious thing to write, because you know the value and there is nowhere else to put it
- **the tell:** the run fails with `input extraction failed: missing required fields: [work_item_id]` — naming the field that is **sitting in the config you just wrote**. It reads as "you forgot it", so the instinct is to add it again, or to add `work_item_id_field`, rather than to suspect the value's *shape*
- **why:** Strategy 0 resolves `config[field]` only `if strings.Contains(pathStr, ".")` (`action_inputs.go:472-488`); a value with no dot is skipped entirely and never becomes an input. Strategies 1/2 then look for a key of that name inside `collectedData` — which, for a value that exists only in your config, is not there
- **the check:** put the value in the message and reference it — for a `scheduled_tasks`-fired one-shot, `input_data` becomes `body.input_data`, so:
  ```
  scheduled_tasks.input_data  = {"work_item_id": "<uuid>", "handler_result": {...}}
  step config                 = {"work_item_id": "input_data.work_item_id", "result": "input_data.handler_result"}
  ```
  A single-segment path (`"work_item_id": "work_item_id"`) is also skipped — it needs the dot. If a config value must be a literal, the action has to read `params.StepConfig.Config` directly, as `fail_work_item` does for `error_message`
- **source:** 2026-08-08, `bugfix_071_fragment_blindspot` lane — cost one full dispatch round while probing `complete_work_item`
- **added:** 2026-08-08, bugfix_071_fragment_blindspot lane

---

### A completed `site_work_items` row keeps the `error` text from an earlier REFUSED completion — and `result._verification` is overwritten, so the refusal leaves no durable trace at all
- **footprint:** `site_work_items.error`, `site_work_items.result->'_verification'`, `CompleteWorkItemAction` (`load_work_item_actions.go:941-949`), `failUnverifiedCompletion` (`complete_work_item_verification.go:213-240`)
- **fires when:** you audit work items by `error IS NOT NULL`, or read one item's `error` to judge whether it is really fixed, or try to count how often the verifier registry has refused a completion
- **the tell:** none — the row is internally contradictory and every column is individually plausible. Measured 2026-08-08 on one item, minutes apart: `status='complete'`, `result->'_verification'->>'status' = 'verified'`, and `error = 'completion blocked: post-fix verification found the defect still present: …'` **simultaneously.** The success UPDATE writes `status`, `result`, `completed_at`, `handled_by` and never touches `error`
- **the second half, which is worse:** each completion attempt REPLACES `result._verification`. A refusal's `defect_persists` verdict is destroyed by the next attempt that succeeds. So the fleet census this estate quotes — `SELECT … WHERE result ? '_verification'` — counts **surviving** verdicts, not verifications performed, and systematically under-counts exactly the refusals the registry exists to produce
- **the check:** never read `error` alone. Read it beside `status` and `completed_at`, and treat a non-empty `error` on a `complete` row as history, not state:
  ```sql
  SELECT status, attempt_count, completed_at,
         result->'_verification'->>'status' AS verdict, left(error,120) AS error_history
  FROM site_work_items WHERE id = '<uuid>';
  ```
  To count refusals, do **not** ask `result`; the only contemporaneous records are the pod log (`"CompleteWorkItemAction: completion blocked"`, which rotates) and `error` (stale-prone, per above)
- **source:** 2026-08-08, `bugfix_071_fragment_blindspot` lane, watching one item go detected → refused → triaged → complete
- **added:** 2026-08-08, bugfix_071_fragment_blindspot lane

---

### Deleting a whole domain DIRECTORY from `gqls/sites` deploys NOTHING — the run goes green, prints one WARNING, and the bucket copy is orphaned for ever
- **footprint:** `.github/workflows/deploy-to-b2.yml` in the `gqls/sites` repo (the `Get changed domains` step and the `for domain in …; do if [ -d "$domain" ]` guard in `Sync to B2`), `b2://portfolio-sites/<domain>`, any retirement/removal of a site directory or the last file under one
- **fires when:** you remove a domain's whole directory — retiring a site, or (as here) deleting the single junk file that was the directory's only content, which removes the tree with it. It is the obvious way to un-publish something and it is the one case the sync cannot express
- **the tell:** ONE line, in a green run, in the middle of a step whose other branch is chatty: `WARNING: <domain> in changed set but no directory — skipped`. The domain IS correctly detected as changed (`Changed domains: <domain>`), so the "did it notice?" check passes and reads like success. There is no error, no failed step, and the repo is genuinely correct — only the bucket is wrong
- **why:** `--delete` removes what is missing from the SOURCE DIRECTORY, and it only ever runs for domains the loop actually enters. `git diff --name-only` still lists the deleted paths, so the domain lands in the changed set; `[ -d "$domain" ]` is then false, and the `else` branch skips it. **There is no code path that deletes a bucket prefix.** Measured 2026-08-08: run `31266031734`, commit `0709f572b` removing `pool-ai-agents.internal/assets/js/snippets.js` — conclusion `success`, zero `Syncing … to B2` lines, the WARNING above, and the object still in the bucket
- **the check:** after removing a directory, **grep the run for the WARNING, not for failure** — and never conclude from a green run that the bucket followed:
  ```bash
  gh run view <id> --repo gqls/sites --log | grep -E "Changed domains:|Syncing .* to B2|no directory — skipped"
  ```
  A domain in `Changed domains:` with **no** matching `Syncing` line is an orphaned prefix, every time
- **what to use instead:** there is no repo-side fix — emptying the directory cannot reach zero files, because a directory with no files is not a directory. Removing the bucket prefix needs the **B2 CLI with the `B2_APPLICATION_KEY_ID` / `B2_APPLICATION_KEY` credentials**, which live only as GitHub secrets: `b2 rm --recursive "b2://portfolio-sites/<domain>"`. Budget that before deleting a directory, or leave the directory in place with the content you actually want served
- **⚠ this is NOT the `--skip-newer` landmine above it, and the check for that one does not catch this one.** That entry is about a file whose bucket copy is newer being silently skipped **inside** a sync; this is the sync **never running for the domain at all**. Its remedy (`gh run rerun`) is useless here — a fresh checkout still has no directory
- **⚠ the automated landmine-verifier CANNOT check this entry, or its `--skip-newer` neighbour** — it returned NEEDS_HUMAN_REVIEW on 2026-08-08 saying "the `code_symbols` index does not cover `.github/workflows/deploy-to-b2.yml` (all six lookups returned 0 rows)". That is correct and it is a property of the **verifier**, not a doubt about the entry: the file lives in a **different repository** (`gqls/sites`), which the index does not ingest. **Read that verdict as "unverifiable here", never as "unconfirmed".** This entry is human-confirmed first-hand — the workflow source was read via `gh api repos/gqls/sites/contents/.github/workflows/deploy-to-b2.yml --jq .content | base64 -d`, and the behaviour observed in run `31266031734`'s own log. Any landmine footprinted on another repo's file inherits this
- **source:** 2026-08-08, `bugfix_071_fragment_blindspot` lane, cleaning up after the `nav-link-fixer` commit landmine above; the removal was correct in git and inert at the origin
- **added:** 2026-08-08, bugfix_071_fragment_blindspot lane

---

## After a whole-fleet release the fleet is MIXED, and "the fleet is on vX" is false for hours — the pod that runs YOUR action may still be on the old image

- **footprint:** `make release`, `make redeploy-agents`, `kubectl get pods -n ai-persona-system`,
  any post-roll verification of a Go change

**The trap.** `make release redeploy-agents` rolls the long-lived **Deployments**
(`agent-chassis`, `business-intel`, `vet-intel`, the adapters). It does **not** replace
the **spawned per-agent pods** — `agent-page-rebuild-*`, `agent-build-dispatch-loop-*`,
`agent-landmine-verifier-*`, `agent-diagnose-*` and friends — which keep running on
whatever image they were created with. Measured 2026-08-08, ~15 minutes after a
release: **20 pods on `v1.0.1264`, 5 on `v1.0.1266`.** A `kubectl get pods | head`
lands on either group depending on sort order, and both answers look authoritative.

So the sentence "the fleet is on `v1.0.1266`" was false, and the sentence "the fleet is
still on `v1.0.1264`" was equally false. **Neither is a state the cluster was in.**

**Why it bites even when you follow the pod-grep rule.** Grepping "a chassis pod" is
not enough: if you happen to exec into a *spawned* pod you will prove your fix is
ABSENT and conclude the release failed; if your action actually executes in a spawned
pod, you can prove the fix PRESENT on the deployment and still run old code. The
grep is right and the pod is wrong, which is the worst shape — a correct method
producing a confident wrong answer.

**the check:** ask for the distribution, never a sample, and then name the pod that
will run *your* action:

```bash
# 1. what is actually out there — a mixture is the expected state, not an anomaly
kubectl get pods -n ai-persona-system \
  -o jsonpath='{range .items[*]}{.spec.containers[0].image}{"\n"}{end}' | sort | uniq -c | sort -rn

# 2. grep the pod that consumes YOUR topic, with a POSITIVE and a NEGATIVE marker
kubectl exec -n ai-persona-system <that-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<string your change ADDED>";
   strings /app/agent-chassis | grep -c "<string your change REMOVED>"'   # expect >=1 then 0
```

A pod still on the old tag is the **free control** this rule otherwise lacks: run the
same two greps there and you should get the exact inverse (`0` then `>=1`). If you get
the same answer on both pods, your marker is wrong and neither result means anything.
Confirmed working on 2026-08-08 (`bugs_open/219`): `v1.0.1266` replicas gave `1 / 0`,
a `v1.0.1264` pod gave `0 / 1`.

⚠ **And do not wait for the mixture to clear before dispatching.** The spawned pods are
per-job and age out on their own; the ones you saw may be `Succeeded` already. What
matters is the image of the pod that will pick up your message, not fleet uniformity.

---

### A voice_tells retraction cannot tell "the prose was fixed" from "we read nothing" — and one of them closes a live human-review row
- **footprint:** `platform/orchestration/actions/revalidate_voice_tells.go`, `platform/orchestration/actions/discovery_checks/check_voice_tells.go`, `ScanVoiceTells`, `VoicePageScan`, `site_work_items.item_type='voice_tells'`

`ScanVoiceTells` returns an empty `Findings` slice in **three** unrelated situations, and only one
of them means the copy improved:

1. components were examined and are clean → **the prose was fixed**
2. **nothing was read at all** — page deleted, not `active`/`deployed`, or no rendered components
3. **only human-LOCKED components were read** — the emit side has always skipped `locked_at IS NOT NULL`

**the check:** never branch on `len(scan.Findings) == 0` alone. Assert `scan.ComponentsExamined > 0`
first, and treat `scan.ComponentsSkippedLocked > 0` as unanswerable rather than clean — a page whose
only offending component was pinned by a human scans clean while the copy is untouched. The verdict
ladder in `voiceTellsVerdict` is ordered so `still_holds` is answered before the locked arms; reorder
it and a page that still trips the gate gets reported as unreadable. Each of the three guards has a
mutation test (`if false` → distinct failure), so a future edit that drops one fails loudly.

### A `voice_tells` item can retract with the copy completely unchanged, because the gate is a moving standard
- **footprint:** `platform/orchestration/actions/revalidate_voice_tells.go`, `site_specs` aspect `voice`, `voice_gate`, `datahelpers.VoiceGate`, `resolution_path='auto:revalidated'`

The revalidator re-runs the site's **current** `voice_gate` thresholds. If a site loosens them — or
disables `voice_gate` entirely and later re-enables it looser — prose that failed on 17 July passes
today and the item closes, **with no edit to the page at all**. The stored evidence records
`components_examined` and the page name; it does **not** record which thresholds were in force, so a
`resolved` stamp read later cannot distinguish "rewritten" from "standard relaxed".

**the check:** before citing `auto:revalidated` on a `voice_tells` row as evidence that copy was
improved, compare the page's churn against the item's age —
```sql
SELECT swi.id, swi.created_at, max(pc.updated_at) AS last_component_edit
FROM site_work_items swi
JOIN page_components pc ON pc.page_id = (swi.spec->>'page_id')::uuid
WHERE swi.item_type='voice_tells' AND swi.resolution_path='auto:revalidated'
GROUP BY 1,2;
```
`last_component_edit` earlier than `created_at` means the page was never touched and the *standard*
moved, not the prose. (The gate being **absent** is already refused — that arm returns `unknown`,
because opting out of an audit is not evidence of passing it.)

### A grep needle short enough to be convenient is long enough to be someone else's string
- **footprint:** `kubectl exec … strings /app/agent-chassis | grep -c`, any pre-roll deploy baseline

Taking a pre-roll baseline for an additive Go change on 2026-08-08, I grepped `no longer exists` — a
fragment of a new error string — and got **6 on both replicas of a build that did not contain the
change at all**. Had the post-roll reading been 6 or 7 I would have had no idea whether it shipped.
The whole value of an additive proof is the dated **0 → 1** transition, and a needle with a non-zero
baseline destroys it silently.

**the check:** every needle must be **verified to return 0 on a build that predates your commit**,
in the same command that records the baseline — a needle you did not disconfirm is not a control.
Prefer a whole distinctive clause (`opting out is not evidence the copy was fixed`) over a fragment,
keep it **ASCII-only** (an em-dash mangles crossing `exec -- sh -c` and returns a 0 that reads like a
failed deploy), and always carry a positive control string that is non-zero on the *current* binary
so a post-roll 0 is distinguishable from a broken probe.

### `cmd | head -N && echo "OK"` reports success on a FAILED command
- **footprint:** any `go build`/`go test` wrapped in a pipeline, `${PIPESTATUS[0]}`

`TMPDIR=… go build ./platform/... 2>&1 | head -20 && echo "=== BUILD OK ==="` printed **BUILD OK**
under a real compile error on 2026-08-08. `&&` tests the exit status of the **last** command in the
pipeline — `head` — which succeeds regardless. The truncation compounds it: `head -10` cut the error
list down to one line, and I then reasoned confidently from the survivor about which session had
broken the tree.

**the check:** gate on the pipeline's real status, and never let a truncating pager decide what you
diagnose from —
```bash
TMPDIR=/home/ant/.cache/buildtmp go build ./platform/... 2>&1 | head -40; echo "exit: ${PIPESTATUS[0]}"
```
On a shared tree, also confirm a failure is actually yours before saying so: another session's
mid-write can leave the tree uncompilable for seconds, and the honest test is
`git archive HEAD` + your own files, not the working tree.

### The `page_build_failed` park must be inserted RAW — routed through `insertWorkItem` it is stillborn and bounds nothing
- **footprint:** `platform/orchestration/actions/page_build_failure_guard.go`, `parkPageBuildFailure`, `insertWorkItem`, `writeWorkItem`
- **fires when:** a refactor "tidies" the park insert in `parkPageBuildFailure` onto the shared `insertWorkItem`/`writeWorkItem` helper — the natural move, since hand-rolled `INSERT INTO site_work_items` is exactly what that helper exists to retire, and a council seat has objected to hand-rolled copies of it before.
- **the tell:** none at review time and none in tests that mock the insert — the two-strike block counts terminal (`complete`,`failed`) predecessors on `(site_id, item_key)`, and the park's key `needs_page:<name>` has BY CONSTRUCTION 2+ terminal predecessors when the park fires (they are the failed builds that earned it, dishonestly `complete` — that dishonesty is bugs_open/210 itself). The park is born `status='unresolved'`: TERMINAL, outside `idx_swi_dedup`'s predicate, so it holds no page slot, blocks no producer, and the retry loop it exists to stop runs on. Everything still "works".
- **the check:** after any change to how the park is written, run `TestUpdatePageStatus_ThirdRefusalParksThePage` (it pins the raw statement and the `needs_human_review` birth status), and live: park a page, then `SELECT status FROM site_work_items WHERE item_type='page_build_failed' ORDER BY created_at DESC LIMIT 1` — anything but `needs_human_review` means the bound is gone.

### An `insertWorkItem` false return on a `needs_page:<name>` key may be a PARKED page, not a duplicate request
- **footprint:** `site_work_items` where `item_key LIKE 'needs_page:%'`, `needs_tool_recreation`, `needs_content_page`, `WriteBuildItemsAction`, `page_build_failed`
- **fires when:** your lane emits a build/recreation item for a page and the emitter silently no-ops (inserted=false / 0 rows). The ordinary reading — "an open request already covers this" — is also the reading when bugs_open/210's guard has PARKED the page after 3 failed content generations: the open `page_build_failed` row holds the same `(site_id, 'needs_page:<name>')` dedup slot, deliberately, so all automatic producers stop paying for retries.
- **the check:** before diagnosing your emitter, `SELECT item_type, status, spec->>'skip_reason' FROM site_work_items WHERE site_id=$1 AND item_key='needs_page:<name>' AND status='needs_human_review'` — a `page_build_failed` hit means the page's content generation is failing repeatedly and a human must close the park (or fix the cause and let the next successful deploy auto-close it) before any producer can claim the slot again.

### `toolgolden.py` drives each page by scaling the page's OWN markup defaults — a cross-page compare feeds the two sides DIFFERENT inputs and reports the difference as behaviour
- **footprint:** `docs024_key_docs_latest/loancalculator_couk/toolgolden.py`, `DRIVE_JS`, `compare_rebuilt.py`, `GOLDEN_*` files, `acceptance/`
- **fires when:** anyone points the toolgolden harness at TWO different implementations of a tool (original vs rebuild, old vs new) expecting a differential arithmetic verdict. DRIVE_JS computes every driven value as `parseFloat(e.getAttribute('value')) × factor` — the page under test supplies its own stimulus — and fills a fixed 1000 into any numeric field with no `value` attribute; a `step>=1` attribute additionally rounds the driven value. Two correct calculators with different example defaults (or one missing `value` attrs, or one added `step="1"`) then "DIVERGE" on every vector.
- **the tell:** none in the report — the diff lines look exactly like arithmetic divergence (`£1,390 -> £1,169.18`), and the numbers are each internally consistent. On 2026-08-08 this produced a committed, owner-reported claim that 6 of 12 mortgagecalculator rebuilds compute wrong formulas; all 6 dissolved (rebuilt repayment answers £1,389.58 vs original £1,390 on identical hand-driven inputs). Absurd magnitudes are the loudest hint: a "1200% yield" is the 1000-into-every-empty-field branch.
- **the check:** before believing any cross-page divergence, hand-drive ONE case: CDP-fill both pages with the SAME literal values by id, press, compare; and check one golden value against an independent calculation. The golden already records the absolute fill plan (`sel`/`action`/`value` per control) — a comparator that REPLAYS that plan on the other page is the fix, and until one exists toolgolden certifies only page-against-itself regressions, never A-vs-B arithmetic.

### The provocation calibration fixture's `body` is copied from production's `detail_body` — comparing same-named columns reports 7 of 9 fixtures as diverged when they are byte-identical
- **footprint:** `provocations` (columns `body`, `detail_body`, `teaser`, `card_desc`, `headline`), `calibration.vonc.com`, `docs/agent_docs/sql_for_agents/319_provocation_gate_calibration_harness.sql`, `gate_provocation`, `provocation-gate-calibration`
- **fires when:** you act on HANDOFF_2026-08-08's §6 instruction — "re-copy the must-approve half if production text has changed since" — and check it the obvious way, `c.body = p.body` joined on slug. `provocations` has FIVE prose columns and production splits across them by vintage: the 8 owner-written `approved` rows (updated 2026-07-31) carry their prose in **`detail_body`** with `body` empty, while the 7 newer `draft` rows (2026-08-07) populate `body` AND `detail_body`. The fixture copy flattens whichever is present into `body`. So the same-name comparison returns `body_same=false` for 7 of 9 rows with `prod_len=0`, and the two rows it calls equal are the two coincidences (one genuinely uses `body`, one is empty in both).
- **the tell:** it reads as something far more alarming than a column mismatch — **"production's bodies have been emptied"**, 7 rows of `prod_len 0` where the fixture has 400-600 characters. Both available responses are damaging: re-copy the fixture from a pool you believe is drained (destroying a fixture that was correct, and by ruling 3 you must not retype it), or stop the lane to chase a production incident that never happened. Nothing about the output looks like a schema misread. Verified 2026-08-08: all 9 fixtures match `COALESCE(NULLIF(detail_body,''), body)` on md5, exactly.
- **the check:** never compare one prose column on this table — enumerate all five before concluding anything, `SELECT slug, length(COALESCE(teaser,'')), length(COALESCE(card_desc,'')), length(COALESCE(headline,'')), length(COALESCE(detail_body,'')), length(COALESCE(body,'')) FROM provocations WHERE domain='vonc.com'`, and compare the fixture against the precedence **the gate itself uses** — `COALESCE(NULLIF(body,''), COALESCE(detail_body,''))`, `provocation_gate_action.go:663` in `loadGateCandidates` — on md5. The genuine zero-body row is `group-chats-replaced-friendship`, which is empty in **both** columns — that one IS a real pool defect (it fails the gate as `body_too_short`) and by owner ruling 3 the framework must generate the body, not a session.
- **CORRECTED 2026-08-08 (same day, by the landmine-verifier), and the error is instructive:** this bullet and the tell above originally prescribed `COALESCE(NULLIF(detail_body,''), body)` — **`detail_body` first, which is backwards.** The gate reads `body` first. My md5 check passed 9 of 9 and could not have failed: no row in the must-approve fixture has BOTH columns populated, so on that set the two precedences are indistinguishable. They disagree on exactly the rows the fixture does not contain — all 7 newer `draft` rows (2026-08-07 vintage), where the gate judges the ~400-char `body` and my formula returned the ~780-char `detail_body`. **Cost if uncorrected:** anyone building or checking a fixture from the newer vintage would read a false mismatch, i.e. the very failure this entry exists to prevent, one vintage along. The lesson is the one this estate keeps re-learning — a comparative claim needs an input where the two candidates DIFFER, and a check that cannot come out otherwise is not evidence.

### Running `landmines-sync.py --apply` before `landmines-verify-dispatch.sh` consumes the "new entry" status — the verifier then fires for NOTHING and your entry is never checked
- **footprint:** `scripts/landmines-sync.py` (`--apply`, its `new`/`changed` detection, the `NEEDS_VERIFICATION:` lines), `scripts/landmines-verify-dispatch.sh`, `scripts/trigger-landmine-verifier.sh`, `doc_notes` rows with `categories ? 'landmine-verification'`, RFC_005 §3.2
- **fires when:** you append a landmine and follow CLAUDE.md's own instruction — *"After you append, run `./scripts/landmines-sync.py --apply`"* — and then run the verify dispatch. `--apply` is what WRITES the `doc_notes` rows, and "new" means "not already in `doc_notes`". So the plain sync marks your entry as seen; the wrapper re-runs sync, finds nothing new, and dispatches zero verifiers. The two documented steps are individually correct and silently incompatible **in the order CLAUDE.md implies**.
- **the tell:** the wrapper prints `Nothing needs verification (no new or changed entries this run)`, which is exactly what a correctly-verified corpus prints. The `doc_notes` rows ARE present and the sync's own `--check` passes, so every downstream signal says done — the only missing artefact is the verification verdict, and nothing lists its absence. Hit on 2026-08-08: sync reported `NEEDS_VERIFICATION:LANDMINES.md#the-provocation-calibration-fixture…`, the dispatch two minutes later reported nothing to do, and no verifier had run.
- **the check:** run `./scripts/landmines-verify-dispatch.sh` **INSTEAD OF** `landmines-sync.py --apply`, never after it — the wrapper does the sync itself (its header says so). If you have already applied, do not re-sync hoping to re-arm it: fire the entry by hand, `./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>'`, taking the slug from the `NEEDS_VERIFICATION:` line the first sync printed (keep that output — it is the only place the slug appears). Confirm a verdict actually landed: `SELECT created_at, left(body,120) FROM doc_notes WHERE subject_key='LANDMINES.md#<slug>' AND categories ? 'landmine-verification';`

### A `landmine-verification` verdict of STALE / "does not exist" is NOT evidence against your entry — the index it consults is 100% Go, and 81% of all footprints are not
- **footprint:** `doc_notes` rows with `categories ? 'landmine-verification'`, `code_symbols`, `landmine-verifier`, `scripts/trigger-landmine-verifier.sh`, `scripts/landmines-verify-dispatch.sh`, `diagnose_code_lookup`, RFC_005 §3.2, `bugs_open/223`
- **fires when:** you read a verdict on any entry whose footprint names a **script, migration, SQL file, table, column, command, config value or agent type** — i.e. most of this file. `code_symbols` holds **5,755 symbols across 668 files, all `.go`** (measured 2026-08-08: the extension census returns exactly ONE row). A non-Go footprint therefore returns 0 rows from both halves of the repaired lookup, and the verifier narrates that structural absence as non-existence. Measured: **1,116 of 1,371 footprint rows (81%), spanning 284 of 288 entries**, can never resolve. Distinct from `bugs_closed/163` (the Go-symbol half, genuinely fixed and live) — no query form can fix a row class the index never ingested.
- **the tell:** there isn't one in the verdict — it reads as a diligent negative, in the same confident voice as its true findings, and a **mixed** footprint list is worst because the confirmed Go half makes the whole verdict look checked. The 2026-08-08 case is self-refuting and still landed: a STALE verdict asserting that the `landmine-verification` category "does not exist anywhere in the indexed codebase" was written **into that category** (row 32 of 32) by the very script chain it says has no footprint. **Do not delete or downgrade an entry on this evidence** — STALE is exactly the signal that argues for removing a correct trap.
- **CORRECTED 2026-08-08 (same hour, by the verifier itself) — the blind spot is deterministic, the CONCLUSION is not, and that cuts both ways.** Given identical 0-row lookups, four verdicts in one hour reached four different conclusions: one flat `STALE` ("the entire described workflow has no footprint" — false), one hedge ("or has been removed"), one correct abstention ("cannot be mechanically verified"), and one that reasoned about the blindness correctly and confirmed this very entry as `STILL_VALID`. **So you cannot predict which you got** — and the corollary is the one to carry: a `STILL_VALID` on a non-Go-footprint entry is **NOT evidence FOR the entry either**, because the footprints went equally unchecked. For 284 of 288 entries only the prose reasoning carries signal, in both directions. Full account + the measured census: `bugs_open/223`.
- **the check:** before believing any such verdict, ask whether the footprint is a Go symbol at all. If not, the verdict is silent, not negative — verify by the footprint's own kind instead: `git ls-files <path>` for a repo file, `information_schema.columns` for a table/column, a row lookup for an `agent_definitions.type` or a `domain` value. And read the verdict rather than a summary of it: `SELECT created_at, left(body,300) FROM doc_notes WHERE categories ? 'landmine-verification' AND subject_key LIKE '%<slug>%' ORDER BY created_at DESC LIMIT 1;`. **The corollary is the useful half: the verifier IS trustworthy on Go footprints** — it caught a genuine backwards column-precedence error in an entry the same day, by reading the function my own passing check had never opened.

### `load_page_record` resolves the NAME before the ID — wiring a correct page_id into a step config changes nothing while any page_name resolves

- **footprint:** `platform/orchestration/actions/load_page_record_action.go` (`LoadPageRecordAction`, `LoadPageRecordInputSpec`, `authoritative_page_id`), `agent_definitions` rows `page-build-handler` / `tool-recreation-handler` (`load_page_record` step config), `build-dispatch-loop` (`call_handler.input_mapping`, `page_id?`), `site_work_items.page_id`
- **fires when:** you route a work item at a specific page by supplying its id — a new cross-page check, a dispatcher mapping, a handler step config — and expect the id to decide which page loads. It does not: the action's documented priority is `page_name` first, `page_id` only when the name is empty or a non-page marker, AND the empty-name branch actively re-fills the name from `input_data.spec.page_name` and three other fallback paths before the id is consulted. A config carrying both plain keys therefore loads the NAMED page, whatever the id says
- **the tell:** none at the failure — the wrong page loads successfully, builds successfully, deploys successfully, and every status is green. `bugs_open/220` ran this loop for six days: the dispatch honoured `spec.page_name` (the page CONTAINING a broken link), rebuilt it, completed the item, and the never-deployed TARGET the item's `page_id` column named stayed a live 404, re-detected and re-dispatched each discovery pass
- **the check:** if the id must win, map it into `authoritative_page_id` (added 2026-08-08, mig 340) — when that input resolves to a valid uuid the lookup is by id and the name inputs are ignored; malformed → loud error, absent → the old behaviour untouched. If you instead find yourself passing `page_id` as the plain optional key and the name "happens to be empty", you are one spec-shape change away from this bug — the fallback paths mean "empty" is a property of four locations, not one
- **source:** 2026-08-08, `bugs_open/220` (`docs024_key_docs_latest/bugfix_220_unbuilt_link_dispatch/`, register WII-012). Found because the bug file's own fix candidate 1 ("map page_id and prefer it") turned out to be insufficient as written — the priority order defeated it, which nothing in the mapping or the item shape reveals
- **added:** 2026-08-08, bugfix_220 lane

### A behavioural golden certifies a calculator that has ALWAYS been wrong — and on this estate the golden is what "we check the calculators" means
- **footprint:** `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/toolgolden.py`, `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/golden_compare_post.py`, `acceptance/GOLDEN_2026-08-05_prechange.json`, `tool_acceptance`, `computed_values`, `--emit-criteria`
- **fires when:** you are asked whether a site's calculators produce correct output, and reach for the golden — which is the only calculator-checking machinery this estate has, is well built, and answers a DIFFERENT question. It records what each tool answered on a given day and re-runs it. If a tool has been wrong since it was written, the golden captured the wrong number faithfully and every comparison since has confirmed it. Tier-2 `tool_acceptance` cannot help: its own header says its static checks "CONFIRM, never refute".
- **the tell:** there is none in the output — a green `golden_compare_post.py` looks exactly the same whether the tool is right or has been wrong for a year, and the word "golden" reads as a correctness baseline. Measured 2026-08-08 on loanandmortgagecalculator.co.uk: an independent oracle found **27 failing boundary vectors across 8 of 23 tools**, including a stamp-duty calculator running a tax rule that expired 16 months earlier (£5,000 under-quote) and six loan calculators that print `£NaN` or a stale answer at 0% — all of them green in the golden, all of them green since it was recorded (`bugs_open/224`, `bugs_open/225`).
- **the check:** before repeating "the calculators are checked", ask which question the check answers. For correctness you need an oracle computed from the DEFINITION — `oracle.py` / `oracles.py` in the loanandmortgagecalculator lane, plus `python3 oracle.py --selftest-parse` and `--mutate expectation|crosstool` in the same session. **And never run `--emit-criteria` from a tool you have not oracle-checked**: it writes the tool's current answers into the platform's acceptance record as the expected ones, which pins a wrong answer in and then defends it.

### An oracle authored from the page's own `<script>` agrees with the bug — and finding the input ids is one line away from reading the arithmetic
- **footprint:** `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/oracle.py`, `oracles.py`, `inventory.py`, `oracle_driver.py`
- **fires when:** you write any independent checker for a page you can read the source of. The oracle legitimately needs to know WHICH box takes the principal and WHICH element holds the answer, and the fastest way to learn that is to open `calculateLoan()` and see what it touches — at which point the line below tells you what the rate is divided by, and your "independent" expectation is the same claim written twice.
- **the tell:** none, and it is worse than none — a contaminated oracle produces a **clean green run**, which is the most reassuring output the harness can emit. It will agree with every defect it was transcribed from and disagree with nothing.
- **the check:** get the interface from the LABELS, not the script: `python3 inventory.py --out /tmp/inv.json` reports the visible `<label for=…>` bound to each control, the button text and the caption above each result box, which is the site's own claim about what each number means. Author the spec from that, compute from the published definition, and **open the calculation body only AFTER a check has failed** — that ordering makes contamination impossible rather than merely unlikely, and it is what let the 2026-08-08 run find both defect families before any of that source was read. If a label and the arithmetic disagree, that disagreement is the finding.

### A calculator gated `if (rate > 0)` writes NOTHING at 0% and the previous answer stays on screen — and comparing against a primed reading MISSES it
- **footprint:** `loans/standard-calc.html`, `loans/settlement-calculator.html`, `loans/car-finance-calculator.html`, `loans/consolidation.html` (sites repo, loanandmortgagecalculator.co.uk), `oracle.py` `run_determinism`, `case(prime=…)`
- **fires when:** you test a boundary input that a tool's guard rejects. The guard makes the handler return before it touches the DOM, so the page keeps displaying the answer to the previous inputs — no error, no blank, no zero, just a confident wrong number that a user has no way to distinguish from a fresh one. Three of this site's tools do it; three others print `£NaN` from the same missing zero-rate branch (`bugs_open/224`).
- **the tell:** the reading is a plausible, well-formatted money figure, so a value comparison reports an ordinary numeric mismatch and you go looking for an arithmetic error that is not there. **The obvious detector does not work either:** priming with a known vector and checking whether the reading changed misses it, because the stale figure is whatever the last ACCEPTED vector produced — including an intermediate state created halfway through typing the new one. Measured: on `standard-calc` the stale figure was £143.47, the answer to (£10,000, 12%, 10y), a combination never deliberately entered and never recorded by the harness.
- **the check:** drive the SAME final vector by TWO different routes and compare the readings — they must agree whatever the tool computes, so this needs no oracle, no formula and no sight of the source. `oracle.py`'s `determinism` spec does it; the 2026-08-08 output is the statement to quote, because it is legible to anyone: *"the SAME final inputs give '£143.47' by one route and '£429.81' by another"*. Related but distinct: a boundary vector is only a test where a BROKEN implementation gives a DIFFERENT answer — `consolidation` passed a 0%-APR-*debt* vector because its guard returns 0 and 0 is correct there; only a 0% *new loan* exposes it, where 0 means a £0.00 monthly payment.

### A `090` diagnosis run on a NON-GO artefact completes as a SUCCESS with no verdict — the code index holds one repo and one extension, so every site's own HTML/JS is invisible to it
- **footprint:** `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh`, `diagnose-agent`, `diagnose-dispatch-loop`, `diagnosis_artifacts`, `code_symbols`, `diagnose_code_lookup`, `site_work_items` (`item_type='needs_diagnosis'`), CLAUDE.md "Diagnosis before debugging", OWNER RULING 2026-07-31
- **fires when:** you follow CLAUDE.md's own norm — *"a `bugs_open/` file asserting a cross-cutting or structural root cause is not 'filed' until it has been through the loop"* — for a bug in anything that is not Go. Every hand-built or adopted site's inline `<script>`, every `assets/js/*.js`, every Python harness in `docs/agent_docs/`, every SQL file: the index is **one repo (`gqls/agentchassis`) and one extension (`.go`), 5,755 symbols**, measured 2026-08-08. The agent cannot open the file your symptom names.
- **the tell:** **there isn't one, and the absence points the wrong way.** The orchestration reaches `current_step='complete'`, `status='COMPLETED'`; the work item reaches `status='complete'`; `diagnosis_artifacts` fills with `bundle` rows so it plainly did work. What is missing is only the verdict artifact and the `doc_notes` row — and nothing lists their absence. Measured 2026-08-08 on correlation `3e18a949-8732-4603-b19b-f0c159860fa5`: 5 bundles, ~9 minutes, no verdict; the bundles it fetched were `page_sections` rows, i.e. the DB half of the symptom, because the JS half was unreachable. **Do not read this as "the loop looked and found nothing"**, and do not retry it hoping for a verdict — it is the same Go-only index that makes the landmine verifier narrate unindexed footprints as non-existent (`bugs_open/223`); one index, two consumers, and in both the silence wears a finding's clothes.
- **the check:** before firing `090`, ask whether the symptom's evidence is Go in `gqls/agentchassis` — `SELECT count(*) FROM code_symbols WHERE path LIKE '%<your file>%';` must be non-zero, and `SELECT DISTINCT repo FROM code_symbols;` tells you the ceiling in one query. If it is zero, the loop can still read the **live DB** (name tables and rows in the symptom and it will fetch them), but it cannot read your code — so take CLAUDE.md's stated escape hatch **explicitly**: state in the bug file why you substituted equivalent first-hand verification, and make that verification something that does not depend on having read the right file (reproduce it against the live artefact; prefer a check like "same inputs by two routes give two answers", which needs no source at all). Confirm what your own run actually produced rather than assuming: `SELECT kind, count(*) FROM diagnosis_artifacts WHERE correlation_id LIKE '<run corr>%' GROUP BY kind;` — **all-`bundle` and no verdict is the signature.**

### toolgolden presses ONE button per page — a golden of a MULTI-CALCULATOR page records the other calculators' outputs in their UNPRESSED state, and a cross-implementation compare then convicts whichever side DID compute
- **footprint:** `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/toolgolden.py` (`PRESS_JS` — `all.find(...) || all[0]`, one click), every `GOLDEN_*.json` captured from a page with more than one Calculate button, `mortgagecalculator_couk_adoption/acceptance/compare_rebuilt.py`
- **fires when:** you read a golden-vs-anything diff for a page hosting TWO tools (yield + LTV on `investor.html`, limits + roll-up projection on `equity-release.html`). PRESS_JS deliberately presses one button — the first with an onclick, else the first enabled — so the second calculator's result ids are captured at their placeholder values (`0%`, `£0`) with the inputs already driven. Any implementation that computes on input, or on a shared button, then "diverges" from the golden — showing a computed number where the golden holds the placeholder.
- **the tell:** the golden side is a suspiciously inert `0%`/`£0` across EVERY vector while the other side moves plausibly with the inputs — and it reads as "the original is broken" or "the rebuild invents numbers". Measured 2026-08-08 (mortgagecalculator.co.uk): golden `ltvResult` 0% on all four vectors vs rebuilt 75.0%/35.9% — and 225000/300000 IS 75.0%, the rebuilt was simply answering the question the original was never asked; same shape on `debt10/20/30` (golden £0s; rebuilt penny-exact `100000×1.065^y`).
- **the check:** before convicting either side, count the Calculate buttons on the ORIGINAL page (`curl -s <url> | grep -c 'btn-primary\|onclick="calc'` or read it) and ask which one the golden's `pressed.label` names. An id that never moves in ANY golden vector while its section's inputs were driven is an unpressed-section suspect, not a measurement. Judge only ids downstream of the button the golden actually pressed; the other section needs its own capture (press ITS button) before any claim about it.
- **added:** 2026-08-08, mortgagecalculator lane (replay-comparator session)

### `MIGRATIONS_DIR=… ./run-migrations.sh --apply` scopes NOTHING if the assignment lands on its own line — and the unscoped run applies ~100 other threads' pending files

- **footprint:** `scripts/migration/run-migrations.sh`, `docs/agent_docs/sql_for_agents/`, `schema_migrations`, and **any** `VAR=value cmd` you paste into a terminal or hand to someone
- **fires when:** you scope a migration run with the documented `MIGRATIONS_DIR` env override and the command wraps, gets copied as two lines, or is pasted into a prompt that splits it. `VAR=value cmd` is ONE statement and exports `VAR` to that child only. `VAR=value` followed by a newline is a plain shell assignment — **not exported** — so the child process never sees it and silently uses its default. There is no error, no warning, and the run looks exactly like the scoped one for the first few seconds.
- **why it is severe here and not merely untidy:** this repo's migrations directory carries a very deep pending backlog — **98 pending files on 2026-08-08**, spanning weeks and dozens of threads, because most sessions apply their own file by hand and never record it. `--apply` takes **every** pending file in number order. So the failure mode is not "my migration ran in the wrong place", it is "**~100 other people's migrations ran, oldest first, in a database none of them were re-verified against**".
- **the worked case, 2026-08-08:** the intended scope was one file (`338`). The assignment landed on its own line. The runner applied and RECORDED four other threads' migrations before halting:
  - `198_tools_api_gauntlet_rounds.sql` — **created `gauntlet_rounds` in `clients_db`**, while migration 276's own guard says that table belongs on *"the ISLAND, not clients_db"*. Empty (0 rows, 11 cols), but now in the ledger.
  - `203_link_resolver_sections_optional.sql` — genuine no-op, said so itself.
  - `204_robot_hands_matchmatrix_normalized_specs.sql` — **10 live `products.content_data` rows** on robot-hands now carry `matchmatrix`.
  - `207_report_dossier_component.sql` — one `content_components` row updated.
  It then FAILED on `208_robot_hands_report_island_config.sql` (which uses psql `:'var'` interpolation and needs `-v`), and **stopped — so the file we actually wanted, 338, never ran at all.**
- **the check, and it is the command shape, not a query:** keep it on ONE line, or `export` it, or use `env`. Verify the scope BEFORE `--apply` by reading the dry run's own `Pending (N)` line:
  ```bash
  env MIGRATIONS_DIR=/abs/path ./scripts/migration/run-migrations.sh          # dry run FIRST
  # the output must say  Pending (1):  and name YOUR file. If it says Pending (98), STOP.
  ```
  **`Pending (N)` is the scope assertion.** It is printed before anything is applied, it cannot be faked, and on this tree the correct N for a scoped run is almost always 1.
- **and never hand this command to a human or a prompt as a wrapped line.** A two-line `VAR=x`/`cmd` pair is indistinguishable from the one-liner in most renderings and behaves completely differently. Write `env VAR=x cmd`, which survives being split because it is a single command word.
- **source:** 2026-08-08, `bugfix_122_contrast_ink_slots` lane; `WRONG_CALLS.md` same date. Related: `bugs_open/007` (unrecorded-but-applied migrations, which is *why* the backlog is 98 deep).

### Firecrawl serves a CACHED snapshot by default — changing what a domain serves does NOT change what a scrape returns, and your own diagnostic probe CREATES the poisoned entry

- **footprint:** `internal/adapters/webscrape/providers/firecrawl.go` (`buildScrapePayload`/`buildCrawlPayload` `max_age`), every `scrape_web`/`firecrawl_*` step config, `domain-research-classifier` `scrape_site`
- **fires when:** you change what a URL serves (park a redirect, take a site down, deploy a fix) and then scrape it to see the new state — or dispatch any pipeline whose scrape step omits `scrape_config.max_age`. Firecrawl v2 caches 200-responses and serves them for its default TTL when the request carries no `maxAge`; the platform's provider only sends `maxAge` when the step config has `max_age`, and as of 2026-08-08 no step in `agent_definitions` set it except `domain-research-classifier.scrape_site` (added that day). Worse: error responses do NOT overwrite the cache — a fresh fetch that gets a 522 leaves the stale 200 snapshot in place, so "I forced one fresh fetch" does not clean it.
- **the tell:** the scrape returns `statusCode: 200` and coherent content while `curl` against the same URL shows something else entirely (a 522, a redirect, new content). Measured 2026-08-08 on `https://webdesign.uk`: edge PROVEN parked (curl 522), yet a default scrape returned the pre-parking redirect target (`webdesign.co.uk`, full markdown) with `metadata.cacheState: "hit"` and a `cachedAt` matching the earlier probe — the probe run to MEASURE the contamination is what wrote the cache entry the pipeline would then have read.
- **the check:** read `metadata.cacheState`/`cachedAt` in every scrape you treat as evidence — `"hit"` means you are reading the past. To measure current state, send `maxAge: 0` explicitly (step config: `"scrape_config": {"max_age": 0}`). Before dispatching a pipeline whose decision anchors on a scrape (the classifier anchors the ENTIRE build), confirm its step config carries `max_age: 0` or accept that it may read a snapshot up to the default TTL old — and remember your own probes populate the same cache the pipeline reads.
- **source:** 2026-08-08, webdesign_uk_build_service lane (rebuild resubmission); WRONG_CALLS.md same date.

### `build-dispatch-loop` invoked with bare `action=orchestrate` reports COMPLETED and processes NOTHING — the supported invocation is the trigger's spawn+call

- **footprint:** `build-dispatch-loop` (agent_definitions), `platform/orchestration/coordinator.go` (`ErrLoopExpansionHandled`), any hand-rolled kcat dispatch of a loop-bearing agent
- **fires when:** you bypass a stalled build queue by orchestrating `build-dispatch-loop` directly (the obvious move — the same envelope shape works for every linear agent). The run loads the site's items (`pending.has_items: true`), logs "Loop expansion handled — outer continueExecution exiting", then the workflow completes: status COMPLETED, current_step complete, no error — and the items are untouched, `status='triaged'`, `attempt_count` 0. Nothing failed, nothing was claimed, nothing was dispatched.
- **the tell:** COMPLETED with `pending.items` populated but NO `claim_result`/`handler_spawned`/`item_completed` keys in `collected_data`. Compare a trigger-driven run (spawn_agent + call_agent with `action: process`): it carries `claim_result` and its item reaches `complete`. Measured 2026-08-08: direct orch `4e26e881…` no-opped on webdesign.uk's item; trigger-driven `67fe4fae…` processed leopardess's item the same evening.
- **the check:** after ANY dispatch-loop invocation, read the ITEM rows, not the orchestration status — `SELECT status, attempt_count FROM site_work_items WHERE id='<item>';` still-`triaged` after a COMPLETED loop run means the run was a no-op. To hand-drive a starved queue, skip the loop entirely: claim the item yourself, orchestrate its `handler_agent` directly with the loop's `input_mapping` shape (worked recipe: webdesign_uk_build_service/HANDOFF_2026-08-08 §1), and mark it complete on verified output.
- **source:** 2026-08-08, webdesign_uk_build_service lane; NOTES 08-08 §3–4. Mechanism NOT yet diagnosed (090 candidate) — this entry records the behaviour, not the cause.

---

### `revalidate_review_queue` cannot be scoped by a dispatch — `site_id` and `item_type` come from STEP CONFIG, and a filter in `input_data` is silently ignored
- **footprint:** `platform/orchestration/actions/revalidate_review_queue_action.go`, `agent_definitions` type `diagnosis-review-queue-revalidator`, `scheduled_tasks.review-queue-revalidate-daily`, `revalidation_result`

Both filters are read from the step config and nowhere else:

```go
config := params.StepConfig.Config              // :266
siteFilter, _ := config["site_id"].(string)     // :275
typeFilter, _ := config["item_type"].(string)   // :276
```

The live `sweep` step has **no `input_mapping`** and its entire config is
`{"dry_run": false, "max_items": 1500}`. **So a hand-dispatch carrying
`{"site_id":…,"item_type":…}` in `input_data` runs FLEET-WIDE over every covered type, and looks
exactly like a scoped run** — gate 2 stamps `result.revalidation` on every row it scans and does not
close (~150 rows today), and a run that closes 1 row is indistinguishable from a filter that worked.

This has already produced a false record: the 2026-08-06 hand-dispatch was written up as
*"**Scoped deliberately**, not fleet-wide"* and was not scoped. **The same file had documented the
trap two entries earlier** — which is the real lesson: *documenting a trap creates a feeling of
having handled it.*

**the check:** before believing any dispatch of this action was scoped, read the step config, not
your payload —
```sql
SELECT jsonb_pretty(s.value) FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.type='diagnosis-review-queue-revalidator' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL AND s.key='sweep';
```
If you genuinely need a scoped run, the honest routes are to add `site_id`/`item_type` to the **step
config** (config, live immediately, no build — but you are editing a shared definition every
scheduled run then reads), or to add an `input_mapping` first. **Do not infer scope from what you
published.** And confirm effect at `site_work_items.resolution_path='auto:revalidated'` and the
run's `scanned` count, never at the payload you sent.

---

### A chassis pod's retrievable log holds LESS THAN A SECOND of history under load — a 0-hit grep for a past event is evidence of nothing
- **footprint:** `kubectl logs`, `agent-chassis`, `platform/orchestration/coordinator.go`, `agent-job-cleanup`

Measured 2026-08-08 22:27:17Z on `agent-chassis-778b7c77c7-rd27g` (v1.0.1268): the oldest
retrievable log line was stamped 22:27:16.878Z — **0.4 seconds old**. The coordinator's
`DEBUGaa` whole-state dumps run to hundreds of KB per step transition, so container log
rotation eats the file within seconds whenever an orchestration is active. Ephemeral
`agent-<type>-<id>` pods stack a second destroyer on top: `agent-job-cleanup` deletes
Completed pods within minutes, taking even the rotated remnant. A `--since=3h` sweep over 23
pods therefore returns 0 for an event that WAS logged minutes earlier, and that 0 reads
exactly like "the code path never ran" — it cost `bugs_open/136` §9 a witness (the alias
warn line fired and was unreachable before any grep could run).

- **the check:** to witness a runtime behaviour, do not sweep logs after the fact. Arrange a
  **DB-visible observable** whose value could come out otherwise (the `bugs_open/136` §11
  witness pattern), or attach `kubectl logs -f` BEFORE the event. If you must grep for a
  recent event, first bound the window you actually have:
  `kubectl logs <pod> --tail=-1 | grep -oE '"ts":"[^"]*' | head -1` — if the oldest line
  postdates your event, the grep was theatre, not absence.
- **source:** 2026-08-08, `bugfix_136_config_key_aliases` lane; `bugs_open/136` §9/§11.

### An UPDATE of `site_components.rendered_html` fires a DB trigger no grep of Go will ever show you — and a failing chrome rebuild may be the trigger REFUSING to destroy an unarchived artefact

- **footprint:** `site_components`, `rendered_html`, `rendered_html_digest`, `trg_site_component_archive`, `site_component_history`, `renderAndStoreSiteComponent`, `sql_for_agents/344`
- **fires when:** (a) you UPDATE `rendered_html` by any means — Go, admin dashboard, raw psql — and a row appears in `site_component_history` you did not write (that is the bugs_open/226 archive working: the outgoing bytes are kept whenever the new value differs); or (b) a chrome overwrite ERRORS mentioning `site_component_history` and it looks like your change broke the render path — it did not; the trigger is FAIL-CLOSED and is refusing to destroy bytes it cannot archive (ledger table dropped/unwritable is the usual cause).
- **the check:** the mechanism is `trg_site_component_archive` (mig `344_site_component_history_divergence_guard.sql`) — a trigger, so `grep -rn` over `platform/` finds nothing; read the migration, and `SELECT tgname, tgenabled FROM pg_trigger WHERE tgname='trg_site_component_archive'`. Do NOT "fix" a failing rebuild by dropping the trigger (the ROLLBACK sidecar exists for a deliberate, stated rollback). And never set `rendered_html_digest` from any writer except `renderAndStoreSiteComponent`'s store statement — a stamp written beside your own bytes classifies them machine-made and silences the hand-patch detector for the slot.
- **added:** 2026-08-08, bugfix_226_chrome_divergence lane (STY-054)

### A step with NO `input_fields` resolves its inputs by RANDOMISED recursive search — the wrong sibling's id wins ~86% of the time and looks exactly like the right one

- **footprint:** `ExtractActionInputs`, `ExtractFields`, `extractSingleField`, `findFieldRecursive`, `platform/orchestration/datahelpers/action_inputs.go`, `platform/orchestration/datahelpers/unified_extractor.go`, `input_fields`, `deploy_image_asset`, `asset_id`
- **fires when:** you add, re-point or "simplify" a step whose config has **no `input_fields`** and no explicit dotted path for the field in question, in a workflow where two sibling steps each emit a map containing the **same key name** (`asset_id`, `work_item_id`, `page_id`, `site_id` …). `ExtractActionInputs` falls to Strategy 2, which reaches `extractSingleField` Strategy 4 — `findFieldRecursive`, which walks `for key, val := range m`. **Go randomises map iteration order**, so the field resolves to an arbitrary sibling's value, differently between runs of the same code on the same data. Nothing errors; the value is well-formed and of the right type. Measured 2026-08-08 on the real helper, 400 iterations, identical input: a **logo** deploy step resolved the **hero's** `asset_id` in **344/400 (86%)** runs.
- **why it hides:** the code already carries a comment acknowledging this ("aggressive recursive search that can find stale values from previous loop iterations", `action_inputs.go:466-471`) — but that comment only documents **Strategy 0**, the escape hatch for configs that *do* supply an explicit multi-segment path. A step with neither `input_fields` nor a dotted config path gets no protection at all, and reading the comment leaves you believing the case is handled. A single-asset test passes every time, because with one sibling there is nothing to pick wrongly.
- **the check:** never reason about which value a no-`input_fields` step will get — **run it, repeatedly.** Extract the resolution into a test and iterate ≥100× on the same input, asserting the value is *stable*, not merely correct:
  ```go
  seen := map[string]int{}
  for i := 0; i < 400; i++ {
      in, _ := datahelpers.ExtractActionInputs(freshCollectedData(), cfg, TheInputSpec, zap.NewNop())
      seen[in.Get("asset_id")]++
  }
  // len(seen) > 1  =>  the step is picking a sibling at random
  ```
  One run proves nothing here — a single pass is right ~14% of the time in the measured case, and 100% of the time when only one sibling exists. Before believing a step is safe, ask the DB whether the field name is unique in that run's `collected_data`:
  `SELECT k, count(*) FROM orchestration_states o, LATERAL jsonb_object_keys(o.collected_data) k WHERE o.owner_agent_type='<type>' GROUP BY k;` then enumerate the keys *inside* the sibling outputs, not just the top level. **The fix is to name `input_fields` (or an explicit dotted path) on the step — config, live immediately, no roll.**
- **⚠ corollary that has already misled one fix plan:** because of this, "remove the purpose-keyed lookup and let the `asset_id` path resolve it" is a *downgrade* wherever the caller has no `input_fields` — it replaces a correct discriminator with an 86%-wrong one. `bugs_open/209` ranked exactly that fix first; see the 2026-08-08 verification block in that file before acting on any similar "just use the id" simplification.
- **added:** 2026-08-08, `bugfix_209_deploy_purpose_keyed_source` lane; evidence in `platform/orchestration/actions/deploy_image_asset_purpose_source_test.go`

### An `empty_section` item's `item_key` names the slot as it was WHEN FILED — a fix that replaces the component renames the slot, so grepping `item_key` for the slot you can see returns 0

- **footprint:** `site_work_items`, `item_key`, `page_components`, `slot_name`, `platform/orchestration/actions/discovery_checks/check_empty_sections.go`, `findEmptySections`, `VerifyEmptySectionResolved`
- **fires when:** you are looking at a live page with a visibly empty section, you read its slot out of `page_components.slot_name`, and you ask whether it was ever detected — `WHERE item_key LIKE '%<slot>%'`. The handler is allowed to satisfy an `empty_section` item by **replacing** the component rather than filling it, and the replacement can land in a **differently named slot**. Measured 2026-08-08 on `finetuning.uk` pages `ai-guides` and `insights`: items exist under slot `featured-article`, while the component sitting empty on both pages today is slot `featured-content` — a different string, so the `item_key` grep returns **0** and reads as "never detected".
- **why it hides:** the zero is *true*, it is just about a different slot, and nothing on the page records the old name. Both halves look authoritative — you took the needle off the live artefact and searched the table that is supposed to hold the answer. It is worse than a plain naming drift because the same divergence ALSO defeats the completion gate: `VerifyEmptySectionResolved` looks up the component id the item names, finds it deleted, and (pre-`RFC_017`) failed OPEN — so the item reads `complete` while the page still serves a 334-byte shell. An item that says `complete` and a grep that says `never filed` are the two things you would check, and both lie in the same direction.
- **the check:** never let `item_key` carry an absence claim for this type. Search the **item's own prose and spec as well**, then reconcile against the page by *page id*, not by slot name:
  ```sql
  SELECT item_key, status, summary FROM site_work_items
  WHERE item_type='empty_section' AND page_id='<page uuid>' ORDER BY created_at DESC;
  ```
  If an item shows `complete` for this type, read `result->'_verification'` before believing it — `{"status":"error", …"no longer exists (genuinely fixed or silently deleted — indistinguishable here)"}` means the verifier never confirmed anything. **Post-`RFC_017` (live `v1.0.1268`) that same case now lands `triaged`/`failed` instead of `complete`** — so on a pre-roll row the identical payload means the OPPOSITE outcome; date the row against the roll.
- **⚠ corollary — this is a rebuild burner now.** `empty_section` is one of the 4-of-8 verifiers whose target can legitimately vanish. With fail-closed live and the handler still free to replace-not-fill, an absent-target case burns up to `max_attempts` (3) page rebuilds before a human sees it. `bugs_closed/032`'s own "stronger option" — ask whether the page still declares the slot, return `Resolved:false` — is the cheap fix, and this is the recurrence that argues for it and for `RFC_017` option 3.
- **added:** 2026-08-08, `bugfix_201_page_content_writer_dispatch` lane

### `landmines-sync.py --apply` CONSUMES the NEEDS_VERIFICATION signal — run it directly, as CLAUDE.md tells you to, and your new entry can never be verified

- **footprint:** `scripts/landmines-sync.py`, `scripts/landmines-verify-dispatch.sh`, `scripts/trigger-landmine-verifier.sh`, `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`, `doc_notes`, `landmine-verification`
- **fires when:** you append a landmine entry and run `./scripts/landmines-sync.py --apply` — which is exactly what **CLAUDE.md instructs** ("After you append, run `./scripts/landmines-sync.py --apply` so the `doc_notes` rows follow"). The sync computes its `NEEDS_VERIFICATION` list as `new + changed` **relative to the rows already in `doc_notes`**. Applying writes those rows, so on the next invocation your entry is neither new nor changed. Running the wrapper `landmines-verify-dispatch.sh` afterwards then prints **"Nothing needs verification (no new or changed entries this run)"** and dispatches nothing — permanently. Observed 2026-08-08, first run of the wrapper after a direct `--apply`.
- **why it hides:** every visible signal says success. The entry parsed (it lists its footprint count, not `skipped (no footprint)`), the rows are really in `doc_notes` and verifiable by identity, the owned-row total went up by exactly the footprint count, and the wrapper exits 0 with a sentence that reads like good news. Nothing anywhere says "this entry has no verification and now cannot get one". The two scripts are also documented as interchangeable — `landmines-verify-dispatch.sh`'s header says plain `--apply` "still works standalone ... for anyone who wants the sync without the dispatch", which describes the *choice* but not that the choice is **irreversible for that entry**.
- **the check:** to get an entry synced *and* verified, run the wrapper **instead of**, never after, the plain sync: `./scripts/landmines-verify-dispatch.sh`. If you have already applied directly, the signal is gone and the only route left is to fire the verifier by hand with the entry's `doc_notes.source` slug (take the slug from the sync's own output line, do not hand-derive it):
  ```bash
  ./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug-from-sync-output>'
  ```
  Then confirm the dispatch actually landed rather than trusting exit 0 — `kcat -P` is known to send nothing and succeed:
  `SELECT current_step, status FROM orchestration_states WHERE collected_data::text LIKE '%<correlation>%';`
  To find entries that silently lost their pass, ask which landmines have no verdict:
  ```sql
  SELECT DISTINCT n.source FROM doc_notes n WHERE n.categories ? 'landmine'
    AND NOT EXISTS (SELECT 1 FROM doc_notes v
                    WHERE v.categories ? 'landmine-verification' AND v.subject_key = n.source);
  ```
- **added:** 2026-08-08, `bugfix_201_page_content_writer_dispatch` lane

### A form component's validation being real is not evidence its SUBMIT is — `contact-block` prints "your message has been sent" from a `setTimeout` and has no transport at all

- **footprint:** `contact-block`, `cb-contact-form`, `cb-submit-btn`, `/tools/assets/contact-block.js`, `content_components`, `js_content`, `html_template`
- **fires when:** you touch, fence, port, restyle or QA any form-bearing section component and satisfy yourself that it works by *using* it. `contact-block`'s client-side validation is genuine and specific — mistype the email and you get "Please enter a valid email address"; leave the message short and you get the character count. Submit it correctly and after ~1.2s you get a green "Your message has been sent. We'll be in touch shortly." **Nothing is sent.** The served form carries no `action` and no `method`, and the served `/tools/assets/contact-block.js` (2,100 bytes) contains **zero** `fetch(` / `XMLHttpRequest` / `sendBeacon` / `form.submit(`. The 1,200 ms delay exists only to look like a network round-trip, and `form.reset()` then wipes the visitor's text. Live on three client pages including `robot-hands.com/contact.html`. Filed as `bugs_open/228`.
- **why it hides:** every signal a human checks points the right way. The validation errors are correct and specific, so the form is visibly *wired up*; the success message is styled as a success; the pause reads as a request in flight; and the fields clearing afterwards is exactly what a real submit does. A browser check that drives the form and asserts the success text — the obvious acceptance check to write — **passes**, and then the contract vouches for the lie (`bugs_open/161`'s class: the record ratifies the claim it caused).
- **the check:** ask the ARTEFACT for a destination, never the behaviour for a verdict. Two greps, both on what is actually served, not on the DB row alone:
  ```bash
  curl -s https://<site>/<page> | grep -oE '<form[^>]*id="cb-contact-form"[^>]*>'   # expect action= and method=
  curl -s https://<site>/tools/assets/<component>.js | grep -cE 'fetch\(|XMLHttpRequest|sendBeacon|form\.submit\('
  ```
  A `0` from the second with no `action` in the first means the form is inert no matter what the page says back to you. To sweep the fleet for the same shape rather than this one component, ask all three questions at once — a component can be honest by EITHER route:
  ```sql
  SELECT cc.function,
         (cc.html_template ~* '<form[^>]*action=')                          AS form_has_action,
         (coalesce(cc.js_content,'') ~ 'fetch\(|XMLHttpRequest|sendBeacon')  AS js_has_transport,
         count(DISTINCT pc.page_id) AS pages
    FROM content_components cc
    LEFT JOIN page_components pc ON pc.component_id = cc.id
    LEFT JOIN pages p ON p.id = pc.page_id AND p.status = 'active'
   WHERE cc.is_active AND cc.html_template ~* '<form'
   GROUP BY 1,2,3 ORDER BY 4 DESC;
  ```
  30 rows on 2026-08-08 and `contact-block` was the only one false on both. **And when you write its acceptance fence, assert the VALIDATION path, not the success message** — otherwise the contract makes the defect permanent.
- **added:** 2026-08-08, `staged_component_build` lane (D10 batch 6)

### Adding a template variable to a prompt WITHOUT adding its field to that step's `input_fields` renders EMPTY and errors nothing — the migration looks applied and does nothing

- **footprint:** `input_fields`, `execute_llm_prompt`, `prompt_template`, `agent_definitions`,
  `default_config`, `llm_call_log.prompt_rendered`, `docs/agent_docs/sql_for_agents/`
- **fires when:** you add a new load step (`query_database` → `output_field: foo`) and
  reference `{{.foo.text}}` in a downstream prompt, but leave that step's
  `input_fields` list as it was. This is the normal shape of "give the agent one more
  piece of context", and it is how most config-only agent fixes are written.
- **the tell:** there isn't one. The step only receives the fields it lists; an
  unreferenced variable renders empty, the LLM call succeeds, the workflow completes, and
  every structural check passes — the new step ran, its `output_field` is populated in
  `collected_data`, the prompt contains the placeholder, the row updated. The failure is
  visible in exactly one place: the **rendered** prompt. It is invisible in the template.
- **the check:** assert the field is listed, at apply time —
  `default_config #> '{workflow,steps,<step>,config,input_fields}' @> '["<field>"]'::jsonb` —
  and then prove it at runtime from `llm_call_log.prompt_rendered`, which is the only
  record of what the model was actually handed. **The runtime assertion needs a POSITIVE
  CONTROL or it cannot fail usefully:** a channel that loads nothing produces the same
  "absent" reading as a subject that legitimately has nothing. Pick one subject whose
  content you know is non-empty and one you know is empty, and require the two to come out
  OPPOSITE ways in the same query. Checking only the empty case passes perfectly on a
  channel that has never worked.
- **source:** found while writing `bugs_open/227`'s fix (345, experience-planner) — caught
  in review of my own SQL, before apply, not by a failure
- **added:** 2026-08-08, loancalculator_couk lane

### `golden_compare_post.py` reports every VERBATIM page as "diverged" — the content-shape assertion is a decomposition check, not a regression signal

- **footprint:** `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/golden_compare_post.py` · `acceptance/GOLDEN_2026-08-05_prechange.json` · any post-change check of a `loans/*` or other adopted-verbatim page
- **fires when:** you run the golden comparator against a page that was never decomposed — the natural move after changing one, since the RUNBOOK's step 5 says "the calculator must still compute"
- **the tell:** `1 of 1 tool(s) diverged`, with every divergence line naming the `content` field and "expected the empty chrome span ('|inline')". On a verbatim page `#content` is the full page wrapper (`|block`, all the prose), not the decomposed chrome's empty span, so the assertion fails by DESIGN — and the message reads exactly like your change moved prose into a live wrapper. Numeric fields matching is the real verdict; the shape line is scope noise.
- **the check:** before reading "diverged" as damage, run the identical compare on a verbatim page you did NOT touch — identical single content-shape divergence on the control means comparator scope; any NUMERIC element divergence is real. Proven 2026-08-09: six edited verbatim pages + untouched `loan-vs-savings` all report the same single shape line; `consolidation` (actually decomposed) reports `MATCHES (arithmetic exact)`.
- **source:** bugfix 224 session, verifying the 0% fix — the control was run before believing either reading
- **added:** 2026-08-09, bugfix 224 session

### `who-owns.py`'s VERDICT line goes OWNED on a bare digit-substring match — a bug number that appears in ANY unrelated commit subject reads as claimed

- **footprint:** `scripts/who-owns.py` (`subject_commits`, the `owners or subject_commits` verdict condition at the file's end) · any bug number that is short or round enough to coincide with a count, a version fragment, or another bug's number inside unrelated commit text
- **fires when:** you run `who-owns.py <N>` to check whether an open bug is already claimed before starting work on it — the exact use CLAUDE.md tells you to make first.
- **the tell:** the script prints `VERDICT: OWNED or recently active.` with `(none identified)` under "likely OWNING workstream(s)" and zero rows under "commits touching the bug file(s)" — the ONLY evidence for the verdict is a "commits whose SUBJECT is about this bug" list, and reading those subjects shows the digits appearing incidentally (a site count, an unrelated bug's number, a version fragment), never actually about the bug in question. Run cold across every number in `bugs_open/` on 2026-08-08/09: 42 of 44 bugs checked came back "OWNED" this way, including a bug (`228`) filed *that same day* with zero prior commits anywhere and a brand-new, empty workstream directory.
- **the check:** never trust the printed VERDICT line alone. Read the "likely OWNING workstream(s)" section specifically — `(none identified)` there, regardless of what the VERDICT line says, means no directory actually discusses the bug. Corroborate with a fresh `git log --since="<window>" --all` grep for the bug's slug (not its bare number) and, if a candidate workstream directory exists, check whether it was created/touched in the last few minutes (mid-flight session) versus hours/days ago (stale, safe to read past).
- **source:** bugfix_228 session triage — caught by reading the "(none identified)" sections rather than the verdict line, before routing work at a falsely-claimed bug; not a wrong call, but the near-miss that would have produced one
- **added:** 2026-08-09, bugfix_228_contact_block_transport lane

### A STATIC step-config value for a spec-DEFAULTED field is DEAD — the default wins silently, and the action's own `if empty` fallback can never fire

- **footprint:** `ExtractActionInputs`, `ActionInputSpec`, `platform/orchestration/datahelpers/action_inputs.go`, `Defaults`, `agent_definitions` step config authoring
- **fires when:** you write (or trust) a step config that sets a plain static value — `"purpose": "logo"`, `"severity": "warning"`, any non-dotted string — for a field whose `ActionInputSpec` carries a `Defaults` entry. The config LOOKS authoritative and is silently ignored: Defaults are copied into `Values` first (`action_inputs.go:457-460`), Strategies 1/2/3 each skip a field that already holds a value (`:499`, `:511`, `:523`), and Strategy 0 only reads **dotted** config paths (`strings.Contains(pathStr, ".")`, `:478`). A static value has no strategy that can carry it. Any in-action fallback shaped `if inputs.Get(f) == "" { read config }` is unreachable — the default guarantees non-empty. The deprecated `*_field` bridges are equally inert for defaulted fields (Strategy 3's has-value skip).
- **the measured case:** `pageflow-builder` / `site-work-orchestrator` `deploy_logo_image` carries static `"purpose": "logo"`; effective purpose is **"hero"** (spec default, since `34d2315ce` 2026-02-20). Consequence if run: hero resize class, and the logo's bytes committed to the HERO's deploy path (`BuildAssetPaths`: filename = purpose + ext). `bugs_open/231`; pinned in `deploy_image_asset_purpose_source_test.go` (`TestLegacyLogoStep_StaticPurposeIsShadowedByDefault`).
- **the check:** before trusting any static config value, open the action's `ActionInputSpec` — if the field appears in `Defaults`, the static value is dead. The only config shape that defeats a default is a Strategy-0 **dotted path** to real data (`"purpose": "logo_stored.purpose"` — proven deterministic in the same test file). To find whether your value is being read at all, run the resolution: `datahelpers.ExtractActionInputs(collected, cfg, TheSpec, zap.NewNop())` in a scratch test and print the field — thirty seconds, and it answers for your exact shapes. Same family as bugfix_136's `Deprecated`-alias landmine; the general rule is **against a defaulted field, only a dotted path can win.**
- **added:** 2026-08-09, `bugfix_209_deploy_purpose_keyed_source` lane; fleet-class census handed to 090 run `e952039b`

### A `lock_blocked_change` item does NOT mean the copy differed — it fires on ANY incoming section matching a locked slot, and records no proposed content

- **footprint:** `site_work_items.item_type='lock_blocked_change'` · `platform/orchestration/actions/lock_helpers.go` (`emitLockBlockedChangeItem`, `emitLockBlockedChange`) · `platform/orchestration/actions/save_page_sections_action.go` (`matchLockedRow`, the locked-slot guard ~line 769) · any locked `page_components` row on a page being rebuilt
- **fires when:** you lock some sections of a page, run a rewrite aimed at the unlocked one, and then read the resulting `lock_blocked_change` items as evidence of what the writer tried to do — i.e. as a leak detector. It is the obvious reading: the summary literally says *"save_page_sections wanted to overwrite locked section X"*.
- **the tell:** there isn't one, and that is the trap — the item looks identical whether the writer rewrote the section or handed it back byte-for-byte. The composer emits **every** section on every run, and the guard fires on a `matchLockedRow` **slot-name** match alone. `emitLockBlockedChange`'s spec carries `surface`, `slot_name`, `locked_by`, `lock_type`, `blocked_action`, `source`, `fix` — **and no content, proposed or stored.** So the item count equals "locked slots on the page", a number you already knew before dispatching, and it is guaranteed non-zero by your own locking.
- **the check:** to learn whether the guidance actually leaked, compare what the writer **proposed** against what was stored *before* the run — a different table answering a different question:
```sql
WITH prop AS (
  SELECT step_name, (response_text::jsonb)->>'content' AS proposed
    FROM llm_call_log WHERE correlation_id='<CORR>')
SELECT prop.step_name, length(prop.proposed) AS proposed_len,
       length(bak.stored) AS stored_len, (prop.proposed = bak.stored) AS byte_identical
  FROM prop JOIN (SELECT slot_name, content_data->>'content' AS stored
                    FROM <your_pre_run_backup_table>) bak
    ON bak.slot_name = <slot for that iteration>;
```
  This requires a **pre-run backup table** — `page_components` itself is no use afterwards, because `save_page_sections` DELETEs and re-inserts every agent-writable row, and the locked rows it preserved are by definition the ones whose proposals were discarded. Proven 2026-08-09 on loancalculator.co.uk `index`: three `lock_blocked_change` items were raised, and the comparison showed **two of the three sections were byte-identical** (obeyed) while one had genuinely rewritten itself — a distinction the items could not make and which changed the lane's conclusion about whether conditional prompt phrasing works.
- **source:** loancalculator_couk lane, closing `HANDOFF_2026-08-08b` §2 — caught while writing up "the locks caught the leak", before the claim was recorded; the leak was real but these items were not the evidence for it
- **added:** 2026-08-09, loancalculator_couk lane

### Re-enabling `improvement-sweep` will never examine the fleet's busiest sites — the cap that protects the queue is an exit door from examination

- **footprint:** `scheduled_tasks` (`improvement-sweep`, `pre_query`), `sites.updated_at`, any plan that says "just re-enable the sweep" or cites it as the thing that will restore fleet-wide discovery coverage
- **fires when:** you flip `improvement-sweep` to `enabled=true` (or argue about doing so) expecting every site to start being examined again — the natural reading of its description, and the obvious remedy for `bugs_open/230`
- **the tell: none, and that is the trap.** The sweep runs, `last_triggered_at` advances, sites get examined, findings appear — everything looks like coverage. But its live pre_query skips any site with ≥50 open build-pipeline items, and `bugs_open/083` means those queues do not drain, so a site that crosses 50 falls out of examination PERMANENTLY while the sweep reports healthy activity. Measured 2026-08-09: webdesign.co.uk 85, dartsonline.com 79 — the two most-worked sites in the fleet are already over the cap. Its `ORDER BY s.updated_at ASC NULLS FIRST` also starves (register IMP-010, known since April): nothing the sweep does advances its own sort key.
- **the check:** before treating the sweep as a coverage mechanism, run the cap query against today's fleet: `SELECT s.domain, count(*) FROM site_work_items wi JOIN sites s ON s.id=wi.site_id WHERE wi.status IN ('triaged','detected') AND wi.pipeline='build' GROUP BY 1 HAVING count(*) >= 50;` — every row returned is a site the sweep will NEVER examine. For detection coverage use SCH-025's `site-discovery-rotation-*` tasks (no backlog cap, stamp-on-selection fairness) instead; the sweep's remit is the triage/fix loop, whose re-enable is `bugs_open/083`'s pending owner decision.
- **source:** 2026-08-09, bugfix_230_discovery_driver — found while establishing why `bugs_open/230`'s candidate "re-enable the designed driver" was not viable as-is
- **added:** 2026-08-09, bugfix_230_discovery_driver

### A retraction sweep's `uncovered_backlog` total can stay FLAT while the type you just adopted leaves it entirely

- **footprint:** `orchestration_states.collected_data->'revalidation_result'` (`uncovered_backlog`, `uncovered_types`), `revalidate_review_queue_action.go` (`reportUncoveredBacklog`), and any handoff step that says "confirm the adoption by watching `uncovered_backlog` fall"
- **fires when:** you register a new `item_type` in `reviewRevalidators`, roll the chassis, and check the next scheduled sweep to confirm it took effect
- **the tell: none, and that is the trap** — the total is a *sum across ~40 types*, and every other type is growing underneath you while yours leaves. Measured 2026-08-09 on the `voice_tells` adoption: `uncovered_backlog` was **625 before the roll and 625 after**, which reads as "the change did nothing". It did: `voice_tells` went **25 → absent** from `uncovered_types`, and nine other types grew by **exactly 25** in the same window (`claims_unverified` +5, `content_rewrite` +5, `lock_blocked_change` +5, `save_refused_incomplete` +4, `empty_internal_href` +2, and five more at +1). The coincidence is not the point — any inflow at all makes the total uninformative about a single type.
- **the check:** confirm at the **per-type map**, never the total. Your type must be ABSENT from `uncovered_types` (not merely smaller), and `scanned` must rise by its live row count — decompose `scanned` by type to prove it rather than trusting the delta:
```sql
-- the type must have LEFT the map
SELECT collected_data #>> '{revalidation_result,uncovered_types}' AS uncovered_types,
       collected_data #>> '{revalidation_result,scanned}'         AS scanned,
       collected_data #>> '{revalidation_result,cap_binding}'     AS cap_binding
FROM orchestration_states WHERE orchestration_name ILIKE '%reval%' ORDER BY created_at DESC LIMIT 1;
-- and scanned, decomposed — these must sum to `scanned`, with your type at its full live count
SELECT item_type, count(*) FROM site_work_items
WHERE result #>> '{revalidation,at}' >= CURRENT_DATE GROUP BY 1 ORDER BY 2 DESC;
-- ⚠ the stamp key is `at`, NOT `checked_at` — a wrong key returns 0 rows and reads as "nothing was scanned"
```
- **source:** 2026-08-09, bugfix_168_deployed_asset_path — `HANDOFF_2026-08-08b` §0b instructed the next session to confirm by watching `uncovered_backlog` fall by ~32. It did not move. The adoption had worked perfectly (all 32 rows scanned, one item closed unattended), and the number the handoff nominated was the one number that could not show it
- **added:** 2026-08-09, bugfix_168_deployed_asset_path

### `evidence_base` is DATA, so a `claims_unverified` retraction can fire with the copy untouched

- **footprint:** `site_specs` (`aspect='evidence_base'`), `site_work_items` (`item_type='claims_unverified'`, `resolution_path='auto:revalidated'`), `revalidate_unverified_claims.go`, `check_unverified_claims.go`
- **fires when:** you read a `resolved` stamp on a `claims_unverified` item as evidence someone fixed the page's copy
- **the tell: none** — the verdict reason says the page was re-scanned and no unsupported claim was found, which is true. But the register the scan compares against is an editable `site_specs` row: **adding a fact makes a previously unregistered number verifiable, so the item retracts although nothing on the page changed.** Arguably correct (the claim is now substantiated) and still not what "the copy was fixed" means. Same class as the `voice_gate` moving standard (CQ-020) but sharper, because a register row is edited far more casually than a gate threshold.
- **the check:** compare the page's components against the item's filing date before believing the copy moved:
```sql
SELECT w.item_key, w.created_at AS filed, max(pc.updated_at) AS newest_component,
       max(pc.updated_at) > w.created_at AS copy_actually_changed
FROM site_work_items w JOIN page_components pc ON pc.page_id = (w.spec->>'page_id')::uuid
WHERE w.item_type='claims_unverified' AND w.resolution_path='auto:revalidated'
GROUP BY 1,2;
```
  `copy_actually_changed = false` means the register moved, not the page. Also check `site_specs.updated_at` for the `evidence_base` aspect around the close.
- **source:** 2026-08-09, bugfix_168_deployed_asset_path — recorded with CQ-021 as the seam shipped, not after something looked wrong
- **added:** 2026-08-09, bugfix_168_deployed_asset_path

### A `SEED_SCOPE` naming a package-level `var` or `const` can NEVER be read — the code index holds no such kind, and the 090 loop burns its iteration cap returning UNVERIFIABLE

- **footprint:** `090_TRIGGER_needs_diagnosis_v1.sh` (`SEED_SCOPE`), `code_symbols`, and any symptom about a status list, registry map, regex, threshold or allow-list — `workItemRevalidatableStatuses`, `workItemTerminalStatuses`, `reviewRevalidators`, `itemTypesWithoutVerifiers`, `RUNTIME_FILL_ALLOWED` and every sibling
- **fires when:** you file a 090 whose hypothesis turns on the MEMBERSHIP of a Go package-level `var`/`const`, and name it in `SEED_SCOPE` as `path.go:theSymbol` — the natural spelling, and the one the trigger's own usage examples suggest
- **the tell: none, and the failure looks like a hard bug rather than a broken lookup.** The run completes, the trail is rich and well-reasoned, and the verdict is `UNVERIFIABLE — stopped: iteration-cap`. Read carefully it says so itself ("symbol/content searches for the literal declaration returned 0 rows — per the index-staleness caveat this is **unknown, not proof**"), but the headline reads as "we could not confirm your theory", which is easy to take as evidence against it. **It is evidence about the index, not about your premise.**
- **why:** `code_symbols.kind` takes only **`func` (3,592), `method` (1,114), `struct` (973), `alias` (40), `interface` (36)** — measured 2026-08-09 across the whole index. **There is no `var` or `const` kind at all**, so a package-level declaration is not merely stale, it was never indexed. `platform/orchestration/actions/work_items_common.go` indexes 4 funcs and none of its vars.
- **the check, BEFORE you spend a run — confirm every seed symbol actually exists in the index:**
```sql
SELECT symbol, kind, line_start FROM code_symbols
WHERE path = '<your/file.go>' AND symbol = '<YourSymbol>';
-- 0 rows ⇒ the loop cannot read it, no matter how the symptom is worded
```
  Zero rows ⇒ **name the enclosing FUNCTION instead** (function bodies are read in full, and the `var` is usually referenced inside one), or state the membership as a fact in the symptom text so the hypothesis does not depend on fetching it. Re-frame the question around what IS readable: a function body plus a `DataRequest` the loop can run beats a symbol it cannot open.
- **source:** 2026-08-09, bugfix_168_deployed_asset_path — run `f3d18013-0b78-472f-b2cb-5bf5e4e893b8` came back UNVERIFIABLE on exactly this, having named `work_items_common.go:workItemRevalidatableStatuses`. Re-filed function-scoped as `a174b184-dac2-47a1-95ca-df2d192e183a`. The premise was independently true and first-hand verified; the loop simply never got to see the list
- **added:** 2026-08-09, bugfix_168_deployed_asset_path

### A prompt census that iterates `default_config->'workflow'->'steps'` MISSES every loop-nested prompt — and the miss reads as a clean, confident absence

- **footprint:** `agent_definitions.default_config->'workflow'->'steps'` · `process_sections_loop` / any step whose config holds a **`sub_workflow`** · `page-content-writer`, `page-build-handler` · any fleet-wide audit of prompt text ("which agents carry the house voice / this rule / this claim?")
- **fires when:** you census prompts with the obvious query — `LATERAL jsonb_each(default_config->'workflow'->'steps') s(name, step)` reading `step->'config'->>'prompt_template'`. It is correct for flat agents and silently blind for looping ones.
- **the tell:** there is none in the output. A loop step's real prompt lives at `config->'sub_workflow'->'steps'-><inner>->'config'->'prompt_template'`, **two levels below where the census looks**, so the agent is simply absent from the result — it does not appear with a NULL, it does not error. Worse, the natural follow-up ("let me check that one agent directly") reproduces the blindness: `…->'config'->'steps'->…` is the *wrong* path — the key is **`sub_workflow`**, not `steps` — and a wrong path returns NULL, which reads exactly like "this agent has no house voice".
- **the check:** for a census, **search the whole config as text** — it cannot be fooled by nesting depth:
```sql
SELECT type FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config::text ILIKE '%<distinctive phrase>%' ORDER BY type;
```
  And before believing any single-agent NULL, **induce a non-zero**: `SELECT length(<the same path>)`. A non-NULL length is the positive control that the path is real; a NULL length means you measured your own typo, not the agent. Proven 2026-08-09: the step-level census returned **6** agents carrying the house voice and confidently excluded `page-content-writer` — the agent that writes every decomposed site's sections. Its prompt is 12,813 chars, nested under `sub_workflow`, and **does** carry it. True count **7**. The blind census was about to be written into a handoff as a correction to a figure that was right all along.
- **source:** loancalculator_couk lane, grounding the "SEVEN copies across seven agents" figure before carrying it into a new handoff — caught by running a positive control on the single-agent read, not by anything in the census output
- **added:** 2026-08-09, loancalculator_couk lane

### `output_field` nested inside a workflow step's `config` is INERT — and the reaper's own seed models the wrong form

- **footprint:** `agent_definitions.default_config->'workflow'->'steps'` authoring · `output_field` · `docs/agent_docs/sql_for_agents/028_thunder_reaper.sql` / `114_thunder_reaper.sql` (the template most task-workflows are copied from) · `pkg/models/contracts.go` `Step.OutputField` · `platform/messaging/processor.go:434` · any step whose result a LATER step must read
- **fires when:** you author a workflow seed by copying an existing one (the reaper is the canonical task-mode example) and put `output_field` inside the step's `config` block, then a later step reads `collected_data.<that name>`. The parser takes `Step.OutputField` from the STEP level only (`stepMap["output_field"]`, processor.go:434); a config-nested copy is read by nothing, so an awaited response stores ONLY under the STEP NAME and the later step's read returns missing.
- **the tell:** almost none before the failure. The reaper never trips it because nothing reads its response downstream — its config-nested `output_field` has been silently inert since it shipped, which is exactly why it looks like a working example to copy. A single live specimen can even CONFIRM the wrong model: the reaper's step name (`dispatch_decommission`) and its output_field (`decommission_dispatch`) are word-reversals, so a session reading its `collected_data` sees a key that "matches" whichever model it already believes (this file's author did precisely that, twice — once when authoring, once when "verifying" for a council answer).
- **the check:** before asserting where a step's output lands, read the STORER, not a specimen: `applyResponseToState` (coordinator.go:2636) stores under stepName always, plus step-level `OutputField` when set. When authoring: `output_field` as a SIBLING of `action`, never inside `config`. When verifying: pick a case where step name and output_field DIFFER — a specimen whose two names collide cannot discriminate.
- **source:** finetuning_uk_service lane, FTW-042 first live run (both runs FAILED at `reconcile`: `collected_data.thunder_list missing`); the council guardian seat had flagged the wiring as its low-severity objection and the round-2 answer asserted the wrong model with a non-discriminating specimen as proof
- **added:** 2026-08-09, finetuning_uk_service lane

### A `LEFT JOIN` from `agent_error_log` to `orchestration_states` drops ~97% of rows to retention, and the survivors are a RECENT-BIASED sample that looks like the whole answer

- **footprint:** `agent_error_log` · `orchestration_states` · any join between a long-lived log table and a pruned state table (`awaited_requests`, `processed_messages` have the same shape) · any question of the form "does the error row agree with its run?"
- **fires when:** you check a provenance/attribution claim by joining the error row to the orchestration that produced it — the obvious and correct-looking way to test whether `agent_error_log.agent_type` matches `orchestration_states.owner_agent_type`.
- **the tell:** the `LEFT JOIN` returns a huge bucket under `'<no orchestration row>'` — measured 2026-08-08: **550 of 555** `agent_type='generic'` rows. It reads as a finding ("these errors have no run"), and it is **retention**: `agent_error_log` spans ~30 days (20,022 rows, oldest 2026-07-09), `orchestration_states` retains ~24h (1,667 rows). The five that DO join are a sample of one day, and an inner join quietly makes that sample the entire study — mine returned 1,137 rows and 18 disagreements, which is a true statement about yesterday dressed as a true statement about the month.
- **the check:** before reading a join result, **measure both spans in the same pass and put them in the output**, so the reader cannot miss which one bounds the answer:
  ```sql
  SELECT 'agent_error_log' t, min(occurred_at)::date lo, max(occurred_at)::date hi, count(*) FROM agent_error_log
  UNION ALL SELECT 'orchestration_states', min(created_at)::date, max(created_at)::date, count(*) FROM orchestration_states;
  ```
  If the log's span exceeds the state table's, the join **cannot** answer a question about the log's history — only about its last day. Say which you asked. When the answer must cover the history, the evidence has to come from somewhere durable (the log's own columns, `agent_run_stats`) or from a **post-change measurement**, not from archaeology.
- **source:** rfc012 lane §1a (`RFC_019` / `RSH-009`), verifying whether the actions door's `generic` rows disagreed with their run. They do — but only 5 rows could be evaluated, so the claim had to be scoped to "the last day" and the real proof deferred to a post-roll baseline. Caught by asking why one bucket held 99% of the rows, not by anything failing.
- **added:** 2026-08-09, rfc012_await_findings lane

### A whole-table count on `agent_error_log` spans a MONTH, so it prices a defect its own fix already retired — the bad rows are history, and nothing in the number says so

- **footprint:** `agent_error_log` (`agent_type`, `error_code`, `action`) · `SELECT count(*) … GROUP BY agent_type` · any bug/handoff/RFC sizing a write-path defect from this table
- **fires when:** you size a provenance or classification defect the obvious way — `SELECT agent_type, count(*), count(DISTINCT step_name) FROM agent_error_log GROUP BY 1` — and quote the number as the live damage. The table keeps ~30 days, so a fix that shipped three weeks ago is invisible in it.
- **the tell:** there is none. That is the point: **555 rows across 25 distinct `step_name`s is a true count and a false claim**, and it looks identical whether the defect is raging or was fixed a fortnight ago. Measured 2026-08-08: **499 of those 555 `agent_type='generic'` rows predate 2026-07-26**, the day `RunAgentType` shipped (`baf887a8e`) — ~89% of the "live damage" was already fixed, the dominant producer (`call_agent`/`call_dispatch`, 394 rows) stops dead the day before, and the 25 `REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows quoted as evidence were **all written on one day, 2026-07-23**. A handoff, a code comment and a register entry all carried the whole-table figure forward as current.
- **the check:** never quote a bucket from this table without **splitting it at the date of the last commit that could have changed it**, and print `min`/`max(occurred_at)` per group so a dead producer is visible as a stale `max`:
  ```sql
  SELECT action, step_name, count(*), min(occurred_at)::date first, max(occurred_at)::date last
  FROM agent_error_log WHERE agent_type = '<value>' GROUP BY 1,2 ORDER BY 3 DESC;
  -- then: git log -S'<the symbol that would fix it>' --format='%h %ad %s' --date=short
  ```
  A group whose `last` is weeks old is not evidence of anything except that it stopped. The **rate** (rows/day since the last relevant commit) is the number that can be acted on; the total is the number that gets quoted.
- **source:** rfc012 lane §1a — the handoff commissioning the work sized it at "559 rows across 25 distinct `step_name`s, the widest spread of any `agent_type`". Re-measured before building, which turned a volume argument into a structural one and moved the honest residue to ~36 rows in 13 days. Nothing failed; the figure was simply never bucketed.
- **added:** 2026-08-09, rfc012_await_findings lane

### `evidence_base.facts[]` is bookkeeping; `writer_block` is the wire — a fact not copied into writer_block does not exist for the content writer, and nothing warns about the divergence

- **footprint:** `site_specs` aspect `evidence_base` (keys `facts` vs `writer_block`), `page-content-writer` prompt template ("## Verified Facts" section), `llm_call_log.prompt_rendered`
- **fires when:** you register a new fact in `evidence_base.facts[]` (with claim/value/writer_line, properly sourced) and expect the writer to state it. The writer prompt's "Verified Facts (the ONLY numbers … you may assert)" section renders from **`writer_block`** — a hand-composed PROSE key — not from `facts[]`. A fact present only in `facts[]` never reaches any writer prompt; the copy stays vague or omits the number entirely, while every spec-level check you run says the fact is registered.
- **the tell:** rounds of rebuilds where the writer "refuses" to state a registered number despite rules instructing it. Measured 2026-08-09 (webdesign.uk caps): THREE full page rounds with the facts registered, identity concretised and an imperative writing_rule — zero statements; `prompt_rendered` showed the Verified Facts header followed immediately by style prose, with the numbers absent from the whole 24.5KB prompt. One append of the terms to `writer_block` → next round stated them on all four pages, first pass. **Beware the false check:** grepping the ORCHESTRATION's collected_data for the fact text returns true (the spec rides along in memory) — only `llm_call_log.prompt_rendered` shows what the model saw.
- **the check:** after adding a fact the copy must state, append its writer_line (with any imperative) to `evidence_base.writer_block` as well — and verify at `SELECT prompt_rendered FROM llm_call_log … ` that the number appears in an actual writer call before blaming the model. The validator's `unregistered_stat` check DOES read fact values, so the asymmetry is silent: stats are blocked against `facts[]` but written from `writer_block`.
- **source:** 2026-08-09, webdesign_uk_build_service lane; NOTES 08-09 (5).

### `render_audit.py`'s total is not the figure to quote — a homepage run understates ~100×, and ~9% of the rows are the probe's own guess

- **footprint:** `scripts/render_audit.py`, `internal/adapters/browserrunner/render_audit_action.go`, `write_render_audit_findings`, `contrast_failure`
- **fires when:** running the render audit, or repeating any per-site contrast figure from `bugs_open/113`, `bugs_open/122` or a `contrast_failure` work item
- **the tell:** none in either direction. The tool prints a confident total and a clean per-page table; both wrong numbers look exactly like right ones, and the two errors push opposite ways so they do not cancel
- **the check:** two arms, both cheap.
  1. **Pass `--sitemap`, and say whether you did.** Without it the tool renders the one URL you named. Measured 2026-08-09 on the same sites the same week: `robot-hands.com` 3 failures homepage-only → **193** full sitemap (19 pages); `dartsonline.com` **1 → 125**. Nothing regressed between the runs — the defects live on tool, guide and news pages, and a homepage never opens them. A per-site figure with no `--sitemap` is a claim about `index.html`.
  2. **Discount `overImage` rows before totalling.** `render_audit.py:111-114` pushes a mid-grey `rgb(128,128,128)` under any text whose backdrop is a background *image or gradient*, because the real colour is unknowable, and sets `overImage: true` — its own comment says this is "so a reader can discount it". Any `on rgb(128,128,128)` row, classically `rgb(255,255,255) on rgb(128,128,128) = 3.95:1`, is the probe's placeholder and not a measurement. It was **41 of 483 rows (8.5%)** across three sites, concentrated wherever `--color-cta-bg` is a `linear-gradient`. `[c.get('overImage') for c in page['contrast']]` in the `--json` output is the filter; the terminal output does not mark them.
- **source:** hit directly, `bugs_open/113` three-site sitemap audit, 2026-08-09; both figures recorded in 113 and 122
- **added:** 2026-08-09, brochure/113 lane

### Four agent types have TWO active definition rows, and only the higher `version` is ever loaded — a config change applied to "the" row can be applied to the dead one

- **footprint:** `agent_definitions`, `loadAgentDefinitionForAction`, `platform/orchestration/actions/ai_actions.go`, any `UPDATE agent_definitions … WHERE type = '<x>'`, `snapshot_agent`
- **fires when:** you change live agent config by hand or by migration — a cap, a model, a prompt, a step's config — on a type that happens to have two rows. Measured 2026-08-09: `chief-strategist`, `content-creator`, `content-creator-contact`, `site-component-architect` (4 of ~180).
- **the tell:** none. `UPDATE … WHERE type = 'chief-strategist' AND is_active AND NOT is_snapshot AND deleted_at IS NULL` reports `UPDATE 2` — and a guard asserting "1 row" fails while a guard asserting "the value is set" passes. Pick the wrong single row and every later read of *that* row confirms your change, while the fleet runs the other one. `chief-strategist` is the live example: version 1 carries a `generate_build_plan` cap of 8192 chosen 2025-11-15, version 2 carries 16000 — **only version 2 is ever loaded**, so the 8192 has been inert for nine months and reads exactly like live config.
- **the check:** two arms.
  1. **Count before you write, and scope to what LOADS.** The loader is `WHERE type = $1 AND is_active = true ORDER BY version DESC LIMIT 1` (`ai_actions.go:1313`) — newest version wins, deterministically (no version ties exist fleet-wide, verified 2026-08-09). So either write EVERY active row (safe, what `sql_for_agents/347` did) or scope explicitly to the max-version row; never `LIMIT 1` yourself and never assume one row. ```sql SELECT type, count(*) FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL GROUP BY 1 HAVING count(*) > 1; ```
  2. **Know that the loader's predicate is NARROWER than every census in this estate** — it filters `is_active` alone, with no `is_snapshot` and no `deleted_at` arm. An `is_active` snapshot or soft-deleted row would win on version and be loaded silently. That population is **empty today** (0 active snapshots, 0 active-but-deleted, of 183) and is held shut only by `snapshot_agent()` writing `is_active = false` — an invariant nothing states or tests, so re-measure it rather than trusting this line: ```sql SELECT count(*) FILTER (WHERE is_active AND COALESCE(is_snapshot,false)) AS active_snapshots, count(*) FILTER (WHERE is_active AND deleted_at IS NOT NULL) AS active_but_deleted, count(*) FILTER (WHERE is_active) AS loader_visible FROM agent_definitions; ```
- **note:** RFC_006's `single-owner-carriers-check` does NOT cover this. It counts how many distinct AGENTS carry one single-owner ACTION; this is one agent TYPE carrying two definition rows. Different failure, no detector, unowned.
- **source:** found while verifying `bugs_open/205`'s cap migration landed on the loaded row, 2026-08-09; NOTES + WRONG_CALLS in `bugfix_205_poison_pill_reaper/`
- **added:** 2026-08-09, bugfix_205 lane

### A page's served URL is NOT derivable from `pages.name` — and the 404 you get from guessing is indistinguishable from the defect you are hunting

- **footprint:** `pages.url` vs `pages.name` · any `curl`/`curl -sI` that verifies a page deployed, a link resolved, or a fix shipped · acceptance sections of `bugs_open/` files that quote a URL
- **fires when:** you verify anything at the served artefact — which this estate tells you to do at every turn — and build the URL from the page's `name` because that is the column every work item, spec and log line carries. The name is an identifier, not a path. Measured 2026-08-09 on two different sites in one session: `beginners` serves at **`/blog/beginners.html`**; `guide-first-time-buyer` serves at **`/guides/first-time-buyer/index.html`**. Neither is `/{name}` or `/{name}.html`, and blog posts, guides, entity pages and section indexes each nest differently.
- **the tell:** **there is none, and that is the whole trap.** A guessed URL returns `404` — the exact signal that means "this page was never deployed", which is very often the precise thing you are testing for. It cost this lane a moment of believing a live container page had vanished, on the same day the lane's own notes had just warned about the beginners spelling. A 200 is self-proving; **a 404 proves nothing until you have read `pages.url`.**
- **the check:** ask the DB for the path, never compose it — and check the container and the target in the same breath so a typo cannot masquerade as a finding:
  ```sql
  SELECT name, url, build_status, COALESCE(deployed_at::text,'NEVER') AS deployed
  FROM pages WHERE site_id = '<site>' AND name IN ('<container>','<target>');
  ```
  Then `curl` the `url` values verbatim. If you are asserting a dead link, the honest pair is **container 200 + target 404 + the href present in the container's served bytes** (`grep -c 'href="<the exact href>"'`); any of the three missing and you have not shown a dead link. Note the href in `site_work_items.spec->>'href'` is already the real path — it is the one string in the whole flow you do NOT have to derive.
- **source:** 2026-08-09, bugfix_220 lane — the residue census in `bugs_open/220` (§ 2026-08-09 midday), where the guessed container URL 404'd and briefly read as evidence the page was gone; the correct URL served 200 and carried the dead link, which is what made the finding real
- **added:** 2026-08-09, bugfix_220_unbuilt_link_dispatch lane

### The deploy-stamp guard does NOT cover `page-rerender` — that path's skip flag is a different key, and its only protection is a workflow conditional

- **footprint:** `page-rerender` (agent_definitions), `check_skipped`, `rendered_page`, `upstreamAssemblySkipped`, `platform/orchestration/actions/owned_page_guard.go`, `rerender_single_page_action.go`, `update_page_status`, PBP-038, `bugs_open/210`
- **fires when:** you edit the `page-rerender` workflow, or you reason about where a page can be stamped `deployed` after work that did not happen, and you carry over PBP-038's headline — "`UpdatePageStatusAction` refuses the stamp on ANY assembly skip" (`bugs_open/210`, live v1.0.1268). It is true of the three *build* loops and false of the rerender path, which uses the same `update_page_status` action.
- **the tell:** none at the action. `upstreamAssemblySkipped` reads **`collected_data["assembled_page"]` and nothing else** (`owned_page_guard.go:308-319`). `page-rerender` renders via `rerender_single_page` into **`rendered_page`** (`output_field`, measured 2026-08-09), and that action emits `{"skipped": true, "html": ""}` on "no component rows" (`rerender_single_page_action.go:198-209`). Same skip, different key, so the code guard cannot see it and the `deployed` stamp would be written normally — the exact outcome bug 210 exists to prevent, on a path whose docs now say it is prevented.
- **what actually holds it shut:** a **config-level** conditional in the workflow — `render_page → check_skipped` (`rendered_page.skipped == true` → `complete_skipped`, else → `deploy_page`). Present and correct in the live definition, verified 2026-08-09. Delete it, rename the render step's `output_field`, or add a second route into `update_status`, and there is no code guard underneath: the page is stamped `deployed`, `built_from_plan_version` is set, and reconcile's `decideEmit` returns `skip_built` for ever. That is 210, reproduced, on the fleet's busiest page path.
- **the check:** before changing that workflow, or before citing PBP-038 as fleet-wide cover, ask which KEY the step writes and whether the guard reads it:
  ```sql
  -- every live route into update_page_status, and the skip key its predecessor writes
  SELECT ad.type, e.path, e.step->>'action', e.step->>'output_field'
  FROM agent_definitions ad,
    LATERAL jsonb_path_query(ad.default_config,'$.**.steps') AS steps,
    LATERAL jsonb_each(steps) AS e(path, step)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND (e.step->>'action' IN ('update_page_status','assemble_page','rerender_single_page'));
  ```
  Guard cover follows `output_field = 'assembled_page'`, NOT the presence of `update_page_status`. Today: covered = `page-rebuild`, `pageflow-builder`, `site-work-orchestrator`; **uncovered = `page-rerender` (config only), `section-editor`** (commits `edit_result.html`, writes no skip flag at all).
- **the strongest single line of evidence, added on the human pass:** `grep -rn '"rendered_page"' platform/orchestration/actions/*.go` (excluding tests) returns **nothing**. No Go code anywhere in the actions layer reads that key — it is consumed only by the workflow conditional. So the config IS the guard, with no code path behind it, which is the whole claim in one command.
- **source:** 2026-08-09, bugfix_210 lane — found while establishing which paths can reach the guard at all, after the behavioural canary was stood down; NOTES § "A scope boundary nobody had written down"
- **verification:** landmine-verifier (corr `d65dbd74`) returned **NEEDS_HUMAN_REVIEW** — not a contradiction: it resolved every symbol and found nothing against the entry, but could not do a line-level pass because the code index was ~2 days stale and it failed to resolve `rerender_single_page_action.go` (its own path-prefix issue — the file is present, 41KB). **Human pass done the same hour, first-hand at the tree:** guard read `owned_page_guard.go:309`; skip emitter `rerender_single_page_action.go:200`; the grep above. Not re-dispatched, because the blocker was index staleness rather than anything about the entry.
- **added:** 2026-08-09, bugfix_210_content_failed_build_stamped_deployed lane

### A census of live step config written as `->'workflow'->'steps'` is TOP-LEVEL ONLY, and it returns a confident wrong number rather than an error

- **footprint:** `agent_definitions`, `default_config`, `jsonb_each(default_config->'workflow'->'steps')`, `sub_workflow`, `substeps`, `validation.WalkSteps`, `platform/validation/subworkflow.go`, `cmd/config-key-audit`, `bugs_open/144`, `bugs_open/136`
- **fires when:** you ask "which live agents carry config key X?", "which steps use action Y?", or "how many places do I have to change?" — and you write, copy or inherit the obvious SQL. Nearly every runbook in this estate contains a version of it, because it is the natural way to write the query and it is right about most definitions.
- **the tell:** **there is none.** A step nested in a loop's `sub_workflow.steps` (or `substeps`, which WINS at execution when both are present) is simply not in the result set. No error, no NULL, no empty result — a plausible number, arrived at early, that you then plan the work against. Measured 2026-08-09 on `bugs_open/136`'s own key census: the query reported **13** carriers of three deprecated config keys; there were **19**. Six lived one level down, in `component-quality-auditor`, `internal-linker`, `tool-auditor` (×2) and `tool-suggester` (×2). A 32% undercount, in a runbook whose entire subject is finding every carrier of a key.
- **why it survives:** the Go side was FIXED — `validation.WalkSteps` was extracted precisely to abolish this descent (`bugs_open/144`: "two hand-written traversals blind in the same direction, agreeing with each other"), and `cmd/config-key-audit` adopted it on 2026-07-29 and now prints its own coverage banner (`68 pairs inside loop sub-workflows, 25 of which exist ONLY there`). **The SQL in the runbooks did not learn about that fix.** A query written before a platform fix keeps the bug for ever, and it is read as authoritative because it is written down.
- **the check:** never trust a single-depth descent for a count you will act on. Cheapest honest census is a text scan **carrying a positive control**, so a zero is readable and a disagreement is visible:
  ```sql
  SELECT count(*) FILTER (WHERE default_config::text ~ '<the key>')      AS hits,
         count(*) FILTER (WHERE default_config::text ~ '<a key you KNOW is present>') AS pos_control,
         count(*) FILTER (WHERE default_config::text ~ 'zzz_invented')   AS neg_control
  FROM agent_definitions
  WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active;
  ```
  If the text count of *definitions* exceeds the step query's count of *agents*, you have found nested carriers — that discrepancy is what caught this one. For the exact paths (which you need in order to write a `jsonb_set`), recurse; note Postgres refuses two references to the CTE, so objects and arrays must be folded into ONE recursive term:
  ```sql
  WITH RECURSIVE walk(agent_type, path, node) AS (
    SELECT ad.type, ARRAY[]::text[], ad.default_config FROM agent_definitions ad
     WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
    UNION ALL
    SELECT w.agent_type, w.path || e.k, e.v FROM walk w CROSS JOIN LATERAL jsonb_each(
      CASE jsonb_typeof(w.node) WHEN 'object' THEN w.node
        WHEN 'array' THEN COALESCE((SELECT jsonb_object_agg((i-1)::text, v)
            FROM jsonb_array_elements(w.node) WITH ORDINALITY a(v,i)), '{}'::jsonb)
        ELSE '{}'::jsonb END) AS e(k,v))
  SELECT agent_type, array_to_string(path,' > '), node #>> '{}' FROM walk
  WHERE path[array_length(path,1)] = '<the key>' ORDER BY 1,2;
  ```
  From Go, call `validation.WalkSteps` and inherit the fix instead of re-deriving the bug. And if a migration must touch every carrier, **drive it from the recursive walk rather than a hand-typed list** — a list is a snapshot of whatever your census could see.
- **source:** 2026-08-09, bugfix_136_config_key_aliases lane — `bugs_open/136` §12 and the lane NOTES; the corrected query now lives in `RUNBOOK_config_key_aliases.md`
- **added:** 2026-08-09, bugfix_136_config_key_aliases lane

### Renaming a key in a SEED can break a LATER migration that deletes that same key by name — and the replay silently restores the defect the later migration removed

- **footprint:** `docs/agent_docs/sql_for_agents/`, `051_build_dispatch_loop.sql`, `052_build_pipeline_trigger.sql`, `#- 'key'`, `? 'key'`, seed replay, `bugs_open/134`, `bugs_open/136`
- **fires when:** you follow the (correct, standing) rule "fix the live row AND the seed in the same commit, so a replay cannot reintroduce the dead key" (`bugs_open/134`), and you do it with a grep-driven rename across the seed directory. The rule is right; applying it file-by-file without reading forward is what bites.
- **the tell:** none at the time — the rename is a 1:1 line edit, `git diff --numstat` shows added == deleted, the SQL still parses, and the live DB is untouched by seed files. The damage only appears on a full replay, months later, in a definition nobody was thinking about.
- **the case:** `051` seeds `build-dispatch-loop` with `"item_domain": "build"` on `load_next_item` and `check_remaining`. `052` **deletes exactly those keys by name** — `(config) - 'item_domain'`, guarded by `... ? 'item_domain'` — because the filter was a defect: it meant work items on any pipeline other than `build` were never dispatched. Renaming `051`'s spelling to `item_pipeline` leaves `052` matching nothing, so a replayed chain keeps a live pipeline filter that `052` exists to remove. Caught before commit by asking "does anything downstream key on this exact spelling?"; six lines reverted, `051` left untouched.
- **the check:** before renaming a key in any seed, grep the WHOLE seed directory for that key as a **predicate**, not just as data — the two spellings look nothing alike:
  ```bash
  grep -rn "'<key>'" docs/agent_docs/sql_for_agents/*.sql | grep -E "\?|#-|\- '|jsonb_set|delete"
  ```
  A hit on a HIGHER-numbered file means the chain is keyed on the old spelling and your rename breaks it. The general rule: **a seed is one frame of a chain, not a standalone statement of intent** — the next frame may be keyed on the exact string you are tidying. When in doubt, leave the historical pair internally consistent and fix only the live row.
- **source:** 2026-08-09, bugfix_136_config_key_aliases lane — lane NOTES § "The near-miss"; migration `349`'s header records why `051` is deliberately excluded
- **added:** 2026-08-09, bugfix_136_config_key_aliases lane
