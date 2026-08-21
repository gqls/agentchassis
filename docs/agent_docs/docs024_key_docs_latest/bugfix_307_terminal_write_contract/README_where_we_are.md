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

## 2026-08-21, morning — it deployed, and the first real blip was handled the way you asked

The review approved it at the second round yesterday afternoon, and the deploy that afternoon
carried it — on someone else's build, which is how this shared tree works by design. It has been
running since four o'clock yesterday.

And it has already done its job once, without any of us arranging it. Just after six-thirty
yesterday evening, two page-rebuild jobs hit a short messaging-system hiccup — the kind of
passing infrastructure blip this whole piece of work was about. Under the old behaviour each
would have spent one of its three lives on a problem that wasn't its fault, or died outright.
Under the new behaviour: both were put back in the queue without losing a life, each with a
short "don't retry before" time stamped on it; the system respected that time, picked them up
again afterwards, and both finished successfully. That is exactly the sentence you gave us —
"a transient blip should return the item to queued" — happening on its own, on real work.

The other number worth saying out loud: in the eighteen hours since it went live, not a single
work item has been marked permanently failed. Before this fix the estate averaged roughly
thirty a day, most of them dying on their first attempt. Failures still happened overnight —
nearly three hundred error events — they just stopped being fatal.

What's left is one supervised drill this morning: I'll feed the system a single deliberately
doomed item on a test site and watch it use its three lives properly, wait out its cooldowns,
refuse to trample a human decision, and only then be declared failed for good. You approved
that this morning, along with closing the bug once the drill passes — with the "next real
outage leaves nothing dead behind" test kept as a standing watch rather than a reason to hold
the file open for weeks. Then the ticket closes and the drill row gets deleted.

## 2026-08-21, midday — the drill found two real faults, so the ticket stays open

The supervised drill ran this morning, and it earned its keep twice over.

The good news first: the retry machinery itself does exactly what you asked. The doomed test
item was put back in the queue with its first life spent and a half-hour cooldown; on its
second failure the cooldown correctly doubled to an hour. And overnight, two real jobs that hit
a passing infrastructure blip were requeued without losing a life and finished fine on retry.

The two faults. First: two seconds after the retry machinery correctly requeues a failed item,
a different piece of housekeeping — the one that marks finished work as done — comes along and
stamps it "complete", because the failed build still ends with a polite "I finished" message.
So a failed piece of work is recorded as done and its retry is cancelled. This started the
moment our fix went live: before, the same situation ended in an honest "failed" which that
housekeeping refuses to touch. A real page on the mortgage calculator site hit this
overnight — its record says done, its page never got the content. Filed as bug 344 with the
repair options costed.

Second: when an item genuinely runs out of lives, the write that should mark it permanently
failed turns out to crash on a one-character database technicality that our fifteen tests
could never see (they use a stand-in database that doesn't check that particular rule). So
nothing can currently run out of lives — a doomed item just cycles for ever, quietly burning a
real build attempt each lap. That one I have fixed today, with a new kind of test that checks
the rule the stand-in can't; the fix rides the next deploy and is with the review council now.

So the ticket does not close today, and that's the system working as designed: the drill
existed precisely to catch what the tests couldn't. The path to closing is now: next deploy
carries the crash fix, a decision on bug 344, then the same drill re-run clean.

## 2026-08-21, midday — I have to correct what I told you this morning

This morning I said the fix was live and proven. **Half of that was right and I need to take the
other half back.**

What is genuinely working, on real traffic and unprompted: when a passing infrastructure fault hits
a job, the item goes back in the queue **without spending one of its three tries**, waits, is picked
up again, and finishes. Two real items did exactly that on Wednesday evening. That is your ruling —
"a transient blip should return it to queued" — behaving as asked, and it is not a test.

**What does not work is the opposite end, and it is my mistake.** When an item genuinely deserves to
give up — three failures, nothing transient, time to stop — the code that writes "failed" **crashes
instead**. A parameter is passed to the database that the statement no longer mentions, and the
database refuses the whole write. So an item that should die honestly at the third attempt doesn't
die at all. That is the single most important behaviour in the change after the retry itself, and
it is the exact case I was most confident about, because I had written a specific test for it.

**Why my tests missed it, which is the part worth your attention.** I wrote fifteen tests and
deliberately broke my own code five different ways to prove they'd notice. They all passed, and
they were all asking the wrong kind of question. They test the SQL as *text* — does the sentence
contain the right words. The fault is in the SQL as a *program* — the sentence is well-worded and
the database still rejects it. A fake database was standing in for a real one, and a fake database
never objects to anything a real one would. Fifteen tests with proofs read as thorough; that
thoroughness is precisely why I never asked what the whole approach structurally couldn't see.

**How it was caught: by the canary you authorised.** Another session ran the synthetic item through
three real attempts rather than trusting the tests, and attempt three produced the error. That is
the entire argument for canaries in one sentence — and the reason I'd resist any future suggestion
to skip one because "the tests are good".

Two other things came out of the same run, neither mine to fix:

- **A second, separate defect** (now `bugs_open/344`): when the retry machinery puts an item back in
  the queue, the surrounding job reports overall success two seconds later and stamps the item
  "complete" — wiping the retry. That was harmless before this change and became load-bearing
  because of it. It has a contained fix proposed and is waiting on a decision.
- **My own bug file for the leftover piece overstated its case.** I'd written that the old
  database-side copy of this logic lacks a safety guard. It doesn't — its own filter is *stronger*
  than the guard. I'd compared two pieces of code on a checklist instead of on what each can
  actually reach. Corrected in place; the fix I built for it shrank accordingly.

**So: 307 cannot close, and I was wrong to imply this morning that it was close to closing.** The
other session is fixing the crash now. My leftover piece is built, checked, committed and
**deliberately not switched on** — turning it on before `344` is decided would spread `344`'s damage
wider, so it is filed under a name the tooling itself refuses to apply until someone releases it.

The honest position: the hard, novel half of this change works in production and is proven. The
simple half — giving up correctly — is broken, understood, and being fixed today.

## 2026-08-21, evening — the drill passed everything, and the ticket is closed

This afternoon's deploy carried both repairs: my fix for the crash-on-last-life, and the other
session's fix for the housekeeping that was stamping failed work "complete". So I ran the whole
drill again, from scratch, on the live system.

It passed every part. The doomed test item lost its first life and went back in the queue with
its half-hour wait — and this time nothing stamped it "complete"; the record stayed honest. I
then marked it "we are not doing this one" mid-run, on purpose, and watched the failure
machinery decline to touch it — the deliberate decision stood, and the system's own log says
so in as many words. Put back in the queue, its second failure doubled the wait to an hour, as
designed. And on its third and final failure it was marked permanently failed, correctly and
quietly — the write that used to crash. The test row was then deleted.

The review council approved the crash fix first time, with two pieces of advice, both of which
I acted on rather than filed: a test now proves the old-database fallback really works, and
the full list of the eleven agents that use this machinery is written down rather than assumed.

So bug 307 is closed and moved to the closed pile. The one test we cannot run on demand — a
real infrastructure outage leaving nothing dead behind — is recorded as a standing watch with
an explicit "reopen if" attached, exactly as you ruled this morning. What remains open is
smaller and has a named owner: the other session is finishing the database-side half of the
completion fix (bug 344) and its own cooldown work (bug 341), both of which this closure
points at rather than absorbs.
