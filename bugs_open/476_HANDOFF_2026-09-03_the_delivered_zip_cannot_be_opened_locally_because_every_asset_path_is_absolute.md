# 476 — the delivered ZIP cannot be opened locally: every asset path is absolute, so a double-click gives an unstyled page

Filed 2026-09-03 by `site_delivery_and_editor`. Found while writing the customer instructions the
owner asked for (`bugs_open/475`), against the ZIP actually delivered in the `idea.uk` rehearsal.

## The measurement

Across the ten `.html` pages in the delivered ZIP:

| reference form | count |
|---|---|
| absolute — `href="/assets/…"`, `src="/tools/…"` | **87** |
| relative — `href="assets/…"`, `src="./…"` | **0** |

`index.html` alone asks for 1 stylesheet, 4 scripts and 1 image by absolute path.

## What that does to a customer

They unzip the folder and double-click `index.html`, which is the obvious thing to do and the thing
any instructions would tell them to do. The browser resolves `/assets/css/styles.css` against the
**root of their hard disk** — `file:///assets/css/styles.css` — not against the folder the file is
sitting in. That path does not exist.

So they get the text of the page with no styling, no images, and no working header, footer or news
section. Nothing is actually missing from the ZIP; it just looks like everything is.

**The likely reading is "this download is broken",** which is the same wrong conclusion
`bugs_open/475` produces by a different route, on the same delivery.

## How it was found, and why that matters more than the defect

The owner asked for a section explaining how to view the site from the unzipped folder. Writing it
took one grep. **The previous draft of that very section already claimed it worked** —

> *"Open `index.html` in any browser and the whole site works, offline, exactly as it does online."*

— written by me, into a document whose entire purpose is fixing a false claim about this ZIP. It was
never checked because it is the kind of sentence that sounds like a fact about HTML in general rather
than a claim about this artefact.

**Asking for the section is what found it.** The general form: a request to document something is a
request to check it, and documenting is cheap enough that it is one of the better ways to discover
what the estate cannot actually do.

## What it costs today, and what it would buy

Today the customer's route to seeing their own site is to put it online first, or to run a local web
server from a terminal. Both work, both are in the instructions, and neither is what somebody expects
of a folder of files they were told is theirs to keep.

Relative paths would make the ZIP self-contained: double-click, works, offline, for ever, with no
account and no host. For a deliverable sold as *"the files are yours"*, that is most of what "yours"
should mean.

## Why this is NOT a delivery-lane fix

The paths are emitted by the site framework and they are correct for the live site — a served site at
a domain root wants absolute paths, and they survive a page moving between directories, which relative
ones do not. Changing them affects every rendered page on every live site, not just what goes in a
ZIP.

**So the cheap-looking fix is the wrong one.** Rewriting paths inside the zipper would make the ZIP
work and put the ZIP's HTML permanently out of step with the pages we serve — two versions of the
same site differing in a way nothing checks. That is a worse defect than the one it fixes.

## Fix candidates, ordered by what closes the door

1. **Emit relative paths from the framework, for both the served site and the ZIP.** Closes it
   everywhere and makes served and delivered bytes identical again. Needs the rendering lane, needs
   care about pages at different depths (`/tools/x.html` referencing `../assets/`), and needs a
   fleet-wide re-render. Biggest, and the only one that makes the two artefacts agree.
2. **A `<base href="/">`-style shim, or a tiny bundled viewer.** A single-file HTML launcher in the
   ZIP that serves the folder locally. Self-contained, no framework change — but it is a second way
   to view the site, which is a thing to maintain and explain.
3. **Say so plainly in the instructions and give the two working routes** (upload it, or
   `python3 -m http.server`). **TAKEN for now** — `DRAFT_2026-09-03b_customer_instructions_copy.md`
   §1. It is honest and it costs nothing, and it does not pretend the deliverable is self-contained.
4. Weakest: rewrite paths in the zipper only. **Rejected**, for the reason above: it makes the
   delivered HTML differ from the served HTML with nothing comparing them.

## How to verify a real fix

Unzip a delivered ZIP into an empty directory, open `index.html` by double-click, and confirm the
page renders with styling and images and no console 404s. Today's negative control is the `idea.uk`
delivery: same ZIP, 87 absolute references, unstyled on open.

## Related

- `bugs_open/475` — the same delivery, the same "customer concludes the download is broken" outcome
  reached by a different route.
- `DRAFT_2026-09-03b_customer_instructions_copy.md` §1 — the honest wording, and the record of the
  false version it replaced.
