# Where we are — bugs sweep (plain prose, append-only)

## 2026-07-28

We pointed a session at the open-bugs folder and asked it to work through them, preferring
fixes that help the whole platform over ones that patch a single site.

Four bugs are properly finished and closed: a page assembly that reported success while
producing nothing, tool pages that were publishing their internal build instructions to
Google, a field that was documented and read by no code at all, and a publish script whose
main branch could never run. Three more are fixed and shipped but stay open because
something real is left over — in one case two duplicate news pages on a live site that no
code change can tidy up.

Three things we learned are worth more than the fixes.

**We had already decided the news URL question, and I re-opened it by mistake.** The
convention is written into the code and there is a site already doing it correctly. I even
quoted the file that says so, as evidence for something else, and still asked for a ruling.
Escalating feels like the careful thing to do, and it isn't free — it writes "undecided"
into the record of a decided question.

**One planned fix turned out to be impossible, and finding that out was the work.** We
wanted the system to spot a page that is secretly the news page under the wrong label. It
can't: on robot-hands there is a catalogue page and a news page that look *identical* to
any check we can write. Shipping the fix would have re-labelled the catalogue and broken it.

**The diagnosis tool is being lied to by its own code index.** It searched for a function,
was told "found nothing, and this is a real answer, not a gap", and gave up. The function
exists. The index is pinned to a snapshot from four days ago, about a thousand commits
behind. Until that is fixed, asking the diagnosis loop anything about recent code wastes
the run — which is exactly what happened to us.

I also broke something myself: deploying a new build killed a review that was running at
that moment, and then I misread a dead job as a slow one for over an hour. Both are written
up so the next person checks what is in flight before deploying.

## 2026-07-28, later — the news feeds were never news

The second session went to pick up the code-index bug from the list above and found
another thread already six commits deep into it — so it took the newest fleet-wide bug
instead: every "news feed" on every site was actually running a plain web search.

The system asks for news in the clearest possible way. A function whose own documentation
says "forces search type to news" sets the value, the message carries it, the receiving
service logs it. And then the value dies at the last doorway: the function that actually
calls the search providers had no slot for it. It was passed along faithfully through
three layers and dropped at the fourth. The struct designed to carry it — with a recency
control as well — had existed in the codebase the whole time, and nothing had ever used it.

You could see the consequence in what the feeds ingested: not news but timeless reference
pages, "best fonts of 2026" listicles, Reddit threads, and — the giveaway — marketing
pages for market forecasts dated 2034 and 2035. Nothing had a publication date at all.
Every health signal was green while every item was wrong. What makes this one worth
remembering: normally a dead setting at least looks dead. This one looked *better* than
alive — forced by a named function, promised in a comment, logged with the right value.
Every artefact a reader would check said "news". Only the function signature one hop
further on told the truth.

The fix widens that doorway so the value cannot be dropped again — the code will not
compile without it. Each provider now uses its genuine news mode; the one provider that
has no news mode now says so honestly and the request moves to the next, and if nobody
can serve news the request fails visibly instead of delivering mislabelled content.
Publication dates now arrive and are translated into the strict format our database
expects — without that last piece the fix would have looked done while the dates stayed
empty.

One expected side effect to know about: we have always had a rule "skip news older than
30 days" that never fired, because no item ever had a date. It now works. So some feeds
will ingest fewer items than before — that is the rule finally doing its job, not a
regression.

The change went through the review council (approved first round), is live on the search
service, and I watched it work: a test query went to the real news API and came back with
three dated news stories, and the honest-refusal path behaved exactly as designed. Two
loose ends keep the bug file open: an optional "only last week" recency control waits for
the next main-service deploy, and the final proof — real feed items with real dates from
the regular afternoon refresh — was still pending when this was written.

## 2026-07-28, afternoon — the news feeds now carry news, and the case is closed

The afternoon refresh settled it. The same feeds that this morning brought back font
listicles and marketing pages for 2034 delivered four actual stories — a July typeface
round-up, an open-source design system release from Meta, a current piece on email
markup — every one carrying a real publication date from the last three weeks. This site
had never ingested a single dated item before today.

Two things about that result were predicted in advance, which is worth saying because
predictions are how you tell a fix from a coincidence. First, the feeds brought back
four items, not the usual ten: we have always had a rule that discards news older than a
month, it had just never fired because nothing ever had a date — it woke up today and
did its job. Second, the recency control I added also went live in the main service
without me deploying anything: another of our parallel work streams shipped a build that
naturally carried my committed change, which is exactly how our
one-commit-per-finished-task discipline is supposed to pay off.

One sour note, not about this bug: our AI provider account hit its usage cap mid-
afternoon and locks us out until the 1st of August. The review council can't sit until
then — this fix got its approval in before the door shut; my other change today (a
tidy-up of how page data reaches templates, also now live) is committed and tested but
will carry its review receipt late, through no fault of its own. The cap is written up
separately and needs an owner decision.
