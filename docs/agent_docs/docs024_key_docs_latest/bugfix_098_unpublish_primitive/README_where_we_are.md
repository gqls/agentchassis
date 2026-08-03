# Where we are — the unpublish primitive (bugs_open/098)

Plain-prose log, append-only, newest at the bottom. Owner's document.

---

## 2026-08-02, evening — what this is and why it was picked

Picked `bugs_open/098` off the open-bugs list after checking it wasn't already being
worked. That check matters more than it sounds: about thirty Claude sessions share this
one working tree, and the usual ownership script reads git commits, so it is blind to a
session that is halfway through a fix and hasn't committed. Grepping the live session
transcripts is what actually answers "is someone on this".

The bug: when we archive a page, the platform stops re-rendering it and stops listing it,
but **nothing removes the file that was already published**. So the page carries on being
served, frozen at whatever it last said. The live example is a "learning centre" index on
robot-hands.com whose article list was last written on 3 July and links to an article
that has since been archived — so a visitor gets a working-looking page full of dead
links. Thirteen pages fleet-wide are in this state.

The interesting part is *why* nobody had fixed it. It wasn't neglect. A different bug
(125) had hit the same wall from the other side — it published a page to a wrong path and
couldn't remove the duplicate; you deleted that one by hand. Its review asked for the
shared gap to be written down rather than quietly absorbed, and it was written down here.
The gap in one sentence: **the platform can publish a page but has no way to unpublish
one.** The only deletion the git adapter had was "delete the whole repository", which
returns "not yet implemented".

## 2026-08-02, evening — the finding that made the job small

The bug file's most ambitious fix option was "make deploys reconciling instead of
additive", flagged as a big piece of work needing its own survey. Before designing
anything I read the deploy workflow, and it turns out **half of it already exists**: the
sync from our repo to the file host already runs with `--delete` and purges the CDN. So
the far end already removes files that leave the repo. Nothing downstream was missing —
only the ability to remove the file. That turned a pipeline redesign into one small,
well-understood capability.

I built it so that a deletion is *a kind of commit* rather than a separate operation.
That sounds like a detail but it's the whole design: it means a deletion automatically
inherits the retry logic that stops two sites' deploys clobbering each other, the
path-naming rules, and — usefully — it means "move a page" becomes one atomic commit
instead of a delete and a write that can half-fail. A separate deletion path would have
had none of that.

Proved it end to end on a scratch file in a corner of the repo that the deploy workflow
ignores, so nothing real was touched: wrote a file, deleted it, then deleted it again to
check that repeating a repair is safe rather than an error.

## 2026-08-03, morning — I stopped before deleting anything, and I'm glad I did

You approved retracting the one page. Before doing it I re-checked the page, and found
something that changes the picture.

**The page isn't frozen. It's being republished twice a day.** Every day at about 08:05
and 20:10 something re-renders it and commits it back. The bug file says archived pages
stop being re-rendered; that was true when it was written on 26 July and stopped being
true on the 31st.

The cause is a query that picks which pages need refreshing when news arrives. It asks
"has this page been deployed?" and never asks "does the platform still want this page?".
Those are two different columns, and archiving only sets one of them. So the retired page
keeps getting picked, for ever.

**This is why it mattered to check.** If I'd deleted the file and verified with a quick
fetch, I'd have seen my 404, reported the bug fixed, and it would have quietly come back
that evening. The bug would have been closed and still broken — which is a worse outcome
than not fixing it, because the next person would trust the closure.

Fleet-wide, exactly one page is affected: this one. So it's a small fix, but it had to
come first.

## 2026-08-03, morning — the review pushed back, and it was right to

I put the code through the review council. It came back "revise", with three of the
reviewers independently making the same point: the retraction works out which file to
delete using one function, and there's a known trap in this codebase where two
near-identically-named functions decide a page's file path and disagree. If I'd used the
wrong one, the deletion would quietly delete nothing while the page carried on serving —
the exact bug I was claiming to fix, just inverted.

That trap was fixed a few days ago and the functions now agree, so I *thought* I was
fine. But the reviewers were right that I'd asserted it rather than shown it. So I
checked it properly: ran the real function over all thirteen pages and looked for the
file in each site's actual repository. Eleven of twelve found their file at exactly the
predicted path; the twelfth is genuinely already gone (it's the dead link the frozen page
points at); the thirteenth lives in a different repository and is correctly absent there.
No mismatches.

Worth saying plainly: the review cost a round trip and it improved the work. The thing it
objected to was a real risk, and "I read the code and it looks fine" was not good enough
evidence for a change that deletes live files.

## 2026-08-03, morning — your point about links

You said that when we retract a page we should look across the whole site for links
pointing *in*, nav entries, and pages that only the retracted page linked to.

That is the same thing this bug found from the other end. Its own notes record a case on
another site where a page was archived and its file correctly deleted — and then every
page on the site advertised the resulting 404 from its own footer, because the navigation
entry was still live. So deleting the file alone just swaps one broken state for another.

I've built that in, with the three cases handled deliberately differently:

- **A link in body copy or in the header/footer blocks the retraction outright** and names
  what's linking. Rewriting page copy is an editorial decision, not something a cleanup
  job should do on its own. Refusing also means "we created a dead link by retracting
  something" simply can't happen, rather than being something we detect afterwards.
- **A navigation entry is retired automatically**, just before the file goes. A nav entry
  is a pointer, not writing — there's no judgement call in removing it, and leaving it is
  precisely the failure above.
- **A page that loses its last link is reported, not touched.** What to do about a page
  nothing links to is a decision about the shape of the site, and we already have a
  standing check that owns that question.

For the page you approved, I ran all of this: nothing links to it, no nav entry points at
it, and nothing becomes unreachable when it goes. It's a clean retraction.

One correction worth recording, because it nearly went the other way: my *first* attempt
at the "what becomes unreachable" check looked in only one table and reported that ten of
sixteen linked pages would be stranded — which would have meant retracting one page broke
most of the site. That was wrong. Most links live in the navigation and the shared
header/footer, which my query couldn't see. Once I looked in all three places, the answer
was "nothing is stranded". A check that only looks in one place gives a confident answer
about that place, not about the site.

## 2026-08-03, late morning — where it stands

The deletion capability is built, reviewed once, revised, and resubmitted. It's live on
the git service and proven working on a scratch file. The twice-a-day republishing is
fixed and built into a new chassis image, waiting to be rolled — I'm holding the roll
until the review finishes, because rolling the chassis kills any review that's mid-flight.

Still to do: roll the chassis, then retract the page, then check it's *still* gone after
the next news refresh — that second check being the one that actually tests anything.
