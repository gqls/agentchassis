# README — bugfix_424_logo_transparency, plain-English log

## 2026-09-02

You asked for the background behind a logo not to be part of the logo. The system asked the image
model for exactly that — a transparent background — and the model drew a grey-and-white
checkerboard instead. That checkerboard pattern is the little icon design tools use to MEAN
"nothing here", but the model painted it as if it were a real pattern, in solid opaque colour. So
the logo you got has a picture of transparency behind it, not actual transparency.

I checked with the outside world today and confirmed this isn't a fluke or a wording problem: the
model family this system uses for images (Google's "Nano Banana" models) simply cannot produce a
see-through picture at all. Every image it makes is a flat, solid rectangle of colour, always. So
no amount of rewording the request would ever fix this — the fix has to be a change in how the
system handles the picture after the model hands it back, not in what we ask the model for.

What I built instead: ask the model to paint the background a very specific, unmistakable colour —
bright magenta pink, a colour no real logo would ever use — instead of asking for transparency at
all. Then, after the picture comes back, a small piece of code removes that exact colour and turns
it into real transparency, the same way a green-screen works in film. It's careful about it: it only
removes colour that's connected to the edges of the picture, so if the model happened to paint a
similar shade somewhere inside the actual logo mark, that wouldn't get erased by mistake. And if the
model doesn't paint the magenta the way it's told to — draws a checkerboard again, say, or ignores
the instruction some other way — the system now refuses to save the result at all, rather than
storing something worse than what's already live. That refusal matters because there's currently no
way to undo a bad logo regeneration once it's saved; there's no "previous version" to fall back to.

This is written as a general capability, not a one-off patch for your site — every logo the system
generates from now on goes through the same process, once it's rolled out.

Where it stands right now: the code is written and tested — I built a set of synthetic test images
in the small check-suite itself (a plain background, a background colour blended into an edge, a
colour trapped inside a shape, and so on) and confirmed the code does the right thing on each, 20
tests in total, all passing. I also caught one thing in my own first draft: the piece of code that
decides where to put the new instruction would have hidden it from the system's own audit trail —
another session working the same area caught that before I wrote any code, which is exactly the kind
of thing that going and checking with the other people working on this system is for.

What's NOT done yet: the code hasn't been built into a running service or rolled out anywhere, so
nothing has actually changed for your live site today — it's still showing the checkerboard-free
interim version another session put up earlier today (a solid near-black background matching your
header, which looks fine on the header itself but isn't truly transparent, so it would show a box on
any other background). I also haven't run the system's own internal review process on the change
yet, which I'll do before or alongside committing it, per how this platform normally checks
platform-wide code changes. And I found — separately, not something I caused — that the system
often forgets to record what TYPE of file an image is (PNG vs JPEG) when it saves it; that's real,
but unrelated to your logo, so I've written it up as its own item rather than mixing it in here.
