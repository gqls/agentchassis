# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-06 — picked up the loose ends of the phantom-link bug

Yesterday a session found why several sites had buttons pointing at the wrong page:
when the system couldn't work out where a button should lead, a shared helper quietly
filled in "/contact.html" rather than leaving the button out. So a real button label —
"Read the tungsten percentage guide", say — got glued to the contact page. That helper
is now fixed, the review council approved the fix, and we've confirmed the fix is in
the software actually running in production.

What's left, and what this thread is doing: thirteen wrong buttons are still live on
seven sites, because fixing the machine doesn't retouch what it already shipped. We'll
get those pages rebuilt through the normal pipeline — first letting the link resolver
have another go at finding each button's real target, so buttons come back pointing at
the right place rather than simply vanishing.

Two things the review council asked for alongside, which we're taking on: check the
same file for OTHER quietly-invented defaults (we already found two more — a
"primary" and "secondary" button link that still default to /contact.html and
/about.html on a backup rendering path), and check why the automatic detector only
caught 2 of the 13 wrong buttons on its own.

## 2026-08-06 (evening) — what we actually found, and why we didn't change any code

Three things worth knowing, in plain terms.

**The fix is genuinely in production.** We didn't take anyone's word for it, and we didn't
rely on the version number either — a version number here tells you nothing about what's
inside it. What settles it is that our fix sits underneath another fix that a different
thread proved, at the running process, against real traffic earlier today. Since every
build is taken from committed history, anything underneath a proven change is necessarily
in the same binary.

**The wrong buttons are fewer, and less broken, than the bug file says.** Yesterday's count
was thirteen. The honest count today is four genuinely misleading buttons — a "Run the Risk
Checker" and a "Run MatchMatrix" and two similar — plus four more on a blog that say
"Get Started" and were never really written by anyone; the system invented the words as
well as the destination. The remaining ones we'd been counting say things like "Talk to us"
and "book a discovery call" and point at the contact page, which is exactly where they
should point. And the contact page exists on all seven sites, so none of this was ever a
broken link — it was buttons pointing somewhere other than what they promised.

Worth adding: the page that started the whole bug, the darts news page, quietly repaired
itself this evening. It got rebuilt through the normal pipeline on the fixed software and
the wrong link simply isn't there any more. That's the clean-up route working, on the
original casualty, without anyone hand-editing anything.

**We deliberately made no code change, and that's the interesting part.** The review
council had asked us to check whether the same "invent a plausible default" habit was
hiding elsewhere in that file. It is — in two more places. But when we looked at what
would happen if we simply deleted them, we found we'd trade a quiet wrong link for
something worse: raw template gibberish printed into the page where the address should be.
One page on another site is already doing exactly that, so this isn't theoretical. The
council reviewer who warned about this was right. Fixing it properly means changing how
that older rendering path behaves when a value is missing, which is a bigger change to
shared machinery and deserves its own review round rather than being tacked onto a tidy-up.
Both of the two remaining cases are, in any event, currently harmless: we checked every
page on the fleet and neither has produced a single bad link.

**One correction to the bug file, and one to ourselves.** The bug file blamed the automatic
detector for not running often enough. That's not true — it has run across ten sites and
filed 123 findings. The problem is that nothing ever picks those findings up: they're all
sitting in a queue waiting for a human, with no handler assigned. That's a different and
more serious problem than the one we thought we had. And a mistake of my own: I tagged a
documentation-only commit as council-reviewed, which quietly pollutes the report that
tracks which code changes went unreviewed. Logged it rather than let it pass.
