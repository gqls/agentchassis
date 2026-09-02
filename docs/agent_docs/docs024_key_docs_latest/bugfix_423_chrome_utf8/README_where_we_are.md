# Where we are — the footer that would not save

*(plain-prose log, append-only, newest at the bottom)*

## 2026-09-02

Two sites on the fleet have a footer the system cannot update. Boxing Online's has been
stuck since Sunday; Garden Tools' has been stuck since the 23rd of August and **nobody
knew**, which is the more interesting half of this.

Here is what was happening. When the system builds a footer it puts a little "Our Services"
column in it, listing a few of the site's pages. It tidies each page's name up first by
capitalising every word. The code that did that capitalising grabbed "the first letter" of
each word — except it grabbed the first **byte**, and some characters take three bytes.
Boxing Online has a page called *"Boxing Quiz — Test Your Knowledge"*. That dash is a long
one, an em-dash, and because it has spaces round it the tidier treated it as a word of its
own. So it grabbed one third of a character, mangled it, and glued the other two thirds back
on. The result is not valid text, and the database refuses to store text that isn't valid.
The footer never saved.

That is a small bug. The expensive part is what happened next: **the code noticed the
database had refused, wrote a line in the log, and then told everybody it had succeeded.**
The build reported "done", with a single `footer: false` buried in the output and no reason
attached anywhere. That is why Garden Tools sat broken for ten days, and why two people
spent about an hour last Sunday checking locks and permissions and page content before
anyone thought to watch the logs live.

So the fix has three parts, and only the first is about the em-dash.

**One.** The capitalising is now done properly, character by character rather than byte by
byte — and not just in the footer. The same broken shortcut was written out by hand in eight
different places in the codebase. They now all call one shared piece of code that does it
right. Worth saying: we'd *already* fixed a close cousin of this bug in July, in a different
place, and nobody went looking for the rest of the family. That is exactly what happened
again here, so this time the whole family got fixed.

**Two.** When the database refuses to store something, the system now says so — through the
same channel it already uses when a template fails, so it shows up in the build's own output
and files an item somebody can see. No more silent "success".

**Three**, and this is the part I'd defend hardest: a check just before the footer is saved
that catches invalid text and **says where it is**. The database's own error names the bad
character but not its position, which on a 40,000-character page is nearly useless — it's
why finding this took as long as it did. The next time somebody writes this kind of mistake
anywhere in the footer pipeline, the error will point straight at it.

**One thing you should know, because it will look like a new problem when it isn't.** Garden
Tools has *never* had a footer stored — not a stale one, none at all. Under the old
behaviour its build quietly said "fine". Under the new behaviour a site with no footer at
all makes the build **fail**, loudly. I think that's right — a site shouldn't go live with a
missing footer, and a build that fails is better than ten more days of silence — but it is a
real change and the next person to build that site will see it. I've flagged it to the
review council as the bit I most want argued rather than nodded through.

**Not fixed yet, strictly speaking.** The code is written, tested and committed, but code
on this system does nothing until a new image is built and rolled out. Until then both
footers stay as they are, and Boxing Online keeps serving the hand-patched one.

## 2026-09-02, later — it's fixed, it's live, and both sites are repaired

The new build went out and carried the fix. I checked the running program itself rather
than trusting the version number — the giveaway is that a line of text my change *deletes*
is genuinely gone from the binary, which is harder to fake than the presence of something
new.

Then the real test. Garden Tools — the site nobody knew was broken — had had no footer
stored at all since the 23rd of August. I asked the pipeline to rebuild its chrome and it
worked: a footer, 2,427 bytes, stored cleanly. Boxing Online's followed a few minutes
later, replacing the hand-patch, and it passes the delivery check that lane set: no contact
block on the page, because that site has no email address on file.

The detail I'd point at if you only read one line: **the page title that caused all this
now renders with its long dash intact.** That matters because a footer that saved simply
because the offending title had been quietly thrown away would have looked identical in
every other respect. So the fix repaired the text rather than avoiding it.

The bug is closed.

**The part worth your attention is where I was wrong.** When the review board pushed back
on my first attempt, they asked a fair question: how much damage could this change do?
I went and measured it properly — seven pipelines use this piece of machinery, and none of
them has anything to catch a failure. So I built a safety switch, defaulted to off.

That was the wrong measurement, and it took two more rounds and a different reviewer to
see it. Seven is *how many things are affected when the switch matters*. It says nothing
about *how often the switch can matter*. The second number was one — a single row in the
whole estate. I had built a safety switch protecting seven pipelines from a population of
one, and in doing so I'd withheld a protection we'd already agreed we wanted.

It survived that long because it looked rigorous. It had a real query and a real date
attached, and everyone downstream — me included — treated that as proof it was the right
question. It's written up in `016b` and in my own notes, because I don't think it's a
one-off habit.

Thank you for breaking the tie on the last one, by the way. Three reviewers had pushed in
three different directions across four rounds, and that's exactly the situation our own
notes say shouldn't be settled by going round again.
