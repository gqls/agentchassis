# Gripper dossier — summary, 2026-09-03

**What we're trying to do.** Prove that one of our sites can sell real engineering
help, not just pages: a visitor on robot-hands.com describes their pick-and-place
application in a short chat, and our platform sends them a proper gripper selection
dossier — deterministic scoring, honest prose around the computed numbers, a
published report page, a link by email.

**Where we've come from.** By late July the cluster half — scoring, report building,
publishing — was proven end to end on staged fixtures. By late August the public
half existed as a second tenant on the island's API, and the pilot had been driven
by two real requests end to end in production, both branches proven. But nothing on
the site linked to any of it, and the visitor-facing widget — the chat box a real
visitor would actually see and click — had not been built yet.

**What we've done.** The widget got built and deployed, and then we found and fixed
two defects a real browser exposed that no server-side check could have caught.
First: the widget never rendered at all. The site loads its scripts synchronously,
before the page body exists, and the widget looked for its mount point the instant
it ran — found nothing, and quietly gave up. Every automated check we had said it
was fine, because the code was in the file and the mount point was on the page;
neither fact means a button gets drawn. We only found this by simulating a real
browser's loading order and watching the old code fail and the new code succeed.
Fixed with a small, well-precedented change (wait for the page, then look), applied,
rebuilt, and confirmed both mechanically and by you loading the page yourself.

Second, the moment that screenshot came back: the chat text was unreadable. The
widget had been styled assuming a white page, and robot-hands.com is dark, so its
text was rendering in colours close to invisible. We measured it properly rather
than guessing — the readability ratio was around 1-to-1 where the standard for body
text is 4.5-to-1 — fixed it with two small, deliberately different changes (the
widget's own bubbles get an explicit text colour, since it owns those; the smaller
print inherits the page's colour instead, since it doesn't), and confirmed the fix
the same way: on the live served file, with a check that the broken version is
genuinely gone and the first fix hadn't been undone.

Both changes went through the platform's reviewer council. The first round objected
that fixing one interactive widget didn't rule out the same silent-failure pattern
existing elsewhere in the site's script library — a fair question, and a previous
pass on this thread had already investigated it thoroughly and closed it (the risk
is real but dormant: eight other scripts share the shape, none of them switched on).
This session re-investigated the same question independently, without first checking
whether it had already been answered, used a cruder method than the original
investigation had explicitly warned against, and briefly filed a second report that
overstated the risk as live. That was caught and withdrawn within the same
afternoon, before it went anywhere — the honest lesson being that checking a
mechanism carefully is not the same as checking how many things it applies to, and
being careful about the first made the second feel more solid, not less.

**Where we are now.** The pilot is complete on the build side. The chat widget
renders, reads clearly, and both fixes passed a full council review — all twelve
reviewers approved, including both that had raised concerns on the first pass.
Nothing further is owed there. The only thing left is a decision that's always been
yours: whether the report page gets a quiet link from the site now, or stays
unlinked for a while longer.

**Where we're going.** Once you make that call, the longer-term plan is unchanged:
generalise the scoring engine so the next site's dossier is a configuration
exercise rather than a build, following the pattern already proven elsewhere on the
platform (one shared table of rules, not code per site). Separately, and not
blocking any of this: the platform's script bundle loads in a way that makes this
class of silent failure possible for any future interactive widget, on any site.
That's a shared-infrastructure fix, not a page fix, and it's on record as its own,
still-unclaimed piece of work.
