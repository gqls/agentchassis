# Where we are — bug 387 (plain prose, append-only)

## 2026-08-25

Someone filed a bug yesterday saying three pages on the AI-orchestration site claim to be
published but give visitors a "page not found". I checked today and the alarming half is not
true: the pages are fine. The person who filed it tested the wrong form of the address — this
site serves every page as "/name.html", and they asked for "/name". Any page on the site fails
the way they tested, including ones that have worked for weeks. This is the third time in a
month someone has made exactly this mistake, so rather than writing another warning I am adding
a small tool that looks up the real address and tests it properly. I also swept every published
page on every site at its real address: not one is missing.

The second half of the bug is real, and worse than filed: one page is showing visitors the
letters "NNN+" where a number should be ("tracks NNN+ agent types"). The instructions we give
the writing system for that site literally contain "NNN+" as a fill-in-the-number example, and
the writing system copies it out about one time in ten — and the number it is supposed to use
is never actually shown to it. The fix is in three parts: reword those instructions (immediate),
teach the publish-time checker to refuse anything shaped like a stand-in number (needs the next
software release), and — the durable part — let the machine substitute live numbers into the
instructions so a human never types a stand-in again; that last piece needs a small change owned
by another active thread, so I have written it up for them rather than barging in.

The lane that filed the bug has already accepted the correction and fixed its own records, and
caught one mistake of mine in return (I pointed at a rebuild as proof their fix worked, but the
rebuild ran three hours before their fix was deployed). Both mistakes are logged.

### Afternoon: the fix worked on the first try, and the reviewers made it better

The page that was showing "NNN+" to visitors rebuilt itself at 12:41 through the corrected
instructions, and the fix held: the live page now says "more than 150 distinct agent types"
— a real statement that stays true as the number grows — and the stand-in is gone from the
public site. No hand-editing; the system's own scheduled rebuild did it.

The safety-net change (teaching the publish-time checker to refuse anything shaped like a
stand-in number) has been through two rounds of review. The reviewers pushed back twice and
were right both times: one of my four patterns was guesswork rather than evidence, so it came
out; and the way I proposed to verify the change after release was a method known not to work
on this particular service, so it's been replaced with one proven today. Third round is in.

The lasting piece also landed: the reason sites hand-type these instruction blocks (and make
these mistakes) is that the automatic system used to DELETE a site's hand-written "never say
this" list whenever it regenerated the numbers. That's now fixed — a site can keep its
hand-written guidance and let the machine fill in the live numbers — built with the agreement
of the thread that owns that code, reviewed, and awaiting release. When the AI-orchestration
site opts in, today's interim wording fix retires.

Also of note: a second copy of this investigation was almost started by another session this
afternoon; it checked first, found this one active, and stood down — and its one question
("does the interim fix survive the nightly refresh?") is now pinned as a dated check for
tomorrow morning.
