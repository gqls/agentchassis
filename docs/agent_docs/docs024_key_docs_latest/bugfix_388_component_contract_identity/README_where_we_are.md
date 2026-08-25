# Where we are — bug 388, the component writer and the component store disagreeing about which component they mean

Plain prose, append-only, newest at the bottom. No jargon where a sentence will do.

---

## 2026-08-25, morning — picking it up

**What the system does here, first, because none of the rest makes sense without it.**

When a site needs a section it doesn't have a design for — a pricing block, a hero, a calculator — the
platform asks an AI writer to build one. That writer is called `component-creator`. Before it writes
anything, the platform looks up whether a component like this already exists, and if one does, it
tells the writer: *this already exists, and pages across the estate have stored their content under
these exact field names, so reuse them or you'll strand everybody's content.* After the writer
finishes, a second piece of code takes what it wrote and saves it — overwriting the existing component
if this was a rewrite, or filing a new one if it wasn't.

So there are two steps: **the advice before**, and **the saving after**. They have to be talking about
the same component. That is the whole of it.

**What the bug says.** They pick that component by different means. The advice looks it up by the
*section type* — "the pricing block" — and among several candidates picks the one most sites actually
use. The saving looks it up by the component's *function name*, which is a separate column and does
not always match. On 27 of the 120 section types we have, those two ways of looking land on different
components. So the writer can be told to preserve one component's field names, and then have its work
judged against a different component's field names — and be refused for breaking a promise it was
never asked to make.

**What I found when I checked, and it is not quite what the bug says.**

The bug was filed yesterday by a lane that was careful enough to mark its own weak spot: it said, in
effect, *I inferred how the saving step picks the component; I did not read that code end to end.*
That marker was well placed, because that was the part that was wrong.

The saving step doesn't derive the name itself. It uses **whatever name the AI writer put in its
output**. And three days ago somebody added a line to the writer's instructions that says: *set your
function name to exactly this one, don't choose a different one, because a different name silently
creates a duplicate instead of rewriting the existing component.*

So the two halves are joined after all. **But the join is a sentence in a prompt.** It is not code. No
part of the system checks that the writer complied, records it when it doesn't, or stops anything if
it does. And the sentence only appears when there are field names to show — so a component that exists
but has an empty schema gets picked, and the writer is never told its name. Five components are in
that state today. Four of them happen to have matching names anyway, so nothing is currently going
wrong. That is luck, and I'd rather not keep relying on it.

**How much has this actually cost us? Nothing that I can measure, and I want to be straight about
that.** I looked for damage in the error log and found none since the fixes that landed on 19 and 22
August. I did find two duplicate components from May that carry the exact fingerprint of this
fault — the same section type saved twice under two names, months apart, each one a paid AI
generation — but those predate everything and can't be pinned on today's code with confidence.

I also nearly got this badly wrong. A neighbouring lane offered me eight failed jobs that looked like
a perfect match, and I said so before checking the dates. The dates killed it: the component that made
those eight look like this bug was created four days *after* they failed. They were a different fault,
since fixed. I've written that up properly, because the mistake is more instructive than the finding —
I used a photograph of today to explain something that happened last week, and the neighbouring lane
hadn't done anything wrong either. Its statement was true; it was just true *now*.

**So what is this bug, honestly stated?** It is a door standing open, not a fire. If the writer ever
ignores that one sentence — and we have exactly eleven observations of it obeying, which is not enough
to promise anything — then on 15 of those 27 section types we get a loud refusal that blames the wrong
thing, and on the other 12 we silently get a duplicate component that nobody is told about. It is also
a door that widens on its own: every new component named to a slightly different convention adds one.

**What I think we should do, and I've asked for it to be argued against.** The bug's own top-ranked
fix is, I believe, backwards — it would have the advice step defer to the saving step, and the saving
step is the one that can't see the problem. The better direction is the reverse: decide which component
we mean **once, in code, before generation**, and have the saving step honour that decision instead of
re-reading a name out of the AI's output. Identity should never have come from a language model in the
first place. There is a precedent in the estate for exactly this, from the components programme:
*convert by row id, never by function name.*

I have a second model drafting that plan with instructions to try to refute me, and a separate
diagnosis run reading the same code independently. Neither has reported yet. I've also spoken to the
four neighbouring lanes whose work touches these files; none of them is blocked by this and two came
back with things I'd have hit later — one has a migration waiting to drop a column I must not
reintroduce, and another warns that its retry detector compares refusal messages byte for byte, so if
I change the wording of an error I have to say so out loud.
