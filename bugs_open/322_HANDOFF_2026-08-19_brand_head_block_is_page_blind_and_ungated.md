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

## ITEM 4 — LIVE AND PROVEN 2026-08-21, plus one residual this file now owns

**Live on chassis `v1.0.1322`** — `declaresHeadTag` probed PRESENT on both replicas, positive control
`injectCanonicalLink` present, fabricated control absent. Council **APPROVED round 2**
(`Council-Reviewed: 54c660f8-1e05-4b88-9910-0d1427b1d805`, 3 advisories, none high).

**Proven at the artefact on the motivating site.** `webdesign.co.uk/guides/tool-aria-builder-guide.html`
now serves its **own** hand-authored `<link rel="icon" href="/favicon.ico">` — preserved exactly, and
still only ONE — **plus** the `apple-touch-icon`, `og:type`, `og:site_name`, `og:image` and
`twitter:card` the wholesale guard had denied it on 117 pages. That is per-tag idempotence doing
exactly what it says: authored tags untouched, missing ones added.

**Round 2's durability objection is also fixed.** A decline (head with no `</head>`) now returns a
reason and the caller writes an `agent_error_log` row (`BRAND_HEAD_TAGS_DECLINED`) with domain,
consequence and remedy — because chassis pod logs retain on the order of minutes, so a `zap.Warn`
alone is write-only. Mutation-proven in both directions.

### ⚠ CORRECTION to the repair count, found while doing the repair

I reported "**two** stored heads short of brand tags". **It was one.** The second,
`loanandmortgagecalculator.co.uk`, has `lock_type='permanent'`
(`locked_by='loanandmortgage_authored_chrome_20260805'`, `component_id IS NULL`) — deliberately
hand-authored chrome the platform must not touch. Its missing `og:image` is **a decision, not a
defect**, and the lock guard correctly refused the re-render I dispatched at it. **Do not "fix" that
site.** The count I gave the council was right in arithmetic and wrong in meaning.

### RESIDUAL this file now owns — the other producer's pages get NO brand tags at all

Raised by the council's `debug_historian` seat and correct: my "N stored heads short" count is scoped
to the **`render_site_components`-driven** population. Pages built through `AssemblePageAction` →
`InjectHead` → `RenderHead` never read `site_components.rendered_html` at all — `RenderHead`
(`component_library.go:2017`) resolves via `ResolveChromeComponent` and falls back to
`RenderFallbackHead`, which emits **no brand tags whatsoever**. So those pages get no `og:image`, no
twitter card and no favicon links, and **nothing in item 4's fix can reach them**.

**I could not size that population honestly.** All three agent types carrying `assemble_page`
(`pageflow-builder`, `page-rebuild`, `site-work-orchestrator`) are `is_active`, and
`orchestration_states` shows **zero runs for any of them** — but that table retains ~24 hours, so that
is a **weak negative, not proof the path is dead** `[UNMEASURED beyond 24h]`. Sizing it needs either a
longer-lived signal or a per-page producer attribution neither of which exists today.

Recorded here rather than spent on a new bug number, because it is the same seam as this file's other
items and because `docs026_concept_register/register/seo.md` **SEO-003** already holds the
two-head-producer convergence open as an architecture-scope question — with a threshold set by the
council's architecture seat: **a fifth one-producer head fix raises an RFC rather than taking a fifth
patch.** This is the fourth.

## ITEM 4 FIXED 2026-08-21 (owner-directed) — the wholesale guard is now PER TAG

Done by the `bugfix_252_og_lang_assembly` lane on the owner's instruction, after the council's
`bug_historian` seat objected on `bugs_closed/252`'s round that removing `og:url` fixed a **symptom**
while this guard — the generic mechanism — stayed exploitable. That framing was better than this
file's and it is why item 4 was promoted from tidy-up to the priority here.

**What changed.** `injectBrandHeadTags` no longer returns the head untouched when it finds
`rel="icon"` OR `og:image`. Each tag is now checked on its own: a tag the head **already declares is
left exactly as authored** (both quote styles), only the missing ones are added, and a head that
already declares everything comes back **byte-identical** rather than splicing an empty string — the
steady state must be free and must not churn `site_components.rendered_html`, whose archive trigger
fires on a real change.

Commit `c2f050036`; council `Council-Submitted: 54c660f8-1e05-4b88-9910-0d1427b1d805`. Mutation-proven
both ways: restoring the wholesale guard fails `TestInjectBrandHeadTags`, and dropping the per-tag
`og:image` check fails it. The test's old assertion (`expected no-op on head with existing favicon`)
**pinned the defect, not a contract** — replaced deliberately, with the reason written into the test.

**What this unblocks, and it is the reason item 4 mattered:** webdesign.co.uk's hand-authored
`rel="icon"` had been costing it **every** og and twitter tag — on **117 assembled pages, the most of
any site in the fleet** — while every caller reported success. It now receives the block. **Checked
before shipping rather than after:** unblocking it means it starts emitting `og:image`, so its assets
were probed first — `/assets/images/og-card.png` and `/assets/images/favicon.png` both return **200**,
so no broken tag is introduced on the motivating case.

**Item 3 is deliberately UNTOUCHED and its landmine is intact.** Nothing in this change consults the
`assets` table, and no tag is gated on an `og_card` row existing. Item 3 (og:image emitted whether or
not the file exists — the five loan-family 404s) remains open and remains yours.

**Item 2 shrinks but does not close.** `og:title`/`og:description` are still site-level fallbacks
here; `bugs_closed/252`'s `spliceOpenGraph` strips and restates both per page at assembly, so what
this function writes is now only what a consumer of the **stored head alone** would see. Their
fallback quality is still item 2's question.

**Item 5** (wide logos making illegible favicons) untouched.

⚠ **Nothing served changes until each site's chrome re-renders, and then only as its pages
re-assemble.** The owner ruled 2026-08-21 that page rebuilds are NOT to be forced; the residual is
tracked in `bugs_open/346`.

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
