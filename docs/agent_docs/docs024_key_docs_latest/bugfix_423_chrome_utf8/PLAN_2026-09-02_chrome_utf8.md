# PLAN — bugs_open/423: a chrome store that Postgres refuses, reported as silence

**Started 2026-09-02.** Bug filed 2026-08-31 by the delivery lane (boxingonline session),
which captured the mechanism live and left half 2's source open. No lane had taken it:
`scripts/who-owns.py 423` said OWNED-or-recently-active on the strength of the FILING
commits, and the filer's own NOTES said *"Until a lane takes 423, the hand-patch serves"*.
So the fix thread was unclaimed and is resumed here.

## The design decision, and why it is not "fix line 1622"

Three fixes were available and they are not alternatives — they close different doors:

| | closes | leaves open |
|---|---|---|
| fix `buildServicesHTML`'s `w[:1]` | this footer, on these two sites | the other 7 call sites; the next hand-roll |
| + one shared primitive, all 8 sites converted | the whole known class | a NINTH site written tomorrow |
| + a pre-store UTF-8 gate that names the OFFSET | any future member, wherever it is written | (nothing, for this seam) |

Ranked by **what makes the bad state unrepresentable**, per the estate norm. The reason
all three shipped together is historical and specific: the estate had **already** fixed
the TRUNCATION shape of this exact class on 2026-07-20 (`SafeCut`, `bugs_open/027` §4b)
and never went looking for the CASING shape. One shape of a class getting fixed while its
siblings survive **is** the failure mode here, so a one-site fix would have been the same
mistake a third time.

### Why REFUSE and not sanitise (the one real design choice)

`strings.ToValidUTF8` would make the store succeed. Rejected: this path has **no gate
downstream** — whatever it stores is what the site serves (`bugs_open/260`) — so
sanitising would ship silently mangled text over working chrome AND leave the upstream
cutter in place to keep producing it. The precedent for sanitising
(`platform/messaging/processor.go:638`) is for INBOUND raw bodies, where there is no
authored artefact to protect. Refusing keeps the previous bytes serving, which is the
disposition this function already takes for every other kind of bad render.

### Why the gate reports an OFFSET

Postgres names the offending BYTE and never its POSITION. On a 40 KB document that is
"0x80" and nothing else, and the only way to locate it has been to bisect the pipeline by
hand — which is most of what this bug cost. `InvalidUTF8At` returns the offset plus a
`QuoteToASCII`'d window. **ASCII-only is load-bearing, not tidiness:** the report is
logged, persisted and put in a work item summary, so a window carrying the raw bad bytes
would reproduce the defect it describes. That is not hypothetical — it is edit 5 below.

## Phasing

1. **Primitives** — `datahelpers.UpperFirst`, `datahelpers.InvalidUTF8At`, beside `SafeCut`.
2. **The class** — all 8 call sites converted in one pass, licensed by an ASCII-parity test.
3. **Half 1** — the store-failure branch takes the existing `chrome_render_failed` surface.
4. **The latent sibling** — three `summary[:247]` byte slices → `SafeCut`. Same pass as 3,
   non-negotiably: 3 puts an arbitrary error string through 5's truncation.
5. **The gate** — before the `UPDATE`, after every transform.
6. **The prose** — `emitChromeRenderFailedItem` gains a `phase`. Reusing a surface means
   fixing its operator-facing text; it said "the template could not be executed", which is
   now false for two of its three callers.

## Decisions, with reasons

- **`phase` reaches the summary and spec but NOT the item_key.** A slot that is broken is
  one problem however far down the path it broke; a phase in the key would mint a second
  open item when the same slot failed differently on the next run.
- **The `no row matched` branch (~:1357) is scoped OUT.** Row-locked-or-gone is a different
  case with its own lock arm above it and a different blast radius. Recorded as considered.
- **`gofmt -w` NOT run on `render_site_components_action.go`.** It is unformatted at HEAD
  from another lane's commit; formatting it would take their line as a same-file passenger.

## Accepted behaviour change

A slot whose store fails **and which has nothing stored to serve** now fails the step.
**garden-tools.uk's footer `rendered_html` is NULL**, so this lands on a live site at its
next build. Held to be `bugs_open/260`'s ruling applied consistently; flagged to the
council as the edit most wanting an argument rather than assumed benign.
