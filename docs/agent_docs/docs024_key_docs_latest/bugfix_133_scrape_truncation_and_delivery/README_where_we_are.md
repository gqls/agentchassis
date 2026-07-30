# Where we are — bug 133, the scrape reply that lied and the scrape reply that vanished

Plain prose, append-only, newest at the bottom.

---

## 2026-07-30 — the whole thing, start to finish

**What was wrong.** When we scrape a page, the result has to travel to whoever
asked for it over our message bus, and the bus has a size limit. So the adapter
cuts anything over 50,000 characters and adds a note saying "the full version is
in S3". That note was added *every single time*. The upload that would put a copy
in S3 is optional, and four — it turns out nine — of the live scrape steps never
turn it on. So on those, we threw the end of the page away and left a note saying
where to find it. There was nothing to find. Someone reading that reply, human or
model, would believe the missing text was one fetch away.

The second fault was quieter and worse. If the bus refused a reply for being too
big, the adapter wrote a line in its own log and gave up. The thing waiting for
that reply is not reading the adapter's log; it is waiting on a message that is
never coming. It waits out its whole retry budget and then fails for a reason
that has nothing to do with what actually happened.

**Neither was ours to find.** Another thread found both on 28 July while doing
something else entirely, wrote them up properly, and said plainly that they were
not its bugs to fix. So they sat.

**What I did about the second one, and why it took longer than copying.** The
batch version of this same code — scraping many URLs at once — already does the
right thing: it shrinks the reply, tries once more, and if that still won't go
through it sends a *real error* so the caller learns something. The obvious fix
was to copy that block across.

I didn't, and this is the part worth explaining. I checked how many places in the
whole system knew that "too big" is a permanent failure rather than a temporary
one. **One.** And I checked how many places send a reply at all. **Nine, across
five services** — eight of them with the give-up-quietly behaviour. So the rule
we wrote down after the last time this bit us was being followed in one place out
of nine. Copying the block would have made it two out of nine and left us with
two hand-written copies of one rule, which is exactly how they drift apart. We
closed a bug last week that *was* that drift.

So the policy now lives in one place, both scrape paths call it, and the batch
version's private copy is gone. The two things that genuinely differ per service
— how to shrink *this* kind of reply, and what *this* service's error message
looks like — stay where they belong.

**What I did about the first one.** The problem was not the wording of the note.
The problem was that the sentence "full version in S3" could be written by code
that had no idea whether anything had been uploaded. So now the note is built
*from* the storage address: hand it an address and it names it, hand it nothing
and it says the remainder was discarded. There is no longer a way to phrase the
claim without having the evidence in hand. Someone editing this later can't
reintroduce the lie without deleting a function argument and noticing.

It resolves the address per *field*, because the uploads happen one field at a
time and any one of them can fail quietly. A single "did we upload?" flag would
have gone back to lying about the fields whose upload failed.

**Three things I got wrong along the way**, all the same mistake. Twice I read a
340-line function through a window that stopped short of the part that mattered,
and twice I was a minute away from writing a confident wrong finding into the bug
file — first "we never upload raw HTML" (we do, forty lines past where I stopped
looking), then "we never upload page content" (we do). The third time it wasn't
my measurement, it was the bug's: its table of affected steps was built from a
query naming three actions, and there are six. The missing one is the busiest
scraper we have. That is why the count in the bug — four steps affected — is
really **nine**, and why it includes the multi-page crawls where a stored copy
mostly cannot exist at all.

I only found that last one because I went to read a real message off the queue
before firing a test one, which is a habit the filing thread wrote down after
making the opposite mistake.

**One nice moment.** I wrote a test that scans the code for the old lying
sentence, to stop anyone putting it back. It failed immediately — on my own
comment, where I'd quoted the old sentence to explain what had been wrong with
it. The test couldn't tell a real message from a note *about* a message. That
distinction is not academic: it's the same confusion that made another thread's
evidence unsound last week, when four of its five checks were looking for
sentences that only ever existed in documentation. So the test now parses the
code properly and looks only at text that can actually be sent.

**Is it live?** Yes, and checked properly rather than assumed. Before shipping I
recorded what the *old* binary said, so the check afterwards had a right answer
and a wrong answer instead of just a right one. Then I fired one real scrape of
one of our own sites, reproducing the exact case from the bug report, and read
what came back off the queue. The reply now says the text was discarded, says so
in a machine-readable field as well as in prose, and the old sentence appears
nowhere in it.

Midway through, another session rebuilt and rolled the same service. Its version
number is higher than mine, so it "must" contain my change — that assumption is
one we have a whole bug about, so I re-checked all three copies. It does.

**The review.** Put through the council; approved in nine minutes. I flagged the
one judgement I couldn't settle by measuring: putting a shared mechanism into a
core piece of messaging plumbing is the kind of thing that's supposed to go
through a heavier architecture review, and I wanted to be told if I'd read the
rule wrong. Two seats disagreed with each other about it — one said hold the
concern anyway, the other said it's a point fix because only one package uses it
so far. The second view carried, but with a caveat I've written into the register
word for word: **the moment anyone adopts this in the other services, that is an
architecture review, and today's approval does not cover it.**

Two reviewers asked good questions I could answer by measuring rather than
arguing. One asked whether some other adapter already had a helper like this that
I should have extended — no, the similar-looking things are all just trimming
text for log lines. The other asked why I wrote a new test stub instead of
extending the shared one; the answer turned out to be more interesting than the
question: the shared one has no users at all and doesn't even implement the
interface it claims to, while two other packages have each quietly written their
own. So there are three, and mine is a fourth. I've recorded that rather than
widened this fix to chase it.

**What I deliberately did not do.** I did not turn the upload on by default —
that changes cost and behaviour across four other teams' agents and the bug
itself says it's an owner's call. I did not adopt the new shared policy in the
other eight places, because that changes what four other services' callers see
when something fails. And I did not fix a related flaw I found in how page
storage addresses are indexed, because it changes the shape of data we hand
downstream. All three are written up rather than dropped.

**Still needing a decision from you:** whether replies should be allowed up to
1MB or 5MB. The cluster currently implies 1MB for these topics while our own
topic-creating code sets 5MB on topics it creates, so we have two numbers and no
stated intent. The bug flagged it; I have not resolved it, because guessing would
just add a third number.
