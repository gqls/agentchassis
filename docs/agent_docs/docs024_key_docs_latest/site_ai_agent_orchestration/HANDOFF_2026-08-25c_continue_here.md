# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-08-25 ~20:30Z.

**Supersedes `HANDOFF_2026-08-25b_continue_here.md`** (same evening). That file's correction —
that the "zero" was four pages of forty-two — stands and is the reason this lane now audits the
whole site. Its remaining-work table is superseded: most of it is fixed.

> ## Site-wide, all 42 active pages: **17 firm failures → 5**. Two pages remain unmeasurable.
>
> | | |
> |---|---|
> | `/tools/agent-complexity-estimator.html` | 6 → **0** |
> | `/tools/password-entropy.html` | 2 → **0** |
> | `/tools/tool-llm-cost-calculator.html` | 1 → **0** |
> | `/contact.html` | 4 → **1** |
> | `/tools/automation-savings-estimator/index.html` | 3 → 3 |
> | `/tools/build-vs-buy-analyzer/index.html` | 1 → 1 |
>
> **Carousels, images and the four original pages are unchanged and still clean.**
> All measured 2026-08-25 evening; re-run rather than quote.

---

## 1. What shipped tonight, and the one manoeuvre worth reading before you touch a tool page

**`625`** cleared 9 failures on the three `owned` tool pages. Both obvious routes were wrong:

- ⚠ **Flipping `rebuild_policy` to `generic`** so an ordinary rerender goes through is the
  documented **tool-clobber**: the composition loop commits freshly-written HTML to the deploying
  repo **one step BEFORE** `save_page_sections` refuses, so the calculator ships as prose.
  Calculators have already been destroyed this way (`367` → `377` re-lock).
- ⚠ **`refresh_owned_page_chrome.sh`** is safe but was **inert** for this: assemble mode
  re-assembles the **stored** HTML, which is exactly what carried the stale CSS.

**The working sequence, and it is a two-step:**

1. **Patch the stored `rendered_html` surgically** (migration `625`; precedent `393`). Safe only
   because `bugs_closed/229` (live v1.0.1276) gave those writes a comparison and an archive.
2. **Then** `refresh_owned_page_chrome.sh <site_id> <domain> <marker> <page…>` — because once the
   stored HTML is right, assemble mode is precisely what ships it.

⚠ **A DB patch alone does NOT reach the live site.** After step 1 the pages still served the old CSS.
⚠ **Verify `rebuild_policy='owned'` afterwards.** The script restores it under an EXIT trap and did
so on all three, but assert it — a page left `generic` is exposed to any session's wide rebuild.

## 2. ⚠ Instrument caveats — both cost me time tonight

- **`render_audit.py` can sample BEFORE the stylesheet applies.** `/contact.html`'s button measured
  `2.08 white on #F0A500` in one run and `1.15 white on rgb(239,239,239)` in the next —
  `rgb(239,239,239)` is the UA default `buttonface`. The stored CSS was byte-identical either side.
  **`getComputedStyle` on the live page is the tiebreak**: `bg rgb(240,165,0)`, true ratio **2.08**.
  A differing measurement of an unchanged thing is an instrument reading, not a finding.
- **A colour can be applied by an inline `style=` attribute, which no CSS-rule query can see.**
  `password-entropy`'s two failures were `style="… color: #666 …"`. I searched `{...}` blocks inside
  `<style>` twice and the page came back clean both times.
- **2 pages CANNOT be measured at all** — `ai-readiness-quiz.html`,
  `tool-ai-agent-roi-estimator.html`: *"probe produced no result"*, reproducible, both HTTP 200.
  **Until that is understood no "the whole site is clean" claim is complete**, and one parked
  `contrast_failure` item sits on `ai-readiness-quiz` and can be neither confirmed nor retracted.

## 3. The 5 remaining, with what is known about each

| page | n | state |
|---|---|---|
| `/tools/automation-savings-estimator/index.html` | 3 | `SUMMARY` + `BUTTON` at 1.00, `A.cta-link` at 1.09. **Stored HTML already carries the ink token**, so this is NOT the 456 defect — cause unknown, needs diagnosis, do not guess |
| `/tools/build-vs-buy-analyzer/index.html` | 1 | `BUTTON` at 1.00, same situation |
| `/contact.html` | 1 | `.form-submit`, white on amber, **2.08**. Fix is known: `--color-accent-text` = `#294155` here → **5.09:1**, two-level fallback, exactly `457`'s shape. ⚠ **`contact-form` is on 20 SITES** — a fleet change; tell the consumers first (owner ruling 2026-07-29 §3) |

## 4. Everything else, so it is not re-derived

- Carousels live on `index` + `enterprise-reference-deployment`; opt-in, other 2 sites verified OFF.
- 10/10 case-study images serving 200.
- `NNN+` incident closed; caused by my `557`, fixed by `611`+`613`, WRONG_CALLS logged.
- `bugs_open/364` (clock times) LIVE. Chassis was mid-roll: **two** stamps live
  (`a7459a44b`, `4c996e1b5`), **both contain it** — checked with `merge-base`, not a binary grep.
- Migrations in force: `469`, `557`, `559`, `560`, `611`, `613`, `625`.
- **9 parked `contrast_failure` items** (not 17 — 8 were cancelled 08-24). 6 sit on pages now
  measuring 0; `news`/`tools` measure 0 firm; `ai-readiness-quiz` is unmeasurable. So
  `bugs_open/296`'s test here is **8 of 9 should retract, 1 undecidable**.
- ⚠ **`writer_block_managed`: do NOT flip by hand.** Another session's `617` (HOLD, council
  APPROVED r1) carries `611`'s prohibitions into `writer_block_guidance` and adds a chassis guard.
  It applies deliberately. Not this lane's to trigger.

## 5. Next actions, cheapest first

1. **Diagnose the two `html_ink`-TRUE tool pages** (4 failures). They already carry the token, so
   the cause is elsewhere — read the actual winning declaration with `getComputedStyle`, do not
   grep the stylesheet (this site's served stylesheet has lied about headings before).
2. **The `contact-form` button** — fleet change; CONTRIB the 20 consumers, then repoint.
3. **The 2 unmeasurable pages** — instrument problem; likely worth a bug against `render_audit.py`.
4. Let the render audit retract the 9 parked items; report the outcome to `bugs_open/296`.

## 6. Practice notes earned this week

- **Question the denominator.** "Contrast is at zero" was four pages of forty-two for a week.
- **A guard must test AUTHORSHIP, not presence.** `625`'s first scope guard asserted no other
  component carried the new literal; it fired on 8 that already did, because the literal is an
  existing fleet idiom. Same mistake as `469`'s first bare-var guard — twice now.
- **Prose written into a prompt is INPUT, not documentation** (`557` shipped `NNN+` publicly).
- **Ask the CAPABILITY, not the commit** — `service_binary_capabilities` + `merge-base`; never grep
  a binary for your own sha.
- **Before concluding nothing is happening, ask what is `claimed`.**
- **A peer lane's CONTRIB is another doc** — twice the reported defect was narrower than the class.
