# 084 — A published page can lose its JavaScript, and nothing anywhere says so

**Filed** 2026-07-26, from the webdesign.co.uk port.
**Class** silent-loss / missing sensor. Sibling of `041` (chrome `js_content`
publishes but never loads) and `024` (tool fixes never reach the page).
**Status** OPEN. One instance fixed in the tool that caused it; the general gap
is unaddressed.

## Symptom

A page publishes successfully, renders correctly, reports `complete`, and is
completely dead: every interactive control does nothing, because the JavaScript
that drives it is not on the page.

Nothing in the platform notices. Not the work item (it succeeds), not the
assembler (it has nothing to compare against), not the deploy (the file is
valid HTML), not the immune sweep (it has no check for this).

## The instance that produced this file

`cmd/webdesignport` ports ~97 hand-built pages onto webdesign.co.uk. Both source
sites put their `<script>` tags after `</main>`, at body level. The transform
chose its content root (`<main>`) **before** extracting scripts, and therefore
harvested only the scripts inside it — of which there were none.

Result: **60 of 63 tools would have shipped as static markup.** The run reported
`transformed 97 pages, 0 warnings`. Every count was correct. Nothing was wrong
with the output except that it did not work.

It was found by chance — grepping one fragment for `<script` while chasing an
unrelated question about sibling assets.

## Why this is a platform concern and not just one tool's bug

The tool has been fixed and now carries a mandatory parity gate (below). But the
same silence protects every other path that can drop JavaScript:

- **`041`** — chrome `js_content` publishes to `/tools/assets/<function>.js` and
  the page loads it only if the template itself carries the `<script src>` tag.
  Forget the tag and the asset is a 200 nobody requests. Fixed for chrome; the
  same shape is available to any component author.
- **`js_snippets`** — bundled into `assets/js/snippets.js` by
  `render_js_snippets_for_site`, selected by `applies_to` overlapping the site's
  component functions. An `applies_to` typo silently selects nothing.
- **Owned pages** — content lives in `page_components.rendered_html`, written by
  whatever tool authored it. Nothing validates that what went in still has its
  scripts.
- **Section re-render** — `save_page_sections` replaces `rendered_html`
  wholesale. A component whose template loses its script tag loses it everywhere
  at once.

In every case the failure is *invisible from the database*. `build_status` is
`deployed`, the artefact exists, the HTML is well-formed.

## Root cause, stated generally

**There is no point in the pipeline where "this page's JavaScript works" is
asserted.** Every check we have is a check of *presence* (a row exists, a file
exists, a status is `complete`) and none is a check of *integrity* — that what
was published still contains what it needs to function. JavaScript is uniquely
exposed to this because, unlike prose or CSS, its absence changes nothing
visible until a human clicks something.

## Fix candidates

**1. A live script-integrity sweep (recommended).** For each deployed page:
extract every `<script src>`, request it, and flag any that is not 200; and flag
any page whose stored `rendered_html` contains an event-handler attribute
(`onclick=`, `oninput=`, …) or an element id referenced by no loaded script.
This is a discovery check in the existing sweep, not new machinery. It catches
`041`, `024` and this one with a single sensor.

**2. Parity at the write site.** Any writer that transforms existing HTML should
compare script counts before and after and refuse on a shortfall. This is what
was added to `cmd/webdesignport` — see `checkScriptParity` in
`cmd/webdesignport/transform.go`. It is cheap and it is the only check that can
catch loss *at the moment it happens* rather than after publication.

**3. Make the sensor red before trusting it.** Whatever ships, prove it by
inducing the fault. The parity gate above was verified by re-introducing the
original defect in a scratch build: it produced **60 failures**, one per tool
that would have shipped dead. A gate that has only ever been seen passing has
not been tested — it has been observed not complaining.

## How to verify a fix

```bash
# every script a live page loads must actually exist
curl -s https://<domain>/<page> \
  | grep -oP '(?<=<script src=")[^"]+' \
  | while read -r s; do
      printf '%s %s\n' "$s" "$(curl -s -o /dev/null -w '%{http_code}' "https://<domain>${s}")"
    done
```

For webdesign.co.uk on 2026-07-26 this returns 200 for
`/tools/assets/webdesign-couk-header.js` and for each tool's sibling engine
(`/tools/bayesian-rank/bayes.js` and friends).

## Evidence

- `cmd/webdesignport/transform.go` — the fix (scripts harvested from the whole
  body before the content root is chosen) and `checkScriptParity`, the gate.
- `docs/agent_docs/docs024_key_docs_latest/webdesign_couk/NOTES_webdesign_couk.md`
  — the misstep log, including how it was found.
- Live proof of the working end state: `https://webdesign.co.uk/tools/bayesian-rank/index.html`
  carries `bayes.js`; the header carries `webdesign-couk-header.js`; both 200.
