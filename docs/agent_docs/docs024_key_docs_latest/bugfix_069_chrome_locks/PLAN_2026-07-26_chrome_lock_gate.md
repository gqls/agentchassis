# PLAN — 2026-07-26 — chrome (`site_components`) lock gate, `bugs_open/069`

## What this workstream is

`bugs_open/069`: the chrome table `site_components` carries the same human-lock columns as
`page_components` (`locked_at / locked_by / lock_type / lock_expires_at`, same 053/115 migration,
same CHECK `permanent|timed|review`), the admin dashboard sets them, and **no chrome writer reads
them back**. It is the residual `bugs_open/058` deliberately spun out to keep that reviewed change
narrow. 058 is CLOSED and live on v1.0.1165; its machinery is what this fix reuses.

A second defect surfaced during the writer audit and the owner ruled it in scope for this session:
**a snapshot revert silently destroys every lock** on both tables (`bugs_open/085`). Filed and fixed
separately — it is a DB-function change, live immediately, not part of the Go commit.

## Ownership check (done before touching anything)

- `scripts/who-owns.py 069` says "OWNED or recently active", but the only commit touching the file is
  `2c7fd3be9` — the 058 fix record that *filed* 069. 058 closed this morning (`3f66627bf`) stating
  069 is a separate bug, not its residual. Nobody is on it.
- No open `site_work_items` row mentions chrome locks (`status NOT IN (complete,cancelled,rejected)`).
- The 049 thread edited `render_site_components_action.go` today up to 16:30 and closed at 17:43. Its
  work is committed. **Re-grep the writer set immediately before committing** — 058's council caught a
  writer that landed nine minutes before its fix commit.

## Live measurements that set the framing (2026-07-26)

| measure | value |
|---|---|
| `site_components` rows / locked | 42 / **0** |
| `page_components` rows / locked / unstamped `lock_type` | 931 / 39 / 0 |
| `lock_blocked_change` items since 058 shipped | **0** |
| admin chrome edits ever (`spec->>'reason'='admin_site_component_edit'`) | **0** |
| snapshots taken / last | 11 / 2026-06-24 |

So 069 is preventative: the path is reachable and ungated, but has never been exercised. Verification
must **induce** the fault; a green run proves nothing on its own.

## Decisions and their reasons

1. **Reuse 058's predicate verbatim** (`pageComponentAgentWritableSQL`) rather than write a chrome
   one. It emits bare column names and both tables share the column set; its own doc comment already
   says so. Not renamed despite now being a misnomer across four tables — six call sites in
   concurrently-owned files for a cosmetic gain.
2. **The gate sits BELOW the `!force` early exit** in `renderAndStoreSiteComponent`. Two of the six
   live agents call with `force_rerender: false`, and that path already no-ops on a populated slot.
   Checking above it would file a work item claiming a writer "wanted to change this" for a call that
   was never going to write. Consequence, stated plainly: **the gate only bites on
   `force_rerender: true`**, so the live proof must force or it passes vacuously.
3. **Uniform blocking, including a locked slot whose `rendered_html` is empty** — which freezes that
   slot site-wide until a human unlocks, because `getSiteComponents` filters empty HTML. A softer
   policy cannot be expressed in the WHERE predicate and would reopen the TOCTOU the predicate exists
   to close. The item carries `artefact_empty: true` so the human sees it.
4. **Skip-result, never an error.** An error would retry the orchestration against a state only a
   human unlock can change (058 took the same line).
5. **A struct-based emitter core, not more positional arguments.** `emitLockBlockedChangeItem`
   already takes 12 positional params, five adjacent bare strings; chrome needs `surface`,
   `severity`, its own `item_key`, its own `fix` text and extra spec fields. The existing signature
   stays as a wrapper, so its four other call-site files are untouched.
6. **`item_key = lock_blocked_change:site_component:<slot>`**, not `…:chrome:<slot>` — the page-side
   key is `lock_blocked_change:<pageName>:<slot>`, so `chrome` would collide with a page named
   "chrome".
7. **085's semantics: a revert restores content; it never locks or unlocks anything.** Restoring the
   snapshot's lock state as-captured would silently release a lock added after the snapshot — the
   very class being fixed.

## Phasing

1. Helpers (`lock_helpers.go` surgical edits + new `site_component_lock_guard.go`).
2. The four writers + the paired detector predicate.
3. Unit tests on a `git archive HEAD` overlay (the shared tree has ~20 files modified by others).
4. Re-grep, commit by pathspec, submit to the council gate (advisory, ~30 min).
5. `bugs_open/085` + migration + live verification + close 085.
6. Build v1.0.1168 from the commit, roll, induced-fault probe, close 069.

## Corrections to the originating brief (`bugs_open/069`)

> **CORRECTION 2026-07-26:** the handoff says `render_site_components`'s INSERT "is birth-only". It
> is not — it is `ON CONFLICT (site_id, slot_name) DO UPDATE SET component_id = $3`, so it repoints
> an existing locked row at a *generic default* template, and it discards both its error and its
> result. That branch (a slot with `component_id IS NULL`, a real detected state) is the worst case
> of the bug, not an exempt one.

> **CORRECTION 2026-07-26:** the handoff lists three writer files. There is a fourth mutation path —
> `revert_site_to_snapshot` — which deletes and re-inserts every chrome row for a site. It is out of
> 069's scope (human-initiated, and it hits `page_components` identically) and is filed as
> `bugs_open/085`.
