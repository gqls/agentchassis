# README — where we are (plain prose, append-only, newest at the bottom)

**2026-07-20** — Kicked off. You asked for more visually interesting components for
our consultancy-style brochure sites — the kind of thing Bain, BCG and McKinsey do:
a hero made of a few cards in a carousel that changes itself, each card with a photo
that gently zooms when you look at it instead of a video, a short title, a "read
more" link, and hardly any other text. Then further down the page, several different
kinds of blocks, a lot of them carousels you can swipe on a phone. And you want the
pages those links go to to look different from each other too, not one template
repeated.

You mentioned fundamentallyai.com as a possible new brand to try this on. I checked
it just now — right now that domain isn't yours. It's sitting on a "domain for sale"
marketplace page (Afternic), which means someone else currently owns the registration
and is trying to sell it. Before we design anything for that specific brand, you'd
need to either buy it or tell me to use a domain you already have. I haven't
assumed either way — just flagging it so we don't build a plan around a name we
don't hold yet.

I also looked at the consultancy site you already have running through this
platform — leopardessconsulting. Useful context: it deliberately uses flat
illustration-style images, not photos, partly because the current image generator
can't reliably render legible text in a photo-style image, and partly because that's
the house style you chose. If the new brand is meant to look like Bain/BCG/McKinsey
— real (or real-looking) photography, people-focused — that's a step further than
what leopardess does today, and it raises the exact question you already anticipated:
how do we show "people" imagery without our veracity checker (correctly) treating it
as inventing a person who doesn't exist. That's a real design question, not just an
engineering one, and I want your steer on it rather than picking for you.

Right now I've got two things running in the background: a deep external
research pass on what Bain/BCG/McKinsey and similar firms actually do
(specific techniques — hover-zoom instead of video, swipeable carousels, how
much text each card carries) and an internal pass mapping exactly how our own
system turns a site's spec into a rendered page, so any new component we design
slots into the real pipeline rather than being a one-off. I'll come back here once
both land with a concrete proposal for the first component or two to build, and the
open questions above for you to weigh in on.
