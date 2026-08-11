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
