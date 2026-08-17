# Where we are — bugfix 291 (plain prose, append-only, newest at the bottom)

## 2026-08-17 — picked up, understood, and a plan that had to change once already

The tool auditor is doing good work: it inspects the little interactive tools on our sites
and writes up genuine problems (missing labels on inputs, scripts that can never load, layout
that breaks on narrow phones). But when it decides a finding needs a human to look at it, it
addresses that finding to a reviewer called "hitl-review" — and no such reviewer has ever
existed. It was an idea written down in April that nobody ever built. The system tries to
deliver the item, finds nobody home, and stamps it "blocked" forever. Fourteen findings are
stuck like that, the pile grows daily, and — worse — while a finding for a page is stuck, any
NEW findings for that same page are silently thrown away.

What we found underneath: the auditor's instructions never say what state a review item
should start in, and the system's default is "ready to dispatch". One missing line of
config. The sibling route (findings the auditor can fix itself) works fine.

The fix has three parts, and the order matters because of a trap we nearly walked into. The
obvious tidy-up — "stop naming the imaginary reviewer" — would today make the software refuse
to file review items AT ALL, and the errors would be swallowed silently. So: first a small
config change that parks review items in the "a human should look at this" state (that works
today and stops the bleed); then a code change so the platform catches this whole class at
the moment of writing (any item addressed to a nonexistent agent gets parked as blocked
immediately, with a clear note, and it un-parks itself automatically if that agent is ever
built); and only after that code is live, the cosmetic rename. The fourteen stuck findings
get repaired — they are real findings and there is already a working button in the admin
that turns a confirmed review item into a fix task.

Also worth saying: the bug write-up we inherited had two small errors in it (about what the
existing parked items look like, and about a second producer being about to blow up the same
way). We measured, corrected them in the file, and the corrections do not change the fix.

A formal diagnosis run is in flight to double-check our reading before the code ships, and
the code change will go through the review council as usual.

## 2026-08-17, end of session — the bleed is stopped and the findings are back

The config fix and the repair are live: the auditor's "a human should look at this"
items now start life parked in the human-review queue instead of being dispatched to
a reviewer that doesn't exist, and all fourteen stranded findings have been recovered
into that queue — each one stamped with a note saying what happened to it, and each
one now actionable through the existing admin confirm button (confirming one files a
fix task, which is the lifecycle these findings were always meant to have).

The deeper platform change is written, tested, and committed: from the next release,
any work item filed to a non-existent handler gets caught at the moment of writing —
parked as "blocked" with a clear note, and automatically released if that handler is
ever built. We proved the tests bite by deliberately breaking the code three ways and
watching them fail. The change is with the review council now.

Two honest confessions, both written into the fleet-wide mistakes log. First, my
initial test for "the guard doesn't overreach" couldn't actually detect overreach —
the guard's own politeness swallowed the evidence; rebuilt it so it bites. Second, I
filed the formal diagnosis and then fixed the database before the diagnosis got
around to reading it, so its report says "couldn't reproduce" on three points that
were absolutely real two hours ago — the write-ups now explain that timeline so
nobody is misled later.

Still to come (waiting on the next release rolling out): flip the now-harmless
leftover reviewer name to the standard empty value, then close the bug. The staged
script for that carries loud guards so nobody can run it early by accident.

## 2026-08-17, evening — the new build went out, and it contains none of the new code

I checked the deployed chassis before doing the last step, and it is a good job I did:
the release that went out does not contain this fix, or anybody else's code from today.

The pods really did restart, and they really are running a freshly deployed image —
but the image is labelled with the same version number as the previous one
(v1.0.1305), and when the label does not change, the machine keeps serving the copy it
already had. So: new pods, same old programme inside. I proved it two ways — the fix's
own distinctive marker is missing from the running programme, and it IS present in the
image that was built on this machine this afternoon. The two copies have different
fingerprints (6039e19c… built here, f90a7e88… actually running).

This is not just us: another lane measured the same thing this afternoon and found the
running chassis is missing **203 commits** of work from today. Anything anyone changed
in the code today is sitting inert. Anything changed in the database configuration —
including our fix that stopped the bleed and recovered the fourteen findings — is live
and unaffected, which is exactly why the situation is confusing to look at.

The remedy is one thing, and it belongs to you because releases go out fleet-wide:
**bump the version number when you release** — `make release IMAGE_TAG=v1.0.1306`.
Re-deploying at the same number will just serve the same cached copy again. I have
written the ten-second check into the fleet-wide traps file so nobody has to discover
this the slow way: compare the fingerprint of the image you built against the
fingerprint of the image that is running; if they differ under the same version
number, your code is not live no matter what anything else says.

I deliberately did NOT do the last step of our fix. It only becomes safe once the new
code is actually running — done today it would make the tool auditor fail to file any
review item at all, silently. The staged script refused to run for exactly that
reason, which is what those guards are for.

## 2026-08-17, later — this time the release really did go out, and the bug is closed

The second release (v1.0.1307) is the real thing. I checked it three ways before
touching anything, because this afternoon taught us not to trust the word "deployed":
the fingerprint of the image built here matches the fingerprint of the image actually
running; the fix's distinctive marker is present in the running programme (with a
control that proves the check could have said no); and the image's own record of which
commit it was built from contains both of our commits — and correctly does not contain
a commit made after the build.

So I finished the job. The leftover phantom reviewer name is gone: the auditor's
configuration now uses the standard "no handler, a human decides" form, and there is
no longer a single row anywhere in the work queue addressed to the reviewer that never
existed. That went in as migration 457, with its undo script written first.

Then I did the part that matters more than "it deployed": I proved the new safety net
actually catches things, in production. I sent the live system two work items in one
go — one addressed to a made-up handler, one addressed to a real one — and watched what
it did with them. The made-up one was stopped at the moment of writing, marked blocked
with a clear explanation, and never dispatched. The real one went through untouched.
That second half is the important half: without it, a net that caught *everything*
would have looked like success. Both test items were cleaned up afterwards.

**Bug 291 is closed.** The findings that were being lost are back and actionable, the
cause is fixed at the framework level rather than just for this one agent, and the
whole thing is proven at the artefact rather than asserted.

Three things stay open, and all three belong to other lanes or later work, not here:
the review-item filing key is still one-per-page (so a second, different finding about
the same page can still be squeezed out — that needs per-finding keys, which the tools
lane already owns); about forty places in the code write to the work queue directly and
bypass the new door (the older check still catches those, just later); and five other
places still name a different non-existent handler harmlessly. All are written down.
