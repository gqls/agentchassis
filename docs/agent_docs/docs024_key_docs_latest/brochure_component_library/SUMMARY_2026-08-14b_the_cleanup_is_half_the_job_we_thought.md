# SUMMARY — the cleanup is about half the job we thought it was (2026-08-14, later)

*Duplicate-pages front of the brochure lane. Follows
`SUMMARY_2026-08-14_the_prevention_holds_and_the_cleanup_is_a_content_job.md`, written this
afternoon. It exists because that read-out put a genuine choice to you — "is this cleanup still
worth its price?" — on a price that has since turned out to be roughly half what I quoted.*

## What we're trying to do

Some of our sites carry the same page twice: a plain version at an address like
`/llm-cost-calculator.html`, and a second version under `/tools/`, both live, both quietly
maintained by the machine. It is untidy for a visitor, it splits whatever search traffic the
page earns, and it means our own records disagree with the website.

Two halves: **stop new duplicates appearing**, and **clean up the seven pairs that already
exist.** The first half is done. This read-out is about the second half getting cheaper.

## Where we've come from

The prevention is built, reviewed, live and holding — unchanged since this afternoon. It still
has not had to refuse anything, because nothing has asked it to, and that remains the expected
reading rather than a disappointment.

The cleanup is where the movement has been. You decided all seven pairs, then reversed two of
them when I found that a step I had relied on — pointing an old address at a new one — does not
exist in our system at all. We then executed the first pair and the platform stopped us,
correctly: the page was still linked from the site navigation, from the footer that appears on
every page, and from an article. Nothing was deleted.

From that one example I drew a conclusion and gave it to you as the basis for a decision: that
the first pair had been chosen as the *easiest* of the seven, that it still needed two pieces of
writing, and so **"there is no reason to expect the other six to be tidier."** The cleanup, I
said, was no longer a tidy-up but a content project across four sites.

## What we've done

**I measured the other six instead of guessing, and the guess was wrong.** Three of the six need
no writing at all. Nothing links to them; they can be retired cleanly.

**It cost nothing and changed nothing.** The plan I had written for finding this out was to
retire each of the six in turn, read the platform's refusal to see what linked to it, then put
each one back — six changes to live pages purely to ask a question. Before doing that I went and
read how the refusal is actually assembled, and found it is built from three straightforward
database queries that never look at the page being retired, only at the pages doing the linking.
So I ran those queries directly. Everything today was read-only. Not one page was touched.

**I checked the measurement against a known answer before trusting it.** Almost everything this
kind of check returns is *nothing*, and nothing is the comfortable answer that nobody
re-examines — it is exactly the trap that produced the redirect mistake I had to come back to you
about yesterday. So I ran the first pair through it, because that is the one page the platform
has already given us a real refusal for, and required my version to produce the same answer. It
reproduced all three links exactly. I also confirmed each of the seven pages had genuinely been
found by the query before accepting any of its blanks.

**A saving fell out of it that I would not have seen otherwise.** One of the two
fundamentally.ai guides is itself one of the three things linking to the other one. Retire them
in the right order and the second one's repair list drops from three items to two on its own.
I have written the order down and passed it to the workstream that owns that site.

## Where we are now

**The remaining work splits cleanly in two.** Of the six pairs still to do, three need no
content work whatsoever, and three need between two and four pieces of copy repaired — links
sitting in article text and in site footers. Across the whole job that is about nine pieces of
writing, not the uniform content project I described this afternoon, and it is concentrated in
half the pairs rather than spread across all of them.

**The three clean ones are not all equally cheap, and I want to be straight about that.** One of
them — a fundamentally.ai guide — is genuinely cheap: decided, nothing links to it, no other
preparation needed. The other two are on robot-hands, where the site's page plan lists both
versions, so the plan has to be edited first or the retired page simply comes back. And one of
those two is the pair where you asked for the two versions to be merged rather than one thrown
away; that merge is around 1,700 words and remains the single largest piece of work in the
cleanup. Its cost was never the links.

**Your reversal of the two fundamentally.ai guides paid off twice.** You made it to protect the
older, better-indexed addresses, since without a redirect retiring one just produces a "not
found". It also happens to retire the side that almost nothing links to — which is why one of
those two is now the cheapest pair remaining. The decision was right for a second reason neither
of us knew at the time.

**The first pair is exactly where it was.** Its older version is marked retired but still
serving, navigation and footer still pointing at a working page, nothing broken for a visitor,
and one command puts it back. It still needs its footer and one article repaired before it can
be finished.

**One caution on all of the above.** What I measured is what the platform will refuse, which is
not quite the same as what links to a page anywhere in the world. The check is deliberately
precise about the form of link it recognises and will miss unusual ones — that is the platform's
own stated limitation, not something I introduced, and it means a blank is a strong signal
rather than a guarantee.

## Where we're going

**The choice I put to you this afternoon still stands, but at a materially lower price.** I
asked whether a cleanup that had become a content project across four sites was still worth
doing. It is roughly half that: three pairs need no writing, and the writing the others need is
nine pieces of copy, all of it repairs to links rather than new prose.

If we proceed, the sensible order is now clear. Start with the fundamentally.ai guide that needs
nothing — it is decided, clean and requires no input from you. That site's execution belongs to
another workstream on the same site, so I have handed them the findings and the ordering rather
than acting behind them. The robot-hands pairs follow, plan edit first. The pairs needing copy
repairs go through the content pipeline, which writes them — not a person typing into the
database.

Two threads stay open and neither has moved. The prevention will not count as finished until it
has actually refused something once, and the cleanup is still the natural way to make that
happen. And the finetuning site's pair stays decided but untouched, because that site has a
separate known fault that has to be fixed first.

---

*Correction carried forward, since summaries are never rewritten: this afternoon's read-out
stated "there is no reason to expect the other six to be tidier". That was an inference from a
single example and it is false — three of the six are tidier. The earlier file stands as
written; this one supersedes it on that point.*
