# Where we are — the contact details that were there all along

Plain-prose log, append-only, newest at the bottom.

---

**2026-07-31, afternoon.**

I picked up bug 072. The complaint behind it is old and simple: you gave us a
phone number for fundamentallyai.com back on the 24th, and it never appeared on
the contact page. Somebody investigated on the 25th and wrote a good bug file
about it, and their explanation was this — the thing that displays contact details
asks the system for "the site's email" at one address, and the thing that
researches the site writes it to a slightly different address. Two halves of the
machine disagreeing about where a fact lives. They proposed fixing it by making
the reader look in the right place.

It's a real disagreement. It just isn't why your sites have no contact block.

The thing that convinced them was a striking pattern: contact details showed up on
exactly the five sites that had the fact stored the "flat" way, and on none of the
eight that had it the "nested" way. No exceptions either direction. That looks
like proof.

It's a coincidence with a boring explanation. I put all the places a contact email
could be stored side by side, for every site, and the eight "broken" sites don't
have the email in the nested place either. **The nested box exists on almost every
site and it's empty.** So the fix as filed would have moved the reader from one
empty shelf to another and changed nothing on any of the eight sites.

The five that work are simply the five where somebody, at some point, typed the
email in by hand.

**So where is the data?** It's in the site's own record — the plain `sites` row
with columns for email, phone, address, company name. Twelve of your fifteen sites
have an email sitting right there, including five of the eight with no contact
block. And the bug file *says so*, in one sentence near the bottom, describing your
phone number: *"written only to `sites.phone`, which no component reads."* They
wrote down the answer and treated it as a footnote about a workaround.

That's the actual bug. The part of the system that plans a page can look up four
different kinds of data, and the site's own record is not one of them. Your phone
number was never lost. Nothing could see it.

**The bit that makes me confident this is right, rather than just my turn to be
wrong about it:** this is the third time we've hit this. There are three code paths
that build or rebuild a page, and the other two were fixed for exactly this reason
already — one of them has a comment saying it now reads the site's own record
"making both render paths agree", from a bug filed weeks ago. This third path got
missed both times. So I'm not inventing a mechanism; I'm bringing the last one into
line with the two that already agree, on the store we'd already decided was the
authoritative one.

I also put the diagnosis through our own diagnosis service before writing any code,
because I was contradicting a filed explanation and I didn't want to be the second
person this month to be confidently wrong about this page. It confirmed it on the
first pass, in about four minutes, and went and found its own evidence — it pulled
a live row for vetcomparison.uk showing the email present in the site record and
absent everywhere else.

**What the fix buys you.** Once this ships, five sites (oufe, robot-hands,
vetcomparison, vonc, webdesign) can find their contact email and will render a
contact block when their pages next rebuild. Three cannot, because there genuinely
is no contact detail anywhere for them — gamesdesign, loancalculator, and
relojistas. That last one is deliberate: you ruled it has no contact route at all,
so it must keep showing nothing, and I've made that a test rather than a hope.

**One thing I want to flag, because it's a judgement call and it's yours.** I made
the page-planner fall back to the site record when the researched data is missing.
The alternative was to make the researcher write to the place the display asks for.
I chose the fallback because it fixes every component at once and needs no data
migration — but it has a cost: the address a component *declares* is no longer the
whole truth about where its value came from. There's a log line whenever the
fallback fires, and that's the only trace. I've written the question down in the
concept register rather than quietly deciding it.

**And one thing I deliberately did not do.** While measuring this I found something
bigger and messier: of the hundred data addresses our components declare,
**seventy-four point at categories that don't exist on any site** — things like
`nav`, `blog`, `legal`, `pricing`. Those fields can never resolve for anybody.
That's a different defect with a different fix, and folding it into a bug about
contact details would have let me make a much larger change than the evidence
justifies. It's filed on its own.

The code is committed. It's Go, so nothing changes on any live site until the
chassis is rebuilt and rolled. I've sent the change to the reviewer council and
I'm waiting on its verdict before doing anything that would disturb a live run.

---

**2026-07-31, later.**

The council approved it first time round, with four advisory notes. Two of them
were worth acting on and I'm glad they were raised.

The sharper one said: the half of my fix that reads the *researched* nested shape
fixes none of the eight sites, so it's extra scope riding along with the real fix.
Reading it back, that's a fair challenge — and rather than argue the point I went
looking for evidence. It turns out there's an action in the codebase whose entire
job is to copy those researched contact details into the site's own record. If it
ran, my nested half really would be redundant.

**It has never run.** It's wired into zero live agents, and its own header says it
"should be added as a step in the build flow" — nobody ever added it. So nothing in
the whole pipeline reads what the research agent writes, which is why a brand new
site would still come out broken. The half the reviewer wanted removed is the only
thing reading it. I've kept it and put the measurement in the code next to it, so
the next person to ask the same fair question finds the answer rather than my
opinion.

That's also a finding in its own right, and probably the more interesting one: we
have a mechanism designed exactly to keep these two stores in step, built, and
never switched on. I haven't switched it on — that's a change to live build
workflows other lanes own, and it wants a decision rather than a drive-by. It's
written down where it'll be found.

The other note asked for a proper record of which fields got their value from
somewhere other than where they claim to, instead of just a log line. Reasonable —
we've been bitten before by treating logs as an audit trail. Added.

**On closing the ticket.** I've left it open, and I want to be straight about why,
because you asked me to close it. The code is committed and approved, but it's Go —
it does nothing on any live site until the chassis is rebuilt and rolled, and until
then the bug is still reproducible on all fifteen sites. Our own rule is that
`bugs_open` answers "what is biting production right now", and this still is.

I could have rolled it myself. I didn't, because another session had a council
review in flight and a roll kills those — one thread lost a whole review that way
last week. Destroying someone else's work to tick my ticket seemed like the wrong
trade.

So the file stays where it is, but with a banner at the top that says the cause is
settled, names the commits, and lists the three checks that close it. Nobody should
re-diagnose it. Once it rolls, verifying is about ten minutes, and the five sites
need their contact pages rebuilt to actually show the block.

One thing I deliberately left for you rather than deciding: I fixed this by making
the reader look in more places. The alternative was to make the research agent
write where the readers already look. Mine fixes every component at once and needs
no data migration, but it costs something real — the address a component declares
is no longer the whole truth about where its value came from. I've written that
question into the register rather than burying it.
