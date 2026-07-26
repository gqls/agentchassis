# Summary — 2026-07-26: the site is verified sound, and the work turned into platform findings

*Read-out for the owner, written to be read aloud. Previous:
`SUMMARY_2026-07-25_pipeline_reproducible_and_link_sound.md`, which claimed the
links were sound before they were. This one differs because that claim is now
actually true and independently verified, and because the day produced three
platform defects that matter more than the site does.*

## What we're trying to do

Build genuinely good consultancy-brochure components — the interactive kind the
big firms use — **through the framework**, so any site we run can have them. And
run fundamentallyai.com as a site that markets this platform's real, verified
capabilities, where every claim has a source.

## Where we've come from

Onboarded and built overnight on 20 July. A content check mistook our own approved
reference to a sibling site for contamination; that plus hosting held it back to
22 July. The five interactive components went live on 24 July, but **hand-placed** —
so a routine rebuild would have wiped them. Yesterday closed that: every page was
rebuilt by the platform and each one restored its own component from the site plan.

## What we've done

**Verified the site properly, for the first time.** Every internal link on all
seven pages, crawled live, three attempts each: **43 targets, 0 broken.** It had
been 21 of 22 broken. The verification is deliberately independent of the fix —
that matters, because my first attempt at it wasn't.

**Finished the contact page.** Your phone number is live and tappable.

**Gave the trust story its own page.** The self-correction account names
leopardessconsulting.co.uk directly and doesn't repeat the invented details as
fact.

**Found why the copy still sounded generated.** Not the model.

**Filed three platform defects and one feature spec**, each from friction hit
rather than from speculation, and contributed measurements to a fourth case owned
by another thread.

## Where we are now

Ten pages deployed. Every link resolves. Contact details display. The five
components live in the site plan, so rebuilds keep them. The site is in good order.

The more valuable position is what the work exposed, because it applies to every
site we run:

**The platform finds broken links and throws the answer away.** The content check
detected all eight broken links on the newest page, by name, at build time — then
deployed it, because they count as warnings. On the success path the findings are
never saved: they sit in run state for about a day and expire. The code comment
justifies this by saying the improvement loop will repair them, and that loop is
switched off. So nothing needs *building* to detect this class. It needs enforcing
and keeping. (`bugs_open/071`)

**Most of our sites cannot show contact details at all.** The block reads the
phone and email as top-level fields; the thing that writes that record files them
one level deeper. The lookup fails, and because a missing email is configured to
withhold the whole block, the section vanishes with no error. Across thirteen live
sites the correlation is exact: the five with the flat shape are the five with a
contact block. The default way a site is created produces the broken shape, so
**every new site is born without contact details**. (`bugs_open/072`)

**Deliberate rebuilds look like runaway loops to the housekeeping.** A job parks
build requests idle for 48 hours but measures the age of the request row, not how
long it waited. Reusing an old row — the only way to ask for a rebuild today — is
therefore stale on arrival; it parked one of ours twelve minutes after we made it.
Creating a fresh row works, and then batches are safe too. (`bugs_open/070`) The
paved road for this already exists in the code, wired end to end with a
"requested_by" field, and has no trigger and two rows ever, last used in February.
The spec is to revive it. (`features_open/021`)

**The em-dash mystery, solved.** The prompt that forbids em dashes contained
seventeen of them, fourteen in its own instructions including the heading. The
model was shown fourteen examples of the banned style in the most authoritative
text it gets. The rule also described the wrong shape: it warned about long asides,
while every actual failure was a noun-dash-gloss. Both fixed. **Whether it worked
is unmeasured** — the change is live but no page has been written since.

**The sharpest remaining gap in the brief is charts.** Your brief requires numbers
as real code-generated charts, you asked for charts and infographics, and **no
chart component exists anywhere in the fleet** to select. What we have renders
verified numbers as text, which is not the same thing. There is prior art already
scoped — the leopardess L7 component — so this is one shared build, not a new one
per site.

## Where we're going

**Two decisions are yours.** The decision-record page: the self-correction page
says the record is "something you can read", which overstates things while that
page doesn't exist — either we build it, or we soften the sentence. I've
deliberately not built it, because publishing internal review records outward is
your call. And the chart component: worth doing properly once, sourced from the
verified-facts register so a chart structurally cannot display an unverified
figure.

**Then the automation set, in the order already agreed.** The brief-fidelity
auditor is built and its findings have now been re-tested against the rebuilt site
— two resolved, imagery improved but still thin, charts and the decision-record
page outstanding. Next is the component-adoption check: the planner still never
*chooses* the new components, it only uses them where we placed them. Then the
specialist design critic on rendered screenshots. Sweep enrolment stays last until
the improvement loop is back on, though `071` sharpens why it matters.

**Three small platform fixes are queued**, all fleet-wide and none large: keep the
gate's warnings, correct the contact-field paths, and make a silently-withheld
section loud.

## The part worth saying out loud

Four times yesterday I reported something as verified when it wasn't, and every
instance was the same mistake in a different costume: **concluding from a check
that could not see the evidence.** A search pattern blind to anchored links, used
to find, fix and confirm — all three agreed, all three wrong. A database query
trimming the one field that named the cause I then guessed at twice. A dispatch
declared dead after two minutes when the normal wait is seven to nine, published
as a false landmine in two documents. And a monitor filtering for a timestamp in
the future, which reports exactly like a failure.

All four are corrected in place, with the corrections left visible, and logged in
the fleet-wide `WRONG_CALLS.md` — the tally is the useful part, not the entries.
The rule now written into this workstream's command book: **a check that shares its
logic with the fix cannot test the fix, only agree with it.** Verify against the
served artefact. And before calling something broken, establish what normal looks
like.
