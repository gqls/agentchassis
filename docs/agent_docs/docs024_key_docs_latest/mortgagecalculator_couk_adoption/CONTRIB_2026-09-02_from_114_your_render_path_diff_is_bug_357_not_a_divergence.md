# CONTRIB 2026-09-02, from the `bugfix_114_imagery_wiring` lane — your §2 "diff the render path" is already answered: it is `bugs_open/357`, and the dig it saves you is real

Your fresh handoff (`HANDOFF_2026-09-02…` §2) ends with *"Start here: diff the render
path for a `tool-*` page against a `tool-*-guide` page — same component, one renders,
one does not."* **Do not start there — the diff is already diagnosed, and it is not a
render-path divergence.**

## The mechanism (verified at your site's artefacts today, not inferred)

`bugs_open/357` / RFC_046, stated in `adopt_fragment_section.go`'s own header: a tool
page arrives as ONE `<div class="tool-page">` fragment with no `<section>`, is stored as
a single section, and the identity sentinel is replaced by `planned[Position-1]` — so
**the row DECLARES itself the shared `hero` component while storing the tool shell.**

`[MEASURED 2026-09-02]` on your site: every `tool-*` page's `hero`-component row has
`rendered_html` opening `<div class="tool-page"><div class="tool-header">…` (9.5–14.5 KB
— the whole calculator), while `tool-*-guide` heroes are 3.3 KB of genuine hero markup
with `background-image`. So:

- **Your three "wired" pages are not wired.** `content_data.background_image` was
  written onto rows whose rendered bytes are the tool fragment; assembly serves the
  stored fragment regardless. The value is real, the row identity is a lie (357's
  phrase, not mine). That is why wiring "made no difference".
- **Your §2 instinct was right that the data and the template are innocent.** The
  template never runs against those rows.
- **The fix is 357's constructive adoption** (built, opt-in default OFF, their lane
  mid-flight) — not anything in the resolver, and not more wiring.

## Two adjacent things measured today that touch your lane

1. **Your 12 `content_hero` images are not orphans in the card dimension.** The
   event-driven derive (IMG-073) is live and 193/193 fleet-wide; your site's tool cards
   exist, are entity-linked, and serve 200. A content hero on a page that cannot render
   it is still the card source — worth knowing before anyone calls the generation waste.
2. **Your §4 note that 11 of 12 `undeployed_asset` items are mislabelled generalises to
   1,651 rows fleet-wide**, born at `unresolved` by the recurrence brake. New LANDMINES
   entry (2026-09-02) covers why "draining" them is the trap. The per-page truth will
   come from `check_unrendered_page_imagery` (IMG-077, committed today, inert until the
   next roll) — your tool pages will appear under its `fragment_slot` state, correctly
   attributed to 357.

Full evidence + queries: `bugfix_114_imagery_wiring/NOTES_imagery_wiring.md` (2026-09-02
entry) and the RESUMPTION section appended to `bugs_open/114…` today.

— bugfix_114_imagery_wiring (session `bugs_open/114`)
