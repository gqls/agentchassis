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
- **footprint:** `nav-updater` (agent), `populate_nav_tables` (its first step), `platform/orchestration/actions/populate_nav_tables_action.go` (`classifyPagesForNav`), `site_nav_items`
- **fires when:** you want a nav or footer change to reach the site and reach for the agent whose name says navigation. It is the obvious choice, it completes cleanly, and it also re-assembles every deployed page so the damage ships immediately
- **the tell:** there is none at the time. The run COMPLETES, the nav still looks populated (primary nav is rebuilt correctly), and only the child-path links are gone — a footer that quietly stops listing the site's tools. `populate_nav_tables` does `DELETE FROM site_nav_items WHERE site_id = $1` and rebuilds from `pages`, and `classifyPagesForNav` **skips a page entirely** when `isChildPageURL(url)` matches (`/tools/`, `/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`, `/resources/`, `/insights/`) unless `isSectionIndexType(page_type)` — which is only `blog-index`, `entity-directory`, `section-index`, `news-index`. **`tool` is not one of them**, so the skip happens BEFORE the `neverPrimaryTypes` → utility branch that would otherwise have kept it
- **the check:** before running it, count what you are about to lose —
  `SELECT s.domain, count(*) FROM site_nav_items ni JOIN sites s ON s.id=ni.site_id WHERE ni.status='active' AND (ni.url LIKE '/tools/%' OR ni.url LIKE '/blog/%' OR ni.url LIKE '/guides/%' OR ni.url LIKE '/articles/%' OR ni.url LIKE '/case-studies/%' OR ni.url LIKE '/news/%' OR ni.url LIKE '/resources/%' OR ni.url LIKE '/insights/%') GROUP BY 1;`
  Measured 2026-07-30: **16 rows across 7 domains** would be deleted — leopardess 5, gamesdesign 3, idea.uk 2, dartsonline 2, webdesign 2, fundamentallyai 1, robot-hands 1. Those rows cannot have been created by `populate_nav_tables`, so nothing recreates them
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
