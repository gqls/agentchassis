# SUMMARY — bugfix_424_logo_transparency, 2026-09-03

## What we're trying to do

Give every site's logo a genuinely transparent background, because the image model this platform
uses cannot produce one on its own — and make sure the platform never again silently ships a logo
that looks fine to the system and isn't.

## Where we've come from

Someone asked for a transparent logo background. The model painted a grey-and-white checkerboard
instead — the picture of transparency, not the real thing — and the platform stored it and served
it as if it had succeeded. Nobody had asked whether the model was even capable of transparency
before asking it for one.

## What we've done

Confirmed, both by testing the code and by checking outside sources, that the model genuinely
cannot do this — it isn't a wording problem, it's a hard limit. So instead of asking for
transparency, the system now asks the model to paint a very specific, deliberately unnatural
colour — bright magenta — as the background, and a small piece of new code removes that exact
colour afterwards, mathematically, turning it into a real see-through area. It only removes colour
that's actually connected to the edges of the picture, so it can't accidentally punch a hole in the
middle of a logo that happens to use a similar shade. And if the model doesn't paint the colour the
way it's told to, the system now refuses to save the result at all, rather than quietly keeping
something worse than what was already there.

That went through this platform's own review process twice. The first review caught a genuine
contradiction in the instructions being sent to the model — one part of the prompt told it to paint
magenta, another part told it not to use magenta, which could have made the whole thing unreliable.
The second issue was more serious and was found by another team testing it on real sites: the
safety check that was supposed to catch a bad result was actually checking the wrong thing, so it
would confidently accept a completely failed image as a success. That let three real sites end up
with a broken, nearly-invisible logo that the platform believed was fine. Once found, it was fixed,
tested, and proven against a broken version of the code before it was trusted.

Three real customer sites and one portfolio site had already been damaged by the underlying bug
before any of this shipped. Once the fix was live — confirmed live, not just committed — each of
those sites was given another attempt, and all four now have a real, working, transparent logo,
individually checked by fetching the actual image and inspecting it, not by trusting a status
message.

Along the way, another team's testing found something else worth knowing: even when the
transparency mechanism works perfectly, a logo can still turn out nearly invisible if the mark
itself happens to be a very light colour on a light page. That isn't a flaw in what was built here
— it's a genuinely different problem (does the logo work at all, versus is the logo actually
visible once placed on a real page) — and it's now being tracked and fixed separately, with the
owner having already made a call on how to handle it.

## Where we are now

Done. All four affected sites have a verified, working, transparent logo. The fix is live across
the platform, not just for these four — every future logo goes through it. This case is closed.

## Where we're going

Nothing further is owed by this piece of work. What's left belongs to other, related efforts
already under way elsewhere: making sure a logo is actually legible once placed on a page (a
different team, already started, owner has ruled on the approach), and a couple of smaller,
unrelated things this investigation happened to notice along the way (some image files being
labelled the wrong file type internally, and a temporary billing hiccup with the image-generation
provider that resolved itself the same day). None of those need anything from this piece of work to
move forward.
