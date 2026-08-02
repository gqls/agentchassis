# 180 — the shared link repair rewrites JAVASCRIPT that builds anchors, deleting a working runtime link, and the result still parses

**Filed:** 2026-08-02 by the `bugfix_136_sibling_link_repair` lane, found while re-running
136's census after the v1.0.1229 roll — a "new" census row turned out not to be an anchor
at all.
**Severity:** Medium. Silent content destruction on a live tool, with `success:true`
everywhere. Exposure today is **1 component on 1 site** (measured below), but the mechanism
is fleet-wide and latent in a shared function with three live callers.
**Status:** **CLOSED 2026-08-03 — fixed (`07576d4e1`, LNK-029), LIVE on chassis `v1.0.1233`,
pod-verified on BOTH replicas, and INDUCED on the damaged page.** Fixed by the
`bugfix_136_sibling_link_repair` lane, which filed it. **Not 136** — 136 was about which
writers CALL the repair; this is about what the repair DOES when it is called.

### The close, with its controls

**Live.** Pod-grep, both replicas, one exec each: `NonMarkupSpans` **1**, `ReplaceAllInMarkup`
**2**, `maskNonMarkup` **2**, `dropMatchesInSpans` **1**, `scanSpans` **2**; positive control
`RepairPageLinks` **3**; invented negative control **0**. ⚠ **The load-bearing control is a
real one, not the invented one: `RuntimeFillSpans` went 2 → 1** between `v1.0.1231` and
`v1.0.1233`, because this change turned it from a function body into a one-line wrapper over
the shared walk. A count that MOVED for a reason this change caused discriminates in a way a
newly-added string cannot (`bugs_open/153`).

**It missed one build first, and that was caught rather than assumed.** On `v1.0.1231` the
same grep returned `NonMarkupSpans` **0** with the controls at 3 and 2 — a real negative, not
a mis-cased pattern. Commit 22:37:50 BST, those pods started 22:39:20 BST.

**Induced on the damaged page.** A reason-LESS `page_rerender` on
`tool-cma-obligation-checker` (item_key `page_rerender_..._180_induction`, `complete`) — chosen
deliberately: the *stored* `rendered_html` still HELD the correct anchor, so re-stapling stored
HTML and running the FIXED outbound repair over it exercises exactly the changed path
(`rerender_single_page_action.go:223`), where a `section_data_resolved` rerender would instead
regenerate the very bytes whose survival is the test. Prediction written before the run;
result, against the wire:

| | before | after |
|---|---|---|
| `curl … \| grep -c 'q\.link'` | **0** | **1** |
| page bytes | 27,247 | 27,307 |
| served JS | `' <p>' + statusText + ' See guide section.</p>'` | `'<p>' + statusText + ' <a href="' + q.link + '" target="_blank" rel="noopener">See guide section</a>.</p>'` |

⚠ **The first post-rerender fetch returned 0 and the work item said `complete`.** Neither was
the truth: Cloudflare serves this page `cache-control: max-age=3600`, so the fetch was a cached
copy. `last-modified` on the fresh fetch reads 22:16:19 GMT, seconds after the rerender. **A
`complete` status and an unchanged page together look exactly like "the fix did not work" —
check `last-modified`/`age` before believing either.**

**Council:** `Council-Submitted: ba199c35-516f-44be-a210-9fd982425eb7`. ⚠ The first run
**stalled at `review_constitution`, `updated_at` frozen 21:38:10, one minute before the
`v1.0.1231` roll replaced the pods** — the documented behaviour that a roll kills an in-flight
council. Resubmitted on the same trail (run `612bc0f9-de7e-4627-a712-a3b226694677`); the
verdict is **owed and unread**, which is why no `Council-Reviewed:` trailer exists on the fix
commit and none should be written until someone reads it.

## THE DAMAGE IS CONFIRMED ON THE WIRE, not only in a probe (added 2026-08-03)

The filing measured `page_components.rendered_html` and reproduced the corruption in a probe.
Neither shows the visitor's page. Fetched today:

```
$ curl -s https://vetcomparison.uk/tools/cma-obligation-checker/index.html | grep -c 'q\.link'
0
```

and the served bytes at the site of the damage read:

```js
'<p>' + statusText + ' See guide section.</p>' +
```

The anchor is **gone from the deployed program** while the STORED `rendered_html` still
contains it — which is the signature of `repairOutboundPageLinks` (LNK-023), the caller that
repairs the assembled page on the way out and leaves the DB copy alone. So a DB-only census
would report this page as healthy for ever. ⚠ **Note the trailing space in `' See guide'` —
that is the byte the deleted `<a …>` used to sit against**, and it is the only residue.

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

## THE FIX AS BUILT (2026-08-03, commit `07576d4e1`, register **LNK-029**)

**Candidate 1, widened — and the widening is the finding.** Before writing anything I ran the
SHIPPING function over six inputs. **Five were corruptions and only one was the spelling this
file was filed for**, and they do not share an arm:

| input | shipped output | arm |
|---|---|---|
| `<script>… '<a href="' + q.link + '">See guide</a>' …</script>` | anchor deleted | `LinkScopeEmpty` → unlink |
| `` <script>`<a href="${q.link}">See guide</a>`</script> `` | anchor deleted | `LinkScopePage` → **phantom** |
| `<style>/* <a href="/nope">x</a> */</style>` | CSS rewritten | phantom |
| `<textarea><a href="/nope">x</a></textarea>` | visible text edited | phantom |
| `<!-- <a href="/nope">x</a> -->` | comment rewritten | phantom |
| `<p><a href="/nope">real phantom</a></p>` | **correctly unlinked** | phantom |

That table is why candidate 2 (a denylist of `' +` / `${`) was rejected: it addresses one of
five. Candidate 3 was rejected on the measurement — it would abandon repair on 161 of 1,186
components to protect 1.

**What shipped.** `datahelpers/markup_spans.go`: `NonMarkupSpans` (raw-text elements **and**
comments, whole), `MarkupMatches`, `ReplaceAllInMarkup`. The span set comes off the tag walk
`runtime_fill.go` already performs — `RuntimeFillSpans` keeps its signature and becomes a
caller of a shared `scanSpans`, because two walks over one grammar is `bugs_open/137`'s own
defect. Wired into `RepairPageLinks` (one line, fixing all three live callers at once) and
into `DropDeadURLControls`.

**Two things the fix taught that the diagnosis did not**, both pinned by tests:

1. **Mask, do not filter.** `<script>var t='<a href="/gone">';</script><a href="/gone">Pricing</a>`
   — the non-greedy `</a>` closes at the REAL anchor, so ONE match spans both. Filtering
   matches that start in a span drops the genuine phantom with the decoy, and `FindAll` never
   revisits those bytes.
2. **The filler byte is load-bearing; whitespace is wrong.** `\ssrc\s*=\s*""` cannot cross
   `<style>…</style>` but crosses the spaces that replaced it, and that match begins
   **outside** the span where an offset filter has no view. NUL.

**Measured before submitting, at the altitude of the live caller.** Both matchers over all
**509 assembled pages** / 19 sites: **11** matches inside non-markup, **1** page mutated
destructively today, **0 legitimate repairs lost**. Over the 13 components fleet-wide with
unbalanced raw-text tags (where the degrade-wide arm could cost coverage): **0 and 0**.
`DropDeadURLControls`: 13 + 37 matches fleet-wide, **0** inside a span — byte-identical today.

**Council:** `Council-Submitted: ba199c35-516f-44be-a210-9fd982425eb7` (verdict pending at
time of writing; read it and act on a REVISE — the code is already on the shared branch).

**STILL OPEN AFTER THIS FIX, deliberately: the DETECTORS.** `links.go`'s phantom scan,
`check_dead_controls` and `check_phantom_internal_links` still read a JS-built anchor as a
finding. A false finding costs a human's attention; a false repair costs content, so they are
different decisions with different fail-safe directions, and narrowing a JUDGE is
`bugs_open/137`'s separate ruling. It belongs to the detection lane (`bugs_open/097`,
`bugs_open/116`) and is recorded here rather than closed by assumption. The helper they would
use already exists (`MarkupMatches`), so it is a one-line adoption when that lane wants it.

## Fix candidates (unranked — kept as filed; candidate 1 was taken, see above)

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
