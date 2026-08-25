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

## 2026-08-25, later — I had this half wrong, and the correction matters to what you decide

I put the work above through an independent review before handing it to you, and it found two
things. The diagnosis holds — the password tool wins the button because of that menu-order number,
and that is confirmed. But my recommendation was wrong, and there is a piece of the machine I had
not seen.

**The wrong pick writes the copy that locks it in.** When the system picks a destination, it then
tells the writer *"the destination is fixed — it is the Password Strength Physics tool"*, and the
writer produces a button that says so. Next time round, the system reads that wording first and
matches it back to the same tool — so the choice is now held in place by the words on the button,
not by the menu-order number I said to fix.

**That means correcting the number doesn't fix the buttons you actually saw.** Of the eighty-odd
wrong buttons, about sixty have generic wording and the number fix moves them. About twenty have
wording that names the password tool — and all three of the buttons you reported are in that
twenty. Those need the wording rewritten.

**Which is exactly the content pass you commissioned on 15 August.** I told you earlier that it
might be unnecessary; that was wrong, and wrong in the unhelpful direction. It is needed — but for
twenty specific buttons rather than a sweep across sixteen sites, and I can now list which ones by
query. That makes it much smaller than its own plan assumed.

**One more thing you should know before deciding anything**, which I should have put in this log
this morning: you already saw this problem. On 15 August the same symptom was measured, you
accepted it as a floor, and you commissioned that content pass to raise it. Nothing was ever run.
So your report yesterday is really you withdrawing that decision, not discovering something new —
and that is itself one of the choices below.

### The choices, one at a time

**One — should the password tool be on those three sites at all?** Deleting the page removes it
from consideration entirely. Worth knowing why it is there: it was pushed onto four sites back in
March because the tool library only had two usable tools at the time, so there was nothing else to
choose. That reason has expired — those sites now carry six to nine tools each.

**Two — should I correct the menu-order number?** One line per site. Two caveats: on
ai-agent-orchestration.com that number is also the tool's position in the visible menu, so the fix
would move the menu item; and as above it only reaches the sixty generic buttons, not the twenty
you saw.

**Three — should the framework be able to be told "never use this page as a button"?** Today that
cannot be said at all. I need to correct something I told you: I called this the only option that
stops the problem recurring. It isn't. It is a lever someone still has to pull, so the next
oddly-numbered tool wins by default until a human notices. To actually close it you would pair it
with an automatic check that flags this shape — lever plus alarm. That pairing is what I would
propose.

**Four — when to repair the eighty stored buttons?** After the others, or it rewrites the same
wrong answer. And it must be checked by looking at the live pages: another session proved this week
that this kind of repair reports success whether or not anything changed.

**Five — the commission itself: honour it, re-scope it, or withdraw it?** This is a decision about
your own 15 August instruction, which is why I have not folded it into the others. My suggestion is
re-scope: keep it, but point it at the twenty locked buttons once the number fix has cleared the
rest.

### The fourth site is the useful clue

There was a fourth site in this story — gaswholesalers.com — and I had not looked at it. It has no
password tool at all, and all six of its tools carry the ordinary menu number. So it escaped, while
still showing the *repetitive* version of the problem with a sensible destination. That is the
clean separation: repetition is everywhere and is what your content pass is really for; the
off-topic button is these three sites and is this bug.
