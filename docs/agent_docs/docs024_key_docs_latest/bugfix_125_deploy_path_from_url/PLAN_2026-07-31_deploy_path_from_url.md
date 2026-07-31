# PLAN — bugs_open/125: the deploy path must come from `pages.url`

**Opened 2026-07-31** by the "bugfix 8" session, taking the bug handed off by the
bug-sweep thread (`bug_backlog_clearing/HANDOFF_2026-07-28_bug_sweep_continue_here.md`
§4a: *"pre-work is DONE, the fix is not written"*, and §6 NEXT item 1).

Ownership checked two ways before starting, after a collision earlier the same session
(see `WRONG_CALLS.md`, 2026-07-31): `scripts/who-owns.py 125` names
`bug_backlog_clearing` (2 commits/14d, last 07-28) whose handoff explicitly hands this
on; and a grep of every live session transcript for the **code symbols**
(`resolveFilePath`, `ensureHTMLExtension`, `git_deployer_actions`) found no session
reading the function.

## The defect

`determinePageFilename` (`platform/orchestration/actions/git_deployer_actions.go:374`)
resolves the file a single-page commit writes to. Given the page object it tries
`slug` → `name` → `page_name` → `filename` → `id`. **`url` is not in the list at any
priority.** The page object it is handed carries `url` — the canonical path — and the
resolver discards it in favour of a path synthesised from `name`.

Three live agents share it: `pageflow-builder`, `page-rebuild` (both `page_field:
current_page`) and `site-work-orchestrator` (`current_item.spec`, which
`write_build_items` documents as carrying "the same fields as `current_page`").

## Re-measured 2026-07-31 (the filed figures were 07-28 and have moved)

```sql
SELECT count(*) FILTER (WHERE url <> '/'||name||'.html') AS wrong_path,
       count(*) FILTER (WHERE url =  '/'||name||'.html') AS right_path,
       count(*) AS total
  FROM pages WHERE url IS NOT NULL AND url <> '';
--  wrong_path | right_path | total
--         316 |        156 |   472      (filed as 280 | 151 | 431 on 07-28)
```

**67% of pages would deploy to the wrong path.** Still valid, and larger.

## What the pre-work said, and the two things it got wrong

The handoff's pre-work is good and saved real time, but two of its conclusions do not
survive contact with the data:

1. **"The fix must strip `#…`" — stripping is the WRONG repair, and would be actively
   destructive.** The single fragment-bearing row is `idea.uk` / `tool-audience-check`
   → `/tools.html#audience-check`. Stripping the fragment yields `/tools.html`, and
   `/tools.html` **is a different page's canonical URL** (`idea.uk`/`tools`, measured).
   A rebuild of `tool-audience-check` would then overwrite the `tools` page's file. A
   URL with a fragment points *into* another page; it does not designate a file, so
   the resolver must decline it and fall back — not sanitise it.
2. **A leading `/` cannot be passed through.** `pages.url` is site-absolute on
   **472/472** rows, and the git adapter builds `data.Domain + "/" + path`
   (`internal/adapters/git/github_client.go:69`), so `/tools/x.html` becomes
   `example.com//tools/x.html` — a `//` in a GitHub tree path. The pre-work does not
   mention it. Every existing path in this code is repo-relative and unprefixed
   (`assets/css/styles.css`), which is the convention to match.

## Why the fix is a shared helper — FIVE copies of this derivation exist

The bug's candidate 1 is "three lines in `determinePageFilename`". That closes the
instance; it leaves the class. Grepping the derivation rather than the symptom found
**five** places that turn a page into a deploy path, and — this is the useful part —
**four of them already consult `url` first and get it right**:

| site | consults `url`? |
|---|---|
| `datahelpers/file_extractor.go:194` `determineFilename` | **yes** — "Try url field first" |
| `rerender_single_page_action.go:521` | yes |
| `get_pages_for_rerender_action.go:176` | yes |
| `rerender_pages_actions.go:324` | yes |
| **`git_deployer_actions.go:374` `determinePageFilename`** | **NO — the bug** |

So this is not a missing feature, it is a **duplicated classifier that drifted**, which
016b §9 already names: *"when you find the same predicate twice, the fix is one
exported list plus a lockstep test, not two careful edits."* Two functions eleven
characters apart in name — `determineFilename` (correct) and `determinePageFilename`
(wrong) — live in one repo, and the wrong one is the one the three build pipelines
reach. That is a landmine on its own and is filed as one.

The four "correct" copies are not identical either: none guards the fragment case
(all four would write `tools.html#audience-check.html` for that row), none handles a
directory-style URL, and three would turn `/foo.php` into `foo.php.html`.

## Fix shape

**One definition, in `datahelpers` (imported by every call site):**

- `PageFilePathFromURL(rawURL) (string, ok bool)` — URL → repo-relative path.
  `ok == false` means "this URL does not designate a file of its own"; the caller
  falls back to its own chain rather than inventing a path. Rules, each with a
  measured reason: reject `#`/`?` (the fragment row above), reject `://` and `\`,
  strip the leading `/` (the adapter's `domain + "/" + path`), `"/"` → `index.html`,
  trailing `/` → `+index.html`, reject anything `path.Clean` rewrites (`..`, `//`),
  append `.html` only when the final segment has no extension (a URL that names
  `.php` is authoritative — rewriting it would 404 the canonical URL).
- `PageDeployFilename(rawURL, name)` — the URL-then-name form the three rerender
  sites want, fallback included.

**Call sites changed:** `determinePageFilename` (the bug, uses the first form because
its fallback chain is richer), `determineFilename`, and the three rerender
derivations.

## Blast radius of the change itself, measured before writing it

- 471 of 472 URLs resolve **byte-identically** to today's rerender-path output, so the
  live rerender lane is unaffected except for the one fragment row (which today
  produces `tools.html#audience-check.html`; not present on the live site — 404 — so
  it is latent there too).
- 0 pages named `index`/`home` carry a non-`/index.html` URL, so dropping that
  special case from the rerender sites is inert.
- 0 URLs with a query string, `..`, `//`, whitespace, or a multi-dot final segment.

## Verification plan

1. Unit tests over the helper, table-driven, including every live shape and the four
   rejection cases.
2. `go test ./platform/...`.
3. Pod-grep a discriminating marker after the roll.
4. The acceptance run the bug asks for is `bugs_open/087`'s re-test on a
   non-`rebuild_policy=owned` page — the handoff says 087 and 125 should ship
   together. Assert the **path**, not just success.

## Decisions log

- **2026-07-31** — helper returns `(path, ok)` rather than `(path, error)`. A page
  whose canonical URL is a fragment is not an error condition; it is a page with no
  file, and the caller has a legitimate fallback.
- **2026-07-31** — `ensureHTMLExtension` (actions) is deliberately NOT reused for the
  URL branch. It *replaces* an existing extension (`foo.php` → `foo.html`), which is
  right for a name and wrong for a URL, where the extension is authoritative.
- **2026-07-31** — no refusal-on-disagreement (the bug's candidate 2). Preferring
  `url` makes the correct path the default; a hard refusal would block 316 pages on a
  disagreement that is *expected* until they are next deployed. The disagreement is
  logged instead, which also makes the fix observable in the pod.
