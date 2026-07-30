# HANDOFF C (2026-07-30) — the home page paints the provocation the gauntlet page is built to hide

**Start a fresh thread on this.** Small, sharp, and it silently defeats a feature
that took a day to build. Owner-raised 2026-07-30: *"while the provocation is
hidden on the gauntlet page it is visible on the home page."* **Confirmed by
rendering.**

## The measurement

`scripts/provocation_visibility.py` (committed with this handoff) loads each page
in Chromium, waits for the client-side fetch to settle, and tests whether today's
headline and body appear in `document.body.innerText` — i.e. **painted**, not
merely present in the DOM:

| page | headline painted | body painted | visible chars |
|---|---|---|---|
| **`/` (home)** | **YES** | **YES** | 4,070 |
| `/tools/gauntlet/index.html` | no | no | 906 |
| `/provocations/index.html` | no | no | 1,293 |

On home the body renders in `<p class="pc-body">` at 669×230.

**Neither page contains the text in its served HTML** — both fetch
`/data/provocations.json` client-side — so a `curl` grep reports "absent" on every
page including the one that shows it. That is why this went unnoticed: every
HTML-level check says the provocation is nowhere.

## Why it matters

The sealed reveal is deliberate engineering (131-C, live: 22-check harness + 16
live checks, corr `824c7f1c`, commit `c2969cbff`). Its design, from §5 of the
2026-07-29 handoff:

> The page opens sealed behind one door; pressing Enter the Gauntlet starts a real
> round and the engine's answer reveals the question … only `/round` 200 or a live
> stored-round resume removes `gi-sealed`.

The point is that you commit to arguing **before** you know what you are arguing
about. **A visitor arriving via the home page — the normal path — has already read
the whole provocation before they reach the door.** The seal still works
mechanically and means nothing experientially.

## The decision, because this is a product question not a bug

Do not just delete the home-page block. Ask the owner which he wants:

1. **Tease, don't tell.** Home shows the eyebrow and a hook but not the body —
   enough to make someone click, not enough to pre-empt the seal. Preserves the
   home page's job (it is the page that has to sell the thing) and restores the
   seal.
2. **Home is the front door and the seal is for returning visitors.** Accept the
   leak, and drop the seal's claim to be a reveal. Cheapest, and it makes 131-C's
   whole build pointless — say so plainly if proposing it.
3. **Home shows YESTERDAY's provocation** (or an archive one) as the sample, and
   today's stays sealed everywhere. Honest, gives the archive a purpose — and it
   depends on **HANDOFF B**, because today the archive stops at 5 Jul and today's
   provocation never joins it.

Option 1 is the smallest change that keeps both features intact. Option 3 is
better product but is blocked on B.

## Where the leak is

The home page's provocation section is a page component on `index`
(`/index.html`), site `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`. The census found the
relevant slots as `index/provocation-card@2` and `index/lobby-grid@5`.

**Note while you are in there:** both of those sections' `content_data` also
contain what looks like page chrome — the normalised text of each includes
`"2026 vonc@contactforsales.com vonc.com 2026-07-25T09:30:41Z home how it works
provocations"`, identical between them. That is a footer/nav string sitting inside
a content component. It may be harmless render context or it may be an ingest
artefact; **it is not part of this handoff's fix, but do not "tidy" it without
understanding where it came from.**

## Landmines

- **Render, don't grep.** Covered above: the HTML contains none of this text.
  `scripts/provocation_visibility.py` is the instrument; it needs playwright,
  which is not installed system-wide (`python3 -m venv` + `pip install playwright`
  in a scratchpad).
- **An `<em>` in the headline splits the text node.** My first attempt matched on
  XPath `contains(text(), …)` and reported the headline "NOT IN DOM" on the page
  that was painting it. Use `innerText` of a container, not `text()` of a node.
- **Read `pages.url`; do not construct URLs.** `/about/index.html` 404s with a B2
  `NoSuchKey` JSON body that reads as 286 characters of page content — the real
  path is `/about.html`. Print `%{http_code}`.
- **Write BOTH `content_data` and `rendered_html`**, guarded on the `updated_at`
  you read, then rerender and verify on the served page with a control string you
  did not touch.

## Definition of done

The owner has chosen; a first-time visitor arriving at `/` cannot read today's
provocation body before entering the Gauntlet (or the seal has been explicitly
retired as a feature); verified by rendering both pages, not by reading HTML.
