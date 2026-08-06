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
