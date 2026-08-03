# Where we are — undeliverable replies, and what a crawl result is for (bugs_open/158)

*Plain-prose log, append-only, newest at the bottom.*

---

**2026-08-03.**

I picked up `bugs_open/158`. It is really four separate findings in one ticket, left
behind deliberately when `133` was closed, and its own author ordered them for
whoever came next. I re-checked all four before touching anything, and two of them
had moved.

**One was already done.** The dead test double it complains about was completed on
31 July, by someone fixing unrelated compile errors. Nothing to do.

**One was undercounted, and that matters more than it sounds.** The main finding is
that when a reply cannot be delivered, most of our services just write a line to
their own log and carry on — while the thing that asked the question sits waiting
until it times out, with nothing to say why. The ticket said the good behaviour was
in place at 2 sites out of 9. It counted by searching for the phrase "Failed to
produce". Four more places do the same wrong thing under different wording, and one
of them does not log anything at all, so no phrase search could ever have found it.
The real figure is **2 of 13** — and the four extra ones are not in the outer
adapters, they are in the chassis's own plumbing, which every agent inherits. So the
change being contemplated is bigger than the ticket described.

**How bad is it in practice? Much less bad than it reads.** The failure only bites
when a reply is too big for the message bus, so I measured how big our replies
actually get. The bus refuses anything over about a megabyte. The largest thing this
fleet has ever produced, across 47,577 recorded calls, is **48 kilobytes** — about
five per cent of the limit, and not one call has ever exceeded even half the limit.
The one service that genuinely sends whole documents is the web scraper, and the web
scraper is precisely the one already fixed. So the remaining eleven are a
consistency problem, not an active fault.

I should be straight about the limit of that: it measures the *size* failure. If the
message bus is simply unavailable for a moment, the same eleven places swallow that
too, and no amount of measuring reply sizes tells you how often that happens.

**What I shipped.** I could not fix the eleven — a council ruling says widening that
change is a design decision that needs your sign-off first, because it changes how
four services behave when they fail, and other people depend on that. What I could
do without changing anyone's behaviour is make it impossible to add a twelfth
quietly: there is now an automatic check that flags this pattern whenever anyone
commits code containing it. It catches all four shapes the mistake takes, including
the one where the error is never even looked at.

Two things about building that check are worth telling you, because they are the
same lesson twice. First, my first version passed every test I gave it — it fired on
all four known-bad files and stayed quiet on the two good ones — and it was still
wrong, because I had only ever shown it files that *should* fire. Run against the
whole codebase it flagged three perfectly correct pieces of code. A test set with no
correct examples in it cannot tell you your rule is too greedy. Second, while
gathering numbers for you I wrote a query that reported a clean, confident zero, and
the zero was an artefact of the question being aimed at the wrong thing entirely. I
did that **four separate times today** in different tools. Every time the wrong
answer arrived looking like a discovery. All of it is written up.

**A number I got wrong and corrected before it reached you.** I told myself, from a
first pass, that *no* reply channel carries the larger 5MB limit and they all run at
1MB. That was my own broken script — it was looking for the setting on the wrong
line. Done properly: **97% of reply channels DO carry the 5MB setting.** The ones
that do not are 91 long-lived, per-agent channels — including the main shared one —
which appear to have been created by something other than the code that sets the
limit. So this is not, as the ticket has it, "two numbers and no stated intent". The
intent is clearly the larger number; ninety-one channels simply missed it.

**Where it stands.** Two of the four findings now wait on you, below. The other two
need nothing: one is already fixed, and one I measured and would leave alone — its
worry is about a data structure that, as far as I can find, nothing actually reads.
