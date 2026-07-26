# SUMMARY — idea.uk (2026-07-26b)

*Second summary today. `SUMMARY_2026-07-26` (this morning) marked the content pipeline complete
and said the only thing left was "four pasted commands on the box". That turned out to be wrong
in an instructive way, which is why this one exists. Written to be read aloud.*

## What we're trying to do

idea.uk is where someone with an idea works it out properly: guides for every stage of an idea's
life, free tools that give an honest steer, and one paid product — the £29 Verified Idea Report —
that everything funnels towards. It is also meant to become the worked example the rest of the
portfolio copies.

## Where we've come from

Ten days ago a migration project. Three days ago nine sound pages with an empty Guides section.
By this morning the whole content pipeline existed — nine guides, two free tools, the paid report
extended to match its own sales page. What remained looked purely clerical: clear five stale test
orders off the box and tidy a duplicated setting.

## What we've done since

**The "clerical" job turned out to be a design gap, and that is the real story of the day.**
The tool caps itself at five simultaneous orders. Nothing ever aged an order out — so five orders
that nobody progressed had closed the service to new business permanently, and the only way to
release them was hand-editing a file the running service owns and overwrites. Clearing them by
hand failed twice for that reason before we understood why.

So rather than write a better instruction, we fixed the mechanism:

- **Orders now expire on their own.** An order waiting on us goes cold after fourteen days, one
  waiting on a customer's payment after seven; both then release their slot automatically. It is
  swept at startup, hourly, and every time the site asks how busy we are.
- **"Expired" is deliberately its own status, separate from "declined".** Declined means you
  looked and said no. Expired means it went cold. Both keep the record — nothing is deleted —
  but filing an abandoned order as a judgement you made would corrupt the very record you asked
  us to protect.
- **The site now says so when we're full.** The report page checks the real queue and, when every
  slot is busy, tells visitors honestly that there'll be a wait. It deliberately does *not* block
  ordering, because being full only stops us *starting* work — the form always worked, and
  blocking it would both misdescribe the system and throw away real demand.
- **One address, one source of truth.** You spotted that `OPERATOR_EMAIL` is the variable that
  matters. It was already right; the wart was the report renderer reading a different variable
  with its own hardcoded fallback. That now comes from the same place as everything else.

You then cleared the five orders yourself, restarted properly, and removed the duplicate setting.
The queue is open again.

## Where we are now

**The site is complete and verified**: nine guides in journey order, four tool cards (the £29
report, two free finders, the audience check), every button real, all authored content locked —
and those locks are now genuinely enforced, since the platform's lock gate went live today on the
current chassis image.

**The box is healthy**: queue open (0 of 5), correct address, and yesterday's binary running with
the extended report format.

**One thing is still pending, and it is the important one: nobody has yet received a report in
the new format.** The end-to-end run — submit, confirm, review the draft, approve, pay, delivery
— has not happened. Until it does, the biggest change of the week is unproven in production.

> **UPDATED later the same day:** it happened. The owner submitted a real idea, received the draft,
> judged it good, and declined it to close the test without self-charging. Verified against the
> stored report rather than impression: 13,227 chars of text / 20,207 of HTML, leading with "Your
> idea, assessed", carrying "Check it yourself" source links (16 in the HTML), disclosing AI use in
> the report itself, and with the old format's opening line absent. **The extended report is proven
> in production.** What remains is the second deploy (automatic order expiry).

**A second deploy is also queued**: the automatic expiry, the address fix and the capacity sweep
are written and tested but not yet on the box.

## Where we're going

In order: deploy the second batch so the queue never silently closes again; then the pipeline grows
sideways — more tools where a stage earns one, and the empty News section. Two open questions for
the owner sit behind those: the margin (each report now costs roughly double the AI spend), and
whether five simultaneous orders is still the right ceiling once the funnel starts working.
