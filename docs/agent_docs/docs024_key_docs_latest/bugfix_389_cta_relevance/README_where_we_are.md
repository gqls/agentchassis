# Where we are — CTA destinations (plain prose, newest at the bottom)

## 2026-08-25 — why the password tool keeps turning up as the button

You spotted that an AI-orchestration consultancy was inviting visitors to try a password-strength
tool, and said it wasn't deliberate. It isn't, and the reason is duller and more fixable than a
bad guess by a language model.

When the framework needs a destination for a "call to action" button, it doesn't think about the
subject at all. It takes every tool or game page on the site, throws out the contact and legal
pages and any link back to the page you're already on, sorts what's left by the number that
controls menu order, and takes the first one. That's the whole decision.

On three sites — ai-agent-orchestration.com, finetuning.uk and leopardessconsulting.co.uk — the
password tool happens to carry menu-order **1**, set when it was created back in March. Every
genuinely relevant tool on those sites is numbered 6 or higher. So the password tool wins the top
button on every page, every time, and it always will until something changes.

Two things make this worse than it sounds. First, the button's wording is generated to match
whatever link it was handed, so the page reads as a deliberate recommendation rather than as
something broken — which is why it reached your eye instead of a queue. On the consultancy site it
currently says *"See a working example first: try the Password Strength Physics tool, built and
run on the same platform."* Second, it is still happening: the system wrote a fresh one of these
today. Correcting the pages by hand would be undone.

There's a neat piece of evidence about how invisible this is. Someone working on the consultancy
site earlier already decided this tool shouldn't be prominent — they removed it from the site's
menu and left a note saying "a password tool doesn't belong in the primary nav". It made no
difference at all, because the code that picks the button never looks at whether a page is in the
menu; it only looks at the menu *order*, which they left alone. The one signal a human gave was
invisible to the part of the system that most needed it.

**What I need from you.** There are four separate choices and they're genuinely different:
whether that tool should be on those sites at all; whether to just correct the three numbers
(quick, but on one of the three sites that number also moves the visible menu item, so the quick
fix inherits the same tangle); whether to change the framework so it can be told "never use this
page as a button" (the only option that stops it recurring, and the biggest); and when to
re-run the repair across the eighty-odd stored buttons — which has to come last, or it'll just
write the same wrong answer again.

If you want a recommendation: the third. Today there is simply no way to *say* "don't use this
page as a call to action" — that's why hiding it from the menu was the only move available, and
why it did nothing.

I should flag one thing I got wrong and caught. I initially thought thirteen sites showed this
same "someone hid it, the system ignored them" contradiction, and I was about to write that down.
It's not true: nearly two-thirds of all tool pages are outside the menu as a matter of course, so
that flag doesn't mean anyone made a judgement. Only the consultancy site is a documented case,
and only because whoever did it left a comment explaining themselves.

Separately: the older piece of work (277, the missing-content repairs) is still closed and I
re-checked it on this morning's new build — the repaired pages are still correct, and the queue
still has nothing unclassified in it.
