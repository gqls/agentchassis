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
