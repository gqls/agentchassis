# README — where we are (bugs_open/307), plain prose, newest at the bottom

## 2026-08-20, evening — what this is and why it needed a plan first

You asked me to look at bug 307: a three-hour GitHub outage on 17 August killed a hundred
pieces of queued work outright, and your ruling at the time was that a passing blip should
put the work back in the queue rather than kill it.

The retry machinery already existed and it already ran. The problem was that it retried
*immediately* — three tries inside a few minutes, all three into the same dead dependency,
and then the item was marked failed for good. Retrying with no wait is the same as not
retrying at all, if the thing you depend on is down for two hours.

While reading it I found something bigger than the outage, and it is the reason this is
worth doing properly rather than adding a wait. There are two different bits of code that
write "this work item failed", and only one of them counts attempts. The other one just
writes `failed` and stops. Five live agents use that second one — so for those, **one
failure is the end, on a normal day with nothing wrong**. Measured over the last fortnight:
141 items died before using up their allowance, 139 of them on their very first try. The
outage cost a hundred items once; this costs a handful every day and nobody had noticed,
because a failed item looks the same either way.

There is a third thing, contributed to the bug by another lane last week. When a handler
deliberately decides something — "a human needs to look at this", "we are not doing this
one" — the failure path can silently write over that decision. On the *success* path, two
sibling pieces of code both refuse to do that, with a comment explaining why. The failure
path has no such refusal. Until this week you could not even see it happening, because
everything wrote the same word. As of yesterday that changed: nine items now carry a
deliberate "won't fix", so from now on an overwrite would be a visible wrong answer.

So the fix is one thing, not three: **a single shared piece of code that every failure path
goes through**, which counts the attempt, waits before the next one, refuses to overwrite a
deliberate decision, and — when it can see that the whole fleet is failing on the same error
at once — puts the item back without spending an attempt at all.

That last part deserves explaining, because the naive version is dangerous. A deleted
repository and a GitHub outage produce the *identical* error message: "404 not found". If we
just treated 404 as temporary and retried forever, a genuinely deleted repo would grind away
until someone noticed. So we do not judge the error text at all. We ask a different
question: *is anything else failing the same way right now?* An outage shows up as many
items, across different customer sites and different kinds of agent, failing with the same
message inside a few minutes. A deleted repo fails alone. I measured a week of history for
false alarms: with that test, exactly three things trigger it, and all three were real
infrastructure outages. Nothing that was one item's own fault ever triggered it.

Two things I want to flag rather than bury.

First, I ran this through the platform's own diagnosis loop, as you asked. It read the code
and the database and built four rounds of evidence that agree with the above — and then its
final write-up step ran out of room mid-sentence and the run died. So there is no formal
verdict, only its working. The bug itself was filed on first-hand verification under the
rule that allows that, and I re-checked everything in the code and the live database this
session, so I am confident in the finding. But I am not going to claim a verdict I do not
have. I also think the run died partly because I asked it about three mechanisms at once
when it is designed for one.

Second, the waiting period needs somewhere to live, and the obvious choices were all traps.
Parking an item as "blocked" looks right and is quietly useless: a cleanup job runs every
ten minutes, releases every blocked item, and wipes the note saying why it was blocked.
That is why there are zero blocked items in the database and always have been. So instead
the item stays in the queue and carries a "not before this time" stamp, which the three
places that pick work up learn to respect. And the waiting *numbers* come from a policy
table that already exists for exactly this purpose and has been waiting for a second user
since it was built.

Next: build it, put it through the review council, and then it needs a deploy before any of
the code half takes effect. The database half goes live the moment it is applied and is
harmless before the deploy.

## 2026-08-20, later — built, submitted, and half of it is live

It is built and committed. Here is what actually happened, in order.

**The code.** One new file does the whole job: every failure path in the fleet now goes
through a single piece of code that counts the attempt, waits before allowing the next one,
refuses to overwrite a decision a handler deliberately made, and — when it can see the whole
fleet failing the same way at once — puts the item back without spending an attempt. The two
places that used to do this differently now both call it. That is the entire fix; the rest is
plumbing so the wait is actually honoured.

**The waiting.** An item now carries a "not before this time" stamp, and the four places that
pick work up all check it. Two of those are in the program, two live in the database as
configuration. The waiting periods themselves come from a small policy table that already
existed for exactly this and had been waiting since it was built for a second user — so the
numbers are yours to retune with one line of SQL, no rebuild.

**What is live right now, and what is not.** The database half is applied and I have verified
it changes nothing yet: the column exists, no item carries a stamp, and both dispatch queries
respect it. That was deliberate — it is inert until the program half ships, so there is no
window where work could be withheld by accident. I confirmed the dispatcher is still running
normally afterwards and that zero items are being held back by the new rule. The program half
is committed but does nothing until the next chassis deploy, which any lane's build will carry.

**Three things I got wrong on the way, all caught before they shipped.** I would have parked
waiting items in the "blocked" state, which a cleanup job empties every ten minutes while
wiping the reason. I would have detected outages by counting distinct *sites*, and that column
is empty on nine rows out of ten — it would have failed to fire during the very outage it was
written for. And my own change nearly wrote an empty value into a column whose *emptiness* is
how two other people's investigations tell the two writers apart. All three are written up in
the technical notes rather than tidied away, because they are the useful part.

**One honest gap.** The guard — the part that stops a deliberate decision being overwritten —
cannot be proven from the data we have, because until this week both paths wrote the same word,
so no query can tell a correct write from an overwrite. Nine items now carry a distinct
"won't fix" decision, and those are the ones where an overwrite would finally be visible. I
have written down in advance that "the guard did nothing" is the *expected* reading of a
prophylactic, so nobody later mistakes silence for failure.

**Where it goes next.** It is with the review council now (the architecture seat is reading it,
which is right for a change to something this shared). After the next deploy I will check the
running binary really carries it, watch a normal failure take the new path, and then the real
test is the next outage: it should leave nothing dead behind. The bug stays open until then —
a fix that is committed but not yet running is still a bug in production.

## 2026-08-20, evening — the review council sent it back, and it was right to

The council reviewed it and returned "revise". I want to be plain about why, because the thing it
caught was mine and it was real.

**It found me reintroducing the exact bug I was fixing, one line away from where I fixed it.** Part
of this change stops a deliberate decision being silently overwritten when work *fails*. While I
was in that file I added the same protection to the neighbouring path — the one that runs when work
*succeeds* — and reused the same list of protected statuses. But the two lists have to differ, and
I knew that: I had written three paragraphs explaining why. The failure list deliberately leaves
two statuses unprotected, because moving through them is what a retry *is*. On the success path
those same two must be protected, or completing an item could quietly paper over one that had
already failed. So the one-line freebie I added on the way past would have re-opened the hole.

Worse, I had written in the submission that this new guard was "the same as the two existing ones".
It wasn't — it differed in precisely the two entries that mattered. The reviewer did the one thing
I hadn't: compared my list against the existing one I was citing, instead of against my argument
for why mine was right. Fixed, with two tests that fail if anyone reverts it, and I've written the
lesson up: **the by-the-way edit is the one nobody reviews, including me.** Fifteen tests covered
the part I was concentrating on and not one covered the part I added because it seemed obvious.

**A second reviewer caught me understating my own headline number.** I had said 141 of 270 failed
items in a fortnight died before using up their retries. That came from a table that only keeps
about a week — the older rows move to an archive. Counted properly it is **401 of 558**, seventy-two
per cent rather than fifty-two. My case was nearly three times stronger than I claimed. That sounds
like good news and I've logged it as a mistake anyway, because nobody ever re-checks a number that
argues *against* their own conclusion, so I'd have kept repeating it forever. What stings is that
I'd written the very check into this project's own runbook that same morning, and didn't run it.

**Three reviewers independently refused to accept a promise.** I had written that one leftover
piece — an old copy of this same retry logic that lives in the database rather than the program —
was "a known residual". They said, in three different ways, that a sentence in a rationale is not
a tracked piece of work. They're right, and one of them put it well: the untreated sibling is where
the next incident comes from, not a footnote. It is now filed as its own bug with the options
costed, including the honest obstacle — that database-resident code can't call program code, so
converging them is a question about ownership rather than a patch.

Two smaller ones: a reviewer caught me claiming retried items become reapable after 48 hours when
in fact the clock restarts on every write, so it's 48 hours after the *last* one; and another
pointed out that my database change overwrote a piece of configuration without first checking it
was still what I'd read minutes earlier. That one hadn't caused harm — I verified nothing was
clobbered — but the *undo* script is the thing somebody runs months later in a hurry, so it now
refuses to run if the configuration has changed underneath it.

All of it is fixed and resubmitted. Nothing had reached production: the program half doesn't run
until the next deploy, which is exactly the window this review exists to use. A round that finds a
real defect is cheaper than the defect, and this is the second time on this project that's proved
true.
