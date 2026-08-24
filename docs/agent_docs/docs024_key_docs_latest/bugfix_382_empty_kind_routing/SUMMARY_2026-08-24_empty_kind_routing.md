# SUMMARY — 2026-08-24 — the image model nobody chose

## What we're trying to do

Stop the platform from quietly using its weaker image model. We have two: Google's Gemini, which
renders legible text and can be anchored to a site's own brand images, and an older one, SDXL,
which can do neither. Every kind of image we generate — heroes, logos, icons, diagrams — is
supposed to go to Gemini. A site that genuinely wants the old model can ask for it in writing, in
its own settings. Nobody has.

## Where we've come from

In July we found that hero images were still going to SDXL because of a gap in the code that
chooses the model, and we fixed it — replacing a hand-maintained list with a table the code can
actually question, so that forgetting to add a new image type would no longer be silent. That fix
was good and it worked.

It had one deliberate hole. The chooser decides using a single word — "hero", "logo" — and the fix
made it loud when that word is one we don't recognise. But when the word is *missing entirely*, it
stayed silent, on the stated reasoning that a missing word meant an old caller from before we had
the field, and those were assumed to have chosen the old model on purpose.

Nobody ever checked that assumption. This month the owner looked at a photograph of a nurse on one
of our sites and said the eyes were wrong. They were: it was SDXL. That particular image was a
leftover from the day before the July fix, and the site's own lane had already replaced it. But
looking for other leftovers turned up fifteen hero images made by SDXL *after* the fix, on five
sites, the most recent three weeks ago. None of those sites had asked for the old model. Someone
filed that as a bug, measured the damage, and deliberately left the hardest question — *which part
of our own system is dropping the word?* — for a later session, because answering it needed a read
they hadn't done.

## What we've done

Answered that question, and fixed it twice — once for the case in front of us and once for the
class.

The word travels from a workflow's configuration into the image service. In the configuration
there is a setting that looks exactly like the safety net for a missing word. It is connected to
nothing; it has never done anything. Someone discovered that on the 11th of August, fixed two of
the three places carrying it, and wrote in their own notes that everything else was already fine.
One thing was not: the step that makes per-page heroes — the hero on the About page, the Services
page — still sent no word, and still had the dead setting sitting there looking like it did.

The proof is unusually clean. On the 11th of August that earlier fix went in at 13:42. Later the
same afternoon, five per-page hero jobs finished on one site at 16:28, 16:37, 16:38, 16:39 and
16:42, and the five SDXL images on that site are stamped 16:28, 16:36, 16:38, 16:39 and 16:41 —
each job completing about thirty seconds after the image it made, all of it after the fix that was
supposed to have covered it.

So: a configuration change, live since this afternoon, that gives that step the missing word — and
the site's identity too, which it had also never been sent, meaning per-page heroes were being
generated with no knowledge of the site's palette or style rules while the main hero on the same
site got all of them. And a code change, committed and waiting for the next rebuild of the image
service, that closes the class: when the word is missing, use the good model, and file a record
saying which caller forgot, so the next one is a query rather than a complaint about a face.

We also went back and corrected our own debugging guide, which was actively instructing future
sessions to keep the silence. The reasoning in it was sound — a warning that fires constantly is a
warning nobody reads — and only the classification was wrong. It now says: you may exempt a
fallback only when you can *name* the callers that take it, not when you can imagine them.

## Where we are now

Reviewed twice, and both reviews are worth reporting.

The council of reviewers approved it on the first round. One reviewer refused to accept the
sentence "the configuration fix ships alongside", on the grounds that this exact file has a history
of exactly that promise being false — which it does, and which is why this bug existed. Another
approved and then objected past its own approval, to point out that this is the third bug in six
weeks on the same underlying seam and that none of the three was found by any check: all three were
found by a human looking at an image. That is now written up as a proposal for a human to decide.

The independent diagnosis run we fired to try to refute all this came back *unverifiable*. It ran
out of iterations without being able to read the one function at the centre of the claim, and said
in as many words: hand this to a human, do not auto-conclude. We have recorded that as a
non-result, with a table showing gap by gap how each thing it couldn't reach was verified by hand
instead.

Two other lanes improved the work. One checked its own site rather than taking our word, and
discovered that the naming pattern we had used to attribute nine of the fourteen affected images
proves nothing about which code path produced them — the same pattern appears on a site that never
used that path. Re-run properly the attribution holds, which is the uncomfortable outcome rather
than the reassuring one. Chasing that further turned up a second, smaller door of the same kind,
which we have deliberately *not* fixed, because the obvious fix would make six of eleven possible
cases worse; the refusal and its reasoning are written down so the next person doesn't undo it.

## Where we're going

One thing remains and it is not ours to schedule: the image service has to be rebuilt and
restarted before the code half does anything. Until then the bug stays open, because a fix that
hasn't shipped is still a fix that hasn't shipped.

When it rolls, three checks close it, each with a control attached so that a quiet result cannot be
mistaken for a good one: the census re-run with a demand control, so an empty result means "no bad
images" rather than "no images"; a negative control on the healthy path that predates all of this,
so a regression there could only have one cause; and a record appearing from one of the two
remaining callers we know still drop the word — or evidence that those callers never run at all.
Either answers the question; silence answers neither.
