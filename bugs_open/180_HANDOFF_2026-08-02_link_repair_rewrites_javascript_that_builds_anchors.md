# 180 — the shared link repair rewrites JAVASCRIPT that builds anchors, deleting a working runtime link, and the result still parses

**Filed:** 2026-08-02 by the `bugfix_136_sibling_link_repair` lane, found while re-running
136's census after the v1.0.1229 roll — a "new" census row turned out not to be an anchor
at all.
**Severity:** Medium. Silent content destruction on a live tool, with `success:true`
everywhere. Exposure today is **1 component on 1 site** (measured below), but the mechanism
is fleet-wide and latent in a shared function with three live callers.
**Status:** OPEN, unowned. **Not 136** — 136 was about which writers CALL the repair; this
is about what the repair DOES when it is called.

## Verification standing (RFC_005 / owner ruling 2026-07-31)

**The `090` diagnosis loop was NOT run, and here is the substitute, stated plainly rather
than omitted.** This is not an inference about a causal chain — I executed the shipping
function over the exact bytes stored on the live site and observed the output:

```go
// probe, run against platform/orchestration/datahelpers at HEAD, then deleted
ix := NewPageURLIndex([]string{"/index.html", "/tools/cma-obligation-checker.html"})
in := `<script>var h = '<p>' + statusText + ' <a href="' + q.link + '" target="_blank" rel="noopener">See guide section</a>.</p>';</script>`
got, repairs := RepairPageLinks(in, ix)
```
```
repairs=1
  action=unlink href="" new=""
IN : <script>var h = '<p>' + statusText + ' <a href="' + q.link + '" target="_blank" rel="noopener">See guide section</a>.</p>';</script>
OUT: <script>var h = '<p>' + statusText + ' See guide section.</p>';</script>
```

There is no hidden step between cause and effect here: one deterministic function, real
input, corrupt output, reproducible in three lines. Re-run it before believing this file.

## Root cause

`platform/orchestration/datahelpers/link_repair.go`. `repairAnchorRe` is a byte-level regex
over the WHOLE input:

```go
var repairAnchorRe = regexp.MustCompile(`(?is)<a\b[^>]*\shref\s*=\s*["']([^"']*)["'][^>]*>(.*?)</a>`)
```

Given the JS fragment `<a href="' + q.link + '" ...>`, the href capture `[^"']*` cannot
cross the `'` that immediately follows `href="`, so it captures the **empty string** and the
closing `["']` matches that `'`. Empty href → `ClassifyLinkScope` returns `LinkScopeEmpty` →
the `unlink` arm drops the `<a>` and keeps the inner text. The anchor a tool builds at
runtime is deleted from its own source.

**Two properties make it silent.** The output is still valid JavaScript, so nothing throws;
and unlink keeps the text, so the page still reads sensibly — the visitor simply cannot
click. This is the `rerender_link_repair.go` landmine ("a dead internal link is REPAIRED
into orphaned prose") in a form no author anticipated: here the link was never dead.

**The `data-runtime-fill` exemption does not save it.** That exemption is whole-input for
writers (deliberately, LNK-025), so a marked sibling component would exempt the whole
assembled page — but the affected component carries no marker and neither does any
component on its page (measured).

## Exposure, measured 2026-08-02 (and it is a LOWER BOUND)

```sql
SELECT count(*) AS components, count(DISTINCT s.domain) AS sites,
       count(*) FILTER (WHERE pc.rendered_html ILIKE '%data-runtime-fill%') AS self_exempt
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html ~ 'href="''\s*\+|href=''"\s*\+|href="\s*''\s*\+';
-- 1 component | 1 site | 0 self-exempt
```

`vetcomparison.uk` / `tool-cma-obligation-checker`, and **no component on that page carries
a runtime-fill marker**, so the whole-page exemption does not cover it either.

⚠ **That query only catches ONE SPELLING** — `href="' + …`. It does not catch template
literals (`href="${url}"`), `setAttribute('href', …)`, or JSX-ish forms. A template literal
is arguably worse: `${url}` is a non-empty href, so it takes the `LinkScopePage` arm, fails
the index lookup, and is unlinked as a *phantom*. **Do not quote "1 component" as the size
of the class** — it is the size of one spelling of it.

## Which callers can actually reach it

| caller | reaches tool markup? | live? |
|---|---|---|
| `repairOutboundPageLinks` (rerender, LNK-023) | **YES** — it repairs the ASSEMBLED page, which contains every component | **yes: 21 `CONTENT_LINK_REPAIR_DETAIL` rows since the v1.0.1229 roll** |
| `repairSectionsBeforePersist` (save, LNK-024) | only if a tool component goes through a section save | live |
| `repairComponentHTMLBeforePersist` (LNK-027) | not today — the tool writers do not call it, and this bug is a reason to be careful about wiring them (`bugs_open/136`) | live, dormant |
| `ValidatePageContentAction` (build gate) | its repaired `clean_html` is structurally discarded (079) | live, inert |

## Fix candidates (unranked)

1. **Skip `<script>` and `<style>` spans before matching.** `datahelpers/runtime_fill.go`'s
   tag scanner already "jumps `script`/`style` contents" (LNK-025) — the capability exists
   and is tested; this is a matter of exposing a span set and having `RepairPageLinks`
   consult it. Smallest correct fix, and it fixes every spelling at once, including the
   template-literal form.
2. **Refuse an anchor whose href contains a quote-adjacent concatenation marker** (`' +`,
   `${`, `" +`). Narrow, cheap, and treats the symptom — it would leave
   `setAttribute('href', …)` alone (correctly, that is not an anchor tag) but is a
   denylist, and denylists of syntax rot.
3. **Do not repair inside a component whose markup contains `<script`.** Bluntest; costs
   real repair coverage on tool pages, which is where phantom links also live.

**(1) is almost certainly right**, and its cost is that the whole-input runtime-fill
exemption and a new script-span exemption are two different notions of "do not touch this
region" that must not drift — see LNK-025's own RFC threshold note.

## How to verify a fix

Not on a build. Take the probe above, add a template-literal case
(`<a href="${q.link}">x</a>` inside a `<script>`), and require **byte-identical output**.
Then re-run the exposure query, and check the live tool still has its link:
`curl -s <tool page> | grep -c "q.link"` — the JS must survive a rerender of that page.

## Related

- `bugs_open/136` (section-editor slug) — the lane that found this. Its § "still open"
  names the tool-markup writers as the next candidate to receive the repair; **this bug is
  a reason that wiring must not be a copy-paste**, because tool markup is exactly where
  JS-built anchors live.
- LNK-023 / LNK-024 / LNK-027 — the three live callers.
- LNK-025 — owns the tag scanner that fix candidate 1 would reuse, and the whole-input
  vs element-scope boundary this would extend.
- `bugs_open/097`, `bugs_open/116` — the detection side; neither would see this, because
  the repair happens after detection and the result looks like clean prose.
