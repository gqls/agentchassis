# HANDOFF 2026-08-09b — the CONTRAST front: cold-start for a fresh chat

**This is a third, separate front in this lane. It supersedes nothing.**
`HANDOFF_2026-08-09_sweep_front_continue_here.md` (fundamentallyai sweep) and
`HANDOFF_2026-08-05_continue_here.md` (camera / contact-sheet / 151-checker) are both
still current and are other threads. Do not merge these three.

Read in this order: this file → `bugs_open/113` (the 2026-08-09 sections, both of them) →
`bugs_open/122` (from "Re-measurement 2026-08-06" to the end).

**The front in one sentence:** live sites are serving text nobody can read, the cause is
always a *pairing* of two individually-valid colours, and the only thing that sees it is
`scripts/render_audit.py` rendering the real page.

---

## 0. THE ONE THING OWED — a two-file SQL change, written and measured, NOT APPLIED

```
docs/agent_docs/sql_for_agents/353_news_list_tag_ink_fix.sql
docs/agent_docs/sql_for_agents/353_news_list_tag_ink_fix_ROLLBACK.sql
```

**Blocked, not undecided.** The production `UPDATE` was refused by the session's
permission classifier. Nothing about the change is open — it is measured on all eight
consumer sites, guarded, and reversible. It just needs running:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/353_news_list_tag_ink_fix.sql
```

**What it does.** One declaration in the shared `news-listing` component
(`11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f`): the topic chip's ink goes from
`--color-text-muted` to `--color-text`. The chip's `--color-border` fill is deliberately
left alone.

**Why it is worth doing first:** it is one line, and `.news-list-tag` is **181 of the 442
failures I measured (41%)** — the single largest contributor across the three worst sites.
**7 of its 8 consumer sites fail AA today** (worst `idea.uk` 2.25:1); all 8 pass after,
minimum 9.38:1. Light and dark both, so it is scheme-independent.

**Backup exists**: `bak_cc_newslisting_20260809`, 1 row, 4462 bytes, taken before the
attempt. The rollback file restores from it.

**Two things about 353 that are NOT verified, and you should not read them as verified:**
- **its `DO`/`RAISE` guard has never been induced** — the arithmetic in it is checked
  (anchor occurs exactly once; 4462 − 6 = 4456 bytes) but the abort path has not been
  made to fire, because I could not write to the DB at all. This repo's own rule is
  "use `DO`/`RAISE`, **and induce it**". Induce it before trusting it: temporarily change
  the expected byte count to a wrong number and confirm the transaction aborts.
- the `updated_at` bump is the only other column it touches.

**After applying, the repair is a PAGE re-render, not a stylesheet rebuild** — see §2,
this is the opposite of what `bugs_closed/072` would lead you to expect and it is the
cheap half of this front.

---

## 1. What this front established today (all committed)

Commit `7b7f37d94` — measurement into `bugs_open/113` and `bugs_open/122` plus a LANDMINE.

**Ran the three-site sitemap audit that 113 named as its cheapest next step.**
**442 solid-background AA failures over 61 pages**, against the 7 that one component on
one page had produced:

| site | pages | solid failures | ≤1.1:1 (invisible) |
|---|---|---|---|
| robot-hands.com | 19 | 193 | 32 |
| dartsonline.com | 18 | 125 | 19 |
| ai-agent-orchestration.com | 24 | 124 | **73** |

**The run discriminates between 113 and the adjacent primary-as-ink defect**, which
palette arithmetic could not. `dartsonline` defines **no `card_bg`** in its `palettes`
row yet serves one **equal to its `surface` to the byte** — that is
`fillDarkSchemeSpecialisedSlots`' signature and nothing else produces it. So **113's fix
is now proven on a served artefact of a site nobody repaired by hand**, which is a
stronger claim than the fundamentallyai verification in that file (that site *was* hand-
repaired in the same session).

**`ai-agent-orchestration.com` is the one of the three still carrying 113 itself** —
serving `--color-card-bg: #ffffff` on a `#080B10` background, 44 of its 124 failures.
Not repainted, per 113's own instruction.

**Also closed two open questions in 113:** the demand-side count it said it had not
enumerated (**31 ink placements over 29 deployed pages**, a floor — chrome is not in
`page_components`), and its `53`-component figure, which I checked because its regex
`color:\s*var\(--color-primary\)` **also matches `background-color:`**. It holds: 54
match, **51 are genuinely ink**.

---

## 2. The mechanism finding that changes the repair path

**The `.news-list-tag` rule ships TWICE** and the copies are identical today:

- inline in each page's `<style>`, emitted from `content_components.html_template`;
- again in `styles.css` (frozen — a stylesheet is written once, `bugs_closed/072`).

Measured on `robot-hands.com/news/index.html`: the `<link>` to `styles.css` is at byte
**8412**, the inline rule at byte **38425**. Later, equal specificity → **the inline copy
wins.** Therefore **a page re-render repairs the page with no stylesheet rebuild**.

That is the opposite of 113's situation and it is why this fix is cheap. Six of the eight
sites also carry the stale rule in `styles.css` (`fundamentallyai.com` and
`ai-agent-orchestration.com` do not) — overridden and harmless, but **it will keep reading
as the old value to anyone grepping the stylesheet**, which is a trap for whoever verifies
this. Verify at the rendered page, not at the CSS.

---

## 3. Where to go next, in the order I would take them

1. **Apply 353** (§0), then **re-render one news page** — `robot-hands.com/news/` carries
   **105 chips** — and re-run the audit. Expect robot-hands −128, dartsonline −53.
   **Run `render_audit.py --sitemap` BEFORE as well as after.** That is 113's own
   transferable lesson and it is the only thing that catches a repair making some other
   page worse, which has happened on this exact bug family before (the cost-calculator
   mirror defect, 113, 2026-07-27).
2. **`ai-agent-orchestration.com`** — the biggest single win left (124 failures, 73 of
   them invisible) and the last of the three carrying 113 proper. **Do not just fire a
   re-render at it.** It has **no `palettes` row at all** (`source_domain` → 0 rows) and
   its `site_specs.design_intent.color_scheme` is a **light** scheme (`background:
   #ffffff`) on a site whose own `avoid` list contains "Bright white backgrounds" and
   whose `style_direction` is `professional-dark`. The served sheet is dark, so that spec
   is not what rendered — **where its palette actually comes from is `[UNMEASURED]` and is
   the first thing to establish**, or a re-render may pull the light scheme in and make it
   worse. This is also 122's sub-shape B, whose cause that file deliberately routed to
   `090` rather than guess. **Keep it there — file the `090`, do not guess.**
3. **The rest of the 442.** After the chip fix and aao, what remains is mostly 122's
   sub-shape A (`--color-primary` spent as both fill and ink), which is the renderer
   proposal at the end of 122 (`--color-primary-ink`) and a much larger piece of work.
4. **`render-audit-agent` still has no dispatcher** — 122's 08-06 section: the Go port,
   the orchestration and the work-item drain are all live, and **28 enabled
   `scheduled_tasks`, none targeting it**. Everything on this front was found by hand-
   running a Python script. That is the actual leverage here and it is one row.

---

## 4. Traps this front paid for — read before trusting your own numbers

1. **`render_audit.py`'s printed total is not the figure to quote.** Two errors, opposite
   directions, both look correct. Now a LANDMINE entry (footprint `scripts/render_audit.py`).
   - **No `--sitemap` = you measured `index.html`.** Same tool, same week:
     robot-hands **3 → 193**, dartsonline **1 → 125**. Nothing regressed — the 08-06 fleet
     run in 122 simply never opened those pages. Any per-site figure in either bug file
     that predates 08-09 is a homepage figure.
   - **~9% of rows are the probe's own guess.** `render_audit.py:111-114` pushes a mid-grey
     `rgb(128,128,128)` under text whose backdrop is a background image *or gradient*, and
     flags it `overImage` so it can be discounted. **41 of 483** here, all under the two
     gradient CTAs. The terminal output does not mark them; filter on `overImage` in the
     `--json`.
2. **A "colour" regex cannot tell an ink from a fill.** `color:\s*var\(--color-primary\)`
   matches `background-color: var(--color-primary)` too. It happened to cost only 3
   components here, but that confusion *is* this entire bug family.
3. **`grep -o '\.rule *{[^}]*}'` finds nothing when the rule spans lines** — grep is
   line-based, so this reads as "the rule is absent" on CSS that plainly contains it. I
   briefly concluded two sites lacked a rule they had. Use Python with `re.S`, or gate on
   a plain substring match first.
4. **A shared component's blast radius is its consumer palettes, not its call count.**
   The 07-29 prescription in 122 (`surface` fill + `text` ink) was half wrong, and only
   computing it against all eight served palettes showed which half: `surface` vs the
   section `background` is **1.04–1.22 everywhere**, so that chip stops being a chip.
   Eight sites is one query and one loop; guessing would have shipped a flat chip fleet-wide.

---

## 5. Commit trail (this front)

`7b7f37d94` three-site audit + contributions to 113 and 122 + LANDMINE ·
this file's commit (handoff, 353 SQL pair, 122 fix section).

No council submission: scope is `bugs_open/`, `docs/` and a `content_components` row —
the gate refuses docs client-side and this is not `platform/`, `internal/` or `pkg/`.
