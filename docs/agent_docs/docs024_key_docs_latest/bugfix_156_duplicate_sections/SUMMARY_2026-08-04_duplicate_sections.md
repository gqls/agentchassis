# Summary — bug 156, the duplicate sections prevention gap (2026-08-04)

First in the series. Written at the point where the fix is built, reviewed and approved, and
waiting only on a chassis roll.

## What we are trying to do

Stop a page being saved with the same section on it twice. Not "detect it afterwards" — that
already exists — but make the state impossible to create in the first place, at the one point
in the system every page-composition write passes through.

## Where we have come from

On the 28th of July, vonc.com's About page was saved with twelve sections where its plan said
six: six identical pairs. A visitor read the entire page, then read the entire page again. It
stayed live for two days and was found only because somebody counted the rows by hand.

That page was repaired on the 30th. On the 31st a different thread built a checker that finds
this shape after the fact, plus a repair for it. Both good, both after the horse has gone. The
bug file was narrowed to what neither of them covers: **nothing stops the state being created**
— there is no database constraint, and although the save has seven separate guards on it,
every single one compares the incoming sections against what is *already stored*, or against a
threshold. Not one of them compares the incoming sections against *each other*. A list where
every section appears twice sails through all seven.

Nobody had picked that up, and it had sat unowned since the 31st.

## What we have done

Built the guard, and — more usefully — got the identity rule right, which turned out to be the
whole problem.

Two obvious answers are both wrong. Telling the database "no page may use the same kind of
section twice" breaks eleven live pages that legitimately do exactly that with different
content; that trap was already recorded. The second is the one nobody had spotted: **the bug
report's own recommended fix would have deleted a live section.** It says to compare the
structured content of the two sections — but the same document, forty lines earlier, records a
page on finetuning.uk where two sections have no structured content stored at all. Compare on
that alone and the two look identical, so one gets deleted, even though what they actually
display is completely different. One half of the document was written to warn a future person
*measuring* the problem, the other to instruct a future person *fixing* it, and nobody had ever
read them against each other.

So the rule we shipped is stricter than either: a section is only removed when it would have
been written to the database as an exact copy of one already going in — same slot, same markup,
same component, same content. If the two rows would have been literally indistinguishable once
saved, one is redundant by definition and removing it cannot lose anything. Any difference at
all, however small, and both are kept.

Along the way the work turned up three things the bug report did not know. A doubled list does
not merely duplicate the page — it makes *four other safety checks* report the wrong number,
including one that could then let a page that had genuinely been gutted slip past. If a human
has locked a section, a doubled list was quietly putting a second copy of it next to the locked
one. And the existing after-the-fact repair refuses to touch repetition the site's own plan
asks for, so our guard had to honour precisely the same rule or the two halves of the system
would disagree — it does, but they fail in opposite directions on purpose.

Fourteen tests, and the code was deliberately broken seven different ways to confirm the right
test went red each time. Two of those breakages are the two wrong designs above, so the tests
distinguish the shipped rule from the plausible wrong answers, not merely from broken code. The
council reviewed it and **approved on the first round**.

## Where we are now

Committed, approved, and **not yet live** — which on this system means the bug stays open. That
is measured, not assumed: both running chassis replicas were checked this evening and carry
none of the new code, with a control string proving the check itself works.

Two of the council's three advisory objections were answered with evidence and closed. The
third is a fair one and is left open on purpose: our guard sits at one of seven places that
write page sections, and the fact that the other six cannot cause this problem is a fact about
the six that exist *today*, not a rule anything enforces. A future one could.

## Where we are going

Three things, in order of who owns them.

The **roll** is somebody else's — releases here are whole-fleet and the owner runs them. When
it happens, the check is behavioural rather than cosmetic: feed a save a doubled list and
confirm the page comes out single with one durable record written, plus the control that a
legitimately repeated section is left alone. A guard that removes nothing looks exactly like a
guard that is not running unless you show it declining to act as well as acting.

The **open council question** — a database-level rule covering all seven writers rather than a
guard on one — is an owner call. It is not urgent, since nothing in the fleet is currently
affected, and it is a schema change to a busy table, so it deserves its own piece of work.

And **the original cause is still unknown**, because the evidence had already been deleted by
the time anyone looked. The guard now writes a permanent record when it fires, carrying exactly
what that investigation wanted and could not get — including whether the duplicates were
adjacent or a repeat of the whole run, which is the detail that ruled out one theory last time.
The next occurrence will be traceable. Nobody has to remember to collect it.
