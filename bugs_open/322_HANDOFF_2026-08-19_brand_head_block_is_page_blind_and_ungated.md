# 322 — the brand head block is page-blind and ungated, and a wide logo cannot make a legible favicon

**Spun out of `bugs_open/131` (og-card slug) on its closure, 2026-08-19.** That file's
headline defect — cards that were never generated — is fixed and live (mis-route closed at
producer, router and completion gate; census 2026-08-19: 18 public sites serve both
artefacts 200; the 5 loan-family 404s are logo-absence → `bugs_open/210`'s territory, or
no-discovery-driver → `bugs_open/230`, both routed there). What remains are the defects in
the EMITTER and the DERIVER'S SOURCE, which were documented inside 131 and deserve their
own case rather than riding a closed file. Full history: 131's file (both status sections)
and `docs024_key_docs_latest/bugfix_131_og_card/`.

**Severity:** medium, outward-facing (what every share and every browser tab shows).
**Verified still present in current code 2026-08-19** — line numbers below are from that
read (`render_site_components_action.go`, `git log` head `229e14e74`).

## The emitter cluster — one function, `injectBrandHeadTags`

1. **No page context at all.** The block is built from site fields only and injected into
   every page's head. `og:url` is hardcoded to the site root (`:495`), so every inner page
   advertises the homepage's identity; `rel="canonical"` is never written. Measured
   2026-07-29 (vonc6 contribution in 131's file): 7 pages, 4 sites, no exceptions. The fix
   is structural — give the injector the page it renders (url, title/h1, meta description)
   and emit per-page `og:url`/`og:title`/`og:description` + `canonical`, site values as
   fallback. A signature change on a shared renderer: architecture-shaped, one deliberate
   change, not three patches.
2. **`og:title` falls back to the bare domain** (`:490`, ~8 sites measured 2026-07-28 —
   re-measure before acting, the fleet has grown) and **`og:description` is skipped when
   the source field is empty** (`:492`) instead of falling back to anything.
3. **`og:image` is emitted unconditionally** (`:494`) whether or not the card exists —
   survives as the gap it was on new sites (the 5 loan-family 404s today are this tag
   pointing at nothing). ⚠ **Landmine from 131, still binding: do NOT gate the tag on "an
   `og_card` assets row exists"** — leopardessconsulting.co.uk served a working
   hand-committed card with NO row for weeks (it now has backfilled locked rows, but the
   gate must not assume rows). Follow the `sprites.css` precedent in the same function.
4. **A head that already contains `rel="icon"` or `og:image` is skipped wholesale**
   (`:467`) — why webdesign.co.uk emitted no og:image at all in 131's census. A partial
   hand-authored head permanently opts a site out of every tag in the block.

## The deriver-source defect

5. **`derive_brand_head_assets` always reads `asset_key='logo'`**, so a wide wordmark
   becomes the favicon source. Aspect is preserved (131's July fix), but 19px of ink in a
   64px canvas is a grey smudge at true 16px tab size — **measured, not guessed**, on
   relojistas 2026-07-29. Wide-logo sites counted then: relojistas (source since repaired),
   fundamentallyai, oufe, robot-hands, vetcomparison; cookly.uk joined the class 2026-08-18
   (eyeballed). The real fix is a **square favicon source** the deriver can be pointed at
   (a distinct asset_key with fallback to logo), which it cannot currently express.

## CONTRIBUTION 2026-08-20 — item 1's ASSEMBLY end is being fixed by the 252 lane; items 2-5 stay yours

Contributed by the lane working `bugs_open/252` (og/lang slug),
`docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/`. Filed here so the two files
do not diverge — 252 and this file describe **the same tags from opposite ends**, and item 1's own
text ("give the injector the page it renders") is one of two possible places to fix it.

**Why the fix is landing at assembly rather than in the injector.** `injectBrandHeadTags` writes into
a **per-site artefact** (`site_components.rendered_html`, one row per site), which is then reused by
every page `assemblePage` builds. Giving the injector a page cannot work for that artefact: there is
no single page for a site-level head to be about. So per-page identity has to be stated where the
page is known — at assembly — and the site-level artefact must stop asserting it at all.

**What the 252 lane is taking:**
- **Item 1, the `og:url` half, at both ends.** New `spliceOpenGraph` in a new
  `platform/orchestration/actions/head_assembly.go`, called from `assemblePage`: strips
  `og:title`/`og:description`/`og:url` from the stored head and injects one per-page set, `og:url`
  via `preferredPageURL` so it agrees with the canonical and the JSON-LD `@id` by construction. Plus
  the **one-line removal of the `og:url` emission** from `injectBrandHeadTags` (working tree ~`:510`).
  That measured "7 pages, 4 sites, no exceptions" is now 22 of 24 stored heads / 700 assembled pages.
- **Item 1's `rel="canonical"` clause is already done** and not ours to redo: `injectCanonicalLink`
  has been on the assembly path since 2026-08-02 and `bugs_open/251`'s root normalisation
  (`preferredPageURL`) is live and council-approved. Verified at the artefact — an assembled subpage
  serves a correct per-page canonical today, beside the wrong `og:url`. **Item 1's wording implies
  the canonical is missing; on the `page-rerender` path it is not.**

**What stays with this file, untouched by us:**
- **Item 2** — `og:title` falling back to the bare domain, `og:description` skipped when the tagline
  is empty. Site-level fallback *quality*. The 252 change makes these the fallback-of-last-resort
  rather than what pages actually serve, so the item shrinks but does not close.
- **Item 3** — `og:image` emitted whether or not the card exists. **We are deliberately not touching
  any og:image line**, and we have preserved this file's landmine in our own plan: do NOT gate on an
  `og_card` assets row.
- **Item 4** — the wholesale `rel="icon"`/`og:image` idempotency skip. Note for whoever takes it:
  that guard is also why the 4 `head-seo-standard` sites carry **duplicate** og tags — it cannot see
  the template's blank `og:title`/`og:description`, so it appends a filled second set. Our migration
  removes those two blank lines from the template, which removes the duplication; the guard's own
  over-broad behaviour (opting webdesign.co.uk out of everything) is unchanged and still yours.
- **Item 5** — favicon source / square asset key.

**One correction to this file's verification recipe.** "og:url must be the PAGE url" will be true on
the `page-rerender` path once 252 ships, and will remain **false on the `assemble_page` path**
(`multipage_actions.go`, 3 agent types), which builds its head from `RenderFallbackHead` and calls
none of the injectors. So the check must record which path built the page, or a correct fix will read
as broken on a page rebuilt through the other producer. That divergence is the register's open
architecture question (SEO-003); 252 adds `head_assembly.go` as the named seam for it without
deciding it.

## Related, not duplicated here

- Completion integrity for these artefacts is DONE and live (131: producer emits
  `spec.mode`, `VerifyBrandHeadAssetsResolved` gates completion, 467 routing fallback
  witnessed) — do not rebuild any of it.
- asset-deployer's `input_contract` is `{}` (council advisory on 131's round: neither
  `spec.mode` nor `spec.purpose` is declared). Declare both together **after reading how
  `input_contract` is enforced** — a first contract on an agent that never had one may
  activate validation.
- New sites start with both artefacts 404 until something files a brand-head item:
  discovery has no recurring driver (`bugs_open/230`) and site-build does not derive
  brand-head. Whoever fixes 230 or the build pipeline closes that half; the redrive
  one-liner is in `bugfix_131_og_card/RUNBOOK_og_card.md` (precondition: an active logo).

## How to verify any fix here

Never assert tag presence — fetch what the tag names, per page not per site:

```bash
curl -s "https://<site>/<inner-page>" | grep -oE '(og:url|og:title|og:image|canonical)[^>]*'
# og:url must be the PAGE url; og:image's target must return 200; title must not be the bare domain
```
