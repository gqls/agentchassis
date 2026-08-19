# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-18 — picked up, plan settled

We took on bug 303. The platform has a safety check that stops a half-finished,
cut-off AI generation being saved as a live tool — a good check that has caught real damage
before. But the way it decides "this looks cut off" is naive: it counts how many times the
text `<script` appears versus `</script>` (and a few other tags), anywhere in the file. A tool
whose *own code* talks about those tags — say an HTML minifier with a comment "protect
<style> blocks" — trips the counter and is refused, for ever, with an error that blames a
truncation that never happened. The two HTML tools we have live today only survived by lucky
phrasing, and the same check runs in three other places, including the sweep that files
"truncated component" tickets and the verifier that closes them.

The fix we've planned: teach the counter to only count *actual tags in actual markup* — skip
comments and the inside of script/style blocks, the same way a browser reads the page. One
shared, well-tested function used by all four places (today one of them keeps a hand-made
copy of the list because of a code-layout constraint; the shared home removes that too).
A genuinely cut-off generation still gets caught — a cut mid-JavaScript leaves a script tag
open with no close anywhere, which the new counter still sees.

Before it ships we will re-run the calibration the check's own documentation demands: every
stored component old-vs-new, the nine known real casualties must still be caught, and the
tool that triggered this bug must now pass.

## 2026-08-18 (evening) — fixed, checked against everything we have ever stored, committed

The fix is in. We built the new counter, and before trusting it we replayed it against every
component the platform has ever stored, side by side with the old one. The result was as good as
it gets: every genuinely cut-off generation the old check ever caught, the new one still catches —
all of them, with no disagreements — and the only things it stops flagging are three components we
read by hand and confirmed are perfectly fine (each just has a code comment that mentions a tag).
Two of those three had actually been put on the human review queue as "truncated" — false alarms.
One of the queue entries even advised restoring an older version, which would have thrown away a
good current template. Those entries will clear themselves once the fix is deployed.

While we worked, the person who found the bug hit a second, nastier variant: tools whose job is to
*produce* a script tag (for example a snippet generator) are forced by JavaScript's own rules into
the exact shape the old check punishes — no rewording can avoid it. The new counter handles that
variant naturally, and we added a test so it stays handled.

One coordination note: another session was fixing a different bug in the same file at the same
moment. We sequenced the commits between us so the build never broke, and both fixes are now in.
The change goes live with the next fleet deployment; until then the old behaviour (and the
workaround) still applies. The review council has the change; verdict pending.

## 2026-08-19 — deployed, verified, closed

The fleet was redeployed overnight and we confirmed — by probing the running binaries themselves
on both server replicas, not by trusting the deploy — that the new counter is what's actually
running. Since the deploy, two new tools have been born with no false refusals, and the safety
side is unchanged (a genuinely cut-off generation is still refused, proven by the test suite that
ships with the code). The bug is closed. The one workaround people had been using — avoiding
angle brackets when describing tools — is no longer needed and has been retired everywhere it was
written down. Two review-queue entries that turned out to be false alarms are now clearly labelled
as such, so nobody follows their original (and harmful) advice; they'll clear when someone
processes them normally.

One honest footnote: while verifying the deploy, the verification instructions I'd written the
day before failed twice in my own hands — both were mistakes in the checking method, not in the
fix — and both are recorded in the shared mistakes log with the corrected method in place.
