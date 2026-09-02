# HANDOFF — ai-agent-orchestration.com. ⚠ SUPERSEDED. Written 2026-08-26.

> ## ⛔ SUPERSEDED by `HANDOFF_2026-09-02_continue_here.md` — READ THAT FIRST.
> Its §2 (my `636` patched an orphan) still stands. Its counts do not: the site grew to **45**
> pages and now measures **4** firm failures, two of them on pages that did not exist on 08-26.

**Supersedes `HANDOFF_2026-08-25c_continue_here.md`.** Its numbers still hold; this file adds the
diagnosis of the last tool pages and **corrects a no-op in my own migration `636`**.

> ## Site-wide: **5 firm failures** (of 42 pages, 2 unmeasurable). 1 fix propagating, 4 blocked on a decision.
>
> | page | n | state |
> |---|---|---|
> | `/tools/automation-savings-estimator/index.html` | 3 | ⚠ **diagnosed but NOT fixed — my `636` patched an ORPHAN.** Cross-site component; see §2 |
> | `/tools/build-vs-buy-analyzer/index.html` | 1 | ✅ fixed in `636`; **rerender queued**, verify it landed |
> | `/contact.html` | 1 | fix known & computed; **20-site component**, needs consumers told first |
>
> Plus **one defect the audit cannot see** (§1) and **two pages it cannot read at all**.
> All measured 2026-08-26 — re-run, do not quote.

---

## 1. The diagnosis — and the root is 456's root wearing a different face

Done by asking which declaration **wins** (`getComputedStyle` + a CSSOM walk over
`document.styleSheets`), not by grepping CSS. **None of these is the 456 defect**, which is why
they survived 456/469/625 and why their stored HTML already carried the ink token.

**On this site `--color-primary` and `--color-surface` are THE SAME VALUE (`#0D1117`).** Any rule
pairing them collapses to 1.00:1 — and it reads as *perfectly sensible CSS*, because on every other
site a label in `--color-surface` on a `--color-primary` fill is exactly right. **That is why no
one filed it.**

| declaration | measured | repoint |
|---|---|---|
| `.method-details summary` `var(--color-primary)` | 1.00 | ink 5.66 |
| `#…-calculate-savings-button` `var(--color-surface)` | 1.00 | primary-text 18.92 |
| `.cta-link` `var(--color-secondary)` | 1.09 | accent-ink 9.09 |
| `.bvb-btn-primary` `var(--color-surface)` | 1.00 | primary-text 18.92 |

⚠ **`.cta-link` is an OVERRIDE, not an omission** — the template already has
`a { color: var(--color-accent-ink, …) }` at 9.09:1 and `.cta-link` outranks it.

⚠ **`.result-value` is a FIFTH defect `render_audit.py` CANNOT SEE.** It sits in a result panel
hidden until the calculator runs. Same 1.00:1 collapse. **On a page with conditional UI the audit's
count is a LOWER BOUND** — census the template for the collapsing pair, don't fix only what the
instrument lists.

## 2. ⚠ Half of `636` is a NO-OP — read this before trusting the migration's own header

- `tool-build-vs-buy-analyzer-ai-agent-orchestration-com` — **1 placement**, fix correct. ✅
- `tool-automation-savings-estimator-ai-agent-orchestration-com` — **ZERO placements.** The page
  renders `tool-automation-savings-estimator-**fundamentallyai-com**`, placed on **2 sites**.

**I selected components by CSS CONTENT and never asked which component the PAGE renders.** The
census answered *"where does this rule live?"*; I read it as *"which component does this page
use?"*. The `-ai-agent-orchestration-com` suffix made the wrong one look obviously right.
**What caught it: the propagation INSERT returned `1` when I expected `2`.** The count was the tell.

**So before any component fix here:**
```sql
SELECT p.name, pc.component_id, cc.name FROM pages p
JOIN page_components pc ON pc.page_id=p.id JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.url='<the url>';
```

⚠ **And the remaining diagnosis is INCOMPLETE.** The fundamentallyai fork matches only the
`.result-value` defect on a regex census; the source of the other three live rules is **not
located** — different whitespace in that fork, or the site stylesheet. **Do not assume the orphan's
rules are the fork's rules** — that is the same mistake one layer down.

## 3. What the two blocked items need

- **savings-estimator (3)** — the component is shared with **fundamentallyai.com**. Locate the real
  rules first (§2), then CONTRIB that lane before repointing (owner ruling 2026-07-29 §3).
- **`/contact.html` (1)** — `.form-submit`, white on amber, **2.08**. Fix known:
  `--color-accent-text` = `#294155` here → **5.09:1**, two-level fallback, exactly `457`'s shape.
  ⚠ `contact-form` is on **20 SITES**. Fleet change; tell the consumers first.

## 4. Verified today, do not re-derive

- Roll complete: **one** chassis stamp (`2fb40a960`, ~66,912 pods), was two yesterday.
- `625` intact: all three owned tool pages still `rebuild_policy='owned'`, scripts present, 0 failures.
- Carousels live (opt-in; other 2 sites OFF), 10/10 images 200, `NNN+` incident closed.
- Migrations in force: `469`, `557`, `559`, `560`, `611`, `613`, `625`, `636`.
- ⚠ `writer_block_managed`: **do NOT flip by hand** — another session's `617` (HOLD, council
  APPROVED) applies deliberately.

## 5. Next actions

1. **Verify the build-vs-buy rerender landed** (`created_by='aiao-636-propagate'`), then re-audit
   that page — expect 1 → 0.
2. **Locate the savings-estimator rules in the fundamentallyai fork**, then CONTRIB that lane.
3. **`contact-form` button** — CONTRIB the 20 consumers, then repoint.
4. **The 2 unmeasurable pages** (`ai-readiness-quiz`, `tool-ai-agent-roi-estimator`) — reproducible
   *"probe produced no result"*, both HTTP 200. Likely a bug against `render_audit.py`. **Until this
   is understood no "the whole site is clean" claim is complete.**
5. **9 parked `contrast_failure` items** — 8 should retract once the audit runs, 1
   (`ai-readiness-quiz`) is undecidable. Report to `bugs_open/296`.

## 6. Practice notes this lane has now paid for

- **Ask which component the PAGE renders, not which component contains the rule.** (§2, today)
- **Question the denominator** — "contrast is at zero" was 4 pages of 42 for a week.
- **A guard must test AUTHORSHIP, not presence** — made this mistake twice (`469`, `625`).
- **`render_audit.py` can sample before the stylesheet applies**, and reports a *plausible
  alternative colour* rather than an error. `getComputedStyle` is the tiebreak.
- **A colour can live in an inline `style=`**, invisible to every CSS-rule query.
- **Prose written into a prompt is INPUT, not documentation** (`557` shipped `NNN+` publicly).
- **Ask the CAPABILITY, not the commit** — `service_binary_capabilities` + `merge-base`.
- **Re-check the highest migration number immediately before writing** — `625` and `626` both
  collided with other lanes mid-session.
