# Where we are — bug 184 (markdown symbols showing up in page text)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-18 evening — picking the bug up properly

The bug: our AI content writers sometimes type markdown formatting symbols (like
`**bold**`, `# headings`, and `[link text](url)`) into fields that the site renders as
plain text, so visitors see the raw symbols instead of formatted text. It was filed
2026-08-03 on three rows; it has grown since.

What already works: the *detector* is live and finds these defects reliably. What has
never worked: the *repair*. The current repair asks an AI to rebuild the whole page —
and the rebuilding AI has the same habit, so it typed the markdown straight back in.
That path has succeeded once in twenty-nine tries, and the queue machinery has now
sensibly stopped feeding it work (a "success floor" another lane added on 08-17). So
today: 71 open defect items across 6 sites, new ones still arriving daily, and no
working way to fix them.

What I checked before starting: no other session is working this bug (two lanes
recently *contributed* evidence to it but both explicitly declined to claim it), and
the bug is still valid — I re-counted the open items and the live data tonight.

The direction I'm researching: stop using an AI to fix a mechanical problem. Removing
markdown symbols from a plain-text field is a deterministic string operation — a small
piece of ordinary code can do it perfectly every time, both as a *repair* (clean the
existing rows) and as a *guard* (clean anything a writer tries to save in future, so
the defect class dies rather than being chased). Also: the detector needs widening —
the commonest form live today is markdown links, which it doesn't yet look for.

## 2026-08-18 later — plan settled, half the code committed, waiting on a neighbour

The plan is written and submitted to the review council. The heart of it: a small piece
of ordinary code that deletes markdown symbols from plain text (it only ever removes
characters, so it cannot inject anything), used in three places — when the AI first
writes content, when an editor edits it, and during the "re-render" step that rebuilds
a page from its stored content. That last one is the repair: rerendering a page is
something the system already does thousands of times reliably, and with the cleaner in
place, rerendering a defective page fixes it. The broken AI-rewrites-the-page repair
path is abandoned, not patched.

Half the code is committed. The other half touches a file another session is editing at
the same time (they're fixing an unrelated bug about phone-number buttons). We messaged
each other and agreed an order: I committed the shared piece first so their commit can't
break the build, they commit next, then I finish. The database migrations that switch
the new behaviour on are written and committed but won't be applied until the new code
is actually running in the fleet.

One measurement worth keeping: the live defect today is mostly raw "#" headings and
[link](url) syntax on news pages — not the **asterisks** the bug was filed on — and the
detector didn't look for links at all until today's widening.
