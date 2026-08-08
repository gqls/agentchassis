# Where we are — the 2026-08-05 code review

Plain prose, append-only, newest at the bottom. This is the owner's log.

---

## 2026-08-05, afternoon — the triage got actioned, and two of the fifteen findings were wrong

This morning a session ran a code review over the working diff, got fifteen findings, and
sorted them by which lane owned them. It deliberately fixed nothing, on the reasoning that all
three lanes were still active and you should contribute into a lane rather than compete with
it. That was the right call at 11:02. By 11:03 it had stopped being true — two of the three
lanes closed their bugs and moved to `bugs_closed/`, and neither of their closing handoffs
mentions the code review at all. So ten of the fifteen findings were nobody's. That is what
made it reasonable to just fix them.

Here is what actually happened, in the order it mattered.

**The 140MB of binaries was the easy one, and it turned out we had done this before.** Four
compiled Go binaries were sitting untracked in the repo root, unmatched by any ignore rule. I
went to add them and found the `.gitignore` already has a section for exactly this, written
after a 93MB binary was committed by accident on 2026-07-20. So this was not a new risk, it
was the same accident lining up a second time. Four lines appended to the block that was
already there, with the reason it already gave: our builds use `git archive`, so a tracked
binary at the root gets extracted into *every* service's build context.

**Two findings were simply wrong, and one of them I nearly got wrong the same way.** The
morning's triage had already caught one false positive — a finding that asserted no live agent
sets a particular config key, when exactly one does. I found a second. A finding said we write
unboundedly to an error-log table with nothing cleaning it up. I measured it: the table had a
month of history and my search for cleanup code found nothing, so I wrote down "no reaper".
That was false. There is a reaper, it runs every hour, it had run minutes before I looked, and
it deletes anything older than thirty days. The "month of history" I read as evidence of
neglect *was the thirty-day boundary*, drawn to the minute — the oldest surviving row was
twenty-one minutes inside it.

What makes that worth telling you rather than quietly fixing: the number I relied on could
never have come out any other way. A thirty-day-old oldest row is what a working thirty-day
reaper produces, and it is also what no reaper at all produces. It does not distinguish
between them, and I did not stop to ask what a disconfirming result would have looked like. I
also searched only our Go code, when the cleanup is written in SQL and lives in a database
column. Both errors are in `WRONG_CALLS.md` now. The finding is recorded as a false positive
rather than filed as a bug — the same outcome the morning session reached for its one, by the
same route.

**The most valuable finding was a name collision, and it had already bitten us.** Yesterday's
work gave a config key called `require_sections_metadata` the meaning "refuse to save this
page". That exact spelling was already live, in the same package, meaning something much
milder — "warn me that a check could not run". Worse, one of our agents, `page-build-handler`,
carries both steps in a single definition, so the same word would have meant "warn" on one
line and "refuse the save" a few lines later. The trap is that the natural way to roll a
setting out across the fleet is a bulk update by key name, and that would have armed a hard
refusal on our highest-traffic page-save path.

The proof it is a real trap and not a theoretical one: I found a comment in our own codebase
that had *already* been confused by it, describing the two keys as though they were one thing
set on one step. They are not. I renamed the new key, which cost nothing because we had
deliberately shipped it switched on for nobody, and I corrected the comment in place rather
than deleting it, because a wrong comment that was caused by the collision is evidence for
fixing the collision.

**One finding asked for something impossible.** It said a database write here should call our
shared logging helper instead of writing its own SQL. It cannot: the shared helper lives in a
package that imports this one, so calling it would be a circular import. About twenty files in
that package each write their own copy for that reason. The finding's premise — that the
package's own convention forbids the duplication — was backwards; the duplication *is* the
convention, and it is forced. The half of that finding that was real (the row could not be
linked back to the run that produced it) is fixed.

**Two findings were assigned to the wrong lane, and the method is the reason.** The morning's
triage worked out ownership by asking git who last changed each *file*. Two lanes touched one
file twenty-six minutes apart yesterday, so that method gave the wrong answer for two
findings: it named the still-open lane when the lines in question belong to the closed one.
Asking git who wrote the specific *lines* settles it. Both are fixed rather than handed to a
lane that does not own them.

**What I did not do.** One finding is in a file another session has open work in right now —
their uncommitted changes are sitting in the working tree, and the very comment the finding is
about is part of what they have not committed yet. Touching it would have edited their work
mid-flight and swept it into my commit. I confirmed the finding is correct and worked out the
exact corrected line numbers, and left it for them. And one finding, F6, I could not action at
all: it is named in the triage's ownership table but never described anywhere, and the
original review output was not saved, so there is no record of what it actually claimed. I
have flagged that rather than guess.

**Where it stands.** Twelve findings resolved: nine fixed and committed, two recorded as false
positives with the measurements that refute them, one confirmed and handed back to its owner
with corrected line numbers. One cannot be recovered. All the code changes went to the review
council; the first verdict came back approved, and its one criticism was fair — I had claimed
two functions had no callers anywhere on the strength of a text search, which the reviewers
rightly said they could not verify. I redid it as a proof that could actually fail: rename the
functions, rebuild, and see whether anything breaks. Nothing did, which is what "no callers"
should mean.

## 2026-08-05, evening — the fixes are live, and one owed check now means something different

The chassis rolled to a new build and the code changes from this morning's work are live on
both replicas. I did not take the roll as proof of that — our own notes record that an image
carries no record of which commit built it, so a roll tells you something changed, not what. I
checked inside the running binary instead, on each replica separately, looking for two strings
that only exist because of these changes and one deliberately misspelled string that should not
exist at all. Both real strings present on both replicas, the misspelled one absent on both.
One of the two only exists after the very last code change, so its presence dates the image
past all of it.

One honest gap in that check, which I would rather write down than paper over: the strongest
version of this test also looks for a string the change *deleted*, and confirms it is gone.
This change set cannot supply one — most of what I changed was comments and internal names,
neither of which appear as text inside the binary, and the one piece of SQL I removed appears
in fifty-one other places. So I can prove the new code is in there; I cannot prove by the same
means that the old code is out.

The more interesting consequence is a check we already owed ourselves. When yesterday's work
shipped, it left a note to look back 24 hours later and count how often a particular
"this page just lost its structured content" warning had fired, with a clear rule attached: if
it fires for the re-render agent, the warning is wrongly designed and we must not roll it out
any further. One of today's fixes widened exactly the condition that warning tests. So that
check now asks a broader question than it did when it was written, and I have set out in the
notes what each possible answer will mean before seeing it — which is the part that stops a
result being read the way we would like to read it.

Right now the count is zero, which is the stated pass. It is also a weak pass and I have
recorded it as one: the warning has never fired in any version, so zero is what a correct
change and a broken one would both produce. The fleet is definitely running, but no page-save
traffic has crossed the window yet. The real read-out is due tomorrow evening.

---

**2026-08-06, morning.** Picked this up about thirteen hours before that read-out is due, to
get it ready rather than to take it early. Four things worth telling you, and one of them is a
mistake in yesterday's own instructions.

First, the system was rebuilt and restarted again overnight, for reasons nothing to do with
this work. That quietly broke yesterday's proof — not the fix, which is still in there and I
have re-checked it on both of the new machines, but the *evidence*. Yesterday's note named two
specific machines, and those no longer exist. Anyone reading it later would have been quoting
something that had vanished. So there is a general lesson written down now: after somebody
else's rebuild, you have to go and look again, because your proof has an expiry date you don't
control.

Second, and this is the mistake: yesterday I wrote down a way of checking whether the thing
we're measuring had actually happened at all, and that check cannot work. It counts entries in
an *error* log. When the operation we care about runs perfectly it writes no error, so the
answer is zero whether it ran a thousand times or never. I would have read a healthy system as
a silent one. I have replaced it with a check that looks at what the operation actually leaves
behind, and that one says clearly: it ran, on ten pages, and every single one kept its
structured content. That turns tonight's zero from "we learned nothing" into "nothing went
wrong, and here is the traffic that proves we'd have seen it if it had".

Third, a piece of luck worth noticing. The rule attached to tonight's check says: if this
warning fires for the *re-render* agent specifically, the design is wrong and we stop. It turns
out the re-render agent is the **only** one that has run since the change went live — twenty-three
times. So tonight's reading lands squarely on the case the rule was written about, which is
better than it might have been.

That third point came with a deadline I nearly missed. The records showing *which* agent ran
are themselves deleted after twenty-four hours — meaning they would have disappeared at almost
exactly the moment tonight's check was due to be read. I have taken and saved that evidence now.
The general form of it: if you schedule a check for a day later, and the supporting evidence
lives somewhere that clears itself out after a day, you have to grab the evidence early or you
will arrive to find the answer and no way to interpret it.

Fourth, my own slip, caught rather than shipped. Trying to confirm those twenty-three runs had
really completed, I asked the question in a way that could only ever have answered "no" — I
looked for a name in a list, and the field is actually a plain count, so the comparison was
meaningless and came back empty every time. I only noticed because it contradicted the other
measurement. It is a small thing, but it is the same shape as most of the errors in this
project's log of wrong calls: a check that cannot come out any other way isn't a check. Written
up where those go.

Where that leaves tonight: the reading itself is unchanged and still due at about 20:45. What is
different is that a zero will now mean something. I have also confirmed the second, smaller
outstanding check is settled — I did it by inspecting the configuration rather than the logs,
because the restart had wiped the logs down to half an hour, and configuration cannot expire the
way a log can. All six callers are correctly set up, so that warning genuinely cannot fire until
someone adds a new one.

---

**2026-08-06, mid-morning.** Another fresh build went out — that is the third restart of the
system in under three hours, none of them ours. Checked again on the new machines: our change
is in there, on both. I have stopped treating this as an interruption and written it down as a
standing fact about this cluster instead. A note that says "verified on machine X" has a shelf
life of hours here, so the rule now recorded is: never write a machine name into a document
without the date and the version beside it, and go and look again rather than quoting yourself.

The numbers have moved in the right direction while we waited. The relevant operation has now
run thirty times rather than twenty-three, across thirteen pages rather than ten, and **every
single saved row still carries its structured content** — forty-seven out of forty-seven. Still
no warnings recorded, which remains the result we want. A second agent has joined in too: the
page-rebuild agent ran once this morning, so tonight's reading will cover two callers instead
of one.

Now the part I am less pleased about. The query I wrote a couple of hours ago to work out
*which* agent had run was wrong in two ways, and I found it only by following up an oddity in
the numbers. I had identified each agent by a distinctive-looking setting in its configuration.
It turns out that setting is shared by four different agents, so it only ever gave a clear
answer for the one I happened to be asking about. Worse, the table has a column that simply
names the agent outright — I had missed it because I printed the table's structure and only
looked at the first two-thirds of the output. So I built a workaround for a problem that did
not exist, and the workaround was ambiguous.

The second error in the same query was counting the wrong things: it counted *steps* rather
than *runs*, and reported two where the truth was one run that happens to contain the step
twice. Chasing that down turned up something genuinely useful, though, so it was not wasted:
the record of what a run actually did is not the same object as the agent's stored recipe —
loops get unrolled into repeated copies when a run starts. That means one of them is the right
place to ask "how is this configured" and the other is the right place to ask "what actually
ran", and they are easy to mix up. Both are now written down with the distinction spelled out,
because an earlier mistake in this same lane was the identical confusion approached from the
other side.

None of that changes tonight's reading or what it will mean. It does mean the figures in the
handoff are now the corrected ones, and I would rather a successor inherit the correction than
the tidy-looking version I wrote at eight o'clock.

---

**2026-08-06, late morning.** Nothing to report on tonight's reading yet — it is due at about
quarter to nine this evening and there is no way to bring it forward. I checked anyway: still no
warnings recorded, all history. The other item I keep re-checking, the one sitting in a file
another session is halfway through editing, is still halfway through being edited. So I turned to
the loose end we had parked.

That loose end was a small note saying that when we record an error without knowing which website
it belongs to, we store an empty text box rather than a proper "unknown" marker — which matters
because anyone asking "how many of these are unattributed?" gets a wrong answer that looks
perfectly reasonable. We had written it down yesterday and moved on. Looking at it properly today,
the note was right about the symptom and wrong about almost everything else, in ways that would
have sent whoever picked it up at the wrong target.

Three things. First, the file we said it lived in no longer holds the code — a separate piece of
work moved that exact chunk into a new home *this morning*, and carried the fault across
untouched. There is a general lesson there worth more than the bug: moving code is not the same as
reviewing it, and a tidy-up can transport a defect into a fresh file where it looks newly written
and unexamined.

Second, our explanation of *why* it happens was wrong. We had said the neighbouring fields are
handled properly and this one was overlooked, implying carelessness. In fact those neighbours are
handled properly because the database physically refuses the alternative — they are a different
type of field, and the empty box is illegal for them, so the code had no choice. Which means that
one piece of code tells you nothing at all about what anyone intended for our field. I checked
this by trying it directly and watching the database reject it.

Third — and this is the part that changes the job — I nearly turned that finding into a worse
mistake. Having established that the neighbours were forced, I was about to write "so there is no
house style here, and our field is normal, and yesterday's note was simply wrong". Before writing
it I counted, and found the house style is real and widely followed: I found thirty-two places
doing it the careful way. So the true picture is the opposite of what I was about to record.
Twenty different places in the system write these error records. Seventeen of them handle the
unknown case properly. Three do not — and those three are obviously copies of one another, the
same block of code pasted three times. One of the three happens to be the general-purpose one used
across the whole system, which is why it produces most of the damage.

I then checked whether that story actually holds against the real data, in a way that could have
proved me wrong: if the explanation is right, the records written by the three bad copies should
all show the empty box and the records from the other seventeen should all show a proper unknown.
That is exactly what came back, with no overlap anywhere across fourteen thousand records. So this
is now an explained fault rather than an observed oddity.

What it means practically: the fix is smaller than we said — three characters in three places,
bringing three stragglers into line with what everything else already does, rather than inventing
a new convention. But it is still a change to what gets permanently stored by the busiest writer
we have, so it wants a proper review round and it is not something to slip in quietly. I have not
done it. What I have done is write the whole thing down properly, including in the fleet-wide
"traps" file so that the next person to query this data is warned before they get a wrong number
rather than after. And I logged my own near-miss, because the shape of it is worth remembering: I
had genuinely verified one narrow fact, and let it lend its authority to a much broader claim
sitting in the same sentence that I had not checked at all. The proportion is the thing to watch —
the number I would have quoted, had I not counted, was out by a factor of about eighty.

---

**2026-08-07, small hours.** The reading we had been waiting for is done, and it came out clean.

The short version: **no problems recorded, and this time the silence means something.** Yesterday
I was careful to say that a zero told us very little, because almost nothing had gone through the
part of the system we had changed. That is no longer true. Since the change went live, the
relevant operation has run **forty-eight times across three different agents**, and has saved
**fifty-five pages' worth of sections across sixteen pages** — every single one of them carrying
the structured content that the whole change exists to protect. Fifty-five out of fifty-five. So
the alarm stayed quiet because there was nothing to complain about, not because nobody walked past
it.

One detail is more satisfying than the raw count. The specific agent we were most curious about —
the one whose behaviour we had deliberately widened the check to cover — had never run at all until
yesterday lunchtime. It has now run three times, and it recorded nothing. That is the case the rule
was written for, and it passed.

I want to be honest about what this does and does not prove, because it would be easy to write
"verified" here and it would not be true. We have shown the alarm *can* go off, but only in a
laboratory sense — a test proves the decision logic fires correctly. Nobody has ever seen the whole
chain work end to end: decision, then write the record, then find the record. So the accurate
sentence is "forty-eight runs, no problems, on an alarm we know can sound" — not "this is proven".
I have written that distinction into the permanent record in three places, because it is exactly
the kind of caveat that gets quietly dropped when someone quotes a result later.

I was four and a half hours late doing this, and it is worth saying what that cost, because the
answer is nearly nothing and I had expected worse. The system deletes its records of completed work
after a day, so some of the evidence had already gone. But yesterday morning I had deliberately
copied the important figures out *before* they could expire, precisely because I could see this
coming — so the numbers were reassembled from that saved copy plus the part still on disk. The
decision to take those figures early is the reason the lateness was survivable.

Now the two things I got wrong, both caught in the same sitting.

First, I tried to check the claim that this system deletes completed work after a day. I asked it
for the oldest record it held, and it said **twenty-four days**, which flatly contradicts a
one-day deletion policy. I was about to write that down as "the deletion policy we have been
relying on does not exist". It does exist. The catch is that the deletion only applies to work that
finished — and the twenty-four-day-old record is one lone job that got stuck halfway through in
July and has been sitting there ever since, untouched because it never finished. Asking for the
oldest record of *any* kind cannot tell you anything about a policy that only removes *finished*
ones. When I split the question by outcome, the answer was immediate and precise: finished work is
being removed at almost exactly twenty-four hours. I then found the actual cleanup instruction and
read it, rather than continuing to infer it, and the boundary it produces matches what I measured
to within thirty seconds. This is the second time in two days this lane has been caught by the same
shape of mistake — a number that would have looked identical whether the thing existed or not.

Second, and slightly embarrassing: yesterday I described a count of error records as covering "all
history". It does not. That table is also cleaned out — anything older than a month is deleted.
Worse, **this lane had already discovered that, yesterday, and written it down in the same
document** about a dozen sections above where I then wrote "all history". So the fact was not
missing; I simply did not go back and look at my own notes before making a claim about that table.
The proportion I reported is still correct for the records anyone can actually look at, which is
what matters in practice — but "all history" was wrong and is now marked as wrong everywhere it
appeared.

Finally, the loose end from yesterday. On your go-ahead I have **submitted the empty-domain problem
to the diagnosis system** — the internal reviewer that reads the real code and the live database and
either confirms a diagnosis or refutes it. That is the step our own rules require before writing up
a fault of this kind as established, and I had skipped it yesterday. It is running now and building
up its evidence. I will not write it up as a confirmed fault until I have read what it says, and I
want to flag in advance that **being refuted would be a good outcome, not a wasted one** — it costs
one run, and it is much cheaper to be wrong there than in a document other people go on to trust.

---

**2026-08-07, still the small hours — the diagnosis came back, and the headline is not what it says
on the tin.**

The verdict is **"unverifiable"**, which sounds like a failure and mostly is not one. The reviewer
independently worked out the same fault I had described — it said in its own words that the
mechanism is real — and then could not check the *one* thing I most wanted checked, namely *where*
in the code it lives. The reason is worth understanding, because it will bite anyone else who uses
this reviewer: **it reads the code through an index that has not been rebuilt since 28 July.** The
file I pointed it at was created on 6 August. So from where it was standing, that file does not
exist, and it looked at the older version of the same code instead — where, quite correctly, it
found the fault sitting somewhere else. We were both right about different copies of the codebase.

To its credit it *said* so, unprompted, and listed re-checking that as something it needed. That is
a safeguard someone built deliberately and it worked. But the practical upshot is that **there is no
point asking it again until the index is rebuilt** — it would spend another round and give the same
answer. I have written that warning where the next person will see it.

Now the part I did not expect, and which I think matters more than the domain question.

**The reviewer's answer was very nearly thrown away, and every indicator said everything was fine.**
The job was marked complete. All three internal records said completed. The place where reports are
kept had five entries, and every single one of them was raw working material rather than a
conclusion — no report at all. If I had done the obvious thing, which is check the status and then
go and read the report, I would have found a clean bill of health and an empty shelf, with nothing
anywhere to tell me whether the reviewer had declined to answer or had answered and lost it. It had
answered. The conclusion was sitting in an internal scratch area that nothing points to, **and that
area is automatically wiped twenty-four hours after a job finishes.** I found it by listing what was
in there and spotting a field called "verdict".

What went wrong is narrow and, oddly, half of it is the system behaving well: the reviewer's answer
was rejected in transit by a message check, and the code that noticed this deliberately reported a
failure upwards rather than pretend success — there is even a comment explaining that choice. The
gap is that the failure never reached the job's own status or the report shelf. I have written this
up as a trap with the recovery steps attached, because the cost of not knowing it is an entire
review round plus a day to notice. I have **not** worked out *why* the message was rejected, and I
have deliberately not guessed in the notes — I have an idea, and I have labelled it as an idea.

One more correction to yesterday, from something the reviewer surfaced that I had missed. I had said
the empty-domain records mostly come from the busiest general-purpose writer, as ordinary background
volume. The mechanism was right but the volume story was wrong: **about eighty-eight per cent of
them are a single failure happening right now in the veterinary lane** — roughly twelve thousand
error records in two and a half days, still arriving as I write, largely website-scraping failures
with certificate errors. That changes how the original problem should be argued: the eye-catching
"eighty times under-reported" figure is really a measure of *that incident*, not of the fault, and it
will shrink when the incident stops. So the case for fixing it has to rest on the mechanism and on
readers getting wrong answers — not on the size of the number.

And separately, and not something I have looked into: **twelve thousand failures in two and a half
days out of one lane, still climbing, probably wants somebody's attention on its own account.** I
have named two existing open items that might be related and have deliberately not investigated or
filed anything, because that is somebody else's lane and guessing at it from here would be exactly
the mistake this log keeps recording.

---

**2026-08-08, evening.** The small fix has been through the review council and **passed** — six
reviewers, one minor comment, about fourteen minutes. What is approved is genuinely tiny: two
lines, each adding the same wrapper the neighbouring columns in the very same statement already
have, so that "we don't know which website this error belongs to" is recorded as *unknown* rather
than as an empty box. **The code is not written yet.** The council reviewed the plan; writing and
shipping it is the next job, and it is a small one.

Two things are worth telling you because they changed the shape of this.

First, the problem got both worse and much easier on the same day, and I only found out because
the reviewer looked at fresher code than I had. On the 7th another thread consolidated eighteen
separate copies of this error-writing code into one. That means the flaw is now everywhere by
construction rather than in three odd corners — but it also means **the fix went from three places
to one line in one shared place**. My earlier survey of the code was correct when I made it and
was false about twenty-four hours later. That is worth remembering about this codebase generally:
a survey of how something is built has a shelf life of roughly a day here.

Second, the council raised two criticisms and I checked both rather than pocketing the approval.
One reviewer said, fairly, that I had claimed "only two places remain" without proving it
independently — and my search had indeed only covered part of the tree. I re-ran it across the
whole repository, and in a form that would catch different spellings and spacing: still the same
three places. The other asked whether this exact fix had already been reviewed before; it has not
— the two earlier reviews that looked similar were the consolidation work itself.

That consolidation is worth one more sentence, because it is a bit uncomfortable and a successor
should know: **it was rejected by the council the first time**, on the grounds that it asked for
approval of thirty-four files on the strength of an eight-file sample. It shipped anyway and was
approved a day later. That is how this place is designed to work — review happens after the fact
because no thread can hold code back on a shared branch — but it does mean the shared writer
everything now depends on passed through a rejection on its way in, and nobody noticed the
empty-box flaw in either round.

Separately, and still needing a decision from you: the safety report we have been watching **has
now gone off**, on a live page. I dug into it and the honest answer is that it is a false alarm of
a very particular kind — it fired because a page "had structured content" where that content was
literally an empty container. It is the *same* mistake as the one we are fixing: something empty
being counted as something present. The rule we wrote in advance says that when this report fires
this way, we must not switch on the stricter save-refusal behaviour, so I have not. Whether to
tighten the report itself is your call, because it would change how a fleet-wide check behaves.
