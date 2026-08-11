# 253 — a framework rewrite of a decomposed prose block strips every layout component

**Filed 2026-08-11.** Observed live on `loanandmortgagecalculator.co.uk/index.html`
within four hours of that page being decomposed. **This is the finding that governs
Track B**, and it was not predicted by any of the decomposition briefs.

---

## What happened

Track A decomposed the LMC homepage at 11:16Z into a single `prose-0` component
holding the page's hand-built body byte-for-byte. At **15:47Z** the generic pipeline
rewrote that component. The rewrite kept the words and the links. It removed the
site's entire visual vocabulary.

| markup in the block | before (Track A, byte-identical to the hand-built page) | after the rewrite |
|---|---|---|
| `class="card"` | 18 | **0** |
| `tool-grid` | 3 | **0** |
| `btn-primary` | 15 | **0** |
| `highlight-box` | 1 | **0** |
| `class="hero"` | 1 | **0** |

Links survived — 14 calculator links before and after, and *more* internal links
overall (28 → 34). So this is **not** a content or navigation loss, and the first
read of the diff (mine) wrongly suggested it was. It is a **presentation** loss: the
site's shopfront went from a styled calculator directory to a flat run of headings
with bare "Open calculator" links. `prose-0` went 6,958 → 5,832 bytes.

## Why this is the important one for Track B

A decomposed calculator page is `["prose-0", "tool-1", "prose-2"]`. The tool row is
**locked**, so the calculator itself is safe — `bugs_open/058`'s lock holds it, and
the matching rule is now pinned by
`save_sections_positional_tool_slot_test.go`. **The prose rows around it are not
locked and are exactly what was rewritten here.** So Track B's expected outcome is:
the calculator keeps working and keeps its markup, while the cards, buttons and
grid framing it are silently flattened on the next generic rebuild.

**That is a materially different risk from the one Track B was authorised against.**
Every brief to date framed the calculator page danger as "the widget gets replaced
by prose or moved to the bottom". Both of those are now guarded. This one is not
guarded, is not silent-but-harmless, and lands on 22 live consumer-finance pages.

## What is NOT wrong here

- **The framework rewriting this copy is correct and intended.** Owner ruling
  2026-08-06: the framework writes the content, not a CLI session. Decomposition
  exists precisely so it *can*. Do not read this bug as an argument against
  decomposition.
- **The shrink guard worked.** An earlier attempt at 15:24Z was REFUSED —
  `prose-0 3776→1334 chars (35% kept, floor 50%)`, `bugs_open/178`, nothing written,
  and it raised `save_refused_incomplete:index` for a human. The 15:47Z save was
  within the floor (84% kept) and proceeded. **The guard measures TEXT VOLUME, and
  is blind to markup.** A rewrite can keep 84% of the words and 0% of the components.
- **No other Track A page was touched** — only `/index.html` has a `page_components`
  row updated after 12:00Z. This is one occurrence, not a sweep.

## Root cause — where to look

Not yet established, and this is the honest state of it. `[INFERRED]` The writer is
handed the section's content and asked to produce the section's content; nothing in
that contract obliges it to preserve component classes it did not author, and a
`ported-prose` block has no schema describing what its markup means. The likely fix
sites are the `page-content-writer` prompt/config and whatever validates a section
write — but **that is a hypothesis, not a diagnosis, and it should go through `090`
before anyone asserts it.** The observation above is solid; the cause is not.

## Fix candidates, ordered by what closes the door

1. **A markup-preservation floor beside the text floor.** The shrink guard already
   exists and already raises a human-reviewable item; it is the natural home. Assert
   that a same-named prose slot may not lose more than N% of its *component class
   occurrences* in one save. This makes the bad state detectable by the mechanism
   that already stops the analogous text case, and it is the smallest change that
   would have caught this.
2. **Lock prose rows that carry layout components.** Blunt: it also stops the
   legitimate rewrites decomposition exists to enable, so it trades the whole benefit
   for the protection. Not recommended except as a stopgap on specific pages.
3. **Give `ported-prose` a schema the writer must satisfy.** Correct in the long run,
   much larger, and it is really a question about how verbatim-adopted markup gets
   described to a writer at all.

**Do not "fix" this by restoring the page and calling it done** — it will be rewritten
again by the next rebuild, and the second occurrence will look like a new bug.

## Immediate decision owed on the live page

The LMC homepage is currently serving the flattened version. Options: leave it (the
framework's output stands, the site looks plainer), or repair `prose-0` with the
original markup and accept that a future rebuild may flatten it again until (1) is
in place. **`load_lmc.py --apply index` will REFUSE** — its pre-write guard checks the
stored md5 against the 08-09 baseline and the row has legitimately moved, which is
the guard doing its job. A repair therefore has to be deliberate and targeted, not a
re-run of the lane tooling.

## See also

- `bugs_open/178` — the shrink floor that fired at 15:24 and correctly held.
- `platform/orchestration/actions/save_sections_positional_tool_slot_test.go` — why
  the locked tool row itself is safe.
- `loanandmortgagecalculator_couk/NOTES_…md`, 2026-08-11 — Track A's full record,
  including that a `predicted/` file is only valid until the framework next writes
  the page. That is how this was caught: a post-roll mirror check flagged the
  homepage as differing, and the difference turned out not to be the roll at all.
