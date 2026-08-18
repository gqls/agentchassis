# 315 — `pages.deployed_at` is stamped whether or not the object is written, and one page has now been skipped by FOUR completed rerenders

**Filed 2026-08-18** by the `webdesign_tool_rebuilds` lane. **Status: OPEN, UNOWNED.**
Two findings: a **measurement defect** that makes the failure invisible (§2), and a **live instance**
of the failure it hides (§3). The measurement defect is the more important of the two.

## 1. The one-paragraph version

`webdesign.co.uk/tools/seo-injector/index.html` serves a tool that was replaced hours ago. The
database is correct, four `page_rerender` items have completed with no error, and
`pages.deployed_at` has been stamped fresh each time — while the **origin object has not been
rewritten since 14:12:06Z**. Nothing anywhere in the platform records that the publish did not
happen; `deployed_at` says it did.

## 2. THE MEASUREMENT DEFECT — `deployed_at` is not evidence of publication `[MEASURED 2026-08-18 20:46Z]`

Three pages, each stamped `deployed_at` within the last half hour, against the origin's own
`last-modified` (fetched with a cache-buster; `cf-cache-status: DYNAMIC`, so these are origin headers):

| page | `pages.deployed_at` | origin `last-modified` | serving correct content? |
|---|---|---|---|
| `tool-seo-injector` | 20:45:57 | **14:12:06** | **NO** |
| `tool-json-cleaner` | 20:45:06 | 19:08:55 | yes |
| `tool-smooth-shadow` | 20:15:29 | 19:08:54 | yes |

**All three are stale against their own `deployed_at`** — including the two that are serving correctly.
So the column tracks "a rerender ran", not "bytes were written". Anyone using it to answer *did this
page publish?* gets yes for a page that has not been touched in six hours.

Note the second row's evidence value: the two healthy pages share a `last-modified` **to the second**
(19:08:54 / 19:08:55), which says publication happens in **batches**, decoupled from the per-page
rerender that stamps `deployed_at`. That is the seam where a page can be dropped silently.

## 3. The live instance

- Page `3d1fbd02-ae36-436a-a281-539ac285d4aa`, `/tools/seo-injector/index.html`.
- **DB is correct:** ported slot `15b8323c` `build_status='removed'` (18:57:00); native slot
  `2100c25e` `deployed`, and its stored `rendered_html` **contains the new component's marker
  (`scriptOpenTag`) and does NOT contain the old tool's `b-type`**.
- **Four rerenders, all `complete`, all `error IS NULL`**, orchestrations COMPLETED with no
  `__step_error`: 15:18:58, 17:10:29, 20:12:06, and a purpose-built republish at **20:45:59** filed
  with a distinct `item_key` specifically to rule out dedup silently swallowing it.
- **Origin unchanged throughout:** `last-modified: Tue, 18 Aug 2026 14:12:06 GMT`, content still the
  ported tool (`class="ported-page"` 1, `scriptOpenTag` 0).
- **Isolated, not systemic:** four sibling pages rebuilt the same way today are all serving correctly
  (`html-minifier`, `svg-optimizer`, `json-cleaner`, `smooth-shadow`). The publish seam works; this
  page is being skipped by it.

## 4. Why it stayed invisible until now

Every layer below the artefact reports success: the work item is `complete`, the orchestration is
COMPLETED, `deployed_at` is fresh, and the database holds the right HTML. **This is CLAUDE.md's
"trust the rendered artefact, not the status" with all four lower layers green.** It was caught only
because this lane grades at the served bytes with a cache-buster — and note that an hour earlier the
identical *symptom* on a different page WAS just a stale edge cache, so the cheap explanation was
available and wrong here.

## 5. Fix candidates, ordered by what closes the door

1. **Make `deployed_at` mean what it says**: stamp it only after a confirmed object write, and record
   the written hash/etag alongside. `pages.content_hash` exists and is **empty on all three pages
   above** — populating it at publish time would make "is the origin current?" a comparison rather
   than an assumption.
2. **Fail the rerender when the publish writes nothing.** A completed item that produced no object is
   the defect; it should be `failed` with the reason, not `complete`.
3. **Find why this page is skipped by the batch** — the two healthy pages share a to-the-second
   publish time, so there is a batch boundary; this page is falling outside it. Start from the
   publisher's page selection, not from the rerender.
4. Alert on divergence: a periodic sweep comparing `deployed_at` to the origin's `last-modified` for
   deployed pages would have caught this the first time, at 15:18.

## 6. How to verify a fix

`curl -sI "https://<domain>/<url>?x=$RANDOM" | grep -i last-modified` must move forward after a
rerender completes, on the page named above. Negative control: a page nobody rerendered must NOT move.
**Always cache-bust** — `cf-cache-status: DYNAMIC` confirms you are reading the origin.

## 7. Related

- `docs024_key_docs_latest/webdesign_tool_rebuilds/NOTES_…` 2026-08-18 20:12Z and 20:46Z (full evidence)
- CLAUDE.md "Trust the rendered artefact, not the status" · `MEMORY/prove-a-deploy-at-the-artefact-index`
- The publish-seam canary from another lane (commit `a2a9912c2`, "served sha256 == pre-publish origin
  hash, published_hash written only after acceptance") — that canary proves the seam CAN work; this is
  a page it is not reaching, and the acceptance idea in it is fix candidate 1.
