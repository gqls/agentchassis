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

## 2026-08-16 afternoon — the council said "revise", and the revisions were cheap and right

The review came back "revise" — not because the fix was wrong, but because the write-up had
not SHOWN the one thing the whole design leans on: that the new list-side matching pairs
sections with locked rows in the same order the existing save-side guard does. It does; the
resubmission now quotes the guard's code line by line and explains the one deliberate
difference (the list has no component ids, so a name arm stands in for the guard's id arm — and
why that is what prevents a duplicate row rather than causing one). Three reviewers also asked
for things I agreed with and did: if the new lock lookup ever fails mid-build, the build now
leaves a permanent record of having skipped the merge instead of just a log line; the two
"which rows are locked?" queries (row-side and list-side) are now mechanically pinned to the
same lock test by a test; and there is a test proving that pages WITHOUT locked sections behave
exactly as before except that the cache write is now a genuine no-op when nothing changed. Both
rounds' code is committed; the second review is running on the same trail.

## 2026-08-16 — approved on the second round

The council approved the revised submission (three advisory notes, none blocking). The two
worth keeping in view: other parts of the system also write a page's cached section list and can
still write it without the locked section between builds — the next build corrects it, and the
health check now tolerates that difference, but it is a window; and nothing FORCES a future
reader of section lists to use the new merge — the register names today's readers, and if a
new one appears the mechanical answer is a test that scans for readers. Nothing left to do on
this bug until the next fleet build lands; then the contact-page acceptance run closes it.

## 2026-08-16 late afternoon — the council approved it, and half of it is now actually running

Two things happened since the last entry, one good and one worth being careful about.

The review came back **approved**. Fourteen reviewers, no blocking objections, three advisory
ones I've written down in full. The most substantial is from the architecture reviewer, and it
is a fair point rather than a complaint about this change: the fix adds two shared functions
that two different parts of the code now call to get "the page's real section list", but nothing
*forces* a future piece of code to call them. So someone could, next month, write a third reader
of section lists and be blind to locked sections all over again — which is exactly how this bug
happened the first time. The reviewer's own recommendation was to ship this and raise that
question separately, which is what we're doing; I have not written that proposal yet and have
said so rather than leaving it implied.

The second thing: a fleet build went out at lunchtime, and **the first half of the fix is now
live**. I checked that at the running program itself rather than trusting the version number —
asked the binary whether it contains the new code, with two control checks in the same breath to
prove the question was capable of coming back "no". It contains this morning's half; it does not
contain the extra safety net I added this afternoon, which waits for the next build. That matches
the clock exactly, which is reassuring.

**But live is not the same as working, and I want to be blunt about that.** Since the build went
out, twenty pages have been rebuilt, and every one of them was a page with no locked sections —
so the new code ran, did the correct nothing, and reported nothing. The part that actually
matters, slotting a locked section back into the list, has not happened once in production yet.
Related: there have been no new "the pipeline tried to remove a locked section" complaints since
the build — and that number is worthless as evidence, because none of the affected pages have
rebuilt. Nothing has asked the question, so "no complaints" is silence, not success. I have
recorded it that way everywhere rather than letting a zero look like a result.

What that leaves is one real test: rebuild a page that *does* have a locked section. The obvious
one is the webdesign.uk contact page with the chat box, and the recipe is written out ready to
run. I have not fired it, for two reasons: it rewrites and republishes the other sections of a
live shopfront page, and you assigned that acceptance run to the lane that owns the chat box.
The other way it gets proven is simply the next time one of the twelve loan-calculator pages
rebuilds on its own — that will exercise it with nobody doing anything, and the first one to run
should show the locked calculator back in the list and file no complaint.

Also answered this afternoon, because a reviewer was right to push on it: I had checked "who
else calls this code" but not "who else writes the section list". I've now measured both. The
other writers are either the planner (whose lists the next build re-merges anyway) or paths that
create brand-new pages, where there is no locked section to lose yet. And of the six places that
can save a page's sections, only two have run at all in the last month — one of them goes through
the fixed code, and the other one builds its proposal from the page's actual rows, so it cannot
drop a locked section in the first place. That is why the complaint counts in the database are
mostly "tried to overwrite" and only a handful are "tried to remove".
