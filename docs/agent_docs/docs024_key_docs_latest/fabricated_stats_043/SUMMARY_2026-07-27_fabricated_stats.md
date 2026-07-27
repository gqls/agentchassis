# Summary — stopping the sites publishing numbers that aren't true

*2026-07-27. Written for the owner to read aloud. First summary in this lane's series;
current state only — the chronology lives in `README_where_we_are.md` and `NOTES`.*

---

## What we're trying to do

Our sites are written by machines, and machines are fluent. A model asked to write a
confident page about a business will happily produce "2,400+ models indexed", "38% average
win rate", "thousands of concurrent instances" — because those sound like the sentences that
belong there, not because anything was counted. The numbers are the most persuasive thing on
the page and the least likely to be true.

So the goal is narrow and it is not "make the writing better". It is: **a number may appear
on one of our sites only if something can prove it.** Everything else follows from that — a
register of what each site is allowed to state with the query that proves each figure, a
rail that tells the writer what it may never say, and a checker that reads what actually got
published and complains when it doesn't match.

## Where we've come from

We found the machines inventing statistics across several sites and fixed it in four ways at
once: made the statistic fields optional so a page can honestly show nothing, stripped the
example figures out of the prompts that were seeding the inventions, built a checker that
audits every published number against the site's register, and added a lint that catches
nonsense units like "2,400+%". All of that is live and was proven end-to-end on a real
rebuild.

Two things were left owing, and both turned out to matter more than they looked.

The first was a gap the review council kept objecting to: we had built a general-purpose
checker and then wired it into exactly one place — the moment a page is freshly written. But
pages also get *re-rendered*, which takes the figures we already stored and republishes them
without a writer involved. Nothing checked those. A made-up number, once stored, would
reprint itself for ever.

The second was your request for evidence registers on all the remaining sites.

## What we've done

**The gap is closed.** The checker now also reads stored content, so the re-render route is
covered, along with hand edits and every page that predates the checker. It is live. We
deliberately did *not* bolt the checker onto the re-render path itself: that path is
intentionally simple, and giving it a new way to fail on content it did not write is the
exact mistake that made a page unbuildable last week.

**The registers turned out to be mostly the wrong work, and the survey is the finding.**
Before building anything, we ran the real checker across the sites with an empty register,
which makes every business-shaped number visible. Gaswholesalers has none at all. Dartsonline
has one, a returns window. Webdesign.co.uk — our biggest site — produced fifteen, and every
single one is a false alarm: they are worked examples inside tutorial pages ("100 concurrent
users", "random 502 Bad Gateway errors"). Building a register there would have created
fifteen review items about correct writing and taught whoever reads that queue to ignore it.
The cause is that the checker cannot tell what *kind* of page it is reading, even though the
database records it. That is now filed as a bug with the fix.

**One site did need it, and it paid off immediately.** Vonc.com's figures are real claims and
checkable, so we registered them — eight archetypes, three tools, two guides, eighteen pages
— each taken from the database rather than from the site's own copy, because the copy is the
thing that might be wrong. Before, the checker had fourteen complaints and no way to judge
any of them. After: twelve confirmed correct, and two real. The about page claimed three
archetypes and eight tools; the rest of the site and the database say eight and three. They
had been transposed. Corrected.

**Your two rulings are done.** The unevidenced "~80% reduction in quote preparation time" is
gone from finetuning.uk — and, because deleting a figure doesn't stop a machine reinventing
it, that site now has its first register, with a rail forbidding quantified client outcomes
until you attest one.

And ai-agent-orchestration.com is on the real agent figure. This one was more interesting
than expected. The site said "70+ agents" and "30+ agent types" while the live count is 175
and 174. Both statements were *technically true* — they're lower bounds — which is precisely
why no checker had ever objected: it is possible to mislead by a factor of five without ever
saying anything false. The copy now says 170+ rather than 175, and that choice is deliberate:
the count moved from 176 to 175 *during this session*, and a frozen exact number in a
headline would make the page fail its own check the first time an agent is retired. The exact
figure lives in the register, where a query keeps it current.

Doing that turned up something worse. The site was publishing "thousands of concurrent
instances" in four sentences — a claim its **own rules already forbade**. It survived because
the site had seven proven facts and *zero* banned claims: the rule telling the writer what
never to say, and the list telling the checker what to catch, are two separate hand-written
lists, and only the first had been filled in. So the prevention worked and the detection did
not exist. Fixed for that site; the general problem is recorded.

**And I made a real mistake, which you should know about.** While replacing "70+" with
"170+", my find-and-replace chain fed on its own output — "170+ agents" contains "70+ agent"
— and for about twenty minutes the site claimed **1170+ agents**, roughly seven times the
truth, produced by a fix for understating it. My own check passed, because I had written the
check from the same assumption as the change. What caught it was running the site's own
banned-claims list against the corrected copy — an independent judge that knew nothing about
my intentions. It is fixed, kept as a separate numbered file so the error stays on the
record, and written up.

## Where we are now

The checker covers both routes and is live. Vonc, finetuning and ai-agent-orchestration have
had their claims corrected in storage, and five pages are queued to re-render so those
corrections actually reach the public sites — until that runs, the live pages still show the
old figures, and I would rather say that plainly than call the job done. Nothing is broken
and nothing is half-finished.

The review council has approved five of its ten seats and still asks for two things: that we
attach the queries behind our claims rather than just their results, and a couple of genuine
design improvements. One seat's response failed to parse, so the last verdict was decided a
seat short.

## Where we're going

Three things need you, and none can be looked up.

The eight departments question. Ai-agent-orchestration states it has eight departments as its
own internal structure; leopardess has "eight departments" banned as an invented fabrication.
There is no notion of a department anywhere in the database. Both sites cannot be right, and
which is right is a decision about what the business says it is.

Whether webdesign.co.uk's tutorial pages should be exempt from the number-checking, which is
the practical form of the bug we filed.

And the about pages. Your separate about-page thread settled what those pages should carry —
the quiet "available to acquire" line, the ad-space line, the built-by line — and finetuning
is the live pilot. That thread decides what goes *on* an about page; this one decides what may
be *claimed* there. They have now met: finetuning's about page is the one carrying the
commercial block, and it is also the site I just gave its first register. Worth doing
together rather than separately, because an about page is where a site is most tempted to
quantify itself.

Everything else — the council round, the remaining checks — is ours to finish and doesn't
need you.

---

> **ADDENDUM, same day, written after the above.** Two findings arrived while publishing the
> corrections, and both change what "where we are" means. Recorded here rather than in a second
> summary, because neither is a new milestone — they are the same milestone, understood better.
>
> **1. A site's stored description is a live source, not a memo.** We had the remaining spec
> work filed as tidy-up: old numbers sitting in site descriptions that would mislead the writer
> next time. It is worse. Those descriptions are *resolved into the page* on every re-render, so
> a stale number in one is pushed back onto the published page **over the top of any repair**.
> Proven the hard way: I corrected ai-agent-orchestration's page, the checks passed, the page
> re-rendered, and the old figure was back within fifteen minutes — because the leadership
> biography is pulled from the site description, and I had fixed the page but not the source.
> The order that works is description, then page, then re-render. This probably also explains a
> recurring thing we have blamed on caching or failed deploys: a fix that "did not stick"
> because something rebuilt the page correctly from a source nobody updated.
>
> **2. Two of that site's copy templates claimed client work we have never done.** Sweeping for
> the old agent figure across every part of the site's configuration — not just the parts I had
> already looked at — turned up its call-to-action template saying *"We have shipped 70+ agent
> systems"*, and its writing-standards template offering, as the example to imitate, *"runs 70+
> agents in production today across financial services and logistics environments"*. We run
> fifteen sites and every one is our own; none is in financial services or logistics. So that is
> not an exaggerated number, it is a claim about a business we do not have — and it sat in
> *templates*, which regenerate it by design. Both now state the true, registered claim instead.
> **This one is copy rather than a defect, so it wants your eye**: if the intended claim really
> was about client delivery, the honest version needs a figure you can attest, and none exists.
>
> Everything else in the summary above stands. The corrections are published and verified
> against the live pages — a complete scan of all 626 components on registered sites reports no
> banned claim published anywhere.
