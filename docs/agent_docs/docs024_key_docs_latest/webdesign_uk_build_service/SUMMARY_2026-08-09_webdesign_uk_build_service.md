# SUMMARY 2026-08-09 — webdesign.uk build service

Fourth in the series (08-04, 08-06, 08-08, this one). Written at the point the
shopfront was accepted for now and the lane turned to the functionality behind
it. Cold-start doc: `HANDOFF_2026-08-09_continue_here.md`.

## What we're trying to do

webdesign.uk sells complete websites to small and medium UK businesses: one
fixed price paid once, the whole site built and put live on the customer's own
domain. The shopfront must be built by the machinery it sells, hosted on our own
machine behind our own controls, and carry a chat box that opens a conversation
with a visitor rather than a form. When the owner is happy, the domain switches
over and the shop is open.

## Where we've come from

The first proper build was rejected on sight in early August: one page, no
styling, brittle copy. All four causes were traced and the site was rebuilt
through the framework on the 8th, arriving as five clean pages on the private
preview. The owner then reviewed it properly, and that review is what this
summary is really about, because two of his three complaints were things every
automated check we had said were fine.

## What we've done

The owner said the site looked hand-built, the home page had no picture, and the
copy promised unlimited free changes alongside an open-ended right to reject.

The first complaint was wrong but worth the hour it took to answer properly: the
site is framework-built and the trail proves it, down to the checking system
blocking the machinery's own drafts mid-build. What made it look hand-made was
the second complaint. The home page's main picture was a background image, and
its file had gone to the wrong repository back on the 4th — a background image
that 404s leaves no broken-image icon, just a bare dark block, so it reads as
"unstyled" to a person and as "fine" to a link checker that only looks at `href`
and `src`. Our verification now extracts `url(...)` too, and the picture is
regenerated and showing, along with the four smaller versions of the same
problem found on the way.

The third complaint took most of the day and taught us the most. The uncapped
promise was not a sentence in a page; it had soaked into every layer of the
machine's own notes about the site — the mission, the identity, the strategy,
the briefing — so every rewrite faithfully brought it back. Those were corrected
at the root, with the owner's original wording preserved in history. Then the
owner confirmed the numbers: **two rounds of revisions included, fourteen days
to review**. Getting those numbers onto the pages took three more rounds and a
wrong diagnosis of our own before we read the prompt the writer actually
receives and found the answer: the platform keeps a facts ledger and a writer's
brief as two separate things, and a fact added to the ledger never reaches the
writer unless it is also copied into the brief. Nothing warns about the
divergence. One paste into the brief and every page stated the terms first pass.

## Where we are now

Five pages serving on the private preview, built by the framework, with every
picture showing, zero violations of the owner's eighteen content rules anywhere
a visitor reads, clean titles, and the confirmed terms stated plainly wherever
changes and refunds come up. The owner has accepted it for now. Nothing public
has changed: the domain still forwards to webdesign.co.uk and the preview is the
only window in.

Two platform faults found here are documented with evidence for whoever owns
them: the image-deploying agent is down ("storage client not available"), which
is why the site still has no favicon; and the build queue does not drive itself,
so every pipeline stage on this site is dispatched by hand.

## Where we're going

The chat service — Phase 4 of the hosting plan, and the one piece we write by
hand rather than generate. Then the input box on the page that talks to it,
which is what the site's "get started" buttons are waiting for. Then, on the
owner's approval, cutover: one DNS change and two deleted redirect rules.

### The todo list, in order

**Phase 4 — the chat service** (the current job)
1. Write the service: Go, stdlib-first, `127.0.0.1:8081` on the box, versioned
   in this lane's `box/` directory. Sibling of site-engine, not an extension.
   Endpoints `POST /api/chat` and `GET /health`. **The box has no Go toolchain
   — cross-compile and ship the binary; document the deploy step in the RUNBOOK.**
2. Build all five controls into the first commit, not after: per-IP limiting
   keyed on `CF-Connecting-IP` (through a tunnel anything else is one global
   bucket), a hard turn cap per conversation, a per-day spend ceiling that fails
   closed to the contact details, a request log with tokens and cost, and
   transcripts stored as rows — the demand signal this phase exists to collect.
3. Prove the per-IP key from two networks (`count(DISTINCT ip) > 1`). One
   machine cannot tell a working key from a constant.
4. Build and test the whole thing against a fake provider first, so the owner's
   key is needed only at the end. Model when it is: `claude-haiku-4-5`.
5. **Blocked on the owner: the scoped Anthropic key.**

**Phase 5 — the input box**
6. Add it as a pinned section in the site plan, posting same-origin to
   `/api/chat`, with the estate's generation-time guards (external loader file,
   no inline script). Never ship it before step 1 exists.
7. Landing it resolves the nine parked review items about buttons with nowhere
   to point — that is their intended destination.

**Phase 6 — cutover** (owner-gated)
8. Owner reviews the preview and approves. Then point the domain at the tunnel
   and delete the two redirect rules. No Worker step remains.

**Loose ends, none blocking**
9. Delete the stale `gqls/sites/webdesign.uk/` directory — orphaned files from
   the pre-flip build that nothing serves.
10. The favicon, once the image-deploying agent is fixed platform-side.
11. Page search descriptions are still empty.
12. Owner still owes: the correction-fee number, and written terms before live
    Stripe.
