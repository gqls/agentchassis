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

### A page whose `url` is `/` deploys a file with NO NAME — `getPageInfo` derives the filename by trimming the leading slash
- **footprint:** `pages.url`, `platform/orchestration/actions/rerender_single_page_action.go` (`getPageInfo`, `PageInfo.Filename`), `platform/orchestration/actions/adopt_verbatim.go` (`urlToDeployPath`), any hand-written `INSERT INTO pages` or `UPDATE pages SET url`
- **fires when:** you set a homepage's `url` by hand, or write any code that derives `pages.url` from a crawled/collected URL. A crawler reports the site root as `"/"` or `"https://host/"`, and `"/"` is the obvious, correct-looking value to store for a homepage
- **the tell:** none at write time — the row is valid, the constraint passes, the page looks right in every `SELECT`. The filename is computed downstream as `strings.TrimPrefix(p.URL, "/")`, so `"/"` yields `""`, and the deploy commits a file with an empty name into the site's directory. `getPageInfo` has a guard (`p.URL == "/" || p.Name == "index"` → `index.html`) which is why this has not bitten in the normal path — but the guard keys on the page being NAMED `index`, so an adopted or hand-made homepage under any other name (`home`, `landing`, `main`) falls straight through it
- **the check:** never store a bare directory URL. Normalise `"/"` and any trailing-slash or extensionless path to its `index.html` **before** it reaches `pages.url`, and assert the invariant rather than the string: `strings.TrimPrefix(url, "/") != ""`. In SQL: `SELECT name, url FROM pages WHERE url = '/' OR url = '' OR url LIKE '%/';` should return no rows
- **why it is a landmine and not a bug report:** the value is not wrong, it is wrong *for one downstream consumer three files away*, and the failure is a silently misnamed artefact rather than an error. `urlToDeployPath` now does this normalisation for verbatim adoption and its test asserts the empty-filename invariant directly — but anything writing `pages.url` by another route still needs the check
- **source:** built during `fidelity=locked` verbatim adoption, concept register `ADO-037`, 2026-07-30. Mutation-verified: forcing the root case to return `"/"` fails `TestURLToDeployPath` on the invariant, not just the expected value
- **added:** 2026-07-30, loancalculator_couk lane

### The adoption crawl index stores ONE page under SEVERAL keys — dedupe it by content pointer, never by URL string
- **footprint:** `platform/orchestration/actions/apply_adoption_plan_action.go` (`buildCrawlPageIndex`, `matchCrawlContent`), `platform/orchestration/actions/adopt_verbatim.go` (`crawlPathIndex`), `research_results` rows with `result_type='adoption_crawl_page'`
- **fires when:** you iterate the crawl index to enumerate a crawled site's pages — the natural move when you want "the list of pages we found", rather than looking a single page up by URL
- **the tell:** a plausible-looking page count that is a small multiple of the truth. `buildCrawlPageIndex` deliberately registers the SAME `*crawlPageContent` under every alias it can (absolute URL, path-only form, and `metadata.sourceURL` when it differs) so `matchCrawlContent` finds the page however the LLM plan spells it. Iterating those keys therefore yields 2–3 entries per real page, each with identical content — and since every entry is genuinely valid, nothing errors. A 27-page site reads as 60–80 pages, and if you create a row per entry you get duplicate pages that differ only in URL form
- **the check:** dedupe by the map VALUE's pointer identity, not the key — `seen := map[*crawlPageContent]string{}`. Then pick the surviving alias **deterministically** (sort the keys first): Go map iteration order is randomised, so without a sort a re-run legitimately picks a different alias and writes a different `pages.url` for the same page. Sanity-check the reduced count against the site's own `sitemap.xml` before creating anything
- **source:** built during `fidelity=locked` verbatim adoption, concept register `ADO-037`, 2026-07-30
- **added:** 2026-07-30, loancalculator_couk lane

### `snapshot_agent` has two overloads writing to different places, and the backup table keeps the SOURCE row's `created_at` — so "the newest snapshot" ordered by `created_at` is a tie, not a row
- **footprint:** `snapshot_agent`, `agent_definitions_backup`, `agent_definitions.is_snapshot`, any `sql_for_agents/*.sql` migration with a ROLLBACK block
- **fires when:** you write the rollback instructions for an agent-config migration. The migration itself is fine — this bites the person restoring, i.e. at the worst possible moment
- **the tell:** none at write time, and the wrong instruction reads as completely reasonable. Two traps compose. (1) `snapshot_agent(text)` and `snapshot_agent(text, text)` are **two overloads**; the 2-arg form (the one every migration here calls) writes to **`agent_definitions_backup`**, NOT to an `is_snapshot` row in `agent_definitions` — so a rollback told to find "the newest `is_snapshot` row" searches the wrong table and finds nothing, or finds an unrelated row. (2) That backup copies `id`, `created_at` and `updated_at` **verbatim from the source row**, so `ORDER BY created_at DESC` does not order snapshots at all. Measured 2026-07-30: all three `feature-designer` backups share source `created_at` `2026-07-17 18:06:05` — a three-way tie. Only **`snapshot_taken_at`** discriminates, and `snapshot_reason` is what identifies which migration took it
- **the check:** `SELECT snapshot_taken_at, snapshot_reason FROM agent_definitions_backup WHERE type='<agent>' ORDER BY snapshot_taken_at DESC LIMIT 5;` — if that returns your migration's reason at the top, your rollback block is right. Restore with `UPDATE agent_definitions a SET default_config = b.default_config FROM (SELECT default_config FROM agent_definitions_backup WHERE type='<agent>' AND snapshot_reason='<your reason>' ORDER BY snapshot_taken_at DESC LIMIT 1) b WHERE a.type='<agent>' AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false;`. **Also: a dry run does not leave a snapshot** — `snapshot_agent` inside a transaction you `ROLLBACK` is rolled back with everything else, so a clean dry run is not a backup
- **known-affected:** `272_feature_designer_plan_repair_loop.sql` (corrected 2026-07-30) and `222_feature_designer_one_edit_per_file_per_stage.sql` (**still carries the wrong instruction** — another lane's applied migration, not edited; this entry is the warning)
- **source:** the council gate's `debug_historian` seat, objecting on corr `f4a4628f-3b90-4054-a875-f2cf72b83e72` that the "snapshotted, verified, reversible" claim was unconfirmed against a documented double-overload trap. It was right, and the first version of 272's rollback block was wrong
- **added:** 2026-07-30, bugfix_099 lane

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
