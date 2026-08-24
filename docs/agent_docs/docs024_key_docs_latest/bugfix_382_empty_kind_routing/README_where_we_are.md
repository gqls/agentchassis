# Where we are — bug 382, the image model that nobody chose

*Plain prose, append-only, newest at the bottom. The owner's document.*

## 2026-08-24 — opening the lane

**The complaint, and what it turned out to be.** You looked at a nurse on vetcomparison.uk and
said the eyes were wrong, "reminiscent of the early image creators". They were. That particular
image was made by SDXL, an older model, the day *before* we switched heroes over to Google's
Gemini image model back in July — so on that site it was a leftover, and the vetcomparison lane
has already replaced it and checked the live file byte for byte.

What that lane found on the way is why this lane exists. It went looking for other SDXL leftovers
and found fifteen hero images made by SDXL **after** the July switch, on five different sites, the
most recent on 11 August. Nobody chose SDXL for any of them. There is a sanctioned way for a site
to ask for it — one line in the site's imagery style guide — and not one of those five sites has
it. So the switch we made in July has been quietly bypassed for real traffic ever since, and the
only way anyone found out was you looking at a face.

**Why it was invisible.** The image service picks its model from a single word — "hero", "logo",
"icon". There is a table listing every word we know, and everything in it goes to the good model.
Anything not in the table falls back to SDXL. If the word is one we simply forgot to add, the code
notices and files a warning we can query later. If the word is **missing entirely** — an empty
string — the code says nothing at all, on purpose. The comment explains why: empty means an old
caller from before we had the field, and those were assumed to be deliberate.

That assumption is what failed. The traffic coming through the empty door is not old and it is not
deliberate — it is our own hero pipeline, today, having dropped the word somewhere upstream.

**The upstream drop, found.** The word travels from the pipeline's configuration into the image
service. In the configuration there is a setting called `default_kind`, sitting on three of our
image steps, that looks exactly like the safety net for this. It is not connected to anything. It
has never done anything. Someone else discovered that on 11 August and fixed two of the three
steps that day — but their note says all the remaining steps were already fine, and one of them
was not. That one, the step that makes the per-page heroes (the hero on the About page, the
Services page, and so on), still has no word to send and still has the dead setting sitting there
looking like it does.

The proof is as clean as this sort of thing gets: on 11 August, five per-page hero jobs finished on
one site at 16:28, 16:37, 16:38, 16:39 and 16:42, and the five SDXL images on that site are stamped
16:28, 16:36, 16:38, 16:39 and 16:41. Each job completes about thirty seconds after the image it
made. All of it happened *after* the 11 August fix was applied at 13:42, which is what proves the
fix did not cover this branch.

**What I want to do about it, and why it is bigger than one line of configuration.** The obvious
move is to add the missing word to that one step. I want to do that — it is live the moment it is
applied, no rebuild needed — but I do not want it to be the whole fix, for two reasons.

First, there are two more places in the system carrying exactly the same shape (an image step with
no word at all), on the initial site-build path. I cannot tell you how often they run; the records
that would answer that are kept for about a day and they are not in today's. They are open doors
either way.

Second, and more important: the silence is the actual defect. Fixing one config row leaves the
next dropped word just as invisible as this one was. So the real fix is in the code — when the word
is missing, use the good model rather than the old one, and file a warning we can query. That
closes every door at once, including the ones nobody has audited, and it stops depending on a
future author remembering something.

There is a wrinkle I should flag rather than bury: our own debugging guide currently *instructs*
future sessions to keep this silence — it holds up the empty case as the one that legitimately
shouldn't warn, and there is a test pinning it. Three of our own artefacts agree with each other
and disagree with production. That gets a dated correction, not a quiet edit, because the reasoning
in it is sound and only the premise is wrong: a warning that fires constantly is one nobody reads,
which was true when the empty case was assumed rare.

**One more thing found in passing, on the same step.** That per-page-hero step also fails to pass
the site's identity through, so those images are generated with no knowledge of the site's palette,
its style guide, or the things it has said it wants avoided. Its sibling step does pass it. I have
written it up as a separate change rather than folding it in silently, because it changes what the
images look like and that should be a visible decision, not a side effect.

**Where I am.** The cause is found and evidenced, the plan is written, an independent diagnosis run
is in flight to try to refute it, and I have asked a second model to attack the plan for anything I
have not thought of — particularly whether switching the empty case to the good model could make
requests *fail* rather than merely improve, which would be a worse outcome than the bug. Nothing is
committed yet. I will come back to you before anything ships if the answer to that question is
uncomfortable.
