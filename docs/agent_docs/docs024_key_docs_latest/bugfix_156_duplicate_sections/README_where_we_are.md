# Where we are — the duplicate sections bug (156)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04, evening — what this is and why it was worth doing

Back on the 28th of July, vonc.com's About page went out with everything on it printed twice.
Not a rendering glitch — the database genuinely held twelve sections where the plan said six,
in six identical pairs. A visitor read the whole page, then read the whole page again. It
stayed like that for two days and we only found it because somebody sat down and counted by
hand.

That page was fixed at the end of July, and a separate thread built a detector that would spot
it happening again after the fact. What nobody had done was stop it happening in the first
place. That is what this session did.

**The first thing worth saying is that nothing is broken right now.** I re-ran the census
across the whole fleet before starting: there are twelve pages that use the same kind of
section more than once, and in eleven of them the two copies say different things, which is
completely legitimate — a page can perfectly well have two text blocks. Not one page in the
estate currently has the identical-twice problem. So this change repairs nothing. It exists so
that the next time something upstream hiccups and hands the save a doubled list, the page does
not go out doubled.

That distinction matters for how you judge it: the only way a change like this can hurt is by
deleting something it should have kept. So the whole design is about making that impossible
rather than unlikely.

## The obvious fix is wrong, and so was the one written in the bug report

The obvious fix is to tell the database "a page may not have the same kind of section twice".
That would immediately break eleven live pages. Fine — the bug report already said so, and
proposed instead that we compare the *content* of the two sections and only remove a genuine
copy.

Here is the thing I did not expect. The bug report's own suggested comparison would have
deleted a live section too, and the evidence was in the same document, forty lines further up.
It suggests comparing the structured content of two sections. But there is a page on
finetuning.uk where two sections have **no** structured content stored at all — the field is
empty on both. Compare on that field alone and the two look identical, so one gets deleted,
even though the actual visible markup is completely different.

The two halves of that document were written for different readers — one warning a future
person *measuring* the problem, the other instructing a future person *fixing* it — and
nobody had ever put them side by side. I only caught it because I read the whole file before
writing code, which is not something I can claim as a method so much as luck plus the house
rule that says read the whole file.

**So the rule we shipped is stricter than either suggestion.** We only remove a section when
it would have been written to the database as an exact copy of one already going in — same
slot, same markup, same component, same content. Put simply: if the two rows would have been
completely indistinguishable once saved, one of them is redundant by definition and removing
it cannot lose anything. If they differ in any way at all, both are kept.

## Three things we learned that the bug report did not know

**One.** When a doubled list comes through, it does not just create a doubled page — it makes
four *other* safety checks report the wrong number. There is a check that refuses to overwrite
a good page with a much thinner one; with a doubled list it sees twice as much text as really
arrived, so a page that had genuinely been gutted could slip past it. There is another that
asks "did this rebuild see enough of the page to be replacing it?" — same problem. So putting
the de-duplication first does not just fix duplicates; it makes several existing guards
measure the truth for the first time.

**Two, and this is the one that surprised me.** If a human has locked a section — told the
system "leave this alone, I wrote it" — and a doubled list arrives, the system was quietly
putting a *second* copy of that section next to the locked one. The first copy hits the lock
and gets discarded, as designed; the second copy sails straight past, because nothing was
looking for it any more. So this was not only an agent-content problem. Nobody had noticed
that.

**Three.** There is already a tool that removes these duplicates after the fact, and it
deliberately refuses to touch a repetition that the site's own plan asks for. Our new check
had to respect exactly the same rule, or the two halves of the system would disagree about the
same question — one adding what the other refuses to remove. It does. But the two fail in
opposite directions on purpose: if the plan cannot be read, the after-the-fact tool stops
(because it is about to delete something), whereas ours simply does not de-duplicate (because
*not acting* is the safe direction when your action is a deletion).

## What we can prove, and what we cannot yet

Fourteen tests, and — more to the point — I broke the code seven different ways on purpose and
checked that the right test went red each time. Two of those deliberate breakages are the two
wrong answers described above: reduce the comparison to "same kind of section" and four tests
fail; drop the markup from the comparison and the finetuning-shaped test fails. So the tests
are not merely detecting broken code, they are detecting the two plausible wrong designs.

What we cannot say yet is that it works in production. It is committed but the fleet has not
rebuilt yet, and on this system a change is not real until the image ships. The bug therefore
stays open. When it does ship, the check is behavioural, not cosmetic: feed a save a doubled
list, confirm the page comes out single, and confirm one durable record was written — plus the
control, a legitimate repeated-section page that must come out untouched with no record at
all. A guard that removes nothing looks exactly like a guard that is not running, unless you
also show it leaving the legitimate case alone.

## The other half of the value: we will know who did it next time

The original investigation could not find out what produced the doubled list, because the
evidence had already been deleted — the working data from a build is cleared after about a
day, and the incident was found two days later. That is recorded in the bug file as
"unrecoverable", and it is why nobody has ever fixed the actual cause.

So the new check writes a permanent record when it fires, carrying the things that
investigation wanted and did not have: which code path built the list, which workflow step and
which work item asked for it, and — the useful one — whether the duplicates were adjacent
(A, A, B, B) or a repeat of the whole run (A, B, A, B). That distinction is precisely what
ruled out one theory last time. Nobody has to remember to collect it now.

## One thing I got wrong, and it is slightly embarrassing

There is a rule here that when you add a shared mechanism, you register it in the platform's
concept register **in the same commit** as the code. I did not — the register entry went in
the commit after.

The embarrassing part is that another session had made precisely this mistake a few hours
earlier and written it up, and I read their write-up while deciding whether my change needed a
register entry at all. I took the fact I wanted from it (yes, guards like this get registered)
and walked straight past its actual lesson. Both are now logged. Two lanes making the same
mistake on the same day, one of them having read the other's account of it, is not really two
mistakes — it is a missing automatic check, and I have said so where that argument belongs.

---

## 2026-08-05, late morning — it works, proven on the real page, and the ticket is closed

The overnight build shipped it. I checked that properly rather than trusting the version number
— a new build is not evidence your own change is in it — by looking inside the running binary on
both machines for three strings my change added, plus one string it deliberately does not add
and one that was already there. The three appeared, the other two behaved, so the code is
genuinely running.

That still only proves the code is *present*. To prove it *works* I put the fault back on the
page it originally happened to — vonc.com's About page — and watched.

**I did a dry run first, and I'm glad I did.** Before creating any duplicate I ran the same
rebuild with nothing wrong, to answer two questions. Does an ordinary rebuild of this page
change it anyway? (No — it came back byte-for-byte identical, so anything that changed later
would be down to my guard and not to the rebuild.) And does the queue that picks these jobs up
actually work today? (Yes, about two minutes.) That second one mattered: if the queue had been
dead, I'd have put a fault on a live page with nothing coming to fix it.

The dry run turned out to be worth more than that, because it is also the proof that the guard
keeps quiet when it should. It removed nothing and wrote nothing. Then I duplicated a section
so the page had seven where it should have six, ran the same job again, and it came back with
six — the right six, in the right order — and one permanent record explaining what it had
removed and why. The public page was byte-for-byte identical the whole way through: nobody
visiting the site could have seen any of this happen.

So: same page, same process, minutes apart. Silence when there was nothing wrong, and one clean
correction when there was. That is the pair that makes it believable — on its own, "the guard
removed nothing" is indistinguishable from "the guard never ran".

**One thing came out of this that no amount of testing would have found.** The record it writes
is supposed to name the job that triggered the rebuild, so that next time we can chase down what
is producing these doubled lists. It came out blank — because that particular rebuild agent
isn't configured to pass its job reference along. Every test I wrote supplies that setting
itself, so they all looked fine. Only running it for real showed the field arriving empty. It is
a small settings change on another part of the system, so I have written it down as the next
job rather than reaching into someone else's configuration; but I'd rather record it than quietly
leave a blank in the thing I said was half the point.

The ticket is closed and moved. Two things are left, and neither is mine to decide: that settings
fix, and the bigger question the reviewers raised — our guard protects one of seven places that
write page sections, and the fact that the other six can't cause this problem is true of the six
that exist today, not a rule anything enforces. Closing that properly means a database-level
constraint, which is a change to a busy table and deserves its own piece of work.
