# 263 — decomposition dissolves the tool block's OWN wrapper, so a calculator loses its card panel and its input grid

**Filed 2026-08-12** by the loanandmortgagecalculator lane, during Track B's first three
pages. **Reproduced on 2 of 3 pages, restored on both, Track B stopped.**
**Blocks: Track B (20 remaining pages) and Track C (loancash), and the sibling lane's
`loancalculator_couk/decompose` almost certainly shares it.**

> **On the 2026-07-31 ruling (a cross-cutting root-cause claim goes through `090`, or
> the filer says why they substituted first-hand verification): SUBSTITUTED, and here is
> the substitute.** I reproduced the loss on two independent pages, restored both and
> re-verified the restoration at the artefact, read the two CSS rules that make it a
> visual regression rather than a cosmetic one, censused all 22 pages in scope for the
> shape, and identified the mechanism in `decompose_lmc.py`'s own documented descent
> rule. What a diagnosis run would add is a second opinion on the *mechanism*, and the
> mechanism is stated by the tool's own docstring — quoted below.

---

## The defect

`decompose_lmc.py` walks INTO any node whose subtree holds script-addressed elements and
emits that node's children as separate blocks. Its docstring says so:

> *"a node whose subtree holds script-addressed elements is walked into, separating prose
> siblings from widget machinery"*

That is correct and deliberate for the page-level wrapper (`div#content`, `.container`) —
the assembled-layout shim in the chrome head compensates by making `<main>` the
container. **But the calculator's own presentation wrapper is inside that subtree too,
and nothing compensates for it.** On a page shaped

```html
<div class="container"><div class="card"><div class="calc-grid">
  <div class="form-group">…4 of these, each holding a script-addressed input…
```

the descent dissolves `.container`, `.card` AND `.calc-grid`, emitting the four
`form-group` divs as bare children of the assembled section.

## What it costs, measured (this is not cosmetic)

From the live stylesheet, `style.css`:

```css
.calc-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 34px; }
.card      { background: var(--card-bg); padding: 30px; border-radius: var(--radius);
             box-shadow: 0 10px 15px -3px rgba(0,0,0,.08), …; }
```

So losing them means the calculator's inputs **stop being a two-column grid and stack**,
and the calculator **loses its panel** — background, 30px padding, radius, shadow. The
arithmetic is unaffected: every script-addressed id survives, and the oracle passed.

## Measured per page (2 of 3 hit, restored)

| page | `container` | `card` | grid | verdict |
|---|---|---|---|---|
| `loans-standard-calc` | 1→0 | **4→3** | `input-grid` 1→1 | calculator's panel dissolved; 3 prose cards survived — **RESTORED** |
| `mortgages-repayment` | 1→0 | 2→2 | 1→1 | genuinely clean, left decomposed |
| `mortgages-overpayment` | 1→0 | **1→0** | **`calc-grid` 1→0** | panel AND input grid dissolved — **RESTORED** |

**Census of the 22 pages in Track B's scope, at pin `7e6b993ef`: 20 carry a `card`
and/or a `calc-grid`/`input-grid` wrapper.** Only `mortgages/fact-finder` and
`mortgages/portfolio` have neither. So this is not an edge case — it is the normal shape.

## ⚠ WHY BOTH EXISTING GATES PASSED IT — the part worth reading

Two independent checks certified a page that had lost design:

1. **`deploy_pages.py`'s byte-for-byte diff said IDENTICAL.** It compares the served page
   against a prediction **generated from the same manifest that dropped the wrapper**. It
   proves fidelity to the model, not preservation of the original. A prediction diff can
   never catch a decomposition defect — only a BEFORE/AFTER comparison can.
2. **My own class check said "no classes lost".** It compared class *sets*. The page has
   four `card`s; the calculator's was dissolved and three prose cards survived, so the
   set was unchanged. **And the aggregate attribute count was 18 before and 18 after** —
   because two removals (`container`, one `card`) were exactly offset by two
   `ported-prose` additions. I reported that page to the owner as proven on the strength
   of it.

**The only check that catches this is a PER-CLASS COUNT diff, before vs after:**

```python
def counts(html):
    c = {}
    for m in re.findall(r'class="([^"]+)"', html):
        for x in m.split(): c[x] = c.get(x, 0) + 1
    return c
drops = {k: (v, after.get(k, 0)) for k, v in before.items() if after.get(k, 0) < v}
```

`container: 1→0` and `ported-prose: 0→2` are the expected, compensated changes. **Any
other drop is this bug.**

## Fix candidates, ordered by what closes the door

> **⚠ AMENDED 2026-08-12, before implementation, at the owner's request to check the
> plan.** Candidate 1's rule as first written — *"stop at the first node that is
> machinery only"* — **is underspecified and would still have dissolved 2 of the 22
> pages.** Verified structurally against four real pages plus a wrapper census; the
> corrected rule and its cost are below. Candidate 2 is now definitively rejected on
> evidence. **The check in the "verify" section is unaffected and remains the first
> thing to build.**

1. **Keep the tool block's wrapper chain in the tool block** — emit the wrapper ITSELF
   rather than recursing past it. This is *not* the "whole-child marking" the docstring
   says was tried and refuted: that marked the whole `<article>` as tool, freezing
   `stamp-duty`'s h1 and intro into the locked row. Verified skeletons:

   ```
   stamp-duty:  article > [h1, p, div.card > calc-grid > (inputs, result-box), h2, p]
   standard-calc: container > [h1, p, div.card > (input-grid, results-box),
                               section > (h2, 3× div.card of PROSE)]
   overpayment: container > [p, div.card > calc-grid > (inputs+button, result-box)]
   ```

   **THE RULE, CORRECTED: the tool block is the OUTERMOST ancestor of the widget's
   controls that still has at least one PROSE SIBLING.** Equivalently: never descend
   past a node whose entire content lies within the widget's own subtree. On all three
   skeletons that resolves to `div.card`, and on `standard-calc` it distinguishes the
   calculator's `.card` from the three PROSE `.card`s **by content, not by class name** —
   which no name-based rule can do.

   **THE IMPLEMENTABLE FORM, read off `split_ordered` itself** (`loancalculator_couk/
   decompose_pages.py:104`). The existing code already emits a holder WHOLE when several
   siblings hold marked ids (`len(holders) > 1`); the loss happens only in the
   `len(holders) == 1` branch, which recurses and so dissolves that single wrapper:

   ```python
   if len(holders) == 1:
       split_ordered(c, src, ids, out, depth + 1)   # <- dissolves c
   ```

   **Descend into a single holder only if that holder still has a child that is NOT a
   holder** — i.e. only if there is prose inside it to separate. Otherwise emit it whole.
   Traced against the skeletons: `#content > .container` descends (container holds h1, p
   and a prose `<section>`); `.container > .card` stops (card's children are all
   machinery); `article > .card` on stamp-duty stops, while the article itself still
   descends. That is the whole fix, it is four lines, and it is vocabulary-agnostic.

   > **⚠ CORRECTED AGAIN 2026-08-12, and the correction shrinks the cost from 2 pages to
   > 1.** The amendment above originally claimed `mortgages/overpayment` AND
   > `loans/damage-checker` both hold static copy inside the panel. **`overpayment` was a
   > FALSE POSITIVE**: I measured a fixed 2,400-character window from `class="card"`
   > instead of the element's real extent, and the window ran past the card's close. Walked
   > properly, `div.card`'s only child is `div.calc-grid` — machinery-only, no cost.
   > **Same error family as the day's other three** (a window/aggregate that answers a
   > different question than the one asked); logged in `WRONG_CALLS.md`.

   **THE REAL COST, one page.** `loans/damage-checker`'s card genuinely mixes: its direct
   children are `h3`, a long advisory `<p>` (*"Lenders typically charge between £50 - £150
   per panel…"*), four `div.checklist-item` (each holding controls) and a `results-box`.
   Because several children hold marked ids, the EXISTING `len(holders) > 1` branch
   dissolves that card today and will still dissolve it after this fix. So for this one
   page the choice is explicit: **lose the panel, or freeze its `h3` + advisory paragraph
   into the locked tool row.** Precedent for the latter is already in the lane —
   `mortgages/simple.html` *"is one card containing everything — its widget-internal text
   is out of the voice's scope by the 'copy zones only' rule"* — and
   `decompose_pages.py` already carries a per-page override table for exactly this class
   of exception. **Decide it per page, with the trade-off written down, and do
   `damage-checker` last with `fact-finder` and `portfolio`.**
2. ~~**Re-wrap on load.**~~ **REJECTED on evidence 2026-08-12.** It would hard-code a
   class vocabulary, and the census found **at least eight** wrapper vocabularies across
   the 22 pages — `card`, `calc-grid`, `input-grid`, `result-box`, `question-section`,
   `q-row`, `metric-card`, `table-container`. **Six of the 22 have no `card` at all**
   (`bridging-loan`, `equity-release`, `fact-finder`, `fee-analyser`, `portfolio`,
   `rate-forecaster`), and two of those are structurally unlike everything else:
   `fact-finder` is 11 × `q-row` inside 4 × `question-section`, and `portfolio` is
   4 × `metric-card` plus a `table-container`. Any name-based re-wrap would need all
   eight and would still miss the ninth. **The corrected candidate 1 and the count gate
   are both vocabulary-agnostic, which is the whole reason to prefer them.**

   **Consequence for ordering:** `fact-finder` and `portfolio` are the most complex
   shapes AND the two with no panel to preserve. **Do them LAST, not first** — the
   earlier plan of "simplest first" accidentally had `fact-finder` in the early group.
3. **Add the classes to the assembled-layout shim** (`section.ported-prose > .form-group`
   grid rules). Rejected: it makes every assembled section a calculator grid, and the
   shim is shared by 18 prose pages.

**Whichever is chosen, the acceptance test is the same and it must be added to the lane's
tooling, not run by hand:** per-class count diff, before vs after, on every page Track B
touches. That gate would have stopped this at page 1.

## How to verify a fix

```bash
# per page: capture BEFORE, apply, deploy, then
#   counts(before) vs counts(after) -> only container:1->0 and ported-prose:0->N permitted
python3 gate_component_bytes.py          # bytes still faithful
cd $LANE && python3 oracle.py --tools <tool> && python3 oracle.py --mutate expectation --tools <tool>
```

## State right now

- `loans-standard-calc` and `mortgages-overpayment`: **restored to verbatim, redeployed,
  re-verified at the artefact** — zero class-count drops against the pre-Track-B original.
- `mortgages-repayment`: **left decomposed** (`["prose-0","tool-1","prose-2"]`, tool row
  `lock_type='permanent'`), because its wrapper survived and all four gates passed
  honestly.
- The other 19: untouched, still `["ported-page"]`, still `rebuild_policy='owned'`.
- No page was flipped to `generic` at any point — see the lane NOTES for why that
  deferral was already in force, and note that it also meant this defect could not be
  compounded by a generic rebuild.

## See also

- `bugs_open/253_…strips_every_layout_component` — the *sibling* defect and a genuinely
  different one: 253 is the generic pipeline REWRITING a decomposed page. This is
  faithful decomposition itself. Both destroy layout classes; neither implies the other.
- `docs024_key_docs_latest/loanandmortgagecalculator_couk/NOTES`, 2026-08-12 entries —
  the full Track B run, including the two instrument repairs that preceded this.
- `WRONG_CALLS.md`, 2026-08-12 — the netting-out aggregate that let me report page 1 as
  proven.

---

## CONTRIBUTION 2026-08-12 22:0x (second thread, independent harness) — the fix is CONFIRMED, and the residual is **6 pages, not 1**

I was asked to take this bug on and had a plan written when `71fb31a03` landed the fix
underneath me. I stopped rather than compete, and spent the time re-measuring the landed
fix with a **separately written** descent harness and per-class counter. Everything below
is measured at the pin `decompose_lmc.py` currently names, over the 21 in-scope pages
(`loans/consolidation` and `mortgages/repayment` have no `div#content` at the pin — already
decomposed — and are skipped, which is why 21 and not 22).

**The fix is right and I am not proposing to change it.** Independent corroboration:

| | pages dropping a layout class | pages whose frozen-text share rises |
|---|---|---|
| before (`keep_widget_wrapper=False`) | **13** of 21 | — |
| after (landed, `=True`) | **6** of 21 | **0** |

"0 pages freeze prose" is the property that matters and it holds — the landed rule is
purely additive against the old one. I also ran an *alternative* rule of my own that
preserves three more panels, and rejected it: it buys those three by freezing 33%→69%,
26%→37% and 58%→73% of those pages' visible text, and on two of them the frozen text
includes the page's `<h1>`. Freezing an h1 is exactly what refuted whole-child marking on
2026-08-05. **The landed trade-off is the better one.**

### The correction: this file says the real cost is one page. The gate says six.

`gate_wrapper_parity.py` on a 21-page manifest reports `21 page(s) checked, 6 failing` —
and my independent counter names the same six. This file's amendment names only
`loans/damage-checker`, because it characterises the surviving loss by the branch that
causes it (`len(holders) > 1`, damage-checker's shape). **Five of the six fail through the
other path**: the new single-holder guard descends whenever the holder has *any* non-holder
child, and on these pages that child is the page's own heading and intro sitting inside
the panel.

| page | the panel | its direct children (`[…]` = holds machinery) | h1 inside |
|---|---|---|---|
| `mortgages/rate-forecaster` | `article.card` | `h1`, `p`, `[div.calc-grid]` | **YES** |
| `mortgages/fee-analyser` | `article.card` | `h1`, `p`, `[div.calc-grid]` | **YES** |
| `mortgages/equity-release` | `article.card` | `h1`, `p.intro`, `[div.calc-grid]` | **YES** |
| `mortgages/bridging-loan` | `article.card` | `h1`, `div`, `p.intro`, `[div.calc-grid]` | **YES** |
| `mortgages/simple` | `div.card` | `h2`, `[div.calc-grid]` | no |
| `loans/damage-checker` | `div.card` | `h3`, `p`, `4× [div.checklist-item]`, `[div.results-box]` | no |

So **all six are one shape, not two**: a panel that mixes the page's own copy with the
widget's machinery. Single-holder or multi-holder is incidental.

### What that changes about the choice this file offers

The amendment offers, for `damage-checker`, "lose the panel, or freeze its `h3` + advisory
paragraph", citing `mortgages/simple` as precedent. That reasoning transfers to `simple`
(an `h2`) and to `damage-checker` (an `h3`). **It does not transfer to the other four,
where the frozen text would be the page's `<h1>` and intro** — a materially different
concession from "widget-internal text", and the one the 08-05 refutation was about.

Measured, so the choice is costed rather than discovered:

- **`bridging-loan`, `equity-release`, `damage-checker`** — a descent rule *can* keep these
  panels (I ran one that does). Cost: frozen share 33%→69%, 26%→37%, 58%→73%, with the h1
  frozen on the first two.
- **`fee-analyser`, `rate-forecaster`, `simple`** — **no descent rule can keep these**,
  whatever its stopping condition. Their `#content > .container` holds nothing but the
  panel, so the panel is the first and only node with prose to separate: stop above it and
  the whole page is one frozen row, descend into it and it dissolves. Preserving these
  three needs the panel's own tags *re-emitted* around the tool block, which is the first
  thing that would make a block something other than a byte slice of the source. That is a
  real decision, not a tuning, and it should be taken deliberately or not at all.

**Recommendation, offered not imposed:** leave all six refused by the gate and verbatim,
which is what happens today, and take the six as one decision later. Nothing is lost by
waiting — a refused page is untouched, and the other 15 can proceed now.

### One hazard in `gate_wrapper_parity.py`, inert today

`class_counts()` counts `class="…"` in raw HTML without stripping `<script>`/`<style>`.
**Verified harmless for this lane right now**: zero of the 22 tool pages carry a `<script>`
or `<style>` inside `div#content`, so both sides are code-free. It is a hazard on reuse:
the sibling lane's `tools/consolidation.html` builds rows in a JS template literal, and
counting it yields **five false drops** (`d-bal`, `d-months`, `d-name`, `d-rate`,
`remove-btn`) — measured, and 0 once code is stripped. The same mechanism can also *hide* a
real drop, exactly the way `ported-prose` hid one in this bug.

A hardened, lane-agnostic version of the predicate now exists at
**`scripts/class_count_delta.py`** (`6427f157f`): strips script/style/comments, refuses an
empty side rather than reporting a confident number, exact-equality and
additions-allowlisted modes, and a self-test containing its own induced failure. It
**postdates `gate_wrapper_parity.py` and does not replace it** — offered for the served-page
seam, for `bugs_open/253`, and for the `mortgagecalculator_couk_adoption` port, whose
input is a bucket file of unknown shape.

*Method, so this is checkable: separate descent implementation, separate class counter, same
pin. The two agree on all 21 pages. Where I quote a frozen-text share it is
`visible(strip_code(...))` over tool blocks ÷ over the whole `#content` subtree.*
