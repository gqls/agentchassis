# HANDOFF A (2026-07-30) — the share card should carry the debate, not just the verdict

**Start a fresh thread on this.** Self-contained; you do not need the rest of the
lane's history to begin. Owner-raised, 2026-07-30.

## The owner's words

> "I think perhaps the card should have the whole debate or there might be two
> cards or the card links through to a record of the full debate."

Three options, and he has not chosen between them. **The choice is the first
deliverable — do not just build one.** Put the trade-offs to him with a mock of
each; he decides quickly when there is something to look at.

## Why this matters now, and why it is not cosmetic

The distribution experiment (owner's ruling of 2026-07-29, "3 leading to 2") is
the whole current strategy for vonc.com: HE posts the share card and the daily
provocation where people already argue, and real behaviour chooses between the
examination thesis and the arena thesis. **The card is the travelling artefact.**
It is the only thing a stranger sees before deciding whether to click.

A verdict-only card says "someone scored 7/10 on an argument you cannot read".
That is a claim about a stranger. A card that carries, or leads to, the actual
argument is a sample of the product. **As of 2026-07-30 there is no engineering
blocker left:** og:image is fixed and verified live (`og-card.png`, HTTP 200,
`image/png`, 1200×630 exactly), so links already render with a face.

## The three options, as they actually differ

1. **Whole debate on the card.** One image, the argument on it. Hard limit: at
   1200×630 you get very few words before it is unreadable in a timeline. Likely
   forces a "best exchange" excerpt, which is an editorial choice a machine will
   make badly. Cheapest to ship, most likely to disappoint.
2. **Two cards.** One verdict card (exists), one debate card. Doubles the render
   path and the choice of *which* to post falls to the owner every time. Ask him
   whether he wants that decision each time he posts — that is a workflow cost,
   not a rendering cost.
3. **Card links to a record of the full debate.** The card stays a hook; a real
   URL holds the transcript. Most work, and the only option that survives someone
   actually being interested. **It also creates something the site does not have:
   a permanent, linkable artefact per round.** Note that the opinion ledger
   (shipped 2026-07-29, live: `gi-ledger` markers on the served page) is already a
   *private, localStorage* record of what you argued. Option 3 is its public
   sibling and should be designed with it, not beside it — ask whether the ledger
   entry should link to the public record.

## THE RAIL — non-negotiable, and option 3 is where it bites

From §5 of `HANDOFF_2026-07-29_continue_here.md`, and it is the owner's own rule:

> Nothing on vonc.com claims a number that is not true by construction; no
> control changes state except as the consequence of a real API response … every
> string on the card is a fact of that round.

So for option 3: a public debate record may contain **only** what a real
`/defend` round actually produced. No reconstructed transcripts, no summarised
paraphrase presented as what was said, no placeholder rounds to make the page
look populated. If there is no round, there is no record.

Second rail, also the owner's: **communal/aggregate views come only after
participants exist.** A "most debated positions" list on an empty site is a
fabricated crowd. Option 3 must not grow one by default.

## Where the pieces already are

- **Card generation is live and real** — a 60KB 1200×630 PNG pulled off the click
  in-harness (corr `ba1666a7`, commit `dce85ccd8`). Read that harness before
  designing; it tells you what the render path can and cannot do today.
- **The card button lives inside the verdict step**, hidden until `/defend`
  returns. That gating is deliberate and is part of the rail.
- Round data is stored server-side: `store.CreateRound` in
  `internal/tools-api/handlers/round.go`. **A public record needs a stable
  per-round identifier and a considered decision about what is publishable** —
  rounds currently carry a visitor key (`client_ip_hash`), so check what else is
  in the row before exposing anything.
- Sources, builders and harnesses: `p4_sources/` (`build_g*.py`, `drive_g*.py`).

## Landmines

- **The page serves ONLY at `/tools/gauntlet/index.html`** — both bare variants
  404. Do not invent a URL; read `pages.url`. (I lost a check to exactly this on
  2026-07-30, reading a B2 `NoSuchKey` JSON body as page content.)
- **`scripts/rerender_gauntlet_vonc.sh` is the hardened rerender** for this page.
  Prefer it over `republish_gauntlet_js.sh`, which still carries the racy kcat
  stdin form that publishes nothing at exit 0.
- **Write BOTH `content_data` and `rendered_html`** on any component edit, each
  guarded on the `updated_at` you just read. An assemble-only rerender reads the
  second; a regeneration reads the first. Writing one lets the other restore the
  old text.
- **Verify on the served page, with a control string you did not change.** A grep
  returning zero cannot distinguish "fixed" from "blanked".

## Definition of done

The owner has chosen an option; it is built; and a card posted from a real round
leads a stranger to something true about that round. Verify by rendering — the
page fetches client-side, so `curl` cannot see any of it.
