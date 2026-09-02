# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-09-02.

**Supersedes `HANDOFF_2026-08-26_continue_here.md`.**

> ## 4 firm contrast failures across **45** active pages (2 unmeasurable). But the number is not the story.
>
> **The site grew 42 → 45 pages this week and TWO of the three new pages arrived carrying the same
> defect this lane has been clearing since August.** Fixing instances has not stopped new ones.
>
> | page | n | note |
> |---|---|---|
> | `/tools/token-calculator/index.html` | 1 | **NEW page**, `<TD>` 1.14 |
> | `/tools/model-approach-selector/index.html` | 1 | **NEW page**, button 1.04 — the primary-fill shape |
> | `/tools/automation-savings-estimator/index.html` | 1 | button 1.04 — known, blocked on a cross-site component (§3) |
> | `/contact.html` | 1 | white on amber, **true ratio 2.08** (§4) |
>
> All measured 2026-09-02 — **re-enumerate the pages, do not reuse this list.**

---

## 1. What landed since 08-26

- `636` propagated: `/tools/build-vs-buy-analyzer/` → **0**. ✅
- `625` holding: `password-entropy`, `tool-llm-cost-calculator` → **0**. ✅
- `692` + `bugs_open/434`: the complexity estimator was serving the calculator **twice** — repaired,
  now **0**, `owned`, 12 inputs. See §2.

## 2. ⚠ A defect this lane had CLOSED was silently re-opened by something that arrived afterwards

`/tools/agent-complexity-estimator.html` measured 0 on 08-25 and 1 on 09-02. Nothing regressed my
work — a **second placement** appeared in the same slot on **2026-08-26 14:48** and it never
received `625`'s repoint.

| | bytes | fieldsets | legends | **inputs** |
|---|---|---|---|---|
| incumbent (2026-04-09) | 22,732 | 4 | 4 | **12** |
| the new one (2026-08-26) | 19,964 | 1 | 1 | **1** |

⚠ **A size floor would have passed it** (−12%). **The input count is the axis that shows the loss.**
⚠ **The estate's de-dup rule was VACUOUS**: both rows carry `content_data='{}'`, so
`count(DISTINCT md5(content_data))` = 1 — agreement it never established. `content_hash` is empty.
**Discriminate on the rendered artefact's structure.**
⚠ **It is NOT `bugs_open/430`.** 430 is fork-on-deploy dropping `js_content`; the new component has
**`forked_from IS NULL`** — generated, not forked, and copying fewer columns cannot rewrite markup.
Both have `js_content` length 0, so that signature does **not** discriminate. Use `forked_from`.

**Producer unidentified.** Nothing in the repo names the component. **If the duplicate returns, that
is the finding** — reopen `434` against the producer rather than repairing the damage again.

## 3. The recurring shape, and why instance-fixing is not converging

Three of the four current failures are the **same defect**: a label painted with a token that equals
its own fill, because on this site `--color-primary` == `--color-surface` == `#0D1117`. It reads as
correct CSS everywhere else in the fleet, which is why it keeps being written.

⚠ **Two of the three sit on pages that did not exist a week ago.** So the source these tool pages
are generated from still carries the pattern. **The next useful move is upstream — find what
generates tool markup and fix the pattern there** — not another per-page migration. This lane has
now done four of those (`456`, `469`, `625`, `636`) and the class keeps arriving.

`automation-savings-estimator` specifically: ⚠ **my `636` patched an ORPHAN** — the aiao-named
component has **zero placements**; the page renders the `fundamentallyai.com` fork, shared across
**2 sites**. Fixing it needs that lane told first (owner ruling 2026-07-29 §3), and the rules'
actual location in that fork is still **not** established.

## 4. ⚠ `/contact.html` — the reading is unstable, the fix is known

Measured `1.15 white on rgb(239,239,239)` today and `2.08 white on #F0A500` before.
`rgb(239,239,239)` is the UA default `buttonface` — **the audit sampled before the stylesheet
applied**. `getComputedStyle` settles it: the button is amber, **true ratio 2.08**.
Fix: `--color-accent-text` = `#294155` here → **5.09:1**, two-level fallback, `457`'s shape.
⚠ **`contact-form` is on 20 SITES** — fleet change; tell the consumers first.

## 5. ⚠ Two pages still CANNOT be measured

`ai-readiness-quiz.html`, `tool-ai-agent-roi-estimator.html` — *"probe produced no result"*,
reproducible, both HTTP 200. **No "the whole site is clean" claim is complete while these stand**,
and one parked `contrast_failure` item sits on `ai-readiness-quiz`, so it can be neither confirmed
nor retracted. Likely a bug against `render_audit.py`.

## 6. ⚠ Operational trap I hit today — read before touching an owned tool page

`refresh_owned_page_chrome.sh` flips the page to `generic`, arms `trap restore_all EXIT INT TERM`,
publishes, then waits `24 × 5s`. **Run under a 2-minute foreground timeout, the harness kills it and
the trap does NOT fire.** I left `/tools/agent-complexity-estimator.html` on `generic` from 05:56
to ~14:00. Nothing errored; the damage is a **window**, not a visible state — the page was eligible
for the generic composition loop, which commits rewritten HTML to the deploying repo one step
*before* `save_page_sections` refuses. Nothing hit it (verified 12 inputs, unchanged).

**Run that script with `run_in_background: true`, and assert the policy immediately afterwards:**
```sql
SELECT name, rebuild_policy FROM pages WHERE site_id='…' AND name IN (…);  -- must be 'owned'
```
LANDMINE filed.

## 7. Next actions

1. **Go upstream on the token-collision pattern** (§3) — the highest-value move; per-page fixes are
   not converging while new pages keep arriving with it.
2. **`contact-form` button** — CONTRIB the 20 consumers, then repoint.
3. **`automation-savings-estimator`** — locate the rules in the `fundamentallyai` fork, CONTRIB that
   lane, then fix.
4. **The 2 unmeasurable pages** — file against `render_audit.py`.
5. **9 parked `contrast_failure` items** — 8 should retract once the audit runs; 1 undecidable.
   Report to `bugs_open/296`.

## 8. Standing facts (verified today)

- Carousels live on `index` + `enterprise-reference-deployment`, opt-in, other 2 sites OFF.
- 10/10 case-study images 200. `NNN+` incident closed.
- Migrations in force: `469`, `557`, `559`, `560`, `611`, `613`, `625`, `636`, `692`.
- ⚠ `writer_block_managed`: **do NOT flip by hand** — `617` (HOLD, council APPROVED) applies deliberately.
- ⚠ **Migration and bug numbers collide on this tree.** `625` and `626` both collided mid-session;
  re-check the highest number immediately before writing, and resolve by SLUG.

## 9. Practice notes this lane has paid for

- **Re-enumerate the denominator every time** — 42 pages became 45 in a week.
- **Ask which component the PAGE renders**, not which contains the rule (`636` patched an orphan).
- **A guard must test AUTHORSHIP, not presence** (`469`, `625` — twice).
- **`render_audit.py` samples before CSS applies** and reports a plausible alternative colour, not
  an error. `getComputedStyle` is the tiebreak.
- **A colour can live in an inline `style=`**, invisible to every CSS-rule query.
- **Grep `/bugs_open/` BEFORE filing** — I did it after, and it changed the filing (§2).
- **Prose written into a prompt is INPUT, not documentation** (`557` shipped `NNN+` publicly).
