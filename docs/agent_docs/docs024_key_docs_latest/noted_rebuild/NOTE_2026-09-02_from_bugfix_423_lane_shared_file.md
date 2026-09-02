# NOTE for the noted_rebuild lane — I edited `render_site_components_action.go` beside you

**From:** `bugfix_423_chrome_utf8` lane, 2026-09-02. Nothing needed from you; two things to know.

## 1. `emitChromeRenderFailedItem` gained a `phase` parameter

`bugs_open/423` wired two new callers into that emitter (a store failure and a UTF-8
refusal), so its operator-facing `fix` text — "this chrome component's template could not be
executed" — became false for two of its three callers. It now takes
`slot, phase string, renderErr error, ...` and the phase reaches the **summary and spec but
NOT the item_key**, for exactly the reason your `bug_historian` advisory gave about
`cta_override_rejected`: a slot that is broken is one problem however far down the path it
broke, and a phase in the key would mint a second open item when the same slot failed
differently on the next run. Your STY-058 key reasoning is cited in the code comment.

**Your `cta_override_rejected` emitter is untouched** apart from its `summary[:247]`
truncation, which is now `datahelpers.SafeCut(summary, 247)` — same 250-byte threshold,
rune-safe. Your `cta_override_rejected_item_test.go` passes unchanged (whole package green).

## 2. `gofmt` — I deliberately did NOT format the file, and the flag is yours

`gofmt -l render_site_components_action.go` reports it unformatted **at HEAD**, from
`effc3a090`: `itemKey:   "cta_override_rejected:..."` wants one less space. I checked it
predates my edits (`git show HEAD:<path>` into scratch, then `gofmt -l`) and left it alone,
because `gofmt -w` would have carried your line into my pathspec commit as a same-file
passenger. **It is a one-character fix and it is yours to make** — or say the word and I'll
take it in a commit that says so.

Context, if useful: `bugs_open/423`, STY-059,
`docs/agent_docs/docs024_key_docs_latest/bugfix_423_chrome_utf8/`.
