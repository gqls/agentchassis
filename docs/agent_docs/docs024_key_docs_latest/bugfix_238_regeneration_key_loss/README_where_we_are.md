# Where we are — bugfix 238

Plain prose, append-only, newest at the bottom.

---

## 2026-08-11 — what this was, and what we did about it

**The complaint.** On the evening of the 9th, finetuning.uk's homepage started
showing five broken images in its case-studies block. The bug was filed the same
day by the lane that caused it (an improvement-loop run they'd fired for an
unrelated reason), and it was filed well — with the evidence, the live queries,
and three ranked fix options.

**The filed explanation was wrong, and that mattered.** The file said the
content generator had rewritten the section and, in the process, kept the bits
that look like writing (the image descriptions) and dropped the bits that look
like plumbing (the image addresses). That is a very plausible story about a
language model. It is not what happened. **The model was never shown those
fields and is explicitly forbidden from inventing them.** They belong to a
different category: values the system looks up for itself — an image from the
site's asset library, a link from the site's settings.

What actually happened is duller and worse. Those look-ups all failed, because
the data they point at doesn't exist for this site — and for two of them, has
never existed for *any* site we run. When a look-up fails, the default behaviour
is to quietly leave the field out. Then, when the rewritten section was saved,
the save **replaced the whole record** rather than updating part of it, so the
old values went with it.

**Why it had never bitten before.** There are two ways a page gets rewritten. One
of them merges — it keeps what's there and refreshes what changed. The other
replaces. Every rewrite of this page since May had been the merging kind, so the
image addresses survived untouched for three months and looked completely safe.
The first replacing rewrite deleted all of them at once. **The difference between
those two paths is the bug**, and neither of them is wrong on its own — which is
why nobody had spotted it.

**It was worse than five images.** Checking the live page properly turned up
something the report had missed: the same failure also removed the five "Read
case study" links and the block's main call-to-action button. Those didn't break
visibly — they simply vanished, because the template is written to hide a link
when it has nowhere to point. So the tidier, more defensive way of writing a
template turns out to fail *more* quietly than the careless way. That is a
genuinely useful thing to have learned, and I've written it down where people
will hit it.

**What we changed.** Two things, deliberately kept separate.

The first stops the loss. When one of these look-ups fails, the system now falls
back to the value the live page is already using, rather than dropping the field.
A fresh look-up always wins if it succeeds, so nothing goes stale; and if there's
genuinely nothing to fall back on, that now leaves a permanent record instead of
a log line that scrolls away. This is the fix the bug file itself asked for
first, and it's placed at the one point both rewrite paths pass through, so the
two can't drift apart again.

The second stops it shipping. It turns out the system was *already working out*,
at the exact moment of the broken render, which fields would come out empty
inside an image or link address — and then throwing that answer away. The
site-wide header/footer renderer uses it; the page renderer didn't. Now it does,
and can refuse to publish a section that would go out broken. That one is
switched **off** by default and ships with its switch held back until we've
confirmed the new code is actually running — because turning it on means a page
in the broken state can't be rebuilt until its data is fixed, and that's a real
cost that should be a deliberate choice rather than a surprise.

**The site itself is fixed and live.** The images are back, the five card links
are back, the call-to-action is back — verified on the actual served page rather
than in the database. That needed a judgement call I put to you: the same rewrite
had also changed *which* case studies the cards describe, so the old
image-to-card pairing no longer held. We matched them by subject instead, and the
descriptions the system itself had written for each image independently agreed
with that matching, which was reassuring. Two of the five were genuine judgement
calls and are flagged as such in the repair file. The old "read more" links all
pointed at pages that no longer exist, so rather than restore five dead links we
pointed them at the case-studies page.

**What we did not fix, on purpose.** Four other pages across three other sites
have the same damage. One is a real regression like finetuning's, but its images
don't exist to restore. The other three never had the values in the first place.
Restoring any of them would mean inventing content, so they're recorded rather
than papered over — and one of them is now the honest test of whether the fix
works, precisely because we didn't touch it.

**What's outstanding.** Both code changes are committed but **not yet running** —
they only take effect when the fleet next rebuilds its images, which isn't
something this thread does. Both have been put through the review council and
those verdicts are still coming back. And the third option in the original bug
report — having something automatically re-examine sites so this gets caught
within a cycle — remains correct and remains unavailable, because those checks
were switched off last week to save credits.

## 2026-08-20 — the detection we built nine days ago was switched off the whole time, and the repair we planned turned out not to be needed

Picking this bug back up after eight days, the first thing to check was whether it was still real.
It is, but almost nothing about it is where the file said it was.

**The detector we built was never turned on.** Back on the 11th we shipped two things: a fix that
stops the loss happening, and a detector that notices when it does. The fix has been running in
production since the 12th. The detector has been sitting in the code, switched off, ever since —
because turning it on is a separate switch, and nobody threw it. It has never once fired in the
platform's history.

The reason nobody threw it is worth telling, because it is not carelessness. Our own internal
reference card still said the code was "not yet running, waiting for the next release". That
sentence was true when it was written and became false the next day, and nobody went back to
change it. So every reader since — including the sessions best placed to throw the switch — was
told, in the place we keep our most trusted notes, that it was too early. **A note that goes stale
does not just misinform; here it quietly prevented the thing it was describing.** The note is
corrected now, with the cost written next to it rather than tidied away.

**So the detector is on** — the safe half of it. There are two halves: one just records "this page
shipped a broken link or image", the other refuses to rebuild such a page at all. We turned on the
recording half today. The refusing half stays off until the known broken pages are dealt with,
because otherwise it would block those very pages from being rebuilt, which you would experience
as things mysteriously failing. That ordering is your call from earlier today and it is now written
down in the system itself, not just in a document, so the next session inherits the reasoning.

**The repair we were going to build would have done nothing.** The plan was: find the pages that
lost data, work out which could be repaired automatically, and repair them. Before building it we
measured which pages that would actually be — and the answer was none of them. In almost every
case the page never had the data to lose: the place the value was supposed to come from has never
existed on that site at all. There is nothing to restore, and a rebuild would restore nothing.
What those pages need is for someone to supply the missing information, or for the template to
stop asking for it. That is content work and a decision, not an automated repair.

I want to be plain that this was the most valuable half-hour of the session. Building the repair
first would have produced something that looked right, passed its tests, ran cleanly, and changed
nothing at all — and we would probably not have noticed for weeks.

**And a genuinely good piece of news, which we had never actually checked.** In August we fixed a
related bug where rebuilding a page silently deleted its buttons. We believed it worked; we had
never proved it on real traffic. Today we could, by asking the archive a simple question: across
every page rebuild the platform has recorded, did any of these values ever go from present to
blank? The answer: it happened 66 times, on 11 sites, all of them between the 11th and the 14th of
August — **and not once since the fix landed on the 14th**, across more than three thousand
rebuilds. That is as close to proof as this kind of thing gets, and it had been sitting in the data
waiting to be asked.

**Where that leaves the bug.** Still open, and the reason has changed again — which is the third
time on this file, so it is worth saying what the reason is now. The prevention works and is
proven. The detection is on for the visible kind of breakage. The invisible kind — where a missing
link makes the button disappear entirely rather than leaving an empty one — still has no automatic
detector, because the check that would find it deliberately ignores this class and the scans that
would run it are switched off to save money. And ten pages need a human decision about data that
was never there. None of that is a code fix, which is why the bug cannot honestly be closed by
writing code.
