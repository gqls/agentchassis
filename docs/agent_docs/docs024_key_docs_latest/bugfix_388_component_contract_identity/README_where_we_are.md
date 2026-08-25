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

---

## 2026-08-25, afternoon — it's built, it's been reviewed, and it's committed

The fix is written and on the branch. What it does, in one sentence: **the component store no longer
works out for itself which stored component a rewrite should overwrite — it is told, by the step that
already worked it out, and it is told by row id rather than by name.**

Why by id turned out to matter more than I expected. I had assumed the component's "function" name
identified it. It doesn't. The lookup the store uses filters on the name and nothing else — not
whether the component is live, not even what kind of component it is — and then takes the first row
by recency. **Twenty-five names in the library are held by more than one component.** Two of them
(`site-footer` and `site-header`) are held by five each, and those five span two different kinds of
component. So the store was picking among several rows by which had been touched most recently, and
that winner can change without anybody changing any code — someone edits a sibling, and the answer
moves. A pin carried as a name could never have fixed that. A pin carried as an id does.

**The review process earned its keep, twice.**

The first round came back "revise", on the grounds that a function I was reusing writes to the
database, so calling it earlier would create stray rows. That would have been a serious problem if
true. It wasn't — the function creates a *name*, not a row, and one command over its body proved it.
I said so with the evidence rather than arguing, and added a test that would go red if that ever
stopped being true. The second round approved it.

But the second round also made a point I had genuinely missed. I had added a warning for one half of
the risky case and not the other — the half I'd covered is where a duplicate gets created, and the
half I'd missed is where an existing component gets *overwritten* by a guess. That second one is
worse. I've added it.

**And then I caught myself writing a test that couldn't fail.**

This is the bit I'd want you to know about, because it's the kind of mistake that ships. I wrote a
test to prove the new warning *doesn't* fire when there's nothing to warn about. The way these tests
work, anything the code does that the test didn't expect makes the test fail — so declaring "I expect
no warning" ought to catch a warning that fires wrongly. Except the code that records warnings is
deliberately forgiving: if writing the warning fails, it shrugs and carries on, because a failed log
entry must never break a component build. So the unexpected warning got written, the write failed, the
code shrugged, and **the test passed**. It had passed first time, too, which I nearly took as proof.

I only found out because I deliberately broke the rule the test was guarding and re-ran it, expecting
red. I got green. The test was decorative.

The fix wasn't a better test — it was moving the rule out into a small standalone function with no
database and no logging in it, so the question "does this fire?" can be asked directly. Now breaking
the rule does turn the test red. The general lesson, which I've written into the fleet-wide log: in
this part of the codebase you can prove something *did* happen, but you cannot prove something
*didn't*, because too many things quietly forgive failure. If the "didn't" case matters, the rule has
to live somewhere you can ask it straight.

**Three things I have not done, deliberately, and one is a decision for you.**

The database half — the small migration that switches the new behaviour on — is written, reviewed and
committed, but **not applied**. Applying it is a live change to how the fleet builds components, and
it's your call rather than mine. It's safe in either order relative to the code roll; neither has to
go first.

The code itself does nothing until the next chassis build, which you run.

And there's a daily audit job that keeps a hardcoded count in step with the code. I've updated the
file and committed it, but that file reaches the cluster through a config apply I haven't run. It's
not urgent — the number it would correct is 6 against a limit of 10, so the stale value can't hide a
problem — but it's outstanding and I'd rather name it than leave it.

**How we'll know it worked, when it goes live.** Not by a clean sweep — by driving one case
deliberately. Rebuild one component whose name and section label disagree (`footer` is the cheapest
example) and check that the rewrite landed on the component the writer was actually shown, that no
second copy appeared, and that the new warning recorded what the writer *would* have chosen. A count
of zero problems, on a system nobody has exercised, measures nothing at all.
