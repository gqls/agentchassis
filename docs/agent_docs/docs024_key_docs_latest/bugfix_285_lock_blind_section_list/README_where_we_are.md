# README — where we are (bugfix 285, the section-list-assembler-is-lock-blind case)

Owner's plain-prose log. Append-only, newest at the bottom. (The OTHER bug 285 — the
tool-improver rewriting a shared template — is a different case with its own directory,
`bugfix_285_shared_template_write/`.)

## 2026-08-15 evening — picked up, verified still live, fleet-wide not one-site

What the bug is, in plain terms: when a page is rebuilt, the pipeline first writes down
the list of sections the page should have. It builds that list from the site plan. It
never looks at what is actually ON the page. So if a human has pinned (locked) a section
onto a page that the plan doesn't know about — the chat box on webdesign.uk's contact
page is the case in hand — every rebuild proposes a page WITHOUT it. A separate safety
guard at the very end stops the section being deleted, but it complains each time
("something tried to remove a locked section") and the page's own record of its sections
keeps saying the section isn't there.

Checked today that this is still true (it is: the contact page's list still reads
hero + contact-info, the locked chat box sits on the page as a third row, and the
"tried to remove" complaint from 13 August is still open). Then measured the fleet: this
is not just the chat box. Twelve calculator pages on loancalculator.co.uk are in the same
state, and FIVE fresh "tried to remove a locked section" complaints were filed for them
this afternoon alone. So it is a framework fault firing right now, and the fix must be
in the framework, not on one page.

Two things in the bug file turned out to be wrong when I read the actual code, and I have
recorded both: (1) it said another open bug (282, a name-resolver that ignores tool
components) had to be fixed first — it doesn't, because the step that would need it is on
a different path (the site re-plan), not the page rebuild; (2) the list the safety guard
compares against is produced one step later than the bug file says (by the content
writer, not the planner). Neither changes the fix; both change what to verify.

The fix: make the list-builder ALSO read the page's locked rows and slot them into the
list at their live position, using the very same "is this locked?" rule the safety guard
uses (one rule, not two that can drift), and make the one health check that compares
plan-vs-list understand the same rule so it doesn't start shouting "drift" on every fixed
page. Then the list tells the truth, the cache tells the truth, and the guard stops
complaining. Where we are: design done and written up (PLAN file here); implementation
starting now. Nothing committed yet.

## 2026-08-16 morning — the fix is written, tested and committed; not yet running

Done this morning: the code change is committed (`7d9b7334a`) and put to the council for review
(reference `79f70435`, verdict pending — it was mid-review at 10:07). In plain terms the change
does three things. First, when the pipeline writes down a page's section list, it now also reads
the page's locked rows and slots them in at the position they actually sit — using the very same
"is this locked?" rule the end-of-pipeline guard uses, so the two cannot disagree. Second, the
page's cached section list is now written once with that full list, and only when it actually
changed (it turned out the old "only if changed" test could never be true, so every build was
rewriting the cache). Third, the health check that compares "the plan" with "the cached list" has
been taught the same rule, otherwise on the day this goes live it would have raised thirteen
false alarms — one per page that has a locked section the plan doesn't name.

Nothing is live yet: it is a code change, so it waits for the next fleet build. The acceptance is
still the five checks in the bug file, run over the contact page after that build, by the lane
that owns the chat box. Until then the chat-box lock stays on and the bug stays open. Two things
worth telling the owner plainly: (a) after this ships, the "tried to remove a locked section"
complaints stop, but a locked section that IS in the list still raises one "tried to overwrite"
note per rebuild — that is the older guard's design, not this fix's; making locked sections
completely silent is a separate small piece of work I have written down as an open question;
(b) where earlier rebuilds already pushed a locked calculator to the bottom of its page (which
happened to two more loancalculator pages this afternoon while I was working), the fix keeps it
where it now is — putting it back is a human edit or a re-plan.

I also found and recorded two mistakes in the original bug write-up (a wrong "must fix 282
first" dependency, and the wrong step named as the one that produces the list the guard checks);
neither changed the fix, both changed what to verify, and both are now in the bug file.
