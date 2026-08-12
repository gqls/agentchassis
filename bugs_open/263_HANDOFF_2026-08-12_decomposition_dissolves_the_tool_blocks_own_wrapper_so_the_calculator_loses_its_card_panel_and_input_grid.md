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

1. **Keep the tool block's wrapper chain in the tool block.** When the descent decides a
   node is widget machinery, emit the node ITSELF rather than its children — i.e. mark
   the outermost element whose subtree is entirely widget machinery, instead of
   recursing past it. This is the "whole-child marking" the docstring says was tried and
   refuted — but it was refuted for a DIFFERENT reason (`mortgages/stamp-duty` froze an
   intro paragraph into the locked tool row because prose and widget shared one
   `<article>`). The correct rule is narrower than either: descend only while a node's
   subtree contains BOTH prose and machinery; stop at the first node that is machinery
   only, and emit it whole.
2. **Re-wrap on load.** `load_lmc.py` reconstructs `<div class="card"><div
   class="calc-grid">…` around the tool block's contents. Cheaper, but it hard-codes one
   site's class vocabulary into the loader, and the loan side uses `input-grid` while the
   mortgage side uses `calc-grid` — two vocabularies, per the stylesheet's own header.
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
