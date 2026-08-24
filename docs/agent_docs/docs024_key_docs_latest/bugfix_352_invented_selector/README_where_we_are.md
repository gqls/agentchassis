# Where we are — the fix that would have made things worse

Plain-prose log for the owner. Append-only, newest at the bottom. No jargon where plain words will
do.

---

## 2026-08-24 — opening entry

**The bug, in one go.** When the system takes a photograph of a live web page and finds text that
is too faint to read, it writes down which bit of the page was at fault so that a repair agent can
go and fix the colour. To describe the bit of the page, it records the element's "class" — the label
a designer attaches to something so it can be styled. But if the element has **no** label, the code
substitutes the element's *type* instead — "paragraph", "heading", "link" — and files it in the box
marked "label". The repair agent then reads that box, believes it, and writes a colour rule aimed at
"the paragraph labelled *paragraph*". No such thing exists on any web page. The rule is written,
deployed, and the work is ticked off as done. The text is still unreadable.

**How much of it there is.** 452 of these repair tickets exist. **181 of them name something that
cannot exist** — and **108 of those are already ticked off as fixed**. So on roughly three-fifths of
the impossible ones, we have already recorded a repair that never happened. The bug started
producing on 10 August and produced 92 more tickets in the last seven days, so it is live, not
history.

**The thing I did not expect, and it is the whole reason today took longer than an afternoon.** The
obvious fix is to stop substituting the type for the label — then the rule aims at "paragraph"
instead of "the paragraph labelled paragraph". That looks obviously better. It is worse. At the
moment the impossible rule is *harmless*, because it matches nothing and does nothing. Corrected in
the obvious way, it matches **every paragraph on the site**, and the repair agent appends it to the
site-wide stylesheet. The two commonest cases in our data are paragraphs (77 tickets) and links
(44) — which happen to be the two worst possible things to recolour site-wide. So the naive fix
converts a harmless no-op into a site-wide typography change on thirteen sites.

That reframes the job. The fix is not "stop lying about the label", it is "describe the element
precisely enough that a repair is safe" — anchored to the nearest labelled thing above it, and, most
importantly, **checked in the browser at the moment of measurement**: the code will now confirm that
the description it is about to file actually picks out the very element it just measured. That check
is the real fix, because it cannot be fooled by a future mistake in how the description is built.

**A second surprise, less interesting but more dangerous.** These tickets are identified by a key
that includes the description. Change the description and every existing ticket's key changes with
it — and the audit has a rule that says "if a page was re-measured and this ticket's problem is no
longer on it, the problem is fixed, so close the ticket". Old tickets would suddenly look absent,
and **73 live ones across 13 sites would be closed with a note saying the text now reads fine.** It
does not. Worse, you cannot fix that by being careful about the order you deploy things in — do it
the other way round and the old code closes the new tickets instead. It has to be handled in the
code, not the sequence. That is now part of the change.

**Where it needs to go, which is not where I assumed.** A chassis release does *not* carry this fix.
The measuring code lives in a separate image (the browser-runner), and only the ticket-filing half
rides the chassis. So this is a two-image change and I will say so plainly when it ships, because
"the chassis rolled" would otherwise read as "the fix is live" when half of it is not.

**Who else this touches, and they have been told.** Another team is sitting on an open decision for
you — whether to release 171 long-standing unreadable-text findings to the repair agent in one go.
**73 of those 171, on 13 of their 15 sites, are the impossible kind.** Releasing them today burns
the run; releasing them straight after a naive fix would recolour those sites. They have the number,
the query, and the warning. A third team's bug file describes "six `.H3` headings" on one site as if
`H3` were a label — it is this same substitution, so anyone following that file into the browser
would go looking for a label that was never there. They have a correction that leaves their actual
diagnosis untouched.

**One good piece of news.** A neighbouring team recorded on 22 August that the review council was
down — the model all seventeen reviewer seats use had hit a usage limit until 1 September. That is
no longer true and has not been for a while: the council has run 47 reviews in three days and
approved four changes today. Two teams were holding work on the strength of that note; both have
been told.

**Two things I got wrong today**, both caught by the team that filed this bug, both logged. I added
two numbers off my own table and got 84 where the answer was 73 — every input correct and on screen,
the sum done in my head. And I wrote it as "~84", which made a plain error look like a considered
estimate; that is the bit that would have let it travel into other people's documents. And the tool
we use to check whether a bug already has an owner told me this one was owned — by the team that had
*filed* it. It reads the commit history, and the commit that creates a bug file looks exactly like
the commit of someone working on it. One message settled it in seconds.

**Where this is going.** The plan is being drafted properly and will go to the review council
before it ships, because it touches a shared measurement path and a live database. Nothing is
committed yet beyond notes and the warnings to other teams.

---

## 2026-08-24 (end of session) — it is built, reviewed and committed, and it is not yet live

**Where it got to.** The fix is written, tested, approved by the review council first time round,
and committed. It is **not working yet**, and that is not a hedge — the code sits in two images
that have to be rebuilt and rolled before anything changes. I have been explicit about that in the
bug file, because "the chassis rolled" would otherwise read as "the fix is live" when the half that
does the measuring lives somewhere else entirely.

**What the fix actually does, in one line.** The system no longer *describes* the faulty bit of the
page and hope for the best — it now **asks the browser, on the spot, whether its description
actually picks out the thing it just measured**, and refuses to file the repair ticket if the answer
is no. That is a stronger promise than fixing the one mistake we found: it means the *next* mistake
of this kind reports itself instead of quietly producing another hundred false repairs.

**The council found four things worth having, and one of them corrected me.** I had written, in
three places, that the withdrawn tickets would come back "at the next weekly audit". A reviewer
asked what evidence there was that the audit runs weekly. There wasn't any — I had checked that
audits happen at all and quietly let that stand in for how often each individual site gets one.
Measured properly: every affected site is re-checked within a fortnight, but only three of thirteen
within a week. So the honest word is "fortnight", and it now says fortnight everywhere. Small, but
it is precisely the kind of number that gets repeated by the next person as though someone had
checked it.

The reviewers also asked me to prove something I had assumed rather than tested — that an older
component would ignore the new fields rather than choke on them. It does; I have checked it rather
than argued it. And one objection I have deliberately **not** closed: the fix counts the findings it
refuses to act on, but nothing automatically reads that count yet. That is honest bookkeeping with
no reader, and I have written it down as owed rather than quietly claiming it as done.

**A mistake worth telling you about, because it is a nice one.** I wrote a test to make sure nobody
later deletes an important line. The test looked for a distinctive piece of that line in the source
file. Then I deleted the line to check the test would notice — and it didn't. The reason: I had
also written a *comment* just above that line explaining why it must not be deleted, and my comment
quoted the very text the test was searching for. So the test found my explanation and concluded the
code was still there. **The better I explained it, the more reliably the test lied.** Fixed, and
written up — it is the sort of thing that is obvious once seen and invisible beforehand.

**Two other teams have been given real numbers.** The one holding a decision for you about
releasing 171 long-standing unreadable-text findings now knows that 73 of them, across 13 of their
15 sites, cannot be acted on at all until this ships — and that releasing them straight afterwards
without the scoping work would have been worse than leaving them parked. The other has a correction
to a bug file whose central piece of evidence was mislabelled.

**What is left, and none of it can be done today.** Rebuild and roll the two images; then confirm at
the running services rather than at git; then apply the held database change that withdraws the 73
impossible tickets; then take the measurements that would show it worked — including the one that
could show it did not. The second half of the original bug, where a correct instruction still fails
for a different reason, is untouched and stays open with its own note.

---

## 2026-08-24 (evening) — it is live, and the proof is in the post

**Both halves shipped.** The fresh build went out at 15:39 this afternoon and, unusually, it carried
*both* pieces — the measuring half and the ticket-filing half — because the browser-runner image was
rebuilt in the same round. I had expected to have to ask for a second build; I didn't.

**I checked it properly rather than trusting the version number.** Two of the three services will
tell you what they are running if you ask them; the third had already scrolled its answer off the
end of its log, exactly as our own notes warn, so I read it out of the running binary instead. Then
I did the more useful check: rather than asking *"is this the right version"*, I asked *"does this
binary contain the new behaviour"* — and confirmed a made-up string was **absent**, so the test could
actually fail. All three came back clean.

**One thing I got wrong and caught.** I paired that check with what I called a control — a second
version marker that "must not" be present. It was present. For a moment that looked like the test
being broken; in fact I had simply picked the wrong marker, because a change I made earlier in the
day turned out to predate the build rather than follow it. The control wasn't discriminating, it was
just badly chosen. Had it come back the other way, for the same wrong reason, I'd have recorded a
passing control and believed a result that proved nothing. Redone with a genuine one. It's logged,
because the lesson isn't "check your work" — it's that **a control is only a control if you check
the thing that makes it one**, and here that was a clock, not code.

**What I have not done, deliberately.** There is a database change ready that withdraws the 73
impossible tickets. Its precondition — both images live — is now met, so I could apply it. I have
not, because no page has been re-measured since the build, and I would rather see the system file
*one* correct ticket before I clear out the old ones on the strength of an argument. So I have
kicked off a re-measurement of a single site to watch. It queues behind everything else on the
fleet, so it is roughly half an hour before there is anything to read.

The site I picked is the one from the *other* team's bug — the one whose evidence describes "six
`.H3` headings" that are nothing of the sort. If the fix works, that site is where it will show
most plainly.

**Where a new session picks this up:**
`docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/HANDOFF_2026-08-24_continue_here.md`
— it has the two remaining jobs, the queries that decide them, and the ways each could come out
false.
