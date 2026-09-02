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
