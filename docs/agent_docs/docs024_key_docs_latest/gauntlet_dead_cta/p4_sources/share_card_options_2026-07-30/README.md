# Share-card options — mocks and the measurement behind them (2026-07-30)

For `HANDOFF_2026-07-30_A_share_card_and_the_full_debate.md`, whose first
deliverable is **the owner's choice between three options**, mocked rather than
described. Owner-facing decision page:
https://claude.ai/code/artifact/2cb2166e-ba5e-406d-a6b2-aabfa5fb8d45

## The measurement that decides it

Every figure below is measured, not estimated. All prose is from **round
`39595461-245e-493e-8b48-b0e74faabe1a`**, a real complete round argued
2026-07-30 15:26Z, pulled from the island (`round_real.json`).

| fact | value | how |
|---|---|---|
| complete rounds stored (island) | **51** of 95, 25–30 Jul | `count(*) FILTER (WHERE verdict IS NOT NULL)` |
| average full debate | **3,109 chars** (min 2,396, max 5,073) | sum of the six prose fields, `avg` over the 51 |
| the mocked round | 3,357 chars | same six fields |
| what TODAY's card carries | headline (46) + verdict word (**13**) + date | `buildVerdictCard`, `gauntlet_js_2026-07-29:589` |
| legible capacity of one 1200×630 card | **~700 chars ≈ 23% of a round** | `budget.html`, canvas `measureText` at each size |
| whole debate auto-fits at | **11px** (≈4.6px in a timeline) | `mock1.html`, binary search on fit |
| the exchange auto-fits at | **26px** | `mock1b.html` |
| challenge+defence+reasons auto-fit at | **16px** | `mock2b.html` |
| the hook card | **30px** — most legible of all options | `mock3a.html` |

**Conclusion: no single card can carry a whole debate legibly, and two cards
carry ~46%.** Options 1 and 2 are both *excerpting* strategies; only option 3
carries the round. That is the real choice, and it is not the one the three
options appear to offer.

> **CORRECTION to the handoff's framing.** It warns that option 1 "likely forces
> a 'best exchange' excerpt, which is an editorial choice a machine will make
> badly." **There is no such choice.** A round has exactly one exchange *by
> construction* — `position_text`, `counter`, `challenge`, `defence_text`,
> `verdict` are one column each on one row (`store.Round`,
> `internal/tools-api/store/rounds.go:11`). Nothing is selected from a set; the
> only decision is which of the six fields to include, made once at design time.
> This makes option 1 materially cheaper than the brief implies.

## Landmines earned here

- **`gauntlet_rounds` is NOT in `clients_db`.** It lives on the island
  (`tools_api`/`tools_api` in the island's postgres container). The
  CLAUDE.md `postgres-clients-0` command answers `relation does not exist` —
  which reads exactly like "no rounds have ever been stored".
  Query: `ssh root@toolsapisuk.vs.mythic-beasts.com "docker exec \$(docker ps
  --format '{{.Names}}' | grep -i postgres | head -1) psql -U tools_api -d tools_api -c '…'"`
- **`chromium` here is a SNAP and cannot write to `/tmp`** — `--screenshot` fails
  `Permission denied (13)` while still exiting 0 through a pipe. Render from a
  `$HOME` directory. (These files were built in `~/vonc_mocks_tmp`.)
- **A raw chars-per-line budget over-estimates a card by ~25%** — labels, the
  ruling line and the footer take vertical space the prose then cannot have. My
  first `mock1b` at 32px (599 chars, "inside" a 737 budget) overlapped its own
  verdict line. Auto-fit against the *drawn* layout, not a capacity table.
- **`count(DISTINCT client_ip_hash) = 1` over all 95 rounds.** Every stored round
  is our own harness traffic behind one proxy address; no stranger has argued
  here. So a public record must **start empty**, not be seeded from the table.

## Files

`mock0.html` today's card, reproduced verbatim from the shipped renderer ·
`mock1.html` whole debate, auto-fitted · `mock1b.html` option 1 as it would ship
(the exchange) · `mock2b.html` option 2's second card · `mock3a.html` option 3's
hook card · `mock3b.html` option 3's public record page · `budget.html` prints
the capacity table to `document.title` · `timeline.py` the 504px downscale
composite · `build_artifact.py` inlines the PNGs into the owner's decision page.

Regenerate (from a `$HOME` dir, not `/tmp` — see landmines):

```bash
chromium --headless --disable-gpu --no-sandbox --hide-scrollbars \
  --force-device-scale-factor=1 --window-size=1200,630 \
  --screenshot=mock1b.png mock1b.html
# the measured fit size is written to document.title:
chromium --headless --disable-gpu --no-sandbox --dump-dom mock1b.html | grep -o "fit=[^<]*"
```
