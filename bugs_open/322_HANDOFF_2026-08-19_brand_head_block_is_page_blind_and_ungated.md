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
