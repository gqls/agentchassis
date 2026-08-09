# README — where we are (contact-block transport, bugs_open/228)

Owner's running plain-prose log. Append-only, newest at the bottom.

---

**2026-08-08/09.** Picked up bug 228 from the open backlog — it was unowned
and looked serious: on `robot-hands.com`'s actual contact page (and one other
site's page), the contact form validates properly, then after a second and a
bit shows "Your message has been sent" — and nothing is sent anywhere. No
email, no request, nothing. It's a fake `setTimeout` dressed up to look like a
network call. Anyone who used that form believed their enquiry went through
and never followed up any other way, which is about as bad as a contact form
can fail.

The good news is this platform already solved "how do you deliver a contact
form when there's no server behind the site" once before, for the sibling
`contact-form` component: it rewrites a broken form target into a `mailto:`
link built from the site's real email address, refusing to do that if the
site has no real address configured (so it never silently pretends to fix
something it can't). `contact-block` just never got wired into that existing
mechanism — its form had no destination attribute at all, and nothing was
telling it to use the same trick.

Rather than patch this one component by hand, I made the underlying
mechanism smarter: it now activates for ANY component whose template asks
for a form destination, not just ones whose content happened to be written
with that field already filled in. That's the "fix the class, not the
instance" version — it also protects whatever the next contact-style
component turns out to be.

Wrote the code, wrote tests (including a manual sanity check where I
temporarily reverted the fix and watched the new test correctly fail — cheap
proof the test isn't just green for the wrong reasons), got it through the
council's advisory review queue, and committed it.

Then came the deploy, and this is where it got slower than expected. This
platform's release process is meant to be run by you, not by an individual
session — a session tried a single-service deploy once before and it was
blocked on purpose, because rolling one service out of step with the rest of
the fleet has bitten this project before. So I asked rather than just trying
it. You said a fresh build had gone out — but when I checked the actual
running code (not the version label, the real binary), the fix wasn't in it.
The timing suggests that release was kicked off just before my commit landed,
so it built a version one commit too old. My fix is still safely sitting on
the branch, and I've already built and pushed the correct image — it just
needs the release to run once more, now that my commit isn't at the very tip
of history anymore.

Everything after the deploy is prepared and waiting: the exact database
change for `contact-block`'s form and JavaScript, and which two live pages to
refresh. Nothing has touched the live `contact-block` component yet — that
step is deliberately gated on seeing the new code actually running on the
pods first, because doing it in the wrong order would make the page briefly
*worse* than it is today (an honest-looking failure instead of today's
fake-success one, but still a failure).

**2026-08-09, later.** Good news with a twist: the bug is fixed and live, but
someone else finished it while I was still going back and forth with the
review process. The team that originally found this bug came back to fix it
themselves, without checking whether anyone else was already on it — they
hit the exact problem I'd been so careful to avoid (applying the database
change before the code fix had actually rolled out, which briefly made the
form honestly-broken instead of dishonestly-broken), then found a cleverer
way around it that didn't need my code change to be live at all for those
two specific pages.

Their version of the fix is better than the one I'd prepared — they actually
tested what a real browser does when you try to email a form submission,
which settled a question I'd flagged but not answered, and their
implementation handles more cases properly (a real server endpoint, a mailto
handoff, and an honest refusal when nothing's configured, each with its own
correct message). So I'm deferring to their version rather than swapping in
mine for the sake of it.

The one piece that's still genuinely mine is the underlying platform fix —
the reason this class of bug could happen at all — and that's now confirmed
running live across the fleet. Both sides of this ended up in a good place:
the immediate bug is fixed with better code than either of us would have
shipped alone, and the general mechanism behind it is fixed too, so the next
component built the same way won't hit this. Nothing left to do here.
