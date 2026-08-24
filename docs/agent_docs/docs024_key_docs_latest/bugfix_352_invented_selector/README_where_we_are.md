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
