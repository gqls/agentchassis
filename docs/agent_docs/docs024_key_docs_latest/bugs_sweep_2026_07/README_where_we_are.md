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
