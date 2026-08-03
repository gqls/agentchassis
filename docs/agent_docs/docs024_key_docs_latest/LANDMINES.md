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
> **Still true in BOTH worlds, and the reason this entry is not deleted:** `og-card.png` is not derivable from `og_card`, so `BrandHeadAssetPaths` remains the one declaration — it became the derivation's INPUT, it was not collapsed away. And `deploy_path` still overrides everything and is invisible from `(asset_key, purpose)`.

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
