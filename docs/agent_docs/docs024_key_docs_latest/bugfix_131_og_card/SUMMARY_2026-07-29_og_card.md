# SUMMARY — 29 July 2026 — the estate's social previews now work

Milestone read-out for `bugs_open/131` (og-card slug). Current state only; the chronology is in
NOTES and `README_where_we_are.md`.

## What we're trying to do

Make the estate look like itself when a link to it is shared. Every page of every site tells
WhatsApp, Slack, X, LinkedIn and iMessage "here is a preview picture for this page". On 11 of 14
sites that picture did not exist, so every share rendered blank. It had never worked.

## Where we've come from

The defect was found on 2026-07-28 out of the relojistas discoverability work — an audit of what
crawlers and social unfurlers actually receive. It was filed measured but unfixed, with three
ranked fix candidates, the first being "stop emitting the tag unless the asset exists" and the
second "generate the card", flagged `[UNDIAGNOSED]` as to how the two working sites got theirs.

## What we've done

**Answered the open question: nothing needed building.** `derive_brand_head_assets` has been
registered and live since 2026-07-11, reachable in production by queuing a work item at
`asset-deployer` in `brand_head` mode. Nobody had ever run it. "Nothing generates og-card.png"
was true of what had *run* and false of what *exists*.

**Corrected the fix ordering, on measurement.** All 14 sites have an active `logo` asset — the
generator's only precondition — so generating was available everywhere, needing no code, no
council, no roll and no chrome re-render. Suppressing the tag would have needed all of those, to
deliver *no* preview instead of a *working* one. The case file's ranking was sound in the
abstract and wrong for this estate.

**Ran it.** relojistas first as a pilot (404 → live card in 18 seconds), then the remaining ten.
**12 of 13 sites with an `og:image` tag now serve a real 1200×630 card**, verified on the wire
and re-verified after the v1.0.1196 roll.

**Looked at every single card.** This is the part that mattered. Eight are good —
vetcomparison.uk best, oufe.com and fundamentallyai.com carrying their names clearly. Two are
wrong, and one failed.

## Where we are now

**Working and done: 8 sites.** Nothing owed on them.

**Three sites need a corrected logo asset, and all three are blocked on the same thing.**
relojistas' card is a two-up brand *specification sheet*; gaswholesalers' is a **3×3 contact
sheet of nine rejected logo concepts** with hallucinated lettering ("GAAS", "WALSE", "WHOLACS");
idea.uk's derivation fails outright. The generator is faultless in all three cases — what is
stored as those sites' "logo" is not a logo. The corrected relojistas crop is **made and
verified by eye**, but installing it needs a write to S3, and reading the storage credentials
was refused by the permission classifier. **That decision is with the owner and is the critical
path.**

**Three further defects found, recorded, none fixed:**
- the favicon derivation uses a **non-proportional** `resize.Resize(64,64)`, so every wide logo
  is squashed illegible — `composeOGCard` gets this right, the favicon path does not;
- **almost every card shows a letterbox rectangle**, because logo assets are opaque and get
  painted onto a differently-coloured card. Knock-out fixes it and is already house practice;
- **2 of 14 `logo` rows hold a deployed web path instead of an S3 URI** (idea.uk,
  leopardess) — which is why idea.uk cannot derive, and, inversely, the only reason
  leopardess's owner-approved card has never been overwritten.

**And a bigger one, outside this bug's frame.** relojistas' header logo is that same spec sheet,
rendered at roughly 81×44px with no crop, on every page, for weeks — on a site whose own handoff
called it "finished". The audit behind that word checked what crawlers receive and never once
what a person sees. Logged in `WRONG_CALLS.md`.

## Where we're going

1. **Owner decision on S3 credentials** — unblocks relojistas, gaswholesalers and idea.uk
   together. Everything else is downstream of it.
2. **Fix 1, the code gate** — still worth landing as the structural guard, now with idea.uk as
   live proof the bad state still occurs. Design constraint recorded: it must not key on an
   `og_card` asset row, or it regresses leopardess.
3. **The three recorded defects**, in the order the owner values them.
