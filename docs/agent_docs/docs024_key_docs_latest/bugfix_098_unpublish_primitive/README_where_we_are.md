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

## 2026-08-03, evening — decided, and the page is down

You asked me to decide as I'd recommended, so I did.

**On the vocabulary question**, I took option B: the deletion verb stays on the git adapter,
but comes out of the list any workflow can reach. In practice that means retraction is only
possible through the one action that carries the safety checks, and a future workflow author
can't wire up a file deletion by writing configuration. That was the reviewers' actual
complaint and it costs almost nothing to concede. I've written down what it does *not*
settle — the general question about destructive verbs is still open for whoever adds the
next one.

**The two objections I agreed with are fixed**, and one of them turned out to be a real
hole rather than a theoretical one. The check for "is anything still linking to this page?"
was only reading rendered HTML. Links also live in stored structured content, and when I
widened it, the check immediately found a link it had been missing — a call-to-action on the
gripper cycle-time page. That's a link that would have been left pointing at a deleted page.
The reviewer who raised it was reasoning from a documented trap and was right.

**The twice-daily republishing is proven dead.** The evening news run fired on the fixed
build: the four live news pages were refreshed as normal, and the retired one was skipped.
Before the fix it took all five.

**And the page is retracted.** Dispatched through the platform rather than by hand:
200 → 404, gone from the repo, and I checked six neighbouring pages afterwards — all still
fine. The link checks ran on the live path and agreed with what I'd measured by hand.

Two things I want to flag rather than bury.

**First, I found a fault in my own work by reading the record afterwards.** The action
computes a full report — what it refused, what nav it retired, what got stranded — and the
platform then throws it away, because the step waits for a reply and the reply overwrites
the report. It survives only in logs, which don't survive a restart. For this page it
doesn't matter, since it refused nothing. But a retraction that *did* refuse a page would
today refuse it silently, which is the opposite of the point. It's written up as owed work.

**Second, this bug's own success criterion doesn't work.** It says "count the archived pages
that were once deployed; after the fix it should be zero". That number didn't move when the
page was retracted, because we deliberately don't clear the "was deployed" stamp. I actually
wrote "it's 13 again" into the file, ran the query out of habit, and it was 14. The bug's
test is an example of the very thing the bug is about — a timestamp that records history
being read as if it recorded current state.

**Where that leaves it:** the capability works and is proven. Thirteen pages still serve
frozen copies, and one more joined them today from another lane, so the backlog is growing
at roughly a page a day. I've left those alone as you asked. The bug stays open, because
archiving still doesn't retract *by itself* — someone has to run the tool.

---

2026-08-03, late — the "silent refusal" hole is closed in code, awaiting the next build.

Last time we found that when a retraction runs, everything it worked out — which pages
it refused to touch and why, what still links where — was being thrown away the moment
the git service replied, surviving only in pod logs that vanish on the next restart. So
a retraction that *declined* to remove a page would decline silently, and nobody would
ever know.

That's now fixed and committed. The full audit is kept in a place the reply can't
overwrite, and every refusal is also written to the fleet's error log, where the
monitoring already looks — written *before* we ask the git service to do anything, so
even a failed run leaves its record. The record also says how many of those rows were
actually written, so a lost write can't masquerade as a kept one. Four tests prove the
behaviour, including one that replays the exact overwrite that caused the loss.

It's committed but not running yet — a new build was already going out as I committed,
built from code just before mine, so the *next* build is the one that picks this up.

Still open, unchanged: nothing archives pages automatically (so nothing retracts them
automatically either), 11 old pages still serve frozen copies, and the decision on
whether to retract the rest — 10 of them on leopardess — is still yours to make.

---

2026-08-04, midday — everything we set out to fix is now fixed, reviewed, and running.

The new build went out this morning and I checked it the proper way, on the running
machines themselves: the retraction system's "keep the evidence" fix is live on both.
From now on, when a retraction refuses to touch a page, that refusal is written down in
two places that survive — nobody has to have been watching.

The two code-tidiness debts from the review are also done and approved. The reviewers
pushed back usefully: their "did you check you found ALL the copies?" question turned up
five more I'd missed, now fixed too.

What's left on this topic is genuinely your decision, not more code: whether to take down
the other eleven retired pages that still serve old copies (ten are ordinary content on
leopardess), whether archiving a page should automatically take it off the site in
future, and a ruling on the deeper plumbing question (RFC 012) — which another team's
outage this morning has made more pressing, not less.

---

2026-08-04, evening — the retired pages are down, one decision taken, one humbling find.

You approved taking down the remaining old pages, so I did: ten on leopardess (the
eleventh, the loan-calculator page, had already vanished on its own overnight). Checked
first that nothing on the live site still links to them — nothing does, and I proved the
check itself works before trusting it. All ten now return "not found"; the live pages
around them are untouched. One re-check remains: confirming they stay down past tonight's
scheduled refresh, which is the test that actually matters.

The delegated decision: archiving will NOT automatically take a page off the site. There
is no safe place to wire that in — it would mean building an automatic file-deleter
triggered by a hand-edited database flag, which is precisely the kind of unguarded power
this platform keeps having to take back. Instead the written procedure now says plainly:
archiving is step one, running the retraction is step two.

The humbling find: half of yesterday's "keep the evidence" fix turned out not to work in
real conditions — the platform throws that copy of the evidence away at an earlier point
than the one I guarded (and tested). The important half — refusals written where
monitoring reads — does work. The repair is small and queued; the deeper plumbing
question now has three documented faces and genuinely needs your ruling (RFC 012).

---

2026-08-05, morning — it held.

The overnight and morning refresh cycles have both run, and every page we took down is
still down; nothing tried to republish any of them, and the live site is untouched. That
was the test that mattered — the one that caught us out last week — and it passed
cleanly. The bug's page count is zero for the first time since it was filed. Left to do:
one small code repair (writing the retraction's full audit somewhere that provably
survives), and then closing the bug on your say-so. Your two standing decisions are in
yesterday's summary: the RFC 012 ruling, and — only when the next destructive verb
arrives — RFC 011's deferred question.

---

2026-08-06 — closed, on your word.

The bug is closed and moved to the closed pile. Everything it asked for exists and is
proven: pages can be taken off a live site through a guarded, audited path; the ones
that were wrongly serving are all down and stayed down through the scheduled refreshes;
and when the machinery refuses to remove something, that refusal — and the full account
of what it considered — now lands somewhere permanent that monitoring already reads,
proven by deliberately aiming it at a live page and watching it decline politely.

You also ruled on the plumbing question (RFC 012): option B, the database-backed
helper. That means the trick every affected piece of code has been reinventing — how to
keep your findings when the platform throws them away mid-flight — becomes one shared,
named tool that writes to the database, the only place proven to survive. Nobody is
assigned to build it yet; the ruling is recorded so whoever next needs it builds it
properly instead of improvising again.

Nothing else is open on this lane. The one question left deliberately unanswered (does
a destructive verb deserve its own vocabulary?) waits, as agreed, for the next
destructive verb to show up.
